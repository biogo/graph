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
	"runtime"
	"sync"
)

// FIXME Use Index() instead of ID() on edges and nodes - this requires a change to node.go

const sqrt2 = 1.4142135623730950488016887242096980785696718753769480

var MaxProcs = runtime.GOMAXPROCS(0)

func FastRandMinCut(g *Undirected, iter int) (c []*Edge, w float64) {
	k := newKargerR(g)
	k.init()
	w = math.Inf(1)
	for i := 0; i < iter; i++ {
		k.fastRandMinCut()
		if k.w < w {
			w = k.w
			c = k.c
		}
	}

	return
}

// parallelised outside the recursion tree

func FastRandMinCutPar(g *Undirected, iter, thread int) (c []*Edge, w float64) {
	if thread > MaxProcs {
		thread = MaxProcs
	}
	if thread > iter {
		thread = iter
	}
	iter, rem := iter/thread+1, iter%thread

	type r struct {
		c []*Edge
		w float64
	}
	rs := make([]*r, thread)

	wg := &sync.WaitGroup{}
	for j := 0; j < thread; j++ {
		if rem == 0 {
			iter--
		}
		if rem >= 0 {
			rem--
		}
		wg.Add(1)
		go func(j, iter int) {
			defer wg.Done()
			k := newKargerR(g)
			k.init()
			var (
				w = math.Inf(1)
				c []*Edge
			)
			for i := 0; i < iter; i++ {
				k.fastRandMinCut()
				if k.w < w {
					w = k.w
					c = k.c
				}
			}

			rs[j] = &r{c, w}
		}(j, iter)
	}

	w = math.Inf(1)
	wg.Wait()
	for _, subr := range rs {
		if subr.w < w {
			w = subr.w
			c = subr.c
		}
	}

	return
}

type kargerR struct {
	g     *Undirected
	order int
	ind   []super
	sel   Selector
	c     []*Edge
	w     float64
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

func (self *kargerR) fastRandMinCut() {
	if self.order <= 6 {
		self.randCompact(2)
		return
	}

	t := int(math.Ceil(float64(self.order)/sqrt2 + 1))

	sub := []*kargerR{self, newKargerR(self.g)}
	sub[1].copy(self)
	for i := range sub {
		sub[i].randContract(t)
		sub[i].fastRandMinCut()
	}

	if sub[0].w < sub[1].w {
		*self = *sub[0]
		return
	}
	*self = *sub[1]
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

func (self *kargerR) randCompact(k int) {
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

	self.c, self.w = []*Edge{}, 0
	for _, e := range self.g.Edges() {
		if self.loop(e) {
			continue
		}
		self.c = append(self.c, e)
		self.w += e.Weight()
	}
}

func (self *kargerR) loop(e *Edge) bool {
	return self.ind[e.Head().ID()].label == self.ind[e.Tail().ID()].label
}

// parallelised within the recursion tree

func ParFastRandMinCut(g *Undirected, iter, threads int) (c []*Edge, w float64) {
	k := newKargerRP(g)
	k.split = bits(threads)
	if k.split == 0 {
		k.split = -1
	}
	k.init()
	w = math.Inf(1)
	for i := 0; i < iter; i++ {
		k.fastRandMinCut()
		if k.w < w {
			w = k.w
			c = k.c
		}
	}

	return
}

type kargerRP struct {
	g     *Undirected
	order int
	ind   []super
	sel   Selector
	c     []*Edge
	w     float64
	depth int
	split int
}

func bits(i int) (b int) {
	for ; i > 1; i >>= 1 {
		b++
	}
	return
}

func newKargerRP(g *Undirected) *kargerRP {
	return &kargerRP{
		g:   g,
		ind: make([]super, g.NextNodeID()),
		sel: make(Selector, g.Size()),
	}
}

func (self *kargerRP) init() {
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

func (self *kargerRP) copy(t *kargerRP) {
	self.order = t.order
	self.depth = t.depth
	self.split = t.split
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

func (self *kargerRP) fastRandMinCut() {
	self.depth++
	if self.order <= 6 {
		self.randCompact(2)
		return
	}

	t := int(math.Ceil(float64(self.order)/sqrt2 + 1))

	var wg *sync.WaitGroup
	if self.depth < self.split {
		wg = &sync.WaitGroup{}
	}

	sub := []*kargerRP{self, newKargerRP(self.g)}
	sub[1].copy(self)
	for i := range sub {
		if wg != nil {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				sub[i].randContract(t)
				sub[i].fastRandMinCut()
			}(i)
		} else {
			sub[i].randContract(t)
			sub[i].fastRandMinCut()
		}
	}

	if wg != nil {
		wg.Wait()
	}

	if sub[0].w < sub[1].w {
		*self = *sub[0]
		return
	}
	*self = *sub[1]
}

func (self *kargerRP) randContract(k int) {
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

func (self *kargerRP) randCompact(k int) {
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

	self.c, self.w = []*Edge{}, 0
	for _, e := range self.g.Edges() {
		if self.loop(e) {
			continue
		}
		self.c = append(self.c, e)
		self.w += e.Weight()
	}
}

func (self *kargerRP) loop(e *Edge) bool {
	return self.ind[e.Head().ID()].label == self.ind[e.Tail().ID()].label
}
