package abc

// Go-only tests for src/base/pla.

import (
	"bytes"
	"errors"
	"math/rand"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

const plaSample = `# comment
.type fr
.i 3
.o 2
.p 3
0-1 10
11- 01
--- -~
.e
`

func plaTruth(t *testing.T, p *PlaManager) [][]bool {
	t.Helper()
	if p.inputs > 16 {
		t.Fatal("oracle width")
	}
	r := make([][]bool, 1<<p.inputs)
	for a := range r {
		var err error
		r[a], err = p.Evaluate(uint64(a))
		if err != nil {
			t.Fatal(err)
		}
	}
	return r
}

func plaTruthEqual(a, b [][]bool) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !slices.Equal(a[i], b[i]) {
			return false
		}
	}
	return true
}

func plaCubeSlicesEqual(a, b []PlaCube) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !slices.Equal(a[i].Inputs, b[i].Inputs) || !slices.Equal(a[i].Outputs, b[i].Outputs) {
			return false
		}
	}
	return true
}

func plaIndependentEvaluate(cubes []PlaCube, outputCount int, assignment uint64) []bool {
	result := make([]bool, outputCount)
	for _, cube := range cubes {
		matches := true
		for variable, literal := range cube.Inputs {
			bit := assignment>>variable&1 != 0
			switch literal {
			case PlaLitZero:
				matches = matches && !bit
			case PlaLitOne:
				matches = matches && bit
			case PlaLitFull:
				matches = false
			}
			if !matches {
				break
			}
		}
		if matches {
			for output, literal := range cube.Outputs {
				if literal == PlaLitOne || literal == PlaLitFull {
					result[output] = true
				}
			}
		}
	}
	return result
}

func TestPlaReadWriteManagerAndOwnership(t *testing.T) {
	p, err := ParsePLA([]byte(plaSample), "sample.pla", PlaLimits{})
	if err != nil {
		t.Fatal(err)
	}
	if p.InputCount() != 3 || p.OutputCount() != 2 || p.CubeCount() != 3 || p.FileType() != PlaFileFR {
		t.Fatalf("manager %+v", p)
	}
	if n, _ := p.InputLiteralCount(); n != 4 {
		t.Fatal(n)
	}
	if n, _ := p.OutputLiteralCount(); n != 2 {
		t.Fatal(n)
	}
	cubes, _ := p.Cubes()
	cubes[0].Inputs[0] = PlaLitFull
	again, _ := p.Cubes()
	if again[0].Inputs[0] != PlaLitZero {
		t.Fatal("aliased cube")
	}
	data, err := p.MarshalPLA("fixed")
	if err != nil {
		t.Fatal(err)
	}
	wantPrefix := "# SOP \"sample.pla\" written via PLA package in ABC on fixed\n\n.type fr\n.i 3\n.o 2\n.p 3\n"
	if !bytes.HasPrefix(data, []byte(wantPrefix)) {
		t.Fatalf("format %q", data)
	}
	q, err := ParsePLA(data, "again", PlaLimits{})
	if err != nil {
		t.Fatal(err)
	}
	if !plaTruthEqual(plaTruth(t, p), plaTruth(t, q)) {
		t.Fatal("roundtrip semantics")
	}
	path := filepath.Join(t.TempDir(), "sample.pla")
	if err = p.WritePLAFile(path, "fixed"); err != nil {
		t.Fatal(err)
	}
	fromFile, err := ReadPLAFile(path, PlaLimits{})
	if err != nil || !plaTruthEqual(plaTruth(t, p), plaTruth(t, fromFile)) {
		t.Fatal("file roundtrip", err)
	}
	clone, _ := p.Clone()
	clone.cubes[0].Inputs[0] = PlaLitOne
	if p.cubes[0].Inputs[0] != PlaLitZero {
		t.Fatal("clone aliases")
	}
	clone.Close()
	clone.Close()
	if _, err = clone.Cubes(); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
}

