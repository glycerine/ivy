package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type inventoryPackage struct {
	Package string          `json:"package"`
	Files   []inventoryFile `json:"files"`
}

type inventoryFile struct {
	Path    string    `json:"path"`
	Consts  []nameRow `json:"consts"`
	Vars    []nameRow `json:"vars"`
	Types   []nameRow `json:"types"`
	Funcs   []funcRow `json:"funcs"`
	Methods []funcRow `json:"methods"`
}

type nameRow struct {
	Name string `json:"name"`
	Kind string `json:"kind,omitempty"`
}

type funcRow struct {
	Name     string         `json:"name"`
	Recv     string         `json:"recv,omitempty"`
	Position string         `json:"position"`
	Shape    map[string]int `json:"shape"`
}

func main() {
	dir := flag.String("dir", "", "Go package source directory")
	pkgName := flag.String("pkg", "", "package name")
	out := flag.String("out", "", "output inventory JSON path")
	flag.Parse()

	if *dir == "" || *pkgName == "" || *out == "" {
		fmt.Fprintln(os.Stderr, "usage: go-package-inventory -dir DIR -pkg NAME -out PATH")
		os.Exit(2)
	}
	inv, err := inventoryDir(*dir, *pkgName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "inventory %s: %v\n", *dir, err)
		os.Exit(1)
	}
	data, err := json.MarshalIndent([]inventoryPackage{inv}, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "marshal inventory: %v\n", err)
		os.Exit(1)
	}
	data = append(data, '\n')
	if err := os.WriteFile(*out, data, 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "write %s: %v\n", *out, err)
		os.Exit(1)
	}
}

func inventoryDir(dir, pkgName string) (inventoryPackage, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return inventoryPackage{}, err
	}
	fset := token.NewFileSet()
	files := make([]inventoryFile, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		path := filepath.Join(dir, name)
		file, err := parser.ParseFile(fset, path, nil, parser.ParseComments|parser.SkipObjectResolution)
		if err != nil {
			return inventoryPackage{}, err
		}
		files = append(files, inventoryFileFor(fset, path, file))
	}
	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
	return inventoryPackage{Package: pkgName, Files: files}, nil
}

func inventoryFileFor(fset *token.FileSet, path string, file *ast.File) inventoryFile {
	out := inventoryFile{Path: path}
	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.GenDecl:
			for _, spec := range d.Specs {
				switch s := spec.(type) {
				case *ast.ValueSpec:
					for _, name := range s.Names {
						if d.Tok == token.CONST {
							out.Consts = append(out.Consts, nameRow{Name: name.Name})
						}
						if d.Tok == token.VAR {
							out.Vars = append(out.Vars, nameRow{Name: name.Name})
						}
					}
				case *ast.TypeSpec:
					out.Types = append(out.Types, nameRow{Name: s.Name.Name, Kind: typeKind(s.Type)})
				}
			}
		case *ast.FuncDecl:
			row := funcRow{
				Name:     d.Name.Name,
				Position: fset.Position(d.Pos()).String(),
				Shape:    shapeOf(d.Body),
			}
			if d.Recv != nil {
				row.Recv = recvName(d.Recv)
				out.Methods = append(out.Methods, row)
			} else {
				out.Funcs = append(out.Funcs, row)
			}
		}
	}
	return out
}

func typeKind(expr ast.Expr) string {
	switch x := expr.(type) {
	case *ast.StructType:
		return "struct"
	case *ast.InterfaceType:
		return "interface"
	case *ast.FuncType:
		return "func"
	case *ast.ArrayType:
		return "*ast.ArrayType"
	case *ast.MapType:
		return "*ast.MapType"
	case *ast.Ident:
		return "*ast.Ident"
	case *ast.SelectorExpr:
		return "*ast.SelectorExpr"
	case *ast.StarExpr:
		return "*ast.StarExpr"
	case *ast.IndexExpr:
		return "*ast.IndexExpr"
	case *ast.IndexListExpr:
		return "*ast.IndexListExpr"
	default:
		return fmt.Sprintf("%T", x)
	}
}

func recvName(recv *ast.FieldList) string {
	if recv == nil || len(recv.List) == 0 {
		return ""
	}
	return exprName(recv.List[0].Type)
}

func exprName(expr ast.Expr) string {
	switch x := expr.(type) {
	case *ast.Ident:
		return x.Name
	case *ast.StarExpr:
		return "*" + exprName(x.X)
	case *ast.SelectorExpr:
		return exprName(x.X) + "." + x.Sel.Name
	case *ast.IndexExpr:
		return exprName(x.X)
	case *ast.IndexListExpr:
		return exprName(x.X)
	case *ast.ParenExpr:
		return exprName(x.X)
	default:
		return fmt.Sprintf("%T", x)
	}
}

func shapeOf(body *ast.BlockStmt) map[string]int {
	shape := map[string]int{
		"if":          0,
		"for":         0,
		"range":       0,
		"switch":      0,
		"typeSwitch":  0,
		"select":      0,
		"branch":      0,
		"assign":      0,
		"return":      0,
		"defer":       0,
		"go":          0,
		"call":        0,
		"binary":      0,
		"unary":       0,
		"composite":   0,
		"funcLiteral": 0,
	}
	if body == nil {
		return shape
	}
	ast.Inspect(body, func(n ast.Node) bool {
		switch n.(type) {
		case *ast.IfStmt:
			shape["if"]++
		case *ast.ForStmt:
			shape["for"]++
		case *ast.RangeStmt:
			shape["range"]++
		case *ast.SwitchStmt:
			shape["switch"]++
		case *ast.TypeSwitchStmt:
			shape["typeSwitch"]++
		case *ast.SelectStmt:
			shape["select"]++
		case *ast.BranchStmt:
			shape["branch"]++
		case *ast.AssignStmt:
			shape["assign"]++
		case *ast.ReturnStmt:
			shape["return"]++
		case *ast.DeferStmt:
			shape["defer"]++
		case *ast.GoStmt:
			shape["go"]++
		case *ast.CallExpr:
			shape["call"]++
		case *ast.BinaryExpr:
			shape["binary"]++
		case *ast.UnaryExpr:
			shape["unary"]++
		case *ast.CompositeLit:
			shape["composite"]++
		case *ast.FuncLit:
			shape["funcLiteral"]++
		}
		return true
	})
	return shape
}
