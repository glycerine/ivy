package saig

import (
	"fmt"
	"time"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
	"github.com/glycerine/abcgo/sat/cnf"
)

// Saig_ManStatesAreEqual reports whether two complete states have the same
// values in the solver's current satisfying assignment. A -1 entry denotes a
// latch whose value is not yet represented in that unrolling.
func Saig_ManStatesAreEqual(sat *bsat.Solver, states *Vec_Int_t, registers, first, second int32) int32 {
	assert(first != 0 && second != 0 && first < second, "first and second are ordered nonzero state indices")
	assert(registers*second <= int32(vec.Vec_IntSize(states)), "state vector contains both states")
	baseFirst, baseSecond := registers*first, registers*second
	for r := int32(0); r < registers; r++ {
		if vec.Vec_IntEntry(states, int(baseFirst+r)) >= 0 && vec.Vec_IntEntry(states, int(baseSecond+r)) < 0 {
			return 0
		}
	}
	for r := int32(0); r < registers; r++ {
		a := vec.Vec_IntEntry(states, int(baseFirst+r))
		b := vec.Vec_IntEntry(states, int(baseSecond+r))
		if a >= 0 {
			va, oka := sat.Value(a)
			vb, okb := sat.Value(b)
			if !oka || !okb || va != vb {
				return 0
			}
		}
	}
	return 1
}

// Saig_ManAddUniqueness adds a blocking clause for equal states. It returns
// one when the constraint makes the SAT instance inconsistent.
func Saig_ManAddUniqueness(sat *bsat.Solver, states *Vec_Int_t, registers, first, second int32, satVarNum, clauses *int32, verbose int32) int32 {
	assert(first != 0 && second != 0 && first < second, "first and second are ordered nonzero state indices")
	assert(registers*second <= int32(vec.Vec_IntSize(states)), "state vector contains both states")
	baseFirst, baseSecond := registers*first, registers*second
	for r := int32(0); r < registers; r++ {
		if vec.Vec_IntEntry(states, int(baseFirst+r)) >= 0 && vec.Vec_IntEntry(states, int(baseSecond+r)) < 0 {
			if verbose != 0 {
				fmt.Println("Cannot constrain an incomplete state.")
			}
			return 0
		}
	}
	oldVar := *satVarNum
	for r := int32(0); r < registers; r++ {
		a := vec.Vec_IntEntry(states, int(baseFirst+r))
		if a < 0 {
			continue
		}
		b := vec.Vec_IntEntry(states, int(baseSecond+r))
		*clauses += 4
		if cnf.Cnf_DataAddXorClause(sat, a, b, *satVarNum) == 0 {
			if verbose != 0 {
				fmt.Println("SAT solver became UNSAT after adding a uniqueness constraint.")
			}
			return 1
		}
		*satVarNum++
	}
	*clauses++
	clause := make([]bsat.Literal, 0, *satVarNum-oldVar)
	for v := oldVar; v < *satVarNum; v++ {
		clause = append(clause, bsat.Positive(v))
	}
	if !sat.AddClause(clause...) {
		if verbose != 0 {
			fmt.Println("SAT solver became UNSAT after adding a uniqueness constraint.")
		}
		return 1
	}
	return 0
}

