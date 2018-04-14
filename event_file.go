package main

import (
	"crypto/sha256"
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"time"
)

const BackGroundPollMs = 100

var EventFileHeader = []string{"job_id", "start_time_ns", "end_time_ns", "arg_hash"}

type EventFileEntry struct {
	JobId     int
	StartTime int64
	EndTime   int64
	ArgHash   [sha256.Size]byte
}

func newEventFileWriter(iw io.Writer) *csv.Writer {
	w := csv.NewWriter(iw)
	w.Comma = '\t'
	w.Write(EventFileHeader)
	return w
}

func BackgroundWriteProcess(iw io.Writer) *chanWriter {
	writer := newChanWriter(iw)
	go func() {
		for {
			for efl := range writer.Chan() {
				writer.Write(efl)
			}
			time.Sleep(time.Millisecond * BackGroundPollMs)
		}
	}()
	return writer
}

type chanWriter struct {
	ch     chan EventFileEntry
	writer *csv.Writer
}

func newChanWriter(iw io.Writer) *chanWriter {
	return &chanWriter{
		ch:     make(chan EventFileEntry),
		writer: newEventFileWriter(iw),
	}
}

func (w *chanWriter) Chan() <-chan EventFileEntry {
	return w.ch
}

func (w *chanWriter) Close() error {
	close(w.ch)
	return nil
}

func (cw *chanWriter) Write(efe EventFileEntry) error {
	cw.writer.Write([]string{
		strconv.Itoa(efe.JobId),
		strconv.FormatInt(efe.StartTime, 10),
		strconv.FormatInt(efe.EndTime, 10),
		fmt.Sprintf("%x", efe.ArgHash),
	})
	cw.writer.Flush()
	if err := cw.writer.Error(); err != nil {
		return err
	}
	return nil
}
