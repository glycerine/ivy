package abc

// Hand-checked skeleton of src/base/abci/abc.c.

// CmdCommandId is Cmd_CommandId_t (src/base/cmd/cmd.h).
type CmdCommandId int

const (
	CMD_COMMAND_UNKNOWN CmdCommandId = iota
	CMD_COMMAND_ALIAS
	CMD_COMMAND_SOURCE
	CMD_COMMAND_SET
	CMD_COMMAND_READ_BLIF
	CMD_COMMAND_READ_AIGER
	CMD_COMMAND_STRASH
	CMD_COMMAND_PRINT_STATS
	CMD_COMMAND_PDR
	CMD_COMMAND_WRITE_AIGER_CEX
)

// Every Abc_CommandXxx( Abc_Frame_t * pAbc, int argc, char ** argv ) command
// handler below takes argv as []string and returns the command status int.

// Abc_FrameReplaceCex
// C: void Abc_FrameReplaceCex( Abc_Frame_t * pAbc, Abc_Cex_t ** ppCex )
func (pAbc *MainFrame) AbcFrameReplaceCex(ppCex **Counterexample) {
	panic("TODO, stubbed")
}

// Abc_FrameReplaceCexVec
// C: void Abc_FrameReplaceCexVec( Abc_Frame_t * pAbc, Vec_Ptr_t ** pvCexVec )
func (pAbc *MainFrame) AbcFrameReplaceCexVec(pvCexVec **PointerVector[Counterexample]) {
	panic("TODO, stubbed")
}

// Abc_FrameReplacePoEquivs
// C: void Abc_FrameReplacePoEquivs( Abc_Frame_t * pAbc, Vec_Ptr_t ** pvPoEquivs )
func (pAbc *MainFrame) AbcFrameReplacePoEquivs(pvPoEquivs **PointerVector[IntVector]) {
	panic("TODO, stubbed")
}

// Abc_FrameReplacePoStatuses
// C: void Abc_FrameReplacePoStatuses( Abc_Frame_t * pAbc, Vec_Int_t ** pvStatuses )
func (pAbc *MainFrame) AbcFrameReplacePoStatuses(pvStatuses **IntVector) {
	panic("TODO, stubbed")
}

// Abc_FrameDeriveStatusArray
// C: Vec_Int_t * Abc_FrameDeriveStatusArray( Vec_Ptr_t * vCexes )
func AbcFrameDeriveStatusArray(vCexes *PointerVector[Counterexample]) *IntVector {
	panic("TODO, stubbed")
}

// Abc_FrameDeriveStatusArray2
// C: Vec_Int_t * Abc_FrameDeriveStatusArray2( Vec_Ptr_t * vCexes )
func AbcFrameDeriveStatusArray2(vCexes *PointerVector[Counterexample]) *IntVector {
	panic("TODO, stubbed")
}

// Abc_FrameClearDesign
// C: void Abc_FrameClearDesign()
func AbcFrameClearDesign() {
	panic("TODO, stubbed")
}

// Abc_FrameUpdateGia
// C: void Abc_FrameUpdateGia( Abc_Frame_t * pAbc, Gia_Man_t * pNew )
func (pAbc *MainFrame) AbcFrameUpdateGia(pNew *GiaManager) {
	panic("TODO, stubbed")
}

// Abc_FrameGetGia
// C: Gia_Man_t * Abc_FrameGetGia( Abc_Frame_t * pAbc )
func (pAbc *MainFrame) AbcFrameGetGia() *GiaManager {
	panic("TODO, stubbed")
}

// Abc_Init
// C: void Abc_Init( Abc_Frame_t * pAbc )
func (pAbc *MainFrame) AbcInit() {
	panic("TODO, stubbed")
}

// AbcCommandDispatchById
// C: int AbcCommandDispatchById( Cmd_CommandId_t Id, Abc_Frame_t * pAbc, int argc, char ** argv, int * pResult )
func AbcCommandDispatchById(Id CmdCommandId, pAbc *MainFrame, argc int, argv []string, pResult *int) int {
	panic("TODO, stubbed")
}

// Abc_End
// C: void Abc_End( Abc_Frame_t * pAbc )
func (pAbc *MainFrame) AbcEnd() {
	panic("TODO, stubbed")
}

// Abc_NtkCompareWithBest (static)
// C: int Abc_NtkCompareWithBest( Abc_Ntk_t * pBest, Abc_Ntk_t * p, float * pnBestNtkArea, float * pnBestNtkDelay, int * pnBestNtkNodes, int * pnBestNtkLevels, int fArea )
func abcNtkCompareWithBest(pBest *ABCNetwork, p *ABCNetwork, pnBestNtkArea *float32, pnBestNtkDelay *float32, pnBestNtkNodes *int, pnBestNtkLevels *int, fArea int) int {
	panic("TODO, stubbed")
}

// Gia_ManCompareWithBest (static)
// C: int Gia_ManCompareWithBest( Gia_Man_t * pBest, Gia_Man_t * p, int * pnBestLuts, int * pnBestEdges, int * pnBestLevels, int fArea )
func giaManCompareWithBest(pBest *GiaManager, p *GiaManager, pnBestLuts *int, pnBestEdges *int, pnBestLevels *int, fArea int) int {
	panic("TODO, stubbed")
}

