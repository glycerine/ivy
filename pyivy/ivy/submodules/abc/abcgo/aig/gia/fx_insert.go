package gia

import (
	"strings"

)

// Gia_ManFxInsert rebuilds a network from extracted cube covers. Small
// covers use DSD decomposition; larger covers are factored as SOPs.
func Gia_ManFxInsert(p *Gia_Man_t, cubes *Vec_Wec_t, phases *Vec_Str_t) *Gia_Man_t {
	var first, count *Vec_Int_t
	order := Gia_ManFxTopoOrder(cubes, Gia_ManCiNum(p), int32(Vec_StrSize(phases)), &first, &count)
	if order == nil {
		result := Gia_ManDup(p)
		giaCopyInterfaceMetadata(result, p)
		return result
	}
	result := Gia_ManStart(p.nObjs)
	result.pName, result.PSpec = p.pName, p.PSpec
	// As in Gia_ManFxInsert, DSD balancing reads and updates these levels.
	result.vLevels = Vec_IntStart(int(6*Gia_ManObjNum(p)/5 + 100))
	Gia_ManHashStart(result)
	copies := make([]int32, Vec_IntSize(order))
	for i := range copies {
		copies[i] = -1
	}
	for i := int32(0); i < Gia_ManCiNum(p); i++ {
		copies[i] = Gia_ManAppendCi(result)
	}
	index := make([]int, len(copies))
	for i := range index {
		index[i] = -1
	}
	var fanins []int32
	leaves := Vec_IntAlloc(16)
	patterns := [6]uint64{0xaaaaaaaaaaaaaaaa, 0xcccccccccccccccc, 0xf0f0f0f0f0f0f0f0, 0xff00ff00ff00ff00, 0xffff0000ffff0000, 0xffffffff00000000}
	for k := int(Gia_ManCiNum(p)); k < Vec_IntSize(order); k++ {
		owner := Vec_IntEntry(order, k)
		start, n := Vec_IntEntry(first, int(owner)), Vec_IntEntry(count, int(owner))
		fanins = fanins[:0]
		for c := int32(0); c < n; c++ {
			cube := Vec_WecEntry(cubes, int(start+c))
			for j := 1; j < Vec_IntSize(cube); j++ {
				input := Vec_IntEntry(cube, j) >> 1
				if index[input] < 0 {
					index[input] = len(fanins)
					fanins = append(fanins, input)
				}
			}
		}
		// As in the original, an absent cover represents the same constant
		// as an empty cube; the phase flag distinguishes zero from one.
		lit := int32(1)
		if len(fanins) != 0 {
			Vec_IntClear(leaves)
			for _, input := range fanins {
				if copies[input] < 0 {
					panic("gia: extraction cube has unresolved input")
				}
				Vec_IntPush(leaves, copies[input])
			}
			if len(fanins) <= 6 {
				var truth uint64
				for c := int32(0); c < n; c++ {
					term := ^uint64(0)
					cube := Vec_WecEntry(cubes, int(start+c))
					for j := 1; j < Vec_IntSize(cube); j++ {
						input := Vec_IntEntry(cube, j)
						pattern := patterns[index[input>>1]]
						if input&1 != 0 {
							pattern = ^pattern
						}
						term &= pattern
					}
					truth |= term
				}
				if truth == 0 {
					truth = ^truth
				}
				lit = Dsm_ManTruthToGia(result, []word{truth}, Vec_IntArray(leaves))
			} else {
				var sop strings.Builder
				for c := int32(0); c < n; c++ {
					row := []byte(strings.Repeat("-", len(fanins)))
					cube := Vec_WecEntry(cubes, int(start+c))
					for j := 1; j < Vec_IntSize(cube); j++ {
						input := Vec_IntEntry(cube, j)
						char := byte('1')
						if input&1 != 0 {
							char = '0'
						}
						row[index[input>>1]] = char
					}
					sop.Write(row)
					sop.WriteString(" 1\n")
				}
				lit = Gia_ManFactorNode(result, sop.String(), leaves)
			}
		}
		if int(owner) < Vec_StrSize(phases) && Vec_StrEntry(phases, int(owner)) != 0 {
			lit ^= 1
		}
		for _, input := range fanins {
			index[input] = -1
		}
		copies[owner] = lit
	}
	Gia_ManHashStop(result)
	for i := int32(0); i < Gia_ManCoNum(p); i++ {
		obj := Gia_ManCo(p, i)
		driver := Gia_ObjFanin0(obj)
		lit := int32(0)
		if driver.id != 0 {
			lit = copies[driver.Value]
			if lit < 0 {
				panic("gia: extraction output has unresolved driver")
			}
		}
		Gia_ManAppendCo(result, lit^int32(obj.fCompl0))
	}
	Gia_ManSetRegNum(result, Gia_ManRegNum(p))
	result = Gia_ManCleanup(result)
	giaCopyInterfaceMetadata(result, p)
	return result
}

// giaFxDsdToGia builds truth over leaves in p as Gia_ManFxInsert does.
func giaFxDsdToGia(p *Gia_Man_t, truth uint64, leaves []int32) int32 {
	return Dsm_ManTruthToGia(p, []word{truth}, leaves)
}
