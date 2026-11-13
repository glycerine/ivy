package bac

import "fmt"

import "os"

/**CFile****************************************************************

  FileName    [bacWriteBlif.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Hierarchical word-level netlist.]

  Synopsis    [Verilog parser.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - November 29, 2014.]

  Revision    [$Id: bacWriteBlif.c,v 1.00 2014/11/29 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "bac.h"
//#include "bacPrs.h"
//#include "map/mio/mio.h"
//#include "base/main/main.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Writing parser state into a file.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Psr_ManWriteBlifArray(pFile *os.File, p *Psr_Ntk_t, vFanins *Vec_Int_t) {
    var i, NameId int32;
    // inlining of macro Vec_IntForEachEntry
    for i = 0; i < Vec_IntSize(vFanins); i++ {
        NameId = Vec_IntEntry(vFanins, i)
        fmt.Fprintf(pFile, " %s", Psr_NtkStr(p, NameId));
    }
    fmt.Fprintf(pFile, "\n");
}
func Psr_ManWriteBlifLines(pFile *os.File, p *Psr_Ntk_t) {
    var vBox *Vec_Int_t; 
    var i, k, FormId, ActId int32;
    // inlining of macro Psr_NtkForEachBox
    for i = 0; i < Psr_NtkBoxNum(p); i++ {
        vBox = Psr_BoxSignals(p, i)
        if vBox == nil { break }
        var NtkId int32  = Psr_BoxNtk(p, i);
        assert(Psr_BoxIONum(p, i) > 0, "Psr_BoxIONum(p, i) > 0");
        assert(Vec_IntSize(vBox) % 2 == 0, "Vec_IntSize(vBox) % 2 == 0");
        if NtkId == -1 { // latch
        
            fmt.Fprintf(pFile, ".latch");
            fmt.Fprintf(pFile, " %s", Psr_NtkStr(p, Vec_IntEntry(vBox, 1)));
            fmt.Fprintf(pFile, " %s", Psr_NtkStr(p, Vec_IntEntry(vBox, 3)));
            fmt.Fprintf(pFile, " %c\n", '0' + Psr_BoxName(p, i));
        
        } else if Psr_BoxIsNode(p, i) { // node
        
            fmt.Fprintf(pFile, ".names");
            // inlining of macro Vec_IntForEachEntryDouble
            for k = 0; k+1 < Vec_IntSize(vBox); k += 2 {
                FormId = Vec_IntEntry(vBox, k)
                ActId = Vec_IntEntry(vBox, k+1)
                fmt.Fprintf(pFile, " %s", Psr_NtkStr(p, ActId));
            }
            fmt.Fprintf(pFile, "\n%s", Psr_NtkStr(p, NtkId));
        
        } else { // box
        
            fmt.Fprintf(pFile, ".subckt");
            fmt.Fprintf(pFile, " %s", Psr_NtkStr(p, NtkId));
            // inlining of macro Vec_IntForEachEntryDouble
            for k = 0; k+1 < Vec_IntSize(vBox); k += 2 {
                FormId = Vec_IntEntry(vBox, k)
                ActId = Vec_IntEntry(vBox, k+1)
                fmt.Fprintf(pFile, " %s=%s", Psr_NtkStr(p, FormId), Psr_NtkStr(p, ActId));
            }
            fmt.Fprintf(pFile, "\n");
        }
    }
}
func Psr_ManWriteBlifNtk(pFile *os.File, p *Psr_Ntk_t) {
    // write header
    fmt.Fprintf(pFile, ".model %s\n", Psr_NtkStr(p, p.iModuleName));
    if Vec_IntSize(&p.vInouts) {
    fmt.Fprintf(pFile, ".inouts"); }
    if Vec_IntSize(&p.vInouts) {
    Psr_ManWriteBlifArray( pFile, p, &p.vInouts ); }
    fmt.Fprintf(pFile, ".inputs");
    Psr_ManWriteBlifArray( pFile, p, &p.vInputs );
    fmt.Fprintf(pFile, ".outputs");
    Psr_ManWriteBlifArray( pFile, p, &p.vOutputs );
    // write objects
    Psr_ManWriteBlifLines( pFile, p );
    fmt.Fprintf(pFile, ".end\n\n");
}
func Psr_ManWriteBlif(pFileName string, vPrs *Vec_Ptr_t) {
    var pNtk *Psr_Ntk_t  = Psr_ManRoot(vPrs);
    var pFile *os.File  = os.OpenFile(pFileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666); var i int32;
    if pFile == nil {
        fmt.Printf("Cannot open output file \"%s\".\n", pFileName);
        return;
    }
    fmt.Fprintf(pFile, "# Design \"%s\" written by ABC on %s\n\n", Psr_NtkStr(pNtk, pNtk.iModuleName), Extra_TimeStamp());
    // inlining of macro Vec_PtrForEachEntry
    for i = 0; i < Vec_PtrSize(vPrs); i++ {
        pNtk, _ = Vec_PtrEntry(vPrs, i).(*Psr_Ntk_t)
        Psr_ManWriteBlifNtk( pFile, pNtk );
    }
    _ = (pFile).Close();
}



/**Function*************************************************************

  Synopsis    [Write elaborated design.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bac_ManWriteBlifGate(pFile *os.File, p *Bac_Ntk_t, pGate *Mio_Gate_t, vFanins *Vec_Int_t, iObj int32) {
    var iFanin, i int32;
    // inlining of macro Vec_IntForEachEntry
    for i = 0; i < Vec_IntSize(vFanins); i++ {
        iFanin = Vec_IntEntry(vFanins, i)
        fmt.Fprintf(pFile, " %s=%s", Mio_GateReadPinName(pGate, i), Bac_ObjNameStr(p, iFanin));
    }
    fmt.Fprintf(pFile, " %s=%s", Mio_GateReadOutName(pGate), Bac_ObjNameStr(p, iObj));
    fmt.Fprintf(pFile, "\n");
}
func Bac_ManWriteBlifArray(pFile *os.File, p *Bac_Ntk_t, vFanins *Vec_Int_t, iObj int32) {
    var iFanin, i int32;
    // inlining of macro Vec_IntForEachEntry
    for i = 0; i < Vec_IntSize(vFanins); i++ {
        iFanin = Vec_IntEntry(vFanins, i)
        fmt.Fprintf(pFile, " %s", Bac_ObjNameStr(p, iFanin));
    }
    if iObj >= 0 {
        fmt.Fprintf(pFile, " %s", Bac_ObjNameStr(p, iObj)); }
    fmt.Fprintf(pFile, "\n");
}
func Bac_ManWriteBlifArray2(pFile *os.File, p *Bac_Ntk_t, iObj int32) {
    var iTerm, i int32;
    var pModel *Bac_Ntk_t  = Bac_BoxNtk( p, iObj );
    // inlining of macro Bac_NtkForEachPi
    for i = 0; i < Bac_NtkPiNum(pModel); i++ {
        iTerm = Bac_NtkPi(pModel, i) 
        fmt.Fprintf(pFile, " %s=%s", Bac_ObjNameStr(pModel, iTerm), Bac_ObjNameStr(p, Bac_BoxBi(p, iObj, i)));
    }
    // inlining of macro Bac_NtkForEachPo
    for i = 0; i < Bac_NtkPoNum(pModel); i++ {
        iTerm = Bac_NtkPo(pModel, i) 
        fmt.Fprintf(pFile, " %s=%s", Bac_ObjNameStr(pModel, iTerm), Bac_ObjNameStr(p, Bac_BoxBo(p, iObj, i)));
    }
    fmt.Fprintf(pFile, "\n");
}
func Bac_ManWriteBlifLines(pFile *os.File, p *Bac_Ntk_t) {
    var i, k, iTerm int32;
    // inlining of macro Bac_NtkForEachBox
    for  i = 0; (i < Bac_NtkObjNum(p)); i++  { if !Bac_ObjIsBox(p, i) {} else {
        if Bac_ObjIsBoxUser(p, i) {
            fmt.Fprintf(pFile, ".subckt");
            fmt.Fprintf(pFile, " %s", Bac_NtkName(Bac_BoxNtk(p, i)));
            Bac_ManWriteBlifArray2( pFile, p, i );
        
        } else if Bac_ObjIsGate(p, i) {
            var pGateName string  = Abc_NamStr(p.pDesign.pMods, Bac_BoxNtkId(p, i));
            pLib, _ := Abc_FrameReadLibGen().(*Mio_Library_t)
            var pGate *Mio_Gate_t  = Mio_LibraryReadGateByName( pLib, pGateName, nil );
            fmt.Fprintf(pFile, ".gate %s", pGateName);
            // inlining of macro Bac_BoxForEachBi
            for  iTerm, k = i - 1, 0; iTerm >= 0 && Bac_ObjIsBi(p, iTerm); iTerm, k = iTerm-1, k+1  {
                fmt.Fprintf(pFile, " %s=%s", Mio_GateReadPinName(pGate, k), Bac_ObjNameStr(p, iTerm));
            }
            // inlining of macro Bac_BoxForEachBo
            for  iTerm, k = i + 1, 0; iTerm < Bac_NtkObjNum(p) && Bac_ObjIsBo(p, iTerm); iTerm, k = iTerm+1, k+1  {
                fmt.Fprintf(pFile, " %s=%s", Mio_GateReadOutName(pGate), Bac_ObjNameStr(p, iTerm));
            }
            fmt.Fprintf(pFile, "\n");
        
        } else {
            fmt.Fprintf(pFile, ".names");
            // inlining of macro Bac_BoxForEachBi
            for  iTerm, k = i - 1, 0; iTerm >= 0 && Bac_ObjIsBi(p, iTerm); iTerm, k = iTerm-1, k+1  {
                fmt.Fprintf(pFile, " %s", Bac_ObjNameStr(p, Bac_ObjFanin(p, iTerm)));
            }
            // inlining of macro Bac_BoxForEachBo
            for  iTerm, k = i + 1, 0; iTerm < Bac_NtkObjNum(p) && Bac_ObjIsBo(p, iTerm); iTerm, k = iTerm+1, k+1  {
                fmt.Fprintf(pFile, " %s", Bac_ObjNameStr(p, iTerm));
            }
            fmt.Fprintf(pFile, "\n%s", Ptr_TypeToSop(Bac_ObjType(p, i)));
        }
    }
    }
}
func Bac_ManWriteBlifNtk(pFile *os.File, p *Bac_Ntk_t) {
    assert(Vec_IntSize(&p.vFanin) == Bac_NtkObjNum(p), "Vec_IntSize(&p.vFanin) == Bac_NtkObjNum(p)");
    // write header
    fmt.Fprintf(pFile, ".model %s\n", Bac_NtkName(p));
    fmt.Fprintf(pFile, ".inputs");
    Bac_ManWriteBlifArray( pFile, p, &p.vInputs, -1 );
    fmt.Fprintf(pFile, ".outputs");
    Bac_ManWriteBlifArray( pFile, p, &p.vOutputs, -1 );
    // write objects
    Bac_ManWriteBlifLines( pFile, p );
    fmt.Fprintf(pFile, ".end\n\n");
}
func Bac_ManWriteBlif(pFileName string, p *Bac_Man_t) {
    var pFile *os.File;
    var pNtk *Bac_Ntk_t; 
    var i int32;
    // check the library
    if p.pMioLib && p.pMioLib != Abc_FrameReadLibGen() {
        fmt.Printf("Genlib library used in the mapped design is not longer a current library.\n");
        return;
    }
    pFile = os.OpenFile(pFileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    if pFile == nil {
        fmt.Printf("Cannot open output file \"%s\".\n", pFileName);
        return;
    }
    fmt.Fprintf(pFile, "# Design \"%s\" written via CBA package in ABC on %s\n\n", Bac_ManName(p), Extra_TimeStamp());
    Bac_ManAssignInternWordNames( p );
    // inlining of macro Bac_ManForEachNtk
    for i = 1; i <= Bac_ManNtkNum(p); i++ {
        pNtk = Bac_ManNtk(p, i) 
        Bac_ManWriteBlifNtk( pFile, pNtk );
    }
    _ = (pFile).Close();
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

