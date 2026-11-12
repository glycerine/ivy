package cba

import "fmt"

import "os"

/**CFile****************************************************************

  FileName    [cbaWriteVer.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Hierarchical word-level netlist.]

  Synopsis    [Verilog writer.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - November 29, 2014.]

  Revision    [$Id: cbaWriteVer.c,v 1.00 2014/11/29 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "cba.h"
//#include "cbaPrs.h"
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
  
  Synopsis    []  
  
  Description []  
                 
  SideEffects []  
  
  SeeAlso     []  
  
***********************************************************************/
func Cba_ManCreatePrimMap(pMap *string) {
    memset( pMap, 0, sizeof(char *) * CBA_BOX_LAST );

    pMap[ CBA_BOX_SLICE  ] = "sli";
    pMap[ CBA_BOX_CONCAT ] = "icc";

    pMap[ CBA_BOX_BUF    ] = "";
    pMap[ CBA_BOX_INV    ] = "~";
    pMap[ CBA_BOX_AND    ] = "&"; 
    pMap[ CBA_BOX_NAND   ] = "&"; 
    pMap[ CBA_BOX_OR     ] = "|"; 
    pMap[ CBA_BOX_NOR    ] = "|"; 
    pMap[ CBA_BOX_XOR    ] = "^"; 
    pMap[ CBA_BOX_XNOR   ] = "^"; 
    pMap[ CBA_BOX_SHARP  ] = "&"; 
    pMap[ CBA_BOX_SHARPL ] = "&"; 
    pMap[ CBA_BOX_MUX    ] = "?"; 
    pMap[ CBA_BOX_MAJ    ] = "maj"; 
    pMap[ CBA_BOX_RAND   ] = "&"; 
    pMap[ CBA_BOX_RNAND  ] = "~&"; 
    pMap[ CBA_BOX_ROR    ] = "|"; 
    pMap[ CBA_BOX_RNOR   ] = "~|"; 
    pMap[ CBA_BOX_RXOR   ] = "^"; 
    pMap[ CBA_BOX_RXNOR  ] = "~^"; 
    pMap[ CBA_BOX_LNOT   ] = "!"; 
    pMap[ CBA_BOX_LAND   ] = "&&"; 
    pMap[ CBA_BOX_LNAND  ] = "logic nand"; 
    pMap[ CBA_BOX_LOR    ] = "||"; 
    pMap[ CBA_BOX_LNOR   ] = "logic nor"; 
    pMap[ CBA_BOX_LXOR   ] = "^^"; 
    pMap[ CBA_BOX_LXNOR  ] = "logic xnor"; 
    pMap[ CBA_BOX_NMUX   ] = "nmux"; 
    pMap[ CBA_BOX_SEL    ] = "sel"; 
    pMap[ CBA_BOX_PSEL   ] = "prio-select"; 
    pMap[ CBA_BOX_ENC    ] = "enc"; 
    pMap[ CBA_BOX_PENC   ] = "prio-enc"; 
    pMap[ CBA_BOX_DEC    ] = "dec"; 
    pMap[ CBA_BOX_EDEC   ] = "enable-dec"; 
    pMap[ CBA_BOX_ADD    ] = "+"; 
    pMap[ CBA_BOX_SUB    ] = "-"; 
    pMap[ CBA_BOX_MUL    ] = "*"; 
    pMap[ CBA_BOX_DIV    ] = "/"; 
    pMap[ CBA_BOX_MOD    ] = "%"; 
    pMap[ CBA_BOX_REM    ] = "%"; 
    pMap[ CBA_BOX_POW    ] = "**"; 
    pMap[ CBA_BOX_MIN    ] = "-"; 
    pMap[ CBA_BOX_SQRT   ] = "@"; 
    pMap[ CBA_BOX_ABS    ] = "abs"; 
    pMap[ CBA_BOX_LTHAN  ] = "<"; 
    pMap[ CBA_BOX_LETHAN ] = "<="; 
    pMap[ CBA_BOX_METHAN ] = ">="; 
    pMap[ CBA_BOX_MTHAN  ] = ">"; 
    pMap[ CBA_BOX_EQU    ] = "=="; 
    pMap[ CBA_BOX_NEQU   ] = "!="; 
    pMap[ CBA_BOX_SHIL   ] = "<<"; 
    pMap[ CBA_BOX_SHIR   ] = ">>"; 
    pMap[ CBA_BOX_SHILA  ] = "<<<"; 
    pMap[ CBA_BOX_SHIRA  ] = ">>>"; 
    pMap[ CBA_BOX_ROTL   ] = "rotL"; 
    pMap[ CBA_BOX_ROTR   ] = "rotR"; 

    pMap[ CBA_BOX_TRI    ] = "tri";
    pMap[ CBA_BOX_RAM    ] = "ram";
    pMap[ CBA_BOX_RAMR   ] = "ramR";
    pMap[ CBA_BOX_RAMW   ] = "ramW";
    pMap[ CBA_BOX_RAMWC  ] = "ramWC";
    pMap[ CBA_BOX_RAMBOX ] = "ramBox";

    pMap[ CBA_BOX_LATCH  ] = "lat";
    pMap[ CBA_BOX_LATCHRS] = "latrs";
    pMap[ CBA_BOX_DFF    ] = "dff";
    pMap[ CBA_BOX_DFFRS  ] = "dffrs";
}

