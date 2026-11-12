package super

import (
	"bytes"
	"context"
	"fmt"
	abcmain "github.com/glycerine/abcgo/base/main"
	"github.com/glycerine/abcgo/map/mio"
	"github.com/glycerine/abcgo/misc/vec"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"
)

func testLibrary(t *testing.T) *Mio_Library_t {
	t.Helper()
	path, err := filepath.Abs("test_vectors/tiny.genlib")
	if err != nil {
		t.Fatal(err)
	}
	lib := mio.Mio_LibraryRead(path, "", "", 0)
	if lib == nil {
		t.Fatal("cannot read vector genlib")
	}
	return lib
}
func TestGenerateTruthAreaDelayAndBounds(t *testing.T) {
	lib := testLibrary(t)
	for _, inputs := range []int{2, 3, 6} {
		options := Options{Inputs: inputs, Levels: 1, DelayLimit: 1, AreaLimit: 2, SkipInputInverters: true}
		p, err := Generate(lib, options)
		if err != nil {
			t.Fatal(err)
		}
		if len(p.Gates) <= inputs {
			t.Fatal("enumerator generated no physical gates")
		}
		for _, g := range p.Gates {
			if g.Root == nil {
				continue
			}
			if g.Area > 2 || g.MaxDelay > 1 {
				t.Fatal("area/delay bound exceeded")
			}
			expectedArea := float32(mio.Mio_GateReadArea(g.Root))
			for _, child := range g.Fanins {
				expectedArea += child.Area
			}
			if g.Area != expectedArea {
				t.Fatal("tree area mismatch")
			}
			rootTruth := mio.Mio_GateReadTruth(g.Root)
			for mint := 0; mint < 1<<uint(inputs); mint++ {
				assignment := 0
				for pin, child := range g.Fanins {
					if child.Truth&(1<<uint(mint)) != 0 {
						assignment |= 1 << uint(pin)
					}
				}
				want := rootTruth&(1<<uint(assignment)) != 0
				if got := g.Truth&(1<<uint(mint)) != 0; got != want {
					t.Fatalf("incorrect truth %s", Super_WriteLibraryGateName(g))
				}
			}
		}
	}
}
func TestParetoFrontierAndRetainedChildren(t *testing.T) {
	p := Super_ManStart()
	Super_First(p, 2)
	root := mio.Mio_LibraryReadGateByName(testLibrary(t), "and2", "")
	add := func(area float32, delays []float32) *Super_Gate_t {
		truth := []uint32{8, 0}
		if Super_CompareGates(p, truth, area, delays, 2) == 0 {
			return nil
		}
		g := Super_CreateGateNew(p, root, p.Inputs, 2, truth, area, delays, max(delays[0], delays[1]), 2)
		Super_AddGateToTable(p, g)
		return g
	}
	old := add(2, []float32{2, 4})
	if old == nil || add(2, []float32{4, 2}) == nil || add(2, []float32{3, 3}) == nil {
		t.Fatal("Pareto tradeoffs removed")
	}
	if add(3, []float32{5, 5}) != nil {
		t.Fatal("dominated gate accepted")
	}
	inv := mio.Mio_LibraryReadGateByName(testLibrary(t), "inv", "")
	parent := &Super_Gate_t{Root: inv, Variable: -1, Truth: 7, Fanins: []*Super_Gate_t{old}, Area: 3, Delays: []float32{3, 5}, MaxDelay: 5, serial: p.nextSerial}
	p.nextSerial++
	Super_AddGateToTable(p, parent)
	if add(1, []float32{1, 1}) == nil || len(p.table[8]) != 1 {
		t.Fatal("strong gate did not prune old frontier")
	}
	before, err := p.MarshalTree()
	if err != nil {
		t.Fatal(err)
	}
	after, err := p.MarshalTree()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) || parent.Fanins[0] != old || old.Area != 2 {
		t.Fatal("serialization/replacement mutated retained child")
	}
	if _, err = readTreeMetrics(before, p, testLibrary(t)); err != nil {
		t.Fatal(err)
	}
}
func TestLimitsCancellationAndEmptySerialization(t *testing.T) {
	lib := testLibrary(t)
	for _, o := range []Options{{Inputs: 7, Levels: 1}, {Inputs: 2, Levels: -1}, {Inputs: 2, Levels: 1, MaxGates: 1}, {Inputs: 2, Levels: 1, AreaLimit: -1}} {
		if _, err := Generate(lib, o); err == nil {
			t.Fatalf("invalid options accepted %+v", o)
		}
	}
	p, err := Generate(lib, Options{Inputs: 2, Levels: 0})
	if err != nil {
		t.Fatal(err)
	}
	data, err := p.MarshalTree()
	if err != nil {
		t.Fatal(err)
	}
	if _, err = readTreeMetrics(data, p, lib); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err = Generate(lib, Options{Inputs: 2, Levels: 4, Context: ctx}); err != context.Canceled {
		t.Fatalf("cancellation %v", err)
	}
	p, err = Generate(lib, Options{Inputs: 2, Levels: 4, TimeLimit: time.Nanosecond})
	if err != nil || !p.Stats.TimedOut {
		t.Fatalf("timeout %+v %v", p, err)
	}
	p, err = Generate(lib, Options{Inputs: 2, Levels: 4, MaxGates: 3})
	if err != nil || len(p.table) > 3 || !p.Stats.LimitReached {
		t.Fatalf("gate cap %+v %v", p, err)
	}
}
func TestANDEnumerationAndFormulaTruth(t *testing.T) {
	for _, item := range []struct{ inputs, levels, count int }{{2, 0, 3}, {2, 1, 7}, {2, 2, 8}, {3, 1, 16}} {
		p, err := GenerateAND(item.inputs, item.levels)
		if err != nil {
			t.Fatal(err)
		}
		if len(p.Gates) != item.count {
			t.Fatalf("AND inputs=%d levels=%d count=%d", item.inputs, item.levels, len(p.Gates))
		}
		seen := make(map[uint32]bool)
		for _, g := range p.Gates {
			key := p.canonical(g.Truth)
			if seen[key] {
				t.Fatal("duplicate complement class")
			}
			seen[key] = true
			if g.One.Gate != nil {
				a, b := g.One.Gate.Truth, g.Two.Gate.Truth
				if g.One.Inverted {
					a ^= p.mask()
				}
				if g.Two.Inverted {
					b ^= p.mask()
				}
				if g.Truth != a&b {
					t.Fatal("incorrect AND truth")
				}
			}
		}
		var text bytes.Buffer
		if err = p.Write(&text); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(text.String(), "0$") {
			t.Fatal("constant formula missing")
		}
	}
	if _, err := GenerateAND(6, 1); err == nil {
		t.Fatal("unsupported AND width accepted")
	}
	if Super2_LibWriteCompare("(aB)", "(aB)") != 0 {
		t.Fatal("non-strict formula comparator")
	}
}
func TestSuperCommandsAndCompatibilityVector(t *testing.T) {
	path, err := filepath.Abs("test_vectors/tiny.genlib")
	if err != nil {
		t.Fatal(err)
	}
	lib := testLibrary(t)
	value := Super_PrecomputeStr(lib, 2, 1, 0, 0, 0, 0, 1, 0)
	if value == nil {
		t.Fatal("source-compatible vector missing")
	}
	data := vec.Vec_StrArray(value)
	if len(data) == 0 || data[len(data)-1] != 0 {
		t.Fatal("source string terminator missing")
	}
	dir := t.TempDir()
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(dir, "tiny.genlib"), contents, 0600); err != nil {
		t.Fatal(err)
	}
	path = "tiny.genlib"
	t.Chdir(dir)
	frame := abcmain.NewFrame()
	defer frame.Close()
	frame.Out = &bytes.Buffer{}
	frame.Err = &bytes.Buffer{}
	Super_Init(frame)
	if frame.Execute("super -I2 -L1 -o "+path) != 0 {
		t.Fatal(frame.Err)
	}
	for _, name := range []string{"tiny.super", "tiny.super_old"} {
		if _, err = os.Stat(name); err != nil {
			t.Fatal(err)
		}
	}
	if frame.Execute("super2 -I2 -L2") != 0 {
		t.Fatal(frame.Err)
	}
	if _, err = os.Stat("superI2L2"); err != nil {
		t.Fatal(err)
	}
	if frame.Execute("super -D0 "+path) != 1 || frame.Execute("super2 -I6") != 1 {
		t.Fatal("invalid command accepted")
	}
	Super_End(frame)
	if _, exists := frame.LookupCommand("super"); exists {
		t.Fatal("command lifecycle retained registration")
	}
}

