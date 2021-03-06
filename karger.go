// Copyright ©2012 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package graph

import (
	"math"
	"sync"
)

// Worked from http://www.cs.tau.ac.il/~zwick/grad-algo-08/gmc.pdf, but the wiki page
// http://en.wikipedia.org/wiki/Karger%27s_algorithm#Karger.E2.80.93Stein_algorithm is very good.

// FIXME Use Index() instead of ID() on edges and nodes - this requires a change to node.go

const sqrt2 = 1.4142135623730950488016887242096980785696718753769480

func RandMinCut(g *Undirected, iter int) (c []Edge, w float64) {
	ka := newKarger(g)
	w = math.Inf(1)
	for i := 0; i < iter; i++ {
		ka.fastMinCut()
		if ka.w < w {
			w = ka.w
			c = ka.c
		}
	}

	return c, w
}

func (ka *karger) fastMinCut() {
	if ka.order <= 6 {
		ka.compact(2)
		return
	}

	t := int(math.Ceil(float64(ka.order)/sqrt2 + 1))

	sub := []*karger{ka, ka.clone()}
	for _, ks := range sub {
		ks.contract(t)
		ks.fastMinCut()
	}

	if sub[1].w < sub[0].w {
		*ka = *sub[1]
	}
}

// parallelised within the recursion tree

func RandMinCutPar(g *Undirected, iter, threads int) (c []Edge, w float64) {
	k := newKarger(g)
	k.split = threads
	if k.split == 0 {
		k.split = -1
	}
	w = math.Inf(1)
	for i := 0; i < iter; i++ {
		k.fastMinCutPar()
		if k.w < w {
			w = k.w
			c = k.c
		}
	}

	return c, w
}

func (ka *karger) fastMinCutPar() {
	if ka.order <= 6 {
		ka.compact(2)
		return
	}

	t := int(math.Ceil(float64(ka.order)/sqrt2 + 1))

	var wg *sync.WaitGroup
	if ka.count < ka.split {
		wg = &sync.WaitGroup{}
	}
	ka.count++

	sub := []*karger{ka, ka.clone()}
	for _, ks := range sub {
		if wg != nil {
			wg.Add(1)
			go func(ks *karger) {
				defer wg.Done()
				ks.contract(t)
				ks.fastMinCutPar()
			}(ks)
		} else {
			ks.contract(t)
			ks.fastMinCutPar()
		}
	}

	if wg != nil {
		wg.Wait()
	}

	if sub[1].w < sub[0].w {
		*ka = *sub[1]
	}
}

type karger struct {
	g     *Undirected
	order int
	ind   []super
	sel   Selector
	c     []Edge
	w     float64

	count int
	split int
}

type super struct {
	label int
	nodes []int
}

func newKarger(g *Undirected) *karger {
	ka := karger{
		g:     g,
		order: g.Order(),
		ind:   make([]super, g.NextNodeID()),
		sel:   make(Selector, g.Size()),
	}

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

	return &ka
}

func (ka *karger) clone() *karger {
	c := karger{
		g:     ka.g,
		ind:   make([]super, ka.g.NextNodeID()),
		sel:   make(Selector, ka.g.Size()),
		order: ka.order,
		count: ka.count,
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

	return &c
}

func (ka *karger) contract(k int) {
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

func (ka *karger) compact(k int) {
	ka.contract(k)
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
