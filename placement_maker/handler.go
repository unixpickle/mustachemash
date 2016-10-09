package main

import (
	"encoding/json"
	"html/template"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"sync"
)

type Handler struct {
	FaceDir   string
	FaceFiles []string

	PlacementPath string
	Placements    []Placement

	Lock sync.Mutex
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/":
		h.handleNext(w, r)
	case "/save":
		h.handleSave(w, r)
	case "/image":
		h.handleImage(w, r)
	case "/mustache.svg":
		http.ServeFile(w, r, "mustache.svg")
	}
}

func (h *Handler) handleNext(w http.ResponseWriter, r *http.Request) {
	h.Lock.Lock()
	var allFiles []string
	for _, file := range h.FaceFiles {
		var found bool
		for _, place := range h.Placements {
			if place.ImageFile == file {
				found = true
				break
			}
		}
		if !found {
			allFiles = append(allFiles, file)
		}
	}
	h.Lock.Unlock()
	if len(allFiles) == 0 {
		w.Write([]byte("done all files!"))
		return
	}
	image := allFiles[rand.Intn(len(allFiles))]
	tempData, err := ioutil.ReadFile("index.template")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	parsed := template.Must(template.New("index").Parse(string(tempData)))
	parsed.Execute(w, image)
}

func (h *Handler) handleSave(w http.ResponseWriter, r *http.Request) {
	placement := Placement{
		ImageFile: r.FormValue("image-name"),
		CenterX:   forceParse(r.FormValue("x-coord")),
		CenterY:   forceParse(r.FormValue("y-coord")),
		Radius:    forceParse(r.FormValue("radius")),
		Angle:     forceParse(r.FormValue("angle")),
	}
	h.Lock.Lock()
	h.Placements = append(h.Placements, placement)
	data, _ := json.Marshal(h.Placements)
	ioutil.WriteFile(h.PlacementPath, data, 0755)
	h.Lock.Unlock()
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handler) handleImage(w http.ResponseWriter, r *http.Request) {
	name := path.Clean(r.URL.Query().Get("name"))
	_, baseName := path.Split(name)
	localPath := filepath.Join(h.FaceDir, baseName)
	f, err := os.Open(localPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	defer f.Close()
	io.Copy(w, f)
}

func forceParse(x string) float64 {
	res, _ := strconv.ParseFloat(x, 64)
	return res
}
