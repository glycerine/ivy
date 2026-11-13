package framgr

import (
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
)

// fraCore.c:186-220 simulates the supplied counterexample, reports the
// disproved constant/class candidates and leaves all temporary marks clear.
func TestVerifyCounterExampleReportsDisprovedClasses(t *testing.T) {
	g := aig.NewGraph()
	a, b := g.NewInput(), g.NewInput()
	left := g.AND(a, b)
	right := g.AND(a, b.Not())
	g.AddOutput(left)
	aig.Aig_ManSetPhase(g)
	classes := Fra_ClassesStart(g)
	classes.Constants = []int{left.NodeID()}
	classes.Classes = [][]int{{left.NodeID(), right.NodeID()}}
	p := &Fra_Man_t{PManAig: g, PCla: classes}
	cex := vec.Vec_IntAlloc(2)
	vec.Vec_IntPush(cex, 1)
	vec.Vec_IntPush(cex, 1)
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	defer func() { os.Stdout = old; r.Close() }()
	Fra_FraigVerifyCounterEx(p, cex)
	w.Close()
	os.Stdout = old
	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{fmt.Sprintf("The node %d is not constant under cex!", left.NodeID()), fmt.Sprintf("The nodes %d and %d are not equal under cex!", left.NodeID(), right.NodeID())} {
		if !strings.Contains(string(out), want) {
			t.Errorf("output %q missing %q", out, want)
		}
	}
	for _, edge := range g.LiveEdges() {
		if g.Node(edge).MarkB {
			t.Errorf("node %d retains temporary mark", edge.NodeID())
		}
	}
}

// fMarkB is a one-bit field in C, so assigning 2 records zero rather than
// treating it as a generic nonzero truth value.
func TestVerifyCounterExampleTruncatesInputToMarkBit(t *testing.T) {
	g := aig.NewGraph()
	a, b := g.NewInput(), g.NewInput()
	node := g.AND(a, b)
	classes := Fra_ClassesStart(g)
	classes.Constants = []int{node.NodeID()}
	p := &Fra_Man_t{PManAig: g, PCla: classes}
	cex := vec.Vec_IntAlloc(2)
	vec.Vec_IntPush(cex, 2)
	vec.Vec_IntPush(cex, 1)
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	defer func() { os.Stdout = old; r.Close() }()
	Fra_FraigVerifyCounterEx(p, cex)
	w.Close()
	os.Stdout = old
	output, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	if len(output) != 0 {
		t.Fatalf("unexpected false counterexample diagnostic: %q", output)
	}
}
