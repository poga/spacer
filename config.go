package main

import "github.com/spf13/viper"

const CONFIG_SERVICE_KEY = "dependency"

func init() {
	viper.AutomaticEnv()
	viper.SetConfigType("toml")
	viper.AddConfigPath(".")
	viper.SetConfigName("spacer")

	viper.SetDefault("listen", ":9064")
	viper.SetDefault("verbose", false)

	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
}

type Dep struct {
	Repo string
}

func GetDeps() []Dep {
	var result []Dep
	for _, m := range viper.Get(CONFIG_SERVICE_KEY).([]map[string]interface{}) {
		result = append(result, Dep{m["repo"].(string)})
	}
	return result
}
