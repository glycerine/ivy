package abc

import (
	"testing"

	"github.com/glycerine/abcgo/aig/miniaig"
	"github.com/glycerine/abcgo/map/mio"
)

func TestGateToType(t *testing.T) {
	n, _, gates, _, driver, _ := abcSCLFixture(t)
	if got := Abc_GateToType(driver); got != miniaig.ABC_OPER_BIT_BUF {
		t.Fatalf("buffer operator = %d", got)
	}
	driver.Data = mio.Mio_LibraryReadGateByName(gates, "and2", "")
	if got := Abc_GateToType(driver); got != miniaig.ABC_OPER_BIT_AND {
		t.Fatalf("AND operator = %d", got)
	}
	_ = n
}
