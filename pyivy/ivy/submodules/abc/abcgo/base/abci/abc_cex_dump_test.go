package abci

import (
	"bytes"
	"testing"

	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/misc/util"
)

func TestAbcNtkDumpOneCexSpecialFormat(t *testing.T) {
	n := abc.Abc_NtkAlloc(abc.ABC_NTK_STRASH, abc.ABC_FUNC_AIG, 1)
	pi := abc.Abc_NtkCreatePi(n)
	pi.SetName("a")
	po := abc.Abc_NtkCreatePo(n)
	po.SetName("out")
	abc.Abc_ObjAddFanin(po, pi)
	abc.Abc_NtkOrderCisCos(n)
	cex := &util.Abc_Cex_t{IFrame: 0, NPis: 1, NBits: 1, PData: []uint32{1}}

	var output bytes.Buffer
	Abc_NtkDumpOneCexSpecial(&output, n, cex)
	if got, want := output.String(), "CEX: a@0=1\n"; got != want {
		t.Fatalf("special CEX output = %q, want %q", got, want)
	}
}
