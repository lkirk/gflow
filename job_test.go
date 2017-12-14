package main

import (
	"fmt"
	"os"
	"path"
	"testing"
)

const OutputDir = "testoutput"

func cleanTestData(t *testing.T) {
	err := os.RemoveAll(OutputDir)
	if err != nil {
		t.Errorf("Error while removing output dir %s", OutputDir)
	}
}

func TestRunWorkflow(t *testing.T) {
	testCases := []struct {
		name string
	}{
		{"A"},
		{"B"},
		{"C"},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer cleanTestData(t)
			wf := newWorkflow(path.Join(OutputDir, tc.name))
			fmt.Println(wf.WorkflowDir)
		})
	}
}
