package place

import (
	"fmt"
	"strconv"
	"strings"
)

import "os"

/*===================================================================*/
//  
//     place_test.c
//
//        Aaron P. Hurst, 2003-2007
//              ahurst@eecs.berkeley.edu
//
/*===================================================================*/

//#include <stdlib.h>
//#include <stdio.h>
//#include <string.h>
//#include <assert.h>
//#include "place_base.h"

//ABC_NAMESPACE_IMPL_START



// --------------------------------------------------------------------
// Hash type/functions
//
// --------------------------------------------------------------------

struct hash_element {
  ConcreteCell        *obj;
  struct hash_element *next;
}
 hash_element;

func hash_string(hash_max int32, str string) int32 {
  unsigned int hash = 0;
  var p int32;
  for p = 0; p<len(str); p++ {
    hash += str[p]*p;
  }
  return hash % hash_max;
}

void hash_add(struct hash_element **hash, int hash_max,
              ConcreteCell *cell) {
  var key int32  = hash_string(hash_max, cell.m_label);
  // printf("adding %s key = %d\n", cell->m_label, key);
  struct hash_element *element = malloc(sizeof(struct hash_element));
  assert(element, "element");
  element.obj = cell;
  element.next = hash[key];
  hash[key] = element;
}

ConcreteCell *hash_find(struct hash_element **hash, int hash_max, char *str) {
  var key int32  = hash_string(hash_max, str);
  // printf("looking for %s key = %d\n", str, key);
  struct hash_element *next = hash[key];
  for next {
    if !(int32(strings.Compare(str, next.obj.m_label)) != 0) {
      return next.obj; }
    next = next.next;
  }
  return 0;
}

// --------------------------------------------------------------------
// Global variables
//
// --------------------------------------------------------------------

struct hash_element **hash_cellname;

var numCells, numNets int32;

var abstractCells *AbstractCell;
var concreteCells *ConcreteCell;
var concreteNets *ConcreteNet;

// --------------------------------------------------------------------
// Function implementations
//
// --------------------------------------------------------------------

func readBookshelfNets(filename string) {
  var tok string;
  var buf [1024]byte;
  char *DELIMITERS = " \n\t:";
  var id int32;
  var t int32;
  var cell *ConcreteCell;

  var netsFile *os.File  = os.Open(filename);
  if !(netsFile != nil) {
    fmt.Printf("ERROR: Could not open .nets file\n");
    os.Exit(int(1));
  }

  // line 1 : version
  for fgets(buf, 1024, netsFile) && (buf[0] == '\n' || buf[0] == '#') {; }

  // line 2 : number of nets
  for fgets(buf, 1024, netsFile) && (buf[0] == '\n' || buf[0] == '#') {; }
  tok = strtok(buf, DELIMITERS);
  tok = strtok(nil, DELIMITERS);
  numNets = strconv.Atoi(tok);
  fmt.Printf("READ-20 : number of nets = %d\n", numNets);
  concreteNets = malloc(sizeof(ConcreteNet)*numNets);

  // line 3 : number of pins
  for fgets(buf, 1024, netsFile) && (buf[0] == '\n' || buf[0] == '#') {; }

  // line XXX : net definitions
  for fgets(buf, 1024, netsFile) {
    if buf[0] == '\n' || buf[0] == '#' { continue; }

    concreteNets[id].m_id = id;
    concreteNets[id].m_weight = 1.0;

    tok = strtok(buf, DELIMITERS);
    if !!(int32(strings.Compare(tok, "NetDegree")) != 0) {
      printf("%s\n",buf);
      fmt.Printf("ERROR: Incorrect format in .nets file\n");
      os.Exit(int(1));
    }

    tok = strtok(nil, DELIMITERS);
    concreteNets[id].m_numTerms = strconv.Atoi(tok);
    if concreteNets[id].m_numTerms < 0 ||
        concreteNets[id].m_numTerms > 100000 {
      fmt.Printf("ERROR: Bad net degree\n");
      os.Exit(int(1)); 
    }
    concreteNets[id].m_terms = malloc(sizeof(ConcreteCell*)*
             concreteNets[id].m_numTerms);

    // read terms
    t = 0;
    for t < concreteNets[id].m_numTerms &&
          fgets(buf, 1024, netsFile) {
      if buf[0] == '\n' || buf[0] == '#' { continue; }
      
      // cell name
      tok = strtok(buf, DELIMITERS);
      cell = hash_find(hash_cellname, numCells, tok);
      if !(cell != nil) {
        fmt.Printf("ERROR: Could not find cell %s in .nodes file\n", tok);
        os.Exit(int(1));
      }
      concreteNets[id].m_terms[t] = cell;
      t++;
    }

    // add!
    addConcreteNet(&(concreteNets[id]));

    id++;
  }

  _ = (netsFile).Close();
}

