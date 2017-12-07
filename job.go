package main

import (
	// "bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"strconv"
	"sync"
)

var jState JobState

func init() {
	jState = JobState{0, &sync.Mutex{}}
}

type JobState struct {
	ID    int
	mutex *sync.Mutex
}

func (j *JobState) increment() int {
	j.mutex.Lock()
	j.ID += 1
	j.mutex.Unlock()
	return j.ID
}

type Job struct {
	ID          int      `json:"id"`
	Directories []string `json:"directories"`
	Cmd         string   `json:"cmd"`
	workflow    *Workflow
}

func AddJob(wf *Workflow, cmd string) {
	jobId := jState.increment()

	wf.Jobs = append(wf.Jobs, &Job{jobId, []string{}, templateExecutable(cmd), wf})
}

func (j *Job) initJob() {
	j.createJobDirs()
	j.writeCommandScript()
	j.createDirectories()
}

func (j *Job) pathToExec(s ...string) string {
	jobExec := []string{j.workflow.ExecDir, "jobs", strconv.Itoa(j.ID)}
	return path.Join(append(jobExec, s...)...)
}

func (j *Job) pathToLog(s ...string) string {
	jobLog := []string{j.workflow.LogDir, strconv.Itoa(j.ID)}
	return path.Join(append(jobLog, s...)...)
}

func (j *Job) pathToOutLog() string {
	return j.pathToLog("stdout.log")
}

func (j *Job) pathToErrLog() string {
	return j.pathToLog("stderr.log")
}

func (j *Job) createJobDirs() {
	for _, d := range []string{j.pathToLog(), j.pathToExec()} {
		err := os.MkdirAll(d, 0755)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func (j *Job) createDirectories() {
	for _, d := range j.Directories {
		fmt.Println("creating: ", d)
		err := os.MkdirAll(d, 0755)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func (j *Job) writeCommandScript() {
	err := ioutil.WriteFile(j.pathToExec("exe"), []byte(j.Cmd), 0755)
	if err != nil {
		log.Fatal(err)
	}
}

func (j *Job) openLogs() (outLog, errLog *os.File, err error) {
	outLog, err = os.Create(j.pathToOutLog())
	if err != nil {
		return nil, nil, err
	}
	errLog, err = os.Create(j.pathToErrLog())
	if err != nil {
		return outLog, nil, err
	}
	return outLog, errLog, nil
}

func (j *Job) runJob(wg *sync.WaitGroup) {
	defer wg.Done()
	outLog, errLog, err := j.openLogs()
	if err != nil {
		log.Fatal(err)
	}
	defer outLog.Close()
	defer errLog.Close()

	cmd := exec.Command(j.pathToExec("exe"))

	cmd.Stdout = outLog
	cmd.Stderr = errLog
	cmd.Dir = j.workflow.WorkflowDir

	err = cmd.Run()
	if err != nil {
		log.Println("Job Failed: job_id:", j.ID, err)
		j.workflow.JobsFailed += 1 // TODO: thread safe?
	}

	log.Println("Job Succeeded: job_id:", j.ID)
}
