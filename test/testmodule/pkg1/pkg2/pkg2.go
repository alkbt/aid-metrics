package pkg2

import (
	"testmodule/pkg1"
)

// Interface2 is an interface for testing abstractness
type Interface2 interface {
	DoSomethingElse() error
}

// Struct2 depends on pkg1
type Struct2 struct {
	Pkg1Struct *pkg1.Struct1
}

// Method for Struct2
func (s *Struct2) Process() string {
	s.Pkg1Struct.DoSomething()
	return "Processed in pkg2"
}
