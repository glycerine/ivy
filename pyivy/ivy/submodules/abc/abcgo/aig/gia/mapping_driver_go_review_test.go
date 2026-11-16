package gia_test

import (
	"fmt"
	"os"
	"slices"
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
	if_ "github.com/glycerine/abcgo/map/if"
)

func mappingDriverReviewRun(t *testing.T, src *gia.Gia_Man_t, seed, mode int) *gia.Gia_Man_t {
	return mappingDriverReviewRunConfigured(t, src, seed, mode, false)
}

func mappingDriverReviewRunConfigured(t *testing.T, src *gia.Gia_Man_t, seed, mode int, filterDSD bool) *gia.Gia_Man_t {
	t.Helper()
	progress, err := os.CreateTemp(t.TempDir(), "mapper-progress")
	if err != nil {
		t.Fatal(err)
	}
	original := os.Stdout
	os.Stdout = progress
	defer func() { os.Stdout = original; progress.Close() }()
	if mode == 4 {
		return if_.Gia_ManPerformSopBalance(src, 8, 0, 0)
	}
	if mode == 5 {
		return if_.Gia_ManPerformDsdBalance(src, 6, 8, 0, 0)
	}
	var pars if_.If_Par_t
	if_.Gia_ManSetIfParsDefault(&pars)
	if mode >= 6 {
		size := int32(mode - 4)
		pars.NLutSize = 2*size - 1
		pars.FTruth, pars.FDeriveLuts, pars.FUseDsd, pars.FCutMin = 1, 1, 1, 1
		pars.FExpRed = 0
		if filterDSD {
			// The source's cell-filter branch requires cut storage of at least four
			// inputs, even when the target consists of smaller physical LUTs.
			pars.NLutSize = max(pars.NLutSize, 4)
			pars.PFuncCell = if_.If_CutPerformCheck16
		}
		pars.PLutStruct = "22"
		if size == 3 {
			pars.PLutStruct = "33"
		}
		dsd := if_.If_DsdManAlloc(pars.NLutSize, size)
		defer if_.If_DsdManFree(dsd, 0)
		pars.SetDsdManager(dsd)
		return if_.Gia_ManPerformMapping(src, &pars)
	}
	pars.NLutSize = int32(3 + ((seed >> 1) & 1))
	pars.FHashMapping = int32(mode & 1)
	if mode >= 2 {
		pars.FTruth, pars.FDeriveLuts = 1, 1
	}
	return if_.Gia_ManPerformMapping(src, &pars)
}

// Exercise the complete driver matrix even on hosts without C compilers or
// ABC object archives. The prior source comparisons are additional evidence;
// these scalar truth, interface and metadata checks must never skip with them.
// Structured modes enable the source DSD mark filter so infeasible cuts are
// rejected during selection, rather than reaching two-LUT reconstruction.
func TestMappingDriversGoOnlyTruthAndMetadata(t *testing.T) {
	for seed := 1; seed <= 64; seed++ {
		for mode := 0; mode < 8; mode++ {
			t.Run(fmt.Sprintf("seed%d/mode%d", seed, mode), func(t *testing.T) {
				src := gia.NewMappingDriverReviewGraphForTest(uint32(seed))
				defer gia.Gia_ManStop(src)
				result := mappingDriverReviewRunConfigured(t, src, seed, mode, true)
				if result == nil {
					t.Fatal("mapping failed for a valid input graph")
				}
				defer gia.Gia_ManStop(result)
				if gia.Gia_ManCiNum(result) != gia.Gia_ManCiNum(src) || gia.Gia_ManCoNum(result) != gia.Gia_ManCoNum(src) || gia.Gia_ManRegNum(result) != gia.Gia_ManRegNum(src) {
					t.Fatal("mapping changed the CI/CO/register interface")
				}
				if name, spec := result.Identity(); name != "driver-review" || spec != "driver-spec" {
					t.Fatalf("mapping identity %q/%q", name, spec)
				}
				for assignment := 0; assignment < 16; assignment++ {
					inputs := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0, assignment&8 != 0}
					want, err := src.Evaluate(inputs)
					if err != nil {
						t.Fatal(err)
					}
					got, err := result.Evaluate(inputs)
					if err != nil {
						t.Fatal(err)
					}
					if !slices.Equal(got, want) {
						t.Fatalf("assignment%d got%v want%v", assignment, got, want)
					}
				}
				if mode != 4 && mode != 5 {
					if result.VMapping == nil {
						t.Fatal("LUT mapping is missing")
					}
					if result.VConfigs != nil || result.PCellStr != "" {
						t.Fatal("general mapping retained obsolete cell configuration")
					}
				} else if result.VConfigs == nil || result.PCellStr != "driver-cell" {
					t.Fatal("balancing lost cell configuration")
				}
			})
		}
	}
}

func mappingGoReviewCapture(t *testing.T, run func()) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "choices")
	if err != nil {
		t.Fatal(err)
	}
	old := os.Stdout
	os.Stdout = f
	defer func() { os.Stdout = old; f.Close() }()
	run()
	data, err := os.ReadFile(f.Name())
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestMappingGiaToIfChoiceChainsGoOnly(t *testing.T) {
	for seed := uint32(1); seed <= 64; seed++ {
		t.Run(fmt.Sprint(seed), func(t *testing.T) {
			src := gia.NewMappingDriverReviewGraphForTest(seed)
			defer gia.Gia_ManStop(src)
			var pars if_.If_Par_t
			if_.Gia_ManSetIfParsDefault(&pars)
			pars.NLutSize = 4
			p := if_.Gia_ManToIf(src, &pars)
			defer if_.If_ManStop(p)
			if if_.If_ManObjNum(p) != gia.Gia_ManObjNum(src) {
				t.Fatal("ordinary bridge changed source object correspondence")
			}
			expectedReduction := int32(1)
			if seed&1 != 0 {
				expectedReduction = 0
			}
			if pars.FExpRed != expectedReduction {
				t.Fatalf("choice expansion policy %d, want%d", pars.FExpRed, expectedReduction)
			}
			for id := int32(0); id < gia.Gia_ManObjNum(src); id++ {
				obj := gia.Gia_ManObj(src, id)
				if gia.Gia_ObjValue(obj) != id {
					t.Fatalf("source object%d has IF copy%d", id, gia.Gia_ObjValue(obj))
				}
				if gia.Gia_ObjMark0(obj) != 0 {
					t.Fatalf("source object%d retains driver mark", id)
				}
			}
			for repeat := 0; repeat < 2; repeat++ {
				output := mappingGoReviewCapture(t, func() { if_.If_ManCheckChoices(p) })
				if output != "There are no dangling objects.\n" {
					t.Fatalf("choice alternatives were not reachable: %q", output)
				}
			}
		})
	}
}
