// Command train trains a single-layer Haar classifier
// to identify the nose+mouth region of a face.
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/unixpickle/haar"
)

func main() {
	if len(os.Args) != 4 {
		fmt.Fprintf(os.Stderr, "Usage: %s positives negatives output.json\n", os.Args[0])
		os.Exit(1)
	}

	log.Println("Loading samples ...")

	posDir, negDir := os.Args[1], os.Args[2]
	samples, err := haar.LoadSampleSource(posDir, negDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to load samples:", err)
		os.Exit(1)
	}

	reqs := []*haar.Requirements{
		&haar.Requirements{
			PositiveRetention: 1,
			NegativeExclusion: 1,
			MaxFeatures:       10,
		},
	}

	cascade := haar.Train(reqs, samples, haar.ConsoleLogger{})
	data, err := json.Marshal(cascade)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to marshal data:", err)
		os.Exit(1)
	}
	if err := ioutil.WriteFile(os.Args[3], data, 0755); err != nil {
		fmt.Fprintln(os.Stderr, "Failed to write file:", err)
		os.Exit(1)
	}
}
