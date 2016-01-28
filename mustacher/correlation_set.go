package mustacher

import "sort"

// A Correlation holds information about an appearance of a
// template inside a larger image.
type Correlation struct {
	Template    *Template
	Correlation float64

	X int
	Y int
}

// A CorrelationSet is an easy-to-manipulate list of Correlations.
type CorrelationSet []*Correlation

// NonOverlappingSet returns a copy of this set in which overlapping
// correlations are omitted.
// When picking which correlations to keep, the best ones are prioritized
// over weaker correlations.
func (c CorrelationSet) NonOverlappingSet() CorrelationSet {
	sortedSet := make(CorrelationSet, len(c))
	copy(sortedSet, c)
	sort.Sort(sortedSet)

	res := make(CorrelationSet, 0, len(c))
	for i := len(sortedSet) - 1; i >= 0; i-- {
		correlation := sortedSet[i]
		if !res.Overlaps(correlation) {
			res = append(res, correlation)
		}
	}
	return res
}

// Overlaps returns true if the given correlation overlaps with a
// correlation in the set.
func (c CorrelationSet) Overlaps(corr *Correlation) bool {
	for _, aCorr := range c {
		if aCorr.X+aCorr.Template.image.Width() > corr.X &&
			corr.X+corr.Template.image.Width() > aCorr.X &&
			aCorr.Y+aCorr.Template.image.Height() > corr.Y &&
			corr.Y+corr.Template.image.Height() > aCorr.Y {
			return true
		}
	}
	return false
}

// Len returns len(c).
func (c CorrelationSet) Len() int {
	return len(c)
}

// Less returns true if the correlation at index i
// is less than the correlation at index j.
func (c CorrelationSet) Less(i, j int) bool {
	return c[i].Correlation < c[j].Correlation
}

// Swap swaps the correlations at index i and j.
func (c CorrelationSet) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}
