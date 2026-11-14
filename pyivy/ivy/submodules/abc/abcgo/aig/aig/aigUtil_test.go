package aig

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/misc/util"
)

func TestUtilMuxAndXorRecognition(t *testing.T) {
	p := NewGraph()
	c, a, b := p.NewInput(), p.NewInput(), p.NewInput()
	mux := p.Mux(c, a, b)
	control, yes, no := Aig_ObjRecognizeMux(p, mux)
	for mask := 0; mask < 8; mask++ {
		inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
		values := aigSimulate(p, aigUtilOrder(p), inputs)
		want := aigUtilEdgeValue(no, values)
		if aigUtilEdgeValue(control, values) {
			want = aigUtilEdgeValue(yes, values)
		}
		if want != aigUtilEdgeValue(mux, values) {
			t.Fatal("multiplexer phase", mask, control, yes, no, mux)
		}
	}
	xor := p.XOR(a, b)
	x, y, ok := Aig_ObjRecognizeExor(p, xor)
	if !ok {
		t.Fatal("decomposed xor not recognized")
	}
	for mask := 0; mask < 8; mask++ {
		values := aigSimulate(p, aigUtilOrder(p), []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0})
		if (aigUtilEdgeValue(x, values) != aigUtilEdgeValue(y, values)) != aigUtilEdgeValue(xor, values) {
			t.Fatal("xor phase")
		}
	}
}
func TestUtilReferencesMarksAndTraversalRollover(t *testing.T) {
	p := NewGraph()
	a, b := p.NewInput(), p.NewInput()
	and := p.AND(a, b)
	p.AddOutput(and)
	p.node(a).Refs = 999
	Aig_ManResetRefs(p)
	if err := p.Check(); err != nil {
		t.Fatal(err)
	}
	p.travID = (1 << 30) - 1
	p.node(a).TravID = 1
	p.objectData = map[int]any{a.NodeID(): "a"}
	p.copies = map[int]Edge{a.NodeID(): b}
	Aig_ManIncrementTravId(p)
	if p.travID != 2 || p.node(a).TravID != 0 || p.objectData[a.NodeID()] != "a" || p.copies[a.NodeID()] != b {
		t.Fatal("typed traversal rollover")
	}
	p.node(a).MarkA = true
	p.node(b).MarkB = true
	Aig_ManCleanMarkAB(p)
	if p.node(a).MarkA || p.node(b).MarkB {
		t.Fatal("marks")
	}
	Aig_ManCleanData(p)
	if len(p.objectData) != 0 || len(p.copies) != 0 {
		t.Fatal("scratch cleanup")
	}
}
func TestUtilSupergateAndBufferPhase(t *testing.T) {
	p := NewGraph()
	a, b, c := p.NewInput(), p.NewInput(), p.NewInput()
	root := p.AND(p.AND(a, b), p.AND(a, c))
	frontier := Aig_ObjCollectMulti(p, root)
	if len(frontier) != 3 {
		t.Fatal("unique supergate frontier", frontier)
	}
	for _, e := range []Edge{a, b, c} {
		found := false
		for _, f := range frontier {
			if e == f {
				found = true
			}
		}
		if !found {
			t.Fatal("missing frontier leaf")
		}
	}
	buffer := Aig_ObjCreate(p, &Node{Kind: Buffer, Fanins: [2]Edge{a.Not(), InvalidEdge}})
	if Aig_ObjReal_rec(p, buffer.Not()) != a {
		t.Fatal("buffer complement composition")
	}
}
func TestUtilExpressionsAndDumps(t *testing.T) {
	p := NewGraph()
	p.Name = "demo"
	a, b, c := p.NewInput(), p.NewInput(), p.NewInput()
	p.objectData = map[int]any{a.NodeID(): "a", b.NodeID(): "b", c.NodeID(): "c"}
	shared := p.NewXor(p.NewXor(a, b), p.NewXor(a, c))
	p.AddOutput(shared)
	var expr bytes.Buffer
	if err := Aig_ObjPrintVerilog(&expr, p, shared); err != nil {
		t.Fatal(err)
	}
	if strings.Count(expr.String(), "a") != 2 {
		t.Fatal("XOR parity lost", expr.String())
	}
	var eq bytes.Buffer
	Aig_ObjPrintEqn(&eq, p, p.AND(a, b).Not())
	if !strings.Contains(eq.String(), "!a + !b") {
		t.Fatal(eq.String())
	}
	var blif, verilog bytes.Buffer
	if err := WriteBLIF(&blif, p, nil, nil); err != nil {
		t.Fatal(err)
	}
	if err := WriteVerilog(&verilog, p); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(blif.String(), ".model demo\n") || !strings.Contains(blif.String(), "10 1\n01 1\n") && !strings.Contains(blif.String(), "01 1\n10 1\n") {
		t.Fatal(blif.String())
	}
	if !strings.Contains(verilog.String(), " ^ ") || !strings.Contains(verilog.String(), "endmodule") {
		t.Fatal(verilog.String())
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "a.blif")
	if err := Aig_ManDumpBlif(p, path, []string{"a", "b", "c"}, []string{"out"}); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil || !strings.Contains(string(data), ".inputs a b c\n.outputs out\n") {
		t.Fatal(err, string(data))
	}
	if err := WriteVerilog(failingWriter{}, p); err == nil {
		t.Fatal("writer error lost")
	}
}

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) { return 0, errors.New("write failed") }
func TestUtilSequentialCounterexample(t *testing.T) {
	p := NewGraph()
	input, state := p.NewInput(), p.NewInput()
	p.AddOutput(state)
	p.AddOutput(input)
	Aig_ManSetRegNum(p, 1)
	// Initial register=0, PI(frame0)=1, PI(frame1)=0; PO becomes 1 in frame1.
	cex := &util.Abc_Cex_t{NRegs: 1, NPis: 1, NBits: 3, IFrame: 1, IPo: 0, PData: []uint32{2}}
	Aig_ManCounterExampleValueStart(p, cex)
	if Aig_ManCounterExampleValueLookup(p, Aig_ObjId(state), 0) != 0 || Aig_ManCounterExampleValueLookup(p, Aig_ObjId(state), 1) != 1 || Aig_ManCounterExampleValueLookup(p, Aig_ObjId(input), 1) != 0 {
		t.Fatal("frame simulation")
	}
	Aig_ManCounterExampleValueStop(p)
	if p.cexValues != nil {
		t.Fatal("session not released")
	}
	var verilog bytes.Buffer
	if err := WriteVerilog(&verilog, p); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(verilog.String(), "always @(posedge clock)") || !strings.Contains(verilog.String(), "initial ") {
		t.Fatal(verilog.String())
	}
}
func TestUtilRandomSequenceAndLists(t *testing.T) {
	r := NewAigRandomSource()
	expected := []uint32{r.Uint32(), r.Uint32(), r.Uint32()}
	got := []uint32{Aig_ManRandom(1), Aig_ManRandom(0), Aig_ManRandom(0)}
	if !reflect.DeepEqual(got, expected) {
		t.Fatal("random stream changed", got, expected)
	}
	word := Aig_ManRandom64(1)
	if uint32(word) != expected[0] || uint32(word>>32) != expected[1] {
		t.Fatal("random64 half order")
	}
	if got := Aig_NodeUnionLists([]Edge{2, 6, 10}, []Edge{4, 6}); !reflect.DeepEqual(got, []Edge{2, 4, 6, 10}) {
		t.Fatal(got)
	}
	if got := Aig_NodeIntersectLists([]Edge{2, 6, 10}, []Edge{4, 6}); !reflect.DeepEqual(got, []Edge{6}) {
		t.Fatal(got)
	}
	if Aig_FileNameGenericAppend("/a.b/path\\file.blif", "_x.aig") != "file_x.aig" {
		t.Fatal("portable filename")
	}
}

