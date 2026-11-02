package abc

// Hand-checked skeleton of src/sat/bsat2/MainSimp.cpp (namespace Minisat).
// `extern void printStats(Solver& solver)` is PrintStats in skel_sat_bsat2_MainSat_cpp.go.

// solver_mainSimp is `static Solver* solver`.
var solver_mainSimp *MinisatSolver

// sigintInterrupt_mainSimp is `static void SIGINT_interrupt(int signum)`.
// C: static void SIGINT_interrupt(int signum)
func sigintInterrupt_mainSimp(signum int) {
	panic("TODO, stubbed")
}

// sigintExit_mainSimp is `static void SIGINT_exit(int signum)`.
// C: static void SIGINT_exit(int signum)
func sigintExit_mainSimp(signum int) {
	panic("TODO, stubbed")
}

// MainSimp is `extern "C" int MainSimp(int argc, char** argv)`.
// C: extern "C" int MainSimp(int argc, char** argv)
func MainSimp(argc int, argv []string) int {
	panic("TODO, stubbed")
}
