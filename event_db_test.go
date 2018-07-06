package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/boltdb/bolt"
)

var someArgHash = []byte("some arg hash")

func mustTestDb(t *testing.T, dbPath string) (db *bolt.DB) {
	err := os.MkdirAll(dbPath, 0755)
	if err != nil {
		t.Error(err)
	}

	db, err = setupEventDB(filepath.Join(dbPath, "test.db"))
	if err != nil {
		t.Error(err)
	}
	return
}

func jobMustExist(t *testing.T, db *bolt.DB, argHash []byte) {
	je, err := jobExists(db, argHash)
	if err != nil {
		t.Error(err)
	}

	if je != true {
		t.Errorf("expected the job to exist, arghash: %x", argHash)
	}
}

func TestAddJob(t *testing.T) {
	name := "AddJob"
	db := mustTestDb(t, filepath.Join(OutputDir, name))
	defer db.Close()

	err := addJob(db, []byte("ohai"))
	if err != nil {
		t.Error(err)
	}
}

func TestJobNoExist(t *testing.T) {
	name := "JobNoExist"
	db := mustTestDb(t, filepath.Join(OutputDir, name))
	defer db.Close()

	je, err := jobExists(db, someArgHash)
	if err != nil {
		t.Error(err)
	}

	if je == true {
		t.Errorf("expected the job not to exist")
	}
}

func TestJobExists(t *testing.T) {
	name := "JobExists"
	db := mustTestDb(t, filepath.Join(OutputDir, name))
	defer db.Close()

	err := addJob(db, someArgHash)
	if err != nil {
		t.Error(err)
	}

	jobMustExist(t, db, someArgHash)
}