func readBookshelfNodes(filename string) {
  var tok string;
  var buf [1024]byte;
  char *DELIMITERS = " \n\t:";
  var id int32;

  var nodesFile *os.File  = os.Open(filename);
  if !(nodesFile != nil) {
    fmt.Printf("ERROR: Could not open .nodes file\n");
    os.Exit(int(1));
  }

  // line 1 : version
  for fgets(buf, 1024, nodesFile) && (buf[0] == '\n' || buf[0] == '#') {; }

  // line 2 : num nodes
  for fgets(buf, 1024, nodesFile) && (buf[0] == '\n' || buf[0] == '#') {; }
  tok = strtok(buf, DELIMITERS);
  tok = strtok(nil, DELIMITERS);
  numCells = strconv.Atoi(tok);
  fmt.Printf("READ-10 : number of cells = %d\n", numCells);
  concreteCells = malloc(sizeof(ConcreteCell)*numCells);
  abstractCells = malloc(sizeof(AbstractCell)*numCells);
  hash_cellname = calloc(numCells, sizeof(struct hash_element*));

  // line 3 : num terminals
  for fgets(buf, 1024, nodesFile) && (buf[0] == '\n' || buf[0] == '#') {; }

  // line XXX : cell definitions
  for fgets(buf, 1024, nodesFile) {
    if buf[0] == '\n' || buf[0] == '#' { continue; }

    tok = strtok(buf, DELIMITERS);
    concreteCells[id].m_id = id;;
    
    // label
    concreteCells[id].m_parent = &(abstractCells[id]);
    concreteCells[id].m_label = malloc(sizeof(char)*len(tok)+1);
    strcpy(concreteCells[id].m_label, tok);
    abstractCells[id].m_label = concreteCells[id].m_label;
    hash_add(hash_cellname, numCells, 
             &(concreteCells[id]));

    // dimensions
    tok = strtok(nil, DELIMITERS);
    abstractCells[id].m_width = atof(tok);
    tok = strtok(nil, DELIMITERS);
    abstractCells[id].m_height = atof(tok);
    tok = strtok(nil, DELIMITERS);
    // terminal
    abstractCells[id].m_pad = tok && !int32(strings.Compare(tok, "terminal"));

    // add!
    addConcreteCell(&(concreteCells[id]));

    // DEBUG
    /*
    printf("\"%s\" : %f x %f\n", concreteCells[id].m_label,
           abstractCells[id].m_width,
           abstractCells[id].m_height);
    */
    id++;
  }

  _ = (nodesFile).Close();
}

