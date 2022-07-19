package main

/*
 Note: This is based on code from the following blog:
 https://eli.thegreenplace.net/2021/rewriting-go-source-code-with-ast-tooling/
*/

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
)

func main() {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "sample/goroutines-with-select.go", nil, 0)
	if err != nil {
		log.Fatal(err)
	}

	visitor := &Visitor{fset: fset}
	ast.Walk(visitor, file)
}

type Visitor struct {
	fset *token.FileSet
}

func (v *Visitor) Visit(n ast.Node) ast.Visitor {
	if n == nil {
		return nil
	}

	switch x := n.(type) {
	case *ast.CallExpr:
		id, ok := x.Fun.(*ast.Ident)
		if ok {
			if id.Name == "make" {
				if len(x.Args) == 1 {
					t, ok := x.Args[0].(*ast.ChanType)
					if ok {
						tname, ok := t.Value.(*ast.Ident)
						if ok {
							fmt.Printf("Found a channel of type %s at %s\n", tname.Name, v.fset.Position(n.Pos()))
						}
					}
				}
			}
		}
	}
	return v
}
