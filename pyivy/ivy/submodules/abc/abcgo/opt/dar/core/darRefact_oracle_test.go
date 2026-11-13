package core
import (
	"bufio"
	"bytes"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/bool/kit"
)

func TestRefactorKitMatchesOriginalC(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("oracle requires ELF section collection")
	}
	cc, err := exec.LookPath("cc")
	if err != nil {
		t.Skip("oracle requires C compiler")
	}
	source := os.Getenv("ABC_SOURCE_ROOT")
	if source == "" {
		_, file, _, _ := runtime.Caller(0)
		source = filepath.Clean(filepath.Join(filepath.Dir(file), "../../../../src"))
	}
	bin := filepath.Join(t.TempDir(), "refactor-oracle")
	args := []string{"-O2", "-DABC_USE_STDINT_H", "-Wno-implicit-function-declaration", "-Wno-int-to-pointer-cast", "-ffunction-sections", "-fdata-sections", "-Wl,--gc-sections", "-I", source, "../test_vectors/refactor_kit_oracle.c"}
	for _, name := range []string{"kitIsop.c", "kitFactor.c", "kitSop.c", "kitTruth.c", "kitGraph.c"} {
		args = append(args, filepath.Join(source, "bool/kit", name))
	}
	args = append(args, "-o", bin)
	if output, err := exec.Command(cc, args...).CombinedOutput(); err != nil {
		t.Fatalf("compile C oracle: %v\n%s", err, output)
	}
	random := rand.New(rand.NewSource(8723))
	cases := []uint32{0xaaaaaaaa, 0x69969669, 0x88888888}
	for i := 0; i < 100; i++ {
		cases = append(cases, random.Uint32())
	}
	var in bytes.Buffer
	for _, truth := range cases {
		fmt.Fprintf(&in, "%08x\n", truth)
	}
	cmd := exec.Command(bin)
	cmd.Stdin = &in
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("run C oracle: %v\n%s", err, output)
	}
	scan := bufio.NewReader(bytes.NewReader(output))
	for _, truth := range cases {
		for phase := int32(0); phase < 2; phase++ {
			x := truth
			if phase != 0 {
				x = ^x
			}
			var cover []uint32
			kit.Kit_TruthIsop([]uint32{x}, 5, &cover, 0)
			g := kit.Kit_SopFactor(cover, phase, 5, &cover)
			var nodes int32
			var root uint32
			if _, err = fmt.Fscan(scan, &nodes, &root); err != nil {
				t.Fatalf("oracle output: %v %s", err, output)
			}
			if gotN := kit.Kit_GraphNodeNum(g); gotN != nodes {
				t.Fatalf("truth=%08x phase=%d C nodes=%d Go=%d", truth, phase, nodes, gotN)
			}
			if got := kit.Kit_EdgeToInt(kit.Kit_GraphRootEdge(g)); got != root {
				t.Fatalf("truth=%08x phase=%d root C=%d Go=%d", truth, phase, root, got)
			}
			for _, node := range kit.Kit_GraphForEachNode(g) {
				var c0, c1 uint32
				if _, err = fmt.Fscan(scan, &c0, &c1); err != nil {
					t.Fatal(err)
				}
				a, b := kit.Kit_NodeEdges(node)
				if kit.Kit_EdgeToInt(a) != c0 || kit.Kit_EdgeToInt(b) != c1 {
					t.Fatalf("truth=%08x phase=%d C edges=%d,%d Go=%d,%d", truth, phase, c0, c1, kit.Kit_EdgeToInt(a), kit.Kit_EdgeToInt(b))
				}
			}
		}
	}
}

