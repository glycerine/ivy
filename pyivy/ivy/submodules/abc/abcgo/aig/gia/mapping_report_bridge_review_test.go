package gia

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	abcutil "github.com/glycerine/abcgo/misc/util"
)

func mappingReportReviewWideLut(size int) *Gia_Man_t {
	p := Gia_ManStart(128)
	inputs := make([]int32, size)
	for i := range inputs {
		inputs[i] = Gia_ManAppendCi(p)
	}
	root := inputs[0]
	for _, input := range inputs[1:] {
		root = Gia_ManAppendAnd(p, root, input)
	}
	Gia_ManAppendCo(p, root)
	p.VMapping = Vec_IntStart(int(p.nObjs))
	Vec_IntWriteEntry(p.VMapping, int(root>>1), int32(Vec_IntSize(p.VMapping)))
	Vec_IntPush(p.VMapping, int32(size))
	for _, input := range inputs {
		Vec_IntPush(p.VMapping, input>>1)
	}
	Vec_IntPush(p.VMapping, root>>1)
	return p
}

func TestMappingReportsUseAbcPrintBridge(t *testing.T) {
	average, packed, empty, large := mappingReviewAverageGraph(), mappingReviewReportGraph(3), Gia_ManStart(1), mappingReportReviewWideLut(33)
	defer Gia_ManStop(average)
	defer Gia_ManStop(packed)
	defer Gia_ManStop(empty)
	defer Gia_ManStop(large)
	emptyPacked, wide := Gia_ManStart(1), mappingReportReviewWideLut(32)
	defer Gia_ManStop(emptyPacked)
	defer Gia_ManStop(wide)
	emptyPacked.VPacking = Vec_IntStart(1)
	var wideProfile strings.Builder
	wideProfile.WriteString("LUT = 1 : ")
	for size := 2; size < 32; size++ {
		fmt.Fprintf(&wideProfile, "%d=0 0.0 %%  ", size)
	}
	wideProfile.WriteString("32=1 100.0 %  Ave = 32.00\n")
	for _, tc := range []struct {
		name, want string
		run        func()
	}{
		{"mapping", "Mapping (K=2)  :  \033[1;33mlut =      2\033[0m  \033[1;32medge =       4\033[0m  \033[1;31mlev =    2\033[0m \033[1;31m(1.07)\033[0m  mem = 0.00 MB\n", func() { Gia_ManPrintMappingStats(average, "") }},
		{"packing", "Packing (N=3)  :  1 x LUT = 1   2 x LUT = 1   3 x LUT = 1   Total = 3   Total LUT = 6\n", func() { Gia_ManPrintPackingStats(packed) }},
		{"profile", "LUT = 3 : 2=1 33.3 %  3=2 66.7 %  Ave = 2.67\n", func() { Gia_ManPrintNodeProfile([]int32{0, 0, 1, 2}, 3) }},
		{"lut_stats", "LUT = 4 : 2=2 50.0 %  3=2 50.0 %  Ave = 2.50\n", func() { Gia_ManPrintLutStats(packed) }},
		{"empty_luts", "LUT = 0 : Ave = 0.00\n", func() { Gia_ManPrintLutStats(empty) }},
		{"large_luts", "The max LUT size (33) is too large.\n", func() { Gia_ManPrintLutStats(large) }},
		{"empty_packing", "Packing (N=-1)  :  Total = 0   Total LUT = 0\n", func() { Gia_ManPrintPackingStats(emptyPacked) }},
		{"profile_terminals", "LUT = 4 : 2=1 25.0 %  Ave = 0.75\n", func() { Gia_ManPrintNodeProfile([]int32{2, 1, 1}, 2) }},
		{"wide_luts", wideProfile.String(), func() { Gia_ManPrintLutStats(wide) }},
		{"absent_mapping", "", func() { Gia_ManPrintMappingStats(empty, "") }},
		{"absent_packing", "", func() { Gia_ManPrintPackingStats(empty) }},
	} {
		for _, bridge := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/bridge%t", tc.name, bridge), func(t *testing.T) {
				old := abcutil.PrintBridge
				defer func() { abcutil.PrintBridge = old }()
				var text strings.Builder
				abcutil.PrintBridge = nil
				if bridge {
					abcutil.PrintBridge = func(message string) { text.WriteString(message) }
				}
				output := hashReviewStdout(t, tc.run)
				wantStdout, wantBridge := tc.want, ""
				if bridge {
					wantStdout, wantBridge = "", tc.want
				}
				if output != wantStdout || text.String() != wantBridge {
					t.Fatalf("stdout%q bridge%q want %q/%q", output, text.String(), wantStdout, wantBridge)
				}
			})
		}
	}
}

