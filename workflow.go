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
	logDir := path.Join(absWfDir, "log")
	execDir := path.Join(absWfDir, ".gflow")

	wf := &Workflow{absWfDir, logDir, execDir, []*Job{}, 0}
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
	// io.Copy(f, )
}

func (w *Workflow) Run() int {
	wg := &sync.WaitGroup{}

	// fmt.Printf("%s\n", w.toJson())
	for _, j := range w.Jobs {
		j.initJob()
		wg.Add(1)
		go j.runJob(wg)
	}

	w.writeWorkflowJson()
	wg.Wait()
	return w.inferExitStatus()
}
