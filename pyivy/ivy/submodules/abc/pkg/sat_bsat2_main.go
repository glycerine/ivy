package abc

import (
	"fmt"
	"math"
	"os"
	"syscall"
)

// Port of src/sat/bsat2/MainSat.cpp and MainSimp.cpp (namespace Minisat).
//
// The C++ main functions wrap their body in try { ... } catch (OutOfMemoryException&)
// that prints "INDETERMINATE" and exits; Go allocation failure is fatal and not
// recoverable, so that handler is not ported (see sat_bsat2_vec.go).

// solver_mainSat is `static Solver* solver` of MainSat.cpp.
var solver_mainSat *MinisatSolver

// PrintStats is printStats (not static; MainSimp.cpp declares it extern).
// C: void printStats(Solver& solver)
func PrintStats(solver *MinisatSolver) {
	cpuTime := MinisatCpuTime()
	memUsed := MinisatMemUsedPeak()
	fmt.Printf("restarts              : %-12.0f\n", float64(int64(solver.Starts)))
	fmt.Printf("conflicts             : %-12.0f  (%.0f /sec)\n", float64(int64(solver.Conflicts)), float64(int64(solver.Conflicts))/cpuTime)
	fmt.Printf("decisions             : %-12.0f  (%4.2f %% random) (%.0f /sec)\n", float64(int64(solver.Decisions)), float64(int64(solver.Rnd_decisions))*100/float64(int64(solver.Decisions)), float64(int64(solver.Decisions))/cpuTime)
	fmt.Printf("propagations          : %-12.0f  (%.0f /sec)\n", float64(int64(solver.Propagations)), float64(int64(solver.Propagations))/cpuTime)
	fmt.Printf("conflict literals     : %-12.0f  (%4.2f %% deleted)\n", float64(int64(solver.Tot_literals)), float64(int64(solver.Max_literals-solver.Tot_literals))*100/float64(int64(solver.Max_literals)))
	if memUsed != 0 {
		fmt.Printf("Memory used           : %.2f MB\n", memUsed)
	}
	fmt.Printf("CPU time              : %.6g s\n", cpuTime)
}

// sigintInterrupt_mainSat is `static void SIGINT_interrupt(int signum)`: terminate by notifying the solver.
// C: static void SIGINT_interrupt(int signum)
func sigintInterrupt_mainSat(signum int) { solver_mainSat.Interrupt() }

// sigintExit_mainSat is `static void SIGINT_exit(int signum)` ('_exit()' rather than 'exit()').
// C: static void SIGINT_exit(int signum)
func sigintExit_mainSat(signum int) {
	fmt.Printf("\n")
	fmt.Printf("*** INTERRUPTED ***\n")
	if solver_mainSat.Verbosity > 0 {
		PrintStats(solver_mainSat)
		fmt.Printf("\n")
		fmt.Printf("*** INTERRUPTED ***\n")
	}
	syscall.Exit(1)
}

// minisatLboolWord is the result word printed by both mains.
func minisatLboolWord(ret MinisatLbool) string {
	if ret == MinisatLTrue {
		return "SATISFIABLE\n"
	} else if ret == MinisatLFalse {
		return "UNSATISFIABLE\n"
	}
	return "INDETERMINATE\n"
}

// minisatWriteResult writes the result file of both mains and closes it.
func minisatWriteResult(res *os.File, s *MinisatSolver, ret MinisatLbool) {
	if ret == MinisatLTrue {
		fmt.Fprintf(res, "SAT\n")
		for i := 0; i < s.NVars(); i++ {
			if *s.Model.At(i) != MinisatLUndef {
				sep := " "
				if i == 0 {
					sep = ""
				}
				neg := "-"
				if *s.Model.At(i) == MinisatLTrue {
					neg = ""
				}
				fmt.Fprintf(res, "%s%s%d", sep, neg, i+1)
			}
		}
		fmt.Fprintf(res, " 0\n")
	} else if ret == MinisatLFalse {
		fmt.Fprintf(res, "UNSAT\n")
	} else {
		fmt.Fprintf(res, "INDET\n")
	}
	res.Close()
}

// minisatOpenInput is `gzFile in = gzopen(argv[1], "rb")` plus its error exit.
func minisatOpenInput(path string) *GZIPReader {
	in, err := OpenGZIPReader(path, GZIPReaderOptions{})
	if err != nil || in == nil {
		fmt.Printf("ERROR! Could not open file: %s\n", path)
		os.Exit(1)
	}
	return in
}

