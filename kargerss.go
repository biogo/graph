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
	"math"
)

// FIXME Use Index() instead of ID() on edges and nodes - this requires a change to node.go

func RandMinCutSS(g *Undirected, iter int64, wt float64) (c []*Edge, w float64) {
	k := newKargerSS(g)
	w = math.Inf(1)
	var cw float64
	for i := int64(0); i < iter; i++ {
		c, cw = k.randCut()
		if cw < w {
			w = cw
		}
		if w <= wt {
			break
		}
	}

	return
}

type kargerSS struct {
	g   *Undirected
	ind []super
	sel Selector
}

type super struct {
	label int
	nodes []int
}

func newKargerSS(g *Undirected) *kargerSS {
	return &kargerSS{
		g:   g,
		ind: make([]super, g.NextNodeID()),
		sel: make(Selector, g.Size()),
	}
}

func (self *kargerSS) init() {
	for i := range self.ind {
		self.ind[i].label = -1
		self.ind[i].nodes = nil
	}
	for _, n := range self.g.Nodes() {
		id := n.ID()
		self.ind[id].label = id
	}
	for i, e := range self.g.Edges() {
		self.sel[i] = WeightedItem{Index: e.ID(), Weight: e.Weight()}
	}
	self.sel.Init()
}

func (self *kargerSS) randCut() ([]*Edge, float64) {
	self.init()
	return self.randCompact()
}

func (self *kargerSS) randCompact() (c []*Edge, w float64) {
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

		hid, tid := e.Head().ID(), e.Tail().ID()
		hl, tl := self.ind[hid].label, self.ind[tid].label
		if len(self.ind[hl].nodes) < len(self.ind[tl].nodes) {
			hid, tid = tid, hid
			hl, tl = tl, hl
		}

		if self.ind[hl].nodes == nil {
			self.ind[hl].nodes = []int{hid}
		}
		if self.ind[tl].nodes == nil {
			self.ind[hl].nodes = append(self.ind[hl].nodes, tid)
		} else {
			self.ind[hl].nodes = append(self.ind[hl].nodes, self.ind[tl].nodes...)
			self.ind[tl].nodes = nil
		}
		for _, i := range self.ind[hl].nodes {
			self.ind[i].label = self.ind[hid].label
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

func (self *kargerSS) loop(e *Edge) bool {
	return self.ind[e.Head().ID()].label == self.ind[e.Tail().ID()].label
}
