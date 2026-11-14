package if_

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/opt/dau"
)

// Id_DsdManTuneStr1 matches stored functions to a programmable cell and saves
// each successful input permutation and programmable truth table.
func Id_DsdManTuneStr1(p *If_DsdMan_t, structure string, conflicts, verbose int32) {
	network := Ifn_NtkParse(structure)
	if p == nil || network == nil {
		return
	}
	if p.nVars > network.nInps {
		fmt.Printf("The support of DSD manager (%d) exceeds the support of the structure (%d).\n", p.nVars, network.nInps)
		return
	}
	if p.nVars < network.nInps {
		fmt.Printf("Warning: The support of DSD manager (%d) is less than the support of the structure (%d).\n", p.nVars, network.nInps)
	}
	bits := Ifn_NtkTtBits(structure)
	words := int32(1) + (bits+63)/64
	start := p.nObjsPrev
	if p.pCellStr != structure || p.nConfigWords != words {
		p.vConfigs = nil
		start = 0
	}
	p.pCellStr, p.nTtBits, p.nConfigWords = structure, bits, words
	if start < 0 || start > If_DsdManObjNum(p) {
		panic("if: invalid DSD tuning bookmark")
	}
	if p.vConfigs == nil {
		p.vConfigs = vec.Vec_WrdStart(int(words * If_DsdManObjNum(p)))
	} else {
		vec.Vec_WrdFillExtra(p.vConfigs, int(words*If_DsdManObjNum(p)), 0)
	}
	data := vec.Vec_WrdArray(p.vConfigs)
	if verbose != 0 {
		fmt.Printf("Considering programmable cell: ")
		Ifn_NtkPrint(network)
	}
	for id := start; id < If_DsdManObjNum(p); id++ {
		obj := If_DsdVecObj(&p.vObjs, id)
		obj.fMark = 0
		config := data[id*words : (id+1)*words]
		clear(config)
		input := append([]word(nil), obj.truth...)
		if len(input) == 0 {
			obj.fMark = 1
			continue
		}
		if obj.Type == IF_DSD_VAR && obj.nFans == 0 {
			input = []word{0xaaaaaaaaaaaaaaaa}
		}
		if obj.nFans > 0 {
			vars := make([]int32, obj.nFans)
			if dau.Dau_DsdMinBase(input, obj.nFans, vars) != obj.nSupp {
				panic("if: DSD support mismatch")
			}
		}
		if Ifn_NtkMatch(network, input, obj.nSupp, conflicts, verbose, 0, config) == 0 {
			obj.fMark = 1
			clear(config)
		}
	}
	p.nObjsPrev = 0
	p.LutSize = 0
}

// If_DsdManTune follows ifDsd.c's mark/retry protocol for two-LUT matching.
// spec is retained because the source entry point accepts but does not use it.
func If_DsdManTune(p *If_DsdMan_t, lutSize, fast, add, spec, verbose int32) {
	if add == 0 || lutSize == 0 {
		for id := int32(0); id < If_DsdManObjNum(p); id++ {
			If_DsdVecObj(&p.vObjs, id).fMark = 0
		}
	}
	if lutSize == 0 {
		return
	}
	scratch := vec.Vec_IntAlloc(1000)
	defer vec.Vec_IntFree(scratch)
	solver := If_ManSatBuildXY(lutSize)
	defer If_ManSatUnbuild(solver)
	for id := int32(0); id < If_DsdManObjNum(p); id++ {
		object := If_DsdVecObj(&p.vObjs, id)
		if object.nSupp <= lutSize || add != 0 && object.fMark == 0 {
			continue
		}
		object.fMark = 0
		if If_DsdManCheckXY(p, id<<1, lutSize, 0, 0, 0, 0) != 0 {
			continue
		}
		if fast == 0 {
			truth := append([]word(nil), object.truth...)
			dau.Dau_DsdMinBase(truth, object.nFans, make([]int32, object.nFans))
			if If_ManSatCheckXYall(solver, lutSize, truth, object.nSupp, scratch) != 0 {
				continue
			}
		}
		object.fMark = 1
	}
	if verbose != 0 {
		marked := make([]int, p.nVars+1)
		for id := int32(0); id < If_DsdManObjNum(p); id++ {
			obj := If_DsdVecObj(&p.vObjs, id)
			if obj.fMark != 0 {
				marked[obj.nSupp]++
			}
		}
		fmt.Printf("DSD functions excluded by support: %v\n", marked)
	}
}

// The portable C build uses the serial matcher for this entry point.
func Id_DsdManTuneStr(p *If_DsdMan_t, structure string, conflicts, processes, verbose int32) {
	Id_DsdManTuneStr1(p, structure, conflicts, verbose)
}
