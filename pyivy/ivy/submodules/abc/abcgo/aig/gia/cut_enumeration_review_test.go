package gia

import (
	"fmt"
	"os/exec"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/misc/vec"
)

func TestCutReportsMatchOriginalABC(t *testing.T) {
	exe := compileGiaSourceOracle(t, "cut_enumeration_review.c")
	for _, mode := range []string{"reports", "driver"} {
		t.Run(mode, func(t *testing.T) {
			native, err := exec.Command(exe, mode).CombinedOutput()
			if err != nil {
				t.Fatalf("native cut report: %v\n%s", err, native)
			}
			g := cutEnumReviewGraph(0)
			got := hashReviewStdout(t, func() {
				if mode == "driver" {
					Gia_StoComputeCuts(g)
					return
				}
				p := Gia_StoAlloc(g, 4, 10, 1, 1, 0)
				defer Gia_StoFree(p)
				cut := Gia_Cut_t{nLeaves: 2, pLeaves: [8]int32{1, 12}, Cost: 7, CostLev: 9, nTreeLeaves: 1}
				Gia_CutPrint(p, 3, &cut)
				Gia_CutPrint(p, 3, nil)
			})
			want := string(native)
			if mode == "driver" {
				timing := regexp.MustCompile(`Time = +[0-9]+\.[0-9]{2} sec\n$`)
				if !timing.MatchString(got) {
					t.Fatalf("missing source-format elapsed report: %q", got)
				}
				got, want = timing.ReplaceAllString(got, "Time = <elapsed>\n"), timing.ReplaceAllString(want, "Time = <elapsed>\n")
			}
			if got != want {
				t.Fatalf("cut report differs\nGo:%s\nC:%s", got, want)
			}
		})
	}
}

func BenchmarkCutTruthScratch(b *testing.B) {
	for _, size := range []int32{4, 8} {
		b.Run(fmt.Sprint(size), func(b *testing.B) {
			g := Gia_ManStart(16)
			for i := 0; i < 8; i++ {
				Gia_ManAppendCi(g)
			}
			p := Gia_StoAlloc(g, size, 10, 1, 1, 0)
			defer Gia_StoFree(p)
			words := 1 << uint(max(0, size-6))
			left, right := make([]uint64, words), make([]uint64, words)
			for i := range left {
				left[i] = 0x8888888888888888
				right[i] = 0x6666666666666666
			}
			a := Gia_Cut_t{nLeaves: 2, pLeaves: [8]int32{1, 3}, iFunc: int32(vec.Vec_MemHashInsert(p.vTtMem, left) << 1)}
			c := Gia_Cut_t{nLeaves: 2, pLeaves: [8]int32{2, 3}, iFunc: int32(vec.Vec_MemHashInsert(p.vTtMem, right) << 1)}
			initial := Gia_Cut_t{nLeaves: 3, pLeaves: [8]int32{1, 2, 3}}
			out := initial
			Gia_CutComputeTruth(p, &a, &c, 0, 0, &out, 0)
			if out.nLeaves != 3 {
				b.Fatal("benchmark minimized relevant variables")
			}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				out = initial
				Gia_CutComputeTruth(p, &a, &c, 0, 0, &out, 0)
			}
		})
	}
}

// The driver-only report parser is also used to check finite statistics for
// supported empty interfaces without requiring C's undefined zero divisions.
func cutEnumReportLine(report, prefix string) string {
	for _, line := range strings.Split(report, "\n") {
		if strings.HasPrefix(line, prefix) {
			return line
		}
	}
	return ""
}

func TestCutEmptyInterfacesHaveFiniteStatistics(t *testing.T) {
	for _, kind := range []string{"empty", "input", "constant-output"} {
		t.Run(kind, func(t *testing.T) {
			g := Gia_ManStart(16)
			if kind == "input" {
				Gia_ManAppendCi(g)
			}
			if kind == "constant-output" {
				Gia_ManAppendCo(g, 1)
			}
			report := hashReviewStdout(t, func() { Gia_StoComputeCuts(g) })
			if strings.Contains(report, "NaN") || strings.Contains(report, "Inf") {
				t.Fatalf("nonfinite statistics: %s", report)
			}
			line := cutEnumReportLine(report, "CutPair =")
			if !strings.Contains(line, "Cut/Node = 0.00") || !strings.Contains(line, "Merge = 0 (0.00 %)") {
				t.Fatalf("missing finite empty-interface statistics: %s", report)
			}
		})
	}
}

func TestCutEnumerationPreservesBorrowedGraph(t *testing.T) {
	g := cutEnumReviewGraph(4)
	Gia_ManCreateRefs(g)
	references := slices.Clone(g.Refs)
	for _, obj := range g.pObjs {
		obj.Value = 1234
		obj.fMark0 = 1
		obj.fMark1 = 1
		obj.fPhase = 1
	}
	p := cutEnumReviewEnumerate(g, 8, 25, 1, 1)
	Gia_StoFree(p)
	Gia_StoFree(p)
	Gia_StoFree(nil)
	if p.pGia != nil || p.vTtMem != nil || p.vCuts != nil || p.vRefs != nil {
		t.Fatal("freed store retains ownership")
	}
	if !slices.Equal(references, g.Refs) {
		t.Fatal("private cut references changed graph references")
	}
	for _, obj := range g.pObjs {
		if obj.Value != 1234 || obj.fMark0 != 1 || obj.fMark1 != 1 || obj.fPhase != 1 {
			t.Fatal("cut enumeration changed borrowed graph scratch")
		}
	}
}

