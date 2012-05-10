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
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

import (
	check "launchpad.net/gocheck"
	"math"
	"math/rand"
	"runtime"
	"testing"
)

type N struct {
	id    int
	tails []int
}

// graph - all edges described at both nodes (historical reasons)
var (
	testG = [][]N{
		{
			{id: 1, tails: []int{19, 15, 36, 23, 18, 39}},
			{id: 2, tails: []int{36, 23, 4, 18, 26, 9}},
			{id: 3, tails: []int{35, 6, 16, 11}},
			{id: 4, tails: []int{23, 2, 18, 24}},
			{id: 5, tails: []int{14, 8, 29, 21}},
			{id: 6, tails: []int{34, 35, 3, 16}},
			{id: 7, tails: []int{30, 33, 38, 28}},
			{id: 8, tails: []int{12, 14, 5, 29, 31}},
			{id: 9, tails: []int{39, 13, 20, 10, 17, 2}},
			{id: 10, tails: []int{9, 20, 12, 14, 29}},
			{id: 11, tails: []int{3, 16, 30, 33, 26}},
			{id: 12, tails: []int{20, 10, 14, 8}},
			{id: 13, tails: []int{24, 39, 9, 20}},
			{id: 14, tails: []int{10, 12, 8, 5}},
			{id: 15, tails: []int{26, 19, 1, 36}},
			{id: 16, tails: []int{6, 3, 11, 30, 17, 35, 32}},
			{id: 17, tails: []int{38, 28, 32, 40, 9, 16}},
			{id: 18, tails: []int{2, 4, 24, 39, 1}},
			{id: 19, tails: []int{27, 26, 15, 1}},
			{id: 20, tails: []int{13, 9, 10, 12}},
			{id: 21, tails: []int{5, 29, 25, 37}},
			{id: 22, tails: []int{32, 40, 34, 35}},
			{id: 23, tails: []int{1, 36, 2, 4}},
			{id: 24, tails: []int{4, 18, 39, 13}},
			{id: 25, tails: []int{29, 21, 37, 31}},
			{id: 26, tails: []int{31, 27, 19, 15, 11, 2}},
			{id: 27, tails: []int{37, 31, 26, 19, 29}},
			{id: 28, tails: []int{7, 38, 17, 32}},
			{id: 29, tails: []int{8, 5, 21, 25, 10, 27}},
			{id: 30, tails: []int{16, 11, 33, 7, 37}},
			{id: 31, tails: []int{25, 37, 27, 26, 8}},
			{id: 32, tails: []int{28, 17, 40, 22, 16}},
			{id: 33, tails: []int{11, 30, 7, 38}},
			{id: 34, tails: []int{40, 22, 35, 6}},
			{id: 35, tails: []int{22, 34, 6, 3, 16}},
			{id: 36, tails: []int{15, 1, 23, 2}},
			{id: 37, tails: []int{21, 25, 31, 27, 30}},
			{id: 38, tails: []int{33, 7, 28, 17, 40}},
			{id: 39, tails: []int{18, 24, 13, 9, 1}},
			{id: 40, tails: []int{17, 32, 22, 34, 38}},
		},
		{
			{id: 1, tails: []int{4}},
			{id: 2, tails: []int{3, 4}},
			{id: 3, tails: []int{2, 4}},
			{id: 4, tails: []int{1, 2, 3, 5}},
			{id: 5, tails: []int{4, 6}},
			{id: 6, tails: []int{5}},
		},
	}
	cutExpects = []float64{3, 1}
)

// Helpers
func createGraph(nodes []N) (g *Undirected) {
	g = NewUndirected()
	for _, n := range nodes {
		h, _ := g.Add(n.id)
		for _, tid := range n.tails {
			t, _ := g.Add(tid)
			if n.id < tid {
				g.Connect(h, t, 1, 0)
			}
		}
	}

	return
}

