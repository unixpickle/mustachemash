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

// A Detector uses a face cascade, a nose-mouth classifier,
// and an angler to detect mustache destinations.
type Detector struct {
	Faces      *haar.Cascade
	NoseMouths *haar.Cascade
	Angler     *AnglerNode
}

// LoadDetector loads a detector from the filesystem,
// given the paths to the detection cascades and angler.
func LoadDetector(facesPath, noseMouthsPath, anglerPath string) (*Detector, error) {
	objects := []interface{}{
		new(haar.Cascade), new(haar.Cascade), new(AnglerNode),
	}
	for i, path := range []string{facesPath, noseMouthsPath, anglerPath} {
		data, err := ioutil.ReadFile(path)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(data, objects[i]); err != nil {
			return nil, err
		}
	}
	return &Detector{objects[0].(*haar.Cascade), objects[1].(*haar.Cascade),
		objects[2].(*AnglerNode)}, nil
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

		minY := faceImg.Height() / 3

		for y := minY; y <= faceImg.Height()-d.NoseMouths.WindowHeight; y++ {
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

		bestWindow := faceDual.Window(highestX, highestY, d.NoseMouths.WindowWidth,
			d.NoseMouths.WindowHeight)
		angle := d.Angler.Classify(bestWindow)

		xScale := float64(m.Width) / float64(faceImg.Width())
		yScale := float64(m.Height) / float64(faceImg.Height())

		matches[i] = &Match{
			X: xScale*(float64(highestX)+float64(d.NoseMouths.WindowWidth)/2) +
				float64(m.X),
			Y: yScale*(float64(highestY)+float64(d.NoseMouths.WindowHeight)/2) +
				float64(m.Y),
			Radius: 0.5 * xScale * float64(d.NoseMouths.WindowWidth),
			Angle:  angle,
		}
	}

	return matches
}
