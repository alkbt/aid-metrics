package pkg

// Abstract types (interfaces)
type Interface1 interface {
	Method1()
	Method2()
}

type Interface2 interface {
	Method3()
}

// Concrete types (structs)
type Struct1 struct {
	Field1 string
	Field2 int
}

type Struct2 struct {
	Field3 bool
}

// Methods (should not count as standalone functions)
func (s *Struct1) Method1() {
	// Implementation
}

func (s *Struct1) Method2() {
	// Implementation
}

func (s *Struct2) Method3() {
	// Implementation
}

// Standalone functions (should count as concrete types)
func Function1() string {
	return "function1"
}

func Function2() int {
	return 42
}

func Function3() bool {
	return true
}

func Function4() {
	// No return
}

// Type alias (should count as a concrete type)
type AliasType = string
