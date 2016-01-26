package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/unixpickle/mustachemash/mustacher"
)

var AssetsPath string
var MustacheDb *mustacher.Database

type MustacheInfo struct {
	CenterX    float64
	CenterY    float64
	Angle      float64
	MouthWidth float64
}

func main() {
	if len(os.Args) != 3 {
		log.Fatal("Usage: mustachemash <mustache_db> <port>")
	}

	port, err := strconv.Atoi(os.Args[2])
	if err != nil || port < 0 || port > 65535 {
		log.Fatal("Invalid port: " + os.Args[2])
	}

	MustacheDb, err = mustacher.ReadDatabase(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	// Setup global variables
	_, sourcePath, _, _ := runtime.Caller(0)
	AssetsPath = filepath.Join(filepath.Dir(sourcePath), "assets")

	http.HandleFunc("/mustache", HandleMustache)
	http.HandleFunc("/", HandleRoot)
	if err := http.ListenAndServe(":"+os.Args[2], nil); err != nil {
		log.Fatal(err)
	}
}

func HandleMustache(w http.ResponseWriter, r *http.Request) {
	urlStr := r.URL.Query().Get("url")
	resp, err := http.Get(urlStr)
	if err != nil {
		http.ServeFile(w, r, filepath.Join(AssetsPath, "noimage.html"))
		return
	}

	defer resp.Body.Close()
	img, err := mustacher.ReadImage(resp.Body)
	if err != nil {
		http.ServeFile(w, r, filepath.Join(AssetsPath, "noimage.html"))
		return
	}

	matches := MustacheDb.Search(img)
	infos := make([]*MustacheInfo, len(matches))
	for i, match := range matches {
		infos[i] = &MustacheInfo{
			CenterX:    match.Coordinates.X + match.Entry.MustacheCenter.X,
			CenterY:    match.Coordinates.Y + match.Entry.MustacheCenter.Y,
			Angle:      match.Entry.MustacheAngle,
			MouthWidth: float64(match.Entry.Image.Width()),
		}
	}
	jsonData, _ := json.Marshal(infos)

	page, _ := ioutil.ReadFile(filepath.Join(AssetsPath, "mustache.html"))
	pageStr := strings.Replace(string(page), "INFOS", string(jsonData), 1)
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(pageStr))
}

func HandleRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		cleanPath := strings.Replace(r.URL.Path, "..", "", -1)
		log.Print("Static file: ", cleanPath)
		http.ServeFile(w, r, filepath.Join(AssetsPath, cleanPath))
		return
	}
	http.ServeFile(w, r, filepath.Join(AssetsPath, "index.html"))
}
