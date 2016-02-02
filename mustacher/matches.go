package mustacher

import (
	"math"
	"sort"
)

type FloatCoordinates struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

func (f FloatCoordinates) Distance(f1 FloatCoordinates) float64 {
	return math.Sqrt(math.Pow(f1.X-f.X, 2) + math.Pow(f1.Y-f.Y, 2))
}

// A Match holds information about an appearance of a
// template inside a larger image.
type Match struct {
	Template *Template

	// Correlation is a numerical measurement of how closely
	// the template matched the larger image.
	Correlation float64

	// Center is the center position of the template's target
	// in the larger image's coordinates.
	Center FloatCoordinates

	// Width is the width of the template's target in the larger
	// image's coordinates.
	// This may vary from Template.TargetWidth if the Match was
	// found in a scaled version of the larger image.
	Width float64
}

// A MatchSet is an easy-to-manipulate list of Matches.
type MatchSet []*Match

// FilterNearMatches returns a copy of this set which
// has no matches that are close to one another.
//
// When picking which matches to keep, matches are
// prioritized using c.Less().
func (c MatchSet) FilterNearMatches() MatchSet {
	sortedSet := make(MatchSet, len(c))
	copy(sortedSet, c)
	sort.Sort(sortedSet)

	res := make(MatchSet, 0, len(c))
	for i := len(sortedSet) - 1; i >= 0; i-- {
		match := sortedSet[i]
		if !res.HasNearMatch(match) {
			res = append(res, match)
		}
	}
	return res
}

// HasNearMatch returns true if the given match is uncomfortably
// close to any other matches in the set.
func (c MatchSet) HasNearMatch(m *Match) bool {
	for _, match := range c {
		// This definition of "nearness" is rather arbitrary; it just ensures
		// that the targets will never be touching.
		minDistance := math.Max(match.Width/2, m.Width/2)
		if match.Center.Distance(m.Center) < minDistance {
			return true
		}
	}
	return false
}

// Len returns len(c).
func (c MatchSet) Len() int {
	return len(c)
}

// Less returns true if the correlation at index i
// is less than the correlation at index j.
func (c MatchSet) Less(i, j int) bool {
	return c[i].Correlation < c[j].Correlation
}

// Swap swaps the correlations at index i and j.
func (c MatchSet) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}
