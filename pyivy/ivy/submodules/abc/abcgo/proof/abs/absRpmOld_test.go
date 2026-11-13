package abs

import (
	"reflect"
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
)

func TestInputRegisterConversionRoundTrip(t *testing.T) {
	p := gia.Gia_ManStart(16)
	x, y, state := gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)
	out := gia.Gia_ManAppendAnd(p, x, state^1)
	next := gia.Gia_ManAppendAnd(p, y, state)
	gia.Gia_ManAppendCo(p, out)
	gia.Gia_ManAppendCo(p, next)
	gia.Gia_ManSetRegNum(p, 1)
	converted := Gia_ManDupIn2Ff(p)
	if gia.Gia_ManPiNum(converted) != 3 || gia.Gia_ManRegNum(converted) != 2 || gia.Gia_ManPoNum(converted) != 2 {
		t.Fatalf("unexpected converted interface: pi=%d po=%d regs=%d", gia.Gia_ManPiNum(converted), gia.Gia_ManPoNum(converted), gia.Gia_ManRegNum(converted))
	}
	restored := Gia_ManDupFf2In(converted, 1)
	if gia.Gia_ManPiNum(restored) != 2 || gia.Gia_ManPoNum(restored) != 1 || gia.Gia_ManRegNum(restored) != 1 {
		t.Fatalf("unexpected restored interface: pi=%d po=%d regs=%d", gia.Gia_ManPiNum(restored), gia.Gia_ManPoNum(restored), gia.Gia_ManRegNum(restored))
	}
	originalGraph, restoredGraph := gia.Gia_ManToAigSimple(p), gia.Gia_ManToAigSimple(restored)
	for mask := 0; mask < 8; mask++ {
		inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
		want, err := originalGraph.Evaluate(inputs)
		if err != nil {
			t.Fatal(err)
		}
		got, err := restoredGraph.Evaluate(inputs)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("inputs %v: got %v want %v", inputs, got, want)
		}
	}
}

func TestOldReparameterizationPreservesOutputRange(t *testing.T) {
	p := gia.Gia_ManStart(8)
	a, b := gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)
	gia.Gia_ManAppendCo(p, gia.Gia_ManAppendAnd(p, a, b))
	result := Abs_RpmPerformOld(p, 0)
	if result == nil || gia.Gia_ManPoNum(result) != 1 || gia.Gia_ManRegNum(result) != 0 {
		t.Fatal("reparameterization lost interface")
	}
	g := gia.Gia_ManToAigSimple(result)
	seen := [2]bool{}
	for mask := 0; mask < 1<<len(g.Inputs()); mask++ {
		inputs := make([]bool, len(g.Inputs()))
		for i := range inputs {
			inputs[i] = mask&(1<<i) != 0
		}
		output, err := g.Evaluate(inputs)
		if err != nil {
			t.Fatal(err)
		}
		if output[0] {
			seen[1] = true
		} else {
			seen[0] = true
		}
	}
	if !seen[0] || !seen[1] {
		t.Fatalf("output range changed: %v", seen)
	}
}
