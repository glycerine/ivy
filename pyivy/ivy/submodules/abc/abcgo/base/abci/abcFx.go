package abci

import "fmt"

/**CFile****************************************************************

  FileName    [abcFx.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Network and node package.]

  Synopsis    [Implementation of traditional "fast_extract" algorithm.]

  Author      [Alan Mishchenko, Bruno Schmitt]
  
  Affiliation [UC Berkeley, UFRGS]

  Date        [Ver. 1.0. Started - April 26, 2013.]

  Revision    [$Id: abcFx.c,v 1.00 2013/04/26 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "base/abc/abc.h"
//#include "misc/vec/vecWec.h"
//#include "misc/vec/vecQue.h"
//#include "misc/vec/vecHsh.h"
//#include "opt/fxch/Fxch.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

/*
    The code in this file implements the traditional "fast_extract" algorithm, 
    which extracts two-cube divisors concurrently with single-cube two-literal divisors,
    as proposed in the TCAD'92 paper by J. Rajski and J. Vasudevamurthi.

    Integration notes:

    It is assumed that each object (primary input or internal node) in the original network 
    is associated with a unique integer number, called object identifier (ObjId, for short).

    The user's input data given to 'fast_extract" is an array of cubes (pMan->vCubes).
    Each cube is an array of integers, in which the first entry contains ObjId of the node, 
    to which this cube belongs in the original network. The following entries of a cube are 
    SOP literals of this cube.  Each literal is represtned as 2*FaninId + ComplAttr, where FaninId 
    is ObjId of the fanin node and ComplAttr is 1 if literal is complemented, and 0 otherwise.

    The user's output data produced by 'fast_extract' is also an array of cubes (pMan->vCubes).
    If no divisors have been extracted, the output array is the same as the input array.
    If some divisors have been extracted, the output array contains updated old cubes and new cubes 
    representing the extracted divisors. The new divisors have their ObjId starting from the 
    largest ObjId used in the cubes. To give the user more flexibility, which may be needed when some 
    ObjIds are already used for primary output nodes, which do not participate in fast_extract,
    the parameter ObjIdMax is passed to procedure Fx_FastExtract().  The new divisors will receive
    their ObjId starting from ObjIdMax onward, as divisor extaction proceeds.

    The following two requirements are imposed on the input and output array of cubes:
    (1) The array of cubes should be sorted by the first entry in each cube (that is, cubes belonging 
    to the same node should form a contiguous range). 
    (2) Literals in a cube should be sorted in the increasing order of the integer numbers.
    
    To integrate this code into a calling application, such as ABC, the input cube array should 
    be generated (below this is done by the procedure Abc_NtkFxRetrieve) and the output cube array
    should be incorporated into the current network (below this is done by the procedure Abc_NtkFxInsert).
    In essence, the latter procedure performs the following:
    - removes the current fanins and SOPs of each node in the network
    - adds new nodes for each new divisor introduced by "fast_extract"
    - populates fanins and SOPs of each node, both old and new, as indicaded by the resulting cube array.

    Implementation notes:

    The implementation is optimized for simplicity and speed of computation.
    (1) Main input/output data-structure (pMan->vCubes) is the array of cubes which is dynamically updated by the algorithm.
    (2) Auxiliary data-structure (pMan->vLits) is the array of arrays. The i-th array contains IDs of cubes which have literal i.
    It may be convenient to think about the first (second) array as rows (columns) of a sparse matrix, 
    although the sparse matrix data-structure is not used in the proposed implementation.
    (3) Hash table (pMan->pHash) hashes the normalized divisors (represented as integer arrays) into integer numbers.
    (4) Array of divisor weights (pMan->vWeights), that is, the number of SOP literals to be saved by extacting each divisor.
    (5) Priority queue (pMan->vPrio), which sorts divisor (integer numbers) by their weight
    (6) Integer array (pMan->vVarCube), which maps each ObjId into the first cube of this object, 
    or -1, if there is no cubes as in the case of a primary input.

*/

   
type Fx_Man_t struct {
    // user's data
    vCubes      *Vec_Wec_t;     // cube -> lit
    LitCountMax             int32;// max size of divisor to extract
    fCanonDivs             int32; // use only AND/XOR/MUX
    // internal data
    vLits      *Vec_Wec_t;      // lit -> cube
    vCounts      *Vec_Int_t;    // literal counts (currently not used)
    pHash   *Hsh_VecMan_t;      // hash table for normalized divisors
    vWeights      *Vec_Flt_t;   // divisor weights
    vPrio      *Vec_Que_t;      // priority queue for divisors by weight
    vVarCube      *Vec_Int_t;   // mapping ObjId into its first cube
    vLevels      *Vec_Int_t;    // variable levels
    // temporary data to update the data-structure when a divisor is extracted
    vCubesS      *Vec_Int_t;    // single cubes for the given divisor
    vCubesD      *Vec_Int_t;    // cube pairs for the given divisor
    vCompls      *Vec_Int_t;    // complemented attribute of each cube pair
    vCubeFree      *Vec_Int_t;  // cube-free divisor
    vDiv      *Vec_Int_t;       // selected divisor
    vSCC      *Vec_Int_t;       // single cubes containment cubes
    // statistics 
    timeStart         abctime;  // starting time
    nVars             int32;      // original problem variables
    nLits             int32;      // the number of SOP literals
    nDivs             int32;      // the number of extracted divisors
    nCompls             int32;    // the number of complements
    nPairsS             int32;    // number of lit pairs
    nPairsD             int32;    // number of cube pairs
    nDivsS             int32;     // single cube divisors
    nDivMux [3]int32; // 0 = mux, 1 = compl mux, 2 = no mux
}

func Fx_ManGetFirstVarCube(p *Fx_Man_t, vCube *Vec_Int_t) int32 { return Vec_IntEntry( p.vVarCube, Vec_IntEntry(vCube, 0) ); }

// Original macro:
// #define Fx_ManForEachCubeVec( vVec, vCubes, vCube, i )           \
//     for ( i = 0; (i < Vec_IntSize(vVec)) && ((vCube) = Vec_WecEntry(vCubes, Vec_IntEntry(vVec, i))); i++ )


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Retrieves SOP information for fast_extract.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkFxRetrieve(pNtk *Abc_Ntk_t) *Vec_Wec_t {
    var vCubes *Vec_Wec_t;
    var vCube *Vec_Int_t;
    var pNode *Abc_Obj_t;
    var pCube, pSop string;
    var nVars, i, v, Lit int32;
    assert(Abc_NtkIsSopLogic(pNtk), "Abc_NtkIsSopLogic(pNtk)");
    vCubes = Vec_WecAlloc( 1000 );
    // inlining of macro Abc_NtkForEachNode
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pNode) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pNode) == nil || !Abc_ObjIsNode(pNode) {} else {
        pSop = (char *)pNode.pData;
        nVars = Abc_SopGetVarNum(pSop);
        assert(nVars == Abc_ObjFaninNum(pNode), "nVars == Abc_ObjFaninNum(pNode)");
//        if ( nVars < 2 ) continue;
        // inlining of macro Abc_SopForEachCube
        for  pCube = (pSop); *pCube; pCube += (nVars) + 3  {
            vCube = Vec_WecPushLevel( vCubes );
            Vec_IntPush( vCube, Abc_ObjId(pNode) );
            // inlining of macro Abc_CubeForEachVar
            for  v = 0; (pCube[v] != ' ') && (Lit = pCube[v]); v++  {
                if Lit == '0' {
                    Vec_IntPush( vCube, Abc_Var2Lit(Abc_ObjFaninId(pNode, v), 1) );
                } else if Lit == '1' {
                    Vec_IntPush( vCube, Abc_Var2Lit(Abc_ObjFaninId(pNode, v), 0) ); }
            }
            Vec_IntSelectSort( Vec_IntArray(vCube) + 1, Vec_IntSize(vCube) - 1 );
        }
    }
    }
    return vCubes;
}

