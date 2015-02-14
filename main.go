package main

import (
	"encoding/json"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

var AssetsPath string
var MainMustacher *Mustacher

func main() {
	if len(os.Args) != 2 {
		log.Fatal("Usage: mustachemash <port>")
	}

	port, err := strconv.Atoi(os.Args[1])
	if err != nil || port < 0 || port > 65535 {
		log.Fatal("Invalid port: " + os.Args[1])
	}

	MainMustacher, err = NewMustacher()
	if err != nil {
		log.Fatal(err)
	}

	// Setup global variables
	_, sourcePath, _, _ := runtime.Caller(0)
	AssetsPath = filepath.Join(filepath.Dir(sourcePath), "assets")

	http.HandleFunc("/mustache", HandleMustache)
	http.HandleFunc("/", HandleRoot)
	if err := http.ListenAndServe(":"+os.Args[1], nil); err != nil {
		log.Fatal(err)
	}
}

func HandleMustache(w http.ResponseWriter, r *http.Request) {
	// Request the image they used in the URL.
	urlStr := r.URL.Query().Get("url")
	resp, err := http.Get(urlStr)
	if err != nil {
		http.ServeFile(w, r, filepath.Join(AssetsPath, "noimage.html"))
		return
	}
	defer resp.Body.Close()
	
	// Find mustaches in the image.
	img, _, err := image.Decode(resp.Body)
	if err != nil {
		http.ServeFile(w, r, filepath.Join(AssetsPath, "noimage.html"))
		return
	}
	res := MainMustacher.FindMustaches(img)
	jsonData, _ := json.Marshal(res)
	
	// Write the page.
	page, _ := ioutil.ReadFile(filepath.Join(AssetsPath, "mustache.html"))
	pageStr := strings.Replace(string(page), "INFOS", string(jsonData), 1)
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(pageStr))
}

func HandleRoot(w http.ResponseWriter, r *http.Request) {
	// Serve static files if necessary.
	if r.URL.Path != "/" {
		cleanPath := strings.Replace(r.URL.Path, "..", "", -1)
		log.Print("Static file: ", cleanPath)
		http.ServeFile(w, r, filepath.Join(AssetsPath, cleanPath))
		return
	}

	http.ServeFile(w, r, filepath.Join(AssetsPath, "index.html"))
}
