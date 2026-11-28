/**CFile****************************************************************

  FileName    [ltl_parser.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Liveness property checking.]

  Synopsis    [LTL checker.]

  Author      [Sayak Ray]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - January 1, 2009.]

  Revision    [$Id: ltl_parser.c,v 1.00 2009/01/01 00:00:00 alanmi Exp $]

***********************************************************************/

package live

import (
	"fmt"

	"github.com/glycerine/abcgo/aig/aig"
	abc "github.com/glycerine/abcgo/base/abc"
)

type tokenType int32

const (
	AND tokenType = iota
	OR
	NOT
	IMPLY
	GLOBALLY
	EVENTUALLY
	NEXT
	UNTIL
	BOOL
)

type ltlGrammerTokenType int32

const (
	OPERAND ltlGrammerTokenType = iota
	LTL
	BINOP
	UOP
)

type ltlNode struct {
	type_ tokenType
	name  string
	pObj  Aig_Obj_t // aigNull for NULL
	left  *ltlNode
	right *ltlNode
}

func generateTypedNode(new_type tokenType) *ltlNode {
	return &ltlNode{type_: new_type, pObj: aigNull}
}

func isNotVarNameSymbol(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == ':' || c == 0
}

// The frame's LTL database (pAbc->vLTLProperties_global; nil is NULL).
const ltlPropertiesKey = "vLTLProperties_global"

func Abc_FrameReadLTLProperties(pAbc *Abc_Frame_t) []string {
	v, _ := pAbc.Resource(ltlPropertiesKey).([]string)
	return v
}

func Abc_FrameCopyLTLDataBase(pAbc *Abc_Frame_t, pNtk *Abc_Ntk_t) {
	v := make([]string, 0, len(pNtk.LTLProperties))
	v = append(v, pNtk.LTLProperties...)
	pAbc.SetResource(ltlPropertiesKey, v)
}

// formula[i] with C's terminating NUL (and the byte before the string).
func charAt(formula string, i int) byte {
	if i < 0 || i >= len(formula) {
		return 0
	}
	return formula[i]
}

func getVarName(suffixFormula string, startLoc int, endLocation *int) (string, bool) {
	i := startLoc
	if isNotVarNameSymbol(charAt(suffixFormula, startLoc)) {
		return "", false
	}
	for !isNotVarNameSymbol(charAt(suffixFormula, i)) {
		i++
	}
	*endLocation = i
	return suffixFormula[startLoc:i], true
}

var startOfSuffixString = 0

func isUnexpectedEOS(formula string, index int) bool {
	if index >= len(formula) {
		fmt.Printf("\nInvalid LTL formula: unexpected end of string...")
		return true
	}
	return false
}

func isTemporalOperator(formula string, index int) bool {
	if !(isUnexpectedEOS(formula, index) || formula[index] == 'G' || formula[index] == 'F' || formula[index] == 'U' || formula[index] == 'X') {
		fmt.Printf("\nInvalid LTL formula: expecting temporal operator at the position %d....\n", index)
		return false
	}
	return true
}

// readVar ends readLtlFormula with a variable name starting at i.
func readVar(formula string, i int) *ltlNode {
	var rememberEnd int
	varName, ok := getVarName(formula, i, &rememberEnd)
	if !ok {
		fmt.Printf("\nInvalid LTL formula: expecting valid variable name token...aborting")
		return nil
	}
	curr_node := generateTypedNode(BOOL)
	curr_node.name = varName
	startOfSuffixString = rememberEnd
	return curr_node
}

// readUnary reads the operand of a unary operator at i.
func readUnary(formula string, i int, kind tokenType) *ltlNode {
	startOfSuffixString = i + 1
	temp_node_left := readLtlFormula(formula)
	if temp_node_left == nil {
		return nil
	}
	curr_node := generateTypedNode(kind)
	curr_node.left = temp_node_left
	return curr_node
}

// readBinary reads the two operands of a binary operator at i.
func readBinary(formula string, i int, kind tokenType) *ltlNode {
	startOfSuffixString = i + 1
	temp_node_left := readLtlFormula(formula)
	if temp_node_left == nil {
		return nil
	}
	temp_node_right := readLtlFormula(formula)
	if temp_node_right == nil {
		return nil
	}
	curr_node := generateTypedNode(kind)
	curr_node.left = temp_node_left
	curr_node.right = temp_node_right
	return curr_node
}

