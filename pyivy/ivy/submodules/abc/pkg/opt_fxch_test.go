package abc

// Go-only tests for src/opt/fxch.

import (
	"errors"
	"slices"
	"testing"
)

func fxchEvalCover(cubes []FXCHCube, inputs, output int, assignment uint64) (bool, error) {
	byOutput := make(map[int][]FXCHCube)
	for _, c := range cubes {
		byOutput[c.Output] = append(byOutput[c.Output], c)
	}
	memo, active := make(map[int]bool), make(map[int]bool)
	known := make(map[int]bool)
	var eval func(int) (bool, error)
	eval = func(v int) (bool, error) {
		if v < inputs {
			return assignment&(uint64(1)<<uint(v)) != 0, nil
		}
		if known[v] {
			return memo[v], nil
		}
		if active[v] {
			return false, ErrInvariant
		}
		active[v] = true
		for _, c := range byOutput[v] {
			term := true
			for _, lit := range c.Literals {
				x, err := eval(fxchVar(lit))
				if err != nil {
					return false, err
				}
				if lit&1 != 0 {
					x = !x
				}
				term = term && x
			}
			if term {
				memo[v] = true
				known[v] = true
				active[v] = false
				return true, nil
			}
		}
		memo[v] = false
		known[v] = true
		active[v] = false
		return false, nil
	}
	return eval(output)
}

func fxchEquivalent(t testing.TB, before, after []FXCHCube, inputs, output int) {
	t.Helper()
	for a := uint64(0); a < uint64(1)<<uint(inputs); a++ {
		x, e := fxchEvalCover(before, inputs, output, a)
		y, f := fxchEvalCover(after, inputs, output, a)
		if e != nil || f != nil || x != y {
			t.Fatalf("assignment %b: before=%v after=%v errors=%v/%v", a, x, y, e, f)
		}
	}
}

func TestFXCHCubeGroupingHashAndAlgebra(t *testing.T) {
	if FXCHCountOnes(0xf0f00101) != 10 {
		t.Fatal("population")
	}
	cubes := []FXCHCube{{4, []int{0, 2}}, {5, []int{1}}, {4, []int{0, 2}}}
	grouped, outputs, err := FXCHGroupCubes(cubes)
	if err != nil || len(grouped) != 2 || !slices.Equal(outputs[0], []int{4}) || !slices.Equal(outputs[1], []int{5}) {
		t.Fatal(grouped, outputs, err)
	}
	ungrouped, err := FXCHUngroupCubes(grouped, outputs)
	if err != nil || len(ungrouped) != 2 {
		t.Fatal(ungrouped, err)
	}
	first, second := FXCHSeparateDivisor([]int{4, 7, 8, 11})
	if !slices.Equal(first, []int{2, 4}) || !slices.Equal(second, []int{3, 5}) {
		t.Fatal(first, second)
	}
	if FXCHDivisorNotConstantOne([]int{4, 7}) || !FXCHDivisorNotConstantOne([]int{4, 9}) {
		t.Fatal("constant-one recognition")
	}
	normalized, complemented, err := FXCHNormalizeDivisor([]int{0, 3, 4, 9})
	if err != nil || complemented || !slices.Equal(normalized, []int{0, 3, 4, 9}) {
		t.Fatal(normalized, complemented, err)
	}
	a, b, complement, removed, err := FXCHRemoveDivisorLiterals([]int{2, 4}, []int{2, 6}, []int{4, 7})
	if err != nil || removed != 2 || complement || !slices.Equal(a, []int{4}) || !slices.Equal(b, []int{6}) {
		t.Fatal(a, b, complement, removed, err)
	}
	h, err := NewFXCHSubCubeHash(3)
	if err != nil {
		t.Fatal(err)
	}
	s := FXCHSubCube{ID: 9, Cube: 2, Literal0: 1}
	if err = h.Insert(s); err != nil || h.Entries() != 1 || h.MemoryBytes() <= 0 {
		t.Fatal(h, err)
	}
	if err = h.Remove(s); err != nil || h.Entries() != 0 {
		t.Fatal(h, err)
	}
	if err = h.Remove(s); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	if FXCHMurmur32(0, 0x9747b28c) != 0xa366817d {
		t.Fatalf("murmur %08x", FXCHMurmur32(0, 0x9747b28c))
	}
	a, b, complement, removed, err = FXCHRemoveDivisorLiterals([]int{2, 5}, []int{3, 4}, []int{4, 9})
	if err != nil || removed != 2 || complement || len(a) != 0 || len(b) != 0 {
		t.Fatal("two-literal complement removal", a, b, complement, removed, err)
	}
	a, b, complement, removed, err = FXCHRemoveDivisorLiterals([]int{2, 5}, []int{3, 7}, []int{4, 7, 8, 13})
	if err != nil || removed != 4 || !complement || len(a) != 0 || len(b) != 0 {
		t.Fatal("mux secondary removal", a, b, complement, removed, err)
	}
}

