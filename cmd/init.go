package cmd

import (
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"

	spacer "github.com/poga/spacer/pkg"
	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init [targetDir]",
	Short: "init a new spacer project",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Initialize a spacer project from defined template and setup directory structure for nginx
		targetDir, err := filepath.Abs(args[0])
		if err != nil {
			log.Fatal(err)
		}

		if targetDir == "" {
			log.Fatalf("Target Directory is Required")
		}

		if _, err := os.Stat(targetDir); err == nil {
			log.Fatal("Target Directory already exists")
		}

		err = series([]func() error{
			func() error { return spacer.RestoreAssets(targetDir, "app") },
			func() error { return spacer.RestoreAssets(targetDir, "lib") },
			func() error { return spacer.RestoreAssets(targetDir, "bin") },
			// func() error { return spacer.RestoreAssets(targetDir, "config") },

			func() error { return os.Mkdir(filepath.Join(targetDir, "config"), os.ModePerm) },
			func() error {
				return writeFromTemplate(filepath.Join(targetDir, "config", "nginx.development.conf"), "config/nginx.conf", nginxConfigTmpl{true})
			},
			func() error {
				return writeFromTemplate(filepath.Join(targetDir, "config", "nginx.production.conf"), "config/nginx.conf", nginxConfigTmpl{false})
			},
			func() error {
				return writeFromTemplate(filepath.Join(targetDir, "config", "nginx.test.conf"), "config/nginx.conf", nginxConfigTmpl{false})
			},
			func() error {
				return writeFromTemplate(
					filepath.Join(targetDir, "config", "env.development.yml"),
					"config/env.yml",
					envConfigTmpl{
						"postgres",
						"postgres://localhost/spacer-development?sslmode=disable",
					},
				)
			},
			func() error {
				return writeFromTemplate(
					filepath.Join(targetDir, "config", "env.production.yml"),
					"config/env.yml",
					envConfigTmpl{
						"postgres",
						"postgres://localhost/spacer-production",
					},
				)
			},
			func() error {
				return writeFromTemplate(
					filepath.Join(targetDir, "config", "env.test.yml"),
					"config/env.yml",
					envConfigTmpl{
						"postgres",
						"postgres://localhost/spacer-test?sslmode=disable",
					},
				)
			},
			func() error {
				return writeFile(filepath.Join(targetDir, "config", "application.yml"), "config/application.yml")
			},
			func() error { return writeFile(filepath.Join(targetDir, ".gitignore"), "appignore") },
			func() error { return writeFile(filepath.Join(targetDir, "Dockerfile"), "Dockerfile") },
			func() error { return os.Mkdir(filepath.Join(targetDir, "logs"), os.ModePerm) },
			func() error { return os.Mkdir(filepath.Join(targetDir, "temp"), os.ModePerm) },
			func() error { return os.Mkdir(filepath.Join(targetDir, "test"), os.ModePerm) },
			func() error { return writeFile(filepath.Join(targetDir, "test/hello.t.md"), "hello.t.md") },
		})
		if err != nil {
			log.Fatal(err)
		}
		return
	},
}

func init() {
	RootCmd.AddCommand(initCmd)
}

func writeFile(to string, name string) error {
	data, err := spacer.Asset(name)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(to, data, 0644)
	if err != nil {
		return err
	}
	return nil
}

type nginxConfigTmpl struct {
	NoCache bool
}

type envConfigTmpl struct {
	DefaultDriver     string
	DefaultConnString string
}

func writeFromTemplate(to string, tmplFile string, data interface{}) error {
	tmplData, err := spacer.Asset(tmplFile)
	if err != nil {
		return err
	}
	tmpl, err := template.New("tmp").Parse(string(tmplData))
	if err != nil {
		return err
	}

	out, err := os.Create(to)
	if err != nil {
		return err
	}

	return tmpl.Execute(out, data)
}

func series(funcs []func() error) error {
	for _, f := range funcs {
		err := f()
		if err != nil {
			return err
		}
	}
	return nil
}
