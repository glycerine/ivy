package gia

import "github.com/glycerine/abcgo/aig/aig"

func Gia_ManSeqCleanupClasses(p *Gia_Man_t, constant, equivalent, verbose int32) {
	source := Gia_ManToAigSimple(p)
	cleaned := aig.Aig_ManScl(source, constant, equivalent, 0, -1, -1, verbose, 0)
	Gia_ManReprFromAigRepr(source, p)
	aig.Aig_ManStop(cleaned)
	aig.Aig_ManStop(source)
}
