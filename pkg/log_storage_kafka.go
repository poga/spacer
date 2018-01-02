package spacer

import (
	"fmt"
	"strings"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	log "github.com/sirupsen/logrus"
)

type KafkaProducer struct {
	producer             *kafka.Producer
	produceChannel       chan Message
	kafkaProducerChannel chan *kafka.Message
	eventChannel         chan Message
	appName              string

	logger *log.Entry
}

func NewKafkaProducer(app *Application) (*KafkaProducer, error) {
	producer, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": strings.Join(app.Brokers(), ",")})
	if err != nil {
		return nil, err
	}

	return &KafkaProducer{producer, nil, nil, nil, app.Name(), app.Log}, nil
}

func (kp *KafkaProducer) ProduceChannel() chan Message {
	if kp.produceChannel != nil {
		return kp.produceChannel
	}

	pc := make(chan Message)
	kp.produceChannel = pc
	kp.kafkaProducerChannel = kp.producer.ProduceChannel()

	go func() {
		for msg := range pc {
			// add namespace to topic
			topic := fmt.Sprintf("%s_%s", kp.appName, *msg.Topic)
			kp.kafkaProducerChannel <- &kafka.Message{
				TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
				Key:            msg.Key,
				Value:          msg.Value,
			}
		}
	}()

	return kp.produceChannel
}

func (kp *KafkaProducer) Close() error {
	kp.producer.Close()
	return nil
}

func (kp *KafkaProducer) CreateTopics(topics []string) error {
	// TODO
	return nil
}

func (kp *KafkaProducer) Events() chan Message {
	if kp.eventChannel != nil {
		return kp.eventChannel
	}

	ec := make(chan Message)
	kp.eventChannel = ec
	go func() {
		for e := range kp.producer.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				m := ev
				if m.TopicPartition.Error != nil {
					kp.logger.Fatal("delivery failed", m.TopicPartition.Error)
				} else {
					// remove namespace prefix
					topic := strings.Split(*m.TopicPartition.Topic, "_")[1]

					kp.eventChannel <- Message{
						Topic:  &topic,
						Offset: int(m.TopicPartition.Offset),
						Key:    m.Key,
						Value:  m.Value,
					}
				}
			default:
				kp.logger.Debugf("Ignored event: %s", ev)
			}
		}
	}()

	return kp.eventChannel
}

type KafkaConsumer struct {
	consumer *kafka.Consumer
	logger   *log.Entry
}

func NewKafkaConsumer(app *Application) (*KafkaConsumer, error) {
	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":               strings.Join(app.Brokers(), ","),
		"group.id":                        app.ConsumerGroupID(),
		"session.timeout.ms":              6000,
		"go.application.rebalance.enable": true,
		"default.topic.config":            kafka.ConfigMap{"auto.offset.reset": "earliest"},
		"metadata.max.age.ms":             1000,
	})
	if err != nil {
		return nil, err
	}

	subscriptionTopicPattern := fmt.Sprintf("^%s_*", app.Name())

	err = consumer.SubscribeTopics([]string{subscriptionTopicPattern}, nil)
	if err != nil {
		return nil, err
	}

	kc := &KafkaConsumer{consumer, app.Log}
	go kc.refreshMetadata()
	return kc, nil
}

func (kc *KafkaConsumer) Close() error {
	return kc.consumer.Close()
}

// periodically refresh metadata to know if there's any new topic created
func (kc *KafkaConsumer) refreshMetadata() {
	for {
		time.Sleep(5 * time.Second)
		_, err := kc.consumer.GetMetadata(nil, true, 100)
		if err != nil {
			// somethimes it just timed out, ignore
			kc.logger.Debugf("Unable to refresh metadata: %s", err)
			continue
		}
	}
}

func (kc *KafkaConsumer) Poll(timeoutMs int) (*Message, error) {
	ev := kc.consumer.Poll(timeoutMs)
	if ev == nil {
		return nil, nil
	}

	switch e := ev.(type) {
	case kafka.AssignedPartitions:
		kc.logger.Debug(e)
		kc.consumer.Assign(e.Partitions)
	case kafka.RevokedPartitions:
		kc.logger.Debug(e)
		kc.consumer.Unassign()
	case *kafka.Message:
		kc.logger.Debugf("Message Received %s", e.TopicPartition)
		// remove namespace prefix
		topic := strings.Split(*e.TopicPartition.Topic, "_")[1]

		return &Message{
			Topic:  &topic,
			Value:  e.Value,
			Key:    e.Key,
			Offset: int(e.TopicPartition.Offset),
		}, nil
	case kafka.PartitionEOF:
		kc.logger.Debugf("Reached %v", e)
	case kafka.Error:
		return nil, e
	default:
		kc.logger.Debugf("Unknown Message %v", e)
	}

	return nil, nil
}
