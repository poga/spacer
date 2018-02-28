package spacer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/pkg/errors"

	log "github.com/sirupsen/logrus"
)

type Application struct {
	config          *Config
	Log             *log.Entry
	Triggers        map[Event][]*url.URL
	WorkerPool      *Pool
	ConsumerGroupID string
	Env             string
}

type Event string

func NewApplication(configFile string, env string) (*Application, error) {
	if env == "" {
		return nil, fmt.Errorf("Invalid Environment: %s", env)
	}

	config, err := NewProjectConfig(configFile)
	if err != nil {
		return nil, err
	}

	triggers := make(map[Event][]*url.URL)

	for topic, triggerNameAndFunctions := range config.GetStringMap("events") {
		for triggerName, functions := range triggerNameAndFunctions.(map[string][]string) {
			event := normalizeEventName(topic, triggerName)
			if _, ok := triggers[event]; !ok {
				triggers[event] = make([]*url.URL, 0)
			}

			for _, functionName := range functions {
				url, err := url.Parse(fmt.Sprintf("%s/%s", config.GetString("functionInvoker"), functionName))
				if err != nil {
					return nil, err
				}

				triggers[event] = append(triggers[event], url)
			}
		}
	}
	logger := log.WithFields(log.Fields{"appName": config.GetString("appName")})
	consumerGroupID := strings.Replace(config.GetString("consumerGroup"), "$appName", config.GetString("appName"), -1)
	app := &Application{config, logger, triggers, nil, consumerGroupID, env}
	app.WorkerPool = NewPool(app.InvokeFunc)

	return app, nil
}

func (app *Application) EnvVar() []string {
	return app.config.GetStringSlice("envVar")
}

func (app *Application) ConnString() string {
	return app.config.GetString(fmt.Sprintf("logStorage.%s.connString", app.Env))
}

func (app *Application) Topics() []string {
	return app.config.GetStringSlice("topics")
}

func (app *Application) Brokers() []string {
	return app.config.GetStringSlice(fmt.Sprintf("logStorage.%s.brokers", app.Env))
}

func (app *Application) LogStorageDriver() string {
	return app.config.GetString(fmt.Sprintf("logStorage.%s.driver", app.Env))
}

func (app *Application) Name() string {
	return app.config.GetString("appName")
}

func (app *Application) FunctionInvoker() string {
	return app.config.GetString("functionInvoker")
}

func (app *Application) WriteProxyListen() string {
	return app.config.GetString("writeProxyListen")
}

func (app *Application) Invoke(msg Message) {
	go func() {
		app.WorkerPool.RunTask(msg)
	}()
}

func (app *Application) invoke(url *url.URL, data []byte) error {
	log := app.Log.WithField("fn", url.Path)
	log.Infof("Invoking")
	client := &http.Client{}
	req, err := http.NewRequest(
		"POST",
		url.String(),
		bytes.NewReader(data),
	)

	req.Header.Set("User-Agent", "SpacerEventRouter")

	resp, err := client.Do(req)
	if err != nil {
		return errors.Wrap(err, "post event handler failed")
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("Function not ok: %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "unable to read function return value")
	}

	var ret map[string]json.RawMessage
	err = json.Unmarshal(body, &ret)
	if err != nil {
		return errors.Wrap(err, "Failed to decode JSON")
	}

	if msg, ok := ret["error"]; ok {
		return fmt.Errorf("Function returned error: %s", msg)
	}
	log.Infof("Function Returned %s", string(body))
	return nil
}

// For WorkerPool
func (app *Application) InvokeFunc(msg Message) error {
	event := normalizeEventName(string(*msg.Topic), "APPEND")
	app.Log.Debugf("Looking for triggers on event %s", event)

	if _, ok := app.Triggers[event]; !ok {
		app.Log.Debugf("No triggers found")
	}

	for _, fn := range app.Triggers[event] {
		err := app.invoke(fn, []byte(string(msg.Value)))
		if err != nil {
			app.Log.WithField("function", app.Triggers[event]).Errorf("Invocation Error: %v", err)
			return err
		}
	}

	return nil
}

func (app *Application) Start(readyChan chan int, withWriteProxy bool) error {
	producer, consumer, err := app.createProducerAndConsumer()
	if err != nil {
		app.Log.Fatalf("Failed to create producer or consumer: %s", err)
	}
	defer producer.Close()
	defer consumer.Close()

	go func() {
		for m := range producer.Events() {
			app.Log.Debugf("Delivered message to topic %s at offset %v", *m.Topic, m.Offset)
		}
	}()

	if withWriteProxy {
		// create a proxy to let functions write data back to kafka
		writeProxy, err := NewWriteProxy(app, producer.ProduceChannel())
		if err != nil {
			app.Log.Fatal(err)
		}
		go http.ListenAndServe(app.WriteProxyListen(), writeProxy)
		app.Log.WithField("listen", app.WriteProxyListen()).Infof("Write Proxy Started")
	}

	app.Log.WithField("groupID", app.ConsumerGroupID).Infof("Consumer Started")

	if readyChan != nil {
		readyChan <- 0
	}

	// start the consumer loop
	for {
		msg, err := consumer.Poll(100)
		if err != nil {
			app.Log.Fatalf("Consumer Error %s", err)
		}
		if msg == nil {
			continue
		}
		app.Invoke(*msg)
	}
}

func (app *Application) createProducerAndConsumer() (LogStorageProducer, LogStorageConsumer, error) {
	var producer LogStorageProducer
	var consumer LogStorageConsumer
	var err error

	driver := app.LogStorageDriver()
	app.Log.Infof("Starting Log Storage with driver %s", driver)

	switch driver {
	case "kafka":
		producer, err = NewKafkaProducer(app)
		if err != nil {
			return nil, nil, err
		}
		consumer, err = NewKafkaConsumer(app)
		if err != nil {
			return nil, nil, err
		}
	case "memory":
		producer = NewMemoryProducer()
		consumer = NewMemoryConsumer()
	case "postgres":
		producer, err = NewPGProducer(app)
		if err != nil {
			return nil, nil, err
		}
		consumer, err = NewPGConsumer(app)
		if err != nil {
			return nil, nil, err
		}

	case "":
		return nil, nil, fmt.Errorf("Missing logStorage driver")
	default:
		return nil, nil, fmt.Errorf("Unknown Log Storage driver %s", driver)
	}

	err = producer.CreateTopics(app.Topics())
	if err != nil {
		return nil, nil, err
	}

	return producer, consumer, nil

}

func normalizeEventName(topic string, event string) Event {
	return Event(fmt.Sprintf("%s:%s", topic, strings.ToUpper(event)))
}
