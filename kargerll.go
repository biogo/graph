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
	"fmt"
	"time"
)

// FIXME Use Index() instead of ID() on edges and nodes - this requires a change to node.go

// linked list approach

func RandMinCutLL(g *Undirected) (c []*Edge, w float64) {
	k := newKargerLL(g)
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

type kargerLL struct {
	g   *Undirected
	ind []superLL
	sel Selector
}

type superLL struct {
	label int
	nodes *kargerNode
	end   *kargerNode
}

type kargerNode struct {
	next *kargerNode
	id   int
}

func newKargerLL(g *Undirected) *kargerLL {
	return &kargerLL{
		g:   g,
		ind: make([]superLL, g.NextNodeID()),
		sel: make(Selector, g.Size()),
	}
}

func (self *kargerLL) init() {
	for i := range self.ind {
		self.ind[i].label = -1
		for n := self.ind[i].nodes; n != nil; n, n.next, n.id = n.next, nil, 0 {
		}
		self.ind[i].nodes, self.ind[i].end = nil, nil
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

func (self *kargerLL) randMinCut() (c []*Edge, w float64) {
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

		if self.ind[hl].nodes == nil {
			self.ind[hl].nodes = &kargerNode{id: hid}
			self.ind[hl].end = self.ind[hl].nodes
		}
		if self.ind[tl].nodes == nil {
			self.ind[hl].nodes = &kargerNode{id: tid, next: self.ind[hl].nodes}
		} else {
			self.ind[tl].end.next = self.ind[hl].nodes
			self.ind[hl].nodes = self.ind[tl].nodes
			self.ind[tl].nodes, self.ind[tl].end = nil, nil
		}
		for n := self.ind[hl].nodes; n != nil; n = n.next {
			self.ind[n.id].label = self.ind[hid].label
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

func (self *kargerLL) loop(e *Edge) bool {
	return self.ind[e.Head().ID()].label == self.ind[e.Tail().ID()].label
}
