package bac

import (
	"testing"

	"github.com/glycerine/abcgo/misc/util"
)

func TestPsrReadSignalListPreservesCursorAndItems(t *testing.T) {
	names := util.Abc_NamStart(32, 16)
	p := &Psr_Man_t{
		pStrs: names,
		pNtk:  &Psr_Ntk_t{},
		pCur:  "alpha, beta)\x00",
	}
	var signals Vec_Int_t
	if got := Psr_ManReadSignalList(p, &signals, ')', 1); got != 1 {
		t.Fatalf("read signal list returned %d: %s", got, string(p.ErrorStr[:]))
	}
	if Vec_IntSize(&signals) != 4 {
		t.Fatalf("signal list has %d entries, want two formal/actual pairs", Vec_IntSize(&signals))
	}
	if Psr_ManIsChar(p, ')') == 0 {
		t.Fatalf("cursor stopped at %q, want closing parenthesis", p.pCur)
	}
	if got := Abc_Lit2Var2(Vec_IntEntry(&signals, 1)); got != Abc_NamStrFindOrAdd(names, "alpha", nil) {
		t.Fatalf("first signal id = %d", got)
	}
	if got := Abc_Lit2Var2(Vec_IntEntry(&signals, 3)); got != Abc_NamStrFindOrAdd(names, "beta", nil) {
		t.Fatalf("second signal id = %d", got)
	}
}
