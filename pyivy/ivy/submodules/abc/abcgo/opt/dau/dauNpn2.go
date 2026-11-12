package dau

import "fmt"

import "os"

/**CFile****************************************************************

  FileName    [dauNpn2.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Functional enumeration.]

  Synopsis    [Functional enumeration.]

  Author      [Siang-Yun Lee]
  
  Affiliation [National Taiwan University]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: dauNpn2.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "dauInt.h"
//#include "misc/util/utilTruth.h"
//#include "misc/extra/extra.h"
//#include "aig/gia/gia.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

   
type Dtt_Man_t struct {
    nVars            int32;          // variable number
    nPerms            int32;         // number of permutations
    nComps            int32;         // number of complementations
    pPerms           *int32;         // permutations
    pComps           *int32;         // complementations
    pPres          *word;          // function marks
    vFanins     *Vec_Int_t;        // node fanins
    vTruths     *Vec_Int_t;        // node truth tables
    vConfigs     *Vec_Int_t;       // configurations
    vClasses     *Vec_Int_t;       // node NPN classes
    vTruthNpns     *Vec_Int_t;     // truth tables of the classes
    vFunNodes     *Vec_Wec_t;      // nodes by NPN class
    vTemp     *Vec_Int_t;          // temporary
    vTemp2     *Vec_Int_t;         // temporary
    FunMask       uint32;        // function mask
    CmpMask       uint32;        // function mask
    BinMask       uint32;        // hash mask
    pBins      *uint32;          // hash bins
    vUsedBins     *Vec_Int_t;      // used bins
    Counts [32]int32;     // node counts
    nClasses            int32;       // count of classes
    pTable      *uint32;         // mapping of funcs into their classes
    pNodes           *int32;         // the number of nodes in min-node network
    pTimes           *int32;         // the number of different min-node networks
    pVisited          string;       // visited classes
    vVisited     *Vec_Int_t;       // the number of visited classes
}

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////
 
/**Function*************************************************************

  Synopsis    [Parse one formula into a truth table.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dau_ParseFormulaEndToken(pForm string) string {
    var Counter int32;
    var pThis string;
    for  pThis = pForm; *pThis; pThis++  {
        if *pThis == '~' {
            continue; }
        if *pThis == '(' {
            Counter++;
        } else if *pThis == ')' {
            Counter--; }
        if Counter == 0 {
            return pThis + 1; }
    }
    assert(false, "0");
    return nil;
}
func Dau_ParseFormula_rec(pBeg string, pEnd string) word {
    var iFans [2]word;
    var Res word;
    var Oper word = -1;
    var pEndNew string;
    var fCompl int32;
    for  pBeg[0] == '~'  {
        pBeg++;
        fCompl ^= 1;
    }
    if pBeg + 1 == pEnd {
        if pBeg[0] >= 'a' && pBeg[0] <= 'f' {
            if fCompl != 0 {
                return ~s_Truths6[pBeg[0] - 'a'];
            } else {
                return s_Truths6[pBeg[0] - 'a'];
            } }
        assert(false, "0");
        return -1;
    }
    if pBeg[0] == '(' {
        pEndNew = Dau_ParseFormulaEndToken( pBeg );
        if pEndNew == pEnd {
            assert(pBeg[0] == '(', "pBeg[0] == '('");
            assert(pBeg[pEnd-pBeg-1] == ')', "pBeg[pEnd-pBeg-1] == ')'");
            Res = Dau_ParseFormula_rec( pBeg + 1, pEnd - 1 );
            if fCompl != 0 {
                return ~Res;
            } else {
                return Res;
            }
        }
    }
    // get first part
    pEndNew  = Dau_ParseFormulaEndToken( pBeg );
    iFans[0] = Dau_ParseFormula_rec( pBeg, pEndNew );
    iFans[0] = fCompl ? ~iFans[0] : iFans[0];
    Oper     = pEndNew[0];
    // get second part
    pBeg     = pEndNew + 1;
    pEndNew  = Dau_ParseFormulaEndToken( pBeg );
    iFans[1] = Dau_ParseFormula_rec( pBeg, pEndNew );
    // derive the formula
    if Oper == '&' {
        return iFans[0] & iFans[1]; }
    if Oper == '^' {
        return iFans[0] ^ iFans[1]; }
    assert(false, "0");
    return -1;
}
func Dau_ParseFormula(p string) word {
    return Dau_ParseFormula_rec( p, p + len(p) );
}
func Dau_ParseFormulaTest() {
    var p string  = "~((~~d&~(~~b&c))^(~(~a&~d)&~(~c^~b)))";
    var r word  = ABC_CONST(0x037d037d037d037d);
    var t word  = Dau_ParseFormula( p );
    assert(r == t, "r == t");
}

 
/**Function*************************************************************

  Synopsis    [Parse one formula into a AIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dau_ParseFormulaAig_rec(p *Gia_Man_t, pBeg string, pEnd string) int32 {
    var iFans [2]int32;
    var Res int32;
    var Oper int32 = -1;
    var pEndNew string;
    var fCompl int32;
    for  pBeg[0] == '~'  {
        pBeg++;
        fCompl ^= 1;
    }
    if pBeg + 1 == pEnd {
        if pBeg[0] >= 'a' && pBeg[0] <= 'f' {
            return Abc_Var2Lit( 1 + pBeg[0] - 'a', fCompl ); }
        assert(false, "0");
        return -1;
    }
    if pBeg[0] == '(' {
        pEndNew = Dau_ParseFormulaEndToken( pBeg );
        if pEndNew == pEnd {
            assert(pBeg[0] == '(', "pBeg[0] == '('");
            assert(pBeg[pEnd-pBeg-1] == ')', "pBeg[pEnd-pBeg-1] == ')'");
            Res = Dau_ParseFormulaAig_rec( p, pBeg + 1, pEnd - 1 );
            return Abc_LitNotCond( Res, fCompl );
        }
    }
    // get first part
    pEndNew  = Dau_ParseFormulaEndToken( pBeg );
    iFans[0] = Dau_ParseFormulaAig_rec( p, pBeg, pEndNew );
    iFans[0] = Abc_LitNotCond( iFans[0], fCompl );
    Oper     = pEndNew[0];
    // get second part
    pBeg     = pEndNew + 1;
    pEndNew  = Dau_ParseFormulaEndToken( pBeg );
    iFans[1] = Dau_ParseFormulaAig_rec( p, pBeg, pEndNew );
    // derive the formula
    if Oper == '&' {
        return Gia_ManHashAnd( p, iFans[0], iFans[1] ); }
    if Oper == '^' {
        return Gia_ManHashXor( p, iFans[0], iFans[1] ); }
    assert(false, "0");
    return -1;
}
func Dau_ParseFormulaAig(p *Gia_Man_t, pStr string) int32 {
    return Dau_ParseFormulaAig_rec( p, pStr, pStr + len(pStr) );
}
func Dau_ParseFormulaAigTest() *Gia_Man_t {
    var i int32;
    var pStr string  = "~((~~d&~(~~b&c))^(~(~a&~d)&~(~c^~b)))";
    var p *Gia_Man_t  = Gia_ManStart( 1000 );
    p.pName = Abc_UtilStrsav( "func_enum_aig" );
    Gia_ManHashAlloc( p );
    for  i = 0; i < 5; i++  {
        Gia_ManAppendCi( p );
    }
    Gia_ManAppendCo( p, Dau_ParseFormulaAig(p, pStr) );
    return p;
}


/**Function*************************************************************

  Synopsis    [Verify one file.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dau_VerifyFile(pFileName string) {
    var Buffer [1000]byte;
    var uTruth, uTruth2 uint32; 
    var nFails, nLines int32;
    var pFile *os.File  = os.Open(pFileName);
    for  fgets( Buffer, 1000, pFile )  {
        if Buffer[strlen(Buffer)-1] == '\n' {
            Buffer[strlen(Buffer)-1] = 0; }
        if Buffer[strlen(Buffer)-1] == '\r' {
            Buffer[strlen(Buffer)-1] = 0; }
        Extra_ReadHexadecimal( &uTruth2, Buffer, 5 );
        uTruth = uint32(Dau_ParseFormula( Buffer + 11 ));
        if uTruth != uTruth2 {
            printf( "Verification failed in line %d:  %s\n", nLines, Buffer );
nFails++; }
        nLines++;
    }
    fmt.Printf("Verification succeeded for %d functions and failed for %d functions.\n", nLines-nFails, nFails);
}
func Dau_VerifyFileTest() {
    var pFileName string  = "lib4var.txt";
    Dau_VerifyFile( pFileName );
}


/**Function*************************************************************

  Synopsis    [Create AIG for one file.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dau_ConstructAigFromFile(pFileName string) *Gia_Man_t {
    var Buffer [1000]byte;
    var i, nLines int32;
    var pFile *os.File  = os.Open(pFileName);
    var p *Gia_Man_t  = Gia_ManStart( 1000 );
    p.pName = Abc_UtilStrsav( "func_enum_aig" );
    Gia_ManHashAlloc( p );
    for  i = 0; i < 5; i++  {
        Gia_ManAppendCi( p );
    }
    for  fgets( Buffer, 1000, pFile )  {
        if Buffer[strlen(Buffer)-1] == '\n' {
            Buffer[strlen(Buffer)-1] = 0; }
        if Buffer[strlen(Buffer)-1] == '\r' {
            Buffer[strlen(Buffer)-1] = 0; }
        Gia_ManAppendCo( p, Dau_ParseFormulaAig(p, Buffer + 11) );
        nLines++;
    }
    fmt.Printf("Finish constructing AIG for %d structures.\n", nLines);
    return p;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dau_ReadFile2(pFileName string, nSizeW int32) *uint32 {
    var clk abctime  = Abc_Clock();
    var p *uint32;
    var RetValue int32;
    var pFile *os.File  = os.Open(pFileName);
    if pFile == nil { return nil; }
    // inlining of macro ABC_CALLOC
    p = (unsigned *)((word *) calloc((size_t)(nSizeW), sizeof(word)));
    if pFile != nil {
        RetValue = fread( p, sizeof(word), nSizeW, pFile );
    } else {
        RetValue = 0;
    }
    RetValue = 0;
    if pFile != nil {
        fmt.Printf("Finished reading file \"%s\".\n", pFileName);
        _ = (pFile).Close();
    
    } else {
        fmt.Printf("Cannot open input file \"%s\".\n", pFileName); }
    Abc_PrintTime( 1, "File reading", Abc_Clock() - clk );
    return p;
}
func Dtt_ManRenum(nVars int32, pTable *uint32, pnClasses *int32) {
    var i uint32;
    var Limit uint32 = 1 << ((1 << nVars)-1);
    var Count uint32;
    for  i = 0; i < Limit; i++  {
        if pTable[i] == i {
            pTable[i] = Count++;
        } else {
            assert(pTable[i] < i, "pTable[i] < i");
            pTable[i] = pTable[pTable[i]];
        }
    }
    fmt.Printf("The total number of NPN classes = %d.\n", Count);
    *pnClasses = Count;
}
func Dtt_ManLoadClasses(nVars int32, pnClasses *int32) *uint32 {
    extern void Dau_TruthEnum(int nVars);
    var pTable *uint32  = nil;
    var nSizeLog int32  = (1<<nVars) -2;
    var nSizeW int32  = 1 << nSizeLog;
    var pFileName string // was [200]byte in C

    pFileName = fmt.Sprintf("tableW%d.data", nSizeLog);
    pTable = Dau_ReadFile2( pFileName, nSizeW );
    if pTable == nil {
        Dau_TruthEnum(nVars);
        pTable = Dau_ReadFile2( pFileName, nSizeW );
    }
    Dtt_ManRenum( nVars, pTable, pnClasses );
    return pTable;
}
 
/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dtt_ManAddVisited(p *Dtt_Man_t, Truth2 uint32, n int32) {
    var Truth uint32;
    if Truth2 & p.CmpMask {
        Truth = ~Truth2;
    } else {
        Truth = Truth2;
    }
    var Class uint32  = p.pTable[Truth & p.FunMask];
    assert(Class < uint32(p.nClasses), "Class < uint32(p.nClasses)");
    if p.pNodes[Class] < n {
        return; }
    assert(p.pNodes[Class] == n, "p.pNodes[Class] == n");
    if p.pVisited[Class] {
        return; }
    p.pVisited[Class] = 1;
    Vec_IntPush( p.vVisited, Class );
}
func Dtt_ManProcessVisited(p *Dtt_Man_t) {
    var i, Class int32;
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(p.vVisited)) && (((Class) = Vec_IntEntry(p.vVisited, i)), 1); i++  {
        assert(p.pVisited[Class], "p.pVisited[Class]");
        p.pVisited[Class] = 0;
        p.pTimes[Class]++;
    }
    Vec_IntClear( p.vVisited );
}
 
/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dtt_ManAlloc(nVars int32, fMulti int32) *Dtt_Man_t {
    // inlining of macro ABC_CALLOC
    var p *Dtt_Man_t  = &Dtt_Man_t{};
    p.nVars      = nVars;
    p.nPerms     = Extra_Factorial( nVars );
    p.nComps     = 1 << nVars;
    p.pPerms     = Extra_PermSchedule( nVars );
    p.pComps     = Extra_GreyCodeSchedule( nVars );
    // inlining of macro ABC_CALLOC
    p.pPres      = ((word *) calloc((size_t)(1 << (p.nComps - 7)), sizeof(word)));
    p.vFanins    = Vec_IntAlloc( 2*617000 );
    p.vTruths    = Vec_IntAlloc( 617000 );
    p.vConfigs   = Vec_IntAlloc( 617000 );
    p.vClasses   = Vec_IntAlloc( 617000 );
    p.vTruthNpns = Vec_IntAlloc( 617000 );
    p.vFunNodes  = Vec_WecStart( 16 );
    p.vTemp      = Vec_IntAlloc( 4000 );
    p.vTemp2     = Vec_IntAlloc( 4000 );
    if nVars == 5 {
        p.FunMask = ~0;
    } else {
        if nVars == 4 {
            p.FunMask = 0xFFFF;
        } else {
            p.FunMask = 0xFF;
        }
    }
    if nVars == 5 {
        p.CmpMask = 1 << 31;
    } else {
        if nVars == 4 {
            p.CmpMask = 1 << 15;
        } else {
            p.CmpMask = 1 << 7;
        }
    }
    p.BinMask    = 0x3FFF;
    // inlining of macro ABC_FALLOC
    p.pBins      = ((unsigned *) memset(malloc(sizeof(unsigned) * (size_t)(p.BinMask + 1)), 0xff, sizeof(unsigned) * (size_t)(p.BinMask + 1)));
    p.vUsedBins  = Vec_IntAlloc( 4000 );
    if !(fMulti != 0) { return p; }
    p.pTable     = Dtt_ManLoadClasses( p.nVars, &p.nClasses );
    // inlining of macro ABC_CALLOC
    p.pNodes     = ((int *) calloc((size_t)(p.nClasses), sizeof(int)));
    // inlining of macro ABC_CALLOC
    p.pTimes     = ((int *) calloc((size_t)(p.nClasses), sizeof(int)));
    // inlining of macro ABC_CALLOC
    p.pVisited   = ((char *) calloc((size_t)(p.nClasses), sizeof(char)));
    p.vVisited   = Vec_IntAlloc( 1000 );
    return p;
}
func Dtt_ManFree(p *Dtt_Man_t) {
    Vec_IntFreeP( &p.vVisited );
    p.pTable = nil;
    p.pNodes = nil;
    p.pTimes = nil;
    p.pVisited = nil;
    Vec_IntFreeP( &p.vFanins );
    Vec_IntFreeP( &p.vTruths );
    Vec_IntFreeP( &p.vConfigs );
    Vec_IntFreeP( &p.vClasses );
    Vec_IntFreeP( &p.vTruthNpns );
    Vec_WecFreeP( &p.vFunNodes );
    Vec_IntFreeP( &p.vTemp );
    Vec_IntFreeP( &p.vTemp2 );
    Vec_IntFreeP( &p.vUsedBins );
    p.pPerms = nil;
    p.pComps = nil;
    p.pPres = nil;
    p.pBins = nil;
    p = nil;
}

/**Function*************************************************************

  Synopsis    [Collect representatives of the same class.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_Dtt_ManHashKey_s_P [4]uint32 = [4]uint32{ 1699, 5147, 7103, 8147 };
func Dtt_ManHashKey(p *Dtt_Man_t, Truth uint32) uint32 {
    
    unsigned char * pD = (unsigned char*)&Truth;
    return pD[0] * spStatic_Dtt_ManHashKey_s_P[0] + pD[1] * spStatic_Dtt_ManHashKey_s_P[1] + pD[2] * spStatic_Dtt_ManHashKey_s_P[2] + pD[3] * spStatic_Dtt_ManHashKey_s_P[3];
}
func Dtt_ManCheckHash(p *Dtt_Man_t, Truth uint32) int32 {
    var Hash uint32  = Dtt_ManHashKey(p, Truth);
    var pSpot *uint32  = p.pBins + (Hash & p.BinMask);
    for  ; ~*pSpot; Hash++, pSpot = p.pBins + (Hash & p.BinMask)  {
        if *pSpot == Truth { // equal
            return 0; }
    }
    Vec_IntPush( p.vUsedBins, pSpot - p.pBins );
    *pSpot = Truth;
    return 1;
}
func Dtt_ManCollect(p *Dtt_Man_t, Truth uint32, vFuns *Vec_Int_t) *Vec_Int_t {
    var i, k, Entry int32;
    var tCur word  = ((word)Truth << 32) | (word)Truth;
    Vec_IntClear( vFuns );
    for  i = 0; i < p.nPerms; i++  {
        for  k = 0; k < p.nComps; k++  {
//            unsigned tTemp = (unsigned)(tCur & 1 ? ~tCur : tCur);
            var tTemp uint32  = uint32((tCur & p.CmpMask ? ~tCur : tCur));
            if Dtt_ManCheckHash( p, tTemp ) != 0 {
                Vec_IntPush( vFuns, tTemp ); }
            tCur = Abc_Tt6Flip( tCur, p.pComps[k] );
        }
        tCur = Abc_Tt6SwapAdjacent( tCur, p.pPerms[i] );
    }
    assert(tCur == (((word)Truth << 32) | (word)Truth), "tCur == (((word)Truth << 32) | (word)Truth)");
    // clean hash table
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(p.vUsedBins)) && (((Entry) = Vec_IntEntry(p.vUsedBins, i)), 1); i++  {
        p.pBins[Entry] = ~0;
    }
    Vec_IntClear( p.vUsedBins );
    //printf( "%d ", Vec_IntSize(vFuns) );
    return vFuns;
}
 
/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dtt_ManGetFun(p *Dtt_Man_t, tFun uint32, n int32) int32 {
    var Class uint32;
    if tFun & p.CmpMask {
        tFun = ~tFun;
    } else {
        tFun = tFun;
    }
    //return Abc_TtGetBit( p->pPres, tFun & p->FunMask );
    if !Abc_TtGetBit( p.pPres, tFun & p.FunMask ) { return 0; }
    if p.pTable == nil { return 1; }

    Class = p.pTable[tFun & p.FunMask];
    assert(Class < uint32(p.nClasses), "Class < uint32(p.nClasses)");
    if p.pNodes[Class] < n {
        return 1; }
    assert(p.pNodes[Class] == n, "p.pNodes[Class] == n");
    if p.pVisited[Class] {
        return 1; }
    
    return 0;
}
func Dtt_ManSetFun(p *Dtt_Man_t, tFun uint32) {
    if tFun & p.CmpMask {
        tFun = ~tFun;
    } else {
        tFun = tFun;
    }
    //assert( !Dtt_ManGetFun(p, fFun & p->FunMask );
    Abc_TtSetBit( p.pPres, tFun & p.FunMask );
}
func Dtt_ManAddFunction(p *Dtt_Man_t, n int32, FanI int32, FanJ int32, Type int32, Truth uint32) {
    var vFuncs *Vec_Int_t  = Dtt_ManCollect( p, Truth, p.vTemp2 );
    var Min uint32  = Vec_IntFindMin( vFuncs ); 
    var i int32;
    var nObjs int32 = Vec_IntSize(p.vFanins)/2;
    var nNodesI int32  = 0xF & (Vec_IntEntry(p.vConfigs, FanI) >> 3);
    var nNodesJ int32  = 0xF & (Vec_IntEntry(p.vConfigs, FanJ) >> 3);
    var nNodes int32   = nNodesI + nNodesJ + 1;
    assert(nObjs == Vec_IntSize(p.vTruths), "nObjs == Vec_IntSize(p.vTruths)");
    assert(nObjs == Vec_IntSize(p.vConfigs), "nObjs == Vec_IntSize(p.vConfigs)");
    assert(nObjs == Vec_IntSize(p.vClasses), "nObjs == Vec_IntSize(p.vClasses)");
    Vec_WecPush( p.vFunNodes, n, nObjs );
    Vec_IntPushTwo( p.vFanins, FanI, FanJ );
    Vec_IntPush( p.vTruths, Truth );
    Vec_IntPush( p.vConfigs, (nNodes << 3) | Type );
    Vec_IntPush( p.vClasses, Vec_IntSize(p.vTruthNpns) );
    Vec_IntPush( p.vTruthNpns, Min );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vFuncs)) && (((Min) = Vec_IntEntry(vFuncs, i)), 1); i++  {
        Dtt_ManSetFun( p, Min );
    }
    assert(nNodes < 32, "nNodes < 32");
    p.Counts[nNodes]++;        

    if p.pTable == nil {
        return; }
    if Truth & p.CmpMask {
        Truth = ~Truth;
    } else {
        Truth = Truth;
    }
    Truth &= p.FunMask;
    //assert( p->pNodes[p->pTable[Truth]] == 0 );
    p.pNodes[p.pTable[Truth]] = n;
}

func Dtt_PrintStats(nNodes int32, nVars int32, vFunNodes *Vec_Wec_t, nSteps word, clk abctime, fDelay int32, nMultis word) int32 {
    var nNew int32  = Vec_IntSize(Vec_WecEntry(vFunNodes, nNodes));
    fmt.Printf("%c =%2d  |  ", fDelay ? 'D':'N', nNodes);
    fmt.Printf("C =%12.0f  |  ", float64((iword))nSteps);
    fmt.Printf("New%d =%10d   ", nVars, nNew + int32((nNodes==0)));
    fmt.Printf("All%d =%10d  |  ", nVars, Vec_WecSizeSize(vFunNodes)+1);
    fmt.Printf("Multi =%10d  |  ", int32(nMultis));
    Abc_PrintTime( 1, "Time",  Abc_Clock() - clk );
    //Abc_Print(1, "%9.2f sec\n", 1.0*(Abc_Clock() - clk)/(CLOCKS_PER_SEC));
    fflush(os.Stdout);
    return nNew;
}
func Dtt_PrintDistrib(p *Dtt_Man_t) {
    var i int32;
    fmt.Printf("NPN classes for each node count (N):\n");
    for  i = 0; i < 32; i++  {
        if p.Counts[i] {
            fmt.Printf("N = %2d : NPN = %6d\n", i, p.Counts[i]); }
    }
}
func Dtt_PrintMulti2(p *Dtt_Man_t) {
    var i, n int32;
    for  n = 0; n <= 7; n++  {
        fmt.Printf("n=%d : ", n);
        for  i = 0; i < p.nClasses; i++  {
            if p.pNodes[i] == n {
                fmt.Printf("%d ", p.pTimes[i]); }
        }
        fmt.Printf("\n");
    }
}
func Dtt_PrintMulti1(p *Dtt_Man_t) {
    var i, n, Entry, Count, Prev int32;
    for  n = 0; n < 16; n++  {
        var vTimes *Vec_Int_t  = Vec_IntAlloc( 100 );
        var vUsed *Vec_Int_t   = Vec_IntAlloc( 100 );
        for  i = 0; i < p.nClasses; i++  {
            if p.pNodes[i] == n {
                Vec_IntPush( vTimes, p.pTimes[i] ); }
        }
        if Vec_IntSize(vTimes) == 0 {
            Vec_IntFree(vTimes);
            Vec_IntFree(vUsed);
            break;
        }
        Vec_IntSort( vTimes, 0 );
        Count = 1;
        Prev = Vec_IntEntry( vTimes, 0 );
        // inlining of macro Vec_IntForEachEntryStart
        for  i = 1; (i < Vec_IntSize(vTimes)) && (((Entry) = Vec_IntEntry(vTimes, i)), 1); i++  {
            if Prev == Entry {
                Count++;
            } else {
                assert(Prev < Entry, "Prev < Entry");
                Vec_IntPushTwo( vUsed, Prev, Count );
                Count = 1;
                Prev = Entry;
            }
        }
        if Count > 0 {
            Vec_IntPushTwo( vUsed, Prev, Count ); }
        fmt.Printf("n=%d : ", n);
        // inlining of macro Vec_IntForEachEntryDouble
        for  i = 0; (i+1 < Vec_IntSize(vUsed)) && (((Prev) = Vec_IntEntry(vUsed, i)), 1) && (((Entry) = Vec_IntEntry(vUsed, i+1)), 1); i += 2  {
            fmt.Printf("%d=%d ", Prev, Entry);
        }
        fmt.Printf("\n");
        Vec_IntFree( vTimes );
        Vec_IntFree( vUsed );
    }
}
func Dtt_PrintMulti(p *Dtt_Man_t) {
    var n int32;
    var Counts [13][15]int32 = [13][15]int32{{0}};
    for  n = 0; n < 13; n++  {
        var i, Total, Count int32;
        for  i = 0; i < p.nClasses; i++  {
            if p.pNodes[i] == n {
                var Log int32  = Abc_Base2Log(p.pTimes[i]);
                assert(Log < 15, "Log < 15");
                if p.pTimes[i] < 2 {
                    Counts[n][0]++;
                } else {
                    Counts[n][Log]++; }
                Total += p.pTimes[i];
                Count++;
            }
        }
        if Count == 0 {
            break; }
        fmt.Printf("n=%2d : ", n);
        fmt.Printf("All = %7d  ", Count);
        fmt.Printf("Ave = %6.2f  ", 1.0*Total/Count);
        for  i = 0; i < 15; i++  {
            if Counts[n][i] {
                fmt.Printf("%6d", Counts[n][i]);
            } else {
                fmt.Printf("%6s", ""); }
        }
        fmt.Printf("\n");
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
   
type Dtt_FunImpl_t struct {
    Type int32;
    NP1 int32;     // 4 bits for an input (NPPP)
    FI1 int32;     // the id (in vLibFun) of the first fanin function
    NP2 int32;
    FI2 int32;     // the id (in vLibFun) of the second fanin function
}

func Dtt_FunImplFI2Str(FI int32, NP int32, vLibFun *Vec_Int_t, str string) {
    var i int32;
    var P [5]int32;
    var N [5]int32;
    for  i = 0; i < 5; i++  {
        P[i] = NP & 0x7;
        NP = NP >> 3;
        N[i] = NP & 0x1;
        NP = NP >> 1 ;
    }
    str = fmt.Sprintf("[%08x(%03d),%d%d%d%d%d,%d%d%d%d%d]", Vec_IntEntry( vLibFun, FI ), FI, P[0], P[1], P[2], P[3], P[4], N[0], N[1], N[2], N[3], N[4]); 
}

func Dtt_FunImpl2Str(Type int32, sFI1 string, sFI2 string, str string) {
    // 0: 1&2, 1: ~1&2, 2: 1&~2, 3: 1|2 = ~(~1&~2), 4: 1^2
    // 5: ~(1&2), 6: ~(~1&2), 7: ~(1&~2), 8: ~1&~2, 9: ~(1^2)
    switch  Type  {
        case 0: str = fmt.Sprintf("(%s&%s)", sFI1, sFI2); break;
        case 1: str = fmt.Sprintf("(~%s&%s)", sFI1, sFI2); break;
        case 2: str = fmt.Sprintf("(%s&~%s)", sFI1, sFI2); break;
        case 3: str = fmt.Sprintf("~(~%s&~%s)", sFI1, sFI2); break;
        case 4: str = fmt.Sprintf("(%s^%s)", sFI1, sFI2); break;
        case 5: str = fmt.Sprintf("~(%s&%s)", sFI1, sFI2); break;
        case 6: str = fmt.Sprintf("~(~%s&%s)", sFI1, sFI2); break;
        case 7: str = fmt.Sprintf("~(%s&~%s)", sFI1, sFI2); break;
        case 8: str = fmt.Sprintf("(~%s&~%s)", sFI1, sFI2); break;
        case 9: str = fmt.Sprintf("~(%s^%s)", sFI1, sFI2); break;
        /*case 0: sprintf( str, " ( %s& %s)", sFI1, sFI2 ); break;
        case 1: sprintf( str, " (~%s& %s)", sFI1, sFI2 ); break;
        case 2: sprintf( str, " ( %s&~%s)", sFI1, sFI2 ); break;
        case 3: sprintf( str, "~(~%s&~%s)", sFI1, sFI2 ); break;
        case 4: sprintf( str, " ( %s^ %s)", sFI1, sFI2 ); break;
        case 5: sprintf( str, "~( %s& %s)", sFI1, sFI2 ); break;
        case 6: sprintf( str, "~(~%s& %s)", sFI1, sFI2 ); break;
        case 7: sprintf( str, "~( %s&~%s)", sFI1, sFI2 ); break;
        case 8: sprintf( str, " (~%s&~%s)", sFI1, sFI2 ); break;
        case 9: sprintf( str, "~( %s^ %s)", sFI1, sFI2 ); break;*/
    }
}

