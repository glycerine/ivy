package abc

// Go-only formatting models for src/misc/espresso/cvrout.c.

import (
	"errors"
	"slices"
	"strconv"
	"strings"
	"testing"
)

func espressoOutputCube(t testing.TB, bits int, elements ...int) *EspressoSet {
	t.Helper()
	cube, err := NewEspressoSet(bits)
	if err != nil {
		t.Fatal(err)
	}
	for _, element := range elements {
		if err = cube.Set(element, true); err != nil {
			t.Fatal(err)
		}
	}
	return cube
}

func espressoOutputFamily(t testing.TB, bits int, cubes ...*EspressoSet) *EspressoSetFamily {
	t.Helper()
	family, err := NewEspressoSetFamily(len(cubes), bits)
	if err != nil {
		t.Fatal(err)
	}
	for _, cube := range cubes {
		if err = family.Add(cube); err != nil {
			t.Fatal(err)
		}
	}
	return family
}

func espressoOutputFixture(t testing.TB) (*EspressoPLA, []*EspressoSet) {
	t.Helper()
	layout, err := NewEspressoCubeLayout([]int{2, 2, 3, 2}, 2)
	if err != nil {
		t.Fatal(err)
	}
	cubes := []*EspressoSet{
		espressoOutputCube(t, 9, 0, 2, 3, 4, 5, 7),
		espressoOutputCube(t, 9, 1, 3, 4, 5, 8),
		espressoOutputCube(t, 9, 0, 1, 2, 6, 7, 8),
		espressoOutputCube(t, 9, 0, 1, 2, 3, 4, 5, 6, 7),
	}
	dc := espressoOutputCube(t, 9, 0, 2, 4, 8)
	off := espressoOutputCube(t, 9, 1, 3, 6, 8)
	empty := espressoOutputFamily(t, 9)
	labels := []string{"a.bar", "a", "b.bar", "b", "red", "green", "blue", "y0", "y1"}
	pla, err := NewEspressoPLA(layout, espressoOutputFamily(t, 9, cubes...), espressoOutputFamily(t, 9, dc), espressoOutputFamily(t, 9, off), labels)
	if err != nil {
		t.Fatal(err)
	}
	_ = empty
	return pla, cubes
}

func TestEspressoOutputCubeHeaderAndPLA(t *testing.T) {
	pla, cubes := espressoOutputFixture(t)
	formatted, err := pla.Layout.FormatCube(cubes[0], "~1", EspressoPLAOutputOptions{})
	if err != nil || formatted != "0- 110 1~" {
		t.Fatal(formatted, err)
	}
	expanded, err := pla.Layout.FormatExpandedCube(cubes[0], EspressoPLAOutputOptions{Phase: []bool{true, false}})
	if err != nil || expanded != "1~11~~1 1~" {
		t.Fatal(expanded, err)
	}
	header, err := pla.FormatHeader(EspressoPLAOutputOn|EspressoPLAOutputDontCare|EspressoPLAOutputOff, EspressoPLAOutputOptions{Phase: []bool{true, false}})
	wantHeader := ".type fdr\n.mv 4 2 3 2\n.ilb a b\n.ob y0 y1\n.label var=2 red green blue\n#.phase 10\n"
	if err != nil || header != wantHeader {
		t.Fatalf("header:\n%q\nwant:\n%q\nerr=%v", header, wantHeader, err)
	}
	raw, err := pla.MarshalPLA(EspressoPLAOutputOn|EspressoPLAOutputDontCare|EspressoPLAOutputOff, EspressoPLAOutputOptions{})
	want := ".type fdr\n.mv 4 2 3 2\n.ilb a b\n.ob y0 y1\n.label var=2 red green blue\n.p 6\n" +
		"0- 110 1~\n11 110 ~1\n-0 001 11\n-- 111 1~\n00 100 ~2\n11 001 ~0\n.end\n"
	if err != nil || string(raw) != want {
		t.Fatalf("PLA:\n%q\nwant:\n%q\nerr=%v", raw, want, err)
	}
	on, err := pla.MarshalPLA(EspressoPLAOutputOn, EspressoPLAOutputOptions{})
	if err != nil || !strings.HasPrefix(string(on), ".mv 4 2 3 2\n") || !strings.HasSuffix(string(on), ".e\n") || strings.Contains(string(on), ".type") {
		t.Fatal(string(on), err)
	}
}

