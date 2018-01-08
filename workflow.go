package main

import (
	"encoding/json"
	// "fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"sync"
)

const (
	ExitJobsFailed int = 1 + iota
)

type Workflow struct {
	WorkflowDir string `json:"workflow_dir"`
	LogDir      string `json:"log_dir"`
	ExecDir     string `json:"exec_dir"`
	TmpDir      string `json:"tmp_dir"`
	Jobs        []*Job `json:"jobs"`
	JobsFailed  int    `json:"jobs_failed"`
}

func (w *Workflow) InitFlags() {
	// TODO: add flag parsing here
}

func (w *Workflow) toJson() []byte {
	rawJson, err := json.MarshalIndent(w, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	return rawJson
}

func (w *Workflow) initWorkflow() {
	w.createWorkflowDirs()
}

func (wf *Workflow) AddJob(j *Job) {
	wf.Jobs = append(wf.Jobs, j)
}

func (w *Workflow) createWorkflowDirs() {
	// TODO: where do responsibilities stop?
	// _, err := os.Stat(w.WorkflowDir)
	// if err != nil {
	// 	if os.IsNotExist(err) {
	// 		err = os.Mkdir(w.WorkflowDir, 0775)
	// 		if err != nil {
	// 			log.Fatal(err)
	// 		}
	// 		return
	// 	}
	// 	log.Fatal(err)
	// }

	for _, d := range []string{w.WorkflowDir, w.ExecDir, w.LogDir} {
		err := os.MkdirAll(d, 0755)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func newWorkflow(wfDir string) *Workflow {
	absWfDir, err := filepath.Abs(wfDir)
	if err != nil {
		log.Fatal(err)
	}
	logDir := path.Join(absWfDir, ".gflow", "log")
	execDir := path.Join(absWfDir, ".gflow", "exec")
	tmpDir := path.Join(absWfDir, ".gflow", "tmp")

	wf := &Workflow{absWfDir, logDir, execDir, tmpDir, []*Job{}, 0}
	wf.createWorkflowDirs()
	return wf
}

func (w *Workflow) inferExitStatus() int {
	// TODO: channel listening for errors from running jobs.
	if w.JobsFailed > 0 {
		log.Println("Error:", strconv.Itoa(w.JobsFailed), "jobs failed")
		return ExitJobsFailed
	}
	return 0
}

func (w *Workflow) writeWorkflowJson() {
	f, err := os.Create(path.Join(w.ExecDir, "wf.json"))
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if enc.Encode(w); err != nil {
		log.Fatal(err)
	}
}

func (w *Workflow) Run() int {
	wg := &sync.WaitGroup{}

	for _, j := range w.Jobs {
		err := j.initJob()
		if err != nil {
			log.Fatal("Failed initializing job_id:%d", j.ID)
		}
		wg.Add(1)
		go j.runJob(wg)
	}

	wg.Wait()
	w.writeWorkflowJson()
	exitStatus := w.inferExitStatus()
	if exitStatus != 0 {
		log.Printf("Workflow failed: exit status: %d", exitStatus)
		return exitStatus
	}
	log.Printf("Workflow success")
	return exitStatus
}
