package end2end

import (
	"fmt"
	"testing"

	"github.com/glycerine/goivy/ast"
	"github.com/glycerine/goivy/compiler"
	il "github.com/glycerine/goivy/ivylogic"
	"github.com/glycerine/goivy/lexer"
	"github.com/glycerine/goivy/module"
	"github.com/glycerine/goivy/parser"
)

func TestDiagnostic(t *testing.T) {
	src := `#lang ivy1.7

type t

relation r(X:t)

after init {
    r(X) := false
}

action step(x:t) = {
    r(x) := true
}

export step

conjecture r(X) | ~r(X)
`
	version := lexer.Version{1, 7}
	p := parser.New(src, version)
	decls, err := p.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	fmt.Printf("Parsed %d declarations:\n", len(decls))
	for i, d := range decls {
		fmt.Printf("  [%d] %T: %s\n", i, d, d)
		switch dd := d.(type) {
		case *ast.TypeDecl:
			fmt.Printf("       DeclArgs: %d\n", len(dd.DeclArgs))
			for j, a := range dd.DeclArgs {
				fmt.Printf("         [%d] %T: %s\n", j, a, a)
			}
		case *ast.RelationDecl:
			fmt.Printf("       DeclArgs: %d\n", len(dd.DeclArgs))
			for j, a := range dd.DeclArgs {
				fmt.Printf("         [%d] %T: %s\n", j, a, a)
			}
		case *ast.ConjectureDecl:
			fmt.Printf("       DeclArgs: %d\n", len(dd.DeclArgs))
			for j, a := range dd.DeclArgs {
				fmt.Printf("         [%d] %T: %s\n", j, a, a)
			}
		}
	}
	
	mod := module.New()
	mod.Sig = il.NewSig()
	err = compiler.IvyCompile(decls, mod)
	if err != nil {
		t.Fatalf("compile error: %v", err)
	}
	
	fmt.Printf("\nAfter compile:\n")
	fmt.Printf("  Sorts: %d\n", len(mod.Sig.Sorts))
	for k, v := range mod.Sig.Sorts {
		fmt.Printf("    %s: %T %s\n", k, v, v)
	}
	fmt.Printf("  Symbols: %d\n", len(mod.Sig.Symbols))
	for k, v := range mod.Sig.Symbols {
		fmt.Printf("    %s: %s\n", k, v.Sort)
	}
	fmt.Printf("  LabeledConjs: %d\n", len(mod.LabeledConjs))
	fmt.Printf("  Actions: %d\n", len(mod.Actions))
	for k := range mod.Actions {
		fmt.Printf("    %s\n", k)
	}
}
