package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func TestTruthCoreNative(t *testing.T) {
	data, e := os.ReadFile("test_vectors/oracle/util_truth.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		Variables, Support, Implies, Intersects, Compare, CompareRev int
		Source, And, Or, Xor, Mux, Majority, Cof0, Cof1              []uint64
	}
	if e = json.Unmarshal(data, &rows); e != nil {
		t.Fatal(e)
	}
	if len(rows) != 9 {
		t.Fatal(len(rows))
	}
	for _, r := range rows {
		source, e := TruthTableFromWords(r.Variables, r.Source, 0)
		if e != nil {
			t.Fatal(e)
		}
		b, _ := NewVariableTruthTable(r.Variables, r.Variables-1, 0)
		ci := 0
		if r.Variables > 2 {
			ci = 1
		}
		c, _ := NewVariableTruthTable(r.Variables, ci, 0)
		and, e := source.And(b, false, false)
		if e != nil {
			t.Fatal(e)
		}
		if r.Variables&1 != 0 {
			and = and.Not()
		}
		or, _ := source.Or(b)
		xor, _ := source.Xor(b, true)
		mux, _ := MuxTruthTables(c, source, b)
		maj, _ := MajorityTruthTables(source, b, c)
		cof0, _ := source.Cofactor(r.Variables-1, false)
		cof1, _ := source.Cofactor(r.Variables-1, true)
		for _, pair := range []struct {
			got  *TruthTable
			want []uint64
		}{{and, r.And}, {or, r.Or}, {xor, r.Xor}, {mux, r.Mux}, {maj, r.Majority}, {cof0, r.Cof0}, {cof1, r.Cof1}} {
			if pair.got == nil || !slices.Equal(pair.got.Words(), pair.want) {
				t.Fatal(r.Variables, pair.got, pair.want)
			}
		}
		support, _ := source.Support()
		mask := 0
		for _, v := range support {
			mask |= 1 << uint(v)
		}
		implies, _ := source.Implies(b)
		intersects, _ := source.Intersects(b, false)
		cmp, _ := source.Compare(b, false)
		rev, _ := source.Compare(b, true)
		if mask != r.Support || implies != (r.Implies != 0) || intersects != (r.Intersects != 0) || cmp != r.Compare || rev != r.CompareRev {
			t.Fatal(r.Variables, mask, cmp, rev)
		}
	}
}
func TestTruthCoreDimensionsOwnershipAndBits(t *testing.T) {
	for variables := 0; variables <= 9; variables++ {
		table, _ := NewConstantTruthTable(variables, false, 0)
		if !table.IsConstant(false) || table.CountOnes() != 0 {
			t.Fatal("false")
		}
		table.Set(0, true)
		if table.CountOnes() != 1 {
			t.Fatal("logical count")
		}
		word, _ := table.Get(0)
		if !word {
			t.Fatal("set")
		}
		clone := table.Clone()
		clone.Set(0, false)
		if !clone.IsConstant(false) || table.IsConstant(false) {
			t.Fatal("clone")
		}
		snapshot := table.Words()
		snapshot[0] = 0
		if table.CountOnes() != 1 {
			t.Fatal("snapshot")
		}
		if !table.Not().Not().Equal(table) {
			t.Fatal("not involution")
		}
		opposite, e := table.Opposite(table.Not())
		if e != nil || !opposite {
			t.Fatal(e)
		}
		if e := table.Set(uint64(1)<<uint(variables), true); !errors.Is(e, ErrDimension) {
			t.Fatal(e)
		}
		if _, e := table.Cofactor(variables, false); !errors.Is(e, ErrDimension) {
			t.Fatal(e)
		}
		for v := 0; v < variables; v++ {
			variable, _ := NewVariableTruthTable(variables, v, 0)
			support, _ := variable.Support()
			if !slices.Equal(support, []int{v}) {
				t.Fatal(support)
			}
			for bits := 0; bits < 1<<uint(variables); bits++ {
				value, e := variable.Get(uint64(bits))
				if e != nil || value != (bits&(1<<uint(v)) != 0) {
					t.Fatal(v, bits, value, e)
				}
			}
		}
	}
	if _, e := TruthTableFromWords(2, []uint64{1}, 0); !errors.Is(e, ErrInvariant) {
		t.Fatal(e)
	}
	if _, e := TruthTableFromWords(7, []uint64{0}, 0); !errors.Is(e, ErrDimension) {
		t.Fatal(e)
	}
	if _, e := NewConstantTruthTable(7, false, 1); !errors.Is(e, ErrLimit) {
		t.Fatal(e)
	}
	if _, e := NewConstantTruthTable(1, false, -1); !errors.Is(e, ErrDimension) {
		t.Fatal(e)
	}
	if _, e := NewVariableTruthTable(2, 2, 0); !errors.Is(e, ErrDimension) {
		t.Fatal(e)
	}
	a, _ := NewConstantTruthTable(1, false, 0)
	b, _ := NewConstantTruthTable(2, false, 0)
	if r, e := a.And(b, false, false); r != nil || !errors.Is(e, ErrDimension) {
		t.Fatal(e)
	}
	if r, e := MuxTruthTables(nil, a, a); r != nil || !errors.Is(e, ErrDimension) {
		t.Fatal(e)
	}
	var zero TruthTable
	if e := zero.Check(); !errors.Is(e, ErrDimension) {
		t.Fatal(e)
	}
}
func FuzzTruthCore(f *testing.F) {
	f.Add(uint8(7), []byte{1, 2, 3, 4, 5})
	f.Add(uint8(0), []byte{255})
	f.Fuzz(func(t *testing.T, flags uint8, data []byte) {
		if len(data) == 0 {
			return
		}
		if len(data) > 128 {
			data = data[:128]
		}
		variables := int(flags % 10)
		n := 1 << uint(variables)
		count, _ := TruthTableWordCount(variables)
		tables := [3]*TruthTable{}
		models := [3][]bool{}
		for side := 0; side < 3; side++ {
			models[side] = make([]bool, n)
			for i := range models[side] {
				v := data[(i/8+side)%len(data)]
				models[side][i] = v&(1<<uint((i+side)%8)) != 0
			}
			words := make([]uint64, count)
			for bit := 0; bit < count*64; bit++ {
				if models[side][bit%n] {
					words[bit/64] |= uint64(1) << uint(bit%64)
				}
			}
			var e error
			tables[side], e = TruthTableFromWords(variables, words, 0)
			if e != nil {
				t.Fatal(e)
			}
		}
		a, b, c := tables[0], tables[1], tables[2]
		ca, cb := flags&16 != 0, flags&32 != 0
		and, _ := a.And(b, ca, cb)
		or, _ := a.Or(b)
		xor, _ := a.Xor(b, flags&64 != 0)
		mux, _ := MuxTruthTables(c, a, b)
		maj, _ := MajorityTruthTables(a, b, c)
		countOnes := 0
		wantImplies, wantIntersect := true, false
		for assignment := 0; assignment < n; assignment++ {
			av, bv, cv := models[0][assignment], models[1][assignment], models[2][assignment]
			if av {
				countOnes++
			}
			wantImplies = wantImplies && (!av || bv)
			wantIntersect = wantIntersect || (av && bv)
			wantMux := bv
			if cv {
				wantMux = av
			}
			wants := []bool{(av != ca) && (bv != cb), av || bv, (av != bv) != (flags&64 != 0), wantMux, (av && bv) || (av && cv) || (bv && cv)}
			for i, table := range []*TruthTable{and, or, xor, mux, maj} {
				got, e := table.Get(uint64(assignment))
				if e != nil || got != wants[i] {
					t.Fatal(variables, assignment, i, got, wants[i], e)
				}
			}
		}
		implies, _ := a.Implies(b)
		intersects, _ := a.Intersects(b, false)
		if a.CountOnes() != countOnes || implies != wantImplies || intersects != wantIntersect {
			t.Fatal("aggregate")
		}
		wantSupport := []int{}
		for variable := 0; variable < variables; variable++ {
			used := false
			for assignment := 0; assignment < n; assignment++ {
				if models[0][assignment] != models[0][assignment^(1<<uint(variable))] {
					used = true
				}
			}
			if used {
				wantSupport = append(wantSupport, variable)
			}
			for _, value := range []bool{false, true} {
				cof, e := a.Cofactor(variable, value)
				if e != nil {
					t.Fatal(e)
				}
				for assignment := 0; assignment < n; assignment++ {
					source := assignment &^ (1 << uint(variable))
					if value {
						source |= 1 << uint(variable)
					}
					got, e := cof.Get(uint64(assignment))
					if e != nil || got != models[0][source] {
						t.Fatal("cofactor", variable, value, assignment)
					}
				}
			}
		}
		support, e := a.Support()
		if e != nil || !slices.Equal(support, wantSupport) {
			t.Fatal(support, wantSupport, e)
		}
	})
}
