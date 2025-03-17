package soa

import (
	"iter"
	"math"
	"reflect"
	"slices"
	"strings"
	"testing"
)

type User struct {
	ID   int
	Name string
}

type UserSlice struct {
	ID   []int
	Name []string

	fakeLen int
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
	return max(min(len(s.ID), len(s.Name)), s.fakeLen)
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

func TestClone(t *testing.T) {
	users := []User{
		{ID: 1, Name: "Alice"},
		{ID: 2, Name: "Bob"},
		{ID: 3, Name: "Charlie"},
	}
	s := Append(Make[UserSlice](0, 5), users...)

	clone := Clone(s)
	if !reflect.DeepEqual(clone, s) {
		t.Errorf("Clone didn't match: %v != %v", clone, s)
	}

	s.Set(1, User{ID: 4, Name: "Dan"})

	if reflect.DeepEqual(clone, s) {
		t.Error("Clone didn't copy")
	}
}

func TestCollect(t *testing.T) {
	s := Collect[UserSlice](slices.Values([]User{
		{ID: 1, Name: "Alice"},
		{ID: 2, Name: "Bob"},
		{ID: 3, Name: "Charlie"},
	}))

	if !reflect.DeepEqual(s, UserSlice{
		ID:   []int{1, 2, 3},
		Name: []string{"Alice", "Bob", "Charlie"},
	}) {
		t.Errorf("Collect didn't match: %v", s)
	}
}

func TestCompact(t *testing.T) {
	tests := []struct {
		title  string
		s      UserSlice
		result UserSlice
	}{
		{
			title: "empty",
		},
		{
			title: "no compaction needed",
			s: UserSlice{
				ID:   []int{1, 2, 3},
				Name: []string{"Alice", "Bob", "Charlie"},
			},
			result: UserSlice{
				ID:   []int{1, 2, 3},
				Name: []string{"Alice", "Bob", "Charlie"},
			},
		},
		{
			title: "compaction needed",
			s: UserSlice{
				ID:   []int{1, 1, 2, 2, 3},
				Name: []string{"Alice", "Alice", "Bob", "Bob", "Charlie"},
			},
			result: UserSlice{
				ID:   []int{1, 2, 3},
				Name: []string{"Alice", "Bob", "Charlie"},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			result := Compact(test.s)
			if !reflect.DeepEqual(result, test.result) {
				t.Errorf("CompactFunc didn't match: %v != %v", test.result, test.result)
			}
		})
	}
}

func TestCompactFunc(t *testing.T) {
	tests := []struct {
		title  string
		s      UserSlice
		eq     func(User, User) bool
		result UserSlice
	}{
		{
			title: "empty",
		},
		{
			title: "no compaction needed",
			s: UserSlice{
				ID:   []int{1, 2, 3},
				Name: []string{"Alice", "Bob", "Charlie"},
			},
			eq: func(a, b User) bool { return a == b },
			result: UserSlice{
				ID:   []int{1, 2, 3},
				Name: []string{"Alice", "Bob", "Charlie"},
			},
		},
		{
			title: "compaction needed",
			s: UserSlice{
				ID:   []int{1, 1, 2, 2, 3},
				Name: []string{"Alice", "Alice", "Bob", "Bob", "Charlie"},
			},
			eq: func(a, b User) bool { return a == b },
			result: UserSlice{
				ID:   []int{1, 2, 3},
				Name: []string{"Alice", "Bob", "Charlie"},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			result := CompactFunc(test.s, test.eq)
			if !reflect.DeepEqual(result, test.result) {
				t.Errorf("CompactFunc didn't match: %v != %v", test.result, test.result)
			}
		})
	}
}

func TestCompareFunc(t *testing.T) {
	tests := []struct {
		title  string
		s1     UserSlice
		s2     UserSlice
		cmp    func(User, User) int
		result int
	}{
		{
			title: "empty",
		},
		{
			title: "different element",
			s1:    UserSlice{ID: []int{1, 2, 3}, Name: []string{"Alice", "Bob", "Charlie"}},
			s2:    UserSlice{ID: []int{1, 4, 3}, Name: []string{"Alice", "Dan", "Charlie"}},
			cmp: func(a, b User) int {
				if o := a.ID - b.ID; o != 0 {
					return o
				}
				return strings.Compare(a.Name, b.Name)
			},
			result: -2,
		},
		{
			title: "s1 is longer",
			s1:    UserSlice{ID: []int{1, 2, 3}, Name: []string{"Alice", "Bob", "Charlie"}},
			s2:    UserSlice{ID: []int{1, 2}, Name: []string{"Alice", "Bob"}},
			cmp: func(a, b User) int {
				if o := a.ID - b.ID; o != 0 {
					return o
				}
				return strings.Compare(a.Name, b.Name)
			},
			result: +1,
		},
		{
			title: "s2 is longer",
			s1:    UserSlice{ID: []int{1, 2}, Name: []string{"Alice", "Bob"}},
			s2:    UserSlice{ID: []int{1, 2, 3}, Name: []string{"Alice", "Bob", "Charlie"}},
			cmp: func(a, b User) int {
				if o := a.ID - b.ID; o != 0 {
					return o
				}
				return strings.Compare(a.Name, b.Name)
			},
			result: -1,
		},
	}

	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			if o := CompareFunc(test.s1, test.s2, test.cmp); test.result != o {
				t.Errorf("CompareFunc didn't match: %v != %v", o, test.result)
			}
		})
	}
}

