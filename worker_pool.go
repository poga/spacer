package main

import (
	"strings"
	"sync"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type WorkerObjectType string
type WorkerKey string
type Work func(*kafka.Message) error

type Pool struct {
	workers map[WorkerObjectType]map[WorkerKey]*Worker
	Work    Work
	*sync.RWMutex
}

func NewPool(work Work) *Pool {
	p := Pool{make(map[WorkerObjectType]map[WorkerKey]*Worker), work, &sync.RWMutex{}}

	return &p
}

func (p *Pool) RunTask(msg *kafka.Message) {
	parts := strings.Split(*msg.TopicPartition.Topic, "_")
	objectType := parts[1]

	worker := p.getWorker(WorkerObjectType(objectType), WorkerKey(string(msg.Key)))
	worker.TaskChan <- msg
}

func (p *Pool) getWorker(objectType WorkerObjectType, key WorkerKey) *Worker {
	p.RLock()
	_, ok := p.workers[objectType]
	p.RUnlock()
	if !ok {
		p.Lock()
		p.workers[objectType] = make(map[WorkerKey]*Worker)
		p.Unlock()
	}
	p.RLock()
	w, ok := p.workers[objectType][key]
	p.RUnlock()
	if !ok {
		p.Lock()
		w := NewWorker(p.Work)
		p.workers[objectType][key] = w
		p.Unlock()
		return w
	}

	return w
}

type Worker struct {
	CloseChan chan int
	TaskChan  chan *kafka.Message
	Error     error
}

func NewWorker(work Work) *Worker {
	w := &Worker{
		CloseChan: make(chan int),
		TaskChan:  make(chan *kafka.Message),
	}
	w.Start(work)

	return w
}

func (w *Worker) Start(work Work) {
	go func() {
		for {
			select {
			case <-w.CloseChan:
				break
			case msg := <-w.TaskChan:
				err := work(msg)
				if err != nil {
					w.Error = err
					break
				}
			}
		}
	}()
}
