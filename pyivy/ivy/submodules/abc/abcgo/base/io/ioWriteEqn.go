package io

import "fmt"

import "os"

/**CFile****************************************************************

  FileName    [ioWriteEqn.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Command processing package.]

  Synopsis    [Procedures to write equation representation of the network.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: ioWriteEqn.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "ioAbc.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// static void Io_NtkWriteEqnOne( FILE * pFile, Abc_Ntk_t * pNtk );

// static void Io_NtkWriteEqnCis( FILE * pFile, Abc_Ntk_t * pNtk );

// static void Io_NtkWriteEqnCos( FILE * pFile, Abc_Ntk_t * pNtk );

// static int Io_NtkWriteEqnCheck( Abc_Ntk_t * pNtk );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Writes the logic network in the equation format.]

  Description []
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_WriteEqn(pNtk *Abc_Ntk_t, pFileName string) {
    var pFile *os.File;

    assert(Abc_NtkIsAigNetlist(pNtk), "Abc_NtkIsAigNetlist(pNtk)");
    if Abc_NtkLatchNum(pNtk) > 0 {
        fmt.Printf("Warning: only combinational portion is being written.\n"); }

    // check that the names are fine for the EQN format
    if !(Io_NtkWriteEqnCheck(pNtk) != 0) {
        return; }

    // start the output stream
    pFile = os.OpenFile(pFileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    if pFile == nil {
        fmt.Fprintf(os.Stdout, "Io_WriteEqn(): Cannot open the output file \"%s\".\n", pFileName);
        return;
    }
    fmt.Fprintf(pFile, "# Equations for \"%s\" written by ABC on %s\n", pNtk.pName, Extra_TimeStamp());

    // write the equations for the network
    Io_NtkWriteEqnOne( pFile, pNtk );
    fmt.Fprintf(pFile, "\n");
    _ = (pFile).Close();
}

/**Function*************************************************************

  Synopsis    [Write one network.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_NtkWriteEqnOne(pFile *os.File, pNtk *Abc_Ntk_t) {
    var vLevels *Vec_Vec_t;
    var pProgress *ProgressBar;
    var pNode, pFanin *Abc_Obj_t;
    var i, k int32;

    // write the PIs
    fmt.Fprintf(pFile, "INORDER =");
    Io_NtkWriteEqnCis( pFile, pNtk );
    fmt.Fprintf(pFile, ";\n");

    // write the POs
    fmt.Fprintf(pFile, "OUTORDER =");
    Io_NtkWriteEqnCos( pFile, pNtk );
    fmt.Fprintf(pFile, ";\n");

    // write each internal node
    vLevels = Vec_VecAlloc( 10 );
    pProgress = Extra_ProgressBarStart( os.Stdout, Abc_NtkObjNumMax(pNtk) );
    // inlining of macro Abc_NtkForEachNode
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pNode) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pNode) == nil || !Abc_ObjIsNode(pNode) {} else {
        Extra_ProgressBarUpdate( pProgress, i, nil );
        fmt.Fprintf(pFile, "%s = ", Abc_ObjName(Abc_ObjFanout0(pNode)));
        // set the input names
        // inlining of macro Abc_ObjForEachFanin
        for  k = 0; (k < Abc_ObjFaninNum(pNode)) && (((pFanin) = Abc_ObjFanin(pNode, k)), 1); k++  {
            Hop_IthVar((Hop_Man_t *)pNtk.pManFunc, k).pData = Abc_ObjName(pFanin);
        }
        // write the formula
        Hop_ObjPrintEqn( pFile, (Hop_Obj_t *)pNode.pData, vLevels, 0 );
        fmt.Fprintf(pFile, ";\n");
    }
    }
    Extra_ProgressBarStop( pProgress );
    Vec_VecFree( vLevels );
}


/**Function*************************************************************

  Synopsis    [Writes the primary input list.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_NtkWriteEqnCis(pFile *os.File, pNtk *Abc_Ntk_t) {
    var pTerm, pNet *Abc_Obj_t;
    var LineLength int32;
    var AddedLength int32;
    var NameCounter int32;
    var i int32;

    LineLength  = 9;
    NameCounter = 0;

    // inlining of macro Abc_NtkForEachCi
    for  i = 0; (i < Abc_NtkCiNum(pNtk)) && (((pTerm) = Abc_NtkCi(pNtk, i)), 1); i++  {
        pNet = Abc_ObjFanout0(pTerm);
        // get the line length after this name is written
        AddedLength = len(Abc_ObjName(pNet)) + 1;
        if (NameCounter != 0) && LineLength + AddedLength + 3 > IO_WRITE_LINE_LENGTH { // write the line extender
            fmt.Fprintf(pFile, " \n");
            // reset the line length
            LineLength  = 0;
            NameCounter = 0;
        }
        fmt.Fprintf(pFile, " %s", Abc_ObjName(pNet));
        LineLength += AddedLength;
        NameCounter++;
    }
}

/**Function*************************************************************

  Synopsis    [Writes the primary input list.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_NtkWriteEqnCos(pFile *os.File, pNtk *Abc_Ntk_t) {
    var pTerm, pNet *Abc_Obj_t;
    var LineLength int32;
    var AddedLength int32;
    var NameCounter int32;
    var i int32;

    LineLength  = 10;
    NameCounter = 0;

    // inlining of macro Abc_NtkForEachCo
    for  i = 0; (i < Abc_NtkCoNum(pNtk)) && (((pTerm) = Abc_NtkCo(pNtk, i)), 1); i++  {
        pNet = Abc_ObjFanin0(pTerm);
        // get the line length after this name is written
        AddedLength = len(Abc_ObjName(pNet)) + 1;
        if (NameCounter != 0) && LineLength + AddedLength + 3 > IO_WRITE_LINE_LENGTH { // write the line extender
            fmt.Fprintf(pFile, " \n");
            // reset the line length
            LineLength  = 0;
            NameCounter = 0;
        }
        fmt.Fprintf(pFile, " %s", Abc_ObjName(pNet));
        LineLength += AddedLength;
        NameCounter++;
    }
}

/**Function*************************************************************

  Synopsis    [Make sure the network does not have offending names.]

  Description []
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_NtkWriteEqnCheck(pNtk *Abc_Ntk_t) int32 {
    var pObj *Abc_Obj_t;
    var pName string  = nil;
    var i, k, Length int32;
    var RetValue int32  = 1;

    // make sure the network does not have proper names, such as "0" or "1" or containing parentheses
    // inlining of macro Abc_NtkForEachObj
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {            if (pObj) == nil {} else {
        pName = Nm_ManFindNameById(pNtk.pManName, i);
        if pName == nil {
            continue; }
        Length = len(pName);
        if pName[0] == '0' || pName[0] == '1' {
            RetValue = 0;
            break;
        }
        for  k = 0; k < Length; k++  {
            if pName[k] == '(' || pName[k] == ')' || pName[k] == '!' || pName[k] == '*' || pName[k] == '+' {
                RetValue = 0;
                break;
            }
        }
        if k < Length {
            break; }
    }
    }
    if RetValue == 0 {
        fmt.Printf("The network cannot be written in the EQN format because object %d has name \"%s\".\n", i, pName);
        fmt.Printf("Consider renaming the objects using command \"short_names\" and trying again.\n");
    }
    return RetValue;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

