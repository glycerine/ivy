package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/sat/satoko"
)

// Gia_ManFramesForCexMin unrolls a zero-initialized transition system.
func Gia_ManFramesForCexMin(p *Gia_Man_t, frames int32) *Gia_Man_t {
	if Gia_ManPoNum(p) != 1 || frames <= 0 {
		panic("gia: invalid counterexample unrolling")
	}
	out := Gia_ManStart(p.nObjs)
	out.pName, out.PSpec = p.pName, p.PSpec
	Gia_ManHashAlloc(out)
	mapped := make([]int32, p.nObjs)
	state := make([]int32, Gia_ManRegNum(p))
	for frame := int32(0); frame < frames; frame++ {
		for i := int32(0); i < Gia_ManPiNum(p); i++ {
			mapped[Gia_ObjId(p, Gia_ManCi(p, i))] = Gia_ManAppendCi(out)
		}
		for i := range state {
			mapped[Gia_ObjId(p, Gia_ManCi(p, Gia_ManPiNum(p)+int32(i)))] = state[i]
		}
		lit := func(v int32) int32 { return mapped[v>>1] ^ (v & 1) }
		for id := int32(1); id < p.nObjs; id++ {
			a, b, c, kind := p.Gate(id)
			switch kind {
			case 1:
			case 2:
				mapped[id] = Gia_ManHashAnd(out, lit(a), lit(b))
			case 3:
				mapped[id] = Gia_ManHashXor(out, lit(a), lit(b))
			case 4:
				mapped[id] = Gia_ManHashMux(out, lit(a), lit(b), lit(c))
			case 5, 6:
				mapped[id] = lit(a)
			default:
				panic("gia: invalid gate during counterexample unrolling")
			}
		}
		for i := range state {
			state[i] = mapped[Gia_ObjId(p, Gia_ManCo(p, Gia_ManPoNum(p)+int32(i)))]
		}
	}
	for i := int32(0); i < Gia_ManCoNum(p); i++ {
		Gia_ManAppendCo(out, mapped[Gia_ObjId(p, Gia_ManCo(p, i))])
	}
	clean := Gia_ManCleanup(out)
	Gia_ManStop(out)
	return clean
}

// Bmc_CexCareDeriveCex builds a mask from the selected input SAT literals.
func Bmc_CexCareDeriveCex(cex *util.Abc_Cex_t, first int32, lits []int32, count int32) *util.Abc_Cex_t {
	if count < 0 || int(count) > len(lits) {
		panic("gia: invalid care literals")
	}
	out := util.Abc_CexAlloc(cex.NRegs, cex.NPis, cex.IFrame+1)
	out.IPo, out.IFrame = cex.IPo, cex.IFrame
	for _, lit := range lits[:count] {
		index := (lit >> 1) - first
		if index < 0 || index >= cex.NBits-cex.NRegs {
			panic("gia: care literal out of range")
		}
		index += out.NRegs
		out.PData[index>>5] |= 1 << uint(index&31)
	}
	return out
}

// Bmc_CexCareSatBasedMinimizeAig finds input bits sufficient to force the
// failed output. The returned counterexample holds a care mask, not values.
func Bmc_CexCareSatBasedMinimizeAig(p *Gia_Man_t, cex *util.Abc_Cex_t, highEffort, verbose int32) *util.Abc_Cex_t {
	if Gia_ManVerifyCex(p, cex, 0) == 0 || cex.IPo != 0 || Gia_ManPoNum(p) != 1 {
		panic("gia: expected a valid single-output counterexample")
	}
	for i := int32(0); i < cex.NRegs; i++ {
		if cex.PData[i>>5]&(1<<uint(i&31)) != 0 {
			panic("gia: expected zero initial state")
		}
	}
	graph := Gia_ManFramesForCexMin(p, cex.IFrame+1)
	defer Gia_ManStop(graph)
	solver := satoko.New()
	solver.SetVariables(int(graph.nObjs))
	solver.AddClause([]int{1}) // constant-zero variable
	for id := int32(1); id < graph.nObjs; id++ {
		a, b, c, kind := graph.Gate(id)
		if kind == 1 {
			continue
		}
		args := []int32{a}
		if kind == 2 || kind == 3 {
			args = append(args, b)
		} else if kind == 4 {
			args = append(args, b, c)
		}
		for pattern := 0; pattern < (1 << len(args)); pattern++ {
			x := pattern&1 != 0
			y, z := false, false
			if len(args) > 1 {
				y = pattern&2 != 0
			}
			if len(args) > 2 {
				z = pattern&4 != 0
			}
			value := x
			switch kind {
			case 2:
				value = x && y
			case 3:
				value = x != y
			case 4:
				if x {
					value = y
				} else {
					value = z
				}
			case 5, 6:
			default:
				panic("gia: invalid gate in care SAT")
			}
			clause := make([]int, 0, len(args)+1)
			for i, arg := range args {
				lit := int(arg)
				if pattern&(1<<i) != 0 {
					lit ^= 1
				}
				clause = append(clause, lit)
			}
			out := int(id << 1)
			if !value {
				out ^= 1
			}
			clause = append(clause, out)
			solver.AddClause(clause)
		}
	}
	// Ask whether the failed output could be false with the selected bits fixed.
	output := Gia_ObjId(graph, Gia_ManCo(graph, 0))
	solver.AddClause([]int{int(output<<1) | 1})
	inputs := graph.InputLiterals()
	assumptions := make([]int, len(inputs))
	for i, lit := range inputs {
		assumptions[i] = int(lit)
		bit := int64(cex.NRegs) + int64(i)
		if cex.PData[bit>>5]&(1<<uint(bit&31)) == 0 {
			assumptions[i] ^= 1
		}
	}
	best, status := solver.MinimizeAssumptions(assumptions, 0)
	if status != satoko.SATOKO_UNSAT {
		panic("gia: counterexample does not force output")
	}
	if highEffort != 0 {
		reversed := append([]int(nil), assumptions...)
		for i, j := 0, len(reversed)-1; i < j; i, j = i+1, j-1 {
			reversed[i], reversed[j] = reversed[j], reversed[i]
		}
		other, status := solver.MinimizeAssumptions(reversed, 0)
		if status == satoko.SATOKO_UNSAT && len(other) < len(best) {
			best = other
		}
	}
	selected := make(map[int]bool, len(best))
	for _, lit := range best {
		selected[lit>>1] = true
	}
	result := util.Abc_CexAlloc(cex.NRegs, cex.NPis, cex.IFrame+1)
	result.IPo, result.IFrame = cex.IPo, cex.IFrame
	for i, lit := range inputs {
		if selected[int(lit>>1)] {
			bit := int64(cex.NRegs) + int64(i)
			result.PData[bit>>5] |= 1 << uint(bit&31)
		}
	}
	if verbose != 0 {
		fmt.Printf("SAT care bits = %d of %d.\n", len(best), len(inputs))
	}
	return result
}

func Gia_ManMinCex(p *Gia_Man_t, cex *util.Abc_Cex_t) {
	Bmc_CexCareSatBasedMinimizeAig(p, cex, 1, 1)
}
