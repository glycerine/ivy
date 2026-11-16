package gia

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	abcutil "github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

func isoDriverReviewGraph() *Gia_Man_t {
	graph := Gia_ManStart(10)
	a, b, c := Gia_ManAppendCi(graph), Gia_ManAppendCi(graph), Gia_ManAppendCi(graph)
	state := Gia_ManAppendCi(graph)
	next := Gia_ManAppendAnd(graph, Gia_ManAppendAnd(graph, a, b^1), c)
	Gia_ManAppendCo(graph, state)
	Gia_ManAppendCo(graph, next)
	Gia_ManSetRegNum(graph, 1)
	return graph
}

func TestIsoCexPipelineAllPIPermutationsAndInitialStates(t *testing.T) {
	for _, order := range [][]int32{{0, 1, 2}, {0, 2, 1}, {1, 0, 2}, {1, 2, 0}, {2, 0, 1}, {2, 1, 0}} {
		t.Run(fmt.Sprint(order), func(t *testing.T) {
			source := isoDriverReviewGraph()
			permutation := vec.Vec_IntAllocArrayCopy(order, len(order))
			permuted := Gia_ManDupPerm(source, permutation)
			combined := Gia_ManDupAppendNew(source, permuted)
			var groups, perms *Vec_Ptr_t
			var reduced *Gia_Man_t
			isoReviewCapture(t, func() { reduced = Gia_ManIsoReduce(combined, &groups, &perms, 0, 0, 0, 0) })
			if reduced == nil || Gia_ManPoNum(reduced) != 1 || groups == nil || vec.Vec_PtrSize(groups) != 1 || !slices.Equal(vec.Vec_IntArray(vec.Vec_PtrEntry(groups, 0).(*Vec_Int_t)), []int32{0, 1}) {
				t.Fatal("isomorphic registered copies did not reduce together")
			}
			old := vec.Vec_IntArray(vec.Vec_PtrEntry(perms, 0).(*Vec_Int_t))
			next := vec.Vec_IntArray(vec.Vec_PtrEntry(perms, 1).(*Vec_Int_t))
			if len(old) != 3 || len(next) != 3 {
				t.Fatal("driver lost full PI support")
			}
			for initial := 0; initial < 2; initial++ {
				for frames := int32(1); frames <= 3; frames++ {
					for sequence := 0; sequence < 1<<uint(3*frames); sequence++ {
						cex := abcutil.Abc_CexAlloc(1, 3, frames)
						cex.IPo, cex.IFrame = 0, frames-1
						cex.PData[0] = uint32(sequence<<1 | initial)
						before := slices.Clone(cex.PData)
						mapped := abcutil.Abc_CexPermuteTwo(cex, old, next)
						if mapped.PData[0]&1 != uint32(initial) {
							t.Fatal("CEX remapping changed the initial register state")
						}
						// The unequal input depths and negative b edge make this graph's
						// structural PI correspondence unique, independently of the metadata.
						for frame := int32(0); frame < frames; frame++ {
							for slot, original := range order {
								actual := mapped.PData[0]&(1<<uint(1+3*frame+int32(slot))) != 0
								expected := sequence&(1<<uint(3*frame+original)) != 0
								if actual != expected {
									t.Fatalf("initial%d frames%d sequence%d PI%d mapped incorrectly", initial, frames, sequence, slot)
								}
							}
						}
						expected := initial != 0
						if frames > 1 {
							previous := (sequence >> uint(3*(frames-2))) & 7
							expected = previous&1 != 0 && previous&2 == 0 && previous&4 != 0
						}
						if (Gia_ManVerifyCex(source, cex, 0) != 0) != expected || (Gia_ManVerifyCex(permuted, mapped, 0) != 0) != expected {
							t.Fatalf("initial%d frames%d sequence%d counterexample validity changed", initial, frames, sequence)
						}
						if !slices.Equal(cex.PData, before) {
							t.Fatal("CEX mapping/replay mutated the original witness")
						}
					}
				}
			}
		})
	}
}

