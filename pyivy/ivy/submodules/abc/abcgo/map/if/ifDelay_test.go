package if_

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/glycerine/abcgo/misc/vec"
)

func ifDelayTestCover(cubes []int32) *Vec_Int_t { return vec.Vec_IntAllocArrayCopy(cubes, len(cubes)) }
func ifDelayTestTruth(cubes []int32, n int) uint64 {
	var truth uint64
	for minterm := 0; minterm < 64; minterm++ {
		match := false
		for _, cube := range cubes {
			term := true
			for pin := 0; pin < n; pin++ {
				literal := uint32(cube) >> uint(pin*2) & 3
				if literal == 1 && minterm>>uint(pin)&1 != 0 || literal == 2 && minterm>>uint(pin)&1 == 0 {
					term = false
				}
			}
			match = match || term
		}
		if match {
			truth |= 1 << uint(minterm)
		}
	}
	return truth
}
func TestSopDelayRecordedVectors(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("test_vectors", "sop_delay_cases.json"))
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct {
		Name         string
		Cubes, Times []int32
		Delay, Area  int32
	}
	if err = json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	for _, test := range cases {
		t.Run(test.Name, func(t *testing.T) {
			cover := ifDelayTestCover(test.Cubes)
			n := int32(len(test.Times))
			var area int32
			delay := If_CutSopBalanceEvalIntInt(cover, n, test.Times, nil, 0, &area)
			if delay != test.Delay || area != test.Area {
				t.Fatal("timing-only", delay, area)
			}
			aig := vec.Vec_IntAlloc(0)
			area = 0
			delay = If_CutSopBalanceEvalIntInt(cover, n, test.Times, aig, 0, &area)
			if delay != test.Delay || area != test.Area {
				t.Fatal("AIG cost", delay, area)
			}
			if got, want := If_AigVerifyArray(aig, n), ifDelayTestTruth(test.Cubes, int(n)); got != want {
				t.Fatalf("SOP AIG %016x != %016x", got, want)
			}
			permutation := make([]byte, n)
			if pins := If_CutSopBalancePinDelaysIntInt(cover, test.Times, n, permutation); pins != test.Delay {
				t.Fatal("pin timing", pins, permutation)
			}
			maximum := int32(0)
			for pin, time := range test.Times {
				maximum = max(maximum, time+int32(permutation[pin]))
			}
			if maximum != delay {
				t.Fatal("pin profile disagrees", maximum, delay, permutation)
			}
		})
	}
}
func TestSopDelayComplementAndFailure(t *testing.T) {
	cover := ifDelayTestCover([]int32{10, 33})
	times := []int32{0, 0, 0}
	for _, coverComplement := range []bool{false, true} {
		for outputComplement := int32(0); outputComplement < 2; outputComplement++ {
			aig := vec.Vec_IntAlloc(0)
			area := int32(0)
			if If_CutSopBalanceEvalIntInt(cover, 3, times, aig, outputComplement, &area, coverComplement) < 0 {
				t.Fatal("complemented derivation")
			}
			want := ifDelayTestTruth([]int32{10, 33}, 3)
			if coverComplement != (outputComplement != 0) {
				want = ^want
			}
			if If_AigVerifyArray(aig, 3) != want {
				t.Fatal("cover/output phase composition")
			}
		}
	}
	permutation := []byte{9, 9, 9}
	bad := ifDelayTestCover([]int32{3})
	if If_CutSopBalancePinDelaysIntInt(bad, times, 3, permutation) != -1 || !reflect.DeepEqual(permutation, []byte{9, 9, 9}) {
		t.Fatal("failed profile changed output")
	}
	duplicate := ifDelayTestCover([]int32{42, 42})
	if If_CutSopBalanceEvalIntInt(duplicate, 3, times, nil, 0, nil) != -1 {
		t.Fatal("duplicate cover")
	}
}
func ifDelayTestManager(t *testing.T) (*If_Man_t, *If_Cut_t) {
	t.Helper()
	pars := &If_Par_t_{NLutSize: 4, NCutsMax: 4, NGateSize: 70, PLutStruct: "44"}
	p := If_ManStart(pars)
	leaves := make([]int32, 3)
	for pin := range leaves {
		node := If_ManCreateCi(p)
		node.CutBest.Delay = float32(pin)
		leaves[pin] = node.Id
	}
	c := &If_Cut_t{nLeaves: 3, pLeaves: leaves, iCutFunc: 2}
	p.vTtIsops[3] = vec.Vec_WecStart(2)
	cover := vec.Vec_WecEntry(p.vTtIsops[3], 1)
	vec.Vec_IntPush(cover, 10)
	vec.Vec_IntPush(cover, 33)
	p.vTtIsopComplement[3] = []bool{false, true}
	return p, c
}
func TestSopDelayManagerMetadataAndSingleLut(t *testing.T) {
	p, c := ifDelayTestManager(t)
	defer If_ManStop(p)
	times := ifDelayTimes(p, c)
	aig := vec.Vec_IntAlloc(0)
	delay := If_CutSopBalanceEval(p, c, aig)
	if delay < 0 || c.Cost != 3 || c.fUser == 0 {
		t.Fatal("manager SOP evaluation", delay, c.Cost)
	}
	if If_AigVerifyArray(aig, 3) != ^ifDelayTestTruth([]int32{10, 33}, 3) {
		t.Fatal("ISOP metadata phase lost")
	}
	permutation := make([]byte, 3)
	if If_CutSopBalancePinDelays(p, c, permutation) != delay {
		t.Fatal("manager pin profiles")
	}
	if sop := If_CutDelaySop(p, c); sop < 0 || len(c.permutation) != 3 || c.Cost != 3 {
		t.Fatal("gate-level SOP cost", sop, c.Cost)
	}
	if lut := If_CutLutBalanceEval(p, c); lut != times[2]+1 || c.Cost != 1 {
		t.Fatal("single LUT timing", lut, c.Cost)
	}
	if If_CutLutBalancePinDelays(p, c, permutation) != times[2]+1 || !reflect.DeepEqual(permutation, []byte{1, 1, 1}) {
		t.Fatal("single LUT pin delays", permutation)
	}
}
