package mustacher

import (
	"encoding/json"
	"io/ioutil"

	"github.com/unixpickle/haar"
)

const (
	faceOverlapThreshold = 0.7
	faceScanStride       = 1.5
)

// A Match represents the destination for a mustache.
type Match struct {
	// These are the coordinates of the mustache's center.
	X float64
	Y float64

	// Radius is the radius of the mustache in pixels.
	Radius float64

	// Angle is the rotation of the mustache in radians.
	Angle float64
}

// A Detector uses a face and nose-mouth classifier to
// detect locations at which mustaches should be added.
type Detector struct {
	Faces      *haar.Cascade
	NoseMouths *haar.Cascade
}

// LoadDetector loads a detector from the filesystem,
// given the paths to the detection cascades.
func LoadDetector(facesPath, noseMouthsPath string) (*Detector, error) {
	var cascades [2]*haar.Cascade
	for i, path := range []string{facesPath, noseMouthsPath} {
		data, err := ioutil.ReadFile(path)
		if err != nil {
			return nil, err
		}
		var cascade haar.Cascade
		if err := json.Unmarshal(data, &cascade); err != nil {
			return nil, err
		}
		cascades[i] = &cascade
	}
	return &Detector{cascades[0], cascades[1]}, nil
}

// Match finds all of the mustache destinations in
// an image.
func (d *Detector) Match(img *haar.DualImage) []*Match {
	faceMatches := d.Faces.Scan(img, 0, faceScanStride)
	faceMatches = faceMatches.JoinOverlaps(faceOverlapThreshold)

	matches := make([]*Match, len(faceMatches))
	for i, m := range faceMatches {
		var highestSum float64
		var highestX, highestY int

		faceImg := haar.ScaleIntegralImage(img.Window(m.X, m.Y, m.Width, m.Height),
			d.Faces.WindowWidth, d.Faces.WindowHeight)
		faceDual := haar.NewDualImage(faceImg)

		for y := 0; y <= faceImg.Height()-d.NoseMouths.WindowHeight; y++ {
			for x := 0; x <= faceImg.Width()-d.NoseMouths.WindowWidth; x++ {
				cropped := faceDual.Window(x, y, d.NoseMouths.WindowWidth,
					d.NoseMouths.WindowHeight)
				sum := d.NoseMouths.Layers[0].Sum(cropped)
				if sum > highestSum || (x == 0 && y == 0) {
					highestX = x
					highestY = y
					highestSum = sum
				}
			}
		}

		xScale := float64(m.Width) / float64(faceImg.Width())
		yScale := float64(m.Height) / float64(faceImg.Height())

		matches[i] = &Match{
			X: xScale*(float64(highestX)+float64(d.NoseMouths.WindowWidth)/2) +
				float64(m.X),
			Y: yScale*(float64(highestY)+float64(d.NoseMouths.WindowHeight)/2) +
				float64(m.Y),
			Radius: 0.5 * xScale * float64(d.NoseMouths.WindowWidth),
			Angle:  0,
		}
	}

	return matches
}
