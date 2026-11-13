package live

import (
	"github.com/glycerine/abcgo/aig/aig"
	"io"
	"os"
	"strings"
	"testing"
)

func TestSourceUnfinishedHelpersReturnEmpty(t *testing.T) {
	reader := strings.NewReader("fairness: x\nmore\n")
	got := getVecOfVecFairness(reader)
	if got == nil || len(got) != 0 || reader.Len() != 0 {
		t.Fatalf("source reader got %v with %d bytes left", got, reader.Len())
	}
	if checkMonotoneSignal() != nil {
		t.Fatal("source placeholder returns NULL")
	}
	g := aig.NewGraph()
	a, b := g.NewInput(), g.NewInput()
	g.AND(a, b)
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	old := os.Stdout
	os.Stdout = w
	result := gatherMonotoneSignals(g)
	os.Stdout = old
	_ = w.Close()
	output, err := io.ReadAll(r)
	_ = r.Close()
	if err != nil {
		t.Fatal(err)
	}
	if result != nil {
		t.Fatal("source gather returns NULL")
	}
	if string(output) != "Node    3 : AND(    1 ,    2  ) (refs =   0)\n" {
		t.Fatalf("source print %q", output)
	}
}

func TestSourceGatherPrintsChoiceMetadata(t *testing.T) {
	g := aig.NewGraph()
	a, b := g.NewInput(), g.NewInput()
	head, other := g.AND(a, b), g.AND(a, b.Not())
	aig.Aig_ObjSetEquiv(g, head, other)
	aig.Aig_ObjSetRepr(g, other, head)
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	old := os.Stdout
	os.Stdout = w
	gatherMonotoneSignals(g)
	os.Stdout = old
	_ = w.Close()
	output, err := io.ReadAll(r)
	_ = r.Close()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(output), "  {    3     4  }") || !strings.Contains(string(output), "  class of 4") {
		t.Fatalf("choice diagnostic %q", output)
	}
}