func readBookshelfPlacement(filename string) {
  var tok string;
  var buf [1024]byte;
  char *DELIMITERS = " \n\t:";
  var cell *ConcreteCell;

  var plFile *os.File  = os.Open(filename);
  var netsFile *os.File  = os.Open(filename);
  if !(plFile != nil) {
    fmt.Printf("ERROR: Could not open .pl file\n");
    os.Exit(int(1));
  }
  if !(netsFile != nil) {
    fmt.Printf("ERROR: Could not open .nets file\n");
    os.Exit(int(1));
  }

  // line 1 : version
  for fgets(buf, 1024, plFile) && (buf[0] == '\n' || buf[0] == '#') {; }

  // line XXX : placement definitions
  for fgets(buf, 1024, plFile) {
    if buf[0] == '\n' || buf[0] == '#' { continue; }

    tok = strtok(buf, DELIMITERS);

    // cell name
    cell = hash_find(hash_cellname, numCells, tok);
    if !(cell != nil) {
      fmt.Printf("ERROR: Could not find cell %s in .nodes file\n", tok);
      os.Exit(int(1));
    }

    // position
    tok = strtok(nil, DELIMITERS);
    cell.m_x = atof(tok);
    tok = strtok(nil, DELIMITERS);
    cell.m_y = atof(tok);

    // hfixed
    cell.m_fixed = strtok(nil, DELIMITERS) && 
      (tok = strtok(nil, DELIMITERS)) &&
      !int32(strings.Compare(tok, "\\FIXED"));
  }
  
  _ = (plFile).Close();
}

func writeBookshelfPlacement(filename string) {
  var c int32;

  var plFile *os.File  = os.OpenFile(filename, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
  if !(plFile != nil) {
    fmt.Printf("ERROR: Could not open .pl file\n");
    os.Exit(int(1));
  }

  fmt.Fprintf(plFile, "UCLA pl 1.0\n");
  for c=0; c<numCells; c++ {
    fmt.Fprintf(plFile, "%s %f %f : N %s\n", concreteCells[c].m_label, concreteCells[c].m_x, concreteCells[c].m_y, (concreteCells[c].m_fixed ? "\\FIXED" : ""));
  }

  _ = (plFile).Close();
}

// deletes all connections to a cell
func delNetConnections(cell *ConcreteCell) {
  var n, t, t2, count int32;
  var old **ConcreteCell  = malloc(sizeof(ConcreteCell*)*g_place_numCells);

  for n=0; n<g_place_numNets; n++ { if g_place_concreteNets[n] {
    var net *ConcreteNet  = g_place_concreteNets[n];
    count = 0;
    for t=0; t<net.m_numTerms; t++ {
      if net.m_terms[t] == cell { count++; }
    }
    if count != 0 {
      memcpy(old, net.m_terms, sizeof(ConcreteCell*)*net.m_numTerms);
      net.m_terms = realloc(net.m_terms, 
                                   sizeof(ConcreteCell*)*(net.m_numTerms-count));
      t2 = 0;
      for t=0; t<net.m_numTerms; t++ {
        if old[t] != cell { net.m_terms[t2++] = old[t]; }
      }
      net.m_numTerms -= count;
    }
  }
  }
  free(old);
}

func main() {

  if len(os.Args) != 4 {
    fmt.Printf("Usage: %s [nodes] [nets] [pl]\n", os.Args[0]);
    os.Exit(int(1));
  }

  readBookshelfNodes(os.Args[1]);
  readBookshelfNets(os.Args[2]);
  readBookshelfPlacement(os.Args[3]);

  globalPreplace(0.8);
  globalPlace();

  // DEBUG net/cell removal/addition
  /*
  int i;
  for(i=1000; i<2000; i++) {
    delConcreteNet(g_place_concreteNets[i]);
    delNetConnections(g_place_concreteCells[i]);
    delConcreteCell(g_place_concreteCells[i]);
  }
  
  ConcreteCell newCell[2];
  newCell[0].m_id = g_place_numCells+1;
  newCell[0].m_x = 1000;
  newCell[0].m_y = 1000;
  newCell[0].m_fixed = false;
  newCell[0].m_parent = &(abstractCells[1000]);
  newCell[0].m_label = " ";
  addConcreteCell(&newCell[0]);
  newCell[1].m_id = g_place_numCells+3;
  newCell[1].m_x = 1000;
  newCell[1].m_y = 1000;
  newCell[1].m_fixed = false;
  newCell[1].m_parent = &(abstractCells[1000]);
  newCell[1].m_label = " ";
  addConcreteCell(&newCell[1]);
  */

  globalIncremental();

  writeBookshelfPlacement(os.Args[3]);

  free(hash_cellname);

  return ;
}
//ABC_NAMESPACE_IMPL_END