func Dtt_ComposeNP(NP1 int32, NP2 int32) int32 {
    // NP[i] = NP1[NP2[i]]
    var NP, i int32;
    for  i = 0; i < 5; i++  {
        NP |= ( ( NP1 >> ((NP2&0x7)<<2) ) & 0x7 ) << (i<<2); // P'[i] = P1[P2[i]]
        NP |= ( ( NP1 >> ((NP2&0x7)<<2) ^ NP2 ) & 0x8 ) << (i<<2); // N'[i] = N1[P2[i]] ^ N2[i]
        NP2 = NP2 >> 4;
    }
    return NP; 
}

func Dtt_MakePI(NP int32, str string) {
    // apply P'[i], find the i s.t. P'[i]==0, correspond to N'[i]
    var i int32;
    for  i = 0; i < 5; i++  {
        if ( NP & 0x7 ) == 0 {
            if NP & 0x8 { 
                str = fmt.Sprintf("~%c", 'a'+i);
            } else {
                str = fmt.Sprintf("%c", 'a'+i); }
            return;
        }
        NP = NP >> 4;
    }
    assert(false, "0");
}

// void Dtt_MakeFormula( unsigned tFun, Dtt_FunImpl_t* pFun, Vec_Vec_t* vLibImpl, int NP, char* sF, int fPrint, FILE* pFile );

