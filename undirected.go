// Copyright ©2012 The bíogo.graph Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package graph

import (
	"errors"
	"fmt"
)

var (
	NodeExists       = errors.New("graph: node exists")
	NodeDoesNotExist = errors.New("graph: node does not exist")
	NodeIDOutOfRange = errors.New("graph: node id out of range")
	EdgeDoesNotExist = errors.New("graph: edge does not exist")
)

// An Undirected is a container for an undirected graph representation.
type Undirected struct {
	nodes, compNodes Nodes
	edges, compEdges Edges
}

// NewUndirected creates a new empty Undirected graph.
func NewUndirected() *Undirected {
	return &Undirected{
		nodes:     Nodes{},
		compNodes: Nodes{},
		edges:     Edges{},
		compEdges: Edges{},
	}
}

// NextNodeID returns the next unused available node ID. Unused IDs may be available for nodes with
// ID in [0, NextNodeID()) from deletion of nodes.
func (g *Undirected) NextNodeID() int {
	return len(g.nodes)
}

func (g *Undirected) NewNode() Node {
	return &node{id: len(g.nodes)}
}

// NextEdgeID returns the next unused available edge ID.
func (g *Undirected) NextEdgeID() int {
	return len(g.edges)
}

// Order returns the number of nodes in the graph.
func (g *Undirected) Order() int {
	return len(g.compNodes)
}

// Size returns the number of edges in the graph.
func (g *Undirected) Size() int {
	return len(g.compEdges)
}

// Nodes returns the complete set of nodes in the graph.
func (g *Undirected) Nodes() Nodes {
	return g.compNodes
}

// Node returns the node with the specified ID.
func (g *Undirected) Node(id int) Node {
	if id >= len(g.nodes) {
		return nil
	}
	return g.nodes[id]
}

// Edges returns the complete set of edges in the graph.
func (g *Undirected) Edges() []Edge {
	return g.compEdges
}

// Edge returns the edge with the specified ID.
func (g *Undirected) Edge(id int) Edge {
	if id >= len(g.edges) {
		return nil
	}
	return g.edges[id]
}

// Node methods

// Add adds a node n to the graph. If a node with already exists in the graph with the same id
// an error NodeExists is returned.
func (g *Undirected) Add(n Node) error {
	id := n.ID()
	if ok, _ := g.HasNodeID(id); ok {
		return NodeExists
	}

	g.addNode(n, id)
	n.setIndex(len(g.compNodes))
	g.compNodes = append(g.compNodes, n)

	return nil
}

// AddID adds a node with a specified ID. If a node with this ID already exists,
// it is returned with an error NodeExists.
func (g *Undirected) AddID(id int) (Node, error) {
	if ok, _ := g.HasNodeID(id); ok {
		return g.Node(id), NodeExists
	}

	n := newNode(id)
	g.addNode(n, id)
	n.setIndex(len(g.compNodes))
	g.compNodes = append(g.compNodes, n)

	return n, nil
}

func (g *Undirected) addNode(n Node, id int) {
	switch {
	case id == len(g.nodes):
		g.nodes = append(g.nodes, n)
	case id > len(g.nodes):
		ns := make(Nodes, id+1)
		copy(ns, g.nodes)
		g.nodes = ns
		g.nodes[id] = n
	default:
		g.nodes[id] = n
	}
}

// DeleteByID deletes the node with the specified from the graph. If the specified node does not exist
// an error, NodeDoesNotExist is returned.
func (g *Undirected) DeleteByID(id int) error {
	ok, _ := g.HasNodeID(id)
	if !ok {
		return NodeDoesNotExist
	}
	g.deleteNode(id)

	return nil
}

// Delete deletes the node n from the graph. If the specified node does not exist an error,
// NodeDoesNotExist is returned.
func (g *Undirected) Delete(n Node) error {
	return g.DeleteByID(n.ID())
}

func (g *Undirected) deleteNode(id int) {
	n := g.nodes[id]
	g.nodes[n.ID()] = nil
	f := func(_ Edge) bool { return true }
	for _, h := range n.Hops(f) {
		h.Edge.disconnect(h.Node)
		g.compEdges = g.compEdges.delFromGraph(h.Edge.index())
	}
	g.compNodes = g.compNodes.delFromGraph(n.index())
	n.setID(-1)
}