func readLtlFormula(formula string) *ltlNode {
	i := startOfSuffixString

	formulaLength := len(formula)
	if isUnexpectedEOS(formula, startOfSuffixString) {
		fmt.Printf("\nFAULTING POINT: formula = %s\nstartOfSuffixString = %d, formula[%d] = %c\n\n", formula, startOfSuffixString, startOfSuffixString-1, charAt(formula, startOfSuffixString-1))
		return nil
	}

	for i < formulaLength {
		ch := formula[i]

		switch ch {
		case ' ', '\n', '\r', '\t', '\v', '\f':
			i++
			startOfSuffixString = i
		case ':':
			i++
			if !isTemporalOperator(formula, i) {
				return nil
			}
			startOfSuffixString = i
		case 'G', 'F', 'X':
			if charAt(formula, i-1) == ':' { // the letter is a temporal operator
				kind := GLOBALLY
				if ch == 'F' {
					kind = EVENTUALLY
				} else if ch == 'X' {
					kind = NEXT
				}
				return readUnary(formula, i, kind)
			}
			// the letter must be starting a variable name
			return readVar(formula, i)
		case 'U':
			if charAt(formula, i-1) == ':' { // 'U' is a temporal operator
				return readBinary(formula, i, UNTIL)
			}
			return readVar(formula, i)
		case '+':
			return readBinary(formula, i, OR)
		case '&':
			return readBinary(formula, i, AND)
		case '!':
			return readUnary(formula, i, NOT)
		default:
			return readVar(formula, i)
		}
	}
	return nil
}

func resetGlobalVar() {
	startOfSuffixString = 0
}

func parseFormulaCreateAST(inputFormula string) *ltlNode {
	temp := readLtlFormula(inputFormula)
	resetGlobalVar()
	return temp
}

func traverseAbstractSyntaxTree(node *ltlNode) {
	switch node.type_ {
	case AND:
		fmt.Printf("& ")
		traverseAbstractSyntaxTree(node.left)
		traverseAbstractSyntaxTree(node.right)
	case OR:
		fmt.Printf("+ ")
		traverseAbstractSyntaxTree(node.left)
		traverseAbstractSyntaxTree(node.right)
	case NOT:
		fmt.Printf("~ ")
		traverseAbstractSyntaxTree(node.left)
	case GLOBALLY:
		fmt.Printf("G ")
		traverseAbstractSyntaxTree(node.left)
	case EVENTUALLY:
		fmt.Printf("F ")
		traverseAbstractSyntaxTree(node.left)
	case NEXT:
		fmt.Printf("X ")
		traverseAbstractSyntaxTree(node.left)
	case UNTIL:
		fmt.Printf("U ")
		traverseAbstractSyntaxTree(node.left)
		traverseAbstractSyntaxTree(node.right)
	case BOOL:
		fmt.Printf("%s ", node.name)
	default:
		fmt.Printf("\nUnsupported token type: Exiting execution\n")
		exit0()
	}
}

func traverseAbstractSyntaxTree_postFix(node *ltlNode) {
	switch node.type_ {
	case AND:
		fmt.Printf("( ")
		traverseAbstractSyntaxTree_postFix(node.left)
		fmt.Printf("& ")
		traverseAbstractSyntaxTree_postFix(node.right)
		fmt.Printf(") ")
	case OR:
		fmt.Printf("( ")
		traverseAbstractSyntaxTree_postFix(node.left)
		fmt.Printf("+ ")
		traverseAbstractSyntaxTree_postFix(node.right)
		fmt.Printf(") ")
	case NOT:
		fmt.Printf("~ ")
		traverseAbstractSyntaxTree_postFix(node.left)
	case GLOBALLY:
		fmt.Printf("G ")
		traverseAbstractSyntaxTree_postFix(node.left)
	case EVENTUALLY:
		fmt.Printf("F ")
		traverseAbstractSyntaxTree_postFix(node.left)
	case NEXT:
		fmt.Printf("X ")
		traverseAbstractSyntaxTree_postFix(node.left)
	case UNTIL:
		fmt.Printf("( ")
		traverseAbstractSyntaxTree_postFix(node.left)
		fmt.Printf("U ")
		traverseAbstractSyntaxTree_postFix(node.right)
		fmt.Printf(") ")
	case BOOL:
		fmt.Printf("%s ", node.name)
	default:
		fmt.Printf("\nUnsupported token type: Exiting execution\n")
		exit0()
	}
}

// signalIndex is Vec_PtrFind on vSignal, pushing the signal when absent.
func signalIndex(vSignal *[]Aig_Obj_t, pObj Aig_Obj_t) int {
	serialNumSignal := aigObjFind(*vSignal, pObj)
	if serialNumSignal == -1 {
		*vSignal = append(*vSignal, pObj)
		serialNumSignal = aigObjFind(*vSignal, pObj)
	}
	return serialNumSignal
}

