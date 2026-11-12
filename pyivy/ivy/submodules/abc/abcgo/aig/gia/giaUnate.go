package gia

import (
	"fmt"
	"time"
)

// Vec_IntFlopBit reverses the polarity of each unpaired literal in a sorted
// support vector; variables represented in both polarities are unchanged.
func Vec_IntFlopBit(p *Vec_Int_t) *Vec_Int_t {
	values := Vec_IntArray(p)
	for i := 0; i < len(values); i++ {
		if i+1 < len(values) && values[i]>>1 == values[i+1]>>1 {
			i++
		} else {
			values[i] ^= 1
		}
	}
	return p
}
func Gia_ManCheckUnateVec(p *Gia_Man_t, ciIDs, coIDs *Vec_Int_t) *Vec_Wec_t {
	inputs, outputs := ciIDs, coIDs
	if inputs == nil {
		inputs = Vec_IntStartNatural(int(Gia_ManCiNum(p)))
	}
	if outputs == nil {
		outputs = Vec_IntStartNatural(int(Gia_ManCoNum(p)))
	}
	result := Vec_WecStart(Vec_IntSize(outputs))
	unates := Vec_WecStart(int(Gia_ManObjNum(p)))
	for _, id := range Vec_IntArray(inputs) {
		obj := Gia_ManCi(p, id)
		Vec_IntPush(Vec_WecEntry(unates, int(obj.id)), id<<1)
	}
	for _, obj := range p.pObjs {
		if Gia_ObjIsAnd(obj) == 0 {
			continue
		}
		a, b := Vec_WecEntry(unates, int(Gia_ObjFaninId0(obj, obj.id))), Vec_WecEntry(unates, int(Gia_ObjFaninId1(obj, obj.id)))
		if Gia_ObjFaninC0(obj) != 0 {
			Vec_IntFlopBit(a)
		}
		if Gia_ObjFaninC1(obj) != 0 {
			Vec_IntFlopBit(b)
		}
		Vec_IntTwoMerge2(a, b, Vec_WecEntry(unates, int(obj.id)))
		if Gia_ObjFaninC0(obj) != 0 {
			Vec_IntFlopBit(a)
		}
		if Gia_ObjFaninC1(obj) != 0 {
			Vec_IntFlopBit(b)
		}
	}
	for i, id := range Vec_IntArray(outputs) {
		obj := Gia_ManCo(p, id)
		fan := Vec_WecEntry(unates, int(Gia_ObjFaninId0p(p, obj)))
		if Gia_ObjFaninC0(obj) != 0 {
			Vec_IntFlopBit(fan)
		}
		Vec_IntAppend(Vec_WecEntry(result, i), fan)
		if Gia_ObjFaninC0(obj) != 0 {
			Vec_IntFlopBit(fan)
		}
	}
	return result
}
func Gia_ManCheckUnate_rec(p *Gia_Man_t, id int32) int32 {
	if difference := p.nTravIds - p.pTravIds[id]; difference <= 3 {
		return difference
	}
	obj := Gia_ManObj(p, id)
	p.pTravIds[id] = p.nTravIds - 3
	if id == 0 || Gia_ObjIsCi(obj) != 0 {
		return 3
	}
	a, b := Gia_ManCheckUnate_rec(p, Gia_ObjFaninId0(obj, id)), Gia_ManCheckUnate_rec(p, Gia_ObjFaninId1(obj, id))
	if (a == 1 || a == 2) && Gia_ObjFaninC0(obj) != 0 {
		a ^= 3
	}
	if (b == 1 || b == 2) && Gia_ObjFaninC1(obj) != 0 {
		b ^= 3
	}
	p.pTravIds[id] = p.nTravIds - (a & b)
	return a & b
}

// Gia_ManCheckUnate reports non-unate=0, negative=1, positive=2, absent=3.
func Gia_ManCheckUnate(p *Gia_Man_t, ci, co int32) int32 {
	input, output := Gia_ManCi(p, ci), Gia_ManCo(p, co)
	Gia_ManIncrementTravId(p)
	Gia_ManIncrementTravId(p)
	Gia_ObjSetTravIdCurrentId(p, input.id)
	Gia_ManIncrementTravId(p)
	Gia_ManIncrementTravId(p)
	value := Gia_ManCheckUnate_rec(p, Gia_ObjFaninId0p(p, output))
	if (value == 1 || value == 2) && Gia_ObjFaninC0(output) != 0 {
		value ^= 3
	}
	return value
}
func Gia_ManCheckUnateVecTest(p *Gia_Man_t, verbose int32) { Gia_ManCheckUnateTest(p, 1, verbose) }
func Gia_ManCheckUnateTest(p *Gia_Man_t, all, verbose int32) {
	start := time.Now()
	var vectors *Vec_Wec_t
	if all != 0 {
		vectors = Gia_ManCheckUnateVec(p, nil, nil)
	}
	if verbose != 0 {
		fmt.Print("Inputs  : ")
		for i := int32(0); i < Gia_ManCiNum(p); i++ {
			fmt.Print(i % 10)
		}
		fmt.Println()
	}
	var unate, nonUnate int
	for o := int32(0); o < Gia_ManCoNum(p); o++ {
		row := make([]byte, Gia_ManCiNum(p))
		for i := range row {
			row[i] = ' '
		}
		if all != 0 {
			lits := Vec_IntArray(Vec_WecEntry(vectors, int(o)))
			for i := 0; i < len(lits); i++ {
				lit := lits[i]
				if i+1 < len(lits) && lit>>1 == lits[i+1]>>1 {
					row[lit>>1] = '.'
					i++
					nonUnate++
				} else {
					row[lit>>1] = 'p'
					if lit&1 != 0 {
						row[lit>>1] = 'n'
					}
					unate++
				}
			}
		} else {
			for i := int32(0); i < Gia_ManCiNum(p); i++ {
				switch Gia_ManCheckUnate(p, i, o) {
				case 0:
					row[i] = '.'
					nonUnate++
				case 1:
					row[i] = 'n'
					unate++
				case 2:
					row[i] = 'p'
					unate++
				}
			}
		}
		if verbose != 0 {
			fmt.Printf("Out%4d : %s\n", o, row)
		}
	}
	fmt.Printf("Ins/Outs = %4d/%4d.  Total supp = %5d.  Total unate = %5d.\nTotal time = %.2f sec\n", Gia_ManCiNum(p), Gia_ManCoNum(p), unate+nonUnate, unate, time.Since(start).Seconds())
}
