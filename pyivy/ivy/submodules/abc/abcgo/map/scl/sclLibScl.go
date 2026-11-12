package scl

import "fmt"

import "os"

/**CFile****************************************************************

  FileName    [sclLibScl.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Standard-cell library representation.]

  Synopsis    [Liberty abstraction for delay-oriented mapping.]

  Author      [Alan Mishchenko, Niklas Een]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - August 24, 2012.]

  Revision    [$Id: sclLibScl.c,v 1.0 2012/08/24 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "sclLib.h"
//#include "misc/st/st.h"
//#include "map/mio/mio.h"
//#include "bool/kit/kit.h"
//#include "misc/extra/extra.h"
//#include "misc/util/utilNam.h"
//#include "map/scl/sclCon.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////


/**Function*************************************************************

  Synopsis    [Reading library from file.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SclReadSurfaceGenlib(p *SC_Surface) {
    var vVec *Vec_Flt_t;
    var vVecI *Vec_Int_t;
    var i, j int32;

    Vec_FltPush( &p.vIndex0, 0 );
    Vec_IntPush( &p.vIndex0I, Scl_Flt2Int(0) );

    Vec_FltPush( &p.vIndex1, 0 );
    Vec_IntPush( &p.vIndex1I, Scl_Flt2Int(0) );

    for  i = 0; i < Vec_FltSize(&p.vIndex0); i++  {
        vVec = Vec_FltAlloc( Vec_FltSize(&p.vIndex1) );
        Vec_PtrPush( &p.vData, vVec );
        vVecI = Vec_IntAlloc( Vec_FltSize(&p.vIndex1) );
        Vec_PtrPush( &p.vDataI, vVecI );
        for  j = 0; j < Vec_FltSize(&p.vIndex1); j++  {
            Vec_FltPush( vVec, 1 );
            Vec_IntPush( vVecI, Scl_Flt2Int(1) );
        }
    }
}
func Abc_SclReadLibraryGenlib(p *SC_Lib, pLib *Mio_Library_t) int32 {
    var pGate *Mio_Gate_t;
    var pGatePin *Mio_Pin_t;
    var j, k int32;

    // Read non-composite fields:
    p.pName                 = Abc_UtilStrsav( Mio_LibraryReadName(pLib) );
    p.default_wire_load     = 0;
    p.default_wire_load_sel = 0;
    p.default_max_out_slew  = 0;

    p.unit_time             = 12;
    p.unit_cap_fst          = 1.0;
    p.unit_cap_snd          = 15;

    Mio_LibraryForEachGate( pLib, pGate )
    {
        var pCell *SC_Cell  = Abc_SclCellAlloc();
        pCell.Id = SC_LibCellNum(p);
        Vec_PtrPush( &p.vCells, pCell );

        pCell.pName          = Abc_UtilStrsav( Mio_GateReadName( pGate ) );     
        pCell.area           = Mio_GateReadArea( pGate );
        pCell.leakage        = 0;
        pCell.drive_strength = 0;

        pCell.n_inputs       = Mio_GateReadPinNum( pGate );    
        pCell.n_outputs      = 1;

        pCell.areaI          = Scl_Flt2Int(pCell.area);
        pCell.leakageI       = Scl_Flt2Int(pCell.leakage);

        Mio_GateForEachPin( pGate, pGatePin )
        {
            var pPin *SC_Pin  = Abc_SclPinAlloc();
            Vec_PtrPush( &pCell.vPins, pPin );

            pPin.dir      = sc_dir_Input;
            pPin.pName    = Abc_UtilStrsav( Mio_PinReadName( pGatePin ) ); 
            pPin.rise_cap = 0;
            pPin.fall_cap = 0;

            pPin.rise_capI = Scl_Flt2Int(pPin.rise_cap);
            pPin.fall_capI = Scl_Flt2Int(pPin.fall_cap);
        }

        for  j = 0; j < pCell.n_outputs; j++  {
            var pTruth *word  = Mio_GateReadTruthP( pGate );
            var pPin *SC_Pin  = Abc_SclPinAlloc();
            Vec_PtrPush( &pCell.vPins, pPin );

            pPin.dir          = sc_dir_Output;
            pPin.pName        = Abc_UtilStrsav( Mio_GateReadOutName( pGate ) );
            pPin.max_out_cap  = 0;
            pPin.max_out_slew = 0;

            // read function
            pPin.func_text = Abc_UtilStrsav( Mio_GateReadForm( pGate ) );
            Vec_WrdGrow( &pPin.vFunc, Abc_Truth6WordNum(pCell.n_inputs) );
            for  k = 0; k < Vec_WrdCap(&pPin.vFunc); k++  {
                Vec_WrdPush( &pPin.vFunc, pTruth[k] );
            }

            // read pins
            Mio_GateForEachPin( pGate, pGatePin )
            {
                var Value Mio_PinPhase_t  = Mio_PinReadPhase( pGatePin );
                var pRTime *SC_Timings  = Abc_SclTimingsAlloc();
                Vec_PtrPush( &pPin.vRTimings, pRTime );
                pRTime.pName = Abc_UtilStrsav( Mio_PinReadName( pGatePin ) ); 
                if true {
                    var pTime *SC_Timing  = Abc_SclTimingAlloc();
                    Vec_PtrPush( &pRTime.vTimings, pTime );
                    if Value == MIO_PHASE_UNKNOWN {
                        pTime.tsense = (SC_TSense)sc_ts_Non;
                    } else if Value == MIO_PHASE_INV {
                        pTime.tsense = (SC_TSense)sc_ts_Neg;
                    } else if Value == MIO_PHASE_NONINV {
                        pTime.tsense = (SC_TSense)sc_ts_Pos;
                    } else { assert(false, "0"); }
                    Abc_SclReadSurfaceGenlib( &pTime.pCellRise );
                    Abc_SclReadSurfaceGenlib( &pTime.pCellFall );
                    Abc_SclReadSurfaceGenlib( &pTime.pRiseTrans );
                    Abc_SclReadSurfaceGenlib( &pTime.pFallTrans );
                }
            }
        }
    }
    return 1;
}
func Abc_SclReadFromGenlib(pLib0 any) *SC_Lib {
    var pLib *Mio_Library_t  = (Mio_Library_t *)pLib0;
    var p *SC_Lib  = Abc_SclLibAlloc();
    if !(Abc_SclReadLibraryGenlib( p, pLib ) != 0) {
        return nil; }
    // hash gates by name
    Abc_SclHashCells( p );
    Abc_SclLinkCells( p );
    return p;
}


/**Function*************************************************************

  Synopsis    [Reading library from file.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SclReadSurface(vOut *Vec_Str_t, pPos *int32, p *SC_Surface) {
    var vVec *Vec_Flt_t;
    var vVecI *Vec_Int_t;
    var i, j int32;

    for  i = Vec_StrGetI(vOut, pPos); i != 0; i--  {
        var Num float32  = Vec_StrGetF(vOut, pPos);
        Vec_FltPush( &p.vIndex0, Num );
        Vec_IntPush( &p.vIndex0I, Scl_Flt2Int(Num) );
    }

    for  i = Vec_StrGetI(vOut, pPos); i != 0; i--  {
        var Num float32  = Vec_StrGetF(vOut, pPos);
        Vec_FltPush( &p.vIndex1, Num );
        Vec_IntPush( &p.vIndex1I, Scl_Flt2Int(Num) );
    }

    for  i = 0; i < Vec_FltSize(&p.vIndex0); i++  {
        vVec = Vec_FltAlloc( Vec_FltSize(&p.vIndex1) );
        Vec_PtrPush( &p.vData, vVec );
        vVecI = Vec_IntAlloc( Vec_FltSize(&p.vIndex1) );
        Vec_PtrPush( &p.vDataI, vVecI );
        for  j = 0; j < Vec_FltSize(&p.vIndex1); j++  {
            var Num float32  = Vec_StrGetF(vOut, pPos);
            Vec_FltPush( vVec, Num );
            Vec_IntPush( vVecI, Scl_Flt2Int(Num) );
        }
    }

    for  i = 0; i < 3; i++  { 
        p.approx[0][i] = Vec_StrGetF( vOut, pPos );
    }
    for  i = 0; i < 4; i++  { 
        p.approx[1][i] = Vec_StrGetF( vOut, pPos );
    }
    for  i = 0; i < 6; i++  { 
        p.approx[2][i] = Vec_StrGetF( vOut, pPos );
    }
}
func Abc_SclReadLibrary(vOut *Vec_Str_t, pPos *int32, p *SC_Lib) int32 {
    var i, j, k, n int32;
    var version int32  = Vec_StrGetI( vOut, pPos );
    if version != ABC_SCL_CUR_VERSION { 
        Abc_Print( -1, "Wrong version of the SCL file.\n" ); 
        return 0; 
    }
    assert(version == ABC_SCL_CUR_VERSION, "version == ABC_SCL_CUR_VERSION"); // wrong version of the file

    // Read non-composite fields:
    p.pName                 = Vec_StrGetS(vOut, pPos);
    p.default_wire_load     = Vec_StrGetS(vOut, pPos);
    p.default_wire_load_sel = Vec_StrGetS(vOut, pPos);
    p.default_max_out_slew  = Vec_StrGetF(vOut, pPos);

    p.unit_time             = Vec_StrGetI(vOut, pPos);
    p.unit_cap_fst          = Vec_StrGetF(vOut, pPos);
    p.unit_cap_snd          = Vec_StrGetI(vOut, pPos);

    // Read 'wire_load' vector:
    for  i = Vec_StrGetI(vOut, pPos); i != 0; i--  {
        var pWL *SC_WireLoad  = Abc_SclWireLoadAlloc();
        Vec_PtrPush( &p.vWireLoads, pWL );

        pWL.pName = Vec_StrGetS(vOut, pPos);
        pWL.cap   = Vec_StrGetF(vOut, pPos);
        pWL.slope = Vec_StrGetF(vOut, pPos);

        for  j = Vec_StrGetI(vOut, pPos); j != 0; j--  {
            Vec_IntPush( &pWL.vFanout, Vec_StrGetI(vOut, pPos) );
            Vec_FltPush( &pWL.vLen,    Vec_StrGetF(vOut, pPos) );
        }
    }

    // Read 'wire_load_sel' vector:
    for  i = Vec_StrGetI(vOut, pPos); i != 0; i--  {
        var pWLS *SC_WireLoadSel  = Abc_SclWireLoadSelAlloc();
        Vec_PtrPush( &p.vWireLoadSels, pWLS );

        pWLS.pName = Vec_StrGetS(vOut, pPos);
        for  j = Vec_StrGetI(vOut, pPos); j != 0; j--  {
            Vec_FltPush( &pWLS.vAreaFrom,      Vec_StrGetF(vOut, pPos) );
            Vec_FltPush( &pWLS.vAreaTo,        Vec_StrGetF(vOut, pPos) );
            Vec_PtrPush( &pWLS.vWireLoadModel, Vec_StrGetS(vOut, pPos) );
        }
    }

    for  i = Vec_StrGetI(vOut, pPos); i != 0; i--  {
        var pCell *SC_Cell  = Abc_SclCellAlloc();
        pCell.Id = SC_LibCellNum(p);
        Vec_PtrPush( &p.vCells, pCell );

        pCell.pName          = Vec_StrGetS(vOut, pPos);     
        pCell.area           = Vec_StrGetF(vOut, pPos);
        pCell.leakage        = Vec_StrGetF(vOut, pPos);
        pCell.drive_strength = Vec_StrGetI(vOut, pPos);

        pCell.n_inputs       = Vec_StrGetI(vOut, pPos);
        pCell.n_outputs      = Vec_StrGetI(vOut, pPos);

        pCell.areaI          = Scl_Flt2Int(pCell.area);
        pCell.leakageI       = Scl_Flt2Int(pCell.leakage);
/*
        printf( "%s\n", pCell->pName );
        if ( !strcmp( "XOR3_X4M_A9TL", pCell->pName ) )
        {
            int s = 0;
        }
*/
        for  j = 0; j < pCell.n_inputs; j++  {
            var pPin *SC_Pin  = Abc_SclPinAlloc();
            Vec_PtrPush( &pCell.vPins, pPin );

            pPin.dir      = sc_dir_Input;
            pPin.pName    = Vec_StrGetS(vOut, pPos); 
            pPin.rise_cap = Vec_StrGetF(vOut, pPos);
            pPin.fall_cap = Vec_StrGetF(vOut, pPos);

            pPin.rise_capI = Scl_Flt2Int(pPin.rise_cap);
            pPin.fall_capI = Scl_Flt2Int(pPin.fall_cap);
        }

        for  j = 0; j < pCell.n_outputs; j++  {
            var pPin *SC_Pin  = Abc_SclPinAlloc();
            Vec_PtrPush( &pCell.vPins, pPin );

            pPin.dir          = sc_dir_Output;
            pPin.pName        = Vec_StrGetS(vOut, pPos); 
            pPin.max_out_cap  = Vec_StrGetF(vOut, pPos);
            pPin.max_out_slew = Vec_StrGetF(vOut, pPos);

            k = Vec_StrGetI(vOut, pPos);
            assert(k == pCell.n_inputs, "k == pCell.n_inputs");

            // read function
            // (possibly empty) formula is always given
            assert(version == ABC_SCL_CUR_VERSION, "version == ABC_SCL_CUR_VERSION");
            assert(pPin.func_text == nil, "pPin.func_text == nil");
            pPin.func_text = Vec_StrGetS(vOut, pPos); 
            if pPin.func_text[0] == 0 {
                // formula is not given - read truth table
                ABC_FREE( pPin.func_text );
                assert(Vec_WrdSize(&pPin.vFunc) == 0, "Vec_WrdSize(&pPin.vFunc) == 0");
                Vec_WrdGrow( &pPin.vFunc, Abc_Truth6WordNum(pCell.n_inputs) );
                for  k = 0; k < Vec_WrdCap(&pPin.vFunc); k++  {
                    Vec_WrdPush( &pPin.vFunc, Vec_StrGetW(vOut, pPos) );
                }
            
            } else {
                // formula is given - derive truth table
                var pPin2 *SC_Pin;
                var vNames *Vec_Ptr_t;
                var vFunc *Vec_Wrd_t;
                // collect input names
                vNames = Vec_PtrAlloc( pCell.n_inputs );
                SC_CellForEachPinIn( pCell, pPin2, n )
                    Vec_PtrPush( vNames, pPin2.pName );
                // derive truth table
                assert(Vec_WrdSize(&pPin.vFunc) == 0, "Vec_WrdSize(&pPin.vFunc) == 0");
                Vec_WrdErase( &pPin.vFunc );
                vFunc = Mio_ParseFormulaTruth( pPin.func_text, (char **)Vec_PtrArray(vNames), pCell.n_inputs );
                pPin.vFunc = *vFunc;
                ABC_FREE( vFunc );
                Vec_PtrFree( vNames );
                // skip truth table
                assert(Vec_WrdSize(&pPin.vFunc) == Abc_Truth6WordNum(pCell.n_inputs), "Vec_WrdSize(&pPin.vFunc) == Abc_Truth6WordNum(pCell.n_inputs)");
                for  k = 0; k < Vec_WrdSize(&pPin.vFunc); k++  {
                    var Value word  = Vec_StrGetW(vOut, pPos);
                    assert(Value == Vec_WrdEntry(&pPin.vFunc, k), "Value == Vec_WrdEntry(&pPin.vFunc, k)");
                }
            }

            // Read 'rtiming': (pin-to-pin timing tables for this particular output)
            for  k = 0; k < pCell.n_inputs; k++  {
                var pRTime *SC_Timings  = Abc_SclTimingsAlloc();
                Vec_PtrPush( &pPin.vRTimings, pRTime );

                pRTime.pName = Vec_StrGetS(vOut, pPos);
                n = Vec_StrGetI(vOut, pPos); assert(n <= 1, "n <= 1");
                if n == 1 {
                    var pTime *SC_Timing  = Abc_SclTimingAlloc();
                    Vec_PtrPush( &pRTime.vTimings, pTime );

                    pTime.tsense = (SC_TSense)Vec_StrGetI(vOut, pPos);
                    Abc_SclReadSurface( vOut, pPos, &pTime.pCellRise );
                    Abc_SclReadSurface( vOut, pPos, &pTime.pCellFall );
                    Abc_SclReadSurface( vOut, pPos, &pTime.pRiseTrans );
                    Abc_SclReadSurface( vOut, pPos, &pTime.pFallTrans );
                
                } else {
                    assert(Vec_PtrSize(&pRTime.vTimings) == 0, "Vec_PtrSize(&pRTime.vTimings) == 0"); }
            }
        }
    }
    return 1;
}
func Abc_SclReadFromStr(vOut *Vec_Str_t) *SC_Lib {
    var p *SC_Lib;
    var Pos int32;
    // read the library
    p = Abc_SclLibAlloc();
    if !(Abc_SclReadLibrary( vOut, &Pos, p ) != 0) {
        return nil; }
    assert(Pos == Vec_StrSize(vOut), "Pos == Vec_StrSize(vOut)");
    // hash gates by name
    Abc_SclHashCells( p );
    Abc_SclLinkCells( p );
    return p;
}
func Abc_SclReadFromFile(pFileName string) *SC_Lib {
    var p *SC_Lib;
    var pFile *os.File;
    var vOut *Vec_Str_t;
    var nFileSize int32;
    pFile = os.Open(pFileName);
    if pFile == nil {
        fmt.Printf("Cannot open file \"%s\" for reading.\n", pFileName);
        return nil;
    }
    // get the file size, in bytes
    fseek( pFile, 0, SEEK_END );  
    nFileSize = ftell( pFile );  
    rewind( pFile ); 
    // load the contents
    vOut = Vec_StrAlloc( nFileSize );
    vOut.nSize = vOut.nCap;
    assert(nFileSize == Vec_StrSize(vOut), "nFileSize == Vec_StrSize(vOut)");
    nFileSize = fread( Vec_StrArray(vOut), 1, Vec_StrSize(vOut), pFile );
    assert(nFileSize == Vec_StrSize(vOut), "nFileSize == Vec_StrSize(vOut)");
    _ = (pFile).Close();
    // read the library
    p = Abc_SclReadFromStr( vOut );
    if p != nil {
        p.pFileName = Abc_UtilStrsav( pFileName ); }
    if p != nil {
        Abc_SclLibNormalize( p ); }
    Vec_StrFree( vOut );
    return p;
}

