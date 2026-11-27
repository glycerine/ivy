package dsd

import (
	"fmt"
	"io"

	bdd "github.com/glycerine/abcgo/bdd/cudd/core"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/opt/dau/canon"
)

// dsdTree.c: managing the decomposition tree.

// Dsd_TreeNodeCreate creates the DSD node.
func Dsd_TreeNodeCreate(Type, nDecs, BlockNum int) *Dsd_Node_t {
	// allocate memory for this node
	p := &Dsd_Node_t{}
	p.Type = Dsd_Type_t(Type) // the type of this block
	if nDecs > 0 {
		p.pDecs = make([]*Dsd_Node_t, nDecs) // the number of decompositions
	}
	p.inverse = &Dsd_Node_t{regular: p, inverse: p}
	return p
}

// Dsd_TreeNodeDelete frees the DSD node.
func Dsd_TreeNodeDelete(dd *DdManager, pNode *Dsd_Node_t) {
	if pNode.G != nil {
		bdd.Cudd_RecursiveDeref(dd, pNode.G)
	}
	if pNode.S != nil {
		bdd.Cudd_RecursiveDeref(dd, pNode.S)
	}
	pNode.G, pNode.S = nil, nil
	pNode.pDecs = nil
}

// Dsd_TreeUnmark unmarks the decomposition tree. This function assumes
// that originally pNode->nVisits are set to zero!
func Dsd_TreeUnmark(pDsdMan *Dsd_Manager_t) {
	for i := 0; i < len(pDsdMan.roots); i++ {
		dsdTreeUnmark_rec(Dsd_Regular(pDsdMan.roots[i]))
	}
}

// dsdTreeUnmark_rec should be called with a non-complemented pointer.
func dsdTreeUnmark_rec(pNode *Dsd_Node_t) {
	if pNode.nVisits <= 0 {
		panic("dsd: Dsd_TreeUnmark_rec: pNode->nVisits > 0")
	}
	pNode.nVisits--
	if pNode.nVisits != 0 { // if this is not the last visit, return
		return
	}
	// upon the last visit, go through the list of successors and call recursively
	if pNode.Type != DSD_NODE_BUF && pNode.Type != DSD_NODE_CONST1 {
		for i := 0; i < len(pNode.pDecs); i++ {
			dsdTreeUnmark_rec(Dsd_Regular(pNode.pDecs[i]))
		}
	}
}

// Dsd_TreeNodeGetInfo computes the max depth and the max gate size of the
// trees rooted at the manager's roots.
func Dsd_TreeNodeGetInfo(pDsdMan *Dsd_Manager_t, DepthMax, GateSizeMax *int) {
	s_DepthMax, s_GateSizeMax := 0, 0
	for i := 0; i < len(pDsdMan.roots); i++ {
		dsdTreeGetInfo_rec(Dsd_Regular(pDsdMan.roots[i]), 0, &s_DepthMax, &s_GateSizeMax)
	}
	if DepthMax != nil {
		*DepthMax = s_DepthMax
	}
	if GateSizeMax != nil {
		*GateSizeMax = s_GateSizeMax
	}
}

// Dsd_TreeNodeGetInfoOne computes the max depth and the max gate size of
// the tree rooted at the node.
func Dsd_TreeNodeGetInfoOne(pNode *Dsd_Node_t, DepthMax, GateSizeMax *int) {
	s_DepthMax, s_GateSizeMax := 0, 0
	dsdTreeGetInfo_rec(Dsd_Regular(pNode), 0, &s_DepthMax, &s_GateSizeMax)
	if DepthMax != nil {
		*DepthMax = s_DepthMax
	}
	if GateSizeMax != nil {
		*GateSizeMax = s_GateSizeMax
	}
}

// dsdTreeGetInfo_rec performs the recursive step of Dsd_TreeNodeGetInfo().
func dsdTreeGetInfo_rec(pNode *Dsd_Node_t, RankCur int, s_DepthMax, s_GateSizeMax *int) {
	var GateSize int
	// we don't want the two-input gates to count for non-decomposable blocks
	if pNode.Type == DSD_NODE_OR || pNode.Type == DSD_NODE_EXOR {
		GateSize = 2
	} else {
		GateSize = len(pNode.pDecs)
	}

	// update the max size of the node
	if *s_GateSizeMax < GateSize {
		*s_GateSizeMax = GateSize
	}

	if len(pNode.pDecs) < 2 {
		return
	}

	// update the max rank
	if *s_DepthMax < RankCur+1 {
		*s_DepthMax = RankCur + 1
	}

	// call recursively
	for i := 0; i < len(pNode.pDecs); i++ {
		dsdTreeGetInfo_rec(Dsd_Regular(pNode.pDecs[i]), RankCur+1, s_DepthMax, s_GateSizeMax)
	}
}

