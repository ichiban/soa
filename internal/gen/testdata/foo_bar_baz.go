package testdata

import (
	t "time"
)

type NonStruct int

type Foo struct {
	ID        int
	CreatedAt t.Time
	UpdatedAt t.Time
}

type Bar struct {
	ID        int
	CreatedAt t.Time
	UpdatedAt t.Time
}

type Baz struct {
	ID        int
	CreatedAt t.Time
	UpdatedAt t.Time
}
