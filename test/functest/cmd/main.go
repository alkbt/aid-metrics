package main

import (
	"fmt"
)

// Standalone functions in main package
func StandaloneFunc1() string {
	return "standalone1"
}

func StandaloneFunc2() int {
	return 100
}

// Multiple interfaces
type MainInterface1 interface {
	MainMethod1()
}

type MainInterface2 interface {
	MainMethod2()
}

// Several structs
type MainStruct1 struct {
	Field1 string
}

type MainStruct2 struct {
	Field2 int
}

func (s *MainStruct1) MainMethod1() {
	fmt.Println("MainMethod1")
}

func (s *MainStruct2) MainMethod2() {
	fmt.Println("MainMethod2")
}

func main() {
	s1 := &MainStruct1{Field1: "test"}
	s2 := &MainStruct2{Field2: 42}

	s1.MainMethod1()
	s2.MainMethod2()

	fmt.Println(StandaloneFunc1())
	fmt.Println(StandaloneFunc2())
}
