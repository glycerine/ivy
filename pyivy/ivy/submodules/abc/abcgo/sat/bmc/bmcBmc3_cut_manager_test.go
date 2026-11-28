package bmc

import (
	"context"
	"fmt"
	"github.com/glycerine/abcgo/internal/testoracle"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/util"
)

func TestScalableBMCNormalizesNativeXORToAIG(t *testing.T) {
	g := aig.NewGraph()
	a, b := g.NewInput(), g.NewInput()
	g.AddOutput(g.NewXor(a, b))
	params := Saig_ParBmc_t{NFramesMax: 1}
	if result := Saig_ManBmcScalable(g, &params); result != 0 {
		t.Fatalf("result=%d, want XOR counterexample", result)
	}
	cex := g.SequentialCounterexample()
	if cex == nil {
		t.Fatal("missing XOR witness")
	}
	inputs := make([]bool, 2)
	inputs[0] = util.Abc_InfoHasBit(cex.PData, 0)
	inputs[1] = util.Abc_InfoHasBit(cex.PData, 1)
	out, err := g.Evaluate(inputs)
	if err != nil || !out[0] {
		t.Fatalf("witness inputs=%v outputs=%v err=%v", inputs, out, err)
	}
}

func TestScalableBMCMultiOutputMatchesCUnderConflictBudget(t *testing.T) {
	binary := testoracle.NativeABCPath()
	if binary == "" {
		binary = testoracle.NativeABCPath()
	}
	absolute, err := filepath.Abs(binary)
	if err != nil {
		t.Fatal(err)
	}
	if info, err := testoracle.Stat(absolute); err != nil || info.IsDir() || info.Mode()&0111 == 0 {
		t.Skip("C ABC executable unavailable")
	}
	g := aig.NewGraph()
	pi, state := g.NewInput(), g.NewInput()
	g.AddOutput(g.AND(pi, state))
	g.AddOutput(pi)
	g.AddOutput(pi)
	aig.Aig_ManSetRegNum(g, 1)
	fixture, err := bmcAIGFixture(g)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "multi.aig")
	if err := os.WriteFile(path, fixture.aiger(), 0600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	out, err := testoracle.CommandContext(t, ctx, absolute, "-c", fmt.Sprintf("read_aiger %s; bmc3 -v -a -x -C 1 -F 3", path)).CombinedOutput()
	if err != nil {
		t.Fatalf("C ABC: %v\n%s", err, out)
	}
	re := regexp.MustCompile(`Output\s+(\d+) was asserted in frame\s+(\d+)`)
	cFrames := map[int32]int32{}
	for _, match := range re.FindAllStringSubmatch(string(out), -1) {
		o, _ := strconv.Atoi(match[1])
		f, _ := strconv.Atoi(match[2])
		cFrames[int32(o)] = int32(f)
	}
	if len(cFrames) != 2 || cFrames[0] != 1 || cFrames[1] != 0 {
		t.Fatalf("C multi-output frames=%v\n%s", cFrames, out)
	}
	params := Saig_ParBmc_t{NFramesMax: 3, NConfLimit: 1, FSolveAll: 1, FStoreCex: 1}
	if result := Saig_ManBmcScalable(g, &params); result != 0 {
		t.Fatalf("Go result=%d, want failing", result)
	}
	for output, want := range cFrames {
		cex := bmcStoredCex(g, output)
		if cex == nil || cex.IPo != output || cex.IFrame != want {
			t.Fatalf("output %d Go CEX=%+v, C frame=%d\n%s", output, cex, want, out)
		}
	}
}

