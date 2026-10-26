package abc

// Stand-alone PLA parser models for src/misc/espresso/cvrin.c.

import (
	"errors"
	"slices"
	"strings"
	"testing"
)

func TestEspressoInputStandardRoundTripAndSummary(t *testing.T) {
	source, _ := espressoOutputFixture(t)
	raw, err := source.MarshalPLA(EspressoPLAOutputOn|EspressoPLAOutputDontCare|EspressoPLAOutputOff, EspressoPLAOutputOptions{})
	if err != nil {
		t.Fatal(err)
	}
	withComments := append([]byte("# permanent Go round trip\n.unknown retained\n"), raw...)
	result, err := ParseEspressoPLA(withComments, EspressoPLAInputOptions{Filename: "mixed.pla"})
	if err != nil {
		t.Fatal(err)
	}
	if result.Lines == 0 || result.Products != 6 || len(result.Comments) != 1 || !slices.Equal(result.Unknown, []string{".unknown retained"}) {
		t.Fatal(result.Lines, result.Products, result.Comments, result.Unknown)
	}
	if !slices.Equal(result.PLA.Layout.partSizes, source.Layout.partSizes) || result.PLA.Layout.binaryVariables != source.Layout.binaryVariables || !slices.Equal(result.PLA.Labels, source.Labels) {
		t.Fatal(result.PLA.Layout.partSizes, result.PLA.Labels)
	}
	for _, pair := range [][2]*EspressoSetFamily{{result.PLA.On, source.On}, {result.PLA.DontCare, source.DontCare}, {result.PLA.Off, source.Off}} {
		if !slices.Equal(espressoFamilyMasks(t, pair[0]), espressoFamilyMasks(t, pair[1])) {
			t.Fatal(espressoFamilyMasks(t, pair[0]), espressoFamilyMasks(t, pair[1]))
		}
	}
	summary, err := result.Summary(EspressoCoverCostOptions{})
	if err != nil || !strings.HasPrefix(summary, "# PLA is mixed.pla with 4 variables (2 binary, mv sizes 3 2)\n") || !strings.Contains(summary, "# ON-set cost is ") || !strings.Contains(summary, "# OFF-set cost is ") || !strings.Contains(summary, "# DC-set cost is ") {
		t.Fatal(summary, err)
	}
}

func TestEspressoInputMissingCoversPhaseAndPositive(t *testing.T) {
	input := []byte(".i 2\n.o 1\n.ilb a b\n.ob y\n.type f\n00 1\n11 1\n.end\n")
	base, err := ParseEspressoPLA(input, EspressoPLAInputOptions{NeedsOff: true})
	if err != nil {
		t.Fatal(err)
	}
	universe := espressoTruthUniverse(base.PLA.Layout)
	onTruth := espressoFamilyTruth(t, base.PLA.Layout, base.PLA.On)
	offTruth := espressoFamilyTruth(t, base.PLA.Layout, base.PLA.Off)
	if onTruth != 0b1001 || offTruth != universe&^onTruth || len(base.PLA.DontCare.sets) != 0 {
		t.Fatal(onTruth, offTruth, universe)
	}

	phasedInput := []byte(".i 2\n.o 1\n.type f\n.phase 0\n00 1\n11 1\n.e\n")
	phased, err := ParseEspressoPLA(phasedInput, EspressoPLAInputOptions{NeedsOff: true})
	if err != nil || !slices.Equal(phased.Phase, []bool{false}) || espressoFamilyTruth(t, phased.PLA.Layout, phased.PLA.On) != universe&^onTruth {
		t.Fatal(phased.Phase, espressoFamilyTruth(t, phased.PLA.Layout, phased.PLA.On), err)
	}
	positive, err := ParseEspressoPLA(input, EspressoPLAInputOptions{NeedsOff: true, Positive: true})
	if err != nil || !slices.Equal(positive.Phase, []bool{false}) || espressoFamilyTruth(t, positive.PLA.Layout, positive.PLA.On) != universe&^onTruth {
		t.Fatal(positive.Phase, espressoFamilyTruth(t, positive.PLA.Layout, positive.PLA.On), err)
	}

	offInput := []byte(".i 2\n.o 1\n.type r\n00 0\n11 0\n.end\n")
	fromOff, err := ParseEspressoPLA(offInput, EspressoPLAInputOptions{})
	if err != nil || espressoFamilyTruth(t, fromOff.PLA.Layout, fromOff.PLA.Off) != onTruth || espressoFamilyTruth(t, fromOff.PLA.Layout, fromOff.PLA.On) != universe&^onTruth {
		t.Fatal(espressoFamilyTruth(t, fromOff.PLA.Layout, fromOff.PLA.On), espressoFamilyTruth(t, fromOff.PLA.Layout, fromOff.PLA.Off), err)
	}

	frInput := []byte(".i 1\n.o 1\n.type fr\n0 1\n1 0\n.end\n")
	fromFR, err := ParseEspressoPLA(frInput, EspressoPLAInputOptions{NeedsDontCare: true})
	if err != nil || len(fromFR.PLA.DontCare.sets) != 0 {
		t.Fatal(espressoFamilyMasks(t, fromFR.PLA.DontCare), err)
	}
}

