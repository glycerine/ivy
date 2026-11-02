package abc

// Hand-checked skeleton of src/map/scl/sclLib.h.

// Hand-written skeleton of the types in src/map/scl/sclLib.h. ABC_SCL_CUR_VERSION
// is 8. The by-value Vec_* and SC_Surface members are written as values. The
// SC_LibForEach* / SC_CellForEachPin* / SC_RingForEachCell macros are not
// represented.

// SCDir is the SC_Dir enum.
type SCDir int

const (
	sc_dir_NULL SCDir = iota
	sc_dir_Input
	sc_dir_Output
	sc_dir_InOut
	sc_dir_Internal
)

// SCTSense is the SC_TSense enum: timing sense, positive-, negative- or non-unate.
type SCTSense int

const (
	sc_ts_NULL SCTSense = iota
	sc_ts_Pos
	sc_ts_Neg
	sc_ts_Non
)

// SCPair is SC_Pair.
type SCPair struct {
	rise float32
	fall float32
}

// SCPairI is SC_PairI.
type SCPairI struct {
	rise int
	fall int
}

// SCSizePars is SC_SizePars.
type SCSizePars struct {
	nIters        int
	nIterNoChange int
	Window        int // used for upsizing
	Ratio         int // used for upsizing
	Notches       int
	DelayUser     int
	DelayGap      int
	TimeOut       int
	BuffTreeEst   int // ratio for buffer tree estimation
	BypassFreq    int // frequency to try bypassing
	fUseDept      int
	fDumpStats    int
	fUseWireLoads int
	fVerbose      int
	fVeryVerbose  int
}

// SCBusPars is SC_BusPars.
type SCBusPars struct {
	GainRatio     int // target gain
	Slew          int // target slew
	nDegree       int // max branching factor
	fSizeOnly     int // perform only sizing
	fAddBufs      int // add buffers
	fBufPis       int // use CI buffering
	fUseWireLoads int // wire loads
	fVerbose      int // verbose
	fVeryVerbose  int // verbose
}

// SCWireLoad is SC_WireLoad.
type SCWireLoad struct {
	pName   string      // char *
	cap     float32     // }- multiply estimation in 'fanout_len[].snd' with this value
	slope   float32     // used to extrapolate wireload for large fanout count
	vFanout IntVector   // Vec_Int_t (by value) : Vec<Pair<uint,float> > -- pairs '(#fanouts, est-wire-len)'
	vLen    FloatVector // Vec_Flt_t (by value)
}

// SCWireLoadSel is SC_WireLoadSel.
type SCWireLoadSel struct {
	pName          string                // char *
	vAreaFrom      FloatVector           // Vec_Flt_t (by value) : Vec<Trip<float,float,Str> > -- triplets '(from-area, upto-area, wire-load-model)'; range is [from, upto[
	vAreaTo        FloatVector           // Vec_Flt_t (by value)
	vWireLoadModel PointerVector[string] // Vec_Ptr_t (by value) : wire-load model names
}

// SCTableTempl is SC_TableTempl.
type SCTableTempl struct {
	pName  string                     // char *
	vVars  PointerVector[string]      // Vec_Ptr_t (by value) : Vec<Str> -- name of variable (numbered from 0, not 1 as in the Liberty file)
	vIndex PointerVector[FloatVector] // Vec_Ptr_t (by value) : Vec<Vec<float> > -- this is the point of measurement in table for the given variable
}

// SCSurface is SC_Surface.
type SCSurface struct {
	pName    string                     // char *
	vIndex0  FloatVector                // Vec_Flt_t (by value) : correspondes to "index_1" in the liberty file (for timing: slew)
	vIndex1  FloatVector                // Vec_Flt_t (by value) : correspondes to "index_2" in the liberty file (for timing: load)
	vData    PointerVector[FloatVector] // Vec_Ptr_t (by value) : Vec<Vec<float> > -- 'data[i0][i1]' gives value at '(index0[i0], index1[i1])'
	vIndex0I IntVector                  // Vec_Int_t (by value) : correspondes to "index_1" in the liberty file (for timing: slew)
	vIndex1I IntVector                  // Vec_Int_t (by value) : correspondes to "index_2" in the liberty file (for timing: load)
	vDataI   PointerVector[IntVector]   // Vec_Ptr_t (by value) : Vec<Vec<float> > -- 'data[i0][i1]' gives value at '(index0[i0], index1[i1])'
	approx   [3][6]float32
}

