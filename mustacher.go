package main

import (
	"github.com/unixpickle/go-flandmark"
	"image"
	"math"
)

const MustacheWidthByHeight = 1854.0 / 473.0

// MustacheInfo stores data which the JavaScript front-end can use to overlay a
// mustache.
type MustacheInfo struct {
	Angle  float64 `json:"angle"`
	Height float64 `json:"height"`
	Left   float64 `json:"left"`
	Top    float64 `json:"top"`
	Width  float64 `json:"width"`
}

// A Mustacher is a queue for finding mustaches in images.
type Mustacher struct {
	faces *flandmark.Cascade
	model *flandmark.Model
	input chan mustacherInput
}

// NewMustacher creates a new Mustacher with one goroutine.
func NewMustacher() (*Mustacher, error) {
	faces, err := flandmark.LoadFaceCascade()
	if err != nil {
		return nil, err
	}
	model, err := flandmark.LoadDefaultModel()
	if err != nil {
		return nil, err
	}
	res := &Mustacher{faces, model, make(chan mustacherInput)}
	go res.loop()
	return res, nil
}

// Close closes the Mustacher, ending its background goroutine.
func (m *Mustacher) Close() {
	close(m.input)
}

// FindMustaches sends an image to the queue and waits for a response from it.
func (m *Mustacher) FindMustaches(img image.Image) []MustacheInfo {
	response := make(chan []MustacheInfo)
	m.input <- mustacherInput{img, response}
	return <-response
}

// defaultFaces creates one mustache in the center of the image.
func (m *Mustacher) defaultMustaches(img image.Image) []MustacheInfo {
	imgWidth := float64(img.Bounds().Dx())
	imgHeight := float64(img.Bounds().Dy())

	width := imgWidth
	height := imgWidth / MustacheWidthByHeight
	y := (imgHeight - height) / 2

	return []MustacheInfo{MustacheInfo{0, height, 0, y, width}}
}

// loop takes mustaches from the queue and finds their mustaches.
func (m *Mustacher) loop() {
	for input := range m.input {
		img, err := flandmark.GoGrayImage(input.img)
		if err != nil {
			input.out <- m.defaultMustaches(input.img)
			continue
		}

		// Detect the faces in the image.
		faces, err := m.faces.Detect(img, 1.1, 2, flandmark.Size{40, 40},
			flandmark.Size{1000000, 1000000})
		if err != nil || len(faces) == 0 {
			input.out <- m.defaultMustaches(input.img)
			continue
		}

		// Detect the mouth of each face.
		mustaches := make([]MustacheInfo, 0, len(faces))
		for _, face := range faces {
			points, err := m.model.Detect(img, face)
			if err != nil || len(points) < 5 {
				continue
			}
			leftMouth := points[3]
			rightMouth := points[4]
			mustache := mustacheForMouth(leftMouth, rightMouth)
			mustaches = append(mustaches, mustache)
		}
		if len(mustaches) == 0 {
			input.out <- m.defaultMustaches(input.img)
		} else {
			input.out <- mustaches
		}
	}
}

// mustacheForMouth computes the CSS layout of a mustache given mouth
// coordinates.
func mustacheForMouth(left, right flandmark.Point) MustacheInfo {
	// The width of the mustache is equal to the distance of the two points.
	xDiff := float64(right.X - left.X)
	yDiff := float64(right.Y - left.Y)
	width := math.Sqrt(math.Pow(xDiff, 2) + math.Pow(yDiff, 2))
	height := width / MustacheWidthByHeight

	// The mustache must be centered between the two points.
	// TODO: perhaps raise the mustache up above the mouth!
	centerX := float64(right.X+left.X) / 2
	centerY := float64(right.Y+left.Y) / 2
	cssLeft := centerX - width/2
	cssTop := centerY - height/2

	// The mustache must span the two points, meaning it must be rotated at a
	// specific angle which we can find using trig.
	angle := math.Atan2(yDiff, xDiff)

	return MustacheInfo{angle, height, cssLeft, cssTop, width}
}

type mustacherInput struct {
	img image.Image
	out chan []MustacheInfo
}
