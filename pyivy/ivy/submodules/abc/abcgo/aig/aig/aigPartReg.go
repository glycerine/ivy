package aig

import "fmt"

/**CFile****************************************************************

  FileName    [aigPartReg.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [AIG package.]

  Synopsis    [Register partitioning algorithm.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - April 28, 2007.]

  Revision    [$Id: aigPartReg.c,v 1.00 2007/04/28 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "aig.h"
//#include "fra.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

           

type Aig_ManPre_t struct {
    // input data    
    pAig      *Aig_Man_t;            // seq AIG manager 
    vMatrix      *Vec_Ptr_t;         // register dependency
    nRegsMax             int32;        // the max number of registers in the cluster
    // information about partitions
    vParts      *Vec_Ptr_t;          // the partitions
    pfUsedRegs           string;      // the registers already included in the partitions
    // info about the current partition
    vRegs      *Vec_Int_t;           // registers of this partition
    vUniques      *Vec_Int_t;        // unique registers of this partition
    vFreeVars      *Vec_Int_t;       // free variables of this partition
    vPartCost      *Vec_Flt_t;       // costs of adding each variable
    pfPartVars           string;      // input/output registers of the partition
}


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Computes partitioning of registers.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManRegManStart(pAig *Aig_Man_t, nPartSize int32) *Aig_ManPre_t {
    var p *Aig_ManPre_t;
    // inlining of macro ABC_ALLOC
    p = ((Aig_ManPre_t *) malloc(sizeof(Aig_ManPre_t) * (size_t)(1)));
    memset( p, 0, sizeof(Aig_ManPre_t) );
    p.pAig      = pAig;
    p.vMatrix   = Aig_ManSupportsRegisters( pAig );
    p.nRegsMax  = nPartSize;
    p.vParts    = Vec_PtrAlloc(256);
    p.vRegs     = Vec_IntAlloc(256);
    p.vUniques  = Vec_IntAlloc(256);
    p.vFreeVars = Vec_IntAlloc(256);
    p.vPartCost = Vec_FltAlloc(256);
    // inlining of macro ABC_ALLOC
    p.pfUsedRegs = ((char *) malloc(sizeof(char) * (size_t)(Aig_ManRegNum(p.pAig))));
    memset( p.pfUsedRegs, 0, sizeof(char) * Aig_ManRegNum(p.pAig) );
    // inlining of macro ABC_ALLOC
    p.pfPartVars  = ((char *) malloc(sizeof(char) * (size_t)(Aig_ManRegNum(p.pAig))));
    return p;
}

/**Function*************************************************************

  Synopsis    [Computes partitioning of registers.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManRegManStop(p *Aig_ManPre_t) {
    Vec_VecFree( (Vec_Vec_t *)p.vMatrix );
    if p.vParts {
        Vec_VecFree( (Vec_Vec_t *)p.vParts ); }
    Vec_IntFree( p.vRegs );
    Vec_IntFree( p.vUniques );
    Vec_IntFree( p.vFreeVars );
    Vec_FltFree( p.vPartCost );
    // inlining of macro ABC_FREE
    ((p.pfUsedRegs) ? (free((char *) (p.pfUsedRegs)), (p.pfUsedRegs) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pfPartVars) ? (free((char *) (p.pfPartVars)), (p.pfPartVars) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}

/**Function*************************************************************

  Synopsis    [Determines what register to use as the seed.]

  Description [The register is selected as the one having the largest
  number of non-taken registers in its support.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManRegFindSeed(p *Aig_ManPre_t) int32 {
    var vRegs *Vec_Int_t;
    var i, k, iReg int32;
    var iMax int32  = -1; // Suppress "might be used uninitialized"
    var nRegsCur int32;
    var nRegsMax int32 = -1;
    for  i = 0; i < Aig_ManRegNum(p.pAig); i++  {
        if p.pfUsedRegs[i] {
            continue; }
        nRegsCur = 0;
        vRegs = (Vec_Int_t *)Vec_PtrEntry( p.vMatrix, i );
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(vRegs)) && (((iReg) = Vec_IntEntry(vRegs, k)), 1); k++  {
            nRegsCur += !p.pfUsedRegs[iReg];
        }
        if nRegsMax < nRegsCur {
            nRegsMax = nRegsCur;
            iMax = i;
        }
    }
    return iMax;
}

/**Function*************************************************************

  Synopsis    [Computes the next register to be added to the set.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManRegFindBestVar(p *Aig_ManPre_t) int32 {
    var vSupp *Vec_Int_t;
    var nNewVars int32;
    var nNewVarsBest int32 = ABC_INFINITY;
    var iVarFree int32;
    var iVarSupp int32;
    var iVarBest int32 = -1;
    var i int32;
    var k int32;
    // go through the free variables
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(p.vFreeVars)) && (((iVarFree) = Vec_IntEntry(p.vFreeVars, i)), 1); i++  {
//        if ( p->pfUsedRegs[iVarFree] )
//            continue;
        // get support of this variable
        vSupp = (Vec_Int_t *)Vec_PtrEntry( p.vMatrix, iVarFree );
        // count the number of new vars
        nNewVars = 0;
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(vSupp)) && (((iVarSupp) = Vec_IntEntry(vSupp, k)), 1); k++  {
            if p.pfPartVars[iVarSupp] {
                continue; }
            nNewVars += 1 + 3 * p.pfUsedRegs[iVarSupp];
        }
        // quit if there is no new variables
        if nNewVars == 0 {
            return iVarFree; }
        // compare the cost of this
        if nNewVarsBest > nNewVars {
            nNewVarsBest = nNewVars;
            iVarBest = iVarFree;
        }
    }
    return iVarBest;
}

/**Function*************************************************************

  Synopsis    [Computes partitioning of registers.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManRegPartitionAdd(p *Aig_ManPre_t, iReg int32) {
    var vSupp *Vec_Int_t;
    var RetValue, iVar, i int32;
    // make sure this is a new variable
//    assert( !p->pfUsedRegs[iReg] );
    if !p.pfUsedRegs[iReg] {
        p.pfUsedRegs[iReg] = 1;
        Vec_IntPush( p.vUniques, iReg );
    }
    // remove it from the free variables
    if Vec_IntSize(p.vFreeVars) > 0 {
        assert(p.pfPartVars[iReg], "p.pfPartVars[iReg]");
        RetValue = Vec_IntRemove( p.vFreeVars, iReg );
        assert(RetValue, "RetValue");
    
    } else {
        assert(!p.pfPartVars[iReg], "!p.pfPartVars[iReg]"); }
    // add it to the partition
    p.pfPartVars[iReg] = 1;
    Vec_IntPush( p.vRegs, iReg );
    // add new variables
    vSupp = (Vec_Int_t *)Vec_PtrEntry( p.vMatrix, iReg );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vSupp)) && (((iVar) = Vec_IntEntry(vSupp, i)), 1); i++  {
        if p.pfPartVars[iVar] {
            continue; }
        p.pfPartVars[iVar] = 1;
        Vec_IntPush( p.vFreeVars, iVar );
    }
    // add it to the cost
    Vec_FltPush( p.vPartCost, 1.0*Vec_IntSize(p.vFreeVars)/Vec_IntSize(p.vRegs) );
}

/**Function*************************************************************

  Synopsis    [Creates projection of 1-hot registers onto the given partition.]

  Description [Assumes that the relevant register outputs are labeled with
  the current traversal ID.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManRegProjectOnehots(pAig *Aig_Man_t, pPart *Aig_Man_t, vOnehots *Vec_Ptr_t, fVerbose int32) *Vec_Ptr_t {
    var vOnehotsPart *Vec_Ptr_t  = nil;
    var vGroup, vGroupNew *Vec_Int_t;
    var pObj, pObjNew *Aig_Obj_t;
    var nOffset, iReg, i, k int32;
    // set the PI numbers
    // inlining of macro Aig_ManForEachCi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pPart.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pPart.vCis, i)), 1); i++  {
        pObj.iData = i;
    }
    // go through each group and check if registers are involved in this one
    nOffset = Aig_ManCiNum(pAig)-Aig_ManRegNum(pAig);
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vOnehots)) && (((vGroup) = (Vec_Int_t *)Vec_PtrEntry(vOnehots, i)), 1); i++  {
        vGroupNew = nil;
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(vGroup)) && (((iReg) = Vec_IntEntry(vGroup, k)), 1); k++  {
            pObj = Aig_ManCi( pAig, nOffset+iReg );
            if !Aig_ObjIsTravIdCurrent(pAig, pObj) {
                continue; }
            if vGroupNew == nil {
                vGroupNew = Vec_IntAlloc( Vec_IntSize(vGroup) ); }
            pObjNew = (Aig_Obj_t *)pObj.pData;
            Vec_IntPush( vGroupNew, pObjNew.iData );
        }
        if vGroupNew == nil {
            continue; }
        if Vec_IntSize(vGroupNew) > 1 {
            if vOnehotsPart == nil {
                vOnehotsPart = Vec_PtrAlloc( 100 ); }
            Vec_PtrPush( vOnehotsPart, vGroupNew );
        
        } else {
            Vec_IntFree( vGroupNew ); }
    }
    // clear the PI numbers
    // inlining of macro Aig_ManForEachCi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pPart.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pPart.vCis, i)), 1); i++  {
        pObj.iData = 0;
    }
    // print out
    if (vOnehotsPart != nil) && (fVerbose != 0) {
        fmt.Printf("Partition contains %d groups of 1-hot registers: { ", Vec_PtrSize(vOnehotsPart));
        // inlining of macro Vec_PtrForEachEntry
        for  k = 0; (k < Vec_PtrSize(vOnehotsPart)) && (((vGroup) = (Vec_Int_t *)Vec_PtrEntry(vOnehotsPart, k)), 1); k++  {
            fmt.Printf("%d ", Vec_IntSize(vGroup));
        }
        fmt.Printf("}\n");
    }
    return vOnehotsPart;
}

/**Function*************************************************************

  Synopsis    [Computes partitioning of registers.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManRegCreatePart(pAig *Aig_Man_t, vPart *Vec_Int_t, pnCountPis *int32, pnCountRegs *int32, ppMapBack **int32) *Aig_Man_t {
    var pNew *Aig_Man_t;
    var pObj, pObjNew *Aig_Obj_t;
    var vNodes *Vec_Ptr_t;
    var vRoots *Vec_Ptr_t;
    var nOffset, iOut, i int32;
    var nCountPis, nCountRegs int32;
    var pMapBack *int32;
    // collect roots
    vRoots = Vec_PtrAlloc( Vec_IntSize(vPart) );
    nOffset = Aig_ManCoNum(pAig)-Aig_ManRegNum(pAig);
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vPart)) && (((iOut) = Vec_IntEntry(vPart, i)), 1); i++  {
        pObj = Aig_ManCo(pAig, nOffset+iOut);
        Vec_PtrPush( vRoots, Aig_ObjFanin0(pObj) );
    }
    // collect/mark nodes/PIs in the DFS order
    vNodes = Aig_ManDfsNodes( pAig, (Aig_Obj_t **)Vec_PtrArray(vRoots), Vec_PtrSize(vRoots) );
    Vec_PtrFree( vRoots );
    // unmark register outputs
    nOffset = Aig_ManCiNum(pAig)-Aig_ManRegNum(pAig);
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vPart)) && (((iOut) = Vec_IntEntry(vPart, i)), 1); i++  {
        pObj = Aig_ManCi(pAig, nOffset+iOut);
        Aig_ObjSetTravIdPrevious( pAig, pObj );
    }
    // count pure PIs
    nCountPis = nCountRegs = 0;
    // inlining of macro Aig_ManForEachPiSeq
    // inlining of macro Vec_PtrForEachEntryStop
    for  i = 0; (i < Aig_ManCiNum(pAig)-Aig_ManRegNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCis, i)), 1); i++  {
        nCountPis += Aig_ObjIsTravIdCurrent(pAig, pObj);
    }
    // count outputs of other registers
    // inlining of macro Aig_ManForEachLoSeq
    // inlining of macro Vec_PtrForEachEntryStart
    for  i = Aig_ManCiNum(pAig)-Aig_ManRegNum(pAig); (i < Vec_PtrSize(pAig.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCis, i)), 1); i++  {
        nCountRegs += Aig_ObjIsTravIdCurrent(pAig, pObj);
    } 
    if pnCountPis != nil {
        *pnCountPis = nCountPis; }
    if pnCountRegs != nil {
        *pnCountRegs = nCountRegs; }
    // create the new manager
    pNew = Aig_ManStart( Vec_PtrSize(vNodes) );
    Aig_ManConst1(pAig).pData = Aig_ManConst1(pNew);
    // create the PIs
    // inlining of macro Aig_ManForEachCi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pAig.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCis, i)), 1); i++  {
        if Aig_ObjIsTravIdCurrent(pAig, pObj) {
            pObj.pData = Aig_ObjCreateCi(pNew); }
    }
    // add variables for the register outputs
    // create fake POs to hold the register outputs
    nOffset = Aig_ManCiNum(pAig)-Aig_ManRegNum(pAig);
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vPart)) && (((iOut) = Vec_IntEntry(vPart, i)), 1); i++  {
        pObj = Aig_ManCi(pAig, nOffset+iOut);
        pObj.pData = Aig_ObjCreateCi(pNew);
        Aig_ObjCreateCo( pNew, (Aig_Obj_t *)pObj.pData );
        Aig_ObjSetTravIdCurrent( pAig, pObj ); // added
    }
    // create the nodes
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        if Aig_ObjIsNode(pObj) {
            pObj.pData = Aig_And(pNew, Aig_ObjChild0Copy(pObj), Aig_ObjChild1Copy(pObj) ); }
    }
    // add real POs for the registers
    nOffset = Aig_ManCoNum(pAig)-Aig_ManRegNum(pAig);
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vPart)) && (((iOut) = Vec_IntEntry(vPart, i)), 1); i++  {
        pObj = Aig_ManCo( pAig, nOffset+iOut );
        Aig_ObjCreateCo( pNew, Aig_ObjChild0Copy(pObj) );
    }
    pNew.nRegs = Vec_IntSize(vPart);
    // create map
    if ppMapBack {
        // inlining of macro ABC_ALLOC
        pMapBack = ((int *) malloc(sizeof(int) * (size_t)(Aig_ManObjNumMax(pNew))));
        memset( pMapBack, 0xff, sizeof(int) * Aig_ManObjNumMax(pNew) );
        // map constant nodes
        pMapBack[0] = 0;
        // logic cones of register outputs
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
            pObjNew = Aig_Regular((Aig_Obj_t *)pObj.pData);
            pMapBack[pObjNew.Id] = pObj.Id;
        }
        // map register outputs
        nOffset = Aig_ManCiNum(pAig)-Aig_ManRegNum(pAig);
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(vPart)) && (((iOut) = Vec_IntEntry(vPart, i)), 1); i++  {
            pObj = Aig_ManCi(pAig, nOffset+iOut);
            pObjNew = (Aig_Obj_t *)pObj.pData;
            pMapBack[pObjNew.Id] = pObj.Id;
        }
        *ppMapBack = pMapBack;
    }
    Vec_PtrFree( vNodes );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Computes partitioning of registers.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManRegPartitionSmart(pAig *Aig_Man_t, nPartSize int32) *Vec_Ptr_t {
    extern void Ioa_WriteAiger( Aig_Man_t * pMan, char * pFileName, int fWriteSymbols, int fCompact );

    var p *Aig_ManPre_t;
    var vResult *Vec_Ptr_t;
    var iSeed, iNext, i, k int32;
    // create the manager
    p = Aig_ManRegManStart( pAig, nPartSize );
    // add partitions as long as registers remain
    for  i = 0; (iSeed = Aig_ManRegFindSeed(p)) >= 0; i++  {
//printf( "Seed variable = %d.\n", iSeed );
        // clean the current partition information
        Vec_IntClear( p.vRegs );
        Vec_IntClear( p.vUniques );
        Vec_IntClear( p.vFreeVars );
        Vec_FltClear( p.vPartCost );
        memset( p.pfPartVars, 0, sizeof(char) * Aig_ManRegNum(p.pAig) );
        // add the register and its partition support
        Aig_ManRegPartitionAdd( p, iSeed );
        // select the best var to add
        for  k = 0; Vec_IntSize(p.vRegs) < p.nRegsMax; k++  {
            // get the next best variable
            iNext = Aig_ManRegFindBestVar( p );
            if iNext == -1 {
                break; }
            // add the register to the support of the partition
            Aig_ManRegPartitionAdd( p, iNext );
            // report the result
//printf( "Part %3d  Reg %3d : Free = %4d. Total = %4d. Ratio = %6.2f. Unique = %4d.\n", i, k, 
//                Vec_IntSize(p->vFreeVars), Vec_IntSize(p->vRegs), 
//                1.0*Vec_IntSize(p->vFreeVars)/Vec_IntSize(p->vRegs), Vec_IntSize(p->vUniques) );
            // quit if there are not free variables
            if Vec_IntSize(p.vFreeVars) == 0 {
                break; }
        }
        // add this partition to the set
        Vec_PtrPush( p.vParts, Vec_IntDup(p.vRegs) );        
fmt.Printf("Part %3d  SUMMARY:  Free = %4d. Total = %4d. Ratio = %6.2f. Unique = %4d.\n", i, Vec_IntSize(p.vFreeVars), Vec_IntSize(p.vRegs), 1.0*Vec_IntSize(p.vFreeVars)/Vec_IntSize(p.vRegs), Vec_IntSize(p.vUniques));
//printf( "\n" ); 
    }
    vResult = p.vParts; p.vParts = nil;
    Aig_ManRegManStop( p );
    return vResult;
}

/**Function*************************************************************

  Synopsis    [Computes partitioning of registers.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManRegPartitionSimple(pAig *Aig_Man_t, nPartSize int32, nOverSize int32) *Vec_Ptr_t {
    var vResult *Vec_Ptr_t;
    var vPart *Vec_Int_t;
    var i, Counter int32;
    if nOverSize >= nPartSize {
        fmt.Printf("Overlap size (%d) is more or equal than the partition size (%d).\n", nOverSize, nPartSize);
        fmt.Printf("Adjusting it to be equal to half of the partition size.\n");
        nOverSize = nPartSize/2;
    }
    assert(nOverSize < nPartSize, "nOverSize < nPartSize");
    vResult = Vec_PtrAlloc( 100 );
    for  Counter = 0; Counter < Aig_ManRegNum(pAig); Counter -= nOverSize  {
        vPart = Vec_IntAlloc( nPartSize );
        for  i = 0; i < nPartSize; i, Counter = i+1, Counter+1  {
            if Counter < Aig_ManRegNum(pAig) {
                Vec_IntPush( vPart, Counter ); }
        }
        if Vec_IntSize(vPart) <= nOverSize {
            Vec_IntFree(vPart);
        } else {
            Vec_PtrPush( vResult, vPart ); }
    }
    return vResult;
}


/**Function*************************************************************

  Synopsis    [Divides a large partition into several ones.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManPartDivide(vResult *Vec_Ptr_t, vDomain *Vec_Int_t, nPartSize int32, nOverSize int32) {
    var vPart *Vec_Int_t;
    var i, Counter int32;
    assert((nPartSize != 0) && Vec_IntSize(vDomain) > nPartSize, "nPartSize && Vec_IntSize(vDomain) > nPartSize");
    if nOverSize >= nPartSize {
        fmt.Printf("Overlap size (%d) is more or equal than the partition size (%d).\n", nOverSize, nPartSize);
        fmt.Printf("Adjusting it to be equal to half of the partition size.\n");
        nOverSize = nPartSize/2;
    }
    assert(nOverSize < nPartSize, "nOverSize < nPartSize");
    for  Counter = 0; Counter < Vec_IntSize(vDomain); Counter -= nOverSize  {
        vPart = Vec_IntAlloc( nPartSize );
        for  i = 0; i < nPartSize; i, Counter = i+1, Counter+1  {
            if Counter < Vec_IntSize(vDomain) {
                Vec_IntPush( vPart, Vec_IntEntry(vDomain, Counter) ); }
        }
        if Vec_IntSize(vPart) <= nOverSize {
            Vec_IntFree(vPart);
        } else {
            Vec_PtrPush( vResult, vPart ); }
    }
}

/**Function*************************************************************

  Synopsis    [Computes partitioning of registers.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManRegPartitionTraverse_rec(p *Aig_Man_t, pObj *Aig_Obj_t, vLos *Vec_Ptr_t) {
    if Aig_ObjIsTravIdCurrent(p, pObj) {
        return; }
    Aig_ObjSetTravIdCurrent( p, pObj );
    if Aig_ObjIsCi(pObj) {
        if pObj.iData >= Aig_ManCiNum(p) - Aig_ManRegNum(p) {
            Vec_PtrPush( vLos, pObj );
            fmt.Printf("%d ", pObj.iData - (Aig_ManCiNum(p) - Aig_ManRegNum(p)));
        }
        return;
    }
    Aig_ManRegPartitionTraverse_rec( p, Aig_ObjFanin0(pObj), vLos );
    Aig_ManRegPartitionTraverse_rec( p, Aig_ObjFanin1(pObj), vLos );
}

/**Function*************************************************************

  Synopsis    [Computes partitioning of registers.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManRegPartitionTraverse(p *Aig_Man_t) *Vec_Ptr_t {
    var vLos *Vec_Ptr_t;
    var pObj *Aig_Obj_t;
    var i, nPrev, Counter int32;
    // mark the registers
    // inlining of macro Aig_ManForEachCi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCis, i)), 1); i++  {
       pObj.iData = i;
    } 
    // collect registers
    vLos = Vec_PtrAlloc( Aig_ManRegNum(p) );
    nPrev = 0;
    Counter = 0;
    Aig_ManIncrementTravId( p );
    // inlining of macro Aig_ManForEachLiSeq
    // inlining of macro Vec_PtrForEachEntryStart
    for  i = Aig_ManCoNum(p)-Aig_ManRegNum(p); (i < Vec_PtrSize(p.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
        ++Counter;
        fmt.Printf("Latch %d: ", Counter);
        Aig_ManRegPartitionTraverse_rec( p, Aig_ObjFanin0(pObj), vLos );
fmt.Printf("%d=%d \n", Counter, Vec_PtrSize(vLos)-nPrev);
        nPrev = Vec_PtrSize(vLos);
    }
    fmt.Printf("Total collected = %d. Total regs = %d.\n", Vec_PtrSize(vLos), Aig_ManRegNum(p));
    return vLos;
}

/**Function*************************************************************

  Synopsis    [Computes partitioning of registers.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManRegPartitionLinear(pAig *Aig_Man_t, nPartSize int32) *Vec_Ptr_t {
    var vLos *Vec_Ptr_t;
    vLos = Aig_ManRegPartitionTraverse( pAig );
    Vec_PtrFree( vLos );
    return nil;
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

