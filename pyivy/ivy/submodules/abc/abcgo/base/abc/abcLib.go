package abc

import "github.com/glycerine/abcgo/misc/util"

/**CFile****************************************************************

  FileName    [abcLib.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Network and node package.]

  Synopsis    [Functions to manipulate verilog libraries.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: abcLib.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "abc.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Create the library.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Removes all pointers to the manager.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Frees the library.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Duplicated the library.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Frees the library.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Prints the library.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Create the library.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Create the library.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Frees the library.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Detects the top-level models.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Derive the AIG of the logic in the netlist.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
// Abc_NodeStrashUsingNetwork_rec maps a netlist node and its fanins into dst.
func Abc_NodeStrashUsingNetwork_rec(dst *Network, object *Object) {
	if object == nil {
		panic("nil hierarchy node")
	}
	if getCopy(object).Object != nil {
		return
	}
	if object.Kind == ObjectNet {
		if len(object.fanins) != 1 {
			panic("net requires one driver")
		}
		edge := object.fanins[0]
		Abc_NodeStrashUsingNetwork_rec(dst, edge.Object)
		mapped := getCopy(edge.Object)
		if edge.Complement {
			mapped = mapped.Not()
		}
		setCopy(object, mapped)
		return
	}
	if object.Kind != ObjectNode {
		panic("unmapped hierarchy terminal")
	}
	for _, edge := range object.fanins {
		Abc_NodeStrashUsingNetwork_rec(dst, edge.Object)
	}
	mapped := Abc_ConvertAigToAig(dst, object)
	setCopy(object, mapped)
	for _, output := range object.fanouts {
		if output.Kind == ObjectNet {
			setCopy(output, mapped)
		}
	}
}

// Abc_NodeStrashUsingNetwork expands a combinational box model into dst.
func Abc_NodeStrashUsingNetwork(dst *Network, box *Object) {
	if box == nil || Abc_ObjIsBox(box) == 0 {
		panic("expected hierarchy box")
	}
	model, ok := box.Data.(*Network)
	if !ok || model == nil || Abc_NtkIsNetlist(model) == 0 || Abc_NtkLatchNum(model) != 0 {
		panic("expected combinational netlist model")
	}
	inputs, outputs := model.ObjectsOfKind(ObjectPI), model.ObjectsOfKind(ObjectPO)
	if len(inputs) != len(box.fanins) || len(outputs) != len(box.fanouts) {
		panic("hierarchy interface size mismatch")
	}
	clearCopies(model)
	setCopy(model.constant, dst.One())
	for i, input := range inputs {
		edge := box.fanins[i]
		value := getCopy(edge.Object)
		if value.Object == nil {
			panic("unmapped box input")
		}
		if edge.Complement {
			value = value.Not()
		}
		if i/32 < len(box.BoxInputPolarity) && util.Abc_InfoHasBit(box.BoxInputPolarity, int32(i)) {
			value = value.Not()
		}
		setCopy(input, value)
		for _, net := range input.fanouts {
			if net.Kind == ObjectNet {
				setCopy(net, value)
			}
		}
	}
	for i, output := range outputs {
		if len(output.fanins) != 1 {
			panic("model output requires one driver")
		}
		edge := output.fanins[0]
		Abc_NodeStrashUsingNetwork_rec(dst, edge.Object)
		value := getCopy(edge.Object)
		if edge.Complement {
			value = value.Not()
		}
		setCopy(box.fanouts[i], value)
	}
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END
