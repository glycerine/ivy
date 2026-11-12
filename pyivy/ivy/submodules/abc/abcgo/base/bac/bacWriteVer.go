package bac

import "fmt"

import "os"

/**CFile****************************************************************

  FileName    [bacWriteVer.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Hierarchical word-level netlist.]

  Synopsis    [Verilog writer.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - November 29, 2014.]

  Revision    [$Id: bacWriteVer.c,v 1.00 2014/11/29 00:00:00 alanmi Exp $]

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
func Psr_ManWriteVerilogConcat(pFile *os.File, p *Psr_Ntk_t, Con int32) {
    extern void Psr_ManWriteVerilogArray( FILE * pFile, Psr_Ntk_t * p, Vec_Int_t * vSigs, int Start, int Stop, int fOdd );
    var vSigs *Vec_Int_t  = Psr_CatSignals(p, Con);
    fmt.Fprintf(pFile, "{");
    Psr_ManWriteVerilogArray( pFile, p, vSigs, 0, Vec_IntSize(vSigs), 0 );
    fmt.Fprintf(pFile, "}");
}
func Psr_ManWriteVerilogSignal(pFile *os.File, p *Psr_Ntk_t, Sig int32) {
    var Value int32  = Abc_Lit2Var2( Sig );
    var Type Psr_ManType_t  = (Psr_ManType_t)Abc_Lit2Att2( Sig );
    if Type == BAC_PRS_NAME || Type == BAC_PRS_CONST {
        fmt.Fprintf(pFile, "%s", Psr_NtkStr(p, Value));
    } else if Type == BAC_PRS_SLICE {
        fmt.Fprintf(pFile, "%s%s", Psr_NtkStr(p, Psr_SliceName(p, Value)), Psr_NtkStr(p, Psr_SliceRange(p, Value)));
    } else if Type == BAC_PRS_CONCAT {
        Psr_ManWriteVerilogConcat( pFile, p, Value );
    } else { assert(false, "0"); }
}
func Psr_ManWriteVerilogArray(pFile *os.File, p *Psr_Ntk_t, vSigs *Vec_Int_t, Start int32, Stop int32, fOdd int32) {
    var i, Sig int32;
    assert(Vec_IntSize(vSigs) > 0, "Vec_IntSize(vSigs) > 0");
    // inlining of macro Vec_IntForEachEntryStartStop
    for  i = Start; (i < Stop) && (((Sig) = Vec_IntEntry(vSigs, i)), 1); i++  {
        if (fOdd != 0) && !(i & 1) {
            continue; }
        Psr_ManWriteVerilogSignal( pFile, p, Sig );
        fmt.Fprintf(pFile, "%s", i == Stop - 1 ? "" : ", ");
    }
}
func Psr_ManWriteVerilogArray2(pFile *os.File, p *Psr_Ntk_t, vSigs *Vec_Int_t) {
    var i, FormId, ActSig int32;
    assert(Vec_IntSize(vSigs) % 2 == 0, "Vec_IntSize(vSigs) % 2 == 0");
    // inlining of macro Vec_IntForEachEntryDouble
    for  i = 0; (i+1 < Vec_IntSize(vSigs)) && (((FormId) = Vec_IntEntry(vSigs, i)), 1) && (((ActSig) = Vec_IntEntry(vSigs, i+1)), 1); i += 2  {
        fmt.Fprintf(pFile, ".");
        fmt.Fprintf(pFile, "%s", Psr_NtkStr(p, FormId));
        fmt.Fprintf(pFile, "(");
        Psr_ManWriteVerilogSignal( pFile, p, ActSig );
        fmt.Fprintf(pFile, ")%s", (i == Vec_IntSize(vSigs) - 2) ? "" : ", ");
    }
}
func Psr_ManWriteVerilogMux(pFile *os.File, p *Psr_Ntk_t, vSigs *Vec_Int_t) {
    var i, FormId, ActSig int32;
    var pStrs [4]string = [4]string{ " = ", " ? ", " : ", ";\n" };
    assert(Vec_IntSize(vSigs) == 8, "Vec_IntSize(vSigs) == 8");
    fmt.Fprintf(pFile, "  assign ");
    Psr_ManWriteVerilogSignal( pFile, p, Vec_IntEntryLast(vSigs) );
    fmt.Fprintf(pFile, "%s", pStrs[0]);
    // inlining of macro Vec_IntForEachEntryDouble
    for  i = 0; (i+1 < Vec_IntSize(vSigs)) && (((FormId) = Vec_IntEntry(vSigs, i)), 1) && (((ActSig) = Vec_IntEntry(vSigs, i+1)), 1); i += 2  {
        Psr_ManWriteVerilogSignal( pFile, p, ActSig );
        fmt.Fprintf(pFile, "%s", pStrs[1+i/2]);
        if i == 4 {
            break; }
    }
}
func Psr_ManWriteVerilogBoxes(pFile *os.File, p *Psr_Ntk_t) {
    var vBox *Vec_Int_t; var i int32;
    // inlining of macro Psr_NtkForEachBox
    for  i = 0; i < Psr_NtkBoxNum(p) && ((vBox) = Psr_BoxSignals(p, i)); i++  {
        var NtkId Bac_ObjType_t  = (Bac_ObjType_t)Psr_BoxNtk(p, i);
        if NtkId == BAC_BOX_MUX {
            Psr_ManWriteVerilogMux( pFile, p, vBox );
        } else if Psr_BoxIsNode(p, i) { // node   ------- check order of fanins
        
            fmt.Fprintf(pFile, "  %s (", Ptr_TypeToName(NtkId));
            Psr_ManWriteVerilogSignal( pFile, p, Vec_IntEntryLast(vBox) );
            if Psr_BoxIONum(p, i) > 1 {
                fmt.Fprintf(pFile, ", "); }                
            Psr_ManWriteVerilogArray( pFile, p, vBox, 0, Vec_IntSize(vBox)-2, 1 );
            fmt.Fprintf(pFile, ");\n");
        
        } else { // box
        
            //char * s = Psr_NtkStr(p, Vec_IntEntry(vBox, 0));
            fmt.Fprintf(pFile, "  %s %s (", Psr_NtkStr(p, NtkId), Psr_BoxName(p, i) ? Psr_NtkStr(p, Psr_BoxName(p, i)) : "");
            Psr_ManWriteVerilogArray2( pFile, p, vBox );
            fmt.Fprintf(pFile, ");\n");
        }
    }
}
func Psr_ManWriteVerilogIos(pFile *os.File, p *Psr_Ntk_t, SigType int32) {
    var NameId, RangeId, i int32;
    var pSigNames [4]string = [4]string{ "inout", "input", "output", "wire" }; 
    var vSigs [4]*Vec_Int_t = [4]*Vec_Int_t{ &p.vInouts,  &p.vInputs,  &p.vOutputs,  &p.vWires };
    var vSigsR [4]*Vec_Int_t = [4]*Vec_Int_t{ &p.vInoutsR, &p.vInputsR, &p.vOutputsR, &p.vWiresR };
    if SigType == 3 {
        fmt.Fprintf(pFile, "\n"); }
    // inlining of macro Vec_IntForEachEntryTwo
    for  i = 0; (i < Vec_IntSize(vSigs[SigType])) && (((NameId) = Vec_IntEntry(vSigs[SigType], i)), 1) && (((RangeId) = Vec_IntEntry(vSigsR[SigType], i)), 1); i++  {
        fmt.Fprintf(pFile, "  %s %s%s;\n", pSigNames[SigType], RangeId ? Psr_NtkStr(p, RangeId) : "", Psr_NtkStr(p, NameId));
    }
}
func Psr_ManWriteVerilogIoOrder(pFile *os.File, p *Psr_Ntk_t, vOrder *Vec_Int_t) {
    var i, NameId int32;
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vOrder)) && (((NameId) = Vec_IntEntry(vOrder, i)), 1); i++  {
        fmt.Fprintf(pFile, "%s%s", Psr_NtkStr(p, NameId), i == Vec_IntSize(vOrder) - 1 ? "" : ", ");
    }
}
func Psr_ManWriteVerilogNtk(pFile *os.File, p *Psr_Ntk_t) {
    var s int32;
    // write header
    fmt.Fprintf(pFile, "module %s (\n    ", Psr_NtkStr(p, p.iModuleName));
    Psr_ManWriteVerilogIoOrder( pFile, p, &p.vOrder );
    fmt.Fprintf(pFile, "\n  );\n");
    // write declarations
    for  s = 0; s < 4; s++  {
        Psr_ManWriteVerilogIos( pFile, p, s );
    }
    fmt.Fprintf(pFile, "\n");
    // write objects
    Psr_ManWriteVerilogBoxes( pFile, p );
    fmt.Fprintf(pFile, "endmodule\n\n");
}
func Psr_ManWriteVerilog(pFileName string, vPrs *Vec_Ptr_t) {
    var pNtk *Psr_Ntk_t  = Psr_ManRoot(vPrs); var i int32;
    var pFile *os.File  = os.OpenFile(pFileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    if pFile == nil {
        fmt.Printf("Cannot open output file \"%s\".\n", pFileName);
        return;
    }
    fmt.Fprintf(pFile, "// Design \"%s\" written by ABC on %s\n\n", Psr_NtkStr(pNtk, pNtk.iModuleName), Extra_TimeStamp());
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vPrs)) && (((pNtk) = (Psr_Ntk_t *)Vec_PtrEntry(vPrs, i)), 1); i++  {
        Psr_ManWriteVerilogNtk( pFile, pNtk );
    }
    _ = (pFile).Close();
}



/**Function*************************************************************

  Synopsis    [Writing word-level Verilog.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
// compute range of a name (different from range of a multi-bit wire)
func Bac_ObjGetRange(p *Bac_Ntk_t, iObj int32) int32 {
    var i int32;
    var NameId int32 = Bac_ObjName(p, iObj);
    assert(Bac_ObjIsCi(p, iObj), "Bac_ObjIsCi(p, iObj)");
//    if ( Bac_NameType(NameId) == BAC_NAME_INDEX )
//        NameId = Bac_ObjName(p, iObj - Abc_Lit2Var2(NameId));
    assert(Bac_NameType(NameId) == BAC_NAME_WORD || Bac_NameType(NameId) == BAC_NAME_INFO, "Bac_NameType(NameId) == BAC_NAME_WORD || Bac_NameType(NameId) == BAC_NAME_INFO");
    for  i = iObj + 1; i < Bac_NtkObjNum(p); i++  {
        if !Bac_ObjIsCi(p, i) || Bac_ObjNameType(p, i) != BAC_NAME_INDEX {
            break; }
    }
    return i - iObj;
}

func Bac_ManWriteVar(p *Bac_Ntk_t, RealName int32) {
    Vec_StrPrintStr( p.pDesign.vOut, Bac_NtkStr(p, RealName) );
}
func Bac_ManWriteRange(p *Bac_Ntk_t, Beg int32, End int32) {
    var vStr *Vec_Str_t  = p.pDesign.vOut;
    Vec_StrPrintStr( vStr, "[" );
    if End >= 0 {
        Vec_StrPrintNum( vStr, End );
        Vec_StrPrintStr( vStr, ":" );
    }
    Vec_StrPrintNum( vStr, Beg );
    Vec_StrPrintStr( vStr, "]" );
}
func Bac_ManWriteConstBit(p *Bac_Ntk_t, iObj int32, fHead int32) {
    var vStr *Vec_Str_t  = p.pDesign.vOut;
    var Const int32  = Bac_ObjGetConst(p, iObj);
    assert((Const != 0), "Const");
    if fHead != 0 {
        Vec_StrPrintStr( vStr, "1\'b" ); }
    if Const == BAC_BOX_CF {
        Vec_StrPush( vStr, '0' );
    } else if Const == BAC_BOX_CT {
        Vec_StrPush( vStr, '1' );
    } else if Const == BAC_BOX_CX {
        Vec_StrPush( vStr, 'x' );
    } else if Const == BAC_BOX_CZ {
        Vec_StrPush( vStr, 'z' );
    } else { assert(false, "0"); }
}
func Bac_ManFindRealNameId(p *Bac_Ntk_t, iObj int32) int32 {
    var NameId int32  = Bac_ObjName(p, iObj);
    assert(Bac_ObjIsCi(p, iObj), "Bac_ObjIsCi(p, iObj)");
    if Bac_NameType(NameId) == BAC_NAME_INDEX {
        NameId = Bac_ObjName(p, iObj - Abc_Lit2Var2(NameId)); }
    if Bac_NameType(NameId) == BAC_NAME_INFO {
        return Bac_NtkInfoName(p, Abc_Lit2Var2(NameId)); }
    assert(Bac_NameType(NameId) == BAC_NAME_BIN || Bac_NameType(NameId) == BAC_NAME_WORD, "Bac_NameType(NameId) == BAC_NAME_BIN || Bac_NameType(NameId) == BAC_NAME_WORD");
    return Abc_Lit2Var2(NameId);
}
func Bac_ManFindRealIndex(p *Bac_Ntk_t, iObj int32) int32 {
    var iBit int32;
    var NameId int32 = Bac_ObjName(p, iObj);
    assert(Bac_ObjIsCi(p, iObj), "Bac_ObjIsCi(p, iObj)");
    assert(Bac_NameType(NameId) != BAC_NAME_BIN, "Bac_NameType(NameId) != BAC_NAME_BIN");
    if Bac_NameType(NameId) == BAC_NAME_INDEX {
        NameId = Bac_ObjName(p, iObj - (iBit = Abc_Lit2Var2(NameId))); }
    if Bac_NameType(NameId) == BAC_NAME_INFO {
        return Bac_NtkInfoIndex(p, Abc_Lit2Var2(NameId), iBit); }
    assert(Bac_NameType(NameId) == BAC_NAME_WORD, "Bac_NameType(NameId) == BAC_NAME_WORD");
    return iBit;
}
func Bac_ManWriteSig(p *Bac_Ntk_t, iObj int32) {
    if Bac_ObjIsCo(p, iObj) {
        iObj = Bac_ObjFanin(p, iObj); }
    assert(Bac_ObjIsCi(p, iObj), "Bac_ObjIsCi(p, iObj)");
    if Bac_ObjGetConst(p, iObj) {
        Bac_ManWriteConstBit( p, iObj, 1 );
    } else {
        var NameId int32  = Bac_ObjName(p, iObj);
        if Bac_NameType(NameId) == BAC_NAME_BIN {
            Bac_ManWriteVar( p, Abc_Lit2Var2(NameId) );
        } else {
            Bac_ManWriteVar( p, Bac_ManFindRealNameId(p, iObj) );
            Bac_ManWriteRange( p, Bac_ManFindRealIndex(p, iObj), -1 );
        }
    }
}
func Bac_ManWriteConcat(p *Bac_Ntk_t, iStart int32, nObjs int32) {
    var vStr *Vec_Str_t  = p.pDesign.vOut; 
    assert(nObjs >= 1, "nObjs >= 1");
    if nObjs == 1 {
        Bac_ManWriteSig( p, iStart );
        return;
    }
    Vec_StrPrintStr( vStr, "{" );
    if Bac_ObjIsBo(p, iStart) { // box output
    
        var i int32;
        for  i = iStart + nObjs - 1; i >= iStart; i--  {
            if Bac_ObjNameType(p, i) == BAC_NAME_INDEX {
                continue; }
            if Vec_StrEntryLast(vStr) != '{' {
                Vec_StrPrintStr( vStr, ", " ); }
            Bac_ManWriteVar( p, Bac_ManFindRealNameId(p, i) );
        }
    
    } else if Bac_ObjIsBi(p, iStart) { // box input
    
        var e, b, k, NameId int32;
        for  e = iStart - nObjs + 1; e <= iStart;  {
            if Vec_StrEntryLast(vStr) != '{' {
                Vec_StrPrintStr( vStr, ", " ); }
            // write constant
            if Bac_ObjGetConst(p, Bac_ObjFanin(p, e)) {
                var fBinary int32  = Bac_ObjIsConstBin(p, Bac_ObjFanin(p, e)-1);                
                for  b = e + 1; b <= iStart; b++  {
                    if !Bac_ObjGetConst(p, Bac_ObjFanin(p, b)) {
                        break; }
                    if !Bac_ObjIsConstBin(p, Bac_ObjFanin(p, b)-1) {
                        fBinary = 0; }
                }
                Vec_StrPrintNum( vStr, b - e );
                if (fBinary != 0) && b - e > 8 { // write hex if more than 8 bits
                
                    var Digit int32;
                    var nBits int32;
                    if (b - e) & 3 {
                        nBits = (b - e) & 3;
                    } else {
                        nBits = 4;
                    }
                    Vec_StrPrintStr( vStr, "\'h" );
                    for  k = e; k < b; k++  {
                        Digit = 2*Digit + Bac_ObjGetConst(p, Bac_ObjFanin(p, k)) - BAC_BOX_CF;
                        assert(Digit < 16, "Digit < 16");
                        if nBits--; nBits == 0 {
                            Vec_StrPush( vStr, byte((Digit < 10 ? '0' + Digit : 'a' + Digit - 10)) );
                            nBits = 4;
                            Digit = 0;
                        }
                    }
                    assert(nBits == 4, "nBits == 4");
                    assert(Digit == 0, "Digit == 0");
                
                } else {
                    Vec_StrPrintStr( vStr, "\'b" );
                    for  k = e; k < b; k++  {
                        Bac_ManWriteConstBit( p, Bac_ObjFanin(p, k), 0 );
                    }
                }
                e = b;
                continue;
            }
            // try replication
            for  b = e + 1; b <= iStart; b++  {
                if Bac_ObjFanin(p, b) != Bac_ObjFanin(p, e) {
                    break; }
            }
            if b > e + 2 { // more than two
            
                Vec_StrPrintNum( vStr, b - e );
                Vec_StrPrintStr( vStr, "{" );
                Bac_ManWriteSig( p, e );
                Vec_StrPrintStr( vStr, "}" );
                e = b;
                continue;
            }
            NameId = Bac_ObjName(p, Bac_ObjFanin(p, e));
            if Bac_NameType(NameId) == BAC_NAME_BIN {
                Bac_ManWriteVar( p, Abc_Lit2Var2(NameId) );
                e++;
                continue;
            }
            // find end of the slice
            for  b = e + 1; b <= iStart; b++  {
                if Bac_ObjFanin(p, e) - Bac_ObjFanin(p, b) != b - e {
                    break; }
            }
            // write signal name
            Bac_ManWriteVar( p, Bac_ManFindRealNameId(p, Bac_ObjFanin(p, e)) );
            if b == e + 1 { // literal
                Bac_ManWriteRange( p, Bac_ManFindRealIndex(p, Bac_ObjFanin(p, e)), -1 );
            } else { // slice or complete variable
            
                // consider first variable of the slice
                var f int32  = Bac_ObjFanin( p, b-1 );
                assert(Bac_ObjNameType(p, f) != BAC_NAME_BIN, "Bac_ObjNameType(p, f) != BAC_NAME_BIN");
                if Bac_ObjNameType(p, f) == BAC_NAME_INDEX || Bac_ObjGetRange(p, f) != b - e { // slice
                    Bac_ManWriteRange( p, Bac_ManFindRealIndex(p, f), Bac_ManFindRealIndex(p, Bac_ObjFanin(p, e)) ); }
                // else this is complete variable
            }
            e = b;
        }
    
    } else { assert(false, "0"); }
    Vec_StrPrintStr( vStr, "}" );
}
func Bac_ManWriteGate(p *Bac_Ntk_t, iObj int32) {
    var vStr *Vec_Str_t  = p.pDesign.vOut; var iTerm, k int32;
    var pGateName string  = Abc_NamStr(p.pDesign.pMods, Bac_BoxNtkId(p, iObj));
    var pLib *Mio_Library_t  = (Mio_Library_t *)Abc_FrameReadLibGen();
    var pGate *Mio_Gate_t  = Mio_LibraryReadGateByName( pLib, pGateName, nil );
    Vec_StrPrintStr( vStr, "  " );
    Vec_StrPrintStr( vStr, pGateName );
    Vec_StrPrintStr( vStr, " " );
    Vec_StrPrintStr( vStr, Bac_ObjName(p, iObj) ? Bac_ObjNameStr(p, iObj) : "" );
    Vec_StrPrintStr( vStr, " (" );
    // inlining of macro Bac_BoxForEachBi
    for  iTerm, k = iObj - 1, 0; iTerm >= 0 && Bac_ObjIsBi(p, iTerm); iTerm, k = iTerm-1, k+1  {
        Vec_StrPrintStr( vStr, k ? ", ." : "." );
        Vec_StrPrintStr( vStr, Mio_GateReadPinName(pGate, k) );
        Vec_StrPrintStr( vStr, "(" );
        Bac_ManWriteSig( p, iTerm );
        Vec_StrPrintStr( vStr, ")" );
    }
    // inlining of macro Bac_BoxForEachBo
    for  iTerm, k = iObj + 1, 0; iTerm < Bac_NtkObjNum(p) && Bac_ObjIsBo(p, iTerm); iTerm, k = iTerm+1, k+1  {
        Vec_StrPrintStr( vStr, Bac_BoxBiNum(p, iObj) ? ", ." : "." );
        Vec_StrPrintStr( vStr, Mio_GateReadOutName(pGate) );
        Vec_StrPrintStr( vStr, "(" );
        Bac_ManWriteSig( p, iTerm );
        Vec_StrPrintStr( vStr, ")" );
    }
    Vec_StrPrintStr( vStr, ");\n" );
}
func Bac_ManWriteAssign(p *Bac_Ntk_t, iObj int32) {
    var vStr *Vec_Str_t  = p.pDesign.vOut;
    var Type Bac_ObjType_t  = Bac_ObjType(p, iObj);
    var nInputs int32  = Bac_BoxBiNum(p, iObj);
    var nOutputs int32  = Bac_BoxBoNum(p, iObj);
    assert(nOutputs == 1, "nOutputs == 1");
    Vec_StrPrintStr( vStr, "  assign " );
    Bac_ManWriteSig( p, iObj + 1 );
    Vec_StrPrintStr( vStr, " = " );
    if nInputs == 0 {
        if Type == BAC_BOX_CF {
            Vec_StrPrintStr( vStr, "1\'b0" );
        } else if Type == BAC_BOX_CT {
            Vec_StrPrintStr( vStr, "1\'b1" );
        } else if Type == BAC_BOX_CX {
            Vec_StrPrintStr( vStr, "1\'bx" );
        } else if Type == BAC_BOX_CZ {
            Vec_StrPrintStr( vStr, "1\'bz" );
        } else { assert(false, "0"); }       
    
    } else if nInputs == 1 {
        if Type == BAC_BOX_INV {
            Vec_StrPrintStr( vStr, "~" );
        } else { assert(Type == BAC_BOX_BUF, "Type == BAC_BOX_BUF"); }
        Bac_ManWriteSig( p, iObj - 1 );
    
    } else if nInputs == 2 {
        if Type == BAC_BOX_NAND || Type == BAC_BOX_NOR || Type == BAC_BOX_XNOR || Type == BAC_BOX_SHARPL {
            Vec_StrPrintStr( vStr, "~" ); }
        Bac_ManWriteSig( p, iObj - 1 );
        if Type == BAC_BOX_AND || Type == BAC_BOX_SHARPL {
            Vec_StrPrintStr( vStr, " & " );
        } else if Type == BAC_BOX_SHARP || Type == BAC_BOX_NOR {
            Vec_StrPrintStr( vStr, " & ~" );
        } else if Type == BAC_BOX_OR {
            Vec_StrPrintStr( vStr, " | " );
        } else if Type == BAC_BOX_NAND {
            Vec_StrPrintStr( vStr, " | ~" );
        } else if Type == BAC_BOX_XOR || Type == BAC_BOX_XNOR {
            Vec_StrPrintStr( vStr, " ^ " );
        } else { assert(false, "0"); }
        Bac_ManWriteSig( p, iObj - 2 );
    }
    Vec_StrPrintStr( vStr, ";\n" );
}
func Bac_ManWriteVerilogBoxes(p *Bac_Ntk_t, fUseAssign int32) {
    var vStr *Vec_Str_t  = p.pDesign.vOut;
    var iObj, k, i, o, StartPos int32;
    // inlining of macro Bac_NtkForEachBox
    for  iObj = 0; (iObj < Bac_NtkObjNum(p)); iObj++  { if !Bac_ObjIsBox(p, iObj) {} else { // .subckt/.gate/box (formal/actual binding) 
    
        // skip constants
        if Bac_ObjIsConst(p, iObj) {
            continue; }
        // write mapped
        if Bac_ObjIsGate(p, iObj) {
            Bac_ManWriteGate( p, iObj );
            continue;
        }
        // write primitives as assign-statements
        if !Bac_ObjIsBoxUser(p, iObj) && (fUseAssign != 0) {
            Bac_ManWriteAssign( p, iObj );
            continue;
        }
        // write header
        StartPos = Vec_StrSize(vStr);
        if Bac_ObjIsBoxUser(p, iObj) {
            var Value, Beg, End, Range int32;
            var pModel *Bac_Ntk_t  = Bac_BoxNtk( p, iObj );
            Vec_StrPrintStr( vStr, "  " );
            Vec_StrPrintStr( vStr, Bac_NtkName(pModel) );
            Vec_StrPrintStr( vStr, " " );
            Vec_StrPrintStr( vStr, Bac_ObjName(p, iObj) ? Bac_ObjNameStr(p, iObj) : ""  );
            Vec_StrPrintStr( vStr, " (" );
            // write arguments
            i = o = 0;
            assert(Bac_NtkInfoNum(pModel), "Bac_NtkInfoNum(pModel)");
            // inlining of macro Vec_IntForEachEntryTriple
            for  k = 0; (k+2 < Vec_IntSize(&pModel.vInfo)) && (((Value) = Vec_IntEntry(&pModel.vInfo, k)), 1) && (((Beg) = Vec_IntEntry(&pModel.vInfo, k+1)), 1) && (((End) = Vec_IntEntry(&pModel.vInfo, k+2)), 1); k += 3  {
                var NameId int32  = Abc_Lit2Var2( Value );
                var Type int32  = Abc_Lit2Att2( Value );
                Vec_StrPrintStr( vStr, k ? ", " : "" );
                if Vec_StrSize(vStr) > StartPos + 70 {
                    StartPos = Vec_StrSize(vStr);
                    Vec_StrPrintStr( vStr, "\n    " );
                }
                Vec_StrPrintStr( vStr, "." );
                Vec_StrPrintStr( vStr, Bac_NtkStr(p, NameId) );
                Vec_StrPrintStr( vStr, "(" );
                Range = Bac_InfoRange( Beg, End );
                assert(Range > 0, "Range > 0");
                if Type == 1 {
                    Bac_ManWriteConcat( p, Bac_BoxBi(p, iObj, i), Range );
i += Range;
                } else if Type == 2 {
                    Bac_ManWriteConcat( p, Bac_BoxBo(p, iObj, o), Range );
o += Range;
                } else { assert(false, "0"); }
                Vec_StrPrintStr( vStr, ")" );
            }
            assert(i == Bac_BoxBiNum(p, iObj), "i == Bac_BoxBiNum(p, iObj)");
            assert(o == Bac_BoxBoNum(p, iObj), "o == Bac_BoxBoNum(p, iObj)");
        
        } else {
            var iTerm, k, Range, iSig int32;
            var vBits *Vec_Int_t  = Bac_BoxCollectRanges( p, iObj );
            var pName string  = Bac_NtkGenerateName( p, Bac_ObjType(p, iObj), vBits );
            var pSymbs string  = Bac_ManPrimSymb( p.pDesign, Bac_ObjType(p, iObj) );
            Vec_StrPrintStr( vStr, "  " );
            Vec_StrPrintStr( vStr, pName );
            Vec_StrPrintStr( vStr, " " );
            Vec_StrPrintStr( vStr, Bac_ObjName(p, iObj) ? Bac_ObjNameStr(p, iObj) : ""  );
            Vec_StrPrintStr( vStr, " (" );
            // write inputs
            // inlining of macro Bac_BoxForEachBiMain
            for  iTerm, k = iObj - 1, 0; iTerm >= 0 && Bac_ObjIsBi(p, iTerm); iTerm, k = iTerm-1, k+1  {                if Bac_ObjBit(p, iTerm) {} else {
                Range = Vec_IntEntry( vBits, iSig );
                Vec_StrPrintStr( vStr, iSig ? ", " : "" );
                if Vec_StrSize(vStr) > StartPos + 70 {
                    StartPos = Vec_StrSize(vStr);
                    Vec_StrPrintStr( vStr, "\n    " );
                }
                Vec_StrPrintStr( vStr, "." );
                Vec_StrPush( vStr, pSymbs[iSig] );
                Vec_StrPrintStr( vStr, "(" );
                Bac_ManWriteConcat( p, iTerm, Range );
                Vec_StrPrintStr( vStr, ")" );
                iSig++;
            }
            }
            // inlining of macro Bac_BoxForEachBoMain
            for  iTerm, k = iObj + 1, 0; iTerm < Bac_NtkObjNum(p) && Bac_ObjIsBo(p, iTerm); iTerm, k = iTerm+1, k+1  {  if Bac_ObjBit(p, iTerm) {} else {
                Range = Vec_IntEntry( vBits, iSig );
                Vec_StrPrintStr( vStr, iSig ? ", " : "" );
                if Vec_StrSize(vStr) > StartPos + 70 {
                    StartPos = Vec_StrSize(vStr);
                    Vec_StrPrintStr( vStr, "\n    " );
                }
                Vec_StrPrintStr( vStr, "." );
                Vec_StrPush( vStr, pSymbs[iSig] );
                Vec_StrPrintStr( vStr, "(" );
                Bac_ManWriteConcat( p, iTerm, Range );
                Vec_StrPrintStr( vStr, ")" );
                iSig++;
            }
            }
            assert(iSig == Vec_IntSize(vBits), "iSig == Vec_IntSize(vBits)");
        }
        Vec_StrPrintStr( vStr, ");\n" );
    }
    }
}
func Bac_ManWriteVerilogNtk(p *Bac_Ntk_t, fUseAssign int32) {
    var pKeyword [4]string = [4]string{ "wire ", "input ", "output ", "inout " };
    var vStr *Vec_Str_t  = p.pDesign.vOut;
    var k, iObj, iTerm, Value, Beg, End, Length, fHaveWires, StartPos int32;
//    assert( Bac_NtkInfoNum(p) );
    assert(Vec_IntSize(&p.vFanin) == Bac_NtkObjNum(p), "Vec_IntSize(&p.vFanin) == Bac_NtkObjNum(p)");
//    Bac_NtkPrint( p );
    // write header
    Vec_StrPrintStr( vStr, "module " );
    Vec_StrPrintStr( vStr, Bac_NtkName(p) );
    Vec_StrPrintStr( vStr, " (\n    " );
    StartPos = Vec_StrSize(vStr);
    // inlining of macro Vec_IntForEachEntryTriple
    for  k = 0; (k+2 < Vec_IntSize(&p.vInfo)) && (((Value) = Vec_IntEntry(&p.vInfo, k)), 1) && (((Beg) = Vec_IntEntry(&p.vInfo, k+1)), 1) && (((End) = Vec_IntEntry(&p.vInfo, k+2)), 1); k += 3  {
    if Abc_Lit2Att2(Value) != 0 {
        Vec_StrPrintStr( vStr, k ? ", " : "" );
        if Vec_StrSize(vStr) > StartPos + 70 {
            StartPos = Vec_StrSize(vStr);
            Vec_StrPrintStr( vStr, "\n    " );
        }
        Bac_ManWriteVar( p, Abc_Lit2Var2(Value) );
    }
    }
    Vec_StrPrintStr( vStr, "\n  );\n" );
    // write inputs/outputs
    // inlining of macro Vec_IntForEachEntryTriple
    for  k = 0; (k+2 < Vec_IntSize(&p.vInfo)) && (((Value) = Vec_IntEntry(&p.vInfo, k)), 1) && (((Beg) = Vec_IntEntry(&p.vInfo, k+1)), 1) && (((End) = Vec_IntEntry(&p.vInfo, k+2)), 1); k += 3  {
    if Abc_Lit2Att2(Value) != 0 {
        Vec_StrPrintStr( vStr, "  " );
        Vec_StrPrintStr( vStr, pKeyword[Abc_Lit2Att2(Value)] );
        if Beg >= 0 {
            Bac_ManWriteRange( p, Beg, End ); }
        Bac_ManWriteVar( p, Abc_Lit2Var2(Value) );
        Vec_StrPrintStr( vStr, ";\n" );
    }
    }
    Vec_StrPrintStr( vStr, "\n" );
    // write word-level wires
    // inlining of macro Bac_NtkForEachBox
    for  iObj = 0; (iObj < Bac_NtkObjNum(p)); iObj++  { if !Bac_ObjIsBox(p, iObj) {} else
    if !Bac_ObjIsConst(p, iObj) {
        // inlining of macro Bac_BoxForEachBo
        for  iTerm, k = iObj + 1, 0; iTerm < Bac_NtkObjNum(p) && Bac_ObjIsBo(p, iTerm); iTerm, k = iTerm+1, k+1  {
        if Bac_ObjNameType(p, iTerm) == BAC_NAME_WORD || Bac_ObjNameType(p, iTerm) == BAC_NAME_INFO {
            Vec_StrPrintStr( vStr, "  wire " );
            Bac_ManWriteRange( p, Bac_ManFindRealIndex(p, iTerm), Bac_ManFindRealIndex(p, iTerm + Bac_ObjGetRange(p, iTerm) - 1) );
            Bac_ManWriteVar( p, Bac_ManFindRealNameId(p, iTerm) );
            Vec_StrPrintStr( vStr, ";\n" );
        }
        } }
    }
    // check if there are any wires left
    fHaveWires = 0;
    // inlining of macro Bac_NtkForEachBox
    for  iObj = 0; (iObj < Bac_NtkObjNum(p)); iObj++  { if !Bac_ObjIsBox(p, iObj) {} else
    if !Bac_ObjIsConst(p, iObj) {
        // inlining of macro Bac_BoxForEachBo
        for  iTerm, k = iObj + 1, 0; iTerm < Bac_NtkObjNum(p) && Bac_ObjIsBo(p, iTerm); iTerm, k = iTerm+1, k+1  {
        if Bac_ObjNameType(p, iTerm) == BAC_NAME_BIN { fHaveWires = 1; iObj = Bac_NtkObjNum(p); break; }
        } }
    }
    // write bit-level wires
    if fHaveWires {
        Length = 7;
        Vec_StrPrintStr( vStr, "\n  wire " );
        // inlining of macro Bac_NtkForEachBox
        for  iObj = 0; (iObj < Bac_NtkObjNum(p)); iObj++  { if !Bac_ObjIsBox(p, iObj) {} else
        if !Bac_ObjIsConst(p, iObj) {
            // inlining of macro Bac_BoxForEachBo
            for  iTerm, k = iObj + 1, 0; iTerm < Bac_NtkObjNum(p) && Bac_ObjIsBo(p, iTerm); iTerm, k = iTerm+1, k+1  {
            if Bac_ObjNameType(p, iTerm) == BAC_NAME_BIN {
                if Length > 72 {
                    Vec_StrPrintStr( vStr, ";\n  wire " );
Length = 7; }
                if Length > 7 {
                    Vec_StrPrintStr( vStr, ", " ); }
                Vec_StrPrintStr( vStr, Bac_ObjNameStr(p, iTerm) ); 
                Length += len(Bac_ObjNameStr(p, iTerm));
            }
            } }
        }
        Vec_StrPrintStr( vStr, ";\n" );
    }
    Vec_StrPrintStr( vStr, "\n" );
    // write objects
    Bac_ManWriteVerilogBoxes( p, fUseAssign );
    Vec_StrPrintStr( vStr, "endmodule\n\n" );
}
func Bac_ManWriteVerilog(pFileName string, p *Bac_Man_t, fUseAssign int32) {
    var pNtk *Bac_Ntk_t; var i int32;
    // check the library
    if p.pMioLib && p.pMioLib != Abc_FrameReadLibGen() {
        fmt.Printf("Genlib library used in the mapped design is not longer a current library.\n");
        return;
    }
    // derive the stream
    p.vOut = Vec_StrAlloc( 10000 );
    p.vOut2 = Vec_StrAlloc( 1000 );
    Vec_StrPrintStr( p.vOut, "// Design \"" );
    Vec_StrPrintStr( p.vOut, Bac_ManName(p) );
    Vec_StrPrintStr( p.vOut, "\" written via CBA package in ABC on " );
    Vec_StrPrintStr( p.vOut, Extra_TimeStamp() );
    Vec_StrPrintStr( p.vOut, "\n\n" );
    Bac_ManAssignInternWordNames( p );
    // inlining of macro Bac_ManForEachNtk
    for  i = 1; (i <= Bac_ManNtkNum(p)) && (((pNtk) = Bac_ManNtk(p, i)), 1); i++  { 
        Bac_ManWriteVerilogNtk( pNtk, fUseAssign );
    }
    // dump into file
    if p.vOut && Vec_StrSize(p.vOut) > 0 {
        var pFile *os.File  = os.OpenFile(pFileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
        if pFile == nil {
            fmt.Printf("Cannot open file \"%s\" for writing.\n", pFileName);
        } else {
            fwrite( Vec_StrArray(p.vOut), 1, Vec_StrSize(p.vOut), pFile );
            _ = (pFile).Close();
        }
    }
    Vec_StrFreeP( &p.vOut );
    Vec_StrFreeP( &p.vOut2 );
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

