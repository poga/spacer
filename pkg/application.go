package spacer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

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
	return app.GetStringSlice("brokers")
}

func (app *Application) Subscription() string {
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
	app.Log.Infof("Invoking %s", string(fn))
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
	app.Log.Debugf("Invoke %s complete", string(fn))
	return nil
}

// For WorkerPool
func (app *Application) InvokeFunc(msg Message) error {
	parts := strings.Split(*msg.Topic, "_")
	objectType := parts[1]
	routePath := GetRouteEvent(string(objectType), "UPDATE")
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
	// create a producer
	producer, err := NewKafkaProducer(app)
	if err != nil {
		app.Log.Fatal("Unable to create producer:", err)
	}

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

	consumer, err := NewKafkaConsumer(app)
	if err != nil {
		app.Log.Fatal("Failed to create consumer", err)
	}

	defer consumer.Close()

	err = consumer.SubscribeTopics([]string{app.Subscription()})
	if err != nil {
		app.Log.Fatal("Failed to subscribe topics %v, %s", app.Subscription(), err)
	}
	app.Log.WithField("groupID", app.ConsumerGroupID()).Infof("Consumer Started")

	// close consumers when user press ctrl+c
	run := true
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt)
	go func() {
		for _ = range c {
			fmt.Println("Exiting...")
			run = false
		}
	}()

	// periodically refresh metadata to know if there's any new topic created
	go refreshMetadata(consumer, app.Log)

	// start the consumer loop
	for run {
		msg, err := consumer.Poll(100)
		if err != nil {
			app.Log.Fatalf("Consumer Error %s", msg)
		}
		if msg == nil {
			continue
		}
		app.Invoke(*msg)
	}

	return nil
}

func refreshMetadata(consumer LogStorageConsumer, logger *log.Entry) {
	for {
		time.Sleep(5 * time.Second)
		err := consumer.GetMetadata()
		if err != nil {
			// somethimes it just timed out, ignore
			logger.Debugf("Unable to refresh metadata: %s", err)
			continue
		}
	}
}
