package main

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

type Application struct {
	*viper.Viper
	Log    *log.Entry
	Routes map[Event]FuncPath
}

type Event string
type FuncPath string

func NewApplication() (*Application, error) {
	config := viper.New()
	config.AddConfigPath(".")
	config.SetConfigName("spacer")

	// default configs
	config.SetDefault("consumer_group_prefix", "spacer")
	config.SetDefault("delegator", "http://localhost:9064")
	config.SetDefault("write_proxy_listen", ":9065")

	err := config.ReadInConfig()
	if err != nil {
		return nil, err
	}

	router := make(map[Event]FuncPath)
	// load routes
	routeConfig := config.Sub("routes")
	for _, objectAndEvent := range routeConfig.AllKeys() {
		parts := strings.Split(objectAndEvent, ".")
		object := parts[0]
		eventType := parts[1]
		fmt.Println(objectAndEvent, parts)
		for _, funcName := range routeConfig.GetStringSlice(objectAndEvent) {
			router[GetRouteEvent(object, eventType)] = FuncPath(strings.Join([]string{
				config.GetString("delegator"),
				funcName,
			}, "/"))
		}
	}

	logger := log.WithFields(log.Fields{"app_name": config.GetString("app_name")})
	app := &Application{config, logger, router}

	err = ValidateApp(app)
	if err != nil {
		return nil, errors.Wrap(err, "application validation failed")
	}

	return app, nil
}

func ValidateApp(app *Application) error {
	if strings.Contains(app.Name(), "_") {
		return fmt.Errorf("app.name %s cannot contains \"_\"", app.GetString("app_name"))
	}

	return nil
}

func GetRouteEvent(object string, eventType string) Event {
	return Event(fmt.Sprintf("%s:%s", object, strings.ToUpper(eventType)))
}

func GetAbsoluteFuncPath(delegator string, funcName string) FuncPath {
	return FuncPath(fmt.Sprintf("%s/%s", delegator, funcName))
}

func (app *Application) ConsumerGroupID() string {
	return strings.Join([]string{app.GetString("consumer_group_prefix"), app.Name()}, "-")
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
	return app.GetString("app_name")
}

func (app *Application) Invoke(fp FuncPath, data []byte) error {
	app.Log.Infof("Invoking %s\n", string(fp))
	resp, err := http.Post(string(fp), "application/json", bytes.NewReader(data))

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
	return nil

}
