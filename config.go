package main

import (
	"path/filepath"
	"regexp"

	"github.com/spf13/viper"
)

const CONFIG_DEPENDENCY_KEY = "dependency"

var self *viper.Viper

func init() {
	self = viper.New()
	self.AutomaticEnv()

	self.SetDefault("listen", ":9064")
	self.SetDefault("prefix", "_services")
	self.SetDefault("verbose", false)

	self.AddConfigPath(".")
	self.SetConfigType("toml")
	self.SetConfigName("spacer")

	err := self.ReadInConfig()
	if err != nil {
		panic(err)
	}
}

func getDependencies() []Service {
	var deps []Service

	for _, serviceConfig := range self.Get(CONFIG_DEPENDENCY_KEY).([]map[string]interface{}) {
		if v, ok := serviceConfig["local"]; ok {
			localPath := v.(string)
			deps = append(deps, Service{
				LocalPath: localPath,
				Name:      filepath.Base(localPath),
				Path:      self.GetString("prefix") + "/",
			})
			continue
		}
		if v, ok := serviceConfig["git"]; ok {
			remotePath := v.(string)
			name := regexp.MustCompile(":(.+)/(.+).git").FindStringSubmatch(remotePath)[2]
			deps = append(deps, Service{
				RemotePath: remotePath,
				Name:       name,
				Path:       self.GetString("prefix") + "/",
			})
			continue
		}
	}
	return deps
}