func TestEspressoInputMultipleValuedLabelsPairAndKISS(t *testing.T) {
	mv := []byte(".mv 3 1 -3 2\n.ilb a\n.ob y0 y1\n.type fdr\n0 red 10\n1 blue 01\n- ANY 11\n? ~ ~~\n.end\n")
	parsed, err := ParseEspressoPLA(mv, EspressoPLAInputOptions{})
	if err != nil {
		t.Fatal(err)
	}
	symbolic, _ := parsed.PLA.Layout.SymbolicVariables()
	if !slices.Equal(parsed.PLA.Layout.partSizes, []int{2, 3, 2}) || !slices.Equal(symbolic, []bool{false, true, false}) || !slices.Equal(parsed.PLA.Labels, []string{"a.bar", "a", "red", "blue", "", "y0", "y1"}) {
		t.Fatal(parsed.PLA.Layout.partSizes, symbolic, parsed.PLA.Labels)
	}
	if parsed.Products != 4 || len(parsed.PLA.On.sets) != 3 || len(parsed.PLA.Off.sets) != 2 || len(parsed.PLA.DontCare.sets) != 0 {
		t.Fatal(parsed.Products, len(parsed.PLA.On.sets), len(parsed.PLA.Off.sets), len(parsed.PLA.DontCare.sets))
	}

	pairedInput := []byte(".i 4\n.o 1\n.ilb a b c d\n.pair 2 (a b) (c d)\n0000 1\n.end\n")
	paired, err := ParseEspressoPLA(pairedInput, EspressoPLAInputOptions{})
	if err != nil || !slices.Equal(paired.PLA.Layout.partSizes, []int{4, 4, 1}) || paired.PLA.Layout.binaryVariables != 0 || len(paired.Pairing.Pairs) != 2 {
		t.Fatal(paired.PLA.Layout.partSizes, paired.PLA.Layout.binaryVariables, paired.Pairing, err)
	}

	kiss := []byte(".mv 4 1 -2 -2 1\n.kiss\n.ilb in\n.label var=1 S0 S1\n.ob out\n.type fr\n0 S0 S1 1\n.end\n")
	kissResult, err := ParseEspressoPLA(kiss, EspressoPLAInputOptions{})
	if err != nil || !kissResult.KISS || !slices.Equal(kissResult.PLA.Layout.partSizes, []int{2, 2, 3}) || !slices.Equal(kissResult.PLA.Labels, []string{"in.bar", "in", "S0", "S1", "S0", "S1", "out"}) {
		t.Fatal(kissResult.KISS, kissResult.PLA.Layout.partSizes, kissResult.PLA.Labels, err)
	}
	if len(kissResult.PLA.On.sets) != 1 || len(kissResult.PLA.Off.sets) != 1 {
		t.Fatal(espressoFamilyMasks(t, kissResult.PLA.On), espressoFamilyMasks(t, kissResult.PLA.Off))
	}
}

