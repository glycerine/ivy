package bmc

import (
	"fmt"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
)

// Bmc_CexCareMinimizeAig computes a sufficient input care mask for cex.
// It follows ABC's priority-based justification, trying up to four orders and
// selecting the smallest mask. Initial register values are assumed zero.
func Bmc_CexCareMinimizeAig(p *gia.Gia_Man_t, realPis int32, cex *util.Abc_Cex_t, tries, check, verbose int32) *util.Abc_Cex_t {
	if p == nil || cex == nil || cex.NPis != gia.Gia_ManPiNum(p) || cex.NRegs != gia.Gia_ManRegNum(p) || cex.IFrame < 0 || cex.IPo < 0 || cex.IPo >= gia.Gia_ManPoNum(p) || realPis < 0 || realPis > cex.NPis || tries < 1 || tries > 4 || int64(cex.NRegs)+int64(cex.NPis)*int64(cex.IFrame+1) > int64(len(cex.PData))*32 {
		return nil
	}
	// ABC's justification runs on AND-inverter graphs. Gia_ManDupAbs can
	// retain native XORs and MUXes, so lower those while preserving CI and CO
	// order, including the register interfaces.
	for id := int32(1); id < gia.Gia_ManObjNum(p); id++ {
		_, _, _, kind := p.Gate(id)
		if kind == 3 || kind == 4 {
			p = cexCareLowerSpecialGates(p)
			break
		}
	}
	pis, regs, frames := cex.NPis, cex.NRegs, cex.IFrame+1
	n := gia.Gia_ManObjNum(p)
	ci := p.InputLiterals()
	co := make([]int32, gia.Gia_ManCoNum(p))
	for i := range co {
		co[i] = gia.Gia_ObjId(p, gia.Gia_ManCo(p, int32(i)))
	}
	isPI := make([]bool, n)
	for i := int32(0); i < pis; i++ {
		isPI[ci[i]>>1] = true
	}
	var best *util.Abc_Cex_t
	for attempt := int32(0); attempt < tries; attempt++ {
		priorities := make([]int32, pis*frames)
		counter := int32(0)
		assign := func(start, end, step int32) {
			for f := frames - 1; f >= 0; f-- {
				for i := start; i != end; i += step {
					v := counter << 1
					if util.Abc_InfoHasBit(cex.PData, regs+f*pis+i) {
						v++
					}
					priorities[f*pis+i] = v
					counter++
				}
			}
		}
		if attempt < 2 {
			assign(pis-1, realPis-1, -1)
		} else {
			assign(realPis, pis, 1)
		}
		if attempt%2 == 0 {
			assign(realPis-1, -1, -1)
		} else {
			assign(0, realPis, 1)
		}
		values := make([]int32, n)
		values[0] = counter << 1
		for i := int32(0); i < regs; i++ {
			values[co[int32(len(co))-regs+i]] = values[0]
		}
		ff := make([]int32, frames*regs)
		forward := func(f int32) bool {
			for i := int32(0); i < pis; i++ {
				values[ci[i]>>1] = priorities[f*pis+i]
			}
			for id := int32(1); id < n; id++ {
				a, b, _, kind := p.Gate(id)
				switch kind {
				case 2:
					x, y := values[a>>1], values[b>>1]
					v0, v1 := (x^a)&1, (y^b)&1
					prio := x >> 1
					if v0 != 0 && v1 != 0 {
						if y>>1 < prio {
							prio = y >> 1
						}
					} else if v0 != 0 {
						prio = y >> 1
					} else if v1 == 0 && y>>1 > prio {
						prio = y >> 1
					}
					values[id] = prio<<1 | (v0 & v1)
				case 5, 6:
					values[id] = values[a>>1] ^ (a & 1)
				}
			}
			return true
		}
		for f := int32(0); f < frames; f++ {
			for i := int32(0); i < regs; i++ {
				v := values[co[int32(len(co))-regs+i]]
				values[ci[pis+i]>>1] = v
				ff[f*regs+i] = v
			}
			if !forward(f) {
				return nil
			}
		}
		if values[co[cex.IPo]]&1 == 0 {
			return nil
		}
		care := util.Abc_CexAlloc(regs, pis, frames)
		care.IPo, care.IFrame = cex.IPo, cex.IFrame
		marked := make([]bool, n)
		for f := frames - 1; f >= 0; f-- {
			marked[co[cex.IPo]] = f == frames-1
			for i := int32(0); i < regs; i++ {
				values[ci[pis+i]>>1] = ff[f*regs+i]
			}
			forward(f)
			for id := int32(1); id < n; id++ {
				_, _, _, kind := p.Gate(id)
				if kind == 2 || kind == 5 {
					marked[id] = false
				}
			}
			for _, lit := range ci {
				marked[lit>>1] = false
			}
			for _, id := range co {
				if marked[id] {
					a, _, _, _ := p.Gate(id)
					marked[a>>1] = true
				}
			}
			for id := n - 1; id > 0; id-- {
				a, b, _, kind := p.Gate(id)
				if !marked[id] {
					continue
				}
				if kind == 5 {
					marked[a>>1] = true
					continue
				}
				if kind != 2 {
					continue
				}
				x, y := a>>1, b>>1
				v0, v1 := (values[x]^a)&1, (values[y]^b)&1
				if v0 != 0 && v1 != 0 {
					marked[x], marked[y] = true, true
				} else if v0 != 0 {
					marked[y] = true
				} else if v1 != 0 {
					marked[x] = true
				} else if !marked[x] && !marked[y] {
					if isPI[x] || (!isPI[y] && values[x]>>1 > values[y]>>1) {
						marked[x] = true
					} else {
						marked[y] = true
					}
				}
			}
			for i := int32(0); i < pis; i++ {
				if marked[ci[i]>>1] {
					util.Abc_InfoSetBit(care.PData, regs+f*pis+i)
				}
			}
			for i := int32(0); i < regs; i++ {
				marked[co[int32(len(co))-regs+i]] = marked[ci[pis+i]>>1]
			}
		}
		if verbose != 0 {
			Bmc_CexPrint(care, realPis, 0)
		}
		if best == nil || util.Abc_CexCountOnes(care) < util.Abc_CexCountOnes(best) {
			best = care
		}
	}
	if Bmc_CexVerify(p, cex, best) == 0 {
		fmt.Println("Counter-example verification has failed.")
	} else if check != 0 {
		fmt.Println("Counter-example verification succeeded.")
	}
	return best
}

