package main

import (
	"path/filepath"
	"regexp"

	"github.com/spf13/viper"
)

const CONFIG_DEPENDENCY_KEY = "dependency"

func init() {
	viper.AutomaticEnv()

	viper.SetDefault("listen", ":9064")
	viper.SetDefault("prefix", "_services")
	viper.SetDefault("verbose", false)

	viper.AddConfigPath(".")
	viper.SetConfigType("toml")
	viper.SetConfigName("spacer")

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
			deps = append(deps, Service{
				LocalPath: localPath,
				Name:      filepath.Base(localPath),
				Path:      viper.GetString("prefix") + "/",
			})
			continue
		}
		if v, ok := serviceConfig["git"]; ok {
			remotePath := v.(string)
			name := regexp.MustCompile(":(.+)/(.+).git").FindStringSubmatch(remotePath)[2]
			deps = append(deps, Service{
				RemotePath: remotePath,
				Name:       name,
				Path:       viper.GetString("prefix") + "/",
			})
			continue
		}
	}
	return deps
}
