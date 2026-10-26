package abc

// Go-only semantic and state-model checks for src/map/cov.

import (
	"errors"
	"math/rand"
	"strings"
	"testing"
)

func covMustCube(t testing.TB, pattern string) CovCube {
	t.Helper()
	cube, err := CovCubeFromPattern(pattern)
	if err != nil {
		t.Fatal(err)
	}
	return cube
}

func covMinterm(variables, assignment int) CovCube {
	values := make([]CovValue, variables)
	for variable := range values {
		values[variable] = CovZero
		if assignment>>variable&1 != 0 {
			values[variable] = CovOne
		}
	}
	return CovCube{values: values}
}

func covTruthCover(t testing.TB, variables int, truth uint64, exclusive bool) CovCover {
	t.Helper()
	cubes := make([]CovCube, 0)
	for assignment := 0; assignment < 1<<variables; assignment++ {
		if truth>>assignment&1 != 0 {
			cubes = append(cubes, covMinterm(variables, assignment))
		}
	}
	cover, err := NewCovCover(variables, exclusive, cubes)
	if err != nil {
		t.Fatal(err)
	}
	return cover
}

func covTruth(t testing.TB, cover CovCover) uint64 {
	t.Helper()
	if cover.variables > 6 {
		t.Fatal("test truth width")
	}
	var truth uint64
	for assignment := 0; assignment < 1<<cover.variables; assignment++ {
		if cover.Evaluate(uint64(assignment)) {
			truth |= uint64(1) << assignment
		}
	}
	return truth
}

func TestCovCubeEncodingAlgebraAndOwnership(t *testing.T) {
	cube := covMustCube(t, "0-1")
	if cube.Pattern() != "0-1" || cube.LiteralCount() != 2 || !cube.Matches(4) || cube.Matches(5) {
		t.Fatalf("cube behavior %+v", cube)
	}
	values := cube.Values()
	values[0] = CovOne
	if cube.Pattern() != "0-1" {
		t.Fatal("Values aliases cube")
	}
	other := covMustCube(t, "-1-")
	product, ok := covCubeProduct(cube, other)
	if !ok || product.Pattern() != "011" {
		t.Fatalf("product %q %v", product.Pattern(), ok)
	}
	if _, ok := covCubeProduct(covMustCube(t, "0"), covMustCube(t, "1")); ok {
		t.Fatal("contradictory product accepted")
	}
	merged, ok := covCubeMerge(covMustCube(t, "00-"), covMustCube(t, "01-"))
	if !ok || merged.Pattern() != "0--" {
		t.Fatalf("merge %q %v", merged.Pattern(), ok)
	}
	if !covCubeContains(covMustCube(t, "---"), cube) || covCubeContains(cube, covMustCube(t, "---")) {
		t.Fatal("containment direction")
	}
	distance, err := covMustCube(t, "00-").DistanceVariables(covMustCube(t, "11-"))
	if err != nil || len(distance) != 2 || distance[0] != 0 || distance[1] != 1 {
		t.Fatalf("distance %v %v", distance, err)
	}
	disjoint, err := covMustCube(t, "0--").Disjoint(covMustCube(t, "1--"))
	if err != nil || !disjoint {
		t.Fatalf("disjoint %v %v", disjoint, err)
	}
}

func TestCovParseFormatAndOwnedSnapshots(t *testing.T) {
	cover, err := ParseCovCover("# comment\n0- 1\n11 1\n", 2, false, CovOptions{})
	if err != nil {
		t.Fatal(err)
	}
	text, err := cover.Format('1')
	if err != nil || text != "0- 1\n11 1\n" {
		t.Fatalf("format %q %v", text, err)
	}
	cubes := cover.Cubes()
	cubes[0].values[0] = CovOne
	if got, _ := cover.Format('1'); got != text {
		t.Fatal("Cubes aliases cover")
	}
	for _, bad := range []string{"0 1", "0x 1", "00", "00 2"} {
		if _, err := ParseCovCover(bad, 2, false, CovOptions{}); err == nil {
			t.Fatalf("accepted %q", bad)
		}
	}
}

func TestCovSOPAndESOPIdentities(t *testing.T) {
	manager, err := NewCovManager(3, CovOptions{})
	if err != nil {
		t.Fatal(err)
	}
	sop, _ := NewCovCover(3, false, []CovCube{
		covMustCube(t, "00-"), covMustCube(t, "01-"), covMustCube(t, "001"),
	})
	minimized, err := manager.MinimizeSOP(sop)
	if err != nil || minimized.CubeCount() != 1 || minimized.cubes[0].Pattern() != "0--" || covTruth(t, minimized) != covTruth(t, sop) {
		t.Fatalf("SOP result %+v %v", minimized, err)
	}
	esop, _ := NewCovCover(3, true, []CovCube{
		covMustCube(t, "00-"), covMustCube(t, "01-"), covMustCube(t, "111"), covMustCube(t, "111"),
	})
	exclusive, err := manager.MinimizeESOP(esop)
	if err != nil || exclusive.CubeCount() != 1 || exclusive.cubes[0].Pattern() != "0--" || covTruth(t, exclusive) != covTruth(t, esop) {
		t.Fatalf("ESOP result %+v %v", exclusive, err)
	}
}

