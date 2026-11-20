package kit

import (
	"math/rand"
	"testing"
)

func kitCoverValue(cover []uint32, x int) bool {
	for _, c := range cover {
		yes := true
		for v := 0; v < 16; v++ {
			lit := c >> uint(2*v) & 3
			if lit == 1 && x&(1<<uint(v)) != 0 || lit == 2 && x&(1<<uint(v)) == 0 || lit == 3 {
				yes = false
				break
			}
		}
		if yes {
			return true
		}
	}
	return false
}
func TestIsopFactoringAllThreeVariableFunctions(t *testing.T) {
	for f := uint32(0); f < 256; f++ {
		truth := f * 0x01010101
		for phase := int32(0); phase < 2; phase++ {
			var cover, memory []uint32
			compl := Kit_TruthIsop([]uint32{truth}, 3, &cover, phase)
			if compl < 0 {
				t.Fatal("ISOP failed", f)
			}
			g := Kit_SopFactor(cover, compl, 3, &memory)
			got := Kit_GraphToTruth(g)
			if got != truth {
				t.Fatalf("f=%x phase=%d cover=%x graph=%x", truth, compl, cover, got)
			}
		}
	}
}
func TestIsopMultiwordAndDontCare(t *testing.T) {
	rng := rand.New(rand.NewSource(13))
	for nVars := int32(5); nVars <= 9; nVars++ {
		for trial := 0; trial < 10; trial++ {
			truth := make([]uint32, Kit_TruthWordNum(nVars))
			for i := range truth {
				truth[i] = rng.Uint32()
			}
			var cover []uint32
			phase := Kit_TruthIsop(truth, nVars, &cover, 1)
			if phase < 0 {
				t.Fatal("ISOP failed")
			}
			for x := 0; x < 1<<uint(nVars); x++ {
				want := truth[x>>5]&(1<<uint(x&31)) != 0
				got := kitCoverValue(cover, x)
				if phase != 0 {
					got = !got
				}
				if got != want {
					t.Fatal(nVars, trial, x)
				}
			}
			off, on := make([]uint32, len(truth)), make([]uint32, len(truth))
			for i := range truth {
				care := rng.Uint32()
				on[i] = truth[i] & care
				off[i] = ^truth[i] & care
			}
			phase = Kit_TruthIsop2(off, on, nVars, &cover, 1, 0)
			if phase < 0 {
				t.Fatal("ISOP2 failed")
			}
			for x := 0; x < 1<<uint(nVars); x++ {
				got := kitCoverValue(cover, x)
				if phase != 0 {
					got = !got
				}
				bit := uint32(1) << uint(x&31)
				if on[x>>5]&bit != 0 && !got || off[x>>5]&bit != 0 && got {
					t.Fatal("care violation", nVars, trial, x)
				}
			}
		}
	}
}
func TestAlgebraicDivisionPreservesCover(t *testing.T) {
	// (a+b)(c+d)+e, using one positive literal bit per variable.
	sop := Kit_Sop_t{pCubes: []uint32{1 | 4, 1 | 8, 2 | 4, 2 | 8, 16}}
	div := Kit_Sop_t{pCubes: []uint32{1, 2}}
	var q, r Kit_Sop_t
	Kit_SopDivideInternal(&sop, &div, &q, &r, nil)
	if q.cubeNum() != 2 || r.cubeNum() != 1 || r.pCubes[0] != 16 {
		t.Fatal(q, r)
	}
	if sop.pCubes[0] != 5 {
		t.Fatal("source cover mutated")
	}
}

