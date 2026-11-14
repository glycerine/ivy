package aig

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/misc/parse"
)

func TestExpressionRootMatchesSourceFormatting(t *testing.T) {
	g := NewGraph()
	a, b := g.NewInput(), g.NewInput()
	g.objectData = map[int]any{a.NodeID(): "a", b.NodeID(): "b"}
	var eq bytes.Buffer
	if err := Aig_ObjPrintEqn(&eq, g, g.AND(a, b.Not())); err != nil {
		t.Fatal(err)
	}
	if eq.String() != "a * !b" {
		t.Errorf("Eqn = %q, want source a * !b", eq.String())
	}
	var verilog bytes.Buffer
	if err := Aig_ObjPrintVerilog(&verilog, g, g.XOR(a, b)); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(verilog.String(), " ^ ") || strings.Contains(verilog.String(), " ? ") {
		t.Fatalf("expanded XOR not recognized: %q", verilog.String())
	}
}

func TestPrintingAgainstCompiledSource(t *testing.T) {
	var input, want strings.Builder
	queriesTotal := 0
	for table := 0; table < 256; table++ {
		g := splitThreeInputTruth(uint8(table))
		g.objectData = map[int]any{g.inputs[0].NodeID(): "a", g.inputs[1].NodeID(): "b", g.inputs[2].NodeID(): "c"}
		firstOutput := g.outputs[0].NodeID()
		for id := 1; id < firstOutput; id++ {
			g.AddOutput(Edge(id) << 1)
		}
		var queries strings.Builder
		queryCount := 0
		apply := func(mode int, root Edge) {
			fmt.Fprintf(&queries, "%d %d\n", mode, root)
			queryCount++
			switch mode {
			case 0:
				want.WriteString("EQN:")
				if err := Aig_ObjPrintEqn(&want, g, root); err != nil {
					t.Fatal(err)
				}
				want.WriteString("\nVERILOG:")
				if err := Aig_ObjPrintVerilog(&want, g, root); err != nil {
					t.Fatal(err)
				}
				want.WriteByte('\n')
			case 1:
				want.WriteString(captureAigOutput(t, func() { Aig_ObjPrintVerbose(g, root, 0); fmt.Println() }))
			case 2:
				want.WriteString(captureAigOutput(t, func() { Aig_ObjPrintVerboseCone(g, root, 0) }))
			case 3:
				text := captureAigOutput(t, func() { Aig_ManPrintVerbose(g, 0) })
				if !strings.HasPrefix(text, "PIs:  1 2 3\n") {
					t.Fatalf("manager CI IDs not printed in interface order: %q", text)
				}
				want.WriteString(text)
			}
			fmt.Fprintf(&want, "traversal %d:", g.travID)
			for _, n := range g.nodes {
				fmt.Fprintf(&want, " %d", n.TravID)
			}
			want.WriteByte('\n')
		}
		for id := 0; id < firstOutput; id++ {
			apply(0, Edge(id)<<1)
			apply(0, (Edge(id) << 1).Not())
		}
		for id := range g.nodes {
			apply(1, Edge(id)<<1)
		}
		apply(2, g.outputs[0])
		apply(3, g.Const1())
		var eq bytes.Buffer
		if err := Aig_ObjPrintEqn(&eq, g, g.outputs[0]); err != nil {
			t.Fatal(err)
		}
		parsed, err := parse.ParseEquation(eq.String(), []string{"a", "b", "c"})
		if err != nil {
			t.Fatal(err)
		}
		for mask := 0; mask < 8; mask++ {
			got, err := parsed.Evaluate([]bool{mask&1 != 0, mask&2 != 0, mask&4 != 0})
			if err != nil || got != (table&(1<<mask) != 0) {
				t.Fatalf("table %d mask %d: formula %q = %t (%v)", table, mask, eq.String(), got, err)
			}
		}
		fmt.Fprintf(&input, "%d %d %d\n", firstOutput-4, len(g.outputs), queryCount)
		for id := 4; id < firstOutput; id++ {
			fmt.Fprintf(&input, "%d %d\n", g.nodes[id].Fanins[0], g.nodes[id].Fanins[1])
		}
		for _, out := range g.outputs {
			fmt.Fprintf(&input, "%d ", g.outputDriver(out))
		}
		input.WriteByte('\n')
		input.WriteString(queries.String())
		queriesTotal += queryCount
	}
	got := runAigSourceFixtureInput(t, "printing_compare.c", input.String(), "aigUtil.c", "aigDfs.c")
	normalize := func(text string) string {
		lines := strings.Split(text, "\n")
		for i, line := range lines {
			if strings.HasPrefix(line, "PIs: ") {
				lines[i] = "PIs: <interface>"
			}
		}
		return strings.Join(lines, "\n")
	}
	if normalize(got) != normalize(want.String()) {
		gotLines, wantLines := strings.Split(normalize(got), "\n"), strings.Split(normalize(want.String()), "\n")
		for i := 0; i < len(gotLines) && i < len(wantLines); i++ {
			if gotLines[i] != wantLines[i] {
				t.Fatalf("line %d: C %q, Go %q", i, gotLines[i], wantLines[i])
			}
		}
		t.Fatalf("C trace length %d, Go %d", len(gotLines), len(wantLines))
	}
	t.Logf("compared %d expression/verbose queries", queriesTotal)
}

func TestUnnamedExpressionInputsRemainDistinct(t *testing.T) {
	g := NewGraph()
	a, b := g.NewInput(), g.NewInput()
	var eq bytes.Buffer
	if err := Aig_ObjPrintEqn(&eq, g, g.AND(a, b)); err != nil {
		t.Fatal(err)
	}
	expression, err := parse.ParseEquation(eq.String(), []string{"pi0", "pi1"})
	if err != nil {
		t.Fatal(err)
	}
	for mask := 0; mask < 4; mask++ {
		got, err := expression.Evaluate([]bool{mask&1 != 0, mask&2 != 0})
		if err != nil || got != (mask == 3) {
			t.Fatalf("mask %d: %q = %t (%v), want %t", mask, eq.String(), got, err, mask == 3)
		}
	}
}

func TestVerbosePrintingUsesSourceScope(t *testing.T) {
	g := NewGraph()
	a, b := g.NewInput(), g.NewInput()
	root := g.AND(a, b.Not())
	out := g.AddOutput(root)
	text := captureAigOutput(t, func() { Aig_ObjPrintVerboseCone(g, root, 0) })
	if !strings.HasPrefix(text, "Node 0 : constant 1") {
		t.Fatalf("cone omitted source constant prelude: %q", text)
	}
	text = captureAigOutput(t, func() { Aig_ManPrintVerbose(g, 0) })
	for _, want := range []string{"Node 0 : constant 1", "Node 1 : CI", "Node 2 : CI", "Node 4 : CO"} {
		if !strings.Contains(text, want) {
			t.Errorf("missing %q in manager dump: %q", want, text)
		}
	}
	if g.node(out).TravID != g.travID {
		t.Error("verbose manager did not perform source CO traversal")
	}
}
