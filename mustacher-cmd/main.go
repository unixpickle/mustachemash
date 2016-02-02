package main

import (
	"flag"
	"fmt"
	"image"
	"image/png"
	"os"
	"strings"

	"github.com/unixpickle/mustachemash/mustacher"
)

func main() {
	filterFlag := flag.Bool("filter", true, "filter near matches")
	sensitivityFlag := flag.Float64("sensitivity", 0.2, "db threshold sensitivity")
	flag.Parse()

	if len(flag.Args()) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: mustacher-cmd [flags] <image> db1 [db2 ...]")
		flag.PrintDefaults()
		os.Exit(1)
	}

	imagePath := flag.Args()[0]
	file, err := os.Open(imagePath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	image, _, err := image.Decode(file)
	file.Close()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	dbs := make([]*mustacher.Database, len(flag.Args())-1)
	for i := range dbs {
		dbs[i], err = mustacher.LoadDatabase(flag.Args()[i+1], *sensitivityFlag)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		dbs[i].AddMirrors()
	}

	matches := mustacher.MatchSet{}
	for _, db := range dbs {
		matches = append(matches, mustacher.ElasticSearch(db, image, nil)...)
	}

	if *filterFlag {
		matches = matches.FilterNearMatches()
	}

	for _, match := range matches {
		fmt.Println("Match", match, "for template", match.Template)
	}

	drawn := mustacher.DrawMustaches(image, matches)
	output, err := os.Create(strings.Replace(imagePath, ".", "_mustached.", 1))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	png.Encode(output, drawn)
}
