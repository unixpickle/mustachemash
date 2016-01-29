package mustacher

import "math"

// A Template is a two-dimensional representation of an object
// which can be matched against parts of a larger image.
type Template struct {
	image     *Image
	magnitude float64

	// See RemainingMagSquared in subregionInfo for info on this field.
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
	subregionInfo := newSubregionInfo(t.image.Width(), t.image.Height())
	for y := 0; y < img.Height()-t.image.Height(); y++ {
		subregionInfo.StartNewRow(img, y)
		for x := 0; x < img.Width()-t.image.Width(); x++ {
			corr := t.correlation(subregionInfo, img, threshold)
			if corr > threshold {
				res = append(res, &Correlation{
					Template:    t,
					Correlation: corr,
					X:           x,
					Y:           y,
				})
			}
			subregionInfo.Roll(img)
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
	subregionInfo := newSubregionInfo(t.image.Width(), t.image.Height())
	for y := 0; y < img.Height()-t.image.Height(); y++ {
		subregionInfo.StartNewRow(img, y)
		for x := 0; x < img.Width()-t.image.Width(); x++ {
			corr := t.correlation(subregionInfo, img, res)
			res = math.Max(res, corr)
			subregionInfo.Roll(img)
		}
	}
	return res
}

func (t *Template) correlation(region *subregionInfo, img *Image, threshold float64) float64 {
	if region.MagSquared == 0 || t.magnitude == 0 {
		if region.MagSquared == t.magnitude {
			// Let's just say that two zero vectors are perfectly correlated.
			return 1
		}
		return 0
	}

	finalNormalization := 1.0 / (math.Sqrt(region.MagSquared) * t.magnitude)

	var dotProduct float64
	for y := 0; y < t.image.Height(); y++ {
		for x := 0; x < t.image.Width(); x++ {
			imgPixel := img.BrightnessValue(region.X+x, region.Y+y)
			templatePixel := t.image.BrightnessValue(x, y)
			dotProduct += imgPixel * templatePixel
		}
		optimalRemainingDot := math.Sqrt(region.RemainingMagSquared[y] * t.remainingMagSquared[y])
		if (dotProduct+optimalRemainingDot)*finalNormalization < threshold {
			return 0
		}
	}

	return dotProduct * finalNormalization
}

// subregionInfo stores information about a subregion of an image.
// This information can be "rolled" to an adjacent subregion, meaning
// that an imageSubregionInfo can be updated for the "next" subregion
// without recomputing everything.
type subregionInfo struct {
	width  int
	height int

	X int
	Y int

	// MagSquared is the sum of the squares of the brightness values
	// in the current subregion.
	MagSquared float64

	// RemainingMagSquared maps y values to magnitudes.
	// The y-th entry is equal to magSquared minus the sum of the
	// squares of the pixel values of the first (y+1) rows of the
	// subregion.
	RemainingMagSquared []float64
}

func newSubregionInfo(width, height int) *subregionInfo {
	return &subregionInfo{
		width:  width,
		height: height,

		RemainingMagSquared: make([]float64, height),
	}
}

// StartNewRow computes completely fresh values for the leftmost
// subregion at the given y coordinate.
func (c *subregionInfo) StartNewRow(img *Image, newY int) {
	c.MagSquared = 0
	c.X = 0
	c.Y = newY
	for y := c.height - 1; y >= 0; y-- {
		c.RemainingMagSquared[y] = c.MagSquared
		for x := 0; x < c.width; x++ {
			imgPixel := img.BrightnessValue(x, newY+y)
			c.MagSquared += imgPixel * imgPixel
		}
	}
}

// Roll computes the info for the subregion to the right of the current one.
func (c *subregionInfo) Roll(img *Image) {
	c.X++
	var compoundedChange float64
	for y := c.height - 1; y >= 0; y-- {
		c.RemainingMagSquared[y] += compoundedChange

		var change float64
		imgPixel := img.BrightnessValue(c.X-1, c.Y+y)
		change -= imgPixel * imgPixel
		imgPixel = img.BrightnessValue(c.X+c.width-1, c.Y+y)
		change += imgPixel * imgPixel

		compoundedChange += change
	}
	c.MagSquared += compoundedChange
}
