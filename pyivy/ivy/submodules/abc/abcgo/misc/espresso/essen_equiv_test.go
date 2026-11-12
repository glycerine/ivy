package espresso

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

//go:embed test_vectors/essen_oracle.json
var essenOracleJSON []byte

func essenFamily(rows []uint32) pcover {
	f := new_cover(int32(len(rows)))
	for _, bits := range rows {
		p := new_cube()
		p[1] = bits
		sf_addset(f, p)
	}
	return f
}
func essenRows(f pcover) []uint32 {
	rows := make([]uint32, 0, f.count)
	for i := int32(0); i < f.count; i++ {
		rows = append(rows, GETSET(f, i)[1])
	}
	return rows
}
func TestEssentialAndEquivalenceOriginalC(t *testing.T) {
	var cases []struct {
		Input, DC                       []uint32
		Query                           uint32
		Equivalent                      int
		EssentialCubes                  []int `json:"essential_cubes"`
		Consensus, Essential, Remaining []uint32
		DCAfter                         []uint32 `json:"dc_after"`
	}
	if err := json.Unmarshal(essenOracleJSON, &cases); err != nil {
		t.Fatal(err)
	}
	cube = cube_struct{num_vars: 3, num_binary_vars: 1, part_size: []int32{2, 3, 2}}
	cube_setup()
	for index, test := range cases {
		f, d := essenFamily(test.Input), essenFamily(test.DC)
		query := new_cube()
		query[1] = test.Query
		for i := int32(0); i < f.count; i++ {
			p := GETSET(f, i)
			SET(p, RELESSEN)
			if i%3 == 1 {
				SET(p, NONESSEN)
			}
		}
		if check_equiv(f, d) != (test.Equivalent != 0) {
			t.Fatalf("case%d equivalence", index)
		}
		for i, want := range test.EssentialCubes {
			if essen_cube(f, d, GETSET(f, int32(i))) != (want != 0) {
				t.Fatalf("case%d cube%d essential", index, i)
			}
		}
		h := cb_consensus(f, query)
		if !reflect.DeepEqual(essenRows(h), test.Consensus) {
			t.Fatalf("case%d consensus got%v want%v", index, essenRows(h), test.Consensus)
		}
		e := essential(&f, &d)
		for _, pair := range []struct {
			name string
			f    pcover
			want []uint32
		}{{"essential", e, test.Essential}, {"remaining", f, test.Remaining}, {"dc", d, test.DCAfter}} {
			if !reflect.DeepEqual(essenRows(pair.f), pair.want) {
				t.Fatalf("case%d %s got%v want%v", index, pair.name, essenRows(pair.f), pair.want)
			}
		}
	}
}
func TestCheckEquivSemanticAndSharedRows(t *testing.T) {
	irredGeometry(2, 3)
	a := new_cover(2)
	sf_addset(a, irredCube(1, 7))
	sf_addset(a, irredCube(2, 7))
	b := new_cover(1)
	sf_addset(b, cube.fullset)
	if !check_equiv(a, b) || !check_equiv(a, a) {
		t.Fatal("equivalent decomposed/shared covers rejected")
	}
	c := new_cover(1)
	sf_addset(c, irredCube(1, 7))
	if check_equiv(a, c) || check_equiv(c, a) {
		t.Fatal("strict containment treated as equivalence")
	}
	empty := new_cover(0)
	if !check_equiv(empty, empty) || check_equiv(empty, a) {
		t.Fatal("empty cover equivalence")
	}
}

func TestEquivalentOutputDiagnostics(t *testing.T) {
	cube = cube_struct{num_vars: 2, num_binary_vars: 1, part_size: []int32{2, 3}}
	cube_setup()
	pla := new_PLA()
	PLA_labels(pla)
	pla.R = new_cover(2)
	sf_addset(pla.R, irredCube(1, 3))
	sf_addset(pla.R, irredCube(2, 4))
	var out bytes.Buffer
	writeEquivalentOutputs(&out, pla)
	for _, text := range []string{"Outputs 0 and 1", "Outputs 0 and NOT 2", "Outputs 1 and NOT 2"} {
		if !strings.Contains(out.String(), text) {
			t.Fatalf("missing %q in%s", text, out.String())
		}
	}
}
