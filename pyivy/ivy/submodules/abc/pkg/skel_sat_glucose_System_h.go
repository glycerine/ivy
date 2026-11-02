package abc

// Hand-checked skeleton of src/sat/glucose/System.h.

// All three live in namespace Gluco.

// Gluco::cpuTime (static inline; CPU-time in seconds)
// C: double Gluco::cpuTime(void)
func GlucoCpuTime() float64 {
	panic("TODO, stubbed")
}

// Gluco::memUsed (extern, defined in System.cpp; memory in megabytes (0 for unsupported architectures))
// C: double Gluco::memUsed()
func GlucoMemUsed() float64 {
	panic("TODO, stubbed")
}

// Gluco::memUsedPeak (extern, defined in System.cpp; peak memory in megabytes)
// C: double Gluco::memUsedPeak()
func GlucoMemUsedPeak() float64 {
	panic("TODO, stubbed")
}
