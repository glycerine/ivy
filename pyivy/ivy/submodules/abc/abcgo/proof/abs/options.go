package abs

// Options configures the GLA and VTA abstraction engines. Frame and conflict
// limits use the same units and zero-value conventions as Abs_Par_t in ABC.
// Obtain the source defaults with DefaultOptions before changing settings.
type Options struct {
	FramesMax, FramesStart, FramesPast       int32
	ConflictLimit                            int32
	LearnedMax, LearnedStart                 int32
	LearnedDelta, LearnedPercent             int32
	TimeoutSeconds                           int32
	RatioMin, RatioMinRefine, RatioMax       int32
	UseTerminalVariables, UseRollback        bool
	PropagateFanout, AddLayer, NewRefinement bool
	SkipIntermediate, SimpleCNF, SkipHash    bool
	// UseFullProof selects old GLA's full proof graph instead of its default
	// incremental proof manager.
	UseFullProof                                   bool
	DumpVariableAbstraction, DumpMappedAbstraction bool
	CallProver, SimplifyBeforeProver               bool
	VariableAbstractionFile                        string
	Verbose, VeryVerbose                           bool
	FramesNoChangeLimit                            int32
}

// Progress contains output fields updated by the abstraction engines.
type Progress struct {
	Frame, FrameProved, FramesNoChange int32
}

// DefaultOptions returns the exact defaults set by Abs_ParSetDefaults.
func DefaultOptions() Options {
	var p Abs_Par_t
	Abs_ParSetDefaults(&p)
	return p.Options()
}

// NewParameters creates an abstraction parameter record with ABC defaults.
func NewParameters() *Abs_Par_t {
	p := new(Abs_Par_t)
	Abs_ParSetDefaults(p)
	return p
}

func flag(value bool) int32 {
	if value {
		return 1
	}
	return 0
}

// Options reads the configuration without changing result fields.
func (p *Abs_Par_t) Options() Options {
	return Options{
		FramesMax: p.nFramesMax, FramesStart: p.nFramesStart, FramesPast: p.nFramesPast,
		ConflictLimit: p.nConfLimit, LearnedMax: p.nLearnedMax, LearnedStart: p.nLearnedStart,
		LearnedDelta: p.nLearnedDelta, LearnedPercent: p.nLearnedPerce, TimeoutSeconds: p.nTimeOut,
		RatioMin: p.nRatioMin, RatioMinRefine: p.nRatioMin2, RatioMax: p.nRatioMax,
		UseTerminalVariables: p.fUseTermVars != 0, UseRollback: p.fUseRollback != 0,
		PropagateFanout: p.fPropFanout != 0, AddLayer: p.fAddLayer != 0, NewRefinement: p.fNewRefine != 0,
		SkipIntermediate: p.fUseSkip != 0, SimpleCNF: p.fUseSimple != 0, SkipHash: p.fSkipHash != 0,
		UseFullProof: p.fUseFullProof != 0, DumpVariableAbstraction: p.fDumpVabs != 0,
		DumpMappedAbstraction: p.fDumpMabs != 0, CallProver: p.fCallProver != 0,
		SimplifyBeforeProver: p.fSimpProver != 0, VariableAbstractionFile: p.pFileVabs,
		Verbose: p.fVerbose != 0, VeryVerbose: p.fVeryVerbose != 0, FramesNoChangeLimit: p.nFramesNoChangeLim,
	}
}

// SetOptions replaces the configuration, preserving progress from prior runs.
func (p *Abs_Par_t) SetOptions(o Options) {
	p.nFramesMax, p.nFramesStart, p.nFramesPast = o.FramesMax, o.FramesStart, o.FramesPast
	p.nConfLimit, p.nLearnedMax, p.nLearnedStart = o.ConflictLimit, o.LearnedMax, o.LearnedStart
	p.nLearnedDelta, p.nLearnedPerce, p.nTimeOut = o.LearnedDelta, o.LearnedPercent, o.TimeoutSeconds
	p.nRatioMin, p.nRatioMin2, p.nRatioMax = o.RatioMin, o.RatioMinRefine, o.RatioMax
	p.fUseTermVars, p.fUseRollback = flag(o.UseTerminalVariables), flag(o.UseRollback)
	p.fPropFanout, p.fAddLayer, p.fNewRefine = flag(o.PropagateFanout), flag(o.AddLayer), flag(o.NewRefinement)
	p.fUseSkip, p.fUseSimple, p.fSkipHash = flag(o.SkipIntermediate), flag(o.SimpleCNF), flag(o.SkipHash)
	p.fUseFullProof, p.fDumpVabs, p.fDumpMabs = flag(o.UseFullProof), flag(o.DumpVariableAbstraction), flag(o.DumpMappedAbstraction)
	p.fCallProver, p.fSimpProver = flag(o.CallProver), flag(o.SimplifyBeforeProver)
	p.pFileVabs, p.fVerbose, p.fVeryVerbose = o.VariableAbstractionFile, flag(o.Verbose), flag(o.VeryVerbose)
	p.nFramesNoChangeLim = o.FramesNoChangeLimit
}

