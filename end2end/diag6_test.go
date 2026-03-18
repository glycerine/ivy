package end2end

import (
	"fmt"
	"testing"

	"github.com/glycerine/goivy/actions"
	"github.com/glycerine/goivy/compiler"
	il "github.com/glycerine/goivy/ivylogic"
	"github.com/glycerine/goivy/lexer"
	"github.com/glycerine/goivy/module"
	"github.com/glycerine/goivy/parser"
	"github.com/glycerine/goivy/isolate"
)

func TestDiagnostic6(t *testing.T) {
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

	ds := compiler.NewDomainSetup(c)
	if err := ds.ProcessDecls(decls); err != nil {
		t.Fatalf("pass 1: %v", err)
	}

	// Check action types
	fmt.Printf("Actions after pass 1:\n")
	for name, act := range mod.Actions {
		_, isAction := act.(actions.Action)
		fmt.Printf("  %s: %T, implements actions.Action: %v\n", name, act, isAction)
	}

	// Check what GetIsolateInfoFull returns
	fmt.Printf("\nIsolates: %d\n", len(mod.Isolates))
	for k := range mod.Isolates {
		fmt.Printf("  %s\n", k)
	}

	// Check exported info
	fmt.Printf("\nExports: %d\n", len(mod.Exports))
	for i, exp := range mod.Exports {
		fmt.Printf("  [%d] %T: %v\n", i, exp, exp)
	}

	// Check what verified/present are for "this" isolate
	fmt.Printf("\nPublicActions: %d\n", len(mod.PublicActions))
	for k := range mod.PublicActions {
		fmt.Printf("  %s\n", k)
	}

	// Check LookupAction
	for name := range mod.Actions {
		act, err := isolate.LookupAction(mod, name)
		fmt.Printf("LookupAction(%s): %T, err=%v\n", name, act, err)
	}
}