func TestMuxDuplicateReportRetainsSourceStdout(t *testing.T) {
	p := Gia_ManStart(32)
	defer Gia_ManStop(p)
	c, a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	var roots []int32
	for i := 0; i < 3; i++ {
		root := Gia_ManAppendMux(p, c, a, b)
		roots = append(roots, root>>1)
		Gia_ManAppendCo(p, root)
	}
	p.VMapping = Vec_IntStart(int(p.nObjs))
	for _, id := range roots {
		Vec_IntWriteEntry(p.VMapping, int(id), int32(Vec_IntSize(p.VMapping)))
		for _, value := range []int32{3, c >> 1, a >> 1, b >> 1, -id} {
			Vec_IntPush(p.VMapping, value)
		}
	}
	old := abcutil.PrintBridge
	defer func() { abcutil.PrintBridge = old }()
	abcutil.PrintBridge = func(message string) { t.Errorf("source printf was redirected: %q", message) }
	for pass := 0; pass < 2; pass++ {
		for _, obj := range p.pObjs {
			obj.fMark0, obj.fMark1 = 1, 1
		}
		var count int32
		output := hashReviewStdout(t, func() { count = Gia_ManCountDupLut(p) })
		want := "Dup fanins = 4.  CI fanins = 6.  MUX fanins = 0.  Total = 10.  (333.33 %)\n"
		if output != want || count != 10 {
			t.Fatalf("pass%d: count%d output%q want%q", pass, count, output, want)
		}
		for _, obj := range p.pObjs {
			if obj.fMark0 != 0 || obj.fMark1 != 0 {
				t.Fatal("duplicate report retained marks")
			}
		}
	}
}

func TestMappingReportBoxedLevelsRouting(t *testing.T) {
	for seed := uint32(1); seed <= 32; seed++ {
		for _, bridge := range []bool{false, true} {
			t.Run(fmt.Sprintf("seed%d/bridge%t", seed, bridge), func(t *testing.T) {
				p, scalarLevels, _ := mappingReviewBoxGraph(seed, false)
				defer Gia_ManStop(p)
				var maximum int32
				for i := int32(0); i < Gia_ManCoNum(p); i++ {
					maximum = max(maximum, scalarLevels[Gia_ManCo(p, i).id])
				}
				old := abcutil.PrintBridge
				defer func() { abcutil.PrintBridge = old }()
				var text strings.Builder
				abcutil.PrintBridge = nil
				if bridge {
					abcutil.PrintBridge = func(message string) { text.WriteString(message) }
				}
				output := hashReviewStdout(t, func() { Gia_ManPrintMappingStats(p, "") })
				if bridge {
					if output != "" {
						t.Fatalf("boxed report bypassed bridge: %q", output)
					}
					output = text.String()
				} else if text.Len() != 0 {
					t.Fatal("boxed report unexpectedly used bridge")
				}
				want := fmt.Sprintf("\033[1;31mlevB =%5d\033[0m  ", maximum)
				if !strings.Contains(output, want) {
					t.Fatalf("boxed scalar level %q missing from %q", want, output)
				}
			})
		}
	}
}

func TestMappingReportMuxFaninsPreserveSourceRoles(t *testing.T) {
	for phase := 0; phase < 8; phase++ {
		p := mappingReviewReportGraph(phase)
		first := Gia_ObjFanin0(Gia_ManCo(p, 0))
		second := Gia_ObjFanin0(Gia_ManCo(p, 1))
		var got [3]int32
		Gia_ManPrintGetMuxFanins(p, first, got[:])
		want := []int32{1, 2, 3}
		if phase&1 != 0 {
			want[1], want[2] = want[2], want[1]
		}
		if !slices.Equal(got[:], want) {
			t.Errorf("phase%d first mux fanins%v want%v", phase, got, want)
		}
		Gia_ManPrintGetMuxFanins(p, second, got[:])
		if !slices.Equal(got[:], []int32{4, first.id, 3}) {
			t.Errorf("phase%d nested mux fanins%v", phase, got)
		}
		Gia_ManStop(p)
	}
}

func TestMuxDuplicateReportTwoInputSelection(t *testing.T) {
	p := Gia_ManStart(32)
	defer Gia_ManStop(p)
	a, b, c, d := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	u, v := Gia_ManAppendAnd(p, a, b), Gia_ManAppendAnd(p, c, d)
	first := Gia_ManAppendAnd(p, u, v)
	roots := []int32{u, v, first, Gia_ManAppendAnd(p, u, v), Gia_ManAppendAnd(p, first, a), Gia_ManAppendAnd(p, first, a), Gia_ManAppendAnd(p, first, a)}
	for _, root := range roots {
		Gia_ManAppendCo(p, root)
	}
	p.VMapping = Vec_IntStart(int(p.nObjs))
	for i, root := range roots {
		id := root >> 1
		obj := Gia_ManObj(p, id)
		marker := id
		if i >= 2 {
			marker = -id
		}
		Vec_IntWriteEntry(p.VMapping, int(id), int32(Vec_IntSize(p.VMapping)))
		for _, value := range []int32{2, Gia_ObjFaninId0(obj, id), Gia_ObjFaninId1(obj, id), marker} {
			Vec_IntPush(p.VMapping, value)
		}
	}
	// The first two marked cuts consume ordinary u then v. The next consumes
	// CI a; the last two consume the first marked root, counting it twice as
	// a mux fanin and once as a duplicate. Seven LUTs form the denominator.
	for pass := 0; pass < 2; pass++ {
		var count int32
		output := hashReviewStdout(t, func() { count = Gia_ManCountDupLut(p) })
		want := "Dup fanins = 1.  CI fanins = 1.  MUX fanins = 2.  Total = 2.  (57.14 %)\n"
		if output != want || count != 2 {
			t.Fatalf("pass%d count%d output%q want%q", pass, count, output, want)
		}
		for _, obj := range p.pObjs {
			if obj.fMark0 != 0 || obj.fMark1 != 0 {
				t.Fatal("two-input duplicate report retained marks")
			}
		}
	}
}