func Dtt_MakeFormulaFI2(tFun uint32, pFun *Dtt_FunImpl_t, vLibImpl *Vec_Vec_t, NP int32, sFI1 string, sF string, fPrint int32, pFile *os.File) {
    var j int32;
    var pImpl2 *Dtt_FunImpl_t;
    var sFI2 [100]byte ; //sprintf( sFI2, "" );

    if pFun.FI2 == 0 { // PI
    
        Dtt_MakePI( Dtt_ComposeNP( pFun.NP2, NP ), sFI2 );
        Dtt_FunImpl2Str( pFun.Type, sFI1, sFI2, sF );
        if fPrint != 0 {
            fmt.Fprintf(pFile, "%08x = %s\n", tFun, sF); }
    
    } else {
        // inlining of macro Vec_VecForEachEntryLevel
        // inlining of macro Vec_PtrForEachEntry
        for  j = 0; (j < Vec_PtrSize(Vec_VecEntry(vLibImpl, pFun.FI2))) && (((pImpl2) = (Dtt_FunImpl_t*)Vec_PtrEntry(Vec_VecEntry(vLibImpl, pFun.FI2), j)), 1); j++  {
            Dtt_MakeFormula( tFun, pImpl2, vLibImpl, Dtt_ComposeNP( pFun.NP2, NP ), sFI2, 0, pFile );
            Dtt_FunImpl2Str( pFun.Type, sFI1, sFI2, sF );
            if fPrint != 0 {
                fmt.Fprintf(pFile, "%08x = %s\n", tFun, sF); }
        }
    }   
}

