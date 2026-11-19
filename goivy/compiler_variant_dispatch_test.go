package goivy

import (
	"os"
	"path/filepath"
	"testing"
)

// Python compile_inline_call dispatches a method call on a type with
// variants through IfAction(ivy_ast.Some(tmpsym, *>(arg, tmpsym)), ...):
// the condition is an AST Some (goivy SomeCondition), not a logic Exists
// (test/annot1.ivy).
func TestVariantMethodDispatchUsesSome(t *testing.T) {
	src := `#lang ivy1.7

object t = {
    type this
    action a(x:this)
}

object q = {
    variant this of t = struct {
        b : bool
    }
    action a(x:this) = {
    }
}

action a(x:t) = {
    call x.a
}

export a
`
	fn := filepath.Join(t.TempDir(), "dispatch.ivy")
	if err := os.WriteFile(fn, []byte(src), 0644); err != nil {
		t.Fatal(err)
	}
	mod, err := IvyLoadFile(fn)
	if err != nil {
		t.Fatal(err)
	}
	act0, ok := mod.Actions.Get2("ext:a")
	if !ok {
		t.Fatal("no action a")
	}
	act, _ := act0.(ActionsAction)
	var found bool
	// IterSubactions, like Python's iter_subactions, yields the action
	// and all its descendants.
	for _, a := range act.IterSubactions() {
		if ifa, ok := a.(*LogicIfAction); ok {
			if _, isSome := ifa.Cond.(*SomeCondition); !isSome {
				t.Fatalf("dispatch condition is %T, want *SomeCondition", ifa.Cond)
			}
			found = true
		}
	}
	if !found {
		t.Fatalf("no dispatch if in %v", act)
	}
}
