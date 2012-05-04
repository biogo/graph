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

const sqrt2 = 1.4142135623730950488016887242096980785696718753769480

func FastRandMinCut(g *Undirected, iter int, wt float64) (c []*Edge, w float64) {
	k := newKargerR(g)
	k.init()
	w = math.Inf(1)
	var cw float64
	for i := 0; i < iter; i++ {
		c, cw = k.fastRandMinCut()
		if cw < w {
			w = cw
		}
		if w <= wt {
			break
		}
	}

	return
}

type kargerR struct {
	g     *Undirected
	order int
	ind   []super
	sel   Selector
}

func newKargerR(g *Undirected) *kargerR {
	return &kargerR{
		g:   g,
		ind: make([]super, g.NextNodeID()),
		sel: make(Selector, g.Size()),
	}
}

func (self *kargerR) init() {
	self.order = self.g.Order()
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

func (self *kargerR) copy(t *kargerR) {
	self.order = t.order
	copy(self.sel, t.sel)
	for i, n := range t.ind {
		s := &self.ind[i]
		s.label = n.label
		if n.nodes != nil {
			s.nodes = make([]int, len(n.nodes))
			copy(s.nodes, n.nodes)
		}
	}
}

func (self *kargerR) fastRandMinCut() (c []*Edge, w float64) {
	if self.order <= 6 {
		return self.randCompact(2)
	}

	t := int(math.Ceil(float64(self.order)/sqrt2 + 1))

	sub := make([]*kargerR, 2)
	ct := make([][]*Edge, 2)
	wt := make([]float64, 2)
	for i := range sub {
		sub[i] = newKargerR(self.g)
		sub[i].copy(self)
		sub[i].randContract(t)
		ct[i], wt[i] = sub[i].fastRandMinCut()
	}

	if wt[0] < wt[1] {
		*self = *sub[0]
		return ct[0], wt[0]
	}
	*self = *sub[1]
	return ct[1], wt[1]
}

func (self *kargerR) randContract(k int) {
	for self.order > k {
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

		self.order--
	}
}

func (self *kargerR) randCompact(k int) (c []*Edge, w float64) {
	for self.order > k {
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

		self.order--
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

func (self *kargerR) loop(e *Edge) bool {
	return self.ind[e.Head().ID()].label == self.ind[e.Tail().ID()].label
}