func bmcAIGFixture(graph *aig.Graph) (abcFixture, error) {
	if graph == nil {
		return abcFixture{}, fmt.Errorf("nil graph")
	}
	nodes, inputs, outputs := graph.Nodes(), graph.Inputs(), graph.Outputs()
	mapped := make([]int, len(nodes))
	mapped[0] = 1
	for i, e := range inputs {
		mapped[e.NodeID()] = 2 * (i + 1)
	}
	fixture := abcFixture{pis: len(inputs) - int(aig.Aig_ManRegNum(graph)), regs: int(aig.Aig_ManRegNum(graph))}
	child := func(e aig.Edge) int {
		v := mapped[e.NodeID()]
		if e.Complemented() {
			v ^= 1
		}
		return v
	}
	for id, n := range nodes {
		switch n.Kind {
		case aig.Constant, aig.Input:
		case aig.And:
			a, b := child(n.Fanins[0]), child(n.Fanins[1])
			mapped[id] = 2 * (len(inputs) + len(fixture.ands) + 1)
			fixture.ands = append(fixture.ands, [2]int{a, b})
		case aig.Output, aig.Buffer:
			mapped[id] = child(n.Fanins[0])
		case aig.ExclusiveOr:
			return abcFixture{}, fmt.Errorf("fixture requires source AND graph")
		}
	}
	pos := len(outputs) - fixture.regs
	for i, e := range outputs {
		n := nodes[e.NodeID()]
		lit := child(n.Fanins[0])
		if i < pos {
			fixture.outs = append(fixture.outs, lit)
		} else {
			fixture.next = append(fixture.next, lit)
		}
	}
	return fixture, nil
}

func TestScalableBMCConflictUnknownMatchesC(t *testing.T) {
	binary := testoracle.NativeABCPath()
	if binary == "" {
		binary = testoracle.NativeABCPath()
	}
	absolute, err := filepath.Abs(binary)
	if err != nil {
		t.Fatal(err)
	}
	if info, err := testoracle.Stat(absolute); err != nil || info.IsDir() || info.Mode()&0111 == 0 {
		t.Skip("C ABC executable unavailable")
	}
	g := bmcPigeonholeGraph(4, 3)
	fixture, err := bmcAIGFixture(g)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "php.aig")
	if err := os.WriteFile(path, fixture.aiger(), 0600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	out, err := testoracle.CommandContext(t, ctx, absolute, "-c", fmt.Sprintf("read_aiger %s; bmc3 -v -C 1", path)).CombinedOutput()
	if err != nil {
		t.Fatalf("C ABC: %v\n%s", err, out)
	}
	text := string(out)
	if strings.Contains(text, "Stopping BMC because all 2^0 reachable states are visited") || regexp.MustCompile(`Output\s+0 was asserted`).MatchString(text) {
		t.Fatalf("C did not stop at conflict-budget UNKNOWN:\n%s", text)
	}
	params := Saig_ParBmc_t{NConfLimit: 1}
	if result := Saig_ManBmcScalable(g, &params); result != -1 {
		t.Fatalf("Go result=%d, want unknown (-1); C output:\n%s", result, text)
	}
}

func bmcPigeonholeGraph(pigeons, holes int) *aig.Graph {
	g := aig.NewGraph()
	x := make([]aig.Edge, pigeons*holes)
	for i := range x {
		x[i] = g.NewInput()
	}
	clause := func(lits ...aig.Edge) aig.Edge {
		result := g.Const0()
		for _, lit := range lits {
			result = g.OR(result, lit)
		}
		return result
	}
	constraints := make([]aig.Edge, 0)
	for p := 0; p < pigeons; p++ {
		atLeast := make([]aig.Edge, holes)
		for h := 0; h < holes; h++ {
			atLeast[h] = x[p*holes+h]
		}
		constraints = append(constraints, clause(atLeast...))
		for a := 0; a < holes; a++ {
			for b := a + 1; b < holes; b++ {
				constraints = append(constraints, clause(x[p*holes+a].Not(), x[p*holes+b].Not()))
			}
		}
	}
	for h := 0; h < holes; h++ {
		for a := 0; a < pigeons; a++ {
			for b := a + 1; b < pigeons; b++ {
				constraints = append(constraints, clause(x[a*holes+h].Not(), x[b*holes+h].Not()))
			}
		}
	}
	property := g.Const1()
	for _, c := range constraints {
		property = g.AND(property, c)
	}
	g.AddOutput(property)
	return g
}
