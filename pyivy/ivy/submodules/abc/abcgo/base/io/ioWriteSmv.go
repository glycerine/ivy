package io

import "fmt"

import "os"

/**CFile****************************************************************

  FileName    [ioWriteSmv.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Command processing package.]

  Synopsis    [Procedures to write the network in SMV format.]

  Author      [Satrajit Chatterjee]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: ioWriteSmv.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "ioAbc.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// static int Io_WriteSmvCheckNames( Abc_Ntk_t * pNtk );


// static int Io_WriteSmvOne( FILE * pFile, Abc_Ntk_t * pNtk );

// static int Io_WriteSmvOneNode( FILE * pFile, Abc_Obj_t * pNode );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

// This returns a pointer to a static area, so be careful in using results
// of this function i.e. don't call this twice in the same printf call.
//
// This function replaces '|' with '_' I think abc introduces '|' when
// flattening hierarchy. The '|' is interpreted as a or function by nusmv
// which is unfortunate. This probably should be fixed elsewhere.
var spStatic_cleanUNSAFE_buffer [1024]byte;
func cleanUNSAFE(s string) string {
    var t string;
    
    assert(len(s) < 1024, "strlen(s) < 1024");
    strcpy(spStatic_cleanUNSAFE_buffer, s);
    for t = spStatic_cleanUNSAFE_buffer; *t != 0; ++t { *t = (*t == '|') ? '_' : *t; }
    return spStatic_cleanUNSAFE_buffer;
}

func hasPrefix(needle string, haystack string) int32 {
    return (strncmp(haystack, needle, len(needle)) == 0);
}

/**Function*************************************************************

  Synopsis    [Writes the network in SMV format.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_WriteSmv(pNtk *Abc_Ntk_t, pFileName string) int32 {
    var pExdc *Abc_Ntk_t;
    var pFile *os.File;
    assert(Abc_NtkIsSopNetlist(pNtk), "Abc_NtkIsSopNetlist(pNtk)");
    if !(Io_WriteSmvCheckNames(pNtk) != 0) {
        fmt.Fprintf(os.Stdout, "Io_WriteSmv(): Signal names in this benchmark contain parentheses making them impossible to reproduce in the SMV format. Use \"short_names\".\n");
        return 0;
    }
    pFile = os.OpenFile(pFileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    if pFile == nil {
        fmt.Fprintf(os.Stdout, "Io_WriteSmv(): Cannot open the output file.\n");
        return 0;
    }
    fmt.Fprintf(pFile, "-- benchmark \"%s\" written by ABC on %s\n", pNtk.pName, Extra_TimeStamp());
    // write the network
    Io_WriteSmvOne( pFile, pNtk );
    // write EXDC network if it exists
    pExdc = Abc_NtkExdc( pNtk );
    if pExdc != nil {
        fmt.Printf("Io_WriteSmv: EXDC is not written (warning).\n"); }
    // finalize the file
    _ = (pFile).Close();
    return 1;
}

/**Function*************************************************************

  Synopsis    [Writes the network in SMV format.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_WriteSmvOne(pFile *os.File, pNtk *Abc_Ntk_t) int32 {
    var pProgress *ProgressBar;
    var pNode *Abc_Obj_t;
    var i int32;

    // write the PIs/POs/latches
    fmt.Fprintf(pFile, "MODULE main\n");   // nusmv needs top module to be main
    fmt.Fprintf(pFile, "\n");

    fmt.Fprintf(pFile, "VAR  -- inputs\n");
    // inlining of macro Abc_NtkForEachPi
    for  i = 0; (i < Abc_NtkPiNum(pNtk)) && (((pNode) = Abc_NtkPi(pNtk, i)), 1); i++  {
        fmt.Fprintf(pFile, "    %s : boolean;\n", cleanUNSAFE(Abc_ObjName(Abc_ObjFanout0(pNode))));
    }
    fmt.Fprintf(pFile, "\n");

    fmt.Fprintf(pFile, "VAR  -- state variables\n");
    // inlining of macro Abc_NtkForEachLatch
    for  i = 0; (i < Vec_PtrSize((pNtk).vBoxes)) && (((pNode) = Abc_NtkBox(pNtk, i)), 1); i++  {           if !Abc_ObjIsLatch(pNode) {} else {
        fmt.Fprintf(pFile, "    %s : boolean;\n", cleanUNSAFE(Abc_ObjName(Abc_ObjFanout0(Abc_ObjFanout0(pNode))))); }
    } 
    fmt.Fprintf(pFile, "\n");

    // No outputs needed for NuSMV: 
    // TODO: Add sepcs by recognizing assume_.* and assert_.*
    //
    // Abc_NtkForEachPo( pNtk, pNode, i )
    //    fprintf( pFile, "OUTPUT(%s)\n", Abc_ObjName(Abc_ObjFanin0(pNode)) );
    
    // write internal nodes
    fmt.Fprintf(pFile, "DEFINE\n");
    pProgress = Extra_ProgressBarStart( os.Stdout, Abc_NtkObjNumMax(pNtk) );
    // inlining of macro Abc_NtkForEachNode
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pNode) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pNode) == nil || !Abc_ObjIsNode(pNode) {} else {
        Extra_ProgressBarUpdate( pProgress, i, nil );
        Io_WriteSmvOneNode( pFile, pNode );
    }
    }
    Extra_ProgressBarStop( pProgress );
    fmt.Fprintf(pFile, "\n");

    fmt.Fprintf(pFile, "ASSIGN\n");
    // inlining of macro Abc_NtkForEachLatch
    for  i = 0; (i < Vec_PtrSize((pNtk).vBoxes)) && (((pNode) = Abc_NtkBox(pNtk, i)), 1); i++  {           if !Abc_ObjIsLatch(pNode) {} else {
        var Reset int32  = int32((ABC_PTRUINT_T))Abc_ObjData( pNode );
        assert(Reset >= 1, "Reset >= 1");
        assert(Reset <= 3, "Reset <= 3");

        if Reset != 3 {
            fmt.Fprintf(pFile, "    init(%s) := %d;\n", cleanUNSAFE(Abc_ObjName(Abc_ObjFanout0(Abc_ObjFanout0(pNode)))), Reset - 1); 
        }
        fmt.Fprintf(pFile, "    next(%s) := ", cleanUNSAFE(Abc_ObjName(Abc_ObjFanout0(Abc_ObjFanout0(pNode))))); 
        fmt.Fprintf(pFile, "%s;\n", cleanUNSAFE(Abc_ObjName(Abc_ObjFanin0(Abc_ObjFanin0(pNode)))));
    }
    }
      
    fmt.Fprintf(pFile, "\n");
    // inlining of macro Abc_NtkForEachPo
    for  i = 0; (i < Abc_NtkPoNum(pNtk)) && (((pNode) = Abc_NtkPo(pNtk, i)), 1); i++  {
        char *n = cleanUNSAFE(Abc_ObjName(Abc_ObjFanin0(pNode)));
        // fprintf( pFile, "-- output %s;\n", n );
        if hasPrefix("assume_fair_", n) != 0 {
            fmt.Fprintf(pFile, "FAIRNESS %s;\n", n);
        
        } else if (hasPrefix("Assert_", n) != 0) || 
                (hasPrefix("assert_safety_", n) != 0) {
            fmt.Fprintf(pFile, "INVARSPEC %s;\n", n);
        
        } else if hasPrefix("assert_fair_", n) != 0 {
            fmt.Fprintf(pFile, "LTLSPEC G F %s;\n", n);
        }
    }

    return 1;
}

/**Function*************************************************************

  Synopsis    [Writes the network in SMV format.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_WriteSmvOneNode(pFile *os.File, pNode *Abc_Obj_t) int32 {
    var nFanins int32;

    assert(Abc_ObjIsNode(pNode), "Abc_ObjIsNode(pNode)");
    nFanins = Abc_ObjFaninNum(pNode);
    if nFanins == 0 {   // write the constant 1 node
        assert(Abc_NodeIsConst1(pNode), "Abc_NodeIsConst1(pNode)");
        fmt.Fprintf(pFile, "    %s", cleanUNSAFE(Abc_ObjName(Abc_ObjFanout0(pNode)) ));
        fmt.Fprintf(pFile, " := 1;\n");
    
    } else if nFanins == 1 {   // write the interver/buffer
        if Abc_NodeIsBuf(pNode) {
            fmt.Fprintf(pFile, "    %s := ", cleanUNSAFE(Abc_ObjName(Abc_ObjFanout0(pNode))));
            fmt.Fprintf(pFile, "%s;\n", cleanUNSAFE(Abc_ObjName(Abc_ObjFanin0(pNode))));
        
        } else {
            fmt.Fprintf(pFile, "    %s := !", cleanUNSAFE(Abc_ObjName(Abc_ObjFanout0(pNode))));
            fmt.Fprintf(pFile, "%s;\n", cleanUNSAFE(Abc_ObjName(Abc_ObjFanin0(pNode))));
        }
    
    } else {   // write the AND gate
        fmt.Fprintf(pFile, "    %s", cleanUNSAFE(Abc_ObjName(Abc_ObjFanout0(pNode))));
        fmt.Fprintf(pFile, " := %s & ", cleanUNSAFE(Abc_ObjName(Abc_ObjFanin0(pNode))));
        fmt.Fprintf(pFile, "%s;\n", cleanUNSAFE(Abc_ObjName(Abc_ObjFanin1(pNode))));
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Returns 1 if the names cannot be written into the bench file.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_WriteSmvCheckNames(pNtk *Abc_Ntk_t) int32 {
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

