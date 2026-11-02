package abc

// Hand-checked skeleton of src/map/scl/scl.c.

// Scl_ConGetMan (static)
// C: Scl_Con_t * Scl_ConGetMan( Abc_Frame_t * pAbc )
func (pAbc *MainFrame) sclConGetMan() *SclCon {
	panic("TODO, stubbed")
}

// Scl_ConFreeMan (static)
// C: void Scl_ConFreeMan( Abc_Frame_t * pAbc )
func (pAbc *MainFrame) sclConFreeMan() {
	panic("TODO, stubbed")
}

// Scl_ConUpdateMan (static)
// C: void Scl_ConUpdateMan( Abc_Frame_t * pAbc, Scl_Con_t * p )
func (pAbc *MainFrame) sclConUpdateMan(p *SclCon) {
	panic("TODO, stubbed")
}

// Scl_ConReadMan
// C: Scl_Con_t * Scl_ConReadMan()
func SclConReadMan() *SclCon {
	panic("TODO, stubbed")
}

// Abc_SclLoad (ppScl is an in/out parameter)
// C: void Abc_SclLoad( SC_Lib * pLib, SC_Lib ** ppScl )
func (pLib *SCLib) AbcSclLoad(ppScl **SCLib) {
	panic("TODO, stubbed")
}

// Scl_Init
// C: void Scl_Init( Abc_Frame_t * pAbc )
func (pAbc *MainFrame) SclInit() {
	panic("TODO, stubbed")
}

// Scl_End
// C: void Scl_End( Abc_Frame_t * pAbc )
func (pAbc *MainFrame) SclEnd() {
	panic("TODO, stubbed")
}

// Scl_CommandReadLib (static)
// C: int Scl_CommandReadLib( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) sclCommandReadLib(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Scl_CommandWriteLib (static)
// C: int Scl_CommandWriteLib( Abc_Frame_t * pAbc, int argc, char **argv )
func (pAbc *MainFrame) sclCommandWriteLib(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Scl_CommandPrintLib (static)
// C: int Scl_CommandPrintLib( Abc_Frame_t * pAbc, int argc, char **argv )
func (pAbc *MainFrame) sclCommandPrintLib(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Scl_CommandLeak2Area (static)
// C: int Scl_CommandLeak2Area( Abc_Frame_t * pAbc, int argc, char **argv )
func (pAbc *MainFrame) sclCommandLeak2Area(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Scl_CommandReadScl (static)
// C: int Scl_CommandReadScl( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) sclCommandReadScl(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Scl_CommandWriteScl (static)
// C: int Scl_CommandWriteScl( Abc_Frame_t * pAbc, int argc, char **argv )
func (pAbc *MainFrame) sclCommandWriteScl(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Scl_CommandDumpGen (static)
// C: int Scl_CommandDumpGen( Abc_Frame_t * pAbc, int argc, char **argv )
func (pAbc *MainFrame) sclCommandDumpGen(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Scl_CommandPrintGS (static)
// C: int Scl_CommandPrintGS( Abc_Frame_t * pAbc, int argc, char **argv )
func (pAbc *MainFrame) sclCommandPrintGS(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Scl_CommandStime (static)
// C: int Scl_CommandStime( Abc_Frame_t * pAbc, int argc, char **argv )
func (pAbc *MainFrame) sclCommandStime(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Scl_CommandTopo (static)
// C: int Scl_CommandTopo( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) sclCommandTopo(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Scl_CommandUnBuffer (static)
// C: int Scl_CommandUnBuffer( Abc_Frame_t * pAbc, int argc, char **argv )
func (pAbc *MainFrame) sclCommandUnBuffer(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Scl_CommandBuffer (static)
// C: int Scl_CommandBuffer( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) sclCommandBuffer(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Scl_CommandBufferOld (static)
// C: int Scl_CommandBufferOld( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) sclCommandBufferOld(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Scl_CommandMinsize (static)
// C: int Scl_CommandMinsize( Abc_Frame_t * pAbc, int argc, char **argv )
func (pAbc *MainFrame) sclCommandMinsize(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Scl_CommandMaxsize (static)
// C: int Scl_CommandMaxsize( Abc_Frame_t * pAbc, int argc, char **argv )
func (pAbc *MainFrame) sclCommandMaxsize(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Scl_CommandUpsize (static)
// C: int Scl_CommandUpsize( Abc_Frame_t * pAbc, int argc, char **argv )
func (pAbc *MainFrame) sclCommandUpsize(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Scl_CommandDnsize (static)
// C: int Scl_CommandDnsize( Abc_Frame_t * pAbc, int argc, char **argv )
func (pAbc *MainFrame) sclCommandDnsize(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Scl_CommandPrintBuf (static)
// C: int Scl_CommandPrintBuf( Abc_Frame_t * pAbc, int argc, char **argv )
func (pAbc *MainFrame) sclCommandPrintBuf(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Scl_CommandReadConstr (static)
// C: int Scl_CommandReadConstr( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) sclCommandReadConstr(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Scl_CommandWriteConstr (static)
// C: int Scl_CommandWriteConstr( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) sclCommandWriteConstr(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Scl_CommandPrintConstr (static)
// C: int Scl_CommandPrintConstr( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) sclCommandPrintConstr(argc int, argv []string) int {
	panic("TODO, stubbed")
}

// Scl_CommandResetConstr (static)
// C: int Scl_CommandResetConstr( Abc_Frame_t * pAbc, int argc, char ** argv )
func (pAbc *MainFrame) sclCommandResetConstr(argc int, argv []string) int {
	panic("TODO, stubbed")
}
