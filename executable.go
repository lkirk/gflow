package main

import (
	"bytes"
	"log"
	"text/template"
)

func templateExecutable(body string) string {
	shell := "/bin/bash"
	preamble := "set -eo pipefail"
	exeTemplate, err := template.New("exe").Parse("#!{{.Shell}}\n{{.Body}}\n") // TODO: if not endswith \n
	if err != nil {
		log.Fatal(err)
	}
	templateResult := bytes.Buffer{}
	err = exeTemplate.Execute(&templateResult, struct {
		Shell    string
		Body     string
		Preamble string
	}{shell, body, preamble})
	if err != nil {
		log.Fatal(err)
	}
	return templateResult.String()
}
