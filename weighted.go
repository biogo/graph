package graph

// Copyright ©2012 Dan Kortschak <dan.kortschak@adelaide.edu.au>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

import (
	"math/rand"
)

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
func (self Selector) Init() {
	for i := range self {
		self[i].total = self[i].Weight
	}
	for i := len(self) - 1; i > 0; i-- {
		// sometimes 1-based counting makes sense
		self[(i+1)>>1-1].total += self[i].total
	}
}

// Select returns the value of the Index field of the chosen WeightedItem and the item is weighted 
// zero to prevent further selection.
func (self Selector) Select() (index int) {
	s, i := self[0].total*rand.Float64(), 1

	for {
		if s -= self[i-1].Weight; s <= 0 {
			break // fall within item i-1
		}
		i <<= 1 // move to left child
		if d := self[i-1].total; s > d {
			s -= d
			// if enough s to pass left child
			// move to right child
			// state will be caught at break above
			i++
		}
	}

	w, index := self[i-1].Weight, self[i-1].Index

	self[i-1].Weight = 0
	for i > 0 {
		self[i-1].total -= w
		i >>= 1
	}

	return
}

// Weight alters the weight of item i in the Selector.
func (self Selector) Weight(i int, w float64) {
	w, self[i].Weight = self[i].Weight-w, w
	i++
	for i > 0 {
		self[i-1].total -= w
		i >>= 1
	}
}
