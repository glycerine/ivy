package gia

/**CFile****************************************************************

  FileName    [giaFx.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    [Interface to fast_extract package.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaFx.c,v 1.00 2013/09/29 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "gia.h"
//#include "bool/kit/kit.h"
//#include "misc/vec/vecWec.h"
//#include "bool/dec/dec.h"
//#include "opt/dau/dau.h"
//#include "misc/util/utilTruth.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Create GIA for SOP.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManGraphToAig(p *Gia_Man_t, pGraph *Dec_Graph_t) int32 {
    var pNode *Dec_Node_t  = nil; // Suppress "might be used uninitialized"
    var i, iAnd0, iAnd1 int32;
    // check for constant function
    if Dec_GraphIsConst(pGraph) {
        return Abc_LitNotCond( 1, Dec_GraphIsComplement(pGraph) ); }
    // check for a literal
    if Dec_GraphIsVar(pGraph) {
        return Abc_LitNotCond( Dec_GraphVar(pGraph).iFunc, Dec_GraphIsComplement(pGraph) ); }
    // build the AIG nodes corresponding to the AND gates of the graph
    // inlining of macro Dec_GraphForEachNode
    for  i = (pGraph).nLeaves; (i < (pGraph).nSize) && (((pNode) = Dec_GraphNode(pGraph, i)), 1); i++  {
        iAnd0 = Abc_LitNotCond( Dec_GraphNode(pGraph, pNode.eEdge0.Node).iFunc, pNode.eEdge0.fCompl ); 
        iAnd1 = Abc_LitNotCond( Dec_GraphNode(pGraph, pNode.eEdge1.Node).iFunc, pNode.eEdge1.fCompl ); 
        pNode.iFunc = Gia_ManHashAnd( p, iAnd0, iAnd1 );
    }
    // complement the result if necessary
    return Abc_LitNotCond( pNode.iFunc, Dec_GraphIsComplement(pGraph) );
}
func Gia_ManSopToAig(p *Gia_Man_t, pSop string, vLeaves *Vec_Int_t) int32 {
    var i, iAnd, iSum, Value, nFanins int32;
    var pCube string;
    // get the number of variables
    nFanins = Kit_PlaGetVarNum(pSop);
    // go through the cubes of the node's SOP
    iSum = 0;
    // inlining of macro Kit_PlaForEachCube
    for  pCube = (pSop); *pCube; pCube += (nFanins) + 3  {
        // create the AND of literals
        iAnd = 1;
        // inlining of macro Kit_PlaCubeForEachVar
        for  i = 0; (pCube[i] != ' ') && (Value = pCube[i]); i++  {
            assert(Vec_IntEntry(vLeaves, i) >= 0, "Vec_IntEntry(vLeaves, i) >= 0");
            if Value == '1' {
                iAnd = Gia_ManHashAnd( p, iAnd, Vec_IntEntry(vLeaves, i) );
            } else if Value == '0' {
                iAnd = Gia_ManHashAnd( p, iAnd, Abc_LitNot(Vec_IntEntry(vLeaves, i)) );
            } else { assert(Value == '-', "Value == '-'"); }
        }
        // add to the sum of cubes
        iSum = Gia_ManHashOr( p, iSum, iAnd );
    }
    // decide whether to complement the result
    if Kit_PlaIsComplement(pSop) {
        iSum = Abc_LitNot(iSum); }
    return iSum;
}
func Gia_ManFactorGraph(p *Gia_Man_t, pFForm *Dec_Graph_t, vLeaves *Vec_Int_t) int32 {
    var pFFNode *Dec_Node_t;
    var i, Lit int32;
    // assign fanins
    // inlining of macro Dec_GraphForEachLeaf
    for  i = 0; (i < (pFForm).nLeaves) && (((pFFNode) = Dec_GraphNode(pFForm, i)), 1); i++  {
        assert(Vec_IntEntry(vLeaves, i) >= 0, "Vec_IntEntry(vLeaves, i) >= 0");
        pFFNode.iFunc = Vec_IntEntry(vLeaves, i);
    }
    // perform strashing
    Lit = Gia_ManGraphToAig( p, pFForm );
    return Lit;
}
func Gia_ManFactorNode(p *Gia_Man_t, pSop string, vLeaves *Vec_Int_t) int32 {
    if Kit_PlaGetVarNum(pSop) == 0 {
        return Abc_LitNotCond( 1, Kit_PlaIsConst0(pSop) ); }
    assert(Kit_PlaGetVarNum(pSop) == Vec_IntSize(vLeaves), "Kit_PlaGetVarNum(pSop) == Vec_IntSize(vLeaves)");
    if Kit_PlaGetVarNum(pSop) > 2 && Kit_PlaGetCubeNum(pSop) > 1 {
        var pFForm *Dec_Graph_t  = Dec_Factor( pSop );
        var Lit int32  = Gia_ManFactorGraph( p, pFForm, vLeaves );
        Dec_GraphFree( pFForm );
        return Lit;
    }
    return Gia_ManSopToAig( p, pSop, vLeaves );
}

/**Function*************************************************************

  Synopsis    [Computing truth tables for the mapped network.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManComputeTruths(p *Gia_Man_t, nCutSize int32, nLutNum int32, fReverse int32) *Vec_Wrd_t {
    var vTruths *Vec_Wrd_t;
    var vLeaves Vec_Int_t;
    var pTruth *word;
    var i, k, nWords int32;
    nWords = Abc_Truth6WordNum( nCutSize );
    vTruths = Vec_WrdAlloc( nWords * nLutNum );
    Gia_ObjComputeTruthTableStart( p, nCutSize );
    Gia_ManForEachLut( p, i )
    {
        // collect and sort fanins
        vLeaves.nCap = vLeaves.nSize = Gia_ObjLutSize( p, i );
        vLeaves.pArray = Gia_ObjLutFanins( p, i );
        assert(Vec_IntCheckUniqueSmall(&vLeaves), "Vec_IntCheckUniqueSmall(&vLeaves)");
        Vec_IntSelectSort( Vec_IntArray(&vLeaves), Vec_IntSize(&vLeaves) );
        if !(fReverse != 0) {
            Vec_IntReverseOrder( &vLeaves ); }
        // compute truth table
        pTruth = Gia_ObjComputeTruthTableCut( p, Gia_ManObj(p, i), &vLeaves );
        for  k = 0; k < nWords; k++  {
            Vec_WrdPush( vTruths, pTruth[k] );
        }
//        Kit_DsdPrintFromTruth( (unsigned *)pTruth, 6 ); printf( "\n" );
    }
    Gia_ObjComputeTruthTableStop( p );
    assert(Vec_WrdCap(vTruths) == 16 || Vec_WrdSize(vTruths) == Vec_WrdCap(vTruths), "Vec_WrdCap(vTruths) == 16 || Vec_WrdSize(vTruths) == Vec_WrdCap(vTruths)");
    return vTruths;
}

/**Function*************************************************************

  Synopsis    [Extracts information about the network.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManAssignNumbers(p *Gia_Man_t) int32 {
    var pObj *Gia_Obj_t;
    var i, Counter int32;
    Gia_ManFillValue( p );
    Gia_ManForEachCi( p, pObj, i )
        pObj.Value = Counter++;
    Gia_ManForEachLut( p, i )
        Gia_ManObj(p, i).Value = Counter++;
    return Counter;
}
func Gia_ManFxRetrieve(p *Gia_Man_t, pvCompl **Vec_Str_t, fReverse int32) *Vec_Wec_t {
    var vCubes *Vec_Wec_t;
    var vTruths *Vec_Wrd_t;
    var vCube, vCover *Vec_Int_t;
    var nItems, nCutSize, nWords int32;
    var i, c, v, Lit, Cube, Counter int32;
//    abctime clk = Abc_Clock();
    nItems = Gia_ManAssignNumbers( p );
    // compute truth tables
    nCutSize = Gia_ManLutSizeMax( p );
    nWords = Abc_Truth6WordNum( nCutSize );
    vTruths = Gia_ManComputeTruths( p, nCutSize, nItems - Gia_ManCiNum(p), fReverse );
    vCover = Vec_IntAlloc( 1 << 16 );
    // collect cubes
    vCubes = Vec_WecAlloc( 1000 );
    *pvCompl = Vec_StrStart( nItems );
    Gia_ManForEachLut( p, i )
    {
        var pObj *Gia_Obj_t  = Gia_ManObj( p, i );
        var nVars int32  = Gia_ObjLutSize( p, i );
        var pVars *int32  = Gia_ObjLutFanins( p, i );
        var pTruth *word  = Vec_WrdEntryP( vTruths, Counter++ * nWords );
        var Status int32  = Kit_TruthIsop( (unsigned *)pTruth, nVars, vCover, 1 );
        if Vec_IntSize(vCover) == 0 || (Vec_IntSize(vCover) == 1 && Vec_IntEntry(vCover,0) == 0) {
            Vec_StrWriteEntry( *pvCompl, pObj.Value, byte((Vec_IntSize(vCover) == 0)) );
            vCube = Vec_WecPushLevel( vCubes );
            Vec_IntPush( vCube, pObj.Value );
            continue;
        }
        Vec_StrWriteEntry( *pvCompl, pObj.Value, byte(Status) );
        Vec_IntForEachEntry( vCover, Cube, c )
        {
            vCube = Vec_WecPushLevel( vCubes );
            Vec_IntPush( vCube, pObj.Value );
            for  v = 0; v < nVars; v++  {
                Lit = 3 & (Cube >> (v << 1));
                if Lit == 1 {
                    Vec_IntPush( vCube, Abc_Var2Lit(Gia_ManObj(p, pVars[v]).Value, 1) );
                } else if Lit == 2 {
                    Vec_IntPush( vCube, Abc_Var2Lit(Gia_ManObj(p, pVars[v]).Value, 0) );
                } else if Lit != 0 {
                    assert(false, "0"); }
            }
            Vec_IntSelectSort( Vec_IntArray(vCube) + 1, Vec_IntSize(vCube) - 1 );
        }        
    }
    assert(Counter * nWords == Vec_WrdSize(vTruths), "Counter * nWords == Vec_WrdSize(vTruths)");
    Vec_WrdFree( vTruths );
    Vec_IntFree( vCover );
//    Abc_PrintTime( 1, "Setup time", Abc_Clock() - clk );
    return vCubes;
}

/**Function*************************************************************

  Synopsis    [Generates GIA after factoring the resulting SOPs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManFxTopoOrder_rec(vCubes *Vec_Wec_t, vFirst *Vec_Int_t, vCount *Vec_Int_t, vVisit *Vec_Int_t, vOrder *Vec_Int_t, iObj int32) {
    var c, v, Lit int32;
    var iFirst int32  = Vec_IntEntry( vFirst, iObj );
    var nCubes int32  = Vec_IntEntry( vCount, iObj ); 
    assert(!Vec_IntEntry( vVisit, iObj ), "!Vec_IntEntry( vVisit, iObj )");
    Vec_IntWriteEntry( vVisit, iObj, 1 );
    for  c = 0; c < nCubes; c++  {
        var vCube *Vec_Int_t  = Vec_WecEntry( vCubes, iFirst + c );
        assert(Vec_IntEntry(vCube, 0) == iObj, "Vec_IntEntry(vCube, 0) == iObj");
        Vec_IntForEachEntryStart( vCube, Lit, v, 1 )
            if ( !Vec_IntEntry( vVisit, Abc_Lit2Var(Lit) ) )
                Gia_ManFxTopoOrder_rec( vCubes, vFirst, vCount, vVisit, vOrder, Abc_Lit2Var(Lit) );
    }
    Vec_IntPush( vOrder, iObj );
}
func Gia_ManFxTopoOrder(vCubes *Vec_Wec_t, nInputs int32, nStart int32, pvFirst **Vec_Int_t, pvCount **Vec_Int_t) *Vec_Int_t {
    var vOrder, vFirst, vCount, vVisit, vCube *Vec_Int_t;
    var i int32;
    var iFanin int32;
    var nNodeMax int32 = -1;
    // find the largest index
    Vec_WecForEachLevel( vCubes, vCube, i )
        nNodeMax = Abc_MaxInt( nNodeMax, Vec_IntEntry(vCube, 0) );
    nNodeMax++;
    // quit if there is no new nodes
    if nNodeMax == nStart {
        //printf( "The network is unchanged by fast extract.\n" );
        return nil;
    }
    // find first cube and how many cubes
    vFirst = Vec_IntStart( nNodeMax );
    vCount = Vec_IntStart( nNodeMax );
    Vec_WecForEachLevel( vCubes, vCube, i )
    {
        iFanin = Vec_IntEntry( vCube, 0 );
        assert(iFanin >= nInputs, "iFanin >= nInputs");
        if Vec_IntEntry(vCount, iFanin) == 0 {
            Vec_IntWriteEntry( vFirst, iFanin, i ); }
        Vec_IntAddToEntry( vCount, iFanin, 1 );
    }
    // put all of them in a topo order
    vOrder = Vec_IntStart( nInputs );
    vVisit = Vec_IntStart( nNodeMax );
    for  i = 0; i < nInputs; i++  {
        Vec_IntWriteEntry( vVisit, i, 1 );
    }
    for  i = nInputs; i < nNodeMax; i++  {
        if !Vec_IntEntry( vVisit, i ) {
            Gia_ManFxTopoOrder_rec( vCubes, vFirst, vCount, vVisit, vOrder, i ); }
    }
    assert(Vec_IntSize(vOrder) == nNodeMax, "Vec_IntSize(vOrder) == nNodeMax");
    Vec_IntFree( vVisit );
    // return topological order of new nodes
    *pvFirst = vFirst;
    *pvCount = vCount;
    return vOrder;
}
func Gia_ManFxInsert(p *Gia_Man_t, vCubes *Vec_Wec_t, vCompls *Vec_Str_t) *Gia_Man_t {
    var pNew, pTemp *Gia_Man_t;
    var pObj *Gia_Obj_t; var vSop *Vec_Str_t;
    var vOrder, vFirst, vCount, vFanins, vCover *Vec_Int_t;
    var vCopies, vCube, vMap *Vec_Int_t;
    var k, c, v, Lit, Var, iItem int32;
//    abctime clk = Abc_Clock();
    // prepare the cubes
    vOrder = Gia_ManFxTopoOrder( vCubes, Gia_ManCiNum(p), Vec_StrSize(vCompls), &vFirst, &vCount );
    if vOrder == nil {
        return Gia_ManDup( p ); }
    assert(Vec_IntSize(vOrder) > Vec_StrSize(vCompls), "Vec_IntSize(vOrder) > Vec_StrSize(vCompls)");
    // create new manager
    pNew = Gia_ManStart( Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    pNew.vLevels = Vec_IntStart( 6*Gia_ManObjNum(p)/5 + 100 );
    Gia_ManHashStart( pNew );
    // create primary inputs
    vMap = Vec_IntStartFull( Vec_IntSize(vOrder) );
    vCopies = Vec_IntAlloc( Vec_IntSize(vOrder) );
    Gia_ManForEachCi( p, pObj, k )
        Vec_IntPush( vCopies, Gia_ManAppendCi(pNew) );
    Vec_IntFillExtra( vCopies, Vec_IntSize(vOrder), -1 );
    // add AIG nodes in the topological order
    vSop = Vec_StrAlloc( 1000 );
    vCover = Vec_IntAlloc( 1 << 16 );
    vFanins = Vec_IntAlloc( 100 );
    Vec_IntForEachEntryStart( vOrder, iItem, k, Gia_ManCiNum(p) )
    {
        var iFirst int32  = Vec_IntEntry( vFirst, iItem );
        var nCubes int32  = Vec_IntEntry( vCount, iItem );
        // collect fanins
        Vec_IntClear( vFanins );
        for  c = 0; c < nCubes; c++  {
            vCube = Vec_WecEntry( vCubes, iFirst + c );
            Vec_IntForEachEntryStart( vCube, Lit, v, 1 )
                if Vec_IntEntry(vMap, Abc_Lit2Var(Lit)) == -1 {
                    Vec_IntWriteEntry( vMap, Abc_Lit2Var(Lit), Vec_IntSize(vFanins) );
                    Vec_IntPush( vFanins, Abc_Lit2Var(Lit) );
                }
        }
        if Vec_IntSize(vFanins) > 6 {
            // create SOP
            Vec_StrClear( vSop );
            for  c = 0; c < nCubes; c++  {
                for  v = 0; v < Vec_IntSize(vFanins); v++  {
                    Vec_StrPush( vSop, '-' );
                }
                vCube = Vec_WecEntry( vCubes, iFirst + c );
                Vec_IntForEachEntryStart( vCube, Lit, v, 1 )
                {
                    Lit = Abc_Lit2LitV( Vec_IntArray(vMap), Lit );
                    assert(Lit >= 0 && Abc_Lit2Var(Lit) < Vec_IntSize(vFanins), "Lit >= 0 && Abc_Lit2Var(Lit) < Vec_IntSize(vFanins)");
                    Vec_StrWriteEntry( vSop, Vec_StrSize(vSop) - Vec_IntSize(vFanins) + Abc_Lit2Var(Lit), byte((Abc_LitIsCompl(Lit)? '0' : '1')) );
                }
                Vec_StrPush( vSop, ' ' );
                Vec_StrPush( vSop, '1' );
                Vec_StrPush( vSop, '\n' );
            }
            Vec_StrPush( vSop, '\0' );
            // collect fanins
            Vec_IntForEachEntry( vFanins, Var, v )
            {
                Vec_IntWriteEntry( vMap, Var, -1 );
                Vec_IntWriteEntry( vFanins, v, Vec_IntEntry(vCopies, Var) );
            }
            // derive new AIG
            Lit = Gia_ManFactorNode( pNew, Vec_StrArray(vSop), vFanins );
        
        } else {
            var uTruth, uCube word;
            for  c = 0; c < nCubes; c++  {
                uCube = ~(word)0;
                vCube = Vec_WecEntry( vCubes, iFirst + c );
                Vec_IntForEachEntryStart( vCube, Lit, v, 1 )
                {
                    Lit = Abc_Lit2LitV( Vec_IntArray(vMap), Lit );
                    assert(Lit >= 0 && Abc_Lit2Var(Lit) < Vec_IntSize(vFanins), "Lit >= 0 && Abc_Lit2Var(Lit) < Vec_IntSize(vFanins)");
                    uCube &= Abc_LitIsCompl(Lit) ? ~s_Truths6[Abc_Lit2Var(Lit)] : s_Truths6[Abc_Lit2Var(Lit)];
                }
                uTruth |= uCube;
            }
            // complement constant
            if uTruth == 0 {
                uTruth = ~uTruth; }
            // collect fanins
            Vec_IntForEachEntry( vFanins, Var, v )
            {
                Vec_IntWriteEntry( vMap, Var, -1 );
                Vec_IntWriteEntry( vFanins, v, Vec_IntEntry(vCopies, Var) );
            }
            // create truth table
            Lit = Dsm_ManTruthToGia( pNew, &uTruth, vFanins, vCover );
        }
        // complement if the original SOP was complemented
        Lit = Abc_LitNotCond( Lit, (iItem < Vec_StrSize(vCompls)) && (Vec_StrEntry(vCompls, iItem) > 0) );
        // remeber this literal
        assert(Vec_IntEntry(vCopies, iItem) == -1, "Vec_IntEntry(vCopies, iItem) == -1");
        Vec_IntWriteEntry( vCopies, iItem, Lit );
    }
    Gia_ManHashStop( pNew );
    // create primary outputs
    Gia_ManForEachCo( p, pObj, k )
    {
        if Gia_ObjFaninId0p(p, pObj) {
            Lit = Vec_IntEntry(vCopies, Gia_ObjFanin0(pObj).Value);
        } else {
            Lit = 0;
        }
        Gia_ManAppendCo( pNew, Abc_LitNotCond( Lit, Gia_ObjFaninC0(pObj) ) );
    }
    Gia_ManSetRegNum( pNew, Gia_ManRegNum(p) );
    // cleanup
    Vec_IntFree( vOrder );
    Vec_IntFree( vFirst );
    Vec_IntFree( vCount );
    Vec_IntFree( vFanins );
    Vec_IntFree( vCopies );
    Vec_IntFree( vMap );
    Vec_StrFree( vSop );
    Vec_IntFree( vCover );
    // remove dangling nodes
    pNew = Gia_ManCleanup( pTemp = pNew );
    Gia_ManStop( pTemp );
//    Abc_PrintTime( 1, "Setdn time", Abc_Clock() - clk );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Performs classical fast_extract on logic functions.]

  Description []
               
  SideEffects [Sorts the fanins of each cut in the increasing order.]

  SeeAlso     []

***********************************************************************/
func Gia_ManPerformFx(p *Gia_Man_t, nNewNodesMax int32, LitCountMax int32, fReverse int32, fVerbose int32, fVeryVerbose int32) *Gia_Man_t {
    extern int Fx_FastExtract( Vec_Wec_t * vCubes, int ObjIdMax, int nNewNodesMax, int LitCountMax, int fCanonDivs, int fVerbose, int fVeryVerbose );
    var pNew *Gia_Man_t  = nil;
    var vCubes *Vec_Wec_t;
    var vCompl *Vec_Str_t;
    if Gia_ManAndNum(p) == 0 {
        pNew = Gia_ManDup(p);
        Gia_ManTransferTiming( pNew, p );
        return pNew;
    }
//    abctime clk;
    assert(Gia_ManHasMapping(p), "Gia_ManHasMapping(p)");   
    // collect information
    vCubes = Gia_ManFxRetrieve( p, &vCompl, fReverse );
    // call the fast extract procedure
//    clk = Abc_Clock();
    Fx_FastExtract( vCubes, Vec_StrSize(vCompl), nNewNodesMax, LitCountMax, 0, fVerbose, fVeryVerbose );
//    Abc_PrintTime( 1, "Fx runtime", Abc_Clock() - clk );
    // insert information
    pNew = Gia_ManFxInsert( p, vCubes, vCompl );
    Gia_ManTransferTiming( pNew, p );
    // cleanup
    Vec_WecFree( vCubes );
    Vec_StrFree( vCompl );
    return pNew;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

