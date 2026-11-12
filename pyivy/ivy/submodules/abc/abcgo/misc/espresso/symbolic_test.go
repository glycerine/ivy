package espresso

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
)

func readPLATest(input string, dc, off bool) pPLA {
	cube = cube_struct{}
	kiss = false
	pos = false
	pla := new_PLA()
	if read_pla(strings.NewReader(input), dc, off, FD_type, &pla) != 1 {
		panic("test input unexpectedly empty")
	}
	return pla
}
func TestReadPLAComplementsAndSequentialUnbufferedReader(t *testing.T) {
	cube = cube_struct{}
	kiss = false
	pos = false
	reader := strings.NewReader(".i 1\n.o 1\n0 1\n.e\n1 1\n.end\n")
	var first, second pPLA
	if read_pla(reader, true, true, F_type, &first) != 1 || read_pla(reader, true, true, F_type, &second) != 1 {
		t.Fatal("sequential PLA read failed")
	}
	if pc1(GETSET(first.F, 0)) != "0 1" || pc1(GETSET(first.R, 0)) != "1 1" || pc1(GETSET(second.F, 0)) != "1 1" {
		t.Fatal("input buffering lost a document")
	}
	var eof pPLA
	if read_pla(reader, true, true, F_type, &eof) != -1 {
		t.Fatal("EOF status")
	}
}
func TestReadPLASymbolicInputs(t *testing.T) {
	pla := readPLATest(".i 2\n.o 1\n.symbolic 0 1 ; A B C D ;\n00 1\n11 1\n.e\n", false, true)
	if cube.num_binary_vars != 0 || fmt.Sprint(cube.part_size) != "[4 1]" {
		t.Fatal(cube.part_size)
	}
	if fmt.Sprint(familyMasks(pla.F)) != "[17 24]" {
		t.Fatal(familyMasks(pla.F))
	}
	if fmt.Sprint(pla.label) != "[A B C D ]" {
		t.Fatal(pla.label)
	}
	if set_ord(minterms(pla.F)) != 2 || set_ord(minterms(pla.R)) != 2 {
		t.Fatal("symbolic mapping changed function")
	}
}
func TestReadPLASymbolicOutputs(t *testing.T) {
	pla := readPLATest(".i 1\n.o 2\n.symbolic-output 0 1 ; A B C D ;\n0 01\n1 10\n.e\n", false, true)
	if fmt.Sprint(cube.part_size) != "[2 4]" {
		t.Fatal(cube.part_size)
	}
	if fmt.Sprint(familyMasks(pla.F)) != "[9 18]" {
		t.Fatal(familyMasks(pla.F))
	}
	if fmt.Sprint(pla.label) != "[  A B C D]" {
		t.Fatal(pla.label)
	}
	if set_ord(minterms(pla.F)) != 2 || set_ord(minterms(pla.R)) != 6 {
		t.Fatal("symbolic output complement changed function")
	}
}
func TestReadPLAKissMergeAndPhase(t *testing.T) {
	pla := readPLATest(".kiss\n.mv 4 1 -2 -2 1\n.label var=1 idle run\n0 idle run 1\n.e\n", false, false)
	defer func() { kiss = false }()
	if fmt.Sprint(cube.part_size) != "[2 2 3]" || cube.num_vars != 3 {
		t.Fatal(cube.part_size)
	}
	if pc1(GETSET(pla.F, 0)) != "0 10 011" {
		t.Fatal(pc1(GETSET(pla.F, 0)))
	}
	pla = readPLATest(".i 1\n.o 1\n.phase 0\n0 1\n.e\n", false, false)
	if pc1(GETSET(pla.F, 0)) != "1 1" || pc1(GETSET(pla.R, 0)) != "0 1" {
		t.Fatal("phase did not swap ON/OFF")
	}
}

func TestSymbolicMappingOriginalCFixtures(t *testing.T) {
	data, err := os.ReadFile("test_vectors/symbolic_oracle.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct {
		Input       string
		On, DC, Off []uint32
	}
	if err = json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	for i, c := range cases {
		pla := readPLATest(c.Input, true, true)
		for name, pair := range map[string]struct {
			family pcover
			want   []uint32
		}{"on": {pla.F, c.On}, "dc": {pla.D, c.DC}, "off": {pla.R, c.Off}} {
			got := minterms(pair.family)
			want := minterms(maskFamily(cube.size, pair.want...))
			if !setp_equal(got, want) {
				t.Fatalf("case%d %s minterms got%v want%v", i, name, got, want)
			}
		}
	}
}
