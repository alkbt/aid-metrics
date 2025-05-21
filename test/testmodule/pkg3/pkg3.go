package pkg3

import (
	"github.com/alkbt/testmodule/pkg1"
	"github.com/alkbt/testmodule/pkg2"
)

// No interfaces here, only concrete types
type Struct3 struct {
	Pkg1Data *pkg1.Struct1
	Pkg2Data *pkg2.Struct2
}

func NewStruct3() *Struct3 {
	pkg1Data := &pkg1.Struct1{
		Field1: "data",
		Field2: 42,
	}

	pkg2Data := &pkg2.Struct2{
		Pkg1Struct: pkg1Data,
	}

	return &Struct3{
		Pkg1Data: pkg1Data,
		Pkg2Data: pkg2Data,
	}
}

func (s *Struct3) Run() string {
	s.Pkg1Data.DoSomething()
	return s.Pkg2Data.Process()
}
