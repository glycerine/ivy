package abc

// Go-only validation for src/bdd/cas/cas.h, casCore.c, and casDec.c.

import (
	"errors"
	"slices"
	"strconv"
	"strings"
	"testing"
)

func casTruth(t *testing.T, variables int, bits uint64) *TruthTable {
	t.Helper()
	table, err := NewConstantTruthTable(variables, false, 1<<16)
	if err != nil {
		t.Fatal(err)
	}
	for assignment := 0; assignment < 1<<uint(variables); assignment++ {
		if bits&(uint64(1)<<uint(assignment)) != 0 {
			if err = table.Set(uint64(assignment), true); err != nil {
				t.Fatal(err)
			}
		}
	}
	return table
}

func evalCascadeBLIF(data []byte, inputs map[string]bool) (bool, error) {
	values := make(map[string]bool, len(inputs)+32)
	for name, value := range inputs {
		values[name] = value
	}
	lines := strings.Split(string(data), "\n")
	for line := 0; line < len(lines); {
		fields := strings.Fields(lines[line])
		if len(fields) == 0 || fields[0] != ".names" {
			line++
			continue
		}
		if len(fields) < 2 {
			return false, ErrFormat
		}
		output := fields[len(fields)-1]
		fanins := fields[1 : len(fields)-1]
		value := false
		line++
		for line < len(lines) {
			row := strings.Fields(lines[line])
			if len(row) == 0 {
				line++
				continue
			}
			if strings.HasPrefix(row[0], ".") || strings.HasPrefix(row[0], "#") {
				break
			}
			if len(fanins) == 0 {
				value = row[0] == "1"
				line++
				continue
			}
			if len(row) != 2 || len(row[0]) != len(fanins) || row[1] != "1" {
				return false, ErrFormat
			}
			matches := true
			for index, symbol := range row[0] {
				if symbol == '-' {
					continue
				}
				if (symbol == '1') != values[fanins[index]] {
					matches = false
					break
				}
			}
			value = value || matches
			line++
		}
		values[output] = value
	}
	return values["F"], nil
}

func TestCascadeGraySupportAndEncoding(t *testing.T) {
	for value := uint32(0); value < 1024; value++ {
		if CascadeGrayCode(value) != value^(value>>1) {
			t.Fatal(value)
		}
	}
	x, _ := NewVariableTruthTable(3, 0, 1)
	y, _ := NewVariableTruthTable(3, 1, 1)
	xy, _ := x.And(y, false, false)
	order, err := CascadeSupportOrder([]*TruthTable{x, xy, y})
	if err != nil || !slices.Equal(order, []int{1, 0, 2}) {
		t.Fatal(order, err)
	}
	mintermOrder, err := CascadeMintermOrder([]int{3, 7, 7, 1})
	if err != nil || !slices.Equal(mintermOrder, []int{1, 2, 0, 3}) {
		t.Fatal(mintermOrder, err)
	}
	encoded, err := EncodeCascadeOutputs([]*TruthTable{x, xy, y}, CascadeOptions{})
	if err != nil || encoded.InputCount != 3 || encoded.EncodingBits != 2 || !slices.Equal(encoded.OutputOrder, order) {
		t.Fatal(encoded, err)
	}
	for input := uint64(0); input < 8; input++ {
		for code, output := range order {
			selector := cascadeEncodingAssignment(code, encoded.EncodingBits)
			got, getErr := encoded.Relation.Get(input | selector<<3)
			want, _ := []*TruthTable{x, xy, y}[output].Get(input)
			if getErr != nil || got != want {
				t.Fatalf("input=%d code=%d got=%v want=%v err=%v", input, code, got, want, getErr)
			}
		}
		unused, _ := encoded.Relation.Get(input | 3<<3)
		if unused {
			t.Fatal("unused encoding is not zero")
		}
	}
	clone := encoded.Clone()
	clone.OutputOrder[0] = 99
	if encoded.OutputOrder[0] == 99 {
		t.Fatal("clone aliases order")
	}
}

