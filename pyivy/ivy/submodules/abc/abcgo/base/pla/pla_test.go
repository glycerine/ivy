package pla

import (
	"bytes"
	"errors"
	"fmt"
	abcmain "github.com/glycerine/abcgo/base/main"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func nativeDump(p *Pla_Man_t) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%d %d %d %d\n", p.Type, p.Ins, p.Outs, len(p.Cubes))
	for _, c := range p.Cubes {
		for _, v := range c.In {
			fmt.Fprintf(&b, "%016x ", v)
		}
		for _, v := range c.Out {
			fmt.Fprintf(&b, "%016x ", v)
		}
		b.WriteByte('\n')
	}
	fmt.Fprintf(&b, "dist %d hash %d\n", Pla_ManDist1Num(p), Pla_ManHashDistance1(p))
	return b.String()
}
func TestOriginalCFormatAndGeneratorOracle(t *testing.T) {
	cases := []struct{ file, want string }{
		{"test_vectors/fdr.pla", "3 5 2 4\n0000000000000155 000000000000000e \n0000000000000156 0000000000000009 \n000000000000015a 0000000000000004 \n000000000000016a 000000000000000c \ndist 3 hash 3\n"},
		{"test_vectors/fd.pla", "0 4 1 5\n0000000000000055 0000000000000002 \n0000000000000056 0000000000000002 \n000000000000005a 0000000000000002 \n000000000000006a 0000000000000002 \n00000000000000aa 0000000000000001 \ndist 4 hash 4\n"}}
	for _, c := range cases {
		p, err := ReadFile(c.file)
		if err != nil {
			t.Fatal(err)
		}
		if got := nativeDump(p); got != c.want {
			t.Fatalf("%s\ngot%s\nwant%s", c.file, got, c.want)
		}
	}
	Pla_SeedRandom(0)
	p := Pla_ManGenerate(8, 4, 5, 0)
	want := "0 8 4 5\n000000000000559a 0000000000000059 \n0000000000005699 000000000000005a \n000000000000595a 0000000000000066 \n0000000000006599 0000000000000069 \n0000000000009669 00000000000000aa \ndist 0 hash 0\n"
	if got := nativeDump(p); got != want {
		t.Fatalf("native RNG/generator mismatch\n%s", got)
	}
}
func TestPackedRoundtripAllTypesAndWidths(t *testing.T) {
	rng := rand.New(rand.NewSource(731))
	for _, width := range []int32{1, 31, 32, 33, 64, 65, 129} {
		for kind := PLA_FILE_FD; kind <= PLA_FILE_FDR; kind++ {
			p := Pla_ManAlloc("test.pla", width, 33, 9)
			p.Type = kind
			p.Directives = []string{".ilb " + strings.Repeat("x ", int(width))}
			for i := range p.Cubes {
				for v := int32(0); v < p.Ins; v++ {
					Pla_CubeSetLit(p.Cubes[i].In, v, Pla_Lit_t(rng.Intn(4)))
				}
				for v := int32(0); v < p.Outs; v++ {
					Pla_CubeSetLit(p.Cubes[i].Out, v, Pla_Lit_t(rng.Intn(4)))
				}
			}
			var b bytes.Buffer
			if err := p.Write(&b); err != nil {
				t.Fatal(err)
			}
			q, err := Parse(&b, "test.pla")
			if err != nil {
				t.Fatal(err)
			}
			if p.Type != q.Type || !reflect.DeepEqual(p.Cubes, q.Cubes) {
				t.Fatalf("type%d width%d lost literal data", kind, width)
			}
			Pla_ManConvertFromBits(p)
			Pla_ManConvertToBits(p)
			if !reflect.DeepEqual(p.Cubes, q.Cubes) {
				t.Fatal("literal array conversion lost FULL/word-boundary data")
			}
		}
	}
}
func TestPLAParserValidationAndTypeSemantics(t *testing.T) {
	for kind, want := range map[string]Pla_Lit_t{"f": PLA_LIT_ZERO, "fd": PLA_LIT_ZERO, "fr": PLA_LIT_DASH, "fdr": PLA_LIT_FULL} {
		p, err := Parse(strings.NewReader(".type "+kind+"\n.i\t1\n.o 1\n.p 3\n- ~ #inline\n.e\nnot data\n"), "t")
		if err != nil {
			t.Fatal(err)
		}
		if Pla_CubeGetLit(p.Cubes[0].Out, 0) != want || len(p.Warnings) != 1 {
			t.Fatal(kind, p.Warnings)
		}
	}
	for _, text := range []string{".i 0\n.o 1\n", ".i 1\n.o 1\n11 1\n", ".i 1\n.o 1\n0 x\n", ".i 1\n.o 1\n.type bad\n", ".i 1\n.o 1\n.mv 2 0 2 2\n", ".i 1\n.o 1\n.p -1\n"} {
		if _, err := Parse(strings.NewReader(text), "bad"); err == nil {
			t.Fatal("accepted malformed input", text)
		}
	}
	p, err := Parse(strings.NewReader(".i 1\n.o 1\n.p 0\n.e"), "empty")
	if err != nil || len(p.Cubes) != 0 {
		t.Fatal(p, err)
	}
}
func TestDistanceAlgorithmsAndPrimes(t *testing.T) {
	p := Pla_ManPrimesDetector(6)
	if len(p.Cubes) != 18 {
		t.Fatal("prime sieve", len(p.Cubes))
	}
	rng := rand.New(rand.NewSource(41))
	for trial := 0; trial < 100; trial++ {
		permutation := rng.Perm(256)
		p = Pla_GenFromMinterms("random", ints32(permutation[:30]), 8)
		if got, want := Pla_ManHashDistance1(p), Pla_ManDist1Num(p); got != want {
			t.Fatal("distance hash", got, want)
		}
		Pla_ManConvertFromBits(p)
		pairs := Pla_ManComputeDistance1(p)
		for i := 0; i < len(pairs); i += 4 {
			a, b := pairs[i], pairs[i+2]
			if Pla_CubesAreDistance1(p.Cubes[a].In, p.Cubes[b].In, 1) != 1 {
				t.Fatal("incorrect skip-signature pair", pairs[i:i+4])
			}
		}
	}
}
func ints32(a []int) []int32 {
	out := make([]int32, len(a))
	for i, v := range a {
		out[i] = int32(v)
	}
	return out
}
func evalPackedCube(c uint64, bits, n int) bool {
	for v := 0; v < n; v++ {
		lit := Pla_CubeGetLit([]uint64{c}, int32(v))
		if lit == PLA_LIT_FULL || lit == PLA_LIT_ONE && bits&(1<<v) == 0 || lit == PLA_LIT_ZERO && bits&(1<<v) != 0 {
			return false
		}
	}
	return true
}
func TestSimpleCoverAllThreeVariableFunctions(t *testing.T) {
	for truth := uint64(0); truth < 256; truth++ {
		cubes := Pla_ManFxMinimize([]uint64{truth}, 3)
		for bits := 0; bits < 8; bits++ {
			got := false
			for _, cube := range cubes {
				got = got || evalPackedCube(cube, bits, 3)
			}
			if got != (truth&(1<<bits) != 0) {
				t.Fatal(truth, bits, cubes)
			}
		}
	}
}
func TestPLAFrameCommandsAndFailedReadOwnership(t *testing.T) {
	f := abcmain.NewFrame()
	var out, errs bytes.Buffer
	f.Out = &out
	f.Err = &errs
	Pla_Init(f)
	for _, name := range []string{"|read", "|write", "|ps", "|gen", "|merge", "|test"} {
		if _, ok := f.LookupCommand(name); !ok {
			t.Fatal(name)
		}
	}
	if Abc_CommandGen(f, 9, []string{"|gen", "-I", "4", "-O", "1", "-P", "4", "-S", "0"}) != 0 {
		t.Fatal(errs.String())
	}
	old := Pla_AbcGetMan(f)
	name := filepath.Join(t.TempDir(), "out.pla")
	if Abc_CommandWritePla(f, 2, []string{"|write", name}) != 0 {
		t.Fatal("write")
	}
	if _, err := os.Stat(name); err != nil {
		t.Fatal(err)
	}
	if Abc_CommandReadPla(f, 2, []string{"|read", "missing.pla"}) != 0 || Pla_AbcGetMan(f) != old {
		t.Fatal("failed read destroyed current design")
	}
	if Abc_CommandReadPla(f, 2, []string{"|read", name}) != 0 || Pla_AbcGetMan(f) == old {
		t.Fatal("read replacement")
	}
	if Abc_CommandPs(f, 1, []string{"|ps"}) != 0 || !strings.Contains(out.String(), "Cube") {
		t.Fatal(out.String())
	}
	Pla_End(f)
	if Pla_AbcGetMan(f) != nil {
		t.Fatal("end retained freed pointer")
	}
	Pla_End(f)
}

type brokenReader struct{}

func (brokenReader) Read([]byte) (int, error) { return 0, errors.New("read failed") }

type brokenWriter struct{}

func (brokenWriter) Write([]byte) (int, error) { return 0, io.ErrClosedPipe }
func TestPLAIOErrors(t *testing.T) {
	if _, err := Parse(brokenReader{}, "bad"); err == nil {
		t.Fatal("read failure lost")
	}
	p := Pla_ManAlloc("p", 1, 1, 0)
	if err := p.Write(brokenWriter{}); err == nil {
		t.Fatal("write failure lost")
	}
}

func TestHashDistanceCountsDuplicateCubeMultiplicity(t *testing.T) {
	p := Pla_GenFromMinterms("duplicates", []int32{0, 0, 1, 3}, 2)
	if got, want := Pla_ManHashDistance1(p), Pla_ManDist1Num(p); got != want || want != 3 {
		t.Fatal("duplicate-cube distance pairs", got, want)
	}
}
