package hop

import "fmt"

/**CFile****************************************************************

  FileName    [hopCheck.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Minimalistic And-Inverter Graph package.]

  Synopsis    [AIG checking procedures.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - May 11, 2006.]

  Revision    [$Id: hopCheck.c,v 1.00 2006/05/11 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "hop.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/*
*Function*************************************************************

	Synopsis    [Checks the consistency of the AIG manager.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Hop_ManCheck(p *Hop_Man_t) int32 {
	// Check primary inputs and outputs in their manager order.
	for i := int32(0); i < Vec_PtrSize(p.vPis); i++ {
		obj := Vec_PtrEntry(p.vPis, i).(*Hop_Obj_t)
		if Hop_ObjFanin0(obj) != nil || Hop_ObjFanin1(obj) != nil {
			fmt.Printf("Hop_ManCheck: The PI node %p has fanins.\n", obj)
			return 0
		}
	}
	for i := int32(0); i < Vec_PtrSize(p.vPos); i++ {
		obj := Vec_PtrEntry(p.vPos, i).(*Hop_Obj_t)
		if Hop_ObjFanin0(obj) == nil {
			fmt.Printf("Hop_ManCheck: The PO node %p has NULL fanin.\n", obj)
			return 0
		}
		if Hop_ObjFanin1(obj) != nil {
			fmt.Printf("Hop_ManCheck: The PO node %p has second fanin.\n", obj)
			return 0
		}
	}
	for _, head := range p.pTable {
		for obj := head; obj != nil; obj = obj.pNext {
			a, b := Hop_ObjFanin0(obj), Hop_ObjFanin1(obj)
			if a == nil || b == nil {
				fmt.Printf("Hop_ManCheck: The AIG has internal node %p with a NULL fanin.\n", obj)
				return 0
			}
			if a.Id >= b.Id {
				fmt.Printf("Hop_ManCheck: The AIG has node %p with a wrong ordering of fanins.\n", obj)
				return 0
			}
			if Hop_TableLookup(p, obj) != obj {
				fmt.Printf("Hop_ManCheck: Node %p is not in the structural hashing table.\n", obj)
				return 0
			}
		}
	}
	if Hop_ManObjNum(p) != 1+Hop_ManPiNum(p)+Hop_ManPoNum(p)+Hop_ManAndNum(p)+Hop_ManExorNum(p) {
		fmt.Println("Hop_ManCheck: The number of created nodes is wrong.")
		return 0
	}
	if Hop_TableCountEntries(p) != Hop_ManAndNum(p)+Hop_ManExorNum(p) {
		fmt.Println("Hop_ManCheck: The number of nodes in the structural hashing table is wrong.")
		return 0
	}
	return 1
}
