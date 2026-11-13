package live

import (
	"math"
	"testing"
	"time"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/util"
)

func TestCheckSafetyPDRBoundaryResults(t *testing.T) {
	g := aig.NewGraph()
	x := g.NewInput()
	g.AddOutput(x)

	for name, got := range map[string]SafetyResult{
		"nil graph":          CheckSafetyPDR(nil, 0, SafetyOptions{}),
		"negative output":    CheckSafetyPDR(g, -1, SafetyOptions{}),
		"missing output":     CheckSafetyPDR(g, 1, SafetyOptions{}),
		"negative conflicts": CheckSafetyPDR(g, 0, SafetyOptions{ConflictLimit: -1}),
		"large conflicts":    CheckSafetyPDR(g, 0, SafetyOptions{ConflictLimit: int64(math.MaxInt32) + 1}),
	} {
		if got.Status != -1 || got.Error == nil {
			t.Errorf("%s: expected unknown with error, got %+v", name, got)
		}
	}
	if got := CheckSafetyPDR(g, 0, SafetyOptions{Deadline: time.Now().Add(-time.Second)}); got.Status != -1 || got.Error != nil {
		t.Fatalf("expired deadline should return unknown without solver error: %+v", got)
	}
	if got := CheckSafetyPDR(g, 0, SafetyOptions{Deadline: time.Now().Add(time.Minute)}); got.Status != 0 || got.Error != nil {
		t.Fatalf("future deadline changed a failing property result: %+v", got)
	}
}

func TestDecodeSafetyWitnessRejectsIllegalOrNonfailingTrace(t *testing.T) {
	t.Run("constraint violation", func(t *testing.T) {
		g := aig.NewGraph()
		x := g.NewInput()
		g.AddOutput(x) // selected property
		g.AddOutput(x) // trailing constraint violation
		aig.Aig_ManSetConstrNum(g, 1)
		cex := &util.Abc_Cex_t{NRegs: 0, NPis: 1, NBits: 1, PData: []uint32{1}, IFrame: 0}
		got := decodeSafetyWitness(g, 0, cex, 0)
		if got.Status != -1 || got.Error == nil {
			t.Fatalf("illegal trace was accepted: %+v", got)
		}
	})

	t.Run("selected output never fails", func(t *testing.T) {
		g := aig.NewGraph()
		x := g.NewInput()
		g.AddOutput(x)
		cex := &util.Abc_Cex_t{NRegs: 0, NPis: 1, NBits: 1, PData: []uint32{0}, IFrame: 0}
		got := decodeSafetyWitness(g, 0, cex, 0)
		if got.Status != -1 || got.Error == nil {
			t.Fatalf("nonfailing trace was accepted: %+v", got)
		}
	})
}
