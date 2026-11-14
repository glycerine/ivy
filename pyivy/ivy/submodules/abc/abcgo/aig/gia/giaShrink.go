package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
	"time"
)

// Gia_ManMapShrink4 resynthesizes each mapped LUT using the four-input DAR library.
func Gia_ManMapShrink4(p *Gia_Man_t, keepLevel, verbose int32) *Gia_Man_t {
	started := time.Now()
	if Gia_ManHasMapping(p) == 0 {
		panic("gia: LUT shrinking requires mapping")
	}
	if Gia_ManLutSizeMax(p) > 4 {
		fmt.Println("Resynthesis is not performed when nodes have more than 4 inputs.")
		return nil
	}
	Gia_ManFillValue(p)
	Gia_ManConst0(p).Value = 0
	q := Gia_ManStart(p.nObjs)
	q.pName, q.PSpec = p.pName, p.PSpec
	Gia_ManCleanLevels(q, p.nObjs)
	Gia_ManHashStart(q)
	for id := int32(1); id < p.nObjs; id++ {
		obj := Gia_ManObj(p, id)
		switch {
		case Gia_ObjIsCi(obj) != 0:
			lit := Gia_ManAppendCi(q)
			obj.Value = uint32(lit)
			if p.vLevels != nil {
				Gia_ObjSetLevelId(q, lit>>1, Gia_ObjLevelId(p, id))
			}
		case Gia_ObjIsCo(obj) != 0:
			obj.Value = uint32(Gia_ManAppendCo(q, Gia_ObjFanin0Copy(obj)))
		case Gia_ObjIsLut(p, id) != 0:
			fans := Gia_ObjLutFanins(p, id)
			truth := uint32(giaNewTruthEval(p, int32(len(fans)), fans).node(id)[0]) & 0xffff
			leaves := vec.Vec_IntAlloc(4)
			for _, fan := range fans {
				value := Gia_ManObj(p, fan).Value
				if value == ^uint32(0) {
					vec.Vec_IntPush(leaves, 0)
				} else {
					vec.Vec_IntPush(leaves, int32(value))
				}
			}
			for vec.Vec_IntSize(leaves) < 4 {
				vec.Vec_IntPush(leaves, 0)
			}
			obj.Value = uint32(Dar_LibEvalBuild(q, leaves, truth, keepLevel, nil))
		}
	}
	Gia_ManHashStop(q)
	if Gia_ManHasDangling(q) != 0 {
		q = Gia_ManCleanup(q)
	}
	Gia_ManSetRegNum(q, Gia_ManRegNum(p))
	if verbose != 0 {
		fmt.Printf("Total gain in AIG nodes = %d. Total runtime = %.2f sec\n", Gia_ManObjNum(p)-Gia_ManObjNum(q), time.Since(started).Seconds())
	}
	return q
}
