package main

import (
	"fmt"

	"github.com/unixpickle/mustachemash/mustacher"
)

type templatePair struct {
	Template    *mustacher.Template
	BestMatch   *mustacher.Template
	Correlation float64
}

func (t templatePair) String() string {
	if t.BestMatch == nil {
		return t.Template.UserInfo + " - no match"
	} else {
		return fmt.Sprintf("%s - %s - %f (threshold %f)", t.Template.UserInfo,
			t.BestMatch.UserInfo, t.Correlation, t.Template.Threshold)
	}
}

type templatePairList []templatePair

func (t templatePairList) Len() int {
	return len(t)
}

func (t templatePairList) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

func (t templatePairList) Less(i, j int) bool {
	return t[i].Correlation > t[j].Correlation
}
