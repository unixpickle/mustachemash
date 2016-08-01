// Command train_angler trains a decision tree to identify
// the angle at which a mouth/nose is rotated.
//
// The command takes a directory containing three sub-dirs:
// straight, to_left, and to_right. These contain faces not
// slanting, slanting down to the left, and slanting
// down to the right respectively.
package main

import (
	"encoding/json"
	"fmt"
	"image"
	"io/ioutil"
	"log"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/unixpickle/haar"
	"github.com/unixpickle/mustachemash/mustacher"
	"github.com/unixpickle/weakai/idtrees"
)

const (
	slantAngle = 10 * math.Pi / 180
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, "Usage: %s sample_dir output.json\n", os.Args[0])
		os.Exit(1)
	}

	log.Println("Loading samples ...")
	straight, right, left := readAllImages()

	width, height := straight[0].Width(), straight[1].Height()
	features := haar.AllFeatures(width, height)

	var samples []idtrees.Sample
	for _, img := range straight {
		samples = append(samples, &Sample{img, 0, features})
	}
	for _, img := range right {
		samples = append(samples, &Sample{img, slantAngle, features})
	}
	for _, img := range left {
		samples = append(samples, &Sample{img, -slantAngle, features})
	}

	attrs := make([]idtrees.Attr, len(features))
	for i := range features {
		attrs[i] = i
	}

	log.Println("Building tree ...")

	tree := idtrees.ID3(samples, attrs, 0)
	savableTree := BuildTreeNode(features, tree)

	encoded, err := json.Marshal(savableTree)
	if err != nil {
		panic(err)
	}
	if err := ioutil.WriteFile(os.Args[2], encoded, 0755); err != nil {
		fmt.Fprintln(os.Stderr, "Failed to save tree:", err)
		os.Exit(1)
	}
}

func readAllImages() (straightImgs, rightImgs, leftImgs []haar.IntegralImage) {
	straight, err1 := ioutil.ReadDir(filepath.Join(os.Args[1], "straight"))
	right, err2 := ioutil.ReadDir(filepath.Join(os.Args[1], "to_right"))
	left, err3 := ioutil.ReadDir(filepath.Join(os.Args[1], "to_left"))
	if err1 != nil {
		fmt.Fprintln(os.Stderr, "Missing straight/ sub-directory:", err1)
	}
	if err2 != nil {
		fmt.Fprintln(os.Stderr, "Missing to_right/ sub-directory:", err2)
	}
	if err3 != nil {
		fmt.Fprintln(os.Stderr, "Missing to_left/ sub-directory:", err3)
	}
	if err1 != nil || err2 != nil || err3 != nil {
		os.Exit(1)
	}

	straightImgs, err1 = readImages(filepath.Join(os.Args[1], "straight"), straight)
	if err1 != nil {
		fmt.Fprintln(os.Stderr, "Failed to load straights:", err1)
	}
	rightImgs, err2 = readImages(filepath.Join(os.Args[1], "to_right"), right)
	if err2 != nil {
		fmt.Fprintln(os.Stderr, "Failed to load rights:", err2)
	}
	leftImgs, err3 = readImages(filepath.Join(os.Args[1], "to_left"), left)
	if err3 != nil {
		fmt.Fprintln(os.Stderr, "Failed to load lefts:", err3)
	}
	if err1 != nil || err2 != nil || err3 != nil {
		os.Exit(1)
	}

	if len(straightImgs) == 0 {
		fmt.Fprintln(os.Stderr, "No straight images.")
		os.Exit(1)
	}

	return
}

func readImages(dirPath string, listing []os.FileInfo) ([]haar.IntegralImage, error) {
	var res []haar.IntegralImage
	for _, item := range listing {
		if strings.HasPrefix(item.Name(), ".") {
			continue
		}
		path := filepath.Join(dirPath, item.Name())
		f, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		img, _, err := image.Decode(f)
		f.Close()
		if err != nil {
			return nil, err
		}
		// Use a DualImage to normalize the colors.
		intImg := haar.ImageIntegralImage(img)
		dualImg := haar.NewDualImage(intImg)
		res = append(res, dualImg.Window(0, 0, intImg.Width(), intImg.Height()))
	}
	return res, nil
}

type Sample struct {
	Image    haar.IntegralImage
	Angle    float64
	Features []*haar.Feature
}

func (s *Sample) Class() idtrees.Class {
	return s.Angle
}

func (s *Sample) Attr(key idtrees.Attr) idtrees.Val {
	return s.Features[key.(int)].Value(s.Image)
}

func BuildTreeNode(features []*haar.Feature, t *idtrees.Tree) *mustacher.AnglerNode {
	if t.Classification != nil {
		var maxKey, maxVal float64
		for key, val := range t.Classification {
			if val >= maxVal {
				maxVal = val
				maxKey = key.(float64)
			}
		}
		return &mustacher.AnglerNode{Classification: maxKey}
	}
	return &mustacher.AnglerNode{
		Feature:   features[t.Attr.(int)],
		Cutoff:    t.NumSplit.Threshold.(float64),
		LessEqual: BuildTreeNode(features, t.NumSplit.LessEqual),
		Greater:   BuildTreeNode(features, t.NumSplit.Greater),
	}
}
