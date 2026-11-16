package if_

import (
	"fmt"
	"math"
	"slices"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
)

func TestIfSwitchingSourceObjectOrder(t *testing.T) {
	for mask := 0; mask < 16; mask++ {
		t.Run(fmt.Sprintf("phase%d", mask), func(t *testing.T) {
			p := If_ManStart(&If_Par_t{NLutSize: 4, NCutsMax: 4, FVerbose: int32(mask & 1)})
			defer If_ManStop(p)
			graph := Gia_ManStart(16)
			defer Gia_ManStop(graph)
			a, b := If_ManCreateCi(p), If_ManCreateCi(p)
			x, y := Gia_ManAppendCi(graph), Gia_ManAppendCi(graph)
			first := If_ManCreateAnd(p, If_NotCond(a, int32(mask&1)), If_NotCond(b, int32(mask>>1&1)))
			node := gia.Gia_ManAppendAnd(graph, x^int32(mask&1), y^int32(mask>>1&1))
			root := If_ManCreateAnd(p, If_NotCond(first, int32(mask>>2&1)), a)
			lit := gia.Gia_ManAppendAnd(graph, node^int32(mask>>2&1), x)
			// Preserve a duplicate, dangling node: simulation must not hash it away.
			If_ManCreateAnd(p, a, b)
			gia.Gia_ManAppendAnd(graph, x, y)
			for _, output := range []struct {
				obj *If_Obj_t
				lit int32
			}{
				{If_NotCond(root, int32(mask>>3&1)), lit ^ int32(mask>>3&1)},
				{If_Not(root), lit ^ 1}, {If_Not(a), x ^ 1},
				{If_ManConst1(p), 1}, {If_Not(If_ManConst1(p)), 0},
			} {
				If_ManCreateCo(p, output.obj)
				Gia_ManAppendCo(graph, output.lit)
			}
			want := gia.Gia_ManComputeSwitchProbs(graph, 48, 16, 0)
			defer gia.Vec_IntFree(want)
			output := dsdReportCapture(t, func() { If_ManComputeSwitching(p) })
			if !slices.Equal(Vec_IntArray(p.vSwitching), gia.Vec_IntArray(want)) {
				t.Fatalf("switching object order differs: got %v want %v", Vec_IntArray(p.vSwitching), gia.Vec_IntArray(want))
			}
			if strings.Count(output, "Computing switching activity") != mask&1 {
				t.Fatalf("unexpected verbose report %q", output)
			}
			for _, bits := range Vec_IntArray(p.vSwitching) {
				value := math.Float32frombits(uint32(bits))
				if math.IsNaN(float64(value)) || value < 0 || value > 1 {
					t.Fatalf("invalid probability %g", value)
				}
			}
		})
	}
}

func TestMappingPowerDefaultSwitching(t *testing.T) {
	for _, override := range []bool{false, true} {
		name := "default"
		if override {
			name = "override"
		}
		t.Run(name, func(t *testing.T) {
			src := driverCallbackReviewGraph(t, false, false)
			probabilities := gia.Gia_ManComputeSwitchProbs(src, 48, 16, 0)
			defer gia.Vec_IntFree(probabilities)
			want := slices.Clone(gia.Vec_IntArray(probabilities))
			var pars If_Par_t
			Gia_ManSetIfParsDefault(&pars)
			pars.NLutSize, pars.FTruth, pars.FDeriveLuts, pars.FPower = 4, 1, 1, 1
			providers, writes := 0, 0
			if override {
				values := make([]float32, len(want))
				for i := range want {
					values[i] = 0.125
					want[i] = int32(math.Float32bits(values[i]))
				}
				pars.SetSwitchingProvider(func(p *If_Man_t) {
					providers++
					p.SetSwitchingProbabilities(values)
				})
			}
			pars.pFuncWrite = func(p *If_Man_t) int32 {
				writes++
				if !slices.Equal(Vec_IntArray(p.vSwitching), want) {
					t.Errorf("switching vector = %v, want %v", Vec_IntArray(p.vSwitching), want)
				}
				var power float32
				for _, entry := range Vec_PtrArray(p.vObjs) {
					obj := entry.(*If_Obj_t)
					if obj.Type != IF_AND || obj.nRefs == 0 {
						continue
					}
					for _, id := range ifLeaves(If_ObjCutBest(obj)) {
						power += math.Float32frombits(uint32(want[id]))
					}
				}
				if p.dPower != power {
					t.Errorf("power %g, want %g", p.dPower, power)
				}
				return 0
			}
			defer func() {
				if caught := recover(); caught != nil {
					t.Fatalf("power mapping panicked: %v", caught)
				}
			}()
			result := Gia_ManPerformMappingInt(src, &pars)
			if result == nil {
				t.Fatal("power mapping failed")
			}
			defer Gia_ManStop(result)
			if writes != 1 || override && providers != 1 || !override && providers != 0 {
				t.Fatalf("write calls %d, provider calls %d", writes, providers)
			}
			for x := 0; x < 8; x++ {
				want := x&1 != 0 && x&2 == 0 && x&4 != 0
				got, err := result.Evaluate([]bool{x&1 != 0, x&2 != 0, x&4 != 0})
				if err != nil || !slices.Equal(got, []bool{want, !want}) {
					t.Fatalf("assignment %d: %v, %v", x, got, err)
				}
			}
		})
	}
}

// Switching vectors store float32 bit patterns in integer slots, not counts.
func TestMappingPowerUsesProbabilities(t *testing.T) {
	p := If_ManStart(&If_Par_t{NLutSize: 4, NCutsMax: 4})
	defer If_ManStop(p)
	a, b := If_ManCreateCi(p), If_ManCreateCi(p)
	root := If_ManCreateAnd(p, a, b)
	root.CutBest = If_Cut_t{nLimit: 4, nLeaves: 2, pLeaves: []int32{a.Id, b.Id}}
	If_ManCreateCo(p, root)
	If_ManCreateCo(p, If_Not(root))
	p.SetSwitchingProbabilities([]float32{0, 0.25, 0.5, 0.375, 0.375, 0.375})
	for pass := 0; pass < 2; pass++ {
		If_ManMarkMapping(p)
		if p.dPower != 0.75 {
			t.Fatalf("pass %d: power = %g, want 0.75", pass, p.dPower)
		}
		if root.nRefs != 2 || a.nRefs != 1 || b.nRefs != 1 || p.nNets != 2 {
			t.Fatal("shared output cone was counted more than once")
		}
	}
}
