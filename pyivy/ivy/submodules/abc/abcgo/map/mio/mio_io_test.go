package mio

import (
	"bytes"
	"errors"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	abcmain "github.com/glycerine/abcgo/base/main"
	"github.com/glycerine/abcgo/misc/vec"
)

func TestGenlibFormatRoundTrip(t *testing.T) {
	lib, err := ReadLibrary("fixture", testGenlib, nil)
	if err != nil {
		t.Fatal(err)
	}
	lib.pGateInv.pPins.dLoadInput = .0123456789012345
	var out bytes.Buffer
	if err = Mio_WriteLibrary(&out, lib, 1, 0, 0); err != nil {
		t.Fatal(err)
	}
	round, err := ReadLibrary("round", out.String(), nil)
	if err != nil {
		t.Fatal(err, out.String())
	}
	for _, g := range lib.ppGates0 {
		other := Mio_LibraryReadGateByName(round, g.pName, g.pOutName)
		if other == nil || other.dArea != g.dArea || other.uTruth != g.uTruth || other.nInputs != g.nInputs {
			t.Fatal("roundtrip gate", g.pName)
		}
		for p, q := g.pPins, other.pPins; p != nil; p, q = p.pNext, q.pNext {
			if q == nil || Mio_CheckPins(p, q) == 0 || p.pName != q.pName {
				t.Fatal("roundtrip timing", g.pName)
			}
		}
	}
	ids := make(map[*Mio_Gate_t]int32)
	for _, g := range lib.ppGates0 {
		ids[g] = g.Cell
	}
	out.Reset()
	if err = Mio_WriteLibrary(&out, lib, 0, 0, 1); err != nil {
		t.Fatal(err)
	}
	for g, id := range ids {
		if g.Cell != id {
			t.Fatal("writer changed mapping IDs")
		}
	}
	out.Reset()
	if err = Mio_WriteLibraryVerilog(&out, lib, 0, 0, 0); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "module and2") || !strings.Contains(out.String(), "assign O") {
		t.Fatal(out.String())
	}
	if Mio_WriteLibrary(errorWriter{}, lib, 0, 0, 0) == nil || Mio_WriteLibraryVerilog(errorWriter{}, lib, 0, 0, 0) == nil {
		t.Fatal("lost write error")
	}
}

type errorWriter struct{}

