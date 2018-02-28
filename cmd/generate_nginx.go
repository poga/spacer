package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

	spacer "github.com/poga/spacer/pkg"
	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

var genCmd = &cobra.Command{
	Use:   "nginx-config <projectDirectory>",
	Short: "Generate nginx config for specified environment",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		projectDir, err := filepath.Abs(args[0])
		if err != nil {
			log.Fatal(err)
		}

		if configFile == "" {
			configFile = filepath.Join(projectDir, "config", "application.yml")
		}

		if env == "" {
			log.Fatal(fmt.Sprintf("Invalid Environment: %s", env))
		}

		app, err := spacer.NewApplication(configFile, env)
		if err != nil {
			log.Fatal(err)
		}

		nginxConfig := spacer.NginxConfig{}

		if app.Env == "development" {
			nginxConfig.NoCodeCache = true
		}
		nginxConfig.WriteProxyPort = strings.Split(app.WriteProxyListen(), ":")[1]
		nginxConfig.EnvVar = make([]string, 0)
		for _, envVar := range app.EnvVar() {
			nginxConfig.EnvVar = append(nginxConfig.EnvVar, envVar)
		}

		nginxConfig.FunctionInvokerPort = strings.Split(app.FunctionInvoker(), ":")[2]

		config, err := nginxConfig.Generate(filepath.Join(projectDir, "config", "nginx.conf"))
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(config)
	},
}

func init() {
	genCmd.Flags().StringVarP(&env, "env", "e", "", "Environment Name")
	genCmd.Flags().StringVarP(&configFile, "config", "c", "", "config file")
	RootCmd.AddCommand(genCmd)
}