// dsdTreeGetAigCost_rec counts AIG nodes needed to implement this node.
func dsdTreeGetAigCost_rec(pNode *Dsd_Node_t) int {
	Counter := 0
	if len(pNode.pDecs) < 2 {
		return 0
	}
	// we don't want the two-input gates to count for non-decomposable blocks
	if pNode.Type == DSD_NODE_OR {
		Counter += len(pNode.pDecs) - 1
	} else if pNode.Type == DSD_NODE_EXOR {
		Counter += 3 * (len(pNode.pDecs) - 1)
	} else if pNode.Type == DSD_NODE_PRIME && len(pNode.pDecs) == 3 {
		Counter += 3
	}
	// call recursively
	for i := 0; i < len(pNode.pDecs); i++ {
		Counter += dsdTreeGetAigCost_rec(Dsd_Regular(pNode.pDecs[i]))
	}
	return Counter
}

// Dsd_TreeGetAigCost counts AIG nodes needed to implement this node.
// Assumes that the only primes of the DSD tree are MUXes.
func Dsd_TreeGetAigCost(pNode *Dsd_Node_t) int {
	return dsdTreeGetAigCost_rec(Dsd_Regular(pNode))
}

// Dsd_TreeCountNonTerminalNodes counts non-terminal nodes of the DSD tree:
// all the nodes with the support more than 1 (OR, EXOR, and PRIME nodes).
func Dsd_TreeCountNonTerminalNodes(pDsdMan *Dsd_Manager_t) int {
	Counter := 0
	for i := 0; i < len(pDsdMan.roots); i++ {
		Counter += dsdTreeCountNonTerminalNodes_rec(Dsd_Regular(pDsdMan.roots[i]))
	}
	Dsd_TreeUnmark(pDsdMan)
	return Counter
}

func Dsd_TreeCountNonTerminalNodesOne(pRoot *Dsd_Node_t) int {
	// go through the list of successors and call recursively
	Counter := dsdTreeCountNonTerminalNodes_rec(Dsd_Regular(pRoot))
	dsdTreeUnmark_rec(Dsd_Regular(pRoot))
	return Counter
}

// dsdTreeCountNonTerminalNodes_rec counts non-terminal nodes for one root.
func dsdTreeCountNonTerminalNodes_rec(pNode *Dsd_Node_t) int {
	Counter := 0
	pNode.nVisits++
	if pNode.nVisits > 1 { // if this is not the first visit, return zero
		return 0
	}
	if len(pNode.pDecs) <= 1 {
		return 0
	}
	// upon the first visit, go through the list of successors and call recursively
	for i := 0; i < len(pNode.pDecs); i++ {
		Counter += dsdTreeCountNonTerminalNodes_rec(Dsd_Regular(pNode.pDecs[i]))
	}
	return Counter + 1
}

// Dsd_TreeCountPrimeNodes counts prime nodes of the DSD tree: nodes with
// the support more than 2, that is not an OR or EXOR gate.
func Dsd_TreeCountPrimeNodes(pDsdMan *Dsd_Manager_t) int {
	Counter := 0
	for i := 0; i < len(pDsdMan.roots); i++ {
		Counter += dsdTreeCountPrimeNodes_rec(Dsd_Regular(pDsdMan.roots[i]))
	}
	Dsd_TreeUnmark(pDsdMan)
	return Counter
}

// Dsd_TreeCountPrimeNodesOne counts prime nodes for one root.
func Dsd_TreeCountPrimeNodesOne(pRoot *Dsd_Node_t) int {
	// go through the list of successors and call recursively
	Counter := dsdTreeCountPrimeNodes_rec(Dsd_Regular(pRoot))
	dsdTreeUnmark_rec(Dsd_Regular(pRoot))
	return Counter
}

