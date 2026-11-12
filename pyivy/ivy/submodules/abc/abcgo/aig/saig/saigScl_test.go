package saig

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
)

func TestReportComplementsCountsRepeatedRegisterDriversAndClearsMarks(t *testing.T) {
	p := aig.NewGraph()
	_ = p.NewInput() // primary input
	lo0 := p.NewInput()
	_ = p.NewInput()        // second latch output
	p.AddOutput(p.Const0()) // primary output
	p.AddOutput(lo0)
	p.AddOutput(lo0.Not()) // same regular driver, opposite phase
	aig.Aig_ManSetRegNum(p, 2)

	if got := Saig_ManReportComplements(p); got != 1 {
		t.Fatalf("repeated register drivers = %d, want 1", got)
	}
	if aig.Aig_ObjIsMarkA(p, lo0) {
		t.Fatal("temporary register mark leaked")
	}
}

func TestReportComplementsPreservesUnrelatedMarks(t *testing.T) {
	p := aig.NewGraph()
	input := p.NewInput()
	lo := p.NewInput()
	p.AddOutput(p.Const0())
	p.AddOutput(input)
	aig.Aig_ManSetRegNum(p, 1)
	aig.Aig_ObjSetMarkA(p, lo)

	if got := Saig_ManReportComplements(p); got != 0 {
		t.Fatalf("repeated register drivers = %d, want 0", got)
	}
	if !aig.Aig_ObjIsMarkA(p, lo) {
		t.Fatal("pre-existing unrelated mark was cleared")
	}
}

func TestReportUselessRegistersUsesTerminalReferenceCounts(t *testing.T) {
	p := aig.NewGraph()
	pi := p.NewInput()
	lo0 := p.NewInput()
	_ = p.NewInput()
	p.AddOutput(lo0) // a register output drives the sole ordinary PO
	p.AddOutput(pi)  // a PI drives one latch input
	p.AddOutput(p.Const0())
	aig.Aig_ManSetRegNum(p, 2)

	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	Saig_ManReportUselessRegisters(p)
	_ = w.Close()
	os.Stdout = oldStdout
	output, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	got := string(output)
	if !strings.Contains(got, "Network has 1 (out of 2) registers driving POs.") {
		t.Fatalf("missing register-to-PO report: %q", got)
	}
	if !strings.Contains(got, "Network has 1 (out of 2) registers driven by PIs.") {
		t.Fatalf("missing PI-to-register report: %q", got)
	}
}
