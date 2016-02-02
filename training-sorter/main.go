package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/unixpickle/mustachemash/mustacher"
)

func main() {
	if len(os.Args) < 3 {
		printErrors("Usage: training-sorter <sort dir> db1.json [db2.json ...]",
			"",
			"training-sorter runs many images against template",
			"databases and separates the ones that work from the ones",
			"that don't.",
			"",
			"At first, the sort directory may just contain a bunch of",
			"images; this will create directories to sort them into.",
			"",
			"Image filenames must be of the form X_Y_D.(png|jpeg),",
			"where X and Y indicate the mustache center and D",
			"specifies the allowed margin of error.")
		os.Exit(1)
	}

	imageFiles, err := findImageFiles(os.Args[1])
	if err != nil {
		printErrors(err)
		os.Exit(1)
	}

	databases := make([]*mustacher.Database, len(os.Args)-2)
	for i := 2; i < len(os.Args); i++ {
		databases[i-2] = &mustacher.Database{}
		contents, err := ioutil.ReadFile(os.Args[i])
		if err != nil {
			printErrors(err)
			os.Exit(1)
		}
		if err := json.Unmarshal(contents, databases[i-2]); err != nil {
			printErrors(err)
			os.Exit(1)
		}
		if err != nil {
			printErrors(err)
			os.Exit(1)
		}
	}

	SortImages(os.Args[1], imageFiles, databases)
}

func printErrors(errors ...interface{}) {
	for _, err := range errors {
		fmt.Fprintln(os.Stderr, err)
	}
}

func findImageFiles(dir string) ([]string, error) {
	subdirs := []Category{"", FalsePositive, TruePositive, BothPositive, Negative}
	res := []string{}
	for _, subdir := range subdirs {
		dirPath := filepath.Join(dir, string(subdir))
		file, err := os.Open(dirPath)
		if err != nil {
			if subdir == "" {
				return nil, err
			}
			if err := os.Mkdir(dirPath, 0755); err != nil {
				return nil, err
			}
			file, err = os.Open(dirPath)
			if err != nil {
				return nil, err
			}
			continue
		}
		contents, err := file.Readdirnames(-1)
		file.Close()
		if err != nil {
			return nil, err
		}
		for _, name := range contents {
			categoryName := Category(name)
			if subdir == "" && (categoryName == FalsePositive || categoryName == TruePositive ||
				categoryName == BothPositive || categoryName == Negative) {
				continue
			} else if strings.HasPrefix(name, ".") {
				continue
			}
			res = append(res, filepath.Join(dirPath, name))
		}
	}
	return res, nil
}