// SCTiming is SC_Timing.
type SCTiming struct {
	related_pin string    // char * : related pin
	tsense      SCTSense  // timing sense (positive_unate, negative_unate, non_unate)
	when_text   string    // char * : logic condition on inputs triggering this delay model for the output (currently not used)
	pCellRise   SCSurface // used to compute pin-to-pin delay
	pCellFall   SCSurface
	pRiseTrans  SCSurface // used to compute output slew
	pFallTrans  SCSurface
}

// SCTimings is SC_Timings.
type SCTimings struct {
	pName    string                  // char * : the 'related_pin' field
	vTimings PointerVector[SCTiming] // Vec_Ptr_t (by value) : structures of type SC_Timing
}

// SCPin is SC_Pin.
type SCPin struct {
	pName        string // char *
	dir          SCDir
	cap          float32                  // this value is used if 'rise_cap' and 'fall_cap' is missing (copied by 'postProcess()'). (not used)
	rise_cap     float32                  // }- used for input pins ('cap' too).
	fall_cap     float32                  // }
	rise_capI    int                      // }- used for input pins ('cap' too).
	fall_capI    int                      // }
	max_out_cap  float32                  // } (not used)
	max_out_slew float32                  // }- used only for output pins (max values must not be exceeded or else mapping is illegal) (not used)
	func_text    string                   // char * : }
	vFunc        WordVector               // Vec_Wrd_t (by value) : }
	vRTimings    PointerVector[SCTimings] // Vec_Ptr_t (by value) : for output pins
	//    SC_Timing      Timing;         // -- for output pins   (commented out in the C source)
}

// SCCell is SC_Cell.
type SCCell struct {
	pName          string // char *
	Id             int
	fSkip          int // skip this cell during genlib computation
	seq            int // set to TRUE by parser if a sequential element
	unsupp         int // set to TRUE by parser if cell contains information we cannot handle
	area           float32
	leakage        float32
	areaI          int
	leakageI       int
	drive_strength int                  // some library files provide this field (currently unused, but may be a good hint for sizing) (not used)
	vPins          PointerVector[SCPin] // Vec_Ptr_t (by value) : NamedSet<SC_Pin>
	n_inputs       int                  // 'pins[0 .. n_inputs-1]' are input pins
	n_outputs      int                  // 'pins[n_inputs .. n_inputs+n_outputs-1]' are output pins
	pNext          *SCCell              // same-functionality cells linked into a ring by area
	pPrev          *SCCell              // same-functionality cells linked into a ring by area
	pRepr          *SCCell              // representative of the class
	pAve           *SCCell              // average size cell of this class
	Order          int                  // order of the gate in the list
	nGates         int                  // the number of gates in the list
}

// SCLib is SC_Lib.
type SCLib struct {
	pName                 string  // char *
	pFileName             string  // char *
	default_wire_load     string  // char *
	default_wire_load_sel string  // char *
	default_max_out_slew  float32 // 'default_max_transition'; this is copied to each output pin where 'max_transition' is not defined  (not used)
	unit_time             int     // Valid 9..12. Unit is '10^(-val)' seconds (e.g. 9=1ns, 10=100ps, 11=10ps, 12=1ps)
	unit_cap_fst          float32 // First part is a multiplier, second either 12 or 15 for 'pf' or 'ff'.
	unit_cap_snd          int
	vWireLoads            PointerVector[SCWireLoad]    // Vec_Ptr_t (by value) : NamedSet<SC_WireLoad>
	vWireLoadSels         PointerVector[SCWireLoadSel] // Vec_Ptr_t (by value) : NamedSet<SC_WireLoadSel>
	vTempls               PointerVector[SCTableTempl]  // Vec_Ptr_t (by value) : NamedSet<SC_TableTempl>
	vCells                PointerVector[SCCell]        // Vec_Ptr_t (by value) : NamedSet<SC_Cell>
	vCellClasses          PointerVector[SCCell]        // Vec_Ptr_t (by value) : NamedSet<SC_Cell>
	pBins                 []int                        // int * : hashing gateName -> gateId
	nBins                 int
}

