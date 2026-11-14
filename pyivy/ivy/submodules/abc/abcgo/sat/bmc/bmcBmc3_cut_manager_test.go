package bmc

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/util"
)

func TestBmcCutManagerMapsPIsAndLatchFramesLazily(t *testing.T) {
	g := aig.NewGraph()
	pi, state := g.NewInput(), g.NewInput()
	property := g.AND(pi, state)
	g.AddOutput(property)
	g.AddOutput(pi)
	aig.Aig_ManSetRegNum(g, 1)
	m, err := newBmc3CutManager(g, 1, 1, 1)
	if err != nil {
		t.Fatal(err)
	}
	if lit, ok := m.outputLiteral(0, 0); !ok || lit != 0 || m.variableCount() != 0 {
		t.Fatalf("frame-0 known-false output literal=%d known=%t vars=%d", lit, ok, m.variableCount())
	}
	if len(m.drainClauses()) != 0 {
		t.Fatal("known-false frame emitted CNF")
	}
	lit, ok := m.outputLiteral(1, 0)
	if !ok || lit < 2 || lit&1 != 0 {
		t.Fatalf("frame-1 output literal=%d known=%t", lit, ok)
	}
	if got := m.variableCount(); got != 3 {
		t.Fatalf("variables=%d, want two PI vars and one cut var", got)
	}
	clauses := m.drainClauses()
	if len(clauses) == 0 {
		t.Fatal("unknown frame emitted no cut clauses")
	}
	for i, clause := range clauses {
		for _, cLit := range clause {
			if cLit < 2 {
				t.Fatalf("clause %d contains a constant literal: %v", i, clause)
			}
			if _, err := bmcCToBSATLiteral(cLit); err != nil {
				t.Fatalf("clause %d literal %d: %v", i, cLit, err)
			}
		}
	}
}

func TestBmcCutCountsMatchCompiledC(t *testing.T) {
	binary := os.Getenv("ABC_BINARY")
	if binary == "" {
		binary = filepath.Join("..", "..", "..", "abc")
	}
	absolute, err := filepath.Abs(binary)
	if err != nil {
		t.Fatal(err)
	}
	if info, err := os.Stat(absolute); err != nil || info.IsDir() || info.Mode()&0111 == 0 {
		t.Skip("C ABC executable unavailable")
	}
	// One PI, one zero-initialized register, an AND property, and PI-fed next state.
	aiger := append([]byte("aig 3 1 1 1 1\n2\n6\n"), 2, 2)
	path := filepath.Join(t.TempDir(), "cut.aig")
	if err := os.WriteFile(path, aiger, 0600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	command := fmt.Sprintf("read_aiger %s; bmc3 -v -F 3", path)
	output, err := exec.CommandContext(ctx, absolute, "-c", command).CombinedOutput()
	if err != nil {
		t.Fatalf("C ABC: %v\n%s", err, output)
	}
	match := regexp.MustCompile(`(?m)^\s*(\d+)\s+[+-] : Var =\s*(\d+)\. Cla =\s*(\d+)\.`).FindAllStringSubmatch(string(output), -1)
	if len(match) < 2 {
		t.Fatalf("missing per-frame C stats:\n%s", output)
	}
	frame, _ := strconv.Atoi(match[1][1])
	cVars, _ := strconv.Atoi(match[1][2])
	cClauses, _ := strconv.Atoi(match[1][3])
	if frame != 1 {
		t.Fatalf("C witness frame stats=%d, want frame 1", frame)
	}
	g := aig.NewGraph()
	pi, state := g.NewInput(), g.NewInput()
	g.AddOutput(g.AND(pi, state))
	g.AddOutput(pi)
	aig.Aig_ManSetRegNum(g, 1)
	m, err := newBmc3CutManager(g, 1, 1, 1)
	if err != nil {
		t.Fatal(err)
	}
	var goClauses int
	for f := 0; f <= frame; f++ {
		if _, ok := m.outputLiteral(f, 0); !ok {
			t.Fatalf("Go output literal missing at frame %d", f)
		}
		goClauses += len(m.drainClauses())
	}
	goVars := int(m.variableCount() + 1)
	if goVars != cVars || goClauses != cClauses {
		t.Fatalf("frame %d C vars/clauses=%d/%d Go=%d/%d; C output:\n%s", frame, cVars, cClauses, goVars, goClauses, strings.TrimSpace(string(output)))
	}
}

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
	binary := os.Getenv("ABC_BINARY")
	if binary == "" {
		binary = filepath.Join("..", "..", "..", "abc")
	}
	absolute, err := filepath.Abs(binary)
	if err != nil {
		t.Fatal(err)
	}
	if info, err := os.Stat(absolute); err != nil || info.IsDir() || info.Mode()&0111 == 0 {
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
	out, err := exec.CommandContext(ctx, absolute, "-c", fmt.Sprintf("read_aiger %s; bmc3 -v -a -x -C 1 -F 3", path)).CombinedOutput()
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
	binary := os.Getenv("ABC_BINARY")
	if binary == "" {
		binary = filepath.Join("..", "..", "..", "abc")
	}
	absolute, err := filepath.Abs(binary)
	if err != nil {
		t.Fatal(err)
	}
	if info, err := os.Stat(absolute); err != nil || info.IsDir() || info.Mode()&0111 == 0 {
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
	out, err := exec.CommandContext(ctx, absolute, "-c", fmt.Sprintf("read_aiger %s; bmc3 -v -C 1", path)).CombinedOutput()
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
