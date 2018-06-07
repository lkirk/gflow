package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/boltdb/bolt"
)

func TestEventFile(t *testing.T) {
	b := bytes.Buffer{}
	bwp := backgroundWriteProcess(&b)
	defer bwp.Close()

	command := "ls -la"
	sum := sha256.Sum256([]byte(command))
	startTime := time.Now().UnixNano()
	endtime := time.Now().UnixNano()
	jobEntry := strings.Join([]string{
		"1", strconv.FormatInt(startTime, 10), strconv.FormatInt(endtime, 10),
		"1de700c29687cae34561545f50d3c8b3d9afe88e04cc11069f8a6dc6e4ce9464"}, "\t")

	expectedFileContents := bytes.Buffer{}
	expectedFileContents.WriteString(jobEntry)
	expectedFileContents.WriteRune('\n')

	efe := EventFileEntry{JobID: 1, StartTime: startTime, EndTime: endtime, ArgHash: sum}
	bwp.Write(efe)
	if b.String() != expectedFileContents.String() {
		t.Errorf("Unexpected test file contents:\nExpected:\n>>>%s<<<\nGot:\n>>>%s<<<\n",
			expectedFileContents.String(), b.String())
	}
}

func setupEventFile() (db *bolt.DB, err error) {
	db, err = bolt.Open("test.db", 0600, nil)
	if err != nil {
		err = fmt.Errorf("could not open event file: %s", err)
		return
	}
	err = db.Update(func(tx *bolt.Tx) (err error) {
		root, err := tx.CreateBucketIfNotExists([]byte("DB"))
		if err != nil {
			err = fmt.Errorf("could not create root bucket: %s", err)
			return
		}
		_, err = root.CreateBucketIfNotExists([]byte("Job"))
		if err != nil {
			err = fmt.Errorf("could not create Job bucket: %s", err)
			return
		}
		return
	})
	if err != nil {
		err = fmt.Errorf("could not set up buckets: %s", err)
		return
	}
	return
}

func addJob(db *bolt.DB, id int, arghash []byte) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("DB")).Bucket([]byte("Job"))
		return b.Put([]byte(id), arghash)
	})
}

func TestBoltIntegration(t *testing.T) {
	db, err := setupEventFile()
	if err != nil {
		t.Error(err)
	}
	defer db.Close()
	addJob(db, 1, []byte("ohai"))
}