func Dtt_MakeFormula(tFun uint32, pFun *Dtt_FunImpl_t, vLibImpl *Vec_Vec_t, NP int32, sF string, fPrint int32, pFile *os.File) {
    var j int32;
    var pImpl1 *Dtt_FunImpl_t;
    var sFI1 [100]byte;
    var sCopy string // was [100]byte in C
 //sprintf( sFI1, "" );

    if pFun.FI1 == 0 { // PI
    
        Dtt_MakePI( Dtt_ComposeNP( pFun.NP1, NP ), sFI1 );
        sCopy = fmt.Sprintf("%s", sF);
        Dtt_MakeFormulaFI2( tFun, pFun, vLibImpl, NP, sFI1, sF, fPrint, pFile );
    
    } else {
        // inlining of macro Vec_VecForEachEntryLevel
        // inlining of macro Vec_PtrForEachEntry
        for  j = 0; (j < Vec_PtrSize(Vec_VecEntry(vLibImpl, pFun.FI1))) && (((pImpl1) = (Dtt_FunImpl_t*)Vec_PtrEntry(Vec_VecEntry(vLibImpl, pFun.FI1), j)), 1); j++  {
            Dtt_MakeFormula( tFun, pImpl1, vLibImpl, Dtt_ComposeNP( pFun.NP1, NP ), sFI1, 0, pFile );
            sCopy = fmt.Sprintf("%s", sF);
            Dtt_MakeFormulaFI2( tFun, pFun, vLibImpl, NP, sFI1, sF, fPrint, pFile );
        }
    }
}

