package mustacher

import (
	"image"

	"github.com/nfnt/resize"
)

// SmallElasticSizes are reasonable sizes to use for ElasticSearch.
// This does not contain any particularly large sizes, because faces
// can usually be found in smaller images with no problems but better
// performance.
var SmallElasticSizes = []int{400, 370, 340, 310, 280, 250,
	220, 205, 190, 175, 160, 145, 130, 115, 100, 85}

// ElasticSearch finds results from a database, but unlike
// Database.Matches(), it tries the search for a bunch of
// different scaled version of the image.
// It will not automatically remove near matches.
//
// The sizes argument specifies sizes to try.
// For each size S, the image will be resized so as to fit
// within an SxS rectangle.
// If sizes is nil, SmallElasticSizes will be used.
func ElasticSearch(d *Database, img image.Image, sizes []int) MatchSet {
	if sizes == nil {
		sizes = SmallElasticSizes
	}
	matches := MatchSet{}
	for _, size := range sizes {
		var scaledImage image.Image
		var scale float64
		if img.Bounds().Dx() > img.Bounds().Dy() {
			scale = float64(size) / float64(img.Bounds().Dx())
			scaledImage = resize.Resize(uint(size), 0, img, resize.Bilinear)
		} else {
			scale = float64(size) / float64(img.Bounds().Dy())
			scaledImage = resize.Resize(0, uint(size), img, resize.Bilinear)
		}
		subMatches := d.Matches(NewImage(scaledImage))
		for _, match := range subMatches {
			match.Width /= scale
			match.Center.X /= scale
			match.Center.Y /= scale
		}
		matches = append(matches, subMatches...)
	}
	return matches
}
