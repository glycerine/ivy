package if_

import (
	"fmt"
	"math/bits"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/misc/vec"
)

type structureReportReviewLut struct{ n, threshold int }

// Build actual AIG logic, recording complete cuts. Expected prime diagnostics
// come from scalar threshold truth, accounting for the returned root phase.
func structureReportReviewGraph(t testing.TB, specs []structureReportReviewLut) (*Gia_Man_t, []string) {
	t.Helper()
	p := Gia_ManStart(64)
	t.Cleanup(func() { Gia_ManStop(p) })
	nInputs := 7
	for _, spec := range specs {
		nInputs = max(nInputs, spec.n)
	}
	inputs := make([]int32, nInputs)
	for i := range inputs {
		inputs[i] = Gia_ManAppendCi(p)
	}
	offsets, records := Vec_IntAlloc(0), vec.Vec_IntAllocArrayCopy([]int32{0}, 1)
	defer Vec_IntFree(records)
	var primeLines []string
	for _, spec := range specs {
		truth := make([]word, 1<<max(0, spec.n-6))
		for x := 0; x < 64*len(truth); x++ {
			if bits.OnesCount(uint(x&((1<<spec.n)-1))) >= spec.threshold {
				truth[x>>6] |= 1 << uint(x&63)
			}
		}
		leaves := vec.Vec_IntAllocArrayCopy(inputs[:spec.n], spec.n)
		lit := Gia_ManFromIfLogicCreateLut(p, truth, leaves, nil, offsets, records)
		Vec_IntFree(leaves)
		Gia_ManAppendCo(p, lit)
		if spec.n >= 6 && spec.threshold != spec.n {
			var text strings.Builder
			text.WriteString("F = ")
			for i := len(truth) - 1; i >= 0; i-- {
				value := truth[i]
				if lit&1 != 0 {
					value = ^value
				}
				fmt.Fprintf(&text, "%016X", value)
			}
			text.WriteByte('(')
			for i := 0; i < spec.n; i++ {
				if i != 0 {
					text.WriteByte(',')
				}
				text.WriteByte(byte('a' + i))
			}
			text.WriteString(")\n")
			primeLines = append(primeLines, text.String())
		}
	}
	count := int(Gia_ManObjNum(p))
	vec.Vec_IntFillExtra(offsets, count, 0)
	for i, offset := range Vec_IntArray(offsets) {
		if offset != 0 {
			vec.Vec_IntWriteEntry(offsets, i, offset+int32(count))
		}
	}
	vec.Vec_IntAppend(offsets, records)
	if err := p.SetMappedData(offsets, nil, nil, ""); err != nil {
		t.Fatal(err)
	}
	// Check the graph against the scalar construction, independently of reporting.
	for x := 0; x < 1<<nInputs; x++ {
		values := make([]bool, nInputs)
		for i := range values {
			values[i] = x>>i&1 != 0
		}
		got, err := p.Evaluate(values)
		if err != nil {
			t.Fatal(err)
		}
		for i, spec := range specs {
			if got[i] != (bits.OnesCount(uint(x&((1<<spec.n)-1))) >= spec.threshold) {
				t.Fatalf("fixture LUT%d input%d has wrong truth", i, x)
			}
		}
	}
	return p, primeLines
}

func TestStructureReportSourceDiagnostics(t *testing.T) {
	for _, tc := range []struct {
		name    string
		specs   []structureReportReviewLut
		summary string
	}{
		{"empty", nil, "LUT5 = 0    LUT6 = 0  NonDec = 0 (0.00 %)    LUT7 = 0  NonDec = 0 (0.00 %)    \n"},
		{"small", []structureReportReviewLut{{2, 2}, {3, 3}, {4, 4}, {5, 5}}, "LUT5 = 4    LUT6 = 0  NonDec = 0 (0.00 %)    LUT7 = 0  NonDec = 0 (0.00 %)    \n"},
		{"mixed", []structureReportReviewLut{{5, 3}, {6, 6}, {6, 3}, {6, 3}, {7, 7}, {7, 7}, {7, 4}}, "LUT5 = 1    LUT6 = 3  NonDec = 2 (66.67 %)    LUT7 = 3  NonDec = 1 (33.33 %)    \n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p, primeLines := structureReportReviewGraph(t, tc.specs)
			before := slices.Clone(Vec_IntArray(p.VMapping))
			for pass := 0; pass < 2; pass++ {
				output := dsdReportCapture(t, func() { Gia_ManTestStruct(p) })
				want := strings.Join(primeLines, "") + tc.summary
				if output != want {
					t.Errorf("pass%d: output %q, want %q", pass, output, want)
				}
				if !slices.Equal(Vec_IntArray(p.VMapping), before) {
					t.Fatal("report changed mapping records")
				}
			}
		})
	}
}

func BenchmarkStructureReportSmallLuts(b *testing.B) {
	specs := make([]structureReportReviewLut, 200)
	for i := range specs {
		specs[i] = structureReportReviewLut{5, 5}
	}
	p, _ := structureReportReviewGraph(b, specs)
	quiet, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err != nil {
		b.Fatal(err)
	}
	old := os.Stdout
	os.Stdout = quiet
	b.Cleanup(func() { os.Stdout = old; quiet.Close() })
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Gia_ManTestStruct(p)
	}
}

func TestStructureReportCapsEachWidthAtTen(t *testing.T) {
	var specs []structureReportReviewLut
	for i := 0; i < 11; i++ {
		specs = append(specs, structureReportReviewLut{6, 3})
	}
	for i := 0; i < 12; i++ {
		specs = append(specs, structureReportReviewLut{7, 4})
	}
	p, primeLines := structureReportReviewGraph(t, specs)
	want := strings.Join(primeLines[:10], "") + strings.Join(primeLines[11:21], "") +
		"LUT5 = 0    LUT6 = 11  NonDec = 11 (100.00 %)    LUT7 = 12  NonDec = 12 (100.00 %)    \n"
	for pass := 0; pass < 2; pass++ {
		if output := dsdReportCapture(t, func() { Gia_ManTestStruct(p) }); output != want {
			t.Fatalf("pass%d: output %q, want %q", pass, output, want)
		}
	}
}

func TestStructureReportRejectsEightInputCut(t *testing.T) {
	p, _ := structureReportReviewGraph(t, []structureReportReviewLut{{8, 8}})
	before := slices.Clone(Vec_IntArray(p.VMapping))
	var caught any
	output := dsdReportCapture(t, func() {
		defer func() { caught = recover() }()
		Gia_ManTestStruct(p)
	})
	if caught != "if: invalid LUT in structure report" || output != "" {
		t.Fatalf("size rejection = %v, output %q", caught, output)
	}
	if !slices.Equal(Vec_IntArray(p.VMapping), before) {
		t.Fatal("size rejection changed mapping")
	}
}