func TestUtilMuxRecognitionAllPhasesAndOrderings(t *testing.T) {
	for position := 0; position < 3; position++ {
		for phase := 0; phase < 8; phase++ {
			p := NewGraph()
			inputs := []Edge{p.NewInput(), p.NewInput(), p.NewInput()}
			c := inputs[position]
			a := inputs[(position+1)%3]
			b := inputs[(position+2)%3]
			if phase&1 != 0 {
				c = c.Not()
			}
			if phase&2 != 0 {
				a = a.Not()
			}
			if phase&4 != 0 {
				b = b.Not()
			}
			e := p.Mux(c, a, b)
			for _, root := range []Edge{e, e.Not()} {
				control, yes, no := Aig_ObjRecognizeMux(p, root)
				for mask := 0; mask < 8; mask++ {
					v := aigSimulate(p, aigUtilOrder(p), []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0})
					got := aigUtilEdgeValue(no, v)
					if aigUtilEdgeValue(control, v) {
						got = aigUtilEdgeValue(yes, v)
					}
					if got != aigUtilEdgeValue(root, v) {
						t.Fatalf("phase/order recognition: position %d phase %d mask %d", position, phase, mask)
					}
				}
			}
		}
	}
}
func TestUtilRejectsCyclesAndCounterexampleFailures(t *testing.T) {
	panicExpected := func(f func()) {
		t.Helper()
		defer func() {
			if recover() == nil {
				t.Fatal("invalid state accepted")
			}
		}()
		f()
	}
	p := NewGraph()
	a, b := p.NewInput(), p.NewInput()
	e := p.AND(a, b)
	p.node(e).Fanins[0] = e
	panicExpected(func() { Aig_ObjCollectMulti(p, e) })
	q := NewGraph()
	pi, state := q.NewInput(), q.NewInput()
	q.AddOutput(state)
	q.AddOutput(pi)
	Aig_ManSetRegNum(q, 1)
	invalid := &util.Abc_Cex_t{NRegs: 1, NPis: 1, NBits: 2, IFrame: 0, IPo: 0, PData: []uint32{0}}
	panicExpected(func() { Aig_ManCounterExampleValueStart(q, invalid) })
	if q.cexValues != nil {
		t.Fatal("failed session left scratch state")
	}
	invalid.PData[0] = 1
	panicExpected(func() { Aig_ManCounterExampleValueStart(q, invalid) })
	if q.cexValues != nil {
		t.Fatal("nonzero initial state left session")
	}
	panicExpected(func() { Aig_ManCounterExampleValueLookup(q, 0, 0) })
}
func TestUtilPhaseNonTopologicalAndConstraintFlip(t *testing.T) {
	p := NewGraph()
	a, b := p.NewInput(), p.NewInput()
	buffer := Aig_ObjCreate(p, &Node{Kind: Buffer, Fanins: [2]Edge{a, InvalidEdge}})
	later := p.AND(a.Not(), b.Not())
	p.node(buffer).Fanins[0] = later
	Aig_ManSetPhase(p)
	if !p.node(buffer).Phase || !p.node(later).Phase {
		t.Fatal("phase calculation assumed ID ordering")
	}
	q := NewGraph()
	x := q.NewInput()
	out := q.AddOutput(x)
	q.nConstrs = 1
	Aig_ManSetPhase(q)
	before := q.node(out).Phase
	Aig_ManInvertConstraints(q)
	if q.node(out).Phase == before {
		t.Fatal("constraint phase stale")
	}
	for _, value := range []bool{false, true} {
		got, err := q.Evaluate([]bool{value})
		if err != nil || got[0] == value {
			t.Fatal(err, got)
		}
	}
}
func TestUtilCioSnapshotsAndZeroRandomSource(t *testing.T) {
	p := NewGraph()
	input := p.NewInput()
	output := p.AddOutput(input)
	ids := Aig_ManSetCioIds(p)
	if len(ids) != 2 || ids[input] != 0 || ids[output] != 0 {
		t.Fatal(ids)
	}
	Aig_ManCleanCioIds(p)
	if Aig_ObjCioId(p, input) != 0 {
		t.Fatal("derived Cio IDs changed")
	}
	var zero AigRandomSource
	seeded := NewAigRandomSource()
	if zero.Uint32() != seeded.Uint32() {
		t.Fatal("zero RNG source is unusable")
	}
}

