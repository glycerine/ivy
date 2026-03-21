package compiler

import (
	"math/rand"
	"testing"

	"github.com/glycerine/goivy/actions"
	"github.com/glycerine/goivy/ast"
	"github.com/glycerine/goivy/module"
)

// makeTestIsolateDef creates an *ast.IsolateDef with the given verified and present names.
func makeTestIsolateDef(verified, present []string) *ast.IsolateDef {
	// Elems layout: [name, verified..., present...]
	// WithArgs = len(present)
	elems := []ast.Node{ast.NewAtom("test")}
	for _, v := range verified {
		elems = append(elems, ast.NewAtom(v))
	}
	for _, p := range present {
		elems = append(elems, ast.NewAtom(p))
	}
	return &ast.IsolateDef{Elems: elems, WithArgs: len(present)}
}

// helper to get labels from an action via GetLabels interface.
func getTestLabels(act interface{}) []string {
	if gl, ok := act.(interface{ GetLabels() []string }); ok {
		return gl.GetLabels()
	}
	return nil
}

// TestHandleTemporals_EmptyActions — No actions → no panic.
func TestHandleTemporals_EmptyActions(t *testing.T) {
	mod := module.New()
	// no actions, no isolates
	HandleTemporals(mod) // should not panic
}

// TestHandleTemporals_ActionGetsLabels — One action in one isolate → gets that isolate name.
func TestHandleTemporals_ActionGetsLabels(t *testing.T) {
	mod := module.New()
	seq := actions.NewSequence()
	mod.Actions["act1"] = seq
	mod.Isolates["iso1"] = makeTestIsolateDef([]string{"act1"}, nil)

	HandleTemporals(mod)

	labels := getTestLabels(seq)
	if len(labels) != 1 || labels[0] != "iso1" {
		t.Fatalf("expected [iso1], got %v", labels)
	}
}

// TestHandleTemporals_ActionNotInAnyIsolate — action not in any isolate → Labels set (not nil).
// Verifies Bug 1 fix: Python's defaultdict returns [] for missing keys.
func TestHandleTemporals_ActionNotInAnyIsolate(t *testing.T) {
	mod := module.New()
	seq := actions.NewSequence()
	// Give it non-nil labels to prove SetLabels was called (overwriting with nil from imap).
	seq.Labels = []string{"should-be-cleared"}
	mod.Actions["lonely"] = seq
	// No isolates at all

	HandleTemporals(mod)

	labels := getTestLabels(seq)
	// Python would set labels = [] (empty list). Go sets nil from map miss.
	// The key point: SetLabels WAS called (labels changed from original).
	if len(labels) != 0 {
		t.Fatalf("expected empty labels, got %v", labels)
	}
	// Verify the original value was overwritten
	if seq.Labels != nil && len(seq.Labels) > 0 {
		t.Fatalf("expected Labels to be cleared, got %v", seq.Labels)
	}
}

// TestHandleTemporals_MultipleIsolates — action in 2 isolates → Labels has both names.
func TestHandleTemporals_MultipleIsolates(t *testing.T) {
	mod := module.New()
	seq := actions.NewSequence()
	mod.Actions["act1"] = seq
	mod.Isolates["isoA"] = makeTestIsolateDef([]string{"act1"}, nil)
	mod.Isolates["isoB"] = makeTestIsolateDef([]string{"act1"}, nil)

	HandleTemporals(mod)

	labels := getTestLabels(seq)
	if len(labels) != 2 {
		t.Fatalf("expected 2 labels, got %v", labels)
	}
	has := make(map[string]bool)
	for _, l := range labels {
		has[l] = true
	}
	if !has["isoA"] || !has["isoB"] {
		t.Fatalf("expected isoA and isoB, got %v", labels)
	}
}

