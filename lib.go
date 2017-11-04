package main

// #cgo LDFLAGS: -Wl,-U,_InsertHighlight,-U,_Errored
// #include "golang.h"
import "C"

import (
	"go/ast"
	"go/token"

	"github.com/richardsamuels/color_goded_ast/walker"
)

type Callback = C.struct_Callback

func (c Callback) InsertHighlight(group, name string, pos token.Position) {
	C.InsertHighlight(c, C.CString(group), C.int(pos.Line), C.int(pos.Column), C.CString(name))
}

func (c Callback) Error(msg string) {
	C.Errored(c, C.CString(msg))
}

//export GoGetTokens
func GoGetTokens(fname_c *C.char, c Callback) bool {
	defer func() {
		if r := recover(); r != nil {
			msg := ""
			if s, ok := r.(error); ok {
				msg = s.Error()
			}
			c.Error(msg)
		}
	}()

	fname := C.GoString(fname_c)

	w := walker.New(fname, c.InsertHighlight, c.Error)

	if w.Err == nil {
		//pp.Println(w.File)
		//pp.Println(tree.Imports)
		ast.Walk(w, w.File)
	} else {
		c.Error(w.Err.Error())
		return true
	}

	return w.Err == nil
}

func main() {}
