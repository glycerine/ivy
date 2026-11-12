package saig

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

import "os"

/**CFile****************************************************************

  FileName    [saigIoa.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Sequential AIG package.]

  Synopsis    [Input/output for sequential AIGs using BLIF files.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: saigIoa.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include <math.h>
//#include "saig.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_Saig_ObjName_Buffer string // was [16]byte in C

func Saig_ObjName(p *Aig_Man_t, pObj *Aig_Obj_t) string {
    
    if Aig_ObjIsNode(pObj) || Aig_ObjIsConst1(pObj) {
        spStatic_Saig_ObjName_Buffer = fmt.Sprintf("n%0*d", Abc_Base10Log(Aig_ManObjNumMax(p)), Aig_ObjId(pObj));
    } else if Saig_ObjIsPi(p, pObj) {
        spStatic_Saig_ObjName_Buffer = fmt.Sprintf("pi%0*d", Abc_Base10Log(Saig_ManPiNum(p)), Aig_ObjCioId(pObj));
    } else if Saig_ObjIsPo(p, pObj) {
        spStatic_Saig_ObjName_Buffer = fmt.Sprintf("po%0*d", Abc_Base10Log(Saig_ManPoNum(p)), Aig_ObjCioId(pObj));
    } else if Saig_ObjIsLo(p, pObj) {
        spStatic_Saig_ObjName_Buffer = fmt.Sprintf("lo%0*d", Abc_Base10Log(Saig_ManRegNum(p)), Aig_ObjCioId(pObj) - Saig_ManPiNum(p));
    } else if Saig_ObjIsLi(p, pObj) {
        spStatic_Saig_ObjName_Buffer = fmt.Sprintf("li%0*d", Abc_Base10Log(Saig_ManRegNum(p)), Aig_ObjCioId(pObj) - Saig_ManPoNum(p));
    } else { 
        assert(false, "0"); }
    return spStatic_Saig_ObjName_Buffer;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManDumpBlif(p *Aig_Man_t, pFileName string) {
    var pFile *os.File;
    var pObj, pObjLi, pObjLo *Aig_Obj_t;
    var i int32;
    if Aig_ManCoNum(p) == 0 {
        fmt.Printf("Aig_ManDumpBlif(): AIG manager does not have POs.\n");
        return;
    }
    Aig_ManSetCioIds( p );
    // write input file
    pFile = os.OpenFile(pFileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    if pFile == nil {
        fmt.Printf("Saig_ManDumpBlif(): Cannot open file for writing.\n");
        return;
    }
    fmt.Fprintf(pFile, "# BLIF file written by procedure Saig_ManDumpBlif()\n");
    fmt.Fprintf(pFile, "# If unedited, this file can be read by Saig_ManReadBlif()\n");
    fmt.Fprintf(pFile, "# AIG stats: pi=%d po=%d reg=%d and=%d obj=%d maxid=%d\n", Saig_ManPiNum(p), Saig_ManPoNum(p), Saig_ManRegNum(p), Aig_ManNodeNum(p), Aig_ManObjNum(p), Aig_ManObjNumMax(p));
    fmt.Fprintf(pFile, ".model %s\n", p.pName);
    // write primary inputs
    fmt.Fprintf(pFile, ".inputs");
    // inlining of macro Aig_ManForEachPiSeq
    // inlining of macro Vec_PtrForEachEntryStop
    for  i = 0; (i < Aig_ManCiNum(p)-Aig_ManRegNum(p)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCis, i)), 1); i++  {
        fmt.Fprintf(pFile, " %s", Saig_ObjName(p, pObj));
    }
    fmt.Fprintf(pFile, "\n");
    // write primary outputs
    fmt.Fprintf(pFile, ".outputs");
    // inlining of macro Aig_ManForEachPoSeq
    // inlining of macro Vec_PtrForEachEntryStop
    for  i = 0; (i < Aig_ManCoNum(p)-Aig_ManRegNum(p)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
        fmt.Fprintf(pFile, " %s", Saig_ObjName(p, pObj));
    }
    fmt.Fprintf(pFile, "\n");
    // write registers
    if Aig_ManRegNum(p) {
        // inlining of macro Aig_ManForEachLiLoSeq
        for  i = 0; (i < Aig_ManRegNum(p)) && (((pObjLi) = Aig_ManLi(p, i)), 1)            && (((pObjLo)=Aig_ManLo(p, i)), 1); i++  {
            fmt.Fprintf(pFile, ".latch");
            fmt.Fprintf(pFile, " %s", Saig_ObjName(p, pObjLi));
            fmt.Fprintf(pFile, " %s", Saig_ObjName(p, pObjLo));
            fmt.Fprintf(pFile, " 0\n");
        }
    } 
    // check if constant is used
    if Aig_ObjRefs(Aig_ManConst1(p)) {
        fmt.Fprintf(pFile, ".names %s\n 1\n", Saig_ObjName(p, Aig_ManConst1(p))); }
    // write the nodes in the DFS order
    // inlining of macro Aig_ManForEachNode
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
        fmt.Fprintf(pFile, ".names");
        fmt.Fprintf(pFile, " %s", Saig_ObjName(p, Aig_ObjFanin0(pObj)));
        fmt.Fprintf(pFile, " %s", Saig_ObjName(p, Aig_ObjFanin1(pObj)));
        fmt.Fprintf(pFile, " %s", Saig_ObjName(p, pObj));
        fmt.Fprintf(pFile, "\n%d%d 1\n", !Aig_ObjFaninC0(pObj), !Aig_ObjFaninC1(pObj));
    }
    }
    // write the POs
    // inlining of macro Aig_ManForEachCo
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
        fmt.Fprintf(pFile, ".names");
        fmt.Fprintf(pFile, " %s", Saig_ObjName(p, Aig_ObjFanin0(pObj)));
        fmt.Fprintf(pFile, " %s", Saig_ObjName(p, pObj));
        fmt.Fprintf(pFile, "\n%d 1\n", !Aig_ObjFaninC0(pObj));
    }
    fmt.Fprintf(pFile, ".end\n");
    _ = (pFile).Close();
}

/**Function*************************************************************

  Synopsis    [Reads one token from file.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_Saig_ManReadToken_Buffer [1000]byte;
func Saig_ManReadToken(pFile *os.File) string {
    
    if fscanf( pFile, "%s", spStatic_Saig_ManReadToken_Buffer ) == 1 {
        return spStatic_Saig_ManReadToken_Buffer; }
    return nil;
}

/**Function*************************************************************

  Synopsis    [Returns the corresponding number.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManReadNumber(p *Aig_Man_t, pToken string) int32 {
    if pToken[0] == 'n' {
        return strconv.Atoi(pToken + 1); }
    if pToken[0] == 'p' {
        return strconv.Atoi(pToken + 2); }
    if pToken[0] == 'l' {
        return strconv.Atoi(pToken + 2); }
    assert(false, "0");
    return -1;
}

/**Function*************************************************************

  Synopsis    [Returns the corresponding node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManReadNode(p *Aig_Man_t, pNum2Id *int32, pToken string) *Aig_Obj_t {
    var Num int32;
    if pToken[0] == 'n' {
        Num = strconv.Atoi(pToken + 1);
        return Aig_ManObj( p, pNum2Id[Num] );
    }
    if pToken[0] == 'p' {
        pToken++;
        if pToken[0] == 'i' {
            Num = strconv.Atoi(pToken + 1);
            return Aig_ManCi( p, Num );
        }
        if pToken[0] == 'o' {
            return nil; }
        assert(false, "0");
        return nil;
    }
    if pToken[0] == 'l' {
        pToken++;
        if pToken[0] == 'o' {
            Num = strconv.Atoi(pToken + 1);
            return Saig_ManLo( p, Num );
        }
        if pToken[0] == 'i' {
            return nil; }
        assert(false, "0");
        return nil;
    }
    assert(false, "0");
    return nil;
}

/**Function*************************************************************

  Synopsis    [Reads BLIF previously dumped by Saig_ManDumpBlif().]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManReadBlif(pFileName string) *Aig_Man_t {
    var pFile *os.File;
    var p *Aig_Man_t;
    var pFanin0, pFanin1, pNode *Aig_Obj_t;
    var pToken string;
    var i, nPis, nPos, nRegs, Number int32;
    var pNum2Id *int32  = nil;  // mapping of node numbers in the file into AIG node IDs
    // open the file
    pFile = os.Open(pFileName);
    if pFile == nil {
        fmt.Printf("Saig_ManReadBlif(): Cannot open file for reading.\n");
        return nil;
    }
    // skip through the comments
    for  i = 0; (pToken = Saig_ManReadToken( pFile )) && pToken[0] != '.'; i++  {}
    if pToken == nil { fmt.Printf("Saig_ManReadBlif(): Error 1.\n"); return nil; }
    // get he model
    pToken = Saig_ManReadToken( pFile );
    if pToken == nil { fmt.Printf("Saig_ManReadBlif(): Error 2.\n"); return nil; }
    // start the package
    p = Aig_ManStart( 10000 );
    p.pName = Abc_UtilStrsav( pToken );
    p.pSpec = Abc_UtilStrsav( pFileName );
    // count PIs
    pToken = Saig_ManReadToken( pFile );
    if pToken == nil || (int32(strings.Compare(pToken, ".inputs")) != 0) { fmt.Printf("Saig_ManReadBlif(): Error 3.\n"); Aig_ManStop(p); return nil; }
    for  nPis = 0; (pToken = Saig_ManReadToken( pFile )) && pToken[0] != '.'; nPis++  {}
    // count POs
    if pToken == nil || (int32(strings.Compare(pToken, ".outputs")) != 0) { fmt.Printf("Saig_ManReadBlif(): Error 4.\n"); Aig_ManStop(p); return nil; }
    for  nPos = 0; (pToken = Saig_ManReadToken( pFile )) && pToken[0] != '.'; nPos++  {}
    // count latches
    if pToken == nil { fmt.Printf("Saig_ManReadBlif(): Error 5.\n"); Aig_ManStop(p); return nil; }
    for  nRegs = 0; int32(strings.Compare(pToken, ".latch")) == 0; nRegs++  {
        pToken = Saig_ManReadToken( pFile );
        if pToken == nil { fmt.Printf("Saig_ManReadBlif(): Error 6.\n"); Aig_ManStop(p); return nil; }
        pToken = Saig_ManReadToken( pFile );
        if pToken == nil { fmt.Printf("Saig_ManReadBlif(): Error 7.\n"); Aig_ManStop(p); return nil; }
        pToken = Saig_ManReadToken( pFile );
        if pToken == nil { fmt.Printf("Saig_ManReadBlif(): Error 8.\n"); Aig_ManStop(p); return nil; }
        pToken = Saig_ManReadToken( pFile );
        if pToken == nil { fmt.Printf("Saig_ManReadBlif(): Error 9.\n"); Aig_ManStop(p); return nil; }
    }
    // create PIs and LOs
    for  i = 0; i < nPis + nRegs; i++  {
        Aig_ObjCreateCi( p );
    }
    Aig_ManSetRegNum( p, nRegs );
    // create nodes
    for  i = 0; int32(strings.Compare(pToken, ".names")) == 0; i++  {
        // first token
        pToken = Saig_ManReadToken( pFile );
        if i == 0 && pToken[0] == 'n' { // constant node
            // read 1
            pToken = Saig_ManReadToken( pFile );
            if pToken == nil || (int32(strings.Compare(pToken, "1")) != 0) { fmt.Printf("Saig_ManReadBlif(): Error 10.\n"); Aig_ManStop(p); return nil; }
            // read next
            pToken = Saig_ManReadToken( pFile );
            if pToken == nil { fmt.Printf("Saig_ManReadBlif(): Error 11.\n"); Aig_ManStop(p); return nil; }
            continue;
        }
        pFanin0 = Saig_ManReadNode( p, pNum2Id, pToken );

        // second token
        pToken = Saig_ManReadToken( pFile );
        if (pToken[0] == 'p' && pToken[1] == 'o') || (pToken[0] == 'l' && pToken[1] == 'i') { // buffer
            // read complemented attribute
            pToken = Saig_ManReadToken( pFile );
            if pToken == nil { fmt.Printf("Saig_ManReadBlif(): Error 12.\n"); Aig_ManStop(p); return nil; }
            if pToken[0] == '0' {
                pFanin0 = Aig_Not(pFanin0); }
            // read 1
            pToken = Saig_ManReadToken( pFile );
            if pToken == nil || (int32(strings.Compare(pToken, "1")) != 0) { fmt.Printf("Saig_ManReadBlif(): Error 13.\n"); Aig_ManStop(p); return nil; }
            Aig_ObjCreateCo( p, pFanin0 );
            // read next
            pToken = Saig_ManReadToken( pFile );
            if pToken == nil { fmt.Printf("Saig_ManReadBlif(): Error 14.\n"); Aig_ManStop(p); return nil; }
            continue;
        }

        // third token
        // regular node
        pFanin1 = Saig_ManReadNode( p, pNum2Id, pToken );
        pToken = Saig_ManReadToken( pFile );
        Number = Saig_ManReadNumber( p, pToken );
        // allocate mapping
        if pNum2Id == nil {
//            extern double pow( double x, double y );
            var Size int32  = int32(math.Pow(float64(10.0), float64(float64((len(pToken) - 1)))));
            // inlining of macro ABC_CALLOC
            pNum2Id = ((int *) calloc((size_t)(Size), sizeof(int)));
        }

        // other tokens
        // get the complemented attributes
        pToken = Saig_ManReadToken( pFile );
        if pToken == nil { fmt.Printf("Saig_ManReadBlif(): Error 15.\n"); Aig_ManStop(p); return nil; }
        if pToken[0] == '0' {
            pFanin0 = Aig_Not(pFanin0); }
        if pToken[1] == '0' {
            pFanin1 = Aig_Not(pFanin1); }
        // read 1
        pToken = Saig_ManReadToken( pFile );
        if pToken == nil || (int32(strings.Compare(pToken, "1")) != 0) { fmt.Printf("Saig_ManReadBlif(): Error 16.\n"); Aig_ManStop(p); return nil; }
        // read next
        pToken = Saig_ManReadToken( pFile );
        if pToken == nil { fmt.Printf("Saig_ManReadBlif(): Error 17.\n"); Aig_ManStop(p); return nil; }

        // create new node
        pNode = Aig_And( p, pFanin0, pFanin1 );
        if Aig_IsComplement(pNode) { fmt.Printf("Saig_ManReadBlif(): Error 18.\n"); Aig_ManStop(p); return nil; }
        // set mapping
        pNum2Id[ Number ] = pNode.Id;
    }
    if pToken == nil || (int32(strings.Compare(pToken, ".end")) != 0) { fmt.Printf("Saig_ManReadBlif(): Error 19.\n"); Aig_ManStop(p); return nil; }
    if nPos + nRegs != Aig_ManCoNum(p) { fmt.Printf("Saig_ManReadBlif(): Error 20.\n"); Aig_ManStop(p); return nil; }
    // add non-node objects to the mapping
    // inlining of macro Aig_ManForEachCi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vCis)) && (((pNode) = (Aig_Obj_t *)Vec_PtrEntry(p.vCis, i)), 1); i++  {
        pNum2Id[pNode.Id] = pNode.Id;
    }
//    ABC_FREE( pNum2Id );
    p.pData = pNum2Id;
    // check the new manager
    Aig_ManSetRegNum( p, nRegs );
    if !Aig_ManCheck(p) {
        fmt.Printf("Saig_ManReadBlif(): Check has failed.\n"); }
    return p;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

