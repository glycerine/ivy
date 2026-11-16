package if_

import (
	"fmt"
	abcutil "github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/opt/dau"
	"github.com/glycerine/abcgo/sat/bsat"
)

// If_ManSatBuildXY constructs the original parameterized two-LUT MUX CNF.
func If_ManSatBuildXY(nLutSize int32) any {
	if nLutSize < 2 || nLutSize > 6 {
		panic("if: two-LUT decomposition supports LUT sizes 2..6")
	}
	mintsL := int32(1) << uint(nLutSize)
	mintsF := int32(1) << uint(2*nLutSize-1)
	solver := bsat.NewSolver()
	solver.SetNumVars(2*mintsL + mintsF)
	for m := int32(0); m < mintsF; m++ {
		solver.AddMux(2*mintsL+m, m%mintsL, mintsL+2*(m/mintsL)+1, mintsL+2*(m/mintsL), false, false, false, false)
	}
	return solver
}

// If_ManSatBuildXYZ constructs the original three-LUT four-way MUX CNF.
func If_ManSatBuildXYZ(nLutSize int32) any {
	if nLutSize < 2 || nLutSize > 6 {
		panic("if: three-LUT decomposition supports LUT sizes 2..6")
	}
	mintsL := int32(1) << uint(nLutSize)
	mintsF := int32(1) << uint(3*nLutSize-2)
	solver := bsat.NewSolver()
	solver.SetNumVars(3*mintsL + mintsF)
	for m := int32(0); m < mintsF; m++ {
		base := 2*mintsL + 4*(m>>uint(2*nLutSize))
		solver.AddMux4(3*mintsL+m, m%mintsL, mintsL+(m>>uint(nLutSize))%mintsL, base, base+1, base+2, base+3)
	}
	return solver
}
func If_ManSatUnbuild(handle any) {
	if handle == nil {
		return
	}
	solver, ok := handle.(*bsat.Solver)
	if !ok {
		panic("if: invalid SAT solver handle")
	}
	if solver != nil {
		solver.Close()
	}
}

func If_ManSat6ComposeLut4(table uint32, functions []word, nVars int32) word {
	if nVars < 0 || nVars > 4 || len(functions) < int(nVars) {
		panic("if: invalid 4-LUT composition inputs")
	}
	var result word
	for assignment := 0; assignment < 64; assignment++ {
		index := 0
		for v := int32(0); v < nVars; v++ {
			if functions[v]&(word(1)<<uint(assignment)) != 0 {
				index |= 1 << uint(v)
			}
		}
		if table&(1<<uint(index)) != 0 {
			result |= word(1) << uint(assignment)
		}
	}
	return result
}

func If_ManSat6Truth(bound, free word, boundSet, sharedSet, freeSet []int32) word {
	boundVars := make([]word, 0, len(boundSet)+len(sharedSet))
	for _, v := range sharedSet {
		boundVars = append(boundVars, truth6Variable(v))
	}
	for _, v := range boundSet {
		boundVars = append(boundVars, truth6Variable(v))
	}
	q := If_ManSat6ComposeLut4(uint32(bound), boundVars, int32(len(boundVars)))
	freeVars := []word{q}
	for _, v := range sharedSet {
		freeVars = append(freeVars, truth6Variable(v))
	}
	for _, v := range freeSet {
		freeVars = append(freeVars, truth6Variable(v))
	}
	return If_ManSat6ComposeLut4(uint32(free), freeVars, int32(len(freeVars)))
}

func truth6Variable(v int32) word {
	if v < 0 || v >= 6 {
		panic("if: truth-table variable outside 0..5")
	}
	var truth word
	for assignment := 0; assignment < 64; assignment++ {
		if assignment&(1<<uint(v)) != 0 {
			truth |= word(1) << uint(assignment)
		}
	}
	return truth
}

// If_ManSatCheckXY remaps truth minterms to assumptions on the original
// parameterized SAT instance, solves it, and extracts both LUT configurations.
func If_ManSatCheckXY(sat any, lutSize int32, truth []word, nVars int32, set uint32, boundOut, freeOut *word, scratch *Vec_Int_t) int32 {
	return ifManSatCheckXY(sat, lutSize, truth, nVars, set, boundOut, freeOut, scratch)
}

