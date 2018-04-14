package main

import (
	"bytes"
	"crypto/sha256"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestEventFile(t *testing.T) {
	b := bytes.Buffer{}
	bwp := BackgroundWriteProcess(&b)
	defer bwp.Close()

	command := "ls -la"
	sum := sha256.Sum256([]byte(command))
	startTime := time.Now().UnixNano()
	endtime := time.Now().UnixNano()
	jobEntry := strings.Join([]string{
		"1", strconv.FormatInt(startTime, 10), strconv.FormatInt(endtime, 10),
		"1de700c29687cae34561545f50d3c8b3d9afe88e04cc11069f8a6dc6e4ce9464"}, "\t")

	expectedFileContents := bytes.Buffer{}
	expectedFileContents.WriteString(strings.Join(EventFileHeader, "\t"))
	expectedFileContents.WriteRune('\n')
	expectedFileContents.WriteString(jobEntry)
	expectedFileContents.WriteRune('\n')

	efe := EventFileEntry{JobId: 1, StartTime: startTime, EndTime: endtime, ArgHash: sum}
	bwp.Write(efe)
	if b.String() != expectedFileContents.String() {
		t.Errorf("Unexpected test file contents:\nExpected:\n>>>%s<<<\nGot:\n>>>%s<<<\n",
			expectedFileContents.String(), b.String())
	}
}
