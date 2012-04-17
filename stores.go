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
)

var (
	queueIndexOutOfRange = errors.New("graph: queue index out of range")
	emptyQueue           = errors.New("graph: queue empty")
	stackIndexOutOfRange = errors.New("graph: stack index out of range")
	emptyStack           = errors.New("graph: stack empty")
)

type queue struct {
	head int
	data []*Node
}

func (self *queue) Enqueue(n *Node) {
	if len(self.data) == cap(self.data) && self.head > 0 {
		l := self.Len()
		copy(self.data, self.data[self.head:])
		self.head = 0
		self.data = append(self.data[:l], n)
	} else {
		self.data = append(self.data, n)
	}
}

func (self *queue) Dequeue() (n *Node, err error) {
	if self.Len() == 0 {
		return nil, emptyQueue
	}

	n, self.data[self.head] = self.data[self.head], nil
	self.head++

	if self.Len() == 0 {
		self.head = 0
		self.data = self.data[:0]
	}

	return
}

func (self *queue) Peek(i int) (n *Node, err error) {
	if i < self.head || i >= len(self.data) {
		return nil, queueIndexOutOfRange
	}
	return self.data[i+self.head], nil
}

func (self *queue) Clear() {
	self.head = 0
	self.data = self.data[:0]
}

func (self *queue) Len() int { return len(self.data) - self.head }

type stack struct {
	data []*Node
}

func (self *stack) Push(n *Node) { self.data = append(self.data, n) }

func (self *stack) Pop() (n *Node, err error) {
	if len(self.data) == 0 {
		return nil, emptyStack
	}

	n, self.data, self.data[len(self.data)-1] = self.data[len(self.data)-1], self.data[:len(self.data)-1], nil

	return
}

func (self *stack) Peek(i int) (n *Node, err error) {
	if i < 0 || i >= len(self.data) {
		return nil, stackIndexOutOfRange
	}
	return self.data[i], nil
}

func (self *stack) Clear() {
	self.data = self.data[:0]
}

func (self *stack) Len() int { return len(self.data) }
