// Copyright ©2012 The bíogo.graph Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package graph

import (
	"fmt"
	check "launchpad.net/gocheck"
)

// Tests
var (
	uv = []e{
		{1, 4},
		{4, 7},
		{7, 1},
		{9, 7},
		{6, 9},
		{3, 6},
		{9, 3},
		{8, 6},
		{8, 5},
		{5, 2},
		{2, 8},
	}
	nodeEdges  = []int{1: 2, 2: 2, 3: 2, 4: 2, 5: 2, 6: 3, 7: 3, 8: 3, 9: 3}
	deleteNode = 9
	parts      = []int{1: 0, 4: 0, 7: 0, 3: 1, 6: 1, 8: 1, 2: 1, 5: 1}
	partSizes  = []int{3, 5}
)

func undirected(c *check.C, edges []e) (g *Undirected) {
	g = NewUndirected()
	for _, e := range edges {
		u, _ := g.AddID(e.u)
		v, _ := g.AddID(e.v)
		g.Connect(u, v)
	}

	return
}

func (s *S) TestUndirected(c *check.C) {
	g := undirected(c, uv)
	nodes := make(map[int]int)
	for _, n := range uv {
		nodes[n.u], nodes[n.v] = 1, 1
	}
	c.Check(g.Order(), check.Equals, len(nodes))
	c.Check(g.Size(), check.Equals, len(uv))
}

func (s *S) TestUndirectedMerge(c *check.C) {
	g := undirected(c, uv)
	order := g.Order()
	size := g.Size()
	err := g.Merge(g.Node(7), g.Node(9))
	if err != nil {
		c.Fatal(err)
	}
	conn, err := g.ConnectingEdges(g.Node(7), g.Node(7))
	if err != nil {
		c.Fatal(err)
	}
	c.Check(len(conn), check.Equals, 1)
	c.Check(fmt.Sprint(conn[0]), check.Equals, "7--7")
	c.Check(g.Order(), check.Equals, order-1)
	c.Check(g.Size(), check.Equals, size)
	c.Check(g.Node(7).Degree(), check.Equals, 6)
	c.Check(len(g.Node(7).Edges()), check.Equals, 5)

	err = g.Merge(g.Node(6), g.Node(3))
	if err != nil {
		c.Fatal(err)
	}
	conn, err = g.ConnectingEdges(g.Node(7), g.Node(6))
	if err != nil {
		c.Fatal(err)
	}
	c.Check(len(conn), check.Equals, 2)
}

func (s *S) TestUndirectedConnected(c *check.C) {
	g := undirected(c, uv)
	n := g.Nodes()
	conns := 0
	for i := 0; i < g.Order(); i++ {
		for j := 0; j < g.Order(); j++ {
			if ok, err := g.Connected(n[i], n[j]); ok {
				conns++
			} else if err != nil {
				c.Fatal(err)
			}
		}
	}
	c.Check(conns, check.Equals, 2*g.Size()+g.Order())
}

func (s *S) TestUndirectedConnectedComponent(c *check.C) {
	g := undirected(c, uv)
	c.Check(len(ConnectedComponents(g, nil)), check.Equals, 1)
	g.DeleteByID(deleteNode)
	nodes, edges := make(map[int]int), make(map[int]int)
	for _, n := range uv {
		nodes[n.u], nodes[n.v] = 1, 1
		edges[n.u]++
		edges[n.v]++
	}
	c.Check(g.Order(), check.Equals, len(nodes)-1)
	c.Check(g.Size(), check.Equals, len(uv)-edges[deleteNode])
	cc := ConnectedComponents(g, nil)
	c.Check(len(cc), check.Equals, 2)
	for i, p := range cc {
		c.Check(len(p), check.Equals, partSizes[i])
		for _, n := range p {
			c.Check(parts[n.ID()], check.Equals, i)
		}
		g0, err := p.BuildUndirected(true)
		if err != nil {
			c.Fatal(err)
		}
		c.Check(g0.Order(), check.Equals, partSizes[i])
		c.Check(g0.Size(), check.Equals, partSizes[i])
	}
}

