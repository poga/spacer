package main

import (
	"path/filepath"

	"github.com/spf13/viper"
)

const CONFIG_DEPENDENCY_KEY = "dependency"

func init() {
	viper.AutomaticEnv()
	viper.SetConfigType("toml")
	viper.AddConfigPath(".")
	viper.SetConfigName("spacer")

	viper.SetDefault("listen", ":9064")
	viper.SetDefault("verbose", false)
	viper.SetDefault("prefix", "_services")

	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
}

func getDependencies() []Service {
	var deps []Service

	for _, serviceConfig := range viper.Get(CONFIG_DEPENDENCY_KEY).([]map[string]interface{}) {
		if v, ok := serviceConfig["local"]; ok {
			localPath := v.(string)
			deps = append(deps, Service{LocalPath: localPath,
				Name: filepath.Base(localPath),
				Path: viper.GetString("prefix") + "/",
			})
		}
	}
	return deps
}