func dsdTreeCountPrimeNodes_rec(pNode *Dsd_Node_t) int {
	Counter := 0
	pNode.nVisits++
	if pNode.nVisits > 1 { // if this is not the first visit, return zero
		return 0
	}
	if len(pNode.pDecs) <= 1 {
		return 0
	}
	// upon the first visit, go through the list of successors and call recursively
	for i := 0; i < len(pNode.pDecs); i++ {
		Counter += dsdTreeCountPrimeNodes_rec(Dsd_Regular(pNode.pDecs[i]))
	}
	if pNode.Type == DSD_NODE_PRIME {
		Counter++
	}
	return Counter
}

// Dsd_TreeCollectDecomposableVars collects the decomposable vars on the
// PI side.
func Dsd_TreeCollectDecomposableVars(pDsdMan *Dsd_Manager_t, pVars []int) int {
	// set the vars collected to 0
	nVars := 0
	dsdTreeCollectDecomposableVars_rec(pDsdMan.dd, Dsd_Regular(pDsdMan.roots[0]), pVars, &nVars)
	// return the number of collected vars
	return nVars
}

// dsdTreeCollectDecomposableVars_rec adds decomposable variables as they
// are found to pVars and increments nVars. Returns 1 if a non-dec node with
// more than 4 inputs was encountered in the processed subtree.
func dsdTreeCollectDecomposableVars_rec(dd *DdManager, pNode *Dsd_Node_t, pVars []int, nVars *int) bool {
	if len(pNode.pDecs) <= 1 {
		return false
	}

	// go through the list of successors and call recursively
	fSkipThisNode := false
	for i := 0; i < len(pNode.pDecs); i++ {
		if dsdTreeCollectDecomposableVars_rec(dd, Dsd_Regular(pNode.pDecs[i]), pVars, nVars) {
			fSkipThisNode = true
		}
	}

	if !fSkipThisNode && (pNode.Type == DSD_NODE_OR || pNode.Type == DSD_NODE_EXOR || len(pNode.pDecs) <= 4) {
		for i := 0; i < len(pNode.pDecs); i++ {
			pTemp := Dsd_Regular(pNode.pDecs[i])
			if pTemp.Type == DSD_NODE_BUF {
				if pVars != nil {
					pVars[*nVars] = int(pTemp.S.Index)
				}
				*nVars++
			}
		}
	} else {
		fSkipThisNode = true
	}

	return fSkipThisNode
}

// Dsd_TreeCollectNodesDfs creates the DFS ordered array of DSD nodes in
// the tree. The collected nodes do not include the terminal nodes and the
// constant 1 node.
func Dsd_TreeCollectNodesDfs(pDsdMan *Dsd_Manager_t, pnNodes *int) []*Dsd_Node_t {
	nNodesAlloc := Dsd_TreeCountNonTerminalNodes(pDsdMan)
	ppNodes := make([]*Dsd_Node_t, 0, nNodesAlloc)
	for i := 0; i < len(pDsdMan.roots); i++ {
		dsdTreeCollectNodesDfs_rec(Dsd_Regular(pDsdMan.roots[i]), &ppNodes)
	}
	Dsd_TreeUnmark(pDsdMan)
	if nNodesAlloc != len(ppNodes) {
		panic("dsd: Dsd_TreeCollectNodesDfs: nNodesAlloc == nNodes")
	}
	if pnNodes != nil {
		*pnNodes = len(ppNodes)
	}
	return ppNodes
}

// Dsd_TreeCollectNodesDfsOne creates the DFS ordered array of DSD nodes in
// the tree rooted at pNode.
func Dsd_TreeCollectNodesDfsOne(pDsdMan *Dsd_Manager_t, pNode *Dsd_Node_t, pnNodes *int) []*Dsd_Node_t {
	nNodesAlloc := Dsd_TreeCountNonTerminalNodesOne(pNode)
	ppNodes := make([]*Dsd_Node_t, 0, nNodesAlloc)
	dsdTreeCollectNodesDfs_rec(Dsd_Regular(pNode), &ppNodes)
	dsdTreeUnmark_rec(Dsd_Regular(pNode))
	if nNodesAlloc != len(ppNodes) {
		panic("dsd: Dsd_TreeCollectNodesDfsOne: nNodesAlloc == nNodes")
	}
	if pnNodes != nil {
		*pnNodes = len(ppNodes)
	}
	return ppNodes
}