func TestConcat(t *testing.T) {
	tests := []struct {
		title  string
		slices []UserSlice
		result UserSlice
		panics bool
	}{
		{
			title: "empty",
		},
		{
			title: "ok",
			slices: []UserSlice{
				{ID: []int{1, 2}, Name: []string{"Alice", "Bob"}},
				{ID: []int{3, 4}, Name: []string{"Charlie", "Dan"}},
			},
			result: UserSlice{ID: []int{1, 2, 3, 4}, Name: []string{"Alice", "Bob", "Charlie", "Dan"}},
		},
		{
			title: "too long",
			slices: []UserSlice{
				{ID: []int{1, 2}, Name: []string{"Alice", "Bob"}},
				{fakeLen: math.MaxInt},
			},
			panics: true,
		},
	}

	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			defer func() {
				r := recover()
				if (r != nil) != test.panics {
					t.Errorf("Concat didn't panic: %v", test.panics)
				}
			}()
			s := Concat(test.slices...)
			if !reflect.DeepEqual(s, test.result) {
				t.Errorf("Concat didn't match: %v != %v", test.result, s)
			}
		})
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		title string
		s     UserSlice
		e     User
		ok    bool
	}{
		{title: "ok", s: UserSlice{ID: []int{1, 2, 3}, Name: []string{"Alice", "Bob", "Charlie"}}, e: User{ID: 2, Name: "Bob"}, ok: true},
		{title: "ng", s: UserSlice{ID: []int{1, 2, 3}, Name: []string{"Alice", "Bob", "Charlie"}}, e: User{ID: 4, Name: "Dan"}, ok: false},
	}

	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			ok := Contains(test.s, test.e)
			if ok != test.ok {
				t.Errorf("Contains didn't match: %v != %v", test.ok, ok)
			}
		})
	}
}

