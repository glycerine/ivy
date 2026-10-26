package abc

// Go-only semantic, atomicity and fuzz checks for DSD-to-GIA lowering.

import (
	"errors"
	"math/bits"
	"slices"
	"testing"
)

func dauGiaInputs(t *testing.T, count int, hashing bool) (*GiaManager, []GiaLit) {
	t.Helper()
	graph := NewGiaManager()
	inputs := make([]GiaLit, count)
	for i := range inputs {
		var err error
		inputs[i], err = graph.AppendCI()
		if err != nil {
			t.Fatal(err)
		}
	}
	if hashing {
		if err := graph.HashStart(); err != nil {
			t.Fatal(err)
		}
	}
	return graph, inputs
}

func checkDauGiaFormula(t *testing.T, formula string, options DauGiaOptions, prehashed bool) {
	t.Helper()
	parsed, err := ParseDauDSD(formula, DauDSDOptions{})
	if err != nil {
		t.Fatal(err)
	}
	graph, inputs := dauGiaInputs(t, parsed.VariableCount(), prehashed)
	before := graph.ObjectCount()
	result, err := graph.WithDauDSD(formula, inputs, options)
	if err != nil {
		t.Fatalf("%q: %v", formula, err)
	}
	if result.Graph == graph || graph.ObjectCount() != before || result.Graph.Hashing() != prehashed {
		t.Fatalf("%q ownership/hash state", formula)
	}
	truth, err := parsed.TruthTable(parsed.VariableCount(), 0, nil)
	if err != nil {
		t.Fatal(err)
	}
	assignments := uint64(1) << uint(parsed.VariableCount())
	for assignment := uint64(0); assignment < assignments; assignment++ {
		values := make([]bool, parsed.VariableCount())
		for i := range values {
			values[i] = assignment>>uint(i)&1 != 0
		}
		actual, err := result.Graph.evaluateValues(values)
		if err != nil {
			t.Fatal(err)
		}
		got := actual[result.Root.ID()] != result.Root.Complemented()
		want, _ := truth.Get(assignment)
		if got != want {
			t.Fatalf("%q assignment %b: got %t want %t", formula, assignment, got, want)
		}
	}
}

func TestDauDSDToGiaSemantics(t *testing.T) {
	formulas := []string{"0", "1", "a", "!a", "(abcd)", "[abcd]", "!<abc>", "96{abc}", "<abc>{d!e[fg]}"}
	for _, formula := range formulas {
		for mode := 0; mode < 8; mode++ {
			t.Run(formula+string(rune('0'+mode)), func(t *testing.T) {
				checkDauGiaFormula(t, formula, DauGiaOptions{NativeGates: mode&1 != 0, Balance: mode&2 != 0}, mode&4 != 0)
			})
		}
	}
	graph, inputs := dauGiaInputs(t, 3, false)
	result, err := graph.WithDauDSD("96{abc}", inputs, DauGiaOptions{})
	if err != nil || result.PrimeCount != 1 {
		t.Fatalf("prime count=%d, %v", result.PrimeCount, err)
	}
}

func TestDauDSDToGiaErrorsAtomic(t *testing.T) {
	graph, inputs := dauGiaInputs(t, 4, false)
	before := graph.ObjectCount()
	tests := []struct {
		formula string
		inputs  []GiaLit
		options DauGiaOptions
		want    error
	}{
		{"(abcd)", inputs[:3], DauGiaOptions{}, ErrDimension},
		{"(abcd)", []GiaLit{inputs[0], inputs[1], inputs[2], InvalidGiaLit}, DauGiaOptions{}, ErrInvalidLiteral},
		{"(abcd)", inputs, DauGiaOptions{MaxObjects: before}, ErrLimit},
		{"(abcd)", inputs, DauGiaOptions{MaxObjects: -1}, ErrDimension},
		{"(abcd)", inputs, DauGiaOptions{Stop: func() bool { return true }}, ErrCancelled},
		{"(abc", inputs[:3], DauGiaOptions{}, ErrInvariant},
	}
	for _, test := range tests {
		result, err := graph.WithDauDSD(test.formula, test.inputs, test.options)
		if result.Graph != nil || result.Root != 0 || !errors.Is(err, test.want) {
			t.Errorf("%q: %+v, %v", test.formula, result, err)
		}
		if graph.ObjectCount() != before {
			t.Fatal("source graph mutated on error")
		}
	}
}

