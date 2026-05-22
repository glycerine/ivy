package goivy

import "testing"

func TestDerivedUpdateExtractsAppliedLHSSymbol(t *testing.T) {
	workNeeded := actionsMkConst("work_needed[7]")
	memcCpl := actionsMkConst("sys_live.memc_cpl")
	tSym := actionsMkConst("T")
	lhs := &Apply{Func: workNeeded, Terms: []Expr{tSym}, aSort: Boolean}
	defn := &LogicDefinition{Lhs: lhs, Rhs: memcCpl}
	upd := NewDerivedUpdate(lhs, defn)

	got, tr, pre := upd.GetUpdateAxioms([]*Const{memcCpl}, nil)
	if tr != nil || pre != nil {
		t.Fatalf("DerivedUpdate should not produce transition/precondition clauses, got tr=%v pre=%v", tr, pre)
	}
	for _, sym := range got {
		if sym.Name == workNeeded.Name {
			return
		}
	}
	t.Fatalf("expected %s to be added to modified symbols, got %#v", workNeeded.Name, got)
}
