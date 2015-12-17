package main

import (
	"path/filepath"
	"regexp"

	"github.com/spf13/viper"
)

const CONFIG_DEPENDENCY_KEY = "dependency"

type Spacer struct {
	*viper.Viper
}

func NewSpacer(configPath string) (*Spacer, error) {
	v := viper.New()
	v.AutomaticEnv()

	v.SetDefault("listen", ":9064")
	v.SetDefault("prefix", "_services")
	v.SetDefault("verbose", false)

	v.AddConfigPath(configPath)
	v.SetConfigType("toml")
	v.SetConfigName("spacer")

	err := v.ReadInConfig()
	if err != nil {
		return nil, err
	}

	return &Spacer{v}, nil
}

func (s Spacer) Dependencies() []Service {
	var deps []Service

	for _, serviceConfig := range s.Get(CONFIG_DEPENDENCY_KEY).([]map[string]interface{}) {
		if v, ok := serviceConfig["local"]; ok {
			localPath := v.(string)
			deps = append(deps, Service{
				LocalPath: localPath,
				Name:      filepath.Base(localPath),
				Path:      s.GetString("prefix") + "/",
			})
			continue
		}
		if v, ok := serviceConfig["git"]; ok {
			remotePath := v.(string)
			name := regexp.MustCompile(":(.+)/(.+).git").FindStringSubmatch(remotePath)[2]
			deps = append(deps, Service{
				RemotePath: remotePath,
				Name:       name,
				Path:       s.GetString("prefix") + "/",
			})
			continue
		}
	}
	return deps
}
