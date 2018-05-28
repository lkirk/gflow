package jobdb

// import (
// 	"testing"
// )

// func TestJobDB(t *testing.T) {
// 	jobs := make(chan Job)

// 	eventFile, err := newEventFile("./events.json")
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	defer eventFile.Close()

// 	go ProcessJobs(jobs, eventFile)

// 	client := &InfoClient{Jobs: jobs}

// 	job, err := client.SaveInfo(Info{Value: "SOMETHING"})
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	t.Log(job)
// }
