package abc

import (
	"errors"
	"slices"
	"testing"
)

func bdcTestTruth(t testing.TB, variables int, bits uint64) *TruthTable {
	t.Helper()
	truth, err := NewConstantTruthTable(variables, false, 0)
	if err != nil {
		t.Fatal(err)
	}
	for assignment := 0; assignment < 1<<uint(variables); assignment++ {
		if bits&(uint64(1)<<uint(assignment)) != 0 {
			if err = truth.Set(uint64(assignment), true); err != nil {
				t.Fatal(err)
			}
		}
	}
	return truth
}

func bdcIndependentContainment(variables int, on, off, truth uint64) bool {
	mask := uint64(1)<<uint(1<<uint(variables)) - 1
	return on&^truth == 0 && truth&off == 0 && on&off&mask == 0
}

func TestBdcTableContainmentExhaustive(t *testing.T) {
	const variables = 3
	limit := uint64(1) << uint(1<<variables)
	for on := uint64(0); on < limit; on++ {
		for off := uint64(0); off < limit; off++ {
			if on&off != 0 {
				continue
			}
			isf := BdcISF{Support: 7, On: bdcTestTruth(t, variables, on), Off: bdcTestTruth(t, variables, off)}
			for truthBits := uint64(0); truthBits < limit; truthBits++ {
				got, err := BdcTableCheckContainment(isf, bdcTestTruth(t, variables, truthBits))
				if err != nil || got != bdcIndependentContainment(variables, on, off, truthBits) {
					t.Fatalf("on=%02x off=%02x truth=%02x got=%v err=%v", on, off, truthBits, got, err)
				}
			}
		}
	}
}

func bdcPreparedTable(t testing.TB, variables, maximum int) *BdcTable {
	t.Helper()
	table, err := NewBdcTable(BdcTableOptions{Variables: variables, MaxFunctions: maximum})
	if err != nil {
		t.Fatal(err)
	}
	constant, err := NewConstantTruthTable(variables, true, 0)
	if err != nil {
		t.Fatal(err)
	}
	literal, err := table.Add(0, constant)
	if err != nil || literal != 0 {
		t.Fatal(literal, err)
	}
	return table
}

func TestBdcTableLookupOrderComplementAndConstants(t *testing.T) {
	table := bdcPreparedTable(t, 2, 8)
	xor := bdcTestTruth(t, 2, 0x6)
	and := bdcTestTruth(t, 2, 0x8)
	xorLit, err := table.Add(3, xor)
	if err != nil {
		t.Fatal(err)
	}
	andLit, err := table.Add(3, and)
	if err != nil {
		t.Fatal(err)
	}
	zero := bdcTestTruth(t, 2, 0)
	any := BdcISF{Support: 3, Unique: 1, On: zero, Off: zero}
	if got, found, err := table.Lookup(any); err != nil || !found || got != andLit {
		t.Fatal("newest direct", got, found, err, andLit)
	}
	complement := BdcISF{Support: 3, On: xor.Not(), Off: xor.Clone()}
	beforeOn, beforeOff := complement.On.Words(), complement.Off.Words()
	if got, found, err := table.Lookup(complement); err != nil || !found || got != xorLit.Not() {
		t.Fatal("complement", got, found, err, xorLit.Not())
	}
	if !slices.Equal(beforeOn, complement.On.Words()) || !slices.Equal(beforeOff, complement.Off.Words()) {
		t.Fatal("lookup mutated caller ISF")
	}
	one := bdcTestTruth(t, 2, 0xf)
	for _, test := range []struct {
		isf  BdcISF
		want BdcLit
	}{
		{BdcISF{On: one, Off: zero}, 0},
		{BdcISF{On: zero, Off: one}, 1},
	} {
		got, found, err := table.Lookup(test.isf)
		if err != nil || !found || got != test.want {
			t.Fatal("constant", got, found, err, test.want)
		}
	}
}

