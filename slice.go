package soa

import "iter"

// Slice is a structure-of-arrays slice derived from a struct T.
type Slice[S Slicer[S], T any] interface {
	Slicer[S]

	// Get gets the value of the index. i.e. s[n]
	Get(int) T
	// Set sets the value of the index. i.e. s[n] = v
	Set(int, T)
	// Len returns the length of the slice. i.e. len(s)
	Len() int
	// Cap returns the capacity of the slice. i.e. cap(s)
	Cap() int
}

// Slicer is a type constraint for a Slice which methods return another Slice.
type Slicer[S any] interface {
	// Slice i.e. s[low:high:max]
	Slice(low, high, max int) S
	// Grow increases the slice's capacity.
	Grow(int) S
}

// Make creates a new Slice.
func Make[S Slice[S, T], T any](len, cap int) S {
	var s S
	s = s.Grow(cap)
	return s.Slice(0, len, cap)
}

// Append appends elements to a Slice.
func Append[S Slice[S, T], T any](slice S, elems ...T) S {
	oldLen := slice.Len()
	newLen := oldLen + len(elems)
	s := slice.Grow(max(0, newLen-slice.Cap()))
	s = s.Slice(0, newLen, s.Cap())
	for i, e := range elems {
		s.Set(oldLen+i, e)
	}
	return s
}

// All returns an iterator over the Slice.
func All[S Slice[S, T], T any](slice S) iter.Seq2[int, T] {
	return func(yield func(int, T) bool) {
		for i := 0; i < slice.Len(); i++ {
			if !yield(i, slice.Get(i)) {
				return
			}
		}
	}
}