func TestCutRejectsBuffersAndMuxes(t *testing.T) {
	for _, mux := range []bool{false, true} {
		g := Gia_ManStart(16)
		a, b, c := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
		if mux {
			g.EnableMuxes()
			Gia_ManAppendMuxReal(g, a, b, c)
		} else {
			Gia_ManAppendBuf(g, a)
		}
		cexReviewMustPanic(t, func() { p := cutEnumReviewEnumerate(g, 6, 10, 1, 1); Gia_StoFree(p) })
	}
	for _, parameters := range [][2]int32{{0, 10}, {1, 10}, {9, 10}, {4, 0}, {4, 1}, {4, 51}} {
		cexReviewMustPanic(t, func() { Gia_StoAlloc(Gia_ManStart(16), parameters[0], parameters[1], 1, 1, 0) })
	}
}

func TestCutTruthInputsAndFormalComposition(t *testing.T) {
	for _, size := range []int32{3, 6, 7, 8} {
		for minimum := int32(0); minimum < 2; minimum++ {
			g := Gia_ManStart(16)
			for i := 0; i < 8; i++ {
				Gia_ManAppendCi(g)
			}
			p := Gia_StoAlloc(g, size, 10, 1, minimum, 0)
			a := Gia_Cut_t{nLeaves: 2, pLeaves: [8]int32{1, 3}}
			c := Gia_Cut_t{nLeaves: 2, pLeaves: [8]int32{2, 3}}
			union := []int32{1, 2, 3}
			if size == 8 {
				a.nLeaves = 5
				a.pLeaves = [8]int32{1, 2, 4, 7, 8}
				c.nLeaves = 5
				c.pLeaves = [8]int32{1, 3, 5, 6, 8}
				union = []int32{1, 2, 3, 4, 5, 6, 7, 8}
			}
			words := 1 << uint(max(0, size-6))
			left, right := make([]uint64, words), make([]uint64, words)
			random := uint32(size + minimum + 613)
			for side, table := range [][]uint64{left, right} {
				arity := a.nLeaves
				if side != 0 {
					arity = c.nLeaves
				}
				var pattern uint64
				for row := 1; row < 1<<arity; row++ {
					random = random*1664525 + 1013904223
					if random>>31 != 0 {
						pattern |= 1 << uint(row)
					}
				}
				for row := 0; row < 64*words; row++ {
					if pattern&(1<<uint(row%int(1<<arity))) != 0 {
						table[row>>6] |= 1 << uint(row&63)
					}
				}
			}
			a.iFunc = int32(vec.Vec_MemHashInsert(p.vTtMem, left) << 1)
			c.iFunc = int32(vec.Vec_MemHashInsert(p.vTtMem, right) << 1)
			beforeA, beforeC := a, c
			inputValue := func(table []uint64, leaves []int32, row int) bool {
				index := 0
				for bit, leaf := range leaves {
					if row&(1<<uint(slices.Index(union, leaf))) != 0 {
						index |= 1 << uint(bit)
					}
				}
				return table[index>>6]&(1<<uint(index&63)) != 0
			}
			for phaseA := int32(0); phaseA < 2; phaseA++ {
				for phaseC := int32(0); phaseC < 2; phaseC++ {
					for xor := int32(0); xor < 2; xor++ {
						out := Gia_Cut_t{nLeaves: uint32(len(union))}
						copy(out.pLeaves[:], union)
						reduced := Gia_CutComputeTruth(p, &a, &c, phaseA, phaseC, &out, xor)
						if reduced != giaBool(int(out.nLeaves) < len(union)) || a != beforeA || c != beforeC {
							t.Fatal("composition changed input cut or reduction contract")
						}
						if !slices.Equal(Gia_CutTruth(p, &a), left) || !slices.Equal(Gia_CutTruth(p, &c), right) {
							t.Fatal("composition changed borrowed input truth")
						}
						result := Gia_CutTruth(p, &out)
						for row := 0; row < 1<<uint(len(union)); row++ {
							va := inputValue(left, a.pLeaves[:a.nLeaves], row) != (phaseA != 0)
							vc := inputValue(right, c.pLeaves[:c.nLeaves], row) != (phaseC != 0)
							want := va && vc
							if xor != 0 {
								want = va != vc
							}
							index := 0
							for bit, leaf := range out.pLeaves[:out.nLeaves] {
								if row&(1<<uint(slices.Index(union, leaf))) != 0 {
									index |= 1 << uint(bit)
								}
							}
							got := (result[index>>6]&(1<<uint(index&63)) != 0) != (out.iFunc&1 != 0)
							if got != want {
								t.Fatalf("size%d minimize%d phases%d,%d xor%d row%d", size, minimum, phaseA, phaseC, xor, row)
							}
						}
					}
				}
			}
			Gia_StoFree(p)
		}
	}
}
