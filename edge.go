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

var alreadyConnected = errors.New("graph: edge already fully connected")

// EdgeFilter is a function type used for assessment of edges during graph traversal. 
type EdgeFilter func(*Edge) bool

// EdgeFlags is a type that can be used to arbitrarily alter the behavior of edges.
type EdgeFlags uint32

const (
	EdgeCut EdgeFlags = 1 << iota // Set and use this flag to prevent traversal of temporarily cut edges.
)

// An Edge is an edge in a graph.
type Edge struct {
	name   string
	id     int
	i      int
	u, v   *Node
	weight float64
	flags  EdgeFlags
}

// newEdge returns a new edge. edges have no meaning outside the context of a
// graph, so this is private.
func newEdge(id, i int, u, v *Node, w float64, f EdgeFlags) *Edge {
	return &Edge{id: id, i: i, u: u, v: v, weight: w, flags: f}
}

// Name returns the name of a node.
func (self *Edge) Name() string {
	return self.name
}

// SetName sets the name of a node to n.
func (self *Edge) SetName(n string) {
	self.name = n
}

// ID returns the id of the edge.
func (self *Edge) ID() int {
	return self.id
}

// Index returns the index of the edge in the compact edge list of the graph. The value returned
// cannot be reliably used after an edge deletion.
func (self *Edge) Index() int {
	return self.i
}

// Nodes returns the two nodes, u and v, that are joined by the edge.
func (self *Edge) Nodes() (u, v *Node) {
	return self.u, self.v
}

// Head returns the first node of an edge's node pair.
func (self *Edge) Head() (v *Node) {
	return self.v
}

// Tail returns the second node of an edge's node pair.
func (self *Edge) Tail() (u *Node) {
	return self.u
}

// Weight returns the weight of the edge.
func (self *Edge) Weight() (w float64) {
	return self.weight
}

// SetWeight sets the weight of the edge to w.
func (self *Edge) SetWeight(w float64) {
	self.weight = w
}

// Flags returns the flags value for the edge. One flag is currently defined, EdgeCut.
func (self *Edge) Flags() EdgeFlags {
	return self.flags
}

// SetFlags sets the flags of the edge. One flag is currently defined, EdgeCut.
func (self *Edge) SetFlags(f EdgeFlags) {
	self.flags = f
}

func (self *Edge) reconnect(u, v *Node) {
	switch u {
	case self.u:
		self.u = v
	case self.v:
		self.v = v
	}
}

func (self *Edge) disconnect(n *Node) {
	switch n {
	case self.u:
		self.u.drop(self)
		self.u = nil
	case self.v:
		self.v.drop(self)
		self.v = nil
	}
}

func (self *Edge) connect(n *Node) (err error) {
	switch (*Node)(nil) {
	case self.u:
		self.u = n
		self.u.add(self)
	case self.v:
		self.v = n
		self.v.add(self)
	default:
		err = alreadyConnected
	}

	return
}

func (self *Edge) String() string {
	return fmt.Sprintf("%d--%d", self.u.ID(), self.v.ID())
}

// Edges is a collection of edges used for internal representation of edge lists in a graph. 
type Edges []*Edge

func (self Edges) delFromGraph(i int) Edges {
	self[i], self[len(self)-1] = self[len(self)-1], self[i]
	self[i].i = i
	self[len(self)-1].i = -1
	return self[:len(self)-1]
}

func (self Edges) delFromNode(i int) Edges {
	self[i], self[len(self)-1] = self[len(self)-1], self[i]
	return self[:len(self)-1]
}
