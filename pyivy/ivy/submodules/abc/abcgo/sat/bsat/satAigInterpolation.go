package bsat

import (
	"fmt"
	"io"
	"os"

	"github.com/glycerine/abcgo/aig/aig"
)

// Both AIG interpolation variants share the resolution engine. Variant B uses
// the shared pivot as a selector; variant A uses McMillan conjunction instead.
type aigInterpolation struct {
	pCnf                               *Sto_Man_t
	pAig                               *aig.Graph
	vVarsAB, pVarTypes                 []int32
	pInters                            []aig.Edge
	fVerbose, fProofVerif, fProofWrite int32
	Counter                            int32
	muxShared                          bool
	lastErr                            error
}

func (p *aigInterpolation) root(c *Sto_Cls_t) {
	value := p.pAig.Const1()
	if c.fA != 0 {
		value = p.pAig.Const0()
		for _, l := range c.pLits {
			if kind := p.pVarTypes[lit_var(l)]; kind < 0 {
				variable := aig.Aig_IthVar(p.pAig, -kind-1)
				if lit_sign(l) != 0 {
					variable = variable.Not()
				}
				value = p.pAig.OR(value, variable)
			}
		}
	}
	p.pInters[c.Id] = value
}
func (p *aigInterpolation) begin(final, conflict *Sto_Cls_t) {
	p.pInters[final.Id] = p.pInters[conflict.Id]
}
func (p *aigInterpolation) resolve(final, reason *Sto_Cls_t, pivot lit) {
	left, right := p.pInters[final.Id], p.pInters[reason.Id]
	kind := p.pVarTypes[lit_var(pivot)]
	switch {
	case kind == 1:
		left = p.pAig.OR(left, right)
	case kind < 0 && p.muxShared:
		variable := aig.Aig_IthVar(p.pAig, -kind-1)
		if lit_sign(pivot) != 0 {
			left = p.pAig.Mux(variable, right, left)
		} else {
			left = p.pAig.Mux(variable, left, right)
		}
	default:
		left = p.pAig.AND(left, right)
	}
	p.pInters[final.Id] = left
}
func (p *aigInterpolation) interpolate(cnf *Sto_Man_t, globals []int32, deadline int64, mux bool, verbose int32) *aig.Graph {
	return p.interpolateWithClock(cnf, globals, deadline, mux, verbose, nil)
}

func (p *aigInterpolation) interpolateWithClock(cnf *Sto_Man_t, globals []int32, deadline int64, mux bool, verbose int32, clock func() int64) *aig.Graph {
	p.pCnf, p.muxShared, p.fVerbose = cnf, mux, verbose
	p.pVarTypes, p.vVarsAB, p.lastErr = interpolationVariables(cnf, globals)
	if p.lastErr != nil {
		return nil
	}
	p.pAig = aig.NewGraph()
	for range p.vVarsAB {
		p.pAig.NewInput()
	}
	p.pInters = make([]aig.Edge, cnf.nClauses)
	var writer io.Writer
	var file *os.File
	if p.fProofWrite != 0 {
		file, p.lastErr = os.Create("proof.cnf_")
		if p.lastErr != nil {
			return nil
		}
		writer = file
	}
	trace, err := reconstructResolutionWithClock(cnf, p, writer, deadline, clock)
	if file != nil {
		if closeErr := file.Close(); err == nil {
			err = closeErr
		}
	}
	p.lastErr = err
	if trace != nil {
		p.Counter = trace.counter
	}
	if err != nil {
		p.pAig = nil
		return nil
	}
	p.pAig.AddOutput(p.pInters[cnf.pEmpty.Id])
	result := p.pAig
	p.pAig = nil
	if verbose != 0 {
		fmt.Printf("Vars = %d. Roots = %d. Learned = %d. Resol steps = %d.\n", cnf.nVars, cnf.nRoots, cnf.nClauses-cnf.nRoots, p.Counter)
	}
	return result
}
func (p *aigInterpolation) deriveClauses(cnf *Sto_Man_t, partA int32) *aig.Graph {
	graph := aig.NewGraph()

	product := graph.Const1()
	for c := cnf.pHead; c != nil; c = c.pNext {
		if c.fRoot == 0 || c.fA != uint32(partA) {
			continue
		}
		sum := graph.Const0()
		for _, l := range c.pLits {
			v := lit_var(l)
			index := int32(len(p.vVarsAB)) + 1 + v
			if kind := p.pVarTypes[v]; kind < 0 {
				index = -kind - 1
			}
			edge := aig.Aig_IthVar(graph, index)
			if lit_sign(l) != 0 {
				edge = edge.Not()
			}
			sum = graph.OR(sum, edge)
		}
		product = graph.AND(product, sum)
	}
	graph.AddOutput(product)
	return graph
}