func TestFXCHSingleAndDoubleExtraction(t *testing.T) {
	single := []FXCHCube{{4, []int{0, 2, 4}}, {4, []int{0, 2, 5}}, {4, []int{0, 2, 6}}}
	r, err := FXCHFastExtract(single, FXCHOptions{ObjectIDMax: 5, MaxExtractions: 1})
	if err != nil || r.Extracted != 1 || r.Variables != 6 {
		t.Fatal(r, err)
	}
	fxchEquivalent(t, single, r.Cubes, 4, 4)
	double := []FXCHCube{{3, []int{0, 2}}, {3, []int{0, 4}}}
	r, err = FXCHFastExtract(double, FXCHOptions{ObjectIDMax: 4, MaxExtractions: 1})
	if err != nil || r.Extracted != 1 {
		t.Fatal(r, err)
	}
	fxchEquivalent(t, double, r.Cubes, 3, 3)
	mux := []FXCHCube{{4, []int{0, 2, 6}}, {4, []int{1, 4, 6}}}
	r, err = FXCHFastExtract(mux, FXCHOptions{ObjectIDMax: 5, MaxExtractions: 1})
	if err != nil || r.Extracted != 1 {
		t.Fatal(r, err)
	}
	fxchEquivalent(t, mux, r.Cubes, 4, 4)
}