func dsdTreeCollectNodesDfs_rec(pNode *Dsd_Node_t, ppNodes *[]*Dsd_Node_t) {
	pNode.nVisits++
	if pNode.nVisits > 1 { // if this is not the first visit, return zero
		return
	}
	if len(pNode.pDecs) <= 1 {
		return
	}
	// upon the first visit, go through the list of successors and call recursively
	for i := 0; i < len(pNode.pDecs); i++ {
		dsdTreeCollectNodesDfs_rec(Dsd_Regular(pNode.pDecs[i]), ppNodes)
	}
	*ppNodes = append(*ppNodes, pNode)
}

// Dsd_TreePrint prints the decompostion tree into file.
func Dsd_TreePrint(pFile io.Writer, pDsdMan *Dsd_Manager_t, pInputNames, pOutputNames []string, fShortNames, Output int) {
	var pNode *Dsd_Node_t
	SigCounter := 1
	if Output == -1 {
		for i := 0; i < len(pDsdMan.roots); i++ {
			pNode = Dsd_Regular(pDsdMan.roots[i])
			dsdTreePrint_rec(pFile, pNode, pNode != pDsdMan.roots[i], pInputNames, pOutputNames[i], 0, &SigCounter, fShortNames != 0)
		}
	} else {
		if Output < 0 || Output >= len(pDsdMan.roots) {
			panic("dsd: Dsd_TreePrint: Output >= 0 && Output < pDsdMan->nRoots")
		}
		pNode = Dsd_Regular(pDsdMan.roots[Output])
		dsdTreePrint_rec(pFile, pNode, pNode != pDsdMan.roots[Output], pInputNames, pOutputNames[Output], 0, &SigCounter, fShortNames != 0)
	}
}

// printChar is fprintf( pFile, "%c", c ): the low byte of c.
func printChar(pFile io.Writer, c int) {
	pFile.Write([]byte{byte(c)})
}

func printSpaces(pFile io.Writer, n int) {
	for i := 0; i < n; i++ {
		fmt.Fprint(pFile, " ")
	}
}

// dsdTreePrint_rec prints the decompostion tree into file.
func dsdTreePrint_rec(pFile io.Writer, pNode *Dsd_Node_t, fComp bool, pInputNames []string, pOutputName string, nOffset int, pSigCounter *int, fShortNames bool) {
	if pNode.Type != DSD_NODE_BUF && pNode.Type != DSD_NODE_CONST1 &&
		pNode.Type != DSD_NODE_PRIME && pNode.Type != DSD_NODE_OR && pNode.Type != DSD_NODE_EXOR {
		panic("dsd: Dsd_TreePrint_rec: bad node type")
	}

	printSpaces(pFile, nOffset)

	if !fComp {
		fmt.Fprintf(pFile, "%s = ", pOutputName)
	} else {
		fmt.Fprintf(pFile, "NOT(%s) = ", pOutputName)
	}
	pInputNums := make([]int, len(pNode.pDecs))
	if pNode.Type == DSD_NODE_CONST1 {
		fmt.Fprintf(pFile, " Constant 1.\n")
	} else if pNode.Type == DSD_NODE_BUF {
		if fShortNames {
			fmt.Fprintf(pFile, "%d", 'a'+int(pNode.S.Index))
		} else {
			fmt.Fprintf(pFile, "%s", pInputNames[pNode.S.Index])
		}
		fmt.Fprintf(pFile, "\n")
	} else {
		// PRIME, OR and EXOR print the same way, except for the gate name
		// and for PRIME's short input names, which C prints with "%d"
		var name string
		switch pNode.Type {
		case DSD_NODE_PRIME:
			name = "PRIME("
		case DSD_NODE_OR:
			name = "OR("
		default:
			name = "EXOR("
		}
		// print the line
		fmt.Fprintf(pFile, "%s", name)
		for i := 0; i < len(pNode.pDecs); i++ {
			pInput := Dsd_Regular(pNode.pDecs[i])
			fCompNew := pInput != pNode.pDecs[i]
			if i != 0 {
				fmt.Fprintf(pFile, ",")
			}
			if fCompNew {
				fmt.Fprintf(pFile, " NOT(")
			} else {
				fmt.Fprintf(pFile, " ")
			}
			if pInput.Type == DSD_NODE_BUF {
				pInputNums[i] = 0
				if fShortNames {
					if pNode.Type == DSD_NODE_PRIME {
						fmt.Fprintf(pFile, "%d", pInput.S.Index)
					} else {
						printChar(pFile, 'a'+int(pInput.S.Index))
					}
				} else {
					fmt.Fprintf(pFile, "%s", pInputNames[pInput.S.Index])
				}
			} else {
				pInputNums[i] = *pSigCounter
				*pSigCounter++
				fmt.Fprintf(pFile, "<%d>", pInputNums[i])
			}
			if fCompNew {
				fmt.Fprintf(pFile, ")")
			}
		}
		fmt.Fprintf(pFile, " )\n")
		// call recursively for the following blocks
		for i := 0; i < len(pNode.pDecs); i++ {
			if pInputNums[i] != 0 {
				Buffer := fmt.Sprintf("<%d>", pInputNums[i])
				dsdTreePrint_rec(pFile, Dsd_Regular(pNode.pDecs[i]), false, pInputNames, Buffer, nOffset+6, pSigCounter, fShortNames)
			}
		}
	}
}

