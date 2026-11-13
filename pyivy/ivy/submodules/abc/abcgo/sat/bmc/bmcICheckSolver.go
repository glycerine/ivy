package bmc

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/sat/bsat"
	"github.com/glycerine/abcgo/sat/cnf"
	"time"
)

// Bmc_DeriveSolver constructs the two-copy M-inductiveness SAT problem.
func Bmc_DeriveSolver(p, m *gia.Gia_Man_t, data *cnf.Cnf_Dat_t, frames, timeout, verbose int32) *bsat.Solver {
	if p == nil || m == nil || data == nil || frames < 0 {
		panic("bmc: invalid induction solver")
	}
	_ = verbose
	regs, cos := gia.Gia_ManRegNum(p), gia.Gia_ManCoNum(p)
	base := regs + cos
	stride := int32(data.Variables)
	s := bsat.NewSolver()
	for s.NumVars() < base+(frames+1)*stride {
		s.NewVariable()
	}
	if timeout > 0 {
		s.SetDeadline(time.Now().Add(time.Duration(timeout) * time.Second))
	}
	variable := func(obj *gia.Gia_Obj_t, f int32) int32 {
		id := gia.Gia_ObjId(m, obj)
		if int(id) >= len(data.VarNums) || data.VarNums[id] < 0 {
			panic("bmc: missing induction CNF variable")
		}
		return base + stride*f + data.VarNums[id]
	}
	add := func(lits ...bsat.Literal) { s.AddClause(lits...) }
	equal := func(a, b, enable int32) {
		x, y := bsat.Positive(a), bsat.Positive(b)
		if enable < 0 {
			add(x^1, y)
			add(x, y^1)
		} else {
			e := bsat.Negative(enable)
			add(e, x^1, y)
			add(e, x, y^1)
		}
	}
	xor := func(a, b, z, enable int32) {
		n := 3
		if enable >= 0 {
			n = 4
		}
		for bits := 0; bits < 1<<n; bits++ {
			av, bv, zv := bits&1 != 0, bits&2 != 0, bits&4 != 0
			want := av != bv
			if enable >= 0 {
				want = want && (bits&8 != 0)
			}
			if zv == want {
				continue
			}
			ids := []int32{a, b, z}
			if enable >= 0 {
				ids = append(ids, enable)
			}
			clause := make([]bsat.Literal, len(ids))
			for i, id := range ids {
				clause[i] = bsat.Positive(id)
				if bits&(1<<i) != 0 {
					clause[i] ^= 1
				}
			}
			add(clause...)
		}
	}
	diff := make([]bsat.Literal, cos)
	for i := range diff {
		diff[i] = bsat.Positive(regs + int32(i))
	}
	add(diff...)
	for i := int32(0); i < gia.Gia_ManPoNum(p); i++ {
		xor(variable(gia.Gia_ManPo(m, 2*i), 0), variable(gia.Gia_ManPo(m, 2*i+1), 0), regs+i, -1)
	}
	for i := int32(0); i < regs; i++ {
		xor(variable(gia.Gia_ManRi(m, i), 0), variable(gia.Gia_ManRi(m, i+regs), 0), regs+gia.Gia_ManPoNum(p)+i, i)
	}
	for f := int32(0); f <= frames; f++ {
		for _, clause := range data.Clauses {
			lits := make([]bsat.Literal, len(clause))
			for i, l := range clause {
				lits[i] = l + 2*(base+f*stride)
			}
			add(lits...)
		}
		if f == 0 {
			continue
		}
		for i := int32(0); i < gia.Gia_ManRegNum(m); i++ {
			ro := gia.Gia_ManRo(m, i)
			id := gia.Gia_ObjId(m, ro)
			if data.VarNums[id] >= 0 {
				equal(variable(gia.Gia_ManRi(m, i), f), variable(ro, f-1), -1)
			}
		}
		for i := int32(0); i < gia.Gia_ManPoNum(p); i++ {
			equal(variable(gia.Gia_ManPo(m, 2*i), f), variable(gia.Gia_ManPo(m, 2*i+1), f), -1)
		}
		for i := int32(0); i < regs; i++ {
			equal(variable(gia.Gia_ManRi(m, i), f), variable(gia.Gia_ManRi(m, i+regs), f), i)
		}
	}
	return s
}
