// Copyright ©2012 The bíogo.graph Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package graph

import (
	"flag"
	check "launchpad.net/gocheck"
	"math/rand"
	"time"
)

var prob = flag.Bool("prob", false, "enables probabilistic testing of the random selector")

// Tests

var (
	exp     = []float64{1 << 0, 1 << 1, 1 << 2, 1 << 3, 1 << 4, 1 << 5, 1 << 6, 1 << 7, 1 << 8, 1 << 9}
	sigChi2 = 16.92 // p = 0.05 df = 9
	sel     = Selector{
		{Index: 1, Weight: exp[0]},
		{Index: 2, Weight: exp[1]},
		{Index: 3, Weight: exp[2]},
		{Index: 4, Weight: exp[3]},
		{Index: 5, Weight: exp[4]},
		{Index: 6, Weight: exp[5]},
		{Index: 7, Weight: exp[6]},
		{Index: 8, Weight: exp[7]},
		{Index: 9, Weight: exp[8]},
		{Index: 10, Weight: exp[9]},
	}
	tot = Selector{
		{Index: 1, Weight: exp[0], total: exp[0] + exp[1] + exp[3] + exp[4] + exp[7] + exp[8] + exp[9] + exp[2] + exp[5] + exp[6]},
		{Index: 2, Weight: exp[1], total: exp[1] + exp[3] + exp[4] + exp[7] + exp[8] + exp[9]},
		{Index: 3, Weight: exp[2], total: exp[2] + exp[5] + exp[6]},
		{Index: 4, Weight: exp[3], total: exp[3] + exp[7] + exp[8]},
		{Index: 5, Weight: exp[4], total: exp[4] + exp[9]},
		{Index: 6, Weight: exp[5], total: exp[5]},
		{Index: 7, Weight: exp[6], total: exp[6]},
		{Index: 8, Weight: exp[7], total: exp[7]},
		{Index: 9, Weight: exp[8], total: exp[8]},
		{Index: 10, Weight: exp[9], total: exp[9]},
	}
	dnw = Selector{
		{Index: 1, Weight: exp[0], total: exp[0] + exp[1] + exp[3] + exp[4] + exp[7] + exp[8] + exp[9] + exp[2] + exp[5]},
		{Index: 2, Weight: exp[1], total: exp[1] + exp[3] + exp[4] + exp[7] + exp[8] + exp[9]},
		{Index: 3, Weight: exp[2], total: exp[2] + exp[5]},
		{Index: 4, Weight: exp[3], total: exp[3] + exp[7] + exp[8]},
		{Index: 5, Weight: exp[4], total: exp[4] + exp[9]},
		{Index: 6, Weight: exp[5], total: exp[5]},
		{Index: 7, Weight: 0, total: 0},
		{Index: 8, Weight: exp[7], total: exp[7]},
		{Index: 9, Weight: exp[8], total: exp[8]},
		{Index: 10, Weight: exp[9], total: exp[9]},
	}
	upw = Selector{
		{Index: 1, Weight: exp[0], total: exp[0] + exp[1] + exp[3] + exp[4] + exp[7] + exp[8] + exp[9] + exp[2] + exp[5] + exp[9]*2},
		{Index: 2, Weight: exp[1], total: exp[1] + exp[3] + exp[4] + exp[7] + exp[8] + exp[9]},
		{Index: 3, Weight: exp[2], total: exp[2] + exp[5] + exp[9]*2},
		{Index: 4, Weight: exp[3], total: exp[3] + exp[7] + exp[8]},
		{Index: 5, Weight: exp[4], total: exp[4] + exp[9]},
		{Index: 6, Weight: exp[5], total: exp[5]},
		{Index: 7, Weight: exp[9] * 2, total: exp[9] * 2},
		{Index: 8, Weight: exp[7], total: exp[7]},
		{Index: 9, Weight: exp[8], total: exp[8]},
		{Index: 10, Weight: exp[9], total: exp[9]},
	}

	obt = []float64{973, 1937, 3898, 7897, 15769, 31284, 62176, 125408, 250295, 500363}
)

