package acec

import "fmt"

/**CFile****************************************************************

  FileName    [acecPolyn.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [CEC for arithmetic circuits.]

  Synopsis    [Polynomial extraction.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: acecPolyn.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "acecInt.h"
//#include "misc/vec/vecWec.h"
//#include "misc/vec/vecHsh.h"
//#include "misc/vec/vecQue.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

/*
!a            ->   1 - a
a & b         ->   ab
a | b         ->   a + b - ab
a ^ b         ->   a + b - 2ab
MUX(a, b, c)  ->   ab | (1 - a)c = ab + (1-a)c - ab(1-a)c = ab + c - ac

!a & b        ->   (1 - a)b = b - ab
 a & !b       ->   a(1 - b) = a - ab
!a & !b       ->   1 - a - b + ab
*/

   
type Pln_Man_t struct {
    pGia     *Gia_Man_t;      // AIG manager
    pHashC  *Hsh_VecMan_t;    // hash table for constants
    pHashM  *Hsh_VecMan_t;    // hash table for monomials
    vQue     *Vec_Que_t;      // queue by largest node
    vCounts     *Vec_Flt_t;   // largest node
    vCoefs     *Vec_Int_t;    // coefficients for each monomial
    Vec_Int_t *    vTempC[2]; // polynomial representation
    Vec_Int_t *    vTempM[4]; // polynomial representation
    vOrder     *Vec_Int_t;    // order of collapsing
    nBuilds            int32;   // built monomials
    nUsed            int32;     // used monomials
}

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Computation manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Pln_ManAlloc(pGia *Gia_Man_t, vOrder *Vec_Int_t) *Pln_Man_t {
    // inlining of macro ABC_CALLOC
    var p *Pln_Man_t  = &Pln_Man_t{};
    p.pGia      = pGia;
    p.pHashC    = Hsh_VecManStart( 1000 );
    p.pHashM    = Hsh_VecManStart( 1000 );
    p.vQue      = Vec_QueAlloc( 1000 );
    p.vCounts   = Vec_FltAlloc( 1000 );
    p.vCoefs    = Vec_IntAlloc( 1000 );
    p.vTempC[0] = Vec_IntAlloc( 100 );
    p.vTempC[1] = Vec_IntAlloc( 100 );
    p.vTempM[0] = Vec_IntAlloc( 100 );
    p.vTempM[1] = Vec_IntAlloc( 100 );
    p.vTempM[2] = Vec_IntAlloc( 100 );
    p.vTempM[3] = Vec_IntAlloc( 100 );
    if vOrder != nil {
        p.vOrder = Vec_IntDup(vOrder);
    } else {
        p.vOrder = Vec_IntStartNatural( Gia_ManObjNum(pGia) );
    }
    assert(Vec_IntSize(p.vOrder) == Gia_ManObjNum(pGia), "Vec_IntSize(p.vOrder) == Gia_ManObjNum(pGia)");
    Vec_QueSetPriority( p.vQue, Vec_FltArrayP(p.vCounts) );
    // add 0-constant and 1-monomial
    Hsh_VecManAdd( p.pHashC, p.vTempC[0] );
    Hsh_VecManAdd( p.pHashM, p.vTempM[0] );
    Vec_FltPush( p.vCounts, 0 );
    Vec_IntPush( p.vCoefs, 0 );
    return p;
}
func Pln_ManStop(p *Pln_Man_t) {
    Hsh_VecManStop( p.pHashC );
    Hsh_VecManStop( p.pHashM );
    Vec_QueFree( p.vQue );
    Vec_FltFree( p.vCounts );
    Vec_IntFree( p.vCoefs );
    Vec_IntFree( p.vTempC[0] );
    Vec_IntFree( p.vTempC[1] );
    Vec_IntFree( p.vTempM[0] );
    Vec_IntFree( p.vTempM[1] );
    Vec_IntFree( p.vTempM[2] );
    Vec_IntFree( p.vTempM[3] );
    Vec_IntFree( p.vOrder );
    p = nil;
}
func Pln_ManCompare3(pData0 *int32, pData1 *int32) int32 {
    if pData0[0] < pData1[0] { return -1; }
    if pData0[0] > pData1[0] { return  1; }
    if pData0[1] < pData1[1] { return -1; }
    if pData0[1] > pData1[1] { return  1; }
    return 0;
}
func Pln_ManPrintFinal(p *Pln_Man_t, fVerbose int32, fVeryVerbose int32) {
    var vArray *Vec_Int_t;
    var i, k, Entry, iMono, iConst int32;
    // collect triples
    var vPairs *Vec_Int_t  = Vec_IntAlloc( 100 );
    // inlining of macro Vec_IntForEachEntry
    for  iMono = 0; (iMono < Vec_IntSize(p.vCoefs)) && (((iConst) = Vec_IntEntry(p.vCoefs, iMono)), 1); iMono++  {
        if iConst == 0 { 
            continue; }
        vArray = Hsh_VecReadEntry( p.pHashC, iConst );
        Vec_IntPush( vPairs, Vec_IntEntry(vArray, 0) );
        vArray = Hsh_VecReadEntry( p.pHashM, iMono );
        Vec_IntPush( vPairs, Vec_IntSize(vArray) ? Vec_IntEntry(vArray, 0) : 0 );
        Vec_IntPushTwo( vPairs, iConst, iMono );
    }
    // sort triples
    qsort( Vec_IntArray(vPairs), (size_t)(Vec_IntSize(vPairs)/4), 16, (int (*)(void *, void *))Pln_ManCompare3 );
    // print
    if fVerbose != 0 {
    // inlining of macro Vec_IntForEachEntryDouble
    for  i = 0; (i+1 < Vec_IntSize(vPairs)) && (((iConst) = Vec_IntEntry(vPairs, i)), 1) && (((iMono) = Vec_IntEntry(vPairs, i+1)), 1); i += 2  {
        if i % 4 == 0 {
            continue; }
        fmt.Printf("%-6d : ", i/4);
        vArray = Hsh_VecReadEntry( p.pHashC, iConst );
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(vArray)) && (((Entry) = Vec_IntEntry(vArray, k)), 1); k++  {
            fmt.Printf("%s%d", Entry < 0 ? "-" : "+", (1 << (Abc_AbsInt(Entry)-1)));
        }
        vArray = Hsh_VecReadEntry( p.pHashM, iMono );
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(vArray)) && (((Entry) = Vec_IntEntry(vArray, k)), 1); k++  {
            fmt.Printf(" * %d", Entry);
        }
        fmt.Printf("\n");
    } }
    fmt.Printf("HashC = %d. HashM = %d.  Total = %d. Used = %d.  ", Hsh_VecSize(p.pHashC), Hsh_VecSize(p.pHashM), p.nBuilds, Vec_IntSize(vPairs)/4);
    Vec_IntFree( vPairs );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_IntAppendMinus2x(vVec1 *Vec_Int_t, vVec2 *Vec_Int_t) {
    var Entry, i int32;
    Vec_IntClear( vVec1 );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vVec2)) && (((Entry) = Vec_IntEntry(vVec2, i)), 1); i++  {
        Vec_IntPush( vVec1, Entry > 0 ? -Entry-1 : -Entry+1 );
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_PolynMergeConstOne(vConst *Vec_Int_t, New int32) {
    var i, Old int32;
    assert(New != 0, "New != 0");
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vConst)) && (((Old) = Vec_IntEntry(vConst, i)), 1); i++  {
        assert(Old != 0, "Old != 0");
        if Old == New { // A == B
        
            Vec_IntDrop( vConst, i );
            Gia_PolynMergeConstOne( vConst, New > 0 ? New + 1 : New - 1 );
            return;
        }
        if Abc_AbsInt(Old) == Abc_AbsInt(New) { // A == -B
        
            Vec_IntDrop( vConst, i );
            return;
        }
        if Old + New == 1 || Old + New == -1 {  // sign(A) != sign(B)  &&  abs(abs(A)-abs(B)) == 1 
        
            var Value int32  = Abc_MinInt( Abc_AbsInt(Old), Abc_AbsInt(New) );
            Vec_IntDrop( vConst, i );
            Gia_PolynMergeConstOne( vConst, (Old + New == 1) ? Value : -Value );
            return;
        }
    }
    Vec_IntPushUniqueOrder( vConst, New );
}
func Gia_PolynMergeConst(vConst *Vec_Int_t, p *Pln_Man_t, iConstAdd int32) {
    var i, New int32;
    var vConstAdd *Vec_Int_t  = Hsh_VecReadEntry( p.pHashC, iConstAdd );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vConstAdd)) && (((New) = Vec_IntEntry(vConstAdd, i)), 1); i++  {
        Gia_PolynMergeConstOne( vConst, New );
        vConstAdd = Hsh_VecReadEntry( p.pHashC, iConstAdd );
    }
}
func Gia_PolynBuildAdd(p *Pln_Man_t, vTempC *Vec_Int_t, vTempM *Vec_Int_t) {
    var iConst int32;
    var iConstNew int32;
    var iMono int32;
    if vTempM != nil {
        iMono = Hsh_VecManAdd(p.pHashM, vTempM);
    } else {
        iMono = 0;
    }
    p.nBuilds++;
    if iMono == Vec_IntSize(p.vCoefs) { // new monomial
    
        iConst = Hsh_VecManAdd( p.pHashC, vTempC );
        Vec_IntPush( p.vCoefs, iConst );
//        Vec_FltPush( p->vCounts, Vec_IntEntryLast(vTempM) );
        Vec_FltPush( p.vCounts, float32(Vec_IntEntry(p.vOrder, Vec_IntEntryLast(vTempM))) );
        Vec_QuePush( p.vQue, iMono );
//        Vec_QueUpdate( p->vQue, iMono );
        if iConst != 0 {
            p.nUsed++; }
        return;
    }
    // this monomial exists
    iConst = Vec_IntEntry( p.vCoefs, iMono );
    if iConst != 0 {
        Gia_PolynMergeConst( vTempC, p, iConst ); }
    iConstNew = Hsh_VecManAdd( p.pHashC, vTempC );
    Vec_IntWriteEntry( p.vCoefs, iMono, iConstNew );
    if (iConst != 0) && !(iConstNew != 0) {
        p.nUsed--;
    } else if !(iConst != 0) && (iConstNew != 0) {
        p.nUsed++; }
    //assert( p->nUsed == Vec_IntSize(p->vCoefs) - Vec_IntCountZero(p->vCoefs) );
}
func Gia_PolynBuildOne(p *Pln_Man_t, iMono int32) {
    var pObj *Gia_Obj_t; 
    var vArray, vConst *Vec_Int_t;
    var iFan0, iFan1 int32;
    var k, iConst, iDriver int32;

    assert(Vec_IntSize(p.vCoefs) == Hsh_VecSize(p.pHashM), "Vec_IntSize(p.vCoefs) == Hsh_VecSize(p.pHashM)");
    vArray  = Hsh_VecReadEntry( p.pHashM, iMono );
    iDriver = Vec_IntEntryLast( vArray );
    pObj    = Gia_ManObj( p.pGia, iDriver );
    if !Gia_ObjIsAnd(pObj) {
        return; }
    assert(!Gia_ObjIsMux(p.pGia, pObj), "!Gia_ObjIsMux(p.pGia, pObj)");

    iConst = Vec_IntEntry( p.vCoefs, iMono );
    if iConst == 0 {
        return; }
    Vec_IntWriteEntry( p.vCoefs, iMono, 0 );
    p.nUsed--;

    iFan0 = Gia_ObjFaninId0p(p.pGia, pObj);
    iFan1 = Gia_ObjFaninId1p(p.pGia, pObj);
    for  k = 0; k < 4; k++  {
        Vec_IntClear( p.vTempM[k] );
        Vec_IntAppend( p.vTempM[k], vArray );
        Vec_IntPop( p.vTempM[k] );
        if k == 1 || k == 3 {
            Vec_IntPushUniqueOrderCost( p.vTempM[k], iFan0, p.vOrder ); }    // x
//            Vec_IntPushUniqueOrder( p->vTempM[k], iFan0 );    // x
        if k == 2 || k == 3 {
            Vec_IntPushUniqueOrderCost( p.vTempM[k], iFan1, p.vOrder ); }    // y
//            Vec_IntPushUniqueOrder( p->vTempM[k], iFan1 );    // y
    }

    vConst = Hsh_VecReadEntry( p.pHashC, iConst );

    if !Gia_ObjIsXor(pObj) {
        for  k = 0; k < 2; k++  {
            Vec_IntAppendMinus( p.vTempC[k], vConst, k );
        } }

    if Gia_ObjIsXor(pObj) {
        vConst = Hsh_VecReadEntry( p.pHashC, iConst );
        Vec_IntAppendMinus( p.vTempC[0], vConst, 0 );
        Gia_PolynBuildAdd( p, p.vTempC[0], p.vTempM[1] );   //  C * x 

        vConst = Hsh_VecReadEntry( p.pHashC, iConst );
        Vec_IntAppendMinus( p.vTempC[0], vConst, 0 );
        Gia_PolynBuildAdd( p, p.vTempC[0], p.vTempM[2] );   //  C * y 

        //if ( !p->pGia->vXors || Vec_IntFind(p->pGia->vXors, iDriver) == -1 || Vec_IntFind(p->pGia->vXors, iDriver) == 5 )
        { 
            vConst = Hsh_VecReadEntry( p.pHashC, iConst );
            Vec_IntAppendMinus2x( p.vTempC[0], vConst );
            Gia_PolynBuildAdd( p, p.vTempC[0], p.vTempM[3] );   // -C * x * y 
        }
    
    } else if Gia_ObjFaninC0(pObj) && Gia_ObjFaninC1(pObj) {  //  C * (1 - x) * (1 - y)
    
        Gia_PolynBuildAdd( p, p.vTempC[0], p.vTempM[0] );   //  C * 1
        Gia_PolynBuildAdd( p, p.vTempC[1], p.vTempM[1] );   // -C * x
        vConst = Hsh_VecReadEntry( p.pHashC, iConst );
        for  k = 0; k < 2; k++  {
            Vec_IntAppendMinus( p.vTempC[k], vConst, k );
        }
        Gia_PolynBuildAdd( p, p.vTempC[1], p.vTempM[2] );   // -C * y 
        Gia_PolynBuildAdd( p, p.vTempC[0], p.vTempM[3] );   //  C * x * y
    
    } else if Gia_ObjFaninC0(pObj) && !Gia_ObjFaninC1(pObj) { //  C * (1 - x) * y
    
        Gia_PolynBuildAdd( p, p.vTempC[0], p.vTempM[2] );   //  C * y 
        Gia_PolynBuildAdd( p, p.vTempC[1], p.vTempM[3] );   // -C * x * y
    
    } else if !Gia_ObjFaninC0(pObj) && Gia_ObjFaninC1(pObj) { //  C * x * (1 - y)
    
        Gia_PolynBuildAdd( p, p.vTempC[0], p.vTempM[1] );   //  C * x 
        Gia_PolynBuildAdd( p, p.vTempC[1], p.vTempM[3] );   // -C * x * y
    
    } else {   
        Gia_PolynBuildAdd( p, p.vTempC[0], p.vTempM[3] ); }   //  C * x * y
}
func Gia_PolynBuild(pGia *Gia_Man_t, vOrder *Vec_Int_t, fSigned int32, fVerbose int32, fVeryVerbose int32) {
    var clk abctime  = Abc_Clock();//, clk2 = 0;
    var pObj *Gia_Obj_t; 
    var vPres *Vec_Bit_t  = Vec_BitStart( Gia_ManObjNum(pGia) );
    var i, iMono, iDriver, LevPrev, LevCur, Iter, Line int32;
    var p *Pln_Man_t  = Pln_ManAlloc( pGia, vOrder );
    // inlining of macro Gia_ManForEachCoReverse
    for  i = Vec_IntSize(pGia.vCos) - 1; (i >= 0) && ((pObj) = Gia_ManCo(pGia, i)); i--  {
        Vec_IntFill( p.vTempC[0], 1,  i+1 );      //  2^i
        Vec_IntFill( p.vTempC[1], 1, -i-1 );      // -2^i

        iDriver = Gia_ObjFaninId0p( pGia, pObj );
        Vec_IntFill( p.vTempM[0], 1, iDriver );   //  Driver

        if (fSigned != 0) && i == Gia_ManCoNum(pGia)-1 {
            if Gia_ObjFaninC0(pObj) {
                Gia_PolynBuildAdd( p, p.vTempC[1], nil );           // -C
                Gia_PolynBuildAdd( p, p.vTempC[0], p.vTempM[0] );   //  C * Driver
            
            } else {
                Gia_PolynBuildAdd( p, p.vTempC[1], p.vTempM[0] ); }   // -C * Driver
        
        } else {
            if Gia_ObjFaninC0(pObj) {
                Gia_PolynBuildAdd( p, p.vTempC[0], nil );           //  C
                Gia_PolynBuildAdd( p, p.vTempC[1], p.vTempM[0] );   // -C * Driver
            
            } else {
                Gia_PolynBuildAdd( p, p.vTempC[0], p.vTempM[0] ); }   //  C * Driver
        }
    }
    LevPrev = -1;
    for  Iter = 0; ; Iter++  {
        var vTempM *Vec_Int_t;
        //abctime temp = Abc_Clock();
        if Vec_QueSize(p.vQue) == 0 {
            break; }
        iMono = Vec_QuePop(p.vQue);

        // report
        vTempM = Hsh_VecReadEntry( p.pHashM, iMono );
        //printf( "Removing var %d\n", Vec_IntEntryLast(vTempM) );
        LevCur = Vec_IntEntryLast(vTempM);
        if !Gia_ObjIsAnd(Gia_ManObj(pGia, LevCur)) {
            continue; }

        if LevPrev != LevCur {
            if Vec_BitEntry( vPres, LevCur ) && (fVerbose != 0) {
                fmt.Printf("Repeating entry %d\n", LevCur);
            } else {
                Vec_BitSetEntry( vPres, LevCur, 1 ); }

            if fVeryVerbose != 0 {
                fmt.Printf("Line%5d   Iter%10d : Obj =%6d.  Order =%6d.  HashC =%6d. HashM =%10d.  Total =%10d. Used =%10d.\n", Line++, Iter, LevCur, Vec_IntEntry(p.vOrder, LevCur), Hsh_VecSize(p.pHashC), Hsh_VecSize(p.pHashM), p.nBuilds, p.nUsed); }
        }
        LevPrev = LevCur;

        Gia_PolynBuildOne( p, iMono );
        //clk2 += Abc_Clock() - temp;
    }
    //Abc_PrintTime( 1, "Time2", clk2 );
    Pln_ManPrintFinal( p, fVerbose, fVeryVerbose );
    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
    Pln_ManStop( p );
    Vec_BitFree( vPres );
}



/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_PolynBuild2(pGia *Gia_Man_t, fSigned int32, fVerbose int32, fVeryVerbose int32) {
    var pHashC *Hsh_VecMan_t  = Hsh_VecManStart( 1000 );    // hash table for constants
    var pHashM *Hsh_VecMan_t  = Hsh_VecManStart( 1000 );    // hash table for monomials
    //Vec_Wec_t * vLit2Mono = Vec_WecStart( Gia_ManObjNum(pGia) * 2 );

    Hsh_VecManStop( pHashC );
    Hsh_VecManStop( pHashM );
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

