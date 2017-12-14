package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"path"
	"path/filepath"
	"time"
)

func randomMilliseconds(min, max int) float64 {
	randInt := rand.Intn(max-min) + min
	return (time.Millisecond * time.Duration(randInt)).Seconds()
}

func main() {
	var (
		workflowDir string
	)

	flag.StringVar(&workflowDir, "wfdir", "", "directory for job to exist")
	flag.Parse()

	if workflowDir == "" {
		fmt.Println("wfdir not specified")
		flag.Usage()
		os.Exit(2)
	}

	wf := newWorkflow(workflowDir)

	// cmds := []string{
	// 	`sleep 3`,
	// 	`sleep 6`,
	// 	`sleep 8`,
	// 	`echo wef | sed -re's/(w)(e)f/\2\1/'`,
	// 	`echo bef`,
	// 	`echo lef`,
	// 	`ls -la`,
	// 	`pwd`,
	// 	`exit 1`,
	// }

	// for _, c := range cmds {
	// 	wf.AddJob(newJob(wf, []string{}, []Dependency{}, c))
	// }

	samples := []string{"a", "b", "c", "d"}
	deps := []string{"e", "f", "g", "h"}
	replace := []string{"hello", "this", "is", "the"}

	tmpDir := "test/job/tmp"
	for i, s := range samples {

		sampleTmp := path.Join(tmpDir, s)

		tmpFile, _ := filepath.Abs(path.Join(sampleTmp, s))
		depTmp, _ := filepath.Abs(path.Join(sampleTmp, deps[i]))

		sampleCmd := fmt.Sprintf(
			`echo "hello this is the world speaking" > %s && sleep %g`,
			tmpFile, randomMilliseconds(1, 5000))

		depCmd := fmt.Sprintf(`sed -e's|%s||g' %s > %s`,
			replace[i], tmpFile, depTmp)

		depJob := newJob(
			wf,
			[]string{sampleTmp},
			[]*Job{},
			[]string{depTmp},
			false,
			depCmd,
		)

		j := newJob(
			wf,
			[]string{sampleTmp},
			[]*Job{depJob},
			[]string{tmpFile},
			false,
			sampleCmd,
		)
		wf.AddJob(j)
	}

	os.Exit(wf.Run())
}