func TestCovComplementProductUnionAndSupport(t *testing.T) {
	manager, _ := NewCovManager(3, CovOptions{})
	first, _ := NewCovCover(3, false, []CovCube{covMustCube(t, "0--"), covMustCube(t, "-10")})
	second, _ := NewCovCover(3, false, []CovCube{covMustCube(t, "--1")})
	complement, err := manager.ComplementSOP(first)
	if err != nil || covTruth(t, complement) != (^covTruth(t, first)&0xff) {
		t.Fatalf("complement %08b %v", covTruth(t, complement), err)
	}
	product, err := manager.Product(first, second, false)
	if err != nil || covTruth(t, product) != covTruth(t, first)&covTruth(t, second) {
		t.Fatalf("product %08b %v", covTruth(t, product), err)
	}
	union, err := manager.Union(first, second, false)
	if err != nil || covTruth(t, union) != covTruth(t, first)|covTruth(t, second) {
		t.Fatalf("union %08b %v", covTruth(t, union), err)
	}
	if got := first.Support(); len(got) != 3 || got[0] != 0 || got[1] != 1 || got[2] != 2 {
		t.Fatalf("support %v", got)
	}
}

func TestCovExhaustiveSmallTruthFunctions(t *testing.T) {
	for variables := 0; variables <= 4; variables++ {
		manager, _ := NewCovManager(variables, CovOptions{MaxOperations: 1 << 22})
		functionCount := 1 << (1 << variables)
		for truth := 0; truth < functionCount; truth++ {
			cover := covTruthCover(t, variables, uint64(truth), false)
			minimized, err := manager.MinimizeSOP(cover)
			if err != nil || covTruth(t, minimized) != uint64(truth) {
				t.Fatalf("n=%d truth=%x got=%x err=%v", variables, truth, covTruth(t, minimized), err)
			}
			complement, err := manager.ComplementSOP(cover)
			mask := uint64(1)<<(1<<variables) - 1
			if err != nil || covTruth(t, complement) != ^uint64(truth)&mask {
				t.Fatalf("complement n=%d truth=%x got=%x err=%v", variables, truth, covTruth(t, complement), err)
			}
		}
	}
}

func TestCovLimitsCancellationAndAtomicInputs(t *testing.T) {
	cover := covTruthCover(t, 3, 0x96, false)
	before, _ := cover.Format('1')
	limited, _ := NewCovManager(3, CovOptions{MaxOperations: 1})
	if _, err := limited.MinimizeSOP(cover); !errors.Is(err, ErrLimit) {
		t.Fatalf("limit %v", err)
	}
	cancelled, _ := NewCovManager(3, CovOptions{Stop: func() bool { return true }})
	if _, err := cancelled.ComplementSOP(cover); !errors.Is(err, ErrCancelled) {
		t.Fatalf("cancel %v", err)
	}
	after, _ := cover.Format('1')
	if before != after {
		t.Fatal("failed operation mutated input")
	}
	if _, err := ParseCovCover("000 1\n111 1", 3, false, CovOptions{MaxCubes: 1}); !errors.Is(err, ErrLimit) {
		t.Fatalf("parse limit %v", err)
	}
}

func TestCovDeterministicIndependentModel(t *testing.T) {
	random := rand.New(rand.NewSource(0x434f56))
	for execution := 0; execution < 5000; execution++ {
		variables := 1 + random.Intn(5)
		manager, _ := NewCovManager(variables, CovOptions{})
		truth := random.Uint64() & (uint64(1)<<(1<<variables) - 1)
		cover := covTruthCover(t, variables, truth, false)
		minimized, err := manager.MinimizeSOP(cover)
		if err != nil || covTruth(t, minimized) != truth {
			t.Fatalf("execution=%d n=%d truth=%x err=%v", execution, variables, truth, err)
		}
		complement, err := manager.ComplementSOP(minimized)
		mask := uint64(1)<<(1<<variables) - 1
		if err != nil || covTruth(t, complement) != ^truth&mask {
			t.Fatalf("complement execution=%d", execution)
		}
	}
}

func FuzzCovCoverModel(f *testing.F) {
	for _, seed := range [][]byte{{1, 0}, {2, 0, 1, 2, 3}, {3, 0xff, 0x55, 0xaa}, {5, 1, 2, 3, 4, 5, 6}} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) == 0 {
			return
		}
		variables := 1 + int(data[0]%5)
		cubes := make([]CovCube, 0, len(data)-1)
		for offset := 1; offset < len(data) && len(cubes) < 24; offset++ {
			values := make([]CovValue, variables)
			value := data[offset]
			for i := range values {
				values[i] = CovValue((value>>uint(2*i))%3 + 1)
			}
			cubes = append(cubes, CovCube{values: values})
		}
		cover, _ := NewCovCover(variables, false, cubes)
		manager, _ := NewCovManager(variables, CovOptions{MaxOperations: 1 << 18, MaxCubes: 1 << 12})
		before := covTruth(t, cover)
		minimized, err := manager.MinimizeSOP(cover)
		if err != nil {
			return
		}
		if got := covTruth(t, minimized); got != before {
			t.Fatalf("minimize %x != %x", got, before)
		}
		complement, err := manager.ComplementSOP(cover)
		if err != nil {
			return
		}
		mask := uint64(1)<<(1<<variables) - 1
		if got := covTruth(t, complement); got != ^before&mask {
			t.Fatalf("complement %x != %x", got, ^before&mask)
		}
		formatted, err := minimized.Format('1')
		if err != nil {
			t.Fatal(err)
		}
		reparsed, err := ParseCovCover(formatted, variables, false, CovOptions{})
		if err != nil || covTruth(t, reparsed) != before {
			t.Fatalf("round trip %v %q", err, strings.TrimSpace(formatted))
		}
	})
}