func TestDauTruthToGiaAndStatistics(t *testing.T) {
	truth := dauDSDTestTable(t, 3, func(assignment uint64) bool {
		return bits.OnesCount64(assignment) >= 2
	})
	graph, inputs := dauGiaInputs(t, 3, false)
	result, err := graph.WithDauTruth(truth, inputs, DauGiaOptions{Balance: true, NativeGates: true})
	if err != nil || result.LargestPrime != 3 || result.PrimeCount != 1 {
		t.Fatal(result, err)
	}
	for assignment := uint64(0); assignment < 8; assignment++ {
		values := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0}
		objects, err := result.Graph.evaluateValues(values)
		if err != nil {
			t.Fatal(err)
		}
		got := objects[result.Root.ID()] != result.Root.Complemented()
		want, _ := truth.Get(assignment)
		if got != want {
			t.Fatal(assignment, got, want)
		}
	}
	constant := dauDSDTestTable(t, 3, func(uint64) bool { return true })
	if lowered, err := graph.WithDauTruth(constant, inputs, DauGiaOptions{}); err != nil || lowered.Root != GiaConst1 {
		t.Fatal(lowered, err)
	}
	stats := DauGiaStatistics{Calls: 4, NonDSD: 2, PrimeBlocks: 3}
	if text := FormatDauGiaStatistics(stats); text != "Calls = 4. NonDSD = 2. Non1Step = 3.\n" {
		t.Fatal(text)
	}
}

func TestDeriveDauGiaMappedNetwork(t *testing.T) {
	source := NewGiaManager()
	source.Name, source.Spec = "mapped", "mapped.aig"
	a, _ := source.AppendCI()
	b, _ := source.AppendCI()
	c, _ := source.AppendCI()
	x, _ := source.AppendAnd(a, b)
	y, _ := source.AppendAnd(x, c.Not())
	buffer, _ := source.AppendBuffer(y.Not())
	source.AppendCO(y)
	source.AppendCO(buffer)
	if err := source.SetRegisterCount(1); err != nil {
		t.Fatal(err)
	}
	if err := source.SetConstraintCount(1); err != nil {
		t.Fatal(err)
	}
	mapping, err := NewGiaLUTMapping(source.ObjectCount(), []GiaLUTRecord{
		{Root: x.ID(), Fanins: []GiaID{a.ID(), b.ID()}},
		{Root: y.ID(), Fanins: []GiaID{x.ID(), c.ID()}},
	}, 0)
	if err != nil {
		t.Fatal(err)
	}
	before := slices.Clone(source.objects)
	result, err := source.DeriveDauGia(mapping, DauGiaDeriveOptions{DauGiaOptions: DauGiaOptions{Balance: true, NativeGates: true}})
	if err != nil {
		t.Fatal(err)
	}
	if result.Graph == source || result.Graph.Name != source.Name || result.Graph.Spec != source.Spec ||
		result.Graph.RegisterCount() != 1 || result.Graph.ConstraintCount() != 1 ||
		result.Statistics != (DauGiaStatistics{Calls: 2}) || !slices.Equal(before, source.objects) {
		t.Fatal(result.Statistics)
	}
	for assignment := 0; assignment < 8; assignment++ {
		inputs := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0}
		want, err := source.Evaluate(inputs)
		if err != nil {
			t.Fatal(err)
		}
		got, err := result.Graph.Evaluate(inputs)
		if err != nil || !slices.Equal(got, want) {
			t.Fatal(assignment, got, want, err)
		}
	}
	for _, options := range []DauGiaDeriveOptions{
		{MaxMappingEntries: 1},
		{DauGiaOptions: DauGiaOptions{MaxObjects: 1}},
		{DauGiaOptions: DauGiaOptions{Stop: func() bool { return true }}},
	} {
		if failed, err := source.DeriveDauGia(mapping, options); failed.Graph != nil || err == nil {
			t.Fatal(failed, err)
		}
	}
}

