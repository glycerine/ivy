package cudd

import (
	"github.com/glycerine/abcgo/bdd/cudd/core"
	"testing"
)

func TestADDNumericPublicForwarders(t *testing.T) {
	dd := NewBDDManager(1)
	v := core.Cudd_addIthVar(dd, 0)
	f := core.Cudd_addIte(dd, v, core.Cudd_addConst(dd, 4), core.Cudd_addConst(dd, 2))
	if Cudd_addExistAbstract(dd, f, v).Value != 6 || Cudd_addUnivAbstract(dd, f, v).Value != 8 || Cudd_addOrAbstract(dd, v, v) != dd.One {
		t.Fatal("abstraction forwarders")
	}
	if Cudd_addFindMax(dd, f).Value != 4 || Cudd_addFindMin(dd, f).Value != 2 || Cudd_addIthBit(dd, f, 1) == nil {
		t.Fatal("find forwarders")
	}
	inverse := Cudd_addScalarInverse(dd, f, core.Cudd_addConst(dd, 1))
	if inverse == nil {
		t.Fatal("inverse forwarder")
	}
	if value, err := dd.EvaluateADD(inverse, []bool{true}); err != nil || value != .25 {
		t.Fatal(value, err)
	}
}
