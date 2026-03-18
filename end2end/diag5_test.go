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

func TestDiagnostic5(t *testing.T) {
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
	fmt.Printf("After pass 1:\n")
	fmt.Printf("  Sorts: %v\n", mod.Sig.Sorts)
	fmt.Printf("  Symbols: %d\n", len(mod.Sig.Symbols))
	fmt.Printf("  LabeledConjs: %d\n", len(mod.LabeledConjs))
	for i, lc := range mod.LabeledConjs {
		fmt.Printf("    [%d] %v\n", i, lc.Formula)
	}
	fmt.Printf("  Actions: %d\n", len(mod.Actions))
	for k := range mod.Actions {
		fmt.Printf("    %s\n", k)
	}
	fmt.Printf("  Mixins: %d keys\n", len(mod.Mixins))
	for k, v := range mod.Mixins {
		fmt.Printf("    %s: %d\n", k, len(v))
	}
	fmt.Printf("  Exports: %d\n", len(mod.Exports))
	fmt.Printf("  Relations: %d\n", len(mod.Relations))
	for k := range mod.Relations {
		fmt.Printf("    %s\n", k)
	}
}
