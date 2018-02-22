package spacer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	yaml "gopkg.in/yaml.v2"

	"github.com/pkg/errors"

	log "github.com/sirupsen/logrus"
)

const CONFIG_VERSION = 1

type ApplicationConfig struct {
	SpacerVersion int                            `yaml:"spacerVersion"`
	AppName       string                         `yaml:"appName"`
	Topics        []string                       `yaml:"topics"`
	Events        map[string]map[string][]string `yaml:"events"`

	ConsumerGroupID string
}

type EnvConfig struct {
	LogStorage struct {
		Driver string `yaml:"driver"`

		// pg
		ConnString string `yaml:"connString"`

		// kafka
		Brokers []string `yaml:"brokers"`
	} `yaml:"logStorage"`

	Delegator        string `yaml:"delegator"`
	WriteProxyListen string `yaml:"writeProxyListen"`
}

type Application struct {
	appConfig       *ApplicationConfig
	envConfig       *EnvConfig
	Log             *log.Entry
	Triggers        map[Event][]*url.URL
	WorkerPool      *Pool
	ConsumerGroupID string
}

type Event string

func NewApplicationConfig(file string) (*ApplicationConfig, error) {
	configData, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	config := ApplicationConfig{}
	err = yaml.Unmarshal(configData, &config)
	if err != nil {
		return nil, err
	}

	// defaults
	if config.ConsumerGroupID == "" {
		config.ConsumerGroupID = "spacer-$appName"
	}
	log.Debugf("environment config: %v", config)

	// validation
	if config.SpacerVersion != CONFIG_VERSION {
		return nil, fmt.Errorf("Expect config version %d, got %d", CONFIG_VERSION, config.SpacerVersion)
	}
	if strings.Contains(config.AppName, "_") {
		return nil, fmt.Errorf("appName %s cannot contains \"_\"", config.AppName)
	}

	return &config, nil
}

func NewEnvConfig(file string) (*EnvConfig, error) {
	configData, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	config := EnvConfig{}
	err = yaml.Unmarshal(configData, &config)
	if err != nil {
		return nil, err
	}

	// defaults
	if config.Delegator == "" {
		config.Delegator = "http://localhost:3000"
	}
	if config.WriteProxyListen == "" {
		config.WriteProxyListen = ":9065"
	}
	log.Debugf("application config: %v", config)

	return &config, nil
}

func NewApplication(configPath string, configName string, env string) (*Application, error) {
	appConfig, err := NewApplicationConfig(filepath.Join(configPath, "config", "application.yml"))
	if err != nil {
		return nil, err
	}
	envConfig, err := NewEnvConfig(filepath.Join(configPath, "config", fmt.Sprintf("env.%s.yml", env)))
	if err != nil {
		return nil, err
	}

	triggers := make(map[Event][]*url.URL)

	for topic, handlers := range appConfig.Events {
		for eventName, functions := range handlers {
			event := normalizeEventName(topic, eventName)
			if _, ok := triggers[event]; !ok {
				triggers[event] = make([]*url.URL, 0)
			}

			for _, functionName := range functions {
				url, err := url.Parse(fmt.Sprintf("%s/%s", envConfig.Delegator, functionName))
				if err != nil {
					return nil, err
				}

				triggers[event] = append(triggers[event], url)
			}
		}
	}
	logger := log.WithFields(log.Fields{"appName": appConfig.AppName})
	consumerGroupID := strings.Replace(appConfig.ConsumerGroupID, "$appName", appConfig.AppName, -1)
	app := &Application{appConfig, envConfig, logger, triggers, nil, consumerGroupID}
	app.WorkerPool = NewPool(app.InvokeFunc)

	return app, nil
}

func (app *Application) Brokers() []string {
	return app.envConfig.LogStorage.Brokers
}

func (app *Application) LogStorageDriver() string {
	return app.envConfig.LogStorage.Driver
}

func (app *Application) Name() string {
	return app.appConfig.AppName
}

func (app *Application) Delegator() string {
	return app.envConfig.Delegator
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

func (app *Application) Start(readyChan chan int) error {
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

	// create a proxy to let functions write data back to kafka
	writeProxy, err := NewWriteProxy(app, producer.ProduceChannel())
	if err != nil {
		app.Log.Fatal(err)
	}
	go http.ListenAndServe(app.envConfig.WriteProxyListen, writeProxy)
	app.Log.WithField("listen", app.envConfig.WriteProxyListen).Infof("Write Proxy Started")

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

	return nil
}

func (app *Application) createProducerAndConsumer() (LogStorageProducer, LogStorageConsumer, error) {
	var producer LogStorageProducer
	var consumer LogStorageConsumer
	var err error

	driver := app.envConfig.LogStorage.Driver
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

	err = producer.CreateTopics(app.appConfig.Topics)
	if err != nil {
		return nil, nil, err
	}

	return producer, consumer, nil

}

func normalizeEventName(topic string, event string) Event {
	return Event(fmt.Sprintf("%s:%s", topic, strings.ToUpper(event)))
}