func TestScalarISOPHonorsNativeStoreBoundary(t *testing.T) {
	storage := make([]uint32, KIT_ISOP_MEM_LIMIT)
	for _, tc := range []struct {
		name          string
		on, dc, truth uint32
		free, cubes   int
	}{
		{"zero needs no words", 0, 0, 0, 0, 0},
		{"constant cube needs one word", 1, ^uint32(0), ^uint32(0), 1, 1},
		{"one literal needs two words", 0xaaaaaaaa, 0xaaaaaaaa, 0xaaaaaaaa, 2, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			memory := storage[:len(storage)-tc.free]
			var sop Kit_Sop_t
			got := Kit_TruthIsop5_rec(tc.on, tc.dc, 5, &sop, &memory)
			if got != tc.truth || sop.cubeNum() != int32(tc.cubes) || len(memory) != len(storage) {
				t.Fatalf("truth %08x cubes %d store %d, want %08x and %d cubes at full capacity", got, sop.cubeNum(), len(memory), tc.truth, tc.cubes)
			}
			if tc.free > 0 {
				memory = storage[:len(storage)-tc.free+1]
				Kit_TruthIsop5_rec(tc.on, tc.dc, 5, &sop, &memory)
				if !sop.failed {
					t.Error("one-word shortage accepted")
				}
			}
		})
	}
	// Four output words followed by two cube words for one essential variable.
	memory := storage[:len(storage)-6]
	var sop Kit_Sop_t
	words := []uint32{0xaaaaaaaa, 0xaaaaaaaa, 0xaaaaaaaa, 0xaaaaaaaa}
	got := Kit_TruthIsop_rec(words, words, 7, &sop, &memory)
	if len(got) != 4 || sop.cubeNum() != 1 || sop.pCubes[0] != 2 {
		t.Fatalf("lifted scalar cover %v, truth %x", sop, got)
	}
}
func fiveVariableCoverDigest() uint64 {
	var cover []uint32
	h := uint64(1)
	mix := func(v uint32) { h = (h ^ uint64(v)) * 1099511628211 }
	state := uint32(0x12345678)
	for i := 0; i < 256; i++ {
		state = 1664525*state + 1013904223
		phase := Kit_TruthIsop([]uint32{state}, 5, &cover, int32(i&1))
		mix(uint32(phase))
		mix(uint32(len(cover)))
		for _, cube := range cover {
			mix(cube)
		}
	}
	return h
}
func TestFiveVariableISOPCubesMatchNative(t *testing.T) {
	if h := fiveVariableCoverDigest(); h != 0xaa1cc4a627bf6337 {
		t.Fatalf("ordered cube digest %016x", h)
	}
}
func BenchmarkFiveVariableISOP(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		if fiveVariableCoverDigest() != 0xaa1cc4a627bf6337 {
			b.Fatal("wrong ordered cube digest")
		}
	}
}

func TestISOPKeepsCallerResultsAcrossStoreReuse(t *testing.T) {
	for _, returnTruth := range []int32{0, 1} {
		t.Run(map[int32]string{0: "cover", 1: "truth"}[returnTruth], func(t *testing.T) {
			t.Parallel()
			off, on := []uint32{0x0030ca0e, 0x46305a11}, []uint32{0x490a1400, 0x000b0080}
			var result []uint32
			phase := Kit_TruthIsop2(off, on, 6, &result, 1, returnTruth)
			if phase < 0 {
				t.Fatal("failed to derive incompletely specified cover")
			}
			saved := append([]uint32(nil), result...)
			for i := range 20 {
				truth := []uint32{uint32(i+1) * 0xabcdef01, 0x13579bdf}
				var scratch []uint32
				Kit_TruthIsop(truth, 6, &scratch, int32(i&1))
			}
			for i, v := range result {
				if v != saved[i] {
					t.Fatal("subsequent call changed earlier result")
				}
			}
			for row := range 64 {
				var value bool
				if returnTruth != 0 {
					value = result[row>>5]>>uint(row&31)&1 != 0
				} else {
					value = kitCoverValue(result, row)
				}
				if phase != 0 {
					value = !value
				}
				if on[row>>5]>>uint(row&31)&1 != 0 && !value || off[row>>5]>>uint(row&31)&1 != 0 && value {
					t.Fatalf("care row %d", row)
				}
			}
		})
	}
}
