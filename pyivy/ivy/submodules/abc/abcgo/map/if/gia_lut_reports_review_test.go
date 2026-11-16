package if_

import (
	"fmt"
	"math/bits"
	"slices"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/bool/kit"

	"github.com/glycerine/abcgo/misc/vec"
)

func TestLutStructureFailureSourceDiagnostics(t *testing.T) {
	for _, tc := range []struct {
		name, structure, hex, variables string
		n, threshold                    int
		check75                         int32
	}{
		{"two", "33", "FEE8E880", "a,b,c,d,e", 5, 3, 0},
		{"three", "333", "FFFEFEE8FEE8E880FEE8E880E8808000", "a,b,c,d,e,f,g", 7, 4, 0},
		{"check75", "", "FFFFFFFEFFFEFEE8FFFEFEE8FEE8E880FFFEFEE8FEE8E880FEE8E880E8808000", "a,b,c,d,e,f,g,h", 8, 4, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := If_ManStart(&If_Par_t{NLutSize: int32(tc.n), NCutsMax: 4})
			defer If_ManStop(p)
			dst, leaves, temp, mapping, records, packing := structureReviewVectors(tc.n)
			defer Gia_ManStop(dst)
			for _, v := range []*Vec_Int_t{leaves, temp, mapping, records, packing} {
				defer vec.Vec_IntFree(v)
			}
			truth := make([]word, 1<<max(0, tc.n-6))
			for x := 0; x < 64*len(truth); x++ {
				if bits.OnesCount(uint(x&((1<<tc.n)-1))) >= tc.threshold {
					truth[x>>6] |= 1 << uint(x&63)
				}
			}
			before := slices.Clone(truth)
			output := dsdReportCapture(t, func() {
				if result := Gia_ManFromIfLogicNode(p, dst, 17, leaves, temp, truth, tc.structure, nil, mapping, records, packing, tc.check75, 0); result != -1 {
					t.Fatalf("result%d want failure", result)
				}
			})
			// Majority functions have no disjoint-support decomposition. The source
			// prints their prime truth in uppercase, with comma-separated variables.
			expected := fmt.Sprintf("0x%s    F = %s(%s)\nNode 17 is not decomposable. Deriving LUT structures has failed.\n", tc.hex, tc.hex, tc.variables)
			if output != expected {
				t.Fatalf("output%q want%q", output, expected)
			}
			if Gia_ManObjNum(dst) != int32(tc.n+1) || Vec_IntSize(mapping) != 0 || !slices.Equal(Vec_IntArray(records), []int32{0}) || !slices.Equal(Vec_IntArray(packing), []int32{0}) || !slices.Equal(truth, before) {
				t.Fatal("failed reconstruction changed graph, mapping, packing or truth")
			}
		})
	}
}

func TestLutDiagnosticDsdNodeFormatting(t *testing.T) {
	for _, tc := range []struct {
		name, want string
		kind       uint32
		fans       int32
		truth      uint32
	}{
		{"constant", "Const1", kit.KIT_DSD_CONST1, 0, 0},
		{"variable", "(a)", kit.KIT_DSD_VAR, 1, 0},
		{"and", "(a*a*a)", kit.KIT_DSD_AND, 3, 0},
		{"xor", "(a+a+a)", kit.KIT_DSD_XOR, 3, 0},
		{"prime", "CA(a,a,a)", kit.KIT_DSD_PRIME, 3, 0xca},
	} {
		for phase := 0; phase < 2; phase++ {
			t.Run(fmt.Sprintf("%s/phase%d", tc.name, phase), func(t *testing.T) {
				ntk := kit.Kit_DsdNtkAlloc(3)
				defer kit.Kit_DsdNtkFree(ntk)
				obj := kit.Kit_DsdObjAlloc(ntk, tc.kind, tc.fans)
				if tc.kind == kit.KIT_DSD_PRIME {
					kit.Kit_DsdObjTruth(obj)[0] = tc.truth
				}
				literal := int32(2*obj.Id) ^ int32(phase)
				var out strings.Builder
				stdout := dsdReportCapture(t, func() { giaWriteSourceDsd(&out, ntk, literal) })
				want := tc.want
				if phase != 0 {
					want = "!" + want
				}
				if out.String() != want || stdout != "" {
					t.Fatalf("writer%q stdout%q want%q", out.String(), stdout, want)
				}
				if obj.Type != tc.kind || kit.Kit_DsdObjFanNum(obj) != tc.fans {
					t.Fatal("formatting changed node metadata")
				}
			})
		}
	}
	ntk := kit.Kit_DsdNtkAlloc(3)
	defer kit.Kit_DsdNtkFree(ntk)
	for phase := int32(0); phase < 2; phase++ {
		var out strings.Builder
		giaWriteSourceDsd(&out, ntk, 4^phase)
		want := "c"
		if phase != 0 {
			want = "!c"
		}
		if out.String() != want {
			t.Fatalf("leaf%q want%q", out.String(), want)
		}
	}
}

