package abc

// Hand-checked skeleton of src/sat/bsat2/System.h.

// All three live in namespace Minisat.

// Minisat::cpuTime (static inline; CPU-time in seconds)
// C: double Minisat::cpuTime(void)
func MinisatCpuTime() float64 {
	panic("TODO, stubbed")
}

// Minisat::memUsed (extern, defined in System.cpp; memory in megabytes (0 for unsupported architectures))
// C: double Minisat::memUsed()
func MinisatMemUsed() float64 {
	panic("TODO, stubbed")
}

// Minisat::memUsedPeak (extern, defined in System.cpp; peak memory in megabytes)
// C: double Minisat::memUsedPeak()
func MinisatMemUsedPeak() float64 {
	panic("TODO, stubbed")
}
