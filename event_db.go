package main

import (
	"fmt"
	"strconv"

	"github.com/boltdb/bolt"
)

func setupEventDB(f string) (db *bolt.DB, err error) {
	db, err = bolt.Open(f, 0600, nil)
	if err != nil {
		err = fmt.Errorf("could not open event file: %s", err)
		return
	}
	err = db.Update(func(tx *bolt.Tx) (err error) {
		_, err = tx.CreateBucketIfNotExists([]byte("Job"))
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

var jobBucket = []byte("Job")

// TODO: investigate whether or not its ok to pass *bolt.Tx through a goroutine.
func addJob(db *bolt.DB, arghash []byte) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(jobBucket)
		return b.Put(arghash, []byte(strconv.Itoa(0)))
	})
}

func deleteJob(db *bolt.DB, arghash []byte) error {
	return nil
}

func jobExists(db *bolt.DB, argHash []byte) (exists bool, err error) {
	err = db.View(func(tx *bolt.Tx) (err error) {
		if err != nil {
			return
		}
		exists = tx.Bucket(jobBucket).Get(argHash) != nil
		return
	})
	return
}
