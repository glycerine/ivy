package extrab

import (
	"testing"

	cudd "github.com/glycerine/abcgo/bdd/cudd/core"
)

func TestRootMaxMinForwarders(t *testing.T) {
	dd := cudd.NewBDDManager(0)
	a := cudd.Cudd_zddChange(dd, dd.One, 0)
	b := cudd.Cudd_zddChange(dd, dd.One, 1)
	ab := cudd.Cudd_zddChange(dd, a, 1)
	family := cudd.Cudd_zddUnion(dd, a, ab)
	if Extra_zddMinimal(dd, family) != a || Extra_zddMaxUnion(dd, family, dd.Zero) != ab || Extra_zddMinUnion(dd, family, dd.Zero) != a || Extra_zddDotProduct(dd, a, b) != ab || Extra_zddCrossProduct(dd, a, b) != dd.One || Extra_zddMaxDotProduct(dd, family, dd.One) != ab {
		t.Fatal("MaxMin forwards")
	}
}
