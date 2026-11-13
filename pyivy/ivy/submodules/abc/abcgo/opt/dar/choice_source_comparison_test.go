package dar_test

import (
	"bufio"
	"bytes"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/opt/dar"
	"github.com/glycerine/abcgo/proof/dch"
	"github.com/glycerine/abcgo/proof/fra"
)

func TestPartitionedChoiceUsesExplicitFraig(t *testing.T) {
	g := aig.NewGraph()
	a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
	g.AddOutput(g.AND(g.OR(a, b), c))
	calls := 0
	backend := func(part *aig.Graph, conflicts, level int32) *aig.Graph {
		calls++
		return fra.Fra_FraigChoice(part, conflicts, level)
	}
	result, err := dar.Dar_ManChoiceWithFraig(g, true, true, false, 1000, 0, false, backend)
	if err != nil {
		t.Fatal(err)
	}
	if calls == 0 {
		t.Fatal("choice orchestration did not invoke FRAIG")
	}
	for mask := 0; mask < 8; mask++ {
		input := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
		want, err := g.Evaluate(input)
		if err != nil {
			t.Fatal(err)
		}
		got, err := result.Evaluate(input)
		if err != nil || got[0] != want[0] {
			t.Fatalf("input=%v got=%v want=%v err=%v", input, got, want, err)
		}
	}
}

func TestComposeActiveCChoicePipeline(t *testing.T) {
	g := aig.NewGraph()
	a, b, c, d := g.NewInput(), g.NewInput(), g.NewInput(), g.NewInput()
	g.AddOutput(g.AND(g.OR(g.AND(a, b), g.AND(a, c)), d))
	options := dar.DefaultChoiceOptions()
	versions, err := dar.Dar_NewChoiceSynthesis(g, &options)
	if err != nil {
		t.Fatal(err)
	}
	snapshots := make([]*gia.Gia_Man_t, 0, len(versions))
	for _, version := range versions {
		snapshots = append(snapshots, gia.Gia_ManFromAig(version))
	}
	miter := gia.Gia_ManChoiceMiter(snapshots)
	input := gia.Gia_ManToAigSkip(miter, 3)
	params := dch.DefaultParams()
	result, err := dch.ComputeChoices(input, &params)
	if err != nil {
		t.Fatal(err)
	}
	for mask := 0; mask < 16; mask++ {
		values := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0}
		want, _ := g.Evaluate(values)
		got, err := result.Evaluate(values)
		if err != nil || got[0] != want[0] {
			t.Fatalf("input=%v got=%v want=%v err=%v", values, got, want, err)
		}
	}
	abc, err := filepath.Abs(filepath.Join("..", "..", "..", "abc"))
	if err != nil {
		t.Fatal(err)
	}
	if info, err := os.Stat(abc); err != nil || info.IsDir() || info.Mode()&0111 == 0 {
		t.Skip("local ABC executable unavailable for full-pipeline comparison")
	}
	blif := ".model c\n.inputs a b c d\n.outputs y\n.names a b n\n11 1\n.names a c m\n11 1\n.names n m t\n1- 1\n-1 1\n.names t d y\n11 1\n.end\n"
	path := filepath.Join(t.TempDir(), "choice.blif")
	if err := os.WriteFile(path, []byte(blif), 0600); err != nil {
		t.Fatal(err)
	}
	out, err := exec.Command(abc, "-c", "read_blif "+path+"; strash; &get; &dch; &ps").CombinedOutput()
	if err != nil {
		t.Fatalf("C choice pipeline: %v\n%s", err, out)
	}
	readStat := func(name string) int {
		match := regexp.MustCompile(name + `\s*=\s*(\d+)`).FindSubmatch(out)
		if len(match) != 2 {
			t.Fatalf("missing C %s statistic: %s", name, out)
		}
		value, err := strconv.Atoi(string(match[1]))
		if err != nil {
			t.Fatal(err)
		}
		return value
	}
	if nodes, choices := readStat("and"), readStat("ch"); nodes != int(aig.Aig_ManNodeNum(result)) || choices != int(aig.Aig_ManChoiceNum(result)) {
		t.Fatalf("C nodes=%d choices=%d, Go nodes=%d choices=%d", nodes, choices, aig.Aig_ManNodeNum(result), aig.Aig_ManChoiceNum(result))
	}
}

