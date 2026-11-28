package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/internal/testoracle"
	"slices"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/misc/vec"
)

func recordRepeatedReview(t *testing.T, records *strings.Builder, out, src *Gia_Man_t, times int32) {
	t.Helper()
	if Gia_ManPiNum(out) != times*Gia_ManPiNum(src) || Gia_ManPoNum(out) != times*Gia_ManPoNum(src) || out.nRegs != times*src.nRegs || out.nObjs != 1+times*(src.nObjs-1) {
		t.Fatal("repeated duplication changed terminal grouping or source object counts")
	}
	if out.nMuxes != times*src.nMuxes || out.nXors != times*src.nXors || out.nBufs != times*src.nBufs {
		t.Fatal("repeated duplication changed real gate kinds")
	}
	for i := int32(0); i < Gia_ManCiNum(out); i++ {
		if Gia_ObjCioId(Gia_ManCi(out, i)) != i {
			t.Fatal("duplicated CI has stale ordinal")
		}
	}
	for i := int32(0); i < Gia_ManCoNum(out); i++ {
		if Gia_ObjCioId(Gia_ManCo(out, i)) != i {
			t.Fatal("duplicated CO has stale ordinal")
		}
	}
	name, spec := src.Identity()
	if gotName, gotSpec := out.Identity(); gotName != name || gotSpec != spec {
		t.Fatal("repeated duplication changed identity")
	}
	for mask := 0; mask < 1<<uint(Gia_ManCiNum(out)); mask++ {
		inputs := make([]bool, Gia_ManCiNum(out))
		for i := range inputs {
			inputs[i] = mask&(1<<uint(i)) != 0
		}
		var pos, ris []bool
		for copy := int32(0); copy < times; copy++ {
			sourceInputs := slices.Clone(inputs[copy*Gia_ManPiNum(src) : (copy+1)*Gia_ManPiNum(src)])
			start := Gia_ManPiNum(out) + copy*src.nRegs
			sourceInputs = append(sourceInputs, inputs[start:start+src.nRegs]...)
			values, err := src.Evaluate(sourceInputs)
			if err != nil {
				t.Fatal(err)
			}
			pos = append(pos, values[:Gia_ManPoNum(src)]...)
			ris = append(ris, values[Gia_ManPoNum(src):]...)
		}
		checkQuantificationOutputs(t, out, inputs, append(pos, ris...))
		values, err := out.Evaluate(inputs)
		if err != nil {
			t.Fatal(err)
		}
		packed := 0
		for i, value := range values {
			if value {
				packed |= 1 << uint(i)
			}
		}
		fmt.Fprintf(records, " %d", packed)
	}
	records.WriteByte('\n')
}

