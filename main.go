package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	var (
		workflowYaml string
	)
	flag.StringVar(&workflowYaml, "yaml", "", "path to workflow yaml file")
	flag.Parse()

	if workflowYaml == "" {
		fmt.Println("Error: workflow yaml not specified")
		flag.Usage()
		os.Exit(2)
	}

	RunFromYaml(workflowYaml)
}