func TestChoiceMiterAndSkipAgainstCompiledC(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("C oracle uses ELF section collection")
	}
	cc, err := exec.LookPath("cc")
	if err != nil {
		t.Skip("C compiler unavailable")
	}
	source := os.Getenv("ABC_SOURCE_ROOT")
	if source == "" {
		_, file, _, _ := runtime.Caller(0)
		source = filepath.Clean(filepath.Join(filepath.Dir(file), "../../../src"))
	}
	binary := filepath.Join(t.TempDir(), "choice-oracle")
	args := []string{"-w", "-O1", "-DABC_USE_STDINT_H", "-ffunction-sections", "-fdata-sections", "-Wl,--gc-sections", "-I", source, "test_vectors/choice_miter_oracle.c"}
	for _, file := range []string{
		"aig/gia/giaMan.c", "aig/gia/giaDup.c", "aig/gia/giaAig.c", "aig/gia/giaHash.c", "aig/gia/giaFanout.c", "aig/gia/giaExist.c", "aig/gia/giaSim.c", "aig/gia/giaUtil.c",
		"aig/aig/aigMan.c", "aig/aig/aigObj.c", "aig/aig/aigTable.c", "aig/aig/aigOper.c", "aig/aig/aigMem.c", "aig/aig/aigUtil.c", "aig/aig/aigFanout.c",
	} {
		args = append(args, filepath.Join(source, file))
	}
	args = append(args, "-o", binary)
	if output, err := exec.Command(cc, args...).CombinedOutput(); err != nil {
		t.Fatalf("compile C oracle: %v\n%s", err, output)
	}
	output, err := exec.Command(binary).CombinedOutput()
	if err != nil {
		t.Fatalf("run C oracle: %v\n%s", err, output)
	}
	var snapshots []*gia.Gia_Man_t
	for i := 0; i < 3; i++ {
		g := gia.Gia_ManStart(32)
		x, y, z := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
		if i == 1 {
			x ^= 1
		}
		v := z
		if i == 0 {
			v = y
		}
		first := gia.Gia_ManAppendAnd(g, x, v)
		v = z
		if i == 2 {
			v = y
		}
		second := gia.Gia_ManAppendAnd(g, first, v)
		gia.Gia_ManAppendCo(g, first)
		gia.Gia_ManAppendCo(g, second)
		snapshots = append(snapshots, g)
	}
	got := gia.Gia_ManToAigSkip(gia.Gia_ManChoiceMiter(snapshots), 3)
	reader := bufio.NewReader(bytes.NewReader(output))
	var count int
	if _, err := fmt.Fscan(reader, &count); err != nil || count != len(got.Nodes()) {
		t.Fatalf("C node count %d, Go %d: %v", count, len(got.Nodes()), err)
	}
	for i, node := range got.Nodes() {
		var kind, first, second int
		if _, err := fmt.Fscan(reader, &kind, &first, &second); err != nil {
			t.Fatal(err)
		}
		lit := func(edge aig.Edge) int {
			if edge == aig.InvalidEdge {
				return -1
			}
			return int(edge)
		}
		firstGo, secondGo := lit(node.Fanins[0]), lit(node.Fanins[1])
		if node.Kind == aig.Constant || node.Kind == aig.Input {
			firstGo, secondGo = -1, -1
		}
		if int(node.Kind) != kind || firstGo != first || secondGo != second {
			t.Fatalf("node %d: C=(%d,%d,%d) Go=(%d,%d,%d)", i, kind, first, second, node.Kind, lit(node.Fanins[0]), lit(node.Fanins[1]))
		}
	}
}

// Compare the graphs entering DCH, not merely output truth tables: ordering
// and sharing can affect simulation classes and the resulting choice links.
func TestChoiceMiterAndTotalAIGHaveDifferentNodeOrder(t *testing.T) {
	rng := rand.New(rand.NewSource(913))
	different := false
	for trial := 0; trial < 20; trial++ {
		g := aig.NewGraph()
		pool := []aig.Edge{g.NewInput(), g.NewInput(), g.NewInput(), g.NewInput()}
		for i := 0; i < 20; i++ {
			a, b := pool[rng.Intn(len(pool))], pool[rng.Intn(len(pool))]
			if rng.Intn(2) != 0 {
				a = a.Not()
			}
			if rng.Intn(2) != 0 {
				b = b.Not()
			}
			pool = append(pool, g.AND(a, b))
		}
		g.AddOutput(pool[len(pool)-1])
		g.AddOutput(pool[len(pool)-3])
		options := dar.DefaultChoiceOptions()
		versions, err := dar.Dar_NewChoiceSynthesis(g, &options)
		if err != nil {
			t.Fatal(err)
		}
		var gias []*gia.Gia_Man_t
		for _, version := range versions {
			gias = append(gias, gia.Gia_ManFromAig(version))
		}
		miter := gia.Gia_ManChoiceMiter(gias)
		want := gia.Gia_ManToAigSkip(miter, 3)
		got, err := dch.DeriveTotalAig(versions)
		if err != nil {
			t.Fatal(err)
		}
		if len(got.Nodes()) != len(want.Nodes()) {
			different = true
			continue
		}
		for i, node := range got.Nodes() {
			other := want.Nodes()[i]
			if node.Kind != other.Kind || node.Fanins != other.Fanins {
				different = true
				break
			}
		}
	}
	if !different {
		t.Fatal("expected a source-path graph with a different node order")
	}
}
