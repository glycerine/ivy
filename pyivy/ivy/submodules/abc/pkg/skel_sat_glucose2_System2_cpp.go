package abc

// Hand-checked skeleton of src/sat/glucose2/System2.cpp (namespace Gluco2).
//
// Gluco2::memUsed and Gluco2::memUsedPeak are defined here (one definition per platform); their stubs
// are Gluco2MemUsed and Gluco2MemUsedPeak in skel_sat_glucose2_System_h.go.

// memReadStat_glucose2 is `static inline int memReadStat(int field)` (reads /proc/<pid>/statm).
// C: static inline int memReadStat(int field)
func memReadStat_glucose2(field int) int {
	panic("TODO, stubbed")
}

// memReadPeak_glucose2 is `static inline int memReadPeak(void)` (reads VmPeak from /proc/<pid>/status).
// C: static inline int memReadPeak(void)
func memReadPeak_glucose2() int {
	panic("TODO, stubbed")
}
