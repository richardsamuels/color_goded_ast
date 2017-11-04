package walker

import (
	"go/ast"
	"go/token"
	"strings"
)

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
				w.field(f)
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
		// TODO oh this is so wrong. Need to parse packages as well to
		// determine their package names
		if v.Name != nil {
			w.Tokenise("Namespace", v.Name.Name, v.Name.NamePos)
			w.packages[v.Name.Name] = true
		} else {
			w.Tokenise("String", v.Path.Value, v.Path.ValuePos)

			pos := v.Path.ValuePos + token.Pos(1)
			path := strings.Split(strings.TrimLeft(strings.TrimRight(v.Path.Value, "\""), "\""), "/")
			last := len(path) - 1
			for i := range path {
				if i != last {
					pos += token.Pos(len(path[i]) + 1)
				}
			}

			w.Tokenise("Namespace", path[last], pos)
			w.packages[path[last]] = true
		}

	case *ast.CompositeLit:
		if t, ok := v.Type.(*ast.Ident); ok {
			w.Tokenise("Type", t.Name, t.NamePos)
		}

	case *ast.KeyValueExpr:
		if k, ok := v.Key.(*ast.Ident); ok {
			w.Tokenise("Member", k.Name, k.NamePos)
		}
		w.expr(&v.Value)

	case *ast.RangeStmt:
		w.Tokenise("Operator", v.Tok.String(), v.TokPos)
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

func (w *Walker) selectorExpr(v *ast.SelectorExpr) {
	if x, ok := v.X.(*ast.Ident); ok {
		if x.Obj != nil {
			w.Tokenise(objKindToGroup(x.Obj.Kind), x.Name, x.NamePos)
		} else {
			w.ident(x)

		}
	}

	w.Tokenise("Member", v.Sel.Name, v.Sel.NamePos)
}
