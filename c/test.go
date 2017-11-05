package main

import (
	"fmt"
	"github.com/k0kubun/pp"
)

type T struct {
	pp         int
	noconflict int
}
type t = int

func (t *T) a() {
	pp.Println(t.field)
}
func (t *t) a() {
	pp.Println(t)
}

func main() {
	dsfsdf := 5
	dsfsdf = 2
	t := T{
		pp:         4,
		noconflict: dsfsdf,
	}
	fmt.Println("vim-go")

	if fmt := dsfsdf + 6; fmt == 8 {
		fmt := T{}
		fmt.a()
	}

	x := []T{}
	pp.Println(x)

	if d, ok := x.(T); ok {
		pp.Println(d)
	}

	m := map[int]int{}
	five := 5
	m[five] = 5

	for k, v := range m {
		pp.Println(k)
		pp.Println(v)
	}
	t.pp = five

	return dsfsdf
}