func TestUtilRandomGoldenSequence(t *testing.T) {
	r := NewAigRandomSource()
	for _, want := range []uint32{929569942, 3633709067, 1962020170, 3748422435, 3781595286} {
		if got := r.Uint32(); got != want {
			t.Fatalf("ABC multiply-with-carry sequence got %d want %d", got, want)
		}
	}
}

type shortUtilWriter struct{}

func (shortUtilWriter) Write([]byte) (int, error) { return 0, nil }
func TestUtilShortWritesAreErrors(t *testing.T) {
	p := NewGraph()
	p.AddOutput(p.NewInput())
	if !errors.Is(WriteBLIF(shortUtilWriter{}, p, nil, nil), io.ErrShortWrite) || !errors.Is(WriteVerilog(shortUtilWriter{}, p), io.ErrShortWrite) {
		t.Fatal("short writer accepted")
	}
}
func TestUtilBlifExhaustiveBooleanSemantics(t *testing.T) {
	p := NewGraph()
	a, b, c := p.NewInput(), p.NewInput(), p.NewInput()
	p.AddOutput(p.Mux(c, a.Not(), b))
	p.AddOutput(p.NewXor(a, b.Not()))
	p.AddOutput(p.Const0())
	p.AddOutput(p.Const1())
	var dump bytes.Buffer
	if err := WriteBLIF(&dump, p, nil, nil); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(dump.String(), "\n")
	for mask := 0; mask < 8; mask++ {
		inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
		values := map[string]bool{}
		for i, e := range p.inputs {
			values[fmt.Sprintf("n%d", e.NodeID())] = inputs[i]
		}
		for i := 0; i < len(lines); i++ {
			words := strings.Fields(lines[i])
			if len(words) == 0 || words[0] != ".names" {
				continue
			}
			output := words[len(words)-1]
			fanins := words[1 : len(words)-1]
			value := false
			for i+1 < len(lines) && !strings.HasPrefix(lines[i+1], ".") {
				i++
				row := strings.Fields(lines[i])
				if len(row) == 0 {
					continue
				}
				match := true
				if len(fanins) > 0 {
					for j, name := range fanins {
						if row[0][j] != '-' && values[name] != (row[0][j] == '1') {
							match = false
						}
					}
				}
				if match && row[len(row)-1] == "1" {
					value = true
				}
			}
			values[output] = value
		}
		want, err := p.Evaluate(inputs)
		if err != nil {
			t.Fatal(err)
		}
		for i, e := range p.outputs {
			if got := values[fmt.Sprintf("n%d", e.NodeID())]; got != want[i] {
				t.Fatalf("BLIF output %d mask %d differs", i, mask)
			}
		}
	}
}

