package framgr

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

// Original Fra_ManPrepareComb assigns nRegs directly; Fra_ManFinalizeComb
// appends outputs without setting the explicit true-terminal counts.
func TestSourceManagerFinalizeLeavesTrueCountsUnset(t *testing.T) {
	for _, regs := range []int32{0, 1} {
		t.Run(map[int32]string{0: "comb", 1: "sequential"}[regs], func(t *testing.T) {
			g := aig.NewGraph()
			a, b := g.NewInput(), g.NewInput()
			g.AddOutput(a)
			g.AddOutput(b)
			aig.Aig_ManSetRegNum(g, regs)
			var pars Fra_Par_t
			Fra_ParamsDefault(&pars)
			m := Fra_ManStart(g, &pars)
			m.PManFraig = Fra_ManPrepareComb(m)
			Fra_ManFinalizeComb(m)
			if m.PManFraig.RegisterCount() != regs || m.PManFraig.TruePiCount() != 0 || m.PManFraig.TruePoCount() != 0 {
				t.Fatalf("regs=%d truePI=%d truePO=%d", m.PManFraig.RegisterCount(), m.PManFraig.TruePiCount(), m.PManFraig.TruePoCount())
			}
		})
	}
}
