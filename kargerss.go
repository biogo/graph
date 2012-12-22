// Copyright ©2012 The bíogo.graph Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package graph

import (
	"math"
)

// FIXME Use Index() instead of ID() on edges and nodes - this requires a change to node.go

func RandMinCutSS(g *Undirected, iter int64, wt float64) (c []Edge, w float64) {
	ka := newKargerSS(g)
	w = math.Inf(1)
	var cw float64
	for i := int64(0); i < iter; i++ {
		c, cw = ka.randCut()
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

func (ka *kargerSS) init() {
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

func (ka *kargerSS) randCut() ([]Edge, float64) {
	ka.init()
	return ka.randCompact()
}

func (ka *kargerSS) randCompact() (c []Edge, w float64) {
	ka.init()
	for k := ka.g.Order(); k > 2; {
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

		k--
	}

	for _, e := range ka.g.Edges() {
		if ka.loop(e) {
			continue
		}
		c = append(c, e)
		w += e.Weight()
	}

	return
}

func (ka *kargerSS) loop(e Edge) bool {
	return ka.ind[e.Head().ID()].label == ka.ind[e.Tail().ID()].label
}
