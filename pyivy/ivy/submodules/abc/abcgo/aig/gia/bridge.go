package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/util"
)

// Gia_ManToBridgeNetlist renumbers CIs before ANDs for the standard bridge
// payload while preserving the manager's PI/RO and PO/RI order.
func Gia_ManToBridgeNetlist(p *Gia_Man_t) (*util.BridgeNetlist, error) {
	if p == nil || p.nRegs < 0 || p.nRegs > Gia_ManCiNum(p) || p.nRegs > Gia_ManCoNum(p) {
		return nil, fmt.Errorf("gia: invalid manager register counts")
	}
	result := &util.BridgeNetlist{Inputs: Gia_ManPiNum(p), Registers: p.nRegs}
	copies := make([]int32, p.nObjs)
	for i := int32(0); i < Gia_ManCiNum(p); i++ {
		copies[Gia_ObjId(p, Gia_ManCi(p, i))] = (i + 1) << 1
	}
	remap := func(lit int32) int32 { return copies[lit>>1] ^ (lit & 1) }
	for id := int32(1); id < p.nObjs; id++ {
		obj := p.pObjs[id]
		if Gia_ObjIsAnd(obj) == 0 {
			continue
		}
		if Gia_ObjIsXor(obj) != 0 || (int(id) < len(p.PMuxes) && p.PMuxes[id] != 0) {
			return nil, fmt.Errorf("gia: normalize specialized XOR/MUX node %d before bridge export", id)
		}
		a, b := Gia_ObjFaninLit0(obj, id), Gia_ObjFaninLit1(obj, id)
		if a < 0 || b < 0 || a>>1 >= id || b>>1 >= id {
			return nil, fmt.Errorf("gia: invalid fanin at object %d", id)
		}
		result.Ands = append(result.Ands, [2]int32{remap(a), remap(b)})
		copies[id] = (Gia_ManCiNum(p) + int32(len(result.Ands))) << 1
	}
	for i := int32(0); i < Gia_ManCoNum(p); i++ {
		obj := Gia_ManCo(p, i)
		lit := remap(Gia_ObjFaninLit0(obj, obj.id))
		if i < Gia_ManPoNum(p) {
			result.Outputs = append(result.Outputs, lit)
		} else {
			result.RegisterInputs = append(result.RegisterInputs, lit)
		}
	}
	return result, nil
}

func Gia_ManFromBridgeNetlist(net *util.BridgeNetlist) (*Gia_Man_t, error) {
	if net == nil || net.Inputs < 0 || net.Registers < 0 || int(net.Registers) != len(net.RegisterInputs) {
		return nil, fmt.Errorf("gia: invalid bridge dimensions")
	}
	count := int64(net.Inputs) + int64(net.Registers) + int64(len(net.Ands)) + int64(len(net.Outputs)) + int64(len(net.RegisterInputs)) + 1
	if count >= int64(objectNoFanin) {
		return nil, fmt.Errorf("gia: bridge exceeds GIA object limit")
	}
	p := Gia_ManStart(int32(count))
	p.fGiaSimple = 1
	ci := net.Inputs + net.Registers
	for i := int32(0); i < ci; i++ {
		Gia_ManAppendCi(p)
	}
	for i, and := range net.Ands {
		limit := 2 * (ci + int32(i) + 1)
		if and[0] < 0 || and[1] < 0 || and[0] >= limit || and[1] >= limit {
			return nil, fmt.Errorf("gia: invalid bridge AND %d", i)
		}
		Gia_ManAppendAnd(p, and[0], and[1])
	}
	limit := 2 * p.nObjs
	for _, list := range [][]int32{net.Outputs, net.RegisterInputs} {
		for _, lit := range list {
			if lit < 0 || lit >= limit {
				return nil, fmt.Errorf("gia: invalid bridge output literal")
			}
			Gia_ManAppendCo(p, lit)
		}
	}
	p.nRegs = net.Registers
	p.fGiaSimple = 0
	return p, nil
}