func (s *S) TestUndirectedBuild(c *check.C) {
	g := undirected(c, uv)
	g0, err := g.Nodes().BuildUndirected(false)
	if err != nil {
		c.Fatal(err)
	}
	for _, n := range g.Nodes() {
		id := n.ID()
		c.Check(g0.Node(id).ID(), check.Equals, g.Node(id).ID())
	}
	for i := range g.Edges() {
		c.Check(g0.Edge(i).ID(), check.Equals, i)
		c.Check(g0.Edge(i).ID(), check.Equals, g.Edge(i).ID())
		c.Check(g0.Edge(i).Head().ID(), check.Equals, g.Edge(i).Head().ID())
		c.Check(g0.Edge(i).Tail().ID(), check.Equals, g.Edge(i).Tail().ID())
	}
}

func (s *S) TestUndirectedRepresentation(c *check.C) {
	g := undirected(c, uv)
	for i, e := range g.Edges() {
		c.Check(fmt.Sprint(e), check.Equals, fmt.Sprintf("%d--%d", uv[i].u, uv[i].v))
	}
	reps := make([]string, len(uv)+1)
	for _, n := range uv {
		if reps[n.u] == "" {
			c.Check(len(g.Node(n.u).Edges()), check.Equals, nodeEdges[n.u])
			reps[n.u] = fmt.Sprintf("%d:%v", n.u, g.Node(n.u).Edges())
		}
	}
	for _, n := range g.Nodes() {
		c.Check(fmt.Sprint(n), check.Equals, reps[n.ID()])
	}
}

func (s *S) TestDeleteEdge(c *check.C) {
	g := undirected(c, uv)
	e, err := g.ConnectingEdges(g.Node(7), g.Node(9))
	c.Assert(err, check.Equals, nil)
	c.Assert(len(e), check.Equals, 1)
	h := e[0].Head()
	t := e[0].Tail()
	c.Check(h.ID(), check.Equals, 7)
	c.Check(t.ID(), check.Equals, 9)
	g.DeleteEdge(e[0])
	for _, ne := range h.Edges() {
		c.Check(ne, check.Not(check.DeepEquals), e[0])
	}
	for _, ne := range t.Edges() {
		c.Check(ne, check.Not(check.DeepEquals), e[0])
	}
	c.Check(e[0].Head(), check.Equals, nil)
	c.Check(e[0].Tail(), check.Equals, nil)
	cc := ConnectedComponents(g, func(e Edge) bool {
		c.Check(e.Head(), check.Not(check.Equals), nil)
		c.Check(e.Tail(), check.Not(check.Equals), nil)
		return true
	})
	c.Check(len(cc), check.Equals, 2)
}

func (s *S) TestDeleteNode(c *check.C) {
	g := NewUndirected()
	cen := g.NewNode()
	g.Add(cen)
	branches := make([]Node, 20)
	edges := make([]Edge, 20)
	for i := range branches {
		b := g.NewNode()
		g.Add(b)
		edges[i], _ = g.Connect(b, cen)
		branches[i] = b
	}
	g.Delete(cen)
	for _, e := range edges {
		c.Check(e.Head(), check.Equals, nil)
		c.Check(e.Tail(), check.Equals, nil)
	}
	for _, b := range branches {
		c.Check(b.Edges(), check.DeepEquals, []Edge(nil))
	}
	c.Check(cen.ID(), check.Equals, -1)
	for _, e := range cen.Edges() {
		c.Check(e.Head(), check.Not(check.Equals), nil)
		c.Check(e.Tail(), check.Not(check.Equals), nil)
	}
	for _, n := range g.compNodes {
		for _, e := range n.Edges() {
			c.Check(e.Head(), check.Not(check.Equals), nil)
			c.Check(e.Tail(), check.Not(check.Equals), nil)
		}
	}
}