func FuzzDauDSDToGia(f *testing.F) {
	for _, seed := range [][]byte{{0, 1}, {1, 2, 3}, {2, 3, 4, 5}, {3, 9, 11}} {
		f.Add(seed, false, false)
	}
	f.Fuzz(func(t *testing.T, data []byte, native, balance bool) {
		formula := fuzzDauFormula(data)
		parsed, err := ParseDauDSD(formula, DauDSDOptions{})
		if err != nil || parsed.VariableCount() > 6 {
			return
		}
		checkDauGiaFormula(t, formula, DauGiaOptions{NativeGates: native, Balance: balance, MaxObjects: 10000}, false)
	})
}

func FuzzDauTruthToGia(f *testing.F) {
	f.Add(uint8(3), []byte{0xe8})
	f.Add(uint8(6), []byte{0x96, 0x3c, 0xa5, 0x5a, 0xc3, 0x69, 0xf0, 0x0f})
	f.Fuzz(func(t *testing.T, width uint8, data []byte) {
		variables := 1 + int(width%6)
		if len(data) == 0 {
			return
		}
		truth := dauDSDTestTable(t, variables, func(assignment uint64) bool {
			bit := assignment % uint64(len(data)*8)
			return data[bit/8]&(1<<uint(bit&7)) != 0
		})
		graph, inputs := dauGiaInputs(t, variables, false)
		result, err := graph.WithDauTruth(truth, inputs, DauGiaOptions{MaxObjects: 10000})
		if err != nil {
			t.Fatal(err)
		}
		for assignment := uint64(0); assignment < 1<<uint(variables); assignment++ {
			values := make([]bool, variables)
			for variable := range values {
				values[variable] = assignment&(1<<uint(variable)) != 0
			}
			objects, err := result.Graph.evaluateValues(values)
			want, truthErr := truth.Get(assignment)
			got := objects[result.Root.ID()] != result.Root.Complemented()
			if err != nil || truthErr != nil || got != want {
				t.Fatal(assignment, got, want, err, truthErr)
			}
		}
	})
}

func FuzzDeriveDauGiaMappedNetwork(f *testing.F) {
	f.Add([]byte{0, 1, 2, 3, 4, 5})
	f.Add([]byte{9, 7, 5, 3, 1})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) == 0 {
			return
		}
		source := NewGiaManager()
		pool := make([]GiaLit, 3)
		for index := range pool {
			pool[index], _ = source.AppendCI()
		}
		records := make([]GiaLUTRecord, 0, 8)
		gates := 1 + int(data[0]%8)
		for gate := 0; gate < gates; gate++ {
			first := int(data[(2*gate+1)%len(data)]) % len(pool)
			second := int(data[(2*gate+2)%len(data)]) % len(pool)
			if second == first {
				second = (second + 1) % len(pool)
			}
			left := pool[first].NotIf(data[(3*gate+1)%len(data)]&1 != 0)
			right := pool[second].NotIf(data[(3*gate+2)%len(data)]&1 != 0)
			root, err := source.AppendAnd(left, right)
			if err != nil {
				t.Fatal(err)
			}
			records = append(records, GiaLUTRecord{Root: root.ID(), Fanins: []GiaID{left.ID(), right.ID()}})
			pool = append(pool, root)
		}
		root := pool[len(pool)-1].NotIf(data[len(data)-1]&1 != 0)
		source.AppendCO(root)
		mapping, err := NewGiaLUTMapping(source.ObjectCount(), records, 0)
		if err != nil {
			t.Fatal(err)
		}
		result, err := source.DeriveDauGia(mapping, DauGiaDeriveOptions{DauGiaOptions: DauGiaOptions{MaxObjects: 10000}})
		if err != nil || result.Statistics.Calls != gates {
			t.Fatal(result.Statistics, err)
		}
		for assignment := 0; assignment < 8; assignment++ {
			inputs := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0}
			want, sourceErr := source.Evaluate(inputs)
			got, resultErr := result.Graph.Evaluate(inputs)
			if sourceErr != nil || resultErr != nil || !slices.Equal(got, want) {
				t.Fatal(assignment, got, want, sourceErr, resultErr)
			}
		}
	})
}