// Has returns a boolean indicating whether the node n exists in the graph. If the ID of n is no in
// [0, NextNodeID()) an error, NodeIDOutOfRange is returned.
func (g *Undirected) Has(n Node) (bool, error) {
	return g.HasNodeID(n.ID())
}

// HasNodeID returns a boolean indicating whether a node with ID is exists in the graph. If ID is no in
// [0, NextNodeID()) an error, NodeIDOutOfRange is returned.
func (g *Undirected) HasNodeID(id int) (bool, error) {
	if id < 0 || id > len(g.nodes)-1 {
		return false, NodeIDOutOfRange
	}
	return g.nodes[id] != nil, nil
}

// Neighbours returns a slice of nodes that are reachable from the node n via edges that satisfy
// the criteria specified by the edge filter ef. If the node does not exist, an error NodeDoesNotExist
// or NodeIDOutOfRange is returned.
func (g *Undirected) Neighbors(n Node, ef EdgeFilter) ([]Node, error) {
	ok, err := g.Has(n)
	if !ok {
		if err == nil {
			err = NodeDoesNotExist
		}
		return nil, err
	}
	return n.Neighbors(ef), nil
}

// Merge merges the node src into the node dst, transfering all the edges of src to dst.
// The node src is then deleted. If either src or dst do not exist in the graph,
// an appropriate error is returned.
func (g *Undirected) Merge(dst, src Node) error {
	var (
		ok  bool
		err error
	)
	ok, err = g.Has(dst)
	if !ok {
		return err
	}
	ok, err = g.Has(src)
	if !ok {
		return err
	}

	for _, e := range src.Edges() {
		e.reconnect(src, dst)
		if e.Head() != e.Tail() {
			dst.add(e)
		}
	}

	src.dropAll()
	g.deleteNode(src.ID())

	return nil
}

// Edge methods

// newEdge makes a new edge joining u and v with weight w and edge flags f. The ID chosen for the
// edge is NextEdgeID().
func (g *Undirected) newEdge(u, v Node, w float64, f EdgeFlags) Edge {
	e := newEdge(len(g.edges), len(g.compEdges), u, v, w, f)
	g.edges = append(g.edges, e)
	g.compEdges = append(g.compEdges, e)

	return e
}

// newEdgeKeepID makes a new edge joining u and v with ID id, weight w and edge flags f.
func (g *Undirected) newEdgeKeepID(id int, u, v Node, w float64, f EdgeFlags) Edge {
	if id < len(g.edges) && g.edges[id] != nil {
		panic("graph: attempted to create a new edge with an existing ID")
	}
	e := newEdge(id, len(g.compEdges), u, v, w, f)

	switch {
	case id == len(g.edges):
		g.edges = append(g.edges, e)
	case id > len(g.edges):
		es := make(Edges, id+1)
		copy(es, g.edges)
		g.edges = es
		g.edges[id] = e
	default:
		g.edges[id] = e
	}
	e.setIndex(len(g.compEdges))
	g.compEdges = append(g.compEdges, e)

	return e
}

// ConnectWith join nodes u and v with the provided edge. An error is returned if
// either of the nodes does not exist.
func (g *Undirected) ConnectWith(u, v Node, with Edge) error {
	var (
		ok  bool
		err error
	)
	ok, err = g.Has(u)
	if !ok {
		return err
	}
	ok, err = g.Has(v)
	if !ok {
		return err
	}

	e := with
	e.setID(len(g.edges))
	e.setIndex(len(g.compEdges))
	e.join(u, v)

	g.edges = append(g.edges, e)
	g.compEdges = append(g.compEdges, e)

	u.add(e)
	if v != u {
		v.add(e)
	}

	return nil
}

