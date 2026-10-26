package abc

// Stand-alone models for src/bool/kit/kitSop.c and kitPla.c.

import (
	"errors"
	"math/rand"
	"slices"
	"testing"
)

func modelLiteralCover(cubes []uint32, assignment uint32) bool {
	for _, cube := range cubes {
		if assignment&cube == cube {
			return true
		}
	}
	return false
}

func modelCoverProduct(left, right []uint32) []uint32 {
	result := make([]uint32, 0, len(left)*len(right))
	for _, a := range left {
		for _, b := range right {
			result = append(result, a|b)
		}
	}
	return result
}

func TestKitSOPConstructionAndLiterals(t *testing.T) {
	input := []uint32{1<<0 | 1<<3, 1<<0 | 1<<2, 1<<1 | 1<<2}
	cover, err := NewKitSOPCover(input, 4, KitCoverOptions{})
	if err != nil || !slices.Equal(cover.Cubes, input) {
		t.Fatalf("create: %#v %v", cover, err)
	}
	input[0] = 0
	if cover.Cubes[0] == 0 {
		t.Fatal("cover aliases caller storage")
	}
	inverse, err := NewInverseKitSOPCover([]uint32{1 | 2<<2, 2 | 1<<2}, 4, KitCoverOptions{})
	if err != nil || !slices.Equal(inverse.Cubes, []uint32{2 | 1<<2, 1 | 2<<2}) {
		t.Fatalf("inverse: %#v %v", inverse, err)
	}
	if got, _ := cover.AnyRepeatedLiteral(4); got != 0 {
		t.Fatalf("repeated literal = %d", got)
	}
	if got, _ := cover.WorstLiteral(4); got != 0 {
		t.Fatalf("worst literal = %d", got)
	}
	if got, _ := cover.BestLiteral(4, 0xf); got != 0 {
		t.Fatalf("best literal tie = %d", got)
	}
	best, err := cover.BestLiteralCover(4, 0xf)
	if err != nil || !slices.Equal(best.Cubes, []uint32{1}) {
		t.Fatalf("best literal cover: %#v %v", best, err)
	}
	common, err := cover.CommonCubeCover()
	if err != nil || common.Cubes[0] != 0 || !cover.IsCubeFree() {
		t.Fatalf("common/free: %#v %v", common, err)
	}
	withCommon, _ := NewKitSOPCover([]uint32{3, 5}, 4, KitCoverOptions{})
	if withCommon.IsCubeFree() || withCommon.CommonCube() != 1 || !slices.Equal(withCommon.MakeCubeFree().Cubes, []uint32{2, 4}) {
		t.Fatal("common cube removal")
	}
	clone := cover.Clone()
	clone.Cubes[0] = 0
	if cover.Cubes[0] == 0 {
		t.Fatal("clone aliases")
	}
}

func TestKitSOPDivisionReconstruction(t *testing.T) {
	cover, _ := NewKitSOPCover([]uint32{1 | 4, 2 | 4, 1 | 8, 2 | 8, 16}, 5, KitCoverOptions{})
	divisor, _ := NewKitSOPCover([]uint32{1, 2}, 5, KitCoverOptions{})
	quotient, remainder, err := cover.Divide(divisor, 5, KitCoverOptions{})
	if err != nil || !slices.Equal(quotient.Cubes, []uint32{4, 8}) || !slices.Equal(remainder.Cubes, []uint32{16}) {
		t.Fatalf("division: q=%#v r=%#v err=%v", quotient, remainder, err)
	}
	reconstructed := append(modelCoverProduct(divisor.Cubes, quotient.Cubes), remainder.Cubes...)
	for assignment := uint32(0); assignment < 1<<5; assignment++ {
		if modelLiteralCover(cover.Cubes, assignment) != modelLiteralCover(reconstructed, assignment) {
			t.Fatalf("reconstruction at %05b", assignment)
		}
	}
	literalQuotient, err := cover.DivideByLiteral(0, 5, KitCoverOptions{})
	if err != nil || !slices.Equal(literalQuotient.Cubes, []uint32{4, 8}) {
		t.Fatalf("literal quotient: %#v %v", literalQuotient, err)
	}
	dividend, _ := NewKitSOPCover([]uint32{1 | 4, 1 | 8, 2 | 4, 2 | 8}, 4, KitCoverOptions{})
	kernel, ok, err := dividend.Divisor(4, KitCoverOptions{})
	if err != nil || !ok || len(kernel.Cubes) == 0 || !kernel.IsCubeFree() {
		t.Fatalf("kernel: %#v %v %v", kernel, ok, err)
	}
	trivial, _ := NewKitSOPCover([]uint32{1, 2}, 2, KitCoverOptions{})
	if got, ok, err := trivial.Divisor(2, KitCoverOptions{}); err != nil || ok || got.Cubes != nil {
		t.Fatalf("trivial divisor: %#v %v %v", got, ok, err)
	}
}

