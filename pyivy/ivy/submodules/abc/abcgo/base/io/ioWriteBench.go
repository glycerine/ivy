package io

import "fmt"

import "os"

/**CFile****************************************************************

  FileName    [ioWriteBench.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Command processing package.]

  Synopsis    [Procedures to write the network in BENCH format.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: ioWriteBench.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "ioAbc.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// static int Io_WriteBenchCheckNames( Abc_Ntk_t * pNtk );


// static int Io_WriteBenchOne( FILE * pFile, Abc_Ntk_t * pNtk );

// static int Io_WriteBenchOneNode( FILE * pFile, Abc_Obj_t * pNode );


// static int Io_WriteBenchLutOne( FILE * pFile, Abc_Ntk_t * pNtk );

// static int Io_WriteBenchLutOneNode( FILE * pFile, Abc_Obj_t * pNode, Vec_Int_t * vTruth );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Writes the network in BENCH format.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_WriteBench(pNtk *Abc_Ntk_t, pFileName string) int32 {
    var pExdc *Abc_Ntk_t;
    var pFile *os.File;
    assert(Abc_NtkIsSopNetlist(pNtk), "Abc_NtkIsSopNetlist(pNtk)");
    if !(Io_WriteBenchCheckNames(pNtk) != 0) {
        fmt.Fprintf(os.Stdout, "Io_WriteBench(): Signal names in this benchmark contain parentheses making them impossible to reproduce in the BENCH format. Use \"short_names\".\n");
        return 0;
    }
    pFile = os.OpenFile(pFileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    if pFile == nil {
        fmt.Fprintf(os.Stdout, "Io_WriteBench(): Cannot open the output file.\n");
        return 0;
    }
    fmt.Fprintf(pFile, "# Benchmark \"%s\" written by ABC on %s\n", pNtk.pName, Extra_TimeStamp());
    // write the network
    Io_WriteBenchOne( pFile, pNtk );
    // write EXDC network if it exists
    pExdc = Abc_NtkExdc( pNtk );
    if pExdc != nil {
        fmt.Printf("Io_WriteBench: EXDC is not written (warning).\n"); }
    // finalize the file
    _ = (pFile).Close();
    return 1;
}

/**Function*************************************************************

  Synopsis    [Writes the network in BENCH format.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_WriteBenchOne(pFile *os.File, pNtk *Abc_Ntk_t) int32 {
    var pProgress *ProgressBar;
    var pNode *Abc_Obj_t;
    var i int32;

    // write the PIs/POs/latches
    // inlining of macro Abc_NtkForEachPi
    for  i = 0; (i < Abc_NtkPiNum(pNtk)) && (((pNode) = Abc_NtkPi(pNtk, i)), 1); i++  {
        fmt.Fprintf(pFile, "INPUT(%s)\n", Abc_ObjName(Abc_ObjFanout0(pNode)));
    }
    // inlining of macro Abc_NtkForEachPo
    for  i = 0; (i < Abc_NtkPoNum(pNtk)) && (((pNode) = Abc_NtkPo(pNtk, i)), 1); i++  {
        fmt.Fprintf(pFile, "OUTPUT(%s)\n", Abc_ObjName(Abc_ObjFanin0(pNode)));
    }
    // inlining of macro Abc_NtkForEachLatch
    for  i = 0; (i < Vec_PtrSize((pNtk).vBoxes)) && (((pNode) = Abc_NtkBox(pNtk, i)), 1); i++  {           if !Abc_ObjIsLatch(pNode) {} else {
        fmt.Fprintf(pFile, "%-11s = DFF(%s)\n", Abc_ObjName(Abc_ObjFanout0(Abc_ObjFanout0(pNode))), Abc_ObjName(Abc_ObjFanin0(Abc_ObjFanin0(pNode)))); }
    }

    // write internal nodes
    pProgress = Extra_ProgressBarStart( os.Stdout, Abc_NtkObjNumMax(pNtk) );
    // inlining of macro Abc_NtkForEachNode
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pNode) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pNode) == nil || !Abc_ObjIsNode(pNode) {} else {
        Extra_ProgressBarUpdate( pProgress, i, nil );
        Io_WriteBenchOneNode( pFile, pNode );
    }
    }
    Extra_ProgressBarStop( pProgress );
    return 1;
}


/**Function*************************************************************

  Synopsis    [Writes the network in BENCH format.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_WriteBenchOneNode(pFile *os.File, pNode *Abc_Obj_t) int32 {
    var nFanins int32;

    assert(Abc_ObjIsNode(pNode), "Abc_ObjIsNode(pNode)");
    nFanins = Abc_ObjFaninNum(pNode);
    if nFanins == 0 {   // write the constant 1 node
        assert(Abc_NodeIsConst1(pNode), "Abc_NodeIsConst1(pNode)");
        fmt.Fprintf(pFile, "%-11s", Abc_ObjName(Abc_ObjFanout0(pNode)));
        fmt.Fprintf(pFile, " = vdd\n");
    
    } else if nFanins == 1 {   // write the interver/buffer
        if Abc_NodeIsBuf(pNode) {
            fmt.Fprintf(pFile, "%-11s = BUFF(", Abc_ObjName(Abc_ObjFanout0(pNode)));
            fmt.Fprintf(pFile, "%s)\n", Abc_ObjName(Abc_ObjFanin0(pNode)));
        
        } else {
            fmt.Fprintf(pFile, "%-11s = NOT(", Abc_ObjName(Abc_ObjFanout0(pNode)));
            fmt.Fprintf(pFile, "%s)\n", Abc_ObjName(Abc_ObjFanin0(pNode)));
        }
    
    } else {   // write the AND gate
        fmt.Fprintf(pFile, "%-11s", Abc_ObjName(Abc_ObjFanout0(pNode)));
        fmt.Fprintf(pFile, " = AND(%s, ", Abc_ObjName(Abc_ObjFanin0(pNode)));
        fmt.Fprintf(pFile, "%s)\n", Abc_ObjName(Abc_ObjFanin1(pNode)));
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Writes the network in BENCH format with LUTs and DFFRSE.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_WriteBenchLut(pNtk *Abc_Ntk_t, pFileName string) int32 {
    var pExdc *Abc_Ntk_t;
    var pFile *os.File;
    assert(Abc_NtkIsAigNetlist(pNtk), "Abc_NtkIsAigNetlist(pNtk)");
    if !(Io_WriteBenchCheckNames(pNtk) != 0) {
        fmt.Fprintf(os.Stdout, "Io_WriteBenchLut(): Signal names in this benchmark contain parentheses making them impossible to reproduce in the BENCH format. Use \"short_names\".\n");
        return 0;
    }
    pFile = os.OpenFile(pFileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    if pFile == nil {
        fmt.Fprintf(os.Stdout, "Io_WriteBench(): Cannot open the output file.\n");
        return 0;
    }
    fmt.Fprintf(pFile, "# Benchmark \"%s\" written by ABC on %s\n", pNtk.pName, Extra_TimeStamp());
    // write the network
    Io_WriteBenchLutOne( pFile, pNtk );
    // write EXDC network if it exists
    pExdc = Abc_NtkExdc( pNtk );
    if pExdc != nil {
        fmt.Printf("Io_WriteBench: EXDC is not written (warning).\n"); }
    // finalize the file
    _ = (pFile).Close();
    return 1;
}

/**Function*************************************************************

  Synopsis    [Writes the network in BENCH format.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_WriteBenchLutOne(pFile *os.File, pNtk *Abc_Ntk_t) int32 {
    var pProgress *ProgressBar;
    var pNode *Abc_Obj_t;
    var vMemory *Vec_Int_t;
    var i int32;

    // write the PIs/POs/latches
    // inlining of macro Abc_NtkForEachPi
    for  i = 0; (i < Abc_NtkPiNum(pNtk)) && (((pNode) = Abc_NtkPi(pNtk, i)), 1); i++  {
        fmt.Fprintf(pFile, "INPUT(%s)\n", Abc_ObjName(Abc_ObjFanout0(pNode)));
    }
    // inlining of macro Abc_NtkForEachPo
    for  i = 0; (i < Abc_NtkPoNum(pNtk)) && (((pNode) = Abc_NtkPo(pNtk, i)), 1); i++  {
        fmt.Fprintf(pFile, "OUTPUT(%s)\n", Abc_ObjName(Abc_ObjFanin0(pNode)));
    }
    // inlining of macro Abc_NtkForEachLatch
    for  i = 0; (i < Vec_PtrSize((pNtk).vBoxes)) && (((pNode) = Abc_NtkBox(pNtk, i)), 1); i++  {           if !Abc_ObjIsLatch(pNode) {} else {
        fmt.Fprintf(pFile, "%-11s = DFFRSE( %s, gnd, gnd, gnd, gnd )\n", Abc_ObjName(Abc_ObjFanout0(Abc_ObjFanout0(pNode))), Abc_ObjName(Abc_ObjFanin0(Abc_ObjFanin0(pNode)))); }
    }
//Abc_NtkLevel(pNtk);
    // write internal nodes
    vMemory = Vec_IntAlloc( 10000 );
    pProgress = Extra_ProgressBarStart( os.Stdout, Abc_NtkObjNumMax(pNtk) );
    // inlining of macro Abc_NtkForEachNode
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pNode) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pNode) == nil || !Abc_ObjIsNode(pNode) {} else {
        Extra_ProgressBarUpdate( pProgress, i, nil );
        Io_WriteBenchLutOneNode( pFile, pNode, vMemory );
    }
    }
    Extra_ProgressBarStop( pProgress );
    Vec_IntFree( vMemory );
    return 1;
}


/**Function*************************************************************

  Synopsis    [Writes the network in BENCH format.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_WriteBenchLutOneNode(pFile *os.File, pNode *Abc_Obj_t, vTruth *Vec_Int_t) int32 {
    var pFanin *Abc_Obj_t;
    var pTruth *uint32;
    var i, nFanins int32;
    assert(Abc_ObjIsNode(pNode), "Abc_ObjIsNode(pNode)");
    nFanins = Abc_ObjFaninNum(pNode);
    assert(nFanins <= 8, "nFanins <= 8");
    // compute the truth table
    pTruth = Hop_ManConvertAigToTruth( (Hop_Man_t *)pNode.pNtk.pManFunc, Hop_Regular((Hop_Obj_t *)pNode.pData), nFanins, vTruth, 0 );
    if Hop_IsComplement((Hop_Obj_t *)pNode.pData) {
        Extra_TruthNot( pTruth, pTruth, nFanins ); }
    // consider simple cases
    if Extra_TruthIsConst0(pTruth, nFanins) {
        fmt.Fprintf(pFile, "%-11s = gnd\n", Abc_ObjName(Abc_ObjFanout0(pNode)));
        return 1;
    }
    if Extra_TruthIsConst1(pTruth, nFanins) {
        fmt.Fprintf(pFile, "%-11s = vdd\n", Abc_ObjName(Abc_ObjFanout0(pNode)));
        return 1;
    }
    if nFanins == 1 {
        fmt.Fprintf(pFile, "%-11s = LUT 0x%d ( %s )\n", Abc_ObjName(Abc_ObjFanout0(pNode)), Abc_NodeIsBuf(pNode)? 2 : 1, Abc_ObjName(Abc_ObjFanin0(pNode)));
        return 1;
    }
    // write it in the hexadecimal form
    fmt.Fprintf(pFile, "%-11s = LUT 0x", Abc_ObjName(Abc_ObjFanout0(pNode)));
    Extra_PrintHexadecimal( pFile, pTruth, nFanins );
/*
    {
extern void Kit_DsdTest( unsigned * pTruth, int nVars );
Abc_ObjForEachFanin( pNode, pFanin, i )
printf( "%c%d ", 'a'+i, Abc_ObjFanin0(pFanin)->Level );
printf( "\n" );
Kit_DsdTest( pTruth, nFanins );
    }
    if ( pNode->Id % 1000 == 0 )
    {
        int x = 0;
    }
*/
    // write the fanins
    fmt.Fprintf(pFile, " (");
    // inlining of macro Abc_ObjForEachFanin
    for  i = 0; (i < Abc_ObjFaninNum(pNode)) && (((pFanin) = Abc_ObjFanin(pNode, i)), 1); i++  {
        fmt.Fprintf(pFile, " %s%s", Abc_ObjName(pFanin), ((i==nFanins-1)? "" : ","));
    }
    fmt.Fprintf(pFile, " )\n");
    return 1;
}


/**Function*************************************************************

  Synopsis    [Returns 1 if the names cannot be written into the bench file.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_WriteBenchCheckNames(pNtk *Abc_Ntk_t) int32 {
    var pObj *Abc_Obj_t;
    var pName string;
    var i int32;
    // inlining of macro Abc_NtkForEachObj
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {            if (pObj) == nil {} else {
        for  pName = Nm_ManFindNameById(pNtk.pManName, i); pName && *pName; pName++  {
            if *pName == '(' || *pName == ')' {
                return 0; }
        } }
    }
    return 1;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

