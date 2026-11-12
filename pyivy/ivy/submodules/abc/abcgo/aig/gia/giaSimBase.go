package gia

import "fmt"

import "os"

/**CFile****************************************************************

  FileName    [giaSim.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    [Fast sequential simulator.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaSim.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "gia.h"
//#include "misc/util/utilTruth.h"
//#include "misc/extra/extra.h"
//#include <immintrin.h>

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








/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/





/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/




/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/





/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/





/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/






/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/



/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/



/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/



/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/




/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/



/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/







/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/





/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/




/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/






/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/





/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/







/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/




/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

   

















/**Function*************************************************************

  Synopsis    [Improving quality of simulation patterns.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/


   






/**Function*************************************************************

  Synopsis    [Improving quality of simulation patterns.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/








/**Function*************************************************************

  Synopsis    [Trying vectorized simulation.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManSimTest(pGia *Gia_Man_t) {
    var n int32;
    var nWords int32 = 4;
    var vSim1, vSim2 *Vec_Wrd_t;
    var vSim0 *Vec_Wrd_t  = Vec_WrdStartRandom( Gia_ManCiNum(pGia) * nWords );
    var clk abctime  = Abc_Clock();

    pGia.vSimsPi = vSim0;
    for  n = 0; n < 20; n++  {
        vSim1 = Gia_ManSimPatSim( pGia );
        Vec_WrdFree( vSim1 );
    }
    Abc_PrintTime( 1, "Time1", Abc_Clock() - clk );

    clk = Abc_Clock();
    for  n = 0; n < 20; n++  {
        vSim2 = Gia_ManSimPatSim2( pGia );
        Vec_WrdFree( vSim2 );
    }
    Abc_PrintTime( 1, "Time2", Abc_Clock() - clk );

    pGia.vSimsPi = nil;
    Vec_WrdFree( vSim0 );
}

/**Function*************************************************************

  Synopsis    [Trying compiled simulation.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManSimGen(pGia *Gia_Man_t) {
    var nWords int32  = 4;
    var pObj *Gia_Obj_t;
    var vSim0 *Vec_Wrd_t  = Vec_WrdStartRandom( Gia_ManCiNum(pGia) * nWords );
    var pFile *os.File  = os.OpenFile("comp_sim.c", os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    var i, k, Id int32;
    fmt.Fprintf(pFile, "#include <stdio.h>\n");
    fmt.Fprintf(pFile, "#include <stdlib.h>\n");
    fmt.Fprintf(pFile, "#include <time.h>\n");
    fmt.Fprintf(pFile, "int main()\n");
    fmt.Fprintf(pFile, "{\n");
    fmt.Fprintf(pFile, "  clock_t clkThis = clock();\n");
    fmt.Fprintf(pFile, "  unsigned long Res = 0;\n");
    fmt.Fprintf(pFile, "  int i;\n");
    fmt.Fprintf(pFile, "  srand(time(NULL));\n");
    fmt.Fprintf(pFile, "  for ( i = 0; i < 2000; i++ )\n");
    fmt.Fprintf(pFile, "  {\n");
    for  k = 0; k < nWords; k++  {
        fmt.Fprintf(pFile, "  unsigned long s%07d_%d = 0x%08x%08x;\n", 0, k, 0, 0);
    }
    // inlining of macro Gia_ManForEachCiId
    for  i = 0; (i < Vec_IntSize(pGia.vCis)) && ((Id) = Gia_ObjId(pGia, Gia_ManCi(pGia, i))); i++  {
        //word * pSim = Vec_WrdEntryP(vSim0, i*nWords);
        //unsigned * pSimU = (unsigned *)pSim;
        for  k = 0; k < nWords; k++  {
            fmt.Fprintf(pFile, "  unsigned long s%07d_%d = ((unsigned long)rand() << 48) | ((unsigned long)rand() << 32) | ((unsigned long)rand() << 16) | (unsigned long)rand();\n", Id, k);
        }
    }
    // inlining of macro Gia_ManForEachAnd
    for  Id = 0; (Id < pGia.nObjs) && ((pObj) = Gia_ManObj(pGia, Id)); Id++  {      if !Gia_ObjIsAnd(pObj) {} else {
        for  k = 0; k < nWords; k++  {
            fmt.Fprintf(pFile, "  unsigned long s%07d_%d = %cs%07d_%d & %cs%07d_%d;\n", Id, k, Gia_ObjFaninC0(pObj) ? '~' : ' ', Gia_ObjFaninId0(pObj, Id), k, Gia_ObjFaninC1(pObj) ? ' ' : '~', Gia_ObjFaninId1(pObj, Id), k);
        }
    }
    }
    // inlining of macro Gia_ManForEachCoId
    for  i = 0; (i < Vec_IntSize(pGia.vCos)) && ((Id) = Gia_ObjId(pGia, Gia_ManCo(pGia, i))); i++  {
        pObj = Gia_ManObj(pGia, Id);
        for  k = 0; k < nWords; k++  {
            fmt.Fprintf(pFile, "  Res ^= %cs%07d_%d;\n", Gia_ObjFaninC0(pObj) ? '~' : ' ', Gia_ObjFaninId0(pObj, Id), k);
        }
    }
    Vec_WrdFree( vSim0 );
    fmt.Fprintf(pFile, "  }\n");
    fmt.Fprintf(pFile, "  printf( \"Res = 0x%%08x    \", (unsigned)Res );\n");
    fmt.Fprintf(pFile, "  printf( \"Time = %%6.2f sec\\n\", (float)(clock() - clkThis)/CLOCKS_PER_SEC );\n");
    fmt.Fprintf(pFile, "  return 1;\n");
    fmt.Fprintf(pFile, "}\n");
    _ = (pFile).Close();
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