func TestEspressoOutputLabelsEquationPleasureAndDebug(t *testing.T) {
	pla, _ := espressoOutputFixture(t)
	withoutLabels, _ := NewEspressoPLA(pla.Layout, pla.On, pla.DontCare, pla.Off, nil)
	labels, err := withoutLabels.DefaultLabels()
	wantLabels := []string{"v0.bar", "v0", "v1.bar", "v1", "v2.0", "v2.1", "v2.2", "v3.0", "v3.1"}
	if err != nil || !slices.Equal(labels, wantLabels) || len(withoutLabels.Labels) != 0 {
		t.Fatal(labels, withoutLabels.Labels, err)
	}

	binaryLayout, _ := NewEspressoCubeLayout([]int{2, 2, 2}, 2)
	first := espressoOutputCube(t, 6, 0, 2, 3, 4)
	second := espressoOutputCube(t, 6, 1, 3, 4, 5)
	empty := espressoOutputFamily(t, 6)
	binary, _ := NewEspressoPLA(binaryLayout, espressoOutputFamily(t, 6, first, second), empty, empty, []string{"a.bar", "a", "b.bar", "b", "y0", "y1"})
	equations, err := binary.MarshalPLA(EspressoPLAOutputEquation, EspressoPLAOutputOptions{})
	if err != nil || string(equations) != "y0 = (!a) | (a&b);\n\ny1 = (a&b);\n\n" {
		t.Fatal(string(equations), err)
	}
	longName := strings.Repeat("x", 70)
	longLayout, _ := NewEspressoCubeLayout([]int{2, 1}, 1)
	longCube := espressoOutputCube(t, 3, 0, 2)
	longEmpty := espressoOutputFamily(t, 3)
	longPLA, _ := NewEspressoPLA(longLayout, espressoOutputFamily(t, 3, longCube), longEmpty, longEmpty, []string{longName + ".bar", longName, "y"})
	wrapped, err := longPLA.MarshalPLA(EspressoPLAOutputEquation, EspressoPLAOutputOptions{})
	if err != nil || string(wrapped) != "y = (\n    !"+longName+");\n\n" {
		t.Fatal(string(wrapped), err)
	}
	pleasure, err := binary.MarshalPLA(EspressoPLAOutputPleasure, EspressoPLAOutputOptions{Phase: []bool{true, false}})
	if err != nil || !strings.Contains(string(pleasure), ".option unmerged\n.label a.bar a b.bar b y0 y1\n.group (a.bar a) ( b.bar b)\n.p 2\n") || !strings.HasSuffix(string(pleasure), ".end\n") {
		t.Fatal(string(pleasure), err)
	}

	debug, err := binary.Layout.FormatCoverDebug(binary.On, "F", 0, true, EspressoPLAOutputOptions{})
	if err != nil || debug != "\nF[0]: ord(T)=2\n   1. 0- 10\n   2. 11 11\n" {
		t.Fatal(debug, err)
	}
	plain, err := binary.Layout.FormatCoverCubes(binary.On, EspressoPLAOutputOptions{})
	if err != nil || plain != "0- 10\n11 11\n" {
		t.Fatal(plain, err)
	}
	list, _ := NewEspressoCubeList(binary.Layout, binary.On)
	listDebug, err := binary.Layout.FormatCubeListDebug(list, "T", 1, false, EspressoPLAOutputOptions{})
	if err != nil || listDebug != "T[1]: ord(T)=2\n" {
		t.Fatal(listDebug, err)
	}
}

func TestEspressoOutputKISSAndConstraints(t *testing.T) {
	pla, cubes := espressoOutputFixture(t)
	numeric, err := pla.MarshalPLA(EspressoPLAOutputConstraints, EspressoPLAOutputOptions{})
	wantNumeric := "# Symbolic constraints for variable 2 (Numeric form)\n# unconstrained weight = 2\nnum_codes=3\nweight=2:  0 1\n"
	if err != nil || string(numeric) != wantNumeric {
		t.Fatal(string(numeric), err)
	}
	symbolic, err := pla.MarshalPLA(EspressoPLAOutputSymbolicConstraints, EspressoPLAOutputOptions{})
	wantSymbolic := "# Symbolic constraints for variable 2 (Symbolic form)\n#   w=2: ( red green )\n"
	if err != nil || string(symbolic) != wantSymbolic {
		t.Fatal(string(symbolic), err)
	}
	if raw, err := pla.MarshalPLA(EspressoPLAOutputKISS, EspressoPLAOutputOptions{}); raw != nil || !errors.Is(err, ErrInvariant) {
		t.Fatal(string(raw), err)
	}

	on := espressoOutputFamily(t, 9, cubes[2], cubes[3])
	kissPLA, _ := NewEspressoPLA(pla.Layout, on, pla.DontCare, pla.Off, pla.Labels)
	kiss, err := kissPLA.MarshalPLA(EspressoPLAOutputKISS, EspressoPLAOutputOptions{})
	if err != nil || string(kiss) != "-0 blue 11\n-- - 1~\n00 red ~2\n" {
		t.Fatal(string(kiss), err)
	}
	combined, err := pla.MarshalPLA(EspressoPLAOutputConstraints|EspressoPLAOutputSymbolicConstraints|EspressoPLAOutputOff, EspressoPLAOutputOptions{})
	if err != nil || !strings.HasPrefix(string(combined), wantNumeric+wantSymbolic+".type r\n") || !strings.HasSuffix(string(combined), ".end\n") {
		t.Fatal(string(combined), err)
	}
}

