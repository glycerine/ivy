package gia

import "fmt"

import "os"

/**CFile****************************************************************

  FileName    [giaGen.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    []

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaGen.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "gia.h"
//#include "misc/util/utilTruth.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////
 

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Populate internal simulation info.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManObjSim(p *Gia_Man_t, iObj int32) *word {
    return Vec_WrdEntryP( p.vSims, p.nSimWords * iObj );
}
func Gia_ManObjSimPi(p *Gia_Man_t, iObj int32) {
    var w int32;
    var pSim *word  = Gia_ManObjSim( p, iObj );
    for  w = 0; w < p.nSimWords; w++  {
        pSim[w] = Gia_ManRandomW( 0 );
    }
//    pSim[0] <<= 1;
}
func Gia_ManObjSimPo(p *Gia_Man_t, iObj int32) {
    var w int32;
    var pObj *Gia_Obj_t  = Gia_ManObj( p, iObj );
    var pSimCo *word   = Gia_ManObjSim( p, iObj );
    var pSimDri *word  = Gia_ManObjSim( p, Gia_ObjFaninId0(pObj, iObj) );
    if Gia_ObjFaninC0(pObj) {
        for  w = 0; w < p.nSimWords; w++  {
            pSimCo[w] = ~pSimDri[w];
        }
    } else {
        for  w = 0; w < p.nSimWords; w++  {
            pSimCo[w] =  pSimDri[w];
        } }
}
func Gia_ManObjSimAnd(p *Gia_Man_t, iObj int32) {
    var w int32;
    var pObj *Gia_Obj_t  = Gia_ManObj( p, iObj );
    var pSim *word   = Gia_ManObjSim( p, iObj );
    var pSim0 *word  = Gia_ManObjSim( p, Gia_ObjFaninId0(pObj, iObj) );
    var pSim1 *word  = Gia_ManObjSim( p, Gia_ObjFaninId1(pObj, iObj) );
    if Gia_ObjFaninC0(pObj) && Gia_ObjFaninC1(pObj) {
        for  w = 0; w < p.nSimWords; w++  {
            pSim[w] = ~pSim0[w] & ~pSim1[w];
        }
    } else if Gia_ObjFaninC0(pObj) && !Gia_ObjFaninC1(pObj) {
        for  w = 0; w < p.nSimWords; w++  {
            pSim[w] = ~pSim0[w] & pSim1[w];
        }
    } else if !Gia_ObjFaninC0(pObj) && Gia_ObjFaninC1(pObj) {
        for  w = 0; w < p.nSimWords; w++  {
            pSim[w] = pSim0[w] & ~pSim1[w];
        }
    } else {
        for  w = 0; w < p.nSimWords; w++  {
            pSim[w] = pSim0[w] & pSim1[w];
        } }
}
func Gia_ManSimulateWords(p *Gia_Man_t, nWords int32) int32 {
    var pObj *Gia_Obj_t; var i int32;
    // allocate simulation info for one timeframe
    Vec_WrdFreeP( &p.vSims );
    p.vSims = Vec_WrdStart( Gia_ManObjNum(p) * nWords );
    p.nSimWords = nWords;
    // perform simulation
    // inlining of macro Gia_ManForEachObj1
    for  i = 1; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
        if Gia_ObjIsAnd(pObj) {
            Gia_ManObjSimAnd( p, i );
        } else if Gia_ObjIsCi(pObj) {
            Gia_ManObjSimPi( p, i );
        } else if Gia_ObjIsCo(pObj) {
            Gia_ManObjSimPo( p, i );
        } else { assert(false, "0"); }
    }
    return 1;
}

func Gia_ManSimulateWordsInit(p *Gia_Man_t, vSimsIn *Vec_Wrd_t) int32 {
    var pObj *Gia_Obj_t; var i, Id int32;
    var nWords int32  = Vec_WrdSize(vSimsIn) / Gia_ManCiNum(p);
    assert(Vec_WrdSize(vSimsIn) == nWords * Gia_ManCiNum(p), "Vec_WrdSize(vSimsIn) == nWords * Gia_ManCiNum(p)");
    // allocate simulation info for one timeframe
    Vec_WrdFreeP( &p.vSims );
    p.vSims = Vec_WrdStart( Gia_ManObjNum(p) * nWords );
    p.nSimWords = nWords;
    // set input sim info
    // inlining of macro Gia_ManForEachCiId
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((Id) = Gia_ObjId(p, Gia_ManCi(p, i))); i++  {
        memcpy( Vec_WrdEntryP(p.vSims, Id*nWords), Vec_WrdEntryP(vSimsIn, i*nWords), sizeof(word)*nWords );
    }
    // perform simulation
    // inlining of macro Gia_ManForEachObj1
    for  i = 1; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
        if Gia_ObjIsAnd(pObj) {
            Gia_ManObjSimAnd( p, i );
        } else if Gia_ObjIsCi(pObj) {
            continue;
        } else if Gia_ObjIsCo(pObj) {
            Gia_ManObjSimPo( p, i );
        } else { assert(false, "0"); }
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Dump data files.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManDumpFiles(p *Gia_Man_t, nCexesT int32, nCexesV int32, Seed int32, pFileName string) {
    var n int32;
    var nSize [2]int32 = [2]int32{nCexesT*64, nCexesV*64};

    var pFileNameOutTX string // was [100]byte in C

    var pFileNameOutTY string // was [100]byte in C

    var pFileNameOutVX string // was [100]byte in C

    var pFileNameOutVY string // was [100]byte in C

    var pFileNameOut string // was [100]byte in C

    
    //sprintf( pFileNameOutTX, "train_%s_%d_%d.data", pFileName ? pFileName : Gia_ManName(p), nSize[0], Gia_ManCiNum(p) );
    //sprintf( pFileNameOutTY, "train_%s_%d_%d.data", pFileName ? pFileName : Gia_ManName(p), nSize[0], Gia_ManCoNum(p) );
    //sprintf( pFileNameOutVX, "test_%s_%d_%d.data",  pFileName ? pFileName : Gia_ManName(p), nSize[1], Gia_ManCiNum(p) );
    //sprintf( pFileNameOutVY, "test_%s_%d_%d.data",  pFileName ? pFileName : Gia_ManName(p), nSize[1], Gia_ManCoNum(p) );
    
    pFileNameOutTX = fmt.Sprintf("%s_x.train.data", pFileName ? pFileName : Gia_ManName(p));
    pFileNameOutTY = fmt.Sprintf("%s_y.train.data", pFileName ? pFileName : Gia_ManName(p));
    pFileNameOutVX = fmt.Sprintf("%s_x.test.data", pFileName ? pFileName : Gia_ManName(p));
    pFileNameOutVY = fmt.Sprintf("%s_y.test.data", pFileName ? pFileName : Gia_ManName(p));

    Gia_ManRandomW( 1 );
    for  n = 0; n < Seed; n++  {
        Gia_ManRandomW( 0 );
    }
    for  n = 0; n < 2; n++  {
        var Res int32  = Gia_ManSimulateWords( p, nSize[n] );

        var vBitX *Vec_Bit_t  = Vec_BitAlloc( nSize[n] * Gia_ManCiNum(p) );
        var vBitY *Vec_Bit_t  = Vec_BitAlloc( nSize[n] * Gia_ManCoNum(p) );

        var pFileOutX *os.File   = os.OpenFile(n ? pFileNameOutVX : pFileNameOutTX, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
        var pFileOutY *os.File   = os.OpenFile(n ? pFileNameOutVY : pFileNameOutTY, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);

        var i, k, Id, Num, Value, nBytes int32;
        for  k = 0; k < nSize[n]; k++  {
            // inlining of macro Gia_ManForEachCiId
            for  i = 0; (i < Vec_IntSize(p.vCis)) && ((Id) = Gia_ObjId(p, Gia_ManCi(p, i))); i++  {
                Vec_BitPush( vBitX, Abc_TtGetBit(Gia_ManObjSim(p, Id), k) );
                //printf( "%d", Abc_TtGetBit(Gia_ManObjSim(p, Id), k) );
            }
            //printf( " " );
            // inlining of macro Gia_ManForEachCoId
            for  i = 0; (i < Vec_IntSize(p.vCos)) && ((Id) = Gia_ObjId(p, Gia_ManCo(p, i))); i++  {
                Vec_BitPush( vBitY, Abc_TtGetBit(Gia_ManObjSim(p, Id), k) );
                //printf( "%d", Abc_TtGetBit(Gia_ManObjSim(p, Id), k) );
            }
            //printf( "\n" );
        }
        assert(Vec_BitSize(vBitX) <= Vec_BitCap(vBitX), "Vec_BitSize(vBitX) <= Vec_BitCap(vBitX)");
        assert(Vec_BitSize(vBitY) <= Vec_BitCap(vBitY), "Vec_BitSize(vBitY) <= Vec_BitCap(vBitY)");

        Num = 2;               Value = fwrite( &Num, 1, 4, pFileOutX ); assert(Value == 4, "Value == 4");
        Num = nSize[n];        Value = fwrite( &Num, 1, 4, pFileOutX ); assert(Value == 4, "Value == 4");
        Num = Gia_ManCiNum(p); Value = fwrite( &Num, 1, 4, pFileOutX ); assert(Value == 4, "Value == 4");

        nBytes = nSize[n] * Gia_ManCiNum(p) / 8;
        assert(nSize[n] * Gia_ManCiNum(p) % 8 == 0, "nSize[n] * Gia_ManCiNum(p) % 8 == 0");
        Value = fwrite( Vec_BitArray(vBitX), 1, nBytes, pFileOutX );
        assert(Value == nBytes, "Value == nBytes");

        Num = 2;               Value = fwrite( &Num, 1, 4, pFileOutY ); assert(Value == 4, "Value == 4");
        Num = nSize[n];        Value = fwrite( &Num, 1, 4, pFileOutY ); assert(Value == 4, "Value == 4");
        Num = Gia_ManCoNum(p); Value = fwrite( &Num, 1, 4, pFileOutY ); assert(Value == 4, "Value == 4");

        nBytes = nSize[n] * Gia_ManCoNum(p) / 8;
        assert(nSize[n] * Gia_ManCoNum(p) % 8 == 0, "nSize[n] * Gia_ManCoNum(p) % 8 == 0");
        Value = fwrite( Vec_BitArray(vBitY), 1, nBytes, pFileOutY );
        assert(Value == nBytes, "Value == nBytes");

        _ = (pFileOutX).Close();
        _ = (pFileOutY).Close();

        Vec_BitFree( vBitX );
        Vec_BitFree( vBitY );

        Res = 0;
    }
    printf( "Finished dumping files \"%s\" and \"%s\".\n", pFileNameOutTX, pFileNameOutTY );
    printf( "Finished dumping files \"%s\" and \"%s\".\n", pFileNameOutVX, pFileNameOutVY );

    pFileNameOut = fmt.Sprintf("%s.flist", pFileName ? pFileName : Gia_ManName(p));
    {
        var pFile *os.File  = fopen( pFileNameOut, "wb" );
        fprintf( pFile, "%s\n", pFileNameOutTX );
        fprintf( pFile, "%s\n", pFileNameOutTY );
        fprintf( pFile, "%s\n", pFileNameOutVX );
        fprintf( pFile, "%s\n", pFileNameOutVY );
        _ = (pFile).Close();
        printf( "Finished dumping file list \"%s\".\n", pFileNameOut );
    }
}

/**Function*************************************************************

  Synopsis    [Dump data files.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManDumpPlaFiles(p *Gia_Man_t, nCexesT int32, nCexesV int32, Seed int32, pFileName string) {
    var n int32;
    var nSize [3]int32 = [3]int32{nCexesT, nCexesV, nCexesV};

    var pFileNameOut [3][100]byte;
    
    sprintf( pFileNameOut[0], "%s.train.pla", pFileName ? pFileName : Gia_ManName(p) );
    sprintf( pFileNameOut[1], "%s.valid.pla", pFileName ? pFileName : Gia_ManName(p) );
    sprintf( pFileNameOut[2], "%s.test.pla",  pFileName ? pFileName : Gia_ManName(p) );

    Gia_ManRandomW( 1 );
    for  n = 0; n < Seed; n++  {
        Gia_ManRandomW( 0 );
    }
    for  n = 0; n < 3; n++  {
        var Res int32  = Gia_ManSimulateWords( p, nSize[n] );
        var i, k, Id int32;

        var pFileOut *os.File   = os.OpenFile(pFileNameOut[n], os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);

        fmt.Fprintf(pFileOut, ".i %d\n", Gia_ManCiNum(p));
        fmt.Fprintf(pFileOut, ".o %d\n", Gia_ManCoNum(p));
        fmt.Fprintf(pFileOut, ".p %d\n", nSize[n]*64);
        fmt.Fprintf(pFileOut, ".type fr\n");
        for  k = 0; k < nSize[n]*64; k++  {
            // inlining of macro Gia_ManForEachCiId
            for  i = 0; (i < Vec_IntSize(p.vCis)) && ((Id) = Gia_ObjId(p, Gia_ManCi(p, i))); i++  {
                //Vec_BitPush( vBitX, Abc_TtGetBit(Gia_ManObjSim(p, Id), k) );
                fmt.Fprintf(pFileOut, "%d", Abc_TtGetBit(Gia_ManObjSim(p, Id), k));
            }
            fmt.Fprintf(pFileOut, " ");
            // inlining of macro Gia_ManForEachCoId
            for  i = 0; (i < Vec_IntSize(p.vCos)) && ((Id) = Gia_ObjId(p, Gia_ManCo(p, i))); i++  {
                //Vec_BitPush( vBitY, Abc_TtGetBit(Gia_ManObjSim(p, Id), k) );
                fmt.Fprintf(pFileOut, "%d", Abc_TtGetBit(Gia_ManObjSim(p, Id), k));
            }
            fmt.Fprintf(pFileOut, "\n");
        }
        fmt.Fprintf(pFileOut, ".e\n");

        _ = (pFileOut).Close();

        Res = 0;
    }
    fmt.Printf("Finished dumping files: \"%s.{train, valid, test}.pla\".\n", pFileName ? pFileName : Gia_ManName(p));
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManSimLogStats(p *Gia_Man_t, pDumpFile string, Total int32, Correct int32, Guess int32) {
    var pTable *os.File  = os.OpenFile(pDumpFile, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    fmt.Fprintf(pTable, "{\n");
    fmt.Fprintf(pTable, "    \"name\" : \"%s\",\n", p.pName);
    fmt.Fprintf(pTable, "    \"input\" : %d,\n", Gia_ManCiNum(p));
    fmt.Fprintf(pTable, "    \"output\" : %d,\n", Gia_ManCoNum(p));
    fmt.Fprintf(pTable, "    \"and\" : %d,\n", Gia_ManAndNum(p));
    fmt.Fprintf(pTable, "    \"level\" : %d,\n", Gia_ManLevelNum(p));
    fmt.Fprintf(pTable, "    \"total\" : %d,\n", Total);
    fmt.Fprintf(pTable, "    \"correct\" : %d,\n", Correct);
    fmt.Fprintf(pTable, "    \"guess\" : %d\n", Guess);
    fmt.Fprintf(pTable, "}\n");
    _ = (pTable).Close();
}
func Gia_ManSimParamRead(pFileName string, pnIns *int32, pnWords *int32) int32 {
    var c int32;
    var nIns int32 = -1;
    var nLines int32;
    var Count int32;
    var fReadDot int32;
    var pFile *os.File  = os.Open(pFileName);
    if pFile == nil {
        fmt.Printf("Cannot open file \"%s\" for reading.\n", pFileName);
        return 0;
    }
    for  (c = fgetc(pFile)) != -1  {
        if c == '.' {
            fReadDot = 1; }
        if c == '\n' {
            if !(fReadDot != 0) {
                if nIns == -1 {
                    nIns = Count;
                } else if nIns != Count {
                    fmt.Printf("The number of symbols (%d) does not match other lines (%d).\n", Count, nIns);
                    _ = (pFile).Close();
                    return 0;
                }
                Count = 0;
                nLines++;
            }
            fReadDot = 0;
        }
        if fReadDot != 0 {
            continue; }
        if c != '0' && c != '1' {
            continue; }
        Count++;
    }
    if nLines % 64 > 0 {
        fmt.Printf("The number of lines (%d) is not divisible by 64.\n", nLines);
        _ = (pFile).Close();
        return 0;
    }
    *pnIns = nIns - 1;
    *pnWords = nLines / 64;
    //printf( "Expecting %d inputs and %d words of simulation data.\n", *pnIns, *pnWords );
    _ = (pFile).Close();
    return 1;
}
func Gia_ManSimFileRead(pFileName string, nIns int32, nWords int32, vSimsIn *Vec_Wrd_t, vValues *Vec_Int_t) {
    var c, nPats, Count, fReadDot int32;
    var pFile *os.File  = os.Open(pFileName);
    if pFile == nil {
        fmt.Printf("Cannot open file \"%s\" for reading.\n", pFileName);
        return;
    }
    assert(Vec_WrdSize(vSimsIn) % nWords == 0, "Vec_WrdSize(vSimsIn) % nWords == 0");
    for  (c = fgetc(pFile)) != -1  {
        if c == '.' {
            fReadDot = 1; }
        if c == '\n' {
            fReadDot = 0; }
        if fReadDot != 0 {
            continue; }
        if c != '0' && c != '1' {
            continue; }
        if Count == nIns {
            Vec_IntPush( vValues, c - '0' );
            Count = 0;
            nPats++;
        
        } else {
            if c == '1' {
                Abc_TtSetBit( Vec_WrdEntryP(vSimsIn, Count * nWords), nPats ); }
            Count++;
        }
    }
    assert(nPats == 64*nWords, "nPats == 64*nWords");
    _ = (pFile).Close();
    fmt.Printf("Finished reading %d simulation patterns for %d inputs. Probability of 1 at the output is %6.2f %%.\n", 64*nWords, nIns, 100.0*Vec_IntSum(vValues)/nPats);
}
func Gia_ManCompareValues(p *Gia_Man_t, vSimsIn *Vec_Wrd_t, vValues *Vec_Int_t, pDumpFile string) {
    var i int32;
    var Value int32;
    var Guess int32;
    var Count int32;
    var nWords int32 = Vec_WrdSize(vSimsIn) / Gia_ManCiNum(p);
    var pSims *word;
    assert(Vec_IntSize(vValues) == nWords * 64, "Vec_IntSize(vValues) == nWords * 64");
    Gia_ManSimulateWordsInit( p, vSimsIn );
    assert(p.nSimWords == nWords, "p.nSimWords == nWords");
    pSims = Gia_ManObjSim( p, Gia_ObjId(p, Gia_ManCo(p, 0)) );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vValues)) && (((Value) = Vec_IntEntry(vValues, i)), 1); i++  {
        if Abc_TtGetBit(pSims, i) == Value {
            Count++; }
    }
    if Vec_IntSum(vValues) > nWords * 32 {
        Guess = Vec_IntSum(vValues);
    } else {
        Guess = nWords * 64 - Vec_IntSum(vValues);
    }
    fmt.Printf("Total = %6d.  Errors = %6d.  Correct = %6d.  (%6.2f %%)   Naive guess = %6d.  (%6.2f %%)\n", Vec_IntSize(vValues), Vec_IntSize(vValues) - Count, Count, 100.0*Count/Vec_IntSize(vValues), Guess, 100.0*Guess/Vec_IntSize(vValues));
    if pDumpFile == nil {
        return; }
    Gia_ManSimLogStats( p, pDumpFile, Vec_IntSize(vValues), Count, Guess );
    fmt.Printf("Finished dumping statistics into file \"%s\".\n", pDumpFile);
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManTestOneFile(p *Gia_Man_t, pFileName string, pDumpFile string) {
    var vSimsIn *Vec_Wrd_t;
    var vValues *Vec_Int_t;
    var nIns, nWords int32;
    if !(Gia_ManSimParamRead( pFileName, &nIns, &nWords ) != 0) {
        return; }
    if nIns != Gia_ManCiNum(p) {
        fmt.Printf("The number of inputs in the file \"%s\" (%d) does not match the AIG (%d).\n", pFileName, nIns, Gia_ManCiNum(p));
        return;
    }
    vSimsIn = Vec_WrdStart( nIns * nWords );
    vValues = Vec_IntAlloc( nWords * 64 );
    Gia_ManSimFileRead( pFileName, nIns, nWords, vSimsIn, vValues );
    Gia_ManCompareValues( p, vSimsIn, vValues, pDumpFile );
    Vec_WrdFree( vSimsIn );
    Vec_IntFree( vValues );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

