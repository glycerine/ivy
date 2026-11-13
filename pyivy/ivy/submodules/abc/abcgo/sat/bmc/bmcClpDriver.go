package bmc

import (
	"bytes"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

// Bmc_CollapseOneOld2 computes both polarities and returns the smaller cover.
func Bmc_CollapseOneOld2(p *gia.Gia_Man_t, cubeLimit, conflictLimit, canonical, reverse, verbose int32) *vec.Vec_Str_t {
	if p == nil || gia.Gia_ManCoNum(p) != 1 {
		return nil
	}
	on := Bmc_CollapseOneInt2(p, cubeLimit, conflictLimit, canonical, reverse, verbose, 0)
	limit := cubeLimit
	if on != nil {
		n := int32(bytes.Count(vec.Vec_StrArray(on), []byte{'\n'}))
		if limit > 0 && n < limit {
			limit = n
		}
	}
	inverse := gia.Gia_ManDupDfs(p)
	gia.Gia_ObjFlipFaninC0(gia.Gia_ManCo(inverse, 0))
	off := Bmc_CollapseOneInt2(inverse, limit, conflictLimit, canonical, reverse, verbose, 1)
	if on == nil {
		return off
	}
	if off == nil {
		return on
	}
	if bytes.Count(vec.Vec_StrArray(on), []byte{'\n'}) <= bytes.Count(vec.Vec_StrArray(off), []byte{'\n'}) {
		return on
	}
	return off
}
func Bmc_CollapseOneOld(p *gia.Gia_Man_t, cubeLimit, conflictLimit, canonical, reverse, verbose int32) *vec.Vec_Str_t {
	return Bmc_CollapseOneOld2(p, cubeLimit, conflictLimit, canonical, reverse, verbose)
}
func Bmc_CollapseOne3(p *gia.Gia_Man_t, cubeLimit, conflictLimit, canonical, reverse, verbose int32) *vec.Vec_Str_t {
	return Bmc_CollapseOneOld2(p, cubeLimit, conflictLimit, canonical, reverse, verbose)
}
func Bmc_CollapseOne(p *gia.Gia_Man_t, cubeLimit, conflictLimit, canonical, reverse, verbose int32) *vec.Vec_Str_t {
	return Bmc_CollapseOneOld2(p, cubeLimit, conflictLimit, canonical, reverse, verbose)
}