// Abc_CommandPrintStats (static)
// C: int Abc_CommandPrintStats( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandPrintStats(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandPrintExdc (static)
// C: int Abc_CommandPrintExdc( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandPrintExdc(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandPrintIo (static)
// C: int Abc_CommandPrintIo( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandPrintIo(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandPrintLatch (static)
// C: int Abc_CommandPrintLatch( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandPrintLatch(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandPrintFanio (static)
// C: int Abc_CommandPrintFanio( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandPrintFanio(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandPrintMffc (static)
// C: int Abc_CommandPrintMffc( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandPrintMffc(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandPrintFactor (static)
// C: int Abc_CommandPrintFactor( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandPrintFactor(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandPrintLevel (static)
// C: int Abc_CommandPrintLevel( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandPrintLevel(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandPrintSupport (static)
// C: int Abc_CommandPrintSupport( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandPrintSupport(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandPrintMint (static)
// C: int Abc_CommandPrintMint( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandPrintMint(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandPrintSymms (static)
// C: int Abc_CommandPrintSymms( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandPrintSymms(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandPrintUnate (static)
// C: int Abc_CommandPrintUnate( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandPrintUnate(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandPrintAuto (static)
// C: int Abc_CommandPrintAuto( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandPrintAuto(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandPrintKMap (static)
// C: int Abc_CommandPrintKMap( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandPrintKMap(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandPrintGates (static)
// C: int Abc_CommandPrintGates( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandPrintGates(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandPrintSharing (static)
// C: int Abc_CommandPrintSharing( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandPrintSharing(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandPrintXCut (static)
// C: int Abc_CommandPrintXCut( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandPrintXCut(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandPrintDsd (static)
// C: int Abc_CommandPrintDsd( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandPrintDsd(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandPrintCone (static)
// C: int Abc_CommandPrintCone( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandPrintCone(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandPrintMiter (static)
// C: int Abc_CommandPrintMiter( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandPrintMiter(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandPrintStatus (static)
// C: int Abc_CommandPrintStatus( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandPrintStatus(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandPrintDelay (static)
// C: int Abc_CommandPrintDelay( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandPrintDelay(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandShow (static)
// C: int Abc_CommandShow( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandShow(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandShowBdd (static)
// C: int Abc_CommandShowBdd( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandShowBdd(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandShowCut (static)
// C: int Abc_CommandShowCut( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandShowCut(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandCollapse (static)
// C: int Abc_CommandCollapse( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandCollapse(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandSatClp (static)
// C: int Abc_CommandSatClp( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandSatClp(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandStrash (static)
// C: int Abc_CommandStrash( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandStrash(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandBalance (static)
// C: int Abc_CommandBalance( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandBalance(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandMuxStruct (static)
// C: int Abc_CommandMuxStruct( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandMuxStruct(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandMulti (static)
// C: int Abc_CommandMulti( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandMulti(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandRenode (static)
// C: int Abc_CommandRenode( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandRenode(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandCleanup (static)
// C: int Abc_CommandCleanup( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandCleanup(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandSweep (static)
// C: int Abc_CommandSweep( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandSweep(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandFastExtract (static)
// C: int Abc_CommandFastExtract( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandFastExtract(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandFxch (static)
// C: int Abc_CommandFxch( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandFxch(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandEliminate (static)
// C: int Abc_CommandEliminate( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandEliminate(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandDisjoint (static)
// C: int Abc_CommandDisjoint( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandDisjoint(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandSparsify (static)
// C: int Abc_CommandSparsify( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandSparsify(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandLutpack (static)
// C: int Abc_CommandLutpack( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandLutpack(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandLutmin (static)
// C: int Abc_CommandLutmin( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandLutmin(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandMfs (static)
// C: int Abc_CommandMfs( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandMfs(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandMfs2 (static)
// C: int Abc_CommandMfs2( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandMfs2(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandMfs3 (static)
// C: int Abc_CommandMfs3( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandMfs3(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandMfse (static)
// C: int Abc_CommandMfse( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandMfse(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandLogicPush (static)
// C: int Abc_CommandLogicPush( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandLogicPush(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandTrace (static)
// C: int Abc_CommandTrace( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandTrace(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandGlitch (static)
// C: int Abc_CommandGlitch( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandGlitch(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandSpeedup (static)
// C: int Abc_CommandSpeedup( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandSpeedup(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandPowerdown (static)
// C: int Abc_CommandPowerdown( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandPowerdown(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAddBuffs (static)
// C: int Abc_CommandAddBuffs( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAddBuffs(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandMerge (static)
// C: int Abc_CommandMerge( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandMerge(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandTestDec (static)
// C: int Abc_CommandTestDec( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandTestDec(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandTestNpn (static)
// C: int Abc_CommandTestNpn( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandTestNpn(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandTestTruth (static)
// C: int Abc_CommandTestTruth( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandTestTruth(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandRunEco (static)
// C: int Abc_CommandRunEco( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandRunEco(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandRunGen (static)
// C: int Abc_CommandRunGen( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandRunGen(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandRunSim (static)
// C: int Abc_CommandRunSim( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandRunSim(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandRunTest (static)
// C: int Abc_CommandRunTest( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandRunTest(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandRewrite (static)
// C: int Abc_CommandRewrite( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandRewrite(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandRefactor (static)
// C: int Abc_CommandRefactor( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandRefactor(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandRestructure (static)
// C: int Abc_CommandRestructure( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandRestructure(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandResubstitute (static)
// C: int Abc_CommandResubstitute( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandResubstitute(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandRr (static)
// C: int Abc_CommandRr( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandRr(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandCascade (static)
// C: int Abc_CommandCascade( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandCascade(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandExtract (static)
// C: int Abc_CommandExtract( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandExtract(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandVarMin (static)
// C: int Abc_CommandVarMin( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandVarMin(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandFaultClasses (static)
// C: int Abc_CommandFaultClasses( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandFaultClasses(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandExact (static)
// C: int Abc_CommandExact( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandExact(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandBmsStart (static)
// C: int Abc_CommandBmsStart( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandBmsStart(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandBmsStop (static)
// C: int Abc_CommandBmsStop( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandBmsStop(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandBmsPs (static)
// C: int Abc_CommandBmsPs( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandBmsPs(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandMajExact (static)
// C: int Abc_CommandMajExact( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandMajExact(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandTwoExact (static)
// C: int Abc_CommandTwoExact( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandTwoExact(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandLutExact (static)
// C: int Abc_CommandLutExact( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandLutExact(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAllExact (static)
// C: int Abc_CommandAllExact( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAllExact(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandTestExact (static)
// C: int Abc_CommandTestExact( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandTestExact(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandMajGen (static)
// C: int Abc_CommandMajGen( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandMajGen(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandLogic (static)
// C: int Abc_CommandLogic( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandLogic(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandComb (static)
// C: int Abc_CommandComb( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandComb(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandMiter (static)
// C: int Abc_CommandMiter( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandMiter(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandMiter2 (static)
// C: int Abc_CommandMiter2( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandMiter2(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandDemiter (static)
// C: int Abc_CommandDemiter( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandDemiter(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandOrPos (static)
// C: int Abc_CommandOrPos( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandOrPos(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAndPos (static)
// C: int Abc_CommandAndPos( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAndPos(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandZeroPo (static)
// C: int Abc_CommandZeroPo( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandZeroPo(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandSwapPos (static)
// C: int Abc_CommandSwapPos( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandSwapPos(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandRemovePo (static)
// C: int Abc_CommandRemovePo( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandRemovePo(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandDropSat (static)
// C: int Abc_CommandDropSat( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandDropSat(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAddPi (static)
// C: int Abc_CommandAddPi( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAddPi(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAppend (static)
// C: int Abc_CommandAppend( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAppend(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandPutOnTop (static)
// C: int Abc_CommandPutOnTop( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandPutOnTop(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandFrames (static)
// C: int Abc_CommandFrames( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandFrames(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandDFrames (static)
// C: int Abc_CommandDFrames( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandDFrames(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandSop (static)
// C: int Abc_CommandSop( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandSop(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandBdd (static)
// C: int Abc_CommandBdd( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandBdd(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAig (static)
// C: int Abc_CommandAig( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAig(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandReorder (static)
// C: int Abc_CommandReorder( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandReorder(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandBidec (static)
// C: int Abc_CommandBidec( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandBidec(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandOrder (static)
// C: int Abc_CommandOrder( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandOrder(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandMuxes (static)
// C: int Abc_CommandMuxes( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandMuxes(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandCubes (static)
// C: int Abc_CommandCubes( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandCubes(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandExpand (static)
// C: int Abc_CommandExpand( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandExpand(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandSplitSop (static)
// C: int Abc_CommandSplitSop( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandSplitSop(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandExtSeqDcs (static)
// C: int Abc_CommandExtSeqDcs( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandExtSeqDcs(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandReach (static)
// C: int Abc_CommandReach( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandReach(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandCone (static)
// C: int Abc_CommandCone( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandCone(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandNode (static)
// C: int Abc_CommandNode( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandNode(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandCof (static)
// C: int Abc_CommandCof( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandCof(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandTopmost (static)
// C: int Abc_CommandTopmost( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandTopmost(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandBottommost (static)
// C: int Abc_CommandBottommost( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandBottommost(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandTopAnd (static)
// C: int Abc_CommandTopAnd( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandTopAnd(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandTrim (static)
// C: int Abc_CommandTrim( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandTrim(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandShortNames (static)
// C: int Abc_CommandShortNames( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandShortNames(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandMoveNames (static)
// C: int Abc_CommandMoveNames( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandMoveNames(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandExdcFree (static)
// C: int Abc_CommandExdcFree( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandExdcFree(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandExdcGet (static)
// C: int Abc_CommandExdcGet( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandExdcGet(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandExdcSet (static)
// C: int Abc_CommandExdcSet( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandExdcSet(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandCareSet (static)
// C: int Abc_CommandCareSet( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandCareSet(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandCut (static)
// C: int Abc_CommandCut( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandCut(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandScut (static)
// C: int Abc_CommandScut( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandScut(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandEspresso (static)
// C: int Abc_CommandEspresso( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandEspresso(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandGen (static)
// C: int Abc_CommandGen( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandGen(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandGenFsm (static)
// C: int Abc_CommandGenFsm( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandGenFsm(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandCover (static)
// C: int Abc_CommandCover( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandCover(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandInter (static)
// C: int Abc_CommandInter( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandInter(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandDouble (static)
// C: int Abc_CommandDouble( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandDouble(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandBb2Wb (static)
// C: int Abc_CommandBb2Wb( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandBb2Wb(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandOutdec (static)
// C: int Abc_CommandOutdec( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandOutdec(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandNodeDup (static)
// C: int Abc_CommandNodeDup( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandNodeDup(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandTestColor (static)
// C: int Abc_CommandTestColor( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandTestColor(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandTest (static)
// C: int Abc_CommandTest( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandTest(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandQuaVar (static)
// C: int Abc_CommandQuaVar( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandQuaVar(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandQuaRel (static)
// C: int Abc_CommandQuaRel( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandQuaRel(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandQuaReach (static)
// C: int Abc_CommandQuaReach( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandQuaReach(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandSenseInput (static)
// C: int Abc_CommandSenseInput( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandSenseInput(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandIStrash (static)
// C: int Abc_CommandIStrash( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandIStrash(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandICut (static)
// C: int Abc_CommandICut( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandICut(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandIRewrite (static)
// C: int Abc_CommandIRewrite( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandIRewrite(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandDRewrite (static)
// C: int Abc_CommandDRewrite( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandDRewrite(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandDRefactor (static)
// C: int Abc_CommandDRefactor( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandDRefactor(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandDc2 (static)
// C: int Abc_CommandDc2( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandDc2(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandDChoice (static)
// C: int Abc_CommandDChoice( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandDChoice(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandDch (static)
// C: int Abc_CommandDch( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandDch(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandDrwsat (static)
// C: int Abc_CommandDrwsat( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandDrwsat(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandIRewriteSeq (static)
// C: int Abc_CommandIRewriteSeq( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandIRewriteSeq(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandIResyn (static)
// C: int Abc_CommandIResyn( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandIResyn(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandISat (static)
// C: int Abc_CommandISat( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandISat(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandIFraig (static)
// C: int Abc_CommandIFraig( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandIFraig(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandDFraig (static)
// C: int Abc_CommandDFraig( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandDFraig(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandCSweep (static)
// C: int Abc_CommandCSweep( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandCSweep(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandIProve (static)
// C: int Abc_CommandIProve( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandIProve(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandQbf (static)
// C: int Abc_CommandQbf( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandQbf(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandNpnLoad (static)
// C: int Abc_CommandNpnLoad( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandNpnLoad(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandNpnSave (static)
// C: int Abc_CommandNpnSave( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandNpnSave(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandSendAig (static)
// C: int Abc_CommandSendAig( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandSendAig(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandSendStatus (static)
// C: int Abc_CommandSendStatus( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandSendStatus(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandBackup (static)
// C: int Abc_CommandBackup( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandBackup(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandRestore (static)
// C: int Abc_CommandRestore( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandRestore(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandMinisat (static)
// C: int Abc_CommandMinisat( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandMinisat(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandMinisimp (static)
// C: int Abc_CommandMinisimp( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandMinisimp(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandFraig (static)
// C: int Abc_CommandFraig( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandFraig(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandFraigTrust (static)
// C: int Abc_CommandFraigTrust( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandFraigTrust(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandFraigStore (static)
// C: int Abc_CommandFraigStore( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandFraigStore(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandFraigRestore (static)
// C: int Abc_CommandFraigRestore( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandFraigRestore(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandFraigClean (static)
// C: int Abc_CommandFraigClean( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandFraigClean(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandFraigSweep (static)
// C: int Abc_CommandFraigSweep( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandFraigSweep(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandFraigDress (static)
// C: int Abc_CommandFraigDress( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandFraigDress(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandDumpEquiv (static)
// C: int Abc_CommandDumpEquiv( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandDumpEquiv(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandRecStart3 (static)
// C: int Abc_CommandRecStart3( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandRecStart3(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandRecStop3 (static)
// C: int Abc_CommandRecStop3( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandRecStop3(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandRecPs3 (static)
// C: int Abc_CommandRecPs3( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandRecPs3(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandRecAdd3 (static)
// C: int Abc_CommandRecAdd3( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandRecAdd3(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandRecDump3 (static)
// C: int Abc_CommandRecDump3( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandRecDump3(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandRecMerge3 (static)
// C: int Abc_CommandRecMerge3( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandRecMerge3(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandMap (static)
// C: int Abc_CommandMap( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandMap(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAmap (static)
// C: int Abc_CommandAmap( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAmap(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandPhaseMap (static)
// C: int Abc_CommandPhaseMap( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandPhaseMap(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandUnmap (static)
// C: int Abc_CommandUnmap( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandUnmap(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAttach (static)
// C: int Abc_CommandAttach( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAttach(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandSuperChoice (static)
// C: int Abc_CommandSuperChoice( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandSuperChoice(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandSuperChoiceLut (static)
// C: int Abc_CommandSuperChoiceLut( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandSuperChoiceLut(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandTimeScale (static)
// C: int Abc_CommandTimeScale( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandTimeScale(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandIf (static)
// C: int Abc_CommandIf( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandIf(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandIfif (static)
// C: int Abc_CommandIfif( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandIfif(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandDsdSave (static)
// C: int Abc_CommandDsdSave( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandDsdSave(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandDsdLoad (static)
// C: int Abc_CommandDsdLoad( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandDsdLoad(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandDsdFree (static)
// C: int Abc_CommandDsdFree( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandDsdFree(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandDsdPs (static)
// C: int Abc_CommandDsdPs( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandDsdPs(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandDsdMatch (static)
// C: int Abc_CommandDsdMatch( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandDsdMatch(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandDsdMerge (static)
// C: int Abc_CommandDsdMerge( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandDsdMerge(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandDsdFilter (static)
// C: int Abc_CommandDsdFilter( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandDsdFilter(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandInit (static)
// C: int Abc_CommandInit( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandInit(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandZero (static)
// C: int Abc_CommandZero( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandZero(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandUndc (static)
// C: int Abc_CommandUndc( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandUndc(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandOneHot (static)
// C: int Abc_CommandOneHot( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandOneHot(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandPipe (static)
// C: int Abc_CommandPipe( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandPipe(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandSeq (static)
// C: int Abc_CommandSeq( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandSeq(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandUnseq (static)
// C: int Abc_CommandUnseq( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandUnseq(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandRetime (static)
// C: int Abc_CommandRetime( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandRetime(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandDRetime (static)
// C: int Abc_CommandDRetime( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandDRetime(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandFlowRetime (static)
// C: int Abc_CommandFlowRetime( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandFlowRetime(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandCRetime (static)
// C: int Abc_CommandCRetime( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandCRetime(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandSeqFpga (static)
// C: int Abc_CommandSeqFpga( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandSeqFpga(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandSeqMap (static)
// C: int Abc_CommandSeqMap( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandSeqMap(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandSeqSweep (static)
// C: int Abc_CommandSeqSweep( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandSeqSweep(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandSeqSweep2 (static)
// C: int Abc_CommandSeqSweep2( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandSeqSweep2(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandTestSeqSweep (static)
// C: int Abc_CommandTestSeqSweep( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandTestSeqSweep(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandTestScorr (static)
// C: int Abc_CommandTestScorr( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandTestScorr(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandLcorr (static)
// C: int Abc_CommandLcorr( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandLcorr(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandSeqCleanup (static)
// C: int Abc_CommandSeqCleanup( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandSeqCleanup(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandCycle (static)
// C: int Abc_CommandCycle( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandCycle(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandXsim (static)
// C: int Abc_CommandXsim( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandXsim(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandSim (static)
// C: int Abc_CommandSim( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandSim(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandSim3 (static)
// C: int Abc_CommandSim3( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandSim3(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandDarPhase (static)
// C: int Abc_CommandDarPhase( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandDarPhase(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandSynch (static)
// C: int Abc_CommandSynch( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandSynch(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandClockGate (static)
// C: int Abc_CommandClockGate( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandClockGate(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandExtWin (static)
// C: int Abc_CommandExtWin( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandExtWin(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandInsWin (static)
// C: int Abc_CommandInsWin( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandInsWin(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandSymFun (static)
// C: int Abc_CommandSymFun( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandSymFun(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandPermute (static)
// C: int Abc_CommandPermute( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandPermute(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandUnpermute (static)
// C: int Abc_CommandUnpermute( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandUnpermute(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandCubeEnum (static)
// C: int Abc_CommandCubeEnum( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandCubeEnum(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandPathEnum (static)
// C: int Abc_CommandPathEnum( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandPathEnum(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandFunEnum (static)
// C: int Abc_CommandFunEnum( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandFunEnum(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandCec (static)
// C: int Abc_CommandCec( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandCec(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandDCec (static)
// C: int Abc_CommandDCec( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandDCec(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandDSec (static)
// C: int Abc_CommandDSec( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandDSec(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandDProve (static)
// C: int Abc_CommandDProve( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandDProve(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbSec (static)
// C: int Abc_CommandAbSec( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbSec(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandSimSec (static)
// C: int Abc_CommandSimSec( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandSimSec(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandMatch (static)
// C: int Abc_CommandMatch( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandMatch(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandSat (static)
// C: int Abc_CommandSat( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandSat(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandDSat (static)
// C: int Abc_CommandDSat( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandDSat(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandXSat (static)
// C: int Abc_CommandXSat( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandXSat(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandSatoko (static)
// C: int Abc_CommandSatoko( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandSatoko(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Satoko (static)
// C: int Abc_CommandAbc9Satoko( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Satoko(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Sat3 (static)
// C: int Abc_CommandAbc9Sat3( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Sat3(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandPSat (static)
// C: int Abc_CommandPSat( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandPSat(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandProve (static)
// C: int Abc_CommandProve( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandProve(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandDebug (static)
// C: int Abc_CommandDebug( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandDebug(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandEco (static)
// C: int Abc_CommandEco( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandEco(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandBmc (static)
// C: int Abc_CommandBmc( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandBmc(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandBmc2 (static)
// C: int Abc_CommandBmc2( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandBmc2(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandBmc3 (static)
// C: int Abc_CommandBmc3( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandBmc3(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandBmcInter (static)
// C: int Abc_CommandBmcInter( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandBmcInter(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandIndcut (static)
// C: int Abc_CommandIndcut( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandIndcut(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandEnlarge (static)
// C: int Abc_CommandEnlarge( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandEnlarge(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandTempor (static)
// C: int Abc_CommandTempor( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandTempor(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandInduction (static)
// C: int Abc_CommandInduction( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandInduction(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandConstr (static)
// C: int Abc_CommandConstr( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandConstr(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandUnfold (static)
// C: int Abc_CommandUnfold( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandUnfold(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandFold (static)
// C: int Abc_CommandFold( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandFold(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandBm (static)
// C: int Abc_CommandBm( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandBm(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandBm2 (static)
// C: int Abc_CommandBm2( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandBm2(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandSaucy (static)
// C: int Abc_CommandSaucy( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandSaucy(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandTestCex (static)
// C: int Abc_CommandTestCex( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandTestCex(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandPdr (static)
// C: int Abc_CommandPdr( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandPdr(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandReconcile (static)
// C: int Abc_CommandReconcile( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandReconcile(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandCexSave (static)
// C: int Abc_CommandCexSave( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandCexSave(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandCexLoad (static)
// C: int Abc_CommandCexLoad( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandCexLoad(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandCexCut (static)
// C: int Abc_CommandCexCut( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandCexCut(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandCexMerge (static)
// C: int Abc_CommandCexMerge( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandCexMerge(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandCexMin
// C: int Abc_CommandCexMin( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) AbcCommandCexMin(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandDualRail (static)
// C: int Abc_CommandDualRail( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandDualRail(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandBlockPo (static)
// C: int Abc_CommandBlockPo( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandBlockPo(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandIso (static)
// C: int Abc_CommandIso( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandIso(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandTraceStart (static)
// C: int Abc_CommandTraceStart( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandTraceStart(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandTraceCheck (static)
// C: int Abc_CommandTraceCheck( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandTraceCheck(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbcSave (static)
// C: int Abc_CommandAbcSave( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbcSave(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbcLoad (static)
// C: int Abc_CommandAbcLoad( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbcLoad(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Read (static)
// C: int Abc_CommandAbc9Read( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Read(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9ReadBlif (static)
// C: int Abc_CommandAbc9ReadBlif( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9ReadBlif(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9ReadCBlif (static)
// C: int Abc_CommandAbc9ReadCBlif( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9ReadCBlif(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9ReadStg (static)
// C: int Abc_CommandAbc9ReadStg( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9ReadStg(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9ReadVer (static)
// C: int Abc_CommandAbc9ReadVer( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9ReadVer(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Get (static)
// C: int Abc_CommandAbc9Get( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Get(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Put (static)
// C: int Abc_CommandAbc9Put( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Put(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Save (static)
// C: int Abc_CommandAbc9Save( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Save(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Save2 (static)
// C: int Abc_CommandAbc9Save2( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Save2(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9SaveAig (static)
// C: int Abc_CommandAbc9SaveAig( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9SaveAig(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Load (static)
// C: int Abc_CommandAbc9Load( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Load(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Load2 (static)
// C: int Abc_CommandAbc9Load2( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Load2(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9LoadAig (static)
// C: int Abc_CommandAbc9LoadAig( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9LoadAig(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9WriteVer (static)
// C: int Abc_CommandAbc9WriteVer( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9WriteVer(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Write (static)
// C: int Abc_CommandAbc9Write( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Write(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9WriteLut (static)
// C: int Abc_CommandAbc9WriteLut( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9WriteLut(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Ps (static)
// C: int Abc_CommandAbc9Ps( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Ps(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9PFan (static)
// C: int Abc_CommandAbc9PFan( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9PFan(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9PSig (static)
// C: int Abc_CommandAbc9PSig( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9PSig(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Status (static)
// C: int Abc_CommandAbc9Status( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Status(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9MuxProfile (static)
// C: int Abc_CommandAbc9MuxProfile( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9MuxProfile(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9MuxPos (static)
// C: int Abc_CommandAbc9MuxPos( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9MuxPos(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9MuxStr (static)
// C: int Abc_CommandAbc9MuxStr( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9MuxStr(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9PrintTruth (static)
// C: int Abc_CommandAbc9PrintTruth( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9PrintTruth(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Unate (static)
// C: int Abc_CommandAbc9Unate( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Unate(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Rex2Gia (static)
// C: int Abc_CommandAbc9Rex2Gia( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Rex2Gia(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9RexWalk (static)
// C: int Abc_CommandAbc9RexWalk( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9RexWalk(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Show (static)
// C: int Abc_CommandAbc9Show( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Show(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9SetRegNum (static)
// C: int Abc_CommandAbc9SetRegNum( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9SetRegNum(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Strash (static)
// C: int Abc_CommandAbc9Strash( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Strash(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Topand (static)
// C: int Abc_CommandAbc9Topand( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Topand(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Add1Hot (static)
// C: int Abc_CommandAbc9Add1Hot( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Add1Hot(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Cof (static)
// C: int Abc_CommandAbc9Cof( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Cof(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Trim (static)
// C: int Abc_CommandAbc9Trim( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Trim(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Dfs (static)
// C: int Abc_CommandAbc9Dfs( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Dfs(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Sim (static)
// C: int Abc_CommandAbc9Sim( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Sim(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Sim3 (static)
// C: int Abc_CommandAbc9Sim3( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Sim3(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9MLGen (static)
// C: int Abc_CommandAbc9MLGen( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9MLGen(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9MLTest (static)
// C: int Abc_CommandAbc9MLTest( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9MLTest(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9ReadSim (static)
// C: int Abc_CommandAbc9ReadSim( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9ReadSim(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9WriteSim (static)
// C: int Abc_CommandAbc9WriteSim( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9WriteSim(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9PrintSim (static)
// C: int Abc_CommandAbc9PrintSim( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9PrintSim(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9GenSim (static)
// C: int Abc_CommandAbc9GenSim( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9GenSim(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9SimRsb (static)
// C: int Abc_CommandAbc9SimRsb( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9SimRsb(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Resim (static)
// C: int Abc_CommandAbc9Resim( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Resim(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9SpecI (static)
// C: int Abc_CommandAbc9SpecI( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9SpecI(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Equiv (static)
// C: int Abc_CommandAbc9Equiv( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Equiv(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Equiv2 (static)
// C: int Abc_CommandAbc9Equiv2( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Equiv2(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Equiv3 (static)
// C: int Abc_CommandAbc9Equiv3( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Equiv3(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Semi (static)
// C: int Abc_CommandAbc9Semi( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Semi(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Times (static)
// C: int Abc_CommandAbc9Times( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Times(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Frames (static)
// C: int Abc_CommandAbc9Frames( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Frames(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Retime (static)
// C: int Abc_CommandAbc9Retime( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Retime(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Enable (static)
// C: int Abc_CommandAbc9Enable( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Enable(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Dc2 (static)
// C: int Abc_CommandAbc9Dc2( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Dc2(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Dsd (static)
// C: int Abc_CommandAbc9Dsd( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Dsd(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Bidec (static)
// C: int Abc_CommandAbc9Bidec( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Bidec(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Shrink (static)
// C: int Abc_CommandAbc9Shrink( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Shrink(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Fx (static)
// C: int Abc_CommandAbc9Fx( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Fx(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Balance (static)
// C: int Abc_CommandAbc9Balance( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Balance(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9BalanceLut (static)
// C: int Abc_CommandAbc9BalanceLut( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9BalanceLut(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Resub (static)
// C: int Abc_CommandAbc9Resub( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Resub(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Syn2 (static)
// C: int Abc_CommandAbc9Syn2( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Syn2(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Synch2 (static)
// C: int Abc_CommandAbc9Synch2( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Synch2(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Syn3 (static)
// C: int Abc_CommandAbc9Syn3( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Syn3(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Syn4 (static)
// C: int Abc_CommandAbc9Syn4( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Syn4(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9False (static)
// C: int Abc_CommandAbc9False( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9False(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Miter (static)
// C: int Abc_CommandAbc9Miter( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Miter(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Miter2 (static)
// C: int Abc_CommandAbc9Miter2( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Miter2(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Append (static)
// C: int Abc_CommandAbc9Append( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Append(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Scl (static)
// C: int Abc_CommandAbc9Scl( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Scl(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Lcorr (static)
// C: int Abc_CommandAbc9Lcorr( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Lcorr(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Scorr (static)
// C: int Abc_CommandAbc9Scorr( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Scorr(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Choice (static)
// C: int Abc_CommandAbc9Choice( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Choice(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Sat (static)
// C: int Abc_CommandAbc9Sat( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Sat(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9SatEnum (static)
// C: int Abc_CommandAbc9SatEnum( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9SatEnum(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Fraig (static)
// C: int Abc_CommandAbc9Fraig( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Fraig(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9CFraig (static)
// C: int Abc_CommandAbc9CFraig( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9CFraig(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Srm (static)
// C: int Abc_CommandAbc9Srm( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Srm(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Srm2 (static)
// C: int Abc_CommandAbc9Srm2( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Srm2(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Filter (static)
// C: int Abc_CommandAbc9Filter( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Filter(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Reduce (static)
// C: int Abc_CommandAbc9Reduce( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Reduce(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9EquivMark (static)
// C: int Abc_CommandAbc9EquivMark( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9EquivMark(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9EquivFilter (static)
// C: int Abc_CommandAbc9EquivFilter( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9EquivFilter(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Cec (static)
// C: int Abc_CommandAbc9Cec( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Cec(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Verify (static)
// C: int Abc_CommandAbc9Verify( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Verify(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Sweep (static)
// C: int Abc_CommandAbc9Sweep( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Sweep(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Force (static)
// C: int Abc_CommandAbc9Force( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Force(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Embed (static)
// C: int Abc_CommandAbc9Embed( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Embed(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Sopb (static)
// C: int Abc_CommandAbc9Sopb( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Sopb(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Dsdb (static)
// C: int Abc_CommandAbc9Dsdb( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Dsdb(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Flow (static)
// C: int Abc_CommandAbc9Flow( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Flow(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Flow2 (static)
// C: int Abc_CommandAbc9Flow2( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Flow2(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Flow3 (static)
// C: int Abc_CommandAbc9Flow3( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Flow3(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9If (static)
// C: int Abc_CommandAbc9If( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9If(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Iff (static)
// C: int Abc_CommandAbc9Iff( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Iff(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Iiff (static)
// C: int Abc_CommandAbc9Iiff( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Iiff(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9If2 (static)
// C: int Abc_CommandAbc9If2( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9If2(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Jf (static)
// C: int Abc_CommandAbc9Jf( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Jf(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Kf (static)
// C: int Abc_CommandAbc9Kf( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Kf(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Lf (static)
// C: int Abc_CommandAbc9Lf( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Lf(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Mf (static)
// C: int Abc_CommandAbc9Mf( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Mf(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Nf (static)
// C: int Abc_CommandAbc9Nf( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Nf(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Of (static)
// C: int Abc_CommandAbc9Of( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Of(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Pack (static)
// C: int Abc_CommandAbc9Pack( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Pack(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Edge (static)
// C: int Abc_CommandAbc9Edge( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Edge(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9SatLut (static)
// C: int Abc_CommandAbc9SatLut( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9SatLut(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Unmap (static)
// C: int Abc_CommandAbc9Unmap( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Unmap(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Struct (static)
// C: int Abc_CommandAbc9Struct( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Struct(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Trace (static)
// C: int Abc_CommandAbc9Trace( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Trace(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Speedup (static)
// C: int Abc_CommandAbc9Speedup( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Speedup(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Era (static)
// C: int Abc_CommandAbc9Era( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Era(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Dch (static)
// C: int Abc_CommandAbc9Dch( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Dch(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Rpm (static)
// C: int Abc_CommandAbc9Rpm( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Rpm(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9BackReach (static)
// C: int Abc_CommandAbc9BackReach( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9BackReach(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Posplit (static)
// C: int Abc_CommandAbc9Posplit( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Posplit(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Permute (static)
// C: int Abc_CommandAbc9Permute( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Permute(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9ReachM (static)
// C: int Abc_CommandAbc9ReachM( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9ReachM(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9ReachP (static)
// C: int Abc_CommandAbc9ReachP( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9ReachP(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9ReachN (static)
// C: int Abc_CommandAbc9ReachN( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9ReachN(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9ReachY (static)
// C: int Abc_CommandAbc9ReachY( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9ReachY(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Undo (static)
// C: int Abc_CommandAbc9Undo( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Undo(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Mesh (static)
// C: int Abc_CommandAbc9Mesh( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Mesh(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Iso (static)
// C: int Abc_CommandAbc9Iso( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Iso(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9IsoNpn (static)
// C: int Abc_CommandAbc9IsoNpn( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9IsoNpn(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9IsoSt (static)
// C: int Abc_CommandAbc9IsoSt( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9IsoSt(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Compare (static)
// C: int Abc_CommandAbc9Compare( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Compare(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9CexInfo (static)
// C: int Abc_CommandAbc9CexInfo( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9CexInfo(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Cycle (static)
// C: int Abc_CommandAbc9Cycle( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Cycle(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Cone (static)
// C: int Abc_CommandAbc9Cone( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Cone(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Slice (static)
// C: int Abc_CommandAbc9Slice( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Slice(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9PoPart (static)
// C: int Abc_CommandAbc9PoPart( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9PoPart(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9PoPart2
// C: int Abc_CommandAbc9PoPart2( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) AbcCommandAbc9PoPart2(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9GroupProve (static)
// C: int Abc_CommandAbc9GroupProve( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9GroupProve(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9MultiProve (static)
// C: int Abc_CommandAbc9MultiProve( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9MultiProve(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9SplitProve (static)
// C: int Abc_CommandAbc9SplitProve( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9SplitProve(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Bmc (static)
// C: int Abc_CommandAbc9Bmc( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Bmc(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9SBmc (static)
// C: int Abc_CommandAbc9SBmc( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9SBmc(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9ChainBmc (static)
// C: int Abc_CommandAbc9ChainBmc( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9ChainBmc(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9BCore (static)
// C: int Abc_CommandAbc9BCore( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9BCore(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9ICheck (static)
// C: int Abc_CommandAbc9ICheck( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9ICheck(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9SatTest (static)
// C: int Abc_CommandAbc9SatTest( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9SatTest(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9FFTest (static)
// C: int Abc_CommandAbc9FFTest( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9FFTest(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Qbf (static)
// C: int Abc_CommandAbc9Qbf( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Qbf(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9QVar (static)
// C: int Abc_CommandAbc9QVar( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9QVar(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9GenQbf (static)
// C: int Abc_CommandAbc9GenQbf( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9GenQbf(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9HomoQbf (static)
// C: int Abc_CommandAbc9HomoQbf( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9HomoQbf(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9SatFx (static)
// C: int Abc_CommandAbc9SatFx( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9SatFx(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9SatClp (static)
// C: int Abc_CommandAbc9SatClp( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9SatClp(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Inse (static)
// C: int Abc_CommandAbc9Inse( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Inse(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Maxi (static)
// C: int Abc_CommandAbc9Maxi( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Maxi(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Bmci (static)
// C: int Abc_CommandAbc9Bmci( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Bmci(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9PoXsim (static)
// C: int Abc_CommandAbc9PoXsim( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9PoXsim(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Demiter (static)
// C: int Abc_CommandAbc9Demiter( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Demiter(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Fadds (static)
// C: int Abc_CommandAbc9Fadds( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Fadds(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9ATree (static)
// C: int Abc_CommandAbc9ATree( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9ATree(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Polyn (static)
// C: int Abc_CommandAbc9Polyn( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Polyn(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Acec (static)
// C: int Abc_CommandAbc9Acec( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Acec(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Anorm (static)
// C: int Abc_CommandAbc9Anorm( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Anorm(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Decla (static)
// C: int Abc_CommandAbc9Decla( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Decla(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Esop (static)
// C: int Abc_CommandAbc9Esop( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Esop(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Exorcism (static)
// C: int Abc_CommandAbc9Exorcism( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Exorcism(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Mfs (static)
// C: int Abc_CommandAbc9Mfs( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Mfs(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Mfsd (static)
// C: int Abc_CommandAbc9Mfsd( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Mfsd(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9DeepSyn (static)
// C: int Abc_CommandAbc9DeepSyn( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9DeepSyn(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9CexCut
// C: int Abc_CommandAbc9CexCut( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) AbcCommandAbc9CexCut(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9CexMerge
// C: int Abc_CommandAbc9CexMerge( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) AbcCommandAbc9CexMerge(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9CexMin
// C: int Abc_CommandAbc9CexMin( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) AbcCommandAbc9CexMin(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9AbsCreate (static)
// C: int Abc_CommandAbc9AbsCreate( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9AbsCreate(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9AbsDerive (static)
// C: int Abc_CommandAbc9AbsDerive( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9AbsDerive(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9AbsRefine (static)
// C: int Abc_CommandAbc9AbsRefine( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9AbsRefine(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9GlaDerive (static)
// C: int Abc_CommandAbc9GlaDerive( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9GlaDerive(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9GlaRefine (static)
// C: int Abc_CommandAbc9GlaRefine( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9GlaRefine(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9GlaShrink (static)
// C: int Abc_CommandAbc9GlaShrink( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9GlaShrink(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Gla (static)
// C: int Abc_CommandAbc9Gla( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Gla(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Vta (static)
// C: int Abc_CommandAbc9Vta( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Vta(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Vta2Gla (static)
// C: int Abc_CommandAbc9Vta2Gla( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Vta2Gla(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Gla2Vta (static)
// C: int Abc_CommandAbc9Gla2Vta( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Gla2Vta(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Fla2Gla (static)
// C: int Abc_CommandAbc9Fla2Gla( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Fla2Gla(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Gla2Fla (static)
// C: int Abc_CommandAbc9Gla2Fla( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Gla2Fla(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Gen (static)
// C: int Abc_CommandAbc9Gen( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Gen(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Cfs (static)
// C: int Abc_CommandAbc9Cfs( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Cfs(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandAbc9Test (static)
// C: int Abc_CommandAbc9Test( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) abcCommandAbc9Test(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Abc_CommandTestRPO
// C: int Abc_CommandTestRPO(Abc_Frame_t * pAbc, int argc, char ** argv)
func (pAbc *MainFrame) abcCommandTestRPO(argc int, argv []string) int {
	panic("TODO, stubbed")
}
