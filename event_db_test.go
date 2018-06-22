package main

import (
	// "fmt"
	"os"
	"path"
	"testing"

	"github.com/boltdb/bolt"
)

var someArgHash = []byte("some arg hash")

func mustTestDb(t *testing.T, dbPath string) (db *bolt.DB) {
	err := os.MkdirAll(dbPath, 0755)
	if err != nil {
		t.Error(err)
	}

	db, err = setupEventDB(path.Join(dbPath, "test.db"))
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
	db := mustTestDb(t, path.Join(OutputDir, name, "test.db"))
	defer db.Close()

	err := addJob(db, []byte("ohai"))
	if err != nil {
		t.Error(err)
	}
}

func TestJobNoExist(t *testing.T) {
	name := "JobNoExist"
	db := mustTestDb(t, path.Join(OutputDir, name))
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
	db := mustTestDb(t, path.Join(OutputDir, name))
	defer db.Close()

	err := addJob(db, someArgHash)
	if err != nil {
		t.Error(err)
	}

	jobMustExist(t, db, someArgHash)
}

// TODO: Table tests will look something like this.... ish

// import (
// 	"path"
// 	"testing"

// 	"github.com/boltdb/bolt"
// )

// func TestJobOperations(t *testing.T) {
// 	testCases := []struct {
// 		name string
// 		fn   func() error
// 	}{
// 		{"TestAddJob", addJob(db, []byte("ohai"), 1)},
// 	}

// 	for _, tc := range testCases {
// 		t.Run(tc.name, func(t *testing.T) {
// 			defer cleanTestData(t)
// 			db, err := setupEventDB(path.Join(OutputDir, tc.name))
// 			if err != nil {
// 				t.Error(err)
// 			}
// 			defer db.Close()
// 		})
// 	}
// }
