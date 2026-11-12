package bbl

/**CFile****************************************************************

  FileName    [bblif.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Binary BLIF representation for logic networks.]

  Synopsis    [External declarations.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - February 28, 2009.]

  Revision    [$Id: bblif.h,v 1.00 2009/02/28 00:00:00 alanmi Exp $]

***********************************************************************/

// Bbl_Type_t describes the combinational object kinds in a BBLIF network.
type Bbl_Type_t int

const (
	BBL_OBJ_NONE Bbl_Type_t = iota
	BBL_OBJ_CI
	BBL_OBJ_CO
	BBL_OBJ_NODE
	BBL_OBJ_VOID
)

type Bbl_Obj_t struct {
	ID         int
	OriginalID int
	Type       Bbl_Type_t
	Fanins     []*Bbl_Obj_t
	SOP        string
	Marked     bool
	Box        bool
	declared   int
	index      int
	function   int
}
type Bbl_Man_t struct {
	Name          string
	Objects       []*Bbl_Obj_t
	byID          map[int]*Bbl_Obj_t
	functions     []string
	handles       map[string]int
	functionBytes int
}
