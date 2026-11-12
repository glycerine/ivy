package dau

import (
	"bytes"
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/gia/mf"
	"github.com/glycerine/abcgo/aig/ioa"
	"github.com/glycerine/abcgo/misc/vec"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestDSDGraphConversionTruthAndBalance(t *testing.T) {
	expressions := []string{"0", "1", "a", "!a", "(abcdef)", "[abcdef]", "![a(bc)<def>]", "<abc>{(ab)[cd](ef)}", "E8{abc}", "6996{abcd}", "!(a[bc]E8{def})"}
	for _, text := range expressions {
		for _, options := range []DSDGraphOptions{{}, {Balance: true}, {Balance: true, NativeXOR: true}} {
			graph := aig.NewGraph()
			leaves := make([]aig.Edge, 6)
			for i := range leaves {
				leaves[i] = graph.NewInput()
			}
			root := Dau_DsdToGiaWithOptions(graph, text, leaves, nil, options)
			graph.AddOutput(root)
			truth := Dau_Dsd6ToTruth(text)
			for pattern := 0; pattern < 64; pattern++ {
				inputs := make([]bool, 6)
				for i := range inputs {
					inputs[i] = pattern>>i&1 != 0
				}
				out, err := graph.Evaluate(inputs)
				if err != nil {
					t.Fatal(err)
				}
				if out[0] != (truth>>uint(pattern)&1 != 0) {
					t.Fatalf("%s options%+v pattern%d", text, options, pattern)
				}
			}
		}
	}
	graph := aig.NewGraph()
	leaves := make([]aig.Edge, 6)
	for i := range leaves {
		leaves[i] = graph.NewInput()
	}
	balanced := Dau_DsdToGia(graph, "(abcdef)", leaves, nil)
	folded := Dau_DsdToGia2(graph, "(abcdef)", leaves, nil)
	if graph.Node(balanced).Level != 3 || graph.Node(folded).Level != 5 {
		t.Fatal("source balance levels", graph.Node(balanced).Level, graph.Node(folded).Level)
	}
}
func TestDSDTruthMappingRandom(t *testing.T) {
	rng := rand.New(rand.NewSource(991))
	for n := 2; n <= 8; n++ {
		for sample := 0; sample < 8; sample++ {
			truth := make([]uint64, 1<<max(0, n-6))
			for i := range truth {
				truth[i] = rng.Uint64()
			}
			if n < 6 {
				mask := uint64(1)<<uint(1<<n) - 1
				truth[0] &= mask
				for shift := 1 << n; shift < 64; shift *= 2 {
					truth[0] |= truth[0] << uint(shift)
				}
			}
			graph := aig.NewGraph()
			leaves := make([]aig.Edge, n)
			for i := range leaves {
				leaves[i] = graph.NewInput()
			}
			root := Dsm_ManTruthToGia(graph, truth, leaves, nil)
			graph.AddOutput(root)
			for pattern := 0; pattern < 1<<n; pattern++ {
				inputs := make([]bool, n)
				for i := range inputs {
					inputs[i] = pattern>>i&1 != 0
				}
				out, err := graph.Evaluate(inputs)
				if err != nil {
					t.Fatal(err)
				}
				if out[0] != (truth[pattern>>6]>>uint(pattern&63)&1 != 0) {
					t.Fatalf("truth mapping n%d sample%d pattern%d", n, sample, pattern)
				}
			}
		}
	}
}
func TestDSDEnumerationAndDivisorProperties(t *testing.T) {
	sets := EnumerateDSD(5)
	want := []int{0, 1, 2, 6}
	for i, n := range want {
		if got := vec.Vec_PtrSize(vec.Vec_PtrEntry(sets, i).(*vec.Vec_Ptr_t)); got != n {
			t.Fatalf("enumerated%d vars count%d want%d", i, got, n)
		}
	}
	for variables, level := range vec.Vec_PtrArray(sets) {
		for _, entry := range vec.Vec_PtrArray(level.(*vec.Vec_Ptr_t)) {
			text := strings.TrimPrefix(entry.(string), "*")
			letters := dsLetters(dsParse(text))
			if len(letters) != variables {
				t.Fatalf("enumerated support %s", text)
			}
		}
	}
	rows := DSDDivisorCounts([]uint64{0x8888888888888888}, 2)
	if len(rows) != 1 || rows[0].Counts != [5]int32{0, 0, 0, 1, 0} {
		t.Fatal("AND divisor", rows)
	}
}
func TestMappedDSDViaRealMF(t *testing.T) {
	model := &ioa.Model{MaxVar: 5, Inputs: []uint32{2, 4, 6}, ANDs: []ioa.And{{Literal: 8, Left: 4, Right: 2}, {Literal: 10, Left: 8, Right: 6}}, Outputs: []uint32{10, 9}}
	network, err := mf.NetworkFromModel(model)
	if err != nil {
		t.Fatal(err)
	}
	mapping, err := mf.GenerateMapping(network, mf.DefaultOptions())
	if err != nil {
		t.Fatal(err)
	}
	graph := Dsm_ManDeriveGia(mapping, 0)
	for pattern := 0; pattern < 8; pattern++ {
		inputs := []bool{pattern&1 != 0, pattern&2 != 0, pattern&4 != 0}
		out, err := graph.Evaluate(inputs)
		if err != nil {
			t.Fatal(err)
		}
		if out[0] != (inputs[0] && inputs[1] && inputs[2]) || out[1] == (inputs[0] && inputs[1]) {
			t.Fatal("mapped derivation", pattern, out)
		}
	}
}

func TestEnumerationAndDivisorsOriginalC(t *testing.T) {
	cc, err := exec.LookPath("cc")
	if err != nil {
		t.Skip("C compiler unavailable")
	}
	_, file, _, _ := runtime.Caller(0)
	src := filepath.Clean(filepath.Join(filepath.Dir(file), "../../../src"))
	dir := t.TempDir()
	enum, err := os.ReadFile(filepath.Join(src, "opt/dau/dauEnum.c"))
	if err != nil {
		t.Fatal(err)
	}
	enum = bytes.Replace(enum, []byte("nVarMax = 10"), []byte("nVarMax = 5"), 1)
	if err := os.WriteFile(filepath.Join(dir, "enum.c"), enum, 0600); err != nil {
		t.Fatal(err)
	}
	code := `#include "enum.c"
#include "opt/dau/dauDivs.c"
int main(void){word t=0x9ef7a8d9c7193a0fULL;Dau_EnumTest();printf("DIVISORS\n");Dau_DsdDivisors(&t,6);return 0;}`
	source, exe := filepath.Join(dir, "oracle.c"), filepath.Join(dir, "oracle")
	if err := os.WriteFile(source, []byte(code), 0600); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command(cc, "-std=gnu99", "-DLIN64", "-ffunction-sections", "-fdata-sections", "-I"+src, "-I"+filepath.Join(src, "opt/dau"), source, "-Wl,--gc-sections", "-o", exe).CombinedOutput(); err != nil {
		t.Fatalf("C enum/divisor oracle%v\n%s", err, out)
	}
	if err := os.MkdirAll(filepath.Join(dir, "_npn/npn"), 0700); err != nil {
		t.Fatal(err)
	}
	command := exec.Command(exe)
	command.Dir = dir
	out, err := command.Output()
	if err != nil {
		t.Fatal(err)
	}
	original, err := os.ReadFile(filepath.Join(dir, "_npn/npn/dsd10.txt"))
	if err != nil {
		t.Fatal(err)
	}
	native := filepath.Join(dir, "native.txt")
	Dau_EnumTestDump(EnumerateDSD(5), native)
	got, err := os.ReadFile(native)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, original) {
		t.Fatal("ordered DSD enumeration differs from original C")
	}
	parts := strings.SplitN(string(out), "DIVISORS\n", 2)
	if len(parts) != 2 {
		t.Fatal("missing divisor marker")
	}
	var report bytes.Buffer
	writeDSDDivisors(&report, []uint64{0x9ef7a8d9c7193a0f}, 6)
	if report.String() != parts[1] {
		t.Fatalf("source divisor report mismatch\ngot%s\nwant%s", report.String(), parts[1])
	}
}