// Connect creates a new edge joining nodes u and v with weight w, and specifying edge flags f.
// The new edge is returned on success. An error is returned if either of the nodes does not
// exist.
func (g *Undirected) Connect(u, v Node, w float64, f EdgeFlags) (Edge, error) {
	var (
		ok  bool
		err error
	)
	ok, err = g.Has(u)
	if !ok {
		return nil, err
	}
	ok, err = g.Has(v)
	if !ok {
		return nil, err
	}

	e := g.newEdge(u, v, w, f)
	u.add(e)
	if v != u {
		v.add(e)
	}

	return e, nil
}

// Connect creates a new edge joining nodes with IDs uid and vid with weight w, and specifying edge
// flags f. The id of the new edge is returned on success. An error is returned if either of the
// nodes does not exist.
func (g *Undirected) ConnectByID(uid, vid int, w float64, f EdgeFlags) (int, error) {
	var (
		ok  bool
		err error
	)
	ok, err = g.HasNodeID(uid)
	if !ok {
		return -1, err
	}
	ok, err = g.HasNodeID(vid)
	if !ok {
		return -1, err
	}

	e := g.newEdge(g.nodes[uid], g.nodes[vid], w, f)
	g.nodes[uid].add(e)
	if vid != uid {
		g.nodes[vid].add(e)
	}

	return e.ID(), nil
}

// Connected returns a boolean indicating whether the nodes u and v share an edge. An error is returned
// if either of the nodes does not exist.
func (g *Undirected) Connected(u, v Node) (bool, error) {
	var (
		ok  bool
		err error
	)
	ok, err = g.Has(u)
	if !ok {
		return false, err
	}
	ok, err = g.Has(v)
	if !ok {
		return false, err
	}

	if u == v {
		return true, nil
	}

	uedges, vedges := u.Edges(), v.Edges()
	if len(uedges) > len(vedges) {
		uedges, v = vedges, u
	}

	for _, e := range uedges {
		if a, b := e.Nodes(); a == v || b == v {
			return true, nil
		}
	}

	return false, nil
}

// ConnectingEdges returns a slice of edges that are shared by nodes u and v. An error is returned
// if either of the nodes does not exist.
func (g *Undirected) ConnectingEdges(u, v Node) ([]Edge, error) {
	var (
		ok  bool
		err error
	)
	ok, err = g.Has(u)
	if !ok {
		return nil, err
	}
	ok, err = g.Has(v)
	if !ok {
		return nil, err
	}

	var c []Edge
	uedges := u.Edges()
	if u == v {
		for _, e := range uedges {
			if a, b := e.Nodes(); a == b {
				c = append(c, e)
			}
		}

		return c, nil
	}

	vedges := v.Edges()
	if len(uedges) > len(vedges) {
		uedges, v = vedges, u
	}

	for _, e := range uedges {
		if a, b := e.Nodes(); a == v || b == v {
			c = append(c, e)
		}
	}

	return c, nil
}

// DeleteEdge deleted the edge e from the graph. An error is returned if the edge does not exist in
// the graph.
func (g *Undirected) DeleteEdge(e Edge) error {
	i := e.index()
	if i < 0 || i > len(g.compEdges)-1 {
		return EdgeDoesNotExist
	}

	e.disconnect(e.Head())
	e.disconnect(e.Tail())
	g.compEdges = g.compEdges.delFromGraph(i)
	g.edges[e.ID()] = nil
	e.setID(-1)

	return nil
}

// Structure methods

// ConnectedComponents returns a slice of slices of nodes. Each top level slice is the set of nodes
// composing a connected component of the graph. Connection is determined by traversal of edges that
// satisfy the edge filter ef.
func ConnectedComponents(g *Undirected, ef EdgeFilter) []Nodes {
	var cc []Nodes
	df := NewDepthFirst()
	c := []Node{}
	f := func(n Node) bool {
		c = append(c, n)
		return false
	}
	for _, s := range g.Nodes() {
		if df.Visited(s) {
			continue
		}
		df.Search(s, ef, f, nil)
		cc = append(cc, []Node{})
		cc[len(cc)-1] = append(cc[len(cc)-1], c...)
		c = c[:0]
	}

	return cc
}

func (g *Undirected) String() string {
	return fmt.Sprintf("G:|V|=%d |E|=%d", g.Order(), g.Size())
}
