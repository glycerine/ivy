package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func TestGiaTruthBridgeNative(t *testing.T) {
	data, e := os.ReadFile("test_vectors/oracle/gia_truth.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		Variables, Mode, Hash, Width int
		Truth, Simple, Mapped        uint64
		Root                         GiaLit
		Inputs                       []GiaLit
		Objects                      [][3]int
	}
	if e = json.Unmarshal(data, &rows); e != nil {
		t.Fatal(e)
	}
	if len(rows) != 624 {
		t.Fatal(len(rows))
	}
	for index, row := range rows {
		source := NewGiaManager()
		for i := 0; i < row.Variables; i++ {
			source.AppendCI()
		}
		before := slices.Clone(source.objects)
		truth, e := TruthTableFromWords(row.Variables, []uint64{row.Truth}, 0)
		if e != nil {
			t.Fatal(e)
		}
		graph, root, e := source.WithTruthTable6(truth, row.Inputs, GiaTruthBuildOptions{Hash: row.Hash != 0})
		if e != nil {
			t.Fatal(index, e)
		}
		if root != row.Root {
			t.Fatal("root", index, root, row.Root)
		}
		if _, e = graph.AppendCO(root); e != nil {
			t.Fatal(e)
		}
		if graph.ObjectCount() != len(row.Objects) {
			t.Fatal("objects", index, graph.ObjectCount(), len(row.Objects))
		}
		for id, expected := range row.Objects {
			object := graph.objects[id]
			actual := [3]int{0, -1, -1}
			if object.IsCI() {
				actual = [3]int{1, int(object.Diff1()), -1}
			} else if object.IsCO() {
				a, _ := object.Fanin0(GiaID(id))
				actual = [3]int{2, int(a), -1}
			} else if id != 0 {
				a, _ := object.Fanin0(GiaID(id))
				b, _ := object.Fanin1(GiaID(id))
				actual = [3]int{3, int(a), int(b)}
			}
			if actual != expected {
				t.Fatal("fanin order", index, id, actual, expected)
			}
		}
		simple, e := graph.OutputTruthWord6(0, nil, GiaCofactorOptions{})
		if e != nil || simple != row.Simple {
			t.Fatal("simple", index, simple, row.Simple, e)
		}
		mapping := make([]int, row.Variables)
		for i := range mapping {
			mapping[i] = row.Variables - 1 - i
		}
		mapped, e := graph.OutputTruthWord6(0, mapping, GiaCofactorOptions{})
		if e != nil || mapped != row.Mapped {
			t.Fatal("mapped", index, mapped, row.Mapped, e)
		}
		if e = graph.Check(); e != nil {
			t.Fatal(e)
		}
		if !slices.Equal(source.objects, before) || source.Hashing() || truth.words[0] != row.Truth {
			t.Fatal("source changed")
		}
	}
}
func TestGiaTruthBridgeLimitsAndDeepCone(t *testing.T) {
	if g, _, e := (*GiaManager)(nil).WithTruthTable6(nil, nil, GiaTruthBuildOptions{}); g != nil || !errors.Is(e, ErrDimension) {
		t.Fatal(e)
	}
	if _, e := (*GiaManager)(nil).OutputTruthWord6(0, nil, GiaCofactorOptions{}); !errors.Is(e, ErrDimension) {
		t.Fatal(e)
	}
	source := NewGiaManager()
	a, _ := source.AppendCI()
	b, _ := source.AppendCI()
	truth, _ := TruthTableFromWords(2, []uint64{0x6666666666666666}, 0)
	before := slices.Clone(source.objects)
	for _, hash := range []bool{false, true} {
		g, root, e := source.WithTruthTable6(truth, []GiaLit{a, b}, GiaTruthBuildOptions{Hash: hash, MaxObjects: source.ObjectCount()})
		if g != nil || root != InvalidGiaLit || !errors.Is(e, ErrLimit) {
			t.Fatal(g, root, e)
		}
	}
	for _, when := range []int{1, 3} {
		calls := 0
		g, root, e := source.WithTruthTable6(truth, []GiaLit{a, b}, GiaTruthBuildOptions{Stop: func() bool { calls++; return calls == when }})
		if g != nil || root != InvalidGiaLit || !errors.Is(e, ErrCancelled) {
			t.Fatal(when, g, root, e)
		}
	}
	if !slices.Equal(source.objects, before) || source.Hashing() {
		t.Fatal("failed build mutated source")
	}
	if g, _, e := source.WithTruthTable6(truth, []GiaLit{a, InvalidGiaLit}, GiaTruthBuildOptions{}); g != nil || !errors.Is(e, ErrInvalidLiteral) {
		t.Fatal(e)
	}
	if _, _, e := source.WithTruthTable6(truth, []GiaLit{a}, GiaTruthBuildOptions{}); !errors.Is(e, ErrDimension) {
		t.Fatal(e)
	}
	if _, _, e := source.WithTruthTable6(nil, nil, GiaTruthBuildOptions{}); !errors.Is(e, ErrDimension) {
		t.Fatal(e)
	}
	larger, _ := NewConstantTruthTable(7, false, 0)
	if _, _, e := source.WithTruthTable6(larger, nil, GiaTruthBuildOptions{}); !errors.Is(e, ErrDimension) {
		t.Fatal(e)
	}
	root, _ := source.AppendAnd(a, b)
	for i := 0; i < 10000; i++ {
		root, _ = source.AppendAnd(root, a)
	}
	source.AppendCO(root)
	word, e := source.OutputTruthWord6(0, nil, GiaCofactorOptions{})
	if e != nil || word != truthVariablePattern(0)&truthVariablePattern(1) {
		t.Fatal("deep iterative cone", word, e)
	}
	for _, options := range []GiaCofactorOptions{{MaxObjects: 10}, {MaxMappingEntries: source.ObjectCount() - 1}} {
		if _, e := source.OutputTruthWord6(0, nil, options); !errors.Is(e, ErrLimit) {
			t.Fatal(e)
		}
	}
	if _, e := source.OutputTruthWord6(0, []int{-1, 1}, GiaCofactorOptions{}); !errors.Is(e, ErrDimension) {
		t.Fatal(e)
	}
	if _, e := source.OutputTruthWord6(0, nil, GiaCofactorOptions{Stop: func() bool { return true }}); !errors.Is(e, ErrCancelled) {
		t.Fatal(e)
	}
	if _, e := source.OutputTruthWord6(1, nil, GiaCofactorOptions{}); !errors.Is(e, ErrDimension) {
		t.Fatal(e)
	}
	// Unreached CIs can be unmapped, even in managers with more than six CIs.
	extra := NewGiaManager()
	for i := 0; i < 8; i++ {
		extra.AppendCI()
	}
	extra.AppendCO(GiaConst1)
	if word, e := extra.OutputTruthWord6(0, nil, GiaCofactorOptions{}); e != nil || word != ^uint64(0) {
		t.Fatal("unreached inputs", word, e)
	}
}
func FuzzGiaTruthBridge(f *testing.F) {
	f.Add(uint8(6), uint64(0x96254718abcdef31))
	f.Add(uint8(128), uint64(0))
	f.Fuzz(func(t *testing.T, flags uint8, word uint64) {
		variables := int(flags % 7)
		if variables < 6 {
			width := 1 << uint(variables)
			word &= (uint64(1) << uint(width)) - 1
			for shift := width; shift < 64; shift *= 2 {
				word |= word << uint(shift)
			}
		}
		truth, e := TruthTableFromWords(variables, []uint64{word}, 0)
		if e != nil {
			t.Fatal(e)
		}
		source := NewGiaManager()
		a, _ := source.AppendCI()
		b, _ := source.AppendCI()
		c, _ := source.AppendCI()
		and, _ := source.AppendAnd(a, b)
		xor, _ := source.AppendXorReal(a, b)
		buffer, _ := source.AppendBuffer(and)
		mux, _ := source.AppendMuxReal(c, xor, and)
		source.AppendCO(mux)
		pool := []GiaLit{GiaConst0, GiaConst1, a, b, c, a.Not(), b.Not(), xor, buffer, mux}
		inputs := make([]GiaLit, variables)
		for i := range inputs {
			inputs[i] = pool[int((word>>uint((i*8)%64))%uint64(len(pool)))]
		}
		before := slices.Clone(source.objects)
		target, root, e := source.WithTruthTable6(truth, inputs, GiaTruthBuildOptions{Hash: flags&128 != 0})
		if e != nil {
			t.Fatal(e)
		}
		target.AppendCO(root)
		packed, e := target.OutputTruthWord6(target.OutputCount()-1, nil, GiaCofactorOptions{})
		if e != nil {
			t.Fatal(e)
		}
		expectedWord := uint64(0)
		for assignment := 0; assignment < 64; assignment++ {
			ci := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0}
			values, e := source.evaluateValues(ci)
			if e != nil {
				t.Fatal(e)
			}
			local := 0
			for i, lit := range inputs {
				if values[lit.ID()] != lit.Complemented() {
					local |= 1 << uint(i)
				}
			}
			expected := word&(uint64(1)<<uint(local)) != 0
			if expected {
				expectedWord |= uint64(1) << uint(assignment)
			}
			outputs, e := target.Evaluate(ci)
			if e != nil || outputs[len(outputs)-1] != expected {
				t.Fatal("Shannon composition", assignment, e)
			}
		}
		if packed != expectedWord {
			t.Fatal("packed output", packed, expectedWord)
		}
		if !slices.Equal(source.objects, before) || truth.words[0] != word || source.Hashing() {
			t.Fatal("source changed")
		}
		if e := target.Check(); e != nil {
			t.Fatal(e)
		}
	})
}
