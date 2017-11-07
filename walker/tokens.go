package walker

import (
	"go/ast"
	"go/token"
	"strings"

	"github.com/k0kubun/pp"
)

func ShutUp() {
	pp.Println("skhjdfsjdlkfdjks")
}

func (w *Walker) onObject(o *ast.Object) {
	if o == nil {
		return
	}

	w.Tokenise(objKindToGroup(o.Kind), o.Name, o.Pos())
}

func objKindToGroup(k ast.ObjKind) string {
	hgroup := "Error"
	switch k {
	case ast.Pkg:
		hgroup = "Namespace"

	case ast.Con:
		hgroup = "Variable"

	case ast.Typ:
		hgroup = "Type"

	case ast.Var:
		hgroup = "Variable"

	case ast.Fun:
		hgroup = "Function"

	case ast.Lbl:
		hgroup = "TODO"
	}
	return hgroup
}

func (w *Walker) badNode(n ast.Node) {
	var pos token.Pos
	toLen := 0

	switch v := n.(type) {
	case *ast.BadDecl:
		pos = v.From
		toLen = int(v.To - v.From)

	case *ast.BadExpr:
		pos = v.From
		toLen = int(v.To - v.From)

	case *ast.BadStmt:
		pos = v.From
		toLen = int(v.To - v.From)

	}

	name := ""
	for i := 0; i < toLen; i++ {
		name += " "
	}
	w.Tokenise("Error", name, pos)
}

func (w *Walker) onNode(n ast.Node) bool {
	if n == nil {
		return false
	}

	switch v := n.(type) {
	case *ast.BadDecl, *ast.BadExpr, *ast.BadStmt:
		w.badNode(n)

	case *ast.FuncDecl:
		w.Tokenise("Function", v.Name.Name, v.Name.Pos())

		if v.Recv != nil {
			for _, f := range v.Recv.List {
				w.field2(f)
			}
		}

	case *ast.CallExpr:
		if x, ok := v.Fun.(*ast.Ident); ok {
			w.Tokenise("Function", x.Name, x.NamePos)
		} else {
			w.onNode(v.Fun)
		}
		for i, _ := range v.Args {
			w.expr(&v.Args[i])
		}

	case *ast.SelectorExpr:
		w.selectorExpr(v)

	case *ast.TypeSpec:
		w.Tokenise("Type", v.Name.Name, v.Name.NamePos)

	case *ast.ValueSpec:
		for i := range v.Names {
			w.onObject(v.Names[i].Obj)
		}

	case *ast.AssignStmt:
		for i := range v.Lhs {
			w.expr(&v.Lhs[i])
		}
		w.Tokenise("Operator", v.Tok.String(), v.TokPos)
		for i := range v.Rhs {
			w.expr(&v.Rhs[i])
		}

	case *ast.BasicLit:
		switch v.Kind {
		case token.INT, token.FLOAT, token.IMAG:
			w.Tokenise("Number", v.Value, v.ValuePos)
		}

	case *ast.Field:
		w.field(v)

	case *ast.BinaryExpr:
		w.xexpr(v.X)
		w.Tokenise("Operator", v.Op.String(), v.OpPos)
		w.xexpr(v.Y)

	case *ast.ImportSpec:
		s := strings.TrimRight(strings.TrimLeft(v.Path.Value, "\""), "\"")
		i, err := w.imp.ImportFrom(s, w.dir, 0)
		if err == nil {
			if v.Name == nil {
				w.packages[i.Name()] = i
			} else {
				w.packages[v.Name.Name] = i
				w.Tokenise("Namespace", v.Name.Name, v.Name.NamePos)
			}
		}

	case *ast.CompositeLit:
		if t, ok := v.Type.(*ast.Ident); ok {
			w.Tokenise("Type", t.Name, t.NamePos)
		}

		for i := range v.Elts {
			w.expr(&v.Elts[i])
		}

	case *ast.KeyValueExpr:
		if k, ok := v.Key.(*ast.Ident); ok {
			w.Tokenise("Member", k.Name, k.NamePos)
		}
		w.expr(&v.Value)

	case *ast.RangeStmt:
		w.Tokenise("Operator", v.Tok.String(), v.TokPos)
		w.expr(&v.Key)
		w.expr(&v.Value)
		w.expr(&v.X)

	case *ast.ReturnStmt:
		for i := range v.Results {
			w.expr(&v.Results[i])
		}

	case *ast.IfStmt:
		w.exprIdent(&v.Cond)

	case *ast.ArrayType:
		w.expr(&v.Elt)

	case *ast.TypeAssertExpr:
		w.expr(&v.X)
		w.expr(&v.Type)

	case *ast.IndexExpr:
		w.expr(&v.X)
		w.expr(&v.Index)

	case *ast.SwitchStmt:
		w.expr(&v.Tag)

	case *ast.MapType:
		switch k := v.Key.(type) {
		case *ast.Ident:
			w.Tokenise("Type", k.Name, k.NamePos)

		}
		switch k := v.Value.(type) {
		case *ast.Ident:
			w.Tokenise("Type", k.Name, k.NamePos)

		}

	case *ast.UnaryExpr:
		w.expr(&v.X)
	}

	return true
}

