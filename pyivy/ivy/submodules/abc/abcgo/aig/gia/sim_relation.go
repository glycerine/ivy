package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
)

func Gia_ManSimRelAssignInputs(p *Gia_Man_t, words int32, sims *Vec_Wrd_t, inputWords int32, inputs *Vec_Wrd_t) {
	if inputWords <= 0 || words%inputWords != 0 || vec.Vec_WrdSize(sims) != int(words*p.nObjs) || vec.Vec_WrdSize(inputs) != int(inputWords*Gia_ManCiNum(p)) {
		panic("gia: invalid relation input dimensions")
	}
	dst, src := vec.Vec_WrdArray(sims), vec.Vec_WrdArray(inputs)
	for i, id := range vec.Vec_IntArray(p.vCis) {
		for m := int32(0); m < words/inputWords; m++ {
			start := id*words + m*inputWords
			copy(dst[start:start+inputWords], src[int32(i)*inputWords:(int32(i)+1)*inputWords])
		}
	}
}
func Gia_ManSimRelCompare(p *Gia_Man_t, words int32, sims *Vec_Wrd_t, outWords int32, outputs *Vec_Wrd_t, pattern, mint int32) int32 {
	data, want := vec.Vec_WrdArray(sims), vec.Vec_WrdArray(outputs)
	for i, id := range vec.Vec_IntArray(p.vCos) {
		start := id*words + mint*outWords
		expected := int32(i) * outWords
		if giaSimBit(data[start:start+outWords], pattern) != giaSimBit(want[expected:expected+outWords], pattern) {
			return 0
		}
	}
	return 1
}
func Gia_ManSimRelCollectOutputs(p *Gia_Man_t, words int32, sims *Vec_Wrd_t, outWords int32, outputs, relation *Vec_Wrd_t) int32 {
	if outWords <= 0 || words%outWords != 0 || vec.Vec_WrdSize(sims) != int(words*p.nObjs) || vec.Vec_WrdSize(outputs) != int(outWords*Gia_ManCoNum(p)) || vec.Vec_WrdSize(relation) != int(words) {
		panic("gia: invalid relation output dimensions")
	}
	mints := words / outWords
	data := vec.Vec_WrdArray(relation)
	var missing int32
	for pattern := int32(0); pattern < 64*outWords; pattern++ {
		var count int32
		for mint := int32(0); mint < mints; mint++ {
			if Gia_ManSimRelCompare(p, words, sims, outWords, outputs, pattern, mint) != 0 {
				bit := pattern*mints + mint
				data[bit>>6] |= uint64(1) << uint(bit&63)
				count++
			}
		}
		if count == 0 {
			missing++
		}
	}
	if missing != 0 {
		fmt.Printf("The relation is not well-defined for %d (out of %d) patterns.\n", missing, 64*outWords)
	}
	return missing
}
func Gia_ManSimRel(p *Gia_Man_t, objects *Vec_Int_t, values *Vec_Wrd_t) *Vec_Wrd_t {
	words := giaPatternWords(p)
	ids := vec.Vec_IntArray(objects)
	if len(ids) >= 31 {
		panic("gia: too many relation variables")
	}
	mints := int32(1) << uint(len(ids))
	width := words * mints
	relation := vec.Vec_WrdStart(int(width))
	sims := vec.Vec_WrdStart(int(p.nObjs * width))
	Gia_ManSimRelAssignInputs(p, width, sims, words, p.VSimsPi)
	data := vec.Vec_WrdArray(sims)
	for i, id := range ids {
		for mint := int32(0); mint < mints; mint++ {
			var value uint64
			if mint&(1<<uint(i)) != 0 {
				value = ^uint64(0)
			}
			for w := int32(0); w < words; w++ {
				data[id*width+mint*words+w] = value
			}
		}
	}
	Gia_ManCleanPhase(p)
	for _, id := range ids {
		Gia_ManObj(p, id).fPhase = 1
	}
	for _, obj := range p.pObjs {
		if Gia_ObjIsAnd(obj) != 0 && obj.fPhase == 0 {
			Gia_ManSimPatSimAnd(p, obj.id, obj, width, sims)
		}
	}
	for _, id := range vec.Vec_IntArray(p.vCos) {
		obj := Gia_ManObj(p, id)
		if obj.fPhase == 0 {
			Gia_ManSimPatSimPo(p, id, obj, width, sims)
		}
	}
	for _, id := range ids {
		Gia_ManObj(p, id).fPhase = 0
	}
	if Gia_ManSimRelCollectOutputs(p, width, sims, words, values, relation) != 0 {
		return nil
	}
	return relation
}
