package abc

// Tests for src/aig/aig/aigCanon.c.

import (
	"bytes"
	"errors"
	"slices"
	"testing"
)

func aigCanonicalTable(t testing.TB, variables int, word uint64) *TruthTable {
	t.Helper()
	word, err := NormalizeSmallTruthWord(word, variables)
	if err != nil {
		t.Fatal(err)
	}
	table, err := TruthTableFromWords(variables, []uint64{word}, 0)
	if err != nil {
		t.Fatal(err)
	}
	return table
}

func aigCanonicalCheckForward(t testing.TB, source *TruthTable, result AigSemiCanonicalization) {
	t.Helper()
	if result.Table == nil || len(result.SourceToTarget) != source.variables || len(result.TargetToSource) != source.variables || len(result.InputPhases) != source.variables {
		t.Fatal("incomplete transformation")
	}
	for assignment := uint64(0); assignment < uint64(1)<<uint(source.variables); assignment++ {
		var original uint64
		for variable := 0; variable < source.variables; variable++ {
			value := assignment&(uint64(1)<<uint(result.SourceToTarget[variable])) != 0
			if result.InputPhases[variable] {
				value = !value
			}
			if value {
				original |= uint64(1) << uint(variable)
			}
		}
		want, err := source.Get(original)
		if err != nil {
			t.Fatal(err)
		}
		got, err := result.Table.Get(assignment)
		if err != nil || got != want {
			t.Fatalf("assignment %d maps to %d: got %v want %v (%v)", assignment, original, got, want, err)
		}
	}
	reconstructed, err := result.Reconstruct()
	if err != nil || !reconstructed.Equal(source) {
		t.Fatalf("reverse transformation: %v", err)
	}
}

func TestAigSemiCanonicalExhaustive(t *testing.T) {
	for variables := 0; variables <= 3; variables++ {
		functions := uint64(1) << (uint64(1) << uint(variables))
		for truth := uint64(0); truth < functions; truth++ {
			table := aigCanonicalTable(t, variables, truth)
			result, err := AigSemiCanonicalize(table, nil)
			if err != nil {
				t.Fatalf("variables=%d truth=%x: %v", variables, truth, err)
			}
			aigCanonicalCheckForward(t, table, result)
			for variable := 0; variable+1 < variables; variable++ {
				if result.Signatures[2*variable].Ones > result.Signatures[2*(variable+1)].Ones {
					t.Fatal("signature order")
				}
			}
			for variable := 0; variable < variables; variable++ {
				if result.Signatures[2*variable].Ones > result.Signatures[2*variable+1].Ones {
					t.Fatal("phase order")
				}
			}
		}
	}
}

