package mustacher

import "math"

// A Template is a two-dimensional representation of an object
// which can be matched against parts of a larger image.
type Template struct {
	image     *Image
	magnitude float64
	valueSum  float64
}

// NewTemplate generates a template from a training image.
func NewTemplate(i *Image) *Template {
	res := &Template{image: i}

	var magSquared float64
	for x := 0; x < i.Width(); x++ {
		for y := 0; y < i.Height(); y++ {
			brightness := i.BrightnessValue(x, y)
			res.valueSum += brightness
			magSquared += brightness * brightness
		}
	}
	res.magnitude = math.Sqrt(magSquared)

	return res
}

// Correlations returns all places in an image where the
// template has a correlation above a given threshold.
func (t *Template) Correlations(img *Image, threshold float64) CorrelationSet {
	res := make(CorrelationSet, 0)
	for y := 0; y < img.Height()-t.image.Height(); y++ {
		for x := 0; x < img.Width()-t.image.Width(); x++ {
			if corr := t.correlation(img, x, y); corr > threshold {
				res = append(res, &Correlation{
					Template:    t,
					Correlation: corr,
					X:           x,
					Y:           y,
				})
			}
		}
	}
	return res
}

// MaxCorrelation returns the maximum correlation (0 to 1)
// for the template anywhere in the given image.
// If the image contains close matches to the template,
// the returned value will be close to 1.
func (t *Template) MaxCorrelation(img *Image) float64 {
	var res float64
	for y := 0; y < img.Height()-t.image.Height(); y++ {
		for x := 0; x < img.Width()-t.image.Width(); x++ {
			res = math.Max(res, t.correlation(img, x, y))
		}
	}
	return res
}

func (t *Template) correlation(img *Image, startX, startY int) float64 {
	var imgSum float64
	var dotProduct float64
	for y := 0; y < t.image.Height(); y++ {
		for x := 0; x < t.image.Width(); x++ {
			imgPixel := img.BrightnessValue(startX+x, startY+y)
			templatePixel := t.image.BrightnessValue(x, y)
			imgSum += imgPixel * imgPixel
			dotProduct += imgPixel * templatePixel
		}
	}

	if imgSum == 0 || t.magnitude == 0 {
		if imgSum == t.magnitude {
			return 1
		} else {
			return 0
		}
	}

	return dotProduct / (math.Sqrt(imgSum) * t.magnitude)
}
