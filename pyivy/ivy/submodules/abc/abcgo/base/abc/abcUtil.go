package abc

/**CFile****************************************************************

  FileName    [abcUtil.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Network and node package.]

  Synopsis    [Various utilities.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: abcUtil.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "abc.h"
//#include "base/main/main.h"
//#include "map/mio/mio.h"
//#include "bool/dec/dec.h"
//#include "opt/fxu/fxu.h"
//#include "aig/miniaig/ndr.h"

// CUDD functionality is provided by the Go core package.

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Frees one attribute manager.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Order CI/COs.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Reads the number of cubes of the node.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Reads the number of cubes of the node.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Reads the number of literals in the SOPs of the nodes.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Counts the number of literals in the factored forms.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Counts the number of nodes with more than 1 reference.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Reads the number of BDD nodes.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Reads the number of BDD nodes.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Reads the number of BDD nodes.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Computes the area of the mapped circuit.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Counts the number of exors.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Counts the number of exors.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Counts the number of exors.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Counts the number of exors.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Returns 1 if it is an AIG with choice nodes.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Reads the maximum number of fanins.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Reads the total number of all fanins.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Cleans the copy field of all objects.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Cleans the copy field of all objects.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Cleans the copy field of all objects.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Counts the number of nodes having non-trivial copies.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Saves copy field of the objects.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Loads copy field of the objects.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Cleans the copy field of all objects.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Cleans the copy field of all objects.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Cleans the copy field of all objects.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Cleans the copy field of all objects.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Cleans the copy field of all objects.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Cleans the copy field of all objects.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Returns the index of the given fanin.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Checks if the internal node has CO fanout.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Checks if the internal node has CO fanout.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Checks if the internal node has CO drivers with the same name.]

  Description [Checks if the internal node can borrow its name from CO fanouts.
  This is possible if all COs with non-complemented fanin edge pointing to this
  node have the same name.]

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Fixes the CO driver problem.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Returns 1 if COs of a logic network are simple.]

  Description [The COs of a logic network are simple under three conditions:
  (1) The edge from CO to its driver is not complemented.
  (2) If CI is a driver of a CO, they have the same name.]
  (3) If two COs share the same driver, they have the same name.]

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Transforms the network to have simple COs.]

  Description [The COs of a logic network are simple under three conditions:
  (1) The edge from CO to its driver is not complemented.
  (2) If CI is a driver of a CO, they have the same name.]
  (3) If two COs share the same driver, they have the same name.
  In some cases, such as FPGA mapping, we prevent the increase in delay
  by duplicating the driver nodes, rather than adding invs/bufs.]

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Transforms the network to have simple COs.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Transforms the network to have simple COs.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Inserts a new node in the order by levels.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Returns 1 if the node is the root of EXOR/NEXOR.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Returns 1 if the node is the root of MUX or EXOR/NEXOR.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Returns 1 if the node is the root of MUX or EXOR/NEXOR.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Returns 1 if the node is the control type of the MUX.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Recognizes what nodes are control and data inputs of a MUX.]

  Description [If the node is a MUX, returns the control variable C.
  Assigns nodes T and E to be the then and else variables of the MUX.
  Node C is never complemented. Nodes T and E can be complemented.
  This function also recognizes EXOR/NEXOR gates as MUXes.]

  SideEffects []

  SeeAlso     []

***********************************************************************/

/*
*Function*************************************************************

	Synopsis    [Prepares two network for a two-argument command similar to "verify".]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/

/**Function*************************************************************

  Synopsis    [Returns 1 if it is an AIG with choice nodes.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Returns 1 if it is an AIG with choice nodes.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Collects all latches in the network.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Procedure used for sorting the nodes in increasing order of levels.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Procedure used for sorting the nodes in decreasing order of levels.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Creates the array of fanout counters.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Collects all objects into one array.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Returns the array of CI IDs.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Puts the nodes into the DFS order and reassign their IDs.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/*
*Function*************************************************************

	Synopsis    [Detect cases when non-trivial FF matching is possible.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/

/*
*Function*************************************************************

	Synopsis    [Compares the pointers.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/

/**Function*************************************************************

  Synopsis    [Adjusts the copy pointers.]

  Description [This procedure assumes that the network was transformed
  into another network, which was in turn transformed into yet another
  network. It makes the pCopy pointers of the original network point to
  the objects of the yet another network.]

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Increaments the cut counter.]

  Description [Returns 1 if it becomes equal to the ref counter.]

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Computes cross-cut of the circuit.]

  Description [Returns 1 if it is the last visit to the node.]

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Computes cross-cut of the circuit.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Prints all 3-var functions.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Compares the supergates by their level.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Analyze choice node support.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Analyze choice node support.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Complements the constraint outputs.]

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

/**Function*************************************************************

  Synopsis    [Returns 1 if all other fanouts of pFanin are below pNode.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Returns 1 if there exist a fanout of pFanin higher than pNode.]

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

/**Function*************************************************************

  Synopsis    [Computes max delay using log(n) delay model.]

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

/**Function*************************************************************

  Synopsis    [Creates precomputed reverse topological order for each node.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Converts multi-output PLA into an AIG with logic sharing.]

  Description [The first argument is an array of char*-strings representing
  individual output of a multi-output PLA. The number of inputs (nInputs)
  and the number of outputs (nOutputs) are the second and third arguments.
  This procedure returns the AIG manager with the given number of inputs
  and outputs representing the PLA as a logic network with sharing.

  For example, if the original PLA is
    1000 10
    0110 01
    0011 01
  the individual PLA for each the two outputs should be
    1000 1
  and
    0110 1
    0011 1

  Reprsentation in terms of two char*-strings will be:
    char * pPlas[2] = { "1000 1\n", "0110 1\n0011 1\n" };
  The call to the procedure may look as follows:
    Abc_Ntk_t * pNtkAig = Abc_NtkFromPla( pPlas, 4, 2 );]

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Checks if the logic network is in the topological order.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Checks if the logic network is in the topological order.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Transfers phase information to the new network.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Starts a new network using existing network as a model.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/*
*Function*************************************************************

	Synopsis    [Derives the AIG representing a property.]

	Description [Given is a sequential logic network (Abc_Ntk_t) and
	an array of nodes (vector of object IDs) and their values (vector of 0s or 1s).
	This procedure creates a sequential AIG (Abc_Ntk_t), which can be given to a
	sequential model checker (in particular "pdr") to prove that the given
	combination of values never appears at the intenal nodes of the network,
	or produce a counter-example showing that it can appear.]

	SideEffects []

	SeeAlso     []

**********************************************************************
*/

