package gia

import "github.com/glycerine/abcgo/misc/util"

// Gia_ManDup copies objects in source order and writes their new literals to
// source Value scratch fields. Real gates retain their actual logic and tags.
func Gia_ManDup(src *Gia_Man_t) *Gia_Man_t {
	dst := Gia_ManStart(src.nObjs)
	dst.CopyIdentityFrom(src)
	dst.fGiaSimple = src.fGiaSimple
	if src.PMuxes != nil {
		dst.EnableMuxes()
	}
	src.pObjs[0].Value = 0
	for _, obj := range src.pObjs[1:] {
		a, b, c, kind := src.Gate(obj.id)
		var literal int32
		switch kind {
		case 1:
			literal = Gia_ManAppendCi(dst)
		case 2:
			literal = Gia_ManAppendAnd(dst, Gia_ObjLitCopy(src, a), Gia_ObjLitCopy(src, b))
		case 3:
			literal = Gia_ManAppendXorReal(dst, Gia_ObjLitCopy(src, a), Gia_ObjLitCopy(src, b))
		case 4:
			literal = Gia_ManAppendMuxReal(dst, Gia_ObjLitCopy(src, a), Gia_ObjLitCopy(src, b), Gia_ObjLitCopy(src, c))
		case 5:
			literal = Gia_ManAppendBuf(dst, Gia_ObjLitCopy(src, a))
		case 6:
			literal = Gia_ManAppendCo(dst, Gia_ObjLitCopy(src, a))
		default:
			panic("gia: unsupported duplicate object")
		}
		obj.Value = uint32(literal)
	}
	if src.pSibls != nil {
		dst.pSibls = append([]int32(nil), src.pSibls...)
	}
	Gia_ManSetRegNum(dst, src.nRegs)
	dst.nConstrs = src.nConstrs
	if src.pCexSeq != nil {
		dst.pCexSeq = util.Abc_CexDup(src.pCexSeq, src.nRegs)
	}
	// Object IDs are unchanged, so serialized node-indexed I/O metadata remains
	// valid and can be independently owned by the duplicate.
	dst.aigerBadOutputs = src.aigerBadOutputs
	dst.aigerComment = append([]byte(nil), src.aigerComment...)
	dst.aigerLatchInits = append([]int32(nil), src.aigerLatchInits...)
	if src.aigerExtensions != nil {
		dst.aigerExtensions = make(map[byte][]byte, len(src.aigerExtensions))
		for key, data := range src.aigerExtensions {
			dst.aigerExtensions[key] = append([]byte(nil), data...)
		}
	}
	if src.aigerSymbols != nil {
		dst.aigerSymbols = make(map[byte][]string, len(src.aigerSymbols))
		for key, names := range src.aigerSymbols {
			dst.aigerSymbols[key] = append([]string(nil), names...)
		}
	}
	return dst
}
