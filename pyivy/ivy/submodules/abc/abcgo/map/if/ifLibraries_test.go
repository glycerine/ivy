package if_

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	abcmain "github.com/glycerine/abcgo/base/main"
)

func readLibraryVector(t *testing.T, name string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("test_vectors", name))
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}
func TestLUTMixedPinTimingAndRoundTrip(t *testing.T) {
	p, err := ParseLUTLibrary("mixed", readLibraryVector(t, "library_mixed.lut"))
	if err != nil {
		t.Fatal(err)
	}
	if p.LutMax != 4 || p.fVarPinDelays != 1 || p.Area(2) != .5 || p.PinDelay(2, 1) != 1.125 || p.PinDelay(4, 3) != 2.25 || If_LibLutFastestPinDelay(p) != 2.25 || If_LibLutSlowestPinDelay(p) != 2.25 {
		t.Fatal("mixed pin normalization", p)
	}
	if If_LibLutDelaysAreDiscrete(p) != 0 || If_LibLutDelaysAreDifferent(p) != 1 {
		t.Fatal("timing predicates")
	}
	p.pLutAreas[2] = 0.123456791
	var encoded bytes.Buffer
	if err = WriteLUTLibrary(&encoded, p); err != nil {
		t.Fatal(err)
	}
	round, err := ParseLUTLibrary("round", encoded.String())
	if err != nil {
		t.Fatal(err)
	}
	if round.pLutAreas != p.pLutAreas || round.pLutDelays != p.pLutDelays || round.fVarPinDelays != p.fVarPinDelays {
		t.Fatal("roundtrip timing")
	}
	copy := If_LibLutDup(p)
	copy.pLutDelays[2][0] = 999
	if p.pLutDelays[2][0] == 999 {
		t.Fatal("array copy aliases")
	}
	if WriteLUTLibrary(libraryErrorWriter{}, p) == nil {
		t.Fatal("lost writer error")
	}
}
func TestLUTBoundaryAndInvalidRows(t *testing.T) {
	var input strings.Builder
	for size := 1; size <= IF_MAX_LUTSIZE; size++ {
		fmt.Fprintf(&input, "%d 1 1\n", size)
	}
	p, err := ParseLUTLibrary("maximum", input.String())
	if err != nil || p.LutMax != IF_MAX_LUTSIZE {
		t.Fatal("maximum LUT rejected", err)
	}
	if _, err = ParseLUTLibrary("overflow", input.String()+"33 1 1\n"); err == nil {
		t.Fatal("accepted oversized LUT")
	}
	bad := []string{"", "2 1 1", "1 1", "1 NaN 1", "1 1 -1", "1 1 1\n2 1 2 1", "1 1 1\n2 1 1 2 3", "1 1 1\n2 1 1\n3 1 1 2", "1 1 1\n1 1 1"}
	for _, text := range bad {
		if _, err = ParseLUTLibrary("bad", text); err == nil {
			t.Fatal("accepted malformed LUT", text)
		}
	}
	// Fractional delay only on a later pin must not be classified discrete.
	p = If_LibLutReadString("1 1 1\n2 1 1 1.5")
	if p == nil || If_LibLutDelaysAreDiscrete(p) != 0 {
		t.Fatal("later fractional pin")
	}
	for size := int32(3); size <= 10; size++ {
		simple := If_LibLutSetSimple(size)
		if simple == nil || simple.LutMax != size || If_LibLutDelaysAreDifferent(simple) != 0 || If_LibLutDelaysAreDiscrete(simple) != 1 {
			t.Fatal(size)
		}
	}
	if If_LibLutSetSimple(2) != nil || If_LibLutSetSimple(11) != nil {
		t.Fatal("invalid simple size")
	}
}
func TestLUTFilesAndFrameIsolation(t *testing.T) {
	path := filepath.Join(t.TempDir(), "lut")
	if err := os.WriteFile(path, []byte("1 1 1\n2 1 1\n3 1 1"), 0600); err != nil {
		t.Fatal(err)
	}
	if p, err := ReadLUTLibrary(path); err != nil || p.LutMax != 3 {
		t.Fatal("missing trailing newline", err)
	}
	frame := abcmain.NewFrame()
	if Abc_FrameSetLutLibrary(frame, "1 1 1") != 1 {
		t.Fatal("set library")
	}
	old := frame.Resource("pLibLut")
	if Abc_FrameSetLutLibrary(frame, "bad") != 0 || frame.Resource("pLibLut") != old {
		t.Fatal("failed replacement changed library")
	}
	if Abc_FrameSetLutLibraryTest(frame) != 1 {
		t.Fatal("source built-in example")
	}
	frame.SetResource("pLibLut", nil)
}
func TestBoxClassicSparseRoundTripAndCopies(t *testing.T) {
	lib, err := ParseBoxLibrary("classic", readLibraryVector(t, "boxes.classic"), false)
	if err != nil {
		t.Fatal(err)
	}
	if If_LibBoxNum(lib) != 2 || If_LibBoxReadBox(lib, 0) != nil || If_LibBoxReadBox(lib, -1) != nil || If_LibBoxFindBox(lib, "adder").Delay(1, 0) != 12 || If_LibBoxFindBox(lib, "adder").Delay(0, 2) != BoxNoArc {
		t.Fatal("sparse box layout")
	}
	if If_LibBoxReadBox(lib, 7).fBlack != 1 {
		t.Fatal("box type")
	}
	clone := If_LibBoxDup(lib)
	clone.Boxes()[0].pDelays[0] = 99
	if lib.Boxes()[0].pDelays[0] == 99 {
		t.Fatal("duplicate shares table")
	}
	var encoded bytes.Buffer
	if err = WriteBoxLibrary(&encoded, lib, false); err != nil {
		t.Fatal(err)
	}
	round, err := ParseBoxLibrary("round", encoded.String(), false)
	if err != nil {
		t.Fatal(err)
	}
	assertSameBoxes(t, lib, round)
	// Matrix row terminators belong to the supplied writer, not process stdout.
	if !strings.Contains(encoded.String(), "10 11 - \n12 13 14 \n") {
		t.Fatal(encoded.String())
	}
	count := lib.nBoxes
	if If_LibBoxAdd(lib, If_BoxDup(lib.Boxes()[0])) == nil || lib.nBoxes != count {
		t.Fatal("duplicate ID changed library")
	}
	If_LibBoxFree(clone)
	if lib.nBoxes != 2 || lib.Boxes()[0].pDelays[0] != 10 {
		t.Fatal("freeing copy damaged original")
	}
}
func assertSameBoxes(t *testing.T, a, b *If_LibBox_t) {
	t.Helper()
	if a.nBoxes != b.nBoxes {
		t.Fatal("box count")
	}
	for _, left := range a.Boxes() {
		right := If_LibBoxReadBox(b, left.Id)
		if right == nil || left.pName != right.pName || left.fSeq != right.fSeq || left.fBlack != right.fBlack || left.fOuter != right.fOuter || left.nPis != right.nPis || left.nPos != right.nPos || !reflect.DeepEqual(left.pDelays, right.pDelays) {
			t.Fatal("box roundtrip", left, right)
		}
	}
}
func TestBoxExtendedFlagsAndCompanion(t *testing.T) {
	text := readLibraryVector(t, "boxes.extended")
	lib, err := ParseBoxLibrary("extended", text, true)
	if err != nil {
		t.Fatal(err)
	}
	if state := If_LibBoxReadBox(lib, 7); state.fSeq != 1 || state.fBlack != 1 || state.fOuter != 1 {
		t.Fatal("attributes", state)
	}
	var encoded bytes.Buffer
	if err = If_LibBoxPrint(&encoded, lib); err != nil {
		t.Fatal(err)
	}
	round, err := ParseBoxLibrary("round", encoded.String(), true)
	if err != nil {
		t.Fatal(err)
	}
	assertSameBoxes(t, lib, round)
	if WriteBoxLibrary(io.Discard, lib, false) == nil {
		t.Fatal("silently lost extended attributes")
	}
	base := filepath.Join(t.TempDir(), "network.aig")
	companion := strings.TrimSuffix(base, ".aig") + ".cdl"
	if err = os.WriteFile(companion, []byte(text), 0600); err != nil {
		t.Fatal(err)
	}
	frame := abcmain.Abc_FrameGetGlobalFrame()
	previous := frame.TakeResource("pLibBox")
	defer func() { frame.SetResource("pLibBox", nil); frame.SetResource("pLibBox", previous) }()
	if If_LibBoxLoad(base) != 1 {
		t.Fatal("companion not loaded")
	}
	loaded := frame.Resource("pLibBox")
	if err = os.WriteFile(companion, []byte(".box broken"), 0600); err != nil {
		t.Fatal(err)
	}
	if If_LibBoxLoad(base) != 0 || frame.Resource("pLibBox") != loaded {
		t.Fatal("malformed companion replaced valid library")
	}

	if If_LibBoxLoad(base+".missing") != 0 || frame.Resource("pLibBox") != loaded {
		t.Fatal("bad companion replacement")
	}
	if If_LibBoxWrite(filepath.Join(t.TempDir(), "missing", "x"), lib) == nil || WriteBoxLibrary(libraryErrorWriter{}, lib, true) == nil {
		t.Fatal("lost I/O failure")
	}
}
func TestBoxMalformedAndComments(t *testing.T) {
	classic := []string{"", "x 1 1 2", "x -1 1 1 1 2", "x 1 2 1 1 2", "x 1 1 -1 1", "x 1 1 2 2 1 2 3", "x 1 1 1 1 nope", "x 1 1 1 1 2 x 2 1 1 1 3"}
	for _, text := range classic {
		if _, err := ParseBoxLibrary("bad", text, false); err == nil {
			t.Fatal("accepted malformed classic", text)
		}
	}
	extended := []string{"", "1 2 3", ".box 1 x", ".box 1 x 1 1 bogus\n1", ".box 1 x 1 1 seq comb\n1", ".box 1 x 1 1\n1 2", ".box 1 x 1 2\n1", ".box 1 x 1 1\n-nope", ".box 1 x 1 1\n1\n.box 1 y 1 1\n2"}
	for _, text := range extended {
		if _, err := ParseBoxLibrary("bad", text, true); err == nil {
			t.Fatal("accepted malformed extended", text)
		}
	}
	reader := strings.NewReader("alpha#comment\nbeta # final comment")
	if If_LibBoxGetToken(reader) != "alpha" || If_LibBoxGetToken(reader) != "beta" || If_LibBoxGetToken(reader) != "" {
		t.Fatal("token comments/EOF")
	}
	if If_BoxStart("bad name", 1, 1, 1, 0, 0, 0) != nil || If_BoxStart("x", 1, -1, 1, 0, 0, 0) != nil {
		t.Fatal("invalid box constructor")
	}
}

type libraryErrorWriter struct{}

func (libraryErrorWriter) Write([]byte) (int, error) { return 0, io.ErrClosedPipe }