/*
*Function*************************************************************

	Synopsis    [Testbench.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

func Abc_SopSynthesize(vSops *Vec_Ptr_t) *Vec_Wec_t {
	var pNtk *Abc_Ntk_t = Abc_NtkCreateFromSops("top", vSops)
	var pNtkNew *Abc_Ntk_t
	Abc_FrameReplaceCurrentNetwork(Abc_FrameReadGlobalFrame(), pNtk)
	//Cmd_CommandExecute( Abc_FrameGetGlobalFrame(), "fx; strash; balance; dc2; map -a" );
	Abc_FrameSetBatchMode(1)
	Cmd_CommandExecute(Abc_FrameGetGlobalFrame(), "st; collapse; sop; fx; strash; &get; &ps; &deepsyn -I 4 -J 50 -T 5 -S 111 -t; &ps; &put; map -a")
	Abc_FrameSetBatchMode(0)
	pNtkNew = Abc_FrameReadNtk(Abc_FrameReadGlobalFrame())
	return abcSerializeMapped(pNtkNew)
}
func Abc_GiaSynthesize(vGias *Vec_Ptr_t, pMulti *Gia_Man_t) *Vec_Wec_t {
	var pNtk *Abc_Ntk_t = Abc_NtkCreateFromGias("top", vGias, pMulti)
	var pNtkNew *Abc_Ntk_t
	Abc_FrameReplaceCurrentNetwork(Abc_FrameReadGlobalFrame(), pNtk)
	Abc_FrameSetBatchMode(1)
	Cmd_CommandExecute(Abc_FrameGetGlobalFrame(), "compress2rs; dch; map -a;  strash; compress2rs; dch; map -a;  strash; compress2rs; dch; map -a")
	Abc_FrameSetBatchMode(0)
	pNtkNew = Abc_FrameReadNtk(Abc_FrameReadGlobalFrame())
	return abcSerializeMapped(pNtkNew)
}
func Abc_GiaSynthesizeInter(p *Gia_Man_t) *Gia_Man_t {
	var pNtkNew, pNtk *Abc_Ntk_t
	var vGias *Vec_Ptr_t = Vec_PtrAlloc(1)
	Vec_PtrPush(vGias, p)
	pNtk = Abc_NtkCreateFromGias("top", vGias, nil)
	Vec_PtrFree(vGias)
	Abc_FrameReplaceCurrentNetwork(Abc_FrameReadGlobalFrame(), pNtk)
	Cmd_CommandExecute(Abc_FrameGetGlobalFrame(), "balance; collapse; muxes; strash; dc2")
	pNtkNew = Abc_FrameReadNtk(Abc_FrameReadGlobalFrame())
	return Abc_NtkClpGia(pNtkNew)
}

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/

func Abc_SopSynthesizeOne(pSop string, fClp int32) *Gia_Man_t {
	var pNtkNew, pNtk *Abc_Ntk_t
	var vSops *Vec_Ptr_t
	if len(pSop) == 3 {
		var pNew *Gia_Man_t = Gia_ManStart(1)
		pNew.pName = Abc_UtilStrsav("top")
		//Gia_ManAppendCi( pNew );
		assert(pSop[1] == '0' || pSop[1] == '1', "pSop[1] == '0' || pSop[1] == '1'")
		Gia_ManAppendCo(pNew, pSop[1] == '1')
		return pNew
	}
	vSops = Vec_PtrAlloc(1)
	Vec_PtrPush(vSops, pSop)
	pNtk = Abc_NtkCreateFromSops("top", vSops)
	Vec_PtrFree(vSops)
	Abc_FrameReplaceCurrentNetwork(Abc_FrameReadGlobalFrame(), pNtk)
	Abc_FrameSetBatchMode(1)
	if fClp != 0 {
		Cmd_CommandExecute(Abc_FrameGetGlobalFrame(), "clp; sop")
	}
	Cmd_CommandExecute(Abc_FrameGetGlobalFrame(), "fx; strash; balance; dc2")
	Abc_FrameSetBatchMode(0)
	pNtkNew = Abc_FrameReadNtk(Abc_FrameReadGlobalFrame())
	return Abc_NtkStrashToGia(pNtkNew)
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END
