package espresso

import (
	"bytes"
	"fmt"
	"math/rand"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

func TestCubeGeometryAgainstPartEnumeration(t *testing.T) {
	savedCube, savedData := cube, cdata
	defer func() { cube, cdata = savedCube, savedData }()
	rng := rand.New(rand.NewSource(27))
	for _, configuration := range []struct {
		binary int32
		multi  []int32
	}{{0, []int32{3, 39}}, {1, []int32{2}}, {17, []int32{3, 39}}, {32, nil}} {
		ConfigureCube(configuration.binary, configuration.multi)
		for repeat := 0; repeat < 200; repeat++ {
			a, b, cof, lower := new_cube(), new_cube(), new_cube(), new_cube()
			for bit := int32(0); bit < cube.size; bit++ {
				if rng.Intn(2) != 0 {
					set_insert(a, bit)
				}
				if rng.Intn(2) != 0 {
					set_insert(b, bit)
				}
				if rng.Intn(2) != 0 {
					set_insert(cof, bit)
				}
				if rng.Intn(8) == 0 {
					set_insert(lower, bit)
				}
			}
			expectedLower := set_save(lower)
			expectedConsensus := new_cube()
			var distance, activeCount int32
			active := int32(-1)
			commonActive := false
			for variable := int32(0); variable < cube.num_vars; variable++ {
				intersects, missingA, missingB, incomplete := false, false, false, false
				for bit := cube.first_part[variable]; bit <= cube.last_part[variable]; bit++ {
					av, bv, cv := is_in_set(a, bit), is_in_set(b, bit), is_in_set(cof, bit)
					intersects = intersects || av && bv
					missingA = missingA || !av && !cv
					missingB = missingB || !bv && !cv
					incomplete = incomplete || !av
				}
				if !intersects {
					distance++
				}
				if incomplete {
					activeCount++
					active = variable
				}
				commonActive = commonActive || missingA && missingB
				for bit := cube.first_part[variable]; bit <= cube.last_part[variable]; bit++ {
					av, bv := is_in_set(a, bit), is_in_set(b, bit)
					if !intersects && av {
						set_insert(expectedLower, bit)
					}
					value := av && bv
					if !intersects {
						value = av || bv
					}
					if value {
						set_insert(expectedConsensus, bit)
					}
				}
			}
			if activeCount != 1 {
				active = -1
			}
			capDistance := distance
			if capDistance > 2 {
				capDistance = 2
			}
			if cdist(a, b) != distance || cdist01(a, b) != capDistance || cdist0(a, b) != (distance == 0) {
				t.Fatal("cube distance", configuration, distance, cdist(a, b))
			}
			force_lower(lower, a, b)
			if !setp_equal(lower, expectedLower) {
				t.Fatal("force lower", configuration)
			}
			result := new_cube()
			SET(result, PRIME)
			consensus(result, a, b)
			if !setp_equal(result, expectedConsensus) || TESTP(result, PRIME) {
				t.Fatal("consensus", configuration)
			}
			alias := set_save(a)
			consensus(alias, alias, b)
			if !setp_equal(alias, expectedConsensus) {
				t.Fatal("aliased consensus")
			}
			if cactive(a) != active || ccommon(a, b, cof) != commonActive {
				t.Fatal("active geometry", configuration, active, cactive(a), commonActive, ccommon(a, b, cof))
			}
			expectedFull := true
			for bit := int32(0); bit < cube.size; bit++ {
				if !is_in_set(a, bit) && !is_in_set(cof, bit) {
					expectedFull = false
				}
			}
			if full_row(a, cof) != expectedFull {
				t.Fatal("full row")
			}
		}
	}
}

func TestCubeComparators(t *testing.T) {
	savedCube, savedData := cube, cdata
	defer func() { cube, cdata = savedCube, savedData }()
	ConfigureCube(2, []int32{2})
	a, b := setOf(cube.size, 0, 2, 4), setOf(cube.size, 1, 3, 5)
	PUTSIZE(a, 3)
	PUTSIZE(b, 4)
	if descend(a, b) != 1 || ascend(a, b) != -1 || desc1(nil, a) != 1 || desc1(a, nil) != -1 || desc1(nil, nil) != 0 {
		t.Fatal("cardinality/null order")
	}
	PUTSIZE(b, 3)
	if lex_order(a, b) != 1 || descend(a, b) != 1 {
		t.Fatal("unsigned lexical order")
	}
	set_copy(cube.temp[0], cube.fullset)
	if d1_order(a, b) != 0 {
		t.Fatal("merge mask should hide all differences")
	}
}

func TestCubeGeometryOriginalCDifferentialFixture(t *testing.T) {
	compiler, err := exec.LookPath("cc")
	if err != nil {
		t.Skip("optional original-C compiler unavailable")
	}
	_, file, _, _ := runtime.Caller(0)
	pkg := filepath.Dir(file)
	root := filepath.Clean(filepath.Join(pkg, "../../.."))
	binary := filepath.Join(t.TempDir(), "setc_oracle")
	command := exec.Command(compiler, "-std=gnu89", "-O0", "-include", filepath.Join(pkg, "testdata/setc_prelude.h"), "-I"+filepath.Join(root, "src/misc/util"), "-I"+filepath.Join(root, "src"), filepath.Join(pkg, "testdata/setc_oracle.c"), filepath.Join(root, "src/misc/espresso/set.c"), filepath.Join(root, "src/misc/espresso/setc.c"), filepath.Join(root, "src/misc/espresso/cubestr.c"), "-o", binary)
	if result, err := command.CombinedOutput(); err != nil {
		t.Fatalf("C cube oracle compile: %v\n%s", err, result)
	}
	expected, err := exec.Command(binary).Output()
	if err != nil {
		t.Fatal(err)
	}
	savedCube, savedData := cube, cdata
	defer func() { cube, cdata = savedCube, savedData }()
	ConfigureCube(17, []int32{3, 39})
	var output bytes.Buffer
	for trial := int32(0); trial < 50; trial++ {
		a, b, cof := new_cube(), new_cube(), new_cube()
		for bit := int32(0); bit < cube.size; bit++ {
			if (bit*7+trial*3)%11 < 5 {
				set_insert(a, bit)
			}
			if (bit*5+trial*7)%13 < 7 {
				set_insert(b, bit)
			}
			if (bit+trial)%5 == 0 {
				set_insert(cof, bit)
			}
		}
		fmt.Fprintf(&output, "%d %d %d %d %d %d\n", cdist(a, b), cdist01(a, b), boolNumber(cdist0(a, b)), cactive(a), boolNumber(ccommon(a, b, cof)), boolNumber(full_row(a, cof)))
		lower, result := new_cube(), new_cube()
		force_lower(lower, a, b)
		consensus(result, a, b)
		set_write(&output, lower)
		set_write(&output, result)
	}
	if !bytes.Equal(output.Bytes(), expected) {
		t.Fatalf("cube geometry differs from original C:\nGo:\n%s\nC:\n%s", output.Bytes(), expected)
	}
}
func boolNumber(b bool) int32 {
	if b {
		return 1
	}
	return 0
}