var s_Truths6 = [6]uint64{
	0xAAAAAAAAAAAAAAAA,
	0xCCCCCCCCCCCCCCCC,
	0xF0F0F0F0F0F0F0F0,
	0xFF00FF00FF00FF00,
	0xFFFF0000FFFF0000,
	0xFFFFFFFF00000000,
}

// Dsd_TreeFunc2Truth_rec converts a BDD over the top six levels into a
// truth table.
func Dsd_TreeFunc2Truth_rec(dd *DdManager, bFunc *DdNode) uint64 {
	if bFunc == dd.LogicZero {
		return 0
	}
	if bFunc == dd.One {
		return ^uint64(0)
	}
	if bdd.Cudd_IsComplement(bFunc) {
		return ^Dsd_TreeFunc2Truth_rec(dd, bdd.Cudd_Not(bFunc))
	}
	Level := dd.Perm[bFunc.Index]
	if Level < 0 || Level >= 6 {
		panic("dsd: Dsd_TreeFunc2Truth_rec: Level >= 0 && Level < 6")
	}
	Cof0 := Dsd_TreeFunc2Truth_rec(dd, bFunc.E)
	Cof1 := Dsd_TreeFunc2Truth_rec(dd, bFunc.T)
	return (s_Truths6[Level] & Cof1) | (^s_Truths6[Level] & Cof0)
}

func dsdTreePrint2_rec(pFile io.Writer, dd *DdManager, pNode *Dsd_Node_t, fComp bool, pInputNames []string) {
	if pNode.Type == DSD_NODE_CONST1 {
		fmt.Fprintf(pFile, "Const%d", b2i(!fComp))
		return
	}
	if pNode.Type != DSD_NODE_BUF && pNode.Type != DSD_NODE_PRIME && pNode.Type != DSD_NODE_OR && pNode.Type != DSD_NODE_EXOR {
		panic("dsd: Dsd_TreePrint2_rec: bad node type")
	}
	if pNode.Type == DSD_NODE_BUF {
		if fComp {
			fmt.Fprintf(pFile, "!")
		}
		fmt.Fprintf(pFile, "%s", pInputNames[pNode.S.Index])
	} else if pNode.Type == DSD_NODE_PRIME {
		fmt.Fprintf(pFile, " ")
		nDecs := len(pNode.pDecs)
		if nDecs <= 6 {
			pCanonPerm := make([]byte, 6)
			// compute truth table
			bFunc := Dsd_TreeGetPrimeFunction(dd, pNode)
			uTruth := []uint64{Dsd_TreeFunc2Truth_rec(dd, bFunc)}
			bdd.Cudd_Ref(bFunc)
			bdd.Cudd_RecursiveDeref(dd, bFunc)
			// canonicize truth table
			uCanonPhase := canon.Canonicalize(uTruth, int32(nDecs), pCanonPerm)
			if fComp != ((uCanonPhase>>uint(nDecs))&1 != 0) {
				fmt.Fprintf(pFile, "!")
			}
			util.Abc_TtPrintHexRev(pFile, uTruth, int32(nDecs))
			fmt.Fprintf(pFile, "{")
			for i := 0; i < nDecs; i++ {
				pInput := pNode.pDecs[int(pCanonPerm[i])]
				dsdTreePrint2_rec(pFile, dd, Dsd_Regular(pInput), Dsd_IsComplement(pInput) != ((uCanonPhase>>uint(i))&1 != 0), pInputNames)
			}
			fmt.Fprintf(pFile, "} ")
		} else {
			fmt.Fprintf(pFile, "|%d|", nDecs)
			fmt.Fprintf(pFile, "{")
			for i := 0; i < nDecs; i++ {
				dsdTreePrint2_rec(pFile, dd, Dsd_Regular(pNode.pDecs[i]), Dsd_IsComplement(pNode.pDecs[i]), pInputNames)
			}
			fmt.Fprintf(pFile, "} ")
		}
	} else if pNode.Type == DSD_NODE_OR {
		if !fComp {
			fmt.Fprintf(pFile, "!")
		}
		fmt.Fprintf(pFile, "(")
		for i := 0; i < len(pNode.pDecs); i++ {
			dsdTreePrint2_rec(pFile, dd, Dsd_Regular(pNode.pDecs[i]), !Dsd_IsComplement(pNode.pDecs[i]), pInputNames)
		}
		fmt.Fprintf(pFile, ")")
	} else if pNode.Type == DSD_NODE_EXOR {
		if fComp {
			fmt.Fprintf(pFile, "!")
		}
		fmt.Fprintf(pFile, "[")
		for i := 0; i < len(pNode.pDecs); i++ {
			dsdTreePrint2_rec(pFile, dd, Dsd_Regular(pNode.pDecs[i]), Dsd_IsComplement(pNode.pDecs[i]), pInputNames)
		}
		fmt.Fprintf(pFile, "]")
	}
}