func ifManSatCheckXY(handle any, lutSize int32, truth []word, nVars int32, set uint32, boundOut, freeOut *word, scratch *Vec_Int_t) int32 {
	if nVars < 0 || lutSize < 2 || lutSize > 6 || nVars > 2*lutSize-1 || len(truth) < (1<<uint(nVars)+63)/64 {
		return 0
	}
	var boundVars, sharedVars, freeVars [IF_MAX_FUNC_LUTSIZE]int32
	var nBound, nShared, nFree int32
	for v := int32(0); v < nVars; v++ {
		switch set >> uint(2*v) & 3 {
		case 0:
			freeVars[nFree] = v
			nFree++
		case 1:
			boundVars[nBound] = v
			nBound++
		case 3:
			sharedVars[nShared] = v
			nShared++
		default:
			return 0
		}
	}
	if nShared+nBound > lutSize || nShared+nFree+1 > lutSize {
		return 0
	}
	// Retain the existing Go convenience of nil handles, using the same real
	// SAT algorithm with temporary ownership rather than a substitute solver.
	if handle == nil {
		handle = If_ManSatBuildXY(lutSize)
		defer If_ManSatUnbuild(handle)
	}
	solver, ok := handle.(*bsat.Solver)
	mintsL := int32(1) << uint(lutSize)
	mintsF := int32(1) << uint(2*lutSize-1)
	if !ok || solver == nil || solver.NumVars() != 2*mintsL+mintsF {
		panic("if: SAT instance does not match two-LUT dimensions")
	}
	if scratch == nil {
		scratch = vec.Vec_IntAlloc(int(mintsF))
	}
	vec.Vec_IntFill(scratch, int(mintsF), -1)
	for m := int32(0); m < int32(1)<<uint(nVars); m++ {
		var mapped, ib, is, iff int32
		for v := int32(0); v < nVars; v++ {
			value := m >> uint(v) & 1
			switch set >> uint(2*v) & 3 {
			case 0:
				if value != 0 {
					mapped |= 1 << uint(lutSize+nShared+iff)
				}
				iff++
			case 1:
				if value != 0 {
					mapped |= 1 << uint(nShared+ib)
				}
				ib++
			case 3:
				if value != 0 {
					mapped |= 1 << uint(is)
					mapped |= 1 << uint(lutSize+is)
				}
				is++
			}
		}
		if vec.Vec_IntEntry(scratch, int(mapped)) != -1 {
			panic("if: duplicate SAT minterm mapping")
		}
		bit := int32(truth[m>>6] >> uint(m&63) & 1)
		vec.Vec_IntWriteEntry(scratch, int(mapped), bit)
	}
	count := 0
	for m, value := range vec.Vec_IntArray(scratch) {
		if value >= 0 {
			vec.Vec_IntWriteEntry(scratch, count, 2*(2*mintsL+int32(m))+(1-value))
			count++
		}
	}
	vec.Vec_IntShrink(scratch, count)
	if solver.SolveWithResourceLimits(vec.Vec_IntArray(scratch), 0, 0, 0, 0) != bsat.Satisfiable {
		return 0
	}
	if boundOut != nil && freeOut != nil {
		var bound, free word
		for m := int32(0); m < int32(1)<<uint(nShared+nBound); m++ {
			value, _ := solver.Value(m)
			if value {
				bound |= 1 << uint(m)
			}
		}
		for m := int32(0); m < int32(1)<<uint(1+nShared+nFree); m++ {
			value, _ := solver.Value(mintsL + m)
			if value {
				free |= 1 << uint(m)
			}
		}
		*boundOut = abcutil.Abc_Tt6Stretch(bound, nShared+nBound)
		*freeOut = abcutil.Abc_Tt6Stretch(free, 1+nShared+nFree)
		if nVars == 6 && lutSize == 4 && If_ManSat6Truth(*boundOut, *freeOut, boundVars[:nBound], sharedVars[:nShared], freeVars[:nFree]) != truth[0] {
			dau.Dau_DsdPrintFromTruth(truth, nVars)
			result := If_ManSat6Truth(*boundOut, *freeOut, boundVars[:nBound], sharedVars[:nShared], freeVars[:nFree])
			dau.Dau_DsdPrintFromTruth([]word{result}, nVars)
			dau.Dau_DsdPrintFromTruth([]word{*boundOut}, nShared+nBound)
			dau.Dau_DsdPrintFromTruth([]word{*freeOut}, 1+nShared+nFree)
			fmt.Println("Verification failed!")
		}
	}
	return 1
}

