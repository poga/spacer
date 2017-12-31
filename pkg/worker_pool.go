package spacer

import (
	"fmt"
	"sync"

	log "github.com/sirupsen/logrus"
)

type Work func(Message) error

type Pool struct {
	workers map[string]*Worker
	Work    Work
	*sync.RWMutex
}

func NewPool(work Work) *Pool {
	p := Pool{make(map[string]*Worker), work, &sync.RWMutex{}}

	return &p
}

func (p *Pool) RunTask(msg Message) {
	workerKey := fmt.Sprintf("%s_%s", *msg.Topic, string(msg.Key))
	worker := p.getWorker(workerKey)
	worker.TaskChan <- msg
}

func (p *Pool) getWorker(workerKey string) *Worker {
	log.Debugf("getting worker %s", workerKey)
	p.RLock()
	_, ok := p.workers[workerKey]
	p.RUnlock()
	if !ok {
		log.Debugf("initializing worker %s", workerKey)
		p.Lock()
		p.workers[workerKey] = NewWorker(p.Work)
		p.Unlock()
	}

	return p.workers[workerKey]
}

type Worker struct {
	CloseChan chan int
	TaskChan  chan Message
	Error     error
}

func NewWorker(work Work) *Worker {
	w := &Worker{
		CloseChan: make(chan int),
		TaskChan:  make(chan Message),
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
