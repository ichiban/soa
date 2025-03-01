package soa

import (
	"iter"
	"reflect"
	"slices"
	"testing"
)

type User struct {
	ID   int
	Name string
}

type UserSlice struct {
	ID   []int
	Name []string
}

var _ Slice[UserSlice, User] = UserSlice{}

func (s UserSlice) Get(i int) User {
	var u User
	u.ID = s.ID[i]
	u.Name = s.Name[i]
	return u
}

func (s UserSlice) Set(i int, t User) {
	s.ID[i] = t.ID
	s.Name[i] = t.Name
}

func (s UserSlice) Len() int {
	return len(s.ID)
}

func (s UserSlice) Cap() int {
	return min(cap(s.ID), cap(s.Name))
}

func (s UserSlice) Slice(low, high, max int) UserSlice {
	return UserSlice{
		ID:   s.ID[low:high:max],
		Name: s.Name[low:high:max],
	}
}

func (s UserSlice) Grow(n int) UserSlice {
	return UserSlice{
		ID:   slices.Grow(s.ID, n),
		Name: slices.Grow(s.Name, n),
	}
}

func take[T any](i iter.Seq[T], n int) iter.Seq[T] {
	return func(yield func(T) bool) {
		for e := range i {
			if n == 0 {
				return
			}
			if !yield(e) {
				return
			}
			n--
		}
	}
}

func TestMake(t *testing.T) {
	s := Make[UserSlice](0, 3)
	if len(s.ID) != 0 {
		t.Error("ID is not empty")
	}
	if cap(s.ID) != 3 {
		t.Error("ID's capacity is not correct")
	}
	if len(s.Name) != 0 {
		t.Error("Name is not empty")
	}
	if cap(s.Name) != 3 {
		t.Error("Name's capacity is not correct")
	}
}

func TestAppend(t *testing.T) {
	s := Make[UserSlice](0, 3)
	s = Append(s, User{ID: 1, Name: "John"})
	s = Append(s, User{ID: 2, Name: "Jane"})
	s = Append(s, User{ID: 3, Name: "Bob"})
	if s.ID[0] != 1 || s.Name[0] != "John" {
		t.Error("User is not appended to slice")
	}
	if s.ID[1] != 2 || s.Name[1] != "Jane" {
		t.Error("User is not appended to slice")
	}
	if s.ID[2] != 3 || s.Name[2] != "Bob" {
		t.Error("User is not appended to slice")
	}
}

func TestAll(t *testing.T) {
	users := []User{
		{ID: 1, Name: "Alice"},
		{ID: 2, Name: "Bob"},
		{ID: 3, Name: "Charlie"},
	}
	s := Append(Make[UserSlice](0, 3), users...)

	for i, u := range All(s) {
		if i == 2 {
			return
		}
		if u != users[i] {
			t.Errorf("All missed a user: %v", u)
		}
	}
}

func TestAppendSeq(t *testing.T) {
	users := []User{
		{ID: 1, Name: "Alice"},
		{ID: 2, Name: "Bob"},
		{ID: 3, Name: "Charlie"},
	}
	s := AppendSeq(Make[UserSlice](0, 3), slices.Values(users))

	if !reflect.DeepEqual(UserSlice{
		ID:   []int{1, 2, 3},
		Name: []string{"Alice", "Bob", "Charlie"},
	}, s) {
		t.Error("AppendSeq didn't append to slice")
	}
}

func TestBackward(t *testing.T) {
	users := []User{
		{ID: 1, Name: "Alice"},
		{ID: 2, Name: "Bob"},
		{ID: 3, Name: "Charlie"},
	}
	s := Append(Make[UserSlice](0, 3), users...)

	for i, u := range Backward(s) {
		if i == 2 {
			return
		}
		if u != users[len(users)-1-i] {
			t.Errorf("Backward missed a user: %v", u)
		}
	}
}

