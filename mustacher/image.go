package mustacher

import (
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"math"
	"os"
)

const thresholdStrictness = 0.7

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
	width  int
	height int

	// brightnessValues go from left to right, then top to bottom.
	// Each brightness value ranges from 0 (black) to 1 (white).
	brightnessValues []float64

	magnitude            float64
	recommendedThreshold float64
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

	bounds := decoded.Bounds()
	res = &Image{
		width:            bounds.Dx(),
		height:           bounds.Dy(),
		brightnessValues: make([]float64, 0, bounds.Dx()*bounds.Dy()),
	}

	var valueSum float64
	var magSquared float64

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			color := decoded.At(x, y)
			r, g, b, _ := color.RGBA()
			num := float64(r+g+b) / (0xffff * 3)
			res.brightnessValues = append(res.brightnessValues, num)
			valueSum += num
			magSquared += num * num
		}
	}

	res.magnitude = math.Sqrt(magSquared)
	unchangingSignalMagnitude := math.Sqrt(float64(res.width * res.height))
	unchangingSignalDot := valueSum / (res.magnitude * unchangingSignalMagnitude)
	res.recommendedThreshold = thresholdStrictness + (1-thresholdStrictness)*unchangingSignalDot

	return
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

// RecommendedThreshold returns a reasonable threshold for
// CorrelationSearch when using this image as a subimage.
func (i *Image) RecommendedThreshold() float64 {
	return i.recommendedThreshold
}

// CorrelationSearch finds all the places in an image that are
// highly correlated with a smaller image.
// It automatically removes overlapping results by choosing the best match.
//
// The threshold argument specifies the minimum correlation (between 0 and 1)
// for which matches will be reported.
func (i *Image) CorrelationSearch(subimage *Image, threshold float64) []*CorrelationMatch {
	res := []*CorrelationMatch{}
	for y := 0; y < i.height-subimage.height; y++ {
		for x := 0; x < i.width-subimage.width; x++ {
			corr := i.correlation(subimage, x, y)
			if corr > threshold {
				c := Coordinates{float64(x), float64(y)}
				match := &CorrelationMatch{Coordinates: c, Correlation: corr}
				res = insertCorrelation(res, match, subimage)
			}
		}
	}
	return res
}

func (i *Image) correlation(subimage *Image, startX, startY int) float64 {
	var iSum float64
	var dotProduct float64
	for y := 0; y < subimage.height; y++ {
		for x := 0; x < subimage.width; x++ {
			iPixel := i.BrightnessValue(startX+x, startY+y)
			subimagePixel := subimage.BrightnessValue(x, y)
			iSum += iPixel * iPixel
			dotProduct += iPixel * subimagePixel
		}
	}

	if iSum == 0 || subimage.magnitude == 0 {
		if iSum == subimage.magnitude {
			return 1
		} else {
			return 0
		}
	}

	return dotProduct / (math.Sqrt(iSum) * subimage.magnitude)
}

func insertCorrelation(matches []*CorrelationMatch, match *CorrelationMatch,
	subimage *Image) []*CorrelationMatch {
	overrideMatches := map[int]bool{}
	for i, otherResult := range matches {
		dx := math.Abs(otherResult.Coordinates.X - match.Coordinates.X)
		dy := math.Abs(otherResult.Coordinates.Y - match.Coordinates.Y)
		if dx <= float64(subimage.width)/2 && dy <= float64(subimage.height)/2 {
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

	res := make([]*CorrelationMatch, 0, len(matches)-len(overrideMatches)+1)
	for i, m := range matches {
		if !overrideMatches[i] {
			res = append(res, m)
		}
	}
	res = append(res, match)
	return res
}
