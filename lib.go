package main

// #cgo darwin LDFLAGS: -Wl,-U,_InsertHighlight,-U,_Errored
// #cgo linux LDFLAGS: -Wl,-u,_InsertHighlight,-u,_Errored
// #include "golang.h"
import "C"

import (
	"go/ast"
	"go/token"
	"sync"

	"github.com/richardsamuels/color_goded_ast/walker"
)

var doWorkMutex sync.RWMutex = sync.RWMutex{}
var doWork bool = true

type Callback = C.struct_Callback

func (c Callback) InsertHighlight(group, name string, pos token.Position) {
	doWorkMutex.RLock()
	defer doWorkMutex.RUnlock()

	if doWork {
		C.InsertHighlight(c, C.CString(group), C.int(pos.Line), C.int(pos.Column), C.CString(name))
	}
}

func (c Callback) Error(msg string) {
	doWorkMutex.RLock()
	defer doWorkMutex.RUnlock()

	if doWork {
		C.Errored(c, C.CString(msg))
	}
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

//export Exit
func Exit() {
	doWorkMutex.Lock()
	defer doWorkMutex.Unlock()

	doWork = false
}

func main() {}
