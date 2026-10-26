package abc

// Standalone exhaustive, ownership, limit, concurrency and fuzz tests for sbdCnf.c.

import (
	"errors"
	"slices"
	"sync"
	"sync/atomic"
	"testing"
)

func sbdClauseFormula(clauses [][]SatLit, assignment []bool) bool {
	for _, clause := range clauses {
		satisfied := false
		for _, literal := range clause {
			value := assignment[literal.Var()]
			satisfied = satisfied || value != literal.Negative()
		}
		if !satisfied {
			return false
		}
	}
	return true
}

func sbdCNFCheckTruth(t testing.TB, truth uint64, variables int, cnf *SBDCNF) {
	t.Helper()
	clauses := cnf.Clauses()
	for assignment := 0; assignment < 1<<uint(variables); assignment++ {
		want := truth&(uint64(1)<<uint(assignment)) != 0
		for output := 0; output < 2; output++ {
			values := make([]bool, variables+1)
			for variable := 0; variable < variables; variable++ {
				values[variable] = assignment&(1<<uint(variable)) != 0
			}
			values[variables] = output != 0
			got := sbdClauseFormula(clauses, values)
			if got != (want == (output != 0)) {
				t.Fatalf("vars=%d assignment=%d output=%d got=%v truth=%016x\n%v", variables, assignment, output, got, truth, clauses)
			}
		}
	}
}

func TestSBDTruthToCNFExactConstantsAndSemantics(t *testing.T) {
	zero, err := SBDTruthToCNF(0, 0, SBDCNFOptions{})
	if err != nil || !slices.Equal(zero.Clauses()[0], []SatLit{1}) || zero.VariableCount() != 1 || zero.CubeCount() != 1 {
		t.Fatalf("zero=%+v clauses=%v err=%v", zero, zero.Clauses(), err)
	}
	one, err := SBDTruthToCNF(^uint64(0), 0, SBDCNFOptions{})
	if err != nil || !slices.Equal(one.Clauses()[0], []SatLit{0}) {
		t.Fatalf("one=%+v clauses=%v err=%v", one, one.Clauses(), err)
	}
	if text, formatErr := SBDFormatCNF(zero, SBDCNFOptions{}); formatErr != nil || text != "-0 \n" {
		t.Fatalf("zero text=%q err=%v", text, formatErr)
	}
	if text, formatErr := SBDFormatCNF(one, SBDCNFOptions{}); formatErr != nil || text != "0 \n" {
		t.Fatalf("one text=%q err=%v", text, formatErr)
	}
	for variables := 1; variables <= 6; variables++ {
		base := uint64(0x96a53cc3f00f5aa5)
		for _, raw := range []uint64{base, 0, ^uint64(0)} {
			truth, normalizeErr := NormalizeSmallTruthWord(raw, variables)
			if normalizeErr != nil {
				t.Fatal(normalizeErr)
			}
			cnf, buildErr := SBDTruthToCNF(truth, variables, SBDCNFOptions{})
			if buildErr != nil {
				t.Fatal(buildErr)
			}
			sbdCNFCheckTruth(t, truth, variables, cnf)
		}
	}
	andTruth, _ := NormalizeSmallTruthWord(8, 2)
	andCNF, err := SBDTruthToCNF(andTruth, 2, SBDCNFOptions{})
	if err != nil {
		t.Fatal(err)
	}
	want := [][]SatLit{{1, 3, 4}, {2, 5}, {0, 5}}
	if !slices.EqualFunc(andCNF.Clauses(), want, slices.Equal[[]SatLit]) {
		t.Fatalf("AND clauses=%v want=%v", andCNF.Clauses(), want)
	}
	text, err := SBDFormatCNF(andCNF, SBDCNFOptions{})
	if err != nil || text != "-0 -1 2 \n1 -2 \n0 -2 \n" {
		t.Fatalf("AND text=%q err=%v", text, err)
	}
}