func TestCascadeRemapAndTransfer(t *testing.T) {
	table := casTruth(t, 4, 0xaaaa)
	wide := casTruth(t, 4, 0x8888)
	remapped, err := EncodeRemappedCascadeOutputs([]*TruthTable{table, wide}, CascadeOptions{})
	if err != nil || remapped.Encoded.InputCount != 2 || !slices.Equal(remapped.SupportMaps[0], []int{0}) || !slices.Equal(remapped.SupportMaps[1], []int{0, 1}) || !slices.Equal(remapped.Encoded.OutputOrder, []int{0, 1}) {
		t.Fatal(remapped, err)
	}
	// Unlike GetSingleOutputFunction, the remapped C entry point does not
	// support-sort its output array before binary encoding.
	for assignment := uint64(0); assignment < 4; assignment++ {
		first, _ := remapped.Encoded.Relation.Get(assignment)
		second, _ := remapped.Encoded.Relation.Get(assignment | 1<<2)
		wantFirst, _ := table.Get(assignment & 1)
		wantSecond, _ := wide.Get(assignment)
		if first != wantFirst || second != wantSecond {
			t.Fatalf("assignment=%d got=%v/%v want=%v/%v", assignment, first, second, wantFirst, wantSecond)
		}
	}
	transferred, err := TransferCascadeTruth(casTruth(t, 3, 0xca), []int{2, 0, 1})
	if err != nil {
		t.Fatal(err)
	}
	for source := uint64(0); source < 8; source++ {
		var target uint64
		permutation := []int{2, 0, 1}
		for bit, position := range permutation {
			if source&(1<<uint(bit)) != 0 {
				target |= 1 << uint(position)
			}
		}
		want, _ := casTruth(t, 3, 0xca).Get(source)
		got, _ := transferred.Get(target)
		if got != want {
			t.Fatalf("%d -> %d", source, target)
		}
	}
}

func TestCascadeNetworkExhaustive(t *testing.T) {
	for variables := 1; variables <= 4; variables++ {
		limit := 1 << uint(1<<uint(variables))
		if variables == 4 {
			limit = 512
		}
		for bits := 0; bits < limit; bits++ {
			first := casTruth(t, variables, uint64(bits))
			second := first.Not()
			network, err := BuildCascade([]*TruthTable{first, second}, 3, CascadeOptions{MaxVariables: 8, MaxEntries: 1 << 16})
			if errors.Is(err, ErrUnsupported) {
				// As in CreateDecomposedNetwork, a cut whose state code consumes
				// the whole LUT cannot admit another primary input.
				continue
			}
			if err != nil {
				t.Fatalf("vars=%d bits=%x: %v", variables, bits, err)
			}
			stages, stageErr := network.Stages()
			if stageErr != nil || len(stages) == 0 {
				t.Fatal(stageErr)
			}
			for _, stage := range stages {
				if stage.PreviousBits+len(stage.InputVariables) > 3 {
					t.Fatalf("oversized stage: %+v", stage)
				}
			}
			for assignment := uint64(0); assignment < uint64(1)<<uint(variables); assignment++ {
				got, evalErr := network.Evaluate(assignment)
				want, _ := first.Get(assignment)
				if evalErr != nil || len(got) != 2 || got[0] != want || got[1] == want {
					t.Fatalf("vars=%d bits=%x assignment=%d got=%v want=%v err=%v", variables, bits, assignment, got, want, evalErr)
				}
			}
			owned, _ := network.Stages()
			owned[0].InputVariables = append(owned[0].InputVariables, 99)
			again, _ := network.Stages()
			if slices.Contains(again[0].InputVariables, 99) {
				t.Fatal("stage snapshot aliases")
			}
			network.Close()
			network.Close()
			if _, evalErr := network.Evaluate(0); !errors.Is(evalErr, ErrInvariant) {
				t.Fatal(evalErr)
			}
		}
	}
}

func TestCascadeMultipleStagesAndOwnership(t *testing.T) {
	parity := casTruth(t, 6, 0x6996966996696996)
	network, err := BuildCascade([]*TruthTable{parity}, 3, CascadeOptions{})
	if err != nil {
		t.Fatal(err)
	}
	stages, err := network.Stages()
	if err != nil || len(stages) < 2 {
		t.Fatalf("stages=%d err=%v", len(stages), err)
	}
	// Construction snapshots the caller's functions, and all observations are
	// snapshots too.
	if err = parity.Set(0, true); err != nil {
		t.Fatal(err)
	}
	got, err := network.Evaluate(0)
	if err != nil || len(got) != 1 || got[0] {
		t.Fatal(got, err)
	}
	relation, err := network.Relation()
	if err != nil || relation.Set(0, true) != nil {
		t.Fatal(err)
	}
	again, err := network.Evaluate(0)
	if err != nil || again[0] {
		t.Fatal(again, err)
	}
	order, _ := network.OutputOrder()
	order[0] = 99
	againOrder, _ := network.OutputOrder()
	if againOrder[0] == 99 {
		t.Fatal("output-order snapshot aliases network")
	}
	blif, err := MarshalCascadeNetworkBLIF(network, []string{"a", "b", "c", "d", "e", "f"}, "parity_cascade", 1<<16)
	if err != nil {
		t.Fatal(err)
	}
	text := string(blif)
	for _, fragment := range []string{".model parity_cascade", "#----------------- LUT #0", "#----------------- LUT #1", "LUT00_00", ".outputs F", ".end\n"} {
		if !strings.Contains(text, fragment) {
			t.Fatalf("cascade BLIF misses %q", fragment)
		}
	}
	againBLIF, err := MarshalCascadeNetworkBLIF(network, []string{"a", "b", "c", "d", "e", "f"}, "parity_cascade", len(blif))
	if err != nil || !slices.Equal(blif, againBLIF) {
		t.Fatal(err)
	}
	if result, err := MarshalCascadeNetworkBLIF(network, []string{"a", "b", "c", "d", "e", "f"}, "parity_cascade", len(blif)-1); result != nil || !errors.Is(err, ErrLimit) {
		t.Fatal(result, err)
	}
	for assignment := uint64(0); assignment < 64; assignment++ {
		inputs := make(map[string]bool, 6)
		for index, name := range []string{"a", "b", "c", "d", "e", "f"} {
			inputs[name] = assignment&(1<<uint(index)) != 0
		}
		got, evalErr := evalCascadeBLIF(blif, inputs)
		want, _ := network.Evaluate(assignment)
		if evalErr != nil || got != want[0] {
			t.Fatalf("assignment=%d got=%v want=%v err=%v", assignment, got, want, evalErr)
		}
	}
}

