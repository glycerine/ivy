package util

import (
	"math/rand"
	"testing"
)

func isopCubeHolds(c int32, n int32, assignment int) bool {
	for v := int32(0); v < n; v++ {
		code := c >> uint(2*v) & 3
		if code == 1 && assignment&(1<<uint(v)) != 0 {
			return false
		}
		if code == 2 && assignment&(1<<uint(v)) == 0 {
			return false
		}
	}
	return true
}
func isopBuildCover(cover []int32, n int32) []word {
	nWords := 1
	if n > 6 {
		nWords = 1 << uint(n-6)
	}
	out := make([]word, nWords)
	for assignment := 0; assignment < nWords*64; assignment++ {
		for _, c := range cover {
			if isopCubeHolds(c, n, assignment) {
				out[assignment>>6] |= word(1) << uint(assignment&63)
				break
			}
		}
	}
	return out
}
func isopStretch(t word, n int32) word {
	var out word
	mask := (1 << uint(n)) - 1
	for a := 0; a < 64; a++ {
		if t&(word(1)<<uint(a&mask)) != 0 {
			out |= word(1) << uint(a)
		}
	}
	return out
}
func TestIsopAllThreeVariableFunctions(t *testing.T) {
	for f := 0; f < 256; f++ {
		truth := isopStretch(word(f), 3)
		cover := make([]int32, 64)
		count := int32(0)
		result := Abc_Tt6IsopCover(truth, truth, 3, cover, &count)
		if result != truth || isopBuildCover(cover[:count], 3)[0] != truth {
			t.Fatal("cover reconstruction", f)
		}
		size := int32(0)
		if Abc_Tt6Isop(truth, truth, 3, &size) != truth || size != count {
			t.Fatal("count mismatch", f)
		}
		for i := int32(0); i < count; i++ {
			private := false
			for a := 0; a < 8; a++ {
				if truth&(1<<uint(a)) == 0 || !isopCubeHolds(cover[i], 3, a) {
					continue
				}
				coveredElsewhere := false
				for j := int32(0); j < count; j++ {
					if i != j && isopCubeHolds(cover[j], 3, a) {
						coveredElsewhere = true
					}
				}
				if !coveredElsewhere {
					private = true
					break
				}
			}
			if !private {
				t.Fatalf("redundant cube function %d", f)
			}
		}
	}
}
func TestIsopDontCaresAndWideTables(t *testing.T) {
	r := rand.New(rand.NewSource(191))
	for n := int32(0); n <= 8; n++ {
		words := 1
		if n > 6 {
			words = 1 << uint(n-6)
		}
		for trial := 0; trial < 40; trial++ {
			on, dc := make([]word, max(words, 4)), make([]word, max(words, 4))
			for i := 0; i < words; i++ {
				dc[i] = r.Uint64()
				on[i] = r.Uint64() & dc[i]
			}
			if n < 6 {
				on[0] = isopStretch(on[0], n)
				dc[0] = isopStretch(dc[0], n)
			}
			for i := words; i < 4; i++ {
				on[i] = on[i%words]
				dc[i] = dc[i%words]
			}
			out := make([]word, 4)
			cover := make([]int32, 256)
			count := int32(0)
			Abc_Tt8IsopCover(on, dc, n, out, cover, &count)
			built := isopBuildCover(cover[:count], n)
			for i := 0; i < words; i++ {
				if on[i]&^out[i] != 0 || out[i]&^dc[i] != 0 || out[i] != built[i] {
					t.Fatalf("bounds/cover n=%d trial=%d", n, trial)
				}
			}
			if size := Abc_Tt8Isop(on, dc, n, make([]word, 4)); size != count {
				t.Fatal("wide count mismatch")
			}
			if n <= 7 {
				out7 := make([]word, 2)
				count7 := Abc_Tt7Isop(on, dc, n, out7)
				if count7 != count || out7[0] != out[0] || out7[1] != out[1] {
					t.Fatal("7/8 mismatch", n)
				}
			}
		}
	}
}
func TestIsopCoverAppendAndConstants(t *testing.T) {
	cover := make([]int32, 64)
	cover[0] = 12345
	count := int32(1)
	Abc_Tt6IsopCover(s_Truths6[2], s_Truths6[2], 6, cover, &count)
	if count != 2 || cover[0] != 12345 || cover[1] != 2<<4 {
		t.Fatal(cover[:count], count)
	}
	for _, truth := range []word{0, ^word(0)} {
		count = 0
		out := Abc_Tt6IsopCover(truth, truth, 0, cover, &count)
		if out != truth || int(count) != int(truth&1) {
			t.Fatal("constant", truth, count)
		}
	}
}
func TestEsopReconstruction(t *testing.T) {
	r := rand.New(rand.NewSource(192))
	for n := int32(0); n <= 6; n++ {
		trials := 100
		if n == 3 {
			trials = 256
		}
		for trial := 0; trial < trials; trial++ {
			truth := isopStretch(r.Uint64(), n)
			if n == 3 {
				truth = isopStretch(word(trial), n)
			}
			cover := make([]int32, 64)
			count := Abc_Tt6Esop(truth, n, cover)
			if Abc_Tt6EsopBuild(n, cover, count) != truth {
				t.Fatalf("ESOP n=%d trial=%d", n, trial)
			}
			if Abc_Tt6Esop(truth, n, nil) != count || Abc_Tt6EsopVerify(truth, n) != count {
				t.Fatal("ESOP count mismatch")
			}
		}
	}
}
func TestCnfEncodedClauses(t *testing.T) {
	for f := 0; f < 256; f++ {
		truth := isopStretch(word(f), 3)
		cover := make([]int32, 64)
		count := Abc_Tt6Cnf(truth, 3, cover)
		if count != Abc_Tt6CnfSize(truth, 3) {
			t.Fatal("CNF count")
		}
		for assignment := 0; assignment < 16; assignment++ {
			forbidden := false
			for _, cube := range cover[:count] {
				if isopCubeHolds(cube, 4, assignment) {
					forbidden = true
					break
				}
			}
			expected := (assignment >> 3 & 1) != int(truth>>uint(assignment&7)&1)
			if forbidden != expected {
				t.Fatal("CNF relation", f, assignment)
			}
		}
	}
}
func TestWideCnf(t *testing.T) {
	r := rand.New(rand.NewSource(193))
	for n := int32(7); n <= 8; n++ {
		truth := make([]word, 4)
		for i := range truth {
			truth[i] = r.Uint64()
		}
		if n == 7 {
			truth[2], truth[3] = truth[0], truth[1]
		}
		cover := make([]int32, 256)
		count := Abc_Tt8Cnf(truth, n, cover)
		if count != Abc_Tt8CnfSize(truth, n) {
			t.Fatal("wide CNF size")
		}
		for a := 0; a < 1<<uint(n+1); a++ {
			forbidden := false
			for _, cube := range cover[:count] {
				if isopCubeHolds(cube, n+1, a) {
					forbidden = true
					break
				}
			}
			input := a & ((1 << uint(n)) - 1)
			expected := (a >> uint(n) & 1) != int(truth[input>>6]>>uint(input&63)&1)
			if forbidden != expected {
				t.Fatal("wide CNF", n, a)
			}
		}
	}
}
func TestOutputDecomposition(t *testing.T) {
	i := int32(2)
	x, g := s_Truths6[i], s_Truths6[0]^s_Truths6[4]
	tables := []word{x & g, ^x & g, ^x | g, x | g, x ^ g}
	for kind, truth := range tables {
		out := []word{0}
		if got := Abc_Tt6CheckOutDec(truth, i, out); got != int32(kind) || out[0] != g {
			t.Fatal("output decomposition", kind, got, out)
		}
		wide := []word{truth, truth, truth, truth}
		outWide := make([]word, 4)
		if got := Abc_TtCheckOutDec(wide, 8, i, outWide); got != int32(kind) {
			t.Fatal("wide output kind", kind, got)
		}
		for _, v := range outWide {
			if v != g {
				t.Fatal("wide output truth")
			}
		}
	}
}
func TestDsdAndAndMux(t *testing.T) {
	i, j, k := int32(1), int32(3), int32(5)
	x, y, z := s_Truths6[i], s_Truths6[j], s_Truths6[k]
	// The source comments label cases 1 and 2 in the opposite order from
	// the actual cofactor comparisons. The original C oracle returns
	// 0,2,1,3,4 for x&y, x&!y, !x&y, !x&!y, x^y respectively.
	pair := []word{x & y, ^x & y, x &^ y, ^x &^ y, x ^ y}
	for expected, p := range pair {
		out := []word{0}
		kind := Abc_TtCheckDsdAnd(p^z, i, j, out)
		if kind != int32(expected) {
			t.Fatal("DSD AND kind", expected, kind)
		}
		for a := 0; a < 64; a++ {
			mapped := a &^ (1 << uint(i))
			if p&(word(1)<<uint(a)) != 0 {
				mapped |= 1 << uint(i)
			}
			if (out[0] >> uint(mapped) & 1) != ((p ^ z) >> uint(a) & 1) {
				t.Fatal("DSD reconstruction", expected, a)
			}
		}
	}
	mux := (x & z) | (^x & y)
	out := []word{0}
	code := Abc_TtCheckDsdMux(mux, i, out)
	if code != (2*k)<<16|2*j || out[0] != x {
		t.Fatal("DSD mux", code, out)
	}
}