func aigCanonicalUniqueExample(t testing.TB) (*TruthTable, AigSemiCanonicalization) {
	t.Helper()
	table := aigCanonicalTable(t, 5, 0xfc88fe88)
	result, err := AigSemiCanonicalize(table, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !result.VariablesUnique {
		t.Fatal("fixed example lost unique signatures")
	}
	for variable := 0; variable < table.variables; variable++ {
		if result.Signatures[2*variable].Ones == result.Signatures[2*variable+1].Ones {
			t.Fatal("fixed example lost phase-distinct signatures")
		}
	}
	return table, result
}

func TestAigSemiCanonicalUniqueNPNInvariance(t *testing.T) {
	source, base := aigCanonicalUniqueExample(t)
	permutations := [][]int{{0, 1, 2, 3, 4}, {3, 1, 0, 4, 2}, {1, 0, 4, 2, 3}}
	phases := [][]bool{{false, false, false, false, false}, {true, false, true, false, true}, {true, true, false, true, false}}
	for i, permutation := range permutations {
		transformed, err := source.TransformNPN(permutation, phases[i], false)
		if err != nil {
			t.Fatal(err)
		}
		result, err := AigSemiCanonicalize(transformed, nil)
		if err != nil {
			t.Fatal(err)
		}
		aigCanonicalCheckForward(t, transformed, result)
		if !result.VariablesUnique || !result.Table.Equal(base.Table) {
			t.Fatalf("unique NPN representative differs for transform %d: source=%x base=%x got=%x baseMap=%v gotMap=%v", i, source.Words(), base.Table.Words(), result.Table.Words(), base.SourceToTarget, result.SourceToTarget)
		}
	}
}

func aigCanonicalCheckGraph(t testing.TB, session *AigCanonicalSession, entry int) {
	t.Helper()
	entries := session.Entries()
	graph := session.Graph()
	if graph == nil || entry < 0 || entry >= len(entries) || graph.OutputCount() != len(entries) {
		t.Fatal("session graph/table shape")
	}
	inputs := make([]bool, graph.InputCount())
	for assignment := 0; assignment < 1<<uint(entries[entry].Variables); assignment++ {
		for variable := 0; variable < entries[entry].Variables; variable++ {
			inputs[variable] = assignment&(1<<uint(variable)) != 0
		}
		outputs, err := graph.Evaluate(inputs)
		want, tableErr := entries[entry].Table.Get(uint64(assignment))
		if err != nil || tableErr != nil || outputs[entry] != want {
			t.Fatalf("entry %d assignment %d: got %v want %v, graph=%v table=%v", entry, assignment, outputs[entry], want, err, tableErr)
		}
	}
}

func TestAigCanonicalSessionRecordOwnershipAndEncoding(t *testing.T) {
	source, _ := aigCanonicalUniqueExample(t)
	newSession := func() *AigCanonicalSession {
		session, err := NewAigCanonicalSession(AigCanonicalOptions{MaxVariables: 5, MaxEntries: 8, MaxAigObjects: 1000, MaxDecompositionNodes: 1000})
		if err != nil {
			t.Fatal(err)
		}
		return session
	}
	first := newSession()
	record, err := first.RecordPrime(source)
	if err != nil || !record.New || record.EntryID != 0 {
		t.Fatalf("first record: %+v %v", record, err)
	}
	aigCanonicalCheckGraph(t, first, 0)
	transformed, err := source.TransformNPN([]int{2, 0, 4, 1, 3}, []bool{true, false, true, false, true}, true)
	if err != nil {
		t.Fatal(err)
	}
	again, err := first.RecordPrime(transformed)
	if err != nil || again.New || again.EntryID != 0 {
		t.Fatalf("equivalent record: %+v %v", again, err)
	}
	entries := first.Entries()
	if len(entries) != 1 || entries[0].Visits != 2 {
		t.Fatalf("entries: %+v", entries)
	}
	entries[0].Table.words[0] ^= 1
	if first.Entries()[0].Table.Equal(entries[0].Table) {
		t.Fatal("entry snapshot aliases session")
	}
	encoded, err := first.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	digest, err := first.SHA256()
	if err != nil || len(encoded) == 0 || digest == [32]byte{} {
		t.Fatal("encoding/hash", err)
	}
	second := newSession()
	if _, err = second.RecordPrime(source); err != nil {
		t.Fatal(err)
	}
	if _, err = second.RecordPrime(transformed); err != nil {
		t.Fatal(err)
	}
	encodedAgain, _ := second.MarshalBinary()
	if !bytes.Equal(encoded, encodedAgain) {
		t.Fatal("deterministic serialization")
	}
	graph := first.Graph()
	if _, err = graph.CreateCO(AigConst1); err != nil {
		t.Fatal(err)
	}
	if first.Graph().OutputCount() != 1 {
		t.Fatal("graph snapshot aliases session")
	}
}

func TestAigCanonicalRecordDSDClassification(t *testing.T) {
	session, err := NewAigCanonicalSession(AigCanonicalOptions{MaxVariables: 4, MaxEntries: 8, MaxAigObjects: 1000})
	if err != nil {
		t.Fatal(err)
	}
	and := aigCanonicalTable(t, 4, 0x8000)
	skipped, err := session.Record(and)
	if err != nil || !skipped.SkippedDSD || skipped.EntryID != -1 {
		t.Fatalf("DSD record: %+v %v", skipped, err)
	}
	var prime *TruthTable
	for truth := uint64(1); truth < 1<<16; truth++ {
		candidate := aigCanonicalTable(t, 4, truth)
		decomposition, decomposeErr := DauDSDDecompose(candidate, DauDSDDecomposeOptions{})
		if decomposeErr != nil {
			t.Fatal(decomposeErr)
		}
		if decomposition.LargestPrime == 4 {
			prime = candidate
			break
		}
	}
	if prime == nil {
		t.Fatal("no four-input prime")
	}
	stored, err := session.Record(prime)
	if err != nil || !stored.New || stored.PrimeInputs != 4 {
		t.Fatalf("prime record: %+v %v", stored, err)
	}
	stats := session.Statistics()
	if stats.Total != 2 || stats.FullDSD != 1 || stats.NonDSD != 1 || stats.VariableFunctions[4] != 1 {
		t.Fatalf("statistics: %+v", stats)
	}
	aigCanonicalCheckGraph(t, session, stored.EntryID)
}

func TestAigCanonicalLimitsCancellationAndLifecycle(t *testing.T) {
	if _, err := NewAigCanonicalSession(AigCanonicalOptions{MaxVariables: -1}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err := NewAigCanonicalSession(AigCanonicalOptions{MaxVariables: 13}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if result, err := AigSemiCanonicalize(nil, nil); result.Table != nil || !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	wide, _ := NewConstantTruthTable(13, false, 0)
	if result, err := AigSemiCanonicalize(wide, nil); result.Table != nil || !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if result, err := AigSemiCanonicalize(aigCanonicalTable(t, 3, 0x96), func() bool { return true }); result.Table != nil || !errors.Is(err, ErrCancelled) {
		t.Fatal(err)
	}
	session, err := NewAigCanonicalSession(AigCanonicalOptions{MaxVariables: 5, MaxEntries: 1, MaxAigObjects: 1000})
	if err != nil {
		t.Fatal(err)
	}
	first, _ := aigCanonicalUniqueExample(t)
	if _, err = session.RecordPrime(first); err != nil {
		t.Fatal(err)
	}
	before, _ := session.MarshalBinary()
	second := aigCanonicalTable(t, 5, 0x69969669)
	if _, err = session.RecordPrime(second); !errors.Is(err, ErrLimit) {
		t.Fatal("entry limit", err)
	}
	after, _ := session.MarshalBinary()
	if !bytes.Equal(before, after) {
		t.Fatal("failed record was not atomic")
	}
	session.options.Stop = func() bool { return true }
	if _, err = session.RecordPrime(first); !errors.Is(err, ErrCancelled) {
		t.Fatal("cancel", err)
	}
	afterCancel, _ := session.MarshalBinary()
	if !bytes.Equal(before, afterCancel) {
		t.Fatal("cancelled record was not atomic")
	}
	if err = session.Close(); err != nil {
		t.Fatal(err)
	}
	if session.Graph() != nil || session.Entries() != nil {
		t.Fatal("closed session retains public state")
	}
	if _, err = session.MarshalBinary(); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if err = session.Close(); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
}

func FuzzAigSemiCanonicalRecord(f *testing.F) {
	for _, seed := range []struct {
		variables uint8
		truth     uint64
	}{{0, 0}, {1, 2}, {2, 8}, {3, 0x96}, {4, 0x6996}, {5, 0x81234567}, {6, 0x96a53cc3f00faa55}} {
		f.Add(seed.variables, seed.truth)
	}
	f.Fuzz(func(t *testing.T, rawVariables uint8, truth uint64) {
		variables := int(rawVariables % 7)
		table := aigCanonicalTable(t, variables, truth)
		result, err := AigSemiCanonicalize(table, nil)
		if err != nil {
			t.Fatal(err)
		}
		aigCanonicalCheckForward(t, table, result)
		session, err := NewAigCanonicalSession(AigCanonicalOptions{MaxVariables: 6, MaxEntries: 2, MaxAigObjects: 5000, MaxDecompositionNodes: 1000})
		if err != nil {
			t.Fatal(err)
		}
		record, err := session.RecordPrime(table)
		if err != nil {
			t.Fatal(err)
		}
		aigCanonicalCheckGraph(t, session, record.EntryID)
		entries := session.Entries()
		if len(entries) != 1 || !slices.Equal(entries[0].Table.Words(), record.Transform.Table.Words()) {
			t.Fatal("record/table mismatch")
		}
	})
}