// Vec_VecPush
func vecVecPush(vAigGFMap *[][]*ltlNode, level int, node *ltlNode) {
	for len(*vAigGFMap) <= level {
		*vAigGFMap = append(*vAigGFMap, nil)
	}
	(*vAigGFMap)[level] = append((*vAigGFMap)[level], node)
}

func populateAigPointerUnitGF(pAigNew *Aig_Man_t, topASTNode *ltlNode, vSignal *[]Aig_Obj_t, vAigGFMap *[][]*ltlNode) {
	switch topASTNode.type_ {
	case AND, OR, IMPLY:
		populateAigPointerUnitGF(pAigNew, topASTNode.left, vSignal, vAigGFMap)
		populateAigPointerUnitGF(pAigNew, topASTNode.right, vSignal, vAigGFMap)
	case NOT:
		populateAigPointerUnitGF(pAigNew, topASTNode.left, vSignal, vAigGFMap)
	case GLOBALLY:
		nextNode := topASTNode.left
		nextToNextNode := nextNode.left
		if nextToNextNode.type_ != BOOL {
			buildLogicFromLTLNode_combinationalOnly(pAigNew, nextToNextNode)
		}
		serialNumSignal := signalIndex(vSignal, nextToNextNode.pObj)
		vecVecPush(vAigGFMap, serialNumSignal, topASTNode)
	case BOOL:
	default:
		fmt.Printf("\nINVALID situation: aborting...\n")
		exit0()
	}
}

func buildLogicFromLTLNode_combinationalOnly(pAigNew *Aig_Man_t, pLtlNode *ltlNode) Aig_Obj_t {
	if pLtlNode.pObj != aigNull {
		return pLtlNode.pObj
	}
	switch pLtlNode.type_ {
	case AND:
		leftAigObj := buildLogicFromLTLNode_combinationalOnly(pAigNew, pLtlNode.left)
		rightAigObj := buildLogicFromLTLNode_combinationalOnly(pAigNew, pLtlNode.right)
		pLtlNode.pObj = aig.Aig_And(pAigNew, leftAigObj, rightAigObj)
		return pLtlNode.pObj
	case OR:
		leftAigObj := buildLogicFromLTLNode_combinationalOnly(pAigNew, pLtlNode.left)
		rightAigObj := buildLogicFromLTLNode_combinationalOnly(pAigNew, pLtlNode.right)
		pLtlNode.pObj = aig.Aig_Or(pAigNew, leftAigObj, rightAigObj)
		return pLtlNode.pObj
	case NOT:
		leftAigObj := buildLogicFromLTLNode_combinationalOnly(pAigNew, pLtlNode.left)
		pLtlNode.pObj = aig.Aig_Not(leftAigObj)
		return pLtlNode.pObj
	case GLOBALLY, EVENTUALLY, NEXT, UNTIL:
		fmt.Printf("FORBIDDEN node: ABORTING!!\n")
		exit0()
	default:
		fmt.Printf("\nSerious ERROR: attempting to create AIG node from a temporal node\n")
		exit0()
	}
	return aigNull
}

func buildLogicFromLTLNode(pAig *Aig_Man_t, pLtlNode *ltlNode) Aig_Obj_t {
	if pLtlNode.pObj != aigNull {
		return pLtlNode.pObj
	}
	switch pLtlNode.type_ {
	case AND:
		leftAigObj := buildLogicFromLTLNode(pAig, pLtlNode.left)
		rightAigObj := buildLogicFromLTLNode(pAig, pLtlNode.right)
		pLtlNode.pObj = aig.Aig_And(pAig, leftAigObj, rightAigObj)
		return pLtlNode.pObj
	case OR:
		leftAigObj := buildLogicFromLTLNode(pAig, pLtlNode.left)
		rightAigObj := buildLogicFromLTLNode(pAig, pLtlNode.right)
		pLtlNode.pObj = aig.Aig_Or(pAig, leftAigObj, rightAigObj)
		return pLtlNode.pObj
	case NOT:
		leftAigObj := buildLogicFromLTLNode(pAig, pLtlNode.left)
		pLtlNode.pObj = aig.Aig_Not(leftAigObj)
		return pLtlNode.pObj
	case GLOBALLY, EVENTUALLY, NEXT, UNTIL:
		fmt.Printf("\nAttempting to create circuit with missing AIG pointer in a TEMPORAL node: ABORTING!!\n")
		exit0()
	default:
		fmt.Printf("\nSerious ERROR: attempting to create AIG node from a temporal node\n")
		exit0()
	}
	return aigNull
}

func isNonTemporalSubformula(topNode *ltlNode) int32 {
	switch topNode.type_ {
	case AND, OR, IMPLY:
		if isNonTemporalSubformula(topNode.left) != 0 && isNonTemporalSubformula(topNode.right) != 0 {
			return 1
		}
		return 0
	case NOT:
		return isNonTemporalSubformula(topNode.left)
	case BOOL:
		return 1
	default:
		return 0
	}
}

