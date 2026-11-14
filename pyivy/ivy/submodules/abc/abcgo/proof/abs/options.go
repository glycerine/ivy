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
		FramesMax: p.NFramesMax, FramesStart: p.NFramesStart, FramesPast: p.NFramesPast,
		ConflictLimit: p.NConfLimit, LearnedMax: p.nLearnedMax, LearnedStart: p.NLearnedStart,
		LearnedDelta: p.NLearnedDelta, LearnedPercent: p.NLearnedPerce, TimeoutSeconds: p.NTimeOut,
		RatioMin: p.NRatioMin, RatioMinRefine: p.NRatioMin2, RatioMax: p.NRatioMax,
		UseTerminalVariables: p.FUseTermVars != 0, UseRollback: p.FUseRollback != 0,
		PropagateFanout: p.FPropFanout != 0, AddLayer: p.FAddLayer != 0, NewRefinement: p.FNewRefine != 0,
		SkipIntermediate: p.FUseSkip != 0, SimpleCNF: p.FUseSimple != 0, SkipHash: p.FSkipHash != 0,
		UseFullProof: p.FUseFullProof != 0, DumpVariableAbstraction: p.FDumpVabs != 0,
		DumpMappedAbstraction: p.FDumpMabs != 0, CallProver: p.FCallProver != 0,
		SimplifyBeforeProver: p.FSimpProver != 0, VariableAbstractionFile: p.PFileVabs,
		Verbose: p.FVerbose != 0, VeryVerbose: p.FVeryVerbose != 0, FramesNoChangeLimit: p.NFramesNoChangeLim,
	}
}

// SetOptions replaces the configuration, preserving progress from prior runs.
func (p *Abs_Par_t) SetOptions(o Options) {
	p.NFramesMax, p.NFramesStart, p.NFramesPast = o.FramesMax, o.FramesStart, o.FramesPast
	p.NConfLimit, p.nLearnedMax, p.NLearnedStart = o.ConflictLimit, o.LearnedMax, o.LearnedStart
	p.NLearnedDelta, p.NLearnedPerce, p.NTimeOut = o.LearnedDelta, o.LearnedPercent, o.TimeoutSeconds
	p.NRatioMin, p.NRatioMin2, p.NRatioMax = o.RatioMin, o.RatioMinRefine, o.RatioMax
	p.FUseTermVars, p.FUseRollback = flag(o.UseTerminalVariables), flag(o.UseRollback)
	p.FPropFanout, p.FAddLayer, p.FNewRefine = flag(o.PropagateFanout), flag(o.AddLayer), flag(o.NewRefinement)
	p.FUseSkip, p.FUseSimple, p.FSkipHash = flag(o.SkipIntermediate), flag(o.SimpleCNF), flag(o.SkipHash)
	p.FUseFullProof, p.FDumpVabs, p.FDumpMabs = flag(o.UseFullProof), flag(o.DumpVariableAbstraction), flag(o.DumpMappedAbstraction)
	p.FCallProver, p.FSimpProver = flag(o.CallProver), flag(o.SimplifyBeforeProver)
	p.PFileVabs, p.FVerbose, p.FVeryVerbose = o.VariableAbstractionFile, flag(o.Verbose), flag(o.VeryVerbose)
	p.NFramesNoChangeLim = o.FramesNoChangeLimit
}

// Progress returns the most recent frame and stability counters.
func (p *Abs_Par_t) Progress() Progress {
	return Progress{Frame: p.IFrame, FrameProved: p.iFrameProved, FramesNoChange: p.nFramesNoChange}
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
