package gia

import (
	"sort"

	"github.com/glycerine/abcgo/bool/kit"
)

// Gia_ManComputeTruths sorts each mapped LUT's fanins in place and computes
// its truth table, using a common width for all LUTs.
func Gia_ManComputeTruths(p *Gia_Man_t, cutSize, lutNum, reverse int32) *Vec_Wrd_t {
	words := int32(giaTruthWordCount64(cutSize))
	truths := Vec_WrdAlloc(int(words * lutNum))
	Gia_ObjComputeTruthTableStart(p, cutSize)
	defer Gia_ObjComputeTruthTableStop(p)
	for id := int32(1); id < p.nObjs; id++ {
		if Gia_ObjIsLut(p, id) == 0 {
			continue
		}
		fanins := Gia_ObjLutFanins(p, id)
		sort.Slice(fanins, func(i, j int) bool {
			if reverse == 0 {
				return fanins[i] > fanins[j]
			}
			return fanins[i] < fanins[j]
		})
		for i := 1; i < len(fanins); i++ {
			if fanins[i] == fanins[i-1] {
				panic("gia: duplicate LUT fanin")
			}
		}
		leaves := Vec_IntAllocArray(fanins, len(fanins))
		truth := Gia_ObjComputeTruthTableCut(p, p.pObjs[id], leaves)
		for _, value := range truth[:words] {
			Vec_WrdPush(truths, value)
		}
	}
	return truths
}

func Gia_ManAssignNumbers(p *Gia_Man_t) int32 {
	Gia_ManFillValue(p)
	var count int32
	for i := int32(0); i < Gia_ManCiNum(p); i++ {
		Gia_ManCi(p, i).Value = uint32(count)
		count++
	}
	for id := int32(1); id < p.nObjs; id++ {
		if Gia_ObjIsLut(p, id) != 0 {
			p.pObjs[id].Value = uint32(count)
			count++
		}
	}
	return count
}

// Gia_ManFxRetrieve emits one vector per cube: owner number followed by
// complemented-edge literals referring to numbered CIs and LUTs.
func Gia_ManFxRetrieve(p *Gia_Man_t, pCompl **Vec_Str_t, reverse int32) *Vec_Wec_t {
	items := Gia_ManAssignNumbers(p)
	cutSize := Gia_ManLutSizeMax(p)
	if cutSize < 0 {
		panic("gia: extraction requires at least one mapped LUT")
	}
	words := int(giaTruthWordCount64(cutSize))
	truths := Vec_WrdArray(Gia_ManComputeTruths(p, cutSize, items-Gia_ManCiNum(p), reverse))
	cubes := Vec_WecAlloc(1000)
	*pCompl = Vec_StrStart(int(items))
	var cover []uint32
	index := 0
	for id := int32(1); id < p.nObjs; id++ {
		if Gia_ObjIsLut(p, id) == 0 {
			continue
		}
		obj, fanins := p.pObjs[id], Gia_ObjLutFanins(p, id)
		truth := make([]uint32, words*2)
		for i, value := range truths[index : index+words] {
			truth[2*i], truth[2*i+1] = uint32(value), uint32(value>>32)
		}
		index += words
		phase := kit.Kit_TruthIsop(truth, int32(len(fanins)), &cover, 1)
		if phase < 0 {
			panic("gia: unable to compute LUT cover")
		}
		if len(cover) == 0 || len(cover) == 1 && cover[0] == 0 {
			if len(cover) == 0 {
				Vec_StrWriteEntry(*pCompl, int(obj.Value), 1)
			}
			Vec_IntPush(Vec_WecPushLevel(cubes), int32(obj.Value))
			continue
		}
		Vec_StrWriteEntry(*pCompl, int(obj.Value), byte(phase))
		for _, term := range cover {
			cube := Vec_WecPushLevel(cubes)
			Vec_IntPush(cube, int32(obj.Value))
			for i, fanin := range fanins {
				value := (term >> (2 * i)) & 3
				if value == 3 {
					panic("gia: invalid cover literal")
				}
				if value != 0 {
					lit := int32(p.pObjs[fanin].Value) << 1
					if value == 1 {
						lit |= 1
					}
					Vec_IntPush(cube, lit)
				}
			}
			sort.Slice(Vec_IntArray(cube)[1:], func(i, j int) bool {
				return Vec_IntEntry(cube, i+1) < Vec_IntEntry(cube, j+1)
			})
		}
	}
	return cubes
}