func timesClassReviewRecords(t *testing.T, dfs bool) string {
	t.Helper()
	var records strings.Builder
	for signs := int32(0); signs < 8; signs++ {
		for regs := int32(0); regs <= 2; regs++ {
			for times := int32(1); times <= 3; times++ {
				src := appendReviewGraph(signs, regs)
				if dfs {
					src = Gia_ManDupOrderDfs(src)
				}
				out := Gia_ManDupTimes(src, times)
				fmt.Fprintf(&records, "0 %d %d %d %d %d %d", signs, times, regs, Gia_ManCiNum(out), Gia_ManCoNum(out), out.nRegs)
				recordRepeatedReview(t, &records, out, src, times)
			}
		}
	}
	if dfs {
		return records.String()
	}
	labels := []int32{-2, 5, 9}
	for signs := int32(0); signs < 8; signs++ {
		for pattern := 0; pattern < 27; pattern++ {
			classes := []int32{labels[pattern%3], labels[(pattern/3)%3], labels[pattern/9]}
			for selection, class := range []int32{-2, 5, 9, 77} {
				src := blockReviewGraph(signs, 3)
				src.VFlopClasses = permutation(classes...)
				out := Gia_ManDupFlopClass(src, class)
				var order []int
				var selected int32
				for pass := 0; pass < 2; pass++ {
					for i, label := range classes {
						if (label == class) == (pass == 1) {
							order = append(order, i)
							if pass == 1 {
								selected++
							}
						}
					}
				}
				if out.nRegs != selected || Gia_ManPiNum(out) != 6-selected || Gia_ManPoNum(out) != 6-selected || out.nObjs != src.nObjs {
					t.Fatal("flop extraction changed source order or class interface")
				}
				if name, spec := out.Identity(); name != "block" || spec != "block.aig" {
					t.Fatal("flop extraction changed identity")
				}
				fmt.Fprintf(&records, "1 %d %d %d %d %d %d", signs, pattern, selection, Gia_ManCiNum(out), Gia_ManCoNum(out), out.nRegs)
				for mask := 0; mask < 64; mask++ {
					inputs, original := make([]bool, 6), make([]bool, 6)
					for i := range inputs {
						inputs[i] = mask&(1<<uint(i)) != 0
					}
					copy(original, inputs[:3])
					for i, old := range order {
						original[3+old] = inputs[3+i]
					}
					values, err := src.Evaluate(original)
					if err != nil {
						t.Fatal(err)
					}
					want := slices.Clone(values[:3])
					for _, old := range order {
						want = append(want, values[3+old])
					}
					checkQuantificationOutputs(t, out, inputs, want)
					got, err := out.Evaluate(inputs)
					if err != nil {
						t.Fatal(err)
					}
					packed := 0
					for i, value := range got {
						if value {
							packed |= 1 << uint(i)
						}
					}
					fmt.Fprintf(&records, " %d", packed)
				}
				records.WriteByte('\n')
			}
		}
	}
	return records.String()
}

func TestRepeatedDuplicationAndFlopClassesTypedTruth(t *testing.T) {
	timesClassReviewRecords(t, false)
}

func TestRepeatedDuplicationOfDfsGraphs(t *testing.T) {
	timesClassReviewRecords(t, true)
}

func TestRepeatedDuplicationEmptyInterfaces(t *testing.T) {
	constants := Gia_ManStart(1)
	Gia_ManAppendCo(constants, 1)
	Gia_ManAppendCo(constants, 0)
	autonomous := Gia_ManStart(1)
	r := Gia_ManAppendCi(autonomous)
	Gia_ManAppendCo(autonomous, r^1)
	Gia_ManSetRegNum(autonomous, 1)
	noOutputs := Gia_ManStart(1)
	a, b := Gia_ManAppendCi(noOutputs), Gia_ManAppendCi(noOutputs)
	Gia_ManAppendAnd(noOutputs, a, b) // Source copying preserves dangling nodes.
	for _, src := range []*Gia_Man_t{Gia_ManStart(1), constants, autonomous, noOutputs} {
		for times := int32(1); times <= 3; times++ {
			out := Gia_ManDupTimes(src, times)
			var records strings.Builder
			recordRepeatedReview(t, &records, out, src, times)
		}
	}
}

func TestRepeatedDuplicationAndClassInvalidArguments(t *testing.T) {
	for _, times := range []int32{0, -1} {
		t.Run(fmt.Sprintf("times%d", times), func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("invalid duplication count accepted")
				}
			}()
			Gia_ManDupTimes(appendReviewGraph(0, 1), times)
		})
	}
	for _, classes := range []*Vec_Int_t{nil, permutation(1), permutation(1, 2, 3, 4)} {
		t.Run(fmt.Sprintf("classes%v", classes), func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("invalid flop-class vector accepted")
				}
			}()
			src := blockReviewGraph(0, 3)
			src.VFlopClasses = classes
			Gia_ManDupFlopClass(src, 1)
		})
	}
}