func TestSBDTranslateCNFOwnershipAndSemantics(t *testing.T) {
	truth, _ := NormalizeSmallTruthWord(8, 2)
	cnf, _ := SBDTruthToCNF(truth, 2, SBDCNFOptions{})
	translated, err := SBDTranslateCNF(cnf, []SatVar{5, 2, 9}, 2, SBDCNFOptions{})
	if err != nil {
		t.Fatal(err)
	}
	want := [][]SatLit{{11, 4, 18}, {5, 19}, {10, 19}}
	if translated.VariableCount() != 10 || !slices.EqualFunc(translated.Clauses(), want, slices.Equal[[]SatLit]) {
		t.Fatalf("translated vars=%d clauses=%v want=%v", translated.VariableCount(), translated.Clauses(), want)
	}
	for x0 := 0; x0 < 2; x0++ {
		for x1 := 0; x1 < 2; x1++ {
			for output := 0; output < 2; output++ {
				global := make([]bool, 10)
				global[5], global[2], global[9] = x0 != 0, x1 == 0, output != 0
				got := sbdClauseFormula(translated.Clauses(), global)
				wantValue := (x0 != 0 && x1 != 0) == (output != 0)
				if got != wantValue {
					t.Fatalf("x0=%d x1=%d output=%d got=%v want=%v", x0, x1, output, got, wantValue)
				}
			}
		}
	}
	copyClauses := translated.Clauses()
	copyClauses[0][0] = 0
	if translated.Clauses()[0][0] != 11 || cnf.Clauses()[0][0] != 1 {
		t.Fatal("clause storage aliases caller or source")
	}
}

func TestSBDCNFLimitsErrorsCancellationAndConcurrency(t *testing.T) {
	truth, _ := NormalizeSmallTruthWord(0x96, 3)
	for _, options := range []SBDCNFOptions{{MaxClauses: -1}, {MaxLiterals: -1}, {MaxBytes: -1}, {ISOP: TruthISOPOptions{MaxElements: -1}}} {
		if cnf, err := SBDTruthToCNF(truth, 3, options); cnf != nil || !errors.Is(err, ErrDimension) {
			t.Fatalf("negative %+v cnf=%v err=%v", options, cnf, err)
		}
	}
	for _, row := range []struct {
		options SBDCNFOptions
		want    error
	}{
		{SBDCNFOptions{MaxClauses: 1}, ErrLimit},
		{SBDCNFOptions{MaxLiterals: 1}, ErrLimit},
		{SBDCNFOptions{ISOP: TruthISOPOptions{MaxElements: 1}}, ErrLimit},
		{SBDCNFOptions{ISOP: TruthISOPOptions{MaxVisits: 1}}, ErrLimit},
		{SBDCNFOptions{Stop: func() bool { return true }}, ErrCancelled},
	} {
		if cnf, err := SBDTruthToCNF(truth, 3, row.options); cnf != nil || !errors.Is(err, row.want) {
			t.Fatalf("options=%+v cnf=%v err=%v want=%v", row.options, cnf, err, row.want)
		}
	}
	if cnf, err := SBDTruthToCNF(1, 0, SBDCNFOptions{}); cnf != nil || err == nil {
		t.Fatalf("noncanonical constant cnf=%v err=%v", cnf, err)
	}
	if cnf, err := SBDTruthToCNF(truth, 7, SBDCNFOptions{}); cnf != nil || !errors.Is(err, ErrDimension) {
		t.Fatalf("wide cnf=%v err=%v", cnf, err)
	}
	cnf, _ := SBDTruthToCNF(truth, 3, SBDCNFOptions{})
	text, _ := SBDFormatCNF(cnf, SBDCNFOptions{})
	if limited, err := SBDFormatCNF(cnf, SBDCNFOptions{MaxBytes: len(text) - 1}); limited != "" || !errors.Is(err, ErrLimit) {
		t.Fatalf("format limit text=%q err=%v", limited, err)
	}
	if translated, err := SBDTranslateCNF(cnf, []SatVar{0}, 0, SBDCNFOptions{}); translated != nil || !errors.Is(err, ErrDimension) {
		t.Fatalf("short map translated=%v err=%v", translated, err)
	}
	if translated, err := SBDTranslateCNF(cnf, []SatVar{0, 1, 2, 3}, 0, SBDCNFOptions{MaxLiterals: 1}); translated != nil || !errors.Is(err, ErrLimit) {
		t.Fatalf("translation limit translated=%v err=%v", translated, err)
	}
	if translated, err := SBDTranslateCNF(cnf, []SatVar{0, 1, 2, 3}, 0, SBDCNFOptions{Stop: func() bool { return true }}); translated != nil || !errors.Is(err, ErrCancelled) {
		t.Fatalf("translation cancel translated=%v err=%v", translated, err)
	}
	if translated, err := SBDTranslateCNF(cnf, []SatVar{0, 1, 2, 3}, 0, SBDCNFOptions{MaxClauses: 1}); translated != nil || !errors.Is(err, ErrLimit) {
		t.Fatalf("translation clause limit translated=%v err=%v", translated, err)
	}
	badEmpty := &SBDCNF{variables: 1, clauses: [][]SatLit{{}}}
	if _, err := SBDFormatCNF(badEmpty, SBDCNFOptions{}); !errors.Is(err, ErrInvariant) {
		t.Fatalf("empty clause err=%v", err)
	}
	badLiteral := &SBDCNF{variables: 1, clauses: [][]SatLit{{2}}}
	if _, err := SBDTranslateCNF(badLiteral, []SatVar{0}, 0, SBDCNFOptions{}); !errors.Is(err, ErrInvalidLiteral) {
		t.Fatalf("invalid literal err=%v", err)
	}
	if _, err := SBDTranslateCNF(cnf, []SatVar{0, 1, 2, SatVar(MaxLiteralID) + 1}, 0, SBDCNFOptions{}); !errors.Is(err, ErrInvalidLiteral) {
		t.Fatalf("invalid map err=%v", err)
	}
	if cancelled, err := SBDFormatCNF(cnf, SBDCNFOptions{Stop: func() bool { return true }}); cancelled != "" || !errors.Is(err, ErrCancelled) {
		t.Fatalf("format cancel text=%q err=%v", cancelled, err)
	}
	if _, err := SBDFormatCNF(nil, SBDCNFOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatalf("nil format err=%v", err)
	}

	var failures atomic.Int32
	var wait sync.WaitGroup
	for worker := 0; worker < 16; worker++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			result, err := SBDTruthToCNF(truth, 3, SBDCNFOptions{})
			if err != nil || !slices.EqualFunc(result.Clauses(), cnf.Clauses(), slices.Equal[[]SatLit]) {
				failures.Add(1)
			}
		}()
	}
	wait.Wait()
	if failures.Load() != 0 {
		t.Fatalf("concurrent failures=%d", failures.Load())
	}
}

