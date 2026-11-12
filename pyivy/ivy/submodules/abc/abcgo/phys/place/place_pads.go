package place

import (
	"fmt"
	"math"
	"os"
)

/*===================================================================*/
//  
//     place_pads.c
//
//        Aaron P. Hurst, 2003-2007
//              ahurst@eecs.berkeley.edu
//
/*===================================================================*/

//#include <stdio.h>
//#include <stdlib.h>
//#include <math.h>
//#include <limits.h>

//#include "place_base.h"

//ABC_NAMESPACE_IMPL_START


// --------------------------------------------------------------------
// globalPreplace()
//
/// \brief Place pad ring, leaving a core area to meet a desired utilization.
//
/// Sets the position of pads that aren't already fixed.
///
/// Computes g_place_coreBounds and g_place_padBounds.  Determines
/// g_place_rowHeight.
//
// --------------------------------------------------------------------
func globalPreplace(utilization float32) {
  var i, c, h, numRows int32;
  var coreArea, totalArea float32;
  var padCount int32;
  var area float32;
  var padCells **ConcreteCell  = nil;
  var padType *AbstractCell  = nil;
  var cell *ConcreteCell;
  var nextPos float32;
  var remainingPads, northPads, southPads, eastPads, westPads int32;

  fmt.Printf("PLAC-00 : Placing IO pads\n");;

  // identify the pads and compute the total core area
  g_place_coreBounds.x = g_place_coreBounds.y = 0;
  g_place_coreBounds.w = g_place_coreBounds.h = -INT_MAX;

  for c=0; c<g_place_numCells; c++ { if g_place_concreteCells[c] {
    cell = g_place_concreteCells[c];
    area = getCellArea(cell);
    if cell.m_parent.m_pad {
      padType = cell.m_parent;
    } else {
      coreArea += area;
      g_place_rowHeight = cell.m_parent.m_height;
    }

    if cell.m_fixed {
      if g_place_coreBounds.x < cell.m_x {
          g_place_coreBounds.x = g_place_coreBounds.x;
      } else {
          g_place_coreBounds.x = cell.m_x;
      }
      if g_place_coreBounds.y < cell.m_y {
          g_place_coreBounds.y = g_place_coreBounds.y;
      } else {
          g_place_coreBounds.y = cell.m_y;
      }
      if g_place_coreBounds.w > cell.m_x {
          g_place_coreBounds.w = g_place_coreBounds.w;
      } else {
          g_place_coreBounds.w = cell.m_x;
      }
      if g_place_coreBounds.h > cell.m_y {
          g_place_coreBounds.h = g_place_coreBounds.h;
      } else {
          g_place_coreBounds.h = cell.m_y;
      }
    } else if cell.m_parent.m_pad {
      padCells = realloc(padCells, sizeof(ConcreteCell **)*(padCount+1));
      padCells[padCount++] = cell;
    }
    totalArea += area;
  }
  }
  if !(padType != nil) {
    fmt.Printf("ERROR: No pad cells\n");
    os.Exit(int(1));
  }
  g_place_padBounds.w -= g_place_padBounds.x;
  g_place_padBounds.h -= g_place_padBounds.y;

  coreArea /= utilization;

  // create the design boundaries
  numRows = math.Sqrt(float64(coreArea))/g_place_rowHeight+1;
  h = numRows * g_place_rowHeight;
  if g_place_coreBounds.h > h {
      g_place_coreBounds.h = g_place_coreBounds.h;
  } else {
      g_place_coreBounds.h = h;
  }
  if g_place_coreBounds.w > coreArea/g_place_coreBounds.h {
      g_place_coreBounds.w = g_place_coreBounds.w;
  } else {
      g_place_coreBounds.w = coreArea/g_place_coreBounds.h;
  }
  // increase the dimensions by the width of the padring
  g_place_padBounds = g_place_coreBounds;
  if padCount != 0 {
    fmt.Printf("PLAC-05 : \tpreplacing %d pad cells\n", padCount);
    g_place_padBounds.x -= padType.m_width;
    g_place_padBounds.y -= padType.m_height;
    g_place_padBounds.w = g_place_coreBounds.w+2*padType.m_width;
    g_place_padBounds.h = g_place_coreBounds.h+2*padType.m_height;
  }

  fmt.Printf("PLAC-05 : \tplaceable rows  : %d\n", numRows);
  fmt.Printf("PLAC-05 : \tcore dimensions : %.0fx%.0f\n", g_place_coreBounds.w, g_place_coreBounds.h);
  fmt.Printf("PLAC-05 : \tchip dimensions : %.0fx%.0f\n", g_place_padBounds.w, g_place_padBounds.h);
  
  remainingPads = padCount;
  c = 0;

  // north pads
  northPads = remainingPads/4; remainingPads -= northPads;
  nextPos = 0;
  for i=0; i<northPads; i++ {
    cell = padCells[c]; c++;
    cell.m_x = g_place_padBounds.x+cell.m_parent.m_width*0.5 + nextPos;
    cell.m_y = g_place_padBounds.y+cell.m_parent.m_height*0.5;
    nextPos += (g_place_padBounds.w-padType.m_width) / northPads;
  }
  
  // south pads
  southPads = remainingPads/3; remainingPads -= southPads;
  nextPos = 0;
  for i=0; i<southPads; i++ {
    cell = padCells[c]; c++;
    cell.m_x = g_place_padBounds.w+g_place_padBounds.x-cell.m_parent.m_width*0.5 - nextPos;
    cell.m_y = g_place_padBounds.h+g_place_padBounds.y-cell.m_parent.m_height*0.5;
    nextPos += (g_place_padBounds.w-2*padType.m_width) / southPads;
  }

  // east pads
  eastPads = remainingPads/2; remainingPads -= eastPads;
  nextPos = 0;
  for i=0; i<eastPads; i++ {
    cell = padCells[c]; c++;
    cell.m_x = g_place_padBounds.w+g_place_padBounds.x-cell.m_parent.m_width*0.5;
    cell.m_y = g_place_padBounds.y+cell.m_parent.m_height*0.5 + nextPos;
    nextPos += (g_place_padBounds.h-padType.m_height) / eastPads;
  }

  // west pads
  westPads = remainingPads;
  nextPos = 0;
  for i=0; i<westPads; i++ {
    cell = padCells[c]; c++;
    cell.m_x = g_place_padBounds.x+cell.m_parent.m_width*0.5;
    cell.m_y = g_place_padBounds.h+g_place_padBounds.y-cell.m_parent.m_height*0.5 - nextPos;
    nextPos += (g_place_padBounds.h-padType.m_height) / westPads;
  }

}

//ABC_NAMESPACE_IMPL_END

