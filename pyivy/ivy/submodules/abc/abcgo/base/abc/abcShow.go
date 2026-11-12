package abc

import "fmt"

import "os"

/**CFile****************************************************************

  FileName    [abcShow.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Network and node package.]

  Synopsis    [Visualization procedures using DOT software and GSView.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: abcShow.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

#ifdef WIN32
#include <process.h> 
#else
#include <unistd.h>
#endif


//#include "abc.h"
//#include "base/main/main.h"
//#include "base/io/ioAbc.h"

#ifdef ABC_USE_CUDD
#include "bdd/extrab/extraBdd.h"
#endif

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// extern void Abc_ShowFile( char * FileNameDot );

// static void Abc_ShowGetFileName( char * pName, char * pBuffer );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

#ifdef ABC_USE_CUDD

/**Function*************************************************************

  Synopsis    [Visualizes BDD of the node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
void Abc_NodeShowBddOne( DdManager * dd, DdNode * bFunc )
{
    char * FileNameDot = "temp.dot";
    FILE * pFile;
    if ( (pFile = fopen( FileNameDot, "w" )) == NULL )
    {
        fprintf( stdout, "Cannot open the intermediate file \"%s\".\n", FileNameDot );
        return;
    }
    Cudd_DumpDot( dd, 1, (DdNode **)&bFunc, NULL, NULL, pFile );
    fclose( pFile );
    Abc_ShowFile( FileNameDot );
}

/**Function*************************************************************

  Synopsis    [Visualizes BDD of the node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
void Abc_NodeShowBdd( Abc_Obj_t * pNode, int fCompl )
{
    FILE * pFile;
    Vec_Ptr_t * vNamesIn;
    char FileNameDot[200];
    char * pNameOut;
    DdManager * dd = (DdManager *)pNode->pNtk->pManFunc;

    assert( Abc_NtkIsBddLogic(pNode->pNtk) );
    // create the file name
    Abc_ShowGetFileName( Abc_ObjName(pNode), FileNameDot );
    // check that the file can be opened
    if ( (pFile = fopen( FileNameDot, "w" )) == NULL )
    {
        fprintf( stdout, "Cannot open the intermediate file \"%s\".\n", FileNameDot );
        return;
    }

    // set the node names 
    vNamesIn = Abc_NodeGetFaninNames( pNode );
    pNameOut = Abc_ObjName(pNode);
    if ( fCompl )
        Cudd_DumpDot( dd, 1, (DdNode **)&pNode->pData, (char **)vNamesIn->pArray, &pNameOut, pFile );
    else
    {
        DdNode * bAdd = Cudd_BddToAdd( dd, (DdNode *)pNode->pData ); Cudd_Ref( bAdd );
        Cudd_DumpDot( dd, 1, (DdNode **)&bAdd, (char **)vNamesIn->pArray, &pNameOut, pFile );
        Cudd_RecursiveDeref( dd, bAdd );
    }
    Abc_NodeFreeNames( vNamesIn );
    Abc_NtkCleanCopy( pNode->pNtk );
    fclose( pFile );

    // visualize the file 
    Abc_ShowFile( FileNameDot );
}
void Abc_NtkShowBdd( Abc_Ntk_t * pNtk, int fCompl )
{
    char FileNameDot[200];
    char ** ppNamesIn, ** ppNamesOut;
    DdManager * dd; DdNode * bFunc;
    Vec_Ptr_t * vFuncsGlob;
    Abc_Obj_t * pObj; int i;
    FILE * pFile;

    assert( Abc_NtkIsStrash(pNtk) );
    dd = (DdManager *)Abc_NtkBuildGlobalBdds( pNtk, 10000000, 1, 1, 0, 0 );
    if ( dd == NULL )
    {
        printf( "Construction of global BDDs has failed.\n" );
        return;
    }
    //printf( "Shared BDD size = %6d nodes.\n", Cudd_ReadKeys(dd) - Cudd_ReadDead(dd) );

    // complement the global functions
    vFuncsGlob = Vec_PtrAlloc( Abc_NtkCoNum(pNtk) );
    Abc_NtkForEachCo( pNtk, pObj, i )
        Vec_PtrPush( vFuncsGlob, Abc_ObjGlobalBdd(pObj) );

    // create the file name
    Abc_ShowGetFileName( pNtk->pName, FileNameDot );
    // check that the file can be opened
    if ( (pFile = fopen( FileNameDot, "w" )) == NULL )
    {
        fprintf( stdout, "Cannot open the intermediate file \"%s\".\n", FileNameDot );
        return;
    }

    // set the node names 
    ppNamesIn = Abc_NtkCollectCioNames( pNtk, 0 );
    ppNamesOut = Abc_NtkCollectCioNames( pNtk, 1 );
    if ( fCompl )
        Cudd_DumpDot( dd, Abc_NtkCoNum(pNtk), (DdNode **)Vec_PtrArray(vFuncsGlob), ppNamesIn, ppNamesOut, pFile );
    else
    {
        DdNode ** pbAdds = ABC_ALLOC( DdNode *, Vec_PtrSize(vFuncsGlob) );
        Vec_PtrForEachEntry( DdNode *, vFuncsGlob, bFunc, i )
            { pbAdds[i] = Cudd_BddToAdd( dd, bFunc ); Cudd_Ref( pbAdds[i] ); }
        Cudd_DumpDot( dd, Abc_NtkCoNum(pNtk), pbAdds, ppNamesIn, ppNamesOut, pFile );
        Vec_PtrForEachEntry( DdNode *, vFuncsGlob, bFunc, i )
            Cudd_RecursiveDeref( dd, pbAdds[i] );
        ABC_FREE( pbAdds );
    }
    ABC_FREE( ppNamesIn );
    ABC_FREE( ppNamesOut );
    fclose( pFile );

    // cleanup
    Abc_NtkFreeGlobalBdds( pNtk, 0 );
    Vec_PtrForEachEntry( DdNode *, vFuncsGlob, bFunc, i )
        Cudd_RecursiveDeref( dd, bFunc );
    Vec_PtrFree( vFuncsGlob );
    Extra_StopManager( dd );
    Abc_NtkCleanCopy( pNtk );

    // visualize the file 
    Abc_ShowFile( FileNameDot );
}

#else
void Abc_NodeShowBdd( Abc_Obj_t * pNode, int fCompl ) {}
void Abc_NtkShowBdd( Abc_Ntk_t * pNtk, int fCompl ) {}
#endif

/**Function*************************************************************

  Synopsis    [Visualizes a reconvergence driven cut at the node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NodeShowCut(pNode *Abc_Obj_t, nNodeSizeMax int32, nConeSizeMax int32) {
    var pFile *os.File;
    var FileNameDot [200]byte;
    var p *Abc_ManCut_t;
    var vCutSmall *Vec_Ptr_t;
    var vCutLarge *Vec_Ptr_t;
    var vInside *Vec_Ptr_t;
    var vNodesTfo *Vec_Ptr_t;
    var pTemp *Abc_Obj_t;
    var i int32;

    assert(Abc_NtkIsStrash(pNode.pNtk), "Abc_NtkIsStrash(pNode.pNtk)");

    // start the cut computation manager
    p = Abc_NtkManCutStart( nNodeSizeMax, nConeSizeMax, 2, ABC_INFINITY );
    // get the recovergence driven cut
    vCutSmall = Abc_NodeFindCut( p, pNode, 1 );
    // get the containing cut
    vCutLarge = Abc_NtkManCutReadCutLarge( p );
    // get the array for the inside nodes
    vInside = Abc_NtkManCutReadVisited( p );
    // get the inside nodes of the containing cone
    Abc_NodeConeCollect( &pNode, 1, vCutLarge, vInside, 1 );

    // add the nodes in the TFO 
    vNodesTfo = Abc_NodeCollectTfoCands( p, pNode, vCutSmall, ABC_INFINITY );
    Vec_PtrForEachEntry( Abc_Obj_t *, vNodesTfo, pTemp, i )
        Vec_PtrPushUnique( vInside, pTemp );

    // create the file name
    Abc_ShowGetFileName( Abc_ObjName(pNode), FileNameDot );
    // check that the file can be opened
    if (pFile = os.OpenFile(FileNameDot, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666)) == nil {
        fmt.Fprintf(os.Stdout, "Cannot open the intermediate file \"%s\".\n", FileNameDot);
        return;
    }
    // add the root node to the cone (for visualization)
    Vec_PtrPush( vCutSmall, pNode );
    // write the DOT file
    Io_WriteDotNtk( pNode.pNtk, vInside, vCutSmall, FileNameDot, 0, 0 );
    // stop the cut computation manager
    Abc_NtkManCutStop( p );

    // visualize the file 
    Abc_ShowFile( FileNameDot );
}

/**Function*************************************************************

  Synopsis    [Visualizes AIG with choices.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkShow(pNtk0 *Abc_Ntk_t, fGateNames int32, fSeq int32, fUseReverse int32) {
    var pFile *os.File;
    var pNtk *Abc_Ntk_t;
    var pNode *Abc_Obj_t;
    var vNodes *Vec_Ptr_t;
    var nBarBufs int32;
    var FileNameDot [200]byte;
    var i int32;

    assert(Abc_NtkIsStrash(pNtk0) || Abc_NtkIsLogic(pNtk0), "Abc_NtkIsStrash(pNtk0) || Abc_NtkIsLogic(pNtk0)");
    if Abc_NtkIsStrash(pNtk0) && Abc_NtkGetChoiceNum(pNtk0) {
        fmt.Printf("Temporarily visualization of AIGs with choice nodes is disabled.\n");
        return;
    }
    // create the file name
    Abc_ShowGetFileName( pNtk0.pName, FileNameDot );
    // check that the file can be opened
    if (pFile = os.OpenFile(FileNameDot, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666)) == nil {
        fmt.Fprintf(os.Stdout, "Cannot open the intermediate file \"%s\".\n", FileNameDot);
        return;
    }
    _ = (pFile).Close();


    // convert to logic SOP
    pNtk = Abc_NtkDup( pNtk0 );
    if Abc_NtkIsLogic(pNtk) && !Abc_NtkHasMapping(pNtk) {
        Abc_NtkToSop( pNtk, -1, ABC_INFINITY ); }

    // collect all nodes in the network
    vNodes = Vec_PtrAlloc( 100 );
    Abc_NtkForEachObj( pNtk, pNode, i )
        Vec_PtrPush( vNodes, pNode );
    // write the DOT file
    nBarBufs = pNtk.nBarBufs;
    pNtk.nBarBufs = 0;
    if fSeq != 0 {
        Io_WriteDotSeq( pNtk, vNodes, nil, FileNameDot, fGateNames, fUseReverse );
    } else {
        Io_WriteDotNtk( pNtk, vNodes, nil, FileNameDot, fGateNames, fUseReverse ); }
    pNtk.nBarBufs = nBarBufs;
    Vec_PtrFree( vNodes );

    // visualize the file 
    Abc_ShowFile( FileNameDot );
    Abc_NtkDelete( pNtk );
}


/**Function*************************************************************

  Synopsis    [Shows the given DOT file.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_ShowFile(FileNameDot string) {
    var pFile *os.File;
    var FileGeneric string;
    var FileNamePs string // was [200]byte in C

    var CommandDot [1000]byte;
    var pDotName string;
    var pDotNameWin string  = "dot.exe";
    var pDotNameUnix string  = "dot";
    var pGsNameWin string  = "gsview32.exe";
    var pGsNameUnix string  = "gv";
    var RetValue int32;

    // get DOT names from the resource file
    if Abc_FrameReadFlag("dotwin") {
        pDotNameWin = Abc_FrameReadFlag("dotwin"); }
    if Abc_FrameReadFlag("dotunix") {
        pDotNameUnix = Abc_FrameReadFlag("dotunix"); }

#ifdef WIN32
    pDotName = pDotNameWin;
#else
    pDotName = pDotNameUnix;
#endif

    // check if the input DOT file is okay
    if (pFile = os.Open(FileNameDot)) == nil {
        fmt.Fprintf(os.Stdout, "Cannot open the intermediate file \"%s\".\n", FileNameDot);
        return;
    }
    _ = (pFile).Close();

    // create the PostScript file name
    FileGeneric = Extra_FileNameGeneric( FileNameDot );
    FileNamePs = fmt.Sprintf("%s.ps", FileGeneric); 
    ABC_FREE( FileGeneric );

    // generate the PostScript file using DOT
    sprintf( CommandDot,  "%s -Tps -o %s %s", pDotName, FileNamePs, FileNameDot ); 
    RetValue = system( CommandDot );
    if RetValue == -1 {
        fprintf( os.Stdout, "Command \"%s\" did not succeed.\n", CommandDot );
        return;
    }
    // check that the input PostScript file is okay
    if (pFile = fopen( FileNamePs, "r" )) == nil {
        fprintf( os.Stdout, "Cannot open intermediate file \"%s\".\n", FileNamePs );
        return;
    }
    _ = (pFile).Close(); 


    // get GSVIEW names from the resource file
    if Abc_FrameReadFlag("gsviewwin") {
        pGsNameWin = Abc_FrameReadFlag("gsviewwin"); }
    if Abc_FrameReadFlag("gsviewunix") {
        pGsNameUnix = Abc_FrameReadFlag("gsviewunix"); }

    // spawn the viewer
#ifdef WIN32
    _unlink( FileNameDot );
    if ( _spawnl( _P_NOWAIT, pGsNameWin, pGsNameWin, FileNamePs, NULL ) == -1 )
        if ( _spawnl( _P_NOWAIT, "C:\\Program Files\\Ghostgum\\gsview\\gsview32.exe", 
            "C:\\Program Files\\Ghostgum\\gsview\\gsview32.exe", FileNamePs, NULL ) == -1 )
            if ( _spawnl( _P_NOWAIT, "C:\\Program Files\\Ghostgum\\gsview\\gsview64.exe", 
                "C:\\Program Files\\Ghostgum\\gsview\\gsview64.exe", FileNamePs, NULL ) == -1 )
            {
                fprintf( stdout, "Cannot find \"%s\".\n", pGsNameWin );
                return;
            }
#else
    {
        char CommandPs[1000];
        unlink( FileNameDot );
        sprintf( CommandPs,  "%s %s &", pGsNameUnix, FileNamePs ); 
        if ( system( CommandPs ) == -1 )
        {
            fprintf( stdout, "Cannot execute \"%s\".\n", CommandPs );
            return;
        }
    }
#endif
}

/**Function*************************************************************

  Synopsis    [Derives the DOT file name.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_ShowGetFileName(pName string, pBuffer string) {
    var pCur string;
    // creat the file name
    pBuffer = fmt.Sprintf("%s.dot", pName);
    // get rid of not-alpha-numeric characters
    for  pCur = pBuffer; *pCur; pCur++  {
        if !((*pCur >= '0' && *pCur <= '9') || (*pCur >= 'a' && *pCur <= 'z') || 
               (*pCur >= 'A' && *pCur <= 'Z') || (*pCur == '.')) {
            *pCur = '_'; }
    }
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkWriteFlopDependency(pNtk *Abc_Ntk_t, pFileName string) {
    var pFile *os.File;
    var vSupp *Vec_Ptr_t;
    var pObj, pTemp *Abc_Obj_t;
    var i, k, Count int32;
    pFile = os.OpenFile(pFileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    if pFile == nil {
        fmt.Printf("Cannot open input file %s.\n", pFileName);
        return;
    }
    fmt.Fprintf(pFile, "# Flop dependency for \"%s\" generated by ABC on %s\n", Abc_NtkName(pNtk), Extra_TimeStamp());
    fmt.Fprintf(pFile, "digraph G {\n");
    fmt.Fprintf(pFile, "  graph [splines=true overlap=false];\n");
    fmt.Fprintf(pFile, "  size = \"7.5,10\";\n");
    fmt.Fprintf(pFile, "  center = true;\n");
//    fprintf( pFile, "  edge [len=3,dir=forward];\n" );
    fmt.Fprintf(pFile, "  edge [dir=forward];\n");
    Abc_NtkForEachLatchInput( pNtk, pObj, i )
    {
        Abc_ObjFanout0( Abc_ObjFanout0(pObj) ).iTemp = i;
        vSupp = Abc_NtkNodeSupport( pNtk, &pObj, 1 );
        Count = 0;
        Vec_PtrForEachEntry( Abc_Obj_t *, vSupp, pTemp, k )
            Count += Abc_ObjIsPi(pTemp);
        Vec_PtrFree( vSupp );
        fmt.Fprintf(pFile, "  { rank = same; %d [label=\"%d(%d)\"]; }\n", i, i, Count);
    }
    Abc_NtkForEachLatchInput( pNtk, pObj, i )
    {
        vSupp = Abc_NtkNodeSupport( pNtk, &pObj, 1 );
        Count = 0;
        Vec_PtrForEachEntry( Abc_Obj_t *, vSupp, pTemp, k )
            if ( !Abc_ObjIsPi(pTemp) )
                fmt.Fprintf(pFile, "  %4d -> %4d\n", pTemp.iTemp, i);
        Vec_PtrFree( vSupp );
    }
    fmt.Fprintf(pFile, "}\n");
    _ = (pFile).Close();
}


/**Function*************************************************************

  Synopsis    [Visualizes AIG with choices.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkShowFlopDependency(pNtk *Abc_Ntk_t) {
    var pFile *os.File;
    var FileNameDot [200]byte;
    assert(Abc_NtkIsStrash(pNtk) || Abc_NtkIsLogic(pNtk), "Abc_NtkIsStrash(pNtk) || Abc_NtkIsLogic(pNtk)");
    // create the file name
    Abc_ShowGetFileName( pNtk.pName, FileNameDot );
    // check that the file can be opened
    if (pFile = os.OpenFile(FileNameDot, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666)) == nil {
        fmt.Fprintf(os.Stdout, "Cannot open the intermediate file \"%s\".\n", FileNameDot);
        return;
    }
    _ = (pFile).Close();
    // write the DOT file
    Abc_NtkWriteFlopDependency( pNtk, FileNameDot );
    // visualize the file 
    Abc_ShowFile( FileNameDot );
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