func TestCascadeBLIFDeterministicAndBounded(t *testing.T) {
	table := casTruth(t, 3, 0x96)
	data, err := MarshalCascadeTruthBLIF(table, []string{"a", "b", "c"}, "xor3", "F", 4096)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	for _, fragment := range []string{".model xor3", ".inputs a b c", ".outputs F", "01- 1", "1-1 1", ".end\n"} {
		if !strings.Contains(text, fragment) {
			t.Fatalf("missing %q in %q", fragment, text)
		}
	}
	again, err := MarshalCascadeTruthBLIF(table, []string{"a", "b", "c"}, "xor3", "F", len(data))
	if err != nil || !slices.Equal(data, again) {
		t.Fatal(err)
	}
	if result, boundedErr := MarshalCascadeTruthBLIF(table, []string{"a", "b", "c"}, "xor3", "F", len(data)-1); result != nil || !errors.Is(boundedErr, ErrLimit) {
		t.Fatal(boundedErr)
	}
	if _, badErr := MarshalCascadeTruthBLIF(table, []string{"a", "bad name", "c"}, "m", "F", 1000); !errors.Is(badErr, ErrFormat) {
		t.Fatal(badErr)
	}
	if _, badErr := MarshalCascadeTruthBLIF(table, []string{"a", "b", "c"}, "bad\nmodel", "F", 1000); !errors.Is(badErr, ErrFormat) {
		t.Fatal(badErr)
	}
	for assignment := uint64(0); assignment < 8; assignment++ {
		got, evalErr := evalCascadeBLIF(data, map[string]bool{"a": assignment&1 != 0, "b": assignment&2 != 0, "c": assignment&4 != 0})
		want, _ := table.Get(assignment)
		if evalErr != nil || got != want {
			t.Fatalf("assignment=%d got=%v want=%v err=%v", assignment, got, want, evalErr)
		}
	}
}

func TestCascadeEncodedNetworkBLIF(t *testing.T) {
	first := casTruth(t, 4, 0xa35c)
	second := casTruth(t, 4, 0x5aa5)
	third := casTruth(t, 4, 0xf00f)
	outputs := []*TruthTable{first, second, third}
	network, err := BuildCascade(outputs, 4, CascadeOptions{MaxVariables: 8, MaxEntries: 1 << 18})
	if err != nil {
		t.Fatal(err)
	}
	defer network.Close()
	order, err := network.OutputOrder()
	if err != nil {
		t.Fatal(err)
	}
	data, err := MarshalCascadeNetworkBLIF(network, []string{"a", "b", "c", "d", "sel0", "sel1"}, "encoded", 1<<18)
	if err != nil {
		t.Fatal(err)
	}
	for assignment := uint64(0); assignment < 16; assignment++ {
		decoded, evalErr := network.Evaluate(assignment)
		if evalErr != nil {
			t.Fatal(evalErr)
		}
		for code := 0; code < 4; code++ {
			inputs := map[string]bool{
				"a": assignment&1 != 0, "b": assignment&2 != 0,
				"c": assignment&4 != 0, "d": assignment&8 != 0,
				"sel0": code&2 != 0, "sel1": code&1 != 0,
			}
			got, blifErr := evalCascadeBLIF(data, inputs)
			want := false
			if code < len(order) {
				want = decoded[order[code]]
			}
			if blifErr != nil || got != want {
				t.Fatalf("assignment=%d code=%d got=%v want=%v err=%v", assignment, code, got, want, blifErr)
			}
		}
	}
}