// SC_PairClean (static inline)
// C: void SC_PairClean( SC_Pair * d )
func (d *SCPair) scPairClean() {
	panic("TODO, stubbed")
}

// SC_PairMax (static inline)
// C: float SC_PairMax( SC_Pair * d )
func (d *SCPair) scPairMax() float32 {
	panic("TODO, stubbed")
}

// SC_PairMin (static inline)
// C: float SC_PairMin( SC_Pair * d )
func (d *SCPair) scPairMin() float32 {
	panic("TODO, stubbed")
}

// SC_PairAve (static inline)
// C: float SC_PairAve( SC_Pair * d )
func (d *SCPair) scPairAve() float32 {
	panic("TODO, stubbed")
}

// SC_PairDup (static inline)
// C: void SC_PairDup( SC_Pair * d, SC_Pair * s )
func (d *SCPair) scPairDup(s *SCPair) {
	panic("TODO, stubbed")
}

// SC_PairMove (static inline)
// C: void SC_PairMove( SC_Pair * d, SC_Pair * s )
func (d *SCPair) scPairMove(s *SCPair) {
	panic("TODO, stubbed")
}

// SC_PairAdd (static inline)
// C: void SC_PairAdd( SC_Pair * d, SC_Pair * s )
func (d *SCPair) scPairAdd(s *SCPair) {
	panic("TODO, stubbed")
}

// SC_PairEqual (static inline)
// C: int SC_PairEqual( SC_Pair * d, SC_Pair * s )
func (d *SCPair) scPairEqual(s *SCPair) int {
	panic("TODO, stubbed")
}

// SC_PairEqualE (static inline)
// C: int SC_PairEqualE( SC_Pair * d, SC_Pair * s, float E )
func (d *SCPair) scPairEqualE(s *SCPair, E float32) int {
	panic("TODO, stubbed")
}

// SC_LibCellNum (static inline)
// C: int SC_LibCellNum( SC_Lib * p )
func (p *SCLib) scLibCellNum() int {
	panic("TODO, stubbed")
}

// SC_LibCell (static inline)
// C: SC_Cell * SC_LibCell( SC_Lib * p, int i )
func (p *SCLib) scLibCell(i int) *SCCell {
	panic("TODO, stubbed")
}

// SC_CellPin (static inline)
// C: SC_Pin * SC_CellPin( SC_Cell * p, int i )
func (p *SCCell) scCellPin(i int) *SCPin {
	panic("TODO, stubbed")
}

// SC_CellFunc (static inline; the output pin's embedded vFunc)
// C: Vec_Wrd_t * SC_CellFunc( SC_Cell * p )
func (p *SCCell) scCellFunc() *WordVector {
	panic("TODO, stubbed")
}

// SC_CellPinCap (static inline)
// C: float SC_CellPinCap( SC_Cell * p, int i )
func (p *SCCell) scCellPinCap(i int) float32 {
	panic("TODO, stubbed")
}

// SC_CellPinCapAve (static inline)
// C: float SC_CellPinCapAve( SC_Cell * p )
func (p *SCCell) scCellPinCapAve() float32 {
	panic("TODO, stubbed")
}

// SC_CellPinOutFunc (static inline)
// C: char * SC_CellPinOutFunc( SC_Cell * p, int i )
func (p *SCCell) scCellPinOutFunc(i int) string {
	panic("TODO, stubbed")
}

// SC_CellPinName (static inline)
// C: char * SC_CellPinName( SC_Cell * p, int i )
func (p *SCCell) scCellPinName(i int) string {
	panic("TODO, stubbed")
}