// Parse the actual emitted tree grammar, independently recomputing node truth,
// tree area, and pin delays from root cells and previously numbered rows.
func readTreeMetrics(data []byte, p *Super_Man_t, lib *Mio_Library_t) ([]string, error) {
	var lines []string
	for _, line := range strings.Split(strings.TrimRight(string(data), "\x00"), "\n") {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			lines = append(lines, line)
		}
	}
	if len(lines) < 4 {
		return nil, fmt.Errorf("short super header")
	}
	inputs, err := strconv.Atoi(lines[1])
	if err != nil {
		return nil, err
	}
	selected, _ := strconv.Atoi(lines[2])
	total, _ := strconv.Atoi(lines[3])
	if inputs < 1 || inputs > 6 || total != inputs+len(lines)-4 {
		return nil, fmt.Errorf("inconsistent tree line count")
	}
	nodes := make([]*Super_Gate_t, inputs)
	for variable := 0; variable < inputs; variable++ {
		g := &Super_Gate_t{Variable: variable, Delays: make([]float32, inputs)}
		for pin := range g.Delays {
			g.Delays[pin] = SUPER_NO_VAR
		}
		g.Delays[variable] = 0
		for mint := 0; mint < 1<<uint(inputs); mint++ {
			if mint&(1<<uint(variable)) != 0 {
				g.Truth |= 1 << uint(mint)
			}
		}
		nodes[variable] = g
	}
	metrics := make([]string, 0)
	for _, line := range lines[4:] {
		fields := strings.Fields(strings.SplitN(line, "#", 2)[0])
		rootSelected := fields[0] == "*"
		if rootSelected {
			fields = fields[1:]
		}
		root := mio.Mio_LibraryReadGateByName(lib, fields[0], "")
		if root == nil {
			return nil, fmt.Errorf("unknown root %s", fields[0])
		}
		g := &Super_Gate_t{Root: root, Variable: -1, Area: float32(mio.Mio_GateReadArea(root)), Delays: make([]float32, inputs)}
		for pin := range g.Delays {
			g.Delays[pin] = SUPER_NO_VAR
		}
		for pin, word := range fields[1:] {
			id, err := strconv.Atoi(word)
			if err != nil || id < 0 || id >= len(nodes) {
				return nil, fmt.Errorf("invalid tree fanin %q", word)
			}
			child := nodes[id]
			g.Fanins = append(g.Fanins, child)
			g.Area += child.Area
			for input, delay := range child.Delays {
				if delay != SUPER_NO_VAR {
					g.Delays[input] = max(g.Delays[input], delay+mio.Mio_GateReadPinDelay(root, int32(pin)))
				}
			}
		}
		if len(g.Fanins) != int(mio.Mio_GateReadPinNum(root)) {
			return nil, fmt.Errorf("arity mismatch")
		}
		for mint := 0; mint < 1<<uint(inputs); mint++ {
			assignment := 0
			for pin, child := range g.Fanins {
				if child.Truth&(1<<uint(mint)) != 0 {
					assignment |= 1 << uint(pin)
				}
			}
			if mio.Mio_GateReadTruth(root)&(1<<uint(assignment)) != 0 {
				g.Truth |= 1 << uint(mint)
			}
		}
		nodes = append(nodes, g)
		if rootSelected {
			metrics = append(metrics, fmt.Sprintf("%016x %.3f %v", g.Truth, g.Area, g.Delays))
		}
	}
	if len(metrics) != selected {
		return nil, fmt.Errorf("selected gate count mismatch")
	}
	sort.Strings(metrics)
	return metrics, nil
}

