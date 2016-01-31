package main

import (
	"errors"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/unixpickle/mustachemash/mustacher"
)

type Category string

var (
	FalsePositive Category = "false_positive"
	TruePositive  Category = "true_positive"
	BothPositive  Category = "both_positive"
	Negative      Category = "negative"
)

func CategorizeImage(path string, dbs []*mustacher.Database) (Category, error) {
	nameExpr := regexp.MustCompile("^([0-9]*)_([0-9]*)_([0-9]*).(jpg|png|jpeg)$")
	match := nameExpr.FindStringSubmatch(filepath.Base(path))
	if match == nil {
		return "", errors.New("Invalid filename: " + filepath.Base(path))
	}
	centerX, _ := strconv.ParseFloat(match[1], 64)
	centerY, _ := strconv.ParseFloat(match[2], 64)
	maxDist, _ := strconv.ParseFloat(match[3], 64)

	reader, err := os.Open(path)
	if err != nil {
		return "", err
	}

	img, _, err := image.Decode(reader)
	reader.Close()
	if err != nil {
		return "", err
	}

	dbMatches := make([][]*mustacher.DatabaseMatch, len(dbs))
	for i, db := range dbs {
		dbMatches[i] = mustacher.ElasticSearch(db, img, nil)
	}

	matches := crossCheckMatches(dbMatches)

	hasFalsePositive := len(matches) > 1
	hasTruePositive := false
	for _, match := range matches {
		dist := math.Sqrt(math.Pow(match.Center.X-centerX, 2) +
			math.Pow(match.Center.Y-centerY, 2))
		if dist > maxDist {
			hasFalsePositive = true
		} else {
			hasTruePositive = true
		}
	}
	if hasFalsePositive && hasTruePositive {
		return BothPositive, nil
	} else if hasFalsePositive {
		return FalsePositive, nil
	} else if hasTruePositive {
		return TruePositive, nil
	} else {
		return Negative, nil
	}
}

func crossCheckMatches(matches [][]*mustacher.DatabaseMatch) []*mustacher.DatabaseMatch {
	res := make([]*mustacher.DatabaseMatch, len(matches[0]))
	copy(res, matches[0])

	for _, matchList := range matches {
	MatchLoop:
		for _, match := range matchList {
			for _, existingMatch := range res {
				minDistance := math.Max(match.Width/2, existingMatch.Width/2)
				if match.Center.Distance(existingMatch.Center) < minDistance {
					continue MatchLoop
				}
			}
			res = append(res, match)
		}
	}

	return res
}