func TestEspressoOutputErrorsLimitsAndCancellation(t *testing.T) {
	pla, cubes := espressoOutputFixture(t)
	checks := []struct {
		options EspressoPLAOutputOptions
		want    error
	}{
		{EspressoPLAOutputOptions{MaxCubes: -1}, ErrDimension},
		{EspressoPLAOutputOptions{Phase: []bool{true}}, ErrDimension},
		{EspressoPLAOutputOptions{MaxCubes: 1}, ErrLimit},
		{EspressoPLAOutputOptions{MaxParts: 1}, ErrLimit},
		{EspressoPLAOutputOptions{MaxBytes: 1}, ErrLimit},
		{EspressoPLAOutputOptions{Stop: func() bool { return true }}, ErrCancelled},
	}
	for _, check := range checks {
		if raw, err := pla.MarshalPLA(EspressoPLAOutputOn|EspressoPLAOutputOff, check.options); raw != nil || !errors.Is(err, check.want) {
			t.Fatal(check.options, string(raw), err, check.want)
		}
	}
	if value, err := pla.Layout.FormatCube(cubes[0], "x", EspressoPLAOutputOptions{}); value != "" || !errors.Is(err, ErrDimension) {
		t.Fatal(value, err)
	}
	if raw, err := pla.MarshalPLA(0, EspressoPLAOutputOptions{}); raw != nil || !errors.Is(err, ErrDimension) {
		t.Fatal(raw, err)
	}
	if raw, err := pla.MarshalPLA(EspressoPLAOutputPleasure|EspressoPLAOutputEquation, EspressoPLAOutputOptions{}); raw != nil || !errors.Is(err, ErrDimension) {
		t.Fatal(raw, err)
	}
	if value, err := pla.Layout.FormatCoverDebug(pla.On, "bad\nname", 0, false, EspressoPLAOutputOptions{}); value != "" || !errors.Is(err, ErrDimension) {
		t.Fatal(value, err)
	}
	if value, err := pla.Layout.FormatCoverDebug(pla.On, "F", 0, false, EspressoPLAOutputOptions{MaxBytes: 1}); value != "" || !errors.Is(err, ErrLimit) {
		t.Fatal(value, err)
	}
	if raw, err := pla.MarshalPLA(EspressoPLAOutputEquation, EspressoPLAOutputOptions{}); raw != nil || !errors.Is(err, ErrDimension) {
		t.Fatal(raw, err)
	}
	noOutputLayout, _ := NewEspressoCubeLayout([]int{2}, 1)
	noOutputFamily := espressoOutputFamily(t, 2, espressoOutputCube(t, 2, 0, 1))
	noOutputPLA, _ := NewEspressoPLA(noOutputLayout, noOutputFamily, espressoOutputFamily(t, 2), espressoOutputFamily(t, 2), nil)
	if raw, err := noOutputPLA.MarshalPLA(EspressoPLAOutputPleasure, EspressoPLAOutputOptions{}); raw != nil || !errors.Is(err, ErrDimension) {
		t.Fatal(raw, err)
	}
}

func espressoReferenceCube(layout *EspressoCubeLayout, cube *EspressoSet, outputMap string) string {
	var result strings.Builder
	for variable := 0; variable < layout.binaryVariables; variable++ {
		first := layout.firstParts[variable]
		low, _ := cube.Has(first)
		high, _ := cube.Has(first + 1)
		result.WriteByte("?01-"[btoi(low)+2*btoi(high)])
	}
	output := layout.OutputVariable()
	for variable := layout.binaryVariables; variable < len(layout.partSizes); variable++ {
		if variable == output {
			continue
		}
		result.WriteByte(' ')
		for part := 0; part < layout.partSizes[variable]; part++ {
			present, _ := cube.Has(layout.firstParts[variable] + part)
			result.WriteByte("01"[btoi(present)])
		}
	}
	if output >= 0 {
		result.WriteByte(' ')
		for part := 0; part < layout.partSizes[output]; part++ {
			present, _ := cube.Has(layout.firstParts[output] + part)
			result.WriteByte(outputMap[btoi(present)])
		}
	}
	return result.String()
}

