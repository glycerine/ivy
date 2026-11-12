package io

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/base/abc"
)

func writePLATestFile(t *testing.T, contents string) string {
	t.Helper()
	filename := filepath.Join(t.TempDir(), "case.pla")
	if err := os.WriteFile(filename, []byte(contents), 0600); err != nil {
		t.Fatal(err)
	}
	return filename
}

func plaOutput(t *testing.T, network *abc.Network, name string, assignment map[string]bool) bool {
	t.Helper()
	inputs := make(map[*abc.Object]bool)
	for _, pi := range network.ObjectsOfKind(abc.ObjectPI) {
		value, ok := assignment[pi.Name()]
		if !ok {
			t.Fatalf("missing input value for %q", pi.Name())
		}
		inputs[pi] = value
	}
	for _, po := range network.ObjectsOfKind(abc.ObjectPO) {
		if po.Name() != name {
			continue
		}
		value, err := network.Evaluate(po.Signal(), inputs)
		if err != nil {
			t.Fatal(err)
		}
		return value
	}
	t.Fatalf("output %q not found", name)
	return false
}

func TestReadPLAOutputModesAndPreprocessing(t *testing.T) {
	filename := writePLATestFile(t, `.model mode_test
.i 2
.o 2
.ilb a b
.ob f g
.p 3
00 10
01 11
1- 0-
.e
`)
	defaultNet := Io_ReadPla(filename, 0, 0, 0, 0, 1)
	if defaultNet == nil {
		t.Fatal("default PLA read failed")
	}
	for input, want := range map[int]bool{0: true, 1: true, 2: false, 3: false} {
		values := map[string]bool{"a": input&2 != 0, "b": input&1 != 0}
		if got := plaOutput(t, defaultNet, "f", values); got != want {
			t.Fatalf("default f(%02b)=%v, want %v", input, got, want)
		}
	}
	if !plaOutput(t, defaultNet, "g", map[string]bool{"a": false, "b": true}) {
		t.Fatal("default mode lost a positive output cube")
	}

	onDC := Io_ReadPla(filename, 0, 0, 1, 0, 1)
	if onDC == nil || !plaOutput(t, onDC, "g", map[string]bool{"a": true, "b": false}) || !plaOutput(t, onDC, "g", map[string]bool{"a": true, "b": true}) {
		t.Fatal("fOnDc did not include output don't-care rows")
	}
	zeros := Io_ReadPla(filename, 1, 0, 0, 0, 1)
	if zeros == nil || !plaOutput(t, zeros, "f", map[string]bool{"a": true, "b": false}) {
		t.Fatal("fZeros did not select zero-output rows")
	}
	both := Io_ReadPla(filename, 0, 1, 0, 0, 1)
	if both == nil || !plaOutput(t, both, "f", map[string]bool{"a": true, "b": false}) {
		t.Fatal("fBoth did not select zero and one rows")
	}

	mergeFile := writePLATestFile(t, ".i 2\n.o 1\n00 1\n01 1\n.e\n")
	merged := Io_ReadPla(mergeFile, 0, 0, 0, 0, 1)
	unmerged := Io_ReadPla(mergeFile, 0, 0, 0, 1, 1)
	if merged == nil || unmerged == nil {
		t.Fatal("preprocessed or source-preserving PLA read failed")
	}
	if got := abc.Abc_SopGetCubeNum(merged.ObjectsOfKind(abc.ObjectNode)[0].Data.(string)); got != 1 {
		t.Fatalf("preprocessor retained %d cubes, want 1", got)
	}
	if got := abc.Abc_SopGetCubeNum(unmerged.ObjectsOfKind(abc.ObjectNode)[0].Data.(string)); got != 2 {
		t.Fatalf("skipPrepro retained %d cubes, want 2", got)
	}
}

func TestReadPLAReportsMalformedCubes(t *testing.T) {
	for _, contents := range []string{
		".i 2\n.o 1\n0 1\n.e\n",
		".i 2\n.o 1\n0x 1\n.e\n",
		".i 2\n.o 1\n00 10\n.e\n",
	} {
		filename := writePLATestFile(t, contents)
		if network, err := readPLAFile(filename, 0, 0, 0, false, true); err == nil || network != nil {
			t.Fatalf("malformed PLA unexpectedly parsed: %q", contents)
		}
	}
}

