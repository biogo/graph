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
	"fmt"
	"math"
	"time"
)

// FIXME Use Index() instead of ID() on edges and nodes - this requires a change to node.go

// array approach

func RandMinCut(g *Undirected) (c []*Edge, w float64) {
	k := newKarger(g)
	w = float64(g.Size())
	var cw float64
	lim := int64(g.Order()) * int64(g.Order()) * int64(math.Log(float64(g.Order()))+1)
	fmt.Println(lim)
	last := time.Now()
	for i := int64(0); i < lim; i++ {
		fmt.Println(i, w, time.Now().Sub(last))
		last = time.Now()
		c, cw = k.randMinCut()
		if cw < w {
			w = cw
		}
	}

	return
}

type karger struct {
	g     *Undirected
	label []int
	sel   Selector
}

func newKarger(g *Undirected) *karger {
	return &karger{
		g:     g,
		label: make([]int, g.NextNodeID()),
		sel:   make(Selector, g.Size()),
	}
}

func (self *karger) init() {
	for i := range self.label {
		self.label[i] = -1
	}
	for _, n := range self.g.Nodes() {
		id := n.ID()
		self.label[id] = id
	}
	for i, e := range self.g.Edges() {
		self.sel[i] = WeightedItem{Index: e.ID(), Weight: e.Weight()}
	}
	self.sel.Init()
}

func (self *karger) randMinCut() (c []*Edge, w float64) {
	self.init()
	for k := self.g.Order(); k > 2; {
		id, err := self.sel.Select()
		if err != nil {
			break
		}

		e := self.g.Edge(id)
		if self.loop(e) {
			continue
		}

		j := self.label[e.Tail().ID()]
		for i, l := range self.label {
			if l != j {
				continue
			}
			self.label[i] = self.label[e.Head().ID()]
		}

		k--
	}
	for _, e := range self.g.Edges() {
		if self.loop(e) {
			continue
		}
		c = append(c, e)
		w += e.Weight()
	}

	return
}

func (self *karger) loop(e *Edge) bool {
	return self.label[e.Head().ID()] == self.label[e.Tail().ID()]
}
