package wln

import "fmt"

/**CFile****************************************************************

  FileName    [wlnNdr.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Word-level network.]

  Synopsis    [Constructing WLN network from NDR data structure.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - September 23, 2018.]

  Revision    [$Id: wlnNdr.c,v 1.00 2018/09/23 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "wln.h"
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
func Wln_NtkToNdr(p *Wln_Ntk_t) any {
    var vFanins *Vec_Int_t;
    var i, k, iObj, iFanin int32;
    // create a new module
    var pDesign any  = Ndr_Create( 1 );
    var ModId int32  = Ndr_AddModule( pDesign, 1 );
    // add primary inputs
    // inlining of macro Wln_NtkForEachPi
    for  i = 0; (i < Wln_NtkPiNum(p)) && (((iObj) = Wln_NtkCi(p, i)), 1); i++  {
        Ndr_AddObject( pDesign, ModId, ABC_OPER_CI, 0,   
            Wln_ObjRangeEnd(p, iObj), Wln_ObjRangeBeg(p, iObj), Wln_ObjIsSigned(p, iObj),   
            0, nil,  1, &iObj,  nil  ); // no fanins
    }
    // add internal nodes 
    vFanins = Vec_IntAlloc( 10 );
    // inlining of macro Wln_NtkForEachObjInternal
    for  iObj = 1; iObj < Wln_NtkObjNum(p); iObj++  { if Wln_ObjIsCio(p, iObj) {} else {
        Vec_IntClear( vFanins );
        // inlining of macro Wln_ObjForEachFanin
        for  k = 0; (k < Wln_ObjFaninNum(p, iObj)) && (((iFanin) = Wln_ObjFanin(p, iObj, k)), 1); k++  { if !(iFanin != 0) {} else {
            Vec_IntPush( vFanins, iFanin ); }
        }
        Ndr_AddObject( pDesign, ModId, Wln_ObjType(p, iObj), 0,   
            Wln_ObjRangeEnd(p, iObj), Wln_ObjRangeBeg(p, iObj), Wln_ObjIsSigned(p, iObj),   
            Vec_IntSize(vFanins), Vec_IntArray(vFanins), 1, &iObj,  
            Wln_ObjIsConst(p, iObj) ? Wln_ObjConstString(p, iObj) : nil ); 
    }
    }
    Vec_IntFree( vFanins );
    // add primary outputs
    // inlining of macro Wln_NtkForEachPo
    for  i = 0; (i < Wln_NtkPoNum(p)) && (((iObj) = Wln_NtkCo(p, i)), 1); i++  {
        Ndr_AddObject( pDesign, ModId, ABC_OPER_CO, 0,                   
            Wln_ObjRangeEnd(p, iObj), Wln_ObjRangeBeg(p, iObj), Wln_ObjIsSigned(p, iObj),   
            1, &iObj,  0, nil,  nil ); 
    }
    return pDesign;
}
func Wln_WriteNdr(p *Wln_Ntk_t, pFileName string) {
    var pDesign any  = Wln_NtkToNdr( p );
    Ndr_Write( pFileName, pDesign );
    Ndr_Delete( pDesign );
    fmt.Printf("Dumped the current design into file \"%s\".\n", pFileName);
}
func Wln_NtkToNdrTest(p *Wln_Ntk_t) {
    // transform
    var pDesign any  = Wln_NtkToNdr( p );

    // collect names     
    // inlining of macro ABC_ALLOC
    var ppNames *string  = ((char * *) malloc(sizeof(char *) * (size_t)(Wln_NtkObjNum(p) + 1))); var i int32;
    // inlining of macro Wln_NtkForEachObj
    for  i = 1; i < Wln_NtkObjNum(p); i++  { 
        ppNames[i] = Abc_UtilStrsav(Wln_ObjName(p, i));
    }

    // verify by writing Verilog
    Ndr_WriteVerilog( nil, pDesign, ppNames );
    Ndr_Write( "test.ndr", pDesign );

    // cleanup
    Ndr_Delete( pDesign );
    // inlining of macro Wln_NtkForEachObj
    for  i = 1; i < Wln_NtkObjNum(p); i++  { 
        ppNames[i] = nil;
    }
    ppNames = nil;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ndr_ObjGetRange(p *Ndr_Data_t, Obj int32, End *int32, Beg *int32) int32 {
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
func Ndr_NtkPrintObjects(pNtk *Wln_Ntk_t) {
    var k, iObj, iFanin int32;
    fmt.Printf("Node IDs and their fanins:\n");
    // inlining of macro Wln_NtkForEachObj
    for  iObj = 1; iObj < Wln_NtkObjNum(pNtk); iObj++  {
        fmt.Printf("%5d = ", iObj);
        // inlining of macro Wln_ObjForEachFanin
        for  k = 0; (k < Wln_ObjFaninNum(pNtk, iObj)) && (((iFanin) = Wln_ObjFanin(pNtk, iObj, k)), 1); k++  { if !(iFanin != 0) {} else {
            fmt.Printf("%5d ", iFanin); }
        }
        for       ; k < 4; k++  {
            fmt.Printf("      ");
        }
        fmt.Printf("    Name Id %d ", Wln_ObjNameId(pNtk, iObj));
        if Wln_ObjIsPi(pNtk, iObj) {
            fmt.Printf("  pi  "); }
        if Wln_ObjIsPo(pNtk, iObj) {
            fmt.Printf("  po  "); }
        fmt.Printf("\n");
    }
}
func Wln_NtkCheckIntegrity(pData any) {
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
func Wln_NtkFromNdr(pData any, fDump int32) *Wln_Ntk_t {
    var p *Ndr_Data_t  = (Ndr_Data_t *)pData;  
    var vName2Obj *Vec_Int_t;
    var vFanins *Vec_Int_t = Vec_IntAlloc( 100 );
    var vConstStrings *Vec_Ptr_t  = Vec_PtrAlloc( 100 );
    var Mod int32 = 2;
    var i int32;
    var k int32;
    var iFanin int32;
    var iObj int32;
    var Obj int32;
    var pArray *int32;
    var nDigits int32;
    var fFound int32;
    var NameId int32;
    var NameIdMax int32;
    var pTemp *Wln_Ntk_t;
    var pNtk *Wln_Ntk_t = Wln_NtkAlloc( "top", Ndr_DataObjNum(p, Mod) );
    Wln_NtkCheckIntegrity( pData );
    //pNtk->pSpec = Abc_UtilStrsav( pFileName );
    // construct network and save name IDs
    Wln_NtkCleanNameId( pNtk );
    Wln_NtkCleanInstId( pNtk );
    // inlining of macro Ndr_ModForEachPi
    // inlining of macro Ndr_ModForEachObj
    for  Obj = Mod + 1; Obj < Ndr_DataEnd(p, Mod); Obj += Ndr_DataSize(p, Obj)  { if Ndr_DataType(p, Obj) != NDR_OBJECT {} else if !Ndr_ObjIsType(p, Obj, ABC_OPER_CI) {} else {
        var End int32;
        var Beg int32;
        var Signed int32 = Ndr_ObjGetRange(p, Obj, &End, &Beg);
        var iObj int32  = Wln_ObjAlloc( pNtk, ABC_OPER_CI, Signed, End, Beg );
        var NameId int32  = Ndr_ObjReadBody( p, Obj, NDR_OUTPUT );
        var InstId int32  = Ndr_ObjReadBody( p, Obj, NDR_NAME );
        Wln_ObjSetNameId( pNtk, iObj, NameId );
        if InstId > 0 { Wln_ObjSetInstId( pNtk, iObj, InstId ); }
    }
    }
    // inlining of macro Ndr_ModForEachNode
    // inlining of macro Ndr_ModForEachObj
    for  Obj = Mod + 1; Obj < Ndr_DataEnd(p, Mod); Obj += Ndr_DataSize(p, Obj)  { if Ndr_DataType(p, Obj) != NDR_OBJECT {} else if Ndr_ObjIsType(p, Obj, ABC_OPER_CI) || Ndr_ObjIsType(p, Obj, ABC_OPER_CO) {} else {
        var End int32;
        var Beg int32;
        var Signed int32 = Ndr_ObjGetRange(p, Obj, &End, &Beg);
        var Type int32     = Ndr_ObjReadBody( p, Obj, NDR_OPERTYPE );
        var nArray int32   = Ndr_ObjReadArray( p, Obj, NDR_INPUT, &pArray );
        var F Vec_Int_t = Vec_Int_t{nArray, nArray, pArray};
        var vTemp *Vec_Int_t = &F;
        var iObj int32     = Wln_ObjAlloc( pNtk, Type, Signed, End, Beg );
        var NameId int32   = Ndr_ObjReadBody( p, Obj, NDR_OUTPUT );
        var InstId int32   = Ndr_ObjReadBody( p, Obj, NDR_NAME );
        Vec_IntClear( vFanins );
        Vec_IntAppend( vFanins, vTemp );
        assert(Type != ABC_OPER_DFF, "Type != ABC_OPER_DFF");
        if Wln_ObjIsSlice(pNtk, iObj) {
            Wln_ObjSetSlice( pNtk, iObj, Hash_Int2ManInsert(pNtk.pRanges, End, Beg, 0) );
        } else if Wln_ObjIsConst(pNtk, iObj) {
            Vec_PtrPush( vConstStrings, (char *)Ndr_ObjReadBodyP(p, Obj, NDR_FUNCTION) ); }
//        else if ( Type == ABC_OPER_BIT_MUX && Vec_IntSize(vFanins) == 3 )
//            ABC_SWAP( int, Wln_ObjFanins(pNtk, iObj)[1], Wln_ObjFanins(pNtk, iObj)[2] );
        Wln_ObjAddFanins( pNtk, iObj, vFanins );
        Wln_ObjSetNameId( pNtk, iObj, NameId );
        if InstId > 0 { Wln_ObjSetInstId( pNtk, iObj, InstId ); }
        if Type == ABC_OPER_ARI_SMUL {
            assert(Wln_ObjFaninNum(pNtk, iObj) == 2, "Wln_ObjFaninNum(pNtk, iObj) == 2");
            Wln_ObjSetSigned( pNtk, Wln_ObjFanin0(pNtk, iObj) );
            Wln_ObjSetSigned( pNtk, Wln_ObjFanin1(pNtk, iObj) );
        }
    }
    }
    // mark primary outputs
    // inlining of macro Ndr_ModForEachPo
    // inlining of macro Ndr_ModForEachObj
    for  Obj = Mod + 1; Obj < Ndr_DataEnd(p, Mod); Obj += Ndr_DataSize(p, Obj)  { if Ndr_DataType(p, Obj) != NDR_OBJECT {} else if !Ndr_ObjIsType(p, Obj, ABC_OPER_CO) {} else {
        var End int32;
        var Beg int32;
        var Signed int32 = Ndr_ObjGetRange(p, Obj, &End, &Beg);
        var nArray int32   = Ndr_ObjReadArray( p, Obj, NDR_INPUT, &pArray );
        var iObj int32     = Wln_ObjAlloc( pNtk, ABC_OPER_CO, Signed, End, Beg );
        var NameId int32   = Ndr_ObjReadBody( p, Obj, NDR_OUTPUT );
        var InstId int32   = Ndr_ObjReadBody( p, Obj, NDR_NAME );
        assert(nArray == 1 && NameId == -1 && InstId == -1, "nArray == 1 && NameId == -1 && InstId == -1");
        Wln_ObjAddFanin( pNtk, iObj, pArray[0] );
    }
    }
    Vec_IntFree( vFanins );
    // remove instance names if they are not given
    //Vec_IntPrint( &pNtk->vInstIds );
    if Vec_IntCountPositive(&pNtk.vInstIds) == 0 {
        Vec_IntErase( &pNtk.vInstIds ); }
    // map name IDs into object IDs
    vName2Obj = Vec_IntInvert( &pNtk.vNameIds, 0 );
    // inlining of macro Wln_NtkForEachObj
    for  i = 1; i < Wln_NtkObjNum(pNtk); i++  {
        // inlining of macro Wln_ObjForEachFanin
        for  k = 0; (k < Wln_ObjFaninNum(pNtk, i)) && (((iFanin) = Wln_ObjFanin(pNtk, i, k)), 1); k++  { if !(iFanin != 0) {} else {
            Wln_ObjSetFanin( pNtk, i, k, Vec_IntEntry(vName2Obj, iFanin) ); }
        }
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
    // add const strings
    i = 0;
    // inlining of macro Wln_NtkForEachObj
    for  iObj = 1; iObj < Wln_NtkObjNum(pNtk); iObj++  {
        if Wln_ObjIsConst(pNtk, iObj) {
            Wln_ObjSetConst( pNtk, iObj, Abc_NamStrFindOrAdd(pNtk.pManName, (char *)Vec_PtrEntry(vConstStrings, i++), nil) ); }
    }
    assert(i == Vec_PtrSize(vConstStrings), "i == Vec_PtrSize(vConstStrings)");
    Vec_PtrFree( vConstStrings );
    //Ndr_NtkPrintObjects( pNtk );
    Wln_WriteVer( pNtk, "temp_ndr.v" );
    fmt.Printf("Dumped design \"%s\" into file \"temp_ndr.v\".\n", pNtk.pName);
    if !Wln_NtkIsAcyclic(pNtk) {
        Wln_NtkFree(pNtk);
        return nil;
    }
    // derive topological order
    pNtk = Wln_NtkDupDfs( pTemp = pNtk );
    Wln_NtkFree( pTemp );
    //Ndr_NtkPrintObjects( pNtk );
    //pNtk->fMemPorts = 1;          // the network contains memory ports
    //pNtk->fEasyFfs = 1;           // the network contains simple flops
    return pNtk;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Wln_ReadNdr(pFileName string) *Wln_Ntk_t {
    var pData any  = Ndr_Read( pFileName );
    var pNtk *Wln_Ntk_t;
    if pData {
        pNtk = Wln_NtkFromNdr( pData, 0 );
    } else {
        pNtk = nil;
    }
    if pNtk != nil { return nil; }
    //char * ppNames[10] = { NULL, "a", "b", "c", "d", "e", "f", "g", "h", "i" };
    //Ndr_WriteVerilog( NULL, pData, ppNames );
    Ndr_Delete( pData );
    return pNtk;
}
func Wln_ReadNdrTest() {
    var pNtk *Wln_Ntk_t  = Wln_ReadNdr( "D:\\temp\\brijesh\\for_alan_dff_warning\\work_fir_filter_fir_filter_proc_out.ndr" );
    //Wln_Ntk_t * pNtk = Wln_ReadNdr( "flop.ndr" );
    Wln_WriteVer( pNtk, "test__test.v" );
    Wln_NtkPrint( pNtk );
    Wln_NtkStaticFanoutTest( pNtk );
    Wln_NtkFree( pNtk );
}
func Wln_NtkRetimeTest(pFileName string, fSkipSimple int32, fDump int32, fVerbose int32) {
    var vMoves *Vec_Int_t;
    var pData any  = Ndr_Read( pFileName );
    var pNtk *Wln_Ntk_t;
    if pData {
        pNtk = Wln_NtkFromNdr( pData, fDump );
    } else {
        pNtk = nil;
    }
    Ndr_Delete( pData );
    if pNtk == nil {
        fmt.Printf("Retiming network is not available.\n");
        return;
    }
    Wln_NtkRetimeCreateDelayInfo( pNtk );
    vMoves = Wln_NtkRetime( pNtk, fSkipSimple, fVerbose );
    //Vec_IntPrint( vMoves );
    Vec_IntFree( vMoves );
    Wln_NtkFree( pNtk );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

