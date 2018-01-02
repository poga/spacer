package spacer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/viper"

	log "github.com/sirupsen/logrus"
)

const CONFIG_VERSION = 1

type Application struct {
	*viper.Viper
	Log        *log.Entry
	Routes     map[Event][]FuncName
	WorkerPool *Pool
}

type Event string
type FuncName string

func NewApplication(configPath string, configName string) (*Application, error) {
	config := viper.New()
	config.AddConfigPath(configPath)
	config.SetConfigName(configName)

	// default configs
	config.SetDefault("consumerGroupPrefix", "spacer")
	config.SetDefault("delegator", "http://localhost:9064")
	config.SetDefault("writeProxyListen", ":9065")

	err := config.ReadInConfig()
	if err != nil {
		return nil, err
	}

	router := make(map[Event][]FuncName)
	// load routes
	routeConfig := config.Sub("events")
	for _, objectAndEvent := range routeConfig.AllKeys() {
		parts := strings.Split(objectAndEvent, ".")
		object := parts[0]
		eventType := parts[1]
		routerKey := GetRouteEvent(object, eventType)
		router[routerKey] = make([]FuncName, 0)
		for _, funcName := range routeConfig.GetStringSlice(objectAndEvent) {
			router[routerKey] = append(router[routerKey], FuncName(funcName))
		}
	}

	logger := log.WithFields(log.Fields{"appName": config.GetString("appName")})
	app := &Application{config, logger, router, nil}
	app.WorkerPool = NewPool(app.InvokeFunc)

	err = validateApp(app)
	if err != nil {
		return nil, errors.Wrap(err, "application validation failed")
	}

	return app, nil
}

func validateApp(app *Application) error {
	if app.GetInt("spacer") != CONFIG_VERSION {
		return fmt.Errorf("Expect config version %d, got %d", CONFIG_VERSION, app.GetInt("spacer"))
	}
	if strings.Contains(app.Name(), "_") {
		return fmt.Errorf("app.name %s cannot contains \"_\"", app.GetString("appName"))
	}

	return nil
}

func GetRouteEvent(object string, eventType string) Event {
	return Event(fmt.Sprintf("%s:%s", object, strings.ToUpper(eventType)))
}

func GetAbsoluteFuncPath(delegator string, funcName string) FuncName {
	return FuncName(fmt.Sprintf("%s/%s", delegator, funcName))
}

func (app *Application) ConsumerGroupID() string {
	return strings.Join([]string{app.GetString("consumerGroupPrefix"), app.Name()}, "-")
}

func (app *Application) Brokers() []string {
	return app.GetStringSlice("logStorage.brokers")
}

func (app *Application) LogStorage() string {
	return app.GetString("logStorage.adapter")
}

func (app *Application) KafkaSubscriptionPattern() string {
	return fmt.Sprintf("^%s_*", app.Name())
}

func (app *Application) GetObjectTopic(objectType string) string {
	return fmt.Sprintf("%s_%s", app.Name(), objectType)
}

func (app *Application) Name() string {
	return app.GetString("appName")
}

func (app *Application) Invoke(msg Message) {
	go func() {
		app.WorkerPool.RunTask(msg)
	}()
}

func (app *Application) invoke(fn FuncName, data []byte) error {
	log := app.Log.WithField("fn", string(fn))
	log.Infof("Invoking")
	client := &http.Client{}
	req, err := http.NewRequest(
		"POST",
		strings.Join([]string{app.GetString("delegator"), string(fn)}, "/"),
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
	routePath := GetRouteEvent(string(*msg.Topic), "UPDATE")
	app.Log.Debugf("Looking up route %s", routePath)

	if _, ok := app.Routes[routePath]; !ok {
		app.Log.Debugf("Route not found")
	}

	for _, fn := range app.Routes[routePath] {
		err := app.invoke(fn, []byte(string(msg.Value)))
		if err != nil {
			app.Log.WithField("route", app.Routes[routePath]).Errorf("Invocation Error: %v", err)
			return err
		}
	}

	return nil
}

func (app *Application) Start() error {
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
	go http.ListenAndServe(app.GetString("writeProxyListen"), writeProxy)
	app.Log.WithField("listen", app.GetString("writeProxyListen")).Infof("Write Proxy Started")

	app.Log.WithField("groupID", app.ConsumerGroupID()).Infof("Consumer Started")

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

	app.Log.Infof("Starting Adapter %s", app.LogStorage())

	switch app.LogStorage() {
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
		return nil, nil, fmt.Errorf("Missing logStorage adapter")
	default:
		return nil, nil, fmt.Errorf("Unknown Log Storage Adapter %s", app.LogStorage())
	}

	err = producer.CreateTopics(app.GetStringSlice("topics"))
	if err != nil {
		return nil, nil, err
	}

	return producer, consumer, nil

}