// Saig_ManInduction performs bounded K-induction on the single output property.
// If fGetCex is set and the final bounded check is SAT, the generated one-frame
// induction witness is stored in p.Data as *util.Abc_Cex_t.
func Saig_ManInduction(p *Aig_Man_t, timeout, framesMax, conflictsMax, unique, uniqueAll, getCex, verbose, veryVerbose int32) int32 {
	if p == nil || Saig_ManPoNum(p) != 1 || aig.Aig_ManRegNum(p) <= 0 || (unique != 0 && uniqueAll != 0) {
		return -1
	}
	if timeout < 0 || framesMax < 0 || conflictsMax < 0 {
		return -1
	}
	registers := aig.Aig_ManRegNum(p)
	pis := Saig_ManPiNum(p)
	top := []aig.Edge{aig.Aig_ManCo(p, 0)}
	var bottom []aig.Edge
	topVarNums := make([]int32, 0, registers)
	states := vec.Vec_IntAlloc(1000)
	var topVarIDs []int32
	sat := bsat.NewSolver()
	if timeout > 0 {
		sat.SetDeadline(time.Now().Add(time.Duration(timeout) * time.Second))
	}
	retValue, status := int32(-1), bsat.Unknown
	var nSatVarNum, nClauses, constraints int32
	started := time.Now()
	if verbose != 0 {
		fmt.Printf("Induction parameters: FramesMax = %5d. ConflictMax = %6d.\n", framesMax, conflictsMax)
	}

	for frame := int32(0); ; frame++ {
		frameStart := time.Now()
		if len(bottom) != 0 {
			bottom = bottom[:0]
		}
		aig.Aig_SupportNodes(p, top, int32(len(top)), &bottom)
		part := aig.Aig_ManDupSimpleDfsPart(p, bottom, top)
		partCnf := cnf.Cnf_Derive(part, int32(len(part.Outputs())))
		if partCnf == nil {
			return -1
		}
		cnf.Cnf_DataLift(partCnf, nSatVarNum)
		nSatVarNum += int32(partCnf.Variables)
		nClauses += int32(len(partCnf.Clauses))

		if getCex != 0 && topVarIDs == nil {
			topVarIDs = make([]int32, aig.Aig_ManCiNum(p))
			for i := range topVarIDs {
				topVarIDs[i] = -1
			}
			for _, pi := range p.Inputs() {
				copyEdge := aig.Aig_ObjCopy(p, pi)
				if copyEdge == aig.InvalidEdge {
					continue
				}
				copyID := aig.Aig_ObjId(copyEdge)
				if Saig_ObjIsPi(p, pi) {
					index := aig.Aig_ObjCioId(p, pi) + registers
					topVarIDs[index] = partCnf.VarNums[copyID]
				} else if Saig_ObjIsLo(p, pi) {
					index := aig.Aig_ObjCioId(p, pi) - pis
					topVarIDs[index] = partCnf.VarNums[copyID]
				}
			}
		}

		partOutputs := part.Outputs()
		assert(int32(len(partOutputs))-1 == int32(len(topVarNums)), "top output variable count matches")
		for i, output := range partOutputs {
			outputVar := partCnf.VarNums[aig.Aig_ObjId(output)]
			if i == 0 {
				lit := bsat.Positive(outputVar)
				if frame > 0 {
					lit = bsat.Negative(outputVar)
				}
				if !sat.AddClause(lit) {
					assert(false, "top output unit clause remains consistent")
				}
				nClauses++
				continue
			}
			topVar := topVarNums[i-1]
			if !sat.AddClause(bsat.Positive(topVar), bsat.Negative(outputVar)) || !sat.AddClause(bsat.Negative(topVar), bsat.Positive(outputVar)) {
				assert(false, "top/bottom stitch clauses remain consistent")
			}
			nClauses += 2
		}
		for _, rawClause := range partCnf.Clauses {
			clause := make([]bsat.Literal, len(rawClause))
			for i, literal := range rawClause {
				if literal&1 != 0 {
					clause[i] = bsat.Negative(literal >> 1)
				} else {
					clause[i] = bsat.Positive(literal >> 1)
				}
			}
			if !sat.AddClause(clause...) {
				retValue = 1
				break
			}
		}
		if retValue == 1 {
			cnf.Cnf_DataFree(partCnf)
			aig.Aig_ManStop(part)
			break
		}

		top = []aig.Edge{aig.Aig_ManCo(p, 0)}
		topVarNums = topVarNums[:0]
		oldStateSize := int32(vec.Vec_IntSize(states))
		vec.Vec_IntFillExtra(states, int(oldStateSize+registers), -1)
		for _, input := range bottom {
			if !Saig_ObjIsLo(p, input) {
				continue
			}
			copyEdge := aig.Aig_ObjCopy(p, input)
			assert(copyEdge != aig.InvalidEdge, "bottom latch output has a copied input")
			top = append(top, Saig_ObjLoToLi(p, input))
			varNum := partCnf.VarNums[aig.Aig_ObjId(copyEdge)]
			topVarNums = append(topVarNums, varNum)
			reg := aig.Aig_ObjCioId(p, input) - pis
			assert(reg >= 0 && reg < registers, "latch output index is in range")
			vec.Vec_IntWriteEntry(states, int(oldStateSize+reg), varNum)
		}
		assert(int32(vec.Vec_IntSize(states))%registers == 0, "state vector contains complete frames")
		lastState := int32(vec.Vec_IntSize(states)) / registers

		if uniqueAll != 0 {
			var i int32
			for i = 1; i < lastState-1; i++ {
				constraints++
				if veryVerbose != 0 {
					fmt.Printf("Adding constraint for state %2d and state %2d.\n", i, lastState-1)
				}
				if Saig_ManAddUniqueness(sat, states, registers, i, lastState-1, &nSatVarNum, &nClauses, verbose) != 0 {
					break
				}
			}
			if i < lastState-1 {
				retValue = 1
				cnf.Cnf_DataFree(partCnf)
				aig.Aig_ManStop(part)
				break
			}
		}

		for {
			nConflictsBefore := sat.Statistics().Conflicts
			sat.SetLimits(int64(conflictsMax), 0)
			status = sat.Solve()
			if verbose != 0 {
				fmt.Printf("Frame %4d : PI =%5d. PO =%5d. AIG =%5d. Var =%7d. Clau =%7d. Conf =%7d. Time = %9.2f sec\n", frame, len(part.Inputs()), len(part.Outputs()), aig.Aig_ManNodeNum(part), nSatVarNum, nClauses, sat.Statistics().Conflicts-nConflictsBefore, time.Since(frameStart).Seconds())
			}
			if status == bsat.Unknown {
				break
			}
			if status == bsat.Unsatisfiable {
				retValue = 1
				break
			}
			if veryVerbose != 0 {
				printInductionStates(sat, states, registers)
			}
			if framesMax != 0 && frame == framesMax-1 {
				if getCex != 0 {
					model := util.Abc_CexAlloc(registers-1, pis, 1)
					model.IFrame, model.IPo = 0, 0
					bit := int32(0)
					for i := 1; i < len(topVarIDs); i++ {
						variable := topVarIDs[i]
						if variable >= 0 {
							if value, ok := sat.Value(variable); ok && value {
								util.Abc_InfoSetBit(model.PData, bit)
							}
						}
						bit++
					}
					assert(bit == model.NBits, "counterexample bit count matches")
					assert(p.SetInductionCounterexample(model) == nil, "compact induction counterexample dimensions are valid")
				}
				retValue = 0
				break
			}
			if unique != 0 {
				added := false
				for i := int32(1); i < lastState && !added; i++ {
					for k := i + 1; k < lastState; k++ {
						if Saig_ManStatesAreEqual(sat, states, registers, i, k) == 0 {
							continue
						}
						constraints++
						added = true
						if veryVerbose != 0 {
							fmt.Printf("Adding constraint for state %2d and state %2d.\n", i, k)
						}
						if Saig_ManAddUniqueness(sat, states, registers, i, k, &nSatVarNum, &nClauses, verbose) != 0 {
							retValue = 1
							break
						}
					}
				}
				if retValue == 1 {
					break
				}
				if added {
					continue
				}
			}
			break
		}
		cnf.Cnf_DataFree(partCnf)
		aig.Aig_ManStop(part)
		if retValue == 1 || status == bsat.Unsatisfiable || status == bsat.Unknown || (framesMax != 0 && frame == framesMax-1) {
			break
		}
	}

	if verbose != 0 {
		if timeout > 0 && time.Since(started) >= time.Duration(timeout)*time.Second {
			fmt.Printf("Timeout (%d sec) was reached during induction.\n", timeout)
		} else if status == bsat.Unknown {
			fmt.Printf("Conflict limit (%d) was reached during induction.\n", conflictsMax)
		} else if unique != 0 || uniqueAll != 0 {
			fmt.Printf("Completed induction and added %d uniqueness constraints.\n", constraints)
		} else {
			fmt.Println("Completed induction.")
		}
	}
	vec.Vec_IntFree(states)
	return retValue
}

func printInductionStates(sat *bsat.Solver, states *Vec_Int_t, registers int32) {
	for i, variable := range vec.Vec_IntArray(states) {
		if i != 0 && int32(i)%registers == 0 {
			fmt.Println()
		}
		if int32(i)%registers == 0 {
			fmt.Printf("       State %3d : ", int32(i)/registers)
		}
		if variable < 0 {
			fmt.Print("x")
			continue
		}
		value, ok := sat.Value(variable)
		if !ok {
			fmt.Print("x")
		} else if value {
			fmt.Print("1")
		} else {
			fmt.Print("0")
		}
	}
	fmt.Println()
}

// Saig_ManInductionCounterexample returns the witness saved by a successful
// final bounded check, if fGetCex was enabled.
func Saig_ManInductionCounterexample(p *Aig_Man_t) *util.Abc_Cex_t {
	if p == nil {
		return nil
	}
	return p.InductionCounterexample()
}
