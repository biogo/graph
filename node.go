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

package graph

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
func (n *Node) Name() string {
	return n.name
}

// SetName sets the name of a node.
func (n *Node) SetName(name string) {
	n.name = name
}

// ID returns the id of a node.
func (n *Node) ID() int {
	return n.id
}

// Edges returns a slice of edges that are incident on the node.
func (n *Node) Edges() []*Edge {
	return n.edges
}

// Degree returns the number of incident edges on a node. Looped edges are counted at both ends.
func (n *Node) Degree() int {
	l := 0
	for _, e := range n.edges {
		if e.Head() == e.Tail() {
			l++
		}
	}
	return l + len(n.edges)
}

// Neighbors returns a slice of nodes that share an edge with the node. Multiply connected nodes are
// repeated in the slice. If the node is n-connected it will be included in the slice, potentially
// repeatedly if there are multiple n-connecting edges.
func (n *Node) Neighbors(ef EdgeFilter) []*Node {
	var nodes []*Node
	for _, e := range n.edges {
		if ef(e) {
			if a := e.Tail(); a == n {
				nodes = append(nodes, e.Head())
			} else {
				nodes = append(nodes, a)
			}
		}
	}
	return nodes
}

// Hops has essentially the same functionality as Neighbors with the exception that the connecting
// edge is also returned.
func (n *Node) Hops(ef EdgeFilter) []*Hop {
	var h []*Hop
	for _, e := range n.edges {
		if ef(e) {
			if a := e.Tail(); a == n {
				h = append(h, &Hop{e, e.Head()})
			} else {
				h = append(h, &Hop{e, a})
			}
		}
	}
	return h
}

func (n *Node) add(e *Edge) { n.edges = append(n.edges, e) }

func (n *Node) dropAll() {
	for i := range n.edges {
		n.edges[i] = nil
	}
	n.edges = n.edges[:0]
}

func (n *Node) drop(e *Edge) {
	for i := 0; i < len(n.edges); {
		if n.edges[i] == e {
			n.edges = n.edges.delFromNode(i)
			return // assumes e has not been added more than once - this should not happen, but we don't check for it
		} else {
			i++
		}
	}
}

func (n *Node) String() string {
	return fmt.Sprintf("%d:%v", n.id, n.edges)
}

// Nodes is a collection of nodes used for internal representation of node lists in a graph.
type Nodes []*Node

func (n Nodes) delFromGraph(i int) Nodes {
	n[i], n[len(n)-1] = n[len(n)-1], n[i]
	n[i].index = i
	n[len(n)-1].index = -1
	return n[:len(n)-1]
}
