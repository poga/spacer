package spacer

import (
	"bytes"
	"html/template"
	"io/ioutil"
)

type NginxConfig struct {
	EnvVar              []string
	NoCodeCache         bool
	WriteProxyPort      string
	FunctionInvokerPort string
}

func (c NginxConfig) Generate(sourcePath string) (string, error) {
	tmplData, err := ioutil.ReadFile(sourcePath)
	if err != nil {
		return "", err
	}

	tmpl, err := template.New("tmp").Parse(string(tmplData))
	if err != nil {
		return "", err
	}

	var out bytes.Buffer

	err = tmpl.Execute(&out, c)
	if err != nil {
		return "", err
	}

	return out.String(), nil
}
