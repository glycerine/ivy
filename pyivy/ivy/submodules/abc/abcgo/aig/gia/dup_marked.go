package gia

func Gia_ManDupMarked(src *Gia_Man_t) *Gia_Man_t {
	if src.pObjs[0].fMark0 != 0 {
		panic("gia: marked duplication must retain constant zero")
	}
	kept := 0
	for _, obj := range src.pObjs {
		if obj.fMark0 == 0 {
			kept++
		}
	}
	Gia_ManFillValue(src)
	dst := Gia_ManStart(int32(kept))
	dst.CopyIdentityFrom(src)
	dst.fGiaSimple = src.fGiaSimple
	if src.PMuxes != nil {
		dst.EnableMuxes()
	}
	src.pObjs[0].Value = 0
	objectMap := make([]int32, len(src.pObjs))
	for i := range objectMap {
		objectMap[i] = -1
	}
	objectMap[0] = 0
	inputs, outputs := make([]int32, 0), make([]int32, 0)
	var registers, inputRegisters []int32
	var ros, ris int32
	oldPOs := Gia_ManPoNum(src)
	ordinary := oldPOs - src.nConstrs - src.aigerBadOutputs
	for _, obj := range src.pObjs[1:] {
		if obj.fMark0 != 0 {
			obj.fMark0 = 0
			continue
		}
		a, b, c, kind := src.Gate(obj.id)
		var literal int32
		switch kind {
		case 1:
			literal = Gia_ManAppendCi(dst)
			index := Gia_ObjCioId(obj)
			inputs = append(inputs, index)
			if Gia_ObjIsRo(src, obj) != 0 {
				ros++
				registers = append(registers, index-Gia_ManPiNum(src))
			}
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
			index := Gia_ObjCioId(obj)
			outputs = append(outputs, index)
			if Gia_ObjIsRi(src, obj) != 0 {
				ris++
				inputRegisters = append(inputRegisters, index-oldPOs)
			} else if index >= oldPOs-src.nConstrs {
				dst.nConstrs++
			} else if index >= ordinary {
				dst.aigerBadOutputs++
			}
		default:
			panic("gia: unsupported marked duplicate object")
		}
		obj.Value = uint32(literal)
		objectMap[obj.id] = literal >> 1
	}
	if ros != ris {
		panic("gia: marked duplication must retain complete RO/RI pairs")
	}
	for i, index := range registers {
		if inputRegisters[i] != index {
			panic("gia: marked duplication retained mismatched register pairs")
		}
	}
	Gia_ManSetRegNum(dst, ros)
	if src.PReprs != nil && src.PNexts != nil {
		if err := dst.TransferRepresentativesFrom(src, objectMap); err != nil {
			panic(err)
		}
	}
	if src.pSibls != nil {
		dst.pSibls = make([]int32, dst.nObjs)
		for old, mapped := range objectMap {
			if mapped < 0 {
				continue
			}
			sibling := src.pSibls[old]
			if sibling != 0 && objectMap[sibling] >= 0 {
				if objectMap[sibling] >= mapped {
					panic("gia: mapped sibling must precede its node")
				}
				dst.pSibls[mapped] = objectMap[sibling]
			}
		}
	}
	if err := dst.CopyIONamesFrom(src, inputs, outputs); err != nil {
		panic(err)
	}
	subset := func(values *Vec_Int_t) *Vec_Int_t {
		if values == nil {
			return nil
		}
		result := Vec_IntAlloc(len(registers))
		for _, index := range registers {
			Vec_IntPush(result, Vec_IntEntry(values, int(index)))
		}
		return result
	}
	dst.vRegClasses, dst.vRegInits, dst.VFlopClasses = subset(src.vRegClasses), subset(src.vRegInits), subset(src.VFlopClasses)
	if src.aigerLatchInits != nil {
		for _, index := range registers {
			dst.aigerLatchInits = append(dst.aigerLatchInits, src.aigerLatchInits[index])
		}
	}
	dst.aigerComment = append([]byte(nil), src.aigerComment...)
	return dst
}