func TestCascadeErrorsLimitsAndCancellation(t *testing.T) {
	table := casTruth(t, 3, 0x96)
	for _, test := range []struct {
		outputs []*TruthTable
		lut     int
		options CascadeOptions
		want    error
	}{
		{nil, 3, CascadeOptions{}, ErrDimension},
		{[]*TruthTable{table}, 1, CascadeOptions{}, ErrDimension},
		{[]*TruthTable{table}, 3, CascadeOptions{MaxVariables: 2}, ErrLimit},
		{[]*TruthTable{table}, 3, CascadeOptions{MaxEntries: 2}, ErrLimit},
		{[]*TruthTable{table}, 3, CascadeOptions{Stop: func() bool { return true }}, ErrCancelled},
	} {
		result, err := BuildCascade(test.outputs, test.lut, test.options)
		if result != nil || !errors.Is(err, test.want) {
			t.Fatalf("result=%v err=%v want=%v", result, err, test.want)
		}
	}
	if _, err := TransferCascadeTruth(table, []int{0, 0, 2}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	calls := 0
	if result, err := BuildCascade([]*TruthTable{table}, 2, CascadeOptions{Stop: func() bool { calls++; return calls > 4 }}); result != nil || !errors.Is(err, ErrCancelled) {
		t.Fatal(result, err, calls)
	}
}

func TestCascadeIndependentModelCampaign(t *testing.T) {
	state := uint64(0x9e3779b97f4a7c15)
	for execution := 0; execution < 5000; execution++ {
		state ^= state << 7
		state ^= state >> 9
		state ^= state << 8
		variables := 1 + int(state%5)
		mask := uint64(1)<<uint(1<<uint(variables)) - 1
		first := casTruth(t, variables, state&mask)
		state = state*6364136223846793005 + 1
		second := casTruth(t, variables, state&mask)
		network, err := BuildCascade([]*TruthTable{first, second}, variables+1, CascadeOptions{MaxVariables: 8, MaxEntries: 1 << 18})
		if err != nil {
			t.Fatalf("execution %d: %v", execution, err)
		}
		for assignment := uint64(0); assignment < uint64(1)<<uint(variables); assignment++ {
			got, evalErr := network.Evaluate(assignment)
			want0, _ := first.Get(assignment)
			want1, _ := second.Get(assignment)
			if evalErr != nil || !slices.Equal(got, []bool{want0, want1}) {
				t.Fatalf("execution=%d assignment=%d got=%v want=%v/%v err=%v", execution, assignment, got, want0, want1, evalErr)
			}
		}
	}
}

func FuzzCascadeModel(f *testing.F) {
	f.Add(uint64(0x96), uint64(0x69), uint8(3), uint8(3))
	f.Add(uint64(0xaaaa), uint64(0xcccc), uint8(4), uint8(4))
	f.Fuzz(func(t *testing.T, firstBits, secondBits uint64, rawVariables, rawLUT uint8) {
		variables := 1 + int(rawVariables%5)
		lut := 2 + int(rawLUT%4)
		mask := uint64(1)<<uint(1<<uint(variables)) - 1
		first := casTruth(t, variables, firstBits&mask)
		second := casTruth(t, variables, secondBits&mask)
		network, err := BuildCascade([]*TruthTable{first, second}, lut, CascadeOptions{MaxVariables: 8, MaxEntries: 1 << 18})
		if errors.Is(err, ErrUnsupported) {
			return
		}
		if err != nil {
			t.Fatal(err)
		}
		for assignment := uint64(0); assignment < uint64(1)<<uint(variables); assignment++ {
			got, evalErr := network.Evaluate(assignment)
			want0, _ := first.Get(assignment)
			want1, _ := second.Get(assignment)
			if evalErr != nil || !slices.Equal(got, []bool{want0, want1}) {
				t.Fatalf("assignment=%d got=%v want=%v/%v err=%v", assignment, got, want0, want1, evalErr)
			}
		}
		names := make([]string, variables+1)
		for index := 0; index < variables; index++ {
			names[index] = "x" + strconv.Itoa(index)
		}
		names[variables] = "selector"
		data, marshalErr := MarshalCascadeNetworkBLIF(network, names, "fuzz_cascade", 1<<18)
		if marshalErr != nil {
			t.Fatal(marshalErr)
		}
		order, orderErr := network.OutputOrder()
		if orderErr != nil {
			t.Fatal(orderErr)
		}
		for assignment := uint64(0); assignment < uint64(1)<<uint(variables); assignment++ {
			for code := range order {
				inputs := make(map[string]bool, len(names))
				for variable := 0; variable < variables; variable++ {
					inputs[names[variable]] = assignment&(1<<uint(variable)) != 0
				}
				inputs["selector"] = code != 0
				got, evalErr := evalCascadeBLIF(data, inputs)
				want, _ := []*TruthTable{first, second}[order[code]].Get(assignment)
				if evalErr != nil || got != want {
					t.Fatalf("BLIF assignment=%d code=%d got=%v want=%v err=%v", assignment, code, got, want, evalErr)
				}
			}
		}
	})
}
