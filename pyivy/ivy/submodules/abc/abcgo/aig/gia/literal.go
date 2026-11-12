package gia

func Gia_ManConst0Lit() int32            { return 0 }
func Gia_ManConst1Lit() int32            { return 1 }
func Gia_ManIsConst0Lit(lit int32) int32 { return giaBool(lit == 0) }
func Gia_ManIsConst1Lit(lit int32) int32 { return giaBool(lit == 1) }
func Gia_ManIsConstLit(lit int32) int32  { return giaBool(lit <= 1) }

const GIA_NONE = objectNoFanin
const GIA_VOID = 0x0FFFFFFF
