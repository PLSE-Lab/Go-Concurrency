package main

/*
 Note: This is based on code from the following blog:
 https://eli.thegreenplace.net/2021/rewriting-go-source-code-with-ast-tooling/
*/

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"log"
	"path/filepath"
)

func main() {
	var filePath string
	flag.StringVar(&filePath, "filePath", "", "The file to be processed")

	var dirPath string
	flag.StringVar(&dirPath, "dirPath", "", "The directory to be processed")

	flag.Parse()

	if dirPath != "" {
		fmt.Printf("Processing all go files in directory %s\n", dirPath)
		processDir(dirPath)
	} else if filePath != "" {
		fmt.Printf("Processing file %s\n", filePath)
		processFile(filePath)
	} else {
		fmt.Printf("No file or directory given\n")
	}

}

func processDir(dirPath string) {
	var err = filepath.Walk(dirPath, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("Encountered an error accessing path %q: %v\n", path, err)
			return err
		} else {
			if filepath.Ext(path) == ".go" {
				fmt.Printf("Processing file %s\n", path)
				processFile(path)
				return nil
			} else {
				return nil
			}
		}
	})

	if err != nil {
		fmt.Printf("Encountered an error walking the directory tree %q: %v", dirPath, err)
	}
}

func processFile(filePath string) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filePath, nil, 0)
	if err != nil {
		log.Print("Could not process file %s", filePath)
		log.Print(err)
	} else {
		visitor := &Visitor{fset: fset}
		ast.Walk(visitor, file)
	}
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

func matchWaitGroupDecl(x *ast.GenDecl, v *Visitor, n ast.Node) {
	spec, ok := x.Specs[0].(*ast.ValueSpec)
	if ok {
		id := spec.Names[0]
		t, ok := spec.Type.(*ast.SelectorExpr)
		if ok {
			tsel, ok := t.X.(*ast.Ident)
			if ok {
				if tsel.Name == "sync" && t.Sel.Name == "WaitGroup" {
					fmt.Printf("Found declaration of waitgroup %s\n", id.Name)
				}
			}
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
	case *ast.GenDecl:
		matchWaitGroupDecl(x, v, n)
	}
	return v
}