// TestHandleTemporals_MultipleActions — different actions with different isolate memberships.
func TestHandleTemporals_MultipleActions(t *testing.T) {
	mod := module.New()
	seq1 := actions.NewSequence()
	seq2 := actions.NewSequence()
	seq3 := actions.NewSequence()
	mod.Actions["act1"] = seq1
	mod.Actions["act2"] = seq2
	mod.Actions["act3"] = seq3

	mod.Isolates["iso1"] = makeTestIsolateDef([]string{"act1", "act2"}, nil)
	mod.Isolates["iso2"] = makeTestIsolateDef([]string{"act2", "act3"}, nil)

	HandleTemporals(mod)

	l1 := getTestLabels(seq1)
	l2 := getTestLabels(seq2)
	l3 := getTestLabels(seq3)

	if len(l1) != 1 || l1[0] != "iso1" {
		t.Fatalf("act1: expected [iso1], got %v", l1)
	}
	if len(l2) != 2 {
		t.Fatalf("act2: expected 2 labels, got %v", l2)
	}
	if len(l3) != 1 || l3[0] != "iso2" {
		t.Fatalf("act3: expected [iso2], got %v", l3)
	}
}

// TestHandleTemporals_GetLabelsWorks — verify GetLabels() interface works after HandleTemporals.
// This tests the Bug 2 fix: GetLabels() added to ActionBase.
func TestHandleTemporals_GetLabelsWorks(t *testing.T) {
	mod := module.New()
	seq := actions.NewSequence()
	mod.Actions["act1"] = seq
	mod.Isolates["iso1"] = makeTestIsolateDef([]string{"act1"}, nil)

	HandleTemporals(mod)

	// Use the interface assertion that temporal.getLabels() uses
	var iface interface{} = seq
	gl, ok := iface.(interface{ GetLabels() []string })
	if !ok {
		t.Fatal("Sequence does not implement GetLabels() interface")
	}
	labels := gl.GetLabels()
	if len(labels) != 1 || labels[0] != "iso1" {
		t.Fatalf("GetLabels: expected [iso1], got %v", labels)
	}
}

// TestHandleTemporals_NoIsolates — module with actions but empty Isolates → all get nil/empty labels.
func TestHandleTemporals_NoIsolates(t *testing.T) {
	mod := module.New()
	seq1 := actions.NewSequence()
	seq2 := actions.NewSequence()
	mod.Actions["act1"] = seq1
	mod.Actions["act2"] = seq2
	// No isolates

	HandleTemporals(mod)

	for name, act := range map[string]*actions.Sequence{"act1": seq1, "act2": seq2} {
		labels := getTestLabels(act)
		if len(labels) != 0 {
			t.Fatalf("%s: expected empty labels, got %v", name, labels)
		}
	}
}

// FuzzHandleTemporals — random combinations of 0-5 actions and 0-3 isolates.
func FuzzHandleTemporals(f *testing.F) {
	f.Add(uint64(0))
	f.Add(uint64(12345))
	f.Add(uint64(99999))

	f.Fuzz(func(t *testing.T, seed uint64) {
		rng := rand.New(rand.NewSource(int64(seed)))

		numActions := rng.Intn(6)  // 0-5
		numIsolates := rng.Intn(4) // 0-3

		mod := module.New()

		// Create actions
		actNames := make([]string, numActions)
		actPtrs := make([]*actions.Sequence, numActions)
		for i := 0; i < numActions; i++ {
			name := string(rune('a' + i))
			actNames[i] = name
			actPtrs[i] = actions.NewSequence()
			mod.Actions[name] = actPtrs[i]
		}

		// Create isolates with random membership
		isoNames := make([]string, numIsolates)
		for i := 0; i < numIsolates; i++ {
			isoName := "iso" + string(rune('0'+i))
			isoNames[i] = isoName
			var verified []string
			for _, an := range actNames {
				if rng.Intn(2) == 1 {
					verified = append(verified, an)
				}
			}
			mod.Isolates[isoName] = makeTestIsolateDef(verified, nil)
		}

		// Should not panic
		HandleTemporals(mod)

		// Every action that implements SetLabels should have had SetLabels called.
		// Verify: Labels is set (not left as the zero value from before).
		// Since we start with nil Labels and SetLabels(nil) sets nil,
		// we can't distinguish "was called with nil" from "was not called" easily.
		// But we CAN verify: any label values are valid isolate names.
		validIso := make(map[string]bool)
		for _, n := range isoNames {
			validIso[n] = true
		}
		for i, act := range actPtrs {
			labels := getTestLabels(act)
			for _, l := range labels {
				if !validIso[l] {
					t.Fatalf("action %s has invalid isolate label %q", actNames[i], l)
				}
			}
		}
	})
}
