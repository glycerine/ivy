package bmc

import (
	"fmt"
	"sync"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

func Abc_InfoGet2Bits(v *vec.Vec_Int_t, words, frame, obj int32) int32 {
	return int32(packed2Get(uint32(vec.Vec_IntEntry(v, int(words*frame+(obj>>4)))), uint32(obj)))
}
func Abc_InfoSet2Bits(v *vec.Vec_Int_t, words, frame, obj, value int32) {
	i := int(words*frame + (obj >> 4))
	vec.Vec_IntWriteEntry(v, i, int32(packed2Set(uint32(vec.Vec_IntEntry(v, i)), uint32(obj), uint32(value))))
}
func Abc_InfoAdd2Bits(v *vec.Vec_Int_t, words, frame, obj, value int32) {
	i := int(words*frame + (obj >> 4))
	vec.Vec_IntWriteEntry(v, i, int32(packed2Add(uint32(vec.Vec_IntEntry(v, i)), uint32(obj), uint32(value))))
}

// The original temporarily stores this annotation in Gia_Man_t.vTruths.
// Keep it here because the Go GIA manager does not expose that scratch field.
type unrollingAnnotation struct {
	n      int32
	frames [][]byte
}

var unrollingAnnotations sync.Map

func annotation(p *gia.Gia_Man_t) *unrollingAnnotation {
	v, ok := unrollingAnnotations.Load(p)
	if !ok {
		panic("bmc: no counterexample annotation")
	}
	return v.(*unrollingAnnotation)
}
func Gia_ManGetTwo(p *gia.Gia_Man_t, f int32, o *gia.Gia_Obj_t) int32 {
	return int32(annotation(p).frames[f][gia.Gia_ObjId(p, o)])
}
func Gia_ManSetTwo(p *gia.Gia_Man_t, f int32, o *gia.Gia_Obj_t, v int32) {
	annotation(p).frames[f][gia.Gia_ObjId(p, o)] = byte(v & 3)
}
func Gia_ManAddTwo(p *gia.Gia_Man_t, f int32, o *gia.Gia_Obj_t, v int32) {
	annotation(p).frames[f][gia.Gia_ObjId(p, o)] |= byte(v & 3)
}

// Gia_ManAnnotateUnrolling records each value and traces the selected output's
// sufficient justification backwards across the register boundaries.
func Gia_ManAnnotateUnrolling(p *gia.Gia_Man_t, c *util.Abc_Cex_t, justifyAll int32) int32 {
	if p == nil || c == nil || c.IFrame < 0 || c.NPis != gia.Gia_ManPiNum(p) || c.IPo < 0 || c.IPo >= gia.Gia_ManPoNum(p) {
		panic("bmc: invalid counterexample")
	}
	n := gia.Gia_ManObjNum(p)
	a := &unrollingAnnotation{n: n, frames: make([][]byte, c.IFrame+1)}
	if _, exists := unrollingAnnotations.LoadOrStore(p, a); exists {
		panic("bmc: counterexample already annotated")
	}
	pis, regs := gia.Gia_ManPiNum(p), gia.Gia_ManRegNum(p)
	state := make([]bool, regs)
	for f := int32(0); f <= c.IFrame; f++ {
		values := make([]bool, n)
		for i := int32(0); i < pis; i++ {
			values[gia.Gia_ObjId(p, gia.Gia_ManCi(p, i))] = util.Abc_InfoHasBit(c.PData, c.NRegs+f*pis+i)
		}
		for i := int32(0); i < regs; i++ {
			values[gia.Gia_ObjId(p, gia.Gia_ManRo(p, i))] = state[i]
		}
		read := func(lit int32) bool { return values[lit>>1] != (lit&1 != 0) }
		for id := int32(1); id < n; id++ {
			x, y, z, k := p.Gate(id)
			switch k {
			case 2:
				values[id] = read(x) && read(y)
			case 3:
				values[id] = read(x) != read(y)
			case 4:
				if read(x) {
					values[id] = read(y)
				} else {
					values[id] = read(z)
				}
			case 5, 6:
				values[id] = read(x)
			}
		}
		a.frames[f] = make([]byte, n)
		for i, v := range values {
			if v {
				a.frames[f][i] = 1
			}
		}
		for i := int32(0); i < regs; i++ {
			state[i] = values[gia.Gia_ObjId(p, gia.Gia_ManRi(p, i))]
		}
	}
	po := gia.Gia_ObjId(p, gia.Gia_ManPo(p, c.IPo))
	ret := int32(a.frames[c.IFrame][po] & 1)
	if ret == 0 {
		fmt.Println("Counter-example is invalid.")
	}
	// A per-frame mark set accumulates reconverging justification paths.
	marks := make([][]bool, c.IFrame+1)
	for i := range marks {
		marks[i] = make([]bool, n)
	}
	marks[c.IFrame][po] = true
	for f := c.IFrame; f >= 0; f-- {
		mark := func(id int32) { marks[f][id] = true }
		read := func(lit int32) bool { return (a.frames[f][lit>>1]&1 != 0) != (lit&1 != 0) }
		for id := n - 1; id >= 0; id-- {
			if !marks[f][id] {
				continue
			}
			a.frames[f][id] |= 2
			x, y, z, k := p.Gate(id)
			switch k {
			case 2:
				vx, vy := read(x), read(y)
				if vx && vy || justifyAll != 0 && !vx && !vy {
					mark(x >> 1)
					mark(y >> 1)
				} else if !vx {
					mark(x >> 1)
				} else {
					mark(y >> 1)
				}
			case 3:
				mark(x >> 1)
				mark(y >> 1)
			case 4:
				mark(x >> 1)
				if read(x) {
					mark(y >> 1)
				} else {
					mark(z >> 1)
				}
			case 5, 6:
				mark(x >> 1)
			case 1:
				if f > 0 {
					for i := int32(0); i < regs; i++ {
						if gia.Gia_ObjId(p, gia.Gia_ManRo(p, i)) == id {
							marks[f-1][gia.Gia_ObjId(p, gia.Gia_ManRi(p, i))] = true
							break
						}
					}
				}
			}
		}
	}
	return ret
}

// Gia_ManCreateUnate constructs the source's justification network for a
// suffix of the annotated trace. Inputs represent justifying boundary objects.
func Gia_ManCreateUnate(p *gia.Gia_Man_t, c *util.Abc_Cex_t, frame, nRealPis, useAll int32) *gia.Gia_Man_t {
	a := annotation(p)
	if c == nil || frame < 0 || frame > c.IFrame || nRealPis >= gia.Gia_ManPiNum(p) {
		panic("bmc: invalid unate frame")
	}
	q := gia.Gia_ManStart(1000)
	gia.Gia_ManHashStart(q)
	pis, regs := gia.Gia_ManPiNum(p), gia.Gia_ManRegNum(p)
	values := make([]int32, a.n)
	for i := int32(0); i < regs; i++ {
		id := gia.Gia_ObjId(p, gia.Gia_ManRo(p, i))
		selected := a.frames[frame][id]&2 != 0
		if nRealPis < 0 {
			if useAll != 0 {
				lit := gia.Gia_ManAppendCi(q)
				if selected {
					values[id] = lit
				}
			} else if selected {
				values[id] = gia.Gia_ManAppendCi(q)
			}
		} else if selected {
			values[id] = 1
		}
	}
	for f := frame; f <= c.IFrame; f++ {
		values[0] = int32(a.frames[f][0] >> 1)
		for i := int32(0); i < pis; i++ {
			id := gia.Gia_ObjId(p, gia.Gia_ManCi(p, i))
			selected := a.frames[f][id]&2 != 0
			values[id] = int32(a.frames[f][id] >> 1)
			if nRealPis >= 0 && i >= nRealPis {
				if useAll != 0 {
					lit := gia.Gia_ManAppendCi(q)
					if selected {
						values[id] = lit
					}
				} else if selected {
					values[id] = gia.Gia_ManAppendCi(q)
				}
			}
		}
		for id := int32(1); id < a.n; id++ {
			x, y, _, kind := p.Gate(id)
			switch kind {
			case 2:
				values[id] = 0
				if a.frames[f][id]&2 == 0 {
					continue
				}
				u, v := values[x>>1], values[y>>1]
				if u != 0 && v != 0 {
					if a.frames[f][id]&1 != 0 {
						values[id] = gia.Gia_ManHashAnd(q, u, v)
					} else {
						values[id] = gia.Gia_ManHashOr(q, u, v)
					}
				} else if u != 0 {
					values[id] = u
				} else if v != 0 {
					values[id] = v
				} else {
					panic("bmc: justified gate has no justified fanin")
				}
			case 3, 4:
				panic("bmc: unate construction requires an AND-inverter graph")
			case 5, 6:
				values[id] = values[x>>1]
			}
		}
		if f < c.IFrame {
			for i := int32(0); i < regs; i++ {
				values[gia.Gia_ObjId(p, gia.Gia_ManRo(p, i))] = values[gia.Gia_ObjId(p, gia.Gia_ManRi(p, i))]
			}
		}
	}
	output := gia.Gia_ObjId(p, gia.Gia_ManPo(p, c.IPo))
	if a.frames[c.IFrame][output]&2 == 0 || values[output] == 0 {
		panic("bmc: missing output justification")
	}
	gia.Gia_ManAppendCo(q, values[output])
	gia.Gia_ManHashStop(q)
	return gia.Gia_ManCleanup(q)
}

// Gia_ManCexMin preserves the source diagnostic entry point; the original
// builds and reports unate networks and returns NULL rather than a new CEX.
func Gia_ManCexMin(p *gia.Gia_Man_t, c *util.Abc_Cex_t, start, nRealPis, justifyAll, useAll, verbose int32) *util.Abc_Cex_t {
	if p == nil || c == nil || start < 0 || start > c.IFrame || nRealPis >= gia.Gia_ManPiNum(p) {
		panic("bmc: invalid CEX minimization")
	}
	Gia_ManAnnotateUnrolling(p, c, justifyAll)
	defer unrollingAnnotations.Delete(p)
	printOne := func(f int32) {
		q := Gia_ManCreateUnate(p, c, f, nRealPis, useAll)
		fmt.Printf("%3d : ", f)
		gia.Gia_ManPrintStats(q, nil)
		if verbose != 0 {
			if err := gia.Gia_AigerWrite(q, "temp.aig", 0, 0, 0); err != nil {
				fmt.Println(err)
			}
		}
	}
	if nRealPis >= 0 {
		printOne(start)
	} else {
		for f := c.IFrame; f >= start; f-- {
			printOne(f)
		}
	}
	return nil
}