func TestContainsFunc(t *testing.T) {
	tests := []struct {
		title string
		s     UserSlice
		f     func(User) bool
		ok    bool
	}{
		{title: "ok", s: UserSlice{ID: []int{1, 2, 3}, Name: []string{"Alice", "Bob", "Charlie"}}, f: func(u User) bool { return u.ID == 2 }, ok: true},
		{title: "ng", s: UserSlice{ID: []int{1, 2, 3}, Name: []string{"Alice", "Bob", "Charlie"}}, f: func(u User) bool { return u.ID == 4 }, ok: false},
	}

	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			ok := ContainsFunc(test.s, test.f)
			if ok != test.ok {
				t.Errorf("ContainsFunc didn't match: %v != %v", test.ok, ok)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	tests := []struct {
		title  string
		s      UserSlice
		i      int
		j      int
		result UserSlice
		panics bool
	}{
		{
			title: "empty",
		},
		{
			title:  "ok",
			s:      UserSlice{ID: []int{1, 2, 3, 4, 5}, Name: []string{"Alice", "Bob", "Charlie", "Dan", "Eve"}},
			i:      2,
			j:      4,
			result: UserSlice{ID: []int{1, 2, 5}, Name: []string{"Alice", "Bob", "Eve"}},
		},
		{
			title:  "out of range i",
			s:      UserSlice{ID: []int{1, 2, 3, 4, 5}, Name: []string{"Alice", "Bob", "Charlie", "Dan", "Eve"}},
			i:      -1,
			j:      4,
			panics: true,
		},
		{
			title:  "out of range j",
			s:      UserSlice{ID: []int{1, 2, 3, 4, 5}, Name: []string{"Alice", "Bob", "Charlie", "Dan", "Eve"}},
			i:      2,
			j:      123,
			panics: true,
		},
	}

	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			defer func() {
				r := recover()
				if (r != nil) != test.panics {
					t.Errorf("Delete didn't panic: %v", test.panics)
				}
			}()
			s := Delete(test.s, test.i, test.j)
			if !reflect.DeepEqual(s, test.result) {
				t.Errorf("Delete didn't match: %v != %v", test.result, s)
			}
		})
	}
}

func TestDeleteFunc(t *testing.T) {
	tests := []struct {
		title  string
		s      UserSlice
		f      func(User) bool
		result UserSlice
		panics bool
	}{
		{
			title: "empty",
		},
		{
			title:  "ok",
			s:      UserSlice{ID: []int{1, 2, 3, 4, 5}, Name: []string{"Alice", "Bob", "Charlie", "Dan", "Eve"}},
			f:      func(u User) bool { return slices.Contains([]int{3, 4}, u.ID) },
			result: UserSlice{ID: []int{1, 2, 5}, Name: []string{"Alice", "Bob", "Eve"}},
		},
	}

	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			defer func() {
				r := recover()
				if (r != nil) != test.panics {
					t.Errorf("DeleteFunc didn't panic: %v", test.panics)
				}
			}()
			s := DeleteFunc(test.s, test.f)
			if !reflect.DeepEqual(s, test.result) {
				t.Errorf("DeleteFunc didn't match: %v != %v", test.result, s)
			}
		})
	}
}

func TestEqual(t *testing.T) {
	tests := []struct {
		title  string
		s1, s2 UserSlice
		ok     bool
	}{
		{
			title: "empty",
			ok:    true,
		},
		{
			title: "same",
			s1:    UserSlice{ID: []int{1, 2, 3}, Name: []string{"Alice", "Bob", "Charlie"}},
			s2:    UserSlice{ID: []int{1, 2, 3}, Name: []string{"Alice", "Bob", "Charlie"}},
			ok:    true,
		},
		{
			title: "different length",
			s1:    UserSlice{ID: []int{1, 2, 3}, Name: []string{"Alice", "Bob", "Charlie"}},
			s2:    UserSlice{ID: []int{1, 3}, Name: []string{"Alice", "Charlie"}},
			ok:    false,
		},
		{
			title: "different element",
			s1:    UserSlice{ID: []int{1, 2, 3}, Name: []string{"Alice", "Bob", "Charlie"}},
			s2:    UserSlice{ID: []int{1, 4, 3}, Name: []string{"Alice", "Don", "Charlie"}},
			ok:    false,
		},
	}

	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			ok := Equal(test.s1, test.s2)
			if ok != test.ok {
				t.Errorf("Equal didn't match: %v != %v", test.ok, ok)
			}
		})
	}
}

