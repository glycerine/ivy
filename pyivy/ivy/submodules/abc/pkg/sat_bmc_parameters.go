package abc

// Public parameter records and defaults from src/sat/bmc/bmc.h. These values
// are concrete command/configuration data; the solving APIs use the narrower
// operation-specific option records declared beside their implementations.

// BMCExactSynthesisParameters mirrors Bmc_EsPar_t. The exact-synthesis engines
// that consume it are tracked separately with bmcMaj.c, bmcMaj2.c and
// bmcMaj3.c.
type BMCExactSynthesisParameters struct {
	Variables, Nodes, LUTSize, MajoritySupport int
	Majority, Incremental, OnlyAND             bool
	UseGlucose, OrderNodes, EnumerateSolutions bool
	FewerVariables, Verbose                    bool
	RuntimeLimit                               int
	Truth                                      string
}

func DefaultBMCExactSynthesisParameters() BMCExactSynthesisParameters {
	return BMCExactSynthesisParameters{LUTSize: 2, Verbose: true}
}

// AigScalableBMCParameters mirrors Saig_ParBmc_t, including its result fields.
// Logical SAT budgets in AigScalableBMCOptions replace the native wall-clock
// and backend-selection controls when an actual solve is started.
type AigScalableBMCParameters struct {
	StartFrame, Frames, ConflictLimit, JumpConflictLimit int
	FramesJump, Timeout, GapTimeout, OutputTimeout       int
	AbstractPrimaryInputs, MaxAddedFlops                 int
	SolveAll, StoreCounterexamples, UseBridge            bool
	DropSATOutputs, SkipRandom, NoRestarts               bool
	UseSatoko, UseGlucose                                bool
	LearnedStart, LearnedDelta, LearnedPercent           int
	Verbose, NotVerbose, Silent                          bool
	LogFile                                              string
	ExploredFrame, FailedOutputs, DroppedOutputs         int
	LastSolvedTicks, RunID                               int64
	OnFailure                                            func(property int, witness *Counterexample) bool
	Stop                                                 func(runID int) bool
}

func DefaultAigScalableBMCParameters() AigScalableBMCParameters {
	return AigScalableBMCParameters{
		LearnedStart:   10000,
		LearnedDelta:   2000,
		LearnedPercent: 80,
		ExploredFrame:  -1,
	}
}

// GiaAndBMCParameters mirrors Bmc_AndPar_t. ABC has two command defaults for
// this shared record, represented by the two constructors below.
type GiaAndBMCParameters struct {
	StartFrame, Frames, AdditionalFrames, ConflictLimit int
	Timeout, LUTSize, Workers                           int
	LoadCNF, DumpFrames, UseSynthesis, UseOldCNF        bool
	UseGlucose, UseElimination                          bool
	Verbose, VeryVerbose, NotVerbose                    bool
	ExploredFrame, FailedOutputs, DroppedOutputs        int
	OnFrameDone                                         func(frame, property int, status BmcStatus)
}

// DefaultGiaAndBMCParameters returns the abc9bmc defaults.
func DefaultGiaAndBMCParameters() GiaAndBMCParameters {
	return GiaAndBMCParameters{AdditionalFrames: 50, LUTSize: 6, Workers: 1}
}

// DefaultGiaSuperBMCParameters returns the abc9bmc3/abc9sbmc defaults.
func DefaultGiaSuperBMCParameters() GiaAndBMCParameters {
	return GiaAndBMCParameters{AdditionalFrames: 1, Workers: 1}
}

// GiaBoundedCoreParameters mirrors Bmc_BCorePar_t and the abc9bcore command
// defaults. File paths are retained as data; the checked Go operations accept
// parsed pivots and return proof data, with explicit read/write helpers.
type GiaBoundedCoreParameters struct {
	Frame, Output, Timeout int
	PivotFile, ProofFile   string
	Verbose                bool
}

func DefaultGiaBoundedCoreParameters() GiaBoundedCoreParameters {
	return GiaBoundedCoreParameters{Frame: 10}
}

// MultiProveParameters mirrors Bmc_MulPar_t and the abc9mult command defaults.
// The Go prover replaces time-based stopping with deterministic work limits,
// but preserving this record keeps command parsing and source defaults exact.
type MultiProveParameters struct {
	GlobalTimeout, LocalTimeout, TimeoutIncrease int
	GapTimeout, PerOutputTimeout                 int
	UseSynthesis, DumpFinal                      bool
	Verbose, VeryVerbose                         bool
}

func DefaultMultiProveParameters() MultiProveParameters {
	return MultiProveParameters{GlobalTimeout: 30, LocalTimeout: 2, TimeoutIncrease: 100}
}

// GiaFaultParameters mirrors Bmc_ParFf_t for the owned fault campaign. Timeout
// is retained as source metadata; deterministic SatLimits control Go work.
// FileName is an explicit Go pattern-input path, while dump options return
// encoded data and structured assignments instead of writing ambient files.
type GiaFaultParameters struct {
	FileName, Formula                            string
	Algorithm, Timeout, CheckIterations          int
	CardinalityConstraint                        int
	ComplementVariables, StartPatterns           bool
	NonStrict, Basic, FlopsOnly, CheckUntestable bool
	Dump, DumpDelay, DumpUntestable              bool
	DumpNewFaults, Verbose                       bool
}

func DefaultGiaFaultParameters() GiaFaultParameters { return GiaFaultParameters{} }
