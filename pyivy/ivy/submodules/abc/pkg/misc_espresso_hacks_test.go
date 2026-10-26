package abc

// Go-only named-don't-care models for src/misc/espresso/hack.c.

import (
	"errors"
	"slices"
	"testing"
)

func espressoDontCarePLA(t testing.TB, truth uint64) *EspressoPLA {
	t.Helper()
	base := espressoPLAFromOutputTruths(t, 2, []uint64{truth & 0xf})
	labels := []string{"DONT_CARE.bar", "DONT_CARE", "a.bar", "a", "y"}
	pla, err := NewEspressoPLA(base.Layout, base.On, base.DontCare, base.Off, labels)
	if err != nil {
		t.Fatal(err)
	}
	return pla
}

func TestEspressoMapDontCareVariableTruthLayoutAndOwnership(t *testing.T) {
	pla := espressoDontCarePLA(t, 0b1000)
	before := espressoFamilyMasks(t, pla.On)
	mapped, err := pla.MapDontCareVariable(EspressoSymbolicOptions{})
	if err != nil || !slices.Equal(mapped.Layout.partSizes, []int{2, 1}) || mapped.Layout.binaryVariables != 1 || !slices.Equal(mapped.Labels, []string{"a.bar", "a", "y"}) {
		t.Fatal(mapped.Layout.partSizes, mapped.Layout.binaryVariables, mapped.Labels, err)
	}
	if got := espressoFamilyTruth(t, mapped.Layout, mapped.On); got != 0 {
		t.Fatal("on", got, espressoFamilyMasks(t, mapped.On))
	}
	if got := espressoFamilyTruth(t, mapped.Layout, mapped.DontCare); got != 0b10 {
		t.Fatal("dc", got, espressoFamilyMasks(t, mapped.DontCare))
	}
	if got := espressoFamilyTruth(t, mapped.Layout, mapped.Off); got != 0b01 {
		t.Fatal("off", got, espressoFamilyMasks(t, mapped.Off))
	}
	if !slices.Equal(espressoFamilyMasks(t, pla.On), before) {
		t.Fatal("mapping mutated source")
	}

	withoutLabels, _ := NewEspressoPLA(pla.Layout, pla.On, pla.DontCare, pla.Off, nil)
	clone, err := withoutLabels.MapDontCareVariable(EspressoSymbolicOptions{})
	if err != nil || !slices.Equal(clone.Layout.partSizes, pla.Layout.partSizes) {
		t.Fatal(clone, err)
	}
}

func TestEspressoMapDontCareVariableIndependentCubeAndErrors(t *testing.T) {
	layout, _ := NewEspressoCubeLayout([]int{2, 2, 1}, 2)
	onCube := espressoOutputCube(t, 5, 0, 1, 3, 4)
	on := espressoOutputFamily(t, 5, onCube)
	empty := espressoOutputFamily(t, 5)
	off, _ := layout.ComplementCover(on, EspressoSharpOptions{})
	pla, _ := NewEspressoPLA(layout, on, empty, off, []string{"dontcare.bar", "dontcare", "a.bar", "a", "y"})
	mapped, err := pla.MapDontCareVariable(EspressoSymbolicOptions{})
	if err != nil || espressoFamilyTruth(t, mapped.Layout, mapped.On) != 0b10 || espressoFamilyTruth(t, mapped.Layout, mapped.DontCare) != 0 {
		t.Fatal(espressoFamilyMasks(t, mapped.On), espressoFamilyMasks(t, mapped.DontCare), err)
	}
	if result, err := pla.MapDontCareVariable(EspressoSymbolicOptions{MaxOperations: 1}); result != nil || !errors.Is(err, ErrLimit) {
		t.Fatal(result, err)
	}
	if result, err := pla.MapDontCareVariable(EspressoSymbolicOptions{MaxCubes: 1}); result != nil || !errors.Is(err, ErrLimit) {
		t.Fatal(result, err)
	}
	if result, err := pla.MapDontCareVariable(EspressoSymbolicOptions{Stop: func() bool { return true }}); result != nil || !errors.Is(err, ErrCancelled) {
		t.Fatal(result, err)
	}
}