func Dtt_ProcessType(Type *int32, nFanin int32) {
    // 0: 1&2, 1: ~1&2, 2: 1&~2, 3: 1|2 = ~(~1&~2), 4: 1^2
    // 5: ~(1&2), 6: ~(~1&2), 7: ~(1&~2), 8: ~1&~2, 9: ~(1^2)
    if nFanin == 3 { // for output negation
        *Type += (*Type<5)? 5: -5;
    } else if *Type == 0 || *Type == 5 {
        *Type += nFanin;
    } else if *Type == nFanin { // 1,1 ; 2,2
        *Type = 0;
    } else if *Type + nFanin == 3 { // 1,2 ; 2,1
        *Type = 8;
    } else if *Type == 3 {
        *Type = (nFanin==1)? 7: 6;
    } else if *Type == 4 {
        *Type = 9;
    } else if *Type == nFanin+5 { // 6,1 ; 7,2
        *Type = 5;
    } else if *Type + nFanin == 8 { // 6,2 ; 7,1
        *Type = 3;
    } else if *Type == 8 {
        *Type = (nFanin==1)? 2: 1;
    } else if *Type == 9 {
        *Type = 4;
    } else {
        assert(false, "0"); }
}

func Dtt_Check(tFun uint32, tGoal uint32, tCur uint32, pType *int32) int32 {
    if !(tGoal != 0) { //for Fanin2 and output
        return ( tCur == tFun || ~tCur == tFun ); }
    //for Fanin1: check if tFun(Type)tCur==tGoal
    switch *pType {
    // 0: 1&2, 1: ~1&2, 2: 1&~2, 3: 1|2 = ~(~1&~2), 4: 1^2
    // 5: ~(1&2), 6: ~(~1&2), 7: ~(1&~2), 8: ~1&~2, 9: ~(1^2)
        case 0,  5: if (~tCur & tFun) == tGoal { Dtt_ProcessType( pType, 1 ); return 1; 
                } else { return ( (tCur & tFun) == tGoal ); }
        fallthrough
case 1,  6: if (tCur & tFun) == tGoal { Dtt_ProcessType( pType, 1 ); return 1; 
                } else { return ( (~tCur & tFun) == tGoal ); }
        fallthrough
case 2,  7: if (~tCur & ~tFun) == tGoal { Dtt_ProcessType( pType, 1 ); return 1; 
                } else { return ( (tCur & ~tFun) == tGoal ); }
        fallthrough
case 3,  8: if (~tCur | tFun) == tGoal { Dtt_ProcessType( pType, 1 ); return 1; 
                } else { return ( (tCur | tFun) == tGoal ); }
        fallthrough
case 4,  9: if (~tCur ^ tFun) == tGoal { Dtt_ProcessType( pType, 1 ); return 1; 
                } else { return ( (tCur ^ tFun) == tGoal ); }
        fallthrough
default: assert(false, "0");
    }
    return -1;
}

