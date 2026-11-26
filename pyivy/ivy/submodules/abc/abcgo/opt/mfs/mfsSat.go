package mfs

import (
	"fmt"
	"os"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/misc/extra"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/sat/bsat"
	"github.com/glycerine/abcgo/sat/cnf"
)

// Port of mfsSat.c.

const (
	l_True  int32 = 1
	l_False int32 = -1
	l_Undef int32 = 0
)

func toLit(v int32) int32               { return v + v }
func toLitCond(v int32, c int32) int32  { return v + v + int32(b2i(c != 0)) }
func lit_neg(l int32) int32             { return l ^ 1 }
func lit_var(l int32) int32             { return l >> 1 }
func satConflicts(s *bsat.Solver) int64 { return s.ConflictCount() }

// sat_solver_solve mirrors the C call with the assumption range given as a slice.
func sat_solver_solve(s *bsat.Solver, assumptions []int32, nConfLimit, nInsLimit, nConfLimitGlobal, nInsLimitGlobal int64) int32 {
	return int32(s.SolveWithResourceLimits(assumptions, nConfLimit, nInsLimit, nConfLimitGlobal, nInsLimitGlobal))
}

func sat_solver_var_value(s *bsat.Solver, v int32) int32 {
	value, ok := s.Value(v)
	if ok && value {
		return 1
	}
	return 0
}

// sat_solver_addclause mirrors the C call (returns 0 on a top-level conflict).
func sat_solver_addclause(s *bsat.Solver, lits []int32) int32 {
	if s.AddClause(lits...) {
		return 1
	}
	return 0
}

// Abc_NtkMfsSolveSat_iter enumerates through the SAT assignments.
func Abc_NtkMfsSolveSat_iter(p *Mfs_Man_t) int32 {
	var Lits [MFS_FANIN_MAX]int32
	if p.nTotConfLim != 0 && int64(p.nTotConfLim) <= satConflicts(p.pSat) {
		return -1
	}
	nBTLimit := int64(0)
	if p.nTotConfLim != 0 {
		nBTLimit = int64(p.nTotConfLim) - satConflicts(p.pSat)
	}
	RetValue := sat_solver_solve(p.pSat, nil, nBTLimit, 0, 0, 0)
	assert(RetValue == l_Undef || RetValue == l_True || RetValue == l_False, "RetValue == l_Undef || RetValue == l_True || RetValue == l_False")
	if RetValue == l_Undef {
		return -1
	}
	if RetValue == l_False {
		return 0
	}
	p.nCares++
	// add SAT assignment to the solver
	Mint := int32(0)
	for b, iVar := range p.vProjVarsSat {
		Lits[b] = toLit(iVar)
		if sat_solver_var_value(p.pSat, iVar) != 0 {
			Mint |= 1 << b
			Lits[b] = lit_neg(Lits[b])
		}
	}
	assert(!util.Abc_InfoHasBit(p.uCare[:], Mint), "!Abc_InfoHasBit(p->uCare, Mint)")
	util.Abc_InfoSetBit(p.uCare[:], Mint)
	// add the blocking clause
	RetValue = sat_solver_addclause(p.pSat, Lits[:len(p.vProjVarsSat)])
	if RetValue == 0 {
		return 0
	}
	return 1
}

// Abc_NtkMfsSolveSat enumerates through the SAT assignments.
func Abc_NtkMfsSolveSat(p *Mfs_Man_t, pNode *abc.Object) int32 {
	// collect projection variables
	p.vProjVarsSat = p.vProjVarsSat[:0]
	nCos := aig.Aig_ManCoNum(p.pAigWin)
	for i := nCos - abc.Abc_ObjFaninNum(pNode); i < nCos; i++ {
		pObjPo := aig.Aig_ManCo(p.pAigWin, i)
		assert(p.pCnf.VarNums[pObjPo.NodeID()] >= 0, "p->pCnf->pVarNums[pObjPo->Id] >= 0")
		p.vProjVarsSat = append(p.vProjVarsSat, p.pCnf.VarNums[pObjPo.NodeID()])
	}

	// prepare the truth table of care set
	p.nFanins = int32(len(p.vProjVarsSat))
	p.nWords = util.Abc_TruthWordNum(p.nFanins)
	clear(p.uCare[:p.nWords])

	// iterate through the SAT assignments
	p.nCares = 0
	p.nTotConfLim = p.pPars.NBTLimit
	var RetValue int32
	for {
		RetValue = Abc_NtkMfsSolveSat_iter(p)
		if RetValue != 1 {
			break
		}
	}
	if RetValue == -1 {
		return 0
	}

	// write statistics
	p.nMintsCare += p.nCares
	p.nMintsTotal += 1 << p.nFanins

	if p.pPars.FVeryVerbose != 0 {
		fmt.Printf("Node %4d : Care = %2d. Total = %2d.  ", pNode.ID, p.nCares, 1<<p.nFanins)
		extra.Extra_PrintBinary(os.Stdout, p.uCare[:], 1<<p.nFanins)
		fmt.Printf("\n")
	}

	// map the care
	if p.nFanins > 4 {
		return 1
	}
	if p.nFanins == 4 {
		p.uCare[0] = p.uCare[0] | (p.uCare[0] << 16)
	}
	if p.nFanins == 3 {
		p.uCare[0] = p.uCare[0] | (p.uCare[0] << 8) | (p.uCare[0] << 16) | (p.uCare[0] << 24)
	}
	if p.nFanins == 2 {
		p.uCare[0] = p.uCare[0] | (p.uCare[0] << 4) | (p.uCare[0] << 8) | (p.uCare[0] << 12) |
			(p.uCare[0] << 16) | (p.uCare[0] << 20) | (p.uCare[0] << 24) | (p.uCare[0] << 28)
	}
	assert(p.nFanins != 1, "p->nFanins != 1")
	return 1
}

// Abc_NtkAddOneHotness adds one-hotness constraints for the window inputs.
func Abc_NtkAddOneHotness(p *Mfs_Man_t) int32 {
	var Lits [2]int32
	nCis := aig.Aig_ManCiNum(p.pAigWin)
	for i := int32(0); i < nCis; i++ {
		for k := i + 1; k < nCis; k++ {
			pObj1 := aig.Aig_ManCi(p.pAigWin, i)
			pObj2 := aig.Aig_ManCi(p.pAigWin, k)
			Lits[0] = toLitCond(p.pCnf.VarNums[pObj1.NodeID()], 1)
			Lits[1] = toLitCond(p.pCnf.VarNums[pObj2.NodeID()], 1)
			if sat_solver_addclause(p.pSat, Lits[:]) == 0 {
				p.pSat.Delete()
				p.pSat = nil
				return 0
			}
		}
	}
	return 1
}

var _ = cnf.Cnf_DataVarNum

func b2i(b bool) int32 {
	if b {
		return 1
	}
	return 0
}