func TestMoplaReductionAndWriterRoundTrip(t *testing.T) {
	filename := writePLATestFile(t, `.model multi
.i 2
.o 2
.ilb a b
.ob f g
.p 4
00 10
01 10
10 01
11 01
.e
`)
	network := Mop_ManTest(filename, 1, 0)
	if network == nil {
		t.Fatal("MOPLA multi-output derivation failed")
	}
	if len(network.ObjectsOfKind(abc.ObjectPO)) != 2 {
		t.Fatal("MOPLA derivation lost outputs")
	}
	for assignment := 0; assignment < 4; assignment++ {
		values := map[string]bool{"a": assignment&2 != 0, "b": assignment&1 != 0}
		if got := plaOutput(t, network, "f", values); got != (assignment < 2) {
			t.Fatalf("f(%02b)=%v", assignment, got)
		}
		if got := plaOutput(t, network, "g", values); got != (assignment >= 2) {
			t.Fatalf("g(%02b)=%v", assignment, got)
		}
	}

	output := filepath.Join(t.TempDir(), "roundtrip.mopla")
	if status := Io_WriteMoPla(network, output); status != 1 {
		t.Fatal("MOPLA write failed")
	}
	parsed, err := ReadNetwork(output, IO_FILE_MOPLA, true)
	if err != nil {
		t.Fatal(err)
	}
	for assignment := 0; assignment < 4; assignment++ {
		values := map[string]bool{"a": assignment&2 != 0, "b": assignment&1 != 0}
		for _, outputName := range []string{"f", "g"} {
			if got, want := plaOutput(t, parsed, outputName, values), plaOutput(t, network, outputName, values); got != want {
				t.Fatalf("roundtrip %s(%02b)=%v, want %v", outputName, assignment, got, want)
			}
		}
	}
	contents, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(contents), ".p 4\n") {
		t.Fatalf("MOPLA writer did not emit four on-set rows:\n%s", contents)
	}
}

func TestMoplaContainmentSubtractsOnlyCoveredOutputs(t *testing.T) {
	filename := writePLATestFile(t, `.model containment
.i 2
.o 2
.ilb a b
.ob f g
.p 2
-- 10
00 11
.e
`)
	network := Mop_ManTest(filename, 0, 0)
	if network == nil {
		t.Fatal("MOPLA containment reduction failed")
	}
	for assignment := 0; assignment < 4; assignment++ {
		values := map[string]bool{"a": assignment&2 != 0, "b": assignment&1 != 0}
		if !plaOutput(t, network, "f", values) {
			t.Fatalf("broad f cube was lost at %02b", assignment)
		}
		if got, want := plaOutput(t, network, "g", values), assignment == 0; got != want {
			t.Fatalf("narrow g(%02b)=%v, want %v", assignment, got, want)
		}
	}
}

func TestStandardPLAWriterRoundTrip(t *testing.T) {
	filename := writePLATestFile(t, `.model standard
.i 2
.o 2
.ilb a b
.ob f g
.p 3
00 10
01 11
1- 0-
.e
`)
	network := Io_ReadPla(filename, 0, 0, 0, 0, 1)
	if network == nil {
		t.Fatal("PLA input failed")
	}
	output := filepath.Join(t.TempDir(), "roundtrip.pla")
	if status := Io_WritePla(network, output); status != 1 {
		t.Fatal("PLA write failed")
	}
	parsed, err := ReadNetwork(output, IO_FILE_PLA, true)
	if err != nil {
		t.Fatal(err)
	}
	for assignment := 0; assignment < 4; assignment++ {
		values := map[string]bool{"a": assignment&2 != 0, "b": assignment&1 != 0}
		for _, outputName := range []string{"f", "g"} {
			if got, want := plaOutput(t, parsed, outputName, values), plaOutput(t, network, outputName, values); got != want {
				t.Fatalf("roundtrip %s(%02b)=%v, want %v", outputName, assignment, got, want)
			}
		}
	}
}
