package soa

import (
	"iter"
	"math"
	"math/bits"
	"sort"
)

// Slice is a structure-of-arrays slice derived from a struct E.
type Slice[S Slicer[S, E], E any] interface {
	Slicer[S, E]

	// Get gets the value of the index. i.e. s[n]
	Get(int) E
	// Set sets the value of the index. i.e. s[n] = v
	Set(int, E)
	// Len returns the length of the slice. i.e. len(s)
	Len() int
	// Cap returns the capacity of the slice. i.e. cap(s)
	Cap() int
}

// Slicer is a type constraint for a Slice which methods return another Slice.
type Slicer[S, E any] interface {
	// Slice i.e. s[low:high:max]
	Slice(low, high, max int) S
	Grow(n int) S
}

// Make creates a new Slice.
func Make[S Slice[S, E], E any](len, cap int) S {
	var s S
	s = s.Grow(cap)
	return s.Slice(0, len, cap)
}

// Append appends elements to a Slice.
func Append[S Slice[S, E], E any](slice S, elems ...E) S {
	oldLen := slice.Len()
	newLen := oldLen + len(elems)
	s := slice.Grow(max(0, newLen-slice.Cap()))
	s = s.Slice(0, newLen, s.Cap())
	for i, e := range elems {
		s.Set(oldLen+i, e)
	}
	return s
}

// Clear clears the slice.
func Clear[S Slice[S, E], E any](slice S) {
	var zero E
	for i := 0; i < slice.Len(); i++ {
		slice.Set(0, zero)
	}
}

// Copy copies elements from one slice to another.
func Copy[S Slice[S, E], E any](dst, src S) int {
	n := min(dst.Len(), src.Len())
	for i := 0; i < n; i++ {
		dst.Set(i, src.Get(i))
	}
	return n
}

// All returns an iterator over the Slice.
func All[S Slice[S, E], E any](slice S) iter.Seq2[int, E] {
	return func(yield func(int, E) bool) {
		for i := 0; i < slice.Len(); i++ {
			if !yield(i, slice.Get(i)) {
				return
			}
		}
	}
}

// AppendSeq appends the values from seq to the slice and returns the extended slice.
func AppendSeq[S Slice[S, E], E any](s S, seq iter.Seq[E]) S {
	for e := range seq {
		s = Append(s, e)
	}
	return s
}

// Backward returns an iterator over index-value pairs in the slice, traversing it backward with descending indices.
func Backward[S Slice[S, E], E any](s S) iter.Seq2[int, E] {
	return func(yield func(int, E) bool) {
		for i := s.Len() - 1; i >= 0; i-- {
			if !yield(i, s.Get(i)) {
				return
			}
		}
	}
}

// BinarySearchFunc searches for target in a sorted Slice.
func BinarySearchFunc[S Slice[S, E], E, T any](x S, target T, cmp func(E, T) int) (int, bool) {
	length := x.Len()
	low, high := 0, length
	for low < high {
		mid := int(uint(low+high) >> 1)
		if cmp(x.Get(mid), target) < 0 {
			low = mid + 1
		} else {
			high = mid
		}
	}
	return low, low < length && cmp(x.Get(low), target) == 0
}

// Chunk returns an iterator over chunks.
func Chunk[S Slice[S, E], E any](s S, n int) iter.Seq[S] {
	if n < 1 {
		panic("cannot be less than 1")
	}

	return func(yield func(S) bool) {
		for i := 0; i < s.Len(); i += n {
			end := min(n, s.Len()-i)
			if !yield(s.Slice(i, i+end, i+end)) {
				return
			}
		}
	}
}

// Clip removes unused capacity from the slice.
func Clip[S Slice[S, E], E any](s S) S {
	return s.Slice(0, s.Len(), s.Len())
}

// Clone returns a copy of the slice.
func Clone[S Slice[S, E], E any](s S) S {
	return Collect[S](Values(s))
}

// Collect collects values from the iterator and returns as a new slice.
func Collect[S Slice[S, E], E any](seq iter.Seq[E]) S {
	var s S
	return AppendSeq(s, seq)
}

// Compact removes duplicates from the slice.
func Compact[S Slice[S, E], E comparable](s S) S {
	return CompactFunc(s, func(e E, e2 E) bool {
		return e == e2
	})
}

