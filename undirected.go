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
	"errors"
	"fmt"
)

var (
	NodeExists       = errors.New("graph: node exists")
	NodeDoesNotExist = errors.New("graph: node does not exist")
	NodeIDOutOfRange = errors.New("graph: node id out of range")
	EdgeDoesNotExist = errors.New("graph: edge does not exist")
)

// An Unidirected is a container for an undirected graph representation.
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

// BuildUndirected creates a new Undirected graph using nodes and edges specified by the
// set of nodes in the slice ns. If edges of nodes in ns connect to nodes not in ns, these extra nodes
// will be included in the resulting graph. If compact is set to true, edge IDs are chosen to minimize
// space consumption, but breaking edge ID consistency between the new graph and the original.
func BuildUndirected(ns []*Node, compact bool) (g *Undirected, err error) {
	seen := make(map[*Edge]struct{})
	g = NewUndirected()
	for _, n := range ns {
		for _, e := range n.Edges() {
			if _, ok := seen[e]; ok {
				continue
			}
			seen[e] = struct{}{}
			u, v := e.Nodes()
			uid, vid := u.ID(), v.ID()
			if uid < 0 || vid < 0 {
				return nil, NodeIDOutOfRange
			}
			g.Add(uid)
			g.Add(vid)
			var ne *Edge
			if compact {
				ne = g.newEdge(g.nodes[uid], g.nodes[vid], e.Weight(), e.Flags())
			} else {
				ne = g.newEdgeKeepID(e.ID(), g.nodes[uid], g.nodes[vid], e.Weight(), e.Flags())
			}
			g.nodes[uid].add(ne)
			g.nodes[vid].add(ne)
		}
	}

	return
}

// NextNodeID returns the next unused available node ID. Unused IDs may be available for nodes with
// ID in [0, NextNodeID()) from deletion of nodes.
func (self *Undirected) NextNodeID() int {
	return len(self.nodes)
}

// NextEdgeID returns the next unused available edge ID.
func (self *Undirected) NextEdgeID() int {
	return len(self.nodes)
}

// Order returns the number of nodes in the graph.
func (self *Undirected) Order() int {
	return len(self.compNodes)
}

// Size returns the number of edges in the graph.
func (self *Undirected) Size() int {
	return len(self.compEdges)
}

// Nodes returns the complete set of nodes in the graph.
func (self *Undirected) Nodes() []*Node {
	return self.compNodes
}

// Node returns the node with ID i.
func (self *Undirected) Node(i int) *Node {
	if i >= len(self.nodes) {
		return nil
	}
	return self.nodes[i]
}

// Edges returns the complete set of edges in the graph.
func (self *Undirected) Edges() []*Edge {
	return self.compEdges
}

// Edge returns the edge with ID i.
func (self *Undirected) Edge(i int) *Edge {
	if i >= len(self.edges) {
		return nil
	}
	return self.edges[i]
}

// Node methods

// Add adds a node with ID is to the graph and returns the node. If a node with that specified ID
// already exists, it is returned and an error NodeExists is also returned.
func (self *Undirected) Add(id int) (n *Node, err error) {
	if ok, _ := self.HasNodeID(id); ok {
		return self.Node(id), NodeExists
	}

	n = newNode(id)

	if id == len(self.nodes) {
		self.nodes = append(self.nodes, n)
	} else if id > len(self.nodes) {
		ns := make(Nodes, id+1)
		copy(ns, self.nodes)
		self.nodes = ns
		self.nodes[id] = n
	} else {
		self.nodes[id] = n
	}
	n.index = len(self.compNodes)
	self.compNodes = append(self.compNodes, n)

	return
}

// DeleteByID deletes the node with ID id from the graph. If the specified node does not exist
// an error, NodeDoesNotExist is returned.
func (self *Undirected) DeleteByID(id int) (err error) {
	ok, err := self.HasNodeID(id)
	if !ok {
		return NodeDoesNotExist
	}
	self.deleteNode(id)

	return
}

// Delete deletes the node n from the graph. If the specified node does not exist an error,
// NodeDoesNotExist is returned.
func (self *Undirected) Delete(n *Node) (err error) {
	ok, err := self.Has(n)
	if !ok {
		return NodeDoesNotExist
	}
	self.deleteNode(n.ID())

	return
}

func (self *Undirected) deleteNode(id int) {
	n := self.nodes[id]
	self.nodes[n.ID()] = nil
	f := func(_ *Edge) bool { return true }
	for _, h := range n.Hops(f) {
		h.Edge.disconnect(h.Node)
		self.compEdges = self.compEdges.delFromGraph(h.Edge.i)
	}
	self.compNodes = self.compNodes.delFromGraph(n.index)
	(*n) = Node{}
}

// Has returns a boolean indicating whether the node n exists in the graph. If the ID of n is no in
// [0, NextNodeID()) an error, NodeIDOutOfRange is returned.
func (self *Undirected) Has(n *Node) (ok bool, err error) {
	if id := n.ID(); id >= 0 && id < len(self.nodes) {
		return self.nodes[id] == n, nil
	}
	return false, NodeIDOutOfRange
}

// HasNodeID returns a boolean indicating whether a node with ID is exists in the graph. If ID is no in
// [0, NextNodeID()) an error, NodeIDOutOfRange is returned.
func (self *Undirected) HasNodeID(id int) (ok bool, err error) {
	if id < 0 || id > len(self.nodes)-1 {
		return false, NodeIDOutOfRange
	}
	return self.nodes[id] != nil, nil
}

