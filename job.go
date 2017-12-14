package main

import (
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
	workflow *Workflow

	ID           int      `json:"id"`
	Directories  []string `json:"directories"`
	Dependencies []*Job   `json:"dependencies"`
	Produces     []string `json:"produces"`
	CleanTmp     bool     `json:"clean_tmp"`
	Cmd          string   `json:"cmd"`
}

func newJob(wf *Workflow, dirs []string, deps []*Job, produces []string, clean bool, cmd string) *Job {
	jobId := jState.increment()
	return &Job{wf, jobId, dirs, deps, produces, clean, templateExecutable(cmd)}
}

func (j *Job) initJob() error {
	j.createJobDirs()
	j.writeCommandScript()
	err := j.createDirectories()
	return err
}

func (j *Job) AddDependency(deps ...*Job) {
	j.Dependencies = append(j.Dependencies, deps...)
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

func (j *Job) createDirectories() (err error) {
	for _, d := range j.Directories {
		exists, err := Exists(d)
		switch {
		case err != nil:
			log.Printf("Failed to stat dir '%s' job_id:%d error:'%s'", d, j.ID, err.Error())
			return err
		case exists:
			return nil
		default:
			log.Println("creating: ", d)
			err = os.MkdirAll(d, 0755)
			if err != nil {
				return err
			}
		}

	}
	return
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

func (j *Job) checkProduces() bool {
	if len(j.Produces) == 0 {
		return false
	}
	for _, f := range j.Produces {
		exists, err := Exists(f)
		if err != nil {
			log.Printf("Failed to stat file '%s' job_id:%d error:'%s'", f, j.ID, err.Error())
			j.workflow.JobsFailed += 1 // TODO: thread safe?
			return exists
		}
		if !exists {
			return exists
		}
	}
	return true
}

func (j *Job) runJob(wg *sync.WaitGroup) {
	defer wg.Done()
	if j.checkProduces() {
		return
	}

	outLog, errLog, err := j.openLogs()
	if err != nil {
		log.Fatal(err)
	}

	defer outLog.Close()
	defer errLog.Close()

	depWg := &sync.WaitGroup{}
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

	for _, d := range j.Dependencies {
		d.initJob()
		depWg.Add(1)
		go d.runJob(wg)
	}
	depWg.Wait()
}
