package abc

// Hand-checked skeleton of src/aig/gia/gia.h.

// GiaRpr is Gia_Rpr_t.
type GiaRpr struct {
	iRepr   uint32 // unsigned : 28 : representative node
	fProved uint32 // unsigned : 1 : marks the proved equivalence
	fFailed uint32 // unsigned : 1 : marks the failed equivalence
	fColorA uint32 // unsigned : 1 : marks cone of A
	fColorB uint32 // unsigned : 1 : marks cone of B
}

// GiaPlc is Gia_Plc_t.
type GiaPlc struct {
	fFixed uint32 // unsigned : 1 : the placement of this object is fixed
	xCoord uint32 // unsigned : 15 : x-ooordinate of the placement
	fUndef uint32 // unsigned : 1 : the placement of this object is not assigned
	yCoord uint32 // unsigned : 15 : y-ooordinate of the placement
}

// GpsPar is Gps_Par_t.
type GpsPar struct {
	fTents    int
	fSwitch   int
	fCut      int
	fNpn      int
	fLutProf  int
	fMuxXor   int
	fMiter    int
	fSkipMap  int
	fSlacks   int
	fNoColor  int
	pDumpFile string // char *
}

// EmbPar is Emb_Par_t.
type EmbPar struct {
	nDims      int // the number of dimension
	nSols      int // the number of solutions (typically, 2)
	nIters     int // the number of iterations of FORCE
	fRefine    int // use refinement by FORCE
	fCluster   int // use clustered representation
	fDump      int // dump Gnuplot file
	fDumpLarge int // dump Gnuplot file for large benchmarks
	fShowImage int // shows image if Gnuplot is installed
	fVerbose   int // verbose flag
}

// GiaParFra is Gia_ParFra_t: frames parameters.
type GiaParFra struct {
	nFrames      int // the number of frames to unroll
	fInit        int // initialize the timeframes
	fSaveLastLit int // adds POs for outputs of each frame
	fDisableSt   int // disables strashing
	fOrPos       int // ORs respective POs in each timeframe
	fVerbose     int // enables verbose output
}

// GiaParSim is Gia_ParSim_t: simulation parameters.
type GiaParSim struct {
	// user-controlled parameters
	nWords      int // the number of machine words
	nIters      int // the number of timeframes
	RandSeed    int // seed to generate random numbers
	TimeLimit   int // time limit in seconds
	fCheckMiter int // check if miter outputs are non-zero
	fVerbose    int // enables verbose output
	iOutFail    int // index of the failed output
}

// GiaManSim is Gia_ManSim_t.
type GiaManSim struct {
	pAig     *GiaManager
	pPars    *GiaParSim
	nWords   int
	vCis2Ids *IntVector
	vConsts  *IntVector
	// simulation information
	pDataSim    []uint32 // unsigned * : simulation data
	pDataSimCis []uint32 // unsigned * : simulation data for CIs
	pDataSimCos []uint32 // unsigned * : simulation data for COs
}

// JfPar is Jf_Par_t.
type JfPar struct {
	nLutSize           int
	nCutNum            int
	nProcNum           int
	nRounds            int
	nRoundsEla         int
	nRelaxRatio        int
	nCoarseLimit       int
	nAreaTuner         int
	nReqTimeFlex       int
	nVerbLimit         int
	nDelayLut1         int
	nDelayLut2         int
	nFastEdges         int
	DelayTarget        int
	fAreaOnly          int
	fPinPerm           int
	fPinQuick          int
	fPinFilter         int
	fOptEdge           int
	fUseMux7           int
	fPower             int
	fCoarsen           int
	fCutMin            int
	fFuncDsd           int
	fGenCnf            int
	fGenLit            int
	fCnfObjIds         int
	fAddOrCla          int
	fCnfMapping        int
	fPureAig           int
	fDoAverage         int
	fCutHashing        int
	fCutSimple         int
	fCutGroup          int
	fVerbose           int
	fVeryVerbose       int
	nLutSizeMax        int
	nCutNumMax         int
	nProcNumMax        int
	nLutSizeMux        int
	Delay              uint64 // word
	Area               uint64 // word
	Edge               uint64 // word
	Clause             uint64 // word
	Mux7               uint64 // word
	WordMapDelay       uint64 // word
	WordMapArea        uint64 // word
	WordMapDelayTarget uint64 // word
	MapDelay           int
	MapArea            float32
	MapAreaF           float32
	MapDelayTarget     float32
	Epsilon            float32
	pTimesArr          []float32 // float *
	pTimesReq          []float32 // float *
}