func TestEspressoInputSymbolicMetadataWarningsAndErrors(t *testing.T) {
	symbolic := []byte(".i 2\n.o 2\n.ilb a b\n.ob y0 y1\n.symbolic a b ; S0 S1 S2 S3 ;\n.symbolic-output y0 y1 ; N0 N1 N2 N3 ;\n00 10\n.end\n")
	mapped, err := ParseEspressoPLA(symbolic, EspressoPLAInputOptions{})
	if err != nil || !slices.Equal(mapped.PLA.Layout.partSizes, []int{4, 4}) || mapped.PLA.Layout.binaryVariables != 0 || !slices.Equal(mapped.PLA.Labels, []string{"S0", "S1", "S2", "S3", "N0", "N1", "N2", "N3"}) {
		t.Fatal(mapped.PLA, err)
	}
	deferred, err := ParseEspressoPLA(symbolic, EspressoPLAInputOptions{DeferSymbolicMappings: true})
	if err != nil || len(deferred.SymbolicInputs) != 1 || len(deferred.SymbolicOutputs) != 1 || len(deferred.SymbolicInputs[0].References) != 2 || !slices.Equal(deferred.SymbolicOutputs[0].Labels, []string{"N0", "N1", "N2", "N3"}) {
		t.Fatal(deferred.SymbolicInputs, deferred.SymbolicOutputs, err)
	}

	malformedProduct := []byte(".i 1\n.o 1\ninvalid\n0 1\n.end\n")
	warned, err := ParseEspressoPLA(malformedProduct, EspressoPLAInputOptions{})
	if err != nil || warned.Products != 1 || len(warned.Warnings) != 1 || len(warned.PLA.On.sets) != 1 {
		t.Fatal(warned.Products, warned.Warnings, err)
	}
	continued := []byte(".i 2\n.o 2\n0\n1 1\n0\n.end\n")
	continuedResult, err := ParseEspressoPLA(continued, EspressoPLAInputOptions{})
	if err != nil || continuedResult.Products != 1 || len(continuedResult.Warnings) != 1 || !strings.Contains(continuedResult.Warnings[0], "spans more than one line") || !slices.Equal(espressoFamilyMasks(t, continuedResult.PLA.On), []uint64{0b011001}) {
		t.Fatal(continuedResult.Products, continuedResult.Warnings, espressoFamilyMasks(t, continuedResult.PLA.On), err)
	}
	badDirectives := [][]byte{
		[]byte(".o 1\n"),
		[]byte(".i 1\n.o 1\n.phase 2\n"),
		[]byte(".mv 2 3 1\n"),
		[]byte(".i 1\n.o 1\n.ilb\n"),
		[]byte(".i 2\n.o 1\n.pair 1 (0 0)\n"),
	}
	for _, input := range badDirectives {
		if result, parseErr := ParseEspressoPLA(input, EspressoPLAInputOptions{}); result.PLA != nil || parseErr == nil {
			t.Fatal(string(input), result, parseErr)
		}
	}
}

func TestEspressoInputLimitsAndCancellation(t *testing.T) {
	input := []byte(".i 1\n.o 1\n0 1\n.end\n")
	checks := []struct {
		options EspressoPLAInputOptions
		want    error
	}{
		{EspressoPLAInputOptions{MaxBytes: 1}, ErrLimit},
		{EspressoPLAInputOptions{MaxLines: 1}, ErrLimit},
		{EspressoPLAInputOptions{MaxVariables: 1}, ErrLimit},
		{EspressoPLAInputOptions{MaxParts: 1}, ErrLimit},
		{EspressoPLAInputOptions{MaxCubes: 1}, nil},
		{EspressoPLAInputOptions{MaxTokens: 1}, ErrLimit},
		{EspressoPLAInputOptions{MaxOperations: 1}, ErrLimit},
		{EspressoPLAInputOptions{Stop: func() bool { return true }}, ErrCancelled},
	}
	for _, check := range checks {
		result, err := ParseEspressoPLA(input, check.options)
		if check.want == nil {
			if err != nil || result.PLA == nil {
				t.Fatal(check.options, result, err)
			}
		} else if result.PLA != nil || !errors.Is(err, check.want) {
			t.Fatal(check.options, result, err, check.want)
		}
	}
	if result, err := ParseEspressoPLA(input, EspressoPLAInputOptions{Type: EspressoPLAOutputEquation}); result.PLA != nil || !errors.Is(err, ErrDimension) {
		t.Fatal(result, err)
	}
}

