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

func (q *queue) Enqueue(n *Node) {
	if len(q.data) == cap(q.data) && q.head > 0 {
		l := q.Len()
		copy(q.data, q.data[q.head:])
		q.head = 0
		q.data = append(q.data[:l], n)
	} else {
		q.data = append(q.data, n)
	}
}

func (q *queue) Dequeue() (*Node, error) {
	if q.Len() == 0 {
		return nil, emptyQueue
	}

	var n *Node
	n, q.data[q.head] = q.data[q.head], nil
	q.head++

	if q.Len() == 0 {
		q.head = 0
		q.data = q.data[:0]
	}

	return n, nil
}

func (q *queue) Peek(i int) (*Node, error) {
	if i < q.head || i >= len(q.data) {
		return nil, queueIndexOutOfRange
	}
	return q.data[i+q.head], nil
}

func (q *queue) Clear() {
	q.head = 0
	q.data = q.data[:0]
}

func (q *queue) Len() int { return len(q.data) - q.head }

type stack struct {
	data []*Node
}

func (s *stack) Push(n *Node) { s.data = append(s.data, n) }

func (s *stack) Pop() (*Node, error) {
	if len(s.data) == 0 {
		return nil, emptyStack
	}

	var n *Node
	n, s.data, s.data[len(s.data)-1] = s.data[len(s.data)-1], s.data[:len(s.data)-1], nil

	return n, nil
}

func (s *stack) Peek(i int) (*Node, error) {
	if i < 0 || i >= len(s.data) {
		return nil, stackIndexOutOfRange
	}
	return s.data[i], nil
}

func (s *stack) Clear() {
	s.data = s.data[:0]
}

func (s *stack) Len() int { return len(s.data) }