// Tests
func (s *S) TestKargerOrig(c *check.C) {
	rand.Seed(0)
	for j, g := range testG {
		k := newKarger(createGraph(g))
		mc := float64(k.g.Size())
		for i := 0; i < k.g.Order()*k.g.Order()*int(math.Log(float64(k.g.Order()))+1); i++ {
			ce, w := k.randMinCut()
			c.Check(len(ce), check.Equals, int(w))
			if w < mc {
				mc = w
			}
		}
		c.Check(mc, check.Equals, cutExpects[j])
	}
}
func (s *S) TestKargerSS(c *check.C) {
	rand.Seed(0)
	for j, g := range testG {
		k := newKargerSS(createGraph(g))
		mc := float64(k.g.Size())
		for i := 0; i < k.g.Order()*k.g.Order()*int(math.Log(float64(k.g.Order()))+1); i++ {
			ce, w := k.randCut()
			c.Check(len(ce), check.Equals, int(w))
			if w < mc {
				mc = w
			}
		}
		c.Check(mc, check.Equals, cutExpects[j])
	}
}
func (s *S) TestKargerFastMinCut(c *check.C) {
	rand.Seed(0)
	for j, g := range testG {
		G := createGraph(g)
		lo := int(math.Log(float64(G.Order())))
		_, mc := FastRandMinCut(G, lo*lo)
		c.Check(mc, check.Equals, cutExpects[j])
	}
}
func (s *S) TestKargerFastMinCutPar(c *check.C) {
	rand.Seed(0)
	for j, g := range testG {
		G := createGraph(g)
		lo := int(math.Log(float64(G.Order())))
		_, mc := FastRandMinCutPar(G, lo*lo, runtime.GOMAXPROCS(0))
		c.Check(mc, check.Equals, cutExpects[j])
	}
}
func (s *S) TestKargerLL(c *check.C) {
	rand.Seed(0)
	for j, g := range testG {
		k := newKargerLL(createGraph(g))
		mc := float64(k.g.Size())
		for i := 0; i < k.g.Order()*k.g.Order()*int(math.Log(float64(k.g.Order()))+1); i++ {
			ce, w := k.randMinCut()
			c.Check(len(ce), check.Equals, int(w))
			if w < mc {
				mc = w
			}
		}
		c.Check(mc, check.Equals, cutExpects[j])
	}
}
func (s *S) TestKargerParFastMinCut(c *check.C) {
	rand.Seed(0)
	for j, g := range testG {
		G := createGraph(g)
		lo := int(math.Log(float64(G.Order())))
		_, mc := ParFastRandMinCut(G, lo*lo, runtime.GOMAXPROCS(0))
		c.Check(mc, check.Equals, cutExpects[j])
	}
}

func BenchmarkKargerOrig(b *testing.B) {
	for j := 0; j < b.N; j++ {
		k := newKarger(createGraph(testG[0]))
		mc := float64(k.g.Size())
		for i := 0; i < k.g.Order()*k.g.Order()*int(math.Log(float64(k.g.Order()))+1); i++ {
			_, w := k.randMinCut()
			if w < mc {
				mc = w
			}
		}
	}
}
func BenchmarkKargerSS(b *testing.B) {
	for j := 0; j < b.N; j++ {
		k := newKargerSS(createGraph(testG[0]))
		mc := float64(k.g.Size())
		for i := 0; i < k.g.Order()*k.g.Order()*int(math.Log(float64(k.g.Order()))+1); i++ {
			_, w := k.randCut()
			if w < mc {
				mc = w
			}
		}
	}
}
func BenchmarkKargerLL(b *testing.B) {
	for j := 0; j < b.N; j++ {
		k := newKargerLL(createGraph(testG[0]))
		mc := float64(k.g.Size())
		for i := 0; i < k.g.Order()*k.g.Order()*int(math.Log(float64(k.g.Order()))+1); i++ {
			_, w := k.randMinCut()
			if w < mc {
				mc = w
			}
		}
	}
}
func BenchmarkFastKarger(b *testing.B) {
	G := createGraph(testG[0])
	lo := int(math.Log(float64(G.Order())))
	for j := 0; j < b.N; j++ {
		FastRandMinCut(G, lo*lo)
	}
}
func BenchmarkFastKargerPar(b *testing.B) {
	G := createGraph(testG[0])
	lo := int(math.Log(float64(G.Order())))
	for j := 0; j < b.N; j++ {
		FastRandMinCutPar(G, lo*lo, runtime.GOMAXPROCS(0))
	}
}
func BenchmarkParFastKarger(b *testing.B) {
	G := createGraph(testG[0])
	lo := int(math.Log(float64(G.Order())))
	for j := 0; j < b.N; j++ {
		ParFastRandMinCut(G, lo*lo, runtime.GOMAXPROCS(0))
	}
}