func Dtt_FindNP(p *Dtt_Man_t, tFun uint32, tGoal uint32, tNpn uint32, NP *int32, pType *int32, NPout int32) {
    var i, k, j int32;
    var P [5]int32 = [5]int32{ 0, 1, 2, 3, 4 };
    var N [5]int32 = [5]int32{ 0, 0, 0, 0, 0 };
    var temp int32;
    
    var tCur word  = ((word)tNpn << 32) | (word)tNpn;
    for  i = 0; i < p.nPerms; i++  {
        for  k = 0; k < p.nComps; k++  {
            if Dtt_Check( tFun, tGoal, uint32(tCur), pType ) != 0 {
                if !(tGoal != 0) && ( tFun == uint32(~tCur) ) { // !tGoal: not FI1
                
                    if !(NPout != 0) { Dtt_ProcessType( pType, 3 );
                    } else { Dtt_ProcessType( pType, 2 ); }
                }
                *NP = 0;
                if !(NPout != 0) {
                    for  j = 0; j < 5; j++  {
                        *NP |= ( ( ( N[j] & 0x1 ) << 3 ) | ( P[j] & 0x7 ) ) << (j<<2) ;
                    }
                
                } else {
                    for  j = 0; j < 5; j++  {
                        *NP |= P[NPout&0x7] << (j<<2); // P'[j] = P[Pout[j]]
                        *NP |= ( N[NPout&0x7] ^ ( (NPout>>3) & 0x1 )) << (j<<2) << 3; // N'[i] = N1[P2[i]] ^ N2[i]
                        NPout = NPout >> 4;
                    }
                }
                
                return;
            }
            tCur = Abc_Tt6Flip( tCur, p.pComps[k] );
            N[p.pComps[k]] ^= 0x1;
        }
        tCur = Abc_Tt6SwapAdjacent( tCur, p.pPerms[i] );
        temp = P[p.pPerms[i]]; P[p.pPerms[i]] = P[p.pPerms[i]+1]; P[p.pPerms[i]+1] = temp;
    }
    assert(false, "0");
}

