package cba

// These predicates preserve the inclusive enum ranges in cba.h.
func Cba_TypeIsBox(kind Cba_ObjType_t) int32 {
	return cbaBool(kind >= CBA_OBJ_BOX && kind < CBA_BOX_LAST)
}

func Cba_TypeIsSeq(kind Cba_ObjType_t) int32 {
	return cbaBool(kind >= CBA_BOX_RAM && kind <= CBA_BOX_DFFRS)
}

func Cba_TypeIsUnary(kind Cba_ObjType_t) int32 {
	return cbaBool(kind == CBA_BOX_BUF || kind == CBA_BOX_INV || kind == CBA_BOX_LNOT ||
		kind == CBA_BOX_MIN || kind == CBA_BOX_SQRT || kind == CBA_BOX_ABS ||
		kind >= CBA_BOX_RAND && kind <= CBA_BOX_RXNOR)
}

func Cba_TypeIsMux(kind Cba_ObjType_t) int32 {
	return cbaBool(kind == CBA_BOX_MUX || kind == CBA_BOX_NMUX || kind == CBA_BOX_SEL || kind == CBA_BOX_PSEL)
}

func cbaBool(value bool) int32 {
	if value {
		return 1
	}
	return 0
}