func TestPlaDistanceHashCombineAndSemantics(t *testing.T) {
	p, _ := NewPlaManager("m", "", 3, 1, PlaFileFD, PlaLimits{MaxOperations: 1000})
	for _, s := range []string{"000", "001", "010", "011", "111"} {
		in := make([]PlaLiteral, 3)
		for i, ch := range []byte(s) {
			in[i] = PlaLitZero
			if ch == '1' {
				in[i] = PlaLitOne
			}
		}
		if err := p.AddCube(in, []PlaLiteral{PlaLitOne}); err != nil {
			t.Fatal(err)
		}
	}
	pairs, err := p.DistanceOnePairs()
	if err != nil {
		t.Fatal(err)
	}
	count, err := p.HashDistanceOneCount()
	if err != nil || count != len(pairs) || count != 5 {
		t.Fatalf("pairs=%v count=%d err=%v", pairs, count, err)
	}
	hashes, _ := p.Hashes()
	again, _ := p.Hashes()
	if !slices.Equal(hashes, again) {
		t.Fatal("nondeterministic hashes")
	}
	want := plaTruth(t, p)
	removed, err := p.CombineDistanceOne()
	if err != nil || removed == 0 {
		t.Fatal(removed, err)
	}
	if got := plaTruth(t, p); !plaTruthEqual(got, want) {
		t.Fatalf("combine changed function %v %v", got, want)
	}
}

func TestPlaFastExtractExhaustive(t *testing.T) {
	p, _ := NewPlaManager("fx", "", 4, 2, PlaFileFD, PlaLimits{MaxOperations: 100000})
	for _, s := range []string{"11--", "110-", "111-", "11-1", "0-11"} {
		in := make([]PlaLiteral, 4)
		for i, ch := range []byte(s) {
			switch ch {
			case '0':
				in[i] = PlaLitZero
			case '1':
				in[i] = PlaLitOne
			default:
				in[i] = PlaLitDash
			}
		}
		_ = p.AddCube(in, []PlaLiteral{PlaLitOne, PlaLitZero})
	}
	network, err := p.FastExtract()
	if err != nil {
		t.Fatal(err)
	}
	if len(network.Divisors) == 0 {
		t.Fatal("expected extracted divisor")
	}
	blif, err := network.MarshalBLIF("fx", "fixed", 4096)
	if err != nil || !bytes.Contains(blif, []byte(".model fx\n.inputs i0 i1 i2 i3\n.outputs o0 o1\n")) || !bytes.HasSuffix(blif, []byte(".end\n\n")) {
		t.Fatalf("BLIF %q %v", blif, err)
	}
	for a := uint64(0); a < 16; a++ {
		want, _ := p.Evaluate(a)
		got, evalErr := network.Evaluate(a)
		if evalErr != nil || !slices.Equal(got, want) {
			t.Fatalf("assignment %d got=%v want=%v err=%v", a, got, want, evalErr)
		}
	}
}

func TestPlaPrimesAndLiteralConventions(t *testing.T) {
	primes, err := PlaPrimesTable(5, PlaLimits{})
	if err != nil {
		t.Fatal(err)
	}
	for i, want := range []bool{false, false, true, true, false, true, false, true, false, false, false, true, false, true, false, false, false, true, false, true, false, false, false, true, false, false, false, false, false, true, false, true} {
		if primes[i] != want {
			t.Fatalf("prime %d", i)
		}
	}
	if !PlaCubesDistanceOne([]PlaLiteral{PlaLitDash}, []PlaLiteral{PlaLitOne}) || PlaCubesDistanceOne([]PlaLiteral{PlaLitZero, PlaLitOne}, []PlaLiteral{PlaLitOne, PlaLitZero}) {
		t.Fatal("distance")
	}
	if variable, ok := PlaCubesConsensus([]PlaLiteral{PlaLitZero, PlaLitDash}, []PlaLiteral{PlaLitOne, PlaLitOne}); !ok || variable != 0 {
		t.Fatal("consensus", variable, ok)
	}
	detector, err := PlaPrimesDetector(5, PlaLimits{MaxOperations: 100000})
	if err != nil {
		t.Fatal(err)
	}
	for assignment, prime := range primes {
		value, evalErr := detector.Evaluate(uint64(assignment))
		if evalErr != nil || value[0] != prime {
			t.Fatal("detector", assignment, value, evalErr)
		}
	}
	generated, err := GeneratePLA(5, 3, 10, 7, PlaLimits{})
	if err != nil || generated.CubeCount() != 10 {
		t.Fatal("generate", err)
	}
}