func TestEqualFunc(t *testing.T) {
	tests := []struct {
		title  string
		s1, s2 UserSlice
		f      func(User, User) bool
		ok     bool
	}{
		{
			title: "empty",
			ok:    true,
		},
		{
			title: "true",
			s1:    UserSlice{ID: []int{1, 2, 3}, Name: []string{"Alice", "Bob", "Charlie"}},
			s2:    UserSlice{ID: []int{1, 2, 3}, Name: []string{"Alice", "Bob", "Charlie"}},
			f: func(User, User) bool {
				return true
			},
			ok: true,
		},
		{
			title: "false",
			s1:    UserSlice{ID: []int{1, 2, 3}, Name: []string{"Alice", "Bob", "Charlie"}},
			s2:    UserSlice{ID: []int{1, 2, 3}, Name: []string{"Alice", "Bob", "Charlie"}},
			f: func(User, User) bool {
				return false
			},
			ok: false,
		},
		{
			title: "different length",
			s1:    UserSlice{ID: []int{1, 2, 3}, Name: []string{"Alice", "Bob", "Charlie"}},
			s2:    UserSlice{ID: []int{1, 3}, Name: []string{"Alice", "Charlie"}},
			ok:    false,
		},
	}

	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			ok := EqualFunc(test.s1, test.s2, test.f)
			if ok != test.ok {
				t.Errorf("EqualFunc didn't match: %v != %v", test.ok, ok)
			}
		})
	}
}

func TestGrow(t *testing.T) {
	tests := []struct {
		title  string
		s      UserSlice
		n      int
		l      int
		c      int
		panics bool
	}{
		{
			title: "empty",
		},
		{
			title: "positive",
			s:     UserSlice{ID: []int{1, 2, 3}, Name: []string{"Alice", "Bob", "Charlie"}},
			n:     4,
			l:     3,
			c:     7,
		},
		{
			title:  "negative",
			s:      UserSlice{ID: []int{1, 2, 3}, Name: []string{"Alice", "Bob", "Charlie"}},
			n:      -4,
			panics: true,
		},
	}

	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			defer func() {
				r := recover()
				if (r != nil) != test.panics {
					t.Errorf("Grow didn't panic: %v", test.panics)
				}
			}()

			s := Grow(test.s, test.n)
			if s.Len() != test.l {
				t.Errorf("Len didn't match: %v != %v", test.l, s.Len())
			}
			if s.Cap() != test.c {
				t.Errorf("Cap didn't match: %v != %v", test.c, s.Cap())
			}
		})
	}
}

func TestIndex(t *testing.T) {
	tests := []struct {
		title string
		s     UserSlice
		e     User
		i     int
	}{
		{
			title: "empty",
			i:     -1,
		},
		{
			title: "found",
			s:     UserSlice{ID: []int{1, 2, 3}, Name: []string{"Alice", "Bob", "Charlie"}},
			e:     User{ID: 2, Name: "Bob"},
			i:     1,
		},
		{
			title: "not found",
			s:     UserSlice{ID: []int{1, 2, 3}, Name: []string{"Alice", "Bob", "Charlie"}},
			e:     User{ID: 4, Name: "Dan"},
			i:     -1,
		},
	}

	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			i := Index(test.s, test.e)
			if i != test.i {
				t.Errorf("Index didn't match: %v != %v", test.i, i)
			}
		})
	}
}

func TestIndexFunc(t *testing.T) {
	tests := []struct {
		title string
		s     UserSlice
		f     func(User) bool
		i     int
	}{
		{
			title: "empty",
			i:     -1,
		},
		{
			title: "found",
			s:     UserSlice{ID: []int{1, 2, 3}, Name: []string{"Alice", "Bob", "Charlie"}},
			f: func(u User) bool {
				return u.ID == 2
			},
			i: 1,
		},
		{
			title: "not found",
			s:     UserSlice{ID: []int{1, 2, 3}, Name: []string{"Alice", "Bob", "Charlie"}},
			f: func(u User) bool {
				return false
			},
			i: -1,
		},
	}

	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			i := IndexFunc(test.s, test.f)
			if i != test.i {
				t.Errorf("IndexFunc didn't match: %v != %v", test.i, i)
			}
		})
	}
}

