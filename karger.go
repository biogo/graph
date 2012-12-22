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

// array approach

func RandMinCut(g *Undirected) (c []Edge, w float64) {
	ka := newKarger(g)
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

func (ka *karger) init() {
	for i := range ka.label {
		ka.label[i] = -1
	}
	for _, n := range ka.g.Nodes() {
		id := n.ID()
		ka.label[id] = id
	}
	for i, e := range ka.g.Edges() {
		ka.sel[i] = WeightedItem{Index: e.ID(), Weight: e.Weight()}
	}
	ka.sel.Init()
}

func (ka *karger) randMinCut() (c []Edge, w float64) {
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

		j := ka.label[e.Tail().ID()]
		for i, l := range ka.label {
			if l != j {
				continue
			}
			ka.label[i] = ka.label[e.Head().ID()]
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

func (ka *karger) loop(e Edge) bool {
	return ka.label[e.Head().ID()] == ka.label[e.Tail().ID()]
}
