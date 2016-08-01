// Command imagestache adds mustaches to an image.
package main

import (
	"fmt"
	"image"
	_ "image/jpeg"
	"image/png"
	"os"

	"github.com/unixpickle/haar"
	"github.com/unixpickle/mustachemash/mustacher"
)

func main() {
	if len(os.Args) != 6 {
		fmt.Fprintf(os.Stderr, "Usage: %s faces.json nosemouths.json angler.json in_img out_img\n",
			os.Args[0])
		os.Exit(1)
	}

	detector, err := mustacher.LoadDetector(os.Args[1], os.Args[2], os.Args[3])
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to load detector:", err)
		os.Exit(1)
	}

	inDualImg, inImg, err := readImage(os.Args[4])
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to load input image:", err)
		os.Exit(1)
	}

	outImg := mustacher.Draw(inImg, detector.Match(inDualImg))
	outFile, err := os.Create(os.Args[5])
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

func readImage(path string) (*haar.DualImage, image.Image, error) {
	imgFile, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer imgFile.Close()
	img, _, err := image.Decode(imgFile)
	if err != nil {
		return nil, nil, err
	}
	intImg := haar.ImageIntegralImage(img)
	return haar.NewDualImage(intImg), img, nil
}
