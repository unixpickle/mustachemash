package mustacher

import (
	"errors"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type FloatCoordinates struct {
	X float64
	Y float64
}

func (f FloatCoordinates) Distance(f1 FloatCoordinates) float64 {
	return math.Sqrt(math.Pow(f1.X-f.X, 2) + math.Pow(f1.Y-f.Y, 2))
}

// A trainingImage stores a training image's template and its metadata.
type trainingImage struct {
	template  *Template
	id        int
	angle     float64
	center    FloatCoordinates
	width     float64
	threshold float64
}

// A DatabaseMatch provides all the information needed to add a mustache
// to an image.
type DatabaseMatch struct {
	TrainingID  int
	Correlation float64

	Width    float64
	Rotation float64
	Center   FloatCoordinates
}

// A Database represents a learned set of training images.
type Database struct {
	trainingImages          []*trainingImage
	templateToTrainingImage map[*Template]*trainingImage
}

// ReadDatabase reads image files from a given directory and
// parses their filenames for metadata.
//
// The strictness argument determines how the database will use its
// negative samples. If strictness is 0.0, then anything that matches
// even a little more than the negative samples will be reported.
// On the other hand, if strictness is 1.0, then nothing will be
// reported at all.
//
// If the mirror argument is set to true, then this will automatically
// generate a second, flipped training sample for every sample.
// The flipped sample will have a negated angle and mirrored X offset.
//
// An image file in the database can be one of two types:
// a negative sample, an image with no facial features, or
// a training sample, an image of a facial feature.
//
// Negative sample filenames should be of the form "negative_N.ext",
// where N is any number.
//
// Training sample filenames should be of the form "N_X_Y_Ddeg_W.ext"
// where N is any number, X and Y are the coordinates of the center of
// the mustache destination relative to the top left corner of the sample
// image, D is an angle (in degrees) that the mustache should be rotated
// rotated (clockwise), and W is the width of the mustache.
//
// The images must be PNG files or JPG files with the file extension
// ".png" or ".jpg".
func ReadDatabase(directory string, strictness float64, mirror bool) (db *Database, err error) {
	f, err := os.Open(directory)
	if err != nil {
		return
	}

	names, err := f.Readdirnames(0)
	if err != nil {
		return
	}

	res := &Database{
		trainingImages:          make([]*trainingImage, 0, len(names)-1),
		templateToTrainingImage: map[*Template]*trainingImage{},
	}
	negatives := make([]*Image, 0, len(names)-1)

	trainingExpr := regexp.MustCompile("([0-9]*)_([-0-9]*)_([-0-9]*)_([-0-9]*)deg_" +
		"([0-9]*)\\.(jpg|png)")
	negativeExpr := regexp.MustCompile("negative_[0-9]*\\.(jpg|png)")
	for _, n := range names {
		if strings.HasPrefix(n, ".") {
			continue
		}

		fullPath := filepath.Join(directory, n)
		if trainingMatch := trainingExpr.FindStringSubmatch(n); trainingMatch != nil {
			if err := res.addTrainingImage(trainingMatch, fullPath, mirror); err != nil {
				return nil, err
			}
		} else if negativeExpr.MatchString(n) {
			if img, err := ReadImageFile(fullPath); err != nil {
				return nil, err
			} else {
				negatives = append(negatives, img)
			}
		} else {
			return nil, errors.New("bad database filename: " + n)
		}
	}

	for _, t := range res.trainingImages {
		var maxCorrelation float64
		for _, negative := range negatives {
			maxCorrelation = math.Max(maxCorrelation, t.template.MaxCorrelation(negative))
		}
		t.threshold = 1*strictness + (1-strictness)*maxCorrelation
	}

	return res, nil
}

// Search finds mustache destinations in an image.
// It removes overlapping results by choosing the best match.
func (db *Database) Search(img *Image) []*DatabaseMatch {
	correlationSet := CorrelationSet{}
	for _, t := range db.trainingImages {
		correlations := t.template.Correlations(img, t.threshold)
		correlationSet = append(correlationSet, correlations...)
	}
	correlationSet = correlationSet.NonOverlappingSet()
	res := make([]*DatabaseMatch, len(correlationSet))
	for i, correlation := range correlationSet {
		t := db.templateToTrainingImage[correlation.Template]
		center := FloatCoordinates{
			X: float64(correlation.X) + t.center.X,
			Y: float64(correlation.Y) + t.center.Y,
		}
		ti := db.templateToTrainingImage[correlation.Template]
		match := &DatabaseMatch{
			TrainingID:  ti.id,
			Correlation: correlation.Correlation,

			Width:    ti.width,
			Rotation: ti.angle,
			Center:   center,
		}
		res[i] = match
	}
	return res
}

func (db *Database) addTrainingImage(match []string, path string, mirror bool) error {
	id, _ := strconv.Atoi(match[1])
	centerX, _ := strconv.Atoi(match[2])
	centerY, _ := strconv.Atoi(match[3])
	angle, _ := strconv.Atoi(match[4])
	width, _ := strconv.Atoi(match[5])
	image, err := ReadImageFile(path)
	if err != nil {
		return err
	}

	t := &trainingImage{
		template: NewTemplate(image),
		angle:    float64(angle),
		center:   FloatCoordinates{X: float64(centerX), Y: float64(centerY)},
		width:    float64(width),
		id:       id,
	}
	db.trainingImages = append(db.trainingImages, t)
	db.templateToTrainingImage[t.template] = t

	if mirror {
		t = &trainingImage{
			template: NewTemplate(image.Mirror()),
			angle:    -float64(angle),
			center:   FloatCoordinates{X: float64(width - centerX), Y: float64(centerY)},
			width:    float64(width),
			id:       -id,
		}
		db.trainingImages = append(db.trainingImages, t)
		db.templateToTrainingImage[t.template] = t
	}

	return nil
}
