package main

import (
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"
	// "time"

	"github.com/boltdb/bolt"
)

// The Job type abstracts the execution of an executable.
// A bash script is written to the filesystem for execution
// the process is waited on to return, then dependent jobs are run.
// If a job fails, dependent jobs will not execute
type Job struct {
	workflow *Workflow

	argHash []byte

	ID           int      `json:"id"`
	Directories  []string `json:"directories"`
	Dependencies []*Job   `json:"dependencies"`
	Outputs      []string `json:"outputs"`
	CleanTmp     bool     `json:"clean_tmp"`
	Cmd          string   `json:"cmd"`
}

func newJob(wf *Workflow, dirs []string, deps []*Job, outputs []string, clean bool, cmd string) *Job {
	JobID := wf.incrementCurrentJobID()
	job := &Job{wf, []byte{}, JobID, dirs, deps, outputs, clean, cmd}
	job.Cmd = templateExecutable(job)
	rawHash := sha256.Sum256([]byte(job.Cmd))
	argHash := append(append([]byte(strconv.Itoa(job.ID)), []byte(".")[:]...), rawHash[:]...)
	job.argHash = argHash
	return job
}

func (j *Job) initJob() error {
	j.createJobDirs()
	j.writeCommandScript()
	err := j.createDirectories()
	return err
}

// AddDependency adds a job dependency the current job instance
func (j *Job) AddDependency(deps ...*Job) {
	j.Dependencies = append(j.Dependencies, deps...)
}

func (j *Job) pathToExec(s ...string) string {
	jobExecDir := []string{j.workflow.execDir, strconv.Itoa(j.ID)}
	return filepath.Join(append(jobExecDir, s...)...)
}

func (j *Job) pathToLog(s ...string) string {
	jobLogDir := []string{j.workflow.logDir, strconv.Itoa(j.ID)}
	return filepath.Join(append(jobLogDir, s...)...)
}

func (j *Job) pathToTmp(s ...string) string {
	jobTmpDir := []string{j.workflow.tmpDir, strconv.Itoa(j.ID)}
	return filepath.Join(append(jobTmpDir, s...)...)
}

func (j *Job) pathToOutLog() string {
	return j.pathToLog("stdout.log")
}

func (j *Job) pathToErrLog() string {
	return j.pathToLog("stderr.log")
}

func (j *Job) createJobDirs() {
	for _, d := range []string{j.pathToLog(), j.pathToExec(), j.pathToTmp()} {
		err := os.MkdirAll(d, 0755)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func pathExists(path string) (exists bool, err error) {
	_, err = os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	return err == nil, nil
}

func (j *Job) createDirectories() (err error) {
	for _, d := range j.Directories {
		abs := j.workflow.pathToWfDir(d)
		exists, err := pathExists(abs)
		switch {
		case err != nil:
			log.Printf("Failed to stat dir '%s' job_id:%d error:'%s'", d, j.ID, err.Error())
			return err
		case exists:
			return nil
		default:
			log.Println("creating:", d)
			err = os.MkdirAll(abs, 0755)
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

func (j *Job) checkOutputs() bool {
	if len(j.Outputs) == 0 {
		return true
	}
	for _, f := range j.Outputs {
		f = j.workflow.pathToWfDir(f)
		exists, err := pathExists(f)
		if err != nil {
			log.Printf("Failed to stat file '%s' job_id:%d error:'%s'", f, j.ID, err.Error())
			return false
		}
		return exists
	}
	return true
}

type failedJobs struct {
	jobs  []*Job
	mutex *sync.Mutex
}

func newFailedJobs() *failedJobs {
	fj := &failedJobs{}
	fj.mutex = &sync.Mutex{}
	return fj
}

func (fj *failedJobs) add(job *Job) {
	fj.mutex.Lock()
	fj.jobs = append(fj.jobs, job)
	fj.mutex.Unlock()
}

func (j *Job) runJob(wg *sync.WaitGroup, db *bolt.DB) {
	defer wg.Done()
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

	exists, err := jobExists(db, j.argHash)
	switch {
	case err != nil:
		log.Fatal(fmt.Errorf("failed to check job existence %s", err))
	case !exists:
		err = cmd.Run()
		switch {
		case j.checkOutputs() == false:
			log.Println("Job Failed: required outputs do not exist: job_id:", j.ID, j.Outputs)
			j.workflow.failedJobs.add(j)
		case err == nil:
			log.Println("Job Succeeded: job_id:", j.ID)
			addJob(db, j.argHash)
		default:
			log.Println("Job Failed: job_id:", j.ID, err)
			j.workflow.failedJobs.add(j)
		}
	case exists:
		log.Println("not running because job", j.ID, "has completed successfully")
	}

	for _, d := range j.Dependencies {
		d.initJob()
		depWg.Add(1)
		go d.runJob(wg, db)
	}
	depWg.Wait()
}