func TestBdcTableOwnershipClearAndErrors(t *testing.T) {
	table := bdcPreparedTable(t, 3, 2)
	variable := bdcTestTruth(t, 3, 0xaa)
	literal, err := table.Add(1, variable)
	if err != nil || literal != 2 || table.Len() != 2 || !slices.Equal(table.OccupiedSupports(), []uint32{0, 1}) {
		t.Fatal(literal, err, table.Len(), table.OccupiedSupports())
	}
	if _, err = table.Add(2, bdcTestTruth(t, 3, 0xcc)); !errors.Is(err, ErrLimit) {
		t.Fatal("limit", err)
	}
	variable.Set(0, true)
	snapshot, err := table.Function(literal.Not())
	if err != nil || snapshot.Support != 1 || snapshot.Truth.Equal(variable) {
		t.Fatal("input aliases table", snapshot, err)
	}
	snapshot.Truth.Set(0, true)
	again, _ := table.Function(literal)
	if again.Truth.Equal(snapshot.Truth) {
		t.Fatal("snapshot aliases table")
	}
	if err = table.Clear(); err != nil || table.Len() != 0 || len(table.OccupiedSupports()) != 0 {
		t.Fatal(err, table.Len(), table.OccupiedSupports())
	}
	if _, _, err = table.Lookup(BdcISF{On: bdcTestTruth(t, 3, 0), Off: bdcTestTruth(t, 3, 0xff)}); !errors.Is(err, ErrInvariant) {
		t.Fatal("lookup before prepare", err)
	}
	if got, err := table.Add(0, bdcTestTruth(t, 3, 0xff)); err != nil || got != 0 {
		t.Fatal("reprepare", got, err)
	}
	if _, err = table.Add(0, bdcTestTruth(t, 3, 0xff)); !errors.Is(err, ErrInvariant) {
		t.Fatal("second support zero", err)
	}

	for _, test := range []struct {
		options BdcTableOptions
		want    error
	}{
		{BdcTableOptions{Variables: -1}, ErrDimension},
		{BdcTableOptions{Variables: 16}, ErrDimension},
		{BdcTableOptions{Variables: 2, MaxFunctions: -1}, ErrDimension},
		{BdcTableOptions{Variables: 2, Stop: func() bool { return true }}, ErrCancelled},
	} {
		got, err := NewBdcTable(test.options)
		if got != nil || !errors.Is(err, test.want) {
			t.Fatal(test.options, got, err)
		}
	}
	if _, err = (*BdcTable)(nil).Add(0, variable); !errors.Is(err, ErrDimension) {
		t.Fatal("nil table", err)
	}
	if _, err = table.Add(1, bdcTestTruth(t, 2, 0xa)); !errors.Is(err, ErrDimension) {
		t.Fatal("truth dimensions", err)
	}
	if _, err = table.Add(2, bdcTestTruth(t, 3, 0xaa)); !errors.Is(err, ErrInvariant) {
		t.Fatal("wrong support", err)
	}
	bad := BdcISF{Support: 1, On: bdcTestTruth(t, 3, 1), Off: bdcTestTruth(t, 3, 1)}
	if _, _, err = table.Lookup(bad); !errors.Is(err, ErrInvariant) {
		t.Fatal("overlapping ISF", err)
	}
	bad = BdcISF{Support: 1, Unique: 2, On: bdcTestTruth(t, 3, 0), Off: bdcTestTruth(t, 3, 0)}
	if _, _, err = table.Lookup(bad); !errors.Is(err, ErrDimension) {
		t.Fatal("unique outside support", err)
	}
	if _, err = table.Function(InvalidBdcLit); !errors.Is(err, ErrInvalidLiteral) {
		t.Fatal("invalid function", err)
	}
}

func TestBdcTableCancellationIsAtomic(t *testing.T) {
	stop := false
	table, err := NewBdcTable(BdcTableOptions{Variables: 2, MaxFunctions: 3, Stop: func() bool { return stop }})
	if err != nil {
		t.Fatal(err)
	}
	one := bdcTestTruth(t, 2, 0xf)
	if _, err = table.Add(0, one); err != nil {
		t.Fatal(err)
	}
	stop = true
	if _, err = table.Add(3, bdcTestTruth(t, 2, 6)); !errors.Is(err, ErrCancelled) || table.Len() != 1 || len(table.OccupiedSupports()) != 1 {
		t.Fatal(err, table.Len(), table.OccupiedSupports())
	}
	if err = table.Clear(); !errors.Is(err, ErrCancelled) || table.Len() != 1 {
		t.Fatal(err, table.Len())
	}
}

func FuzzBdcTableLookup(f *testing.F) {
	for _, seed := range [][]byte{{}, {0}, {6, 8, 9, 0, 6}, {255, 90, 165, 24, 66, 129}} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		const variables = 3
		table := bdcPreparedTable(t, variables, 33)
		type modelFunction struct {
			literal BdcLit
			support uint32
			truth   uint64
		}
		model := []modelFunction{{literal: 0, truth: 0xff}}
		functionCount := len(data) / 2
		if functionCount > 32 {
			functionCount = 32
		}
		for index := 0; index < functionCount; index++ {
			bits := uint64(data[index])
			truth := bdcTestTruth(t, variables, bits)
			support, err := bdcTruthSupportMask(truth)
			if err != nil || support == 0 {
				continue
			}
			literal, err := table.Add(support, truth)
			if err != nil {
				t.Fatal(err)
			}
			model = append(model, modelFunction{literal: literal, support: support, truth: bits})
		}
		var a, b byte
		if len(data) != 0 {
			a = data[len(data)-1]
		}
		if len(data) > 1 {
			b = data[len(data)-2]
		}
		on := uint64(a &^ b)
		off := uint64(b &^ a)
		support := uint32(1 + int(a^b)%7)
		isf := BdcISF{Support: support, Unique: support & uint32(a), On: bdcTestTruth(t, variables, on), Off: bdcTestTruth(t, variables, off)}
		want := InvalidBdcLit
		found := false
		for index := len(model) - 1; index >= 0; index-- {
			candidate := model[index]
			if candidate.support == support && bdcIndependentContainment(variables, on, off, candidate.truth) {
				want, found = candidate.literal, true
				break
			}
		}
		if !found {
			for index := len(model) - 1; index >= 0; index-- {
				candidate := model[index]
				if candidate.support == support && bdcIndependentContainment(variables, off, on, candidate.truth) {
					want, found = candidate.literal.Not(), true
					break
				}
			}
		}
		got, gotFound, err := table.Lookup(isf)
		if err != nil || gotFound != found || got != want {
			t.Fatalf("support=%d on=%02x off=%02x got=(%d,%v,%v) want=(%d,%v)", support, on, off, got, gotFound, err, want, found)
		}
	})
}