func FuzzEspressoPLAInputRoundTrip(f *testing.F) {
	f.Add([]byte{2, 3, 2, 0x55, 0xaa, 0x0f})
	f.Add([]byte{1, 1, 1, 0xff, 0x00})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) < 5 {
			return
		}
		binary := int(data[0]%3) + 1
		mvSize := int(data[1]%3) + 1
		outputs := int(data[2]%3) + 1
		parts := make([]int, binary+2)
		for index := 0; index < binary; index++ {
			parts[index] = 2
		}
		parts[binary], parts[binary+1] = mvSize, outputs
		layout, err := NewEspressoCubeLayout(parts, binary)
		if err != nil {
			t.Fatal(err)
		}
		families := make([]*EspressoSetFamily, 3)
		position := 3
		for familyIndex := range families {
			family, _ := NewEspressoSetFamily(0, layout.bitCount)
			count := int(data[position%len(data)] % 3)
			position++
			for cubeIndex := 0; cubeIndex < count; cubeIndex++ {
				cube, _ := NewEspressoSet(layout.bitCount)
				for part := 0; part < layout.bitCount; part++ {
					if data[position%len(data)]&(1<<uint(part&7)) != 0 {
						_ = cube.Set(part, true)
					}
					position++
				}
				output := layout.OutputVariable()
				selectedOutput := false
				for part := 0; part < layout.partSizes[output]; part++ {
					present, _ := cube.Has(layout.firstParts[output] + part)
					selectedOutput = selectedOutput || present
				}
				if !selectedOutput {
					_ = cube.Set(layout.firstParts[output]+cubeIndex%layout.partSizes[output], true)
				}
				_ = family.Add(cube)
			}
			families[familyIndex] = family
		}
		pla, err := NewEspressoPLA(layout, families[0], families[1], families[2], nil)
		if err != nil {
			t.Fatal(err)
		}
		raw, err := pla.MarshalPLA(EspressoPLAOutputOn|EspressoPLAOutputDontCare|EspressoPLAOutputOff, EspressoPLAOutputOptions{MaxCubes: 16, MaxParts: 1 << 14, MaxBytes: 1 << 16})
		if err != nil {
			t.Fatal(err)
		}
		parsed, err := ParseEspressoPLA(raw, EspressoPLAInputOptions{MaxBytes: 1 << 16, MaxLines: 128, MaxVariables: 16, MaxParts: 128, MaxCubes: 16, MaxTokens: 256, MaxOperations: 1 << 16})
		if err != nil {
			t.Fatal(string(raw), err)
		}
		for familyIndex, got := range []*EspressoSetFamily{parsed.PLA.On, parsed.PLA.DontCare, parsed.PLA.Off} {
			if !slices.Equal(espressoFamilyMasks(t, got), espressoFamilyMasks(t, families[familyIndex])) {
				t.Fatal(parts, familyIndex, espressoFamilyMasks(t, got), espressoFamilyMasks(t, families[familyIndex]), string(raw))
			}
		}
	})
}

func FuzzEspressoPLAInputRobustness(f *testing.F) {
	f.Add([]byte(".i 1\n.o 1\n0 1\n.end\n"))
	f.Add([]byte(".mv 3 1 -2 1\n0 S0 1\n"))
	f.Add([]byte{0, 1, 2, 3, 4, 255})
	f.Fuzz(func(t *testing.T, data []byte) {
		result, err := ParseEspressoPLA(data, EspressoPLAInputOptions{DeferSymbolicMappings: true, MaxBytes: 1 << 12, MaxLines: 128, MaxVariables: 32, MaxParts: 256, MaxCubes: 128, MaxTokens: 512, MaxOperations: 1 << 14, MaxDepth: 128})
		if err == nil {
			if result.PLA == nil || result.PLA.check() != nil || result.Products < 0 || result.Lines < 0 {
				t.Fatal(result)
			}
		}
	})
}