func FuzzSBDTruthToCNF(f *testing.F) {
	f.Add(uint8(2), uint64(8), uint8(0))
	f.Add(uint8(6), uint64(0x6996966996696996), uint8(3))
	f.Add(uint8(3), uint64(0), uint8(1))
	f.Fuzz(func(t *testing.T, variableByte uint8, raw uint64, pivotByte uint8) {
		variables := int(variableByte%6) + 1
		truth, err := NormalizeSmallTruthWord(raw, variables)
		if err != nil {
			t.Fatal(err)
		}
		cnf, err := SBDTruthToCNF(truth, variables, SBDCNFOptions{MaxClauses: 1 << 14, MaxLiterals: 1 << 16, MaxBytes: 1 << 20, ISOP: TruthISOPOptions{MaxElements: 1 << 16, MaxVisits: 1 << 16}})
		if err != nil {
			t.Fatal(err)
		}
		sbdCNFCheckTruth(t, truth, variables, cnf)
		mapping := make([]SatVar, variables+1)
		for index := range mapping {
			mapping[index] = SatVar(variables - index)
		}
		pivot := mapping[int(pivotByte)%len(mapping)]
		translated, err := SBDTranslateCNF(cnf, mapping, pivot, SBDCNFOptions{MaxClauses: 1 << 14, MaxLiterals: 1 << 16})
		if err != nil {
			t.Fatal(err)
		}
		for assignment := 0; assignment < 1<<uint(variables); assignment++ {
			for output := 0; output < 2; output++ {
				local := make([]bool, variables+1)
				global := make([]bool, variables+1)
				for variable := 0; variable < variables; variable++ {
					local[variable] = assignment&(1<<uint(variable)) != 0
				}
				local[variables] = output != 0
				for variable, mapped := range mapping {
					global[mapped] = local[variable]
					if mapped == pivot {
						global[mapped] = !global[mapped]
					}
				}
				got := sbdClauseFormula(translated.Clauses(), global)
				want := sbdClauseFormula(cnf.Clauses(), local)
				if got != want {
					t.Fatalf("vars=%d assignment=%d output=%d pivot=%d got=%v want=%v", variables, assignment, output, pivot, got, want)
				}
			}
		}
	})
}