// MainSat is `extern "C" int MainSat(int argc, char** argv)`.
// C: extern "C" int MainSat(int argc, char** argv)
func MainSat(argc int, argv []string) int {
	MinisatSetUsageHelp("USAGE: %s [options] <input-file> <result-output-file>\n\n  where input may be either in plain or gzipped DIMACS.\n")
	// The C code sets the x87 FPU to double precision here (a no-op for Go's float64 arithmetic).
	fmt.Printf("WARNING: for repeatability, setting FPU to use double precision\n")

	// Extra options:
	verb := NewMinisatIntOption("MAIN", "verb", "Verbosity level (0=silent, 1=some, 2=more).", 1, NewMinisatIntRange(0, 2))
	_ = NewMinisatIntOption("MAIN", "cpu-lim", "Limit on CPU time allowed in seconds.\n", math.MaxInt32, NewMinisatIntRange(0, math.MaxInt32))
	_ = NewMinisatIntOption("MAIN", "mem-lim", "Limit on memory usage in megabytes.\n", math.MaxInt32, NewMinisatIntRange(0, math.MaxInt32))

	if MinisatParseOptions(&argc, argv, true) == 0 {
		return 1
	}

	S := NewMinisatSolver()
	initialTime := MinisatCpuTime()

	S.Verbosity = int(verb.Value())

	solver_mainSat = S
	// (The signal handlers and resource limits are commented out in the C source.)
	if argc == 1 {
		fmt.Printf("Reading from standard input... Use '--help' for help.\n")
		return 1
	}

	in := minisatOpenInput(argv[1])

	if S.Verbosity > 0 {
		fmt.Printf("============================[ Problem Statistics ]=============================\n")
		fmt.Printf("|                                                                             |\n")
	}

	minisatParseDIMACS(in, S)
	in.Close()
	var res *os.File
	if argc >= 3 {
		res, _ = os.Create(argv[2]) // fopen(argv[2], "wb")
	}

	if S.Verbosity > 0 {
		fmt.Printf("|  Number of variables:  %12d                                         |\n", S.NVars())
		fmt.Printf("|  Number of clauses:    %12d                                         |\n", S.NClauses())
	}

	parsedTime := MinisatCpuTime()
	if S.Verbosity > 0 {
		fmt.Printf("|  Parse time:           %12.2f s                                       |\n", parsedTime-initialTime)
		fmt.Printf("|                                                                             |\n")
	}

	if !S.Simplify() {
		if res != nil {
			fmt.Fprintf(res, "UNSAT\n")
			res.Close()
		}
		if S.Verbosity > 0 {
			fmt.Printf("===============================================================================\n")
			fmt.Printf("Solved by unit propagation\n")
			PrintStats(S)
			fmt.Printf("\n")
		}
		fmt.Printf("UNSATISFIABLE\n")
		os.Exit(20)
	}

	var dummy MinisatVec[MinisatLit]
	ret := S.SolveLimited(&dummy)
	if S.Verbosity > 0 {
		PrintStats(S)
		fmt.Printf("\n")
	}
	fmt.Print(minisatLboolWord(ret))
	if res != nil {
		minisatWriteResult(res, S, ret)
	}

	if ret == MinisatLTrue {
		return 10
	} else if ret == MinisatLFalse {
		return 20
	}
	return 0
}

// solver_mainSimp is `static Solver* solver` of MainSimp.cpp (Go cannot assign a
// SimpSolver to a Solver pointer, so the base object is kept).
var solver_mainSimp *MinisatSolver

// sigintInterrupt_mainSimp is `static void SIGINT_interrupt(int signum)`.
// C: static void SIGINT_interrupt(int signum)
func sigintInterrupt_mainSimp(signum int) { solver_mainSimp.Interrupt() }

// sigintExit_mainSimp is `static void SIGINT_exit(int signum)`.
// C: static void SIGINT_exit(int signum)
func sigintExit_mainSimp(signum int) {
	fmt.Printf("\n")
	fmt.Printf("*** INTERRUPTED ***\n")
	if solver_mainSimp.Verbosity > 0 {
		PrintStats(solver_mainSimp)
		fmt.Printf("\n")
		fmt.Printf("*** INTERRUPTED ***\n")
	}
	syscall.Exit(1)
}

