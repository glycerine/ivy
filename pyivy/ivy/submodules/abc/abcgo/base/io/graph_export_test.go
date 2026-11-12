package io

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/glycerine/abcgo/base/abc"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func exportFixture() *abc.Network {
	n := abc.NewNetwork("graph \"test\"")
	a, b := n.NewObject(abc.ObjectPI), n.NewObject(abc.ObjectPI)
	a.SetName("a")
	b.SetName("b")
	gate := n.NewObject(abc.ObjectNode)
	gate.SetName("and")
	n.AddFanin(gate, a.Signal())
	n.AddFanin(gate, b.Signal().Not())
	gate.Data = "10 1\n"
	out := n.NewObject(abc.ObjectPO)
	out.SetName("z")
	n.AddFanin(out, gate.Signal())
	return n
}

type badGraphWriter struct{}

func (badGraphWriter) Write(p []byte) (int, error) { return 0, errors.New("forced writer failure") }
func TestGraphGMLListAndDOT(t *testing.T) {
	n := exportFixture()
	var gml, list, dot bytes.Buffer
	if e := WriteGML(&gml, n); e != nil {
		t.Fatal(e)
	}
	if strings.Count(gml.String(), "    edge [") != 3 || !strings.Contains(gml.String(), `type "triangle"`) {
		t.Fatal("GML connectivity/shape")
	}
	if e := WriteList(&list, n, true); e != nil {
		t.Fatal(e)
	}
	if !strings.Contains(list.String(), "HOST_to_a") || !strings.Contains(list.String(), "z_to_HOST") {
		t.Fatal("host arcs")
	}
	objects := n.Objects()
	objects[1].MarkB = true
	original := objects[1].MarkB
	if e := WriteDOT(&dot, n, nil, []*abc.Object{objects[1]}, false, false, false); e != nil {
		t.Fatal(e)
	}
	if !strings.Contains(dot.String(), "style = dotted") || !strings.Contains(dot.String(), `Benchmark \"graph`) {
		t.Fatalf("DOT labels/arcs %s", dot.String())
	}
	if objects[1].MarkB != original {
		t.Fatal("DOT changed source marks")
	}
	for _, write := range []func() error{func() error { return WriteGML(badGraphWriter{}, n) }, func() error { return WriteList(badGraphWriter{}, n, false) }, func() error { return WriteDOT(badGraphWriter{}, n, nil, nil, false, false, false) }} {
		if write() == nil {
			t.Fatal("writer failure swallowed")
		}
	}
}
func TestDOTSelectionReverseAndLimits(t *testing.T) {
	n := exportFixture()
	o := n.ObjectsOfKind(abc.ObjectNode)[0]
	var b bytes.Buffer
	if e := WriteDOT(&b, n, []*abc.Object{o}, nil, false, true, false); e != nil {
		t.Fatal(e)
	}
	if strings.Contains(b.String(), " [style = solid]") || strings.Contains(b.String(), " [style = dotted]") {
		t.Fatal("selected graph leaked foreign endpoint")
	}
	foreign := abc.NewNetwork("other")
	if WriteDOT(&b, n, foreign.Objects(), nil, false, false, false) == nil {
		t.Fatal("foreign selection")
	}
	if WriteDOT(&b, n, []*abc.Object{}, nil, false, false, false) == nil {
		t.Fatal("empty selection accepted")
	}
}
func TestBookshelfFilesAndDisconnectedPads(t *testing.T) {
	n := exportFixture()
	for i := 0; i < 3; i++ {
		p := n.NewObject(abc.ObjectPI)
		p.SetName(fmt.Sprintf("unused%d", i))
	}
	terms := append(n.ObjectsOfKind(abc.ObjectPI), n.ObjectsOfKind(abc.ObjectPO)...)
	pads := Io_NtkOrderingPads(n, terms)
	seen := map[*abc.Object]bool{}
	for _, p := range pads {
		if seen[p] {
			t.Fatal("duplicate pad")
		}
		seen[p] = true
	}
	if len(pads) != len(terms) {
		t.Fatal("disconnected pads dropped")
	}
	base := filepath.Join(t.TempDir(), "graph")
	if e := Io_WriteBook(n, base); e != nil {
		t.Fatal(e)
	}
	for _, suffix := range []string{".nodes", ".nets", ".scl", ".pl", ".wts", ".aux"} {
		if _, e := os.Stat(base + suffix); e != nil {
			t.Fatal(e)
		}
	}
	nodes, _ := os.ReadFile(base + ".nodes")
	placement, _ := os.ReadFile(base + ".pl")
	if !strings.Contains(string(nodes), "terminal") || strings.Count(string(placement), "/FIXED") != len(terms) {
		t.Fatal("Bookshelf counts/placement")
	}
	nettext, _ := os.ReadFile(base + ".nets")
	if !strings.Contains(string(nettext), "NumPins : 6") {
		t.Fatalf("pin counts %s", nettext)
	}
}
func TestBookshelfZeroAreaAndCellNetRepeatedPins(t *testing.T) {
	n := abc.NewNetwork("empty")
	model, e := makeBook(n)
	if e != nil {
		t.Fatal(e)
	}
	files := bookFiles(model, "empty")
	if strings.Contains(files[".scl"], "NaN") || strings.Contains(files[".pl"], "NaN") {
		t.Fatal("zero area produced invalid layout")
	}
	a := n.NewObject(abc.ObjectPI)
	gate := n.NewObject(abc.ObjectNode)
	gate.Data = "11 1\n"
	n.AddFanin(gate, a.Signal())
	n.AddFanin(gate, a.Signal())
	var out bytes.Buffer
	if e := WriteCellNet(&out, n); e != nil {
		t.Fatal(e)
	}
	line := ""
	for _, l := range strings.Split(out.String(), "\n") {
		if strings.HasPrefix(l, fmt.Sprintf("net %d ", a.ID)) {
			line = l
		}
	}
	if strings.Count(line, fmt.Sprintf("  %d ", gate.ID)) != 2 {
		t.Fatalf("pin multiplicity %q", line)
	}
}
