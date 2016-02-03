package mustacher

import (
	"errors"
	"math"
	"path/filepath"
	"regexp"
	"strconv"
)

// A Template is a two-dimensional representation of an object
// which can be found in a larger image.
//
// A Template is used to find a "target" in an image.
// For example, a Template might be a picture of a nose, but
// indicate where the user's mouth (the target) is likely to be.
// Thus, the template image itself can be distinct from its target.
type Template struct {
	Image *Image `json:"image"`

	// UserInfo is a user-defined "tag" that can be encoded with
	// a Template.
	UserInfo string `json:"user_info"`

	// Threshold is a numerical value from 0 to 1, indicating how
	// closely this template must match an image to trigger a match.
	Threshold float64 `json:"threshold"`

	// TargetAngle is the rotation, in degrees, of the target.
	TargetAngle float64 `json:"target_angle"`

	// TargetCenter is the center of the target, measured from
	// the top left corner of a given match.
	// For example, if the template image appears at (50,30) in
	// the bigger image, and TargetCenter is (5,3), it indicates
	// that the target is located at (55,33) in the larger image.
	TargetCenter FloatCoordinates `json:"target_center"`

	// TargetWidth is the width of the target, measured in the
	// coordinate system of the template image.
	TargetWidth float64 `json:"target_width"`

	// magnitude is the vector magnitude of the image's brightness values.
	magnitude float64

	// See RemainingMagSquared in subregionInfo for info on this field.
	// This will be nil until the optimization data is computed.
	remainingMagSquared []float64
}

// NewTemplate generates a template from a training image.
// The Template's fields--besides the Image field--will be set
// to their zero values.
func NewTemplate(i *Image) *Template {
	return &Template{Image: i}
}

// LoadTemplate reads a template from a file.
//
// The filename should be of the form "U_X_Y_Ddeg_W.ext":
// - U is any alphanumeric string, used as the user info.
// - X and Y are the x and y components of the target center
// - D is the target angle
// - W is the target width
//
// The images must be PNG files or JPG files with the file extension
// ".png", ".jpg", or ".jpeg".
func LoadTemplate(path string) (*Template, error) {
	nameExp := regexp.MustCompile("([0-9]*)_([-\\.0-9]*)_([-\\.0-9]*)_" +
		"([-\\.0-9]*)deg_([\\.0-9]*)\\.(png|jpg|jpeg)")
	nameMatch := nameExp.FindStringSubmatch(filepath.Base(path))
	if nameMatch == nil {
		return nil, errors.New("invalid template filename: " + path)
	}

	centerX, _ := strconv.ParseFloat(nameMatch[2], 64)
	centerY, _ := strconv.ParseFloat(nameMatch[3], 64)
	angle, _ := strconv.ParseFloat(nameMatch[4], 64)
	width, _ := strconv.ParseFloat(nameMatch[5], 64)

	image, err := ReadImageFile(path)
	if err != nil {
		return nil, err
	}

	return &Template{
		Image:        image,
		TargetAngle:  angle,
		TargetCenter: FloatCoordinates{X: centerX, Y: centerY},
		TargetWidth:  width,
		UserInfo:     nameMatch[1],
	}, nil
}

// Mirror returns the mirror image of this template.
// It will automatically mirror the center X coordinate
// and negate the angle.
// It will append "flipped" to the UserInfo.
// It will not set the threshold on the new template.
func (t *Template) Mirror() *Template {
	return &Template{
		Image:       t.Image.Mirror(),
		TargetAngle: -float64(t.TargetAngle),
		TargetCenter: FloatCoordinates{
			X: float64(t.Image.Width) - t.TargetCenter.X,
			Y: t.TargetCenter.Y,
		},
		TargetWidth: t.TargetWidth,
		UserInfo:    t.UserInfo + "flipped",
	}
}

// Matches scans an image and returns all the matches where the
// template has a correlation above t.Threshold.
// This will not filter near matches; that is the caller's job.
func (t *Template) Matches(img *Image) MatchSet {
	if t.Image.Width > img.Width || t.Image.Height > img.Height {
		return MatchSet{}
	}

	if !t.hasOptimizationMetadata() {
		t.computeOptimizationMetadata()
	}

	res := make(MatchSet, 0)
	subregionInfo := newSubregionInfo(t.Image.Width, t.Image.Height)
	for y := 0; y < img.Height-t.Image.Height; y++ {
		subregionInfo.StartNewRow(img, y)
		for x := 0; x < img.Width-t.Image.Width; x++ {
			corr := t.correlation(subregionInfo, img, t.Threshold)
			if corr > t.Threshold {
				res = append(res, &Match{
					Template:    t,
					Correlation: corr,
					Center: FloatCoordinates{
						X: float64(x) + t.TargetCenter.X,
						Y: float64(y) + t.TargetCenter.Y,
					},
					Width: t.TargetWidth,
				})
			}
			subregionInfo.Roll(img)
		}
	}
	return res
}

// MaxCorrelation returns the maximum correlation (0 to 1)
// for the template anywhere in a given image.
// If the image contains close matches to the template,
// the returned value will be close to 1.
func (t *Template) MaxCorrelation(img *Image) float64 {
	return t.MaxCorrelationAll([]*Image{img})
}

// MaxCorrelationAll returns the maximum correlation that this
// template has in any of the supplied images.
//
// Using one call to MaxCorrelationAll() will almost surely
// perform better than calling MaxCorrelation multiple times,
// especially with many samples.
func (t *Template) MaxCorrelationAll(images []*Image) float64 {
	if !t.hasOptimizationMetadata() {
		t.computeOptimizationMetadata()
	}

	var res float64

	for _, img := range images {
		if t.Image.Width > img.Width || t.Image.Height > img.Height {
			continue
		}
		subregionInfo := newSubregionInfo(t.Image.Width, t.Image.Height)
		for y := 0; y < img.Height-t.Image.Height; y++ {
			subregionInfo.StartNewRow(img, y)
			for x := 0; x < img.Width-t.Image.Width; x++ {
				corr := t.correlation(subregionInfo, img, res)
				res = math.Max(res, corr)
				subregionInfo.Roll(img)
			}
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
	for y := 0; y < t.Image.Height; y++ {
		for x := 0; x < t.Image.Width; x++ {
			imgPixel := img.BrightnessValue(region.X+x, region.Y+y)
			templatePixel := t.Image.BrightnessValue(x, y)
			dotProduct += imgPixel * templatePixel
		}
		optimalRemainingDot := math.Sqrt(region.RemainingMagSquared[y] * t.remainingMagSquared[y])
		if (dotProduct+optimalRemainingDot)*finalNormalization < threshold {
			return 0
		}
	}

	return dotProduct * finalNormalization
}

func (t *Template) hasOptimizationMetadata() bool {
	return t.remainingMagSquared != nil
}

func (t *Template) computeOptimizationMetadata() {
	var magSquared float64
	for x := 0; x < t.Image.Width; x++ {
		for y := 0; y < t.Image.Height; y++ {
			brightness := t.Image.BrightnessValue(x, y)
			magSquared += brightness * brightness
		}
	}
	t.magnitude = math.Sqrt(magSquared)

	remaining := magSquared
	t.remainingMagSquared = make([]float64, t.Image.Height)
	for y := 0; y < t.Image.Height; y++ {
		for x := 0; x < t.Image.Width; x++ {
			brightness := t.Image.BrightnessValue(x, y)
			remaining -= brightness * brightness
		}
		t.remainingMagSquared[y] = remaining
	}
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
