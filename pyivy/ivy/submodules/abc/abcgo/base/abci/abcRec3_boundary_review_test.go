package abci

import (
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
	ifpkg "github.com/glycerine/abcgo/map/if"
	"github.com/glycerine/abcgo/misc/vec"
)

func recordedReviewCapture(t *testing.T, run func()) string {
	t.Helper()
	file, err := os.CreateTemp(t.TempDir(), "recorded-report")
	if err != nil {
		t.Fatal(err)
	}
	old := os.Stdout
	os.Stdout = file
	defer func() { os.Stdout = old; file.Close() }()
	run()
	data, err := os.ReadFile(file.Name())
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestRecordedDelayConvertsAfterFloatAddition(t *testing.T) {
	p := ifpkg.If_ManStart(&ifpkg.If_Par_t{NLutSize: 6, NCutsMax: 4})
	defer ifpkg.If_ManStop(p)
	leaf := ifpkg.If_ManCreateCi(p)
	cut := ifpkg.If_ObjCutBest(leaf)
	ifpkg.If_ManSetupCutTriv(p, cut, ifpkg.If_ObjId(leaf))
	for _, tc := range []struct {
		arrival   float32
		pin, want int32
	}{
		{-0.25, 1, 0}, {-4.5, 5, 0}, {16777216, 1, 16777216},
		{2.75, 3, 5}, {0, 15, 15},
	} {
		cut.Delay = tc.arrival
		profile := Lms_DelayInit(0) * word(tc.pin)
		if got := If_CutComputeDelay(p, cut, []byte{0}, profile); got != tc.want {
			t.Errorf("arrival%v pin%d: got%d want%d", tc.arrival, tc.pin, got, tc.want)
		}
	}
}

func TestRecordedLibraryTerminalAreaAndDelayProfiles(t *testing.T) {
	p := gia.Gia_ManStart(16)
	defer gia.Gia_ManStop(p)
	a, b := gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)
	root := gia.Gia_ManAppendAnd(p, a, b)
	for _, lit := range []int32{0, 1, a, a ^ 1, root, root ^ 1} {
		gia.Gia_ManAppendCo(p, lit)
	}
	areas := Lms_GiaAreas(p)
	defer vec.Vec_StrFree(areas)
	if got, want := vec.Vec_StrArray(areas), []byte{0, 0, 0, 0, 1, 1}; !slices.Equal(got, want) {
		t.Fatalf("areas%v want%v", got, want)
	}
	delays := Lms_GiaDelays(p)
	defer vec.Vec_WrdFree(delays)
	if got, want := vec.Vec_WrdArray(delays), []uint64{0, 0, 0, 0, 0x11, 0x11}; !slices.Equal(got, want) {
		t.Fatalf("delays%v want%v", got, want)
	}
}

func TestRecordedFunctionOnlyLibraryReport(t *testing.T) {
	p := Lms_ManStart(nil, 6, 8, 1, 0)
	defer Lms_ManStop(p)
	output := recordedReviewCapture(t, func() { Lms_ManPrint(p) })
	want := "Library with 6 vars has 0 classes and 0 AIG subgraphs with 0 AND nodes.\n"
	if !strings.HasPrefix(output, want) {
		t.Fatalf("missing source header: %q", output)
	}
	if strings.Contains(output, "NaN") || strings.Contains(output, "Inf") {
		t.Fatalf("zero-try counters: %q", output)
	}
	if p.pGia != nil {
		t.Fatal("report created a GIA for a function-only library")
	}
	p.nTried, p.nFilterSize = 10, 3
	vec.Vec_MemHashInsert(p.vTtMem, []uint64{0x8888888888888888})
	output = recordedReviewCapture(t, func() { Lms_ManPrint(p) })
	for _, line := range []string{
		"Library with 6 vars has 1 classes and 0 AIG subgraphs with 0 AND nodes.\n",
		"Subgraphs filtered by support size          =          3. ( 30.00 %)\n",
		"Functions added                             =          1. ( 10.00 %)\n",
	} {
		if !strings.Contains(output, line) {
			t.Fatalf("missing source counter%q in %q", line, output)
		}
	}
}
