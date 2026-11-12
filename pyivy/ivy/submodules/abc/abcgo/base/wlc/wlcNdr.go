package wlc

import "fmt"

/**CFile****************************************************************

  FileName    [wlcNdr.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Experimental procedures.]

  Synopsis    [Constructing WLC network from NDR data structure.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - August 22, 2014.]

  Revision    [$Id: wlcNdr.c,v 1.00 2014/09/12 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "wlc.h"
//#include "aig/miniaig/ndr.h"

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
func Ndr_TypeNdr2Wlc(Type int32) int32 {
    if Type == ABC_OPER_CONST {         return WLC_OBJ_CONST; }         // 06: constant
    if Type == ABC_OPER_BIT_BUF {       return WLC_OBJ_BUF; }           // 07: buffer
    if Type == ABC_OPER_BIT_MUX {       return WLC_OBJ_MUX; }           // 08: multiplexer
    if Type == ABC_OPER_SHIFT_R {       return WLC_OBJ_SHIFT_R; }       // 09: shift right
    if Type == ABC_OPER_SHIFT_RA {      return WLC_OBJ_SHIFT_RA; }      // 10: shift right (arithmetic)
    if Type == ABC_OPER_SHIFT_L {       return WLC_OBJ_SHIFT_L; }       // 11: shift left
    if Type == ABC_OPER_SHIFT_LA {      return WLC_OBJ_SHIFT_LA; }      // 12: shift left (arithmetic)
    if Type == ABC_OPER_SHIFT_ROTR {    return WLC_OBJ_ROTATE_R; }      // 13: rotate right
    if Type == ABC_OPER_SHIFT_ROTL {    return WLC_OBJ_ROTATE_L; }      // 14: rotate left
    if Type == ABC_OPER_BIT_INV {       return WLC_OBJ_BIT_NOT; }       // 15: bitwise NOT
    if Type == ABC_OPER_BIT_AND {       return WLC_OBJ_BIT_AND; }       // 16: bitwise AND
    if Type == ABC_OPER_BIT_OR {        return WLC_OBJ_BIT_OR; }        // 17: bitwise OR
    if Type == ABC_OPER_BIT_XOR {       return WLC_OBJ_BIT_XOR; }       // 18: bitwise XOR
    if Type == ABC_OPER_BIT_NAND {      return WLC_OBJ_BIT_NAND; }      // 19: bitwise AND
    if Type == ABC_OPER_BIT_NOR {       return WLC_OBJ_BIT_NOR; }       // 20: bitwise OR
    if Type == ABC_OPER_BIT_NXOR {      return WLC_OBJ_BIT_NXOR; }      // 21: bitwise NXOR
    if Type == ABC_OPER_SLICE {         return WLC_OBJ_BIT_SELECT; }    // 22: bit selection
    if Type == ABC_OPER_CONCAT {        return WLC_OBJ_BIT_CONCAT; }    // 23: bit concatenation
    if Type == ABC_OPER_ZEROPAD {       return WLC_OBJ_BIT_ZEROPAD; }   // 24: zero padding
    if Type == ABC_OPER_SIGNEXT {       return WLC_OBJ_BIT_SIGNEXT; }   // 25: sign extension
    if Type == ABC_OPER_LOGIC_NOT {     return WLC_OBJ_LOGIC_NOT; }     // 26: logic NOT
    if Type == ABC_OPER_LOGIC_IMPL {    return WLC_OBJ_LOGIC_IMPL; }    // 27: logic implication
    if Type == ABC_OPER_LOGIC_AND {     return WLC_OBJ_LOGIC_AND; }     // 28: logic AND
    if Type == ABC_OPER_LOGIC_OR {      return WLC_OBJ_LOGIC_OR; }      // 29: logic OR
    if Type == ABC_OPER_LOGIC_XOR {     return WLC_OBJ_LOGIC_XOR; }     // 30: logic XOR
    if Type == ABC_OPER_SEL_NMUX {      return WLC_OBJ_MUX; }           // 08: multiplexer
    if Type == ABC_OPER_SEL_SEL {       return WLC_OBJ_SEL; }           // 57: selector
    if Type == ABC_OPER_SEL_DEC {       return WLC_OBJ_DEC; }           // 58: decoder
    if Type == ABC_OPER_COMP_EQU {      return WLC_OBJ_COMP_EQU; }      // 31: compare equal
    if Type == ABC_OPER_COMP_NOTEQU {   return WLC_OBJ_COMP_NOTEQU; }   // 32: compare not equal
    if Type == ABC_OPER_COMP_LESS {     return WLC_OBJ_COMP_LESS; }     // 33: compare less
    if Type == ABC_OPER_COMP_MORE {     return WLC_OBJ_COMP_MORE; }     // 34: compare more
    if Type == ABC_OPER_COMP_LESSEQU {  return WLC_OBJ_COMP_LESSEQU; }  // 35: compare less or equal
    if Type == ABC_OPER_COMP_MOREEQU {  return WLC_OBJ_COMP_MOREEQU; }  // 36: compare more or equal
    if Type == ABC_OPER_RED_AND {       return WLC_OBJ_REDUCT_AND; }    // 37: reduction AND
    if Type == ABC_OPER_RED_OR {        return WLC_OBJ_REDUCT_OR; }     // 38: reduction OR
    if Type == ABC_OPER_RED_XOR {       return WLC_OBJ_REDUCT_XOR; }    // 39: reduction XOR
    if Type == ABC_OPER_RED_NAND {      return WLC_OBJ_REDUCT_NAND; }   // 40: reduction NAND
    if Type == ABC_OPER_RED_NOR {       return WLC_OBJ_REDUCT_NOR; }    // 41: reduction NOR
    if Type == ABC_OPER_RED_NXOR {      return WLC_OBJ_REDUCT_NXOR; }   // 42: reduction NXOR
    if Type == ABC_OPER_ARI_ADD {       return WLC_OBJ_ARI_ADD; }       // 43: arithmetic addition
    if Type == ABC_OPER_ARI_SUB {       return WLC_OBJ_ARI_SUB; }       // 44: arithmetic subtraction
    if Type == ABC_OPER_ARI_MUL {       return WLC_OBJ_ARI_MULTI; }     // 45: arithmetic multiplier
    if Type == ABC_OPER_ARI_DIV {       return WLC_OBJ_ARI_DIVIDE; }    // 46: arithmetic division
    if Type == ABC_OPER_ARI_REM {       return WLC_OBJ_ARI_REM; }       // 47: arithmetic remainder
    if Type == ABC_OPER_ARI_MOD {       return WLC_OBJ_ARI_MODULUS; }   // 48: arithmetic modulus
    if Type == ABC_OPER_ARI_POW {       return WLC_OBJ_ARI_POWER; }     // 49: arithmetic power
    if Type == ABC_OPER_ARI_MIN {       return WLC_OBJ_ARI_MINUS; }     // 50: arithmetic minus
    if Type == ABC_OPER_ARI_SQRT {      return WLC_OBJ_ARI_SQRT; }      // 51: integer square root
    if Type == ABC_OPER_ARI_SQUARE {    return WLC_OBJ_ARI_SQUARE; }    // 52: integer square
    if Type == ABC_OPER_ARI_ADDSUB {    return WLC_OBJ_ARI_ADDSUB; }    // 56: adder-subtractor
    if Type == ABC_OPER_ARI_SMUL {      return WLC_OBJ_ARI_MULTI; }     // 45: signed multiplier
    if Type == ABC_OPER_DFF {           return WLC_OBJ_FO; }            // 03: flop
    if Type == ABC_OPER_DFFRSE {        return WLC_OBJ_FF; }            // 05: flop
    if Type == ABC_OPER_RAMR {          return WLC_OBJ_READ; }          // 54: read port
    if Type == ABC_OPER_RAMW {          return WLC_OBJ_WRITE; }         // 55: write port
    if Type == ABC_OPER_LUT {           return WLC_OBJ_LUT; }           // 59: LUT
    return -1;
}
func Ndr_TypeWlc2Ndr(Type int32) int32 {
    if Type == WLC_OBJ_CONST {          return ABC_OPER_CONST; }        // 06: constant
    if Type == WLC_OBJ_BUF {            return ABC_OPER_BIT_BUF; }      // 07: buffer
    if Type == WLC_OBJ_MUX {            return ABC_OPER_BIT_MUX; }      // 08: multiplexer
    if Type == WLC_OBJ_SHIFT_R {        return ABC_OPER_SHIFT_R; }      // 09: shift right
    if Type == WLC_OBJ_SHIFT_RA {       return ABC_OPER_SHIFT_RA; }     // 10: shift right (arithmetic)
    if Type == WLC_OBJ_SHIFT_L {        return ABC_OPER_SHIFT_L; }      // 11: shift left
    if Type == WLC_OBJ_SHIFT_LA {       return ABC_OPER_SHIFT_LA; }     // 12: shift left (arithmetic)
    if Type == WLC_OBJ_ROTATE_R {       return ABC_OPER_SHIFT_ROTR; }   // 13: rotate right
    if Type == WLC_OBJ_ROTATE_L {       return ABC_OPER_SHIFT_ROTL; }   // 14: rotate left
    if Type == WLC_OBJ_BIT_NOT {        return ABC_OPER_BIT_INV; }      // 15: bitwise NOT
    if Type == WLC_OBJ_BIT_AND {        return ABC_OPER_BIT_AND; }      // 16: bitwise AND
    if Type == WLC_OBJ_BIT_OR {         return ABC_OPER_BIT_OR; }       // 17: bitwise OR
    if Type == WLC_OBJ_BIT_XOR {        return ABC_OPER_BIT_XOR; }      // 18: bitwise XOR
    if Type == WLC_OBJ_BIT_NAND {       return ABC_OPER_BIT_NAND; }     // 19: bitwise AND
    if Type == WLC_OBJ_BIT_NOR {        return ABC_OPER_BIT_NOR; }      // 20: bitwise OR
    if Type == WLC_OBJ_BIT_NXOR {       return ABC_OPER_BIT_NXOR; }     // 21: bitwise NXOR
    if Type == WLC_OBJ_BIT_SELECT {     return ABC_OPER_SLICE; }        // 22: bit selection
    if Type == WLC_OBJ_BIT_CONCAT {     return ABC_OPER_CONCAT; }       // 23: bit concatenation
    if Type == WLC_OBJ_BIT_ZEROPAD {    return ABC_OPER_ZEROPAD; }      // 24: zero padding
    if Type == WLC_OBJ_BIT_SIGNEXT {    return ABC_OPER_SIGNEXT; }      // 25: sign extension
    if Type == WLC_OBJ_LOGIC_NOT {      return ABC_OPER_LOGIC_NOT; }    // 26: logic NOT
    if Type == WLC_OBJ_LOGIC_IMPL {     return ABC_OPER_LOGIC_IMPL; }   // 27: logic implication
    if Type == WLC_OBJ_LOGIC_AND {      return ABC_OPER_LOGIC_AND; }    // 28: logic AND
    if Type == WLC_OBJ_LOGIC_OR {       return ABC_OPER_LOGIC_OR; }     // 29: logic OR
    if Type == WLC_OBJ_LOGIC_XOR {      return ABC_OPER_LOGIC_XOR; }    // 30: logic XOR
    if Type == WLC_OBJ_SEL {            return ABC_OPER_SEL_SEL; }      // 57: selector
    if Type == WLC_OBJ_DEC {            return ABC_OPER_SEL_DEC; }      // 58: decoder
    if Type == WLC_OBJ_COMP_EQU {       return ABC_OPER_COMP_EQU; }     // 31: compare equal
    if Type == WLC_OBJ_COMP_NOTEQU {    return ABC_OPER_COMP_NOTEQU; }  // 32: compare not equal
    if Type == WLC_OBJ_COMP_LESS {      return ABC_OPER_COMP_LESS; }    // 33: compare less
    if Type == WLC_OBJ_COMP_MORE {      return ABC_OPER_COMP_MORE; }    // 34: compare more
    if Type == WLC_OBJ_COMP_LESSEQU {   return ABC_OPER_COMP_LESSEQU; } // 35: compare less or equal
    if Type == WLC_OBJ_COMP_MOREEQU {   return ABC_OPER_COMP_MOREEQU; } // 36: compare more or equal
    if Type == WLC_OBJ_REDUCT_AND {     return ABC_OPER_RED_AND; }      // 37: reduction AND
    if Type == WLC_OBJ_REDUCT_OR {      return ABC_OPER_RED_OR; }       // 38: reduction OR
    if Type == WLC_OBJ_REDUCT_XOR {     return ABC_OPER_RED_XOR; }      // 39: reduction XOR
    if Type == WLC_OBJ_REDUCT_NAND {    return ABC_OPER_RED_NAND; }     // 40: reduction NAND
    if Type == WLC_OBJ_REDUCT_NOR {     return ABC_OPER_RED_NOR; }      // 41: reduction NOR
    if Type == WLC_OBJ_REDUCT_NXOR {    return ABC_OPER_RED_NXOR; }     // 42: reduction NXOR
    if Type == WLC_OBJ_ARI_ADD {        return ABC_OPER_ARI_ADD; }      // 43: arithmetic addition
    if Type == WLC_OBJ_ARI_SUB {        return ABC_OPER_ARI_SUB; }      // 44: arithmetic subtraction
    if Type == WLC_OBJ_ARI_MULTI {      return ABC_OPER_ARI_MUL; }      // 45: arithmetic multiplier
    if Type == WLC_OBJ_ARI_DIVIDE {     return ABC_OPER_ARI_DIV; }      // 46: arithmetic division
    if Type == WLC_OBJ_ARI_REM {        return ABC_OPER_ARI_REM; }      // 47: arithmetic remainder
    if Type == WLC_OBJ_ARI_MODULUS {    return ABC_OPER_ARI_MOD; }      // 48: arithmetic modulus
    if Type == WLC_OBJ_ARI_POWER {      return ABC_OPER_ARI_POW; }      // 49: arithmetic power
    if Type == WLC_OBJ_ARI_MINUS {      return ABC_OPER_ARI_MIN; }      // 50: arithmetic minus
    if Type == WLC_OBJ_ARI_SQRT {       return ABC_OPER_ARI_SQRT; }     // 51: integer square root
    if Type == WLC_OBJ_ARI_SQUARE {     return ABC_OPER_ARI_SQUARE; }   // 52: integer square
    if Type == WLC_OBJ_ARI_ADDSUB {     return ABC_OPER_ARI_ADDSUB; }   // 56: adder-subtractor
    if Type == WLC_OBJ_ARI_MULTI {      return ABC_OPER_ARI_SMUL; }     // 45: signed multiplier
    if Type == WLC_OBJ_FO {             return ABC_OPER_DFFRSE; }       // 03: flop
    if Type == WLC_OBJ_FF {             return ABC_OPER_DFFRSE; }       // 05: flop
    if Type == WLC_OBJ_READ {           return ABC_OPER_RAMR; }         // 54: read port
    if Type == WLC_OBJ_WRITE {          return ABC_OPER_RAMW; }         // 55: write port
    if Type == WLC_OBJ_LUT {            return ABC_OPER_LUT; }          // 59: LUT
    return -1;
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_Ndr_ObjWriteConstant_Buffer string // was [10000]byte in C

func Ndr_ObjWriteConstant(pBits *uint32, nBits int32) string {
     var i, Len int32;
    assert(nBits + 10 < 10000, "nBits + 10 < 10000");
    spStatic_Ndr_ObjWriteConstant_Buffer = fmt.Sprintf("%d\'b", nBits);
    Len = strlen(spStatic_Ndr_ObjWriteConstant_Buffer);
    for  i = nBits-1; i >= 0; i--  {
        spStatic_Ndr_ObjWriteConstant_Buffer[Len++] = '0' + Abc_InfoHasBit(pBits, i);
    }
    spStatic_Ndr_ObjWriteConstant_Buffer[Len] = 0;
    return spStatic_Ndr_ObjWriteConstant_Buffer;
}
func Wlc_NtkToNdr(pNtk *Wlc_Ntk_t) any {
    var pObj *Wlc_Obj_t; 
    var i, k, iFanin, iOutId, Type int32;
    // create a new module
    var pDesign any  = Ndr_Create( 1 );
    var ModId int32  = Ndr_AddModule( pDesign, 1 );
    // add primary inputs
    var vFanins *Vec_Int_t  = Vec_IntAlloc( 10 );
    // inlining of macro Wlc_NtkForEachPi
    for  i = 0; (i < Wlc_NtkPiNum(pNtk)) && (((pObj) = Wlc_NtkPi(pNtk, i)), 1); i++  {
        iOutId = Wlc_ObjId(pNtk, pObj);
        Ndr_AddObject( pDesign, ModId, ABC_OPER_CI, 0,   
            pObj.End, pObj.Beg, pObj.Signed,   
            0, nil,      1, &iOutId,  nil  ); // no fanins
    }
    // add internal nodes 
    // inlining of macro Wlc_NtkForEachObj
    for  iOutId = 1; (iOutId < Wlc_NtkObjNumMax(pNtk)) && (((pObj) = Wlc_NtkObj(pNtk, iOutId)), 1); iOutId++  {
        var pFunction string  = nil;
        if Wlc_ObjIsPi(pObj) || pObj.Type == 0 {
            continue; }
        Vec_IntClear( vFanins );
        // inlining of macro Wlc_ObjForEachFanin
        for  k = 0; (k < Wlc_ObjFaninNum(pObj)) && (((iFanin) = Wlc_ObjFaninId(pObj, k)), 1); k++  {
            Vec_IntPush( vFanins, iFanin );
        }
        if pObj.Type == WLC_OBJ_CONST {
            pFunction = Ndr_ObjWriteConstant( (unsigned *)Wlc_ObjFanins(pObj), Wlc_ObjRange(pObj) ); }
        if pObj.Type == WLC_OBJ_MUX && Wlc_ObjRange(Wlc_ObjFanin0(pNtk, pObj)) > 1 {
            Type = ABC_OPER_SEL_NMUX;
        } else if pObj.Type == WLC_OBJ_FO {
            var pFi *Wlc_Obj_t  = Wlc_ObjFo2Fi( pNtk, pObj );
            assert(Vec_IntSize(vFanins) == 0, "Vec_IntSize(vFanins) == 0");
            Vec_IntPush( vFanins, Wlc_ObjId(pNtk, pFi) );
            Vec_IntFillExtra( vFanins, 7, 0 );
            Type = ABC_OPER_DFFRSE;
        
        } else {
            Type = Ndr_TypeWlc2Ndr(pObj.Type); }
        assert(Type > 0, "Type > 0");
        Ndr_AddObject( pDesign, ModId, Type, 0,   
            pObj.End, pObj.Beg, pObj.Signed,   
            Vec_IntSize(vFanins), Vec_IntArray(vFanins), 1, &iOutId,  pFunction  ); 
    }
    // add primary outputs
    // inlining of macro Wlc_NtkForEachObj
    for  iOutId = 1; (iOutId < Wlc_NtkObjNumMax(pNtk)) && (((pObj) = Wlc_NtkObj(pNtk, iOutId)), 1); iOutId++  {
        if !Wlc_ObjIsPo(pObj) {
            continue; }
        Vec_IntFill( vFanins, 1, iOutId );
        Ndr_AddObject( pDesign, ModId, ABC_OPER_CO, 0,                   
            pObj.End, pObj.Beg, pObj.Signed,   
            1, Vec_IntArray(vFanins),  0, nil,  nil ); 
    }
    Vec_IntFree( vFanins );
    return pDesign;
}
func Wlc_WriteNdr(pNtk *Wlc_Ntk_t, pFileName string) {
    var pDesign any  = Wlc_NtkToNdr( pNtk );
    Ndr_Write( pFileName, pDesign );
    Ndr_Delete( pDesign );
    fmt.Printf("Dumped the current design into file \"%s\".\n", pFileName);
}
func Wlc_NtkToNdrTest(pNtk *Wlc_Ntk_t) {
    // transform
    var pDesign any  = Wlc_NtkToNdr( pNtk );

    // collect names     
    var pObj *Wlc_Obj_t;  var i int32;
    var ppNames *string  = ABC_ALLOC( char *, Wlc_NtkObjNum(pNtk) + 1 );
    // inlining of macro Wlc_NtkForEachObj
    for  i = 1; (i < Wlc_NtkObjNumMax(pNtk)) && (((pObj) = Wlc_NtkObj(pNtk, i)), 1); i++  { 
        ppNames[i] = Wlc_ObjName(pNtk, i);
    }

    // verify by writing Verilog
    Ndr_WriteVerilog( nil, pDesign, ppNames );
    Ndr_Write( "test.ndr", pDesign );

    // cleanup
    Ndr_Delete( pDesign );
    ABC_FREE( ppNames );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ndr_ObjReadRange(p *Ndr_Data_t, Obj int32, End *int32, Beg *int32) int32 {
    var pArray *int32;
    var nArray int32 = Ndr_ObjReadArray( p, Obj, NDR_RANGE, &pArray );
    var Signed int32; *End = *Beg = 0;
    if nArray == 0 {
        return 0; }
    if nArray == 3 {
        Signed = 1; }
    if nArray == 1 {
        *End = *Beg = pArray[0];
    } else {
        *End = pArray[0], *Beg = pArray[1]; }
    return Signed;
}
func Ndr_ObjReadConstant(vFanins *Vec_Int_t, pStr string) {
    var i int32;
    var k int32;
    var Len int32;
    if pStr {
        Len = len(pStr);
    } else {
        Len = 0;
    }
    for  k = 0; k < Len; k++  {
        if pStr[k] == 'b' {
            break; }
    }
    if pStr == nil || pStr[k] != 'b' {
        fmt.Printf("Constants should be represented in binary Verilog notation <nbits>\'b<bits> as char strings (for example, \"4'b1010\").\n");
        return;
    }
    Vec_IntFill( vFanins, Abc_BitWordNum(Len-k-1), 0 );
    for  i = k+1; i < Len; i++  {
        if pStr[i] == '1' {
            Abc_InfoSetBit( (unsigned *)Vec_IntArray(vFanins), Len-i-1 );
        } else if pStr[i] != '0' {
            fmt.Printf("Wrongn symbol (%c) in binary Verilog constant \"%s\".\n", pStr[i], pStr); }
    }
}
func Ndr_NtkPrintNodes(pNtk *Wlc_Ntk_t) {
    var pObj *Wlc_Obj_t; var i, k int32;
    fmt.Printf("Node IDs and their fanins:\n");
    // inlining of macro Wlc_NtkForEachObj
    for  i = 1; (i < Wlc_NtkObjNumMax(pNtk)) && (((pObj) = Wlc_NtkObj(pNtk, i)), 1); i++  {
        var pFanins *int32  = Wlc_ObjFanins(pObj);
        fmt.Printf("%5d = ", i);
        for  k = 0; k < Wlc_ObjFaninNum(pObj); k++  {
            fmt.Printf("%5d ", pFanins[k]);
        }
        for       ; k < 4; k++  {
            fmt.Printf("      ");
        }
        fmt.Printf("    Name Id %d ", Wlc_ObjNameId(pNtk, i));
        if Wlc_ObjIsPi(pObj) {
            fmt.Printf("  pi  "); }
        if Wlc_ObjIsPo(pObj) {
            fmt.Printf("  po  "); }
        fmt.Printf("\n");
    }
}
func Wlc_NtkCheckIntegrity(pData any) {
    var p *Ndr_Data_t  = (Ndr_Data_t *)pData;  
    var vMap *Vec_Int_t  = Vec_IntAlloc( 100 );
    var Mod int32 = 2;
    var Obj int32;
    // inlining of macro Ndr_ModForEachObj
    for  Obj = Mod + 1; Obj < Ndr_DataEnd(p, Mod); Obj += Ndr_DataSize(p, Obj)  { if Ndr_DataType(p, Obj) != NDR_OBJECT {} else {
        var NameId int32   = Ndr_ObjReadBody( p, Obj, NDR_OUTPUT );
        if NameId == -1 {
            var Type int32  = Ndr_ObjReadBody( p, Obj, NDR_OPERTYPE );
            if Type != ABC_OPER_CO {
                fmt.Printf("Internal object %d of type %s has no output name.\n", Obj, Abc_OperName(Type)); }
            continue;
        }
        if Vec_IntGetEntry(vMap, NameId) > 0 {
            fmt.Printf("Output name %d is used more than once (obj %d and obj %d).\n", NameId, Vec_IntGetEntry(vMap, NameId), Obj); }
        Vec_IntSetEntry( vMap, NameId, Obj );
    }
    }
    // inlining of macro Ndr_ModForEachObj
    for  Obj = Mod + 1; Obj < Ndr_DataEnd(p, Mod); Obj += Ndr_DataSize(p, Obj)  { if Ndr_DataType(p, Obj) != NDR_OBJECT {} else {
        var Type int32  = Ndr_ObjReadBody( p, Obj, NDR_OPERTYPE );
        var i int32;
        var pArray *int32;
        var nArray int32 = Ndr_ObjReadArray( p, Obj, NDR_INPUT, &pArray );
        for  i = 0; i < nArray; i++  {
            if Vec_IntGetEntry(vMap, pArray[i]) == 0 && !(Type == ABC_OPER_DFFRSE && (i >= 5 && i <= 7)) {
                fmt.Printf("Input name %d appearing as fanin %d of obj %d is not used as output name in any object.\n", pArray[i], i, Obj); }
        }
    }
    }
    Vec_IntFree( vMap );
}
func Wlc_NtkFromNdr(pData any) *Wlc_Ntk_t {
    var p *Ndr_Data_t  = (Ndr_Data_t *)pData;  
    var pObj *Wlc_Obj_t; var vName2Obj *Vec_Int_t;
var vFanins *Vec_Int_t = Vec_IntAlloc( 100 );
    var Mod int32 = 2;
    var i int32;
    var k int32;
    var Obj int32;
    var pArray *int32;
    var nDigits int32;
    var fFound int32;
    var NameId int32;
    var NameIdMax int32;
    var vTruths *Vec_Wrd_t  = nil;
    var pTemp *Wlc_Ntk_t;
    var pNtk *Wlc_Ntk_t = Wlc_NtkAlloc( "top", Ndr_DataObjNum(p, Mod)+1 );
    Wlc_NtkCheckIntegrity( pData );
    Vec_IntClear( &pNtk.vFfs );
    //pNtk->pSpec = Abc_UtilStrsav( pFileName );
    // construct network and save name IDs
    Wlc_NtkCleanNameId( pNtk );
    // inlining of macro Ndr_ModForEachPi
    // inlining of macro Ndr_ModForEachObj
    for  Obj = Mod + 1; Obj < Ndr_DataEnd(p, Mod); Obj += Ndr_DataSize(p, Obj)  { if Ndr_DataType(p, Obj) != NDR_OBJECT {} else if !Ndr_ObjIsType(p, Obj, ABC_OPER_CI) {} else {
        var End int32;
        var Beg int32;
        var Signed int32 = Ndr_ObjReadRange(p, Obj, &End, &Beg);
        var iObj int32  = Wlc_ObjAlloc( pNtk, WLC_OBJ_PI, Signed, End, Beg );
        var NameId int32  = Ndr_ObjReadBody( p, Obj, NDR_OUTPUT );
        Wlc_ObjSetNameId( pNtk, iObj, NameId );
    }
    }
    // inlining of macro Ndr_ModForEachNode
    // inlining of macro Ndr_ModForEachObj
    for  Obj = Mod + 1; Obj < Ndr_DataEnd(p, Mod); Obj += Ndr_DataSize(p, Obj)  { if Ndr_DataType(p, Obj) != NDR_OBJECT {} else if Ndr_ObjIsType(p, Obj, ABC_OPER_CI) || Ndr_ObjIsType(p, Obj, ABC_OPER_CO) {} else {
        var End int32;
        var Beg int32;
        var Signed int32 = Ndr_ObjReadRange(p, Obj, &End, &Beg);
        var Type int32     = Ndr_ObjReadBody( p, Obj, NDR_OPERTYPE );
        var nArray int32   = Ndr_ObjReadArray( p, Obj, NDR_INPUT, &pArray );
        var F Vec_Int_t = Vec_Int_t{nArray, nArray, pArray};
        var vTemp *Vec_Int_t = &F;
        var iObj int32     = Wlc_ObjAlloc( pNtk, Ndr_TypeNdr2Wlc(Type), Signed, End, Beg );
        var NameId int32   = Ndr_ObjReadBody( p, Obj, NDR_OUTPUT );
        Vec_IntClear( vFanins );
        Vec_IntAppend( vFanins, vTemp );
        if Type == ABC_OPER_DFF {
            // save init state
            if pNtk.vInits == nil {
                pNtk.vInits = Vec_IntAlloc( 100 ); }
            if Vec_IntSize(vFanins) == 2 {
                Vec_IntPush( pNtk.vInits, Vec_IntPop(vFanins) );
            } else { // assume const0 if init is not given
                Vec_IntPush( pNtk.vInits, -(End-Beg+1) ); }
            // save flop output
            pObj = Wlc_NtkObj(pNtk, iObj);
            assert(Wlc_ObjType(pObj) == WLC_OBJ_FO, "Wlc_ObjType(pObj) == WLC_OBJ_FO");
            Wlc_ObjSetNameId( pNtk, iObj, NameId );
            Vec_IntPush( &pNtk.vFfs, NameId );
            // save flop input
            assert(Vec_IntSize(vFanins) == 1, "Vec_IntSize(vFanins) == 1");
            Vec_IntPush( &pNtk.vFfs, Vec_IntEntry(vFanins, 0) );
            continue;
        }
        if Type == ABC_OPER_DFFRSE {
            Vec_IntPush( &pNtk.vFfs2, iObj ); }
        if Type == ABC_OPER_LUT {
            if vTruths == nil {
                vTruths = Vec_WrdStart( 1000 ); }
            if NameId >= Vec_WrdSize(vTruths) {
                Vec_WrdFillExtra( vTruths, 2*NameId, 0 ); }
            Vec_WrdWriteEntry( vTruths, NameId, *((word *)Ndr_ObjReadBodyP(p, Obj, NDR_FUNCTION)) );
        }
        if Type == ABC_OPER_SLICE {
            Vec_IntPushTwo( vFanins, End, Beg );
        } else if Type == ABC_OPER_CONST {
            Ndr_ObjReadConstant( vFanins, (char *)Ndr_ObjReadBodyP(p, Obj, NDR_FUNCTION) );
        } else if Type == ABC_OPER_BIT_MUX && Vec_IntSize(vFanins) == 3 {
            ABC_SWAP( int, Vec_IntEntryP(vFanins, 1)[0], Vec_IntEntryP(vFanins, 2)[0] ); }
        Wlc_ObjAddFanins( pNtk, Wlc_NtkObj(pNtk, iObj), vFanins );
        Wlc_ObjSetNameId( pNtk, iObj, NameId );
        if Type == ABC_OPER_ARI_SMUL {
            pObj = Wlc_NtkObj(pNtk, iObj);
            assert(Wlc_ObjFaninNum(pObj) == 2, "Wlc_ObjFaninNum(pObj) == 2");
            Wlc_ObjFanin0(pNtk, pObj).Signed = 1;
            Wlc_ObjFanin1(pNtk, pObj).Signed = 1;
        }
    }
    }
    // mark primary outputs
    // inlining of macro Ndr_ModForEachPo
    // inlining of macro Ndr_ModForEachObj
    for  Obj = Mod + 1; Obj < Ndr_DataEnd(p, Mod); Obj += Ndr_DataSize(p, Obj)  { if Ndr_DataType(p, Obj) != NDR_OBJECT {} else if !Ndr_ObjIsType(p, Obj, ABC_OPER_CO) {} else {
        var End int32;
        var Beg int32;
        var Signed int32 = Ndr_ObjReadRange(p, Obj, &End, &Beg);
        var nArray int32   = Ndr_ObjReadArray( p, Obj, NDR_INPUT, &pArray );
        var iObj int32     = Wlc_ObjAlloc( pNtk, WLC_OBJ_BUF, Signed, End, Beg );
        var NameId int32   = Ndr_ObjReadBody( p, Obj, NDR_OUTPUT );
        assert(nArray == 1 && NameId == -1, "nArray == 1 && NameId == -1");
        pObj = Wlc_NtkObj( pNtk, iObj );
        Vec_IntFill( vFanins, 1, pArray[0] );
        Wlc_ObjAddFanins( pNtk, pObj, vFanins );
        Wlc_ObjSetCo( pNtk, pObj, 0 );
    }
    }
    Vec_IntFree( vFanins );
    // map name IDs into object IDs
    vName2Obj = Vec_IntInvert( &pNtk.vNameIds, 0 );
    // inlining of macro Wlc_NtkForEachObj
    for  i = 1; (i < Wlc_NtkObjNumMax(pNtk)) && (((pObj) = Wlc_NtkObj(pNtk, i)), 1); i++  {
        var pFanins *int32  = Wlc_ObjFanins(pObj);
        for  k = 0; k < Wlc_ObjFaninNum(pObj); k++  {
            pFanins[k] = Vec_IntEntry(vName2Obj, pFanins[k]);
        }
    }
    if pNtk.vInits {
        Vec_IntForEachEntry( &pNtk.vFfs, NameId, i )
            Vec_IntWriteEntry( &pNtk.vFfs, i, Vec_IntEntry(vName2Obj, NameId) );
        Vec_IntForEachEntry( pNtk.vInits, NameId, i )
            if ( NameId > 0 )
                Vec_IntWriteEntry( pNtk.vInits, i, Vec_IntEntry(vName2Obj, NameId) );
        // move FO/FI to be part of CI/CO
        assert((Vec_IntSize(&pNtk.vFfs) & 1) == 0, "(Vec_IntSize(&pNtk.vFfs) & 1) == 0");
        assert(Vec_IntSize(&pNtk.vFfs) == 2 * Vec_IntSize(pNtk.vInits), "Vec_IntSize(&pNtk.vFfs) == 2 * Vec_IntSize(pNtk.vInits)");
        // inlining of macro Wlc_NtkForEachFf
        for  i = 0; (i < Vec_IntSize(&pNtk.vFfs)) && (((pObj) = Wlc_NtkFf(pNtk, i)), 1); i++  {
            if i & 1 {
                Wlc_ObjSetCo( pNtk, pObj, 1 ); }
        }
            //else
            //    Wlc_ObjSetCi( pNtk, pObj );
        Vec_IntClear( &pNtk.vFfs );
        // convert init values into binary string
        //Vec_IntPrint( &p->pNtk->vInits );
        pNtk.pInits = Wlc_PrsConvertInitValues( pNtk );
        //printf( "%s", p->pNtk->pInits );
    }
    Vec_IntFree(vName2Obj);
    // create fake object names
    NameIdMax = Vec_IntFindMax(&pNtk.vNameIds);
    nDigits = Abc_Base10Log( NameIdMax+1 );
    pNtk.pManName = Abc_NamStart( NameIdMax+1, 10 );
    for  i = 1; i <= NameIdMax; i++  {
        var pName string // was [20]byte in C
 pName = fmt.Sprintf("s%0*d", nDigits, i);
        NameId = Abc_NamStrFindOrAdd( pNtk.pManName, pName, &fFound );
        assert(!(fFound != 0) && i == NameId, "!fFound && i == NameId");
    }
    //Ndr_NtkPrintNodes( pNtk );
    //Wlc_WriteVer( pNtk, "temp_ndr.v", 0, 0 );
    // derive topological order
    pNtk = Wlc_NtkDupDfs( pTemp = pNtk, 0, 1 );
    Wlc_NtkFree( pTemp );
    // copy truth tables
    if vTruths != nil {
        pNtk.vLutTruths = Vec_WrdStart( Wlc_NtkObjNumMax(pNtk) );
        // inlining of macro Wlc_NtkForEachObj
        for  i = 1; (i < Wlc_NtkObjNumMax(pNtk)) && (((pObj) = Wlc_NtkObj(pNtk, i)), 1); i++  {
            var iObj int32    = Wlc_ObjId(pNtk, pObj);
            var NameId int32  = Wlc_ObjNameId(pNtk, iObj);
            var Truth word;
            if pObj.Type != WLC_OBJ_LUT || NameId == 0 {
                continue; }
            Truth = Vec_WrdEntry(vTruths, NameId);
            assert(sizeof(void *) == 8 || Wlc_ObjFaninNum(pObj) < 6, "sizeof(void *) == 8 || Wlc_ObjFaninNum(pObj) < 6");
            Vec_WrdWriteEntry( pNtk.vLutTruths, iObj, Truth );
        }
        Vec_WrdFreeP( &vTruths );
    }
    //Ndr_NtkPrintNodes( pNtk );
    pNtk.fMemPorts = 1;          // the network contains memory ports
    pNtk.fEasyFfs = 1;           // the network contains simple flops
    return pNtk;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Wlc_ReadNdr(pFileName string) *Wlc_Ntk_t {
    var pData any  = Ndr_Read( pFileName );
    var pNtk *Wlc_Ntk_t  = Wlc_NtkFromNdr( pData );
    //char * ppNames[10] = { NULL, "a", "b", "c", "d", "e", "f", "g", "h", "i" };
    //Ndr_WriteVerilog( NULL, pData, ppNames );
    //Ndr_Delete( pData );
    Abc_FrameInputNdr( Abc_FrameGetGlobalFrame(), pData );
    return pNtk;
}
func Wlc_ReadNdrTest() {
    var pNtk *Wlc_Ntk_t  = Wlc_ReadNdr( "top.ndr" );
    Wlc_WriteVer( pNtk, "top.v", 0, 0 );
    Wlc_NtkFree( pNtk );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