func TestIsoRandomPermutationKeepsSourceSwapDomain(t *testing.T) {
	for _, size := range []int32{0, 1, 2, 3, 31, 33, 257} {
		result := Gia_IsoTestGenPerm(size)
		if vec.Vec_IntSize(result) != int(size) {
			t.Fatalf("size%d wrong permutation length", size)
		}
		seen := make([]bool, size)
		for _, value := range vec.Vec_IntArray(result) {
			if value < 0 || value >= size || seen[value] {
				t.Fatalf("size%d invalid permutation%v", size, vec.Vec_IntArray(result))
			}
			seen[value] = true
		}
	}
}

func TestIsoCexDriverUsesSourcePermutationReport(t *testing.T) {
	for _, pis := range []int32{0, 1} {
		for _, frames := range []int32{1, 2} {
			for initial := 0; initial < 2; initial++ {
				t.Run(fmt.Sprintf("PIs%d/frames%d/initial%d", pis, frames, initial), func(t *testing.T) {
					graph := Gia_ManStart(5)
					driver := int32(0)
					if pis != 0 {
						driver = Gia_ManAppendCi(graph)
					}
					state := Gia_ManAppendCi(graph)
					Gia_ManAppendCo(graph, state)
					Gia_ManAppendCo(graph, driver)
					Gia_ManSetRegNum(graph, 1)
					cex := abcutil.Abc_CexAlloc(1, pis, frames)
					cex.IPo, cex.IFrame = 0, frames-1
					cex.PData[0] = uint32(initial)
					for bit := int32(1); bit < cex.NBits; bit++ {
						cex.PData[0] |= 1 << uint(bit)
					}
					expected := initial != 0
					if frames > 1 {
						expected = pis != 0
					}
					status := "valid"
					if !expected {
						status = "not valid"
					}
					got := isoReviewCapture(t, func() { Gia_IsoTest(graph, cex, 0) })
					entries := ""
					if pis != 0 {
						entries = " 0"
					}
					prefix := fmt.Sprintf("Considering random permutation of the primary inputs of the AIG:\nVector has %d entries: {%s }\n", pis, entries)
					if !strings.HasPrefix(got, prefix) {
						t.Errorf("missing source permutation report: %q", got)
					}
					suffix := fmt.Sprintf("CEX for the init AIG is %s.\nCEX for the perm AIG is %s.\n", status, status)
					if !strings.HasSuffix(got, suffix) {
						t.Errorf("driver remapping/replay validity changed: %q", got)
					}
				})
			}
		}
	}
}

func TestIsoOldDriverKeepsSourceLabelCountSemantics(t *testing.T) {
	for outputs := int32(1); outputs <= 2; outputs++ {
		for verbose := int32(0); verbose < 2; verbose++ {
			graph := iso2ReviewInputs(1)
			if outputs == 2 {
				Gia_ManAppendCo(graph, 2)
			}
			got := isoReviewCapture(t, func() { Gia_IsoTestOld(graph, verbose) })
			// The old driver runs full unique labeling, not forward-only candidate
			// grouping, so even duplicate POs become separate labeled singletons.
			expected := fmt.Sprintf("Reduced %d outputs to %d.  Time = ", outputs, outputs)
			if !strings.Contains(got, expected) || strings.Contains(got, "Nontrivial classes:") {
				t.Fatalf("old driver changed source labeling/report policy: %q", got)
			}
		}
	}
}