// CompactFunc removes duplicates from the slice comparing the elements with the given function.
func CompactFunc[S Slice[S, E], E any](s S, eq func(E, E) bool) S {
	if s.Len() <= 1 {
		return s
	}

	// First, we look for the need for compaction.
	for i := 1; i < s.Len(); i++ {
		if !eq(s.Get(i-1), s.Get(i)) {
			// No compaction needed so far.
			continue
		}

		// Starts compaction.
		for j := i + 1; j < s.Len(); j++ {
			if eq(s.Get(j-1), s.Get(j)) {
				continue
			}

			s.Set(i, s.Get(j))
			i++
		}

		Clear(s.Slice(i, s.Len(), s.Cap()))
		return s.Slice(0, i, s.Cap())
	}

	// No compaction was needed.
	return s
}

// CompareFunc compares 2 slices.
func CompareFunc[S1 Slice[S1, E1], S2 Slice[S2, E2], E1, E2 any](s1 S1, s2 S2, cmp func(E1, E2) int) int {
	l1, l2 := s1.Len(), s2.Len()
	for i := 0; i < l1; i++ {
		if i >= l2 {
			return +1
		}
		e1, e2 := s1.Get(i), s2.Get(i)
		if c := cmp(e1, e2); c != 0 {
			return c
		}
	}
	if l1 < l2 {
		return -1
	}
	return 0
}

// Concat concatenates the slices.
func Concat[S Slice[S, E], E any](slices ...S) S {
	size := 0
	for _, s := range slices {
		size += s.Len()
		if size < 0 {
			panic("len out of range")
		}
	}
	var zero S
	slice := zero.Grow(size)
	for _, s := range slices {
		slice = AppendSeq(slice, Values(s))
	}
	return slice
}

// Contains checks if the slice contains the element.
func Contains[S Slice[S, E], E comparable](s S, v E) bool {
	return ContainsFunc(s, func(e E) bool {
		return v == e
	})
}

// ContainsFunc checks if the slice contains an element that satisfies the predicate.
func ContainsFunc[S Slice[S, E], E any](s S, f func(E) bool) bool {
	return IndexFunc(s, f) >= 0
}

// Delete deletes elements from the slice.
func Delete[S Slice[S, E], E any](s S, i, j int) S {
	if i == j {
		return s
	}

	l := s.Len()
	s = AppendSeq(s.Slice(0, i, s.Cap()), Values(s.Slice(j, s.Len(), s.Cap())))
	Clear(s.Slice(s.Len(), l, s.Cap()))
	return s
}

// DeleteFunc deletes elements that satisfy the predicate.
func DeleteFunc[S Slice[S, E], E any](s S, del func(E) bool) S {
	i := IndexFunc(s, del)
	if i == -1 {
		return s
	}
	for j := i + 1; j < s.Len(); j++ {
		if v := s.Get(j); !del(v) {
			s.Set(i, v)
			i++
		}
	}
	Clear(s.Slice(i, s.Len(), s.Cap()))
	return s.Slice(0, i, s.Cap())
}

// Equal checks equality of two slices.
func Equal[S Slice[S, E], E comparable](s1, s2 S) bool {
	return EqualFunc(s1, s2, func(e1 E, e2 E) bool {
		return e1 == e2
	})
}

// EqualFunc checks equality of two slices by a predicate.
func EqualFunc[S1 Slice[S1, E1], S2 Slice[S2, E2], E1, E2 any](s1 S1, s2 S2, eq func(E1, E2) bool) bool {
	l := s1.Len()
	if s2.Len() != l {
		return false
	}
	for i := 0; i < l; i++ {
		if !eq(s1.Get(i), s2.Get(i)) {
			return false
		}
	}
	return true
}

// Grow grows the capacity of the slice.
func Grow[S Slice[S, E], E any](s S, n int) S {
	return s.Grow(n)
}

// Index returns the index of the first occurrence of the element if it exists. Otherwise, it returns -1.
func Index[S Slice[S, E], E comparable](s S, v E) int {
	return IndexFunc(s, func(e E) bool {
		return v == e
	})
}

// IndexFunc returns the index of the first occurrence of the element that satisfies the predicate. If not exists, it returns -1.
func IndexFunc[S Slice[S, E], E any](s S, f func(E) bool) int {
	for i, e := range All(s) {
		if f(e) {
			return i
		}
	}
	return -1
}

// Insert inserts elements to the slice.
func Insert[S Slice[S, E], E any](s S, i int, v ...E) S {
	l := s.Len()
	// Make a room for inserted elements at the end.
	s = s.Grow(l + len(v))
	c := s.Cap()
	s = s.Slice(0, l+len(v), c)
	// Move the elements following the inserted elements to the end.
	for j := l - 1; j >= i; j-- {
		s.Set(j+len(v), s.Get(j))
	}
	// Insert the elements.
	for j, v := range v {
		s.Set(i+j, v)
	}
	return s
}

