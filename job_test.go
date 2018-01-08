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

	deps := []string{"e", "f", "g", "h"}
	replace := []string{"hello", "this", "is", "the"}

	tmpDir := "test/job/tmp"
	for i, s := range samples {

		sampleTmp := path.Join(tmpDir, s)

		tmpFile, _ := filepath.Abs(path.Join(sampleTmp, s))
		depTmp, _ := filepath.Abs(path.Join(sampleTmp, deps[i]))

		sampleCmd := fmt.Sprintf(
			`echo "hello this is the world speaking" > %s && sleep %g`,
			tmpFile, randomMilliseconds(1, 5000))

		depCmd := fmt.Sprintf(`sed -e's| %s||g' %s > %s`,
			replace[i], tmpFile, depTmp)

		depJob := newJob(
			wf,
			[]string{sampleTmp},
			[]*Job{},
			[]string{depTmp},
			false,
			depCmd,
		)

		j := newJob(
			wf,
			[]string{sampleTmp},
			[]*Job{depJob},
			[]string{tmpFile},
			false,
			sampleCmd,
		)
		wf.AddJob(j)
	}
}
