package extrab

import (
	"testing"

	cudd "github.com/glycerine/abcgo/bdd/cudd/core"
)

func TestRootPrimeForwarder(t *testing.T) {
	dd := cudd.NewBDDManager(2)
	f := cudd.Cudd_bddOr(dd, dd.Vars[0], dd.Vars[1])
	p := Extra_zddPrimes(dd, f)
	if p == nil || cudd.Cudd_zddCount(dd, p) != 2 || cudd.Cudd_MakeBddFromZddCover(dd, p) != f {
		t.Fatal("prime forwarder")
	}
}
