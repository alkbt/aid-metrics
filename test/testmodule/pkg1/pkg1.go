package pkg1

import (
	"fmt"
)

// Interface1 is an interface for testing abstractness
type Interface1 interface {
	Method1() string
	Method2() int
}

// Struct1 is a concrete type
type Struct1 struct {
	Field1 string
	Field2 int
}

// Method for Struct1
func (s *Struct1) DoSomething() {
	fmt.Println("Doing something from pkg1")
}
