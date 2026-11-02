package abc

// Hand-checked skeleton of src/sat/glucose/System.cpp (namespace Gluco).
//
// Gluco::memUsed and Gluco::memUsedPeak are defined here (one definition per platform); their stubs
// are GlucoMemUsed and GlucoMemUsedPeak in skel_sat_glucose_System_h.go.

// memReadStat_glucose is `static inline int memReadStat(int field)` (reads /proc/<pid>/statm).
// C: static inline int memReadStat(int field)
func memReadStat_glucose(field int) int {
	panic("TODO, stubbed")
}

// memReadPeak_glucose is `static inline int memReadPeak(void)` (reads VmPeak from /proc/<pid>/status).
// C: static inline int memReadPeak(void)
func memReadPeak_glucose() int {
	panic("TODO, stubbed")
}
