package end2end

import (
	"fmt"
	"testing"

	"github.com/glycerine/goivy/compiler"
	il "github.com/glycerine/goivy/ivylogic"
	"github.com/glycerine/goivy/lexer"
	"github.com/glycerine/goivy/module"
	"github.com/glycerine/goivy/parser"
)

func TestDiagnostic3(t *testing.T) {
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
	sigPtr := mod.Sig
	fmt.Printf("Sig ptr before: %p\n", sigPtr)
	
	// Manual pass 1
	c := compiler.NewFromModule(mod)
	ds := compiler.NewDomainSetup(c)
	if err := ds.ProcessDecls(decls); err != nil {
		t.Fatalf("pass 1 error: %v", err)
	}
	fmt.Printf("After pass 1 - Sig ptr: %p, same? %v\n", mod.Sig, mod.Sig == sigPtr)
	fmt.Printf("After pass 1 - Sorts: %v\n", mod.Sig.Sorts)
	fmt.Printf("After pass 1 - Symbols: %d\n", len(mod.Sig.Symbols))
	for k := range mod.Sig.Symbols {
		fmt.Printf("  sym: %s\n", k)
	}
	
	// Now do full IvyCompile
	mod2 := module.New()
	mod2.Sig = il.NewSig()
	err = compiler.IvyCompile(decls, mod2)
	fmt.Printf("\nAfter IvyCompile - Sig ptr: %p\n", mod2.Sig)
	fmt.Printf("After IvyCompile - Sorts: %v\n", mod2.Sig.Sorts)
	fmt.Printf("After IvyCompile - Symbols: %d\n", len(mod2.Sig.Symbols))
}