func TestRefactorCandidateLimitsMatchOriginalC(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("oracle requires ELF section collection")
	}
	cc, err := exec.LookPath("cc")
	if err != nil {
		t.Skip("oracle requires C compiler")
	}
	source := os.Getenv("ABC_SOURCE_ROOT")
	if source == "" {
		_, file, _, _ := runtime.Caller(0)
		source = filepath.Clean(filepath.Join(filepath.Dir(file), "../../../../src"))
	}
	bin := filepath.Join(t.TempDir(), "cost-oracle")
	args := []string{"-O2", "-DABC_USE_STDINT_H", "-Wno-implicit-function-declaration", "-Wno-int-to-pointer-cast", "-ffunction-sections", "-fdata-sections", "-Wl,--gc-sections", "-I", source, "../test_vectors/refactor_cost_oracle.c", filepath.Join(source, "opt/dar/darRefact.c")}
	for _, name := range []string{"kitIsop.c", "kitFactor.c", "kitSop.c", "kitTruth.c", "kitGraph.c"} {
		args = append(args, filepath.Join(source, "bool/kit", name))
	}
	args = append(args, "-o", bin)
	if output, err := exec.Command(cc, args...).CombinedOutput(); err != nil {
		t.Fatalf("compile cost oracle: %v\n%s", err, output)
	}
	g := aig.NewGraph()
	leaves := make([]aig.Edge, 5)
	for i := range leaves {
		e := g.NewInput()
		for j := 0; j < i%3; j++ {
			e = g.AND(e, g.NewInput())
		}
		leaves[i] = e
	}
	root := g.AND(g.NewInput(), g.NewInput())
	random := rand.New(rand.NewSource(481))
	var input bytes.Buffer
	type entry struct {
		truth        uint32
		nodes, level int
	}
	var cases []entry
	for i := 0; i < 60; i++ {
		truth := random.Uint32()
		nodes := random.Intn(30)
		level := random.Intn(10)
		cases = append(cases, entry{truth, nodes, level})
		fmt.Fprintf(&input, "%08x %d %d\n", truth, nodes, level)
	}
	cmd := exec.Command(bin)
	cmd.Stdin = &input
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("run cost oracle: %v\n%s", err, out)
	}
	scan := bufio.NewReader(bytes.NewReader(out))
	for _, c := range cases {
		for phase := int32(0); phase < 2; phase++ {
			x := c.truth
			if phase != 0 {
				x = ^x
			}
			var cover []uint32
			kit.Kit_TruthIsop([]uint32{x}, 5, &cover, 0)
			k := kit.Kit_SopFactor(cover, phase, 5, &cover)
			got, _, ok := refactorTryKit(g, root, leaves, k, c.nodes, c.level)
			if !ok {
				got = -1
			}
			var want int
			if _, err := fmt.Fscan(scan, &want); err != nil {
				t.Fatal(err)
			}
			if got != want {
				t.Fatalf("truth=%x phase=%d nodes=%d level=%d C=%d Go=%d", c.truth, phase, c.nodes, c.level, want, got)
			}
		}
	}
}

func TestRefactorCutChoicesMatchOriginalC(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("oracle requires ELF section collection")
	}
	cc, err := exec.LookPath("cc")
	if err != nil {
		t.Skip("oracle requires C compiler")
	}
	source := os.Getenv("ABC_SOURCE_ROOT")
	if source == "" {
		_, file, _, _ := runtime.Caller(0)
		source = filepath.Clean(filepath.Join(filepath.Dir(file), "../../../../src"))
	}
	bin := filepath.Join(t.TempDir(), "cut-oracle")
	args := []string{"-O2", "-DABC_USE_STDINT_H", "-ffunction-sections", "-fdata-sections", "-Wl,--gc-sections", "-I", source, "../test_vectors/refactor_cut_oracle.c", "-o", bin}
	if output, err := exec.Command(cc, args...).CombinedOutput(); err != nil {
		t.Fatalf("compile C cut oracle: %v\n%s", err, output)
	}
	output, err := exec.Command(bin).CombinedOutput()
	if err != nil {
		t.Fatalf("run cut oracle: %v\n%s", err, output)
	}
	scan := bufio.NewReader(bytes.NewReader(output))
	for shared := 0; shared <= 1; shared++ {
		g := aig.NewGraph()
		a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
		middle := g.AND(a, b)
		root := g.AND(middle, c)
		g.AddOutput(root)
		if shared != 0 {
			g.AddOutput(middle)
		}
		var leaves, extended []aig.Edge
		gotSize := aig.Aig_NodeMffcSupp(g, root, 0, &leaves)
		var size, n int
		if _, err := fmt.Fscan(scan, &size, &n); err != nil {
			t.Fatal(err)
		}
		if int(gotSize) != size || len(leaves) != n {
			t.Fatalf("shared=%d cut size C=%d,%d Go=%d,%d", shared, size, n, gotSize, len(leaves))
		}
		for _, e := range leaves {
			var id int
			if _, err := fmt.Fscan(scan, &id); err != nil {
				t.Fatal(err)
			}
			if e.NodeID()-1 != id {
				t.Fatalf("shared=%d cut node C=%d Go=%d", shared, id, e.NodeID()-1)
			}
		}
		got := aig.Aig_NodeMffcExtendCut(g, root, leaves, &extended)
		var success, count int
		if _, err := fmt.Fscan(scan, &success, &count); err != nil {
			t.Fatal(err)
		}
		if got != (success != 0) || len(extended) != count {
			t.Fatalf("shared=%d extension C=%d,%d Go=%v,%d", shared, success, count, got, len(extended))
		}
		for _, e := range extended {
			var id int
			if _, err := fmt.Fscan(scan, &id); err != nil {
				t.Fatal(err)
			}
			if e.NodeID()-1 != id {
				t.Fatalf("shared=%d extension node C=%d Go=%d", shared, id, e.NodeID()-1)
			}
		}
	}
}
