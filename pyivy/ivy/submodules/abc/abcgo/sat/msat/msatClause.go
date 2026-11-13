// Go port of msatClause.c by Alan Mishchenko, UC Berkeley (2004).
package msat

import (
	"fmt"
	"io"
	"sort"
)

type Msat_Clause_t struct {
	Num                           int
	learned, mark, typeA, deleted bool
	lits                          []int
	activity                      float32
}

func Msat_ClauseCreate(p *Msat_Solver_t, v *Msat_IntVec_t, learnt int, out **Msat_Clause_t) int {
	*out = nil
	lits := append([]int(nil), v.data...)
	if learnt == 0 {
		for _, lit := range lits {
			if lit < 0 || lit>>1 >= p.nVars {
				panic("literal variable out of range")
			}
		}
		// The C implementation sorts the caller's vector and compacts it in place.
		sort.Ints(v.data)
		Msat_SolverIncrementSeenId(p)
		seenID := Msat_SolverIncrementSeenId(p)
		j := 0
		for i := range v.data {
			lit := v.data[i]
			variable, sign := lit>>1, lit&1
			if p.pSeen[variable] >= seenID-1 {
				if (p.pSeen[variable] != seenID) == (sign != 0) {
					continue
				}
				return 1
			}
			p.pSeen[variable] = seenID - (1 - sign)
			if p.pAssigns[variable] >= 0 {
				if p.pAssigns[variable] == lit {
					return 1
				}
				continue
			}
			v.data[j] = lit
			j++
		}
		v.data = v.data[:j]
		lits = append(lits[:0], v.data...)
	}
	if len(lits) == 0 {
		p.okay = false
		return 0
	}
	if len(lits) == 1 {
		return Msat_SolverEnqueue(p, lits[0], nil)
	}
	c := &Msat_Clause_t{Num: len(p.vClauses.data) + len(p.vLearned.data), learned: learnt != 0, lits: lits}
	if c.learned {
		best := 1
		for i := 2; i < len(lits); i++ {
			if p.pLevel[lits[i]>>1] > p.pLevel[lits[best]>>1] {
				best = i
			}
		}
		c.lits[1], c.lits[best] = c.lits[best], c.lits[1]
		Msat_SolverClaBumpActivity(p, c)
		for _, lit := range lits {
			Msat_SolverVarBumpActivity(p, lit)
		}
	}
	Msat_ClauseVecPush(p.pvWatched[c.lits[0]^1], c)
	Msat_ClauseVecPush(p.pvWatched[c.lits[1]^1], c)
	*out = c
	return 1
}
func Msat_ClauseCreateFake(p *Msat_Solver_t, v *Msat_IntVec_t) *Msat_Clause_t {
	return &Msat_Clause_t{lits: append([]int(nil), v.data...)}
}
func Msat_ClauseCreateFakeLit(p *Msat_Solver_t, lit int) *Msat_Clause_t {
	return &Msat_Clause_t{lits: []int{lit}}
}
func Msat_ClauseReadLearned(c *Msat_Clause_t) int { return boolInt(c.learned) }
func Msat_ClauseReadSize(c *Msat_Clause_t) int    { return len(c.lits) }
func Msat_ClauseReadLits(c *Msat_Clause_t) []int  { return c.lits }
func Msat_ClauseReadMark(c *Msat_Clause_t) int    { return boolInt(c.mark) }
func Msat_ClauseSetMark(c *Msat_Clause_t, v int)  { c.mark = v != 0 }
func Msat_ClauseReadNum(c *Msat_Clause_t) int     { return c.Num }
func Msat_ClauseSetNum(c *Msat_Clause_t, n int)   { c.Num = n }
func Msat_ClauseReadTypeA(c *Msat_Clause_t) int   { return boolInt(c.typeA) }
func Msat_ClauseSetTypeA(c *Msat_Clause_t, v int) { c.typeA = v != 0 }
func Msat_ClauseIsLocked(p *Msat_Solver_t, c *Msat_Clause_t) int {
	for v, r := range p.pReasons[:p.nVars] {
		if r == c && p.pAssigns[v] >= 0 {
			return 1
		}
	}
	return 0
}
func Msat_ClauseReadActivity(c *Msat_Clause_t) float32     { return c.activity }
func Msat_ClauseWriteActivity(c *Msat_Clause_t, a float32) { c.activity = a }
func Msat_ClauseRemoveWatch(v *Msat_ClauseVec_t, c *Msat_Clause_t) {
	for i, other := range v.data {
		if other == c {
			copy(v.data[i:], v.data[i+1:])
			v.data = v.data[:len(v.data)-1]
			return
		}
	}
}
func Msat_ClauseFree(p *Msat_Solver_t, c *Msat_Clause_t, remove int) {
	if remove != 0 && len(c.lits) > 1 {
		Msat_ClauseRemoveWatch(p.pvWatched[c.lits[0]^1], c)
		Msat_ClauseRemoveWatch(p.pvWatched[c.lits[1]^1], c)
	}
	c.deleted = true
	c.lits = nil
}
func Msat_ClausePropagate(c *Msat_Clause_t, lit int, assigns []int, out *int) int {
	*out = -1
	falseLit := lit ^ 1
	if c.lits[0] == falseLit {
		c.lits[0], c.lits[1] = c.lits[1], c.lits[0]
	}
	if assigns[c.lits[0]>>1] == c.lits[0] {
		return 1
	}
	for i := 2; i < len(c.lits); i++ {
		q := c.lits[i]
		if assigns[q>>1] < 0 || assigns[q>>1] == q {
			c.lits[1], c.lits[i] = q, falseLit
			*out = q ^ 1
			return 1
		}
	}
	*out = c.lits[0]
	return 0
}
func Msat_ClauseSimplify(c *Msat_Clause_t, assigns []int) int {
	for _, lit := range c.lits {
		if assigns[lit>>1] == lit {
			return 1
		}
	}
	return 0
}
func Msat_ClauseCalcReason(p *Msat_Solver_t, c *Msat_Clause_t, implied int, out *Msat_IntVec_t) {
	out.data = out.data[:0]
	for _, lit := range c.lits {
		if lit == implied {
			continue
		}
		out.data = append(out.data, lit^1)
	}
}
func Msat_ClauseWriteDimacs(w io.Writer, c *Msat_Clause_t, increment int) {
	for _, lit := range c.lits {
		v := lit >> 1
		if increment != 0 {
			v++
		}
		if lit&1 != 0 {
			fmt.Fprintf(w, "-%d ", v)
		} else {
			fmt.Fprintf(w, "%d ", v)
		}
	}
	if increment != 0 {
		fmt.Fprint(w, "0")
	}
	fmt.Fprintln(w)
}
func Msat_ClausePrint(c *Msat_Clause_t) {
	if c == nil {
		fmt.Println("NULL pointer")
		return
	}
	for _, lit := range c.lits {
		sign := ""
		if lit&1 != 0 {
			sign = "-"
		}
		fmt.Printf(" %s%d", sign, (lit>>1)+1)
	}
	fmt.Println()
}
func Msat_ClausePrintSymbols(c *Msat_Clause_t) { Msat_ClausePrint(c) }