/**Function*************************************************************

  Synopsis    [Inserts SOP information after fast_extract.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkFxInsert(pNtk *Abc_Ntk_t, vCubes *Vec_Wec_t) {
    var vCube, vPres, vFirst, vCount *Vec_Int_t;
    var pNode, pFanin *Abc_Obj_t;
    var pCube, pSop string;
    var i, k, v, Lit, iFanin, iNodeMax int32;
    assert(Abc_NtkIsSopLogic(pNtk), "Abc_NtkIsSopLogic(pNtk)");
    // check that cubes have no gaps and are ordered by first node
    Lit = -1;
    // inlining of macro Vec_WecForEachLevel
    for  i = 0; (i < Vec_WecSize(vCubes)) && (((vCube) = Vec_WecEntry(vCubes, i)), 1); i++  {
        assert(Vec_IntSize(vCube) > 0, "Vec_IntSize(vCube) > 0");
        assert(Lit <= Vec_IntEntry(vCube, 0), "Lit <= Vec_IntEntry(vCube, 0)");
        Lit = Vec_IntEntry(vCube, 0);
    }
    // find the largest index
    // inlining of macro Vec_WecForEachLevel
    for  i = 0; (i < Vec_WecSize(vCubes)) && (((vCube) = Vec_WecEntry(vCubes, i)), 1); i++  {
        iNodeMax = Abc_MaxInt( iNodeMax, Vec_IntEntry(vCube, 0) );
    }
    // quit if nothing changes
    if iNodeMax < Abc_NtkObjNumMax(pNtk) {
        fmt.Printf("The network is unchanged by fast extract.\n");
        return;
    }
    // create new nodes
    for  i = Abc_NtkObjNumMax(pNtk); i <= iNodeMax; i++  {
        pNode = Abc_NtkCreateNode( pNtk );
        assert(i == int32(Abc_ObjId(pNode)), "i == int32(Abc_ObjId(pNode))");
    }
    // create node fanins
    vFirst = Vec_IntStart( Abc_NtkObjNumMax(pNtk) );
    vCount = Vec_IntStart( Abc_NtkObjNumMax(pNtk) );
    // inlining of macro Vec_WecForEachLevel
    for  i = 0; (i < Vec_WecSize(vCubes)) && (((vCube) = Vec_WecEntry(vCubes, i)), 1); i++  {
        iFanin = Vec_IntEntry( vCube, 0 );
        if Vec_IntEntry(vCount, iFanin) == 0 {
            Vec_IntWriteEntry( vFirst, iFanin, i ); }
        Vec_IntAddToEntry( vCount, iFanin, 1 );
    }
    // create node SOPs
    vPres = Vec_IntStartFull( Abc_NtkObjNumMax(pNtk) );
    // inlining of macro Abc_NtkForEachNode
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pNode) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pNode) == nil || !Abc_ObjIsNode(pNode) {} else {
//        if ( Vec_IntEntry(vCount, i) == 0 ) continue;
        Abc_ObjRemoveFanins( pNode );
        // create fanins
        assert(Vec_IntEntry(vCount, i) > 0, "Vec_IntEntry(vCount, i) > 0");
        for  k = 0; k < Vec_IntEntry(vCount, i); k++  {
            vCube = Vec_WecEntry( vCubes, Vec_IntEntry(vFirst, i) + k );
            assert(Vec_IntEntry( vCube, 0 ) == i, "Vec_IntEntry( vCube, 0 ) == i");
            // inlining of macro Vec_IntForEachEntryStart
            for  v = 1; (v < Vec_IntSize(vCube)) && (((Lit) = Vec_IntEntry(vCube, v)), 1); v++  {
                pFanin = Abc_NtkObj(pNtk, Abc_Lit2Var(Lit));
                if Vec_IntEntry(vPres, Abc_ObjId(pFanin)) >= 0 {
                    continue; }
                Vec_IntWriteEntry(vPres, Abc_ObjId(pFanin), Abc_ObjFaninNum(pNode));
                Abc_ObjAddFanin( pNode, pFanin );
            }
        }
        // create SOP
        pSop = pCube = Abc_SopStart( (Mem_Flex_t *)pNtk.pManFunc, Vec_IntEntry(vCount, i), Abc_ObjFaninNum(pNode) );
        for  k = 0; k < Vec_IntEntry(vCount, i); k++  {
            vCube = Vec_WecEntry( vCubes, Vec_IntEntry(vFirst, i) + k );
            assert(Vec_IntEntry( vCube, 0 ) == i, "Vec_IntEntry( vCube, 0 ) == i");
            // inlining of macro Vec_IntForEachEntryStart
            for  v = 1; (v < Vec_IntSize(vCube)) && (((Lit) = Vec_IntEntry(vCube, v)), 1); v++  {
                pFanin = Abc_NtkObj(pNtk, Abc_Lit2Var(Lit));
                iFanin = Vec_IntEntry(vPres, Abc_ObjId(pFanin));
                assert(iFanin >= 0 && iFanin < Abc_ObjFaninNum(pNode), "iFanin >= 0 && iFanin < Abc_ObjFaninNum(pNode)");
                pCube[iFanin] = Abc_LitIsCompl(Lit) ? '0' : '1';
            }
            pCube += Abc_ObjFaninNum(pNode) + 3;
        }
        // complement SOP if the original one was complemented
        if pNode.pData && Abc_SopIsComplement((char *)pNode.pData) {
            Abc_SopComplement( pSop ); }
        pNode.pData = pSop;
        // clean fanins
        // inlining of macro Abc_ObjForEachFanin
        for  v = 0; (v < Abc_ObjFaninNum(pNode)) && (((pFanin) = Abc_ObjFanin(pNode, v)), 1); v++  {
            Vec_IntWriteEntry( vPres, Abc_ObjId(pFanin), -1 );
        }
    }
    }
    Vec_IntFree( vFirst );
    Vec_IntFree( vCount );
    Vec_IntFree( vPres );
}

/**Function*************************************************************

  Synopsis    [Makes sure the nodes do not have complemented and duplicated fanins.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkFxCheck(pNtk *Abc_Ntk_t) int32 {
    var pNode *Abc_Obj_t;
    var i int32;
//    Abc_NtkForEachObj( pNtk, pNode, i )
//        Abc_ObjPrint( stdout, pNode );
    // inlining of macro Abc_NtkForEachNode
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pNode) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pNode) == nil || !Abc_ObjIsNode(pNode) {} else
        if !Vec_IntCheckUniqueSmall( &pNode.vFanins ) {
            return 0; }
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkFxPerform(pNtk *Abc_Ntk_t, nNewNodesMax int32, LitCountMax int32, fCanonDivs int32, fVerbose int32, fVeryVerbose int32) int32 {
    extern int Fx_FastExtract( Vec_Wec_t * vCubes, int ObjIdMax, int nNewNodesMax, int LitCountMax, int fCanonDivs, int fVerbose, int fVeryVerbose );
    var vCubes *Vec_Wec_t;
    assert(Abc_NtkIsSopLogic(pNtk), "Abc_NtkIsSopLogic(pNtk)");
    // check unique fanins
    if !(Abc_NtkFxCheck(pNtk) != 0) {
        fmt.Printf("Abc_NtkFastExtract: Nodes have duplicated fanins. FX is not performed.\n");
        return 0;
    }
    // collect information about the covers
    vCubes = Abc_NtkFxRetrieve( pNtk );
    // call the fast extract procedure
    if Fx_FastExtract( vCubes, Abc_NtkObjNumMax(pNtk), nNewNodesMax, LitCountMax, fCanonDivs, fVerbose, fVeryVerbose ) > 0 {
        // update the network
        Abc_NtkFxInsert( pNtk, vCubes );
        Vec_WecFree( vCubes );
        if !Abc_NtkCheck( pNtk ) {
            fmt.Printf("Abc_NtkFxPerform: The network check has failed.\n"); }
        return 1;
    
    } else {
        fmt.Printf("Warning: The network has not been changed by \"fx\".\n"); }
    Vec_WecFree( vCubes );
    return 0;
}



/**Function*************************************************************

  Synopsis    [Starting the manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fx_ManStart(vCubes *Vec_Wec_t) *Fx_Man_t {
    
    // inlining of macro ABC_CALLOC
    p := &Fx_Man_t{};
    p.vCubes   = vCubes;
    // temporary data
    p.vCubesS   = Vec_IntAlloc( 100 );
    p.vCubesD   = Vec_IntAlloc( 100 );
    p.vCompls   = Vec_IntAlloc( 100 );
    p.vCubeFree = Vec_IntAlloc( 100 );
    p.vDiv      = Vec_IntAlloc( 100 );
    p.vSCC      = Vec_IntAlloc( 100 );
    return p;
}
func Fx_ManStop(p *Fx_Man_t) {
//    Vec_WecFree( p->vCubes );
    Vec_WecFree( p.vLits );
    Vec_IntFree( p.vCounts );
    Hsh_VecManStop( p.pHash );
    Vec_FltFree( p.vWeights );
    Vec_QueFree( p.vPrio );
    Vec_IntFree( p.vVarCube );
    Vec_IntFree( p.vLevels );
    // temporary data
    Vec_IntFree( p.vCubesS );
    Vec_IntFree( p.vCubesD );
    Vec_IntFree( p.vCompls );
    Vec_IntFree( p.vCubeFree );
    Vec_IntFree( p.vDiv );
    Vec_IntFree( p.vSCC );
    p = nil;
}

/**Function*************************************************************

  Synopsis    [Compute levels of the nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fx_ManComputeLevelDiv(p *Fx_Man_t, vCubeFree *Vec_Int_t) int32 {
    var i, Lit, Level int32;
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vCubeFree)) && (((Lit) = Vec_IntEntry(vCubeFree, i)), 1); i++  {
        Level = Abc_MaxInt( Level, Vec_IntEntry(p.vLevels, Abc_Lit2Var(Abc_Lit2Var(Lit))) );
    }
    return Abc_MinInt( Level, 800 );
}
func Fx_ManComputeLevelCube(p *Fx_Man_t, vCube *Vec_Int_t) int32 {
    var k, Lit, Level int32;
    // inlining of macro Vec_IntForEachEntryStart
    for  k = 1; (k < Vec_IntSize(vCube)) && (((Lit) = Vec_IntEntry(vCube, k)), 1); k++  {
        Level = Abc_MaxInt( Level, Vec_IntEntry(p.vLevels, Abc_Lit2Var(Lit)) );
    }
    return Level;
}
func Fx_ManComputeLevel(p *Fx_Man_t) {
    var vCube *Vec_Int_t;
    var i, iVar, iFirst int32;
    iVar = Vec_IntEntry( Vec_WecEntry(p.vCubes,0), 0 );
    p.vLevels = Vec_IntStart( p.nVars );
    // inlining of macro Vec_WecForEachLevel
    for  i = 0; (i < Vec_WecSize(p.vCubes)) && (((vCube) = Vec_WecEntry(p.vCubes, i)), 1); i++  {
        if iVar != Vec_IntEntry(vCube, 0) {
            // add the number of cubes
            Vec_IntAddToEntry( p.vLevels, iVar, i - iFirst );
            iVar = Vec_IntEntry(vCube, 0);
            iFirst = i;
        }
        Vec_IntUpdateEntry( p.vLevels, iVar, Fx_ManComputeLevelCube(p, vCube) );
    }
}

/**Function*************************************************************

  Synopsis    [Printing procedures.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fx_PrintDivArray(vDiv *Vec_Int_t) {
    var i, Lit int32;
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vDiv)) && (((Lit) = Vec_IntEntry(vDiv, i)), 1); i++  {
        if !Abc_LitIsCompl(Lit) {
            fmt.Printf("%d(1)", Abc_Lit2Var(Lit)); }
    }
    fmt.Printf(" + ");
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vDiv)) && (((Lit) = Vec_IntEntry(vDiv, i)), 1); i++  {
        if Abc_LitIsCompl(Lit) {
            fmt.Printf("%d(2)", Abc_Lit2Var(Lit)); }
    }
}
func Fx_PrintDiv(p *Fx_Man_t, iDiv int32) {
    var i int32;
    fmt.Printf("%4d : ", p.nDivs);
    fmt.Printf("Div %7d : ", iDiv);
    fmt.Printf("Weight %12.5f  ", Vec_FltEntry(p.vWeights, iDiv));
    Fx_PrintDivArray( Hsh_VecReadEntry(p.pHash, iDiv) );
    for  i = Vec_IntSize(Hsh_VecReadEntry(p.pHash, iDiv)) + 3; i < 16; i++  {
        fmt.Printf(" ");
    }
    fmt.Printf("Lits =%7d  ", p.nLits);
    fmt.Printf("Divs =%8d  ", Hsh_VecSize(p.pHash));
    Abc_PrintTime( 1, "Time", Abc_Clock() - p.timeStart );
}
func Fx_PrintDivisors(p *Fx_Man_t) {
    var iDiv int32;
    for  iDiv = 0; iDiv < Vec_FltSize(p.vWeights); iDiv++  {
        Fx_PrintDiv( p, iDiv );
    }
}
func Fx_PrintStats(p *Fx_Man_t, clk abctime) {
    fmt.Printf("Cubes =%8d  ", Vec_WecSizeUsed(p.vCubes));
    fmt.Printf("Lits  =%8d  ", Vec_WecSizeUsed(p.vLits));
    fmt.Printf("Divs  =%8d  ", Hsh_VecSize(p.pHash));
    fmt.Printf("Divs+ =%8d  ", Vec_QueSize(p.vPrio));
    fmt.Printf("Compl =%8d  ", p.nDivMux[1]);
    fmt.Printf("Extr  =%7d  ", p.nDivs);
    Abc_PrintTime( 1, "Time", clk );
}

/**Function*************************************************************

  Synopsis    [Returns 1 if the divisor should be complemented.]

  Description [Normalizes the divisor by putting, first, positive control 
  literal first and, second, positive data1 literal. As the result, 
  a MUX divisor is (ab + !ac) and an XOR divisor is (ab + !a!b).]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
static int Fx_ManDivNormalize( Vec_Int_t * vCubeFree ) // return 1 if complemented
{
    var L *int32  = Vec_IntArray(vCubeFree);
    var RetValue int32;
    var LitA0 int32 = -1;
    var LitB0 int32 = -1;
    var LitA1 int32 = -1;
    var LitB1 int32 = -1;
    assert(Vec_IntSize(vCubeFree) == 4, "Vec_IntSize(vCubeFree) == 4");
    if Abc_LitIsCompl(L[0]) != Abc_LitIsCompl(L[1]) && (L[0] >> 2) == (L[1] >> 2) { // diff cubes, same vars
    
        if Abc_LitIsCompl(L[2]) == Abc_LitIsCompl(L[3]) {
            return -1; }
        LitA0 = Abc_Lit2Var(L[0]);
LitB0 = Abc_Lit2Var(L[1]);
        if Abc_LitIsCompl(L[0]) == Abc_LitIsCompl(L[2]) {
            assert(Abc_LitIsCompl(L[1]) == Abc_LitIsCompl(L[3]), "Abc_LitIsCompl(L[1]) == Abc_LitIsCompl(L[3])");
            LitA1 = Abc_Lit2Var(L[2]);
LitB1 = Abc_Lit2Var(L[3]);
        
        } else {
            assert(Abc_LitIsCompl(L[0]) == Abc_LitIsCompl(L[3]), "Abc_LitIsCompl(L[0]) == Abc_LitIsCompl(L[3])");
            assert(Abc_LitIsCompl(L[1]) == Abc_LitIsCompl(L[2]), "Abc_LitIsCompl(L[1]) == Abc_LitIsCompl(L[2])");
            LitA1 = Abc_Lit2Var(L[3]);
LitB1 = Abc_Lit2Var(L[2]);
        }
    
    } else if Abc_LitIsCompl(L[1]) != Abc_LitIsCompl(L[2]) && (L[1] >> 2) == (L[2] >> 2) {
        if Abc_LitIsCompl(L[0]) == Abc_LitIsCompl(L[3]) {
            return -1; }
        LitA0 = Abc_Lit2Var(L[1]);
LitB0 = Abc_Lit2Var(L[2]);
        if Abc_LitIsCompl(L[1]) == Abc_LitIsCompl(L[0]) {
            LitA1 = Abc_Lit2Var(L[0]);
LitB1 = Abc_Lit2Var(L[3]);
        } else {
            LitA1 = Abc_Lit2Var(L[3]);
LitB1 = Abc_Lit2Var(L[0]); }
    
    } else if Abc_LitIsCompl(L[2]) != Abc_LitIsCompl(L[3]) && (L[2] >> 2) == (L[3] >> 2) {
        if Abc_LitIsCompl(L[0]) == Abc_LitIsCompl(L[1]) {
            return -1; }
        LitA0 = Abc_Lit2Var(L[2]);
LitB0 = Abc_Lit2Var(L[3]);
        if Abc_LitIsCompl(L[2]) == Abc_LitIsCompl(L[0]) {
            LitA1 = Abc_Lit2Var(L[0]);
LitB1 = Abc_Lit2Var(L[1]);
        } else {
            LitA1 = Abc_Lit2Var(L[1]);
LitB1 = Abc_Lit2Var(L[0]); }
    
    } else { 
        return -1; }
    assert(LitA0 == Abc_LitNot(LitB0), "LitA0 == Abc_LitNot(LitB0)");
    if Abc_LitIsCompl(LitA0) {
        // inlining of macro ABC_SWAP
        { var t int32  = LitA0; LitA0 = LitB0; LitB0 = t; };
        // inlining of macro ABC_SWAP
        { var t int32  = LitA1; LitA1 = LitB1; LitB1 = t; };
    }
    assert(!Abc_LitIsCompl(LitA0), "!Abc_LitIsCompl(LitA0)");
    if Abc_LitIsCompl(LitA1) {
        LitA1 = Abc_LitNot(LitA1);
        LitB1 = Abc_LitNot(LitB1);
        RetValue = 1;
    }
    assert(!Abc_LitIsCompl(LitA1), "!Abc_LitIsCompl(LitA1)");
    // arrange literals in such as a way that
    // - the first two literals are control literals from different cubes
    // - the third literal is non-complented data input
    // - the forth literal is possibly complemented data input
    L[0] = Abc_Var2Lit( LitA0, 0 );
    L[1] = Abc_Var2Lit( LitB0, 1 );
    L[2] = Abc_Var2Lit( LitA1, 0 );
    L[3] = Abc_Var2Lit( LitB1, 1 );
    return RetValue;
}

/**Function*************************************************************

  Synopsis    [Find a cube-free divisor of the two cubes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fx_ManDivFindCubeFree(vArr1 *Vec_Int_t, vArr2 *Vec_Int_t, vCubeFree *Vec_Int_t, fWarning *int32) int32 {
    var pBeg1 *int32  = Vec_IntArray( vArr1 ) + 1;  // skip variable ID
    var pBeg2 *int32  = Vec_IntArray( vArr2 ) + 1;  // skip variable ID
    var pEnd1 *int32  = Vec_IntLimit( vArr1 );
    var pEnd2 *int32  = Vec_IntLimit( vArr2 );
    var Counter int32;
    var fAttr0 int32;
    var fAttr1 int32 = 1;
    Vec_IntClear( vCubeFree );
    for  pBeg1 < pEnd1 && pBeg2 < pEnd2  {
        if *pBeg1 == *pBeg2 {
            pBeg1++, pBeg2++, Counter++;
        } else if *pBeg1 < *pBeg2 {
            Vec_IntPush( vCubeFree, Abc_Var2Lit(*pBeg1++, fAttr0) );
        } else {
            if Vec_IntSize(vCubeFree) == 0 {
                fAttr0 = 1;
fAttr1 = 0; }
            Vec_IntPush( vCubeFree, Abc_Var2Lit(*pBeg2++, fAttr1) );
        }
    }
    for  pBeg1 < pEnd1  {
        Vec_IntPush( vCubeFree, Abc_Var2Lit(*pBeg1++, fAttr0) );
    }
    for  pBeg2 < pEnd2  {
        Vec_IntPush( vCubeFree, Abc_Var2Lit(*pBeg2++, fAttr1) );
    }
    if Vec_IntSize(vCubeFree) == 0 {
        fmt.Printf("The SOP has duplicated cubes.\n");
    } else if Vec_IntSize(vCubeFree) == 1 {
        return -1;
    } else if Vec_IntSize( vCubeFree ) == 3 {
        var pArray *int32  = Vec_IntArray( vCubeFree ); 

        if Abc_Lit2Var( pArray[0] ) == Abc_LitNot( Abc_Lit2Var( pArray[1] ) ) {
            if Abc_LitIsCompl( pArray[0] ) == Abc_LitIsCompl( pArray[2] ) {
                Vec_IntDrop( vCubeFree, 0 );
            } else {
                Vec_IntDrop( vCubeFree, 1 ); }
        
        } else if Abc_Lit2Var( pArray[1] ) == Abc_LitNot( Abc_Lit2Var( pArray[2] ) ) {
            if Abc_LitIsCompl( pArray[1] ) == Abc_LitIsCompl( pArray[0] ) {
                Vec_IntDrop( vCubeFree, 1 );
            } else {
                Vec_IntDrop( vCubeFree, 2 ); }
        }

        if Vec_IntSize( vCubeFree ) == 2 {
            var Lit0 int32 = Abc_Lit2Var( pArray[0] );
            var Lit1 int32 = Abc_Lit2Var( pArray[1] );

            if Lit0 > Lit1 {
                // inlining of macro ABC_SWAP
                 var t int32  = Lit0; Lit0 = Lit1; Lit1 = t; };

            Vec_IntWriteEntry( vCubeFree, 0, Abc_Var2Lit( Lit0, 0 ) );
            Vec_IntWriteEntry( vCubeFree, 1, Abc_Var2Lit( Lit1, 1 ) );
        }
    }
    assert(!Abc_LitIsCompl(Vec_IntEntry(vCubeFree, 0)), "!Abc_LitIsCompl(Vec_IntEntry(vCubeFree, 0))");
    return Counter;
}

/**Function*************************************************************

  Synopsis    [Procedures operating on a two-cube divisor.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fx_ManDivFindPivots(vDiv *Vec_Int_t, pLit0 *int32, pLit1 *int32) {
    var i, Lit int32;
    *pLit0 = -1;
    *pLit1 = -1;
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vDiv)) && (((Lit) = Vec_IntEntry(vDiv, i)), 1); i++  {
        if Abc_LitIsCompl(Lit) {
            if *pLit1 == -1 {
                *pLit1 = Abc_Lit2Var(Lit); }
        
        } else {
            if *pLit0 == -1 {
                *pLit0 = Abc_Lit2Var(Lit); }
        }
        if *pLit0 >= 0 && *pLit1 >= 0 {
            return; }
    }
}
func Fx_ManDivRemoveLits(vCube *Vec_Int_t, vDiv *Vec_Int_t, fCompl int32) int32 {
    var i, Lit, Count int32;
    assert(!(fCompl != 0) || Vec_IntSize(vDiv) == 4, "!fCompl || Vec_IntSize(vDiv) == 4");
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vDiv)) && (((Lit) = Vec_IntEntry(vDiv, i)), 1); i++  {
        Count += Vec_IntRemove1( vCube, Abc_Lit2Var(Lit) ^ (fCompl && i > 1) );  // the last two lits can be complemented
        if Vec_IntSize( vDiv ) == 2 {
            Count += Vec_IntRemove1( vCube, Abc_LitNot( Abc_Lit2Var(Lit) ) ); }
    }
    return Count;
}
func Fx_ManDivAddLits(vCube *Vec_Int_t, vCube2 *Vec_Int_t, vDiv *Vec_Int_t) {
    var i int32;
    var Lit int32;
    var pArray *int32;
//    Vec_IntClear( vCube );
//    Vec_IntClear( vCube2 );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vDiv)) && (((Lit) = Vec_IntEntry(vDiv, i)), 1); i++  {
        if Abc_LitIsCompl(Lit) {
            Vec_IntPush( vCube2, Abc_Lit2Var(Lit) );
        } else {
            Vec_IntPush( vCube, Abc_Lit2Var(Lit) ); }
    }
    if Vec_IntSize(vDiv) == 4 && Vec_IntSize(vCube) == 3 {
        assert(Vec_IntSize(vCube2) == 3, "Vec_IntSize(vCube2) == 3");
        pArray = Vec_IntArray(vCube);
        if pArray[1] > pArray[2] {
            // inlining of macro ABC_SWAP
             var t int32  = pArray[1]; pArray[1] = pArray[2]; pArray[2] = t; };
        pArray = Vec_IntArray(vCube2);
        if pArray[1] > pArray[2] {
            // inlining of macro ABC_SWAP
             var t int32  = pArray[1]; pArray[1] = pArray[2]; pArray[2] = t; };
    }
}

/**Function*************************************************************

  Synopsis    [Setting up the data-structure.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fx_ManCreateLiterals(p *Fx_Man_t, nVars int32) {
    var vCube *Vec_Int_t;
    var i, k, Lit, Count int32;
    // find the number of variables
    p.nVars = p.nLits = 0;
    // inlining of macro Vec_WecForEachLevel
    for  i = 0; (i < Vec_WecSize(p.vCubes)) && (((vCube) = Vec_WecEntry(p.vCubes, i)), 1); i++  {
        assert(Vec_IntSize(vCube) > 0, "Vec_IntSize(vCube) > 0");
        p.nVars = Abc_MaxInt( p.nVars, Vec_IntEntry(vCube, 0) );
        p.nLits += Vec_IntSize(vCube) - 1;
        // inlining of macro Vec_IntForEachEntryStart
        for  k = 1; (k < Vec_IntSize(vCube)) && (((Lit) = Vec_IntEntry(vCube, k)), 1); k++  {
            p.nVars = Abc_MaxInt( p.nVars, Abc_Lit2Var(Lit) );
        }
    }
//    p->nVars++;
    assert(p.nVars < nVars, "p.nVars < nVars");
    p.nVars = nVars;
    // count literals
    p.vCounts = Vec_IntStart( 2*p.nVars );
    // inlining of macro Vec_WecForEachLevel
    for  i = 0; (i < Vec_WecSize(p.vCubes)) && (((vCube) = Vec_WecEntry(p.vCubes, i)), 1); i++  {
        // inlining of macro Vec_IntForEachEntryStart
        for  k = 1; (k < Vec_IntSize(vCube)) && (((Lit) = Vec_IntEntry(vCube, k)), 1); k++  {
            Vec_IntAddToEntry( p.vCounts, Lit, 1 );
        }
    }
    // start literals
    p.vLits = Vec_WecStart( 2*p.nVars );
    // inlining of macro Vec_IntForEachEntry
    for  Lit = 0; (Lit < Vec_IntSize(p.vCounts)) && (((Count) = Vec_IntEntry(p.vCounts, Lit)), 1); Lit++  {
        Vec_IntGrow( Vec_WecEntry(p.vLits, Lit), Count );
    }
    // fill out literals
    // inlining of macro Vec_WecForEachLevel
    for  i = 0; (i < Vec_WecSize(p.vCubes)) && (((vCube) = Vec_WecEntry(p.vCubes, i)), 1); i++  {
        // inlining of macro Vec_IntForEachEntryStart
        for  k = 1; (k < Vec_IntSize(vCube)) && (((Lit) = Vec_IntEntry(vCube, k)), 1); k++  {
            Vec_WecPush( p.vLits, Lit, i );
        }
    }
    // create mapping of variable into the first cube
    p.vVarCube = Vec_IntStartFull( p.nVars );
    // inlining of macro Vec_WecForEachLevel
    for  i = 0; (i < Vec_WecSize(p.vCubes)) && (((vCube) = Vec_WecEntry(p.vCubes, i)), 1); i++  {
        if Vec_IntEntry(p.vVarCube, Vec_IntEntry(vCube, 0)) == -1 {
            Vec_IntWriteEntry( p.vVarCube, Vec_IntEntry(vCube, 0), i ); }
    }
}
func Fx_ManCubeSingleCubeDivisors(p *Fx_Man_t, vPivot *Vec_Int_t, fRemove int32, fUpdate int32) int32 {
    var k, n, Lit, Lit2, iDiv int32;
    if Vec_IntSize(vPivot) < 2 {
        return 0; }
    // inlining of macro Vec_IntForEachEntryStart
    for  k = 1; (k < Vec_IntSize(vPivot)) && (((Lit) = Vec_IntEntry(vPivot, k)), 1); k++  {
    // inlining of macro Vec_IntForEachEntryStart
    for  n = k+1; (n < Vec_IntSize(vPivot)) && (((Lit2) = Vec_IntEntry(vPivot, n)), 1); n++  {
        assert(Lit < Lit2, "Lit < Lit2");
        Vec_IntClear( p.vCubeFree );
        Vec_IntPush( p.vCubeFree, Abc_Var2Lit(Abc_LitNot(Lit), 0) );
        Vec_IntPush( p.vCubeFree, Abc_Var2Lit(Abc_LitNot(Lit2), 1) );
        iDiv = Hsh_VecManAdd( p.pHash, p.vCubeFree );
        if !(fRemove != 0) {
            if Vec_FltSize(p.vWeights) == iDiv {
                Vec_FltPush(p.vWeights, -2 + 0.9 - 0.001 * Fx_ManComputeLevelDiv(p, p.vCubeFree));
                p.nDivsS++;
            }
            assert(iDiv < Vec_FltSize(p.vWeights), "iDiv < Vec_FltSize(p.vWeights)");
            Vec_FltAddToEntry( p.vWeights, iDiv, 1 );
            p.nPairsS++;
        
        } else {
            assert(iDiv < Vec_FltSize(p.vWeights), "iDiv < Vec_FltSize(p.vWeights)");
            Vec_FltAddToEntry( p.vWeights, iDiv, -1 );
            p.nPairsS--;
        }
        if fUpdate != 0 {
            if Vec_QueIsMember(p.vPrio, iDiv) {
                Vec_QueUpdate( p.vPrio, iDiv );
            } else if !(fRemove != 0) {
                Vec_QuePush( p.vPrio, iDiv ); }
        }
    }
    }
    return Vec_IntSize(vPivot) * (Vec_IntSize(vPivot) - 1) / 2;
}
func Fx_ManCubeDoubleCubeDivisors(p *Fx_Man_t, iFirst int32, vPivot *Vec_Int_t, fRemove int32, fUpdate int32, fWarning *int32) {
    var vCube *Vec_Int_t;
    var i, iDiv, Base int32;
    // inlining of macro Vec_WecForEachLevelStart
    for  i = iFirst; (i < Vec_WecSize(p.vCubes)) && (((vCube) = Vec_WecEntry(p.vCubes, i)), 1); i++  {
        if Vec_IntSize(vCube) == 0 || vCube == vPivot {
            continue; }
        if Vec_WecIntHasMark(vCube) && Vec_WecIntHasMark(vPivot) && vCube > vPivot {
            continue; }
        if Vec_IntEntry(vCube, 0) != Vec_IntEntry(vPivot, 0) {
            break; }
        Base = Fx_ManDivFindCubeFree( vCube, vPivot, p.vCubeFree, fWarning );
        if Base == -1 {
            if fRemove == 0 {
                if Vec_IntSize( vCube ) > Vec_IntSize( vPivot ) {
                    Vec_IntPush( p.vSCC, Vec_WecLevelId( p.vCubes, vCube ) );
                } else {
                    Vec_IntPush( p.vSCC, Vec_WecLevelId( p.vCubes, vPivot ) ); }
            }
            continue;
        }
        if Vec_IntSize(p.vCubeFree) == 4 { 
            var Value int32  = Fx_ManDivNormalize( p.vCubeFree );
            if Value == 0 {
                p.nDivMux[0]++;
            } else if Value == 1 {
                p.nDivMux[1]++;
            } else {
                p.nDivMux[2]++; }
            if p.fCanonDivs && Value < 0 {
                continue; }
        }
        if p.LitCountMax && p.LitCountMax < Vec_IntSize(p.vCubeFree) {
            continue; }
        if p.fCanonDivs && Vec_IntSize(p.vCubeFree) == 3 {
            continue; }
        iDiv = Hsh_VecManAdd( p.pHash, p.vCubeFree );
        if !(fRemove != 0) {
            if iDiv == Vec_FltSize(p.vWeights) {
                Vec_FltPush(p.vWeights, -Vec_IntSize(p.vCubeFree) + 0.9 - 0.0009 * Fx_ManComputeLevelDiv(p, p.vCubeFree)); }
            assert(iDiv < Vec_FltSize(p.vWeights), "iDiv < Vec_FltSize(p.vWeights)");
            Vec_FltAddToEntry( p.vWeights, iDiv, Base + Vec_IntSize(p.vCubeFree) - 1 );
            p.nPairsD++;
        
        } else {
            assert(iDiv < Vec_FltSize(p.vWeights), "iDiv < Vec_FltSize(p.vWeights)");
            Vec_FltAddToEntry( p.vWeights, iDiv, -(Base + Vec_IntSize(p.vCubeFree) - 1) );
            p.nPairsD--;
        }
        if fUpdate != 0 {
            if Vec_QueIsMember(p.vPrio, iDiv) {
                Vec_QueUpdate( p.vPrio, iDiv );
            } else if !(fRemove != 0) {
                Vec_QuePush( p.vPrio, iDiv ); }
        }
    } 
}
func Fx_ManCreateDivisors(p *Fx_Man_t) {
    var vCube *Vec_Int_t;
    var Weight float32;
    var i, fWarning int32;
    // alloc hash table
    assert(p.pHash == nil, "p.pHash == nil");
    p.pHash = Hsh_VecManStart( 1000 );
    p.vWeights = Vec_FltAlloc( 1000 );
    // create single-cube two-literal divisors
    // inlining of macro Vec_WecForEachLevel
    for  i = 0; (i < Vec_WecSize(p.vCubes)) && (((vCube) = Vec_WecEntry(p.vCubes, i)), 1); i++  {
        Fx_ManCubeSingleCubeDivisors( p, vCube, 0, 0 );
    } // add - no update
    assert(p.nDivsS == Vec_FltSize(p.vWeights), "p.nDivsS == Vec_FltSize(p.vWeights)");
    // create two-cube divisors
    // inlining of macro Vec_WecForEachLevel
    for  i = 0; (i < Vec_WecSize(p.vCubes)) && (((vCube) = Vec_WecEntry(p.vCubes, i)), 1); i++  {
        Fx_ManCubeDoubleCubeDivisors( p, i+1, vCube, 0, 0, &fWarning );
    } // add - no update
    // create queue with all divisors
    p.vPrio = Vec_QueAlloc( Vec_FltSize(p.vWeights) );
    Vec_QueSetPriority( p.vPrio, Vec_FltArrayP(p.vWeights) );
    // inlining of macro Vec_FltForEachEntry
    for  i = 0; (i < Vec_FltSize(p.vWeights)) && (((Weight) = Vec_FltEntry(p.vWeights, i)), 1); i++  {
        if Weight > 0.0 {
            Vec_QuePush( p.vPrio, i ); }
    }
}


/**Function*************************************************************

  Synopsis    [Compress the cubes by removing unused ones.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fx_ManCompressCubes(vCubes *Vec_Wec_t, vLit2Cube *Vec_Int_t) {
    var i, CubeId, k int32;
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vLit2Cube)) && (((CubeId) = Vec_IntEntry(vLit2Cube, i)), 1); i++  {
        if Vec_IntSize(Vec_WecEntry(vCubes, CubeId)) > 0 {
            Vec_IntWriteEntry( vLit2Cube, k++, CubeId ); }
    }
    Vec_IntShrink( vLit2Cube, k );
}


/**Function*************************************************************

  Synopsis    [Find command cube pairs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fx_ManGetCubeVar(vCubes *Vec_Wec_t, iCube int32) int32 { return Vec_IntEntry( Vec_WecEntry(vCubes, iCube), 0 );      }
func Fx_ManFindCommonPairs(vCubes *Vec_Wec_t, vPart0 *Vec_Int_t, vPart1 *Vec_Int_t, vPairs *Vec_Int_t, vCompls *Vec_Int_t, vDiv *Vec_Int_t, vCubeFree *Vec_Int_t, fWarning *int32) {
    var pBeg1 *int32  = vPart0.pArray;
    var pBeg2 *int32  = vPart1.pArray;
    var pEnd1 *int32  = vPart0.pArray + vPart0.nSize;
    var pEnd2 *int32  = vPart1.pArray + vPart1.nSize;
    var i, k, i_, k_, fCompl, CubeId1, CubeId2 int32;
    Vec_IntClear( vPairs );
    Vec_IntClear( vCompls );
    for  pBeg1 < pEnd1 && pBeg2 < pEnd2  {
        CubeId1 = Fx_ManGetCubeVar(vCubes, *pBeg1);
        CubeId2 = Fx_ManGetCubeVar(vCubes, *pBeg2);
        if CubeId1 == CubeId2 {
            for  i = 1; pBeg1+i < pEnd1; i++  {
                if CubeId1 != Fx_ManGetCubeVar(vCubes, pBeg1[i]) {
                    break; }
            }
            for  k = 1; pBeg2+k < pEnd2; k++  {
                if CubeId1 != Fx_ManGetCubeVar(vCubes, pBeg2[k]) {
                    break; }
            }
            for  i_ = 0; i_ < i; i_++  {
            for  k_ = 0; k_ < k; k_++  {
                if pBeg1[i_] == pBeg2[k_] {
                    continue; }
                Fx_ManDivFindCubeFree( Vec_WecEntry(vCubes, pBeg1[i_]), Vec_WecEntry(vCubes, pBeg2[k_]), vCubeFree, fWarning );
                fCompl = (Vec_IntSize(vCubeFree) == 4 && Fx_ManDivNormalize(vCubeFree) == 1);
                if !Vec_IntEqual( vDiv, vCubeFree ) {
                    continue; }
                Vec_IntPush( vPairs, pBeg1[i_] );
                Vec_IntPush( vPairs, pBeg2[k_] );
                Vec_IntPush( vCompls, fCompl );
            }
            }
            pBeg1 += i;
            pBeg2 += k;
        
        } else if CubeId1 < CubeId2 {
            pBeg1++;
        } else { 
            pBeg2++; }
    }
}

/**Function*************************************************************

  Synopsis    [Updates the data-structure when one divisor is selected.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fx_ManUpdate(p *Fx_Man_t, iDiv int32, fWarning *int32) {
    var vCube *Vec_Int_t;
    var vCube2 *Vec_Int_t;
    var vLitP *Vec_Int_t = nil;
    var vLitN *Vec_Int_t = nil;
    var vDiv *Vec_Int_t  = p.vDiv;
    var i, k, Lit0, Lit1, iVarNew, RetValue, Level int32;
    var Diff float32  = Vec_FltEntry(p.vWeights, iDiv) - float32((int32(Vec_FltEntry(p.vWeights, iDiv))));
    assert(Diff > 0.0 && Diff < 1.0, "Diff > 0.0 && Diff < 1.0");

    // get the divisor and select pivot variables
    p.nDivs++;
    Vec_IntClear( vDiv );
    Vec_IntAppend( vDiv, Hsh_VecReadEntry(p.pHash, iDiv) );
    Fx_ManDivFindPivots( vDiv, &Lit0, &Lit1 );
    assert(Lit0 >= 0 && Lit1 >= 0, "Lit0 >= 0 && Lit1 >= 0");


    // collect single-cube-divisor cubes
    Vec_IntClear( p.vCubesS );
    if Vec_IntSize(vDiv) == 2 {
        Fx_ManCompressCubes( p.vCubes, Vec_WecEntry(p.vLits, Abc_LitNot(Lit0)) );
        Fx_ManCompressCubes( p.vCubes, Vec_WecEntry(p.vLits, Abc_LitNot(Lit1)) );
        Vec_IntTwoRemoveCommon( Vec_WecEntry(p.vLits, Abc_LitNot(Lit0)), Vec_WecEntry(p.vLits, Abc_LitNot(Lit1)), p.vCubesS );
    }

    // collect double-cube-divisor cube pairs
    Fx_ManCompressCubes( p.vCubes, Vec_WecEntry(p.vLits, Lit0) );
    Fx_ManCompressCubes( p.vCubes, Vec_WecEntry(p.vLits, Lit1) );
    Fx_ManFindCommonPairs( p.vCubes, Vec_WecEntry(p.vLits, Lit0), Vec_WecEntry(p.vLits, Lit1), p.vCubesD, p.vCompls, vDiv, p.vCubeFree, fWarning );

    // subtract cost of single-cube divisors
    // inlining of macro Fx_ManForEachCubeVec
    for  i = 0; (i < Vec_IntSize(p.vCubesS)) && ((vCube) = Vec_WecEntry(p.vCubes, Vec_IntEntry(p.vCubesS, i))); i++  {
        Fx_ManCubeSingleCubeDivisors( p, vCube, 1, 1 );
    }  // remove - update
    // inlining of macro Fx_ManForEachCubeVec
    for  i = 0; (i < Vec_IntSize(p.vCubesD)) && ((vCube) = Vec_WecEntry(p.vCubes, Vec_IntEntry(p.vCubesD, i))); i++  {
        Fx_ManCubeSingleCubeDivisors( p, vCube, 1, 1 );
    }  // remove - update

    // mark the cubes to be removed
    Vec_WecMarkLevels( p.vCubes, p.vCubesS );
    Vec_WecMarkLevels( p.vCubes, p.vCubesD );

    // subtract cost of double-cube divisors
    // inlining of macro Fx_ManForEachCubeVec
    for  i = 0; (i < Vec_IntSize(p.vCubesS)) && ((vCube) = Vec_WecEntry(p.vCubes, Vec_IntEntry(p.vCubesS, i))); i++  {
        Fx_ManCubeDoubleCubeDivisors( p, Fx_ManGetFirstVarCube(p, vCube), vCube, 1, 1, fWarning );
    }  // remove - update
    // inlining of macro Fx_ManForEachCubeVec
    for  i = 0; (i < Vec_IntSize(p.vCubesD)) && ((vCube) = Vec_WecEntry(p.vCubes, Vec_IntEntry(p.vCubesD, i))); i++  {
        Fx_ManCubeDoubleCubeDivisors( p, Fx_ManGetFirstVarCube(p, vCube), vCube, 1, 1, fWarning );
    }  // remove - update

    // unmark the cubes to be removed
    Vec_WecUnmarkLevels( p.vCubes, p.vCubesS );
    Vec_WecUnmarkLevels( p.vCubes, p.vCubesD );

    if Abc_Lit2Var(Lit0) == Abc_Lit2Var(Lit1) && Vec_IntSize(Hsh_VecReadEntry(p.pHash, iDiv)) == 2 {
        goto ExtractFromPairs; }

    // create new divisor
    iVarNew = Vec_WecSize( p.vLits ) / 2;
    assert(Vec_IntSize(p.vVarCube) == iVarNew, "Vec_IntSize(p.vVarCube) == iVarNew");
    Vec_IntPush( p.vVarCube, Vec_WecSize(p.vCubes) );
    vCube = Vec_WecPushLevel( p.vCubes );
    Vec_IntPush( vCube, iVarNew );
    if Vec_IntSize(vDiv) == 2 {
        Vec_IntPush( vCube, Abc_LitNot(Lit0) );
        Vec_IntPush( vCube, Abc_LitNot(Lit1) );
        Level = 1 + Fx_ManComputeLevelCube( p, vCube );
    
    } else {
        vCube2 = Vec_WecPushLevel( p.vCubes );
        vCube = Vec_WecEntry( p.vCubes, Vec_WecSize(p.vCubes) - 2 );
        Vec_IntPush( vCube2, iVarNew );
        Fx_ManDivAddLits( vCube, vCube2, vDiv );
        Level = 2 + Abc_MaxInt( Fx_ManComputeLevelCube(p, vCube), Fx_ManComputeLevelCube(p, vCube2) );
    }
    assert(Vec_IntSize(p.vLevels) == iVarNew, "Vec_IntSize(p.vLevels) == iVarNew");
    Vec_IntPush( p.vLevels, Level );
    // do not add new cubes to the matrix 
    p.nLits += Vec_IntSize( vDiv );
    // create new literals
    vLitP = Vec_WecPushLevel( p.vLits );
    vLitN = Vec_WecPushLevel( p.vLits );
    vLitP = Vec_WecEntry( p.vLits, Vec_WecSize(p.vLits) - 2 );
    // create updated single-cube divisor cubes
    // inlining of macro Fx_ManForEachCubeVec
    for  i = 0; (i < Vec_IntSize(p.vCubesS)) && ((vCube) = Vec_WecEntry(p.vCubes, Vec_IntEntry(p.vCubesS, i))); i++  {
        RetValue  = Vec_IntRemove1( vCube, Abc_LitNot(Lit0) );
        RetValue += Vec_IntRemove1( vCube, Abc_LitNot(Lit1) );
        assert(RetValue == 2, "RetValue == 2");
        Vec_IntPush( vCube, Abc_Var2Lit(iVarNew, 0) );
        Vec_IntPush( vLitP, Vec_WecLevelId(p.vCubes, vCube) );
        p.nLits--;
    }
    // create updated double-cube divisor cube pairs
ExtractFromPairs:
    k = 0;
    p.nCompls = 0;
    assert(Vec_IntSize(p.vCubesD) % 2 == 0, "Vec_IntSize(p.vCubesD) % 2 == 0");
    assert(Vec_IntSize(p.vCubesD) == 2 * Vec_IntSize(p.vCompls), "Vec_IntSize(p.vCubesD) == 2 * Vec_IntSize(p.vCompls)");
    for  i = 0; i < Vec_IntSize(p.vCubesD); i += 2  {
        var fCompl int32  = Vec_IntEntry(p.vCompls, i/2);
        p.nCompls += fCompl;
        vCube  = Vec_WecEntry( p.vCubes, Vec_IntEntry(p.vCubesD, i) );
        vCube2 = Vec_WecEntry( p.vCubes, Vec_IntEntry(p.vCubesD, i+1) );
        RetValue  = Fx_ManDivRemoveLits( vCube, vDiv, fCompl );  // cube 2*i
        RetValue += Fx_ManDivRemoveLits( vCube2, vDiv, fCompl ); // cube 2*i+1
        assert(RetValue == Vec_IntSize(vDiv) || RetValue == Vec_IntSize( vDiv ) + 1, "RetValue == Vec_IntSize(vDiv) || RetValue == Vec_IntSize( vDiv ) + 1");
        if iVarNew > 0 {
            if Vec_IntSize(vDiv) == 2 || (fCompl != 0) {
                Vec_IntPush( vCube, Abc_Var2Lit(iVarNew, 1) );
                Vec_IntPush( vLitN, Vec_WecLevelId(p.vCubes, vCube) ); // MAKE SURE vCube IS SORTED BY ID
            
            } else {
                Vec_IntPush( vCube, Abc_Var2Lit(iVarNew, 0) );
                Vec_IntPush( vLitP, Vec_WecLevelId(p.vCubes, vCube) ); // MAKE SURE vCube IS SORTED BY ID
            }
        }
        p.nLits -= Vec_IntSize(vDiv) + Vec_IntSize(vCube2) - 2;

        // remove second cube
        Vec_IntWriteEntry( p.vCubesD, k++, Vec_WecLevelId(p.vCubes, vCube) );
        Vec_IntClear( vCube2 ); 
    }
    assert(k == Vec_IntSize(p.vCubesD) / 2, "k == Vec_IntSize(p.vCubesD) / 2");
    Vec_IntShrink( p.vCubesD, k );
    Vec_IntSort( p.vCubesD, 0 );
    //Vec_IntSort( vLitN, 0 );
    //Vec_IntSort( vLitP, 0 );

    // add cost of single-cube divisors
    // inlining of macro Fx_ManForEachCubeVec
    for  i = 0; (i < Vec_IntSize(p.vCubesS)) && ((vCube) = Vec_WecEntry(p.vCubes, Vec_IntEntry(p.vCubesS, i))); i++  {
        Fx_ManCubeSingleCubeDivisors( p, vCube, 0, 1 );
    }  // add - update
    // inlining of macro Fx_ManForEachCubeVec
    for  i = 0; (i < Vec_IntSize(p.vCubesD)) && ((vCube) = Vec_WecEntry(p.vCubes, Vec_IntEntry(p.vCubesD, i))); i++  {
        Fx_ManCubeSingleCubeDivisors( p, vCube, 0, 1 );
    }  // add - update

    // mark the cubes to be removed
    Vec_WecMarkLevels( p.vCubes, p.vCubesS );
    Vec_WecMarkLevels( p.vCubes, p.vCubesD );

    // add cost of double-cube divisors
    // inlining of macro Fx_ManForEachCubeVec
    for  i = 0; (i < Vec_IntSize(p.vCubesS)) && ((vCube) = Vec_WecEntry(p.vCubes, Vec_IntEntry(p.vCubesS, i))); i++  {
        Fx_ManCubeDoubleCubeDivisors( p, Fx_ManGetFirstVarCube(p, vCube), vCube, 0, 1, fWarning );
    }  // add - update
    // inlining of macro Fx_ManForEachCubeVec
    for  i = 0; (i < Vec_IntSize(p.vCubesD)) && ((vCube) = Vec_WecEntry(p.vCubes, Vec_IntEntry(p.vCubesD, i))); i++  {
        Fx_ManCubeDoubleCubeDivisors( p, Fx_ManGetFirstVarCube(p, vCube), vCube, 0, 1, fWarning );
    }  // add - update

    // unmark the cubes to be removed
    Vec_WecUnmarkLevels( p.vCubes, p.vCubesS );
    Vec_WecUnmarkLevels( p.vCubes, p.vCubesD );

    // Deal with SCC
    if Vec_IntSize( p.vSCC ) {
        Vec_IntUniqify( p.vSCC );
        // inlining of macro Fx_ManForEachCubeVec
        for  i = 0; (i < Vec_IntSize(p.vSCC)) && ((vCube) = Vec_WecEntry(p.vCubes, Vec_IntEntry(p.vSCC, i))); i++  {
            Fx_ManCubeDoubleCubeDivisors( p, Fx_ManGetFirstVarCube(p, vCube), vCube, 1, 1, fWarning );  // remove - update
            Vec_IntClear( vCube );
        }
        Vec_IntClear( p.vSCC );
    }
    // add cost of the new divisor
    if Vec_IntSize(vDiv) > 2 {
        vCube  = Vec_WecEntry( p.vCubes, Vec_WecSize(p.vCubes) - 2 );
        vCube2 = Vec_WecEntry( p.vCubes, Vec_WecSize(p.vCubes) - 1 );
        Fx_ManCubeSingleCubeDivisors( p, vCube,  0, 1 );  // add - update
        Fx_ManCubeSingleCubeDivisors( p, vCube2, 0, 1 );  // add - update
        // inlining of macro Vec_IntForEachEntryStart
        for  i = 1; (i < Vec_IntSize(vCube)) && (((Lit0) = Vec_IntEntry(vCube, i)), 1); i++  {
            Vec_WecPush( p.vLits, Lit0, Vec_WecLevelId(p.vCubes, vCube) );
        }
        // inlining of macro Vec_IntForEachEntryStart
        for  i = 1; (i < Vec_IntSize(vCube2)) && (((Lit0) = Vec_IntEntry(vCube2, i)), 1); i++  {
            Vec_WecPush( p.vLits, Lit0, Vec_WecLevelId(p.vCubes, vCube2) );
        }
    }

    // remove these cubes from the lit array of the divisor
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vDiv)) && (((Lit0) = Vec_IntEntry(vDiv, i)), 1); i++  {
        Vec_IntTwoRemove( Vec_WecEntry(p.vLits, Abc_Lit2Var(Lit0)), p.vCubesD );
        if (p.nCompls && i > 1) || Vec_IntSize( vDiv ) == 2 { // the last two lits are possibly complemented
            Vec_IntTwoRemove( Vec_WecEntry(p.vLits, Abc_LitNot(Abc_Lit2Var(Lit0))), p.vCubesD ); }
    }

}

/**Function*************************************************************

  Synopsis    [Implements the traditional fast_extract algorithm.]

  Description [J. Rajski and J. Vasudevamurthi, "The testability-
  preserving concurrent decomposition and factorization of Boolean
  expressions", IEEE TCAD, Vol. 11, No. 6, June 1992, pp. 778-793.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fx_FastExtract(vCubes *Vec_Wec_t, ObjIdMax int32, nNewNodesMax int32, LitCountMax int32, fCanonDivs int32, fVerbose int32, fVeryVerbose int32) int32 {
    var fVeryVeryVerbose int32;
    var i, iDiv, fWarning int32;
    var p *Fx_Man_t;
    var clk abctime  = Abc_Clock();
    // initialize the data-structure
    p = Fx_ManStart( vCubes );
    p.LitCountMax = LitCountMax;
    p.fCanonDivs = fCanonDivs;
    Fx_ManCreateLiterals( p, ObjIdMax );
    Fx_ManComputeLevel( p );
    Fx_ManCreateDivisors( p );
    if fVeryVerbose != 0 {
        Fx_PrintDivisors( p ); }
    if fVerbose != 0 {
        Fx_PrintStats( p, Abc_Clock() - clk ); }
    // perform extraction
    p.timeStart = Abc_Clock();
    for  i = 0; i < nNewNodesMax && Vec_QueTopPriority(p.vPrio) > 0.0; i++  {
        iDiv = Vec_QuePop(p.vPrio);
        if fVeryVerbose != 0 {
            Fx_PrintDiv( p, iDiv ); }
        Fx_ManUpdate( p, iDiv, &fWarning );
        if fVeryVeryVerbose != 0 {
            Fx_PrintDivisors( p ); }
    }
    if fVerbose != 0 {
        Fx_PrintStats( p, Abc_Clock() - clk ); }
    Fx_ManStop( p );
    // return the result
    Vec_WecRemoveEmpty( vCubes );
    return 1;
}



////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