func TestKitPLARoundTripAndRecognition(t *testing.T) {
	tests := []struct {
		text       string
		buffer     bool
		inverter   bool
		constant0  bool
		constant1  bool
		complement bool
	}{
		{" 0\n", false, false, true, false, true},
		{" 1\n", false, false, false, true, false},
		{"1 1\n", true, false, false, false, false},
		{"0 0\n", true, false, false, false, true},
		{"0 1\n", false, true, false, false, false},
		{"1 0\n", false, true, false, false, true},
		{"0- 1\n-1 1\n", false, false, false, false, false},
	}
	for _, test := range tests {
		pla, err := ParseKitPLA(test.text, KitCoverOptions{})
		if err != nil {
			t.Fatalf("parse %q: %v", test.text, err)
		}
		got, err := pla.Format()
		if err != nil || got != test.text {
			t.Fatalf("format %q: %q %v", test.text, got, err)
		}
		if pla.IsBuffer() != test.buffer || pla.IsInverter() != test.inverter || pla.IsConst0() != test.constant0 || pla.IsConst1() != test.constant1 || pla.IsComplemented() != test.complement {
			t.Fatalf("recognition %q: %#v", test.text, pla)
		}
		twice := pla.Complement().Complement()
		again, _ := twice.Format()
		if again != test.text {
			t.Fatalf("double complement %q: %q", test.text, again)
		}
	}
	for _, bad := range []string{"", "1 1", "1 1\n0 0\n", "2 1\n", "1 z\n", "12345678901234567 1\n", "1\x00 1\n"} {
		if _, err := ParseKitPLA(bad, KitCoverOptions{}); err == nil {
			t.Fatalf("accepted malformed PLA %q", bad)
		}
	}
}

func TestKitPLAExhaustiveTruth(t *testing.T) {
	for variables := 0; variables <= 3; variables++ {
		functions := 1 << uint(1<<uint(variables))
		for function := 0; function < functions; function++ {
			table, err := NewConstantTruthTable(variables, false, 1)
			if err != nil {
				t.Fatal(err)
			}
			for assignment := 0; assignment < 1<<uint(variables); assignment++ {
				if err := table.Set(uint64(assignment), function>>uint(assignment)&1 != 0); err != nil {
					t.Fatal(err)
				}
			}
			pla, err := KitPLAFromTruth(table, TruthISOPOptions{TryBoth: true}, KitCoverOptions{})
			if err != nil {
				t.Fatalf("from truth n=%d f=%x: %v", variables, function, err)
			}
			text, err := pla.Format()
			if err != nil {
				t.Fatal(err)
			}
			reparsed, err := ParseKitPLA(text, KitCoverOptions{})
			if err != nil {
				t.Fatal(err)
			}
			for assignment := 0; assignment < 1<<uint(variables); assignment++ {
				var got bool
				if reparsed.Variables == 0 {
					got, err = reparsed.Evaluate(0)
				} else {
					got, err = reparsed.Evaluate(uint64(assignment))
				}
				if err != nil || got != (function>>uint(assignment)&1 != 0) {
					t.Fatalf("n=%d f=%x assignment=%d got=%v err=%v PLA=%q", variables, function, assignment, got, err, text)
				}
			}
		}
	}
}

func TestKitPLAComposition(t *testing.T) {
	pla, _ := ParseKitPLA("10 1\n01 1\n", KitCoverOptions{})
	a, _ := NewVariableTruthTable(3, 0, 1)
	b, _ := NewVariableTruthTable(3, 1, 1)
	notB := b.Not()
	result, err := pla.ComposeTruth([]*TruthTable{a, notB})
	if err != nil {
		t.Fatal(err)
	}
	for assignment := uint64(0); assignment < 8; assignment++ {
		got, _ := result.Get(assignment)
		av := assignment&1 != 0
		bv := assignment&2 != 0
		if got != (av == bv) {
			t.Fatalf("composition %d: %v", assignment, got)
		}
	}
}

func TestKitCoverLimitsCancellationAndAtomicity(t *testing.T) {
	input := []uint32{3, 5}
	if _, err := NewKitSOPCover(input, 2, KitCoverOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatalf("out-of-range cube: %v", err)
	}
	cover, _ := NewKitSOPCover([]uint32{1, 1}, 2, KitCoverOptions{})
	original := slices.Clone(cover.Cubes)
	for _, options := range []KitCoverOptions{{MaxSteps: 1}, {Stop: func() bool { return true }}} {
		if _, _, err := cover.Divide(KitSOPCover{Cubes: []uint32{1}}, 2, options); err == nil {
			t.Fatalf("division should stop for %#v", options)
		}
		if !slices.Equal(cover.Cubes, original) {
			t.Fatal("failed division mutated input")
		}
	}
	if _, err := ParseKitPLA("-- 1\n-- 1\n", KitCoverOptions{MaxCubes: 1}); !errors.Is(err, ErrLimit) {
		t.Fatalf("cube limit: %v", err)
	}
	if _, err := StartKitPLA(16, 2, KitCoverOptions{MaxCubeEntries: 16}); !errors.Is(err, ErrLimit) {
		t.Fatalf("entry limit: %v", err)
	}
}

