package aigpartsat

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/opt/dar"
	"github.com/glycerine/abcgo/sat/bsat"
)

type darSynthesizer struct{}

func (darSynthesizer) SynthesizePart(g *aig.Graph) (*aig.Graph, error) {
	return dar.Dar_ManRwsat(g, false, false)
}
func outputDriver(g *aig.Graph, e aig.Edge) aig.Edge {
	for g.Node(e).Kind == aig.Output {
		child := g.Node(e).Fanins[0]
		if e.Complemented() {
			child = child.Not()
		}
		e = child
	}
	return e
}

// PartSATSynthesizer permits an explicit rewrite pass in the modern API. The
// source entry point uses DAR. Terminal order and Boolean functions must be preserved.
type PartSATSynthesizer interface {
	SynthesizePart(*aig.Graph) (*aig.Graph, error)
}
type PartSATOptions struct {
	Algorithm                             int32
	PartitionSize                         int32
	PartConflictLimit, TotalConflictLimit int64
	AlignPolarity                         bool
	Synthesize                            bool
	Synthesizer                           PartSATSynthesizer
	Verbose                               bool
}

// PartitionedSAT uses the source partition/mapped-CNF/BSAT workflow. It returns
// 0 for SAT, 1 for UNSAT and -1 for an incomplete solve. Zero budgets mean
// unlimited. Unlike the C driver, only newly spent conflicts are subtracted,
// and exhausting the total budget before the last partition cannot report SAT.
func PartitionedSAT(g *aig.Graph, options PartSATOptions) (int32, error) {
	if g == nil {
		return -1, errors.New("aig: nil SAT graph")
	}
	if err := g.Check(); err != nil {
		return -1, err
	}
	if options.Synthesize && options.Synthesizer == nil {
		return -1, errors.New("aig: partition synthesis requested without a PartSATSynthesizer")
	}
	if options.PartConflictLimit < 0 || options.TotalConflictLimit < 0 {
		return -1, errors.New("aig: negative SAT conflict limit")
	}
	var labels []int32
	switch options.Algorithm {
	case 0:
		labels = aig.Aig_ManPartitionMonolithic(g)
	case 1, 2, 3:
		if options.PartitionSize <= 0 {
			return -1, errors.New("aig: partition size must be positive")
		}
		if options.Algorithm == 1 {
			labels = aig.Aig_ManPartitionLevelized(g, options.PartitionSize)
		} else {
			pre := int32(0)
			if options.Algorithm == 3 {
				pre = 1
			}
			labels = aig.Aig_ManPartitionDfs(g, options.PartitionSize, pre)
		}
	default:
		return -1, errors.New("aig: unknown partitioning algorithm")
	}
	started := time.Now()
	parts, pio, outputs := aig.Aig_ManPartSplit(g, labels)
	for i, part := range parts {
		var err error
		parts[i], err = normalizePart(part)
		if err != nil {
			return -1, err
		}
	}
	if options.Synthesize {
		for i, part := range parts {
			inputs, outputs := part.InputCount(), part.OutputCount()
			rewritten, err := options.Synthesizer.SynthesizePart(part)
			if err != nil {
				return -1, err
			}
			if rewritten == nil || rewritten.InputCount() != inputs || rewritten.OutputCount() != outputs {
				return -1, errors.New("aig: synthesis changed partition terminal interface")
			}
			if err := rewritten.Check(); err != nil {
				return -1, err
			}
			parts[i], err = normalizePart(rewritten)
			if err != nil {
				return -1, err
			}
		}
	}
	solver := bsat.NewSolver()
	defer solver.Close()
	variables := NewNodeVariables(int(aig.Aig_ManObjNumMax(g)))
	// C only constrains PI outputs through marked partition CIs. An isolated
	// PI has no such terminal; keep the existing explicit constraint repair.
	for _, output := range g.Outputs() {
		driver := outputDriver(g, output)
		if driver.NodeID() == 0 {
			if driver.Complemented() {
				return 1, nil
			}
			continue
		}
		if g.Node(driver).Kind != aig.Input {
			continue
		}
		id := driver.NodeID()
		if variables.IDs[id] == 0 {
			variables.IDs[id] = newPartVariable(solver)
			variables.Phases[id] = options.AlignPolarity && g.Node(driver).Phase
		}
		lit := sharedLiteral(variables, id)
		if driver.Complemented() {
			lit ^= 1
		}
		if !solver.AddClause(lit) {
			return 1, nil
		}
	}
	remaining := options.TotalConflictLimit
	finalStatus := int32(-1)
	for i, part := range parts {
		if options.AlignPolarity {
			aig.Aig_ManPartSetNodePolarity(g, part, pio[i])
		} else {
			aig.Aig_ManPartResetNodePolarity(part)
		}
		if Aig_ManAddNewCnfToSolver(solver, part, variables, pio[i], outputs[i], 0) != 0 {
			return 1, nil
		}
		budget := options.PartConflictLimit
		if options.TotalConflictLimit > 0 && (budget == 0 || remaining < budget) {
			budget = remaining
		}
		before := solver.Statistics().Conflicts
		status := solver.SolveWithResourceLimits(nil, budget, 0, 0, 0)
		used := solver.Statistics().Conflicts - before
		if options.Verbose {
			fmt.Printf("Partition %d/%d: vars=%d conflicts=%d elapsed=%s\n", i+1, len(parts), solver.NumVars(), used, time.Since(started))
		}
		if status == bsat.Unsatisfiable {
			return 1, nil
		}
		if status == bsat.Satisfiable {
			finalStatus = 0
		} else {
			finalStatus = -1
		}
		if options.TotalConflictLimit > 0 {
			remaining -= used
			if remaining <= 0 && i+1 < len(parts) {
				return -1, nil
			}
		}
	}
	if finalStatus == 0 {
		Aig_ManDeriveCounterExample(g, variables, solver)
	}
	return finalStatus, nil
}
func Aig_ManPartitionedSat(g *aig.Graph, algorithm, size, partLimit, totalLimit, align, synthesize, verbose int32) int32 {
	// The C entry point never uses nConfPart. Modern PartSATOptions can set a
	// per-call cap, but this source-compatible wrapper keeps that argument inert.
	options := PartSATOptions{Algorithm: algorithm, PartitionSize: size, TotalConflictLimit: int64(totalLimit), AlignPolarity: align != 0, Synthesize: synthesize != 0, Verbose: verbose != 0}
	if options.Synthesize {
		options.Synthesizer = darSynthesizer{}
	}
	status, err := PartitionedSAT(g, options)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	return status
}