func Dsd_TreePrint2(pFile io.Writer, pDsdMan *Dsd_Manager_t, pInputNames, pOutputNames []string, Output int) {
	if Output == -1 {
		for i := 0; i < len(pDsdMan.roots); i++ {
			fmt.Fprintf(pFile, "%8s = ", pOutputNames[i])
			dsdTreePrint2_rec(pFile, pDsdMan.dd, Dsd_Regular(pDsdMan.roots[i]), Dsd_IsComplement(pDsdMan.roots[i]), pInputNames)
			fmt.Fprintf(pFile, "\n")
		}
	} else {
		if Output < 0 || Output >= len(pDsdMan.roots) {
			panic("dsd: Dsd_TreePrint2: Output >= 0 && Output < pDsdMan->nRoots")
		}
		fmt.Fprintf(pFile, "%8s = ", pOutputNames[Output])
		dsdTreePrint2_rec(pFile, pDsdMan.dd, Dsd_Regular(pDsdMan.roots[Output]), Dsd_IsComplement(pDsdMan.roots[Output]), pInputNames)
		fmt.Fprintf(pFile, "\n")
	}
}

// Dsd_NodePrint prints the decompostion tree into file.
func Dsd_NodePrint(pFile io.Writer, pNode *Dsd_Node_t) {
	SigCounter := 1
	pNodeR := Dsd_Regular(pNode)
	dsdNodePrint_rec(pFile, pNodeR, pNodeR != pNode, "F", 0, &SigCounter)
}

// dsdNodePrint_rec prints one node of the decomposition tree.
func dsdNodePrint_rec(pFile io.Writer, pNode *Dsd_Node_t, fComp bool, pOutputName string, nOffset int, pSigCounter *int) {
	if pNode.Type != DSD_NODE_BUF && pNode.Type != DSD_NODE_CONST1 &&
		pNode.Type != DSD_NODE_PRIME && pNode.Type != DSD_NODE_OR && pNode.Type != DSD_NODE_EXOR {
		panic("dsd: Dsd_NodePrint_rec: bad node type")
	}

	printSpaces(pFile, nOffset)

	if !fComp {
		fmt.Fprintf(pFile, "%s = ", pOutputName)
	} else {
		fmt.Fprintf(pFile, "NOT(%s) = ", pOutputName)
	}
	pInputNums := make([]int, len(pNode.pDecs))
	if pNode.Type == DSD_NODE_CONST1 {
		fmt.Fprintf(pFile, " Constant 1.\n")
	} else if pNode.Type == DSD_NODE_BUF {
		fmt.Fprintf(pFile, " ")
		printChar(pFile, 'a'+int(pNode.S.Index))
		fmt.Fprintf(pFile, "\n")
	} else {
		var name string
		switch pNode.Type {
		case DSD_NODE_PRIME:
			name = "PRIME("
		case DSD_NODE_OR:
			name = "OR("
		default:
			name = "EXOR("
		}
		// print the line
		fmt.Fprintf(pFile, "%s", name)
		for i := 0; i < len(pNode.pDecs); i++ {
			pInput := Dsd_Regular(pNode.pDecs[i])
			fCompNew := pInput != pNode.pDecs[i]
			if fCompNew && pNode.Type != DSD_NODE_OR {
				panic("dsd: Dsd_NodePrint_rec: fCompNew == 0")
			}
			if i != 0 {
				fmt.Fprintf(pFile, ",")
			}
			if pInput.Type == DSD_NODE_BUF {
				pInputNums[i] = 0
				fmt.Fprintf(pFile, " ")
				printChar(pFile, 'a'+int(pInput.S.Index))
			} else {
				pInputNums[i] = *pSigCounter
				*pSigCounter++
				fmt.Fprintf(pFile, " <%d>", pInputNums[i])
			}
			if fCompNew {
				fmt.Fprintf(pFile, "'")
			}
		}
		fmt.Fprintf(pFile, " )\n")
		// call recursively for the following blocks
		for i := 0; i < len(pNode.pDecs); i++ {
			if pInputNums[i] != 0 {
				Buffer := fmt.Sprintf("<%d>", pInputNums[i])
				dsdNodePrint_rec(pFile, Dsd_Regular(pNode.pDecs[i]), false, Buffer, nOffset+6, pSigCounter)
			}
		}
	}
}