func (errorWriter) Write([]byte) (int, error) { return 0, errors.New("write failure") }
func TestGateProfilesAndSelection(t *testing.T) {
	lib, err := ReadLibrary("fixture", testGenlib, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err = Mio_LibraryReadProfile(strings.NewReader("# profile\nand2 17\nxor2 3\n"), lib); err != nil {
		t.Fatal(err)
	}
	var profile bytes.Buffer
	if err = Mio_LibraryWriteProfile(&profile, lib); err != nil {
		t.Fatal(err)
	}
	if Mio_LibraryHasProfile(lib) == 0 || !strings.Contains(profile.String(), "17") {
		t.Fatal("profiles")
	}
	if err = Mio_LibraryReadProfile(strings.NewReader("and2 99\nmissing 2\n"), lib); err == nil || lib.pGateAnd2.Profile != 17 {
		t.Fatal("profile read not transactional")
	}
	roots := Mio_CollectRoots(lib, 6, 0, 1, nil, 0)
	if len(roots) != 2 {
		t.Fatal("root selection", len(roots))
	}
	cells := Mio_CollectRootsNew2(lib, 6, nil, 0)
	if len(cells) != 6 || cells[0].uTruth != 0 || cells[3].uTruth != 0x5555555555555555 {
		t.Fatal("basic root slots", cells)
	}
	Mio_LibraryShortNames(lib)
	if lib.pName != "lib6" {
		t.Fatal(lib.pName)
	}
	for g := lib.pGates; g != nil; g = g.pNext {
		if Mio_LibraryReadGateByName(lib, g.pName, "") != g || Mio_ParseCheckFormula(g, g.pForm) == 0 {
			t.Fatal("short names", g)
		}
	}
}
func TestFrameLibraryCommands(t *testing.T) {
	frame := abcmain.NewFrame()
	var output, diagnostic bytes.Buffer
	frame.Out = &output
	frame.Err = &diagnostic
	Mio_Init(frame)
	dir := t.TempDir()
	input, export := filepath.Join(dir, "in.genlib"), filepath.Join(dir, "out.genlib")
	if err := os.WriteFile(input, []byte(testGenlib), 0600); err != nil {
		t.Fatal(err)
	}
	if status := frame.Dispatch([]string{"read_genlib", input}); status != 0 {
		t.Fatal(status, diagnostic.String())
	}
	lib := mioFrameLibrary(frame)
	if lib == nil {
		t.Fatal("not installed on supplied frame")
	}
	if status := frame.Dispatch([]string{"write_genlib", export}); status != 0 {
		t.Fatal(status, diagnostic.String())
	}
	if _, err := ReadLibraryFile(export, nil); err != nil {
		t.Fatal(err)
	}
	if status := frame.Dispatch([]string{"read_genlib", "-W", "NaN", input}); status == 0 || mioFrameLibrary(frame) != lib {
		t.Fatal("accepted nonfinite or changed failed command state")
	}
	if status := frame.Dispatch([]string{"print_genlib"}); status != 0 || !strings.Contains(output.String(), "GATE") {
		t.Fatal("print command", diagnostic.String())
	}
	if status := Mio_CommandReadLiberty(frame, 2, []string{"read_liberty", input}); status == 0 {
		t.Fatal("missing adapter reported success")
	}
	Mio_End(frame)
	if mioFrameLibrary(frame) != nil {
		t.Fatal("resource cleanup")
	}
}
func TestPinValidationAndDelayComposition(t *testing.T) {
	lib, err := ReadLibrary("names", "GATE and2 2 O=a*aa; PIN a NONINV 1 1 2 0 3 0 PIN aa NONINV 1 1 4 0 5 0", nil)
	if err != nil {
		t.Fatal(err)
	}
	g := lib.pGates
	if Mio_ParseCheckFormula(g, "a+aa") == 0 || Mio_ParseCheckFormula(g, "a+aaa") != 0 {
		t.Fatal("name coverage")
	}
	if Mio_GateReadPinName(g, -1) != "" || !math.IsInf(float64(Mio_GateReadPinDelay(g, -1)), 1) {
		t.Fatal("negative pin index")
	}
	var result [2]uint32
	Mio_DeriveTruthTable(g, [][2]uint32{{0xaaaaaaaa, 0xaaaaaaaa}, {0xcccccccc, 0xcccccccc}}, 2, 2, result[:])
	if result[0] != 0x88888888 || result[1] != 0x88888888 {
		t.Fatal(result)
	}
	var delay [2]float32
	var maxDelay float32
	Mio_DeriveGateDelays(g, [][]float32{{1, -1}, {-1, 2}}, 2, 2, -1, delay[:], &maxDelay)
	if delay != [2]float32{4, 7} || maxDelay != 7 {
		t.Fatal(delay, maxDelay)
	}
}
func TestNPNPreparationOrbit(t *testing.T) {
	lib, err := ReadLibrary("fixture", testGenlib, nil)
	if err != nil {
		t.Fatal(err)
	}
	orbits := PrepareLibrary(lib)
	if orbits["and2"] != 8 || orbits["xor2"] != 2 {
		t.Fatal(orbits)
	}
	result := vec.Vec_WrdAlloc(0)
	Nf_ManPrepareGate(2, 0x8888888888888888, []int32{0, 1, 0, 1}, []int32{0, 0}, result)
	set := make(map[uint64]bool)
	for _, truth := range vec.Vec_WrdArray(result) {
		set[truth] = true
	}
	if vec.Vec_WrdSize(result) != 16 || len(set) != 8 {
		t.Fatal("scheduled NPN orbit", len(set))
	}
}
func TestMatchingUnavailableIsExplicitAndPreservesOutputs(t *testing.T) {
	lib, err := ReadLibrary("fixture", testGenlib, nil)
	if err != nil {
		t.Fatal(err)
	}
	count := int32(123)
	var truth *Vec_Mem_t
	var matches *Vec_Wec_t
	var cells []Mio_Cell2_t
	if err = Mio_LibraryMatchesFetch(lib, &truth, &matches, &cells, &count, 0, 0, 0); err == nil || count != 123 {
		t.Fatal("missing mapper adapter")
	}
	if err = Mio_LibraryMatches2Start(lib); err == nil {
		t.Fatal("missing GIA adapter")
	}
}

func TestTwinVerilogAndEscapedNames(t *testing.T) {
	text := `GATE pair/name 1 X=a; PIN * NONINV 1 1 1 1 1 1
GATE pair/name 1 Y=!a; PIN * INV 1 1 1 1 1 1`
	lib, err := ReadLibrary("twins", text, nil)
	if err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	if err = Mio_WriteLibraryVerilog(&output, lib, 0, 0, 0); err != nil {
		t.Fatal(err)
	}
	result := output.String()
	if strings.Count(result, "module ") != 1 || !strings.Contains(result, "\\pair/name ") || !strings.Contains(result, "output X, Y;") || strings.Count(result, "assign ") != 2 {
		t.Fatal(result)
	}
}
