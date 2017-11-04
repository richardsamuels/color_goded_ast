package walker

import (
	"go/ast"
	"go/parser"
	"go/token"
)

type Walker struct {
	File      *ast.File
	Err       error
	fset      *token.FileSet
	uniquePos map[token.Pos]bool
	packages  map[string]bool
	Insert    func(group, name string, pos token.Position)
	Error     func(msg string)
}

func New(fname string, insert func(group, name string, pos token.Position), errf func(msg string)) *Walker {
	w := &Walker{
		fset:      token.NewFileSet(),
		uniquePos: map[token.Pos]bool{},
		packages:  map[string]bool{},
		Insert:    insert,
		Error:     errf,
	}

	w.File, w.Err = parser.ParseFile(w.fset, fname, nil, parser.AllErrors)

	if w.Err == nil {
		w.Tokenise("Namespace", w.File.Name.Name, w.File.Name.NamePos)
	}

	return w
}

func (w Walker) Visit(n ast.Node) ast.Visitor {
	if w.onNode(n) {
		return w
	}

	return nil
}

func (w *Walker) Tokenise(group, name string, rawPos token.Pos) {
	pos := w.fset.Position(rawPos)

	if b, ok := w.uniquePos[rawPos]; ok && b {
		return
	} else {
		w.uniquePos[rawPos] = true
	}
	w.Insert(group, name, pos)
}
