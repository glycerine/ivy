package abc

// Hand-checked skeleton of src/sat/bsat2/System.cpp (namespace Minisat).
//
// Minisat::memUsed and Minisat::memUsedPeak are defined here (one definition per
// platform: Linux, FreeBSD, Apple, other); their stubs are MinisatMemUsed and
// MinisatMemUsedPeak in skel_sat_bsat2_System_h.go.

// memReadStat_bsat2 is `static inline int memReadStat(int field)` (reads /proc/<pid>/statm).
// C: static inline int memReadStat(int field)
func memReadStat_bsat2(field int) int {
	panic("TODO, stubbed")
}

// memReadPeak_bsat2 is `static inline int memReadPeak(void)` (reads VmPeak from /proc/<pid>/status).
// C: static inline int memReadPeak(void)
func memReadPeak_bsat2() int {
	panic("TODO, stubbed")
}
