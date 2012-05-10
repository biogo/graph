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
	"fmt"
)

// NodeFilter is a function type used for assessment of nodes during graph traversal.
type NodeFilter func(*Node) bool

// A Node is a node in a graph.
type Node struct {
	name  string
	id    int
	index int
	edges Edges
}

// newNode creates a new *Nodes with ID id. Nodes should only ever exist in the context of a
// graph, so this is not a public function.
func newNode(id int) *Node {
	return &Node{
		id: id,
	}
}

// A Hop is an edge/node pair where the edge leads to the node from a neighbor.
type Hop struct {
	Edge *Edge
	Node *Node
}

// Name returns the name of a node.
func (self *Node) Name() string {
	return self.name
}

// SetName sets the name of a node to n.
func (self *Node) SetName(n string) {
	self.name = n
}

// ID returns the id of a node.
func (self *Node) ID() int {
	return self.id
}

// Edges returns a slice of edges that are incident on the node.
func (self *Node) Edges() []*Edge {
	return self.edges
}

// Degree returns the number of incident edges on a node. Looped edges are counted at both ends.
func (self *Node) Degree() int {
	l := 0
	for _, e := range self.edges {
		if e.Head() == e.Tail() {
			l++
		}
	}
	return l + len(self.edges)
}

// Neighbors returns a slice of nodes that share an edge with the node. Multiply connected nodes are
// repeated in the slice. If the node is self-connected it will be included in the slice, potentially
// repeatedly if there are multiple self-connecting edges.
func (self *Node) Neighbors(ef EdgeFilter) (n []*Node) {
	for _, e := range self.edges {
		if ef(e) {
			if a := e.Tail(); a == self {
				n = append(n, e.Head())
			} else {
				n = append(n, a)
			}
		}
	}
	return
}

// Hops has essentially the same functionality as Neighbors with the exception that the connecting
// edge is also returned.
func (self *Node) Hops(ef EdgeFilter) (h []*Hop) {
	for _, e := range self.edges {
		if ef(e) {
			if a := e.Tail(); a == self {
				h = append(h, &Hop{e, e.Head()})
			} else {
				h = append(h, &Hop{e, a})
			}
		}
	}
	return
}

func (self *Node) add(e *Edge) { self.edges = append(self.edges, e) }

func (self *Node) dropAll() {
	for i := range self.edges {
		self.edges[i] = nil
	}
	self.edges = self.edges[:0]
}

func (self *Node) drop(e *Edge) {
	for i := 0; i < len(self.edges); {
		if self.edges[i] == e {
			self.edges = self.edges.delFromNode(i)
			return // assumes e has not been added more than once - this should not happen, but we don't check for it
		} else {
			i++
		}
	}
}

func (self *Node) String() string {
	return fmt.Sprintf("%d:%v", self.id, self.edges)
}

// Nodes is a collection of nodes used for internal representation of node lists in a graph.
type Nodes []*Node

func (self Nodes) delFromGraph(i int) Nodes {
	self[i], self[len(self)-1] = self[len(self)-1], self[i]
	self[i].index = i
	self[len(self)-1].index = -1
	return self[:len(self)-1]
}
