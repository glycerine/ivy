package extrab

import (
	"testing"

	cudd "github.com/glycerine/abcgo/bdd/cudd/core"
)

func TestRootSetForwarders(t *testing.T) {
	dd := cudd.NewBDDManager(0)
	a := cudd.Cudd_zddChange(dd, dd.One, 0)
	ab := cudd.Cudd_zddChange(dd, a, 1)
	x := cudd.Cudd_zddUnion(dd, a, ab)
	if Extra_zddSubSet(dd, x, a) != a || Extra_zddSupSet(dd, x, ab) != ab || Extra_zddNotSubSet(dd, x, a) != ab || Extra_zddNotSupSet(dd, x, ab) != a || Extra_zddMaximal(dd, x) != ab || Extra_zddMaxNotSupSet(dd, x, dd.Zero) != ab {
		t.Fatal("set forwarders")
	}
	if !Extra_zddEmptyBelongs(dd, dd.One) || !Extra_zddIsOneSubset(dd, a) || Extra_zddIsOneSubset(dd, dd.Zero) {
		t.Fatal("predicate forwarders")
	}
}
