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

func addJob(db *bolt.DB, arghash []byte, returnValue int) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Job"))
		return b.Put(arghash, []byte(strconv.Itoa(returnValue)))
	})
}
