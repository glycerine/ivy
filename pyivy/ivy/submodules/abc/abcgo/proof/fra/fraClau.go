package fra

// Counterexample-guided clause induction from fraClau.c.
import (
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/cnf"
)

type Cla_Man_t = clauManager

func Fra_ClauSaveLatchVars(g *aig.Graph, c *cnf.Cnf_Dat_t, current int32) *vec.Vec_Int_t {
	return clauLatchVars(g, c, current != 0)
}
func Fra_ClauSaveOutputVars(g *aig.Graph, c *cnf.Cnf_Dat_t) *vec.Vec_Int_t {
	return clauOutputVars(g, c)
}
func Fra_ClauSaveInputVars(g *aig.Graph, c *cnf.Cnf_Dat_t, start int32) *vec.Vec_Int_t {
	return clauInputVars(g, c, int(start))
}
func Fra_ClauCreateMapping(a, b *vec.Vec_Int_t, n int32) []int32 { return clauMapping(a, b, int(n)) }
func Fra_ClauStart(g *aig.Graph) *Cla_Man_t                      { return clauStart(g) }
func Fra_ClauStop(p *Cla_Man_t) {
	if p != nil {
		*p = Cla_Man_t{}
	}
}
func copyVec(dst, src *vec.Vec_Int_t) {
	vec.Vec_IntClear(dst)
	if src != nil {
		vec.Vec_IntPushArray(dst, vec.Vec_IntArray(src), vec.Vec_IntSize(src))
	}
}
func Fra_ClauCheckProperty(p *Cla_Man_t, out *vec.Vec_Int_t) int32 {
	status, cube := p.checkProperty()
	copyVec(out, cube)
	return status
}
func Fra_ClauCheckBmc(p *Cla_Man_t, c *vec.Vec_Int_t) int32 { return p.checkBMC(c) }
func Fra_ClauRemapClause(mapping []int32, clause, out *vec.Vec_Int_t, invert int32) {
	copyVec(out, clauRemap(mapping, clause, invert != 0))
}
func Fra_ClauCheckClause(p *Cla_Man_t, c, out *vec.Vec_Int_t) int32 {
	status, cube := p.checkClause(c)
	if out != nil {
		copyVec(out, cube)
	}
	return status
}
func Fra_ClauReduceClause(a, b *vec.Vec_Int_t) { clauReduce(a, b) }
func clauComplement(v *vec.Vec_Int_t) {
	for i, x := range vec.Vec_IntArray(v) {
		vec.Vec_IntWriteEntry(v, i, x^1)
	}
}
func clauSplit(v *vec.Vec_Int_t) *vec.Vec_Int_t {
	n := vec.Vec_IntSize(v) / 2
	if n == 0 {
		panic("cannot split short clause")
	}
	out := vec.Vec_IntAllocArrayCopy(vec.Vec_IntArray(v)[n:], vec.Vec_IntSize(v)-n)
	vec.Vec_IntShrink(v, n)
	return out
}
func Fra_ClauMinimizeClause_rec(p *Cla_Man_t, basis, extra *vec.Vec_Int_t) {
	if vec.Vec_IntSize(extra) == 1 {
		return
	}
	n := vec.Vec_IntSize(basis)
	second := clauSplit(extra)
	vec.Vec_IntAppend(basis, extra)
	if Fra_ClauCheckClause(p, basis, nil) == 1 {
		vec.Vec_IntShrink(basis, n)
		Fra_ClauMinimizeClause_rec(p, basis, extra)
		return
	}
	vec.Vec_IntShrink(basis, n)
	vec.Vec_IntAppend(basis, second)
	if Fra_ClauCheckClause(p, basis, nil) == 1 {
		vec.Vec_IntShrink(basis, n)
		Fra_ClauMinimizeClause_rec(p, basis, second)
		return
	}
	Fra_ClauMinimizeClause_rec(p, basis, extra)
	vec.Vec_IntShrink(basis, n)
	vec.Vec_IntAppend(basis, extra)
	Fra_ClauMinimizeClause_rec(p, basis, second)
	vec.Vec_IntShrink(basis, n)
	vec.Vec_IntAppend(extra, second)
}
func Fra_ClauMinimizeClause(p *Cla_Man_t, basis, extra *vec.Vec_Int_t) {
	for i := vec.Vec_IntSize(extra) - 1; i >= 0; i-- {
		vec.Vec_IntClear(basis)
		for j, lit := range vec.Vec_IntArray(extra) {
			if i != j {
				vec.Vec_IntPush(basis, lit)
			}
		}
		if Fra_ClauCheckClause(p, basis, nil) != 1 {
			continue
		}
		values := vec.Vec_IntArray(extra)
		copy(values[i:], values[i+1:])
		vec.Vec_IntShrink(extra, len(values)-1)
	}
}
func Fra_ClauPrintClause(vars, clause *vec.Vec_Int_t) {
	i := 0
	for _, v := range vec.Vec_IntArray(vars) {
		if i < len(vec.Vec_IntArray(clause)) && vec.Vec_IntEntry(clause, i)>>1 == v {
			if vec.Vec_IntEntry(clause, i)&1 == 0 {
				fmt.Print("1")
			} else {
				fmt.Print("0")
			}
			i++
		} else {
			fmt.Print("-")
		}
	}
	if i != vec.Vec_IntSize(clause) {
		panic("clause variables out of order")
	}
}

// Fra_Clau returns 1 for proof, 0 when not proved, or -1 if SAT is unknown.
func Fra_Clau(graph *aig.Graph, nIters, verbose, veryVerbose int32) int32 {
	p := Fra_ClauStart(graph)
	if p == nil {
		fmt.Println("The property is trivially inductive.")
		return 1
	}
	defer Fra_ClauStop(p)
	for iter := int32(0); ; iter++ {
		status, main0 := p.checkProperty()
		// ABC evaluates the property before the bound and tests equality
		// when reporting the result, including for negative bounds.
		if iter >= nIters {
			if iter == nIters {
				return 0
			}
			return 1
		}
		if status == 1 {
			return 1
		}
		if status < 0 {
			return -1
		}
		main := clauRemap(p.mainToTest, main0, false)
		if verbose != 0 && veryVerbose != 0 {
			Fra_ClauPrintClause(p.testCS, main)
		}
		failed := false
		for {
			status, cex := p.checkClause(main)
			if status == 1 {
				break
			}
			if status < 0 {
				return -1
			}
			clauReduce(main, cex)
			if vec.Vec_IntSize(main) == 0 {
				clauComplement(main0)
				if !p.main.AddClause(vec.Vec_IntArray(main0)...) {
					return 0
				}
				failed = true
				break
			}
		}
		if failed {
			continue
		}
		basis := vec.Vec_IntAlloc(vec.Vec_IntSize(main))
		if vec.Vec_IntSize(main) > 1 {
			Fra_ClauMinimizeClause(p, basis, main)
		}
		if vec.Vec_IntSize(main) == 0 {
			panic("empty inductive clause")
		}
		if verbose != 0 && veryVerbose != 0 {
			Fra_ClauPrintClause(p.testCS, main)
		}
		clause := clauRemap(p.testToMain, main, true)
		if !p.main.AddClause(vec.Vec_IntArray(clause)...) {
			if iter+1 == nIters {
				return 0
			}
			return 1
		}
	}
}
