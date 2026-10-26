package abc

// Exact active behavior from src/sat/bsat2/AbcApi.cpp.

// ABCMiniSat22Unavailable is the only result returned by the native
// Abc_CallMiniSat22 stub. The translation unit constructs an otherwise unused
// solver, does not inspect its CNF argument, and returns -1 unconditionally.
const ABCMiniSat22Unavailable = -1

// ABCCallMiniSat22 preserves Abc_CallMiniSat22 exactly. In particular, it does
// not reinterpret this dormant API as a call to another package SAT engine.
// Nil is accepted because the native function never dereferences its argument.
func ABCCallMiniSat22(_ *CNFFormula) int { return ABCMiniSat22Unavailable }
