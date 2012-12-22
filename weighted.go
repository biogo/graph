// Copyright ©2012 The bíogo.graph Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package graph

import (
	"fmt"
	"math/rand"
)

// SelectorEmpty is returned when an attempt is made to select an item from a Selector with
// no remaining selectable items.
var SelectorEmpty = fmt.Errorf("graph: selector empty")

// A WeightedItem is a type that can be be selected from a population with a defined probability
// specified by the field Weight. Index is used as an index to the actual item in another slice.
type WeightedItem struct {
	Index         int
	Weight, total float64
}

// A Selector is a collection of weighted items that can be selected with weighted probabilities
// without replacement.
type Selector []WeightedItem

// Init must be called on a Selector before it is selected from. Init is idempotent.
func (s Selector) Init() {
	for i := range s {
		s[i].total = s[i].Weight
	}
	for i := len(s) - 1; i > 0; i-- {
		// sometimes 1-based counting makes sense
		s[(i+1)>>1-1].total += s[i].total
	}
}

// Select returns the value of the Index field of the chosen WeightedItem and the item is weighted 
// zero to prevent further selection.
func (s Selector) Select() (int, error) {
	if s[0].total == 0 {
		return -1, SelectorEmpty
	}
	r, i := s[0].total*rand.Float64(), 1

	for {
		if r -= s[i-1].Weight; r <= 0 {
			break // fall within item i-1
		}
		i <<= 1 // move to left child
		if d := s[i-1].total; r > d {
			r -= d
			// if enough r to pass left child
			// move to right child
			// state will be caught at break above
			i++
		}
	}

	w, index := s[i-1].Weight, s[i-1].Index

	s[i-1].Weight = 0
	for i > 0 {
		s[i-1].total -= w
		i >>= 1
	}

	return index, nil
}

// Weight alters the weight of item i in the Selector.
func (s Selector) Weight(i int, w float64) {
	w, s[i].Weight = s[i].Weight-w, w
	i++
	for i > 0 {
		s[i-1].total -= w
		i >>= 1
	}
}