func TestFinalDelayLimitAndZeroDelayInputInverterPolicy(t *testing.T) {
	p, err := Generate(testLibrary(t), Options{Inputs: 2, Levels: 3, DelayLimit: 1, SkipInputInverters: true})
	if err != nil {
		t.Fatal(err)
	}
	for _, g := range p.Gates {
		if g.MaxDelay > 1 {
			t.Fatalf("off-by-one delay selection admitted %g", g.MaxDelay)
		}
	}
	text := `GATE zero 0 Y=CONST0;
GATE one 0 Y=CONST1;
GATE buf 1 Y=a; PIN * NONINV 1 999 1 0 1 0
GATE inv 1 Y=!a; PIN * INV 1 999 1 0 1 0
GATE and2 1 Y=a*b; PIN * NONINV 1 999 0 0 0 0
`
	lib := mio.Mio_LibraryRead("zero-delay.genlib", text, "", 0)
	if lib == nil {
		t.Fatal("zero-delay fixture failed")
	}
	p, err = Generate(lib, Options{Inputs: 2, Levels: 2, SkipInputInverters: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(p.table[7]) == 0 {
		t.Fatal("input-only inverter restriction wrongly excluded inversion of a physical zero-delay gate")
	}
}
func TestSerializerRejectsCyclesAndHasNoNumberingSideEffects(t *testing.T) {
	p, err := Generate(testLibrary(t), Options{Inputs: 2, Levels: 1, SkipInputInverters: true})
	if err != nil {
		t.Fatal(err)
	}
	data, err := p.MarshalTree()
	if err != nil {
		t.Fatal(err)
	}
	results := make(chan []byte, 2)
	errors := make(chan error, 2)
	for i := 0; i < 2; i++ {
		go func() { value, e := p.MarshalTree(); results <- value; errors <- e }()
	}
	for i := 0; i < 2; i++ {
		if got := <-results; !bytes.Equal(data, got) {
			t.Fatal("concurrent serialization changed output")
		}
		if e := <-errors; e != nil {
			t.Fatal(e)
		}
	}
	gate := p.selected()[0]
	original := gate.Fanins
	gate.Fanins = append([]*Super_Gate_t(nil), original...)
	gate.Fanins[0] = gate
	if _, err = p.MarshalTree(); err == nil {
		t.Fatal("cyclic gate tree serialized")
	}
	gate.Fanins = original
}
