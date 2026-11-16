package if_

import (
	"fmt"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/misc/vec"
)

// Build source-shaped records directly so report expectations do not depend on
// decomposition/canonicalization choosing a particular root or permutation.
func dsdVerboseNode(p *If_DsdMan_t, kind If_DsdType_t, fans ...int32) *If_DsdObj_t {
	o := If_DsdObjAlloc(p, int32(kind), int32(len(fans)))
	copy(o.pFans, fans)
	for _, fan := range fans {
		o.nSupp += If_DsdVecLitSuppSize(&p.vObjs, fan)
	}
	return o
}

func TestDsdVerbosePartitionReports(t *testing.T) {
	for _, tc := range []struct {
		name        string
		kind        If_DsdType_t
		fans        int
		lut, derive int32
		mask        uint32
		want        uint32
		suffix      string
	}{
		{"trivial", IF_DSD_AND, 2, 2, 1, 0, ^uint32(0), "    Trivial\n"},
		{"and", IF_DSD_AND, 3, 2, 1, 0, 5, "    1010000000000000\n    Using multi-input AND/XOR node\n"},
		{"xor", IF_DSD_XOR, 3, 2, 1, 0, 5, "    1010000000000000\n    Using multi-input AND/XOR node\n"},
		{"and-existence", IF_DSD_AND, 3, 2, 0, 0, ^uint32(0), "    1111111111111111\n    Using multi-input AND/XOR node\n"},
		{"and-masked-first", IF_DSD_AND, 3, 2, 1, 1, 20, "    0010100000000000\n    Using multi-input AND/XOR node\n"},
		{"and-impossible", IF_DSD_AND, 4, 2, 1, 0, 0, "    UNDEC\n"},
		{"and-all-masked", IF_DSD_AND, 3, 2, 1, ^uint32(0), 0, "    UNDEC\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := If_DsdManAlloc(4, 0)
			defer If_DsdManFree(p, 0)
			fans := make([]int32, tc.fans)
			for i := range fans {
				fans[i] = 2
			}
			o := dsdVerboseNode(p, tc.kind, fans...)
			open, close := "(", ")"
			if tc.kind == IF_DSD_XOR {
				open, close = "[", "]"
			}
			text := open + "abcd"[:tc.fans] + close
			prefix := fmt.Sprintf("     2 : %2d        0 0  %s", tc.fans, text)
			var result uint32
			output := dsdReportCapture(t, func() { result = If_DsdManCheckXY_int(p, o.Id<<1|1, tc.lut, tc.derive, tc.mask, 1) })
			if result != tc.want {
				t.Fatalf("partition %x, want %x", result, tc.want)
			}
			if output != prefix+tc.suffix {
				t.Fatalf("verbose %q, want %q", output, prefix+tc.suffix)
			}
			quiet := dsdReportCapture(t, func() { result = If_DsdManCheckXY_int(p, o.Id<<1, tc.lut, tc.derive, tc.mask, 0) })
			if quiet != "" || result != tc.want {
				t.Fatalf("verbosity changed decisions or quiet output: %q %x", quiet, result)
			}
		})
	}
}

func TestDsdVerboseSubtreeCandidates(t *testing.T) {
	for _, mask := range []uint32{0, 4} {
		t.Run(fmt.Sprint(mask), func(t *testing.T) {
			p := If_DsdManAlloc(3, 0)
			defer If_DsdManFree(p, 0)
			child := dsdVerboseNode(p, IF_DSD_XOR, 2, 2)
			root := dsdVerboseNode(p, IF_DSD_AND, 2, child.Id<<1)
			want := "     3 :  3        0 0  (a[bc])    Dec using node      2 :  2        0 0  [ab]\n"
			expected := uint32(20)
			if mask != 0 {
				expected = 0
				want += "    UNDEC\n"
			}
			var got uint32
			output := dsdReportCapture(t, func() { got = If_DsdManCheckXY_int(p, root.Id<<1, 2, 1, mask, 1) })
			if got != expected || output != want {
				t.Fatalf("partition%x report%q; want%x %q", got, output, expected, want)
			}
		})
	}
}

func TestDsdVerboseMuxCandidates(t *testing.T) {
	p := If_DsdManAlloc(5, 0)
	defer If_DsdManFree(p, 0)
	child := dsdVerboseNode(p, IF_DSD_AND, 2, 2)
	root := dsdVerboseNode(p, IF_DSD_MUX, 2, child.Id<<1, child.Id<<1)
	var got uint32
	output := dsdReportCapture(t, func() { got = If_DsdManCheckXY_int(p, root.Id<<1, 3, 1, 0, 1) })
	want := "     3 :  5        0 0  <a(bc)(de)>    UNDEC\n"
	if got != 0 || output != want {
		t.Fatalf("partition%x report%q, want0 %q", got, output, want)
	}
	// Exercise the MUX helper's successful shared-control partition directly.
	// The whole-tree traversal would first accept its 3-input child subtree.
	other := dsdVerboseNode(p, IF_DSD_AND, 2, 2, 2)
	mux := dsdVerboseNode(p, IF_DSD_MUX, 2, 2, other.Id<<1)
	for _, tc := range []struct {
		derive int32
		mask   uint32
		want   uint32
	}{{1, 0, 339}, {1, 1, 0}, {1, 16, 0}, {0, 0, ^uint32(0)}, {0, 1, ^uint32(0)}} {
		if got := If_DsdManCheckMux(p, 0, tc.mask, mux, 5, 4, tc.derive, 1); got != tc.want {
			t.Fatalf("MUX helper %x, want%x", got, tc.want)
		}
	}
}

func TestDsdVerbosePrimeSetFamily(t *testing.T) {
	p := If_DsdManAlloc(4, 0)
	defer If_DsdManFree(p, 0)
	// The direct helper consumes the cached family, independently of a caller's
	// traversal. Test weighted fanins and shared/unique/free set formatting.
	child := dsdVerboseNode(p, IF_DSD_AND, 2, 2)
	prime := dsdVerboseNode(p, IF_DSD_PRIME, 2, 2, child.Id<<1)
	sets := vec.Vec_IntAlloc(2)
	vec.Vec_IntPush(sets, 7)
	vec.Vec_IntPush(sets, 28)
	vec.Vec_PtrPush(p.vTtDecs[3], sets)
	vec.Vec_IntWriteEntry(&p.vTruths, int(prime.Id), 0)
	want := "\nThe 3-variable set family contains 2 sets:\n" +
		"S = 1  D = 2  C = 3   x=Ab y=xAc" + strings.Repeat(" ", 11) + "\n" +
		"S = 1  D = 2  C = 3   x=Bc y=xaB" + strings.Repeat(" ", 11) + "\n\n"
	for _, tc := range []struct {
		derive int32
		mask   uint32
		want   uint32
	}{{1, 0, 92}, {1, 4, 0}, {0, ^uint32(0), ^uint32(0)}} {
		var got uint32
		output := dsdReportCapture(t, func() { got = If_DsdManCheckPrime(p, 0, tc.mask, prime, 4, 3, tc.derive, 1) })
		if got != tc.want || output != want {
			t.Fatalf("partition%x report%q; want%x %q", got, output, tc.want, want)
		}
	}
}