/**Function*************************************************************

  Synopsis    [Writing library into file.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SclWriteSurface(vOut *Vec_Str_t, p *SC_Surface) {
    var vVec *Vec_Flt_t;
    var Entry float32;
    var i, k int32;

    Vec_StrPutI( vOut, Vec_FltSize(&p.vIndex0) );
    Vec_FltForEachEntry( &p.vIndex0, Entry, i )
        Vec_StrPutF( vOut, Entry );

    Vec_StrPutI( vOut, Vec_FltSize(&p.vIndex1) );
    Vec_FltForEachEntry( &p.vIndex1, Entry, i )
        Vec_StrPutF( vOut, Entry );

    Vec_PtrForEachEntry( Vec_Flt_t *, &p.vData, vVec, i )
        Vec_FltForEachEntry( vVec, Entry, k )
            Vec_StrPutF( vOut, Entry );

    for  i = 0; i < 3; i++  { 
        Vec_StrPutF( vOut, p.approx[0][i] );
    }
    for  i = 0; i < 4; i++  { 
        Vec_StrPutF( vOut, p.approx[1][i] );
    }
    for  i = 0; i < 6; i++  { 
        Vec_StrPutF( vOut, p.approx[2][i] );
    }
}
func Abc_SclWriteLibrary(vOut *Vec_Str_t, p *SC_Lib) {
    var pWL *SC_WireLoad;
    var pWLS *SC_WireLoadSel;
    var pCell *SC_Cell;
    var pPin *SC_Pin;
    var n_valid_cells int32;
    var i, j, k int32;

    Vec_StrPutI( vOut, ABC_SCL_CUR_VERSION );

    // Write non-composite fields:
    Vec_StrPutS( vOut, p.pName );
    Vec_StrPutS( vOut, p.default_wire_load );
    Vec_StrPutS( vOut, p.default_wire_load_sel );
    Vec_StrPutF( vOut, p.default_max_out_slew );

    assert(p.unit_time >= 0, "p.unit_time >= 0");
    assert(p.unit_cap_snd >= 0, "p.unit_cap_snd >= 0");
    Vec_StrPutI( vOut, p.unit_time );
    Vec_StrPutF( vOut, p.unit_cap_fst );
    Vec_StrPutI( vOut, p.unit_cap_snd );

    // Write 'wire_load' vector:
    Vec_StrPutI( vOut, Vec_PtrSize(&p.vWireLoads) );
    SC_LibForEachWireLoad( p, pWL, i )
    {
        Vec_StrPutS( vOut, pWL.pName );
        Vec_StrPutF( vOut, pWL.cap );
        Vec_StrPutF( vOut, pWL.slope );

        Vec_StrPutI( vOut, Vec_IntSize(&pWL.vFanout) );
        for  j = 0; j < Vec_IntSize(&pWL.vFanout); j++  {
            Vec_StrPutI( vOut, Vec_IntEntry(&pWL.vFanout, j) );
            Vec_StrPutF( vOut, Vec_FltEntry(&pWL.vLen, j) );
        }
    }

    // Write 'wire_load_sel' vector:
    Vec_StrPutI( vOut, Vec_PtrSize(&p.vWireLoadSels) );
    SC_LibForEachWireLoadSel( p, pWLS, i )
    {
        Vec_StrPutS( vOut, pWLS.pName );
        Vec_StrPutI( vOut, Vec_FltSize(&pWLS.vAreaFrom) );
        for  j = 0; j < Vec_FltSize(&pWLS.vAreaFrom); j++ {
            Vec_StrPutF( vOut, Vec_FltEntry(&pWLS.vAreaFrom, j) );
            Vec_StrPutF( vOut, Vec_FltEntry(&pWLS.vAreaTo, j) );
            Vec_StrPutS( vOut, (char *)Vec_PtrEntry(&pWLS.vWireLoadModel, j) );
        }
    }

    // Write 'cells' vector:
    n_valid_cells = 0;
    SC_LibForEachCell( p, pCell, i )
        if ( !(pCell.seq || pCell.unsupp) )
            n_valid_cells++;

    Vec_StrPutI( vOut, n_valid_cells );
    SC_LibForEachCell( p, pCell, i )
    {
        if pCell.seq || pCell.unsupp {
            continue; }

        Vec_StrPutS( vOut, pCell.pName );
        Vec_StrPutF( vOut, pCell.area );
        Vec_StrPutF( vOut, pCell.leakage );
        Vec_StrPutI( vOut, pCell.drive_strength );

        // Write 'pins': (sorted at this point; first inputs, then outputs)
        Vec_StrPutI( vOut, pCell.n_inputs);
        Vec_StrPutI( vOut, pCell.n_outputs);

        SC_CellForEachPinIn( pCell, pPin, j )
        {
            assert(pPin.dir == sc_dir_Input, "pPin.dir == sc_dir_Input");
            Vec_StrPutS( vOut, pPin.pName );
            Vec_StrPutF( vOut, pPin.rise_cap );
            Vec_StrPutF( vOut, pPin.fall_cap );
        }

        SC_CellForEachPinOut( pCell, pPin, j )
        {
            var pRTime *SC_Timings;
            var uWord word;

            assert(pPin.dir == sc_dir_Output, "pPin.dir == sc_dir_Output");
            Vec_StrPutS( vOut, pPin.pName );
            Vec_StrPutF( vOut, pPin.max_out_cap );
            Vec_StrPutF( vOut, pPin.max_out_slew );
            Vec_StrPutI( vOut, pCell.n_inputs );

            // write function
            Vec_StrPutS( vOut, pPin.func_text ? pPin.func_text : (char *)"" );

            // write truth table
            assert(Vec_WrdSize(&pPin.vFunc) == Abc_Truth6WordNum(pCell.n_inputs), "Vec_WrdSize(&pPin.vFunc) == Abc_Truth6WordNum(pCell.n_inputs)");
            Vec_WrdForEachEntry( &pPin.vFunc, uWord, k ) // -- 'size = 1u << (n_vars - 6)'
                Vec_StrPutW( vOut, uWord );  // -- 64-bit number, written uncompressed (low-byte first)

            // Write 'rtiming': (pin-to-pin timing tables for this particular output)
            assert(Vec_PtrSize(&pPin.vRTimings) == pCell.n_inputs, "Vec_PtrSize(&pPin.vRTimings) == pCell.n_inputs");
            SC_PinForEachRTiming( pPin, pRTime, k )
            {
                Vec_StrPutS( vOut, pRTime.pName );
                Vec_StrPutI( vOut, Vec_PtrSize(&pRTime.vTimings) );
                    // -- NOTE! After post-processing, the size of the 'rtiming[k]' vector is either
                    // 0 or 1 (in static timing, we have merged all tables to get the worst case).
                    // The case with size 0 should only occur for multi-output gates.
                if Vec_PtrSize(&pRTime.vTimings) == 1 {
                    var pTime *SC_Timing  = (SC_Timing *)Vec_PtrEntry( &pRTime.vTimings, 0 );
                        // -- NOTE! We don't need to save 'related_pin' string because we have sorted 
                        // the elements on input pins.
                    Vec_StrPutI( vOut, int32(pTime.tsense));
                    Abc_SclWriteSurface( vOut, &pTime.pCellRise );
                    Abc_SclWriteSurface( vOut, &pTime.pCellFall );
                    Abc_SclWriteSurface( vOut, &pTime.pRiseTrans );
                    Abc_SclWriteSurface( vOut, &pTime.pFallTrans );
                
                } else {
                    assert(Vec_PtrSize(&pRTime.vTimings) == 0, "Vec_PtrSize(&pRTime.vTimings) == 0"); }
            }
        }
    }
}
func Abc_SclWriteScl(pFileName string, p *SC_Lib) {
    var vOut *Vec_Str_t;
    vOut = Vec_StrAlloc( 10000 );
    Abc_SclWriteLibrary( vOut, p );
    if Vec_StrSize(vOut) > 0 {
        var pFile *os.File  = os.OpenFile(pFileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
        if pFile == nil {
            fmt.Printf("Cannot open file \"%s\" for writing.\n", pFileName);
        } else {
            fwrite( Vec_StrArray(vOut), 1, Vec_StrSize(vOut), pFile );
            _ = (pFile).Close();
        }
    }
    Vec_StrFree( vOut );    
}


/**Function*************************************************************

  Synopsis    [Writing library into text file.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SclWriteSurfaceText(s *os.File, p *SC_Surface) {
    var vVec *Vec_Flt_t;
    var Entry float32;
    var i, k int32;

    fmt.Fprintf(s, "          index_1(\"");  
    Vec_FltForEachEntry( &p.vIndex0, Entry, i )
        fmt.Fprintf(s, "%f%s", Entry, i == Vec_FltSize(&p.vIndex0)-1 ? "":", ");
    fmt.Fprintf(s, "\");\n");

    fmt.Fprintf(s, "          index_2(\"");  
    Vec_FltForEachEntry( &p.vIndex1, Entry, i )
        fmt.Fprintf(s, "%f%s", Entry, i == Vec_FltSize(&p.vIndex1)-1 ? "":", ");
    fmt.Fprintf(s, "\");\n");

    fmt.Fprintf(s, "          values (\"");  
    Vec_PtrForEachEntry( Vec_Flt_t *, &p.vData, vVec, i )
    {
        Vec_FltForEachEntry( vVec, Entry, k )
            fmt.Fprintf(s, "%f%s", Entry, i == Vec_PtrSize(&p.vData)-1 && k == Vec_FltSize(vVec)-1 ? "\");":", ");
        if i == Vec_PtrSize(&p.vData)-1 {
            fmt.Fprintf(s, "\n");
        } else {
            fmt.Fprintf(s, "\\\n");
            fmt.Fprintf(s, "                   ");  
        }
    }
/*
    fprintf( s, "       approximations: \n" );
    fprintf( s, "       " );
    for ( i = 0; i < 3; i++ ) 
        fprintf( s, "%f ", p->approx[0][i] );
    fprintf( s, "\n" );
    fprintf( s, "       " );
    for ( i = 0; i < 4; i++ ) 
        fprintf( s, "%f ", p->approx[1][i] );
    fprintf( s, "\n" );
    fprintf( s, "       " );
    for ( i = 0; i < 6; i++ ) 
        fprintf( s, "%f ", p->approx[2][i] );
    fprintf( s, "\n" );
    fprintf( s, "       \n" );
*/
}
func Abc_SclWriteLibraryText(s *os.File, p *SC_Lib) {
    var pWL *SC_WireLoad;
    var pWLS *SC_WireLoadSel;
    var pCell *SC_Cell;
    var pPin *SC_Pin;
    var n_valid_cells int32;
    var i, j, k int32;
    fmt.Fprintf(s, "/* This Liberty file was generated by ABC on %s */\n", Extra_TimeStamp());
    fmt.Fprintf(s, "/* The original unabridged library came from file \"%s\".*/\n\n", p.pFileName);

//    fprintf( s, "%d", ABC_SCL_CUR_VERSION );
    fmt.Fprintf(s, "library(%s) {\n\n", p.pName);
    if p.default_wire_load && len(p.default_wire_load) {
    fmt.Fprintf(s, "  default_wire_load : \"%s\";\n", p.default_wire_load); }
    if p.default_wire_load_sel && len(p.default_wire_load_sel) {
    fmt.Fprintf(s, "  default_wire_load_selection : \"%s\";\n", p.default_wire_load_sel); }
    if p.default_max_out_slew != -1 {
    fmt.Fprintf(s, "  default_max_transition : %f;\n", p.default_max_out_slew); }
    if p.unit_time == 9 {
    fmt.Fprintf(s, "  time_unit : \"1ns\";\n");
    } else if p.unit_time == 10 {
    fmt.Fprintf(s, "  time_unit : \"100ps\";\n");
    } else if p.unit_time == 11 {
    fmt.Fprintf(s, "  time_unit : \"10ps\";\n");
    } else if p.unit_time == 12 {
    fmt.Fprintf(s, "  time_unit : \"1ps\";\n");
    } else { assert(false, "0"); }
    fmt.Fprintf(s, "  capacitive_load_unit(%.1f,%s);\n", p.unit_cap_fst, p.unit_cap_snd == 12 ? "pf" : "ff");
    fmt.Fprintf(s, "\n");

    // Write 'wire_load' vector:
    SC_LibForEachWireLoad( p, pWL, i )
    {
        fmt.Fprintf(s, "  wire_load(\"%s\") {\n", pWL.pName);
        fmt.Fprintf(s, "    capacitance : %f;\n", pWL.cap);
        fmt.Fprintf(s, "    slope : %f;\n", pWL.slope);
        for  j = 0; j < Vec_IntSize(&pWL.vFanout); j++  {
            fmt.Fprintf(s, "    fanout_length( %d, %f );\n", Vec_IntEntry(&pWL.vFanout, j), Vec_FltEntry(&pWL.vLen, j));
        }
        fmt.Fprintf(s, "  }\n\n");
    }

    // Write 'wire_load_sel' vector:
    SC_LibForEachWireLoadSel( p, pWLS, i )
    {
        fmt.Fprintf(s, "  wire_load_selection(\"%s\") {\n", pWLS.pName);
        for  j = 0; j < Vec_FltSize(&pWLS.vAreaFrom); j++ {
            fmt.Fprintf(s, "    wire_load_from_area( %f, %f, %s );\n", Vec_FltEntry(&pWLS.vAreaFrom, j), Vec_FltEntry(&pWLS.vAreaTo, j), (char *)Vec_PtrEntry(&pWLS.vWireLoadModel, j));
        }
        fmt.Fprintf(s, "  }\n\n");
    }

    // Write 'cells' vector:
    n_valid_cells = 0;
    SC_LibForEachCell( p, pCell, i )
        if ( !(pCell.seq || pCell.unsupp) )
            n_valid_cells++;

    SC_LibForEachCell( p, pCell, i )
    {
        if pCell.seq || pCell.unsupp {
            continue; }

        fmt.Fprintf(s, "\n");
        fmt.Fprintf(s, "  cell(%s) {\n", pCell.pName);
        fmt.Fprintf(s, "    /*  n_inputs = %d  n_outputs = %d */\n", pCell.n_inputs, pCell.n_outputs);
        fmt.Fprintf(s, "    area : %f;\n", pCell.area);
        fmt.Fprintf(s, "    cell_leakage_power : %f;\n", pCell.leakage);
        fmt.Fprintf(s, "    drive_strength : %d;\n", pCell.drive_strength);

        SC_CellForEachPinIn( pCell, pPin, j )
        {
            assert(pPin.dir == sc_dir_Input, "pPin.dir == sc_dir_Input");
            fmt.Fprintf(s, "    pin(%s) {\n", pPin.pName);
            fmt.Fprintf(s, "      direction : %s;\n", "input");
            fmt.Fprintf(s, "      fall_capacitance : %f;\n", pPin.fall_cap);
            fmt.Fprintf(s, "      rise_capacitance : %f;\n", pPin.rise_cap);
            fmt.Fprintf(s, "    }\n");
        }

        SC_CellForEachPinOut( pCell, pPin, j )
        {
            var pRTime *SC_Timings;
//            word uWord;
            assert(pPin.dir == sc_dir_Output, "pPin.dir == sc_dir_Output");
            fmt.Fprintf(s, "    pin(%s) {\n", pPin.pName);
            fmt.Fprintf(s, "      direction : %s;\n", "output");
            fmt.Fprintf(s, "      max_capacitance : %f;\n", pPin.max_out_cap);
            fmt.Fprintf(s, "      max_transition : %f;\n", pPin.max_out_slew);
            fmt.Fprintf(s, "      function : \"%s\";\n", pPin.func_text ? pPin.func_text : "?");
            fmt.Fprintf(s, "      /*  truth table = ");
            Extra_PrintHex( s, (unsigned *)Vec_WrdArray(&pPin.vFunc), pCell.n_inputs );
            fmt.Fprintf(s, "  */\n");

            // Write 'rtiming': (pin-to-pin timing tables for this particular output)
            assert(Vec_PtrSize(&pPin.vRTimings) == pCell.n_inputs, "Vec_PtrSize(&pPin.vRTimings) == pCell.n_inputs");
            SC_PinForEachRTiming( pPin, pRTime, k )
            {
                if Vec_PtrSize(&pRTime.vTimings) == 1 {
                    var pTime *SC_Timing  = (SC_Timing *)Vec_PtrEntry( &pRTime.vTimings, 0 );
                    fmt.Fprintf(s, "      timing() {\n");
                    fmt.Fprintf(s, "        related_pin : \"%s\"\n", pRTime.pName);
                    if pTime.tsense == sc_ts_Pos {
                        fmt.Fprintf(s, "        timing_sense : positive_unate;\n");
                    } else if pTime.tsense == sc_ts_Neg {
                        fmt.Fprintf(s, "        timing_sense : negative_unate;\n");
                    } else if pTime.tsense == sc_ts_Non {
                        fmt.Fprintf(s, "        timing_sense : non_unate;\n");
                    } else { assert(false, "0"); }

                    fmt.Fprintf(s, "        cell_rise() {\n");
                    Abc_SclWriteSurfaceText( s, &pTime.pCellRise );
                    fmt.Fprintf(s, "        }\n");

                    fmt.Fprintf(s, "        cell_fall() {\n");
                    Abc_SclWriteSurfaceText( s, &pTime.pCellFall );
                    fmt.Fprintf(s, "        }\n");

                    fmt.Fprintf(s, "        rise_transition() {\n");
                    Abc_SclWriteSurfaceText( s, &pTime.pRiseTrans );
                    fmt.Fprintf(s, "        }\n");

                    fmt.Fprintf(s, "        fall_transition() {\n");
                    Abc_SclWriteSurfaceText( s, &pTime.pFallTrans );
                    fmt.Fprintf(s, "        }\n");
                    fmt.Fprintf(s, "      }\n");
                
                } else {
                    assert(Vec_PtrSize(&pRTime.vTimings) == 0, "Vec_PtrSize(&pRTime.vTimings) == 0"); }
            }
            fmt.Fprintf(s, "    }\n");
        }
        fmt.Fprintf(s, "  }\n");
    }
    fmt.Fprintf(s, "}\n\n");
}
func Abc_SclWriteLiberty(pFileName string, p *SC_Lib) {
    var pFile *os.File  = os.OpenFile(pFileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    if pFile == nil {
        fmt.Printf("Cannot open text file \"%s\" for writing.\n", pFileName);
    } else {
        Abc_SclWriteLibraryText( pFile, p );
        _ = (pFile).Close();
        fmt.Printf("Dumped internal library into Liberty file \"%s\".\n", pFileName);
    }
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