func TestBinarySearchFunc(t *testing.T) {
	users := []User{
		{ID: 1, Name: "Alice"},
		{ID: 2, Name: "Bob"},
		{ID: 3, Name: "Charlie"},
	}
	s := Append(Make[UserSlice](0, 3), users...)

	n, ok := BinarySearchFunc(s, 2, func(u User, n int) int {
		return u.ID - n
	})
	if n != 1 {
		t.Error("BinarySearchFunc didn't find user")
	}
	if !ok {
		t.Error("BinarySearchFunc didn't find user")
	}
}

func TestChunk(t *testing.T) {
	users := []User{
		{ID: 1, Name: "Alice"},
		{ID: 2, Name: "Bob"},
		{ID: 3, Name: "Charlie"},
		{ID: 4, Name: "Dan"},
		{ID: 5, Name: "Eve"},
		{ID: 6, Name: "Frank"},
		{ID: 7, Name: "Grace"},
		{ID: 8, Name: "Heidi"},
		{ID: 9, Name: "Ivan"},
		{ID: 10, Name: "Judy"},
	}
	s := Append(Make[UserSlice](0, 10), users...)

	tests := []struct {
		title  string
		s      UserSlice
		n      int
		chunks []UserSlice
		panics bool
		size   int
	}{
		{title: "empty", s: s, n: 0, chunks: []UserSlice{}, panics: true, size: -1},
		{title: "ok", s: s, n: 3, chunks: []UserSlice{
			{
				ID:   []int{1, 2, 3},
				Name: []string{"Alice", "Bob", "Charlie"},
			},
			{
				ID:   []int{4, 5, 6},
				Name: []string{"Dan", "Eve", "Frank"},
			},
			{
				ID:   []int{7, 8, 9},
				Name: []string{"Grace", "Heidi", "Ivan"},
			},
			{
				ID:   []int{10},
				Name: []string{"Judy"},
			},
		}, size: -1},
		{title: "ok with size limit", s: s, n: 3, chunks: []UserSlice{
			{
				ID:   []int{1, 2, 3},
				Name: []string{"Alice", "Bob", "Charlie"},
			},
			{
				ID:   []int{4, 5, 6},
				Name: []string{"Dan", "Eve", "Frank"},
			},
		}, size: 2},
	}

	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			defer func() {
				r := recover()
				if (r != nil) != test.panics {
					t.Errorf("panic expected: %v", test.panics)
				}
			}()
			var chunks []UserSlice
			for chunk := range take(Chunk(test.s, test.n), test.size) {
				chunks = append(chunks, chunk)
			}
			if !reflect.DeepEqual(chunks, test.chunks) {
				t.Errorf("Chunks didn't match: %v != %v", chunks, test.chunks)
			}
		})
	}
}

func TestClip(t *testing.T) {
	users := []User{
		{ID: 1, Name: "Alice"},
		{ID: 2, Name: "Bob"},
		{ID: 3, Name: "Charlie"},
	}
	s := Append(Make[UserSlice](0, 5), users...)

	tests := []struct {
		title   string
		s       UserSlice
		lenID   int
		capID   int
		lenName int
		capName int
	}{
		{title: "ok", s: s, lenID: 3, capID: 3, lenName: 3, capName: 3},
	}

	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			clip := Clip(test.s)
			if len(clip.ID) != test.lenID {
				t.Errorf("ID didn't match: %v != %v", test.lenID, len(clip.ID))
			}
			if cap(clip.ID) != test.capID {
				t.Errorf("ID didn't match: %v != %v", test.capID, len(clip.ID))
			}
			if len(clip.Name) != test.lenName {
				t.Errorf("Name didn't match: %v != %v", test.lenName, len(clip.Name))
			}
			if cap(clip.Name) != test.capName {
				t.Errorf("Name didn't match: %v != %v", test.capName, len(clip.Name))
			}
		})
	}
}
