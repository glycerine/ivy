package aig

import (
	"fmt"
	"strings"
	"testing"
)

func TestTernaryRandomHelpersConsumeSourceStream(t *testing.T) {
	Aig_ManRandom(1)
	defer Aig_ManRandom(1)
	r := NewAigRandomSource()
	r.Uint32() // Reset returns and consumes the first source word.
	for i := 0; i < 32; i++ {
		want2 := 1 + int32(r.Uint32()&1)
		if got := Aig_XsimRand2(); got != want2 {
			t.Errorf("binary draw %d = %d, want %d", i, got, want2)
		}
		var want3 int32
		for want3 == 0 {
			want3 = int32(r.Uint32() & 3)
		}
		if got := Aig_XsimRand3(); got != want3 {
			t.Errorf("ternary draw %d = %d, want %d", i, got, want3)
		}
	}
	if got, want := Aig_ManRandom(0), r.Uint32(); got != want {
		t.Errorf("shared stream advanced to %d, want %d", got, want)
	}
}

func TestTernaryRandomHelpersAgainstCompiledSource(t *testing.T) {
	out := runAigSourceFixture(t, "ternary_random_compare.c", "aigUtil.c")
	Aig_ManRandom(1)
	defer Aig_ManRandom(1)
	var got strings.Builder
	for i := 0; i < 32; i++ {
		binary, ternary := Aig_XsimRand2(), Aig_XsimRand3()
		fmt.Fprintf(&got, "%d %d\n", binary, ternary)
	}
	fmt.Fprintf(&got, "next %d\n", Aig_ManRandom(0))
	if got.String() != out {
		t.Fatalf("Go stream:\n%sC stream:\n%s", got.String(), out)
	}
}

func TestTernaryMappingAcrossPackedWordsAgainstSource(t *testing.T) {
	out := runAigSourceFixture(t, "ternary_mapping_compare.c", "aigMem.c")
	p := NewGraph()
	input := p.NewInput()
	los := make([]Edge, 18)
	for i := range los {
		los[i] = p.NewInput()
	}
	for i, e := range los {
		switch i % 4 {
		case 0:
			p.AddOutput(e)
		case 1:
			p.AddOutput(e.Not())
		case 2:
			p.AddOutput(input)
		case 3:
			p.AddOutput(p.Const1())
		}
	}
	p.nRegs = 18
	var got strings.Builder
	got.WriteString("mapping:")
	for _, e := range Aig_ManTernarySimulate(p, 0, 0) {
		fmt.Fprintf(&got, " %d", e)
	}
	got.WriteString("\nvalues:")
	for _, n := range p.nodes {
		fmt.Fprintf(&got, " %d", n.Xsim)
	}
	got.WriteByte('\n')
	if got.String() != out {
		t.Fatalf("Go simulation:\n%sC simulation:\n%s", got.String(), out)
	}
}

func TestTernarySimulationSourceScratchScope(t *testing.T) {
	for _, regs := range []int32{0, 1} {
		p := NewGraph()
		a, b := p.NewInput(), p.NewInput()
		gate := p.AND(a, b)
		po := p.AddOutput(gate)
		Aig_ObjSetXsim(p, po, AIG_XVS1)
		if regs != 0 {
			p.AddOutput(b)
		}
		p.nRegs = regs
		Aig_ManTernarySimulate(p, 0, 0)
		if p.node(po).Xsim != AIG_XVS1 {
			t.Errorf("regs=%d: PO scratch overwritten by source LI-only transfer", regs)
		}
		wantGate := AIG_XVSX
		if regs != 0 {
			wantGate = AIG_XVS0
		}
		if p.node(p.Const1()).Xsim != AIG_XVS1 || p.node(a).Xsim != AIG_XVSX || p.node(gate).Xsim != wantGate {
			t.Errorf("regs=%d: source initialization/logic scratch missing: const=%d PI=%d gate=%d", regs, p.node(p.Const1()).Xsim, p.node(a).Xsim, p.node(gate).Xsim)
		}
	}
}

func TestTernaryScratchScopeAgainstCompiledSource(t *testing.T) {
	out := runAigSourceFixture(t, "ternary_scope_compare.c", "aigMem.c")
	var got strings.Builder
	for _, regs := range []int32{0, 1} {
		p := NewGraph()
		a, b := p.NewInput(), p.NewInput()
		po := p.AddOutput(p.AND(a, b))
		Aig_ObjSetXsim(p, po, AIG_XVS1)
		if regs != 0 {
			p.AddOutput(b)
		}
		p.nRegs = regs
		mapping := Aig_ManTernarySimulate(p, 0, 0)
		fmt.Fprintf(&got, "mode %d mapping:", regs)
		for _, e := range mapping {
			fmt.Fprintf(&got, " %d", e)
		}
		got.WriteString("\nvalues:")
		for _, n := range p.nodes {
			fmt.Fprintf(&got, " %d", n.Xsim)
		}
		got.WriteByte('\n')
	}
	if got.String() != out {
		t.Fatalf("Go scratch state:\n%sC scratch state:\n%s", got.String(), out)
	}
}

func TestConstantReductionRepeatsAndPreservesLastManager(t *testing.T) {
	p := NewGraph()
	los := []Edge{p.NewInput(), p.NewInput(), p.NewInput()}
	for _, e := range los {
		p.AddOutput(e)
	}
	for _, e := range los {
		p.AddOutput(e)
	}
	p.nRegs = 3
	var calls []*Graph
	engine := func(g *Graph, symbolic, satur, verbose, veryVerbose int32) []Edge {
		if symbolic != 5 || satur != 7 || verbose != 0 || veryVerbose != 1 {
			t.Fatal("engine options changed")
		}
		calls = append(calls, g)
		if len(calls) == 3 {
			return nil
		}
		mapping := g.Inputs()
		mapping[0] = g.Const0()
		return mapping
	}
	result := Aig_ManConstReduce(p, 1, 5, 7, 0, 1, engine)
	if len(calls) != 3 || calls[2] != result || result.nRegs != 1 {
		t.Fatalf("driver failed repeated reductions: calls=%d registers=%d", len(calls), result.nRegs)
	}
	for _, old := range calls[:2] {
		if old.nodes != nil {
			t.Error("replaced manager was not stopped")
		}
	}
	for _, state := range []bool{false, true} {
		values, err := result.Evaluate([]bool{state})
		if err != nil || len(values) != 4 || values[0] || values[1] || values[2] != state || values[3] != state {
			t.Fatalf("reduced outputs %v, error %v", values, err)
		}
	}
	if err := result.Check(); err != nil {
		t.Fatal(err)
	}
}

func TestConstantReductionWithoutMappingKeepsSourceManager(t *testing.T) {
	p := NewGraph()
	lo := p.NewInput()
	p.AddOutput(lo.Not())
	p.AddOutput(lo.Not())
	p.nRegs = 1
	engine := func(g *Graph, _, _, _, _ int32) []Edge {
		if g != p {
			t.Fatal("wrong manager")
		}
		return nil
	}
	if got := Aig_ManConstReduce(p, 1, 0, 0, 0, 0, engine); got != p || len(p.nodes) == 0 {
		t.Fatal("mapping-free reduction replaced/stopped the source")
	}
}