func TestIsoCexDriverHandlesUnusedGlobalPIs(t *testing.T) {
	for _, allUnused := range []bool{false, true} {
		t.Run(fmt.Sprintf("allUnused%t", allUnused), func(t *testing.T) {
			defer func() {
				if value := recover(); value != nil {
					t.Fatalf("unused global PIs caused driver panic: %v", value)
				}
			}()
			graph := Gia_ManStart(8)
			a := Gia_ManAppendCi(graph)
			Gia_ManAppendCi(graph) // Unused middle input must not shift local PI indices.
			c := Gia_ManAppendCi(graph)
			state := Gia_ManAppendCi(graph)
			next := int32(0)
			frames := int32(1)
			if !allUnused {
				next = Gia_ManAppendAnd(graph, a, c^1)
				frames = 2
			}
			Gia_ManAppendCo(graph, state)
			Gia_ManAppendCo(graph, next)
			Gia_ManSetRegNum(graph, 1)
			cex := abcutil.Abc_CexAlloc(1, 3, frames)
			cex.IPo, cex.IFrame = 0, frames-1
			cex.PData[0] = 1 | 2 | 4 // Initial register=1; first inputs a=1,b=1,c=0.
			got := isoReviewCapture(t, func() { Gia_IsoTest(graph, cex, 0) })
			if !strings.HasSuffix(got, "CEX for the init AIG is valid.\nCEX for the perm AIG is valid.\n") {
				t.Fatalf("unused-input CEX was not preserved: %q", got)
			}
		})
	}
}

func TestIsoCexCanonicalOrdersLiftNonPrefixSupport(t *testing.T) {
	for _, order := range [][]int32{{0, 1, 2}, {0, 2, 1}, {1, 0, 2}, {1, 2, 0}, {2, 0, 1}, {2, 1, 0}} {
		t.Run(fmt.Sprint(order), func(t *testing.T) {
			graph := Gia_ManStart(8)
			a := Gia_ManAppendCi(graph)
			Gia_ManAppendCi(graph)
			c := Gia_ManAppendCi(graph)
			state := Gia_ManAppendCi(graph)
			next := Gia_ManAppendAnd(graph, a, c^1)
			Gia_ManAppendCo(graph, state)
			Gia_ManAppendCo(graph, next)
			Gia_ManSetRegNum(graph, 1)
			permuted := Gia_ManDupPerm(graph, vec.Vec_IntAllocArrayCopy(order, len(order)))
			combined := Gia_ManDupAppendNew(graph, permuted)
			var perms *Vec_Ptr_t
			isoReviewCapture(t, func() { Gia_ManIsoReduce(combined, nil, &perms, 0, 0, 0, 0) })
			oldLocal := vec.Vec_PtrEntry(perms, 0).(*Vec_Int_t)
			newLocal := vec.Vec_PtrEntry(perms, 1).(*Vec_Int_t)
			beforeOld, beforeNew := slices.Clone(vec.Vec_IntArray(oldLocal)), slices.Clone(vec.Vec_IntArray(newLocal))
			if len(beforeOld) != 2 || len(beforeNew) != 2 {
				t.Fatal("fixture did not retain cone-local support size")
			}
			old, nextOrder := isoCexPIOrder(graph, oldLocal), isoCexPIOrder(permuted, newLocal)
			if len(old) != 3 || len(nextOrder) != 3 {
				t.Fatal("canonical order was not lifted to full PI interface")
			}
			for initial := 0; initial < 2; initial++ {
				for sequence := 0; sequence < 64; sequence++ {
					cex := abcutil.Abc_CexAlloc(1, 3, 2)
					cex.IPo, cex.IFrame = 0, 1
					cex.PData[0] = uint32(sequence<<1 | initial)
					mapped := abcutil.Abc_CexPermuteTwo(cex, old, nextOrder)
					if mapped.PData[0]&1 != uint32(initial) {
						t.Fatal("lifting lost initial register state")
					}
					for frame := 0; frame < 2; frame++ {
						for slot, original := range order {
							if ((mapped.PData[0] >> uint(1+3*frame+slot)) & 1) != uint32((sequence>>uint(3*frame+int(original)))&1) {
								t.Fatal("non-prefix support was mapped using local indices")
							}
						}
					}
					expected := sequence&1 != 0 && sequence&4 == 0
					if (Gia_ManVerifyCex(graph, cex, 0) != 0) != expected || (Gia_ManVerifyCex(permuted, mapped, 0) != 0) != expected {
						t.Fatal("lifted witness changed validity")
					}
				}
			}
			if !slices.Equal(vec.Vec_IntArray(oldLocal), beforeOld) || !slices.Equal(vec.Vec_IntArray(newLocal), beforeNew) {
				t.Fatal("lifting mutated library cone-local metadata")
			}
		})
	}
}