// Abc_SclWireLoadAlloc (static inline)
// C: SC_WireLoad * Abc_SclWireLoadAlloc()
func abcSclWireLoadAlloc() *SCWireLoad {
	panic("TODO, stubbed")
}

// Abc_SclWireLoadSelAlloc (static inline)
// C: SC_WireLoadSel * Abc_SclWireLoadSelAlloc()
func abcSclWireLoadSelAlloc() *SCWireLoadSel {
	panic("TODO, stubbed")
}

// Abc_SclTableTemplAlloc (static inline)
// C: SC_TableTempl * Abc_SclTableTemplAlloc()
func abcSclTableTemplAlloc() *SCTableTempl {
	panic("TODO, stubbed")
}

// Abc_SclSurfaceAlloc (static inline)
// C: SC_Surface * Abc_SclSurfaceAlloc()
func abcSclSurfaceAlloc() *SCSurface {
	panic("TODO, stubbed")
}

// Abc_SclTimingAlloc (static inline)
// C: SC_Timing * Abc_SclTimingAlloc()
func abcSclTimingAlloc() *SCTiming {
	panic("TODO, stubbed")
}

// Abc_SclTimingsAlloc (static inline)
// C: SC_Timings * Abc_SclTimingsAlloc()
func abcSclTimingsAlloc() *SCTimings {
	panic("TODO, stubbed")
}

// Abc_SclPinAlloc (static inline)
// C: SC_Pin * Abc_SclPinAlloc()
func abcSclPinAlloc() *SCPin {
	panic("TODO, stubbed")
}

// Abc_SclCellAlloc (static inline)
// C: SC_Cell * Abc_SclCellAlloc()
func abcSclCellAlloc() *SCCell {
	panic("TODO, stubbed")
}

// Abc_SclLibAlloc (static inline)
// C: SC_Lib * Abc_SclLibAlloc()
func abcSclLibAlloc() *SCLib {
	panic("TODO, stubbed")
}

// Abc_SclWireLoadFree (static inline)
// C: void Abc_SclWireLoadFree( SC_WireLoad * p )
func abcSclWireLoadFree(p *SCWireLoad) {
	panic("TODO, stubbed")
}

// Abc_SclWireLoadSelFree (static inline)
// C: void Abc_SclWireLoadSelFree( SC_WireLoadSel * p )
func abcSclWireLoadSelFree(p *SCWireLoadSel) {
	panic("TODO, stubbed")
}

// Abc_SclTableTemplFree (static inline)
// C: void Abc_SclTableTemplFree( SC_TableTempl * p )
func abcSclTableTemplFree(p *SCTableTempl) {
	panic("TODO, stubbed")
}

// Abc_SclSurfaceFree (static inline)
// C: void Abc_SclSurfaceFree( SC_Surface * p )
func abcSclSurfaceFree(p *SCSurface) {
	panic("TODO, stubbed")
}

// Abc_SclTimingFree (static inline)
// C: void Abc_SclTimingFree( SC_Timing * p )
func abcSclTimingFree(p *SCTiming) {
	panic("TODO, stubbed")
}

// Abc_SclTimingsFree (static inline)
// C: void Abc_SclTimingsFree( SC_Timings * p )
func abcSclTimingsFree(p *SCTimings) {
	panic("TODO, stubbed")
}

// Abc_SclPinFree (static inline)
// C: void Abc_SclPinFree( SC_Pin * p )
func abcSclPinFree(p *SCPin) {
	panic("TODO, stubbed")
}

// Abc_SclCellFree (static inline)
// C: void Abc_SclCellFree( SC_Cell * p )
func abcSclCellFree(p *SCCell) {
	panic("TODO, stubbed")
}

// Abc_SclLibFree (static inline)
// C: void Abc_SclLibFree( SC_Lib * p )
func abcSclLibFree(p *SCLib) {
	panic("TODO, stubbed")
}

// Scl_LibLookup (static inline)
// C: float Scl_LibLookup( SC_Surface * p, float slew, float load )
func (p *SCSurface) sclLibLookup(slew float32, load float32) float32 {
	panic("TODO, stubbed")
}