func TestInsert(t *testing.T) {
	tests := []struct {
		title  string
		s      UserSlice
		i      int
		es     []User
		result UserSlice
	}{
		{
			title: "empty",
		},
		{
			title: "front",
			s: UserSlice{
				ID:   []int{1, 2, 3},
				Name: []string{"Alice", "Bob", "Charlie"},
			},
			i: 0,
			es: []User{
				{ID: 4, Name: "Dan"},
				{ID: 5, Name: "Eve"},
			},
			result: UserSlice{
				ID:   []int{4, 5, 1, 2, 3},
				Name: []string{"Dan", "Eve", "Alice", "Bob", "Charlie"},
			},
		},
		{
			title: "middle",
			s: UserSlice{
				ID:   []int{1, 2, 3},
				Name: []string{"Alice", "Bob", "Charlie"},
			},
			i: 2,
			es: []User{
				{ID: 4, Name: "Dan"},
				{ID: 5, Name: "Eve"},
			},
			result: UserSlice{
				ID:   []int{1, 2, 4, 5, 3},
				Name: []string{"Alice", "Bob", "Dan", "Eve", "Charlie"},
			},
		},
		{
			title: "back",
			s: UserSlice{
				ID:   []int{1, 2, 3},
				Name: []string{"Alice", "Bob", "Charlie"},
			},
			i: 3,
			es: []User{
				{ID: 4, Name: "Dan"},
				{ID: 5, Name: "Eve"},
			},
			result: UserSlice{
				ID:   []int{1, 2, 3, 4, 5},
				Name: []string{"Alice", "Bob", "Charlie", "Dan", "Eve"},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			result := Insert(test.s, test.i, test.es...)
			if !reflect.DeepEqual(result, test.result) {
				t.Errorf("Insertion didn't match: %v != %v", result, test.result)
			}
		})
	}
}

func TestIsSortedFunc(t *testing.T) {
	tests := []struct {
		title string
		s     UserSlice
		cmp   func(User, User) int
		ok    bool
	}{
		{
			title: "empty",
			ok:    true,
		},
		{
			title: "sorted",
			s: UserSlice{
				ID:   []int{1, 2, 3},
				Name: []string{"Alice", "Bob", "Charlie"},
			},
			cmp: func(u1, u2 User) int {
				return u1.ID - u2.ID
			},
			ok: true,
		},
		{
			title: "unsorted",
			s: UserSlice{
				ID:   []int{1, 3, 2},
				Name: []string{"Alice", "Charlie", "Bob"},
			},
			cmp: func(u1, u2 User) int {
				return u1.ID - u2.ID
			},
			ok: false,
		},
	}

	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			ok := IsSortedFunc(test.s, test.cmp)
			if ok != test.ok {
				t.Errorf("IsSortedFunc didn't match: %v != %v", ok, test.ok)
			}
		})
	}
}

func TestMaxFunc(t *testing.T) {
	tests := []struct {
		title  string
		s      UserSlice
		cmp    func(User, User) int
		result User
		panics bool
	}{
		{
			title:  "empty",
			panics: true,
		},
		{
			title: "ok",
			s: UserSlice{
				ID:   []int{1, 2, 3},
				Name: []string{"Alice", "Bob", "Charlie"},
			},
			cmp: func(u1, u2 User) int {
				return u1.ID - u2.ID
			},
			result: User{ID: 3, Name: "Charlie"},
		},
	}

	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			defer func() {
				r := recover()
				if (r != nil) != test.panics {
					t.Errorf("MaxFunc didn't match: %v != %v", r, test.panics)
				}
			}()

			result := MaxFunc(test.s, test.cmp)
			if !reflect.DeepEqual(result, test.result) {
				t.Errorf("MaxFunc didn't match: %v != %v", result, test.result)
			}
		})
	}
}