/**Function*************************************************************

  Synopsis    [Writing parser state into a file.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Prs_NameIsLegalInVerilog(pName string, NameId int32) int32 {
    // identifier ::= simple_identifier | escaped_identifier
    // simple_identifier ::= [a-zA-Z_][a-zA-Z0-9_$]
    // escaped_identifier ::= \ {Any_ASCII_character_except_white_space} white_space
    // white_space ::= space | tab | newline
    assert(pName != nil && *pName != '\0', "pName != nil && *pName != '\\0'");
    if *pName == '\\' {
        return 1; }
    if NameId < 13 { // see PRS_VER_UNKNOWN in cbaReadVer.c
        return 0; }
    if (*pName < 'a' || *pName > 'z') && (*pName < 'A' || *pName > 'Z') && *pName != '_' {
        return 0; }
    for  *(++pName)  {
        if (*pName < 'a' || *pName > 'z') && (*pName < 'A' || *pName > 'Z') && (*pName < '0' || *pName > '9') && *pName != '_' && *pName != '$' { 
            return 0; }
    }
    return 1;
}
func Prs_ObjGetName(p *Prs_Ntk_t, NameId int32) string {
    var pName string  = Prs_NtkStr(p, NameId);
    if pName == nil {
        return pName; }
    if Prs_NameIsLegalInVerilog(pName, NameId) != 0 {
        return pName; }
    return Vec_StrPrintF( Abc_NamBuffer(p.pStrs), "\\%s ", pName );
}
func Prs_ManWriteRange(p *Prs_Ntk_t, RangeId int32, fSlice int32) string {
    var Left int32;
    if RangeId != 0 {
        Left = Hash_IntObjData0( p.vHash, RangeId );
    } else {
        Left = 0;
    }
    var Right int32;
    if RangeId != 0 {
        Right = Hash_IntObjData1( p.vHash, RangeId );
    } else {
        Right = 0;
    }
    if Left == Right && (fSlice != 0) {
        return Vec_StrPrintF( Abc_NamBuffer(p.pStrs), "[%d]", Right );
    } else {
        return Vec_StrPrintF( Abc_NamBuffer(p.pStrs), "[%d:%d]", Left, Right ); }
}
func Prs_ManWriteVerilogConcat(pFile *os.File, p *Prs_Ntk_t, Con int32) {
    extern void Prs_ManWriteVerilogArray( FILE * pFile, Prs_Ntk_t * p, Vec_Int_t * vSigs, int fOdd );
    var vSigs *Vec_Int_t  = Prs_CatSignals(p, Con);
    fmt.Fprintf(pFile, "{");
    Prs_ManWriteVerilogArray( pFile, p, vSigs, 0 );
    fmt.Fprintf(pFile, "}");
}
func Prs_ManWriteVerilogSignal(pFile *os.File, p *Prs_Ntk_t, Sig int32) {
    var Value int32  = Abc_Lit2Var2( Sig );
    var Type Prs_ManType_t  = (Prs_ManType_t)Abc_Lit2Att2( Sig );
    if Type == CBA_PRS_NAME {
        fmt.Fprintf(pFile, "%s", Prs_ObjGetName(p, Value));
    } else if Type == CBA_PRS_CONST {
        fmt.Fprintf(pFile, "%s", Prs_NtkConst(p, Value));
    } else if Type == CBA_PRS_SLICE {
        fmt.Fprintf(pFile, "%s%s", Prs_ObjGetName(p, Prs_SliceName(p, Value)), Prs_ManWriteRange(p, Prs_SliceRange(p, Value), 1));
    } else if Type == CBA_PRS_CONCAT {
        Prs_ManWriteVerilogConcat( pFile, p, Value );
    } else { assert(false, "0"); }
}
func Prs_ManWriteVerilogArray(pFile *os.File, p *Prs_Ntk_t, vSigs *Vec_Int_t, fOdd int32) {
    var i int32;
    var Sig int32;
    var fFirst int32 = 1;
    assert(Vec_IntSize(vSigs) > 0, "Vec_IntSize(vSigs) > 0");
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vSigs)) && (((Sig) = Vec_IntEntry(vSigs, i)), 1); i++  {
        if (fOdd != 0) && !(i & 1) {
            continue; }
        fmt.Fprintf(pFile, "%s", fFirst ? "" : ", ");
        Prs_ManWriteVerilogSignal( pFile, p, Sig );
        fFirst = 0;
    }
}
func Prs_ManWriteVerilogArray2(pFile *os.File, p *Prs_Ntk_t, vSigs *Vec_Int_t) {
    var i, FormId, ActSig int32;
    assert(Vec_IntSize(vSigs) % 2 == 0, "Vec_IntSize(vSigs) % 2 == 0");
    // inlining of macro Vec_IntForEachEntryDouble
    for  i = 0; (i+1 < Vec_IntSize(vSigs)) && (((FormId) = Vec_IntEntry(vSigs, i)), 1) && (((ActSig) = Vec_IntEntry(vSigs, i+1)), 1); i += 2  {
        fmt.Fprintf(pFile, ".");
        fmt.Fprintf(pFile, "%s", Prs_ObjGetName(p, FormId));
        fmt.Fprintf(pFile, "(");
        Prs_ManWriteVerilogSignal( pFile, p, ActSig );
        fmt.Fprintf(pFile, ")%s", (i == Vec_IntSize(vSigs) - 2) ? "" : ", ");
    }
}
func Prs_ManWriteVerilogMux(pFile *os.File, p *Prs_Ntk_t, vSigs *Vec_Int_t) {
    var i, FormId, ActSig int32;
    var pStrs [4]string = [4]string{ " = ", " ? ", " : ", ";\n" };
    assert(Vec_IntSize(vSigs) == 8, "Vec_IntSize(vSigs) == 8");
    fmt.Fprintf(pFile, "  assign ");
    Prs_ManWriteVerilogSignal( pFile, p, Vec_IntEntry(vSigs, 1) );
    fmt.Fprintf(pFile, "%s", pStrs[0]);
    // inlining of macro Vec_IntForEachEntryDoubleStart
    for  i = 2; (i+1 < Vec_IntSize(vSigs)) && (((FormId) = Vec_IntEntry(vSigs, i)), 1) && (((ActSig) = Vec_IntEntry(vSigs, i+1)), 1); i += 2  {
        Prs_ManWriteVerilogSignal( pFile, p, ActSig );
        fmt.Fprintf(pFile, "%s", pStrs[i/2]);
    }
}
func Prs_ManWriteVerilogBoxes(pFile *os.File, p *Prs_Ntk_t, pTypeNames *string) {
    var vBox *Vec_Int_t; var i, k int32;
    // inlining of macro Prs_NtkForEachBox
    for  i = 0; i < Prs_NtkBoxNum(p) && ((vBox) = Prs_BoxSignals(p, i)); i++  {
        var NtkId Cba_ObjType_t  = (Cba_ObjType_t)Prs_BoxNtk(p, i);
        //char * pNtkName = Prs_ObjGetName(p, Prs_BoxName(p, i));
        if NtkId == CBA_BOX_MUX && Prs_BoxIsNode(p, i) {
            Prs_ManWriteVerilogMux( pFile, p, vBox );
        } else if Prs_BoxIsNode(p, i) { // node
        
            fmt.Fprintf(pFile, "  assign ");
            Prs_ManWriteVerilogSignal( pFile, p, Vec_IntEntry(vBox, 1) );
            fmt.Fprintf(pFile, " = ");
            if Cba_TypeIsUnary(NtkId) {
                fmt.Fprintf(pFile, "%s", pTypeNames[NtkId]);
                Prs_ManWriteVerilogSignal( pFile, p, Vec_IntEntry(vBox, 3) );
            
            } else if NtkId == CBA_BOX_NMUX {
                Prs_ManWriteVerilogSignal( pFile, p, Vec_IntEntry(vBox, 3) );
                fmt.Fprintf(pFile, " ? ");
                for  k = 5; k < Vec_IntSize(vBox); k += 2  {
                    if k > 5 { fmt.Fprintf(pFile, " : "); }
                    Prs_ManWriteVerilogSignal( pFile, p, Vec_IntEntry(vBox, k) );
                }
            
            } else if NtkId == CBA_BOX_ADD {
                if Vec_IntEntry(vBox, 3) {
                    Prs_ManWriteVerilogSignal( pFile, p, Vec_IntEntry(vBox, 3) );
                    fmt.Fprintf(pFile, " %s ", pTypeNames[NtkId]);
                }
                Prs_ManWriteVerilogSignal( pFile, p, Vec_IntEntry(vBox, 5) );
                fmt.Fprintf(pFile, " %s ", pTypeNames[NtkId]);
                Prs_ManWriteVerilogSignal( pFile, p, Vec_IntEntry(vBox, 7) );
            
            } else {
                Prs_ManWriteVerilogSignal( pFile, p, Vec_IntEntry(vBox, 3) );
                fmt.Fprintf(pFile, " %s ", pTypeNames[NtkId]);
                Prs_ManWriteVerilogSignal( pFile, p, Vec_IntEntry(vBox, 5) );
            }
            fmt.Fprintf(pFile, ";\n");
        
        } else { // box
        
            fmt.Fprintf(pFile, "  %s %s ( ", Prs_ObjGetName(p, NtkId), Prs_BoxName(p, i) ? Prs_ObjGetName(p, Prs_BoxName(p, i)) : "");
            Prs_ManWriteVerilogArray2( pFile, p, vBox );
            fmt.Fprintf(pFile, " );\n");
        }
    }
}
func Prs_ManWriteVerilogIos(pFile *os.File, p *Prs_Ntk_t, SigType int32) {
    var NameId, RangeId, i int32;
    var pSigNames [4]string = [4]string{ "inout", "input", "output", "wire" }; 
    var vSigs [4]*Vec_Int_t = [4]*Vec_Int_t{ &p.vInouts,  &p.vInputs,  &p.vOutputs,  &p.vWires };
    var vSigsR [4]*Vec_Int_t = [4]*Vec_Int_t{ &p.vInoutsR, &p.vInputsR, &p.vOutputsR, &p.vWiresR };
    if SigType == 3 {
        fmt.Fprintf(pFile, "\n"); }
    // inlining of macro Vec_IntForEachEntryTwo
    for  i = 0; (i < Vec_IntSize(vSigs[SigType])) && (((NameId) = Vec_IntEntry(vSigs[SigType], i)), 1) && (((RangeId) = Vec_IntEntry(vSigsR[SigType], i)), 1); i++  {
        fmt.Fprintf(pFile, "  %s %s%s", pSigNames[SigType], Abc_LitIsCompl(RangeId) ? "signed " : "", RangeId ? Prs_ManWriteRange(p, Abc_Lit2Var(RangeId), 0) : "");
        fmt.Fprintf(pFile, "%s;\n", Prs_ObjGetName(p, NameId));
    }
}
func Prs_ManWriteVerilogIoOrder(pFile *os.File, p *Prs_Ntk_t, vOrder *Vec_Int_t) {
    var i, NameId int32;
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vOrder)) && (((NameId) = Vec_IntEntry(vOrder, i)), 1); i++  {
        fmt.Fprintf(pFile, "%s%s", Prs_ObjGetName(p, Abc_Lit2Var2(NameId)), i == Vec_IntSize(vOrder) - 1 ? "" : ", ");
    }
}
func Prs_ManWriteVerilogNtk(pFile *os.File, p *Prs_Ntk_t, pTypeNames *string) {
    var s int32;
    // write header
    fmt.Fprintf(pFile, "module %s (\n    ", Prs_ObjGetName(p, p.iModuleName));
    Prs_ManWriteVerilogIoOrder( pFile, p, &p.vOrder );
    fmt.Fprintf(pFile, "\n  );\n");
    // write declarations
    for  s = 0; s < 4; s++  {
        Prs_ManWriteVerilogIos( pFile, p, s );
    }
    fmt.Fprintf(pFile, "\n");
    // write objects
    Prs_ManWriteVerilogBoxes( pFile, p, pTypeNames );
    fmt.Fprintf(pFile, "endmodule\n\n");
}
func Prs_ManWriteVerilog(pFileName string, vPrs *Vec_Ptr_t) {
    var pTypeNames [CBA_BOX_LAST]string;
    var pNtk *Prs_Ntk_t  = Prs_ManRoot(vPrs); var i int32;
    var pFile *os.File  = os.OpenFile(pFileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    if pFile == nil {
        fmt.Printf("Cannot open output file \"%s\".\n", pFileName);
        return;
    }
    Cba_ManCreatePrimMap( pTypeNames );
    fmt.Fprintf(pFile, "// Design \"%s\" written by ABC on %s\n\n", Prs_NtkStr(pNtk, pNtk.iModuleName), Extra_TimeStamp());
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vPrs)) && (((pNtk) = (Prs_Ntk_t *)Vec_PtrEntry(vPrs, i)), 1); i++  {
        Prs_ManWriteVerilogNtk( pFile, pNtk, pTypeNames );
    }
    _ = (pFile).Close();
}





func Cba_NameIsLegalInVerilog(pName string, NameId int32) int32 {
    // identifier ::= simple_identifier | escaped_identifier
    // simple_identifier ::= [a-zA-Z_][a-zA-Z0-9_$]
    // escaped_identifier ::= \ {Any_ASCII_character_except_white_space} white_space
    // white_space ::= space | tab | newline
    assert(pName != nil && *pName != '\0', "pName != nil && *pName != '\\0'");
    if *pName == '\\' {
        return 1; }
    if NameId < 13 { // see PRS_VER_UNKNOWN in cbaReadVer.c
        return 0; }
    if (*pName < 'a' || *pName > 'z') && (*pName < 'A' || *pName > 'Z') && *pName != '_' {
        return 0; }
    for  *(++pName)  {
        if (*pName < 'a' || *pName > 'z') && (*pName < 'A' || *pName > 'Z') && (*pName < '0' || *pName > '9') && *pName != '_' && *pName != '$' { 
            return 0; }
    }
    return 1;
}
func Cba_ObjGetName(p *Cba_Ntk_t, i int32) string {
    var pName string  = Cba_ObjNameStr(p, i);
    if pName == nil {
        return pName; }
    if Cba_NameIsLegalInVerilog(pName, Cba_ObjName(p, i)) != 0 {
        return pName; }
    return Vec_StrPrintF( Abc_NamBuffer(Cba_NtkNam(p)), "\\%s ", pName );
}
func Cba_FonGetName(p *Cba_Ntk_t, i int32) string {
    var pName string  = Cba_FonNameStr(p, i);
    if pName == nil {
        return pName; }
    if Cba_ObjType(p, Cba_FonObj(p, i)) == CBA_BOX_SLICE {
        return pName; }
    if Cba_NameIsLegalInVerilog(pName, Cba_FonName(p, i)) != 0 {
        return pName; }
    return Vec_StrPrintF( Abc_NamBuffer(Cba_NtkNam(p)), "\\%s ", pName );
}
func Cba_ManGetSliceName(p *Cba_Ntk_t, iFon int32, RangeId int32) string {
    var Left int32   = Cba_NtkRangeLeft(p, RangeId);
    var Right int32  = Cba_NtkRangeRight(p, RangeId);
    var pName string  = Cba_FonNameStr(p, iFon);
    if Cba_NameIsLegalInVerilog(pName, Cba_FonName(p, iFon)) != 0 {
        if Left == Right {
            return Vec_StrPrintF( Abc_NamBuffer(Cba_NtkNam(p)), "%s[%d]", pName, Right );
        } else {
            return Vec_StrPrintF( Abc_NamBuffer(Cba_NtkNam(p)), "%s[%d:%d]", pName, Left, Right ); }
    } else
        if Left == Right {
            return Vec_StrPrintF( Abc_NamBuffer(Cba_NtkNam(p)), "\\%s [%d]", pName, Right );
        } else {
            return Vec_StrPrintF( Abc_NamBuffer(Cba_NtkNam(p)), "\\%s [%d:%d]", pName, Left, Right ); }
}

func Cba_ManWriteFonRange(p *Cba_Ntk_t, iFon int32) {
    var vStr *Vec_Str_t  = &p.pDesign.vOut;
    if !(iFon != 0) || Cba_FonIsConst(iFon) || (Cba_FonRangeSize(p, iFon) == 1 && Cba_FonRight(p, iFon) == 0) {
        return; }
    if Cba_FonSigned(p, iFon) {
        Vec_StrPrintF( vStr, "signed " ); }
    Vec_StrPrintF( vStr, "[%d:%d] ", Cba_FonLeft(p, iFon), Cba_FonRight(p, iFon) );
}
func Cba_ManWriteFonName(p *Cba_Ntk_t, iFon int32, fInlineConcat int32, fInput int32) {
    extern void Cba_ManWriteConcat( Cba_Ntk_t * p, int iObj );
    var vStr *Vec_Str_t  = &p.pDesign.vOut;
    if !(iFon != 0) || (!Cba_FonIsConst(iFon) && !Cba_FonName(p, iFon)) { 
        Vec_StrPrintF( vStr, "Open_%d", Cba_NtkMan(p).nOpens++ );
//        Vec_StrPrintF( vStr, "1\'b0" );
    } else if (fInlineConcat != 0) && !Cba_FonIsConst(iFon) && Cba_ObjIsConcat(p, Cba_FonObj(p, iFon)) {
        Cba_ManWriteConcat( p, Cba_FonObj(p, iFon) );
    } else {
        var Range int32;
        if fInput != 0 {
            Range = Cba_FonRangeSize( p, iFon );
        } else {
            Range = 0;
        }
        if (fInput != 0) && Range > 1 {
            Vec_StrPush( vStr, '{' ); }
        Vec_StrPrintStr( vStr, Cba_FonIsConst(iFon) ? Cba_NtkConst(p, Cba_FonConst(iFon)) : Cba_FonGetName(p, iFon) );
        if (fInput != 0) && Range > 1 {
            Vec_StrPush( vStr, '}' ); }
    }
}
func Cba_ManWriteConcat(p *Cba_Ntk_t, iObj int32) {
    var i, iFin, iFon int32;
    var vStr *Vec_Str_t  = &p.pDesign.vOut;
    assert(Cba_ObjIsConcat(p, iObj), "Cba_ObjIsConcat(p, iObj)");
    Vec_StrPush( vStr, '{' );
    // inlining of macro Cba_ObjForEachFinFon
    for  i = 0, iFin = Cba_ObjFin0(p, iObj); iFin < Cba_ObjFin0(p, iObj+1) && ((iFon = Cba_FinFon(p, iFin)), 1); iFin, i = iFin+1, i+1  {
        Vec_StrPrintStr( vStr, i ? ", " : "" );
        Cba_ManWriteFonName( p, iFon, 1, 0 );
    }
    Vec_StrPush( vStr, '}' );
}
func Cba_ManWriteLineFile(p *Cba_Ntk_t, iObj int32, FileAttr int32, LineAttr int32) int32 {
    var vStr *Vec_Str_t  = &p.pDesign.vOut;
    var FileId, LineId int32;
    if FileAttr && (FileId = Cba_ObjAttrValue(p, iObj, FileAttr)) {
        LineId = Cba_ObjAttrValue(p, iObj, LineAttr);
        Vec_StrPrintF( vStr, "  // %s(%d)", Cba_NtkStr(p, FileId), LineId );
        return 1;
    }
    return 0;
}
func Cba_ManWriteVerilogNtk(p *Cba_Ntk_t, fInlineConcat int32) {
    var vStr *Vec_Str_t  = &p.pDesign.vOut;
    var i, k, iObj, iFin, iFon, StartPos, Status int32;
    var FileAttr int32  = Cba_NtkStrId( p, "file" );
    var LineAttr int32  = Cba_NtkStrId( p, "line" );
    var fUseNewLine int32  = int32((Cba_NtkPioNum(p) > 5));
    // mark PO fons
    var vPoFons *Vec_Bit_t  = Vec_BitStart( Cba_NtkFonNum(p)+1 );
    // inlining of macro Cba_NtkForEachPoDriverFon
    for  i = 0; (i < Cba_NtkPoNum(p))  && (((iObj) = Cba_NtkPo(p, i)), 1)  && (((iFon) = Cba_ObjFinFon(p, iObj, 0)), 1); i++  { 
        if Cba_FonIsReal(iFon) && Cba_FonName(p, iFon) == Cba_ObjName(p, iObj) {
            Vec_BitWriteEntry( vPoFons, iFon, 1 ); }
    }
    // write header
    Vec_StrPrintStr( vStr, "module " );
    Vec_StrPrintStr( vStr, Cba_NtkName(p) );
    Vec_StrPrintStr( vStr, fUseNewLine ? " (\n    " : " ( " );
    StartPos = Vec_StrSize(vStr);
    // inlining of macro Cba_NtkForEachPioOrder
    for  i = 0; (i < Cba_NtkPioOrderNum(p))  && (((iObj) = Cba_NtkPioOrder(p, i)), 1); i++  {
        Vec_StrPrintStr( vStr, i ? ", " : "" );
        if Vec_StrSize(vStr) > StartPos + 70 {
            StartPos = Vec_StrSize(vStr);
            Vec_StrPrintStr( vStr, "\n    " );
        }
        Vec_StrPrintStr( vStr, Cba_ObjGetName(p, iObj) );
    }
    Vec_StrPrintStr( vStr, fUseNewLine ? "\n  );" : " );" );
    Cba_ManWriteLineFile( p, 0, FileAttr, LineAttr );
    Vec_StrPrintStr( vStr, fUseNewLine ? "\n" : "\n\n" );
    // write inputs/outputs
    // inlining of macro Cba_NtkForEachPioOrder
    for  i = 0; (i < Cba_NtkPioOrderNum(p))  && (((iObj) = Cba_NtkPioOrder(p, i)), 1); i++  {
        var Offset int32  = Vec_StrSize(vStr);
        Vec_StrPrintStr( vStr, "  " );
        Vec_StrPrintStr( vStr, Cba_ObjIsPi(p, iObj) ? "input " : "output " );
        Cba_ManWriteFonRange( p, Cba_ObjIsPi(p, iObj) ? Cba_ObjFon0(p, iObj) : Cba_ObjFinFon(p, iObj, 0) );
        Vec_StrPrintStr( vStr, Cba_ObjGetName(p, iObj) );
        Vec_StrPrintF( vStr, ";%*s", Offset + 40 - Vec_StrSize(vStr), "" );
        Cba_ManWriteLineFile( p, iObj, FileAttr, LineAttr );
        Vec_StrPush( vStr, '\n' );
    }
    Vec_StrPrintStr( vStr, "\n" );
    // write objects
    // inlining of macro Cba_NtkForEachBox
    for  iObj = 1; iObj < Vec_StrSize(&p.vObjType); iObj++  { if !Cba_ObjIsBox(p, iObj) {} else {
//        char * pNameNtk = Cba_NtkName(p);
//        char * pNameInst = Cba_ObjGetName(p, iObj);
        var Type int32  = Cba_ObjType(p, iObj);
        if Cba_ObjIsSlice(p, iObj) {
            continue; }
        if (fInlineConcat != 0) && Cba_ObjIsConcat(p, iObj) {
            continue; }
        if Cba_ObjIsBoxUser(p, iObj) {
            var pNtk *Cba_Ntk_t  = Cba_ObjNtk(p, iObj);
            // write output wire declarations
            // inlining of macro Cba_ObjForEachFon
            for  i = 0, iFon = Cba_ObjFon0(p, iObj); iFon < Cba_ObjFon0(p, iObj+1); iFon, i = iFon+1, i+1  {
                if Vec_BitEntry(vPoFons, iFon) {
                    continue; }
                Vec_StrPrintStr( vStr, "  wire " );
                Cba_ManWriteFonRange( p, iFon );
                Cba_ManWriteFonName( p, iFon, 0, 0 );
                Vec_StrPrintStr( vStr, ";\n" );
            }
            // write box
            Vec_StrPrintStr( vStr, "  " );
            Vec_StrPrintStr( vStr, Cba_NtkName(pNtk) );
            Vec_StrPush( vStr, ' ' );
            if Cba_ObjName(p, iObj) {
                Vec_StrPrintF( vStr, "%s ", Cba_ObjGetName(p, iObj) ); }
            // write input binding
            Vec_StrPrintStr( vStr, "( " );
            // inlining of macro Cba_ObjForEachFinFon
            for  i = 0, iFin = Cba_ObjFin0(p, iObj); iFin < Cba_ObjFin0(p, iObj+1) && ((iFon = Cba_FinFon(p, iFin)), 1); iFin, i = iFin+1, i+1  {
                Vec_StrPrintF( vStr, "%s.%s(", i ? ", " : "", Cba_ObjGetName(pNtk, Cba_NtkPi(pNtk, i)) );
                Cba_ManWriteFonName( p, iFon, fInlineConcat, 1 );
                Vec_StrPush( vStr, ')' );
            }
            // write output binding
            // inlining of macro Cba_ObjForEachFon
            for  i = 0, iFon = Cba_ObjFon0(p, iObj); iFon < Cba_ObjFon0(p, iObj+1); iFon, i = iFon+1, i+1  {
                Vec_StrPrintF( vStr, "%s.%s(", Cba_ObjFinNum(p, iObj) ? ", " : "", Cba_ObjGetName(pNtk, Cba_NtkPo(pNtk, i)) );
                Cba_ManWriteFonName( p, iFon, 0, 1 );
                Vec_StrPush( vStr, ')' );
            }
            Vec_StrPrintStr( vStr, ");" );
        
        } else if Type == CBA_BOX_RAMWC || Type == CBA_BOX_RAMBOX || Type == CBA_BOX_RAMR {
            var Num int32;
            if Type == CBA_BOX_RAMWC {
                Num = 0;
            } else {
                if Type == CBA_BOX_RAMR {
                    Num = 1;
                } else {
                    Num = 2;
                }
            } // write / read / box
            var pBoxName [3]string = [3]string{ "RAM_WRITE", "RAM_READ", "RAM_BOX" };
            var pOutputs [3]string = [3]string{ "ram", "rdata", "out" };
            var pInputs [3][4]string = [3][4]string{  {"clk", "wen", "waddr", "wdata"},  {"ren", "raddr", "ram"},   {"in0", "in1", "in2", "in3"} };
            // write declaration for the output
            var iFonOut int32  = Cba_ObjFon0(p, iObj);
            if Vec_BitEntry(vPoFons, iFonOut) { 
                Vec_StrPrintStr( vStr, "  assign " );
            } else {
                Vec_StrPrintStr( vStr, "  wire " );
                Cba_ManWriteFonRange( p, iFonOut );
            }
            Cba_ManWriteFonName( p, iFonOut, 0, 0 );
            Vec_StrPrintStr( vStr, ";\n" );
            // write box
            Vec_StrPrintF( vStr, "  %s( ", pBoxName[Num] );
            // inlining of macro Cba_ObjForEachFinFon
            for  i = 0, iFin = Cba_ObjFin0(p, iObj); iFin < Cba_ObjFin0(p, iObj+1) && ((iFon = Cba_FinFon(p, iFin)), 1); iFin, i = iFin+1, i+1  {
                Vec_StrPrintF( vStr, "%s.%s(", i ? ", " : "", pInputs[Num][i] );
                Cba_ManWriteFonName( p, iFon, fInlineConcat, 1 );
                Vec_StrPush( vStr, ')' );
            }
            Vec_StrPrintF( vStr, "%s.%s(", Cba_ObjFinNum(p, iObj) ? ", " : "", pOutputs[Num] );
            Cba_ManWriteFonName( p, iFonOut, 0, 1 );
            Vec_StrPrintStr( vStr, ") );" );
        
        } else if Type == CBA_BOX_NMUX || Type == CBA_BOX_SEL {
            var fUseSel int32  = Type == CBA_BOX_SEL;
            var nBits int32;
            if fUseSel != 0 {
                nBits = Cba_ObjFinNum(p, iObj) - 1;
            } else {
                nBits = Abc_Base2Log(Cba_ObjFinNum(p, iObj) - 1);
            }
            var iFonIn int32   = Cba_ObjFinFon(p, iObj, 0);
            var iFonOut int32  = Cba_ObjFon0(p, iObj);
            // function [15:0] res;
            Vec_StrPrintStr( vStr, "  function " );
            Cba_ManWriteFonRange( p, iFonOut );
            Vec_StrPrintStr( vStr, "_func_" );
            Cba_ManWriteFonName( p, iFonOut, 0, 0 );
            Vec_StrPrintStr( vStr, ";\n" );
            // input [1:0] s;
            Vec_StrPrintStr( vStr, "    input " );
            Cba_ManWriteFonRange( p, iFonIn );
            Vec_StrPrintStr( vStr, "s;\n" );
            // input [15:0] d0, d1, d2, d3;
            Vec_StrPrintStr( vStr, "    input " );
            Cba_ManWriteFonRange( p, iFonOut );
            // inlining of macro Cba_ObjForEachFinFon
            for  i = 0, iFin = Cba_ObjFin0(p, iObj); iFin < Cba_ObjFin0(p, iObj+1) && ((iFon = Cba_FinFon(p, iFin)), 1); iFin, i = iFin+1, i+1  {
                if i == 0 { continue; }
                Vec_StrPrintF( vStr, "%sd%d", i > 1 ? ", " : "", i-1 );
            }
            Vec_StrPrintStr( vStr, ";\n" );
            // casez (s)
            Vec_StrPrintStr( vStr, "    casez(s)" );
            if fUseSel != 0 {
                Vec_StrPrintStr( vStr, "  // synopsys full_case parallel_case" ); }
            Vec_StrPrintStr( vStr, "\n" );
            // 2'b00: res = b[15:0];
            // inlining of macro Cba_ObjForEachFinFon
            for  i = 0, iFin = Cba_ObjFin0(p, iObj); iFin < Cba_ObjFin0(p, iObj+1) && ((iFon = Cba_FinFon(p, iFin)), 1); iFin, i = iFin+1, i+1  {
                if i == 0 { continue; }
                Vec_StrPrintF( vStr, "      %d\'b", nBits );
                if fUseSel != 0 {
                    Vec_StrFillExtra( vStr, Vec_StrSize(vStr) + nBits, '?' );
                    Vec_StrWriteEntry( vStr, Vec_StrSize(vStr) - i, '1' );
                
                } else {
                    for  k = nBits-1; k >= 0; k--  {
                        Vec_StrPrintNum( vStr, ((i-1) >> k) & 1 );
                    }
                }
                Vec_StrPrintStr( vStr, ": _func_" );
                Cba_ManWriteFonName( p, iFonOut, 0, 0 );
                Vec_StrPrintF( vStr, " = d%d;\n", i-1 );
            }
            Vec_StrPrintStr( vStr, "    endcase\n" );
            Vec_StrPrintStr( vStr, "  endfunction\n" );
            // assign z = res(a, b, s);
            if Vec_BitEntry(vPoFons, iFonOut) { 
                Vec_StrPrintStr( vStr, "  assign " );
            } else {
                Vec_StrPrintStr( vStr, "  wire " );
                Cba_ManWriteFonRange( p, iFonOut );
            }
            Cba_ManWriteFonName( p, iFonOut, fInlineConcat, 0 );
            Vec_StrPrintStr( vStr, " = _func_" );
            Cba_ManWriteFonName( p, iFonOut, 0, 0 );
            Vec_StrPrintStr( vStr, " ( " );
            // inlining of macro Cba_ObjForEachFinFon
            for  i = 0, iFin = Cba_ObjFin0(p, iObj); iFin < Cba_ObjFin0(p, iObj+1) && ((iFon = Cba_FinFon(p, iFin)), 1); iFin, i = iFin+1, i+1  {
                Vec_StrPrintStr( vStr, i ? ", " : "" );
                Cba_ManWriteFonName( p, iFon, fInlineConcat, 0 );
            }
            Vec_StrPrintStr( vStr, " );" );
        
        } else if Type == CBA_BOX_DEC {
            var iFonIn int32    = Cba_ObjFinFon(p, iObj, 0);
            var iFonOut int32   = Cba_ObjFon0(p, iObj);
            var nBitsIn int32   = Cba_FonRangeSize(p, iFonIn);
            var nBitsOut int32  = Cba_FonRangeSize(p, iFonOut);
            assert((1 << nBitsIn) == nBitsOut, "(1 << nBitsIn) == nBitsOut");
            // function [15:0] res;
            Vec_StrPrintStr( vStr, "  function " );
            Cba_ManWriteFonRange( p, iFonOut );
            Vec_StrPrintStr( vStr, "_func_" );
            Cba_ManWriteFonName( p, iFonOut, 0, 0 );
            Vec_StrPrintStr( vStr, ";\n" );
            // input [1:0] i;
            Vec_StrPrintStr( vStr, "    input " );
            Cba_ManWriteFonRange( p, iFonIn );
            Vec_StrPrintStr( vStr, "i;\n" );
            // casez (i)
            Vec_StrPrintStr( vStr, "    casez(i)\n" );
            // 2'b00: res = 4'b0001;
            for  i = 0; i < (1 << nBitsIn); i++  {
                Vec_StrPrintF( vStr, "      %d\'b", nBitsIn );
                for  k = nBitsIn-1; k >= 0; k--  {
                    Vec_StrPrintNum( vStr, (i >> k) & 1 );
                }
                Vec_StrPrintStr( vStr, ": _func_" );
                Cba_ManWriteFonName( p, iFonOut, 0, 0 );
                Vec_StrPrintF( vStr, " = %d\'b%0*d;\n", nBitsOut, nBitsOut, 0 );
                Vec_StrWriteEntry( vStr, Vec_StrSize(vStr) - i - 3, '1' );
            }
            Vec_StrPrintStr( vStr, "    endcase\n" );
            Vec_StrPrintStr( vStr, "  endfunction\n" );
            // assign z = res(i, o);
            if Vec_BitEntry(vPoFons, iFonOut) { 
                Vec_StrPrintStr( vStr, "  assign " );
            } else {
                Vec_StrPrintStr( vStr, "  wire " );
                Cba_ManWriteFonRange( p, iFonOut );
            }
            Cba_ManWriteFonName( p, iFonOut, fInlineConcat, 0 );
            Vec_StrPrintStr( vStr, " = _func_" );
            Cba_ManWriteFonName( p, iFonOut, 0, 0 );
            Vec_StrPrintStr( vStr, " ( " );
            Cba_ManWriteFonName( p, iFonIn, fInlineConcat, 0 );
            Vec_StrPrintStr( vStr, " );" );
         
        } else if Type == CBA_BOX_DFFRS || Type == CBA_BOX_LATCHRS {
            var fUseFlop int32  = Type == CBA_BOX_DFFRS;
            var iFonQ int32     = Cba_ObjFon0(p, iObj);
            var iFonD int32     = Cba_ObjFinFon(p, iObj, 0);
            var iFonSet int32   = Cba_ObjFinFon(p, iObj, 1);
            var iFonRst int32   = Cba_ObjFinFon(p, iObj, 2);
            var iFonC int32     = Cba_ObjFinFon(p, iObj, 3);
            var Range int32     = Cba_FonRangeSize( p, iFonQ );
            assert(Cba_FonRangeSize(p, iFonSet) == 1 && Cba_FonRangeSize(p, iFonRst) == 1, "Cba_FonRangeSize(p, iFonSet) == 1 && Cba_FonRangeSize(p, iFonRst) == 1");
            // reg    [3:0] Q;
            Vec_StrPrintStr( vStr, "  reg " );
            Cba_ManWriteFonRange( p, iFonQ );
            Cba_ManWriteFonName( p, iFonQ, 0, 0 );
            Vec_StrPrintStr( vStr, ";\n" );
            // always @(posedge C or posedge PRE)
            Vec_StrPrintStr( vStr, "  always @(" );
            if fUseFlop != 0 {
                Vec_StrPrintStr( vStr, "posedge " ); }
            Cba_ManWriteFonName( p, iFonC, 0, 0 );
            if !(fUseFlop != 0) {
                Vec_StrPrintStr( vStr, " or " );
                Cba_ManWriteFonName( p, iFonD, 0, 0 );
            }
            if iFonSet > 0 {
                Vec_StrPrintStr( vStr, " or " );
                if fUseFlop != 0 {
                    Vec_StrPrintStr( vStr, "posedge " ); }
                Cba_ManWriteFonName( p, iFonSet, 0, 0 );
            }
            if iFonRst > 0 {
                Vec_StrPrintStr( vStr, " or " );
                if fUseFlop != 0 {
                    Vec_StrPrintStr( vStr, "posedge " ); }
                Cba_ManWriteFonName( p, iFonRst, 0, 0 );
            }
            Vec_StrPrintStr( vStr, ")\n" );
            //  if (Set)  Q <= 4'b1111;
            if iFonSet > 0 {
                Vec_StrPrintStr( vStr, "    if (" );
                Cba_ManWriteFonName( p, iFonSet, 0, 0 );
                Vec_StrPrintStr( vStr, ")  " );
                Cba_ManWriteFonName( p, iFonQ, 0, 0 );
                Vec_StrPrintStr( vStr, fUseFlop ? " <= " : " = " );
                // value 1
                Vec_StrPrintNum( vStr, Range );
                Vec_StrPrintStr( vStr, "\'b" );
                Vec_StrFillExtra( vStr, Vec_StrSize(vStr) + Range, '1' );
                Vec_StrPrintStr( vStr, ";\n" );
//                Vec_StrPrintF( vStr, "\'b1" );
            }
            if iFonRst > 0 {
                Vec_StrPrintStr( vStr, iFonSet > 0 ? "    else if (" : "    if (" );
                Cba_ManWriteFonName( p, iFonRst, 0, 0 );
                Vec_StrPrintStr( vStr, ")  " );
                Cba_ManWriteFonName( p, iFonQ, 0, 0 );
                Vec_StrPrintStr( vStr, fUseFlop ? " <= " : " = " );
                // value 0
                Vec_StrPrintNum( vStr, Range );
                Vec_StrPrintStr( vStr, "\'b" );
                Vec_StrFillExtra( vStr, Vec_StrSize(vStr) + Range, '0' );
                Vec_StrPrintStr( vStr, ";\n" );
//                Vec_StrPrintF( vStr, "\'b0" );
            }
            Vec_StrPrintStr( vStr, (iFonSet > 0 || iFonRst > 0) ? "    else " : "    " );
            if !(fUseFlop != 0) {
                Vec_StrPrintStr( vStr, "    if (" );
                Cba_ManWriteFonName( p, iFonC, 0, 0 );
                Vec_StrPrintStr( vStr, ")  " );
            }
            Cba_ManWriteFonName( p, iFonQ, 0, 0 );
            Vec_StrPrintStr( vStr, fUseFlop ? " <= " : " = " );
            Cba_ManWriteFonName( p, iFonD, fInlineConcat, 0 );
            Vec_StrPrintStr( vStr, ";" );
        
        } else if Type == CBA_BOX_DFFCPL {
            // CPL_FF#32 inst_reg( .d(s1), .arstval(s2), .arst(s3), .clk(s4), .q(s5), .qbar(s6) );
            var iFon0 int32  = Cba_ObjFon0(p, iObj);
            var iFon1 int32  = Cba_ObjFon(p, iObj, 1);
            var Range int32  = Cba_FonRangeSize( p, iFon0 );
            if !Vec_BitEntry(vPoFons, iFon0) {
                Vec_StrPrintStr( vStr, "  wire " );
                Cba_ManWriteFonRange( p, iFon0 );
                Cba_ManWriteFonName( p, iFon0, 0, 0 );
                Vec_StrPrintStr( vStr, ";\n" );
            }
            if !Vec_BitEntry(vPoFons, iFon1) && Cba_FonName(p, iFon1) {
                Vec_StrPrintStr( vStr, "  wire " );
                Cba_ManWriteFonRange( p, iFon1 );
                Cba_ManWriteFonName( p, iFon1, 0, 0 );
                Vec_StrPrintStr( vStr, ";\n" );
            }
            Vec_StrPrintStr( vStr, "  CPL_FF" );
            if Range > 1 {
                Vec_StrPrintF( vStr, "#%d", Range ); }
            Vec_StrPrintStr( vStr, " " );
            if Cba_ObjName(p, iObj) {
                Vec_StrPrintStr( vStr, Cba_ObjGetName(p, iObj) ); }
            Vec_StrPrintStr( vStr, " ( .d(" );
            Cba_ManWriteFonName( p, Cba_ObjFinFon(p, iObj, 0), fInlineConcat, 0 );
            Vec_StrPrintStr( vStr, "), .arstval(" );
            Cba_ManWriteFonName( p, Cba_ObjFinFon(p, iObj, 1), fInlineConcat, 0 );
            Vec_StrPrintStr( vStr, "), .arst(" );
            Cba_ManWriteFonName( p, Cba_ObjFinFon(p, iObj, 2), fInlineConcat, 0 );
            Vec_StrPrintStr( vStr, "), .clk(" );
            Cba_ManWriteFonName( p, Cba_ObjFinFon(p, iObj, 3), fInlineConcat, 0 );
            Vec_StrPrintStr( vStr, "), .q(" );
            Cba_ManWriteFonName( p, iFon0, fInlineConcat, 0 );
            Vec_StrPrintStr( vStr, "), .qbar(" );
            if Cba_FonName(p, iFon1) {
                Cba_ManWriteFonName( p, iFon1, fInlineConcat, 0 ); }
            Vec_StrPrintStr( vStr, ") );" );
        
        } else if Type == CBA_BOX_ADD {
            var iFon0 int32  = Cba_ObjFon0(p, iObj);
            var iFon1 int32  = Cba_ObjFon(p, iObj, 1);
            // write outputs
            if Cba_FonName(p, iFon1) {
                if !Vec_BitEntry(vPoFons, iFon0) {
                    Vec_StrPrintStr( vStr, "  wire " );
                    Cba_ManWriteFonRange( p, iFon0 );
                    Cba_ManWriteFonName( p, iFon0, 0, 0 );
                    Vec_StrPrintStr( vStr, ";\n" );
                }
                if !Vec_BitEntry(vPoFons, iFon1) {
                    Vec_StrPrintStr( vStr, "  wire " );
                    Cba_ManWriteFonRange( p, iFon1 );
                    Cba_ManWriteFonName( p, iFon1, 0, 0 );
                    Vec_StrPrintStr( vStr, ";\n" );
                }
                Vec_StrPrintStr( vStr, "  assign {" );
                Cba_ManWriteFonName( p, iFon1, 0, 0 );
                Vec_StrPrintStr( vStr, ", " );
                Cba_ManWriteFonName( p, iFon0, 0, 0 );
                Vec_StrPrintStr( vStr, "} = " );
            
            } else {
                if Vec_BitEntry(vPoFons, iFon0) { 
                    Vec_StrPrintStr( vStr, "  assign " );
                } else {
                    Vec_StrPrintStr( vStr, "  wire " );
                    Cba_ManWriteFonRange( p, iFon0 );
                }
                Cba_ManWriteFonName( p, iFon0, 0, 0 );
                Vec_StrPrintStr( vStr, " = " );
            }
            // write carry-in
            if Cba_ObjFinFon(p, iObj, 0) && Cba_ObjFinFon(p, iObj, 0) != Cba_FonFromConst(1) {
                Vec_StrPush( vStr, ' ' );
                Cba_ManWriteFonName( p, Cba_ObjFinFon(p, iObj, 0), fInlineConcat, 0 );
                Vec_StrPush( vStr, ' ' );
                Vec_StrPrintStr( vStr, "+" );
            }
            Vec_StrPush( vStr, ' ' );
            Cba_ManWriteFonName( p, Cba_ObjFinFon(p, iObj, 1), fInlineConcat, 0 );
            Vec_StrPush( vStr, ' ' );
            Vec_StrPrintStr( vStr, "+" );
            Vec_StrPush( vStr, ' ' );
            Cba_ManWriteFonName( p, Cba_ObjFinFon(p, iObj, 2), fInlineConcat, 0 );
            Vec_StrPush( vStr, ';' );
        
        } else {
            if Vec_BitEntry(vPoFons, Cba_ObjFon0(p, iObj)) { 
                Vec_StrPrintStr( vStr, "  assign " );
            } else {
                Vec_StrPrintStr( vStr, "  wire " );
                Cba_ManWriteFonRange( p, Cba_ObjFon0(p, iObj) );
            }
            Cba_ManWriteFonName( p, Cba_ObjFon0(p, iObj), 0, 0 );
            Vec_StrPrintStr( vStr, " = " );
            if Cba_ObjIsConcat(p, iObj) {
                Cba_ManWriteConcat( p, iObj );
            } else if Type == CBA_BOX_MUX {
                Cba_ManWriteFonName( p, Cba_ObjFinFon(p, iObj, 0), fInlineConcat, 0 );
                Vec_StrPrintStr( vStr, " ? " );
                Cba_ManWriteFonName( p, Cba_ObjFinFon(p, iObj, 1), fInlineConcat, 0 );
                Vec_StrPrintStr( vStr, " : " );
                Cba_ManWriteFonName( p, Cba_ObjFinFon(p, iObj, 2), fInlineConcat, 0 );
            
            } else if Type == CBA_BOX_ROTL || Type == CBA_BOX_ROTR {
                //  wire [27:0] s4960 = (s57 >> 17) | (s57 << 11);
                var Range int32  = Cba_FonRangeSize( p, Cba_ObjFon0(p, iObj) );
                var iFinFon1 int32  = Cba_ObjFinFon(p, iObj, 1);
                Vec_StrPush( vStr, '(' );
                Cba_ManWriteFonName( p, Cba_ObjFinFon(p, iObj, 0), fInlineConcat, 0 );
                Vec_StrPrintStr( vStr, Type == CBA_BOX_ROTL ? " << " : " >> " );
                if Cba_FonIsConst(iFinFon1) {
                    Vec_StrPrintNum( vStr, Cba_FonConst(iFinFon1) );
                } else {
                    Cba_ManWriteFonName( p, Cba_ObjFinFon(p, iObj, 1), fInlineConcat, 0 ); }
                Vec_StrPrintStr( vStr, ") | (" );
                Cba_ManWriteFonName( p, Cba_ObjFinFon(p, iObj, 0), fInlineConcat, 0 );
                Vec_StrPrintStr( vStr, Type == CBA_BOX_ROTL ? " >> " : " << " );
                if Cba_FonIsConst(iFinFon1) {
                    Vec_StrPrintNum( vStr, Range - Cba_FonConst(iFinFon1) );
                } else {
                    Vec_StrPush( vStr, '(' );
                    Vec_StrPrintNum( vStr, Range );
                    Vec_StrPrintStr( vStr, " - " );
                    Cba_ManWriteFonName( p, Cba_ObjFinFon(p, iObj, 1), fInlineConcat, 0 );
                    Vec_StrPush( vStr, ')' );
                }
                Vec_StrPush( vStr, ')' );
            
            } else if Type == CBA_BOX_LTHAN {
                var fLessThan int32  = (Cba_ObjFinFon(p, iObj, 0) == Cba_FonFromConst(1)); // const0
                Vec_StrPush( vStr, ' ' );
                Cba_ManWriteFonName( p, Cba_ObjFinFon(p, iObj, 1), fInlineConcat, 0 );
                Vec_StrPush( vStr, ' ' );
                Vec_StrPrintStr( vStr, fLessThan ? "<" : "<=" );
                Vec_StrPush( vStr, ' ' );
                Cba_ManWriteFonName( p, Cba_ObjFinFon(p, iObj, 2), fInlineConcat, 0 );
            
            } else if Cba_TypeIsUnary((Cba_ObjType_t)Type) {
                Vec_StrPrintStr( vStr, Cba_NtkTypeName(p, Type) );
                Cba_ManWriteFonName( p, Cba_ObjFinFon(p, iObj, 0), fInlineConcat, 0 );
            
            } else if Cba_NtkTypeName(p, Type) { // binary operation
            
                var fCompl int32  = (Type == CBA_BOX_NAND || Type == CBA_BOX_NOR || Type == CBA_BOX_XNOR);
                if fCompl != 0 {
                    Vec_StrPrintStr( vStr, "!(" ); }
                Cba_ManWriteFonName( p, Cba_ObjFinFon(p, iObj, 0), fInlineConcat, 0 );
                Vec_StrPush( vStr, ' ' );
                Vec_StrPrintStr( vStr, Cba_NtkTypeName(p, Type) );
                Vec_StrPush( vStr, ' ' );
                Cba_ManWriteFonName( p, Cba_ObjFinFon(p, iObj, 1), fInlineConcat, 0 );
                if fCompl != 0 {
                    Vec_StrPrintStr( vStr, ")" ); }
            
            } else { // unknown
            
                var pName string  = Cba_FonGetName(p, Cba_ObjFon0(p, iObj));
                Vec_StrPrintStr( vStr, "<unknown operator>" );
                fmt.Printf("Cba_ManWriteVerilog(): In module \"%s\", cannot write object \"%s\" with output name \"%s\".\n", Cba_NtkName(p), Cba_ObjGetName(p, iObj), pName);
            }
            Vec_StrPush( vStr, ';' );
        }
        // write attributes
        Status = Cba_ManWriteLineFile( p, iObj, FileAttr, LineAttr );
        if !Cba_ObjIsBoxUser(p, iObj) && Cba_ObjName(p, iObj) {
            if !(Status != 0) {
                Vec_StrPrintStr( vStr, "  //" ); }
            Vec_StrPrintStr( vStr, " name=" );
            Vec_StrPrintStr( vStr, Cba_ObjGetName(p, iObj) );
        }
        Vec_StrPush( vStr, '\n' );
    }
    }
    // write remaining outputs
    // inlining of macro Cba_NtkForEachPo
    for  i = 0; (i < Cba_NtkPoNum(p))  && (((iObj) = Cba_NtkPo(p, i)), 1); i++  {
        iFon = Cba_ObjFinFon(p, iObj, 0);
        if !(iFon != 0) || (!Cba_FonIsConst(iFon) && Cba_FonName(p, iFon) == Cba_ObjName(p, iObj)) { // already written
            continue; }
        Vec_StrPrintStr( vStr, "  assign " );
        Vec_StrPrintStr( vStr, Cba_ObjGetName(p, iObj) );
        Vec_StrPrintStr( vStr, " = " );
        Cba_ManWriteFonName( p, iFon, fInlineConcat, 0 );
        Vec_StrPush( vStr, ';' );
        Vec_StrPush( vStr, '\n' );
    }
    Vec_StrPrintStr( vStr, "\n" );
    Vec_StrPrintStr( vStr, "endmodule\n\n" );
    Vec_BitFree( vPoFons );
}
func Cba_ManWriteVerilog(pFileName string, p *Cba_Man_t, fInlineConcat int32) {
    var pNtk *Cba_Ntk_t; var i int32;
    // check the library
    if p.pMioLib && p.pMioLib != Abc_FrameReadLibGen() {
        fmt.Printf("Genlib library used in the mapped design is not longer a current library.\n");
        return;
    }
    Cba_ManCreatePrimMap( p.pTypeNames );
    // derive the stream
    p.nOpens = 1;
    Vec_StrClear( &p.vOut );
    Vec_StrClear( &p.vOut2 );
    Vec_StrPrintStr( &p.vOut, "// Design \"" );
    Vec_StrPrintStr( &p.vOut, Cba_ManName(p) );
    Vec_StrPrintStr( &p.vOut, "\" written via CBA package in ABC on " );
    Vec_StrPrintStr( &p.vOut, Extra_TimeStamp() );
    Vec_StrPrintStr( &p.vOut, "\n\n" );
    // inlining of macro Cba_ManForEachNtk
    for  i = 1; (i <= Cba_ManNtkNum(p)) && (((pNtk) = Cba_ManNtk(p, i)), 1); i++  { 
        Cba_ManWriteVerilogNtk( pNtk, fInlineConcat );
    }
    // dump into file
    if Vec_StrSize(&p.vOut) > 0 {
        var pFile *os.File  = os.OpenFile(pFileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
        if pFile == nil {
            fmt.Printf("Cannot open file \"%s\" for writing.\n", pFileName);
        } else {
            fwrite( Vec_StrArray(&p.vOut), 1, Vec_StrSize(&p.vOut), pFile );
            _ = (pFile).Close();
        }
    }
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

