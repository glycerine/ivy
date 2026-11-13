package bmc

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

func Bmc_GiaGenerateJustNonRec(p *gia.Gia_Man_t, frame int32, values, just *vec.Vec_Bit_t) {
	n := gia.Gia_ManObjNum(p)
	for f := frame; f >= 0; f-- {
		shift := int(f * n)
		for id := n - 1; id > 0; id-- {
			o := gia.Gia_ManObj(p, id)
			if gia.Gia_ObjIsPi(p, o) != 0 || vec.Vec_BitEntry(just, shift+int(id)) == 0 {
				continue
			}
			if gia.Gia_ObjIsAnd(o) != 0 {
				a, b := gia.Gia_ObjFaninId0(o, id), gia.Gia_ObjFaninId1(o, id)
				if gia.Gia_ObjIsXor(o) != 0 || vec.Vec_BitEntry(values, shift+int(id)) != 0 {
					vec.Vec_BitWriteEntry(just, shift+int(a), 1)
					vec.Vec_BitWriteEntry(just, shift+int(b), 1)
				} else if vec.Vec_BitEntry(values, shift+int(a)) == int(gia.Gia_ObjFaninC0(o)) {
					vec.Vec_BitWriteEntry(just, shift+int(a), 1)
				} else {
					vec.Vec_BitWriteEntry(just, shift+int(b), 1)
				}
			} else if gia.Gia_ObjIsCo(o) != 0 {
				vec.Vec_BitWriteEntry(just, shift+int(gia.Gia_ObjFaninId0(o, id)), 1)
			} else if gia.Gia_ObjIsRo(p, o) != 0 && f > 0 {
				vec.Vec_BitWriteEntry(just, shift-int(n)+int(gia.Gia_ObjRoToRiId(p, id)), 1)
			}
		}
	}
}
func Bmc_GiaGenerateJust_rec(p *gia.Gia_Man_t, frame, id int32, values, just *vec.Vec_Bit_t) int32 {
	if frame < 0 {
		return 0
	}
	vec.Vec_BitWriteEntry(just, int(frame*gia.Gia_ManObjNum(p)+id), 1)
	Bmc_GiaGenerateJustNonRec(p, frame, values, just)
	return 0
}
func Bmc_GiaGenerateJust(p *gia.Gia_Man_t, cex *util.Abc_Cex_t, pvValues, pvJust **vec.Vec_Bit_t) {
	if p == nil || cex == nil || pvValues == nil || pvJust == nil || cex.IFrame < 0 || cex.NRegs != gia.Gia_ManRegNum(p) || cex.NPis != gia.Gia_ManPiNum(p) || cex.IPo < 0 || cex.IPo >= gia.Gia_ManPoNum(p) {
		panic("bmc: invalid counterexample")
	}
	n := gia.Gia_ManObjNum(p)
	values := vec.Vec_BitStart(int(n * (cex.IFrame + 1)))
	just := vec.Vec_BitStart(int(n * (cex.IFrame + 1)))
	state := make([]bool, cex.NRegs)
	depend := make([]bool, cex.NRegs)
	for i := range state {
		state[i] = util.Abc_InfoHasBit(cex.PData, int32(i))
	}
	for frame := int32(0); frame <= cex.IFrame; frame++ {
		v := make([]bool, n)
		j := make([]bool, n)
		j[0] = true
		for i := int32(0); i < cex.NPis; i++ {
			id := gia.Gia_ObjId(p, gia.Gia_ManCi(p, i))
			v[id] = util.Abc_InfoHasBit(cex.PData, cex.NRegs+frame*cex.NPis+i)
			j[id] = true
		}
		for i := int32(0); i < cex.NRegs; i++ {
			id := gia.Gia_ObjId(p, gia.Gia_ManRo(p, i))
			v[id], j[id] = state[i], depend[i]
		}
		for id := int32(1); id < n; id++ {
			o := gia.Gia_ManObj(p, id)
			if gia.Gia_ObjIsAnd(o) != 0 {
				a, b := gia.Gia_ObjFaninId0(o, id), gia.Gia_ObjFaninId1(o, id)
				x, y := v[a] != (gia.Gia_ObjFaninC0(o) != 0), v[b] != (gia.Gia_ObjFaninC1(o) != 0)
				if gia.Gia_ObjIsXor(o) != 0 {
					v[id], j[id] = x != y, j[a] && j[b]
				} else {
					v[id] = x && y
					if v[id] {
						j[id] = j[a] && j[b]
					} else if !x && !y {
						j[id] = j[a] || j[b]
					} else if !x {
						j[id] = j[a]
					} else {
						j[id] = j[b]
					}
				}
			} else if gia.Gia_ObjIsCo(o) != 0 {
				a := gia.Gia_ObjFaninId0(o, id)
				v[id], j[id] = v[a] != (gia.Gia_ObjFaninC0(o) != 0), j[a]
			}
		}
		for id := int32(0); id < n; id++ {
			if v[id] {
				vec.Vec_BitWriteEntry(values, int(frame*n+id), 1)
			}
			if j[id] {
				vec.Vec_BitWriteEntry(just, int(frame*n+id), 1)
			}
		}
		for i := int32(0); i < cex.NRegs; i++ {
			id := gia.Gia_ObjId(p, gia.Gia_ManRi(p, i))
			state[i], depend[i] = v[id], j[id]
		}
	}
	po := gia.Gia_ObjId(p, gia.Gia_ManPo(p, cex.IPo))
	vec.Vec_BitWriteEntry(just, int(cex.IFrame*n+po), 1)
	Bmc_GiaGenerateJustNonRec(p, cex.IFrame, values, just)
	*pvValues, *pvJust = values, just
}
