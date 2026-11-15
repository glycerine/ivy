package gia

import (
	"fmt"

	"github.com/glycerine/abcgo/bool/bdc"
)

type Bdc_Man_t = bdc.Bdc_Man_t

// Truth storage uses int32 words in the original vector API. Return copied
// uint32 words instead of reinterpreting signed slice storage with unsafe.
func giaBidecTruthWords(storage *Vec_Int_t, offset, words int) []uint32 {
	result := make([]uint32, words)
	for i, value := range Vec_IntArray(storage)[offset : offset+words] {
		result[i] = uint32(value)
	}
	return result
}

// Offsets survive recursive growth of the signed scratch vector. Only the
// public result is copied to uint32 storage, so callers own it independently.
func Gia_ManConvertAigToTruth_rec(p *Gia_Man_t, obj *Gia_Obj_t, storage *Vec_Int_t, words int32, visited *Vec_Int_t) []uint32 {
	if Gia_IsComplement(obj) != 0 {
		panic("gia: truth recursion expects regular object")
	}
	offset := giaBidecTruthOffset(p, obj, storage, int(words), visited)
	if offset < 0 {
		return make([]uint32, words)
	}
	return giaBidecTruthWords(storage, offset, int(words))
}
func giaBidecReadWord(storage []int32, offset, word int, lit int32) uint32 {
	var value uint32
	if offset >= 0 {
		value = uint32(storage[offset+word])
	}
	if lit&1 != 0 {
		return ^value
	}
	return value
}
func giaBidecTruthOffset(p *Gia_Man_t, obj *Gia_Obj_t, storage *Vec_Int_t, words int, visited *Vec_Int_t) int {
	if index := Vec_IntGetEntryFull(p.vTruths, int(obj.id)); index != -1 {
		return words * int(index)
	}
	if obj.id == 0 {
		return -1
	}
	a, b, c, kind := p.Gate(obj.id)
	if kind == 1 {
		panic("gia: truth cut does not cover input")
	}
	offset0 := giaBidecTruthOffset(p, p.pObjs[a>>1], storage, words, visited)
	offset1, offset2 := -1, -1
	if kind == 2 || kind == 3 || kind == 4 {
		offset1 = giaBidecTruthOffset(p, p.pObjs[b>>1], storage, words, visited)
	}
	if kind == 4 {
		offset2 = giaBidecTruthOffset(p, p.pObjs[c>>1], storage, words, visited)
	}
	offset := Vec_IntSize(storage)
	Vec_IntFillExtra(storage, offset+words, 0)
	// Reacquire the array after all recursion and growth, avoiding stale views.
	data := Vec_IntArray(storage)
	result := data[offset : offset+words]
	for i := range result {
		x := giaBidecReadWord(data, offset0, i, a)
		var value uint32
		switch kind {
		case 2:
			value = x & giaBidecReadWord(data, offset1, i, b)
		case 3:
			value = x ^ giaBidecReadWord(data, offset1, i, b)
		case 4:
			value = x&giaBidecReadWord(data, offset1, i, b) | ^x&giaBidecReadWord(data, offset2, i, c)
		case 5, 6:
			value = x
		default:
			panic("gia: invalid truth gate")
		}
		result[i] = int32(value)
	}
	Vec_IntSetEntry(p.vTruths, int(obj.id), int32(Vec_IntSize(visited)))
	Vec_IntPush(visited, obj.id)
	return offset
}
func Gia_ManConvertAigToTruth(p *Gia_Man_t, root *Gia_Obj_t, leaves, storage, visited *Vec_Int_t) []uint32 {
	variables := Vec_IntSize(leaves)
	if variables > 15 {
		panic("gia: bidecomposition truth supports at most 15 inputs")
	}
	words := 1
	if variables > 5 {
		words = 1 << uint(variables-5)
	}
	if p.vTruths == nil {
		Gia_ManCleanTruth(p)
	}
	Vec_IntClear(storage)
	Vec_IntClear(visited)
	elementary := [...]uint32{0xaaaaaaaa, 0xcccccccc, 0xf0f0f0f0, 0xff00ff00, 0xffff0000}
	for variable, id := range Vec_IntArray(leaves) {
		if id <= 0 || id >= p.nObjs {
			panic("gia: invalid cut leaf")
		}
		index := Vec_IntSize(visited)
		for word := 0; word < words; word++ {
			var pattern uint32
			if variable < 5 {
				pattern = elementary[variable]
			} else if word&(1<<uint(variable-5)) != 0 {
				pattern = ^uint32(0)
			}
			Vec_IntPush(storage, int32(pattern))
		}
		Vec_IntSetEntry(p.vTruths, int(id), int32(index))
		Vec_IntPush(visited, id)
	}
	result := Gia_ManConvertAigToTruth_rec(p, Gia_Regular(root), storage, int32(words), visited)
	if Gia_IsComplement(root) != 0 {
		for i := range result {
			result[i] = ^result[i]
		}
	}
	for _, id := range Vec_IntArray(visited) {
		Vec_IntSetEntry(p.vTruths, int(id), -1)
	}
	return result
}
func Gia_ObjPerformBidec(manager *Bdc_Man_t, dst, src *Gia_Man_t, root *Gia_Obj_t, leaves, storage, visited *Vec_Int_t) int32 {
	Vec_IntClear(leaves)
	for _, id := range Gia_ObjLutFanins(src, root.id) {
		Vec_IntPush(leaves, id)
	}
	variables := int32(Vec_IntSize(leaves))
	if variables > 15 {
		panic("gia: LUT exceeds bidecomposition limit")
	}
	truth := Gia_ManConvertAigToTruth(src, root, leaves, storage, visited)
	zero, one := true, true
	for _, word := range truth {
		zero = zero && word == 0
		one = one && word == ^uint32(0)
	}
	if zero {
		return 0
	}
	if one {
		return 1
	}
	if bdc.Bdc_ManDecompose(manager, truth, nil, variables, nil, 1000) < 0 {
		panic("gia: bidecomposition failed")
	}
	bdc.Bdc_FuncSetCopyInt(bdc.Bdc_ManFunc(manager, 0), 1)
	for i, id := range Vec_IntArray(leaves) {
		bdc.Bdc_FuncSetCopyInt(bdc.Bdc_ManFunc(manager, int32(i)+1), int32(src.pObjs[id].Value))
	}
	for i := variables + 1; i < bdc.Bdc_ManNodeNum(manager); i++ {
		function := bdc.Bdc_ManFunc(manager, i)
		bdc.Bdc_FuncSetCopyInt(function, Gia_ManHashAnd(dst, bdc.Bdc_FunFanin0Copy(function), bdc.Bdc_FunFanin1Copy(function)))
	}
	return bdc.Bdc_FunObjCopy(bdc.Bdc_ManRoot(manager))
}
func Gia_ManPerformBidec(src *Gia_Man_t, verbose int32) *Gia_Man_t {
	maximum := Gia_ManLutSizeMax(src)
	if maximum < 2 {
		fmt.Println("Resynthesis is not performed when nodes have less than 2 inputs.")
		return nil
	}
	if maximum > 15 {
		fmt.Println("Resynthesis is not performed when nodes have more than 15 inputs.")
		return nil
	}
	manager := bdc.Bdc_ManAlloc(&bdc.Bdc_Par_t{NVarsMax: maximum, FVerbose: verbose})
	defer bdc.Bdc_ManFree(manager)
	leaves, storage, visited := Vec_IntAlloc(0), Vec_IntAlloc(1<<16), Vec_IntAlloc(0)
	Gia_ManCleanTruth(src)
	Gia_ManFillValue(src)
	src.pObjs[0].Value = 0
	dst := Gia_ManStart(src.nObjs)
	dst.CopyIdentityFrom(src)
	Gia_ManHashAlloc(dst)
	for _, obj := range src.pObjs[1:] {
		switch {
		case Gia_ObjIsCi(obj) != 0:
			obj.Value = uint32(Gia_ManAppendCi(dst))
		case Gia_ObjIsCo(obj) != 0:
			obj.Value = uint32(Gia_ManAppendCo(dst, Gia_ObjFanin0Copy(obj)))
		case Gia_ObjIsLut(src, obj.id) != 0:
			obj.Value = uint32(Gia_ObjPerformBidec(manager, dst, src, obj, leaves, storage, visited))
		}
	}
	Gia_ManDupRemapCis(dst, src)
	Gia_ManDupRemapCos(dst, src)
	Gia_ManHashStop(dst)
	Gia_ManSetRegNum(dst, src.nRegs)
	giaCopyInterfaceMetadata(dst, src)
	result := Gia_ManCleanup(dst)
	if Gia_ManAndNum(result) != Gia_ManAndNum(dst) {
		fmt.Printf("Gia_ManPerformBidec() node count before and after: %6d and %6d.\n", Gia_ManAndNum(dst), Gia_ManAndNum(result))
	}
	Gia_ManStop(dst)
	return result
}
