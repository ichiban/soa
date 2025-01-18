package main

import (
	"fmt"

	"soa"
)

// User is an example struct. We're going to make an SoA slice for it.
type User struct {
	// exported fields
	ID   int
	Name string

	// unexported fields
	deleted bool
}

// To generate an SoA slice, run `go generate ./...`.
// It'll generate UserSlice in *_soa.go from the declaration of User in this file.
//go:generate go run ../../cmd/soagen

func main() {
	// Now you can use UserSlice to store User.

	// To create a new SoA slice, you can call `soa.Make()`.
	s := soa.Make[UserSlice](0, 4)

	// To append Users, you can call `soa.Append()`.
	s = soa.Append(s, User{ID: 1, Name: "Alice"})
	s = soa.Append(s, User{ID: 2, Name: "Bob", deleted: true})
	s = soa.Append(s, User{ID: 3, Name: "Charlie"})
	s = soa.Append(s, User{ID: 4, Name: "Dave", deleted: true})

	// To iterate over the SoA slice, you can call `soa.All()`.
	for i, u := range soa.All(s) {
		fmt.Println(i, u)
	}
}
