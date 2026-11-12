package cba

import (
	"fmt"
	"math"
	"strconv"
)

/**CFile****************************************************************

  FileName    [cbaNtk.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Hierarchical word-level netlist.]

  Synopsis    [Network manipulation.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - July 21, 2015.]

  Revision    [$Id: cbaNtk.c,v 1.00 2014/11/29 00:00:00 alanmi Exp $]

***********************************************************************/

//#include <math.h>
//#include "cba.h"
//#include "base/main/main.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Prints distribution of operator types.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_WrdSelectSortCost2(pArray *word, nSize int32, pCosts *word) {
    var i, j, best_i int32;
    for  i = 0; i < nSize-1; i++  {
        best_i = i;
        for  j = i+1; j < nSize; j++  {
            if pCosts[j] < pCosts[best_i] {
                best_i = j; }
        }
        // inlining of macro ABC_SWAP
        { var t word  = pArray[i]; pArray[i] = pArray[best_i]; pArray[best_i] = t; };
        // inlining of macro ABC_SWAP
        { var t word  = pCosts[i]; pCosts[i] = pCosts[best_i]; pCosts[best_i] = t; };
    }
}
func Cba_NtkPrintDistribMakeSign(s int32, s0 int32, s1 int32) word {
    return ((word)s1 << 42) | ((word)s0 << 21) | (word)s;
}
func Cba_NtkPrintDistribFromSign(sss word, s *int32, s0 *int32, s1 *int32) {
    *s1 =  int32((sss >> 42));  *s0 = int32((sss >> 21)) & 0x1FFFFF;  *s  =  int32(sss) & 0x1FFFFF;
}
func Cba_NtkPrintDistribAddOne(vTypes *Vec_Ptr_t, vOccurs *Vec_Ptr_t, Type int32, Sign word) {
    var vType *Vec_Wrd_t   = (Vec_Wrd_t *)Vec_PtrEntry( vTypes, Type );
    var vOccur *Vec_Wrd_t  = (Vec_Wrd_t *)Vec_PtrEntry( vOccurs, Type );
    var Entry word; var i int32;
    // inlining of macro Vec_WrdForEachEntry
    for  i = 0; (i < Vec_WrdSize(vType)) && (((Entry) = Vec_WrdEntry(vType, i)), 1); i++  {
        if Entry == Sign {
            Vec_WrdAddToEntry( vOccur, i, 1 );
            return;
        }
    }
    Vec_WrdPush( vType, Sign );
    Vec_WrdPush( vOccur, 1 );
}
func Cba_NtkPrintDistribSortOne(vTypes *Vec_Ptr_t, vOccurs *Vec_Ptr_t, Type int32) {
    var vType *Vec_Wrd_t   = (Vec_Wrd_t *)Vec_PtrEntry( vTypes, Type );
    var vOccur *Vec_Wrd_t  = (Vec_Wrd_t *)Vec_PtrEntry( vOccurs, Type );
    Vec_WrdSelectSortCost2( Vec_WrdArray(vType), Vec_WrdSize(vType), Vec_WrdArray(vOccur) );
    Vec_WrdReverseOrder( vType );
    Vec_WrdReverseOrder( vOccur );
}
func Cba_NtkPrintDistrib(p *Cba_Ntk_t, fVerbose int32) {
    var vTypes, vOccurs *Vec_Ptr_t;
    var vAnds *Vec_Int_t  = Vec_IntStart( CBA_BOX_LAST );
    var iRnObj int32 = -1;
    var nCountRange int32;
    var i, k, s, s0, s1 int32;  var Sign word;
    var Type Cba_ObjType_t;
    var pTypeNames [CBA_BOX_LAST]string;
    Cba_ManCreatePrimMap( pTypeNames );
    // allocate statistics arrays
    vTypes  = Vec_PtrStart( CBA_BOX_LAST );
    vOccurs = Vec_PtrStart( CBA_BOX_LAST );
    for  i = 0; i < CBA_BOX_LAST; i++  {
        Vec_PtrWriteEntry( vTypes, i, Vec_WrdAlloc(16) );
    }
    for  i = 0; i < CBA_BOX_LAST; i++  {
        Vec_PtrWriteEntry( vOccurs, i, Vec_WrdAlloc(16) );
    }
    // add nodes
    // inlining of macro Cba_NtkForEachObj
    for  i = 1; i < Vec_StrSize(&p.vObjType); i++  { if !Cba_ObjType(p, i) {} else {
//        char * pName = Cba_ObjName(p, i);
        Type = Cba_ObjType( p, i );
        if Cba_ObjSign(p, i) > 0x1FFFFF {
            fmt.Printf("Object %6d has range %d, which is reduced to %d in the statistics.\n", i, Cba_ObjRangeSize(p, i), Cba_ObjRangeSize(p, i) & 0xFFFFF); }
        if Cba_ObjLeft(p, i) && Cba_ObjRight(p, i) {
            if iRnObj == -1 {
                iRnObj = 1; }
            nCountRange++;
        }
        // 0-input types
        if Cba_ObjIsPi(p, i) || (Type == CBA_BOX_BUF && Cba_FonIsConst(Cba_ObjFinFon(p, i, 0))) || Type == CBA_BOX_CONCAT {
            Sign = Cba_NtkPrintDistribMakeSign( Cba_ObjSign(p, i), 0, 0 );
        // 1-input types
        } else if Cba_TypeIsUnary(Type) {
            Sign = Cba_NtkPrintDistribMakeSign( Cba_ObjSign(p, i), Cba_ObjSign(p, Cba_ObjFinFon(p, i, 0)), 0 );
        // 2-input types (including MUX)
        } else if Cba_ObjFinNum(p, i) == 1 {
            Sign = Cba_NtkPrintDistribMakeSign( Cba_ObjSign(p, i), Cba_ObjSign(p, Cba_ObjFinFon(p, i, 0)), 0 );
        } else {
            assert(Cba_ObjFinNum(p, i) >= 2, "Cba_ObjFinNum(p, i) >= 2");
            Sign = Cba_NtkPrintDistribMakeSign( Cba_ObjSign(p, i), Cba_ObjSign(p, Cba_ObjFinFon(p, i, 0)), Cba_ObjSign(p, Cba_ObjFinFon(p, i, 1)) );
        }
        // add to storage
        Cba_NtkPrintDistribAddOne( vTypes, vOccurs, Type, Sign );
        // count the number of AIG nodes
        if Type == CBA_BOX_MUX {
            Vec_IntAddToEntry( vAnds, CBA_BOX_MUX,      3 * Cba_ObjRangeSize(p, i) * (Cba_ObjFinNum(p, i) - 2) );
        } else if Type == CBA_BOX_SHIR {      
            Vec_IntAddToEntry( vAnds, CBA_BOX_SHIR,  Abc_MinInt(Cba_ObjRangeSize(p, Cba_ObjFinFon(p, i, 0)), Abc_Base2Log(Cba_ObjRangeSize(p, i))) * 3 );
        } else if Type == CBA_BOX_SHIRA {     
            Vec_IntAddToEntry( vAnds, CBA_BOX_SHIRA, Cba_ObjRangeSize(p, i) * Abc_MinInt(Cba_ObjRangeSize(p, Cba_ObjFinFon(p, i, 0)), Abc_Base2Log(Cba_ObjRangeSize(p, i))) * 3 );
        } else if Type == CBA_BOX_SHIL {      
            Vec_IntAddToEntry( vAnds, CBA_BOX_SHIL,  Cba_ObjRangeSize(p, i) * Abc_MinInt(Cba_ObjRangeSize(p, Cba_ObjFinFon(p, i, 0)), Abc_Base2Log(Cba_ObjRangeSize(p, i))) * 3 );
        } else if Type == CBA_BOX_SHILA {     
            Vec_IntAddToEntry( vAnds, CBA_BOX_SHILA, Cba_ObjRangeSize(p, i) * Abc_MinInt(Cba_ObjRangeSize(p, Cba_ObjFinFon(p, i, 0)), Abc_Base2Log(Cba_ObjRangeSize(p, i))) * 3 );
        } else if Type == CBA_BOX_ROTR {     
            Vec_IntAddToEntry( vAnds, CBA_BOX_ROTR,  Cba_ObjRangeSize(p, i) * Abc_MinInt(Cba_ObjRangeSize(p, Cba_ObjFinFon(p, i, 0)), Abc_Base2Log(Cba_ObjRangeSize(p, i))) * 3 );
        } else if Type == CBA_BOX_ROTL {     
            Vec_IntAddToEntry( vAnds, CBA_BOX_ROTL,  Cba_ObjRangeSize(p, i) * Abc_MinInt(Cba_ObjRangeSize(p, Cba_ObjFinFon(p, i, 0)), Abc_Base2Log(Cba_ObjRangeSize(p, i))) * 3 );
        } else if Type == CBA_BOX_INV {     
            Vec_IntAddToEntry( vAnds, CBA_BOX_INV, 0 );
        } else if Type == CBA_BOX_AND {      
            Vec_IntAddToEntry( vAnds, CBA_BOX_AND,         Cba_ObjRangeSize(p, Cba_ObjFinFon(p, i, 0)) );
        } else if Type == CBA_BOX_OR {       
            Vec_IntAddToEntry( vAnds, CBA_BOX_OR,          Cba_ObjRangeSize(p, Cba_ObjFinFon(p, i, 0)) );
        } else if Type == CBA_BOX_XOR {      
            Vec_IntAddToEntry( vAnds, CBA_BOX_XOR,     3 * Cba_ObjRangeSize(p, Cba_ObjFinFon(p, i, 0)) );
        } else if Type == CBA_BOX_SLICE {   
            Vec_IntAddToEntry( vAnds, CBA_BOX_SLICE, 0 );
        } else if Type == CBA_BOX_CONCAT {   
            Vec_IntAddToEntry( vAnds, CBA_BOX_CONCAT, 0 );
        } else if Type == CBA_BOX_LNOT {    
            Vec_IntAddToEntry( vAnds, CBA_BOX_LNOT,        Cba_ObjRangeSize(p, Cba_ObjFinFon(p, i, 0)) - 1 );
        } else if Type == CBA_BOX_LAND {    
            Vec_IntAddToEntry( vAnds, CBA_BOX_LAND,        Cba_ObjRangeSize(p, Cba_ObjFinFon(p, i, 0)) + Cba_ObjRangeSize(p, Cba_ObjFinFon(p, i, 1)) - 1 );
        } else if Type == CBA_BOX_LOR {     
            Vec_IntAddToEntry( vAnds, CBA_BOX_LOR,         Cba_ObjRangeSize(p, Cba_ObjFinFon(p, i, 0)) + Cba_ObjRangeSize(p, Cba_ObjFinFon(p, i, 1)) - 1 );
        } else if Type == CBA_BOX_LXOR {    
            Vec_IntAddToEntry( vAnds, CBA_BOX_LXOR,        Cba_ObjRangeSize(p, Cba_ObjFinFon(p, i, 0)) + Cba_ObjRangeSize(p, Cba_ObjFinFon(p, i, 1)) + 1 );
        } else if Type == CBA_BOX_EQU {     
            Vec_IntAddToEntry( vAnds, CBA_BOX_EQU,      4 * Cba_ObjRangeSize(p, Cba_ObjFinFon(p, i, 0)) - 1 );
        } else if Type == CBA_BOX_NEQU {    
            Vec_IntAddToEntry( vAnds, CBA_BOX_NEQU,     4 * Cba_ObjRangeSize(p, Cba_ObjFinFon(p, i, 0)) - 1 );
        } else if Type == CBA_BOX_LTHAN {    
            Vec_IntAddToEntry( vAnds, CBA_BOX_LTHAN,    6 * Cba_ObjRangeSize(p, Cba_ObjFinFon(p, i, 0)) - 6 );
        } else if Type == CBA_BOX_MTHAN {    
            Vec_IntAddToEntry( vAnds, CBA_BOX_MTHAN,    6 * Cba_ObjRangeSize(p, Cba_ObjFinFon(p, i, 0)) - 6 );
        } else if Type == CBA_BOX_LETHAN { 
            Vec_IntAddToEntry( vAnds, CBA_BOX_LETHAN,   6 * Cba_ObjRangeSize(p, Cba_ObjFinFon(p, i, 0)) - 6 );
        } else if Type == CBA_BOX_METHAN { 
            Vec_IntAddToEntry( vAnds, CBA_BOX_METHAN,   6 * Cba_ObjRangeSize(p, Cba_ObjFinFon(p, i, 0)) - 6 );
        } else if Type == CBA_BOX_RAND {   
            Vec_IntAddToEntry( vAnds, CBA_BOX_RAND,         Cba_ObjRangeSize(p, Cba_ObjFinFon(p, i, 0)) - 1 );
        } else if Type == CBA_BOX_ROR {    
            Vec_IntAddToEntry( vAnds, CBA_BOX_ROR,          Cba_ObjRangeSize(p, Cba_ObjFinFon(p, i, 0)) - 1 );
        } else if Type == CBA_BOX_RXOR {   
            Vec_IntAddToEntry( vAnds, CBA_BOX_RXOR,     3 * Cba_ObjRangeSize(p, Cba_ObjFinFon(p, i, 0)) - 3 );
        } else if Type == CBA_BOX_ADD {       
            Vec_IntAddToEntry( vAnds, CBA_BOX_ADD,      9 * Cba_ObjRangeSize(p, Cba_ObjFinFon(p, i, 0)) );
        } else if Type == CBA_BOX_SUB {       
            Vec_IntAddToEntry( vAnds, CBA_BOX_SUB,      9 * Cba_ObjRangeSize(p, Cba_ObjFinFon(p, i, 0)) );
        } else if Type == CBA_BOX_MUL {    
            Vec_IntAddToEntry( vAnds, CBA_BOX_MUL,      9 * Cba_ObjRangeSize(p, Cba_ObjFinFon(p, i, 0)) * Cba_ObjRangeSize(p, Cba_ObjFinFon(p, i, 1)) );
        } else if Type == CBA_BOX_DIV {   
            Vec_IntAddToEntry( vAnds, CBA_BOX_DIV,     13 * Cba_ObjRangeSize(p, Cba_ObjFinFon(p, i, 0)) * Cba_ObjRangeSize(p, Cba_ObjFinFon(p, i, 0)) - 19 * Cba_ObjRangeSize(p, Cba_ObjFinFon(p, i, 0)) + 10 );
        } else if Type == CBA_BOX_MOD {  
            Vec_IntAddToEntry( vAnds, CBA_BOX_MOD,     13 * Cba_ObjRangeSize(p, Cba_ObjFinFon(p, i, 0)) * Cba_ObjRangeSize(p, Cba_ObjFinFon(p, i, 0)) - 7 * Cba_ObjRangeSize(p, Cba_ObjFinFon(p, i, 0)) - 2  );
        } else if Type == CBA_BOX_POW { 
            Vec_IntAddToEntry( vAnds, CBA_BOX_POW,     10 * int32(math.Pow(float64(float64(Cba_ObjRangeSize(p, Cba_ObjFinFon(p, i, 0)))), float64(float64(Cba_ObjRangeSize(p, Cba_ObjFinFon(p, i, 0)))))) );
        } else if Type == CBA_BOX_MIN {   
            Vec_IntAddToEntry( vAnds, CBA_BOX_MIN,      4 * Cba_ObjRangeSize(p, Cba_ObjFinFon(p, i, 0)) );
        } else if Type == CBA_BOX_SQRT {    
            Vec_IntAddToEntry( vAnds, CBA_BOX_SQRT,    11 * Cba_ObjRangeSize(p, Cba_ObjFinFon(p, i, 0)) * Cba_ObjRangeSize(p, Cba_ObjFinFon(p, i, 0)) / 8 + 5 * Cba_ObjRangeSize(p, Cba_ObjFinFon(p, i, 0)) / 2 - 5  ); }
    }
    }
    if nCountRange != 0 {
        fmt.Printf("Warning: %d objects of the design have non-zero-based ranges.\n", nCountRange);
        fmt.Printf("In particular, object %6d with name \"%s\" has range %d=[%d:%d]\n", iRnObj, Cba_FonNameStr(p, Cba_ObjFon0(p, iRnObj)), Cba_ObjRangeSize(p, iRnObj), Cba_ObjLeft(p, iRnObj), Cba_ObjRight(p, iRnObj));
    }
    // print by occurrence
    fmt.Printf("ID  :  name  occurrence    and2 (occurrence)<output_range>=<input_range>.<input_range> ...\n");
    for  i = 0; i < CBA_BOX_LAST; i++  {
        var vType *Vec_Wrd_t   = (Vec_Wrd_t *)Vec_PtrEntry( vTypes, i );
        var vOccur *Vec_Wrd_t  = (Vec_Wrd_t *)Vec_PtrEntry( vOccurs, i );
        if p.pDesign.nObjs[i] == 0 {
            continue; }
        fmt.Printf("%2d  :  %-8s  %6d%8d ", i, pTypeNames[i], p.pDesign.nObjs[i], Vec_IntEntry(vAnds, i));
        // sort by occurence
        Cba_NtkPrintDistribSortOne( vTypes, vOccurs, i );
        // inlining of macro Vec_WrdForEachEntry
        for  k = 0; (k < Vec_WrdSize(vType)) && (((Sign) = Vec_WrdEntry(vType, k)), 1); k++  {
            Cba_NtkPrintDistribFromSign( Sign, &s, &s0, &s1 );
            if ((k % 6) == 5 && s1) || ((k % 8) == 7 && !s1) {
                fmt.Printf("\n                                "); }
            fmt.Printf("(%d)", int32(Vec_WrdEntry( vOccur, k )));
            fmt.Printf("%s%d", Abc_LitIsCompl(s)?"-":"", Abc_Lit2Var(s));
            if s0 {
                fmt.Printf("=%s%d", Abc_LitIsCompl(s0)?"-":"", Abc_Lit2Var(s0)); }
            if s1 {
                fmt.Printf(".%s%d", Abc_LitIsCompl(s1)?"-":"", Abc_Lit2Var(s1)); }
            fmt.Printf(" ");
        }
        fmt.Printf("\n");
    }
    Vec_VecFree( (Vec_Vec_t *)vTypes );
    Vec_VecFree( (Vec_Vec_t *)vOccurs );
    Vec_IntFree( vAnds );
}
func Cba_NtkPrintNodes(p *Cba_Ntk_t, Type int32) {
    var i, iFon0, iFon1, Counter int32;
    var pTypeNames [CBA_BOX_LAST]string;
    Cba_ManCreatePrimMap( pTypeNames );
    fmt.Printf("Operation %s\n", pTypeNames[Type]);
    // inlining of macro Cba_NtkForEachObj
    for  i = 1; i < Vec_StrSize(&p.vObjType); i++  { if !Cba_ObjType(p, i) {} else {
        if int32(Type) != Type {
            continue; }
        iFon0 = Cba_ObjFinFon(p, i, 0);
        iFon1 = Cba_ObjFinFon(p, i, 1);

        fmt.Printf("%8d  :", Counter++);
        fmt.Printf("%8d  :  ", i);
        fmt.Printf("%3d%s = ", Cba_ObjRangeSize(p, i), Cba_ObjSigned(p, i) ? "s" : " ");
        fmt.Printf("%3d%s  %s ", Cba_ObjRangeSize(p, Cba_ObjFinFon(p, i, 0)), Cba_ObjSigned(p, iFon0) ? "s" : " ", pTypeNames[Type]);
        fmt.Printf("%3d%s ", Cba_ObjRangeSize(p, Cba_ObjFinFon(p, i, 1)), Cba_ObjSigned(p, iFon1) ? "s" : " ");
        fmt.Printf(" :    ");
        fmt.Printf("%-12s =  ", Cba_ObjNameStr(p, i));
        fmt.Printf("%-12s  %s  ", Cba_FonIsConst(iFon0) ? Cba_NtkConst(p, Cba_FonConst(iFon0)) : Cba_FonNameStr(p, iFon0), pTypeNames[Type]);
        fmt.Printf("%-12s ", Cba_FonIsConst(iFon1) ? Cba_NtkConst(p, Cba_FonConst(iFon1)) : Cba_FonNameStr(p, iFon1));
        fmt.Printf("\n");
    }
    }
}
func Cba_NtkPrintStatsFull(p *Cba_Ntk_t, fDistrib int32, fVerbose int32) {
    var i int32;
    var pTypeNames [CBA_BOX_LAST]string;
    Cba_ManCreatePrimMap( pTypeNames );
    fmt.Printf("%-20s : ", Cba_NtkName(p));
    fmt.Printf("PI = %4d  ", Cba_NtkPiNum(p));
    fmt.Printf("PO = %4d  ", Cba_NtkPoNum(p));
    fmt.Printf("FF = %4d  ", Cba_NtkBoxSeqNum(p));
    fmt.Printf("Obj = %6d  ", Cba_NtkObjNum(p));
    fmt.Printf("Mem = %.3f MB", 1.0*Cba_NtkMemory(p)/(1<<20));
    fmt.Printf("\n");
    if fDistrib != 0 {
        Cba_NtkPrintDistrib( p, fVerbose );
        return;
    }
    if !(fVerbose != 0) {
        return; }
    fmt.Printf("Node type statistics:\n");
    for  i = 1; i < CBA_BOX_LAST; i++  {
        if !p.pDesign.nObjs[i] {
            continue; }
        if p.pDesign.nAnds[0] && p.pDesign.nAnds[i] {
            fmt.Printf("%2d  :  %-8s  %6d  %7.2f %%\n", i, pTypeNames[i], p.pDesign.nObjs[i], 100.0*p.pDesign.nAnds[i]/p.pDesign.nAnds[0]);
        } else {
            fmt.Printf("%2d  :  %-8s  %6d\n", i, pTypeNames[i], p.pDesign.nObjs[i]); }
    }
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cba_NtkCollectDistrib(p *Cba_Ntk_t, pCounts *int32, pUserCounts *int32) {
    var i int32;
    // inlining of macro Cba_NtkForEachBox
    for  i = 1; i < Vec_StrSize(&p.vObjType); i++  { if !Cba_ObjIsBox(p, i) {} else
        if Cba_ObjIsBoxUser(p, i) {
            pUserCounts[Cba_ObjNtkId(p, i)]++;
        } else {
            pCounts[Cba_ObjType(p, i)]++; }
    }
}

func Cba_NtkPrintDistribStat(p *Cba_Ntk_t, pCounts *int32, pUserCounts *int32) {
    var pNtk *Cba_Ntk_t; var i int32;
    fmt.Printf("Primitives (%d):\n", Cba_NtkBoxPrimNum(p));
    for  i = 0; i < CBA_BOX_LAST; i++  {
        if pCounts[i] {
            fmt.Printf("%-20s = %5d\n", Cba_NtkTypeName(p, i), pCounts[i]); }
    }
    fmt.Printf("User hierarchy (%d):\n", Cba_NtkBoxUserNum(p));
    // inlining of macro Cba_ManForEachNtk
    for  i = 1; (i <= Cba_ManNtkNum(p.pDesign)) && (((pNtk) = Cba_ManNtk(p.pDesign, i)), 1); i++  { 
        if pUserCounts[i] {
            fmt.Printf("%-20s = %5d\n", Cba_NtkName(pNtk), pUserCounts[i]); }
    }
}
func Cba_NtkPrintDistribOld(p *Cba_Ntk_t) {
    var pCounts [CBA_BOX_LAST]int32 ;
    // inlining of macro ABC_CALLOC
    var pUserCounts *int32  = ((int *) calloc((size_t)(Cba_ManNtkNum(p.pDesign) + 1), sizeof(int)));
    Cba_ManCreatePrimMap( p.pDesign.pTypeNames );
    Cba_NtkCollectDistrib( p, pCounts, pUserCounts );
    Cba_NtkPrintDistribStat( p, pCounts, pUserCounts );
    // inlining of macro ABC_FREE
    ((pUserCounts) ? (free((char *) (pUserCounts)), (pUserCounts) = 0) : 0);
}

func Cba_ManPrintDistribStat(p *Cba_Man_t, pCounts *int32, pUserCounts *int32) {
    var pNtk *Cba_Ntk_t  = Cba_ManRoot(p); var i int32;
    fmt.Printf("Primitives:\n");
    for  i = 0; i < CBA_BOX_LAST; i++  {
        if pCounts[i] {
            fmt.Printf("%-20s = %5d\n", Cba_NtkTypeName(pNtk, i), pCounts[i]); }
    }
    fmt.Printf("User hierarchy:\n");
    // inlining of macro Cba_ManForEachNtk
    for  i = 1; (i <= Cba_ManNtkNum(p)) && (((pNtk) = Cba_ManNtk(p, i)), 1); i++  { 
        if pUserCounts[i] {
            fmt.Printf("%-20s = %5d\n", Cba_NtkName(pNtk), pUserCounts[i]); }
    }
}
func Cba_ManPrintDistrib(p *Cba_Man_t) {
    var pNtk *Cba_Ntk_t; var i int32;
    var pCounts [CBA_BOX_LAST]int32 ;
    // inlining of macro ABC_CALLOC
    var pUserCounts *int32  = ((int *) calloc((size_t)(Cba_ManNtkNum(p) + 1), sizeof(int)));
    Cba_ManCreatePrimMap( p.pTypeNames );
    // inlining of macro Cba_ManForEachNtk
    for  i = 1; (i <= Cba_ManNtkNum(p)) && (((pNtk) = Cba_ManNtk(p, i)), 1); i++  { 
        Cba_NtkCollectDistrib( pNtk, pCounts, pUserCounts );
    }
    Cba_ManPrintDistribStat( p, pCounts, pUserCounts );
    // inlining of macro ABC_FREE
    ((pUserCounts) ? (free((char *) (pUserCounts)), (pUserCounts) = 0) : 0);
}

/**Function*************************************************************

  Synopsis    [Order objects by box type and then by name.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
// compares two numbers with the first mismatching char in i-th position
func Cba_CharIsDigit(c byte) int32 { return c >= '0' && c <= '9'; }
func Cba_StrCmpInt(p1 string, p2 string, i int32) int32 {
    // check if one of the mismatching chars is a digit
    if (Cba_CharIsDigit(p1[i]) != 0) || (Cba_CharIsDigit(p2[i]) != 0) {
        // if previous (equal) char was a digit or if this is first digit on both sides, scroll back
        if (i > 0 && (Cba_CharIsDigit(p1[i-1]) != 0)) || ((Cba_CharIsDigit(p1[i]) != 0) && (Cba_CharIsDigit(p2[i]) != 0)) {
            var Num1, Num2 int32;
            // find the first digit
            for  --i; i >= 0; i--  {
                if !(Cba_CharIsDigit(p1[i]) != 0) {
                    break; }
            }
            i++;
            // current char is digit
            assert((Cba_CharIsDigit(p1[i]) != 0), "Cba_CharIsDigit(p1[i])");
            assert((Cba_CharIsDigit(p2[i]) != 0), "Cba_CharIsDigit(p2[i])");
            // previous char does not exist or is not a digit
            assert(i == 0 || !(Cba_CharIsDigit(p1[i-1]) != 0), "i == 0 || !Cba_CharIsDigit(p1[i-1])");
            assert(i == 0 || !(Cba_CharIsDigit(p2[i-1]) != 0), "i == 0 || !Cba_CharIsDigit(p2[i-1])");
            // compare numbers
            Num1 = strconv.Atoi(p1 + i);
            Num2 = strconv.Atoi(p2 + i);
            if Num1 < Num2 {
                return -1; }
            if Num1 > Num2 { 
                return 1; }
            assert(false, "0");
            return 0;
        }
    }
    // compare as usual
    if p1[i] < p2[i] {
        return -1; }
    if p1[i] > p2[i] {
        return 1; }
    assert(false, "0");
    return 0;
}
func Cba_StrCmp(pp1 *string, pp2 *string) int32 {
    var p1 string  = *pp1;
    var p2 string  = *pp2; var i int32;
    for  i = 0; p1[i] && p2[i]; i++  {
        if p1[i] != p2[i] {
            return Cba_StrCmpInt( p1, p2, i ); }
    }
    assert(!p1[i] || !p2[i], "!p1[i] || !p2[i]");
    return Cba_StrCmpInt( p1, p2, i );
}
func Cba_NtkObjOrder(p *Cba_Ntk_t, vObjs *Vec_Int_t, vNameIds *Vec_Int_t) {
    var Buffer string // was [1000]byte in C

    var pName string;
    var vNames *Vec_Ptr_t;
    var i, iObj int32;
    if Vec_IntSize(vObjs) < 2 {
        return; }
    vNames = Vec_PtrAlloc( Vec_IntSize(vObjs) );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vObjs)) && (((iObj) = Vec_IntEntry(vObjs, i)), 1); i++  {
        var pTypeName string  = Cba_NtkTypeName( p, Cba_ObjType(p, iObj) );
        var pInstName string;
        if vNameIds != nil {
            pInstName = Cba_NtkStr(p, Vec_IntEntry(vNameIds, i));
        } else {
            pInstName = Cba_ObjNameStr(p, iObj);
        }
        Buffer = fmt.Sprintf("%s_%s_%d", pTypeName, pInstName, iObj);
        Vec_PtrPush( vNames, Abc_UtilStrsav(Buffer) );
    }
    // print before
//    Vec_PtrForEachEntry( char *, vNames, pName, i )
//        printf( "%s \n", pName );
//    printf( "\n" );
    // do the sorting
    Vec_PtrSort( vNames, (int (*)(void))Cba_StrCmp );
    // print after
//    Vec_PtrForEachEntry( char *, vNames, pName, i )
//        printf( "%s \n", pName );
//    printf( "\n" );
    // reload in a new order
    Vec_IntClear( vObjs );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNames)) && (((pName) = (char *)Vec_PtrEntry(vNames, i)), 1); i++  {
        Vec_IntPush( vObjs, strconv.Atoi(strrchr(pName, '_')+1) );
    }
    Vec_PtrFreeFree( vNames );
}


/**Function*************************************************************

  Synopsis    [Returns the number of CI fons.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cba_NtkCiFonNum(p *Cba_Ntk_t) int32 {
    var i int32;
    var iObj int32;
    var Count int32 = Cba_NtkPiNum(p);
    // inlining of macro Cba_NtkForEachBoxSeq
    for  i = 0; (i < Cba_NtkBoxSeqNum(p))  && (((iObj) = Cba_NtkBoxSeq(p, i)), 1); i++  { 
        Count += Cba_ObjFonNum(p, iObj);
    }
    return Count;  
}
func Cba_NtkCoFinNum(p *Cba_Ntk_t) int32 {
    var i int32;
    var iObj int32;
    var Count int32 = Cba_NtkPoNum(p);
    // inlining of macro Cba_NtkForEachBoxSeq
    for  i = 0; (i < Cba_NtkBoxSeqNum(p))  && (((iObj) = Cba_NtkBoxSeq(p, i)), 1); i++  { 
        Count += Cba_ObjFinNum(p, iObj);
    }
    return Count;  
}

/**Function*************************************************************

  Synopsis    [Returns 1 if the manager is in the topo order.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cba_NtkIsTopoOrder(p *Cba_Ntk_t) int32 {
    var i int32;
    var k int32;
    var iObj int32;
    var iFin int32;
    var iFanin int32;
    var fTopo int32 = 1;
    var vVisited *Vec_Bit_t  = Vec_BitStart( Cba_NtkObjNum(p) + 1 );
    // mark PIs and seq objects
    // inlining of macro Cba_NtkForEachPi
    for  i = 0; (i < Cba_NtkPiNum(p))  && (((iObj) = Cba_NtkPi(p, i)), 1); i++  { 
        Vec_BitWriteEntry( vVisited, iObj, 1 );
    }
    // inlining of macro Cba_NtkForEachBoxSeq
    for  i = 0; (i < Cba_NtkBoxSeqNum(p))  && (((iObj) = Cba_NtkBoxSeq(p, i)), 1); i++  { 
        Vec_BitWriteEntry( vVisited, iObj, 1 );
    }
    // visit combinational nodes
    // inlining of macro Cba_NtkForEachBox
    for  iObj = 1; iObj < Vec_StrSize(&p.vObjType); iObj++  { if !Cba_ObjIsBox(p, iObj) {} else
        if !Cba_ObjIsSeq(p, iObj) {
            // inlining of macro Cba_ObjForEachFinFaninReal
            for  k = 0, iFin = Cba_ObjFin0(p, iObj); iFin < Cba_ObjFin0(p, iObj+1) && ((iFanin = Cba_FonObj(p, Cba_FinFon(p, iFin))), 1); iFin, k = iFin+1, k+1  { if !(iFanin != 0) {} else
                if !Vec_BitEntry(vVisited, iFanin) {
                    fTopo = 0; }
            }
            if !(fTopo != 0) {
                break; }
            Vec_BitWriteEntry( vVisited, iObj, 1 );
        }
    }
    // visit POs and seq objects
    if fTopo != 0 {
    // inlining of macro Cba_NtkForEachPo
    for  i = 0; (i < Cba_NtkPoNum(p))  && (((iObj) = Cba_NtkPo(p, i)), 1); i++  {
        // inlining of macro Cba_ObjForEachFinFaninReal
        for  k = 0, iFin = Cba_ObjFin0(p, iObj); iFin < Cba_ObjFin0(p, iObj+1) && ((iFanin = Cba_FonObj(p, Cba_FinFon(p, iFin))), 1); iFin, k = iFin+1, k+1  { if !(iFanin != 0) {} else
            if !Vec_BitEntry(vVisited, iFanin) {
                fTopo = 0; }
        }
        if !(fTopo != 0) {
            break; }
    } }
    if fTopo != 0 {
    // inlining of macro Cba_NtkForEachBoxSeq
    for  i = 0; (i < Cba_NtkBoxSeqNum(p))  && (((iObj) = Cba_NtkBoxSeq(p, i)), 1); i++  {
        // inlining of macro Cba_ObjForEachFinFaninReal
        for  k = 0, iFin = Cba_ObjFin0(p, iObj); iFin < Cba_ObjFin0(p, iObj+1) && ((iFanin = Cba_FonObj(p, Cba_FinFon(p, iFin))), 1); iFin, k = iFin+1, k+1  { if !(iFanin != 0) {} else
            if !Vec_BitEntry(vVisited, iFanin) {
                fTopo = 0; }
        }
        if !(fTopo != 0) {
            break; }
    } }
    Vec_BitFree( vVisited );
    return fTopo;
}
func Cba_ManIsTopoOrder(p *Cba_Man_t) int32 {
    var pNtk *Cba_Ntk_t; var i int32;
    // inlining of macro Cba_ManForEachNtk
    for  i = 1; (i <= Cba_ManNtkNum(p)) && (((pNtk) = Cba_ManNtk(p, i)), 1); i++  { 
        if !(Cba_NtkIsTopoOrder(pNtk) != 0) {
            return 0; }
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Collects user boxes in the DFS order.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cba_NtkCheckComboLoop_rec(p *Cba_Ntk_t, iObj int32) int32 {
    var k, iFin, iFanin int32;
    if Cba_ObjIsPi(p, iObj) {
        return 1; }
    if Cba_ObjCopy(p, iObj) == 1 { // visited
        return 1; }
    if Cba_ObjCopy(p, iObj) == 0 { // loop
        return 0; }
    Cba_ObjSetCopy( p, iObj, 0 );
    // inlining of macro Cba_ObjForEachFinFaninReal
    for  k = 0, iFin = Cba_ObjFin0(p, iObj); iFin < Cba_ObjFin0(p, iObj+1) && ((iFanin = Cba_FonObj(p, Cba_FinFon(p, iFin))), 1); iFin, k = iFin+1, k+1  { if !(iFanin != 0) {} else
        if !Cba_ObjIsSeq(p, iFanin) {
            if !(Cba_NtkCheckComboLoop_rec( p, iFanin ) != 0) {
                return 0; } }
    }
    //Cba_ObjSetCopy( p, iObj, 1 );
    Vec_IntSetEntry( &p.vObjCopy, iObj, 1 );
    return 1;
}
func Cba_NtkCheckComboLoop(p *Cba_Ntk_t) int32 {
    var iObj int32;
    Cba_NtkCleanObjCopies( p ); // -1 = not visited; 0 = on the path; 1 = finished
    // inlining of macro Cba_NtkForEachBox
    for  iObj = 1; iObj < Vec_StrSize(&p.vObjType); iObj++  { if !Cba_ObjIsBox(p, iObj) {} else
        if !Cba_ObjIsSeq(p, iObj) {
            if !(Cba_NtkCheckComboLoop_rec( p, iObj ) != 0) {
                fmt.Printf("Cyclic dependency of user boxes is detected.\n");
                return 0;
            } }
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Find one missing object.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cba_NtkFindMissing(vObjs *Vec_Int_t, nObjs int32) int32 {
    var vMap *Vec_Int_t  = Vec_IntStartFull( nObjs + 1 );
    var i, iObj int32;
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vObjs)) && (((iObj) = Vec_IntEntry(vObjs, i)), 1); i++  {
        Vec_IntWriteEntry( vMap, iObj, i );
    }
    // inlining of macro Vec_IntForEachEntryStart
    for  iObj = 1; (iObj < Vec_IntSize(vMap)) && (((i) = Vec_IntEntry(vMap, iObj)), 1); iObj++  {
        if i == -1 {
            Vec_IntFree( vMap );
            return iObj;
        }
    }
    Vec_IntFree( vMap );
    return -1;
}

/**Function*************************************************************

  Synopsis    [Collect nodes in the DFS order.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cba_NtkCollectDfs_rec(p *Cba_Ntk_t, iObj int32, vObjs *Vec_Int_t) {
    var iFin, iFanin, k int32;
    if !Cba_ObjCopy(p, iObj) {
        return; }
    Cba_ObjSetCopy( p, iObj, 0 );
    // inlining of macro Cba_ObjForEachFinFaninReal
    for  k = 0, iFin = Cba_ObjFin0(p, iObj); iFin < Cba_ObjFin0(p, iObj+1) && ((iFanin = Cba_FonObj(p, Cba_FinFon(p, iFin))), 1); iFin, k = iFin+1, k+1  { if !(iFanin != 0) {} else {
        Cba_NtkCollectDfs_rec( p, iFanin, vObjs ); }
    }
    Vec_IntPush( vObjs, iObj );
}
func Cba_NtkCollectDfs(p *Cba_Ntk_t) *Vec_Int_t {
    var i, k, iObj, iFin, iFanin int32;
    var vObjs *Vec_Int_t  = Vec_IntAlloc( Cba_NtkObjNum(p) );
    // collect PIs and seq boxes
    // inlining of macro Cba_NtkForEachPi
    for  i = 0; (i < Cba_NtkPiNum(p))  && (((iObj) = Cba_NtkPi(p, i)), 1); i++  { 
        Vec_IntPush( vObjs, iObj );
    }
    // inlining of macro Cba_NtkForEachBoxSeq
    for  i = 0; (i < Cba_NtkBoxSeqNum(p))  && (((iObj) = Cba_NtkBoxSeq(p, i)), 1); i++  { 
        Vec_IntPush( vObjs, iObj );
    }
    // prepare leaves
    Cba_NtkCleanObjCopies( p );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vObjs)) && (((iObj) = Vec_IntEntry(vObjs, i)), 1); i++  {
        Cba_ObjSetCopy( p, iObj, 0 );
    }
    // collect internal
    // inlining of macro Cba_NtkForEachPo
    for  i = 0; (i < Cba_NtkPoNum(p))  && (((iObj) = Cba_NtkPo(p, i)), 1); i++  { 
        // inlining of macro Cba_ObjForEachFinFaninReal
        for  k = 0, iFin = Cba_ObjFin0(p, iObj); iFin < Cba_ObjFin0(p, iObj+1) && ((iFanin = Cba_FonObj(p, Cba_FinFon(p, iFin))), 1); iFin, k = iFin+1, k+1  { if !(iFanin != 0) {} else {
            Cba_NtkCollectDfs_rec( p, iFanin, vObjs ); }
        }
    }
    // inlining of macro Cba_NtkForEachBoxSeq
    for  i = 0; (i < Cba_NtkBoxSeqNum(p))  && (((iObj) = Cba_NtkBoxSeq(p, i)), 1); i++  { 
        // inlining of macro Cba_ObjForEachFinFaninReal
        for  k = 0, iFin = Cba_ObjFin0(p, iObj); iFin < Cba_ObjFin0(p, iObj+1) && ((iFanin = Cba_FonObj(p, Cba_FinFon(p, iFin))), 1); iFin, k = iFin+1, k+1  { if !(iFanin != 0) {} else {
            Cba_NtkCollectDfs_rec( p, iFanin, vObjs ); }
        }
    }
    // collect POs
    // inlining of macro Cba_NtkForEachPo
    for  i = 0; (i < Cba_NtkPoNum(p))  && (((iObj) = Cba_NtkPo(p, i)), 1); i++  { 
        Vec_IntPush( vObjs, iObj );
    }
    assert(Vec_IntSize(vObjs) <= Cba_NtkObjNum(p), "Vec_IntSize(vObjs) <= Cba_NtkObjNum(p)");
    if Vec_IntSize(vObjs) != Cba_NtkObjNum(p) && !Abc_FrameReadFlag("silentmode") {
        var iObj int32  = Cba_NtkFindMissing( vObjs, Cba_NtkObjNum(p) );
        fmt.Printf("Warning: DSF ordering for module \"%s\" collected %d out of %d objects.\n", Cba_NtkName(p), Vec_IntSize(vObjs), Cba_NtkObjNum(p));
        fmt.Printf("         For example, object %d with name \"%s\" is not reachable from outputs.\n", iObj, iObj > 0 ? Cba_FonNameStr( p, Cba_ObjFon0(p, iObj) ) : "<unknown>");
    }
    return vObjs;
}


/**Function*************************************************************

  Synopsis    [Count number of objects after collapsing.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cba_ManGetClpStats_rec(p *Cba_Ntk_t, pCountN *int32, pCountI *int32, pCountO *int32) {
    var iObj int32;
    var Id int32 = Cba_NtkId(p);
    if pCountN[Id] >= 0 {
        return; }
    pCountN[Id] = pCountI[Id] = pCountO[Id] = 0;
    // inlining of macro Cba_NtkForEachObj
    for  iObj = 1; iObj < Vec_StrSize(&p.vObjType); iObj++  { if !Cba_ObjType(p, iObj) {} else   
        if Cba_ObjIsBoxPrim(p, iObj) {
            pCountN[Id] += 1;
            pCountI[Id] += Cba_ObjFinNum(p, iObj);
            pCountO[Id] += Cba_ObjFonNum(p, iObj);
        
        } else if Cba_ObjIsBoxUser(p, iObj) {
            var NtkId int32  = Cba_ObjNtkId(p, iObj);
            Cba_ManGetClpStats_rec( Cba_ObjNtk(p, iObj), pCountN, pCountI, pCountO );
            pCountN[Id] += pCountN[NtkId] + Cba_ObjFonNum(p, iObj);
            pCountI[Id] += pCountI[NtkId] + Cba_ObjFonNum(p, iObj);
            pCountO[Id] += pCountO[NtkId] + Cba_ObjFonNum(p, iObj);
        }
    }
}
func Cba_ManGetClpStats(p *Cba_Man_t, nObjs *int32, nFins *int32, nFons *int32) {
    // inlining of macro ABC_FALLOC
    var pCountN *int32  = ((int *) memset(malloc(sizeof(int) * (size_t)(Cba_ManNtkNum(p) + 1)), 0xff, sizeof(int) * (size_t)(Cba_ManNtkNum(p) + 1)));
    // inlining of macro ABC_FALLOC
    var pCountI *int32  = ((int *) memset(malloc(sizeof(int) * (size_t)(Cba_ManNtkNum(p) + 1)), 0xff, sizeof(int) * (size_t)(Cba_ManNtkNum(p) + 1)));
    // inlining of macro ABC_FALLOC
    var pCountO *int32  = ((int *) memset(malloc(sizeof(int) * (size_t)(Cba_ManNtkNum(p) + 1)), 0xff, sizeof(int) * (size_t)(Cba_ManNtkNum(p) + 1)));
    var pRoot *Cba_Ntk_t  = Cba_ManRoot(p);
    Cba_ManGetClpStats_rec( pRoot, pCountN, pCountI, pCountO );
    *nObjs = Cba_NtkPioNum(pRoot) + pCountN[Cba_NtkId(pRoot)];
    *nFins = Cba_NtkPoNum(pRoot)  + pCountI[Cba_NtkId(pRoot)];
    *nFons = Cba_NtkPiNum(pRoot)  + pCountO[Cba_NtkId(pRoot)];
    // inlining of macro ABC_FREE
    ((pCountN) ? (free((char *) (pCountN)), (pCountN) = 0) : 0);
    // inlining of macro ABC_FREE
    ((pCountI) ? (free((char *) (pCountI)), (pCountI) = 0) : 0);
    // inlining of macro ABC_FREE
    ((pCountO) ? (free((char *) (pCountO)), (pCountO) = 0) : 0);
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cba_NtkCollapse_rec(pNew *Cba_Ntk_t, p *Cba_Ntk_t, vSigs *Vec_Int_t) {
    var i, iObj, iObjNew, iFin, iFon int32;
    Cba_NtkCleanObjCopies( p );
    Cba_NtkCleanFonCopies( p );
    // set PI copies
    assert(Vec_IntSize(vSigs) == Cba_NtkPiNum(p), "Vec_IntSize(vSigs) == Cba_NtkPiNum(p)");
    // inlining of macro Cba_NtkForEachPiFon
    for  i = 0; (i < Cba_NtkPiNum(p))  && (((iObj) = Cba_NtkPi(p, i)), 1)  && (((iFon) = Cba_ObjFon0(p, iObj)), 1); i++  { 
        Cba_FonSetCopy( p, iFon, Vec_IntEntry(vSigs, i) );
    }
    // duplicate primitives and create buffers for user instances
    // inlining of macro Cba_NtkForEachObj
    for  iObj = 1; iObj < Vec_StrSize(&p.vObjType); iObj++  { if !Cba_ObjType(p, iObj) {} else   
        if Cba_ObjIsBoxPrim( p, iObj ) {
            iObjNew = Cba_ObjDup( pNew, p, iObj );
            // inlining of macro Cba_ObjForEachFon
            for  i = 0, iFon = Cba_ObjFon0(p, iObj); iFon < Cba_ObjFon0(p, iObj+1); iFon, i = iFon+1, i+1  {
                Cba_FonSetCopy( p, iFon, Cba_ObjFon(pNew, iObjNew, i) );
            }
            if Cba_ObjAttr(p, iObj) {
                Cba_ObjSetAttrs( pNew, iObjNew, Cba_ObjAttrArray(p, iObj), Cba_ObjAttrSize(p, iObj) ); }
        
        } else if Cba_ObjIsBoxUser( p, iObj ) {
            // inlining of macro Cba_ObjForEachFon
            for  i = 0, iFon = Cba_ObjFon0(p, iObj); iFon < Cba_ObjFon0(p, iObj+1); iFon, i = iFon+1, i+1  {
                iObjNew = Cba_ObjAlloc( pNew, CBA_BOX_BUF, 1, 1 );
                Cba_FonSetCopy( p, iFon, Cba_ObjFon0(pNew, iObjNew) );
            }
        }
    }
    // connect primitives and collapse user instances
    // inlining of macro Cba_NtkForEachObj
    for  iObj = 1; iObj < Vec_StrSize(&p.vObjType); iObj++  { if !Cba_ObjType(p, iObj) {} else   
        if Cba_ObjIsBoxPrim( p, iObj ) {
            iObjNew = Cba_ObjCopy( p, iObj );
            // inlining of macro Cba_ObjForEachFinFon
            for  i = 0, iFin = Cba_ObjFin0(p, iObj); iFin < Cba_ObjFin0(p, iObj+1) && ((iFon = Cba_FinFon(p, iFin)), 1); iFin, i = iFin+1, i+1  {
                Cba_ObjSetFinFon( pNew, iObjNew, i, Cba_FonCopy(p, iFon) );
            }
        
        } else if Cba_ObjIsBoxUser( p, iObj ) {
            Vec_IntClear( vSigs );
            // inlining of macro Cba_ObjForEachFinFon
            for  i = 0, iFin = Cba_ObjFin0(p, iObj); iFin < Cba_ObjFin0(p, iObj+1) && ((iFon = Cba_FinFon(p, iFin)), 1); iFin, i = iFin+1, i+1  {
                Vec_IntPush( vSigs, Cba_FonCopy(p, iFon) );
            }
            assert(Vec_IntSize(vSigs) == Cba_ObjFinNum(p, iObj), "Vec_IntSize(vSigs) == Cba_ObjFinNum(p, iObj)");
            Cba_NtkCollapse_rec( pNew, Cba_ObjNtk(p, iObj), vSigs );
            assert(Vec_IntSize(vSigs) == Cba_ObjFonNum(p, iObj), "Vec_IntSize(vSigs) == Cba_ObjFonNum(p, iObj)");
            // inlining of macro Cba_ObjForEachFon
            for  i = 0, iFon = Cba_ObjFon0(p, iObj); iFon < Cba_ObjFon0(p, iObj+1); iFon, i = iFon+1, i+1  {
                iObjNew = Cba_FonObj( pNew, Cba_FonCopy(p, iFon) ); // buffer
                Cba_ObjSetFinFon( pNew, iObjNew, 0, Vec_IntEntry(vSigs, i) );
            }
        }
    }
    // collect POs
    Vec_IntClear( vSigs );
    // inlining of macro Cba_NtkForEachPoDriverFon
    for  i = 0; (i < Cba_NtkPoNum(p))  && (((iObj) = Cba_NtkPo(p, i)), 1)  && (((iFon) = Cba_ObjFinFon(p, iObj, 0)), 1); i++  { 
        Vec_IntPush( vSigs, Cba_FonCopy(p, iFon) );
    }
}
func Cba_ManCollapse(p *Cba_Man_t) *Cba_Man_t {
    var pNew *Cba_Man_t   = Cba_ManAlloc( p.pSpec, 1, Abc_NamRef(p.pStrs), Abc_NamRef(p.pFuns), Abc_NamStart(100, 24), Hash_IntManRef(p.vHash) );
    var pRoot *Cba_Ntk_t = Cba_ManRoot( p );
    var pRootNew *Cba_Ntk_t;
    var vSigs *Vec_Int_t  = Vec_IntAlloc( 1000 );
    var i, iObj, iObjNew, iFon, nObjs, nFins, nFons int32;
    Cba_ManDupTypeNames( pNew, p );
    Cba_ManGetClpStats( p, &nObjs, &nFins, &nFons    );
    pRootNew = Cba_NtkAlloc( pNew, Cba_NtkNameId(pRoot), Cba_NtkPiNum(pRoot), Cba_NtkPoNum(pRoot), nObjs, nFins, nFons );
    Cba_NtkAdd( pNew, pRootNew );
    if Cba_NtkHasObjNames(pRoot) {
        Cba_NtkCleanObjNames( pRootNew ); }
    if Cba_NtkHasFonNames(pRoot) {
        Cba_NtkCleanFonNames( pRootNew ); }
    if Cba_NtkHasObjAttrs(pRoot) {
        Cba_NtkCleanObjAttrs( pRootNew ); }
    if Cba_ObjAttr(pRoot, 0) {
        Cba_ObjSetAttrs( pRootNew, 0, Cba_ObjAttrArray(pRoot, 0), Cba_ObjAttrSize(pRoot, 0) ); }
    Cba_NtkCleanObjCopies( pRoot );
    // inlining of macro Cba_NtkForEachPiFon
    for  i = 0; (i < Cba_NtkPiNum(pRoot))  && (((iObj) = Cba_NtkPi(pRoot, i)), 1)  && (((iFon) = Cba_ObjFon0(pRoot, iObj)), 1); i++  {
        iObjNew = Cba_ObjDup( pRootNew, pRoot, iObj );
        Vec_IntPush( vSigs, Cba_ObjFon0(pRootNew, iObjNew) );
        if Cba_NtkHasObjNames(pRoot) {
            Cba_ObjSetName( pRootNew, iObjNew, Cba_ObjName(pRoot, iObj) ); }
        if Cba_NtkHasFonNames(pRoot) {
            Cba_FonSetName( pRootNew, Cba_ObjFon0(pRootNew, iObjNew), Cba_FonName(pRoot, iFon) ); }
        if Cba_ObjAttr(pRoot, iObj) {
            Cba_ObjSetAttrs( pRootNew, iObjNew, Cba_ObjAttrArray(pRoot, iObj), Cba_ObjAttrSize(pRoot, iObj) ); }
    }
    assert(Vec_IntSize(vSigs) == Cba_NtkPiNum(pRoot), "Vec_IntSize(vSigs) == Cba_NtkPiNum(pRoot)");
    Cba_NtkCollapse_rec( pRootNew, pRoot, vSigs );
    assert(Vec_IntSize(vSigs) == Cba_NtkPoNum(pRoot), "Vec_IntSize(vSigs) == Cba_NtkPoNum(pRoot)");
    // inlining of macro Cba_NtkForEachPoDriverFon
    for  i = 0; (i < Cba_NtkPoNum(pRoot))  && (((iObj) = Cba_NtkPo(pRoot, i)), 1)  && (((iFon) = Cba_ObjFinFon(pRoot, iObj, 0)), 1); i++  {
        iObjNew = Cba_ObjDup( pRootNew, pRoot, iObj );
        Cba_ObjSetFinFon( pRootNew, iObjNew, 0, Vec_IntEntry(vSigs, i) );
        if Cba_NtkHasObjNames(pRoot) {
            Cba_ObjSetName( pRootNew, iObjNew, Cba_ObjName(pRoot, iObj) ); }
        if Cba_NtkHasFonNames(pRoot) {
            Cba_FonSetName( pRootNew, Vec_IntEntry(vSigs, i), Cba_FonName(pRoot, iFon) ); }
        if Cba_ObjAttr(pRoot, iObj) {
            Cba_ObjSetAttrs( pRootNew, iObjNew, Cba_ObjAttrArray(pRoot, iObj), Cba_ObjAttrSize(pRoot, iObj) ); }
    }
    Vec_IntFree( vSigs );
    assert(Cba_NtkObjNum(pRootNew) == Cba_NtkObjNumAlloc(pRootNew), "Cba_NtkObjNum(pRootNew) == Cba_NtkObjNumAlloc(pRootNew)");
    assert(Cba_NtkFinNum(pRootNew) == Cba_NtkFinNumAlloc(pRootNew), "Cba_NtkFinNum(pRootNew) == Cba_NtkFinNumAlloc(pRootNew)");
    assert(Cba_NtkFonNum(pRootNew) == Cba_NtkFonNumAlloc(pRootNew), "Cba_NtkFonNum(pRootNew) == Cba_NtkFonNumAlloc(pRootNew)");
    // create internal node names
    Cba_NtkMissingFonNames( pRootNew, "m" );
    //Cba_NtkPrepareSeq( pRootNew );
    return pNew;
}



/**Function*************************************************************

  Synopsis    [Performs the reverse of collapsing.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cba_NtkCollectInFons(p *Cba_Ntk_t, vObjs *Vec_Int_t) *Vec_Int_t {
    var vFons *Vec_Int_t  = Vec_IntAlloc( 100 );
    var vVisFons *Vec_Bit_t  = Vec_BitStart( Cba_NtkFonNum(p) + 1 );
    var i, k, iObj, iFin, iFon, Entry int32;
    // mark fanin fons
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vObjs)) && (((iObj) = Vec_IntEntry(vObjs, i)), 1); i++  {
        // inlining of macro Cba_ObjForEachFinFon
        for  k = 0, iFin = Cba_ObjFin0(p, iObj); iFin < Cba_ObjFin0(p, iObj+1) && ((iFon = Cba_FinFon(p, iFin)), 1); iFin, k = iFin+1, k+1  {
            if iFon > 0 {
                Vec_BitWriteEntry( vVisFons, iFon, 1 ); }
        }
    }
    // unmark internal fons
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vObjs)) && (((iObj) = Vec_IntEntry(vObjs, i)), 1); i++  {
        // inlining of macro Cba_ObjForEachFon
        for  k = 0, iFon = Cba_ObjFon0(p, iObj); iFon < Cba_ObjFon0(p, iObj+1); iFon, k = iFon+1, k+1  {
            Vec_BitWriteEntry( vVisFons, iFon, 0 );
        }
    }
    // collect fons
    // inlining of macro Vec_BitForEachEntry
    for  iFon = 0; (iFon < Vec_BitSize(vVisFons)) && (((Entry) = Vec_BitEntry(vVisFons, iFon)), 1); iFon++  {
        if Entry != 0 {
            Vec_IntPush( vFons, iFon ); }
    }
    Vec_BitFree( vVisFons );
    return vFons;
}
func Cba_NtkCollectOutFons(p *Cba_Ntk_t, vObjs *Vec_Int_t) *Vec_Int_t {
    var vFons *Vec_Int_t  = Vec_IntAlloc( 100 );
    var vMapObjs *Vec_Bit_t  = Vec_BitStart( Cba_NtkObjNum(p) + 1 );
    var vVisFons *Vec_Bit_t  = Vec_BitStart( Cba_NtkFonNum(p) + 1 );
    var i, k, iObj, iFin, iFon int32;
    // map objects
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vObjs)) && (((iObj) = Vec_IntEntry(vObjs, i)), 1); i++  {
        Vec_BitWriteEntry( vMapObjs, iObj, 1 );
    }
    // mark those used by non-objects
    // inlining of macro Cba_NtkForEachObj
    for  iObj = 1; iObj < Vec_StrSize(&p.vObjType); iObj++  { if !Cba_ObjType(p, iObj) {} else   
        if !Vec_BitEntry(vMapObjs, iObj) {
            // inlining of macro Cba_ObjForEachFinFon
            for  k = 0, iFin = Cba_ObjFin0(p, iObj); iFin < Cba_ObjFin0(p, iObj+1) && ((iFon = Cba_FinFon(p, iFin)), 1); iFin, k = iFin+1, k+1  {
                if iFon > 0 {
                    Vec_BitWriteEntry( vVisFons, iFon, 1 ); }
            } }
    }
    // collect pointed fons among those in objects
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vObjs)) && (((iObj) = Vec_IntEntry(vObjs, i)), 1); i++  {
        // inlining of macro Cba_ObjForEachFon
        for  k = 0, iFon = Cba_ObjFon0(p, iObj); iFon < Cba_ObjFon0(p, iObj+1); iFon, k = iFon+1, k+1  {
            if Vec_BitEntry(vVisFons, iFon) {
                Vec_IntPush( vFons, iFon ); }
        }
    }
    Vec_BitFree( vMapObjs );
    Vec_BitFree( vVisFons );
    return vFons;
}
func Cba_NtkCollectGroupStats(p *Cba_Ntk_t, vObjs *Vec_Int_t, pnFins *int32, pnFons *int32) {
    var i, iObj, nFins, nFons int32;
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vObjs)) && (((iObj) = Vec_IntEntry(vObjs, i)), 1); i++  {
        nFins += Cba_ObjFinNum(p, iObj);
        nFons += Cba_ObjFonNum(p, iObj);
    }
    *pnFins = nFins;
    *pnFons = nFons;
}
func Cba_ManExtractGroupInt(pNew *Cba_Ntk_t, p *Cba_Ntk_t, vObjs *Vec_Int_t, vFonIns *Vec_Int_t, vFonOuts *Vec_Int_t) {
    var i, k, iObj, iObjNew, iFin, iFon int32;
    Cba_NtkCleanObjCopies( p );
    Cba_NtkCleanFonCopies( p );
    // create inputs and map fons
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vFonIns)) && (((iFon) = Vec_IntEntry(vFonIns, i)), 1); i++  {
        iObjNew = Cba_ObjAlloc( pNew, CBA_OBJ_PI, 0, 1 );
        Cba_FonSetCopy( p, iFon, Cba_ObjFon0(pNew, iObjNew) );
        if Cba_NtkHasObjNames(p) {
            Cba_ObjSetName( pNew, iObjNew, Cba_ObjName(p, Cba_FonObj(p, iFon)) ); }
        if Cba_NtkHasFonNames(p) {
            Cba_FonSetName( pNew, Cba_ObjFon0(pNew, iObjNew), Cba_FonName(p, iFon) ); }

    }
    // create internal
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vObjs)) && (((iObj) = Vec_IntEntry(vObjs, i)), 1); i++  {
        iObjNew = Cba_ObjDup( pNew, p, iObj );
        if Cba_NtkHasObjNames(p) {
            Cba_ObjSetName( pNew, iObjNew, Cba_ObjName(p, iObj) ); }
        // inlining of macro Cba_ObjForEachFon
        for  k = 0, iFon = Cba_ObjFon0(p, iObj); iFon < Cba_ObjFon0(p, iObj+1); iFon, k = iFon+1, k+1  {
            Cba_FonSetCopy( p, iFon, Cba_ObjFon(pNew, iObjNew, k) );
            if Cba_NtkHasFonNames(p) {
                Cba_FonSetName( pNew, Cba_ObjFon(pNew, iObjNew, k), Cba_FonName(p, iFon) ); }
        }
    }
    // connect internal
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vObjs)) && (((iObj) = Vec_IntEntry(vObjs, i)), 1); i++  {
        iObjNew = Cba_ObjCopy( p, iObj );
        // inlining of macro Cba_ObjForEachFinFon
        for  k = 0, iFin = Cba_ObjFin0(p, iObj); iFin < Cba_ObjFin0(p, iObj+1) && ((iFon = Cba_FinFon(p, iFin)), 1); iFin, k = iFin+1, k+1  {
            Cba_ObjSetFinFon( pNew, iObjNew, k, Cba_FonCopy(p, iFon) );
        }
    }
    // create POs
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vFonOuts)) && (((iFon) = Vec_IntEntry(vFonOuts, i)), 1); i++  {
        iObjNew = Cba_ObjAlloc( pNew, CBA_OBJ_PO, 1, 0 );
        if Cba_NtkHasObjNames(p) {
            Cba_ObjSetName( pNew, iObjNew, Cba_FonName(p, iFon) ); }
        Cba_ObjSetFinFon( pNew, iObjNew, 0, Cba_FonCopy(p, iFon) );
    }
    assert(Cba_NtkObjNum(pNew) == Cba_NtkObjNumAlloc(pNew), "Cba_NtkObjNum(pNew) == Cba_NtkObjNumAlloc(pNew)");
    assert(Cba_NtkFinNum(pNew) == Cba_NtkFinNumAlloc(pNew), "Cba_NtkFinNum(pNew) == Cba_NtkFinNumAlloc(pNew)");
    assert(Cba_NtkFonNum(pNew) == Cba_NtkFonNumAlloc(pNew), "Cba_NtkFonNum(pNew) == Cba_NtkFonNumAlloc(pNew)");
}
func Cba_ManExtractGroup(p *Cba_Man_t, vObjs *Vec_Int_t) *Cba_Man_t {
    var pNew *Cba_Man_t   = Cba_ManAlloc( p.pSpec, 1, Abc_NamRef(p.pStrs), Abc_NamRef(p.pFuns), Abc_NamStart(100, 24), Hash_IntManRef(p.vHash) );
    var pRoot *Cba_Ntk_t = Cba_ManRoot( p );
    var pRootNew *Cba_Ntk_t;
    var vFonIns *Vec_Int_t  = Cba_NtkCollectInFons( pRoot, vObjs );
    var vFonOuts *Vec_Int_t  = Cba_NtkCollectOutFons( pRoot, vObjs );
    var nObjs, nFins, nFons int32;
    Cba_ManDupTypeNames( pNew, p );
    // collect stats
    Cba_NtkCollectGroupStats( pRoot, vObjs, &nFins, &nFons );
    nObjs  = Vec_IntSize(vObjs) + Vec_IntSize(vFonIns) + Vec_IntSize(vFonOuts);
    nFins += Vec_IntSize(vFonOuts);
    nFons += Vec_IntSize(vFonIns);
    // create network
    pRootNew = Cba_NtkAlloc( pNew, Cba_NtkNameId(pRoot), Vec_IntSize(vFonIns), Vec_IntSize(vFonOuts), nObjs, nFins, nFons );
    Cba_NtkAdd( pNew, pRootNew );
    if Cba_NtkHasObjNames(pRoot) {
        Cba_NtkCleanObjNames( pRootNew ); }
    if Cba_NtkHasFonNames(pRoot) {
        Cba_NtkCleanFonNames( pRootNew ); }
    // add group nodes
    Cba_ManExtractGroupInt( pRootNew, pRoot, vObjs, vFonIns, vFonOuts );
    Cba_NtkMissingFonNames( pRootNew, "b" );
    //Cba_NtkPrepareSeq( pRootNew );
    // cleanup
    Vec_IntFree( vFonIns );
    Vec_IntFree( vFonOuts );
    return pNew;
}


/**Function*************************************************************

  Synopsis    [Derives the design from the GIA manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cba_NtkInsertGiaLit(p *Cba_Ntk_t, iLit int32, vLit2Fon *Vec_Int_t, fUseXor int32) int32 {
    var iObjNew int32;
    if iLit == 0 || iLit == 1 {
        return Cba_FonFromConst(iLit); }
    if Vec_IntEntry(vLit2Fon, iLit) >= 0 {
        return Vec_IntEntry(vLit2Fon, iLit); }
    assert(Abc_LitIsCompl(iLit), "Abc_LitIsCompl(iLit)");
    assert(Vec_IntEntry(vLit2Fon, Abc_LitNot(iLit)) >= 0, "Vec_IntEntry(vLit2Fon, Abc_LitNot(iLit)) >= 0");
    // create inverter
    if fUseXor != 0 {
        iObjNew = Cba_ObjAlloc( p, CBA_BOX_XOR, 2, 1 );
        Cba_ObjSetFinFon( p, iObjNew, 0, Vec_IntEntry(vLit2Fon, Abc_LitNot(iLit)) );
        Cba_ObjSetFinFon( p, iObjNew, 1, Cba_FonFromConst(1) );
    
    } else {
        iObjNew = Cba_ObjAlloc( p, CBA_BOX_INV, 1, 1 );
        Cba_ObjSetFinFon( p, iObjNew, 0, Vec_IntEntry(vLit2Fon, Abc_LitNot(iLit)) );
    }
    // save the result
    Vec_IntWriteEntry( vLit2Fon, iLit, Cba_ObjFon0(p, iObjNew) );
    return Cba_ObjFon0(p, iObjNew);
}
func Cba_NtkInsertGiaObj(p *Cba_Ntk_t, pGia *Gia_Man_t, iObj int32, vLit2Fon *Vec_Int_t, fUseXor int32) int32 {
    var pObj *Gia_Obj_t  = Gia_ManObj( pGia, iObj );
    var iLit0 int32  = Gia_ObjFaninLit0( pObj, iObj );
    var iLit1 int32  = Gia_ObjFaninLit1( pObj, iObj );
    var iFon0 int32  = Cba_NtkInsertGiaLit( p, iLit0, vLit2Fon, fUseXor );
    var iFon1 int32  = Cba_NtkInsertGiaLit( p, iLit1, vLit2Fon, fUseXor );
    var iObjNew int32;
    if Gia_ObjIsMux(pGia, pObj) {
        var iLit2 int32  = Gia_ObjFaninLit2( pGia, iObj );
        var iFon2 int32  = Cba_NtkInsertGiaLit( p, iLit2, vLit2Fon, fUseXor );
        iObjNew = Cba_ObjAlloc( p, CBA_BOX_MUX, 3, 1 );
        Cba_ObjSetFinFon( p, iObjNew, 0, iFon2 );
        Cba_ObjSetFinFon( p, iObjNew, 1, iFon1 );
        Cba_ObjSetFinFon( p, iObjNew, 2, iFon0 );
    
    } else {
        assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
        iObjNew = Cba_ObjAlloc( p, Gia_ObjIsXor(pObj) ? CBA_BOX_XOR : CBA_BOX_AND, 2, 1 );
        Cba_ObjSetFinFon( p, iObjNew, 0, iFon0 );
        Cba_ObjSetFinFon( p, iObjNew, 1, iFon1 );
    }
    Vec_IntWriteEntry( vLit2Fon, Abc_Var2Lit(iObj, 0), Cba_ObjFon0(p, iObjNew) );
    return iObjNew;
}
func Cba_ManDeriveFromGia(pOld *Cba_Man_t, pGia *Gia_Man_t, fUseXor int32) *Cba_Man_t {
    var p *Cba_Man_t  = Cba_ManAlloc( pGia.pSpec, 1, pOld ? Abc_NamRef(pOld.pStrs) : nil, pOld ? Abc_NamRef(pOld.pFuns) : nil, nil, nil );
    var pNtk *Cba_Ntk_t  = Cba_NtkAlloc( p, Abc_NamStrFindOrAdd(p.pStrs, pGia.pName, nil), Gia_ManCiNum(pGia), Gia_ManCoNum(pGia), 1000, 2000, 2000 );
    var vLit2Fon *Vec_Int_t  = Vec_IntStartFull( 2*Gia_ManObjNum(pGia) );
    var i, iObj, iObjNew, NameId, iLit0, iFon0 int32;
    var pObj *Gia_Obj_t;
    //Cba_ManPrepareTypeNames( p );
    Cba_NtkAdd( p, pNtk );
    Cba_NtkCleanObjNames( pNtk );
    // inlining of macro Gia_ManForEachCiId
    for  i = 0; (i < Vec_IntSize(pGia.vCis)) && ((iObj) = Gia_ObjId(pGia, Gia_ManCi(pGia, i))); i++  {
        if pGia.vNamesIn {
            NameId = Abc_NamStrFindOrAdd(p.pStrs, (char*)Vec_PtrEntry(pGia.vNamesIn, i), nil);
        } else {
            NameId = Cba_NtkNewStrId(pNtk, "i%d", i);
        }
        iObjNew = Cba_ObjAlloc( pNtk, CBA_OBJ_PI, 0, 1 );
        Cba_ObjSetName( pNtk, iObjNew, NameId );
        Vec_IntWriteEntry( vLit2Fon, Abc_Var2Lit(iObj, 0), Cba_ObjFon0(pNtk, iObjNew) );
    }
    // inlining of macro Gia_ManForEachAndId
    for  iObj = 0; (iObj < pGia.nObjs); iObj++  {                                     if !Gia_ObjIsAnd(Gia_ManObj(pGia, iObj)) {} else {
        Cba_NtkInsertGiaObj( pNtk, pGia, iObj, vLit2Fon, fUseXor ); }
    }
    // create inverters if needed
    // inlining of macro Gia_ManForEachCoId
    for  i = 0; (i < Vec_IntSize(pGia.vCos)) && ((iObj) = Gia_ObjId(pGia, Gia_ManCo(pGia, i))); i++  {
        pObj = Gia_ManObj( pGia, iObj );
        iLit0 = Gia_ObjFaninLit0( pObj, iObj );
        iFon0 = Cba_NtkInsertGiaLit( pNtk, iLit0, vLit2Fon, fUseXor ); // can be const!
    }
    // inlining of macro Gia_ManForEachCoId
    for  i = 0; (i < Vec_IntSize(pGia.vCos)) && ((iObj) = Gia_ObjId(pGia, Gia_ManCo(pGia, i))); i++  {
        pObj = Gia_ManObj( pGia, iObj );
        iLit0 = Gia_ObjFaninLit0( pObj, iObj );
        iFon0 = Cba_NtkInsertGiaLit( pNtk, iLit0, vLit2Fon, fUseXor ); // can be const!
        iObjNew = Cba_ObjAlloc( pNtk, CBA_BOX_BUF, 1, 1 );
        Cba_ObjSetFinFon( pNtk, iObjNew, 0, iFon0 );
        iFon0 = Cba_ObjFon0(pNtk, iObjNew); // non-const fon unique for this output
        if pGia.vNamesOut {
            NameId = Abc_NamStrFindOrAdd(p.pStrs, (char*)Vec_PtrEntry(pGia.vNamesOut, i), nil);
        } else {
            NameId = Cba_NtkNewStrId(pNtk, "o%d", i);
        }
        iObjNew = Cba_ObjAlloc( pNtk, CBA_OBJ_PO, 1, 0 );
        Cba_ObjSetName( pNtk, iObjNew, NameId );
        Cba_ObjSetFinFon( pNtk, iObjNew, 0, iFon0 );
    }
    Cba_NtkCleanFonNames( pNtk );
    Cba_NtkCreateFonNames( pNtk, "a" );
    Vec_IntFree( vLit2Fon );
    return p;
}


/**Function*************************************************************

  Synopsis    [Inserts the network into the root module instead of objects.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cba_NtkInsertGroup(p *Cba_Ntk_t, vObjs *Vec_Int_t, pSyn *Cba_Ntk_t) {
    var vFonIns *Vec_Int_t   = Cba_NtkCollectInFons( p, vObjs );
    var vFonOuts *Vec_Int_t  = Cba_NtkCollectOutFons( p, vObjs );
    var k, iObj, iObjNew, iFin, iFon int32;
    assert(Cba_NtkPiNum(pSyn) == Vec_IntSize(vFonIns), "Cba_NtkPiNum(pSyn) == Vec_IntSize(vFonIns)");
    assert(Cba_NtkPoNum(pSyn) == Vec_IntSize(vFonOuts), "Cba_NtkPoNum(pSyn) == Vec_IntSize(vFonOuts)");
    // mark AIG with the input fons
    Cba_NtkCleanFonCopies( pSyn );
    // inlining of macro Cba_NtkForEachPiFon
    for  k = 0; (k < Cba_NtkPiNum(pSyn))  && (((iObj) = Cba_NtkPi(pSyn, k)), 1)  && (((iFon) = Cba_ObjFon0(pSyn, iObj)), 1); k++  { 
        Cba_FonSetCopy( pSyn, iFon, Vec_IntEntry(vFonIns, k) );
    }
    Vec_IntFree( vFonIns );
    // build up internal nodes
    Cba_NtkCleanObjCopies( pSyn );
    // inlining of macro Cba_NtkForEachBox
    for  iObj = 1; iObj < Vec_StrSize(&pSyn.vObjType); iObj++  { if !Cba_ObjIsBox(pSyn, iObj) {} else {
        iObjNew = Cba_ObjDup( p, pSyn, iObj );
        // inlining of macro Cba_ObjForEachFon
        for  k = 0, iFon = Cba_ObjFon0(pSyn, iObj); iFon < Cba_ObjFon0(pSyn, iObj+1); iFon, k = iFon+1, k+1  {
            Cba_FonSetCopy( pSyn, iFon, Cba_ObjFon(p, iObjNew, k) );
        }
    }
    }
    // connect internal nodes
    // inlining of macro Cba_NtkForEachBox
    for  iObj = 1; iObj < Vec_StrSize(&pSyn.vObjType); iObj++  { if !Cba_ObjIsBox(pSyn, iObj) {} else {
        iObjNew = Cba_ObjCopy( pSyn, iObj );
        // inlining of macro Cba_ObjForEachFinFon
        for  k = 0, iFin = Cba_ObjFin0(pSyn, iObj); iFin < Cba_ObjFin0(pSyn, iObj+1) && ((iFon = Cba_FinFon(pSyn, iFin)), 1); iFin, k = iFin+1, k+1  {
            Cba_ObjSetFinFon( p, iObjNew, k, Cba_FonCopy(pSyn, iFon) );
        }
    }
    }
    // connect output fons
    Cba_NtkCleanFonCopies( p );
    if Cba_NtkHasFonNames(p) {
        Vec_IntFillExtra( &p.vFonName, Cba_NtkFonNum(p) + 1, 0 ); }
    // inlining of macro Cba_NtkForEachPoDriverFon
    for  k = 0; (k < Cba_NtkPoNum(pSyn))  && (((iObj) = Cba_NtkPo(pSyn, k)), 1)  && (((iFon) = Cba_ObjFinFon(pSyn, iObj, 0)), 1); k++  {
        assert(Cba_FonIsReal(Cba_FonCopy(pSyn, iFon)), "Cba_FonIsReal(Cba_FonCopy(pSyn, iFon))");
        Cba_FonSetCopy( p, Vec_IntEntry(vFonOuts, k), Cba_FonCopy(pSyn, iFon) );
        // transfer names
        if Cba_NtkHasFonNames(p) {
            Cba_FonSetName( p, Cba_FonCopy(pSyn, iFon), Cba_FonName(p, Vec_IntEntry(vFonOuts, k)) );
            Cba_FonCleanName( p, Vec_IntEntry(vFonOuts, k) );
        }
    }
    Vec_IntFree( vFonOuts );
    // delete nodes
//    Vec_IntForEachEntry( vObjs, iObj, k )
//        Cba_ObjDelete( p, iObj );
    // update fins pointing to output fons to point to the new fons
    // inlining of macro Cba_NtkForEachFinFon
    for  iFin = 1; iFin < Vec_IntSize(&p.vFinFon) && (((iFon) = Vec_IntEntry(&p.vFinFon, iFin)), 1); iFin++  { if !(iFon != 0) {} else   
        if Cba_FonIsReal(iFon) && Cba_FonCopy(p, iFon) {
            Cba_PatchFinFon( p, iFin, Cba_FonCopy(p, iFon) ); }
    }
    Cba_NtkMissingFonNames( p, "j" );
/*
    // duplicate in DFS order
    pNew = Cba_NtkDupOrder( p->pDesign, p, Cba_NtkCollectDfs );
    Cba_NtkDupAttrs( pNew, p );
    // replace "p" with "pNew"
    Cba_NtkUpdate( Cba_NtkMan(p), pNew ); // removes "p"
    return pNew;
*/
}
func Cba_ManInsertGroup(p *Cba_Man_t, vObjs *Vec_Int_t, pSyn *Cba_Ntk_t) *Cba_Man_t {
    Cba_NtkInsertGroup( Cba_ManRoot(p), vObjs, pSyn );
    Cba_NtkCheckComboLoop( Cba_ManRoot(p) );
    return Cba_ManDup( p, Cba_NtkCollectDfs );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

