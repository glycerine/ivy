package acb

import (
	"fmt"
	"reflect"
	"testing"
)

func TestComputeFunctionExhaustive(t *testing.T) {
	for n := uint(0); n <= 3; n++ {
		for bits := uint64(0); bits < 1<<(1<<n); bits++ {
			for complement := int32(0); complement < 2; complement++ {
				s := sat_solver_new()
				sat_solver_setnvars(s, int32(n)+2)
				var cover Vec_Int_t
				var encoded Vec_Str_t
				var clause []int32
				truth := stretchTruth(bits, n)
				Acb_DeriveCnfFromTruth(truth, int32(n), &cover, &encoded)
				for _, lit := range Vec_StrArray(&encoded) {
					if lit == 255 {
						if !sat_solver_addclause(s, clause) {
							t.Fatal("truth function CNF is contradictory")
						}
						clause = nil
					} else {
						clause = append(clause, int32(lit))
					}
				}
				divs := Vec_IntStartNatural(int32(n))
				got := Acb_ComputeFunction(s, int32(n), int32(n)+1, divs, complement)
				sat_solver_delete(s)
				want := truth
				if complement != 0 {
					want = ^want
				}
				if got != want {
					t.Fatalf("n=%d truth=%x complement=%d got=%x want=%x", n, bits, complement, got, want)
				}
			}
		}
	}
}

func TestPatchEnumerationAllBinaryFunctions(t *testing.T) {
	for bits := uint64(0); bits < 16; bits++ {
		t.Run(fmt.Sprintf("%x", bits), func(t *testing.T) {
			g := Gia_ManStart(32)
			Gia_ManHashAlloc(g)
			a := Gia_ManAppendCi(g)
			b := Gia_ManAppendCi(g)
			target := Gia_ManAppendCi(g)
			f := int32(0)
			for i := 0; i < 4; i++ {
				if bits>>uint(i)&1 == 0 {
					continue
				}
				x, y := a, b
				if i&1 == 0 {
					x ^= 1
				}
				if i&2 == 0 {
					y ^= 1
				}
				f = Gia_ManHashOr(g, f, Gia_ManHashAnd(g, x, y))
			}
			Gia_ManAppendCo(g, Gia_ManHashXor(g, target, f))
			Gia_ManAppendCo(g, a)
			Gia_ManAppendCo(g, b)
			cnf := Mf_ManGenerateCnf(g, 8, 0, 0, 0, 0)
			defer Cnf_DataFree(cnf)
			defer Gia_ManStop(g)
			used := Vec_IntStartNatural(2)
			sop := Acb_DeriveOnePatchFunction(cnf, 0, 1, 2, used, 0)
			if got := Abc_SopToTruth(sop, Abc_SopGetVarNum(sop)); got != stretchTruth(bits, 2) {
				t.Fatalf("SOP=%q truth=%x want=%x", sop, got, stretchTruth(bits, 2))
			}
			if !reflect.DeepEqual(Vec_IntArray(used), []int32{0, 1}) {
				t.Fatal("divisor remapping was not undone")
			}
		})
	}
}

func TestSupportSelectionUpdatesMask(t *testing.T) {
	s := sat_solver_new()
	defer sat_solver_delete(s)
	sat_solver_addclause(s, []int32{0, 2})
	weights := Vec_IntStart(2)
	Vec_IntWriteEntry(weights, 0, 1)
	Vec_IntWriteEntry(weights, 1, 2)
	pats := Vec_WrdStart(2 * NWORDS)
	Vec_WrdWriteEntry(pats, 0, 1)
	Vec_WrdWriteEntry(pats, NWORDS, 2)
	count := int32(2)
	support := Acb_FindSupportNext(s, 0, weights, pats, &count)
	if support == nil || !reflect.DeepEqual(Vec_IntArray(support), []int32{1, 3}) {
		t.Fatalf("support=%v, count=%d", support, count)
	}
}

func TestCheaperSupportReplacementAcrossWords(t *testing.T) {
	s := sat_solver_new()
	defer sat_solver_delete(s)
	sat_solver_setnvars(s, 2)
	sat_solver_addclause(s, []int32{0})
	sat_solver_addclause(s, []int32{2})
	weights := Vec_IntStart(2)
	Vec_IntWriteEntry(weights, 0, 1)
	Vec_IntWriteEntry(weights, 1, 2)
	pats := Vec_WrdStart(2 * NWORDS)
	for i := int32(0); i < 70; i++ {
		Abc_TtSetBit(Vec_WrdEntryP(pats, 0), i)
		Abc_TtSetBit(Vec_WrdEntryP(pats, NWORDS), i)
	}
	support := Vec_IntAlloc(1)
	Vec_IntPush(support, 3)
	Acb_FindReplace(s, 0, weights, pats, 70, support)
	if !reflect.DeepEqual(Vec_IntArray(support), []int32{1}) {
		t.Fatalf("replacement=%v", Vec_IntArray(support))
	}
}

func TestEnabledBufferAllPhases(t *testing.T) {
	for complement := int32(0); complement < 2; complement++ {
		s := sat_solver_new()
		sat_solver_setnvars(s, 3)
		if sat_solver_add_buffer_enable(s, 0, 1, 2, complement) != 2 {
			t.Fatal("clause creation")
		}
		for a := int32(0); a < 8; a++ {
			assumptions := []int32{Abc_Var2Lit(0, 1^(a&1)), Abc_Var2Lit(1, 1^((a>>1)&1)), Abc_Var2Lit(2, 1^((a>>2)&1))}
			got := sat_solver_solve(s, assumptions, nil, 0, 0, 0, 0) == l_True
			want := a&4 == 0 || a&1 == ((a>>1)&1)^complement
			if got != want {
				t.Fatalf("phase%d assignment%b got%v want%v", complement, a, got, want)
			}
		}
		sat_solver_delete(s)
	}
}

func TestCandidateFunctionValidation(t *testing.T) {
	for f := uint64(0); f < 16; f++ {
		for candidate := uint64(0); candidate < 16; candidate++ {
			s := sat_solver_new()
			sat_solver_setnvars(s, 4)
			var cover Vec_Int_t
			var cnf Vec_Str_t
			var clause []int32
			truth := stretchTruth(f, 2)
			Acb_DeriveCnfFromTruth(truth, 2, &cover, &cnf)
			for _, lit := range Vec_StrArray(&cnf) {
				if lit == 255 {
					sat_solver_addclause(s, clause)
					clause = nil
				} else {
					clause = append(clause, int32(lit))
				}
			}
			support := Vec_IntStartNatural(2)
			// First enumerate, leaving guarded blocking clauses in the solver.
			Acb_ComputeFunction(s, 2, 3, support, 0)
			if got := acbCheckFunction(s, 2, 3, support, stretchTruth(candidate, 2)); got != (f == candidate) {
				t.Fatalf("original%x candidate%x accepted%v", f, candidate, got)
			}
			sat_solver_delete(s)
		}
	}
}