func isWellFormed(topNode *ltlNode) int32 {
	switch topNode.type_ {
	case AND, OR, IMPLY:
		if isWellFormed(topNode.left) != 0 && isWellFormed(topNode.right) != 0 {
			return 1
		}
		return 0
	case NOT:
		return isWellFormed(topNode.left)
	case BOOL:
		return 1
	case GLOBALLY:
		nextNode := topNode.left
		if nextNode.type_ != EVENTUALLY {
			return 0
		}
		return isNonTemporalSubformula(nextNode.left)
	default:
		return 0
	}
}

func checkBooleanConstant(targetName string) int32 {
	if targetName == "true" {
		return 1
	}
	if targetName == "false" {
		return 0
	}
	return -1
}

func checkSignalNameExistence(pNtk *Abc_Ntk_t, topASTNode *ltlNode) int32 {
	switch topASTNode.type_ {
	case BOOL:
		targetName := topASTNode.name
		if checkBooleanConstant(targetName) != -1 {
			return 1
		}
		for i := int32(0); i < abc.Abc_NtkPoNum(pNtk); i++ {
			if abc.Abc_ObjName(abc.Abc_NtkPo(pNtk, i)) == targetName {
				return 1
			}
		}
		fmt.Printf("\nVariable name \"%s\" not found in the PO name list\n", targetName)
		return 0
	case AND, OR, IMPLY, UNTIL:
		if checkSignalNameExistence(pNtk, topASTNode.left) != 0 && checkSignalNameExistence(pNtk, topASTNode.right) != 0 {
			return 1
		}
		return 0
	case NOT, NEXT, GLOBALLY, EVENTUALLY:
		return checkSignalNameExistence(pNtk, topASTNode.left)
	default:
		fmt.Printf("\nUNSUPPORTED LTL NODE TYPE:: Aborting execution\n")
		exit0()
	}
	return 0
}

func populateBoolWithAigNodePtr(pNtk *Abc_Ntk_t, pAigOld, pAigNew *Aig_Man_t, topASTNode *ltlNode) {
	switch topASTNode.type_ {
	case BOOL:
		targetName := topASTNode.name
		if checkBooleanConstant(targetName) == 1 {
			topASTNode.pObj = aig.Aig_ManConst1(pAigNew)
			return
		}
		if checkBooleanConstant(targetName) == 0 {
			topASTNode.pObj = aig.Aig_Not(aig.Aig_ManConst1(pAigNew))
			return
		}
		for i := int32(0); i < abc.Abc_NtkPoNum(pNtk); i++ {
			if abc.Abc_ObjName(abc.Abc_NtkPo(pNtk, i)) == targetName {
				pObj := aig.Aig_ManCo(pAigOld, i)
				topASTNode.pObj = driverNew(pAigOld, pObj)
				return
			}
		}
	case AND, OR, IMPLY, UNTIL:
		populateBoolWithAigNodePtr(pNtk, pAigOld, pAigNew, topASTNode.left)
		populateBoolWithAigNodePtr(pNtk, pAigOld, pAigNew, topASTNode.right)
	case NOT, NEXT, GLOBALLY, EVENTUALLY:
		populateBoolWithAigNodePtr(pNtk, pAigOld, pAigNew, topASTNode.left)
	default:
		fmt.Printf("\nUNSUPPORTED LTL NODE TYPE:: Aborting execution\n")
		exit0()
	}
}

func checkAllBoolHaveAIGPointer(topASTNode *ltlNode) int32 {
	switch topASTNode.type_ {
	case BOOL:
		if topASTNode.pObj != aigNull {
			return 1
		}
		fmt.Printf("\nfaulting PODMANDYO topASTNode->name = %s\n", topASTNode.name)
		return 0
	case AND, OR, IMPLY, UNTIL:
		if checkAllBoolHaveAIGPointer(topASTNode.left) != 0 && checkAllBoolHaveAIGPointer(topASTNode.right) != 0 {
			return 1
		}
		return 0
	case NOT, NEXT, GLOBALLY, EVENTUALLY:
		return checkAllBoolHaveAIGPointer(topASTNode.left)
	default:
		fmt.Printf("\nUNSUPPORTED LTL NODE TYPE:: Aborting execution\n")
		exit0()
	}
	return 0
}

func setAIGNodePtrOfGloballyNode(astNode *ltlNode, pObjLo Aig_Obj_t) {
	astNode.pObj = pObjLo
}

func retriveAIGPointerFromLTLNode(astNode *ltlNode) Aig_Obj_t {
	return astNode.pObj
}