// MainSimp is `extern "C" int MainSimp(int argc, char** argv)`.
// C: extern "C" int MainSimp(int argc, char** argv)
func MainSimp(argc int, argv []string) int {
	MinisatSetUsageHelp("USAGE: %s [options] <input-file> <result-output-file>\n\n  where input may be either in plain or gzipped DIMACS.\n")
	fmt.Printf("WARNING: for repeatability, setting FPU to use double precision\n")

	// Extra options:
	verb := NewMinisatIntOption("MAIN", "verb", "Verbosity level (0=silent, 1=some, 2=more).", 1, NewMinisatIntRange(0, 2))
	pre := NewMinisatBoolOption("MAIN", "pre", "Completely turn on/off any preprocessing.", true)
	dimacs := NewMinisatStringOption("MAIN", "dimacs", "If given, stop after preprocessing and write the result to this file.", "")
	_ = NewMinisatIntOption("MAIN", "cpu-lim", "Limit on CPU time allowed in seconds.\n", math.MaxInt32, NewMinisatIntRange(0, math.MaxInt32))
	_ = NewMinisatIntOption("MAIN", "mem-lim", "Limit on memory usage in megabytes.\n", math.MaxInt32, NewMinisatIntRange(0, math.MaxInt32))

	if MinisatParseOptions(&argc, argv, true) == 0 {
		return 1
	}

	S := NewMinisatSimpSolver()
	initialTime := MinisatCpuTime()

	if !pre.Value() {
		S.Eliminate(true)
	}

	S.Verbosity = int(verb.Value())

	solver_mainSimp = &S.MinisatSolver
	if argc == 1 {
		fmt.Printf("Reading from standard input... Use '--help' for help.\n")
		return 1
	}

	in := minisatOpenInput(argv[1])

	if S.Verbosity > 0 {
		fmt.Printf("============================[ Problem Statistics ]=============================\n")
		fmt.Printf("|                                                                             |\n")
	}

	minisatParseDIMACS(in, S)
	in.Close()
	var res *os.File
	if argc >= 3 {
		res, _ = os.Create(argv[2])
	}

	if S.Verbosity > 0 {
		fmt.Printf("|  Number of variables:  %12d                                         |\n", S.NVars())
		fmt.Printf("|  Number of clauses:    %12d                                         |\n", S.NClauses())
	}

	parsedTime := MinisatCpuTime()
	if S.Verbosity > 0 {
		fmt.Printf("|  Parse time:           %12.2f s                                       |\n", parsedTime-initialTime)
	}

	S.Eliminate(true)
	simplifiedTime := MinisatCpuTime()
	if S.Verbosity > 0 {
		fmt.Printf("|  Simplification time:  %12.2f s                                       |\n", simplifiedTime-parsedTime)
		fmt.Printf("|                                                                             |\n")
	}

	if !S.Okay() {
		if res != nil {
			fmt.Fprintf(res, "UNSAT\n")
			res.Close()
		}
		if S.Verbosity > 0 {
			fmt.Printf("===============================================================================\n")
			fmt.Printf("Solved by simplification\n")
			PrintStats(&S.MinisatSolver)
			fmt.Printf("\n")
		}
		fmt.Printf("UNSATISFIABLE\n")
		os.Exit(20)
	}

	if dimacs.Value() != "" {
		if S.Verbosity > 0 {
			fmt.Printf("==============================[ Writing DIMACS ]===============================\n")
		}
		S.ToDimacsFileNoAssumps(dimacs.Value()) // SimpSolver::toDimacs is only declared in C++; the Solver one runs
		if S.Verbosity > 0 {
			PrintStats(&S.MinisatSolver)
		}
		os.Exit(0)
	}

	var dummy MinisatVec[MinisatLit]
	ret := S.SolveLimited(&dummy, true, false)

	if S.Verbosity > 0 {
		PrintStats(&S.MinisatSolver)
		fmt.Printf("\n")
	}
	fmt.Print(minisatLboolWord(ret))
	if res != nil {
		minisatWriteResult(res, &S.MinisatSolver, ret)
	}

	if ret == MinisatLTrue {
		return 10
	} else if ret == MinisatLFalse {
		return 20
	}
	return 0
}
