package spacer

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	*viper.Viper
}

const CONFIG_VERSION = 1

const DEFAULT_CONSUMER_GROUP = "spacer-$appName"
const DEFAULT_FUNCTION_INVOKER = "http://localhost:3000"
const DEFAULT_WRITE_PROXY_LISTEN = ":9065"

func NewProjectConfig(path string) (*Config, error) {
	c := Config{viper.New()}

	c.AutomaticEnv()

	dir, file := filepath.Split(path)

	c.SetConfigName(strings.TrimSuffix(file, filepath.Ext(file)))
	c.AddConfigPath(dir)

	c.SetDefault("consumerGroup", DEFAULT_CONSUMER_GROUP)
	c.SetDefault("functionInvoker", DEFAULT_FUNCTION_INVOKER)
	c.SetDefault("writeProxyListen", DEFAULT_WRITE_PROXY_LISTEN)

	err := c.ReadInConfig()
	if err != nil {
		return nil, err
	}

	if c.GetInt("spacerVersion") != CONFIG_VERSION {
		return nil, fmt.Errorf("Expect config version %d, got %d", CONFIG_VERSION, c.GetInt("spacerVersion"))

	}
	if strings.Contains(c.GetString("appName"), "_") {
		return nil, fmt.Errorf("appName %s cannot contains \"_\"", c.GetString("appName"))
	}

	return &c, nil
}