func Dtt_DumpLibrary(p *Dtt_Man_t, FileName string) {
    var pFile *os.File;
    var str [100]byte;
    var sFI1 [50]byte;
    var sFI2 [50]byte;
    var i, j, Entry, fRepeat int32;
    var pFun, pFun2 *Dtt_FunImpl_t;
    var vLibFun *Vec_Int_t  = Vec_IntDup( p.vTruthNpns );  // none-duplicating vector of NPN representitives
    var vLibImpl *Vec_Vec_t;
    Vec_IntUniqify( vLibFun );
    vLibImpl = Vec_VecStart( Vec_IntSize( vLibFun ) );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(p.vTruths)) && (((Entry) = Vec_IntEntry(p.vTruths, i)), 1); i++  {
        var NP, Fanin2, Fanin1Npn, Fanin2Npn int32;
        if i<2 { continue; } // skip const 0
        // inlining of macro ABC_CALLOC
        pFun = &Dtt_FunImpl_t{};
        pFun.Type = int32(( 0x7 & Vec_IntEntry(p.vConfigs, i) ));
        //word Fanin1 = Vec_IntEntry( p->vTruths, Vec_IntEntry( p->vFanins, i*2 ) );
        Fanin2 = Vec_IntEntry( p.vTruths, Vec_IntEntry( p.vFanins, i*2+1 ) );
        Fanin1Npn = Vec_IntEntry( p.vTruthNpns, Vec_IntEntry( p.vFanins, i*2 ) );
        Fanin2Npn = Vec_IntEntry( p.vTruthNpns, Vec_IntEntry( p.vFanins, i*2+1 ) );
        pFun.FI1 = Vec_IntFind( vLibFun, Fanin1Npn );
        pFun.FI2 = Vec_IntFind( vLibFun, Fanin2Npn );
        
        fRepeat = 0;
        // inlining of macro Vec_VecForEachEntryLevel
        // inlining of macro Vec_PtrForEachEntry
        for  j = 0; (j < Vec_PtrSize(Vec_VecEntry(vLibImpl, Vec_IntFind( vLibFun, Vec_IntEntry( p.vTruthNpns, i ) )))) && (((pFun2) = (Dtt_FunImpl_t*)Vec_PtrEntry(Vec_VecEntry(vLibImpl, Vec_IntFind( vLibFun, Vec_IntEntry( p.vTruthNpns, i ) )), j)), 1); j++  {
            if ( pFun2.FI1 == pFun.FI1 && pFun2.FI2 == pFun.FI2 ) || ( pFun2.FI2 == pFun.FI1 && pFun2.FI1 == pFun.FI2 ) {
                fRepeat = 1;
                break;
            }
        }
        if fRepeat != 0 {
            pFun = nil;
            continue;
        }

        Dtt_FindNP( p, Vec_IntEntry( p.vTruthNpns, i ), 0, Entry, &NP, &(pFun.Type), 0 ); //out: tGoal=0, NPout=0
        Dtt_FindNP( p, Fanin2, Entry, Fanin1Npn, &(pFun.NP1), &(pFun.Type), NP ); //FI1
        Dtt_FindNP( p, Fanin2, 0, Fanin2Npn, &(pFun.NP2), &(pFun.Type), NP ); //FI2: tGoal=0
        
        Vec_VecPush( vLibImpl, Vec_IntFind( vLibFun, Vec_IntEntry( p.vTruthNpns, i ) ), pFun );
    }

    // print to file
    pFile = os.OpenFile(FileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);

    if false {
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vLibFun)) && (((Entry) = Vec_IntEntry(vLibFun, i)), 1); i++  {
        if !(Entry != 0) { continue; } // skip const 0
        fmt.Fprintf(pFile, "%08x: ", uint32(Entry));
        // inlining of macro Vec_VecForEachEntryLevel
        // inlining of macro Vec_PtrForEachEntry
        for  j = 0; (j < Vec_PtrSize(Vec_VecEntry(vLibImpl, i))) && (((pFun) = (Dtt_FunImpl_t*)Vec_PtrEntry(Vec_VecEntry(vLibImpl, i), j)), 1); j++  {
            Dtt_FunImplFI2Str( pFun.FI1, pFun.NP1, vLibFun, sFI1 ); 
            Dtt_FunImplFI2Str( pFun.FI2, pFun.NP2, vLibFun, sFI2 );
            Dtt_FunImpl2Str( pFun.Type, sFI1, sFI2, str );
            fmt.Fprintf(pFile, "%s, ", str);
        }
        fmt.Fprintf(pFile, "\n");
    } }

    // formula format
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vLibFun)) && (((Entry) = Vec_IntEntry(vLibFun, i)), 1); i++  {
        if i<2 { continue; } // skip const 0 and buffer
        // inlining of macro Vec_VecForEachEntryLevel
        // inlining of macro Vec_PtrForEachEntry
        for  j = 0; (j < Vec_PtrSize(Vec_VecEntry(vLibImpl, i))) && (((pFun) = (Dtt_FunImpl_t*)Vec_PtrEntry(Vec_VecEntry(vLibImpl, i), j)), 1); j++  {
            str[0] = 0; //sprintf( str, "" );
            Dtt_MakeFormula( uint32(Entry), pFun, vLibImpl, (4<<16)+(3<<12)+(2<<8)+(1<<4), str, 1, pFile );
        }
    }

    _ = (pFile).Close();
    fmt.Printf("Dumped file \"%s\". \n", FileName);
    fflush( os.Stdout );
}