func TestPlaConsensusExhaustivePackedLiteralModel(t *testing.T) {
	// The source consensus predicate operates on the two-bit encoding itself:
	// exactly one position must differ in both bits; arbitrary one-bit
	// differences at other positions are permitted.
	for encodedA := 0; encodedA < 64; encodedA++ {
		for encodedB := 0; encodedB < 64; encodedB++ {
			a, b := make([]PlaLiteral, 3), make([]PlaLiteral, 3)
			opposite := -1
			valid := true
			for variable := 0; variable < 3; variable++ {
				a[variable] = PlaLiteral(encodedA >> (2 * variable) & 3)
				b[variable] = PlaLiteral(encodedB >> (2 * variable) & 3)
				difference := a[variable] ^ b[variable]
				if difference&1 != 0 && difference&2 != 0 {
					if opposite >= 0 {
						valid = false
					}
					opposite = variable
				}
			}
			valid = valid && opposite >= 0
			gotVariable, got := PlaCubesConsensus(a, b)
			if got != valid || got && gotVariable != opposite {
				t.Fatalf("%v %v: (%d,%v) want (%d,%v)", a, b, gotVariable, got, opposite, valid)
			}
			differences := 0
			for variable := range a {
				if a[variable] != b[variable] {
					differences++
				}
			}
			if PlaCubesDistanceOne(a, b) != (differences == 1) {
				t.Fatalf("distance %v %v", a, b)
			}
		}
	}
}

func TestPlaSimpleMinimizeAllThreeVariableFunctions(t *testing.T) {
	for function := 0; function < 256; function++ {
		on := make([]bool, 8)
		for assignment := range on {
			on[assignment] = function>>assignment&1 != 0
		}
		manager, err := PlaSimpleMinimize(on, 3, "all3", PlaLimits{MaxOperations: 100000})
		if err != nil {
			t.Fatalf("function %02x: %v", function, err)
		}
		for assignment, want := range on {
			got, e := manager.Evaluate(uint64(assignment))
			if e != nil || len(got) != 1 || got[0] != want {
				t.Fatalf("function %02x assignment %d: %v want %v: %v", function, assignment, got, want, e)
			}
		}
	}
}