// Neighbours returns a slice of nodes that are reachable from the node n via edges that satisfy
// the criteria specified by the edge filter ef. If the node does not exist, an error NodeDoesNotExist
// or NodeIDOutOfRange is returned.
func (self *Undirected) Neighbors(n *Node, ef EdgeFilter) (adj []*Node, err error) {
	ok, err := self.Has(n)
	if !ok {
		if err == nil {
			err = NodeDoesNotExist
		}
		return
	}
	return n.Neighbors(ef), nil
}

// Merge merges the node src into the node dst, transfering all the edges of src to dst.
// The node src is then deleted. If either src or dst do not exist in the graph,
// an appropriate error is returned.
func (self *Undirected) Merge(dst, src *Node) (err error) {
	var ok bool
	ok, err = self.Has(dst)
	if !ok {
		return
	}
	ok, err = self.Has(src)
	if !ok {
		return
	}

	for _, e := range src.Edges() {
		e.reconnect(src, dst)
		if e.Head() != e.Tail() {
			dst.add(e)
		}
	}

	src.dropAll()
	self.deleteNode(src.ID())

	return
}

// Edge methods

// newEdge makes a new edge joining u and v with weight w and edge flags f. The ID chosen for the
// edge is NextEdgeID().
func (self *Undirected) newEdge(u, v *Node, w float64, f EdgeFlags) (e *Edge) {
	e = newEdge(len(self.edges), len(self.compEdges), u, v, w, f)
	self.edges = append(self.edges, e)
	self.compEdges = append(self.compEdges, e)

	return
}

// newEdgeKeepID makes a new edge joining u and v with ID id, weight w and edge flags f.
func (self *Undirected) newEdgeKeepID(id int, u, v *Node, w float64, f EdgeFlags) (e *Edge) {
	if id < len(self.edges) && self.edges[id] != nil {
		panic("graph: attempted to create a new edge with an existing ID")
	}
	e = newEdge(id, len(self.compEdges), u, v, w, f)

	if id == len(self.edges) {
		self.edges = append(self.edges, e)
	} else if id > len(self.edges) {
		es := make(Edges, id+1)
		copy(es, self.edges)
		self.edges = es
		self.edges[id] = e
	} else {
		self.edges[id] = e
	}
	e.i = len(self.compEdges)
	self.compEdges = append(self.compEdges, e)

	return
}

// Connect creats a new edge joining nodes u and v with weight w, and specifying edge flags f.
// The id of the new edge is returned on success. An error is returned if either of the nodes does not
// exist.
func (self *Undirected) Connect(u, v *Node, w float64, f EdgeFlags) (id int, err error) {
	var ok bool
	ok, err = self.Has(u)
	if !ok {
		return -1, err
	}
	ok, err = self.Has(v)
	if !ok {
		return -1, err
	}

	e := self.newEdge(u, v, w, f)
	u.add(e)
	v.add(e)
	id = e.ID()

	return
}

// Connect creats a new edge joining nodes with IDs uid and vid with weight w, and specifying edge
// flags f. The id of the new edge is returned on success. An error is returned if either of the
// nodes does not exist.
func (self *Undirected) ConnectByID(uid, vid int, w float64, f EdgeFlags) (id int, err error) {
	var ok bool
	ok, err = self.HasNodeID(uid)
	if !ok {
		return -1, err
	}
	ok, err = self.HasNodeID(vid)
	if !ok {
		return -1, err
	}

	e := self.newEdge(self.nodes[uid], self.nodes[vid], w, f)
	self.nodes[uid].add(e)
	self.nodes[vid].add(e)
	id = e.ID()

	return
}

// Connected returns a boolean indicating whether the nodes u and v share an edge. An error is returned
// if either of the nodes does not exist.
func (self *Undirected) Connected(u, v *Node) (c bool, err error) {
	var ok bool
	ok, err = self.Has(u)
	if !ok {
		return false, err
	}
	ok, err = self.Has(v)
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

	return
}

// ConnectingEdges returns a slice of edges that are shared by nodes u and v. An error is returned
// if either of the nodes does not exist.
func (self *Undirected) ConnectingEdges(u, v *Node) (c []*Edge, err error) {
	var ok bool
	ok, err = self.Has(u)
	if !ok {
		return nil, err
	}
	ok, err = self.Has(v)
	if !ok {
		return nil, err
	}

	uedges := u.Edges()
	if u == v {
		for _, e := range uedges {
			if a, b := e.Nodes(); a == b {
				c = append(c, e)
			}
		}

		return
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

	return
}

// DeleteEdge deleted the edge e from the graph. An error is returned if the edge does not exist in
// the graph.
func (self *Undirected) DeleteEdge(e *Edge) (err error) {
	i := e.Index()
	if i < 0 || i > len(self.compEdges)-1 {
		return EdgeDoesNotExist
	}

	e.disconnect(e.Head())
	e.disconnect(e.Tail())
	self.compEdges = self.compEdges.delFromGraph(i)
	self.edges[e.ID()] = nil
	(*e) = Edge{}

	return
}

// Structure methods

// ConnectedComponents returns a slice of slices of nodes. Each top level slice is the set of nodes
// composing a connected component of the graph. Connection is determined by traversal of edges that
// satisfy the edge filter ef.
func (self *Undirected) ConnectedComponents(ef EdgeFilter) (cc [][]*Node) {
	df := NewDepthFirst()
	c := []*Node{}
	f := func(n *Node) bool {
		c = append(c, n)
		return false
	}
	for _, s := range self.compNodes {
		if df.Visited(s) {
			continue
		}
		df.Search(s, ef, f)
		cc = append(cc, []*Node{})
		cc[len(cc)-1] = append(cc[len(cc)-1], c...)
		c = c[:0]
	}

	return
}

func (self *Undirected) String() string {
	return fmt.Sprintf("G:|V|=%d |E|=%d", self.Order(), self.Size())
}
