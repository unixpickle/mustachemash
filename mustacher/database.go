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

// A DatabaseEntry includes an image of a nose and a mouth, along with
// information about where and how on that image a mustache should be
// positioned.
type DatabaseEntry struct {
	Image          *Image
	MustacheAngle  float64
	MustacheCenter Coordinates
}

// A DatabaseMatch reports the position of a DatabaseEntry which has
// been located inside a larger image.
type DatabaseMatch struct {
	Entry       *DatabaseEntry
	Coordinates Coordinates
	Correlation float64
}

// A Database is a collection of images that can be used to locate
// mustache targets in images.
type Database struct {
	Entries []*DatabaseEntry
}

// ReadDatabase reads image files from a given directory and
// parses their filenames for metadata.
//
// Image filenames should be of the form "N_X_Y_Ddeg.ext" where
// N is any unique number for the database, X is the X coordinate
// in pixels of the center of the mustache destination, Y is the
// Y coordinate in pixels of the center of the mustache destination,
// and D is an angle (in degrees) that the mustache should be
// rotated clockwise.
// The images must be PNG files or JPG files with the file extension
// ".png" or ".jpg".
func ReadDatabase(directory string) (db *Database, err error) {
	f, err := os.Open(directory)
	if err != nil {
		return
	}
	names, err := f.Readdirnames(0)
	if err != nil {
		return
	}

	expr := regexp.MustCompile("[0-9]*_([0-9]*)_([0-9]*)_([-0-9]*)deg")
	res := &Database{Entries: make([]*DatabaseEntry, 0, len(names))}
	for _, n := range names {
		if !strings.HasSuffix(n, "png") && !strings.HasSuffix(n, "jpg") {
			continue
		}
		match := expr.FindStringSubmatch(n)
		if match == nil {
			return nil, errors.New("bad database filename: " + n)
		}
		mustacheLeft, _ := strconv.Atoi(match[1])
		mustacheTop, _ := strconv.Atoi(match[2])
		mustacheAngle, _ := strconv.Atoi(match[3])
		image, err := ReadImageFile(filepath.Join(directory, n))
		if err != nil {
			return nil, err
		}
		entry := &DatabaseEntry{
			Image:          image,
			MustacheAngle:  float64(mustacheAngle),
			MustacheCenter: Coordinates{X: float64(mustacheLeft), Y: float64(mustacheTop)},
		}
		res.Entries = append(res.Entries, entry)
	}

	return res, nil
}

// Search looks through an Image to find potential matches against the
// Database.
// It automatically removes overlapping results by choosing the best match.
//
// The threshold parameter specifies the minimum correlation for a match to
// be reported. At 0, it will report all possible matches, and at 1 it will report
// only perfect matches.
func (db *Database) Search(i *Image, threshold float64) []*DatabaseMatch {
	res := []*DatabaseMatch{}
	for _, entry := range db.Entries {
		for _, match := range i.CorrelationSearch(entry.Image, threshold) {
			dbMatch := &DatabaseMatch{
				Entry:       entry,
				Coordinates: match.Coordinates,
				Correlation: match.Correlation,
			}
			res = insertDbMatch(res, dbMatch)
		}
	}
	return res
}

func insertDbMatch(matches []*DatabaseMatch, match *DatabaseMatch) []*DatabaseMatch {
	overrideMatches := map[int]bool{}
	for i, m := range matches {
		dx := math.Abs(m.Coordinates.X - match.Coordinates.X)
		dy := math.Abs(m.Coordinates.Y - match.Coordinates.Y)
		minWidth := math.Min(float64(m.Entry.Image.Width), float64(match.Entry.Image.Width))
		minHeight := math.Min(float64(m.Entry.Image.Height), float64(match.Entry.Image.Height))
		if dx <= minWidth/2 && dy <= minHeight/2 {
			if match.Correlation > m.Correlation {
				overrideMatches[i] = true
			} else {
				return matches
			}
		}
	}

	if len(overrideMatches) == 0 {
		return append(matches, match)
	}

	res := make([]*DatabaseMatch, 0, len(matches)-len(overrideMatches)+1)
	for i, m := range matches {
		if !overrideMatches[i] {
			res = append(res, m)
		}
	}
	res = append(res, match)
	return res
}
