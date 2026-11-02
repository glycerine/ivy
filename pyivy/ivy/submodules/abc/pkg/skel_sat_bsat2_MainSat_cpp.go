package abc

// Hand-checked skeleton of src/sat/bsat2/MainSat.cpp (namespace Minisat).

// solver_mainSat is `static Solver* solver`.
var solver_mainSat *MinisatSolver

// PrintStats is printStats (not static; MainSimp.cpp declares it extern).
// C: void printStats(Solver& solver)
func PrintStats(solver *MinisatSolver) {
	panic("TODO, stubbed")
}

// sigintInterrupt_mainSat is `static void SIGINT_interrupt(int signum)`.
// C: static void SIGINT_interrupt(int signum)
func sigintInterrupt_mainSat(signum int) {
	panic("TODO, stubbed")
}

// sigintExit_mainSat is `static void SIGINT_exit(int signum)`.
// C: static void SIGINT_exit(int signum)
func sigintExit_mainSat(signum int) {
	panic("TODO, stubbed")
}

// MainSat is `extern "C" int MainSat(int argc, char** argv)`.
// C: extern "C" int MainSat(int argc, char** argv)
func MainSat(argc int, argv []string) int {
	panic("TODO, stubbed")
}
