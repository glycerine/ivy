package gia

import "fmt"

import "os"

/**CFile****************************************************************

  FileName    [giaShow.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [And-Inverter Graph package.]

  Synopsis    [AIG visualization.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - May 11, 2006.]

  Revision    [$Id: giaShow.c,v 1.00 2006/05/11 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "gia.h"
//#include "proof/cec/cec.h"
//#include "proof/acec/acec.h"
//#include "misc/extra/extra.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

const NODE_MAX = 2000

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Writes the graph structure of AIG for DOT.]

  Description [Useful for graph visualization using tools such as GraphViz: 
  http://www.graphviz.org/]
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Writes the graph structure of AIG for DOT.]

  Description [Useful for graph visualization using tools such as GraphViz: 
  http://www.graphviz.org/]
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Writes the graph structure of AIG for DOT.]

  Description [Useful for graph visualization using tools such as GraphViz: 
  http://www.graphviz.org/]
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Returns DFS ordered array of objects and their levels.]

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
func Gia_ManShow(pMan *Gia_Man_t, vBold *Vec_Int_t, fAdders int32, fFadds int32, fPath int32) {
    extern void Abc_ShowFile( char * FileNameDot );
    var FileNameDot string // was [200]byte in C

    var pFile *os.File;
    var vXors *Vec_Int_t = nil;
    var vAdds *Vec_Int_t;
    if fAdders != 0 {
        vAdds = Ree_ManComputeCuts( pMan, &vXors, 0 );
    } else {
        vAdds = nil;
    }
    FileNameDot = fmt.Sprintf("%s", Extra_FileNameGenericAppend(pMan.pName, ".dot"));
    // check that the file can be opened
    if (pFile = fopen( FileNameDot, "w" )) == nil {
        fprintf( os.Stdout, "Cannot open the intermediate file \"%s\".\n", FileNameDot );
        return;
    }
    _ = (pFile).Close();
    // generate the file
    if fPath != 0 {
        Gia_ShowPath( pMan, FileNameDot );
    } else if fAdders != 0 {
        Gia_ShowProcess( pMan, FileNameDot, vBold, vAdds, vXors, fFadds );
    } else {
        Gia_WriteDotAigSimple( pMan, FileNameDot, vBold ); }
    // visualize the file 
    Abc_ShowFile( FileNameDot );

    Vec_IntFreeP( &vAdds );
    Vec_IntFreeP( &vXors );
}





////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

