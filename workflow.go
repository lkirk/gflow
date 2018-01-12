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
	ExitJobsFailed int = 1 + iota // One or more jobs failed
)

type Workflow struct {
	WorkflowDir string `json:"workflow_dir"`
	LogDir      string `json:"log_dir"`
	ExecDir     string `json:"exec_dir"`
	TmpDir      string `json:"tmp_dir"`
	WFJsonPath  string `json:"wf_json_path"`
	Jobs        []*Job `json:"jobs"`

	currentJobId int
	jobIdLock    *sync.Mutex
	failedJobs   *failedJobs
}

func (w *Workflow) InitFlags() {
	// TODO: add flag parsing here
}

func (w *Workflow) initWorkflow() {
	w.createWorkflowDirs()
}

func (wf *Workflow) AddJob(j ...*Job) {
	wf.Jobs = append(wf.Jobs, j...)
}

func (wf *Workflow) pathToWFDir(s ...string) string {
	return path.Join(append([]string{wf.WorkflowDir}, s...)...)
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

func (w *Workflow) incrementCurrentJobId() int {
	w.jobIdLock.Lock()
	w.currentJobId += 1
	w.jobIdLock.Unlock()
	return w.currentJobId
}

func newWorkflow(wfDir string) *Workflow {
	absWfDir, err := filepath.Abs(wfDir)
	if err != nil {
		log.Fatal(err)
	}
	logDir := path.Join(absWfDir, ".gflow", "log")
	execDir := path.Join(absWfDir, ".gflow", "exec")
	tmpDir := path.Join(absWfDir, ".gflow", "tmp")
	wfJsonPath := path.Join(absWfDir, ".gflow", "wf.json")

	wf := &Workflow{
		absWfDir, logDir, execDir, tmpDir, wfJsonPath,
		[]*Job{}, 0, &sync.Mutex{}, newFailedJobs(),
	}
	wf.createWorkflowDirs()
	return wf
}

func (w *Workflow) inferExitStatus() int {
	numberFailedJobs := len(w.failedJobs.jobs)
	if numberFailedJobs > 0 {
		log.Println("Error:", strconv.Itoa(numberFailedJobs), "jobs failed")
		return ExitJobsFailed
	}
	return 0
}

func (w *Workflow) writeWorkflowJson() {
	f, err := os.Create(w.WFJsonPath)
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
	w.initWorkflow()
	wg := &sync.WaitGroup{}

	for _, j := range w.Jobs {
		err := j.initJob()
		if err != nil {
			log.Fatal("Failed initializing job_id: %d", j.ID)
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
