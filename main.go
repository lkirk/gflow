package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"sync"
	"text/template"
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
	JobDir      string   `json:"job_dir"`
	ExecDir     string   `json:"exec_dir"`
	Cmd         string   `json:"cmd"`
}

func templateExecutable(body string) string {
	shell := "/bin/bash"
	preamble := "set -eo pipefail"
	exeTemplate, err := template.New("exe").Parse("#!{{.Shell}}\n{{.Body}}\n") // TODO: if not endswith \n
	if err != nil {
		log.Fatal(err)
	}
	templateResult := bytes.Buffer{}
	err = exeTemplate.Execute(&templateResult, struct {
		Shell    string
		Body     string
		Preamble string
	}{shell, body, preamble})
	if err != nil {
		log.Fatal(err)
	}
	return templateResult.String()
}

func newJob(jobDir, scriptDir, cmd string) *Job {
	jobId := jState.increment()

	absJobDir, err := filepath.Abs(jobDir)
	if err != nil {
		log.Fatal(err)
	}

	execDir := path.Join(absJobDir, scriptDir, strconv.Itoa(jobId))

	return &Job{jobId, []string{}, jobDir, execDir, templateExecutable(cmd)}
}

func (j *Job) initJob() {
	j.createJobDir()
	j.writeCommandScript()
	j.createDirectories()
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

func (j *Job) createJobDir() {
	// TODO: where do responsibilities stop?
	// _, err := os.Stat(j.JobDir)
	// if err != nil {
	// 	if os.IsNotExist(err) {
	// 		err = os.Mkdir(j.JobDir, 0775)
	// 		if err != nil {
	// 			log.Fatal(err)
	// 		}
	// 		return
	// 	}
	// 	log.Fatal(err)
	// }

	err := os.MkdirAll(j.JobDir, 0755)
	if err != nil {
		log.Fatal(err)
	}
	err = os.MkdirAll(j.ExecDir, 0755)
	if err != nil {
		log.Fatal(err)
	}
}

func (j *Job) writeCommandScript() {
	err := ioutil.WriteFile(path.Join(j.ExecDir, "exe"), []byte(j.Cmd), 0755)
	if err != nil {
		log.Fatal(err)
	}
}

func (j *Job) toJson() []byte {
	rawJson, err := json.MarshalIndent(j, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	return rawJson
}

func (j *Job) runJob(wg *sync.WaitGroup) {
	defer wg.Done()

	cmd := exec.Command(path.Join(j.ExecDir, "exe"))

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Dir = j.JobDir
	err := cmd.Run()
	if err != nil {
		log.Println("Job Failed: job_id:", j.ID, err)
	}
	fmt.Print(out.String())
}

func main() {
	jobDir := "./job"   // TODO: flag for jobDir
	execDir := ".gflow" // TODO: flag for execDir
	jobs := []*Job{
		newJob(jobDir, execDir, `sleep 5`),
		newJob(jobDir, execDir, `sleep 10`),
		newJob(jobDir, execDir, `sleep 15`),
		newJob(jobDir, execDir, `
echo wef | \
sed -re's/(w)(e)f/\2\1/'
`),
		newJob(jobDir, execDir, `echo bef`),
		newJob(jobDir, execDir, `echo lef`),
		newJob(jobDir, execDir, `ls -la`),
		newJob(jobDir, execDir, `pwd`),
		newJob(jobDir, execDir, `exit 1`),
	}

	wg := &sync.WaitGroup{}

	for _, j := range jobs {
		j.initJob()
		fmt.Printf("%s\n", j.toJson())
		wg.Add(1)
		go j.runJob(wg)
	}

	wg.Wait()
}