func TestOutputDecompositionSevenAndEight(t *testing.T) {
	parity := s_Truths6[0] ^ s_Truths6[1] ^ s_Truths6[2] ^ s_Truths6[3] ^ s_Truths6[4] ^ s_Truths6[5]
	seven := []word{0, parity}
	variable, kind := int32(-1), int32(-1)
	if result := Abc_TtCheckDecOutOne7(seven, &variable, &kind); result != parity || variable != 6 || kind != 0 {
		t.Fatal("seven", result, variable, kind)
	}
	eight := []word{0, 0, 0, parity}
	v1, v2, k1, k2 := int32(-1), int32(-1), int32(-1), int32(-1)
	if result := Abc_TtCheckDecOutOne8(eight, &v1, &v2, &k1, &k2); result != parity || v1 != 7 || v2 != 6 || k1 != 0 || k2 != 0 {
		t.Fatal("eight", result, v1, v2, k1, k2)
	}
	// Put the decomposable variable below the most significant position. The
	// helper must permute it to position six before returning the residual.
	original := make([]word, 2)
	for a := 0; a < 128; a++ {
		g := 0
		for v := 0; v < 7; v++ {
			if v != 3 {
				g ^= a >> uint(v) & 1
			}
		}
		if a&(1<<3) != 0 && g != 0 {
			original[a>>6] |= 1 << uint(a&63)
		}
	}
	table := append([]word(nil), original...)
	result := Abc_TtCheckDecOutOne7(table, &variable, &kind)
	if variable != 3 || kind != 0 {
		t.Fatal("lower variable", variable, kind)
	}
	for a := 0; a < 128; a++ {
		mapped := a
		if (a>>3^a>>6)&1 != 0 {
			mapped ^= (1 << 3) | (1 << 6)
		}
		if table[a>>6]>>uint(a&63)&1 != original[mapped>>6]>>uint(mapped&63)&1 {
			t.Fatal("swap mismatch")
		}
		want := word(0)
		if a&(1<<6) != 0 {
			want = result >> uint(a&63) & 1
		}
		if table[a>>6]>>uint(a&63)&1 != want {
			t.Fatal("residual mismatch")
		}
	}
}

func TestIsopWideInputOutputAlias(t *testing.T) {
	on := []word{0x123456789abcdef0, 0x02468ace13579bdf, 0x1122334455667788, 0x8877665544332211}
	dc := []word{^word(0), ^word(0), ^word(0), ^word(0)}
	separate := make([]word, 4)
	wantCount := Abc_Tt8Isop(on, dc, 8, separate)
	alias := append([]word(nil), on...)
	gotCount := Abc_Tt8Isop(alias, dc, 8, alias)
	if gotCount != wantCount {
		t.Fatal("alias count")
	}
	for i := range alias {
		if alias[i] != separate[i] {
			t.Fatal("alias result")
		}
	}
}
