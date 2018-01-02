package spacer

var memoryLogStorage map[string][]Message

type MemoryProducer struct {
	producChan chan Message
}

func NewMemoryProducer() *MemoryProducer {
	if memoryLogStorage == nil {
		memoryLogStorage = make(map[string][]Message)
	}
	return &MemoryProducer{nil}
}

func (mp *MemoryProducer) Close() error {
	// no-op
	return nil
}

func (mp *MemoryProducer) Events() chan Message {
	// noop
	return make(chan Message)
}

func (mp *MemoryProducer) ProduceChannel() chan Message {
	if mp.producChan != nil {
		return mp.producChan
	}

	pc := make(chan Message)
	mp.producChan = pc
	go func() {
		for msg := range pc {
			memoryLogStorage[*msg.Topic] = append(memoryLogStorage[*msg.Topic], msg)
		}
	}()

	return pc
}

func (mp *MemoryProducer) CreateTopics(topics []string) error {
	for _, topic := range topics {
		memoryLogStorage[topic] = make([]Message, 0)
	}
	return nil
}

type MemoryConsumer struct {
	topicOffsets map[string]int
}

func NewMemoryConsumer() *MemoryConsumer {
	return &MemoryConsumer{make(map[string]int)}
}

func (mc *MemoryConsumer) Close() error {
	// noop
	return nil
}

func (mc *MemoryConsumer) Poll(timeoutMs int) (*Message, error) {
	for topic, v := range memoryLogStorage {
		if len(v) > mc.topicOffsets[topic] {
			i := mc.topicOffsets[topic]
			mc.topicOffsets[topic]++
			return &memoryLogStorage[topic][i], nil
		}
	}

	return nil, nil
}
