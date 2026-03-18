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

func TestDiagnostic2(t *testing.T) {
	src := `#lang ivy1.7

type t
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
		if td, ok := d.(*ast.TypeDecl); ok {
			fmt.Printf("    DeclArgs: %d\n", len(td.DeclArgs))
			for j, a := range td.DeclArgs {
				fmt.Printf("    arg[%d] %T: %s\n", j, a, a)
				if tdf, ok := a.(*ast.TypeDef); ok {
					fmt.Printf("      Name: %T = %v\n", tdf.Name, tdf.Name)
					fmt.Printf("      Value: %T = %v\n", tdf.Value, tdf.Value)
				}
			}
		}
	}

	mod := module.New()
	mod.Sig = il.NewSig()
	fmt.Printf("\nBefore compile - Sig addr: %p\n", mod.Sig)
	fmt.Printf("Before compile - Sorts: %v\n", mod.Sig.Sorts)
	
	// Do manual domain setup to trace
	c := compiler.NewFromModule(mod)
	fmt.Printf("Compiler Sig addr: %p\n", c.Sig)
	
	ds := compiler.NewDomainSetup(c)
	for _, decl := range decls {
		fmt.Printf("\nProcessing: %T\n", decl)
		if err := ds.ProcessDecl(decl); err != nil {
			t.Fatalf("process error: %v", err)
		}
		fmt.Printf("After processing - Sorts: %v\n", mod.Sig.Sorts)
	}
	fmt.Printf("\nFinal Sorts: %v\n", mod.Sig.Sorts)
}
