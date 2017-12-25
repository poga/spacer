package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

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
	topic := fmt.Sprintf("%s_%s", p.app.GetString("app_name"), write.Topic)
	for key, value := range write.Entries {
		p.produceChan <- &kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
			Key:            []byte(key),
			Value:          value,
		}
	}
	fmt.Fprintf(w, "ok")
}

type WriteRequest struct {
	Topic   string                     `json:"topic"`
	Entries map[string]json.RawMessage `json:"entries"`
}
