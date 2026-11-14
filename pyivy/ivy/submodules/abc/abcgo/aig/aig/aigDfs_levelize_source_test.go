package aig

import (
	"reflect"
	"testing"
)

func TestLevelizeIncludesAllLiveObjects(t *testing.T) {
	p := NewGraph()
	if got := Aig_ManLevelize(p); !reflect.DeepEqual(got, [][]Edge{{p.Const1()}}) {
		t.Errorf("empty manager levels = %v", got)
	}
	a, b, unused := p.NewInput(), p.NewInput(), p.NewInput()
	root := p.AND(a, b)
	out := p.AddOutput(root)
	want := [][]Edge{{p.Const1(), a, b, unused}, {root, out}}
	if got := Aig_ManLevelize(p); !reflect.DeepEqual(got, want) {
		t.Fatalf("level objects = %v, want %v", got, want)
	}
	// Deleted slots must not become phantom members of a level.
	dead := p.AND(a, unused)
	Aig_ManCleanup(p)
	if !p.deleted[dead.NodeID()] {
		t.Fatal("fixture did not delete the dangling node")
	}
	if got := Aig_ManLevelize(p); !reflect.DeepEqual(got, want) {
		t.Fatalf("levels after cleanup = %v, want %v", got, want)
	}
}

func TestLevelizeRejectsObjectAboveOutputLevel(t *testing.T) {
	p := NewGraph()
	a, b := p.NewInput(), p.NewInput()
	p.AND(a, b)
	p.AddOutput(a)
	defer func() {
		if recover() == nil {
			t.Fatal("accepted node above Aig_ManLevelNum, unlike C's assertion")
		}
	}()
	Aig_ManLevelize(p)
}
