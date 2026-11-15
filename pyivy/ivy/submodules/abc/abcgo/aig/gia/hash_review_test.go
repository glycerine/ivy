package gia

import (
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strings"
	"testing"
)

func hashReviewStdout(t *testing.T, action func()) string {
	t.Helper()
	file, err := os.CreateTemp(t.TempDir(), "hash-profile")
	if err != nil {
		t.Fatal(err)
	}
	original := os.Stdout
	os.Stdout = file
	defer func() { os.Stdout = original; file.Close() }()
	action()
	data, err := os.ReadFile(file.Name())
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestHashProfileIncludesBucketOccupancy(t *testing.T) {
	g := Gia_ManStart(8)
	a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	Gia_ManHashStart(g)
	Gia_ManHashAnd(g, a, b)
	output := hashReviewStdout(t, func() { Gia_ManHashProfile(g) })
	if !strings.HasSuffix(output, "\n1 \n") {
		t.Fatalf("missing source bucket occupancy line: %q", output)
	}
}

type hashReviewTruth struct {
	lit   int32
	truth uint16
}

func TestHashPrimitivesNativeOracle(t *testing.T) {
	exe := compileGiaSourceOracle(t, "hash_primitives.c")
	native, err := exec.Command(exe).CombinedOutput()
	if err != nil {
		t.Fatalf("native: %v\n%s", err, native)
	}
	var got strings.Builder
	for mode := 0; mode < 4; mode++ {
		g := Gia_ManStart(64)
		inputs := make([]int32, 4)
		for i := range inputs {
			inputs[i] = Gia_ManAppendCi(g)
		}
		if mode == 2 {
			g.EnableMuxes()
		}
		if mode == 3 {
			g.fGiaSimple = 1
		} else {
			Gia_ManHashStart(g)
		}
		pool := []int32{0, 1}
		for _, lit := range inputs {
			pool = append(pool, lit, lit^1)
		}
		for i := 0; i < 4; i++ {
			for j := i + 1; j < 4; j++ {
				for phase := int32(0); phase < 4; phase++ {
					x := Gia_ManHashAnd(g, inputs[i]^(phase&1), inputs[j]^((phase>>1)&1))
					pool = append(pool, x, x^1)
				}
			}
		}
		truth := make([]uint16, len(pool))
		for mask := 0; mask < 16; mask++ {
			bits := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0}
			values, _, err := g.evaluateObjects(bits)
			if err != nil {
				t.Fatal(err)
			}
			for i, lit := range pool {
				if values[lit>>1] != (lit&1 != 0) {
					truth[i] |= 1 << mask
				}
			}
		}
		g.fAddStrash = int32(giaBool(mode == 1))
		var checks []hashReviewTruth
		check := func(lit int32, want uint16) { checks = append(checks, hashReviewTruth{lit, want}) }
		for i, a := range pool {
			for j, b := range pool {
				tried := int32(-2)
				if mode != 3 || a < 2 || b < 2 || a>>1 == b>>1 {
					tried = Gia_ManHashAndTry(g, a, b)
				}
				x, y, z := Gia_ManHashAnd(g, a, b), Gia_ManHashOr(g, a, b), Gia_ManHashXor(g, a, b)
				real, lookup := int32(-2), int32(-2)
				if mode == 2 {
					real = Gia_ManHashXorReal(g, a, b)
					check(real, truth[i]^truth[j])
				}
				if mode != 3 && a >= 2 && b >= 2 && a>>1 != b>>1 {
					lookup = Gia_ManHashLookupInt(g, a, b)
				}
				fmt.Fprintf(&got, "pair %d %d %d %d %d %d %d %d %d %d %d %d\n", mode, i, j, tried, x, y, z, real, lookup, g.nObjs, g.nHashHit, g.nHashMiss)
				check(x, truth[i]&truth[j])
				check(y, truth[i]|truth[j])
				check(z, truth[i]^truth[j])
				if tried >= 0 {
					check(tried, truth[i]&truth[j])
				}
			}
		}
		for i := 0; i < 10; i++ {
			for j := 0; j < 10; j++ {
				for k := 0; k < 10; k++ {
					x := Gia_ManHashMux(g, pool[i], pool[j], pool[k])
					y := Gia_ManHashMaj(g, pool[i], pool[j], pool[k])
					real := int32(-2)
					mux := truth[i]&truth[j] | ^truth[i]&truth[k]
					if mode == 2 {
						real = Gia_ManHashMuxReal(g, pool[i], pool[j], pool[k])
						check(real, mux)
					}
					fmt.Fprintf(&got, "triple %d %d %d %d %d %d %d %d %d %d\n", mode, i, j, k, x, y, real, g.nObjs, g.nHashHit, g.nHashMiss)
					check(x, mux)
					check(y, truth[i]&truth[j]|truth[i]&truth[k]|truth[j]&truth[k])
				}
			}
		}
		for length, limit := 0, 1; length <= 5; length, limit = length+1, limit*6 {
			for code := 0; code < limit; code++ {
				var literals []int32
				wantAnd, wantMiter := uint16(0xffff), uint16(0)
				var previous uint16
				for position, value := 0, code; position < length; position, value = position+1, value/6 {
					index := value % 6
					literals = append(literals, pool[index])
					wantAnd &= truth[index]
					if position%2 == 0 {
						previous = truth[index]
					} else {
						wantMiter |= previous ^ truth[index]
					}
				}
				v, v2, v3 := permutation(literals...), permutation(literals...), permutation(literals...)
				x, y, z := Gia_ManHashAndMulti(g, v), Gia_ManHashAndMulti2(g, v2), Gia_ManHashDualMiter(g, v3)
				fmt.Fprintf(&got, "fold %d %d %d %d %d %d vector", mode, length, code, x, y, z)
				for _, lit := range Vec_IntArray(v) {
					fmt.Fprintf(&got, " %d", lit)
				}
				got.WriteString(" kept")
				for _, lit := range Vec_IntArray(v2) {
					fmt.Fprintf(&got, " %d", lit)
				}
				fmt.Fprintf(&got, " %d %d %d\n", g.nObjs, g.nHashHit, g.nHashMiss)
				if !slices.Equal(Vec_IntArray(v2), literals) || !slices.Equal(Vec_IntArray(v3), literals) {
					t.Fatal("non-destructive fold changed its input")
				}
				if length == 0 {
					check(x, 0)
				} else {
					check(x, wantAnd)
				}
				check(y, wantAnd)
				check(z, wantMiter)
			}
		}
		if mode != 3 {
			Gia_ManHashResize(g)
		}
		fmt.Fprintf(&got, "profile %d\n", mode)
		got.WriteString(hashReviewStdout(t, func() { Gia_ManHashProfile(g) }))
		for mask := 0; mask < 16; mask++ {
			inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0}
			values, _, err := g.evaluateObjects(inputs)
			if err != nil {
				t.Fatal(err)
			}
			for _, c := range checks {
				if (values[c.lit>>1] != (c.lit&1 != 0)) != (c.truth>>mask&1 != 0) {
					t.Fatalf("mode%d literal%d assignment%d", mode, c.lit, mask)
				}
			}
		}
	}
	if got.String() != string(native) {
		a, b := strings.Split(string(native), "\n"), strings.Split(got.String(), "\n")
		for i := 0; i < len(a) && i < len(b); i++ {
			if a[i] != b[i] {
				t.Fatalf("line%d\nC: %s\nGo: %s", i+1, a[i], b[i])
			}
		}
		t.Fatalf("line counts C%d Go%d", len(a), len(b))
	}
}
