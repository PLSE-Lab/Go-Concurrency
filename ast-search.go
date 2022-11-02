package main

/*
 Note: This is based on code from the following blog:
 https://eli.thegreenplace.net/2021/rewriting-go-source-code-with-ast-tooling/
*/

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
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
		log.Printf("Could not process file %s", filePath)
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
	for i := 0; i < len(x.Specs); i++ {
		spec, ok := x.Specs[i].(*ast.ValueSpec)
		if ok {
			for j := 0; j < len(spec.Names); j++ {
				id := spec.Names[j]
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
	}
}

func matchWaitGroupParamDecl(x *ast.Field, v *Visitor, n ast.Node) {
	for i := 0; i < len(x.Names); i++ {
		fieldName := x.Names[i]
		fieldType, ok := x.Type.(*ast.SelectorExpr)
		if ok {
			tsel, ok := fieldType.X.(*ast.Ident)
			if ok {
				if tsel.Name == "sync" && fieldType.Sel.Name == "WaitGroup" {
					fmt.Printf("Found declaration of waitgroup field %s\n", fieldName.Name)
				}
			}
		}
	}
}

func matchMutexDecl(x *ast.GenDecl, v *Visitor, n ast.Node) {
	for i := 0; i < len(x.Specs); i++ {
		spec, ok := x.Specs[i].(*ast.ValueSpec)
		if ok {
			for j := 0; j < len(spec.Names); j++ {
				id := spec.Names[j]
				t, ok := spec.Type.(*ast.SelectorExpr)
				if ok {
					tsel, ok := t.X.(*ast.Ident)
					if ok {
						if tsel.Name == "sync" && t.Sel.Name == "Mutex" {
							fmt.Printf("Found declaration of mutex %s\n", id.Name)
						}
					}
				}
			}
		}
	}
}

func matchMutexParamDecl(x *ast.Field, v *Visitor, n ast.Node) {
	for i := 0; i < len(x.Names); i++ {
		fieldName := x.Names[i]
		fieldType, ok := x.Type.(*ast.SelectorExpr)
		if ok {
			tsel, ok := fieldType.X.(*ast.Ident)
			if ok {
				if tsel.Name == "sync" && fieldType.Sel.Name == "Mutex" {
					fmt.Printf("Found declaration of mutex field %s\n", fieldName.Name)
				}
			}
		}
	}
}

func matchRWMutexDecl(x *ast.GenDecl, v *Visitor, n ast.Node) {
	for i := 0; i < len(x.Specs); i++ {
		spec, ok := x.Specs[i].(*ast.ValueSpec)
		if ok {
			for j := 0; j < len(spec.Names); j++ {
				id := spec.Names[j]
				t, ok := spec.Type.(*ast.SelectorExpr)
				if ok {
					tsel, ok := t.X.(*ast.Ident)
					if ok {
						if tsel.Name == "sync" && t.Sel.Name == "RWMutex" {
							fmt.Printf("Found declaration of rwmutex %s\n", id.Name)
						}
					}
				}
			}
		}
	}
}

func matchRWMutexParamDecl(x *ast.Field, v *Visitor, n ast.Node) {
	for i := 0; i < len(x.Names); i++ {
		fieldName := x.Names[i]
		fieldType, ok := x.Type.(*ast.SelectorExpr)
		if ok {
			tsel, ok := fieldType.X.(*ast.Ident)
			if ok {
				if tsel.Name == "sync" && fieldType.Sel.Name == "RWMutex" {
					fmt.Printf("Found declaration of rwmutex field %s\n", fieldName.Name)
				}
			}
		}
	}
}

func matchLockerDecl(x *ast.GenDecl, v *Visitor, n ast.Node) {
	if len(x.Specs) > 0 {
		spec, ok := x.Specs[0].(*ast.ValueSpec)
		if ok {
			for j := 0; j < len(spec.Names); j++ {
				id := spec.Names[j]
				t, ok := spec.Type.(*ast.SelectorExpr)
				if ok {
					tsel, ok := t.X.(*ast.Ident)
					if ok {
						if tsel.Name == "sync" && t.Sel.Name == "Locker" {
							fmt.Printf("Found declaration of locker %s\n", id.Name)
						}
					}
				}
			}
		}
	}
}

func matchLockerParamDecl(x *ast.Field, v *Visitor, n ast.Node) {
	for i := 0; i < len(x.Names); i++ {
		fieldName := x.Names[i]
		fieldType, ok := x.Type.(*ast.SelectorExpr)
		if ok {
			tsel, ok := fieldType.X.(*ast.Ident)
			if ok {
				if tsel.Name == "sync" && fieldType.Sel.Name == "Locker" {
					fmt.Printf("Found declaration of locker field %s\n", fieldName.Name)
				}
			}
		}
	}
}

func matchWaitGroupDone(x *ast.SelectorExpr, v *Visitor, n ast.Node) {
	funName := x.Sel
	if funName.Name == "Done" {
		var buf bytes.Buffer
		printer.Fprint(&buf, v.fset, x.X)
		fmt.Printf("Found call of Done on node %s\n", buf.String())
	}
}

func matchWaitGroupAdd(x *ast.SelectorExpr, v *Visitor, n ast.Node) {
	funName := x.Sel
	if funName.Name == "Add" {
		var buf bytes.Buffer
		printer.Fprint(&buf, v.fset, x.X)
		fmt.Printf("Found call of Add on node %s\n", buf.String())
	}
}

func matchLock(x *ast.SelectorExpr, v *Visitor, n ast.Node) {
	funName := x.Sel
	if funName.Name == "Lock" {
		var buf bytes.Buffer
		printer.Fprint(&buf, v.fset, x.X)
		fmt.Printf("Found call of Lock on node %s\n", buf.String())
	}
}

func matchUnlock(x *ast.SelectorExpr, v *Visitor, n ast.Node) {
	funName := x.Sel
	if funName.Name == "Unlock" {
		var buf bytes.Buffer
		printer.Fprint(&buf, v.fset, x.X)
		fmt.Printf("Found call of Unlock on node %s\n", buf.String())
	}
}

func matchWaitGroupWait(x *ast.SelectorExpr, v *Visitor, n ast.Node) {
	funName := x.Sel
	if funName.Name == "Wait" {
		var buf bytes.Buffer
		printer.Fprint(&buf, v.fset, x.X)
		fmt.Printf("Found call of Wait on node %s\n", buf.String())
	}
}

func matchNewCondCall(x *ast.CallExpr, v *Visitor, n ast.Node) {
	target, ok := x.Fun.(*ast.SelectorExpr)
	if ok {
		targetName, ok := target.X.(*ast.Ident)
		if ok {
			funName := target.Sel
			if funName.Name == "NewCond" && targetName.Name == "sync" {
				fmt.Print("Found call of NewCond\n")
			}
		}
	}
}

func matchCondSignalCall(x *ast.SelectorExpr, v *Visitor, n ast.Node) {
	funName := x.Sel
	if funName.Name == "Signal" {
		var buf bytes.Buffer
		printer.Fprint(&buf, v.fset, x.X)
		fmt.Printf("Found call of Signal on node %s\n", buf.String())
	}
}

func matchCondBroadcastCall(x *ast.SelectorExpr, v *Visitor, n ast.Node) {
	funName := x.Sel
	if funName.Name == "Broadcast" {
		var buf bytes.Buffer
		printer.Fprint(&buf, v.fset, x.X)
		fmt.Printf("Found call of Broadcast on node %s\n", buf.String())
	}
}

func matchOnceDecl(x *ast.GenDecl, v *Visitor, n ast.Node) {
	for i := 0; i < len(x.Specs); i++ {
		spec, ok := x.Specs[i].(*ast.ValueSpec)
		if ok {
			for j := 0; j < len(spec.Names); j++ {
				id := spec.Names[j]
				t, ok := spec.Type.(*ast.SelectorExpr)
				if ok {
					tsel, ok := t.X.(*ast.Ident)
					if ok {
						if tsel.Name == "sync" && t.Sel.Name == "Once" {
							fmt.Printf("Found declaration of once %s\n", id.Name)
						}
					}
				}
			}
		}
	}
}

func matchOnceDoCall(x *ast.SelectorExpr, v *Visitor, n ast.Node) {
	funName := x.Sel
	if funName.Name == "Do" {
		var buf bytes.Buffer
		printer.Fprint(&buf, v.fset, x.X)
		fmt.Printf("Found call of Do on node %s\n", buf.String())
	}
}

func (v *Visitor) Visit(n ast.Node) ast.Visitor {
	if n == nil {
		return nil
	}

	switch x := n.(type) {
	case *ast.CallExpr:
		matchMakeCall(x, v, n)
		matchNewCondCall(x, v, n)
	case *ast.SendStmt:
		matchSendStmt(x, v, n)
	case *ast.UnaryExpr:
		matchUnaryExpr(x, v, n)
	case *ast.GenDecl:
		matchWaitGroupDecl(x, v, n)
		matchMutexDecl(x, v, n)
		matchRWMutexDecl(x, v, n)
		matchLockerDecl(x, v, n)
		matchOnceDecl(x, v, n)
	case *ast.Field:
		matchWaitGroupParamDecl(x, v, n)
		matchMutexParamDecl(x, v, n)
		matchRWMutexParamDecl(x, v, n)
		matchLockerParamDecl(x, v, n)
	case *ast.SelectorExpr:
		matchWaitGroupDone(x, v, n)
		matchWaitGroupAdd(x, v, n)
		matchWaitGroupWait(x, v, n)
		matchLock(x, v, n)
		matchUnlock(x, v, n)
		matchCondSignalCall(x, v, n)
		matchCondBroadcastCall(x, v, n)
		matchOnceDoCall(x, v, n)
	}
	return v
}