// Progress returns the most recent frame and stability counters.
func (p *Abs_Par_t) Progress() Progress {
	return Progress{Frame: p.iFrame, FrameProved: p.iFrameProved, FramesNoChange: p.nFramesNoChange}
}

// LegacyOptions configures the older abstraction/refinement driver.
type LegacyOptions struct {
	Algorithm                                        int32
	FramesMax, ConflictLimit                         int32
	Dynamic, Constraints                             bool
	BMCFrames, BMCConflictLimit, StableFrames, Ratio int32
	TimeoutSeconds, VTATimeoutSeconds, BobParameter  int32
	UseBDDs, UseDprove, UseStart                     bool
	Verbose, VeryVerbose                             bool
}

// LegacyProgress holds status fields updated by the legacy driver.
type LegacyProgress struct{ Status, FramesDone int32 }

// DefaultLegacyOptions returns the values from Gia_ManAbsSetDefaultParams.
func DefaultLegacyOptions() LegacyOptions {
	var p Gia_ParAbs_t
	Gia_ManAbsSetDefaultParams(&p)
	return p.Options()
}

// NewLegacyParameters creates a legacy parameter record with ABC defaults.
func NewLegacyParameters() *Gia_ParAbs_t {
	p := new(Gia_ParAbs_t)
	Gia_ManAbsSetDefaultParams(p)
	return p
}

// Options reads the legacy configuration without changing result fields.
func (p *Gia_ParAbs_t) Options() LegacyOptions {
	return LegacyOptions{
		Algorithm: p.Algo, FramesMax: p.nFramesMax, ConflictLimit: p.nConfMax,
		Dynamic: p.fDynamic != 0, Constraints: p.fConstr != 0,
		BMCFrames: p.nFramesBmc, BMCConflictLimit: p.nConfMaxBmc, StableFrames: p.nStableMax, Ratio: p.nRatio,
		TimeoutSeconds: p.TimeOut, VTATimeoutSeconds: p.TimeOutVT, BobParameter: p.nBobPar,
		UseBDDs: p.fUseBdds != 0, UseDprove: p.fUseDprove != 0, UseStart: p.fUseStart != 0,
		Verbose: p.fVerbose != 0, VeryVerbose: p.fVeryVerbose != 0,
	}
}

// SetOptions replaces the legacy configuration, preserving status counters.
func (p *Gia_ParAbs_t) SetOptions(o LegacyOptions) {
	p.Algo, p.nFramesMax, p.nConfMax = o.Algorithm, o.FramesMax, o.ConflictLimit
	p.fDynamic, p.fConstr = flag(o.Dynamic), flag(o.Constraints)
	p.nFramesBmc, p.nConfMaxBmc, p.nStableMax, p.nRatio = o.BMCFrames, o.BMCConflictLimit, o.StableFrames, o.Ratio
	p.TimeOut, p.TimeOutVT, p.nBobPar = o.TimeoutSeconds, o.VTATimeoutSeconds, o.BobParameter
	p.fUseBdds, p.fUseDprove, p.fUseStart = flag(o.UseBDDs), flag(o.UseDprove), flag(o.UseStart)
	p.fVerbose, p.fVeryVerbose = flag(o.Verbose), flag(o.VeryVerbose)
}

// Progress returns the legacy driver's most recent status and frame count.
func (p *Gia_ParAbs_t) Progress() LegacyProgress {
	return LegacyProgress{Status: p.Status, FramesDone: p.nFramesDone}
}
