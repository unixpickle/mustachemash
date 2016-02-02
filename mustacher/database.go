package mustacher

import (
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// A Database stores and manipulates a set of training images.
type Database struct {
	Templates  []*Template `json:"template"`
	Negatives  []*Image    `json:"negatives"`
	Strictness float64     `json:"strictness"`
}

// LoadDatabase loads template files from a directory and runs the
// templates against a battery of negative samples.
//
// The strictness argument determines how the database will use its
// negative samples to determine threshold values for its templates.
// If strictness is 0.0, then anything that matches even a little more
// than the negative samples will be reported.
// On the other hand, if strictness is 1.0, no matches will be reported.
//
// An image file in the database can be one of two types:
// a negative sample (an image that no templates should match) or
// a training sample, an image to use for a template.
//
// Negative sample filenames should be of the form "negative_N.ext",
// where N is any number.
//
// Training sample filenames should be of the form prescribed for
// LoadTemplate().
func LoadDatabase(directory string, strictness float64) (db *Database, err error) {
	f, err := os.Open(directory)
	if err != nil {
		return
	}

	names, err := f.Readdirnames(0)
	if err != nil {
		return
	}

	res := &Database{
		Strictness: strictness,
		Templates:  make([]*Template, 0, len(names)-1),
		Negatives:  make([]*Image, 0, len(names)-1),
	}

	negativeExpr := regexp.MustCompile("negative_[0-9]*\\.(jpg|png)")
	for _, n := range names {
		if strings.HasPrefix(n, ".") {
			continue
		}

		fullPath := filepath.Join(directory, n)

		if negativeExpr.MatchString(n) {
			if img, err := ReadImageFile(fullPath); err != nil {
				return nil, err
			} else {
				res.Negatives = append(res.Negatives, img)
			}
			continue
		}

		if template, err := LoadTemplate(fullPath); err != nil {
			return nil, err
		} else {
			res.Templates = append(res.Templates, template)
		}
	}

	for _, t := range res.Templates {
		maxCorrelation := t.MaxCorrelationAll(res.Negatives)
		t.Threshold = 1*strictness + (1-strictness)*maxCorrelation
	}

	return res, nil
}

// AddMirrors adds the mirror image of each template to the database.
//
// This will update threshold values of the old templates as well as
// those of the new ones. The new threshold values will be as if the
// negative samples included mirrors themselves, so that a mirrored
// template will always have the same threshold as the original.
func (db *Database) AddMirrors() {
	newTemplates := make([]*Template, len(db.Templates))
	for i, template := range db.Templates {
		mirror := template.Mirror()
		newTemplates[i] = mirror

		maxCorrelation := mirror.MaxCorrelationAll(db.Negatives)
		mirrorThreshold := 1*db.Strictness + (1-db.Strictness)*maxCorrelation

		threshold := math.Max(mirrorThreshold, template.Threshold)
		mirror.Threshold = threshold
		template.Threshold = threshold
	}
	db.Templates = append(db.Templates, newTemplates...)
}

// Matches runs an image against the database,
// reporting matches found for every template.
// It will not automatically remove near matches.
func (db *Database) Matches(img *Image) MatchSet {
	res := MatchSet{}
	for _, t := range db.Templates {
		res = append(res, t.Matches(img)...)
	}
	return res
}