// Scl_LibPinArrival (static inline)
// C: void Scl_LibPinArrival( SC_Timing * pTime, SC_Pair * pArrIn, SC_Pair * pSlewIn, SC_Pair * pLoad, SC_Pair * pArrOut, SC_Pair * pSlewOut )
func sclLibPinArrival(pTime *SCTiming, pArrIn *SCPair, pSlewIn *SCPair, pLoad *SCPair, pArrOut *SCPair, pSlewOut *SCPair) {
	panic("TODO, stubbed")
}

// Scl_LibPinDeparture (static inline)
// C: void Scl_LibPinDeparture( SC_Timing * pTime, SC_Pair * pDepIn, SC_Pair * pSlewIn, SC_Pair * pLoad, SC_Pair * pDepOut )
func sclLibPinDeparture(pTime *SCTiming, pDepIn *SCPair, pSlewIn *SCPair, pLoad *SCPair, pDepOut *SCPair) {
	panic("TODO, stubbed")
}

// Scl_LibLookupI (static inline)
// C: int Scl_LibLookupI( SC_Surface * p, int slew, int load )
func (p *SCSurface) sclLibLookupI(slew int, load int) int {
	panic("TODO, stubbed")
}

// Scl_LibPinArrivalI (static inline)
// C: void Scl_LibPinArrivalI( SC_Timing * pTime, SC_PairI * pArrIn, SC_PairI * pSlewIn, SC_PairI * pLoad, SC_PairI * pArrOut, SC_PairI * pSlewOut, int * pArray )
func sclLibPinArrivalI(pTime *SCTiming, pArrIn *SCPairI, pSlewIn *SCPairI, pLoad *SCPairI, pArrOut *SCPairI, pSlewOut *SCPairI, pArray []int) {
	panic("TODO, stubbed")
}

// Scl_LibPinRequiredI (static inline)
// C: void Scl_LibPinRequiredI( SC_Timing * pTime, SC_PairI * pReqIn, SC_PairI * pReqOut, int * pArray )
func sclLibPinRequiredI(pTime *SCTiming, pReqIn *SCPairI, pReqOut *SCPairI, pArray []int) {
	panic("TODO, stubbed")
}

// Scl_CellPinTime (static inline)
// C: SC_Timing * Scl_CellPinTime( SC_Cell * pCell, int iPin )
func sclCellPinTime(pCell *SCCell, iPin int) *SCTiming {
	panic("TODO, stubbed")
}

// Scl_LibPinArrivalEstimate (static inline)
// C: float Scl_LibPinArrivalEstimate( SC_Cell * pCell, int iPin, float Slew, float Load )
func sclLibPinArrivalEstimate(pCell *SCCell, iPin int, Slew float32, Load float32) float32 {
	panic("TODO, stubbed")
}

// Scl_LibHandleInputDriver (static inline)
// C: void Scl_LibHandleInputDriver( SC_Cell * pCell, SC_Pair * pLoadIn, SC_Pair * pArrOut, SC_Pair * pSlewOut )
func sclLibHandleInputDriver(pCell *SCCell, pLoadIn *SCPair, pArrOut *SCPair, pSlewOut *SCPair) {
	panic("TODO, stubbed")
}

// Scl_LibPinArrivalEstimateI (static inline)
// C: int Scl_LibPinArrivalEstimateI( SC_Cell * pCell, int iPin, int Slew, int Load )
func sclLibPinArrivalEstimateI(pCell *SCCell, iPin int, Slew int, Load int) int {
	panic("TODO, stubbed")
}

// Scl_LibHandleInputDriver2 (static inline)
// C: void Scl_LibHandleInputDriver2( SC_Cell * pCell, SC_PairI * pLoadIn, SC_PairI * pArrOut, SC_PairI * pSlewOut )
func sclLibHandleInputDriver2(pCell *SCCell, pLoadIn *SCPairI, pArrOut *SCPairI, pSlewOut *SCPairI) {
	panic("TODO, stubbed")
}
