package main

import (
	"fmt"

	"github.com/k0kubun/pp"
)

func (t *T) test() {
	fmt.Println("hi")
}

func (t t) a() {
	pp.Println(t)
}
