package gia

import (
	"strings"
	"time"
)

func Gia_TimeStamp() string { return time.Now().Format("Mon Jan _2 15:04:05 2006") }
func Gia_FileNameGenericAppend(base, suffix string) string {
	if index := strings.LastIndex(base, "."); index >= 0 {
		base = base[:index]
	}
	base += suffix
	// Retain both C path conventions independently of the host OS.
	if index := strings.LastIndex(base, "\\"); index >= 0 {
		return base[index+1:]
	}
	if index := strings.LastIndex(base, "/"); index >= 0 {
		return base[index+1:]
	}
	return base
}
func Gia_ManSetMark0(p *Gia_Man_t) {
	for _, obj := range p.pObjs {
		obj.fMark0 = 1
	}
}
func Gia_ManSetMark1(p *Gia_Man_t) {
	for _, obj := range p.pObjs {
		obj.fMark1 = 1
	}
}
func Gia_ManCheckMark0(p *Gia_Man_t) {
	for _, obj := range p.pObjs {
		if obj.fMark0 != 0 {
			panic("gia: mark0 is not clear")
		}
	}
}
func Gia_ManCheckMark1(p *Gia_Man_t) {
	for _, obj := range p.pObjs {
		if obj.fMark1 != 0 {
			panic("gia: mark1 is not clear")
		}
	}
}
func Gia_ManCheckCoPhase(p *Gia_Man_t) int32 {
	var count int32
	for i := int32(0); i < Gia_ManCoNum(p); i++ {
		count += Gia_ObjPhase(Gia_ManCo(p, i))
	}
	return count
}
func Gia_ManCleanTruth(p *Gia_Man_t) {
	if p.vTruths == nil {
		p.vTruths = Vec_IntAlloc(int(p.nObjs))
	}
	Vec_IntFill(p.vTruths, int(p.nObjs), -1)
}
func Gia_ManCreateValueRefs(p *Gia_Man_t) {
	Gia_ManCleanValue(p)
	for _, obj := range p.pObjs {
		for _, id := range giaObjectFanins(p, obj) {
			p.pObjs[id].Value++
		}
	}
}
func Gia_ManMarkDangling(p *Gia_Man_t) int32 {
	Gia_ManMarkFanoutDrivers(p)
	var count int32
	for _, obj := range p.pObjs {
		if Gia_ObjIsAnd(obj) != 0 && obj.fMark0 == 0 {
			count++
		}
	}
	return count
}
func Gia_ManHasDangling(p *Gia_Man_t) int32 {
	count := Gia_ManMarkDangling(p)
	Gia_ManCleanMark0(p)
	return count
}
func Gia_ManGetDangling(p *Gia_Man_t) *Vec_Int_t {
	Gia_ManMarkDangling(p)
	result := Vec_IntAlloc(100)
	for _, obj := range p.pObjs {
		if Gia_ObjIsAnd(obj) != 0 && obj.fMark0 == 0 {
			Vec_IntPush(result, obj.id)
		}
	}
	Gia_ManCleanMark0(p)
	return result
}
func Gia_ManInvertConstraints(p *Gia_Man_t) {
	for i := Gia_ManPoNum(p) - Gia_ManConstrNum(p); i < Gia_ManPoNum(p); i++ {
		Gia_ObjFlipFaninC0(Gia_ManPo(p, i))
	}
}
func Gia_ManInvertPos(p *Gia_Man_t) {
	for i := int32(0); i < Gia_ManPoNum(p); i++ {
		Gia_ObjFlipFaninC0(Gia_ManPo(p, i))
	}
}
