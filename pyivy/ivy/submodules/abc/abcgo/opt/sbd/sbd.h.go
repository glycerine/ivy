package sbd

// Parameters configures SAT-based delay optimization of a mapped native graph.
// LUTSize supports 2..6 inputs; cut support is bounded by ten object IDs.
type Parameters struct {
	LUTSize, LUTCount, CutSize, CutCount                                                 int32
	TFOLevels, TFOFanoutMax, WindowSizeMax, ConflictLimit, SimulationWords               int32
	Mapping, MoreCuts, FindDivisors, CriticalPathOnly, Area, Cover, Verbose, VeryVerbose bool
}
type Sbd_Par_t = Parameters
