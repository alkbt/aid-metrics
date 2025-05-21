package main

import (
	"fmt"

	"github.com/alkbt/testmodule/pkg3"
)

func main() {
	s3 := pkg3.NewStruct3()
	result := s3.Run()
	fmt.Println("Result:", result)
}
