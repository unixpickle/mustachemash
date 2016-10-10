package mustacher

import (
	"encoding/json"
	"fmt"
	"image"
	"io/ioutil"

	"github.com/nfnt/resize"
	"github.com/unixpickle/autofunc"
	"github.com/unixpickle/haar"
	"github.com/unixpickle/weakai/neuralnet"
)

const (
	faceOverlapThreshold = 0.7
	faceScanStride       = 1.5
	placerImageSize      = 28
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
	Faces  *haar.Cascade
	Placer neuralnet.Network
}

// LoadDetector loads a detector from the filesystem,
// given the paths to the detection cascades and placer.
func LoadDetector(facesPath, placerPath string) (*Detector, error) {
	var err error
	data := make([][]byte, 2)
	paths := []string{facesPath, placerPath}
	for i, path := range paths {
		data[i], err = ioutil.ReadFile(path)
		if err != nil {
			return nil, err
		}
	}
	res := &Detector{}
	if err := json.Unmarshal(data[0], &res.Faces); err != nil {
		return nil, fmt.Errorf("deserialize faces cascade: %s", err)
	}
	res.Placer, err = neuralnet.DeserializeNetwork(data[1])
	if err != nil {
		return nil, fmt.Errorf("deserialize placer network: %s", err)
	}
	return res, nil
}

// Match finds all of the mustache destinations in
// an image.
func (d *Detector) Match(img image.Image) []*Match {
	dualImage := haar.NewDualImage(haar.ImageIntegralImage(img))
	faceMatches := d.Faces.Scan(dualImage, 0, faceScanStride)
	faceMatches = faceMatches.JoinOverlaps(faceOverlapThreshold)

	matches := make([]*Match, len(faceMatches))
	for i, m := range faceMatches {
		cropped := image.NewRGBA(image.Rect(0, 0, m.Width, m.Height))
		for y := 0; y < m.Height; y++ {
			for x := 0; x < m.Width; x++ {
				cropped.Set(x, y, img.At(x+m.X+img.Bounds().Min.X,
					y+m.Y+img.Bounds().Min.Y))
			}
		}
		scale := float64(cropped.Bounds().Dx()) / placerImageSize
		scaled := resize.Resize(placerImageSize, placerImageSize, cropped, resize.Bilinear)
		inTensor := neuralnet.NewTensor3(placerImageSize, placerImageSize, 3)
		for y := 0; y < scaled.Bounds().Dy(); y++ {
			for x := 0; x < scaled.Bounds().Dx(); x++ {
				r, g, b, _ := scaled.At(x+scaled.Bounds().Min.X,
					y+scaled.Bounds().Min.Y).RGBA()
				inTensor.Set(x, y, 0, float64(r)/0xffff)
				inTensor.Set(x, y, 1, float64(g)/0xffff)
				inTensor.Set(x, y, 2, float64(b)/0xffff)
			}
		}
		out := d.Placer.Apply(&autofunc.Variable{Vector: inTensor.Data}).Output()
		matches[i] = &Match{
			X:      out[0]*placerImageSize*scale + float64(m.X),
			Y:      out[1]*placerImageSize*scale + float64(m.Y),
			Radius: out[2] * placerImageSize * scale,
			Angle:  out[3],
		}
	}

	return matches
}
