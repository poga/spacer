package spacer

import (
	"strings"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	log "github.com/sirupsen/logrus"
)

type KafkaProducer struct {
	producer             *kafka.Producer
	produceChannel       chan Message
	kafkaProducerChannel chan *kafka.Message
	eventChannel         chan Message

	logger *log.Entry
}

func NewKafkaProducer(app *Application) (*KafkaProducer, error) {
	producer, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": strings.Join(app.Brokers(), ",")})
	if err != nil {
		return nil, err
	}

	return &KafkaProducer{producer, nil, nil, nil, app.Log}, nil
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
			kp.kafkaProducerChannel <- &kafka.Message{
				TopicPartition: kafka.TopicPartition{Topic: msg.Topic, Partition: kafka.PartitionAny},
				Key:            msg.Key,
				Value:          msg.Value,
			}
		}
	}()

	return kp.produceChannel
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
					kp.eventChannel <- Message{
						Topic:  m.TopicPartition.Topic,
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

	logger *log.Entry
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

	return &KafkaConsumer{consumer, app.Log}, nil
}

func (kc *KafkaConsumer) Close() error {
	return kc.consumer.Close()
}

func (kc *KafkaConsumer) SubscribeTopics(topics []string) error {
	return kc.consumer.SubscribeTopics(topics, nil)
}

func (kc *KafkaConsumer) GetMetadata() error {
	_, err := kc.consumer.GetMetadata(nil, true, 100)
	return err
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

		return &Message{
			Topic:  e.TopicPartition.Topic,
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
