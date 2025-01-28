package soa

import (
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