func TestMappedDSDSequentialAndRecursiveAdapters(t *testing.T) {
	model := &ioa.Model{MaxVar: 2, Inputs: []uint32{2}, Latches: []ioa.Latch{{Literal: 4, Next: 3}}, Outputs: []uint32{4}}
	network, err := mf.NetworkFromModel(model)
	if err != nil {
		t.Fatal(err)
	}
	mapping, err := mf.GenerateMapping(network, mf.DefaultOptions())
	if err != nil {
		t.Fatal(err)
	}
	graph := Dsm_ManDeriveGia(mapping, 1)
	if aig.Aig_ManRegNum(graph) != 1 || graph.InputCount() != 2 || graph.OutputCount() != 2 {
		t.Fatal("sequential interface not preserved")
	}
	for pattern := 0; pattern < 4; pattern++ {
		out, err := graph.Evaluate([]bool{pattern&1 != 0, pattern&2 != 0})
		if err != nil {
			t.Fatal(err)
		}
		if out[0] != (pattern&2 != 0) || out[1] != (pattern&1 == 0) {
			t.Fatal("sequential boundary truth", pattern, out)
		}
	}
	p := aig.NewGraph()
	leaves := []aig.Edge{p.NewInput(), p.NewInput(), p.NewInput()}
	text := "(ab)[bc]"
	matches := Dau_DsdComputeMatches(text)
	position := int32(0)
	first := Dau_DsdToGia_rec(p, text, &position, matches, leaves, nil)
	if position != 3 {
		t.Fatal("recursive source cursor position", position)
	}
	position++
	second := Dau_DsdToGia2_rec(p, text, &position, matches, leaves, nil)
	if position != 7 {
		t.Fatal("recursive folded cursor position", position)
	}
	p.AddOutput(first)
	p.AddOutput(second)
	for pattern := 0; pattern < 8; pattern++ {
		out, err := p.Evaluate([]bool{pattern&1 != 0, pattern&2 != 0, pattern&4 != 0})
		if err != nil {
			t.Fatal(err)
		}
		if out[0] != (pattern&1 != 0 && pattern&2 != 0) || out[1] != ((pattern&2 != 0) != (pattern&4 != 0)) {
			t.Fatal("recursive graph conversion")
		}
	}
}