func TestPlaMintermAndSorterGenerators(t *testing.T) {
	from, err := PlaFromMinterms("m", 3, []int{0, 3, 7}, PlaLimits{})
	if err != nil {
		t.Fatal(err)
	}
	for assignment := 0; assignment < 8; assignment++ {
		got, e := from.Evaluate(uint64(assignment))
		want := assignment == 0 || assignment == 3 || assignment == 7
		if e != nil || got[0] != want {
			t.Fatal(assignment, got, want, e)
		}
	}
	if _, err = PlaFromMinterms("bad", 2, []int{1, 1}, PlaLimits{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	sorter, err := GenerateSorterPLA(4, PlaLimits{})
	if err != nil || sorter.CubeCount() != 15 {
		t.Fatal(sorter, err)
	}
	for assignment := 1; assignment < 16; assignment++ {
		got, e := sorter.Evaluate(uint64(assignment))
		if e != nil {
			t.Fatal(e)
		}
		ones := 0
		for value := assignment; value != 0; value >>= 1 {
			ones += value & 1
		}
		for output, value := range got {
			if value != (output < ones) {
				t.Fatal(assignment, output, got, ones)
			}
		}
	}
}

func TestPlaNoneFormatAndFXValidation(t *testing.T) {
	p, err := ParsePLA([]byte(".type ???\n.i 1\n.o 1\n.p 1\n0 1\n.e\n"), "none", PlaLimits{})
	if err != nil || p.FileType() != PlaFileNone {
		t.Fatal(p, err)
	}
	encoded, err := p.MarshalPLA("fixed")
	if err != nil || !bytes.Contains(encoded, []byte(".type ???\n")) {
		t.Fatal(string(encoded), err)
	}
	if _, err = ParsePLA(encoded, "none2", PlaLimits{}); err != nil {
		t.Fatal(err)
	}

	valid := &PlaFXNetwork{Inputs: 2, Terms: [][]int{{0}}, Outputs: [][]PlaLiteral{{PlaLitOne}}}
	if _, err = valid.Evaluate(4); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	malformed := []PlaFXNetwork{
		{Inputs: 2, Terms: [][]int{{4}}, Outputs: [][]PlaLiteral{{PlaLitOne}}},
		{Inputs: 2, Terms: [][]int{{0}}, Outputs: nil},
		{Inputs: 2, Divisors: []PlaFXDivisor{{0, 4}}, Terms: [][]int{{0}}, Outputs: [][]PlaLiteral{{PlaLitOne}}},
		{Inputs: 2, Terms: [][]int{{0}}, Outputs: [][]PlaLiteral{{PlaLiteral(9)}}},
	}
	for i := range malformed {
		if _, err = malformed[i].Evaluate(0); err == nil {
			t.Fatalf("malformed %d accepted", i)
		}
		if _, err = malformed[i].MarshalBLIF("m", "fixed", 1000); err == nil {
			t.Fatalf("malformed BLIF %d accepted", i)
		}
	}
	if _, err = PlaSimpleMinimize([]bool{true}, 0, "constant", PlaLimits{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err = NewPlaManager("cancel", "", 1, 1, PlaFileFD, PlaLimits{Stop: func() bool { return true }}); !errors.Is(err, ErrCancelled) {
		t.Fatal(err)
	}
	empty, _ := NewPlaManager("empty", "", 2, 2, PlaFileFD, PlaLimits{})
	emptyFX, err := empty.FastExtract()
	if err != nil {
		t.Fatal(err)
	}
	values, err := emptyFX.Evaluate(0)
	if err != nil || !slices.Equal(values, []bool{false, false}) {
		t.Fatal(values, err)
	}
	blif, err := emptyFX.MarshalBLIF("empty", "fixed", 4096)
	if err != nil || !bytes.Contains(blif, []byte(".outputs o0 o1\n")) {
		t.Fatal(string(blif), err)
	}
}

func TestPlaParserLimitsCancellationAndAtomicity(t *testing.T) {
	bad := []string{".i 1\n.o 1\n.p 2\n0 1\n.e\n", ".i 1\n.o 1\nx 1\n.e\n", ".i 1\n.i 1\n.o 1\n0 1\n", ".i 1\n.o 1\n0 1 extra\n", ".i 0\n.o 1\n"}
	for _, s := range bad {
		if _, err := ParsePLA([]byte(s), "x", PlaLimits{}); !errors.Is(err, ErrDimension) {
			t.Fatalf("accepted %q: %v", s, err)
		}
	}
	if _, err := ParsePLA([]byte(plaSample), "x", PlaLimits{MaxBytes: 2}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if _, err := ParsePLA([]byte(plaSample), "x", PlaLimits{Stop: func() bool { return true }}); !errors.Is(err, ErrCancelled) {
		t.Fatal(err)
	}
	p, _ := ParsePLA([]byte(".i 2\n.o 1\n.p 2\n00 1\n01 1\n.e\n"), "x", PlaLimits{MaxOperations: 1})
	before, _ := p.Cubes()
	if _, err := p.CombineDistanceOne(); err != nil {
		t.Fatal(err)
	}
	after, _ := p.Cubes()
	if plaCubeSlicesEqual(before, after) {
		t.Fatal("expected commit")
	}
	limited, _ := ParsePLA([]byte(".i 2\n.o 1\n.p 3\n00 1\n01 1\n10 1\n.e\n"), "x", PlaLimits{MaxOperations: 1})
	before, _ = limited.Cubes()
	if _, err := limited.CombineDistanceOne(); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	after, _ = limited.Cubes()
	if !plaCubeSlicesEqual(before, after) {
		t.Fatal("failed transform mutated manager")
	}
}

func TestPlaRandomRoundtripAndModel5000(t *testing.T) {
	r := rand.New(rand.NewSource(91))
	for execution := 0; execution < 5000; execution++ {
		ins := 1 + r.Intn(6)
		outs := 1 + r.Intn(3)
		p, _ := NewPlaManager("random", "", ins, outs, PlaFileFD, PlaLimits{})
		for c := 0; c < 1+r.Intn(10); c++ {
			in := make([]PlaLiteral, ins)
			out := make([]PlaLiteral, outs)
			for i := range in {
				in[i] = PlaLiteral(r.Intn(4))
			}
			for i := range out {
				out[i] = PlaLiteral(r.Intn(4))
			}
			_ = p.AddCube(in, out)
		}
		data, err := p.MarshalPLA("fixed")
		if err != nil {
			t.Fatal(execution, err)
		}
		q, err := ParsePLA(data, "q", PlaLimits{})
		if err != nil {
			t.Fatal(execution, err)
		}
		cubes, _ := p.Cubes()
		for assignment := 0; assignment < 1<<ins; assignment++ {
			got, evalErr := p.Evaluate(uint64(assignment))
			want := plaIndependentEvaluate(cubes, outs, uint64(assignment))
			if evalErr != nil || !slices.Equal(got, want) {
				t.Fatal(execution, assignment, got, want, evalErr)
			}
		}
		if !plaTruthEqual(plaTruth(t, p), plaTruth(t, q)) {
			t.Fatal(execution)
		}
	}
}

func FuzzPlaParserRoundtrip(f *testing.F) {
	f.Add([]byte(plaSample))
	f.Add([]byte(".i 1\n.o 1\n.p 1\n- 1\n.e\n"))
	f.Fuzz(func(t *testing.T, data []byte) {
		p, err := ParsePLA(data, "fuzz", PlaLimits{MaxBytes: 4096, MaxInputs: 32, MaxOutputs: 32, MaxCubes: 128, MaxOperations: 10000})
		if err != nil {
			return
		}
		encoded, err := p.MarshalPLA("fuzz")
		if err != nil {
			t.Fatal(err)
		}
		q, err := ParsePLA(encoded, "q", PlaLimits{MaxBytes: 4096, MaxInputs: 32, MaxOutputs: 32, MaxCubes: 128})
		if err != nil {
			t.Fatal(err)
		}
		first, _ := p.Cubes()
		second, _ := q.Cubes()
		if !plaCubeSlicesEqual(first, second) {
			t.Fatalf("roundtrip\n%s", strings.TrimSpace(string(encoded)))
		}
	})
}

func FuzzPlaSemanticTransforms(f *testing.F) {
	f.Add(uint64(1), uint8(4), uint8(8))
	f.Fuzz(func(t *testing.T, seed uint64, width, count uint8) {
		inputs := int(width%8) + 1
		cubes := int(count%24) + 1
		r := rand.New(rand.NewSource(int64(seed)))
		p, _ := NewPlaManager("f", "", inputs, 1, PlaFileFD, PlaLimits{MaxOperations: 200000})
		for c := 0; c < cubes; c++ {
			in := make([]PlaLiteral, inputs)
			for i := range in {
				in[i] = PlaLiteral(r.Intn(3))
			}
			_ = p.AddCube(in, []PlaLiteral{PlaLitOne})
		}
		want := plaTruth(t, p)
		combined, _ := p.Clone()
		if _, err := combined.CombineDistanceOne(); err != nil {
			t.Fatal(err)
		}
		if !plaTruthEqual(want, plaTruth(t, combined)) {
			t.Fatal("combine")
		}
		fx, err := p.FastExtract()
		if err != nil {
			t.Fatal(err)
		}
		for a := range want {
			got, err := fx.Evaluate(uint64(a))
			if err != nil || !slices.Equal(got, want[a]) {
				t.Fatal("fx", a, err)
			}
		}
	})
}