func TestDumpWritersPreserveGraphScratch(t *testing.T) {
	g := NewGraph()
	a, b := g.NewInput(), g.NewInput()
	root := g.AND(a, b.Not())
	g.AddOutput(root)
	g.AddOutput(root.Not())
	Aig_ManSetRegNum(g, 1)
	for i := range g.nodes {
		g.nodes[i].MarkA = i%2 != 0
		g.nodes[i].MarkB = i%3 != 0
		g.nodes[i].TravID = int32(i + 20)
		g.nodes[i].CioID = int32(i + 40)
	}
	g.objectData = map[int]any{a.NodeID(): "a", root.NodeID(): "saved payload"}
	g.copies = map[int]Edge{b.NodeID(): root.Not()}
	before := g.Nodes()
	for _, writer := range []func(io.Writer) error{
		func(w io.Writer) error { return WriteBLIF(w, g, nil, nil) },
		func(w io.Writer) error { return WriteVerilog(w, g) },
	} {
		if err := writer(io.Discard); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(before, g.nodes) || g.travID != 1 ||
			!reflect.DeepEqual(g.objectData, map[int]any{a.NodeID(): "a", root.NodeID(): "saved payload"}) ||
			!reflect.DeepEqual(g.copies, map[int]Edge{b.NodeID(): root.Not()}) {
			t.Fatal("dumping changed node metadata, payloads or copies")
		}
	}
}

func TestGenericFilenameUsesOnlyBasenameExtension(t *testing.T) {
	for _, tc := range []struct{ input, want string }{
		{"/parent.v/graph", "graph_x.aig"},
		{"C:\\parent.v\\graph.blif", "graph_x.aig"},
		{"/parent\\nested/file.blif", "file_x.aig"},
		{"graph.part.blif", "graph.part_x.aig"},
	} {
		if got := Aig_FileNameGenericAppend(tc.input, "_x.aig"); got != tc.want {
			t.Fatalf("basename %q: got %q, want %q", tc.input, got, tc.want)
		}
	}
}
