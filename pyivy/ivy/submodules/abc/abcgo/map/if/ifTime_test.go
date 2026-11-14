package if_

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/glycerine/abcgo/misc/tim"
)

func ifTimeTestManager(arrivals []float32) (*If_Man_t, *If_Cut_t, []*If_Obj_t) {
	pars := &If_Par_t_{NLutSize: 4, NCutsMax: 4, NAndDelay: 2, DelayTarget: -1}
	p := If_ManStart(pars)
	var leaves []*If_Obj_t
	var ids []int32
	for _, arrival := range arrivals {
		leaf := If_ManCreateCi(p)
		leaf.CutBest.Delay = arrival
		leaf.Required = IF_FLOAT_LARGE
		leaves = append(leaves, leaf)
		ids = append(ids, leaf.Id)
	}
	return p, &If_Cut_t{nLeaves: uint32(len(ids)), pLeaves: ids}, leaves
}
func TestCutVariablePinTimingVector(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("test_vectors", "pin_timing.json"))
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Arrivals     []float32
		PinDelays    []float32 `json:"pin_delays"`
		Arrival      float32
		Required     float32   `json:"output_required"`
		LeafRequired []float32 `json:"leaf_required"`
	}
	if err = json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	p, c, leaves := ifTimeTestManager(fixture.Arrivals)
	defer If_ManStop(p)
	library := If_LibLutSetSimple(3)
	library.fVarPinDelays = 1
	copy(library.pLutDelays[3][:], fixture.PinDelays)
	p.pPars.PLutLib = library
	if arrival := If_CutDelay(p, nil, c); arrival != fixture.Arrival {
		t.Fatal("arrival", arrival)
	}
	permutation, delays := make([]int32, 3), make([]float32, 3)
	If_CutSortInputPins(p, c, permutation, delays)
	if !reflect.DeepEqual(permutation, []int32{1, 2, 0}) || !reflect.DeepEqual(delays, fixture.Arrivals) {
		t.Fatal("pin assignment", permutation, delays)
	}
	If_CutPropagateRequired(p, nil, c, fixture.Required)
	for i, leaf := range leaves {
		if leaf.Required != fixture.LeafRequired[i] {
			t.Fatal("pin required", i, leaf.Required)
		}
	}
}
func TestCutAndUserAndLiftedTiming(t *testing.T) {
	p, c, leaves := ifTimeTestManager([]float32{2, 7, 3})
	defer If_ManStop(p)
	c.fAndCut = 1
	if If_CutDelay(p, nil, c) != 9 {
		t.Fatal("AND delay")
	}
	If_CutPropagateRequired(p, nil, c, 10)
	for _, leaf := range leaves {
		if leaf.Required != 8 {
			t.Fatal("AND required")
		}
		leaf.Required = IF_FLOAT_LARGE
	}
	c.fAndCut = 0
	c.fUser = 1
	c.permutation = []byte{3, 1, 2}
	if If_CutDelay(p, nil, c) != 8 {
		t.Fatal("user pin delay")
	}
	If_CutPropagateRequired(p, nil, c, 10)
	if leaves[0].Required != 7 || leaves[1].Required != 9 || leaves[2].Required != 8 {
		t.Fatal("user required")
	}
	c.fUser = 0
	p.pPars.fLiftLeaves = 1
	p.Period = 10
	c.nLeaves = 2
	c.pLeaves = []int32{leaves[0].Id<<8 | 1, leaves[1].Id << 8}
	leaves[0].CutBest.Delay = 13
	leaves[1].CutBest.Delay = 4
	if If_CutDelay(p, nil, c) != 5 {
		t.Fatal("lifted delay")
	}
}
func ifTimeMappedPair() (*If_Man_t, *If_Obj_t, []*If_Obj_t) {
	p, _, leaves := ifTimeTestManager([]float32{0, 0})
	If_ManSetupCiCutSets(p)
	node := If_ManCreateAnd(p, leaves[0], leaves[1])
	node.CutBest = If_Cut_t{nLeaves: 2, pLeaves: []int32{leaves[0].Id, leaves[1].Id}, Delay: 1}
	If_ManCreateCo(p, node)
	If_ManCreateCo(p, node)
	return p, node, leaves
}
func TestRequiredSharedOutputAndNoInputSelfGate(t *testing.T) {
	p, node, leaves := ifTimeMappedPair()
	defer If_ManStop(p)
	p.pPars.pTimesReq = []float32{2, 5}
	If_ManComputeRequired(p)
	if p.RequiredGlo != 1 || node.Required != 2 {
		t.Fatal("shared output lost tighter constraint", p.RequiredGlo, node.Required)
	}
	for _, leaf := range leaves {
		if leaf.Required != 1 {
			t.Fatal("trivial CI cut counted as another gate", leaf.Required)
		}
	}
}
func TestRequiredTargetRelaxationAreaAndViolation(t *testing.T) {
	p, node, leaves := ifTimeMappedPair()
	defer If_ManStop(p)
	p.pPars.DelayTarget = 4
	If_ManComputeRequired(p)
	if p.RequiredGlo != 4 || node.Required != 4 || leaves[0].Required != 3 || p.fNextRound == 0 {
		t.Fatal("target relaxation", p.RequiredGlo, node.Required, leaves[0].Required)
	}
	p.pPars.DelayTarget = -1
	p.pPars.pTimesReq = []float32{.25, .5}
	If_ManComputeRequired(p)
	if node.Required != 1 || p.fReqTimeWarn == 0 {
		t.Fatal("infeasible output requirements")
	}
	p.pPars.pTimesReq = nil
	p.pPars.fAreaOnly = 1
	If_ManComputeRequired(p)
	if node.Required != IF_FLOAT_LARGE || leaves[0].Required != IF_FLOAT_LARGE {
		t.Fatal("area-only propagated timing")
	}
}
func TestRequiredExternalTimingManager(t *testing.T) {
	p, node, leaves := ifTimeMappedPair()
	defer If_ManStop(p)
	p.pManTim = tim.Tim_ManStart(2, 2)
	If_ManComputeRequired(p)
	if node.Required != 1 || leaves[0].Required != 0 || leaves[1].Required != 0 {
		t.Fatal("external manager required times", node.Required, leaves[0].Required, leaves[1].Required)
	}
}
