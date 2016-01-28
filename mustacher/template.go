package mustacher

import "math"

// A Template is a two-dimensional representation of an object
// which can be matched against parts of a larger image.
type Template struct {
	image               *Image
	magnitude           float64
	remainingMagSquared []float64
}

// NewTemplate generates a template from a training image.
func NewTemplate(i *Image) *Template {
	res := &Template{image: i}

	var magSquared float64
	for x := 0; x < i.Width(); x++ {
		for y := 0; y < i.Height(); y++ {
			brightness := i.BrightnessValue(x, y)
			magSquared += brightness * brightness
		}
	}
	res.magnitude = math.Sqrt(magSquared)

	remaining := magSquared
	res.remainingMagSquared = make([]float64, i.Height())
	for y := 0; y < i.Height(); y++ {
		for x := 0; x < i.Width(); x++ {
			brightness := i.BrightnessValue(x, y)
			remaining -= brightness * brightness
		}
		res.remainingMagSquared[y] = remaining
	}

	return res
}

// Correlations returns all places in an image where the
// template has a correlation above a given threshold.
func (t *Template) Correlations(img *Image, threshold float64) CorrelationSet {
	res := make(CorrelationSet, 0)
	remMagSquared := make([]float64, t.image.Height())
	for y := 0; y < img.Height()-t.image.Height(); y++ {
		var oldMag float64
		for x := 0; x < img.Width()-t.image.Width(); x++ {
			var corr float64
			corr, oldMag = t.correlation(oldMag, remMagSquared, img, x, y, threshold)
			if corr > threshold {
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
	remMagSquared := make([]float64, t.image.Height())
	var res float64
	for y := 0; y < img.Height()-t.image.Height(); y++ {
		var oldMag float64
		for x := 0; x < img.Width()-t.image.Width(); x++ {
			var corr float64
			corr, oldMag = t.correlation(oldMag, remMagSquared, img, x, y, res)
			res = math.Max(res, corr)
		}
	}
	return res
}

func (t *Template) correlation(oldMag float64, oldRemMagSquared []float64, img *Image,
	startX, startY int, threshold float64) (corr float64, imgMag float64) {

	imgMag = oldMag
	if startX == 0 {
		for y := t.image.Height() - 1; y >= 0; y-- {
			oldRemMagSquared[y] = imgMag
			for x := 0; x < t.image.Width(); x++ {
				imgPixel := img.BrightnessValue(startX+x, startY+y)
				imgMag += imgPixel * imgPixel
			}
		}
	} else {
		var compoundedChange float64
		for y := t.image.Height() - 1; y >= 0; y-- {
			oldRemMagSquared[y] += compoundedChange

			var change float64
			imgPixel := img.BrightnessValue(startX-1, startY+y)
			change -= imgPixel * imgPixel
			imgPixel = img.BrightnessValue(startX+t.image.Width()-1, startY+y)
			change += imgPixel * imgPixel

			imgMag += change
			compoundedChange += change
		}
	}

	if imgMag == 0 || t.magnitude == 0 {
		if imgMag == t.magnitude {
			corr = 1
		}
		return
	}

	finalNormalization := 1.0 / (math.Sqrt(imgMag) * t.magnitude)

	var dotProduct float64
	for y := 0; y < t.image.Height(); y++ {
		for x := 0; x < t.image.Width(); x++ {
			imgPixel := img.BrightnessValue(startX+x, startY+y)
			templatePixel := t.image.BrightnessValue(x, y)
			dotProduct += imgPixel * templatePixel
		}
		optimalRemainingDot := math.Sqrt(oldRemMagSquared[y] * t.remainingMagSquared[y])
		if (dotProduct+optimalRemainingDot)*finalNormalization < threshold {
			return
		}
	}

	corr = dotProduct * finalNormalization
	return
}
