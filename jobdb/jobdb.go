package jobdb

// ripped from https://github.com/benschw/jsondb-go, with some modification
// Need to integrate

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
)

func newEventFile(p string) (f *os.File, err error) {
	return os.OpenFile(p, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
}

type Info struct {
	Id    string `json:"id"`
	Value string `json:"value"`
}

// Job to delete a Info from the database
type DeleteInfoJob struct {
	toDelete string
	exitChan chan error
}

func NewDeleteInfoJob(id string) *DeleteInfoJob {
	return &DeleteInfoJob{
		toDelete: id,
		exitChan: make(chan error, 1),
	}
}

func (j DeleteInfoJob) ExitChan() chan error {
	return j.exitChan
}

func (j DeleteInfoJob) Run(infos map[string]Info) (map[string]Info, error) {
	delete(infos, j.toDelete)
	return infos, nil
}

type Job interface {
	ExitChan() chan error
	Run(infos map[string]Info) (map[string]Info, error)
}

func ProcessJobs(jobs chan Job, f *os.File) {
	for {
		j := <-jobs

		infos := make(map[string]Info, 0)
		log.Print("HERE")
		content, err := ioutil.ReadAll(f)
		if err == nil {
			if err = json.Unmarshal(content, &infos); err == nil {
				log.Print("HERE")
				infosMod, err := j.Run(infos)

				if err == nil && infosMod != nil {
					b, err := json.Marshal(infosMod)
					if err == nil {
						_, err = io.WriteString(f, string(b))
					}
				}
			}
		}

		j.ExitChan() <- err
	}
}

// Job to read all infos from the database
type ReadInfosJob struct {
	infos    chan map[string]Info
	exitChan chan error
}

func NewReadInfosJob() *ReadInfosJob {
	return &ReadInfosJob{
		infos:    make(chan map[string]Info, 1),
		exitChan: make(chan error, 1),
	}
}

func (j ReadInfosJob) ExitChan() chan error {
	return j.exitChan
}

func (j ReadInfosJob) Run(infos map[string]Info) (map[string]Info, error) {
	j.infos <- infos

	return nil, nil
}

// Job to add a Info to the database
type SaveInfoJob struct {
	toSave   Info
	saved    chan Info
	exitChan chan error
}

func NewSaveInfoJob(info Info) *SaveInfoJob {
	return &SaveInfoJob{
		toSave:   info,
		saved:    make(chan Info, 1),
		exitChan: make(chan error, 1),
	}
}
func (j SaveInfoJob) ExitChan() chan error {
	return j.exitChan
}
func (j SaveInfoJob) Run(infos map[string]Info) (map[string]Info, error) {
	var info Info
	if j.toSave.Id == "" {
		id, err := newUUID()
		if err != nil {
			return nil, err
		}
		info = Info{Id: id, Value: j.toSave.Value}
	} else {
		info = j.toSave
	}
	infos[info.Id] = info

	j.saved <- info
	return infos, nil
}

// Generate a uuid to use as a unique identifier for each Info
// http://play.golang.org/p/4FkNSiUDMg
func newUUID() (string, error) {
	uuid := make([]byte, 16)
	n, err := io.ReadFull(rand.Reader, uuid)
	if n != len(uuid) || err != nil {
		return "", err
	}
	// variant bits; see section 4.1.1
	uuid[8] = uuid[8]&^0xc0 | 0x80
	// version 4 (pseudo-random); see section 4.1.3
	uuid[6] = uuid[6]&^0xf0 | 0x40
	return fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:]), nil
}

type InfoClient struct {
	Jobs chan Job
}

func (c *InfoClient) SaveInfo(info Info) (Info, error) {
	job := NewSaveInfoJob(info)
	c.Jobs <- job

	if err := <-job.ExitChan(); err != nil {
		return Info{}, err
	}
	return <-job.saved, nil
}

func (c *InfoClient) GetInfos() ([]Info, error) {
	arr := make([]Info, 0)

	infos, err := c.getInfoHash()
	if err != nil {
		return arr, err
	}

	for _, value := range infos {
		arr = append(arr, value)
	}
	return arr, nil
}

func (c *InfoClient) GetInfo(id string) (Info, error) {
	infos, err := c.getInfoHash()
	if err != nil {
		return Info{}, err
	}
	return infos[id], nil
}

func (c *InfoClient) DeleteInfo(id string) error {
	job := NewDeleteInfoJob(id)
	c.Jobs <- job

	if err := <-job.ExitChan(); err != nil {
		return err
	}
	return nil
}

func (c *InfoClient) getInfoHash() (map[string]Info, error) {
	job := NewReadInfosJob()
	c.Jobs <- job

	if err := <-job.ExitChan(); err != nil {
		return make(map[string]Info, 0), err
	}
	return <-job.infos, nil
}
