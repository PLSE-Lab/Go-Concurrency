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

// We should have a consistent return type with information on what
// the match has found
func matchMakeCall(x *ast.CallExpr, v *Visitor, n ast.Node) {
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
			} else if len(x.Args) == 2 {
				t, ok := x.Args[0].(*ast.ChanType)
				if ok {
					tname, ok := t.Value.(*ast.Ident)
					if ok {
						bsize, ok := x.Args[1].(*ast.BasicLit)
						if ok {
							if bsize.Kind == token.INT {
								fmt.Printf("Found a channel of type %s with literal buffer size %s at %s\n", tname.Name, bsize.Value, v.fset.Position(n.Pos()))
							} else {
								fmt.Printf("Found a channel of type %s with buffer size %s at %s\n", tname.Name, bsize.Value, v.fset.Position(n.Pos()))
							}
						} else {
							fmt.Printf("Found a channel of type %s with computed buffer size %s at %s\n", tname.Name, x.Args[1], v.fset.Position(n.Pos()))
						}
					}
				}
			}
		}
	}
}

func matchSendStmt(x *ast.SendStmt, v *Visitor, n ast.Node) {
	id, ok := x.Chan.(*ast.Ident)
	if ok {
		fmt.Printf("Found a send to channel %s for value %s\n", id.Name, x.Value)
	}
}

func matchUnaryExpr(x *ast.UnaryExpr, v *Visitor, n ast.Node) {
	if x.Op == token.ARROW {
		id, ok := x.X.(*ast.Ident)
		if ok {
			fmt.Printf("Found a read of channel %s\n", id.Name)
		} else {
			fmt.Printf("Found a read of channel from expr %s\n", x.X)
		}
	}
}

func (v *Visitor) Visit(n ast.Node) ast.Visitor {
	if n == nil {
		return nil
	}

	switch x := n.(type) {
	case *ast.CallExpr:
		matchMakeCall(x, v, n)
	case *ast.SendStmt:
		matchSendStmt(x, v, n)
	case *ast.UnaryExpr:
		matchUnaryExpr(x, v, n)
	}
	return v
}
