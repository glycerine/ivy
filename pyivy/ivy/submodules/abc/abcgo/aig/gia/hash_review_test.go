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
		g.fAddStrash = giaBool(mode == 1)
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
					lookup = Gia_ManHashLookup(g, Gia_ObjFromLit(g, a), Gia_ObjFromLit(g, b))
					if lookup != Gia_ManHashLookupInt(g, a, b) {
						t.Fatal("pointer and literal lookup disagree")
					}
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
			Gia_ManHashStop(g)
			Gia_ManHashStart(g)
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
	hashStaticReviewSnapshots(t, &got)
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

func BenchmarkHashConstructionAndResize(b *testing.B) {
	b.ReportAllocs()
	for run := 0; run < b.N; run++ {
		g := Gia_ManStart(8)
		inputs := make([]int32, 16)
		for i := range inputs {
			inputs[i] = Gia_ManAppendCi(g)
		}
		Gia_ManHashStart(g)
		for i := 0; i < len(inputs); i++ {
			for j := i + 1; j < len(inputs); j++ {
				for phase := int32(0); phase < 4; phase++ {
					Gia_ManHashAnd(g, inputs[i]^(phase&1), inputs[j]^((phase>>1)&1))
				}
			}
		}
		Gia_ManHashResize(g)
	}
}

func hashStaticReviewSnapshots(t *testing.T, got *strings.Builder) {
	t.Helper()
	literals := []int32{0, 1, 2, 3, 6, 7, 30, 31, 62, 63, 254, 255}
	for _, a := range literals {
		for _, b := range literals {
			for _, control := range []int32{-1, 0, 1, 2, 63, 126, 127} {
				for _, size := range []int32{3, 11, 67, 1009} {
					fmt.Fprintf(got, "key %d %d %d %d %d\n", a, b, control, size, Gia_ManHashOne(a, b, control, size))
				}
			}
		}
	}
	g := Gia_ManStart(64)
	inputs := make([]int32, 4)
	for i := range inputs {
		inputs[i] = Gia_ManAppendCi(g)
	}
	var nodes []int32
	for i := 0; i < 4; i++ {
		for j := i + 1; j < 4; j++ {
			for phase := int32(0); phase < 4; phase++ {
				nodes = append(nodes, Gia_ManAppendAnd(g, inputs[i]^(phase&1), inputs[j]^((phase>>1)&1)))
			}
		}
	}
	values := make([][]bool, 16)
	for mask := range values {
		bits := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0}
		var err error
		values[mask], _, err = g.evaluateObjects(bits)
		if err != nil {
			t.Fatal(err)
		}
	}
	for i, lit0 := range nodes {
		for j, lit1 := range nodes {
			x, y := Gia_ManObj(g, lit0>>1), Gia_ManObj(g, lit1>>1)
			a, b, _, _ := g.Gate(x.id)
			c, d, _, _ := g.Gate(y.id)
			if a != c^1 && a != d^1 && b != c^1 && b != d^1 {
				continue
			}
			var yes, no *Gia_Obj_t
			control := Gia_ObjRecognizeMuxTwo(x, y, &yes, &no)
			ctrl, then, otherwise := Gia_ObjToLit(g, control), Gia_ObjToLit(g, yes), Gia_ObjToLit(g, no)
			fmt.Fprintf(got, "recognize %d %d %d %d %d\n", i, j, ctrl, then, otherwise)
			if ctrl&1 != 0 {
				t.Fatal("recognized control is complemented")
			}
			for _, assignment := range values {
				eval := func(lit int32) bool { return assignment[lit>>1] != (lit&1 != 0) }
				value := eval(otherwise)
				if eval(ctrl) {
					value = eval(then)
				}
				if value != (!eval(lit0) && !eval(lit1)) {
					t.Fatalf("recognizer changed NOR-of-products for nodes %d/%d", i, j)
				}
			}
		}
	}
}

func TestTypedHashLifecycleRetainsCanonicalNodes(t *testing.T) {
	g := Gia_ManStart(8)
	a, b, c, d := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	g.EnableMuxes()
	Gia_ManHashStart(g)
	and := Gia_ManHashAnd(g, b, c)
	xor := Gia_ManHashXorReal(g, b, c)
	mux0, mux1 := Gia_ManHashMuxReal(g, a, b, c), Gia_ManHashMuxReal(g, d, b, c)
	buf := Gia_ManAppendBuf(g, a)
	count := g.nObjs
	verify := func() {
		if Gia_ManHashAnd(g, c, b) != and || Gia_ManHashXorReal(g, c, b) != xor || Gia_ManHashMuxReal(g, a^1, c, b) != mux0 || Gia_ManHashMuxReal(g, d^1, c, b) != mux1 || g.nObjs != count {
			t.Fatal("hash lifecycle changed canonical identity")
		}
	}
	for repeat := 0; repeat < 3; repeat++ {
		Gia_ManHashResize(g)
		verify()
		Gia_ManHashStop(g)
		Gia_ManHashStart(g)
		verify()
		seen := make(map[int32]bool)
		for bucket, head := range Vec_IntArray(&g.vHTable) {
			for head != 0 {
				if head < 1 || head >= g.nObjs || seen[head] {
					t.Fatal("duplicate/invalid hash chain entry")
				}
				seen[head] = true
				obj := Gia_ManObj(g, head)
				key := Gia_ManHashOne(Gia_ObjFaninLit0p(g, obj), Gia_ObjFaninLit1p(g, obj), Gia_ObjFaninLit2p(g, obj), int32(Vec_IntSize(&g.vHTable)))
				if int(key) != bucket {
					t.Fatal("gate stored in wrong bucket")
				}
				head = Vec_IntEntry(&g.vHash, int(head))
			}
		}
		if len(seen) != 4 || seen[buf>>1] {
			t.Fatal("buffer or terminal was hashed as an ordinary gate")
		}
	}
}

func TestEmptyHashLifecycle(t *testing.T) {
	g := Gia_ManStart(1)
	Gia_ManHashStart(g)
	Gia_ManHashResize(g)
	profile := hashReviewStdout(t, func() { Gia_ManHashProfile(g) })
	if !strings.HasSuffix(profile, "Misses = 0.\n\n") {
		t.Fatalf("empty profile %q", profile)
	}
	Gia_ManHashStop(g)
	if Vec_IntSize(&g.vHTable) != 0 || Vec_IntSize(&g.vHash) != 0 {
		t.Fatal("stopped table retained entries")
	}
	Gia_ManHashStart(g)
	if g.nObjs != 1 {
		t.Fatal("hash lifecycle added graph nodes")
	}
}