func TestRepeatedDuplicationAndClassesMatchOriginalExpandedGraphs(t *testing.T) {
	exe := compileGiaSourceOracle(t, "dup_times_classes.c")
	want, err := testoracle.Command(t, exe).CombinedOutput()
	if err != nil {
		t.Fatalf("C repetition/classes: %v: %s", err, want)
	}
	gotRows := strings.Split(strings.TrimSpace(timesClassReviewRecords(t, false)), "\n")
	wantRows := strings.Split(strings.TrimSpace(string(want)), "\n")
	if len(gotRows) != len(wantRows) {
		t.Fatalf("Go cases %d, C cases %d", len(gotRows), len(wantRows))
	}
	for i := range gotRows {
		if gotRows[i] != wantRows[i] {
			t.Fatalf("case %d:\nGo %s\nC  %s", i, gotRows[i], wantRows[i])
		}
	}
	// The original routine also confuses physical and public PI order after
	// DFS. The unused first PI becomes a spurious true output in this case.
	witness, err := testoracle.Command(t, exe, "--dfs-order-witness").CombinedOutput()
	if err != nil || strings.TrimSpace(string(witness)) != "1" {
		t.Fatalf("original DFS-order defect witness: %v: %s", err, witness)
	}
}

func TestRepeatedDuplicationPreservesRemappedOutputOrder(t *testing.T) {
	src := Gia_ManStart(1)
	a, b := Gia_ManAppendCi(src), Gia_ManAppendCi(src)
	Gia_ManAppendCo(src, a)
	Gia_ManAppendCo(src, b^1)
	first, second := Gia_ManCo(src, 0), Gia_ManCo(src, 1)
	vec.Vec_IntWriteEntry(src.vCos, 0, second.id)
	vec.Vec_IntWriteEntry(src.vCos, 1, first.id)
	Gia_ObjSetCioId(second, 0)
	Gia_ObjSetCioId(first, 1)
	out := Gia_ManDupTimes(src, 2)
	for mask := 0; mask < 16; mask++ {
		inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0}
		checkQuantificationOutputs(t, out, inputs, []bool{!inputs[1], inputs[0], !inputs[3], inputs[2]})
	}
}

var repeatedDupReviewSink *Gia_Man_t

func BenchmarkRepeatedDuplicationTerminals(b *testing.B) {
	src := appendReviewGraph(3, 2)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		repeatedDupReviewSink = Gia_ManDupTimes(src, 32)
	}
}

func TestRepeatedDuplicationPreservesDfsInputOrder(t *testing.T) {
	src := Gia_ManStart(1)
	Gia_ManAppendCi(src) // Unused PI must remain first in the public interface.
	b := Gia_ManAppendCi(src)
	Gia_ManAppendCo(src, b)
	dfs := Gia_ManDupOrderDfs(src)
	if Gia_ManCi(dfs, 0).id < Gia_ManCi(dfs, 1).id {
		t.Fatal("fixture did not reorder physical CIs")
	}
	out := Gia_ManDupTimes(dfs, 2)
	for mask := 0; mask < 16; mask++ {
		inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0}
		checkQuantificationOutputs(t, out, inputs, []bool{inputs[1], inputs[3]})
	}
}

func TestRepeatedDuplicationAndFlopClassPreserveSimpleAnds(t *testing.T) {
	for _, operation := range []string{"times", "class", "permutation", "constraints"} {
		t.Run(operation, func(t *testing.T) {
			src := Gia_ManStart(1)
			src.fGiaSimple = 1
			a, r := Gia_ManAppendCi(src), Gia_ManAppendCi(src)
			Gia_ManAppendCo(src, Gia_ManAppendAnd(src, r, r^1))
			Gia_ManAppendCo(src, a)
			Gia_ManSetRegNum(src, 1)
			src.VFlopClasses = permutation(5)
			var out *Gia_Man_t
			switch operation {
			case "times":
				out = Gia_ManDupTimes(src, 2)
			case "class":
				out = Gia_ManDupFlopClass(src, 5)
			case "permutation":
				out = Gia_ManDupPerm(src, permutation(0))
			case "constraints":
				out = Gia_ManDupWithConstraints(src, permutation(0))
			}
			if out.fGiaSimple != 1 {
				t.Fatal("duplication lost simple mode")
			}
			for mask := 0; mask < 1<<uint(Gia_ManCiNum(out)); mask++ {
				inputs := make([]bool, Gia_ManCiNum(out))
				for i := range inputs {
					inputs[i] = mask&(1<<uint(i)) != 0
				}
				want := []bool{false, inputs[0]}
				if operation == "times" {
					want = []bool{false, false, inputs[0], inputs[1]}
				}
				checkQuantificationOutputs(t, out, inputs, want)
			}
		})
	}
}
