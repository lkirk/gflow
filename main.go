package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	var (
		workflowDir string
	)

	flag.StringVar(&workflowDir, "wfdir", "", "directory for job to exist")

	flag.Parse()

	if workflowDir == "" {
		fmt.Println("wfdir not specified")
		flag.Usage()
		os.Exit(2)
	}

	wf := newWorkflow(workflowDir)

	AddJob(wf, `sleep 3`)
	AddJob(wf, `sleep 6`)
	AddJob(wf, `sleep 8`)
	AddJob(wf, `echo wef | sed -re's/(w)(e)f/\2\1/'`)
	AddJob(wf, `echo bef`)
	AddJob(wf, `echo lef`)
	AddJob(wf, `ls -la`)
	AddJob(wf, `pwd`)
	AddJob(wf, `exit 1`)

	os.Exit(wf.Run())
}
