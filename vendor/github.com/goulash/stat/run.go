// Copyright (c) 2016, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package stat

import "math"

// Run calculates the running mean, variance, and standard deviation.
//
// Note: Run contains only plain old datatypes, so a shallow copy is
// a complete copy.
type Run struct {
	n int64
	m float64
	s float64

	max float64
	min float64
}

// Reset the values to zero.
func (r *Run) Reset() {
	r.n = 0
	r.m = 0
	r.s = 0
	r.max = math.Inf(-1)
	r.min = math.Inf(1)
}

// Add a new value to the run.
func (r *Run) Add(x float64) {
	// This is only necessary because we want to allow the zero value of
	// Run to be used. All we need is r.min = Inf and r.max = -Inf.
	// If only we could have a constructor.
	if r.n == 0 {
		r.n = 1
		r.m = x
		r.min = x
		r.max = x
		return
	}

	r.n++
	if r.max < x {
		r.max = x
	}
	if r.min > x {
		r.min = x
	}
	m := r.m + (x-r.m)/float64(r.n)
	r.s = r.s + (x-r.m)*(x-m)
	r.m = m
}

func (r *Run) AddN(n int64, x float64) {
	if r.n == 0 {
		r.n = n
		r.m = x
		r.max = x
		r.min = x
		return
	}

	if r.max < x {
		r.max = x
	}
	if r.min > x {
		r.min = x
	}
	// TODO: Double check this calculation!!!
	i := float64(n - r.n)
	m := r.m + (i*x-i*r.m)/float64(n)
	r.s = r.s + i*(x-r.m)*(x-m)
	r.m = m
	r.n = n
}

func (r *Run) N() int64 { return r.n }

// Max returns the max.
func (r *Run) Max() float64 {
	if r.n == 0 {
		return math.Inf(-1)
	}
	return r.max
}

// Min returns the min.
func (r *Run) Min() float64 {
	if r.n == 0 {
		return math.Inf(1)
	}
	return r.min
}

// Mean returns the mean.
func (r *Run) Mean() float64 {
	if r.n == 0 {
		return math.NaN()
	}
	return r.m
}

// Var returns the sample variance.
func (r *Run) Var() float64 {
	if r.n <= 1 {
		if r.n == 0 {
			return math.NaN()
		}
		return 0
	}
	return r.s / float64(r.n-1)
}

// VarP returns the population variance.
func (r *Run) VarP() float64 {
	if r.n <= 1 {
		if r.n == 0 {
			return math.NaN()
		}
		return 0
	}
	return r.s / float64(r.n)
}

// Std returns the sample standard deviation.
func (r *Run) Std() float64 {
	return math.Sqrt(r.Var())
}

// StdP returns the population standard deviation.
func (r *Run) StdP() float64 {
	return math.Sqrt(r.VarP())
}