func TestLutStructureValidationSourceDiagnostics(t *testing.T) {
	for _, tc := range []struct {
		structure, want string
		n               int
	}{
		{"3", "Wrong LUT struct (3)\n", 3},
		{"3333", "Wrong LUT struct (3333)\n", 3},
		{"23", "The LUT size (2) should belong to {3,4,5,6}.\n", 3},
		{"73", "The LUT size (7) should belong to {3,4,5,6}.\n", 3},
		{"3x", "The LUT size (72) should belong to {3,4,5,6}.\n", 3},
		{"33", "The node size (6) is too large for the LUT structure 33.\n", 6},
		{"333", "The node size (8) is too large for the LUT structure 333.\n", 8},
	} {
		t.Run(tc.structure, func(t *testing.T) {
			dst, leaves, temp, mapping, records, packing := structureReviewVectors(tc.n)
			defer Gia_ManStop(dst)
			for _, v := range []*Vec_Int_t{leaves, temp, mapping, records, packing} {
				defer vec.Vec_IntFree(v)
			}
			// A constant table must still be rejected by structure validation first.
			truth := make([]word, 1<<max(0, tc.n-6))
			for i := range truth {
				truth[i] = ^word(0)
			}
			p := &If_Man_t{pPars: &If_Par_t{}}
			got := dsdReportCapture(t, func() {
				if lit := Gia_ManFromIfLogicNode(p, dst, 17, leaves, temp, truth, tc.structure, nil, mapping, records, packing, 0, 0); lit != -1 {
					t.Fatalf("result%d want-1", lit)
				}
			})
			if got != tc.want {
				t.Fatalf("output%q want%q", got, tc.want)
			}
			if Gia_ManObjNum(dst) != int32(tc.n+1) || Vec_IntSize(mapping) != 0 || !slices.Equal(Vec_IntArray(records), []int32{0}) || !slices.Equal(Vec_IntArray(packing), []int32{0}) {
				t.Fatal("validation failure changed graph or records")
			}
		})
	}
}

func TestLutStructureSpecialFlagsPrecedeValidation(t *testing.T) {
	for _, tc := range []struct {
		name                     string
		n                        int
		check75, check44, filter int32
	}{
		{"check44", 3, 0, 1, 0}, {"check75", 4, 1, 0, 0}, {"filter", 6, 0, 0, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dst, leaves, temp, mapping, records, packing := structureReviewVectors(tc.n)
			defer Gia_ManStop(dst)
			for _, v := range []*Vec_Int_t{leaves, temp, mapping, records, packing} {
				defer vec.Vec_IntFree(v)
			}
			truth := []word{0}
			for x := 0; x < 64; x++ {
				if x&((1<<tc.n)-1) == (1<<tc.n)-1 {
					truth[0] |= 1 << x
				}
			}
			p := &If_Man_t{pPars: &If_Par_t{FLut6Filter: tc.filter}}
			result := int32(-1)
			output := dsdReportCapture(t, func() {
				result = Gia_ManFromIfLogicNode(p, dst, 17, leaves, temp, truth, "invalid", nil, mapping, records, packing, tc.check75, tc.check44)
			})
			if result <= 1 || output != "" {
				t.Fatalf("result%d output%q", result, output)
			}
			if tc.filter != 0 {
				if !slices.Equal(Vec_IntArray(packing), []int32{0}) {
					t.Fatalf("filter unexpectedly added packing%v", Vec_IntArray(packing))
				}
			} else if !slices.Equal(Vec_IntArray(packing), []int32{1, 1, result >> 1}) {
				t.Fatalf("single packing%v", Vec_IntArray(packing))
			}
			gia.Gia_ManAppendCo(dst, result)
			for x := 0; x < 1<<tc.n; x++ {
				input := make([]bool, tc.n)
				logical := 0
				for pin, lit := range Vec_IntArray(leaves) {
					input[pin] = x>>pin&1 != 0
					if (x>>pin&1)^int(lit&1) != 0 {
						logical |= 1 << pin
					}
				}
				got, err := dst.Evaluate(input)
				want := logical == (1<<tc.n)-1
				if err != nil || len(got) != 1 || got[0] != want {
					t.Fatalf("assignment%d got%v err%v want%t", x, got, err, want)
				}
			}
		})
	}
}