func TestKitSOPIndependentModelCampaign(t *testing.T) {
	rng := rand.New(rand.NewSource(0x51a0))
	for trial := 0; trial < 5000; trial++ {
		literals := 1 + rng.Intn(8)
		divisorCount := 1 + rng.Intn(3)
		divisorCubes := make([]uint32, divisorCount)
		for i := range divisorCubes {
			divisorCubes[i] = uint32(rng.Intn(1 << uint(literals)))
		}
		quotientCubes := make([]uint32, 1+rng.Intn(4))
		for i := range quotientCubes {
			quotientCubes[i] = uint32(rng.Intn(1 << uint(literals)))
		}
		dividendCubes := modelCoverProduct(divisorCubes, quotientCubes)
		for i := 0; i < rng.Intn(3); i++ {
			dividendCubes = append(dividendCubes, uint32(rng.Intn(1<<uint(literals))))
		}
		dividend, _ := NewKitSOPCover(dividendCubes, literals, KitCoverOptions{})
		divisor, _ := NewKitSOPCover(divisorCubes, literals, KitCoverOptions{})
		quotient, remainder, err := dividend.Divide(divisor, literals, KitCoverOptions{})
		if err != nil {
			t.Fatal(err)
		}
		reconstructed := append(modelCoverProduct(divisor.Cubes, quotient.Cubes), remainder.Cubes...)
		for assignment := uint32(0); assignment < 1<<uint(literals); assignment++ {
			if modelLiteralCover(dividend.Cubes, assignment) != modelLiteralCover(reconstructed, assignment) {
				t.Fatalf("trial %d assignment %x", trial, assignment)
			}
		}
	}
}

func FuzzKitSOPPLA(f *testing.F) {
	f.Add([]byte{3, 2, 1, 3, 5, 1, 2})
	f.Add([]byte{1, 1, 0, 0})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) < 3 {
			return
		}
		literals := 1 + int(data[0]%8)
		divisorCount := 1 + int(data[1]%3)
		if 2+divisorCount > len(data) {
			return
		}
		mask := uint32(1<<uint(literals)) - 1
		divisorCubes := make([]uint32, divisorCount)
		for i := range divisorCubes {
			divisorCubes[i] = uint32(data[2+i]) & mask
		}
		dividendCubes := make([]uint32, 0, len(data)-2-divisorCount)
		for _, value := range data[2+divisorCount:] {
			dividendCubes = append(dividendCubes, uint32(value)&mask)
		}
		if len(dividendCubes) < len(divisorCubes) {
			return
		}
		dividend, err := NewKitSOPCover(dividendCubes, literals, KitCoverOptions{MaxCubes: 256, MaxCubeEntries: 2048})
		if err != nil {
			return
		}
		divisor, err := NewKitSOPCover(divisorCubes, literals, KitCoverOptions{MaxCubes: 256, MaxCubeEntries: 2048})
		if err != nil {
			return
		}
		quotient, remainder, err := dividend.Divide(divisor, literals, KitCoverOptions{MaxCubes: 256, MaxCubeEntries: 2048, MaxSteps: 1 << 16})
		if errors.Is(err, ErrLimit) {
			return
		}
		if err != nil {
			t.Fatal(err)
		}
		reconstructed := append(modelCoverProduct(divisor.Cubes, quotient.Cubes), remainder.Cubes...)
		for assignment := uint32(0); assignment < 1<<uint(literals); assignment++ {
			if modelLiteralCover(dividend.Cubes, assignment) != modelLiteralCover(reconstructed, assignment) {
				t.Fatalf("division changed function at %x", assignment)
			}
		}
		variables := literals / 2
		if variables == 0 {
			return
		}
		plaCubes := make([]uint32, 0, len(dividend.Cubes))
		plaMask := uint32(1<<uint(2*variables)) - 1
		for _, cube := range dividend.Cubes {
			cube &= plaMask
			valid := true
			for variable := 0; variable < variables; variable++ {
				valid = valid && cube>>uint(2*variable)&3 != 3
			}
			if valid {
				plaCubes = append(plaCubes, cube)
			}
		}
		if len(plaCubes) == 0 {
			return
		}
		pla, err := NewKitPLA(variables, plaCubes, '1', KitCoverOptions{MaxCubes: 256, MaxCubeEntries: 2048})
		if err != nil {
			t.Fatal(err)
		}
		text, err := pla.Format()
		if err != nil {
			t.Fatal(err)
		}
		roundTrip, err := ParseKitPLA(text, KitCoverOptions{MaxCubes: 256, MaxCubeEntries: 2048})
		if err != nil || !slices.Equal(pla.Cubes, roundTrip.Cubes) || pla.Output != roundTrip.Output {
			t.Fatalf("PLA round trip: %#v %#v %v", pla, roundTrip, err)
		}
	})
}
