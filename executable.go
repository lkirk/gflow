package main

import (
	"bytes"
	"errors"
	"log"
	"strings"
	"text/template"
)

func templateCleanTmpTrap(tmpDir string) (string, error) {
	if tmpDir == "" {
		return "", errors.New("Error: tried to create cleanup, no tmpdir")
	}
	// uses bash specific logic (pseudosignal EXIT)
	t := `
	function cleanTmp {
	    echo 'GFLOW: removing tmp dir: "{{.TmpDir}}"'
	    if [[ -d {{.TmpDir}} ]]; then
	        rm -rf {{.TmpDir}}
	    fi
	}
	trap cleanTmp EXIT
	`
	t = strings.Replace(t, "\t", "", -1) // remove leading tabs
	cleanTmpTemplate, err := template.New("cleanTmp").Parse(t)
	if err != nil {
		return "", err
	}

	templateResult := bytes.Buffer{}
	err = cleanTmpTemplate.Execute(&templateResult, struct{ TmpDir string }{tmpDir})
	if err != nil {
		return "", err
	}
	return templateResult.String(), err
}

func templateBody(j *Job) (string, error) {
	bodyTemplate, err := template.New("bodyTemplate").Parse(j.Cmd)
	if err != nil {
		return "", err
	}

	templateResult := bytes.Buffer{}
	err = bodyTemplate.Execute(&templateResult, struct {
		Job    *Job
		TmpDir string
	}{j, j.pathToTmp()})
	if err != nil {
		return "", err
	}
	return templateResult.String(), err
}

func templateExecutable(j *Job) string {
	shell := "/bin/bash"
	preamble := "set -eo pipefail"

	traps := ""
	err := errors.New("")
	if j.CleanTmp {
		traps, err = templateCleanTmpTrap(j.pathToTmp())
		if err != nil {
			log.Fatal(err)
		}
	}

	scriptText := "#!{{.Shell}}\n{{.Preamble}}\n{{.Traps}}\n{{.Body}}\n"

	exeTemplate, err := template.New("exe").Parse(scriptText) // TODO: if not endswith \n
	if err != nil {
		log.Fatal(err)
	}
	body, err := templateBody(j)
	if err != nil {
		log.Fatal(err)
	}

	templateResult := bytes.Buffer{}
	err = exeTemplate.Execute(&templateResult, struct {
		Shell    string
		Body     string
		Preamble string
		Traps    string
	}{shell, body, preamble, traps})
	if err != nil {
		log.Fatal(err)
	}
	return templateResult.String()
}
