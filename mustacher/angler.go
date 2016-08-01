package mustacher

import "github.com/unixpickle/haar"

// An AnglerNode is a node in a decision tree for deciding
// the rotation (i.e. slant) of a mouth+nose image.
type AnglerNode struct {
	// Classification is the final angle (in radians) if
	// this node is a leaf node.
	Classification float64

	// Feature is the feature on which to split this
	// branch.
	// If this is nil, then the node is a leaf node.
	Feature *haar.Feature

	// Cutoff is the feature value above which to take
	// the Greater branch.
	Cutoff float64

	LessEqual *AnglerNode
	Greater   *AnglerNode
}

// Classify follows the decision tree to decide the angle
// of the mouth in a mouth+nose image.
func (a *AnglerNode) Classify(img haar.IntegralImage) float64 {
	if a.Feature == nil {
		return a.Classification
	}
	if a.Feature.Value(img) > a.Cutoff {
		return a.Greater.Classify(img)
	} else {
		return a.LessEqual.Classify(img)
	}
}
