package main

import (
	"flag"
	"os"
	"path"
	"testing"
)

const OutputDir = "testoutput"

var noClean bool

func init() {
	flag.BoolVar(&noClean, "no-clean", false, "do not clean workflow directories")
	flag.Parse()
}

func cleanTestData(t *testing.T) {
	if noClean == false {
		err := os.RemoveAll(OutputDir)
		if err != nil {
			t.Errorf("Error while removing output dir %s", OutputDir)
		}
	}
}

func expectZero(t *testing.T, status int) {
	if status != 0 {
		t.Error("expected exit 0, wf exited", status)
	}
}

func expectNonZero(t *testing.T, status int) {
	if status == 0 {
		t.Error("expected nonzero exit")
	}
}

func TestWorkflowFailure(t *testing.T) {
	defer cleanTestData(t)
	wf := newWorkflow(path.Join(OutputDir, "testJobFailure"))
	wf.AddJob(
		newJob(wf, []string{}, []*Job{}, []string{}, false, "echo hello"),
		newJob(wf, []string{}, []*Job{}, []string{}, false, "echo error failure >2; false"),
	)
	expectNonZero(t, wf.Run())
}

func TestWorkflowSuccess(t *testing.T) {
	defer cleanTestData(t)
	wf := newWorkflow(path.Join(OutputDir, "testJobFailure"))
	wf.AddJob(
		newJob(wf, []string{}, []*Job{}, []string{}, false, "echo hello"),
	)
	expectZero(t, wf.Run())
}

func TestRunWorkflow(t *testing.T) {
	testCases := []struct {
		name string
		jobs func(*Workflow) []*Job
	}{
		{"WorkflowFailure", func(wf *Workflow) []*Job {
			return []*Job{
				newJob(wf, []string{}, []*Job{}, []string{}, false, "echo hello"),
				newJob(wf, []string{}, []*Job{}, []string{}, false, "echo error failure >2; false"),
			}
		}},
		{"WorkflowSuccess", func(wf *Workflow) []*Job {
			return []*Job{
				newJob(wf, []string{}, []*Job{}, []string{}, false, "echo hello"),
			}
		}},
		{"CreateOutput", func(wf *Workflow) []*Job {
			return []*Job{
				newJob(wf, []string{"a", "b"}, []*Job{}, []string{}, false, "echo hello"),
			}
		}},
		{"CleanTmp", func(wf *Workflow) []*Job {
			return []*Job{
				newJob(wf, []string{}, []*Job{}, []string{}, true, "echo hello"),
			}
		}},
		{"DependentJobs", func(wf *Workflow) []*Job {
			j := newJob(wf, []string{"out"},
				[]*Job{newJob(wf, []string{"out"}, []*Job{}, []string{"out/test_out.txt"}, true, "grep -o some out/test_a.txt > out/test_out.txt")},
				[]string{}, true, "echo 'some test output' > out/test_a.txt")
			return []*Job{j}
		}},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer cleanTestData(t)
			wf := newWorkflow(path.Join(OutputDir, tc.name))
			jobs := tc.jobs(wf)
			wf.AddJob(jobs...)
			wf.Run()
		})
	}
}
