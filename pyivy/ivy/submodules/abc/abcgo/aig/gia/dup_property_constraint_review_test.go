package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/internal/testoracle"
	"strings"
	"testing"
)

func propertyConstraintReviewRecords(t *testing.T) string {
	t.Helper()
	var records strings.Builder
	check := func(out, source, property *Gia_Man_t, types []int32, mode int, signs, parameter int32) {
		t.Helper()
		wantPOs, wantConstraints := int32(2), int32(0)
		if mode == 1 {
			wantPOs = Gia_ManPoNum(source)
			for _, kind := range types {
				wantConstraints += kind
			}
		}
		if Gia_ManCiNum(out) != 6 || Gia_ManPoNum(out) != wantPOs || out.nRegs != source.nRegs || out.nConstrs != wantConstraints || Gia_ManIsNormalized(out) == 0 {
			t.Fatalf("mode %d signs %d parameter %d: wrong output interface or constraints", mode, signs, parameter)
		}
		if mode == 1 && out.nObjs != source.nObjs {
			t.Fatal("constraint conversion changed source object count")
		}
		if name, spec := out.Identity(); name != "block" || spec != "block.aig" {
			t.Fatal("lost source identity")
		}
		fmt.Fprintf(&records, "%d %d %d %d %d %d %d %d", mode, signs, parameter, source.nRegs, Gia_ManCiNum(out), Gia_ManCoNum(out), out.nRegs, out.nConstrs)
		for mask := 0; mask < 64; mask++ {
			inputs := make([]bool, 6)
			for i := range inputs {
				inputs[i] = mask&(1<<uint(i)) != 0
			}
			values, err := source.Evaluate(inputs)
			if err != nil {
				t.Fatal(err)
			}
			var want []bool
			if mode == 0 {
				// Property inputs are current register outputs, not the next
				// state computed by the source's register-input cones.
				want, err = property.Evaluate(inputs[Gia_ManPiNum(source):])
				if err != nil {
					t.Fatal(err)
				}
			} else {
				for phase := int32(0); phase <= 1; phase++ {
					for i, kind := range types {
						if kind == phase {
							want = append(want, values[i] != (phase == 1))
						}
					}
				}
			}
			want = append(want, values[Gia_ManPoNum(source):]...)
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
	for signs := int32(0); signs < 8; signs++ {
		for propertySigns := int32(0); propertySigns < 8; propertySigns++ {
			source := blockReviewGraph(signs, 3)
			source.nConstrs = 1 // Replacing old properties also drops old constraint flags.
			property := cascadeReviewSecond(propertySigns)
			out := Gia_ManDupWithNewPo(source, property)
			check(out, source, property, nil, 0, signs, propertySigns)
		}
	}
	for signs := int32(0); signs < 8; signs++ {
		for _, regs := range []int32{0, 3, 6} {
			for flags := int32(0); flags < 1<<uint(6-regs); flags++ {
				source := blockReviewGraph(signs, regs)
				if Gia_ManPoNum(source) != 0 {
					source.nConstrs = 1
				}
				types := make([]int32, 6-regs)
				for i := range types {
					types[i] = (flags >> uint(i)) & 1
				}
				out := Gia_ManDupWithConstraints(source, permutation(types...))
				check(out, source, nil, types, 1, signs, flags)
			}
		}
	}
	return records.String()
}

func TestPropertyAndConstraintTypedTransitionTables(t *testing.T) {
	propertyConstraintReviewRecords(t)
}

func TestPropertyAndConstraintsMatchOriginalExpandedGraphs(t *testing.T) {
	exe := compileGiaSourceOracle(t, "property_constraints.c")
	want, err := testoracle.Command(t, exe).CombinedOutput()
	if err != nil {
		t.Fatalf("C property/constraints: %v: %s", err, want)
	}
	gotRows := strings.Split(strings.TrimSpace(propertyConstraintReviewRecords(t)), "\n")
	wantRows := strings.Split(strings.TrimSpace(string(want)), "\n")
	if len(gotRows) != len(wantRows) {
		t.Fatalf("Go cases %d, C cases %d", len(gotRows), len(wantRows))
	}
	for i := range gotRows {
		if gotRows[i] != wantRows[i] {
			t.Fatalf("case %d:\nGo %s\nC  %s", i, gotRows[i], wantRows[i])
		}
	}
}

func TestPropertyReplacementAndConstraintsEmptyGraphs(t *testing.T) {
	source := Gia_ManStart(1)
	property := Gia_ManStart(1)
	Gia_ManAppendCo(property, 1)
	Gia_ManAppendCo(property, 0)
	out := Gia_ManDupWithNewPo(source, property)
	checkQuantificationOutputs(t, out, nil, []bool{true, false})
	if out.nObjs != 3 || out.nRegs != 0 || out.nConstrs != 0 {
		t.Fatal("constant property replacement added unnecessary objects")
	}
	out = Gia_ManDupWithConstraints(source, permutation())
	if out.nObjs != 1 || Gia_ManCiNum(out) != 0 || Gia_ManCoNum(out) != 0 || out.nRegs != 0 || out.nConstrs != 0 {
		t.Fatal("empty constraint conversion changed interface")
	}
	source = blockReviewGraph(5, 3)
	property = Gia_ManStart(1)
	for i := 0; i < 3; i++ {
		Gia_ManAppendCi(property)
	}
	out = Gia_ManDupWithNewPo(source, property)
	if Gia_ManPoNum(out) != 0 || out.nRegs != 3 || out.nConstrs != 0 {
		t.Fatal("empty property changed transition interface")
	}
	for mask := 0; mask < 64; mask++ {
		inputs := make([]bool, 6)
		for i := range inputs {
			inputs[i] = mask&(1<<uint(i)) != 0
		}
		want, err := source.Evaluate(inputs)
		if err != nil {
			t.Fatal(err)
		}
		checkQuantificationOutputs(t, out, inputs, want[3:])
	}
}

func TestPropertyAndConstraintInterfaceRejection(t *testing.T) {
	cases := []struct {
		name string
		run  func()
	}{
		{"sequential_property", func() {
			Gia_ManDupWithNewPo(blockReviewGraph(0, 3), appendReviewGraph(0, 1))
		}},
		{"property_input_count", func() {
			Gia_ManDupWithNewPo(appendReviewGraph(0, 1), cascadeReviewSecond(0))
		}},
		{"too_few_types", func() {
			Gia_ManDupWithConstraints(blockReviewGraph(0, 3), permutation(0, 1))
		}},
		{"too_many_types", func() {
			Gia_ManDupWithConstraints(blockReviewGraph(0, 3), permutation(0, 1, 0, 1))
		}},
		{"invalid_type", func() {
			Gia_ManDupWithConstraints(blockReviewGraph(0, 3), permutation(0, 3, 1))
		}},
		{"negative_type", func() {
			Gia_ManDupWithConstraints(blockReviewGraph(0, 3), permutation(0, -1, 1))
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("invalid property/constraint interface accepted")
				}
			}()
			tc.run()
		})
	}
}
