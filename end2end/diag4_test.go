package end2end

import (
	"fmt"
	"testing"

	il "github.com/glycerine/goivy/ivylogic"
	"github.com/glycerine/goivy/compiler"
	"github.com/glycerine/goivy/isolate"
	"github.com/glycerine/goivy/lexer"
	"github.com/glycerine/goivy/module"
	"github.com/glycerine/goivy/parser"
	"github.com/glycerine/goivy/ast"
)

func TestDiagnostic4(t *testing.T) {
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

	mod := module.New()
	mod.Sig = il.NewSig()

	topCtx := compiler.CollectActions(decls)
	c := compiler.NewFromModule(mod)
	c.TopCtx = topCtx

	// Pass 1
	ds := compiler.NewDomainSetup(c)
	if err := ds.ProcessDecls(decls); err != nil {
		t.Fatalf("pass 1: %v", err)
	}
	fmt.Printf("After pass 1 - Sig ptr: %p, Sorts: %v\n", mod.Sig, mod.Sig.Sorts)

	compiler.FixConstructors(mod)

	// Pass 2
	cs := compiler.NewConjSetup(c)
	if err := cs.ProcessDecls(decls); err != nil {
		t.Fatalf("pass 2: %v", err)
	}
	fmt.Printf("After pass 2 - Sig ptr: %p\n", mod.Sig)

	// Pass 3
	as := compiler.NewARGSetup(c)
	if err := as.ProcessDecls(decls); err != nil {
		t.Fatalf("pass 3: %v", err)
	}
	fmt.Printf("After pass 3 - Sig ptr: %p\n", mod.Sig)

	// Post-processing
	compiler.CreateSortOrder(mod)
	compiler.CreateConstructorSchemata(mod)
	compiler.AttachProofs(mod)
	compiler.CheckDefinitions(mod)
	compiler.CheckPropertiesPass(mod)
	compiler.CreateConjActions(mod)
	compiler.HandleTemporals(mod)
	fmt.Printf("After post-proc - Sig ptr: %p\n", mod.Sig)

	// Create isolate
	if _, ok := mod.Isolates["this"]; !ok {
		isol := &ast.IsolateDef{
			Elems:    []ast.Node{ast.NewAtom("this"), ast.NewAtom("this")},
			WithArgs: 0,
		}
		mod.Isolates["this"] = isol
	}

	if err := isolate.CreateIsolate("this", mod); err != nil {
		fmt.Printf("CreateIsolate error: %v\n", err)
	}
	fmt.Printf("After CreateIsolate - Sig ptr: %p, Sorts: %v\n", mod.Sig, mod.Sig.Sorts)

	// Final
	mod.LabeledAxioms = append(mod.LabeledAxioms, mod.LabeledProps...)
	fmt.Printf("Final - Sig ptr: %p, Sorts: %v\n", mod.Sig, mod.Sig.Sorts)
}
