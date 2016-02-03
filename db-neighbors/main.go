package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"sort"

	"github.com/unixpickle/mustachemash/mustacher"
)

func main() {
	edgeLeeway := flag.Float64("edge-leeway", 0.15,
		"fraction of edges that can be omitted")

	flag.Parse()

	if len(flag.Args()) != 1 {
		fmt.Fprintln(os.Stderr, "Usage: db-neighbors [flags] <db.json>")
		flag.PrintDefaults()
		os.Exit(1)
	}

	contents, err := ioutil.ReadFile(flag.Args()[0])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	var database mustacher.Database
	if err := json.Unmarshal(contents, &database); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	list := templatePairList{}
	for i, template := range database.Templates {
		var closestMatch *mustacher.Template
		var closestCorrelation float64
		for j, check := range database.Templates {
			if i == j {
				continue
			}
			maxCorr := maxCorrelation(template.Image, check.Image, *edgeLeeway)
			if maxCorr > closestCorrelation {
				closestMatch = check
				closestCorrelation = maxCorr
			}
		}
		list = append(list, templatePair{template, closestMatch, closestCorrelation})
	}

	sort.Sort(list)
	for _, x := range list {
		fmt.Println(x)
	}
}

func maxCorrelation(img1, img2 *mustacher.Image, leeway float64) float64 {
	horizontalLeeway := int(leeway * float64(img1.Width) / 2)
	verticalLeeway := int(leeway * float64(img1.Height) / 2)

	var maxCorr float64
	for left := 0; left < horizontalLeeway; left++ {
		for right := 0; right < horizontalLeeway; right++ {
			for top := 0; top < verticalLeeway; top++ {
				for bottom := 0; bottom < verticalLeeway; bottom++ {
					cropped := cropImage(img1, left, right, top, bottom)
					template := mustacher.NewTemplate(cropped)
					maxCorr = math.Max(maxCorr, template.MaxCorrelation(img2))
				}
			}
		}
	}

	return maxCorr
}

func cropImage(img *mustacher.Image, left, right, top, bottom int) *mustacher.Image {
	width := img.Width - (left + right)
	height := img.Height - (top + bottom)
	res := &mustacher.Image{
		Width:            width,
		Height:           height,
		BrightnessValues: make([]float64, width*height),
	}

	pixelIdx := 0
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			res.BrightnessValues[pixelIdx] = img.BrightnessValue(x+left, y+top)
			pixelIdx++
		}
	}

	return res
}
