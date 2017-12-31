package spacer

type LogStorageProducer interface {
	Events() chan Message
	ProduceChannel() chan Message
}

type LogStorageConsumer interface {
	Close() error
	SubscribeTopics([]string) error
	GetMetadata() error
	Poll(int) (*Message, error)
}

type Message struct {
	Topic  *string
	Value  []byte
	Key    []byte
	Offset int
}