func TestMinFunc(t *testing.T) {
	tests := []struct {
		title  string
		s      UserSlice
		cmp    func(User, User) int
		result User
		panics bool
	}{
		{
			title:  "empty",
			panics: true,
		},
		{
			title: "ok",
			s: UserSlice{
				ID:   []int{2, 1, 3},
				Name: []string{"Bob", "Alice", "Charlie"},
			},
			cmp: func(u1, u2 User) int {
				return u1.ID - u2.ID
			},
			result: User{ID: 1, Name: "Alice"},
		},
	}

	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			defer func() {
				r := recover()
				if (r != nil) != test.panics {
					t.Errorf("MinFunc didn't match: %v != %v", r, test.panics)
				}
			}()

			result := MinFunc(test.s, test.cmp)
			if !reflect.DeepEqual(result, test.result) {
				t.Errorf("MinFunc didn't match: %v != %v", result, test.result)
			}
		})
	}
}

func TestRepeat(t *testing.T) {
	tests := []struct {
		title  string
		s      UserSlice
		n      int
		result UserSlice
		panics bool
	}{
		{
			title: "empty",
		},
		{
			title: "ok",
			s: UserSlice{
				ID:   []int{1, 2, 3},
				Name: []string{"Alice", "Bob", "Charlie"},
			},
			n: 3,
			result: UserSlice{
				ID:   []int{1, 2, 3, 1, 2, 3, 1, 2, 3},
				Name: []string{"Alice", "Bob", "Charlie", "Alice", "Bob", "Charlie", "Alice", "Bob", "Charlie"},
			},
		},
		{
			title: "negative",
			s: UserSlice{
				ID:   []int{1, 2, 3},
				Name: []string{"Alice", "Bob", "Charlie"},
			},
			n:      -1,
			panics: true,
		},
		{
			title: "too long",
			s: UserSlice{
				fakeLen: math.MaxInt / 3,
			},
			n:      4,
			panics: true,
		},
	}

	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			defer func() {
				r := recover()
				if (r != nil) != test.panics {
					t.Errorf("Repeat didn't match: %v != %v", r, test.panics)
				}
			}()

			result := Repeat(test.s, test.n)
			if !reflect.DeepEqual(result, test.result) {
				t.Errorf("Repeat didn't match: %v != %v", result, test.result)
			}
		})
	}
}

func TestReplace(t *testing.T) {
	tests := []struct {
		title  string
		s      UserSlice
		i, j   int
		vs     []User
		result UserSlice
	}{
		{
			title: "empty",
		},
		{
			title: "expand",
			s:     UserSlice{ID: []int{1, 2, 3}, Name: []string{"Alice", "Bob", "Charlie"}},
			i:     1,
			j:     2,
			vs: []User{
				{ID: 4, Name: "Dan"},
				{ID: 5, Name: "Eve"},
			},
			result: UserSlice{ID: []int{1, 4, 5, 3}, Name: []string{"Alice", "Dan", "Eve", "Charlie"}},
		},
	}

	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			result := Replace(test.s, test.i, test.j, test.vs...)
			if !reflect.DeepEqual(result, test.result) {
				t.Errorf("Replaced didn't match: %v != %v", result, test.result)
			}
		})
	}
}

func TestReverse(t *testing.T) {
	tests := []struct {
		title  string
		s      UserSlice
		result UserSlice
	}{
		{
			title: "empty",
		},
		{
			title:  "ok",
			s:      UserSlice{ID: []int{1, 2, 3}, Name: []string{"Alice", "Bob", "Charlie"}},
			result: UserSlice{ID: []int{3, 2, 1}, Name: []string{"Charlie", "Bob", "Alice"}},
		},
	}

	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			Reverse(test.s)
			if !reflect.DeepEqual(test.s, test.result) {
				t.Errorf("Reverse didn't match: %v != %v", test.s, test.result)
			}
		})
	}
}

