package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const SpacerVersion = "0.1"

var rootCmd = &cobra.Command{
	Use:   "spacer",
	Short: "serverless platform",
	Long:  `blah`,
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
		viper.Set("consumer_group_prefix", fmt.Sprintf("spacer-replay-%d", time.Now().Nanosecond()))
		run()
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "show spacer verision",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(SpacerVersion)
	},
}

func init() {
	rootCmd.AddCommand(startCmd, replayCmd, versionCmd)
}

func run() {
	app, err := NewApplication()
	if err != nil {
		log.Fatal(err)
	}

	// create a producer
	producer, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": strings.Join(app.Brokers(), ",")})
	if err != nil {
		app.Log.Fatal("Unable to create producer:", err)
	}

	// generic logging for producer
	go func() {
		for e := range producer.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				m := ev
				if m.TopicPartition.Error != nil {
					app.Log.Fatal("delivery failed", m.TopicPartition.Error)
				} else {
					app.Log.Debugf("Delivered message to topic %s [%d] at offset %v",
						*m.TopicPartition.Topic, m.TopicPartition.Partition, m.TopicPartition.Offset)
				}
			default:
				app.Log.Debugf("Ignored event: %s", ev)
			}
		}
	}()

	// create a proxy to let functions write data back to kafka
	writeProxy, err := NewWriteProxy(app, producer.ProduceChannel())
	if err != nil {
		app.Log.Fatal(err)
	}
	go http.ListenAndServe(app.GetString("write_proxy_listen"), writeProxy)
	app.Log.Infof("Write Proxy Started: %s", app.GetString("write_proxy_listen"))

	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":               strings.Join(app.Brokers(), ","),
		"group.id":                        app.ConsumerGroupID(),
		"session.timeout.ms":              6000,
		"go.application.rebalance.enable": true,
		"default.topic.config":            kafka.ConfigMap{"auto.offset.reset": "earliest"},
		"metadata.max.age.ms":             1000,
	})
	if err != nil {
		app.Log.Fatal("Failed to create consumer", err)
	}

	defer consumer.Close()

	err = consumer.SubscribeTopics([]string{app.Subscription()}, nil)
	if err != nil {
		app.Log.Fatal("Failed to subscribe topics %v, %s", app.Subscription(), err)
	}
	app.Log.Infof("Consumer Started")

	// close consumers when user press ctrl+c
	run := true
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt)
	go func() {
		for _ = range c {
			fmt.Println("Exiting...")
			run = false
		}
	}()

	// periodically refresh metadata to know if there's any new topic created
	go refreshMetadata(consumer, app.Log)

	// start the consumer loop
	for run {
		ev := consumer.Poll(100)
		if ev == nil {
			continue
		}
		switch e := ev.(type) {
		case kafka.AssignedPartitions:
			app.Log.Info(e)
			consumer.Assign(e.Partitions)
		case kafka.RevokedPartitions:
			app.Log.Info(e)
			consumer.Unassign()
		case *kafka.Message:
			app.Log.Infof("Message Received %s", e.TopicPartition)

			app.Invoke(e)
		case kafka.PartitionEOF:
			app.Log.Debugf("Reached %v", e)
		case kafka.Error:
			app.Log.Fatalf("Consumer Error %s", e)
		default:
			app.Log.Debugf("Unknown Message %v", e)
		}
	}
}

func refreshMetadata(consumer *kafka.Consumer, logger *log.Entry) {
	for {
		time.Sleep(5 * time.Second)
		metadata, err := consumer.GetMetadata(nil, true, 100)
		if err != nil {
			// somethimes it just timed out, ignore
			logger.Warn("Unable to refresh metadata: ", err)
			continue
		}
		keys := []string{}
		for k := range metadata.Topics {
			keys = append(keys, k)
		}
		// logger.Info("metadata: ", keys)
	}
}
