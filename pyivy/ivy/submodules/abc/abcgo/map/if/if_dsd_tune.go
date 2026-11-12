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
