package glucose2

import "strconv"

/**CFile****************************************************************

  FileName    [AbcGlucoseCmd.cpp]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [SAT solver Glucose 3.0 by Gilles Audemard and Laurent Simon.]

  Synopsis    [Interface to Glucose.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - September 6, 2017.]

  Revision    [$Id: AbcGlucoseCmd.cpp,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "aig/gia/gia.h"
//#include "base/main/mainInt.h"

//#include "sat/glucose2/AbcGlucose2.h"


//ABC_NAMESPACE_HEADER_START

// extern void Glucose2_Init( Abc_Frame_t *pAbc );

// extern void Glucose2_End( Abc_Frame_t * pAbc );


//ABC_NAMESPACE_HEADER_END

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// static int Abc_CommandGlucose( Abc_Frame_t * pAbc, int argc, char ** argv );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/

func Glucose2_Init(pAbc *Abc_Frame_t) {
    Cmd_CommandAdd( pAbc, "ABC9", "&glucose2",  Abc_CommandGlucose,  0 );
}

func Glucose2_End(pAbc *Abc_Frame_t) {
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_CommandGlucose(pAbc *Abc_Frame_t, argc int32, argv *string) int32 {
    var c int32;
    var pre int32      = 1;
    var verb int32;
    var nConfls int32;

    var pPars Glucose2_Pars;
    Extra_UtilGetoptReset();
    for  ( c = Extra_UtilGetopt( argc, argv, "Cpvh" ) ) != -1  {
        switch  c  {
            case 'C':
                if globalUtilOptind >= argc {
                    Abc_Print( -1, "Command line switch \"-C\" should be followed by an integer.\n" );
                    goto usage;
                }
                nConfls = strconv.Atoi(argv[globalUtilOptind]);
                globalUtilOptind++;
                if nConfls < 0 {
                    goto usage; }
                break;
            case 'p':
                pre ^= 1;
                break;
            case 'v':
                verb ^= 1;
                break;
            case 'h':
                goto usage;
            default:
                goto usage;
        }
    }

    pPars = Glucose_CreatePars(pre,verb,0,nConfls);

    if argc == globalUtilOptind + 1 {
        Glucose2_SolveCnf( argv[globalUtilOptind], &pPars );
        return 0;
    }

    if pAbc.pGia == nil {
        Abc_Print( -1, "Abc_CommandGlucose(): There is no AIG.\n" );
        return 1;
    }
    
    if Glucose2_SolveAig( pAbc.pGia, &pPars ) == 10 {
        Abc_FrameReplaceCex( pAbc, &pAbc.pGia.pCexComb ); }

    return 0;
    
usage:
    Abc_Print( -2, "usage: &glucose2 [-C num] [-pvh] <file.cnf>\n" );
    Abc_Print( -2, "\t             run Glucose 3.0 by Gilles Audemard and Laurent Simon\n" );
    Abc_Print( -2, "\t-C num     : conflict limit [default = %d]\n",  nConfls );
    Abc_Print( -2, "\t-p         : enable preprocessing [default = %d]\n",pre);
    Abc_Print( -2, "\t-v         : verbosity [default = %d]\n",verb);
    Abc_Print( -2, "\t-h         : print the command usage\n");
    Abc_Print( -2, "\t<file.cnf> : (optional) CNF file to solve\n");
    return 1;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END
