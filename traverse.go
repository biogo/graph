// Copyright ©2012 The bíogo.graph Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package graph

// Visit is a function type that is used by a BreadthFirst or DepthFirst to allow side-effects
// on visiting new nodes in a graph traversal.
type Visit func(u, v Node)

// BreadthFirst is a type that can perform a breadth-first search on a graph.
type BreadthFirst struct {
	q      *queue
	visits []bool
}

// NewBreadthFirst creates a new BreadthFirst searcher.
func NewBreadthFirst() *BreadthFirst {
	return &BreadthFirst{q: &queue{}}
}

// Search searches a graph starting from node s until the NodeFilter function nf returns a value of
// true, traversing edges in the graph that allow the Edgefilter function ef to return true. On success
// the terminating node, t is returned. If vo is not nil, it is called with the start and end nodes of an
// edge when the end node has not already been visited.
func (b *BreadthFirst) Search(s Node, ef EdgeFilter, nf NodeFilter, vo Visit) Node {
	b.q.Enqueue(s)
	b.visits = mark(s, b.visits)
	for b.q.Len() > 0 {
		t, err := b.q.Dequeue()
		if err != nil {
			panic(err)
		}
		if nf != nil && nf(t) {
			return t
		}
		for _, n := range t.Neighbors(ef) {
			if !b.Visited(n) {
				if vo != nil {
					vo(t, n)
				}
				b.visits = mark(n, b.visits)
				b.q.Enqueue(n)
			}
		}
	}

	return nil
}

// Visited returns whether the node n has been visited by the searcher.
func (b *BreadthFirst) Visited(n Node) bool {
	id := n.ID()
	if id < 0 || id >= len(b.visits) {
		return false
	}
	return b.visits[id]
}

// Reset clears the search queue and visited list.
func (b *BreadthFirst) Reset() {
	b.q.Clear()
	b.visits = b.visits[:0]
}

// DepthFirst is a type that can perform a depth-first search on a graph.
type DepthFirst struct {
	s      *stack
	visits []bool
}

// NewDepthFirst creates a new DepthFirst searcher.
func NewDepthFirst() *DepthFirst {
	return &DepthFirst{s: &stack{}}
}

// Search searches a graph starting from node s until the NodeFilter function nf returns a value of
// true, traversing edges in the graph that allow the Edgefilter function ef to return true. On success
// the terminating node, t is returned. If vo is not nil, it is called with the start and end nodes of an
// edge when the end node has not already been visited.
func (d *DepthFirst) Search(s Node, ef EdgeFilter, nf NodeFilter, vo Visit) Node {
	d.s.Push(s)
	d.visits = mark(s, d.visits)
	for d.s.Len() > 0 {
		t, err := d.s.Pop()
		if err != nil {
			panic(err)
		}
		if nf != nil && nf(t) {
			return t
		}
		for _, n := range t.Neighbors(ef) {
			if !d.Visited(n) {
				if vo != nil {
					vo(t, n)
				}
				d.visits = mark(n, d.visits)
				d.s.Push(n)
			}
		}
	}

	return nil
}

// Visited returns whether the node n has been visited by the searcher.
func (d *DepthFirst) Visited(n Node) bool {
	id := n.ID()
	if id < 0 || id >= len(d.visits) {
		return false
	}
	return d.visits[id]
}

// Reset clears the search stack and visited list.
func (d *DepthFirst) Reset() {
	d.s.Clear()
	d.visits = d.visits[:0]
}

func mark(n Node, v []bool) []bool {
	id := n.ID()
	switch {
	case id == len(v):
		v = append(v, true)
	case id > len(v):
		t := make([]bool, id+1)
		copy(t, v)
		v = t
		v[id] = true
	default:
		v[id] = true
	}
	return v
}
