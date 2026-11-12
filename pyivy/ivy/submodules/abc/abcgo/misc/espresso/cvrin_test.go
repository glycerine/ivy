package espresso

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
)

func parsePLATest(input string) pPLA {
	cube = cube_struct{}
	kiss = false
	pla := new_PLA()
	pla.pla_type = FD_type
	parse_pla(bufio.NewReader(strings.NewReader(input)), pla)
	return pla
}
func TestPLAParserBinaryLabelsPhaseAndTypes(t *testing.T) {
	pla := parsePLATest("# a comment\n.i 2\n.o 2\n.type fdr\n.ilb a b\n.ob yes no\n.phase 10\n00 10\n1- -1\n.e\n")
	if pla.pla_type != FDR_type || pla.F.count != 2 || pla.D.count != 1 || pla.R.count != 1 {
		t.Fatalf("type/counts=%d %d/%d/%d", pla.pla_type, pla.F.count, pla.D.count, pla.R.count)
	}
	if fmt.Sprint(pla.label) != "[a.bar a b.bar b yes no]" {
		t.Fatal(pla.label)
	}
	if pc1(pla.phase) != "-- 10" {
		t.Fatal(pc1(pla.phase))
	}
	var output bytes.Buffer
	fprint_pla(&output, pla, FDR_type)
	copy := parsePLATest(output.String())
	for name, pair := range map[string][2]pcover{"on": {pla.F, copy.F}, "dc": {pla.D, copy.D}, "off": {pla.R, copy.R}} {
		if fmt.Sprint(familyMasks(pair[0])) != fmt.Sprint(familyMasks(pair[1])) {
			t.Fatalf("%s round-trip mismatch %s", name, output.String())
		}
	}
}
func TestPLAParserMultivaluedSymbolicAndPair(t *testing.T) {
	pla := parsePLATest(".mv 4 2 -3 2\n.ilb a b\n.label var=2 idle run done\n.ob ok bad\n.pair 1 (a b)\n.symbolic a b ; zero one two three ;\n.symbolic-output ok bad ; z o t h ;\n0- idle 10\n11 ANY -1\n.end\n")
	if pla.F.count != 2 || pla.D.count != 1 || pla.pair.cnt != 1 || pla.pair.var1[0] != 1 || pla.pair.var2[0] != 2 {
		t.Fatalf("bad parsed covers/pair %+v", pla.pair)
	}
	if pla.symbolic.symbolic_list_length != 2 || pla.symbolic.symbolic_label_length != 4 || pla.symbolic_output.symbolic_list.pos != 0 {
		t.Fatal("bad symbolic mapping")
	}
	if pc1(GETSET(pla.F, 0)) != "0- 100 10" || pc1(GETSET(pla.F, 1)) != "11 111 01" {
		t.Fatal(pc1(GETSET(pla.F, 0)), pc1(GETSET(pla.F, 1)))
	}
}
func TestPLAParserSequentialDocuments(t *testing.T) {
	cube = cube_struct{}
	reader := bufio.NewReader(strings.NewReader(".i 1\n.o 1\n0 1\n.e\n1 1\n.end\n"))
	first, second := new_PLA(), new_PLA()
	first.pla_type = F_type
	second.pla_type = F_type
	parse_pla(reader, first)
	parse_pla(reader, second)
	if pc1(GETSET(first.F, 0)) != "0 1" || pc1(GETSET(second.F, 0)) != "1 1" {
		t.Fatal("sequential reader lost data")
	}
}
func TestPLAParserRejectsInvalidDirectives(t *testing.T) {
	for _, input := range []string{".o 2\n", ".i -1\n", ".mv 1 2\n", ".i 1\n.o 1\n.phase x\n", ".i 1\n.o 1\n.label var=9 x\n", ".i 2\n.o 1\n.symbolic a b\n"} {
		t.Run(input, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("malformed input accepted")
				}
			}()
			parsePLATest(input)
		})
	}
}
func TestCubeFormattingLongDiagnostic(t *testing.T) {
	ConfigureCube(200, []int32{100})
	p := set_save(cube.fullset)
	if got := pc1(p); len(got) != 301 || got != strings.Repeat("-", 200)+" "+strings.Repeat("1", 100) {
		t.Fatalf("long cube truncated length %d", len(got))
	}
}

func TestPLAParserUnknownDirectiveDoesNotSwallowNextLine(t *testing.T) {
	pla := parsePLATest(".i 1\n.o 1\n.unknown\n0 1\n.e\n")
	if pla.F.count != 1 || pc1(GETSET(pla.F, 0)) != "0 1" {
		t.Fatal("unknown directive consumed the following cube")
	}
}

func TestPLAParserKissStateFields(t *testing.T) {
	pla := parsePLATest(".kiss\n.mv 4 1 -2 -2 1\n.label var=1 idle run\n0 idle run 1\n.end\n")
	defer func() { kiss = false }()
	if pla.F.count != 1 || pla.R.count != 1 {
		t.Fatalf("KISS counts %d/%d", pla.F.count, pla.R.count)
	}
	if pc1(GETSET(pla.F, 0)) != "0 10 01 1" || pc1(GETSET(pla.R, 0)) != "0 10 10 0" {
		t.Fatal(pc1(GETSET(pla.F, 0)), pc1(GETSET(pla.R, 0)))
	}
}

func TestPLAParserOriginalCFixtures(t *testing.T) {
	data, err := os.ReadFile("test_vectors/pla_input_oracle.json")
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
		pla := parsePLATest(c.Input)
		for name, pair := range map[string]struct {
			family pcover
			want   []uint32
		}{"on": {pla.F, c.On}, "dc": {pla.D, c.DC}, "off": {pla.R, c.Off}} {
			if pair.family.count != int32(len(pair.want)) {
				t.Fatalf("case %d %s count got %d want %d", i, name, pair.family.count, len(pair.want))
			}
			for j, want := range pair.want {
				row := GETSET(pair.family, int32(j))
				mask := uint32(0)
				for bit := int32(0); bit < cube.size; bit++ {
					if is_in_set(row, bit) {
						mask |= 1 << bit
					}
				}
				if mask != want {
					t.Fatalf("case %d %s row %d got %d want %d", i, name, j, mask, want)
				}
			}
		}
	}
}
