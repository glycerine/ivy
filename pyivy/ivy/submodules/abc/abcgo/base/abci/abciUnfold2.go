package abci

import "strconv"


/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_CommandUnfold2(pAbc *Abc_Frame_t, argc int32, argv *string) int32 {
    var pNtk, pNtkRes *Abc_Ntk_t;
    var nFrames int32;
    var nConfs int32;
    var nProps int32;
    var fStruct int32;
    var fOldAlgo int32;
    var fVerbose int32;
    var c int32;
    extern Abc_Ntk_t * Abc_NtkDarUnfold2( Abc_Ntk_t * pNtk, int nFrames, int nConfs, int nProps, int fStruct, int fOldAlgo, int fVerbose );
    pNtk = Abc_FrameReadNtk(pAbc);
    // set defaults
    nFrames   =      1;
    nConfs    =   1000;
    nProps    =   1000;
    fVerbose  =      0;
    Extra_UtilGetoptReset();
    for  ( c = Extra_UtilGetopt( argc, argv, "CPvh" ) ) != -1  {
        switch  c  {
        /* case 'F': */
        /*     if ( globalUtilOptind >= argc ) */
        /*     { */
        /*         Abc_Print( -1, "Command line switch \"-F\" should be followed by an integer.\n" ); */
        /*         goto usage; */
        /*     } */
        /*     nFrames = atoi(argv[globalUtilOptind]); */
        /*     globalUtilOptind++; */
        /*     if ( nFrames < 0 ) */
        /*         goto usage; */
        /*     break; */
        case 'C':
            if globalUtilOptind >= argc {
                Abc_Print( -1, "Command line switch \"-C\" should be followed by an integer.\n" );
                goto usage;
            }
            nConfs = strconv.Atoi(argv[globalUtilOptind]);
            globalUtilOptind++;
            if nConfs < 0 {
                goto usage; }
            break;
        case 'P':
            if globalUtilOptind >= argc {
                Abc_Print( -1, "Command line switch \"-P\" should be followed by an integer.\n" );
                goto usage;
            }
            nProps = strconv.Atoi(argv[globalUtilOptind]);
            globalUtilOptind++;
            if nProps < 0 {
                goto usage; }
            break;
        case 'v':
            fVerbose ^= 1;
            break;
        case 'h':
            goto usage;
        default:
            goto usage;
        }
    }
    if pNtk == nil {
        Abc_Print( -1, "Empty network.\n" );
        return 1;
    }
    if Abc_NtkIsComb(pNtk) {
        Abc_Print( -1, "The network is combinational.\n" );
        return 0;
    }
    if !Abc_NtkIsStrash(pNtk) {
        Abc_Print( -1, "Currently only works for structurally hashed circuits.\n" );
        return 0;
    }
    if Abc_NtkConstrNum(pNtk) > 0 {
        Abc_Print( -1, "Constraints are already extracted.\n" );
        return 0;
    }
    if Abc_NtkPoNum(pNtk) > 1 && !(fStruct != 0) {
        Abc_Print( -1, "Functional constraint extraction works for single-output miters (use \"orpos\").\n" );
        return 0;
    }
    // modify the current network
    pNtkRes = Abc_NtkDarUnfold2( pNtk, nFrames, nConfs, nProps, fStruct, fOldAlgo, fVerbose );
    if pNtkRes == nil {
        Abc_Print( 1,"Transformation has failed.\n" );
        return 0;
    }
    // replace the current network
    Abc_FrameReplaceCurrentNetwork( pAbc, pNtkRes );
    return 0;
usage:
    Abc_Print( -2, "usage: unfold2 [-FCP num] [-savh]\n" );
    Abc_Print( -2, "\t         unfold hidden constraints as separate outputs\n" );
    Abc_Print( -2, "\t-C num : the max number of conflicts in SAT solving [default = %d]\n", nConfs );
    Abc_Print( -2, "\t-P num : the max number of constraint propagations [default = %d]\n", nProps );
    Abc_Print( -2, "\t-v     : toggle printing verbose information [default = %s]\n", fVerbose? "yes": "no" );
    Abc_Print( -2, "\t-h     : print the command usage\n");
    return 1;
}



func Abc_CommandFold2(pAbc *Abc_Frame_t, argc int32, argv *string) int32 {
    var pNtk, pNtkRes *Abc_Ntk_t;
    var fCompl int32;
    var fVerbose int32;
    var c int32;
    extern Abc_Ntk_t * Abc_NtkDarFold2( Abc_Ntk_t * pNtk, int fCompl, int fVerbose , int);
    pNtk = Abc_FrameReadNtk(pAbc);
    // set defaults
    fCompl    =   0;
    fVerbose  =   0;
    Extra_UtilGetoptReset();
    for  ( c = Extra_UtilGetopt( argc, argv, "cvh" ) ) != -1  {
        switch  c  {
        /* case 'c': */
        /*     fCompl ^= 1; */
        /*     break; */
        case 'v':
            fVerbose ^= 1;
            break;
        case 'h':
            goto usage;
        default:
            goto usage;
        }
    }
    if pNtk == nil {
        Abc_Print( -1, "Empty network.\n" );
        return 1;
    }
    if !Abc_NtkIsStrash(pNtk) {
        Abc_Print( -1, "Currently only works for structurally hashed circuits.\n" );
        return 0;
    }
    if Abc_NtkConstrNum(pNtk) == 0 {
        Abc_Print( 0, "The network has no constraints.\n" );
        return 0;
    }
    if Abc_NtkIsComb(pNtk) {
        Abc_Print( 0, "The network is combinational.\n" ); }
    // modify the current network
    pNtkRes = Abc_NtkDarFold2( pNtk, fCompl, fVerbose ,0);
    if pNtkRes == nil {
        Abc_Print( 1,"Transformation has failed.\n" );
        return 0;
    }
    // replace the current network
    Abc_FrameReplaceCurrentNetwork( pAbc, pNtkRes );
    return 0;
usage:
    Abc_Print( -2, "usage: fold [-cvh]\n" );
    Abc_Print( -2, "\t         folds constraints represented as separate outputs\n" );
    //    Abc_Print( -2, "\t-c     : toggle complementing constraints while folding [default = %s]\n", fCompl? "yes": "no" );
    Abc_Print( -2, "\t-v     : toggle printing verbose information [default = %s]\n", fVerbose? "yes": "no" );
    Abc_Print( -2, "\t-h     : print the command usage\n");
    return 1;
}
