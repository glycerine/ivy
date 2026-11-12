package nwk

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/hop"
	"github.com/glycerine/abcgo/bool/bdc"
	"github.com/glycerine/abcgo/misc/vec"
)

type Bdc_Man_t = bdc.Bdc_Man_t
type Bdc_Fun_t = bdc.Bdc_Fun_t

func Extra_TruthWordNum(n int32) int32 {
	if n <= 5 {
		return 1
	}
	return 1 << uint(n-5)
}
func Extra_TruthNot(out, in []uint32, n int32) {
	for i := range out[:Extra_TruthWordNum(n)] {
		out[i] = ^in[i]
	}
}
func Extra_TruthOr(out, a, b []uint32, n int32) {
	for i := range out[:Extra_TruthWordNum(n)] {
		out[i] = a[i] | b[i]
	}
}
func Extra_TruthSharp(out, a, b []uint32, n int32) {
	for i := range out[:Extra_TruthWordNum(n)] {
		out[i] = a[i] &^ b[i]
	}
}
func Bdc_FunCopyHop(f *Bdc_Fun_t) *Hop_Obj_t {
	return hop.Hop_NotCond(bdc.Bdc_FuncCopy(bdc.Bdc_Regular(f)).(*Hop_Obj_t), bdc.Bdc_IsComplement(f))
}
func Nwk_NodeIfNodeResyn(p *Bdc_Man_t, h *Hop_Man_t, root *Hop_Obj_t, n int32, scratch *Vec_Int_t, care []uint32, probability float32) *Hop_Obj_t {
	if n > 16 {
		panic("nwk: decomposition support exceeds sixteen")
	}
	truth := append([]uint32(nil), hop.Hop_ManConvertAigToTruth(h, root, n, scratch, 0)...)
	decompositionCare := care
	if probability >= 0 {
		if probability > 1 {
			panic("nwk: invalid switching probability")
		}
		activity := 2 * probability * (1 - probability)
		if activity >= .4 {
			if len(care) < len(truth) {
				panic("nwk: power decomposition needs care table")
			}
			for i := range truth {
				if probability > .5 {
					truth[i] |= ^care[i]
				} else {
					truth[i] &= care[i]
				}
			}
			decompositionCare = nil
		}
	}
	if bdc.Bdc_ManDecompose(p, truth, decompositionCare, n, nil, 1000) < 0 {
		panic("nwk: decomposition failed")
	}
	bdc.Bdc_FuncSetCopy(bdc.Bdc_ManFunc(p, 0), hop.Hop_ManConst1(h))
	for i := int32(0); i < n; i++ {
		bdc.Bdc_FuncSetCopy(bdc.Bdc_ManFunc(p, i+1), hop.Hop_IthVar(h, i))
	}
	for i := n + 1; i < bdc.Bdc_ManNodeNum(p); i++ {
		f := bdc.Bdc_ManFunc(p, i)
		bdc.Bdc_FuncSetCopy(f, hop.Hop_And(h, Bdc_FunCopyHop(bdc.Bdc_FuncFanin0(f)), Bdc_FunCopyHop(bdc.Bdc_FuncFanin1(f))))
	}
	return Bdc_FunCopyHop(bdc.Bdc_ManRoot(p))
}
func Nwk_ManBidecResyn(p *Nwk_Man_t, verbose int32) {
	n := Nwk_ManGetFaninMax(p)
	if n < 2 {
		return
	}
	n = min(n, 15)
	manager := bdc.Bdc_ManAlloc(&bdc.Bdc_Par_t{NVarsMax: n, FVerbose: verbose})
	defer bdc.Bdc_ManFree(manager)
	scratch := vec.Vec_IntAlloc(0)
	var gain int32
	for _, o := range p.Objects {
		if o == nil || o.Type != NWK_OBJ_NODE || len(o.Fanins) > 15 {
			continue
		}
		before := hop.Hop_DagSize(o.Function)
		o.Function = Nwk_NodeIfNodeResyn(manager, p.Hop, o.Function, int32(len(o.Fanins)), scratch, nil, -1)
		gain += before - hop.Hop_DagSize(o.Function)
	}
	if verbose != 0 {
		fmt.Fprintf(writer(p), "Total gain in AIG nodes = %d.\n", gain)
	}
}
