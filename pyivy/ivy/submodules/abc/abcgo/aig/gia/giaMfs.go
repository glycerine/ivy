package gia

import (
	"fmt"

	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/opt/sfm"
)

// Gia_ManExtractMfs extracts the mapped combinational network into SFM form.
// The source copy array is used to map optimized objects back during insertion.
func Gia_ManExtractMfs(p *Gia_Man_t) *sfm.Sfm_Ntk_t {
	d := giaMfsExtract(p)
	n, err := sfm.NewNetwork(d.fanins, d.pis, d.pos, d.truths, d.fixed, d.empty)
	if err != nil {
		panic(fmt.Sprintf("gia: invalid extracted MFS network: %v", err))
	}
	return n
}

// Gia_ManInsertMfs rebuilds a mapped GIA after SFM has optimized the network.
func Gia_ManInsertMfs(p *Gia_Man_t, n *sfm.Sfm_Ntk_t, allBoxes int32) *Gia_Man_t {
	d := &giaMfsNetworkData{fanins: n.Fanins, truths: n.Truths, fixed: n.Fixed, empty: n.Empty, pis: n.Pis, pos: n.Pos}
	return giaMfsInsert(p, d, allBoxes != 0)
}

// Gia_ManPerformMfs optimizes the mapped network and retains timing metadata.
func Gia_ManPerformMfs(p *Gia_Man_t, pars *sfm.Sfm_Par_t) *Gia_Man_t {
	if Gia_ManRegNum(p) != 0 || p.VMapping == nil {
		panic("gia: MFS requires a mapped combinational network")
	}
	if p.PManTime != nil && p.PAigExtra == nil {
		fmt.Println("Timing manager is given but there is no GIA of boxes.")
		return nil
	}
	if p.PManTime != nil && Gia_ManCiNum(p.PAigExtra) > 15 {
		fmt.Println("Currently \"&mfs\" cannot process white boxes with more than 15 inputs.")
		return nil
	}
	if Gia_ManLutSizeMax(p) > 15 {
		fmt.Println("Currently \"&mfs\" cannot process nodes with more than 15 fanins.")
		return nil
	}
	n := Gia_ManExtractMfs(p)
	changed := sfm.Sfm_NtkPerform(n, pars)
	if changed == 0 {
		if p.PManTime != nil {
			fmt.Println("The network is not changed by \"&mfs\".")
		}
		result := Gia_ManDup(p)
		result.VMapping = vec.Vec_IntDup(p.VMapping)
		Gia_ManTransferTiming(result, p)
		return result
	}
	allBoxes := int32(0)
	if pars != nil && pars.AllBoxes {
		allBoxes = 1
	}
	result := Gia_ManInsertMfs(p, n, allBoxes)
	if pars != nil && pars.Verbose {
		fmt.Printf("The network has %d nodes changed by \"&mfs\".\n", changed)
	}
	return result
}
