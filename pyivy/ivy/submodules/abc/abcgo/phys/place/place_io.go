package place

import "fmt"

import "os"

/*===================================================================*/
//  
//     place_io.c
//
//        Aaron P. Hurst, 2003-2007
//              ahurst@eecs.berkeley.edu
//
/*===================================================================*/

//#include <stdlib.h>
//#include <limits.h>
//#include <assert.h>
//#include <string.h>
//#include <stdio.h>

//#include "place_base.h"

//ABC_NAMESPACE_IMPL_START



// --------------------------------------------------------------------
// writeBookshelfNodes()
//
// --------------------------------------------------------------------
func writeBookshelfNodes(filename string) {
  
  var c int32;
  var numNodes, numTerms int32;

  var nodesFile *os.File  = os.OpenFile(filename, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
  if !(nodesFile != nil) {
    fmt.Printf("ERROR: Could not open .nodes file\n");
    os.Exit(int(1));
  }

  numNodes = numTerms = 0;
  for c=0; c<g_place_numCells; c++ { if g_place_concreteCells[c] {
      numNodes++;
      if g_place_concreteCells[c].m_parent.m_pad {
        numTerms++; }
    }
  }



  fmt.Fprintf(nodesFile, "UCLA nodes 1.0\n");
  fmt.Fprintf(nodesFile, "NumNodes : %d\n", numNodes);
  fmt.Fprintf(nodesFile, "NumTerminals : %d\n", numTerms);

  for c=0; c<g_place_numCells; c++ { if g_place_concreteCells[c] {
    fmt.Fprintf(nodesFile, "CELL%d %f %f %s\n", g_place_concreteCells[c].m_id, g_place_concreteCells[c].m_parent.m_width, g_place_concreteCells[c].m_parent.m_height, (g_place_concreteCells[c].m_parent.m_pad ? " terminal" : ""));
  }
  }

  _ = (nodesFile).Close();  
}


// --------------------------------------------------------------------
// writeBookshelfPl()
//
// --------------------------------------------------------------------
func writeBookshelfPl(filename string) {
  
  var c int32;

  var plFile *os.File  = os.OpenFile(filename, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
  if !(plFile != nil) {
    fmt.Printf("ERROR: Could not open .pl file\n");
    os.Exit(int(1));
  }

  fmt.Fprintf(plFile, "UCLA pl 1.0\n");
  for c=0; c<g_place_numCells; c++ {  if g_place_concreteCells[c] {
    fmt.Fprintf(plFile, "CELL%d %f %f : N %s\n", g_place_concreteCells[c].m_id, g_place_concreteCells[c].m_x, g_place_concreteCells[c].m_y, (g_place_concreteCells[c].m_fixed ? "\\FIXED" : ""));
  }
  }

  _ = (plFile).Close();  

}


// --------------------------------------------------------------------
// writeBookshelf()
//
// --------------------------------------------------------------------
func writeBookshelf(filename string) {
  writeBookshelfNodes("out.nodes");
  writeBookshelfPl("out.pl");
}
//ABC_NAMESPACE_IMPL_END

