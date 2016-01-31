package mustacher

import (
	"image"
	"math"
	"sort"

	"github.com/nfnt/resize"
)

// SmallElasticSizes are reasonable sizes to use for ElasticSearch.
// This does not contain any particularly large sizes, because faces
// can usually be found in smaller images with no problems but better
// performance.
var SmallElasticSizes = []int{400, 370, 340, 310, 280, 250,
	220, 205, 190, 175, 160, 145, 130, 115, 100, 85}

// ElasticSearch finds results from a database, but unlike
// Database.Search(), it tries the search for a bunch of
// different scaled version of the image.
//
// The sizes argument specifies sizes to try.
// For each size S, the image will be resized so as to fit
// within an SxS rectangle.
// If sizes is nil, SmallElasticSizes will be used.
func ElasticSearch(d *Database, img image.Image, sizes []int) []*DatabaseMatch {
	if sizes == nil {
		sizes = SmallElasticSizes
	}
	matches := []*DatabaseMatch{}
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
		subMatches := d.Search(NewImage(scaledImage))
		for _, match := range subMatches {
			match.Width /= scale
			match.Center.X /= scale
			match.Center.Y /= scale
		}
		matches = append(matches, subMatches...)
	}
	return removeNearDuplicateMatches(matches)
}

func removeNearDuplicateMatches(m []*DatabaseMatch) []*DatabaseMatch {
	sort.Sort(sortableMatches(m))
	res := make([]*DatabaseMatch, 0, len(m))

MatchLoop:
	for _, match := range m {
		for _, existingMatch := range res {
			minDistance := math.Max(match.Width/2, existingMatch.Width/2)
			if match.Center.Distance(existingMatch.Center) < minDistance {
				continue MatchLoop
			}
		}
		res = append(res, match)
	}

	return res
}

type sortableMatches []*DatabaseMatch

func (s sortableMatches) Len() int {
	return len(s)
}

func (s sortableMatches) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s sortableMatches) Less(i, j int) bool {
	return s[i].Correlation > s[j].Correlation
}