func TestFXCHSourceLevelAccountingAndZeroHash(t *testing.T) {
	cover := []FXCHCube{
		{Output: 2, Literals: []int{0}},
		{Output: 2, Literals: []int{2}},
		{Output: 3, Literals: []int{4}},
	}
	manager, err := NewFXCHManager(cover, FXCHOptions{ObjectIDMax: 4})
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(manager.levels, []int{0, 0, 2, 2}) {
		t.Fatalf("levels = %v", manager.levels)
	}
	zero := &FXCHSubCubeHash{}
	if err = zero.Insert(FXCHSubCube{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if err = zero.Remove(FXCHSubCube{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
}

func TestFXCHIndependentModelCampaign(t *testing.T) {
	state := uint64(0x243f6a8885a308d3)
	next := func() uint64 { state = state*6364136223846793005 + 1442695040888963407; return state }
	for execution := 0; execution < 5000; execution++ {
		inputs := 2 + int(next()%4)
		output := inputs
		pairA, pairB := int(next()%uint64(inputs)), int(next()%uint64(inputs))
		if pairA == pairB {
			pairB = (pairB + 1) % inputs
		}
		if pairA > pairB {
			pairA, pairB = pairB, pairA
		}
		pairLitA, pairLitB := fxchMake(pairA, next()&1 != 0), fxchMake(pairB, next()&1 != 0)
		count := 2 + int(next()%4)
		cover := make([]FXCHCube, 0, count)
		for i := 0; i < count; i++ {
			lits := []int{pairLitA, pairLitB}
			if lits[0] > lits[1] {
				lits[0], lits[1] = lits[1], lits[0]
			}
			extra := int(next() % uint64(inputs))
			lit := fxchMake(extra, next()&1 != 0)
			if !slices.Contains(lits, lit) && !slices.Contains(lits, fxchNot(lit)) {
				lits = append(lits, lit)
				slices.Sort(lits)
			}
			cover = append(cover, FXCHCube{output, lits})
		}
		r, err := FXCHFastExtract(cover, FXCHOptions{ObjectIDMax: output + 1, MaxExtractions: 2, MaxCubes: 100, MaxLiterals: 1000})
		if err != nil {
			t.Fatalf("execution %d: %v", execution, err)
		}
		fxchEquivalent(t, cover, r.Cubes, inputs, output)
	}
}

func TestFXCHBoundsCancellationAndAtomicFailure(t *testing.T) {
	cover := []FXCHCube{{2, []int{0, 2}}, {2, []int{0, 3}}}
	before := fxchCloneCover(cover)
	if r, err := FXCHFastExtract(cover, FXCHOptions{ObjectIDMax: 3, Stop: func() bool { return true }}); r.Cubes != nil || !errors.Is(err, ErrCancelled) {
		t.Fatal(r, err)
	}
	if r, err := FXCHFastExtract(cover, FXCHOptions{ObjectIDMax: 3, MaxCubes: 2, MaxExtractions: 1}); r.Cubes != nil || !errors.Is(err, ErrLimit) {
		t.Fatal(r, err)
	}
	if !slices.Equal(cover[0].Literals, before[0].Literals) || !slices.Equal(cover[1].Literals, before[1].Literals) {
		t.Fatal("input mutated")
	}
	if _, err := FXCHFastExtract([]FXCHCube{{2, []int{2, 0}}}, FXCHOptions{ObjectIDMax: 3}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err := FXCHFastExtract([]FXCHCube{{2, []int{0, 1}}}, FXCHOptions{ObjectIDMax: 3}); !errors.Is(err, ErrDimension) {
		t.Fatal("opposite literals of one variable were accepted", err)
	}
	if _, err := NewFXCHSubCubeHash(1<<24 + 1); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	h, err := NewFXCHSubCubeHash(1)
	if err != nil {
		t.Fatal(err)
	}
	if err = h.Insert(FXCHSubCube{Cube: 0, Literal0: 1 << 16}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	m, err := NewFXCHManager(cover, FXCHOptions{ObjectIDMax: 3})
	if err != nil {
		t.Fatal(err)
	}
	m.Close()
	if _, err = m.Cover(); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
}

func FuzzFXCHExtraction(f *testing.F) {
	f.Add(uint8(3), uint64(0x12345678))
	f.Add(uint8(5), uint64(0xfeedface))
	f.Fuzz(func(t *testing.T, raw uint8, seed uint64) {
		inputs := 2 + int(raw%4)
		output := inputs
		state := seed
		next := func() uint64 { state = state*2862933555777941757 + 3037000493; return state }
		cover := make([]FXCHCube, 0, 5)
		for i := 0; i < 2+int(next()%4); i++ {
			used := make(map[int]bool)
			var lits []int
			for v := 0; v < inputs; v++ {
				if next()&3 == 0 {
					lit := fxchMake(v, next()&1 != 0)
					if !used[v] {
						used[v] = true
						lits = append(lits, lit)
					}
				}
			}
			slices.Sort(lits)
			cover = append(cover, FXCHCube{output, lits})
		}
		r, err := FXCHFastExtract(cover, FXCHOptions{ObjectIDMax: output + 1, MaxExtractions: 2, MaxCubes: 100, MaxLiterals: 1000})
		if err != nil {
			t.Fatal(err)
		}
		fxchEquivalent(t, cover, r.Cubes, inputs, output)
	})
}