func Dtt_EnumerateLf(nVars int32, nNodeMax int32, fDelay int32, fMulti int32, fVerbose int32, pFileName string) {
    var clk abctime  = Abc_Clock(); var nSteps, nMultis word;
    var p *Dtt_Man_t  = Dtt_ManAlloc( nVars, fMulti ); var n, i, j int32;

    // constant zero class
    Vec_IntPushTwo( p.vFanins, 0, 0 );
    Vec_IntPush( p.vTruths, 0 );
    Vec_IntPush( p.vConfigs, 0 );
    Vec_IntPush( p.vClasses, Vec_IntSize(p.vTruthNpns) );
    Vec_IntPush( p.vTruthNpns, 0 );
    Dtt_ManSetFun( p, 0 );
    // buffer class
    Vec_WecPush( p.vFunNodes, 0, Vec_IntSize(p.vFanins)/2 );
    Vec_IntPushTwo( p.vFanins, 0, 0 );
    Vec_IntPush( p.vTruths, uint32(s_Truths6[0]) );
    Vec_IntPush( p.vConfigs, 0 );
    Vec_IntPush( p.vClasses, Vec_IntSize(p.vTruthNpns) );
    Vec_IntPush( p.vTruthNpns, uint32(s_Truths6[0]) );
    for  i = 0; i < nVars; i++  {
        Dtt_ManSetFun( p, uint32(s_Truths6[i]) );
    }
    p.Counts[0] = 2;
    // enumerate
    Dtt_PrintStats(0, nVars, p.vFunNodes, nSteps, clk, fDelay, 0);
    for  n = 1; n <= nNodeMax; n++  {
        for  i, j = 0, n - 1; i < n; i, j = i+1, j - 1 + fDelay  { if i <= j {
            var vFaninI *Vec_Int_t  = Vec_WecEntry( p.vFunNodes, i );
            var vFaninJ *Vec_Int_t  = Vec_WecEntry( p.vFunNodes, j );
            var k, i0, j0, EntryI, EntryJ int32;
            // inlining of macro Vec_IntForEachEntry
            for  i0 = 0; (i0 < Vec_IntSize(vFaninI)) && (((EntryI) = Vec_IntEntry(vFaninI, i0)), 1); i0++  {
                var TruthI uint32  = Vec_IntEntry(p.vTruths, EntryI);
                var vFuns *Vec_Int_t  = Dtt_ManCollect( p, TruthI, p.vTemp );
                var Start int32;
                if i == j {
                    Start = i0;
                } else {
                    Start = 0;
                }
                // inlining of macro Vec_IntForEachEntryStart
                for  j0 = Start; (j0 < Vec_IntSize(vFaninJ)) && (((EntryJ) = Vec_IntEntry(vFaninJ, j0)), 1); j0++  {
                    var Truth uint32;
                    var TruthJ uint32 = Vec_IntEntry(p.vTruths, EntryJ);
                    // inlining of macro Vec_IntForEachEntry
                    for  k = 0; (k < Vec_IntSize(vFuns)) && (((Truth) = Vec_IntEntry(vFuns, k)), 1); k++  {
                        if !(Dtt_ManGetFun(p,  TruthJ &  Truth, n) != 0) {
                            Dtt_ManAddFunction( p, n, EntryI, EntryJ, 0,  TruthJ &  Truth ); }
                        if !(Dtt_ManGetFun(p,  TruthJ & ~Truth, n) != 0) {
                            Dtt_ManAddFunction( p, n, EntryI, EntryJ, 1,  TruthJ & ~Truth ); }
                        if !(Dtt_ManGetFun(p, ~TruthJ &  Truth, n) != 0) {
                            Dtt_ManAddFunction( p, n, EntryI, EntryJ, 2, ~TruthJ &  Truth ); }
                        if !(Dtt_ManGetFun(p,  TruthJ |  Truth, n) != 0) {
                            Dtt_ManAddFunction( p, n, EntryI, EntryJ, 3,  TruthJ |  Truth ); }
                        if !(Dtt_ManGetFun(p,  TruthJ ^  Truth, n) != 0) {
                            Dtt_ManAddFunction( p, n, EntryI, EntryJ, 4,  TruthJ ^  Truth ); }
                        nSteps += 5;

                        if p.pTable { Dtt_ManAddVisited( p,  TruthJ &  Truth, n ); }
                        if p.pTable { Dtt_ManAddVisited( p,  TruthJ & ~Truth, n ); }
                        if p.pTable { Dtt_ManAddVisited( p, ~TruthJ &  Truth, n ); }
                        if p.pTable { Dtt_ManAddVisited( p,  TruthJ |  Truth, n ); }
                        if p.pTable { Dtt_ManAddVisited( p,  TruthJ ^  Truth, n ); }
                    }
                    nMultis++;
                    if p.pTable { Dtt_ManProcessVisited( p ); }
                }
            }
        }
        }
        if Dtt_PrintStats(n, nVars, p.vFunNodes, nSteps, clk, fDelay, nMultis) == 0 {
            break; }
    }
    if fDelay != 0 {
        Dtt_PrintDistrib( p ); }
    //if ( p->pTable ) 
        //Dtt_PrintMulti( p );
    if !(fDelay != 0) && pFileName!=nil {
        Dtt_DumpLibrary( p, pFileName ); }
    Dtt_ManFree( p );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END

