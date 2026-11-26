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
	// C keeps one truth word per object in vTruths and the truth tables of
	// objects with more than six fanins in vTruths2 at offset vStarts.
	vTruths := make([]uint64, len(d.fanins))
	vStarts := make([]int, len(d.fanins))
	var vTruths2 []uint64
	for i, fanins := range d.fanins {
		if len(fanins) <= 6 {
			vTruths[i] = d.truths[i][0]
		} else {
			vStarts[i] = len(vTruths2)
			vTruths2 = append(vTruths2, d.truths[i][:1<<uint(len(fanins)-6)]...)
		}
	}
	n, err := sfm.Sfm_NtkConstruct(d.fanins, d.pis, d.pos, d.fixed, d.empty, vTruths, vStarts, vTruths2)
	if err != nil {
		panic(fmt.Sprintf("gia: invalid extracted MFS network: %v", err))
	}
	return n
}

// Gia_ManInsertMfs rebuilds a mapped GIA after SFM has optimized the network.
func Gia_ManInsertMfs(p *Gia_Man_t, n *sfm.Sfm_Ntk_t, allBoxes int32) *Gia_Man_t {
	d := &giaMfsNetworkData{fanins: make([][]int, len(n.Fanins)), truths: make([][]uint64, len(n.Fanins)), fixed: n.Fixed, empty: n.Empty, pis: n.Pis, pos: n.Pos}
	for i := range d.fanins {
		d.fanins[i] = sfm.Sfm_NodeReadFanins(n, i)
		words := 1
		if len(d.fanins[i]) > 6 {
			words = 1 << uint(len(d.fanins[i])-6)
		}
		d.truths[i] = sfm.Sfm_NodeReadTruth(n, i)[:words]
	}
	return giaMfsInsert(p, d, allBoxes != 0)
}

// Gia_ManPerformMfs optimizes the mapped network and retains timing metadata.
func Gia_ManPerformMfs(p *Gia_Man_t, pars *sfm.Sfm_Par_t) *Gia_Man_t {
	// As ABC asserts: the &mfs command refuses sequential AIGs (CBUGS_FIXED.md #7).
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
		sfm.Sfm_NtkFree(n)
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
	sfm.Sfm_NtkFree(n)
	return result
}
