package espresso

import (
	_ "embed"
	"encoding/json"
	"math/rand"
	"testing"
)

func sharpDisjoint(t *testing.T, f pcover) {
	t.Helper()
	for i := int32(0); i < f.count; i++ {
		for j := i + 1; j < f.count; j++ {
			if cdist0(GETSET(f, i), GETSET(f, j)) {
				t.Fatalf("rows %d and%d overlap", i, j)
			}
		}
	}
}
func TestSharpCubeDifferencesExhaustive(t *testing.T) {
	irredGeometry(2, 3, 2)
	var cubes []pcube
	for a := uint32(1); a < 4; a++ {
		for b := uint32(1); b < 8; b++ {
			for c := uint32(1); c < 4; c++ {
				cubes = append(cubes, irredCube(a, b, c))
			}
		}
	}
	for i, a := range cubes {
		for j, b := range cubes {
			ordinary, disjoint := sharp(a, b), dsharp(a, b)
			sharpDisjoint(t, disjoint)
			irredPoints(func(point []int32) {
				want := irredMatches(a, point) && !irredMatches(b, point)
				if irredCoverMatches(ordinary, point) != want || irredCoverMatches(disjoint, point) != want {
					t.Fatalf("cube pair%d/%d point%v", i, j, point)
				}
			})
		}
	}
}
func TestSharpCoverProperties(t *testing.T) {
	irredGeometry(2, 3, 2)
	rng := rand.New(rand.NewSource(532))
	for iteration := 0; iteration < 120; iteration++ {
		a, b := new_cover(8), new_cover(8)
		for _, f := range []pcover{a, b} {
			for i, n := 0, rng.Intn(9); i < n; i++ {
				sf_addset(f, irredCube(uint32(rng.Intn(3)+1), uint32(rng.Intn(7)+1), uint32(rng.Intn(3)+1)))
			}
		}
		ordinary, ds, intersection, disjoint := cv_sharp(a, b), cv_dsharp(a, b), cv_intersect(a, b), make_disjoint(a)
		sharpDisjoint(t, disjoint)
		irredPoints(func(point []int32) {
			x, y := irredCoverMatches(a, point), irredCoverMatches(b, point)
			if irredCoverMatches(ordinary, point) != (x && !y) || irredCoverMatches(ds, point) != (x && !y) || irredCoverMatches(intersection, point) != (x && y) || irredCoverMatches(disjoint, point) != x {
				t.Fatalf("iteration%d point%v", iteration, point)
			}
		})
	}
}
func TestIntersectionCrossesContainmentBatch(t *testing.T) {
	irredGeometry(2, 2)
	a, b := new_cover(30), new_cover(30)
	for i := 0; i < 30; i++ {
		sf_addset(a, cube.fullset)
		sf_addset(b, cube.fullset)
	}
	result := cv_intersect(a, b)
	if result.count != 1 || !setp_equal(GETSET(result, 0), cube.fullset) {
		t.Fatal("intersection batch union lost duplicate containment")
	}
	if a.count != 30 || b.count != 30 {
		t.Fatal("intersection consumed input covers")
	}
}

//go:embed test_vectors/sharp_oracle.json
var sharpOracleJSON []byte

func TestSharpOriginalCDifferential(t *testing.T) {
	var cases []struct {
		Input, DC                             []uint32
		Query                                 uint32
		Sharp, Dsharp, Intersection, Disjoint []uint32
		CubeSharp                             []uint32 `json:"cube_sharp"`
		CubeDsharp                            []uint32 `json:"cube_dsharp"`
	}
	if err := json.Unmarshal(sharpOracleJSON, &cases); err != nil {
		t.Fatal(err)
	}
	cube = cube_struct{num_vars: 3, num_binary_vars: 1, part_size: []int32{2, 3, 2}}
	cube_setup()
	for i, test := range cases {
		a, b := essenFamily(test.Input), essenFamily(test.DC)
		query := new_cube()
		query[1] = test.Query
		for _, pair := range []struct {
			name string
			got  pcover
			want []uint32
		}{{"sharp", cv_sharp(a, b), test.Sharp}, {"dsharp", cv_dsharp(a, b), test.Dsharp}, {"intersection", cv_intersect(a, b), test.Intersection}, {"disjoint", make_disjoint(a), test.Disjoint}, {"cube sharp", cb_sharp(query, b), test.CubeSharp}, {"cube dsharp", cb_dsharp(query, b), test.CubeDsharp}} {
			oracle := essenFamily(pair.want)
			irredPoints(func(point []int32) {
				if irredCoverMatches(pair.got, point) != irredCoverMatches(oracle, point) {
					t.Fatalf("case%d %s differs from original C at%v", i, pair.name, point)
				}
			})
		}
	}
}

func TestSharpAfterCachedGeometryShrink(t *testing.T) {
	define_cube_size(100)
	define_cube_size(2)
	a, b := irredCube(3, 3), irredCube(1, 1)
	result := dsharp(a, b)
	sharpDisjoint(t, result)
	irredPoints(func(point []int32) {
		if irredCoverMatches(result, point) != (irredMatches(a, point) && !irredMatches(b, point)) {
			t.Fatal("cached geometry leaked inactive variable masks")
		}
	})
}
