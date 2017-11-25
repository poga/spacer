package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var appLogger *log.Entry

var APP *viper.Viper = viper.New()
var SPACER *viper.Viper = viper.New()

var RootCmd = &cobra.Command{
	Use:   "spacer",
	Short: "serverless platform",
	Long:  `blah`,
	Run: func(cmd *cobra.Command, args []string) {
	},
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "start spacer",
	Run: func(cmd *cobra.Command, args []string) {
		run()
	},
}

var replayCmd = &cobra.Command{
	Use:   "replay",
	Short: "replay specified log",
	Run: func(cmd *cobra.Command, args []string) {
		// use an unique consumer group to replay
		s1 := rand.NewSource(time.Now().UnixNano())
		r1 := rand.New(s1)
		viper.Set("consumer_group_prefix", fmt.Sprintf("spacer-replay-%f", r1.Float64()))
		run()
	},
}

var versionCmd = &cobra.Command{
	Use: "version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("0.1")
	},
}

func init() {
	APP.AddConfigPath(".")
	APP.SetConfigName("app")
	SPACER.AddConfigPath(".")
	SPACER.SetConfigName("spacer")

	SPACER.SetDefault("consumer_group_prefix", "spacer")

	err := APP.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
	err = SPACER.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}

	RootCmd.AddCommand(startCmd, replayCmd, versionCmd)
}

func run() {
	var APP_NAME = APP.GetString("app_name")
	var DELEGATOR = SPACER.GetString("delegator")
	var BROKERS = strings.Join(SPACER.GetStringSlice("brokers"), ",")

	routes := make(map[string]string)
	// should support multiple app in one proxy
	routes["PoESocial_stat:UPDATE"] = fmt.Sprintf("%s/%s", DELEGATOR, "get_stashes")

	topic := fmt.Sprintf("^%s_*", APP_NAME)

	appLogger = log.WithFields(log.Fields{"app_name": APP_NAME, "broker": BROKERS})

	// create a producer
	producer, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": BROKERS})
	if err != nil {
		appLogger.Fatal("Unable to create producer:", err)
	}

	// generic logging for producer
	go func() {
		for e := range producer.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				m := ev
				if m.TopicPartition.Error != nil {
					appLogger.Fatal("delivery failed", m.TopicPartition.Error)
				} else {
					appLogger.Infof("Delivered message to topic %s [%d] at offset %v\n",
						*m.TopicPartition.Topic, m.TopicPartition.Partition, m.TopicPartition.Offset)
				}
				return
			default:
				appLogger.Info("Ignored event: %s\n", ev)
			}
		}
	}()

	// create a proxy to let functions write data back to kafka
	writeProxy, err := NewWriteProxy(producer.ProduceChannel())
	if err != nil {
		appLogger.Fatal(err)
	}
	go http.ListenAndServe(SPACER.GetString("write_proxy_listen"), writeProxy)

	groupID := strings.Join([]string{SPACER.GetString("consumer_group_prefix"), APP_NAME}, "-")

	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":               BROKERS,
		"group.id":                        groupID,
		"session.timeout.ms":              6000,
		"go.application.rebalance.enable": true,
		"default.topic.config":            kafka.ConfigMap{"auto.offset.reset": "earliest"},
		"metadata.max.age.ms":             1000,
		"enable.auto.commit":              false,
	})
	if err != nil {
		appLogger.Fatal("Failed to create consumer", err)
	}
	defer consumer.Close()

	err = consumer.SubscribeTopics([]string{topic}, nil)
	if err != nil {
		appLogger.Fatal("Failed to subscribe topics %v, %s\n", topic, err)
	}

	// periodically refresh metadata to know if there's any new topic created
	go refreshMetadata(consumer, appLogger)

	// start the consumer loop
	for {
		ev := consumer.Poll(100)
		if ev == nil {
			continue
		}
		switch e := ev.(type) {
		case kafka.AssignedPartitions:
			appLogger.Info(e)
			consumer.Assign(e.Partitions)
		case kafka.RevokedPartitions:
			appLogger.Info(e)
			consumer.Unassign()
		case *kafka.Message:
			appLogger.Info("%% Message on ", e.TopicPartition)

			routePath := fmt.Sprintf("%s:UPDATE", *e.TopicPartition.Topic)
			appLogger.Info("Looking up route ", routePath)

			if _, ok := routes[routePath]; !ok {
				appLogger.Info("Route not found")
				continue
			}

			err := invoke(routes[routePath], []byte(string(e.Value)))
			if err != nil {
				appLogger.WithField("route", routes[routePath]).Errorf("Invocation Error: %v\n", err)
				continue
			}
			_, err = consumer.CommitMessage(e)
			if err != nil {
				appLogger.Error("Commit Error: %v %v\n", e, err)
			}
		case kafka.PartitionEOF:
			appLogger.Info("%% Reached", e)
		case kafka.Error:
			appLogger.Fatal("%% Error", e)
		default:
			appLogger.Info("Unknown", e)
		}
	}

}

func main() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func invoke(route string, data []byte) error {
	appLogger.Infof("Invoking %s\n", route)
	resp, err := http.Post(route, "application/json", bytes.NewReader(data))

	if err != nil {
		return errors.Wrap(err, "post event handler failed")
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return errors.New(fmt.Sprintf("Function not ok: %d", resp.StatusCode))
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "unable to read function return value")
	}

	// TODO: function return {error: ...} 也要當作出錯
	var ret map[string]json.RawMessage
	err = json.Unmarshal(body, &ret)
	if err != nil {
		return errors.Wrap(err, "Failed to decode JSON")
	}

	if msg, ok := ret["error"]; ok {
		return errors.New(fmt.Sprintf("Function returned error: %s", msg))
	}
	return nil
}

type WriteProxy struct {
	produceChan chan *kafka.Message
}

func NewWriteProxy(produceChan chan *kafka.Message) (*WriteProxy, error) {
	proxy := WriteProxy{produceChan}
	return &proxy, nil
}

func (p WriteProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		appLogger.Error("read body failed", err)
		w.WriteHeader(400)
		return
	}
	var write WriteRequest
	err = json.Unmarshal(body, &write)
	if err != nil {
		appLogger.Errorf("decode body failed %v\n", err)
		w.WriteHeader(400)
		return
	}
	topic := fmt.Sprintf("%s_%s", APP.GetString("app_name"), write.Object)
	for key, value := range write.Data {
		p.produceChan <- &kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
			Key:            []byte(key),
			Value:          value,
		}
	}
	fmt.Fprintf(w, "ok")
}

type WriteRequest struct {
	Object string                     `json:"object"`
	Data   map[string]json.RawMessage `json:"data"`
}

func refreshMetadata(consumer *kafka.Consumer, logger *log.Entry) {
	for {
		metadata, err := consumer.GetMetadata(nil, true, 100)
		if err != nil {
			// somethimes it just timed out, ignore
			logger.Warn("Unable to refresh metadata: ", err)
			continue
		}
		keys := []string{}
		for k, _ := range metadata.Topics {
			keys = append(keys, k)
		}
		// logger.Info("metadata: ", keys)
		time.Sleep(5 * time.Second)
	}
}
