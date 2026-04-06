package end2end

import (
	"fmt"
	"testing"

	"github.com/glycerine/ivy/goivy/ast"
	"github.com/glycerine/ivy/goivy/lexer"
	"github.com/glycerine/ivy/goivy/parser"
)

func TestDebugClientServerInit(t *testing.T) {
	mod := compileIvyFile(t, "client_server.ivy")

	fmt.Printf("InitCond: %v\n", mod.InitCond)
	fmt.Printf("InitCond is nil: %v\n", mod.InitCond == nil)
	if mod.InitCond != nil {
		fmt.Printf("InitCond fmlas: %d\n", len(mod.InitCond.Fmlas))
		for i, f := range mod.InitCond.Fmlas {
			fmt.Printf("  fmla[%d]: %s\n", i, f)
		}
	}
	fmt.Printf("LabeledInits: %d\n", len(mod.LabeledInits))
	for i, li := range mod.LabeledInits {
		fmt.Printf("  init[%d]: %s\n", i, li.Formula)
	}
	fmt.Printf("Initializers: %d\n", len(mod.Initializers))
	for i, na := range mod.Initializers {
		fmt.Printf("  initializer[%d]: name=%s type=%T\n", i, na.Name, na.Action)
	}
	fmt.Printf("Mixins: %d keys\n", mod.Mixins.Len())
	for k, v := range mod.Mixins.All() {
		fmt.Printf("  mixins[%q]: %d entries\n", k, len(v))
		for j, m := range v {
			fmt.Printf("    [%d]: %T = %s\n", j, m, m)
		}
	}
	fmt.Printf("Actions: %d\n", mod.Actions.Len())
	for k, v := range mod.Actions.All() {
		fmt.Printf("  actions[%q]: %T\n", k, v)
	}
}

func TestDebugParseInit(t *testing.T) {
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
`
	version := lexer.Version{1, 7}
	result, err := parser.Parse(src, version)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	decls := result.Decls
	fmt.Printf("Parsed %d declarations:\n", len(decls))
	for i, d := range decls {
		fmt.Printf("  [%d] %T: %s\n", i, d, d)
		if id, ok := d.(*ast.InitDecl); ok {
			fmt.Printf("      InitDecl with %d args\n", len(id.DeclArgs))
			for j, arg := range id.DeclArgs {
				fmt.Printf("        arg[%d]: %T = %s\n", j, arg, arg)
			}
		}
	}
}