func FuzzEspressoOutputCube(f *testing.F) {
	f.Add([]byte{2, 3, 2, 0x55, 0x01})
	f.Add([]byte{0, 2, 4, 0xff})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) < 4 {
			return
		}
		binary := int(data[0] % 4)
		mv := int(data[1]%3) + 1
		parts := make([]int, binary+2)
		for index := 0; index < binary; index++ {
			parts[index] = 2
		}
		parts[binary], parts[binary+1] = mv, int(data[2]%4)+1
		layout, err := NewEspressoCubeLayout(parts, binary)
		if err != nil {
			t.Fatal(err)
		}
		cube, _ := NewEspressoSet(layout.bitCount)
		for part := 0; part < layout.bitCount; part++ {
			if data[(3+part)%len(data)]&(1<<uint(part&7)) != 0 {
				_ = cube.Set(part, true)
			}
		}
		mapping := "~2"
		got, err := layout.FormatCube(cube, mapping, EspressoPLAOutputOptions{MaxParts: 1 << 12, MaxBytes: 1 << 12})
		want := espressoReferenceCube(layout, cube, mapping)
		if err != nil || got != want {
			t.Fatal(parts, got, want, err)
		}
	})
}

func FuzzEspressoPLAMarshal(f *testing.F) {
	f.Add([]byte{2, 3, 2, 0x55, 0xaa, 0x0f})
	f.Add([]byte{1, 1, 4, 0xff, 0x00})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) < 5 {
			return
		}
		binary := int(data[0]%4) + 1
		mvSize := int(data[1]%4) + 1
		outputSize := int(data[2]%4) + 1
		partSizes := make([]int, binary+2)
		for variable := 0; variable < binary; variable++ {
			partSizes[variable] = 2
		}
		partSizes[binary], partSizes[binary+1] = mvSize, outputSize
		layout, err := NewEspressoCubeLayout(partSizes, binary)
		if err != nil {
			t.Fatal(err)
		}
		cubeCount := int(data[3]%4) + 1
		cubes := make([]*EspressoSet, cubeCount)
		for cubeIndex := range cubes {
			cube, _ := NewEspressoSet(layout.bitCount)
			for part := 0; part < layout.bitCount; part++ {
				position := (4 + cubeIndex*layout.bitCount + part) % len(data)
				if data[position]&(1<<uint((cubeIndex+part)&7)) != 0 {
					_ = cube.Set(part, true)
				}
			}
			cubes[cubeIndex] = cube
		}
		on := espressoOutputFamily(t, layout.bitCount, cubes...)
		empty := espressoOutputFamily(t, layout.bitCount)
		pla, err := NewEspressoPLA(layout, on, empty, empty, nil)
		if err != nil {
			t.Fatal(err)
		}
		raw, err := pla.MarshalPLA(EspressoPLAOutputOn, EspressoPLAOutputOptions{MaxCubes: 16, MaxParts: 1 << 12, MaxBytes: 1 << 14})
		if err != nil {
			t.Fatal(err)
		}
		lines := strings.Split(strings.TrimSuffix(string(raw), "\n"), "\n")
		productLine := -1
		for index, line := range lines {
			if strings.HasPrefix(line, ".p ") {
				productLine = index
				if line != ".p "+strconv.Itoa(cubeCount) {
					t.Fatal(line, cubeCount)
				}
				break
			}
		}
		if productLine < 0 || len(lines) != productLine+cubeCount+2 || lines[len(lines)-1] != ".e" {
			t.Fatal(string(raw), productLine, cubeCount)
		}
		for cubeIndex, cube := range cubes {
			fields := strings.Fields(lines[productLine+1+cubeIndex])
			if len(fields) != 3 || len(fields[0]) != binary || len(fields[1]) != mvSize || len(fields[2]) != outputSize {
				t.Fatal(fields, partSizes)
			}
			decoded := make([]bool, layout.bitCount)
			for variable, value := range []byte(fields[0]) {
				switch value {
				case '0':
					decoded[layout.firstParts[variable]] = true
				case '1':
					decoded[layout.firstParts[variable]+1] = true
				case '-':
					decoded[layout.firstParts[variable]], decoded[layout.firstParts[variable]+1] = true, true
				case '?':
				default:
					t.Fatal(value)
				}
			}
			for offset, field := range fields[1:] {
				variable := binary + offset
				for part, value := range []byte(field) {
					if value != '0' && value != '1' {
						t.Fatal(value)
					}
					decoded[layout.firstParts[variable]+part] = value == '1'
				}
			}
			for part, want := range decoded {
				got, _ := cube.Has(part)
				if got != want {
					t.Fatal(partSizes, cubeIndex, part, got, want, string(raw))
				}
			}
		}
	})
}
