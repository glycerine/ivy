package abci

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/base/abc"
)

func captureMeasureOutput(t *testing.T, run func()) string {
	t.Helper()
	read, write, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	previous := os.Stdout
	os.Stdout = write
	run()
	_ = write.Close()
	os.Stdout = previous
	var output bytes.Buffer
	if _, err := io.Copy(&output, read); err != nil {
		t.Fatal(err)
	}
	_ = read.Close()
	return output.String()
}

func TestAbcNtkPrintMeasures(t *testing.T) {
	// Four-variable pattern repetition represents x0 AND x1.
	truth := []uint32{0x88888888}
	output := captureMeasureOutput(t, func() { Abc_NtkPrintMeasures(truth, 2) })
	if !strings.Contains(output, "     16      16") {
		t.Fatalf("expected both individual influence counts to be 16; output:\n%s", output)
	}
	if !strings.HasPrefix(output, "Total = 3.\n") {
		t.Fatalf("expected deterministic total pair count of 3; output:\n%s", output)
	}
	if !strings.Contains(output, "        \n     32 \n     32         \n") {
		t.Fatalf("expected the pair measure matrix to contain 32 in both off-diagonal cells; output:\n%s", output)
	}
}

func TestAbcNtk4VarObj(t *testing.T) {
	network := abc.NewNetwork("measure-test")
	a, b := abc.Abc_NtkCreatePi(network), abc.Abc_NtkCreatePi(network)
	a.CopyData, b.CopyData = uint32(0xAAAAAAAA), uint32(0xCCCCCCCC)
	node := abc.Abc_NtkCreateNodeAnd(network, []*abc.Object{a, b})
	if got := Abc_Ntk4VarObj([]*abc.Object{node}); got != 0x88888888 {
		t.Fatalf("AND truth table = %#08x, want %#08x", got, uint32(0x88888888))
	}
	if node.CopyData != uint32(0x88888888) {
		t.Fatalf("node copy scratch = %#v, want truth table", node.CopyData)
	}
	if node.Data != nil {
		t.Fatalf("truth evaluation clobbered the network's Data field: %#v", node.Data)
	}
}

func TestAbcNtkPrintOneDecompChecks(t *testing.T) {
	cofactors := []uint32{0, 1, 0, 1, 2, 3, 2, 3}
	if got := Abc_NtkPrintOneDecompCheckCofList(cofactors[:4], 4); got != 1 {
		t.Fatalf("two unique cofactors: got %d, want 1", got)
	}
	if got := Abc_NtkPrintOneDecompCheck(cofactors, 8, 1); got != 1 {
		t.Fatalf("mask 1 should leave at most two unique cofactors per group: got %d", got)
	}
	if got := Abc_NtkPrintOneDecompCheck(cofactors, 8, 0); got != 0 {
		t.Fatalf("mask 0 should reject three unique cofactors: got %d", got)
	}
}
