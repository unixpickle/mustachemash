package mustacher

import (
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"os"
)

// An Image represents a black and white rectangular image.
type Image struct {
	width  int
	height int

	// brightnessValues go from left to right, then top to bottom.
	// Each brightness value ranges from 0 (black) to 1 (white).
	brightnessValues []float64
}

// NewImage creates an Image from an image.Image.
// The returned image will not reference the passed image.
func NewImage(i image.Image) *Image {
	bounds := i.Bounds()
	res := &Image{
		width:            bounds.Dx(),
		height:           bounds.Dy(),
		brightnessValues: make([]float64, 0, bounds.Dx()*bounds.Dy()),
	}

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			color := i.At(x, y)
			r, g, b, _ := color.RGBA()
			num := float64(r+g+b) / (0xffff * 3)
			res.brightnessValues = append(res.brightnessValues, num)
		}
	}

	return res
}

// ReadImageFile decodes an image file at a given path.
func ReadImageFile(path string) (res *Image, err error) {
	reader, err := os.Open(path)
	if err != nil {
		return
	}
	defer reader.Close()
	return ReadImage(reader)
}

// ReadImage decodes an image file using an io.Reader.
func ReadImage(reader io.Reader) (res *Image, err error) {
	decoded, _, err := image.Decode(reader)
	if err != nil {
		return
	}
	return NewImage(decoded), nil
}

// BrightnessValue returns the brightness value at x and y coordinates.
func (i *Image) BrightnessValue(x, y int) float64 {
	return i.brightnessValues[x+y*i.width]
}

// Width returns the width of the image in pixels.
func (i *Image) Width() int {
	return i.width
}

// Height returns the width of the image in pixels.
func (i *Image) Height() int {
	return i.height
}