// IsSortedFunc checks if the elements are sorted.
func IsSortedFunc[S Slice[S, E], E any](x S, cmp func(a, b E) int) bool {
	for i := x.Len() - 1; i > 0; i-- {
		if cmp(x.Get(i), x.Get(i-1)) < 0 {
			return false
		}
	}
	return true
}

// MaxFunc returns the maximal element in the slice.
func MaxFunc[S Slice[S, E], E any](x S, cmp func(a, b E) int) E {
	l := x.Len()
	if l < 1 {
		panic("soa.MaxFunc: empty list")
	}
	m := x.Get(0)
	for i := 1; i < l; i++ {
		e := x.Get(i)
		if cmp(e, m) > 0 {
			m = e
		}
	}
	return m
}

// MinFunc returns the minimum element in the slice.
func MinFunc[S Slice[S, E], E any](x S, cmp func(a, b E) int) E {
	l := x.Len()
	if l < 1 {
		panic("soa.MinFunc: empty list")
	}
	m := x.Get(0)
	for i := 1; i < l; i++ {
		e := x.Get(i)
		if cmp(e, m) < 0 {
			m = e
		}
	}
	return m
}

// Repeat returns a slice that repeats the elements the given times.
func Repeat[S Slice[S, E], E any](x S, count int) S {
	if count < 0 {
		panic("cannot be negative")
	}

	l := x.Len()
	if hi, lo := bits.Mul(uint(l), uint(count)); hi > 0 || lo > math.MaxInt {
		panic("the result of (len(x) * count) overflows")
	}

	l = l * count
	s := Make[S](l, l)
	n := 0
	for n < l {
		n += Copy(s.Slice(n, l, l), x)
	}
	return s
}

// Replace replaces values.
func Replace[S Slice[S, E], E any](s S, i, j int, v ...E) S {
	ol, d := s.Len(), len(v)-(j-i)
	nl := ol + d
	if nl > ol {
		// Make a room for inserted elements at the end.
		s = s.Grow(nl)
		s = s.Slice(0, nl, s.Cap())
	}
	// Move the elements following the replaced elements to the end.
	for k := ol - 1; k >= j; k-- {
		s.Set(k+d, s.Get(k))
	}
	// Replace the elements.
	for k, v := range v {
		s.Set(i+k, v)
	}
	return s
}

// Reverse reverses the slice.
func Reverse[S Slice[S, E], E any](s S) {
	for i, j := 0, s.Len()-1; i < j; i, j = i+1, j-1 {
		x, y := s.Get(i), s.Get(j)
		s.Set(i, y)
		s.Set(j, x)
	}
}

// SortFunc sorts the slice.
func SortFunc[S Slice[S, E], E any](x S, cmp func(a, b E) int) {
	sort.Sort(sortable[S, E]{slice: x, cmp: cmp})
}

// SortStableFunc stable sorts the slice.
func SortStableFunc[S Slice[S, E], E any](x S, cmp func(a, b E) int) {
	sort.Stable(sortable[S, E]{slice: x, cmp: cmp})
}

// SortedFunc collects values from seq into a new sorted slice.
func SortedFunc[S Slice[S, E], E any](seq iter.Seq[E], cmp func(E, E) int) S {
	s := Collect[S](seq)
	SortFunc(s, cmp)
	return s
}

// SortedStableFunc collects values from seq into a new stable-sorted slice.
func SortedStableFunc[S Slice[S, E], E any](seq iter.Seq[E], cmp func(E, E) int) S {
	s := Collect[S](seq)
	SortStableFunc(s, cmp)
	return s
}

// Values returns an iterator over the Slice.
func Values[S Slice[S, E], E any](s S) iter.Seq[E] {
	return func(yield func(E) bool) {
		for i := 0; i < s.Len(); i++ {
			if !yield(s.Get(i)) {
				return
			}
		}
	}
}

type sortable[S Slice[S, E], E any] struct {
	slice S
	cmp   func(a, b E) int
}

func (s sortable[S, E]) Len() int {
	return s.slice.Len()
}

func (s sortable[S, E]) Less(i, j int) bool {
	x, y := s.slice.Get(i), s.slice.Get(j)
	return s.cmp(x, y) < 0
}

func (s sortable[S, E]) Swap(i, j int) {
	x, y := s.slice.Get(i), s.slice.Get(j)
	s.slice.Set(i, y)
	s.slice.Set(j, x)
}
