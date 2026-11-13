package bmc

import (
	"fmt"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

func Bmc_GiaTargetStates(p *gia.Gia_Man_t, cex *util.Abc_Cex_t, begin, end, combOnly, useOne, allFrames, verbose int32) *gia.Gia_Man_t {
	if p == nil || cex == nil || begin < 0 || end < begin || end > cex.IFrame {
		if verbose != 0 {
			fmt.Println("Invalid counterexample frame range.")
		}
		return nil
	}
	var initial *vec.Vec_Bit_t
	var result *gia.Gia_Man_t
	switch {
	case useOne != 0:
		result = Bmc_GiaGenerateGiaOne(p, cex, &initial, begin, end)
	case allFrames != 0:
		result = Bmc_GiaGenerateGiaAllFrames(p, cex, &initial, begin, end)
	default:
		result = Bmc_GiaGenerateGiaAllOne(p, cex, &initial, begin, end)
	}
	if combOnly == 0 {
		combined := gia.Gia_ManDupWithNewPo(p, result)
		result = gia.Gia_ManDupFlip(combined, vec.Vec_BitArray(initial))
	}
	return result
}
func Bmc_AigTargetStates(p *aig.Graph, cex *util.Abc_Cex_t, begin, end, combOnly, useOne, allFrames, verbose int32) *aig.Graph {
	if p == nil || cex == nil {
		return nil
	}
	g := gia.Gia_ManFromAigSimple(p)
	if gia.Gia_ManVerifyCex(g, cex, 0) == 0 {
		if verbose != 0 {
			fmt.Printf("Current CEX does not fail AIG %q.\n", p.Name)
		}
		return nil
	}
	result := Bmc_GiaTargetStates(g, cex, begin, end, combOnly, useOne, allFrames, verbose)
	if result == nil {
		return nil
	}
	return gia.Gia_ManToAigSimple(result)
}
