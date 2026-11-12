package io

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/bdd/cudd/core"
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

func TestSequentialDOTAndBookshelfLatch(t *testing.T) {
	n := abc.NewNetwork("sequential")
	a := n.NewObject(abc.ObjectPI)
	a.SetName("a")
	bi := n.NewObject(abc.ObjectBI)
	n.AddFanin(bi, a.Signal().Not())
	latch := n.NewObject(abc.ObjectLatch)
	latch.SetName("q")
	n.AddFanin(latch, bi.Signal())
	bo := n.NewObject(abc.ObjectBO)
	n.AddFanin(bo, latch.Signal())
	po := n.NewObject(abc.ObjectPO)
	po.SetName("z")
	n.AddFanin(po, bo.Signal())
	var b bytes.Buffer
	if e := WriteDOT(&b, n, nil, nil, false, false, true); e != nil {
		t.Fatal(e)
	}
	for _, arc := range []string{fmt.Sprintf("Node%d -> Node%d [style = dotted]", latch.ID, a.ID), fmt.Sprintf("Node%d -> Node%d [style = solid]", po.ID, latch.ID)} {
		if !strings.Contains(b.String(), arc) {
			t.Fatalf("missing collapsed sequential edge %s: %s", arc, b.String())
		}
	}
	model, e := makeBook(n)
	if e != nil {
		t.Fatal(e)
	}
	if model.area != 6 || len(model.cells) != 3 {
		t.Fatal("latch cell area/count")
	}
	for _, net := range model.nets {
		for _, pin := range net.pins {
			if pin.cell == nil {
				t.Fatal("terminal collapse lost cell")
			}
		}
	}
}
func TestGraphExportsRegisterAndFailures(t *testing.T) {
	n := exportFixture()
	root := t.TempDir()
	for _, writer := range []func(*abc.Network, string) error{Io_WriteDot, Io_WriteGml, Io_WriteBookLogic} {
		if e := writer(n, filepath.Join(root, "missing", "bad")); e == nil {
			t.Fatal("file error swallowed")
		}
	}
	var b bytes.Buffer
	if WriteGML(&b, nil) == nil || WriteList(&b, nil, false) == nil || WriteDOT(&b, nil, nil, nil, false, false, false) == nil {
		t.Fatal("nil network accepted")
	}
	long := n.NewObject(abc.ObjectPI)
	long.SetName("quote\" & name\nline")
	b.Reset()
	if e := WriteGML(&b, n); e != nil {
		t.Fatal(e)
	}
	if !strings.Contains(b.String(), "&quot;") || !strings.Contains(b.String(), "&amp;") || !strings.Contains(b.String(), "&#10;") {
		t.Fatal("GML label escaping")
	}
	if _, e := makeBook(n); e == nil {
		t.Fatal("Bookshelf whitespace names accepted")
	}
}

func TestGraphBDDLabelsAndBookConversion(t *testing.T) {
	n := exportFixture()
	n.Type = uint8(abc.ABC_NTK_LOGIC)
	n.Function = uint8(abc.ABC_FUNC_BDD)
	dd := core.Cudd_Init(2, 0, 256, 256, 0)
	n.FunctionManager = dd
	node := n.ObjectsOfKind(abc.ObjectNode)[0]
	f := core.Cudd_bddAnd(dd, core.Cudd_bddIthVar(dd, 0), core.Cudd_Not(core.Cudd_bddIthVar(dd, 1)))
	node.Data = f
	var b bytes.Buffer
	if err := WriteDOT(&b, n, nil, nil, false, false, false); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(b.String(), "10 1") {
		t.Fatalf("BDD label %s", b.String())
	}
	if node.Data != f || n.FunctionManager != dd || n.Function != uint8(abc.ABC_FUNC_BDD) {
		t.Fatal("DOT changed BDD representation")
	}
	if err := Io_WriteBookLogic(n, filepath.Join(t.TempDir(), "bdd")); err != nil {
		t.Fatal(err)
	}
	if node.Data != f || n.FunctionManager != dd {
		t.Fatal("Bookshelf changed BDD representation")
	}
}
