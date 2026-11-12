package bsat2

import (
	"flag"
	"fmt"
	"io"
	"math"
	"os"
)

/*****************************************************************************************[Main.cc]
Copyright (c) 2003-2006, Niklas Een, Niklas Sorensson
Copyright (c) 2007-2010, Niklas Sorensson

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and
associated documentation files (the "Software"), to deal in the Software without restriction,
including without limitation the rights to use, copy, modify, merge, publish, distribute,
sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or
substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT
NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM,
DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT
OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
**************************************************************************************************/

// RunSAT provides the original command-line workflow as an embeddable library
// entry point. SAT/UNSAT exit codes are 10/20; malformed input returns an error.
func RunSAT(arguments []string, input io.Reader, output, diagnostic io.Writer) (int, error) {
	return runCommand(arguments, input, output, diagnostic, false)
}
func runCommand(arguments []string, input io.Reader, output, diagnostic io.Writer, simplify bool) (int, error) {
	options := NewOptions("minisat", diagnostic)
	verbosity := options.Int("verb", 1, 0, 2, "verbosity")
	varDecay := options.Float64("var-decay", 0.95, 0, 1, "variable activity decay")
	clauseDecay := options.Float64("cla-decay", 0.999, 0, 1, "clause activity decay")
	randomFrequency := options.Float64("rnd-freq", 0, 0, 1, "random decision frequency")
	randomSeed := options.Float64("rnd-seed", 91648253, 0, math.MaxFloat64, "random seed")
	restartFirst := options.Int("rfirst", 100, 1, math.MaxInt32, "first restart conflict limit")
	restartInc := options.Float64("rinc", 2, 1, math.MaxFloat64, "restart growth")
	minimization := options.Int("ccmin-mode", 2, 0, 2, "clause minimization mode")
	phaseSaving := options.Int("phase-saving", 2, 0, 2, "phase saving mode")
	lubyRestart := options.Bool("luby", true, "Luby restarts")
	randomInitial := options.Bool("rnd-init", false, "random initial variable activities")
	garbageFraction := options.Float64("gc-frac", 0.2, 0, math.MaxFloat64, "garbage collection threshold")
	simpGarbageFraction := options.Float64("simp-gc-frac", 0.5, 0, math.MaxFloat64, "preprocessing garbage collection threshold")
	pre := options.Bool("pre", true, "run preprocessing")
	asymm := options.Bool("asymm", false, "asymmetric branching")
	elimination := options.Bool("elim", true, "variable elimination")
	redundancy := options.Bool("rcheck", false, "redundancy checking")
	grow := options.Int("grow", 0, math.MinInt32, math.MaxInt32, "allowed elimination clause growth")
	clauseLimit := options.Int("cl-lim", 20, -1, math.MaxInt32, "elimination resolvent length limit")
	subsumptionLimit := options.Int("sub-lim", 1000, -1, math.MaxInt32, "subsumption length limit")
	dimacs := options.String("dimacs", "", "write preprocessed DIMACS instead of solving")
	positional, err := options.Parse(arguments)
	if err != nil {
		if err == flag.ErrHelp {
			return 0, nil
		}
		return 1, err
	}
	if *varDecay <= 0 || *varDecay >= 1 || *clauseDecay <= 0 || *clauseDecay >= 1 || *randomSeed <= 0 || *restartInc <= 1 || *garbageFraction <= 0 || *simpGarbageFraction <= 0 {
		return 1, fmt.Errorf("invalid open-bound solver option")
	}
	if len(positional) > 2 {
		return 1, fmt.Errorf("expected input and optional result filename")
	}
	var inputFile *os.File
	if len(positional) > 0 {
		inputFile, err = os.Open(positional[0])
		if err != nil {
			return 1, err
		}
		defer inputFile.Close()
		input = inputFile
	}
	if input == nil {
		return 1, fmt.Errorf("missing input")
	}
	core := NewSolver()
	var simp *SimpSolver
	var sink ClauseSink = core
	if simplify {
		simp = NewSimpSolver()
		core = simp.Solver
		sink = simp
		simp.UseAsymmetric = *asymm
		simp.UseElimination = *elimination
		simp.UseRedundancyCheck = *redundancy
		simp.Grow = *grow
		simp.ClauseLimit = *clauseLimit
		simp.SubsumptionLimit = *subsumptionLimit
		simp.SimplificationGarbageFraction = *simpGarbageFraction
	}
	core.Verbosity = *verbosity
	core.VarDecay = *varDecay
	core.ClauseDecay = *clauseDecay
	core.RandomFrequency = *randomFrequency
	core.RandomSeed = *randomSeed
	core.RestartFirst = *restartFirst
	core.RestartIncrease = *restartInc
	core.Minimization = *minimization
	core.PhaseSaving = *phaseSaving
	core.LubyRestart = *lubyRestart
	core.RandomInitialActivity = *randomInitial
	core.GarbageFraction = *garbageFraction
	if err := ParseDIMACS(input, sink); err != nil {
		return 1, err
	}
	if *dimacs != "" {
		if simp != nil && *pre {
			simp.Eliminate(true)
		}
		file, err := os.Create(*dimacs)
		if err != nil {
			return 1, err
		}
		err = core.WriteDIMACS(file)
		if closeErr := file.Close(); err == nil {
			err = closeErr
		}
		return 0, err
	}
	status := Undef
	if simp != nil {
		simp.BudgetOff()
		status = simp.SolveSimplified(nil, *pre, false)
	} else {
		status = core.Solve()
	}
	message, code := "INDETERMINATE", 0
	if status == True {
		message, code = "SATISFIABLE", 10
	} else if status == False {
		message, code = "UNSATISFIABLE", 20
	}
	if _, err := fmt.Fprintln(output, message); err != nil {
		return 1, err
	}
	if len(positional) == 2 {
		result, err := os.Create(positional[1])
		if err != nil {
			return 1, err
		}
		label := "INDET"
		if status == True {
			label = "SAT"
		} else if status == False {
			label = "UNSAT"
		}
		_, err = fmt.Fprintln(result, label)
		if err == nil && status == True {
			for v, value := range core.Model {
				if value == Undef {
					continue
				}
				literal := v + 1
				if value == False {
					literal = -literal
				}
				if _, err = fmt.Fprintf(result, "%d ", literal); err != nil {
					break
				}
			}
			if err == nil {
				_, err = fmt.Fprintln(result, "0")
			}
		}
		if closeErr := result.Close(); err == nil {
			err = closeErr
		}
		if err != nil {
			return 1, err
		}
	}
	if *verbosity > 0 {
		fmt.Fprintf(diagnostic, "conflicts: %d; decisions: %d; propagations: %d\n", core.Conflicts, core.Decisions, core.Propagations)
	}
	return code, nil
}
