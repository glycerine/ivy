package gia_test

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
	if_ "github.com/glycerine/abcgo/map/if"
)

func TestMappingBoxedChoiceBridgeGoOnly(t *testing.T) {
	for signs := 0; signs < 256; signs++ {
		p := gia.NewMappingBoxChoiceGraphForTest(signs, 0)
		var pars if_.If_Par_t
		if_.Gia_ManSetIfParsDefault(&pars)
		pars.NLutSize = 4
		manager := if_.Gia_ManToIf(p, &pars)
		if pars.FExpRed != 0 || if_.If_ManObjNum(manager) != gia.Gia_ManObjNum(p) {
			t.Fatal("boxed choice policy/object correspondence changed")
		}
		links := 0
		for id := int32(1); id < gia.Gia_ManObjNum(p); id++ {
			obj := gia.Gia_ManObj(p, id)
			if gia.Gia_ObjMark0(obj) != 0 {
				t.Fatal("boxed bridge retained fanout marks")
			}
			if sibling := gia.Gia_ObjSibl(p, id); sibling != 0 {
				links++
				mapped := if_.If_ManObj(manager, gia.Gia_ObjValue(obj))
				other := mapped.Equivalent()
				if other == nil || if_.If_ObjId(other) != gia.Gia_ObjValue(gia.Gia_ManObj(p, sibling)) {
					t.Fatal("boxed choice link was lost")
				}
			}
		}
		if links != 2 {
			t.Fatal("fixture must expose two independent choice classes")
		}
		for i, want := range []int32{0, 0, 2, 2, 4} {
			if got := if_.If_ObjLevel(if_.If_ManCi(manager, int32(i))); got != want {
				t.Fatalf("CI%d boxed level%d want%d", i, got, want)
			}
		}
		if output := mappingGoReviewCapture(t, func() { if_.If_ManCheckChoices(manager) }); output != "There are no dangling objects.\n" {
			t.Fatalf("boxed choices: %q", output)
		}
		if_.If_ManStop(manager)
		gia.Gia_ManStop(p)
	}
}

func TestMappingBoxedChoiceDriversGoOnly(t *testing.T) {
	modes := []struct {
		name string
		kind int
		hash int32
	}{
		{"structural", 0, 0}, {"structural_hash", 0, 1},
		{"truth", 1, 0}, {"truth_hash", 1, 1},
		{"sop", 2, 0}, {"sop_hash", 2, 1},
		{"dsd", 3, 0}, {"dsd_hash", 3, 1},
		{"sop_wrapper", 4, 0}, {"dsd_wrapper", 5, 0},
	}
	for signs := 0; signs < 256; signs++ {
		for regs := int32(0); regs < 2; regs++ {
			for _, mode := range modes {
				t.Run(fmt.Sprintf("phase%d/regs%d/%s", signs, regs, mode.name), func(t *testing.T) {
					src := gia.NewMappingBoxChoiceGraphForTest(signs, regs)
					defer gia.Gia_ManStop(src)
					timing, _ := src.BoxTiming()
					var result *gia.Gia_Man_t
					output := mappingGoReviewCapture(t, func() {
						if mode.kind == 4 {
							if signs&1 == 0 {
								result = if_.Gia_ManPerformSopBalance(src, 8, 0, 0)
							} else {
								result = gia.Gia_ManPerformSopBalance(src, 8, 0, 0)
							}
							return
						}
						if mode.kind == 5 {
							if signs&1 == 0 {
								result = if_.Gia_ManPerformDsdBalance(src, 6, 8, 0, 0)
							} else {
								result = gia.Gia_ManPerformDsdBalance(src, 6, 8, 0, 0)
							}
							return
						}
						var pars if_.If_Par_t
						if_.Gia_ManSetIfParsDefault(&pars)
						pars.NLutSize = 4
						pars.FHashMapping = mode.hash
						if mode.kind != 0 {
							pars.FTruth, pars.FDeriveLuts = 1, 1
						}
						if mode.kind >= 2 {
							pars.NLutSize, pars.FExpRed, pars.FCutMin = 6, 0, 1
						}
						if mode.kind == 2 {
							pars.FDelayOpt = 1
						}
						if mode.kind == 3 {
							pars.FUseDsd, pars.FDsdBalance = 1, 1
							dsd := if_.If_DsdManAlloc(6, 0)
							defer if_.If_DsdManFree(dsd, 0)
							pars.SetDsdManager(dsd)
						}
						result = if_.Gia_ManPerformMapping(src, &pars)
					})
					if result == nil {
						t.Fatalf("boxed choices failed: %s", output)
					}
					defer gia.Gia_ManStop(result)
					if gia.Gia_ManCiNum(result) != 5 || gia.Gia_ManCoNum(result) != 6 || gia.Gia_ManRegNum(result) != regs || gia.Gia_ManIsNormalized(result) == 0 {
						t.Fatal("boxed choice interface/order changed")
					}
					gotTiming, _ := result.BoxTiming()
					if gotTiming != timing || src.PManTime != nil {
						t.Fatal("boxed choice timing ownership did not move")
					}
					if (result.VMapping != nil) != (mode.kind < 2) {
						t.Fatal("unexpected boxed choice result representation")
					}
					if strings.Contains(output, "Gia_ManMappingVerify:") {
						t.Fatalf("invalid mapping: %s", output)
					}
					wantWarnings := 0
					if regs != 0 && mode.kind < 4 {
						wantWarnings = 1
					}
					if strings.Count(output, "Warning: Shuffled CI order to be correct sequential AIG.\n") != wantWarnings {
						t.Fatal("boxed choice shuffle policy changed")
					}
					for x := 0; x < 32; x++ {
						a, b, u, v, w := x&1 != 0, x&2 != 0, x&4 != 0, x&8 != 0, x&16 != 0
						phase := func(bit uint) bool { return signs>>bit&1 != 0 }
						mid := (u != phase(3)) && (v != phase(4))
						want := []bool{(a != phase(0)) && (b != phase(1)), a && (b != phase(2)), mid, (w != phase(5)) && (mid != phase(6)), (w != phase(7)) && a, !b}
						inputs := []bool{a, b, u, v, w}
						// General normalization moves the FO b last; dedicated
						// balancing wrappers retain the source CI convention.
						if regs != 0 && mode.kind < 4 {
							inputs = []bool{a, u, v, w, b}
						}
						got, err := result.Evaluate(inputs)
						if err != nil || !slices.Equal(got, want) {
							t.Fatalf("input%d: got%v want%v err%v", x, got, want, err)
						}
					}
				})
			}
		}
	}
}
