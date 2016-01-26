package mustacher

import (
	"image"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"os"
)

const correlationThreshold = 0.5

type Coordinates struct {
	X float64
	Y float64
}

type CorrelationMatch struct {
	Coordinates Coordinates
	Correlation float64
}

// An Image represents a black and white rectangular image.
type Image struct {
	Width  int
	Height int

	// BrightnessValues go from left to right, then top to bottom.
	// Each brightness value ranges from 0 (black) to 1 (white).
	BrightnessValues []float64
}

// ReadImage decodes an image file at a given path.
func ReadImage(path string) (res *Image, err error) {
	reader, err := os.Open(path)
	if err != nil {
		return
	}
	defer reader.Close()
	decoded, _, err := image.Decode(reader)
	if err != nil {
		return
	}

	bounds := decoded.Bounds()
	res = &Image{
		Width:            bounds.Dx(),
		Height:           bounds.Dy(),
		BrightnessValues: make([]float64, 0, bounds.Dx()*bounds.Dy()),
	}

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			color := decoded.At(x, y)
			r, g, b, _ := color.RGBA()
			num := float64(r+g+b) / (0xffff * 3)
			res.BrightnessValues = append(res.BrightnessValues, num)
		}
	}

	return
}

// BrightnessValue returns the brightness value at x and y coordinates.
func (i *Image) BrightnessValue(x, y int) float64 {
	return i.BrightnessValues[x+y*i.Width]
}

// CorrelationSearch finds all the places in an image that are
// highly correlated with a smaller image.
// The threshold argument specifies the minimum correlation (between 0 and 1)
// for which matches will be reported.
func (i *Image) CorrelationSearch(subimage *Image, threshold float64) []CorrelationMatch {
	res := []CorrelationMatch{}
	for y := 0; y < i.Height-subimage.Height; y++ {
		for x := 0; x < i.Width-subimage.Width; x++ {
			corr := i.correlation(subimage, x, y)
			if corr > threshold {
				c := Coordinates{float64(x), float64(y)}
				match := CorrelationMatch{Coordinates: c, Correlation: corr}
				res = insertCorrelation(res, match, subimage)
			}
		}
	}
	return res
}

func (i *Image) correlation(subimage *Image, startX, startY int) float64 {
	var iSum float64
	var subimageSum float64
	var dotProduct float64
	for y := 0; y < subimage.Height; y++ {
		for x := 0; x < subimage.Width; x++ {
			iPixel := i.BrightnessValue(startX+x, startY+y)
			subimagePixel := subimage.BrightnessValue(startX+x, startY+y)
			iSum += math.Pow(iPixel, 2)
			subimageSum += math.Pow(subimagePixel, 2)
			dotProduct += iPixel * subimagePixel
		}
	}
	if iSum == 0 || subimageSum == 0 {
		if iSum == subimageSum {
			return 1
		} else {
			return 0
		}
	}
	return dotProduct / (math.Sqrt(iSum) * math.Sqrt(subimageSum))
}

func insertCorrelation(matches []CorrelationMatch, match CorrelationMatch,
	subimage *Image) []CorrelationMatch {
	overrideMatches := map[int]bool{}
	for i, otherResult := range matches {
		dx := math.Abs(otherResult.Coordinates.X - match.Coordinates.X)
		dy := math.Abs(otherResult.Coordinates.Y - match.Coordinates.Y)
		if dx <= float64(subimage.Width)/2 &&
			dy <= float64(subimage.Height)/2 {
			if otherResult.Correlation > match.Correlation {
				return matches
			} else {
				overrideMatches[i] = true
			}
		}
	}

	if len(overrideMatches) == 0 {
		return append(matches, match)
	}

	res := make([]CorrelationMatch, 0, len(matches)-len(overrideMatches)+1)
	for i, m := range matches {
		if !overrideMatches[i] {
			res = append(res, m)
		}
	}
	res = append(res, match)
	return res
}
