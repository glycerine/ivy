//go:build abc_use_cudd

package abci

import (
	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/bdd/cudd/core"
)

const proveBddEnabled = true

func proveBddFallback(n *Abc_Ntk_t, p *Prove_Params_t) (*Abc_Ntk_t, int32, bool) {
	collapsed := Abc_NtkCollapse(n, p.nBddSizeLimit, 0, p.fBddReorder, 0, 0, 0)
	if collapsed == nil {
		return n, -1, false
	}
	result := int32(0)
	manager, ok := collapsed.FunctionManager.(*core.DdManager)
	if ok && Abc_NtkNodeNum(collapsed) == 1 && Abc_NtkPoNum(collapsed) > 0 {
		if driver := abc.Abc_ObjFanin0(Abc_NtkPo(collapsed, 0)); driver != nil {
			if value, ok := driver.Data.(*core.DdNode); ok && value == manager.LogicZero {
				result = 1
			}
		}
	}
	return collapsed, result, true
}