func FuzzEspressoMapDontCareVariable(f *testing.F) {
	f.Add(byte(0x8))
	f.Add(byte(0x6))
	f.Add(byte(0xf))
	f.Fuzz(func(t *testing.T, truthByte byte) {
		truth := uint64(truthByte & 0xf)
		mapped, err := espressoDontCarePLA(t, truth).MapDontCareVariable(EspressoSymbolicOptions{MaxCubes: 1 << 12, MaxOperations: 1 << 18})
		if err != nil {
			t.Fatal(truth, err)
		}
		wantDC := uint64(0)
		for a := 0; a < 2; a++ {
			zero := truth >> uint(2*a) & 1
			one := truth >> uint(2*a+1) & 1
			if zero != one {
				wantDC |= 1 << uint(a)
			}
		}
		gotDC := espressoFamilyTruth(t, mapped.Layout, mapped.DontCare)
		if gotDC != wantDC {
			t.Fatal(truth, gotDC, wantDC, espressoFamilyMasks(t, mapped.DontCare))
		}
		gotOn := espressoFamilyTruth(t, mapped.Layout, mapped.On)
		gotOff := espressoFamilyTruth(t, mapped.Layout, mapped.Off)
		if gotOn&gotDC != 0 || gotOff != 0b11&^(gotOn|gotDC) {
			t.Fatal(truth, gotOn, gotDC, gotOff)
		}
	})
}

func espressoFSMFixture(t testing.TB, external bool) *EspressoPLA {
	t.Helper()
	outputSize := 2
	if external {
		outputSize++
	}
	layout, _ := NewEspressoCubeLayout([]int{2, 2, outputSize}, 1)
	cubes := []*EspressoSet{
		espressoOutputCube(t, layout.bitCount, 0, 2, 3, 4),
		espressoOutputCube(t, layout.bitCount, 1, 2, 5),
	}
	if external {
		_ = cubes[0].Set(6, true)
		cubes = append(cubes, espressoOutputCube(t, layout.bitCount, 0, 3, 6))
	}
	on := espressoOutputFamily(t, layout.bitCount, cubes...)
	empty := espressoOutputFamily(t, layout.bitCount)
	off, err := layout.ComplementCover(on, EspressoSharpOptions{})
	if err != nil {
		t.Fatal(err)
	}
	labels := []string{"in.bar", "in", "S0", "S1", "N0", "N1"}
	if external {
		labels = append(labels, "out")
	}
	pla, err := NewEspressoPLA(layout, on, empty, off, labels)
	if err != nil {
		t.Fatal(err)
	}
	return pla
}

func TestEspressoDisassembleFSMPartitionsDiagnosticsAndKISS(t *testing.T) {
	pla := espressoFSMFixture(t, true)
	before := espressoFamilyMasks(t, pla.On)
	result, err := pla.DisassembleFSM(EspressoFSMDisassembleOptions{Verbose: true, Minimize: EspressoMinimizeOptions{SingleExpand: true}})
	if err != nil || result.PLA == nil || !slices.Equal(result.PLA.Layout.partSizes, []int{2, 2, 2, 1}) || result.PLA.Layout.binaryVariables != 1 {
		t.Fatal(result.PLA, err)
	}
	wantStats := []EspressoFSMArcStatistic{
		{Kind: EspressoFSMEveryState, From: -1, To: 0, Before: 1, After: 1},
		{Kind: EspressoFSMNowhere, From: -1, To: -1, Before: 1, After: 1},
		{Kind: EspressoFSMSpecific, From: 0, To: 1, Before: 1, After: 1},
	}
	if !slices.Equal(result.Statistics, wantStats) {
		t.Fatal(result.Statistics, wantStats)
	}
	wantDiagnostics := "# state EVERY to 0, before=1 after=1\n# state ANY to NOWHERE, before=1 after=1\n# state 0 to 1, before=1 after=1\n"
	if result.Diagnostics != wantDiagnostics {
		t.Fatal(result.Diagnostics)
	}
	wantKISS := "0 - N0 1\n0 S1 ~ 1\n1 S0 N1 ~\n"
	if string(result.KISS) != wantKISS {
		t.Fatal(string(result.KISS), wantKISS, espressoFamilyMasks(t, result.PLA.On))
	}
	if !slices.Equal(espressoFamilyMasks(t, pla.On), before) {
		t.Fatal("FSM disassembly mutated source")
	}
	if len(result.PLA.DontCare.sets) != 0 {
		t.Fatal(espressoFamilyMasks(t, result.PLA.DontCare))
	}
}

