package spacer

type LogStorageProducer interface {
	CreateTopics([]string) error

	Events() chan Message
	ProduceChannel() chan Message

	Close() error
}

type LogStorageConsumer interface {
	Poll(int) (*Message, error)

	Close() error
}

type Message struct {
	Topic  *string
	Value  []byte
	Key    []byte
	Offset int
}
