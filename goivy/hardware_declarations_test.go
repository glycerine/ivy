package goivy

import (
	"os"
	"strings"
	"testing"
)

func TestFrozenWires(t *testing.T) {
	const fixture = "test_vectors/frozen_wires.ivy"
	source, err := os.ReadFile(fixture)
	if err != nil {
		t.Fatal(err)
	}
	if err := StartSourceWithConfig(fixture, string(source), NewConfig()); err != nil {
		t.Fatal(err)
	}

	mod := New()
	if err := SourceString(fixture, string(source), mod, mod.Sig, map[string]interface{}{"create_isolate": false}); err != nil {
		t.Fatal(err)
	}
	if mod.Wires.Len() != 4 || mod.Registers.Len() != 2 {
		t.Fatalf("got %d wires and %d registers, want 4 and 2", mod.Wires.Len(), mod.Registers.Len())
	}
	if len(mod.InputWires) != 1 || mod.InputWires[0].Name != "input_value" {
		t.Fatalf("input wires = %v, want input_value", mod.InputWires)
	}
	if len(mod.OutputWires) != 1 || mod.OutputWires[0].Name != "inline" {
		t.Fatalf("output wires = %v, want inline", mod.OutputWires)
	}
	for _, update := range mod.Updates {
		if derived, ok := update.(*DerivedUpdate); ok {
			if _, wire := mod.Wires.Get2(Key(derived.Symbol)); wire {
				t.Errorf("wire %v acquired a derived update", derived.Symbol)
			}
		}
	}

	mod.UpdateTheory()
	copy := mod.Copy()
	copy.Wires.Delkey(Key(mod.InputWires[0]))
	if mod.Wires.Len() != 4 || copy.Wires.Len() != 3 {
		t.Fatal("a module copy shares its wire table with the original")
	}
	act := mod.Actions.Get("tick")
	ctx := &UpdateContext{Domain: mod, ActCfg: mod.Cfg.ActCfg}
	full := GetUpdate(act, ctx)
	_, modified := uniqueConstSet(full.Modified)
	for key, wire := range mod.Wires.All() {
		if !modified[key] {
			t.Errorf("wire %s is framed across the action boundary", wire.Name)
		}
	}
}

func TestSharedRegisterOnExportedAction(t *testing.T) {
	const fixture = "test_vectors/shared_register.ivy"
	source, err := os.ReadFile(fixture)
	if err != nil {
		t.Fatal(err)
	}
	if err := StartSourceWithConfig(fixture, string(source), NewConfig()); err != nil {
		t.Fatal(err)
	}

	mod := New()
	if err := SourceString(fixture, string(source), mod, mod.Sig, map[string]interface{}{"create_isolate": false}); err != nil {
		t.Fatal(err)
	}
	if err := CreateIsolate("observer", mod); err != nil {
		t.Fatal(err)
	}
	if !mod.PublicActions.Get("ext:tick") {
		t.Fatal("the observer's exported clock is missing")
	}
	var havocsShared bool
	for _, sub := range mod.Actions.Get("ext:tick").IterSubactions() {
		if havoc, ok := sub.(*LogicHavocAction); ok && constName(havoc.Target) == "shared" {
			havocsShared = true
		}
	}
	if !havocsShared {
		t.Fatal("the omitted writer's transitive register write was not summarized")
	}

	ordinary := strings.Replace(string(source), "register shared", "var shared", 1)
	err = StartSourceWithConfig(fixture, ordinary, NewConfig())
	if err == nil || !strings.Contains(err.Error(), "may have visible effect on shared") {
		t.Fatalf("ordinary state interference error = %v", err)
	}
}