// Dsd_TreeGetPrimeFunctionOld returns the function of one node of the
// decomposition tree. This is the old procedure. It is now superceded by
// the procedure Dsd_TreeGetPrimeFunction() found in "dsdLocal.c".
func Dsd_TreeGetPrimeFunctionOld(dd *DdManager, pNode *Dsd_Node_t, fRemap int) *DdNode {
	var bCof0, bCof1, bCube0, bCube1, bNewFunc, bTemp *DdNode
	if pNode == nil || Dsd_IsComplement(pNode) || pNode.Type != DSD_NODE_PRIME {
		panic("dsd: Dsd_TreeGetPrimeFunctionOld: expected a regular PRIME node")
	}

	// transform the function of this block to depend on inputs
	// corresponding to the formal inputs

	// first, substitute those inputs that have some blocks associated with them
	// second, remap the inputs to the top of the manager (then, it is easy to output them)

	// start the function
	bNewFunc = pNode.G
	bdd.Cudd_Ref(bNewFunc)
	// go over all primary inputs
	for i := 0; i < len(pNode.pDecs); i++ {
		if pNode.pDecs[i].Type != DSD_NODE_BUF { // remap only if it is not the buffer
			bCube0 = extraBddFindOneCube(dd, bdd.Cudd_Not(pNode.pDecs[i].G))
			bdd.Cudd_Ref(bCube0)
			bCof0 = bdd.Cudd_Cofactor(dd, bNewFunc, bCube0)
			bdd.Cudd_Ref(bCof0)
			bdd.Cudd_RecursiveDeref(dd, bCube0)

			bCube1 = extraBddFindOneCube(dd, pNode.pDecs[i].G)
			bdd.Cudd_Ref(bCube1)
			bCof1 = bdd.Cudd_Cofactor(dd, bNewFunc, bCube1)
			bdd.Cudd_Ref(bCof1)
			bdd.Cudd_RecursiveDeref(dd, bCube1)

			bdd.Cudd_RecursiveDeref(dd, bNewFunc)

			// use the first variale in the support of the component
			bNewFunc = bdd.Cudd_bddIte(dd, dd.Vars[pNode.pDecs[i].S.Index], bCof1, bCof0)
			bdd.Cudd_Ref(bNewFunc)
			bdd.Cudd_RecursiveDeref(dd, bCof0)
			bdd.Cudd_RecursiveDeref(dd, bCof1)
		}
	}

	if fRemap != 0 {
		// remap the function to the first variables of the manager
		Permute := make([]int, len(dd.Vars))
		for i := range Permute {
			Permute[i] = i
		}
		for i := 0; i < len(pNode.pDecs); i++ {
			Permute[pNode.pDecs[i].S.Index] = i
		}
		bTemp = bNewFunc
		bNewFunc = bdd.Cudd_bddPermute(dd, bTemp, Permute)
		bdd.Cudd_Ref(bNewFunc)
		bdd.Cudd_RecursiveDeref(dd, bTemp)
	}

	bdd.Cudd_Deref(bNewFunc)
	return bNewFunc
}
