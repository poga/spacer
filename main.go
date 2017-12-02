package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	log "github.com/sirupsen/logrus"
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
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
				return
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
		"enable.auto.commit":              false,
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

	// periodically refresh metadata to know if there's any new topic created
	go refreshMetadata(consumer, app.Log)

	// start the consumer loop
	for {
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
			app.Log.Debugf("Message Received %s", e.TopicPartition)

			parts := strings.Split(*e.TopicPartition.Topic, "_")
			object := parts[1]

			routePath := GetRouteEvent(object, "UPDATE")
			app.Log.Debugf("Looking up route %s", routePath)

			if _, ok := app.Routes[routePath]; !ok {
				app.Log.Debugf("Route not found")
				continue
			}

			err := app.Invoke(app.Routes[routePath], []byte(string(e.Value)))

			if err != nil {
				app.Log.WithField("route", app.Routes[routePath]).Errorf("Invocation Error: %v", err)
				continue
			}
			_, err = consumer.CommitMessage(e)
			if err != nil {
				app.Log.Errorf("Commit Error: %v %v", e, err)
			}
		case kafka.PartitionEOF:
			app.Log.Debugf("Reached %v", e)
		case kafka.Error:
			app.Log.Fatalf("Consumer Error %s", e)
		default:
			app.Log.Debugf("Unknown Message %v", e)
		}
	}

}

type WriteProxy struct {
	produceChan chan *kafka.Message
	app         *Application
}

func NewWriteProxy(app *Application, produceChan chan *kafka.Message) (*WriteProxy, error) {
	proxy := WriteProxy{produceChan, app}
	return &proxy, nil
}

func (p WriteProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		p.app.Log.Error("read body failed", err)
		w.WriteHeader(400)
		return
	}
	var write WriteRequest
	err = json.Unmarshal(body, &write)
	if err != nil {
		p.app.Log.Errorf("decode body failed %v", err)
		w.WriteHeader(400)
		return
	}
	topic := fmt.Sprintf("%s_%s", p.app.GetString("app_name"), write.Object)
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
