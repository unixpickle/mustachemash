package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Placement struct {
	ImageFile string
	CenterX   float64
	CenterY   float64
	Radius    float64
	Angle     float64
}

func main() {
	if len(os.Args) != 4 {
		fmt.Fprintln(os.Stderr, "Usage: placement_maker <port> <face_dir> <placements.json>")
		os.Exit(1)
	}
	port, err := strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, "Invalid port:", os.Args[1])
		os.Exit(1)
	}
	faceDir := os.Args[2]
	placementsFile := os.Args[3]
	http.ListenAndServe(":"+strconv.Itoa(port), &Handler{
		FaceDir:       faceDir,
		FaceFiles:     listFaceFiles(faceDir),
		PlacementPath: placementsFile,
		Placements:    readPlacements(placementsFile),
	})
}

func listFaceFiles(dir string) []string {
	listing, err := ioutil.ReadDir(dir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	var res []string
	for _, file := range listing {
		ext := strings.ToLower(filepath.Ext(file.Name()))
		if ext == ".png" || ext == ".jpg" || ext == ".jpeg" {
			res = append(res, file.Name())
		}
	}
	return res
}

func readPlacements(file string) []Placement {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil
	}
	var res []Placement
	if err := json.Unmarshal(data, &res); err != nil {
		fmt.Fprintln(os.Stderr, "Invalid placements file:", err)
		os.Exit(1)
	}
	return res
}
