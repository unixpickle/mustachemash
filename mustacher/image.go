package mustacher

import (
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"os"
)

// An Image stores a black and white raster image.
type Image struct {
	Width  int `json:"width"`
	Height int `json:"height"`

	// BrightnessValues go from left to right, then top to bottom.
	// Each brightness value ranges from 0 (black) to 1 (white).
	BrightnessValues []float64 `json:"brightness_values"`
}

// NewImage creates an Image from an image.Image.
// The returned image will not reference the passed image.
func NewImage(i image.Image) *Image {
	bounds := i.Bounds()
	res := &Image{
		Width:            bounds.Dx(),
		Height:           bounds.Dy(),
		BrightnessValues: make([]float64, 0, bounds.Dx()*bounds.Dy()),
	}

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			color := i.At(x, y)
			r, g, b, _ := color.RGBA()
			num := float64(r+g+b) / (0xffff * 3)
			res.BrightnessValues = append(res.BrightnessValues, num)
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
	return i.BrightnessValues[x+y*i.Width]
}

// Mirror returns the mirror image of this image.
func (i *Image) Mirror() *Image {
	res := &Image{
		Width:            i.Width,
		Height:           i.Height,
		BrightnessValues: make([]float64, len(i.BrightnessValues)),
	}
	valueIdx := 0
	for y := 0; y < i.Height; y++ {
		for x := 0; x < i.Width; x++ {
			res.BrightnessValues[valueIdx] = i.BrightnessValue(i.Width-(x+1), y)
			valueIdx++
		}
	}
	return res
}
