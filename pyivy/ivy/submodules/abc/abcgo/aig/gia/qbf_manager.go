package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/satoko"
	"time"
)

type Qbf_Man_t struct {
	graph                   *Gia_Man_t
	nPars, nVars            int32
	verbose                 int32
	verification, synthesis *satoko.Solver
	parameterMap            []int32
	started                 time.Time
}

func Gia_QbfAlloc(g *Gia_Man_t, nPars, glucose, verbose int32) *Qbf_Man_t {
	_ = glucose
	if g == nil || g.nRegs != 0 || Gia_ManPoNum(g) != 1 || nPars < 0 || nPars > Gia_ManPiNum(g) {
		panic("gia: invalid QBF manager")
	}
	n := Gia_ManPiNum(g)
	ver := satoko.New()
	ver.SetVariables(int(n))
	inputs := make([]int, n)
	for i := range inputs {
		inputs[i] = 2 * i
	}
	output := qbfEncodeGraph(ver, g, inputs)
	ver.AddClause([]int{output ^ 1})
	syn := satoko.New()
	syn.SetVariables(int(nPars))
	mapping := make([]int32, nPars)
	for i := range mapping {
		mapping[i] = -1
	}
	return &Qbf_Man_t{graph: g, nPars: nPars, nVars: n - nPars, verbose: verbose, verification: ver, synthesis: syn, parameterMap: mapping, started: time.Now()}
}
func Gia_QbfFree(p *Qbf_Man_t) {
	if p != nil {
		p.verification = nil
		p.synthesis = nil
	}
}
func Gia_QbfVerify(p *Qbf_Man_t, values *Vec_Int_t) int32 {
	if vec.Vec_IntSize(values) != int(p.nPars) {
		panic("gia: invalid QBF parameter assignment")
	}
	assumptions := make([]int, p.nPars)
	for i, v := range vec.Vec_IntArray(values) {
		assumptions[i] = 2 * i
		if v == 0 {
			assumptions[i] ^= 1
		}
	}
	if p.verification.SolveAssumptions(assumptions) != satoko.SATOKO_SAT {
		return 0
	}
	vec.Vec_IntClear(values)
	for i := int32(0); i < p.nVars; i++ {
		vec.Vec_IntPush(values, int32(p.verification.ModelValue(int(p.nPars+i))))
	}
	return 1
}
func Gia_QbfAddCofactor(p *Qbf_Man_t, cofactor *Gia_Man_t) int32 {
	if Gia_ManPiNum(cofactor) != p.nPars {
		panic("gia: invalid QBF cofactor inputs")
	}
	inputs := make([]int, p.nPars)
	for i := range inputs {
		inputs[i] = 2 * i
	}
	output := qbfEncodeGraph(p.synthesis, cofactor, inputs)
	return giaBool(p.synthesis.AddClause([]int{output}))
}
func Gia_QbfAddCofactorG(p *Qbf_Man_t, cofactor *Gia_Man_t) int32 {
	return Gia_QbfAddCofactor(p, cofactor)
}
func Gia_QbfOnePattern(p *Qbf_Man_t, values *Vec_Int_t) {
	vec.Vec_IntClear(values)
	for i := int32(0); i < p.nPars; i++ {
		vec.Vec_IntPush(values, int32(p.synthesis.ModelValue(int(i))))
	}
}
func Gia_QbfPrint(p *Qbf_Man_t, values *Vec_Int_t, iteration int32) {
	fmt.Printf("%5d : ", iteration)
	for _, v := range vec.Vec_IntArray(values) {
		fmt.Printf("%d", v)
	}
	fmt.Printf("  Var =%7d  Cla =%7d  Time = %s\n", p.synthesis.VariableCount(), p.synthesis.ClauseCount(), time.Since(p.started).Round(time.Millisecond))
}
func Gia_QbfAddSpecialConstr(p *Qbf_Man_t) {
	if p.nPars < 44 {
		panic("gia: QBF special constraints require 44 parameters")
	}
	for i := 0; i < 44; i += 4 {
		if p.parameterMap[i] != -1 {
			panic("gia: QBF parameter already fixed")
		}
		p.parameterMap[i] = 0
		if !p.synthesis.AddClause([]int{2*i + 1}) {
			panic("gia: inconsistent QBF constraints")
		}
	}
}
func Gia_QbfLearnConstraint(p *Qbf_Man_t, values *Vec_Int_t) {
	if vec.Vec_IntSize(values) != int(p.nPars) {
		panic("gia: invalid QBF learned assignment")
	}
	fmt.Print("  Pattern   ")
	for _, v := range vec.Vec_IntArray(values) {
		fmt.Printf("%d", v)
	}
	fmt.Println()
	for i, v := range vec.Vec_IntArray(values) {
		opposite := 2*i + int(v&1)
		status := p.synthesis.SolveAssumptions([]int{opposite})
		fmt.Printf("  Var =%4d ", i)
		if status != satoko.SATOKO_SAT {
			fmt.Println("UNSAT")
			p.synthesis.AddClause([]int{opposite ^ 1})
			continue
		}
		for j := 0; j < int(p.nPars); j++ {
			fmt.Printf("%d", p.synthesis.ModelValue(j))
		}
		fmt.Println()
	}
}
func Cnf_SpecialDataLift(cnf *Cnf_Dat_t, offset, first, last int32) {
	for _, clause := range cnf.Clauses {
		for i, lit := range clause {
			variable := lit >> 1
			if variable < first || variable >= last {
				clause[i] += 2 * offset
			} else {
				clause[i] -= 2 * first
			}
		}
	}
}
