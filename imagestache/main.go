// Command imagestache adds mustaches to an image.
package main

import (
	"fmt"
	"image"
	_ "image/jpeg"
	"image/png"
	"os"

	"github.com/unixpickle/mustachemash/mustacher"
)

func main() {
	if len(os.Args) != 5 {
		fmt.Fprintf(os.Stderr, "Usage: %s faces.json placer in_img out_img\n",
			os.Args[0])
		os.Exit(1)
	}

	detector, err := mustacher.LoadDetector(os.Args[1], os.Args[2])
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to load detector:", err)
		os.Exit(1)
	}

	inImg, err := readImage(os.Args[3])
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to load input image:", err)
		os.Exit(1)
	}

	outImg := mustacher.Draw(inImg, detector.Match(inImg))
	outFile, err := os.Create(os.Args[4])
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to create output:", err)
		os.Exit(1)
	}
	defer outFile.Close()
	if err := png.Encode(outFile, outImg); err != nil {
		fmt.Fprintln(os.Stderr, "Failed to encode output:", err)
		os.Exit(1)
	}
}

func readImage(path string) (image.Image, error) {
	imgFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer imgFile.Close()
	img, _, err := image.Decode(imgFile)
	return img, err
}
