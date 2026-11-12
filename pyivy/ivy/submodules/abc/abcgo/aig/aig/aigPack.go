package aig

import "fmt"

/**CFile****************************************************************
 
  FileName    [aigPack.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [AIG package.]

  Synopsis    [Bit-packing code.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - April 28, 2007.]

  Revision    [$Id: aigPack.c,v 1.00 2007/04/28 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "aig.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// resubstitution manager
   
type Aig_ManPack_t struct {
    pAig         *Aig_Man_t;        // working manager 
    vSigns         *Vec_Wrd_t;      // object signatures
    vPiPats         *Vec_Wrd_t;     // PI assignments
    vPiCare         *Vec_Wrd_t;     // PI care set
    iPatCur                int32;     // current pattern
    fVerbose                int32;    // verbosiness flag
    // statistics
    nPatTotal                int32;   // number of all patterns
    nPatSkip                int32;    // number of skipped patterns
    nPatRepeat                int32;  // number of repeated patterns
}

func Aig_Word6CountOnes(t word) int32 { return Aig_WordCountOnes( uint32((t >> 32)) ) + Aig_WordCountOnes( uint32((t & 0xFFFFFFFF)) ); }
func Aig_Word6HasOneBit(t word) int32 { return (t & (t-1)) == 0; }


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManPackAlloc(pAig *Aig_Man_t) *Aig_ManPack_t {
    
    // inlining of macro ABC_CALLOC
    p := &Aig_ManPack_t{};
    p.pAig    = pAig;
    p.vSigns  = Vec_WrdStart( Aig_ManObjNumMax(pAig) );
    p.vPiPats = Vec_WrdStart( Aig_ManCiNum(pAig) );
    p.vPiCare = Vec_WrdStart( Aig_ManCiNum(pAig) );
    p.iPatCur = 1;
    return p;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManPackCountCares(p *Aig_ManPack_t) int32 {
    var pObj *Aig_Obj_t;
    var i, Total int32;
    // inlining of macro Aig_ManForEachCi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pAig.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCis, i)), 1); i++  {
        Total += Aig_Word6CountOnes( Vec_WrdEntry(p.vPiCare, i) );
    }
    return Total;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManPackPrintCare(p *Aig_ManPack_t) {
    var pObj *Aig_Obj_t;
    var Sign word;
    var i int32;
    // inlining of macro Aig_ManForEachCi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pAig.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCis, i)), 1); i++  {
        Sign = Vec_WrdEntry( p.vPiCare, i );
//        Extra_PrintBinary( stdout, (unsigned *)&Sign, 64 );
//        printf( "\n" );
    }
//    printf( "\n" );
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManPackFree(p *Aig_ManPack_t) {
//    Aig_ManPackPrintCare( p );
    fmt.Printf("Patterns: ");
    fmt.Printf("Total = %6d. ", p.nPatTotal);
    fmt.Printf("Skipped = %6d. ", p.nPatSkip);
    fmt.Printf("Cares = %6.2f %%  ", 100.0*Aig_ManPackCountCares(p)/Aig_ManCiNum(p.pAig)/64);
    fmt.Printf("\n");
    Vec_WrdFree( p.vSigns );
    Vec_WrdFree( p.vPiPats );
    Vec_WrdFree( p.vPiCare );
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManPackSetRandom(p *Aig_ManPack_t) {
    var pObj *Aig_Obj_t;
    var Sign word;
    var i int32;
    // inlining of macro Aig_ManForEachCi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pAig.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCis, i)), 1); i++  {
        Sign = (((word)Aig_ManRandom(0)) << 32) | ((word)Aig_ManRandom(0));
        Vec_WrdWriteEntry( p.vPiPats, i, Sign << 1 );
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManPackSimulate(p *Aig_ManPack_t) {
    var pObj *Aig_Obj_t;
    var Sign, Sign0, Sign1 word;
    var i int32;
    // set the constant
    Vec_WrdWriteEntry( p.vSigns, 0, ~(word)0 );
    // transfer into the array
    // inlining of macro Aig_ManForEachCi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pAig.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCis, i)), 1); i++  {
        Vec_WrdWriteEntry( p.vSigns, Aig_ObjId(pObj), Vec_WrdEntry(p.vPiPats, i) );
    }
    // simulate internal nodes
    // inlining of macro Aig_ManForEachNode
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
        Sign0 = Vec_WrdEntry( p.vSigns, Aig_ObjFaninId0(pObj) );
        Sign1 = Vec_WrdEntry( p.vSigns, Aig_ObjFaninId1(pObj) );
        if Aig_ObjFaninC0(pObj) && Aig_ObjFaninC1(pObj) {
            Sign = ~(Sign0 | Sign1);
        } else if Aig_ObjFaninC0(pObj) {
            Sign = ~Sign0 &  Sign1;
        } else if Aig_ObjFaninC1(pObj) {
            Sign =  Sign0 & ~Sign1;
        } else {
            Sign =  Sign0 &  Sign1; }
        Vec_WrdWriteEntry( p.vSigns, Aig_ObjId(pObj), Sign );
    }
    }
    // set the outputs
    // inlining of macro Aig_ManForEachCo
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pAig.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCos, i)), 1); i++  {
        Sign0 = Vec_WrdEntry( p.vSigns, Aig_ObjFaninId0(pObj) );
        if Aig_ObjFaninC0(pObj) {
            Sign = ~Sign0;
        } else {
            Sign = Sign0;
        }
        Vec_WrdWriteEntry( p.vSigns, Aig_ObjId(pObj), Sign );
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManPackPrintStats(p *Aig_ManPack_t) {
    var Sign word;
    var pObj *Aig_Obj_t;
    var i int32;
    var Total int32;
    var Count int32;
    var Counts [33]int32 ; // the number of nodes having that many patterns
    // inlining of macro Aig_ManForEachNode
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
        Sign  = Vec_WrdEntry( p.vSigns, Aig_ObjId(pObj) );
        Count = Aig_Word6CountOnes( Sign );
        if Count > 32 {
            Count = 64 - Count; }
        Counts[Count]++;        
    }
    }
    // print statistics
    Total = 0;
    for  i = 0; i <= 32; i++  {
        Total += Counts[i];
        fmt.Printf("%2d : ", i);
        fmt.Printf("%6d  ", Counts[i]);
        fmt.Printf("%6.1f %%", 100.0*Counts[i]/Aig_ManNodeNum(p.pAig));
        fmt.Printf("%6.1f %%", 100.0*Total/Aig_ManNodeNum(p.pAig));
        fmt.Printf("\n");
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManPackConstNodes(p *Aig_ManPack_t) *Vec_Int_t {
    var vNodes *Vec_Int_t;
    var pObj *Aig_Obj_t;
    var Sign word;
    var i int32;
    vNodes = Vec_IntAlloc( 1000 );
    // inlining of macro Aig_ManForEachNode
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
        Sign  = Vec_WrdEntry( p.vSigns, Aig_ObjId(pObj) );
        if Sign == 0 || ~Sign == 0 || (Aig_Word6HasOneBit(Sign) != 0) || (Aig_Word6HasOneBit(~Sign) != 0) {
            Vec_IntPush( vNodes, Aig_ObjId(pObj) ); }
    }
    }
    return vNodes;
}


/**Function*************************************************************

  Synopsis    [Packs patterns into array of simulation info.]

  Description []
               
  SideEffects []

  SeeAlso     []

*************************************`**********************************/
func Aig_ManPackAddPatternTry(p *Aig_ManPack_t, iBit int32, vLits *Vec_Int_t) int32 {
    var pInfo, pPres *word;
    var i, Lit int32;
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vLits)) && (((Lit) = Vec_IntEntry(vLits, i)), 1); i++  {
        pInfo = Vec_WrdEntryP( p.vPiPats, Abc_Lit2Var(Lit) );
        pPres = Vec_WrdEntryP( p.vPiCare, Abc_Lit2Var(Lit) );
        if Abc_InfoHasBit( (unsigned *)pPres, iBit ) && 
             Abc_InfoHasBit( (unsigned *)pInfo, iBit ) == Abc_LitIsCompl(Lit) {
             return 0; }
    }
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vLits)) && (((Lit) = Vec_IntEntry(vLits, i)), 1); i++  {
        pInfo = Vec_WrdEntryP( p.vPiPats, Abc_Lit2Var(Lit) );
        pPres = Vec_WrdEntryP( p.vPiCare, Abc_Lit2Var(Lit) );
        Abc_InfoSetBit( (unsigned *)pPres, iBit );
        if Abc_InfoHasBit( (unsigned *)pInfo, iBit ) == Abc_LitIsCompl(Lit) {
             Abc_InfoXorBit( (unsigned *)pInfo, iBit ); }
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManPackAddPattern(p *Aig_ManPack_t, vLits *Vec_Int_t) {
    var k int32;
    for  k = 1; k < 64; k++  {
        if Aig_ManPackAddPatternTry( p, k, vLits ) != 0 {
            break; }
    }
    if k == 64 {
/*
        word * pInfo, * pPres;
        int i, Lit;
        Vec_IntForEachEntry( vLits, Lit, i )
            printf( "%d", Abc_LitIsCompl(Lit) );
        printf( "\n\n" );
        for ( k = 1; k < 64; k++ )
        {
            Vec_IntForEachEntry( vLits, Lit, i )
            {
                pInfo = Vec_WrdEntryP( p->vPiPats, Abc_Lit2Var(Lit) );
                pPres = Vec_WrdEntryP( p->vPiCare, Abc_Lit2Var(Lit) );
                if ( Abc_InfoHasBit( (unsigned *)pPres, k ) )
                    printf( "%d", Abc_InfoHasBit( (unsigned *)pInfo, k ) );
                else
                    printf( "-" );
            }
            printf( "\n" );
        }
*/
        p.nPatSkip++;
    }
    p.nPatTotal++;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManPackStart(pAig *Aig_Man_t) *Aig_ManPack_t {
    var p *Aig_ManPack_t;
    p = Aig_ManPackAlloc( pAig );
    Aig_ManPackSetRandom( p );
    Aig_ManPackSimulate( p );
    Aig_ManPackPrintStats( p );
    return p;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManPackStop(p *Aig_ManPack_t) {
    Aig_ManPackSimulate( p );
    Aig_ManPackPrintStats( p );
    Aig_ManPackFree( p );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