// If_ManSatCheckXYall searches bound sets and increasing numbers of shared
// inputs in the same order as the source implementation.
func If_ManSatCheckXYall(pSat any, lutSize int32, truth []word, nVars int32, scratch *Vec_Int_t) uint32 {
	if lutSize < 2 || lutSize > 6 || nVars <= lutSize || nVars > 2*lutSize-1 || len(truth) < (1<<uint(nVars)+63)/64 {
		return 0
	}
	if pSat == nil {
		pSat = If_ManSatBuildXY(lutSize)
		defer If_ManSatUnbuild(pSat)
	}
	maxShared := 2*lutSize - 1 - nVars
	var result uint32
	bound := make([]int32, lutSize)
	for sharedCount := int32(0); sharedCount <= maxShared; sharedCount++ {
		var chooseBound func(int32, int32) bool
		chooseBound = func(depth, start int32) bool {
			if depth == lutSize {
				var chooseShared func(int32, int32, uint32) bool
				chooseShared = func(remaining, start int32, set uint32) bool {
					if remaining == 0 {
						if ifManSatCheckXY(pSat, lutSize, truth, nVars, set, nil, nil, scratch) != 0 {
							result = set
							return true
						}
						return false
					}
					for i := start; i <= lutSize-remaining; i++ {
						if chooseShared(remaining-1, i+1, set|(3<<uint(2*bound[i]))) {
							return true
						}
					}
					return false
				}
				var set uint32
				for _, variable := range bound {
					set |= 1 << uint(2*variable)
				}
				return chooseShared(sharedCount, 0, set)
			}
			for i := start; i <= nVars-(lutSize-depth); i++ {
				bound[depth] = i
				if chooseBound(depth+1, i+1) {
					return true
				}
			}
			return false
		}
		if chooseBound(0, 0) {
			return result
		}
	}
	return 0
}

func If_ManSatCheckXYall_int(pSat any, lutSize int32, truth []word, nVars int32, scratch *Vec_Int_t) uint32 {
	return If_ManSatCheckXYall(pSat, lutSize, truth, nVars, scratch)
}

func If_ManSatTest2() {
	const nVars, lutSize = 6, 4
	solver := If_ManSatBuildXY(lutSize)
	defer If_ManSatUnbuild(solver)
	truth := dau.Dau_DsdToTruth("0123456789ABCDEF{abcdef}", nVars)
	scratch := vec.Vec_IntAlloc(100)
	defer vec.Vec_IntFree(scratch)
	var bound, free word
	const set = 1<<0 | 3<<2 | 1<<4 | 1<<6
	if If_ManSatCheckXY(solver, lutSize, truth, nVars, set, &bound, &free, scratch) == 0 {
		panic("if: example two-LUT decomposition failed")
	}
	dau.Dau_DsdPrintFromTruth(truth, nVars)
	dau.Dau_DsdPrintFromTruth([]word{bound}, lutSize)
	dau.Dau_DsdPrintFromTruth([]word{free}, lutSize)
}

func If_ManSatTest3() {
	const nVars, lutSize = 6, 4
	solver := If_ManSatBuildXY(lutSize)
	defer If_ManSatUnbuild(solver)
	truth := dau.Dau_DsdToTruth("0123456789ABCDEF{abcdef}", nVars)
	scratch := vec.Vec_IntAlloc(100)
	defer vec.Vec_IntFree(scratch)
	set := If_ManSatCheckXYall(solver, lutSize, truth, nVars, scratch)
	dau.Dau_DecPrintSet(set, nVars, 1)
}
