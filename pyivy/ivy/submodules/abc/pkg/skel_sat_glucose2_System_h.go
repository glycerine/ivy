package abc

// Hand-checked skeleton of src/sat/glucose2/System.h.

// All three live in namespace Gluco2.

// Gluco2::cpuTime (static inline; CPU-time in seconds)
// C: double Gluco2::cpuTime(void)
func Gluco2CpuTime() float64 {
	panic("TODO, stubbed")
}

// Gluco2::memUsed (extern, defined in System2.cpp; memory in megabytes (0 for unsupported architectures))
// C: double Gluco2::memUsed()
func Gluco2MemUsed() float64 {
	panic("TODO, stubbed")
}

// Gluco2::memUsedPeak (extern, defined in System2.cpp; peak memory in megabytes)
// C: double Gluco2::memUsedPeak()
func Gluco2MemUsedPeak() float64 {
	panic("TODO, stubbed")
}
