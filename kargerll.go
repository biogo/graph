// Copyright ©2012 The bíogo.graph Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package graph

import (
	"fmt"
	"math"
	"time"
)

// FIXME Use Index() instead of ID() on edges and nodes - this requires a change to node.go

// linked list approach

func RandMinCutLL(g *Undirected) (c []Edge, w float64) {
	ka := newKargerLL(g)
	w = float64(g.Size())
	var cw float64
	lim := int64(g.Order()) * int64(g.Order()) * int64(math.Log(float64(g.Order()))+1)
	fmt.Println(lim)
	last := time.Now()
	for i := int64(0); i < lim; i++ {
		fmt.Println(i, w, time.Now().Sub(last))
		last = time.Now()
		c, cw = ka.randMinCut()
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

func (ka *kargerLL) init() {
	for i := range ka.ind {
		ka.ind[i].label = -1
		for n := ka.ind[i].nodes; n != nil; n, n.next, n.id = n.next, nil, 0 {
		}
		ka.ind[i].nodes, ka.ind[i].end = nil, nil
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

func (ka *kargerLL) randMinCut() (c []Edge, w float64) {
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

		if ka.ind[hl].nodes == nil {
			ka.ind[hl].nodes = &kargerNode{id: hid}
			ka.ind[hl].end = ka.ind[hl].nodes
		}
		if ka.ind[tl].nodes == nil {
			ka.ind[hl].nodes = &kargerNode{id: tid, next: ka.ind[hl].nodes}
		} else {
			ka.ind[tl].end.next = ka.ind[hl].nodes
			ka.ind[hl].nodes = ka.ind[tl].nodes
			ka.ind[tl].nodes, ka.ind[tl].end = nil, nil
		}
		for n := ka.ind[hl].nodes; n != nil; n = n.next {
			ka.ind[n.id].label = ka.ind[hid].label
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

func (ka *kargerLL) loop(e Edge) bool {
	return ka.ind[e.Head().ID()].label == ka.ind[e.Tail().ID()].label
}
