package if_

import (
	"fmt"
	abcutil "github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/opt/dau"
	"os"
)

func If_DsdManCollect_rec(p *If_DsdMan_t, id int32, nodes, firsts *Vec_Int_t, support *int32) {
	o := If_DsdVecObj(&p.vObjs, id)
	if o.Type == IF_DSD_CONST0 {
		return
	}
	if o.Type == IF_DSD_VAR {
		*support++
		return
	}
	first := *support
	for _, fan := range o.pFans {
		If_DsdManCollect_rec(p, fan>>1, nodes, firsts, support)
	}
	Vec_IntPush(nodes, id)
	Vec_IntPush(firsts, first)
}
func If_DsdManCollect(p *If_DsdMan_t, id int32, nodes, firsts *Vec_Int_t) {
	vec.Vec_IntClear(nodes)
	vec.Vec_IntClear(firsts)
	var support int32
	If_DsdManCollect_rec(p, id, nodes, firsts, &support)
}
func If_DsdSign_rec(p *If_DsdMan_t, o *If_DsdObj_t, support *int32) uint32 {
	if o.Type == IF_DSD_VAR {
		result := uint32(1) << uint(2**support)
		*support++
		return result
	}
	var result uint32
	for _, fan := range o.pFans {
		result |= If_DsdSign_rec(p, dsdObject(p, fan), support)
	}
	return result
}
func If_DsdSign(p *If_DsdMan_t, o *If_DsdObj_t, fan, first, shared int32) uint32 {
	sign := If_DsdSign_rec(p, dsdObject(p, o.pFans[fan]), &first)
	if shared != 0 {
		return sign<<1 | sign
	}
	return sign
}
func If_DsdManGetSuppSizes(p *If_DsdMan_t, o *If_DsdObj_t, sizes []int32) {
	for i, fan := range o.pFans {
		sizes[i] = If_DsdVecLitSuppSize(&p.vObjs, fan)
	}
}
func ifDsdCheckCombination(p *If_DsdMan_t, o *If_DsdObj_t, first, limit, lut, derive int32, mask uint32, indices *[4]int32, depth, start, count int32, sizes, firsts []int32) uint32 {
	if depth == count {
		var inside int32
		for _, i := range indices[:count] {
			inside += sizes[i]
		}
		if inside > lut || o.nSupp-inside > limit {
			return 0
		}
		if derive == 0 {
			return ^uint32(0)
		}
		If_DsdManComputeFirst(p, o, firsts)
		var sign uint32
		for _, i := range indices[:count] {
			sign |= If_DsdSign(p, o, i, first+firsts[i], 0)
		}
		if sign&mask != 0 {
			return 0
		}
		return sign
	}
	for i := start; i < o.nFans; i++ {
		indices[depth] = i
		if sign := ifDsdCheckCombination(p, o, first, limit, lut, derive, mask, indices, depth+1, i+1, count, sizes, firsts); sign != 0 {
			return sign
		}
	}
	return 0
}
func If_DsdManCheckAndXor(p *If_DsdMan_t, first int32, mask uint32, o *If_DsdObj_t, total, lut, derive, verbose int32) uint32 {
	assert(o.nFans > 2 && o.nSupp > lut, "nontrivial associative DSD decomposition")
	var sizes, firsts [16]int32
	var indices [4]int32
	If_DsdManGetSuppSizes(p, o, sizes[:])
	limit := lut - (total - o.nSupp + 1)
	for count := int32(2); count <= 4 && count < o.nFans; count++ {
		if sign := ifDsdCheckCombination(p, o, first, limit, lut, derive, mask, &indices, 0, 0, count, sizes[:], firsts[:]); sign != 0 {
			return sign
		}
	}
	return 0
}
func If_DsdManCheckMux(p *If_DsdMan_t, first int32, mask uint32, o *If_DsdObj_t, total, lut, derive, verbose int32) uint32 {
	assert(o.nFans == 3 && o.nSupp > lut, "nontrivial DSD mux decomposition")
	var sizes, firsts [16]int32
	If_DsdManGetSuppSizes(p, o, sizes[:])
	limit := lut - (total - o.nSupp + 1)
	for _, fan := range [...]int32{1, 2} {
		if sizes[0]+sizes[fan] > lut || sizes[0]+sizes[3-fan]+1 > limit {
			continue
		}
		if derive == 0 {
			return ^uint32(0)
		}
		If_DsdManComputeFirst(p, o, firsts[:])
		sign := If_DsdSign(p, o, 0, first+firsts[0], 1) | If_DsdSign(p, o, fan, first+firsts[fan], 0)
		if sign&mask == 0 {
			return sign
		}
	}
	return 0
}
func If_DsdManCheckPrime(p *If_DsdMan_t, first int32, mask uint32, o *If_DsdObj_t, total, lut, derive, verbose int32) uint32 {
	assert(o.nFans > 2 && o.nSupp > lut, "nontrivial prime DSD decomposition")
	sets := Vec_PtrEntry(p.vTtDecs[o.nFans], int(If_DsdObjTruthId(p, o))).(*Vec_Int_t)
	if verbose != 0 {
		fmt.Println()
		dau.Dau_DecPrintSets(sets, o.nFans)
	}
	var sizes, firsts [16]int32
	If_DsdManGetSuppSizes(p, o, sizes[:])
	limit := lut - (total - o.nSupp + 1)
	for _, raw := range Vec_IntArray(sets) {
		set := uint32(raw)
		var inside, outside int32
		for v := int32(0); v < o.nFans; v++ {
			switch set >> uint(2*v) & 3 {
			case 0:
				outside += sizes[v]
			case 1:
				inside += sizes[v]
			case 3:
				inside += sizes[v]
				outside += sizes[v]
			default:
				panic("if: invalid prime DSD partition")
			}
		}
		if inside > lut || outside > limit {
			continue
		}
		if derive == 0 {
			return ^uint32(0)
		}
		If_DsdManComputeFirst(p, o, firsts[:])
		var sign uint32
		for v := int32(0); v < o.nFans; v++ {
			switch set >> uint(2*v) & 3 {
			case 1:
				sign |= If_DsdSign(p, o, v, first+firsts[v], 0)
			case 3:
				sign |= If_DsdSign(p, o, v, first+firsts[v], 1)
			}
		}
		if sign&mask == 0 {
			return sign
		}
	}
	return 0
}
func If_DsdManCheckXY_int(p *If_DsdMan_t, lit, lut, derive int32, mask uint32, verbose int32) uint32 {
	root := dsdObject(p, lit)
	if verbose != 0 {
		If_DsdManPrintOne(os.Stdout, p, root.Id, nil, 0)
	}
	if root.nSupp <= lut {
		if verbose != 0 {
			fmt.Println("    Trivial")
		}
		return ^uint32(0)
	}
	If_DsdManCollect(p, root.Id, p.vTemp1, p.vTemp2)
	for i, id := range Vec_IntArray(p.vTemp1) {
		o := If_DsdVecObj(&p.vObjs, id)
		if o.nSupp <= lut && root.nSupp-o.nSupp <= lut-1 {
			if verbose != 0 {
				fmt.Print("    Dec using node ")
				If_DsdManPrintOne(os.Stdout, p, id, nil, 1)
			}
			first := Vec_IntEntry(p.vTemp2, i)
			sign := If_DsdSign_rec(p, o, &first)
			if sign&mask == 0 {
				return sign
			}
		}
	}
	for _, kind := range [...]If_DsdType_t{IF_DSD_AND, IF_DSD_MUX, IF_DSD_PRIME} {
		for i, id := range Vec_IntArray(p.vTemp1) {
			o := If_DsdVecObj(&p.vObjs, id)
			if o.nSupp <= lut {
				continue
			}
			first := Vec_IntEntry(p.vTemp2, i)
			var sign uint32
			switch {
			case kind == IF_DSD_AND && (o.Type == IF_DSD_AND || o.Type == IF_DSD_XOR) && o.nFans > 2:
				sign = If_DsdManCheckAndXor(p, first, mask, o, root.nSupp, lut, derive, verbose)
			case kind == IF_DSD_MUX && o.Type == IF_DSD_MUX:
				sign = If_DsdManCheckMux(p, first, mask, o, root.nSupp, lut, derive, verbose)
			case kind == IF_DSD_PRIME && o.Type == IF_DSD_PRIME:
				sign = If_DsdManCheckPrime(p, first, mask, o, root.nSupp, lut, derive, verbose)
			}
			if sign != 0 {
				if verbose != 0 {
					fmt.Print("    ")
					switch kind {
					case IF_DSD_AND:
						abcutil.Abc_TtPrintBinary([]word{word(sign)}, 4)
						fmt.Println("    Using multi-input AND/XOR node")
					case IF_DSD_MUX:
						abcutil.Abc_TtPrintBinary([]word{word(sign)}, 4)
						fmt.Println("    Using multi-input MUX node")
					case IF_DSD_PRIME:
						dau.Dau_DecPrintSet(sign, o.nFans, 0)
						fmt.Println("    Using prime node")
					}
				}
				return sign
			}
		}
	}
	if verbose != 0 {
		fmt.Println("    UNDEC")
	}
	return 0
}
