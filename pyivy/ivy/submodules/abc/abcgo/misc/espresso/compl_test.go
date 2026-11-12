package espresso

import (
	_ "embed"
	"encoding/json"
	"math/rand"
	"testing"
)

func TestComplementAndSimplificationMinterms(t *testing.T) {
	irredGeometry(2, 3, 2)
	rng := rand.New(rand.NewSource(567))
	for iteration := 0; iteration < 250; iteration++ {
		f := new_cover(15)
		for j, n := 0, rng.Intn(13); j < n; j++ {
			sf_addset(f, irredCube(uint32(rng.Intn(3)+1), uint32(rng.Intn(7)+1), uint32(rng.Intn(3)+1)))
		}
		before := sf_save(f)
		off := complement(cube1list(f))
		simple := simplify(cube1list(f))
		var both, bothOff pcover
		simp_comp(cube1list(f), &both, &bothOff)
		irredPoints(func(point []int32) {
			want := irredCoverMatches(f, point)
			if irredCoverMatches(off, point) == want || irredCoverMatches(bothOff, point) == want || irredCoverMatches(simple, point) != want || irredCoverMatches(both, point) != want {
				t.Fatalf("iteration%d point%v complement/simplify mismatch", iteration, point)
			}
		})
		for i := int32(0); i < f.count; i++ {
			if !setp_equal(GETSET(f, i), GETSET(before, i)) {
				t.Fatal("complement mutated borrowed rows")
			}
		}
	}
}

//go:embed test_vectors/compl_oracle.json
var complOracleJSON []byte

func TestComplementOriginalCDifferential(t *testing.T) {
	var cases []struct{ Input, Complement, Simplify, On, Off []uint32 }
	if err := json.Unmarshal(complOracleJSON, &cases); err != nil {
		t.Fatal(err)
	}
	irredGeometry(2, 3, 2)
	for i, test := range cases {
		f := new_cover(int32(len(test.Input)))
		for _, bits := range test.Input {
			p := new_cube()
			p[1] = bits
			sf_addset(f, p)
		}
		off := complement(cube1list(f))
		simple := simplify(cube1list(f))
		var on, bothOff pcover
		simp_comp(cube1list(f), &on, &bothOff)
		for _, pair := range []struct {
			name     string
			got      pcover
			expected []uint32
		}{{"complement", off, test.Complement}, {"simplify", simple, test.Simplify}, {"on", on, test.On}, {"off", bothOff, test.Off}} {
			oracle := new_cover(int32(len(pair.expected)))
			for _, bits := range pair.expected {
				p := new_cube()
				p[1] = bits
				sf_addset(oracle, p)
			}
			irredPoints(func(point []int32) {
				if irredCoverMatches(pair.got, point) != irredCoverMatches(oracle, point) {
					t.Fatalf("case%d %s differs from original C at%v", i, pair.name, point)
				}
			})
		}
	}
}

func TestComplementAllLiftingModes(t *testing.T) {
	irredGeometry(2, 3, 2)
	f := new_cover(5)
	for _, row := range []pcube{irredCube(1, 3, 1), irredCube(2, 6, 2), irredCube(3, 1, 2)} {
		sf_addset(f, row)
	}
	for mode := USE_COMPL_LIFT; mode <= NO_LIFTING; mode++ {
		list := cube1list(f)
		massive_count(list)
		left, right := new_cube(), new_cube()
		variable := binate_split_select(list, left, right, 0)
		l, r := complement(scofactor(list, left, variable)), complement(scofactor(list, right, variable))
		result := compl_merge(list, l, r, left, right, variable, mode)
		irredPoints(func(point []int32) {
			if irredCoverMatches(f, point) == irredCoverMatches(result, point) {
				t.Fatalf("lifting%d incorrect at%v", mode, point)
			}
		})
	}
}
func TestComplementSentinelAndComplexLiftRows(t *testing.T) {
	irredGeometry(2, 2)
	onset := new_cover(1)
	sf_addset(onset, irredCube(1, 1))
	a, b := irredCube(2, 3), irredCube(3, 2)
	SET(a, ACTIVE)
	SET(b, ACTIVE)
	// Both active rows must retain independent lower-bound scratch storage.
	compl_lift_onset_complex([]pset{a, b, nil}, onset, 0)
	if cdist0(a, GETSET(onset, 0)) || cdist0(b, GETSET(onset, 0)) {
		t.Fatal("complex lifting intersected onset")
	}
	compl_d1merge([]pset{nil}, []pset{nil})
}