func (s *S) TestWeightedUnseeded(c *check.C) {
	rand.Seed(0)
	f := make([]float64, len(sel))
	ts := make(Selector, len(sel))

	copy(ts, sel)
	ts.Init()
	c.Check(ts, check.DeepEquals, tot)

	for i := 0; i < 1e6; i++ {
		copy(ts, sel)
		ts.Init()
		item, err := ts.Select()
		if err != nil {
			c.Fatal(err)
		}
		f[item-1]++
	}

	fsum, exsum := 0., 0.
	for i := range f {
		fsum += f[i]
		exsum += exp[i]
	}
	fac := fsum / exsum
	for i := range f {
		exp[i] *= fac
	}

	// Check that we get exactly what we expect
	c.Check(f, check.DeepEquals, obt)

	// Check that this is within statistical expectations - we know this is true for this set.
	X := chi2(f, exp)
	c.Logf("H₀: d(Sample) = d(Expect), H₁: d(S) ≠ d(Expect). df = %d, p = 0.05, X² threshold = %.2f, X² = %f", len(f)-1, sigChi2, X)
	c.Check(X < sigChi2, check.Equals, true)
}

func (s *S) TestWeightedTimeSeeded(c *check.C) {
	if !*prob {
		c.Skip("probabilistic testing not requested")
	}
	c.Log("Note: This test is stochastic and is expected to fail with probability ≈ 0.05.")
	rand.Seed(time.Now().Unix())
	f := make([]float64, len(sel))
	ts := make(Selector, len(sel))

	for i := 0; i < 1e6; i++ {
		copy(ts, sel)
		ts.Init()
		item, err := ts.Select()
		if err != nil {
			c.Fatal(err)
		}
		f[item-1]++
	}

	fsum, exsum := 0., 0.
	for i := range f {
		fsum += f[i]
		exsum += exp[i]
	}
	fac := fsum / exsum
	for i := range f {
		exp[i] *= fac
	}

	// Check that our obtained values are within statistical expectations for p = 0.05.
	// This will not be true approximately 1 in 20 tests.
	X := chi2(f, exp)
	c.Logf("H₀: d(Sample) = d(Expect), H₁: d(S) ≠ d(Expect). df = %d, p = 0.05, X² threshold = %.2f, X² = %f", len(f)-1, sigChi2, X)
	c.Check(X < sigChi2, check.Equals, true)
}

func (s *S) TestWeightZero(c *check.C) {
	rand.Seed(0)
	f := make([]float64, len(sel))
	ts := make(Selector, len(sel))

	copy(ts, sel)
	ts.Init()
	ts.Weight(6, 0)
	c.Check(ts, check.DeepEquals, dnw)

	for i := 0; i < 1e6; i++ {
		copy(ts, sel)
		ts.Init()
		ts.Weight(6, 0)
		item, err := ts.Select()
		if err != nil {
			c.Fatal(err)
		}
		f[item-1]++
	}

	fsum, exsum := 0., 0.
	for i := range f {
		fsum += f[i]
		exsum += exp[i]
	}
	fac := fsum / exsum
	for i := range f {
		exp[i] *= fac
	}

	// Check that we get exactly what we expect
	c.Check(f[:6], check.Not(check.DeepEquals), obt[:6])
	c.Check(f[7:], check.Not(check.DeepEquals), obt[7:])
	c.Check(f[6], check.Equals, 0.)
}

func (s *S) TestWeightIncrease(c *check.C) {
	rand.Seed(0)
	f := make([]float64, len(sel))
	ts := make(Selector, len(sel))

	copy(ts, sel)
	ts.Init()
	ts.Weight(6, sel[len(sel)-1].Weight*2)
	c.Check(ts, check.DeepEquals, upw)

	for i := 0; i < 1e6; i++ {
		copy(ts, sel)
		ts.Init()
		ts.Weight(6, sel[len(sel)-1].Weight*2)
		item, err := ts.Select()
		if err != nil {
			c.Fatal(err)
		}
		f[item-1]++
	}

	fsum, exsum := 0., 0.
	for i := range f {
		fsum += f[i]
		exsum += exp[i]
	}
	fac := fsum / exsum
	for i := range f {
		exp[i] *= fac
	}

	// Check that we get exactly what we expect
	c.Check(f[:6], check.Not(check.DeepEquals), obt[:6])
	c.Check(f[7:], check.Not(check.DeepEquals), obt[7:])
	c.Check(f[6] > f[9], check.Equals, true)
}

func chi2(ob, ex []float64) (sum float64) {
	for i := range ob {
		x := ob[i] - ex[i]
		sum += (x * x) / ex[i]
	}

	return
}