func (w *Walker) exprIdent(v *ast.Expr) {
	if c, ok := (*v).(*ast.Ident); ok {
		w.ident(c)
	}
}

func (w *Walker) expr(v *ast.Expr) {
	if c, ok := (*v).(*ast.Ident); ok {
		w.ident(c)
	} else {
		w.onNode(*v)
	}
}

func (w *Walker) ident(v *ast.Ident) {
	if v.Obj == nil {
		if _, ok := w.packages[v.Name]; ok {
			w.Tokenise("Namespace", v.Name, v.Pos())
		} else {
			for _, f := range w.Package.Files {
				if o := f.Scope.Lookup(v.Name); o != nil {
					w.Tokenise(objKindToGroup(o.Kind), v.Name, v.NamePos)
				}
			}
		}

	} else {
		w.Tokenise(objKindToGroup(v.Obj.Kind), v.Name, v.NamePos)
	}
}

func (w *Walker) xexpr(x interface{}) {
	switch v := x.(type) {
	case *ast.Ident:
		if v.Obj != nil {
			w.Tokenise(objKindToGroup(v.Obj.Kind), v.Name, v.NamePos)
		}
	}

}

func (w *Walker) field(v *ast.Field) {
	for _, name := range v.Names {
		w.onObject(name.Obj)
	}
	switch t := v.Type.(type) {
	case *ast.Ident:
		w.Tokenise("Type", t.Name, t.NamePos)
	case *ast.StarExpr:
		if i, ok := t.X.(*ast.Ident); ok {
			w.ident(i)
		}

	}
}

// TODO: fix this shame
func (w *Walker) field2(v *ast.Field) {

	for _, name := range v.Names {
		w.onObject(name.Obj)
	}
	switch t := v.Type.(type) {
	case *ast.Ident:
		w.Tokenise("Type", t.Name, t.NamePos)
	case *ast.StarExpr:
		if i, ok := t.X.(*ast.Ident); ok {
			w.ident(i)
			if i.Obj == nil {
				w.Tokenise("Type", i.Name, i.NamePos)
			}
		}
	}
}

func (w *Walker) selectorExpr(v *ast.SelectorExpr) {
	if x, ok := v.X.(*ast.Ident); ok {
		if x.Obj != nil {
			w.Tokenise(objKindToGroup(x.Obj.Kind), x.Name, x.NamePos)
		} else {
			w.ident(x)
		}
	}

	// TODO: properly resolve the Selectors; this marks everything as a
	// member, even when more appropriate highlights exist
	w.Tokenise("Member", v.Sel.Name, v.Sel.NamePos)
}
