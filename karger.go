// Copyright ©2012 The bíogo.graph Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package graph

import (
	"math"
	"runtime"
	"sync"
)

// FIXME Use Index() instead of ID() on edges and nodes - this requires a change to node.go

const sqrt2 = 1.4142135623730950488016887242096980785696718753769480

var MaxProcs = runtime.GOMAXPROCS(0)

func FastRandMinCut(g *Undirected, iter int) (c []Edge, w float64) {
	ka := newKarger(g)
	ka.init()
	w = math.Inf(1)
	for i := 0; i < iter; i++ {
		ka.fastRandMinCut()
		if ka.w < w {
			w = ka.w
			c = ka.c
		}
	}

	return
}

type karger struct {
	g     *Undirected
	order int
	ind   []super
	sel   Selector
	c     []Edge
	w     float64
}

type super struct {
	label int
	nodes []int
}

func newKarger(g *Undirected) *karger {
	return &karger{
		g:   g,
		ind: make([]super, g.NextNodeID()),
		sel: make(Selector, g.Size()),
	}
}

func (ka *karger) init() {
	ka.order = ka.g.Order()
	for i := range ka.ind {
		ka.ind[i].label = -1
		ka.ind[i].nodes = nil
	}
	for _, n := range ka.g.Nodes() {
		id := n.ID()
		ka.ind[id].label = id
	}
	for i, e := range ka.g.Edges() {
		ka.sel[i] = WeightedItem{Index: e.ID(), Weight: e.Weight()}
	}
	ka.sel.Init()
}

func (ka *karger) clone() (c *karger) {
	c = &karger{
		g:     ka.g,
		ind:   make([]super, ka.g.NextNodeID()),
		sel:   make(Selector, ka.g.Size()),
		order: ka.order,
	}

	copy(c.sel, ka.sel)
	for i, n := range ka.ind {
		s := &c.ind[i]
		s.label = n.label
		if n.nodes != nil {
			s.nodes = make([]int, len(n.nodes))
			copy(s.nodes, n.nodes)
		}
	}

	return
}

func (ka *karger) fastRandMinCut() {
	if ka.order <= 6 {
		ka.randCompact(2)
		return
	}

	t := int(math.Ceil(float64(ka.order)/sqrt2 + 1))

	sub := []*karger{ka, ka.clone()}
	for i := range sub {
		sub[i].randContract(t)
		sub[i].fastRandMinCut()
	}

	if sub[0].w < sub[1].w {
		*ka = *sub[0]
		return
	}
	*ka = *sub[1]
}

func (ka *karger) randContract(k int) {
	for ka.order > k {
		id, err := ka.sel.Select()
		if err != nil {
			break
		}

		e := ka.g.Edge(id)
		if ka.loop(e) {
			continue
		}

		hid, tid := e.Head().ID(), e.Tail().ID()
		hl, tl := ka.ind[hid].label, ka.ind[tid].label
		if len(ka.ind[hl].nodes) < len(ka.ind[tl].nodes) {
			hid, tid = tid, hid
			hl, tl = tl, hl
		}

		if ka.ind[hl].nodes == nil {
			ka.ind[hl].nodes = []int{hid}
		}
		if ka.ind[tl].nodes == nil {
			ka.ind[hl].nodes = append(ka.ind[hl].nodes, tid)
		} else {
			ka.ind[hl].nodes = append(ka.ind[hl].nodes, ka.ind[tl].nodes...)
			ka.ind[tl].nodes = nil
		}
		for _, i := range ka.ind[hl].nodes {
			ka.ind[i].label = ka.ind[hid].label
		}

		ka.order--
	}
}

func (ka *karger) randCompact(k int) {
	for ka.order > k {
		id, err := ka.sel.Select()
		if err != nil {
			break
		}

		e := ka.g.Edge(id)
		if ka.loop(e) {
			continue
		}

		hid, tid := e.Head().ID(), e.Tail().ID()
		hl, tl := ka.ind[hid].label, ka.ind[tid].label
		if len(ka.ind[hl].nodes) < len(ka.ind[tl].nodes) {
			hid, tid = tid, hid
			hl, tl = tl, hl
		}

		if ka.ind[hl].nodes == nil {
			ka.ind[hl].nodes = []int{hid}
		}
		if ka.ind[tl].nodes == nil {
			ka.ind[hl].nodes = append(ka.ind[hl].nodes, tid)
		} else {
			ka.ind[hl].nodes = append(ka.ind[hl].nodes, ka.ind[tl].nodes...)
			ka.ind[tl].nodes = nil
		}
		for _, i := range ka.ind[hl].nodes {
			ka.ind[i].label = ka.ind[hid].label
		}

		ka.order--
	}

	ka.c, ka.w = []Edge{}, 0
	for _, e := range ka.g.Edges() {
		if ka.loop(e) {
			continue
		}
		ka.c = append(ka.c, e)
		ka.w += e.Weight()
	}
}

func (ka *karger) loop(e Edge) bool {
	return ka.ind[e.Head().ID()].label == ka.ind[e.Tail().ID()].label
}

// parallelised within the recursion tree

func ParFastRandMinCut(g *Undirected, iter, threads int) (c []Edge, w float64) {
	k := newKargerP(g)
	k.split = threads
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

type kargerP struct {
	karger
	count int
	split int
}

func newKargerP(g *Undirected) *kargerP {
	return &kargerP{karger: karger{
		g:   g,
		ind: make([]super, g.NextNodeID()),
		sel: make(Selector, g.Size()),
	}}
}

func (ka *kargerP) clone() (c *kargerP) {
	c = &kargerP{karger: *ka.karger.clone()}
	c.count = ka.count

	return
}

func (ka *kargerP) fastRandMinCut() {
	if ka.order <= 6 {
		ka.randCompact(2)
		return
	}

	t := int(math.Ceil(float64(ka.order)/sqrt2 + 1))

	var wg *sync.WaitGroup
	if ka.count < ka.split {
		wg = &sync.WaitGroup{}
	}
	ka.count++

	sub := []*kargerP{ka, ka.clone()}
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
		*ka = *sub[0]
		return
	}
	*ka = *sub[1]
}