// cexCareLowerSpecialGates leaves the source manager untouched and retains
// the original input, output, and register ordering.
func cexCareLowerSpecialGates(p *gia.Gia_Man_t) *gia.Gia_Man_t {
	out := gia.Gia_ManStart(gia.Gia_ManObjNum(p))
	copies := make([]int32, gia.Gia_ManObjNum(p))
	for _, lit := range p.InputLiterals() {
		copies[lit>>1] = gia.Gia_ManAppendCi(out)
	}
	copyLit := func(lit int32) int32 { return copies[lit>>1] ^ (lit & 1) }
	for id := int32(1); id < gia.Gia_ManObjNum(p); id++ {
		a, b, c, kind := p.Gate(id)
		switch kind {
		case 2:
			copies[id] = gia.Gia_ManAppendAnd2(out, copyLit(a), copyLit(b))
		case 3:
			copies[id] = gia.Gia_ManAppendXor2(out, copyLit(a), copyLit(b))
		case 4:
			copies[id] = gia.Gia_ManAppendMux2(out, copyLit(a), copyLit(b), copyLit(c))
		case 5:
			copies[id] = copyLit(a)
		}
	}
	for _, lit := range p.OutputLiterals() {
		gia.Gia_ManAppendCo(out, copyLit(lit))
	}
	gia.Gia_ManSetRegNum(out, gia.Gia_ManRegNum(p))
	return out
}

// Bmc_CexCareMinimize applies the same minimization to an AIG manager.
func Bmc_CexCareMinimize(p *aig.Graph, realPis int32, cex *util.Abc_Cex_t, tries, check, verbose int32) *util.Abc_Cex_t {
	if p == nil {
		return nil
	}
	return Bmc_CexCareMinimizeAig(gia.Gia_ManFromAigSimple(p), realPis, cex, tries, check, verbose)
}