func TestSortFunc(t *testing.T) {
	tests := []struct {
		title  string
		s      UserSlice
		cmp    func(User, User) int
		result UserSlice
	}{
		{
			title: "empty",
		},
		{
			title: "ok",
			s:     UserSlice{ID: []int{1, 2, 3}, Name: []string{"Alice", "Bob", "Charlie"}},
			cmp: func(u1, u2 User) int {
				return u2.ID - u1.ID
			},
			result: UserSlice{ID: []int{3, 2, 1}, Name: []string{"Charlie", "Bob", "Alice"}},
		},
	}

	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			SortFunc(test.s, test.cmp)
			if !reflect.DeepEqual(test.s, test.result) {
				t.Errorf("SortFunc didn't match: %v != %v", test.s, test.result)
			}
		})
	}
}

func TestSortStableFunc(t *testing.T) {
	tests := []struct {
		title  string
		s      UserSlice
		cmp    func(User, User) int
		result UserSlice
	}{
		{
			title: "empty",
		},
		{
			title: "ok",
			s:     UserSlice{ID: []int{1, 2, 3}, Name: []string{"Alice", "Bob", "Charlie"}},
			cmp: func(u1, u2 User) int {
				return u2.ID - u1.ID
			},
			result: UserSlice{ID: []int{3, 2, 1}, Name: []string{"Charlie", "Bob", "Alice"}},
		},
	}

	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			SortStableFunc(test.s, test.cmp)
			if !reflect.DeepEqual(test.s, test.result) {
				t.Errorf("SortStableFunc didn't match: %v != %v", test.s, test.result)
			}
		})
	}
}

func TestSortedFunc(t *testing.T) {
	tests := []struct {
		title  string
		seq    iter.Seq[User]
		cmp    func(User, User) int
		result UserSlice
	}{
		{
			title: "empty",
			seq:   slices.Values([]User{}),
		},
		{
			title: "ok",
			seq: slices.Values([]User{
				{ID: 1, Name: "Alice"},
				{ID: 2, Name: "Bob"},
				{ID: 3, Name: "Charlie"},
			}),
			cmp: func(u1, u2 User) int {
				return u2.ID - u1.ID
			},
			result: UserSlice{ID: []int{3, 2, 1}, Name: []string{"Charlie", "Bob", "Alice"}},
		},
	}

	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			result := SortedFunc[UserSlice](test.seq, test.cmp)
			if !reflect.DeepEqual(result, test.result) {
				t.Errorf("SortedFunc didn't match: %v != %v", result, test.result)
			}
		})
	}
}

func TestSortedStableFunc(t *testing.T) {
	tests := []struct {
		title  string
		seq    iter.Seq[User]
		cmp    func(User, User) int
		result UserSlice
	}{
		{
			title: "empty",
			seq:   slices.Values([]User{}),
		},
		{
			title: "ok",
			seq: slices.Values([]User{
				{ID: 1, Name: "Alice"},
				{ID: 2, Name: "Bob"},
				{ID: 3, Name: "Charlie"},
			}),
			cmp: func(u1, u2 User) int {
				return u2.ID - u1.ID
			},
			result: UserSlice{ID: []int{3, 2, 1}, Name: []string{"Charlie", "Bob", "Alice"}},
		},
	}

	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			result := SortedStableFunc[UserSlice](test.seq, test.cmp)
			if !reflect.DeepEqual(result, test.result) {
				t.Errorf("SortedStableFunc didn't match: %v != %v", result, test.result)
			}
		})
	}
}

func TestValues(t *testing.T) {
	tests := []struct {
		title  string
		s      UserSlice
		n      int
		result []User
	}{
		{
			title: "empty",
			n:     -1,
		},
		{
			title: "ok",
			s:     UserSlice{ID: []int{1, 2, 3, 4, 5}, Name: []string{"Alice", "Bob", "Charlie", "Dan", "Eve"}},
			n:     3,
			result: []User{
				{ID: 1, Name: "Alice"},
				{ID: 2, Name: "Bob"},
				{ID: 3, Name: "Charlie"},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			result := slices.Collect(take(Values(test.s), test.n))
			if !reflect.DeepEqual(result, test.result) {
				t.Errorf("Values didn't match: %v != %v", result, test.result)
			}
		})
	}
}
