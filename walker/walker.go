package walker

import (
	"errors"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"path/filepath"

	"github.com/k0kubun/pp"
)

func ShutUp2() {
	pp.Println("skhjdfsjdlkfdjks")
}

type Walker struct {
	File      *ast.File
	Package   *ast.Package
	Err       error
	fset      *token.FileSet
	uniquePos map[token.Pos]bool
	packages  map[string]*types.Package
	Insert    func(group, name string, pos token.Position)
	Error     func(msg string)
	imp       types.ImporterFrom
	dir       string
}

func New(fname string, insert func(group, name string, pos token.Position), errf func(msg string)) *Walker {
	w := &Walker{
		fset:      token.NewFileSet(),
		uniquePos: map[token.Pos]bool{},
		packages:  map[string]*types.Package{},
		Insert:    insert,
		Error:     errf,
		imp:       importer.For("source", nil).(types.ImporterFrom),
	}
	w.dir = filepath.Dir(fname)

	p, err := parser.ParseDir(w.fset, w.dir, nil, parser.AllErrors)
	if err != nil {
		w.Err = err
	} else {
		for _, v := range p {
			if file, ok := v.Files[fname]; ok {
				w.File = file
				w.Package = v
				break
			}
		}
	}

	if w.File == nil {
		w.Err = errors.New("parsed file missing from package.")
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
