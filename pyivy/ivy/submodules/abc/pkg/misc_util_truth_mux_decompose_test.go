package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func TestTruthMuxDecompositionNativeAndCorrections(t *testing.T) {
	data, e := os.ReadFile("test_vectors/oracle/util_truth_mux_decompose.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		Variables, Control, Mode, Code       int
		Source, CRemainder, CorrectRemainder uint64
	}
	if e = json.Unmarshal(data, &rows); e != nil {
		t.Fatal(e)
	}
	if len(rows) != 72 {
		t.Fatal(len(rows))
	}
	rejected, recovered := 0, 0
	for _, row := range rows {
		source, e := TruthTableFromWords(row.Variables, []uint64{row.Source}, 0)
		if e != nil {
			t.Fatal(e)
		}
		r, found, e := source.DecomposeMuxInput(row.Control)
		if e != nil {
			t.Fatal(e)
		}
		invalidC := row.Variables >= 4 && row.Mode == 2
		missedC := row.Variables >= 4 && row.Mode == 3
		if invalidC {
			rejected++
			if found || row.Code < 0 {
				t.Fatal("C false positive", row, found)
			}
			mismatch := false
			low, high, neg := (row.Code&0xffff)>>1, (row.Code>>16)>>1, row.Code&1 != 0
			for assignment := 0; assignment < 1<<uint(row.Variables); assignment++ {
				factor := (assignment&(1<<uint(low)) != 0) != neg
				if assignment&(1<<uint(row.Control)) != 0 {
					factor = assignment&(1<<uint(high)) != 0
				}
				target := assignment &^ (1 << uint(row.Control))
				if factor {
					target |= 1 << uint(row.Control)
				}
				got := row.CRemainder&(uint64(1)<<uint(target)) != 0
				want := row.Source&(uint64(1)<<uint(assignment)) != 0
				if got != want {
					mismatch = true
				}
			}
			if !mismatch {
				t.Fatal("fixture must refute C reconstruction")
			}
			continue
		}
		if !found || r.Control != row.Control || r.WhenFalse != (row.Control+1)%row.Variables || r.WhenTrue != (row.Control+2)%row.Variables || r.ComplementFalse != (row.Mode != 0) || !slices.Equal(r.Remainder.Words(), []uint64{row.CorrectRemainder}) {
			t.Fatal(row, r, found)
		}
		if missedC {
			recovered++
			if row.Code != -1 {
				t.Fatal("C missed valid mux", row)
			}
		} else {
			code := ((r.WhenTrue * 2) << 16) | (r.WhenFalse * 2)
			if r.ComplementFalse {
				code |= 1
			}
			if row.Code != code || row.CRemainder != r.Remainder.words[0] {
				t.Fatal("native parity", row, code)
			}
		}
	}
	if rejected != 15 || recovered != 15 {
		t.Fatal(rejected, recovered)
	}
}
func TestTruthMuxDecompositionWideAndErrors(t *testing.T) {
	control, _ := NewVariableTruthTable(13, 12, 0)
	low, _ := NewVariableTruthTable(13, 0, 0)
	high, _ := NewVariableTruthTable(13, 6, 0)
	common, _ := NewVariableTruthTable(13, 1, 0)
	factor, _ := MuxTruthTables(control, high, low.Not())
	source, _ := factor.And(common, false, false)
	before := source.Words()
	r, found, e := source.DecomposeMuxInput(12)
	if e != nil || !found || !r.ComplementFalse || r.WhenFalse != 0 || r.WhenTrue != 6 {
		t.Fatal(found, e)
	}
	want, _ := control.And(common, false, false)
	if !r.Remainder.Equal(want) {
		t.Fatal("wide result")
	}
	r.Remainder.Set(0, true)
	if !slices.Equal(before, source.Words()) {
		t.Fatal("alias")
	}
	if _, found, e := source.DecomposeMuxInput(11); found || e != nil {
		t.Fatal("unused control", e)
	}
	if _, _, e := source.DecomposeMuxInput(13); !errors.Is(e, ErrDimension) {
		t.Fatal(e)
	}
	if _, _, e := (*TruthTable)(nil).DecomposeMuxInput(0); !errors.Is(e, ErrDimension) {
		t.Fatal(e)
	}
	constant, _ := NewConstantTruthTable(4, true, 0)
	if _, found, e := constant.DecomposeMuxInput(0); found || e != nil {
		t.Fatal(e)
	}
}
func FuzzTruthMuxDecomposition(f *testing.F) {
	f.Add(uint8(7), []byte{1, 2, 3, 4, 5})
	f.Add(uint8(64), []byte{255})
	f.Fuzz(func(t *testing.T, flags uint8, data []byte) {
		if len(data) == 0 {
			return
		}
		if len(data) > 96 {
			data = data[:96]
		}
		variables := 3 + int(flags%7)
		n := 1 << uint(variables)
		control := int(data[0]) % variables
		low, high := (control+1)%variables, (control+2)%variables
		common := -1
		for v := 0; v < variables; v++ {
			if v != control && v != low && v != high {
				common = v
				break
			}
		}
		model := make([]bool, n)
		for assignment := range model {
			c, a, b := assignment&(1<<uint(control)) != 0, assignment&(1<<uint(low)) != 0, assignment&(1<<uint(high)) != 0
			z := false
			if common >= 0 {
				z = assignment&(1<<uint(common)) != 0
			}
			factor := a != (flags&16 != 0)
			if c {
				factor = b
			}
			value := factor != z
			if flags&32 != 0 {
				if common < 0 {
					z = true
				}
				value = factor && z
			}
			if flags&64 != 0 {
				if common < 0 {
					z = true
				}
				value = !(a && z)
				if c {
					value = b && z
				}
			}
			if flags&128 != 0 {
				value = data[(assignment/8)%len(data)]&(1<<uint(assignment%8)) != 0
			}
			model[assignment] = value
		}
		count, _ := TruthTableWordCount(variables)
		words := make([]uint64, count)
		for bit := 0; bit < count*64; bit++ {
			if model[bit%n] {
				words[bit/64] |= uint64(1) << uint(bit%64)
			}
		}
		source, e := TruthTableFromWords(variables, words, 0)
		if e != nil {
			t.Fatal(e)
		}
		exclusive := [2][]int{}
		for variable := 0; variable < variables; variable++ {
			if variable == control {
				continue
			}
			used := [2]bool{}
			for branch := 0; branch < 2; branch++ {
				for assignment := 0; assignment < n; assignment++ {
					base := assignment &^ (1 << uint(control))
					if branch == 1 {
						base |= 1 << uint(control)
					}
					if model[base] != model[base^(1<<uint(variable))] {
						used[branch] = true
					}
				}
			}
			if used[0] && !used[1] {
				exclusive[0] = append(exclusive[0], variable)
			}
			if used[1] && !used[0] {
				exclusive[1] = append(exclusive[1], variable)
			}
		}
		wantFound, wantPhase := false, false
		if len(exclusive[0]) == 1 && len(exclusive[1]) == 1 {
			a, b := exclusive[0][0], exclusive[1][0]
			same, cross := true, true
			for assignment := 0; assignment < n; assignment++ {
				base := assignment &^ (1 << uint(control)) &^ (1 << uint(a)) &^ (1 << uint(b))
				c00, c01 := model[base], model[base|(1<<uint(a))]
				c10, c11 := model[base|(1<<uint(control))], model[base|(1<<uint(control))|(1<<uint(b))]
				same = same && c00 == c10 && c01 == c11
				cross = cross && c00 == c11 && c01 == c10
			}
			wantFound = same || cross
			wantPhase = !same && cross
		}
		r, found, e := source.DecomposeMuxInput(control)
		if e != nil || found != wantFound {
			t.Fatal(found, wantFound, e)
		}
		if !found {
			if r.Remainder != nil {
				t.Fatal("partial")
			}
			return
		}
		if r.ComplementFalse != wantPhase || r.WhenFalse != exclusive[0][0] || r.WhenTrue != exclusive[1][0] {
			t.Fatal("projection")
		}
		for _, v := range []int{r.WhenFalse, r.WhenTrue} {
			used, e := r.Remainder.HasVariable(v)
			if e != nil || used {
				t.Fatal("data input remains")
			}
		}
		for assignment := 0; assignment < n; assignment++ {
			factor := (assignment&(1<<uint(r.WhenFalse)) != 0) != r.ComplementFalse
			if assignment&(1<<uint(control)) != 0 {
				factor = assignment&(1<<uint(r.WhenTrue)) != 0
			}
			target := assignment &^ (1 << uint(control))
			if factor {
				target |= 1 << uint(control)
			}
			got, e := r.Remainder.Get(uint64(target))
			if e != nil || got != model[assignment] {
				t.Fatal("reconstruction", variables, assignment)
			}
		}
		if !slices.Equal(words, source.Words()) {
			t.Fatal("mutation")
		}
	})
}