func TestEspressoDisassembleFSMWithoutExternalOutputsAndErrors(t *testing.T) {
	pla := espressoFSMFixture(t, false)
	result, err := pla.DisassembleFSM(EspressoFSMDisassembleOptions{})
	if err != nil || !slices.Equal(result.PLA.Layout.partSizes, []int{2, 2, 2}) || string(result.KISS) != "0 - N0 \n1 S0 N1 \n" {
		t.Fatal(result.PLA, string(result.KISS), err)
	}
	badLayout, _ := NewEspressoCubeLayout([]int{2, 3, 2}, 1)
	empty := espressoOutputFamily(t, badLayout.bitCount)
	bad, _ := NewEspressoPLA(badLayout, empty, empty, empty, nil)
	if got, disassembleErr := bad.DisassembleFSM(EspressoFSMDisassembleOptions{}); got.PLA != nil || !errors.Is(disassembleErr, ErrDimension) {
		t.Fatal(got, disassembleErr)
	}
	binaryLayout, _ := NewEspressoCubeLayout([]int{2, 1}, 1)
	binaryEmpty := espressoOutputFamily(t, binaryLayout.bitCount)
	binary, _ := NewEspressoPLA(binaryLayout, binaryEmpty, binaryEmpty, binaryEmpty, nil)
	if got, disassembleErr := binary.DisassembleFSM(EspressoFSMDisassembleOptions{}); got.PLA != nil || !errors.Is(disassembleErr, ErrDimension) {
		t.Fatal(got, disassembleErr)
	}
	checks := []struct {
		options EspressoFSMDisassembleOptions
		want    error
	}{
		{EspressoFSMDisassembleOptions{MaxCubes: 1}, ErrLimit},
		{EspressoFSMDisassembleOptions{MaxOperations: 1}, ErrLimit},
		{EspressoFSMDisassembleOptions{MaxBytes: 1}, ErrLimit},
		{EspressoFSMDisassembleOptions{Stop: func() bool { return true }}, ErrCancelled},
	}
	for _, check := range checks {
		got, disassembleErr := pla.DisassembleFSM(check.options)
		if got.PLA != nil || !errors.Is(disassembleErr, check.want) {
			t.Fatal(check.options, got, disassembleErr, check.want)
		}
	}
}

func espressoFSMPartRelation(pla *EspressoPLA, partStart, partCount int) []uint8 {
	states := pla.Layout.partSizes[pla.Layout.binaryVariables]
	result := make([]uint8, partCount)
	for part := 0; part < partCount; part++ {
		for input := 0; input < 2; input++ {
			for state := 0; state < states; state++ {
				for _, cube := range pla.On.sets {
					inputPresent, _ := cube.Has(pla.Layout.firstParts[0] + input)
					statePresent, _ := cube.Has(pla.Layout.firstParts[pla.Layout.binaryVariables] + state)
					outputPresent, _ := cube.Has(partStart + part)
					if inputPresent && statePresent && outputPresent {
						result[part] |= 1 << uint(input+2*state)
						break
					}
				}
			}
		}
	}
	return result
}

func FuzzEspressoDisassembleFSM(f *testing.F) {
	f.Add([]byte{3, 0x3, 0x1, 1, 0x1, 0x2, 0, 0x2, 0x0, 1})
	f.Add([]byte{1, 0x3, 0x3, 1})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) < 4 {
			return
		}
		layout, _ := NewEspressoCubeLayout([]int{2, 2, 3}, 1)
		count := int(data[0]%4) + 1
		family, _ := NewEspressoSetFamily(0, layout.bitCount)
		position := 1
		for cubeIndex := 0; cubeIndex < count; cubeIndex++ {
			cube, _ := NewEspressoSet(layout.bitCount)
			inputMask := data[position%len(data)]&3 | 1
			stateMask := data[(position+1)%len(data)]&3 | 1
			nextMask := data[(position+2)%len(data)] & 3
			external := data[(position+3)%len(data)]&1 != 0
			position += 4
			if nextMask == 0 && !external {
				external = true
			}
			for part := 0; part < 2; part++ {
				if inputMask>>uint(part)&1 != 0 {
					_ = cube.Set(part, true)
				}
				if stateMask>>uint(part)&1 != 0 {
					_ = cube.Set(2+part, true)
				}
				if nextMask>>uint(part)&1 != 0 {
					_ = cube.Set(4+part, true)
				}
			}
			if external {
				_ = cube.Set(6, true)
			}
			_ = family.Add(cube)
		}
		empty := espressoOutputFamily(t, layout.bitCount)
		off, err := layout.ComplementCover(family, EspressoSharpOptions{MaxCubes: 1 << 14, MaxOperations: 1 << 18})
		if err != nil {
			t.Fatal(err)
		}
		pla, _ := NewEspressoPLA(layout, family, empty, off, []string{"i0", "i1", "S0", "S1", "N0", "N1", "o"})
		want := espressoFSMPartRelation(pla, 4, 3)
		result, err := pla.DisassembleFSM(EspressoFSMDisassembleOptions{MaxCubes: 1 << 14, MaxOperations: 1 << 18, MaxRows: 1 << 14, MaxDepth: 128, MaxIterations: 16, MaxBytes: 1 << 16, Minimize: EspressoMinimizeOptions{SingleExpand: true}})
		if err != nil {
			t.Fatal(data, err)
		}
		gotNext := espressoFSMPartRelation(result.PLA, result.PLA.Layout.firstParts[2], 2)
		gotExternal := espressoFSMPartRelation(result.PLA, result.PLA.Layout.firstParts[3], 1)
		got := append(gotNext, gotExternal...)
		if !slices.Equal(got, want) {
			t.Fatal(data, got, want, espressoFamilyMasks(t, family), espressoFamilyMasks(t, result.PLA.On))
		}
	})
}
