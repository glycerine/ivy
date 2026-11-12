package live

import (
	"fmt"
	"os"
	"strings"
)

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

//#include <stdio.h>
//#include <string.h>
//#include <assert.h>
//#include <stdlib.h>
//#include "aig/aig/aig.h"
//#include "base/abc/abc.h"
//#include "base/main/mainInt.h"

//ABC_NAMESPACE_IMPL_START


type ltlToken int32
const  ( AND ltlToken = iota
 OR
 NOT
 IMPLY
 GLOBALLY
 EVENTUALLY
 NEXT
 UNTIL
 BOOL 
)

type ltlGrammerToken int32
const  ( OPERAND ltlGrammerToken = iota
 LTL
 BINOP
 UOP 
)

type tokenType = ltlToken ;
type ltlGrammerTokenType = ltlGrammerToken ;

type ltlNode_t struct {
    type tokenType;
    name string;
    pObj *Aig_Obj_t;
    left  *ltlNode_t;
    right  *ltlNode_t;
}

type ltlNode = ltlNode_t;

ltlNode *generateTypedNode( tokenType new_type )
//void generateTypedNode( ltlNode *new_node, tokenType new_type )
{
    var new_node *ltlNode;

    new_node = (ltlNode *)malloc( sizeof(ltlNode) );
    if new_node != nil {
        new_node.type = new_type;
        new_node.pObj = nil;
        new_node.name = nil;
        new_node.left = nil;
        new_node.right = nil;
    }

    return new_node;
}

// Aig_Obj_t *buildLogicFromLTLNode_combinationalOnly( Aig_Man_t *pAig, ltlNode *pLtlNode );


func isNotVarNameSymbol(c byte) int32 {
    return ( c == ' ' || c == '\t' || c == '\n' || c == ':' || c == '\0' );
}

func Abc_FrameCopyLTLDataBase(pAbc *Abc_Frame_t, pNtk *Abc_Ntk_t) {
    var pLtlFormula, tempFormula string;
    var i int32;

    if pAbc.vLTLProperties_global != nil {
//        printf("Deleting exisitng LTL database from the frame\n");
        Vec_PtrFree( pAbc.vLTLProperties_global );
        pAbc.vLTLProperties_global = nil;
    }
    pAbc.vLTLProperties_global = Vec_PtrAlloc(Vec_PtrSize(pNtk.vLtlProperties));
    Vec_PtrForEachEntry( char *, pNtk.vLtlProperties, pLtlFormula, i )
    {
        tempFormula = (char *)malloc( sizeof(char)*(len(pLtlFormula)+1) );
        tempFormula = fmt.Sprintf("%s", pLtlFormula);
        Vec_PtrPush( pAbc.vLTLProperties_global, tempFormula );
    }
}

func getVarName(suffixFormula string, startLoc int32, endLocation *int32) string {
    var i int32 = startLoc;
    var length int32;
    var name string;

    if isNotVarNameSymbol( suffixFormula[startLoc] ) != 0 {
        return nil; }

    for  !(isNotVarNameSymbol( suffixFormula[i] ) != 0)  {
        i++;
    }
    *endLocation = i;
    length = i - startLoc;
    name = (char *)malloc( sizeof(char) * (length + 1));
    for  i=0; i<length; i++  {
        name[i] = suffixFormula[i+startLoc];
    }
    name[i] = '\0';

    return name;
}


var startOfSuffixString int32;

func isUnexpectedEOS(formula string, index int32) int32 {
    assert(formula, "formula");
    if index >= int32(len(formula)) {
        fmt.Printf("\nInvalid LTL formula: unexpected end of string...");
        return 1;
    }
    return 0;
}

func isTemporalOperator(formula string, index int32) int32 {
    if !((isUnexpectedEOS( formula, index ) != 0) || formula[ index ] == 'G' || formula[ index ] == 'F' || formula[ index ] == 'U' || formula[ index ] == 'X') {
        fmt.Printf("\nInvalid LTL formula: expecting temporal operator at the position %d....\n", index);
        return 0;
    }
    return 1;
}

func readLtlFormula(formula string) *ltlNode {
    var ch byte;
    var varName string;
    var formulaLength, rememberEnd int32;
    var i int32  = startOfSuffixString;
    var curr_node, temp_node_left, temp_node_right *ltlNode;
    var prevChar byte;
    
    formulaLength = len(formula);
    if isUnexpectedEOS( formula, startOfSuffixString ) != 0 {
        fmt.Printf("\nFAULTING POINT: formula = %s\nstartOfSuffixString = %d, formula[%d] = %c\n\n", formula, startOfSuffixString, startOfSuffixString - 1, formula[startOfSuffixString-1]);
        return nil;
    }
            
    for  i < formulaLength  {
        ch = formula[i];

        switch ch{
            case ' ',
             '\n',
             '\r',
             '\t',
             '\v',
             '\f':
                        i++;
                        startOfSuffixString = i;
                        break;
            case ':':
                        i++;
                        if !(isTemporalOperator( formula, i ) != 0) {
                            return nil; }
                        startOfSuffixString = i;
                        break;
            case 'G':
                        prevChar = formula[i-1];
                        if prevChar == ':' { //i.e. 'G' is a temporal operator
                        
                            i++;
                            startOfSuffixString = i;
                            temp_node_left = readLtlFormula( formula );
                            if temp_node_left == nil {
                                return nil;
                            } else {
                                curr_node = generateTypedNode(GLOBALLY);
                                curr_node.left = temp_node_left;
                                return curr_node;
                            }
                        
                        } else {    //i.e. 'G' must be starting a variable name
                        
                            varName = getVarName( formula, i, &rememberEnd );
                            if !varName {
                                fmt.Printf("\nInvalid LTL formula: expecting valid variable name token...aborting");
                                return nil;
                            }
                            curr_node = generateTypedNode(BOOL);
                            curr_node.name = varName;
                            i = rememberEnd;
                            startOfSuffixString = i;
                            return curr_node;
                        }
            fallthrough
case 'F':
                        prevChar = formula[i-1];
                        if prevChar == ':' { //i.e. 'F' is a temporal operator
                        
                            i++;
                            startOfSuffixString = i;
                            temp_node_left = readLtlFormula( formula );
                            if temp_node_left == nil {
                                return nil;
                            } else {
                                curr_node = generateTypedNode(EVENTUALLY);
                                curr_node.left = temp_node_left;
                                return curr_node;
                            }
                        
                        } else { //i.e. 'F' must be starting a variable name
                        
                            varName = getVarName( formula, i, &rememberEnd );
                            if !varName {
                                fmt.Printf("\nInvalid LTL formula: expecting valid variable name token...aborting");
                                return nil;
                            }
                            curr_node = generateTypedNode(BOOL);
                            curr_node.name = varName;
                            i = rememberEnd;
                            startOfSuffixString = i;
                            return curr_node;
                        }    
            fallthrough
case 'X':
                        prevChar = formula[i-1];
                        if prevChar == ':' { //i.e. 'X' is a temporal operator
                        
                            i++;
                            startOfSuffixString = i;
                            temp_node_left = readLtlFormula( formula );
                            if temp_node_left == nil {
                                return nil;
                            } else {
                                curr_node = generateTypedNode(NEXT);
                                curr_node.left = temp_node_left;
                                return curr_node;
                            }
                        
                        } else { //i.e. 'X' must be starting a variable name
                        
                            varName = getVarName( formula, i, &rememberEnd );
                            if !varName {
                                fmt.Printf("\nInvalid LTL formula: expecting valid variable name token...aborting");
                                return nil;
                            }
                            curr_node = generateTypedNode(BOOL);
                            curr_node.name = varName;
                            i = rememberEnd;
                            startOfSuffixString = i;
                            return curr_node;
                        }    
            fallthrough
case 'U':
                        prevChar = formula[i-1];
                        if prevChar == ':' { //i.e. 'X' is a temporal operator
                        
                            i++;
                            startOfSuffixString = i;
                            temp_node_left = readLtlFormula( formula );
                            if temp_node_left == nil {
                                return nil; }
                            temp_node_right = readLtlFormula( formula );
                            if temp_node_right == nil {
                                //need to do memory management: if right subtree is NULL then left
                                //subtree must be freed.
                                return nil;
                            }
                            curr_node = generateTypedNode(UNTIL);
                            curr_node.left = temp_node_left;
                            curr_node.right = temp_node_right;
                            return curr_node;
                        
                        } else { //i.e. 'U' must be starting a variable name
                        
                            varName = getVarName( formula, i, &rememberEnd );
                            if !varName {
                                fmt.Printf("\nInvalid LTL formula: expecting valid variable name token...aborting");
                                return nil;
                            }
                            curr_node = generateTypedNode(BOOL);
                            curr_node.name = varName;
                            i = rememberEnd;
                            startOfSuffixString = i;
                            return curr_node;
                        }    
            fallthrough
case '+':
                        i++;
                        startOfSuffixString = i;
                        temp_node_left = readLtlFormula( formula );
                        if temp_node_left == nil {
                            return nil; }
                        temp_node_right = readLtlFormula( formula );
                        if temp_node_right == nil {
                            //need to do memory management: if right subtree is NULL then left
                            //subtree must be freed.
                            return nil;
                        }
                        curr_node = generateTypedNode(OR);
                        curr_node.left = temp_node_left;
                        curr_node.right = temp_node_right;
                        return curr_node;
            case '&':
                        i++;
                        startOfSuffixString = i;
                        temp_node_left = readLtlFormula( formula );
                        if temp_node_left == nil {
                            return nil; }
                        temp_node_right = readLtlFormula( formula );
                        if temp_node_right == nil {
                            //need to do memory management: if right subtree is NULL then left
                            //subtree must be freed.
                            return nil;
                        }
                        curr_node = generateTypedNode(AND);
                        curr_node.left = temp_node_left;
                        curr_node.right = temp_node_right;
                        return curr_node;
            case '!':
                        i++;
                        startOfSuffixString = i;
                        temp_node_left = readLtlFormula( formula );
                        if temp_node_left == nil {
                            return nil;
                        } else {
                            curr_node = generateTypedNode(NOT);
                            curr_node.left = temp_node_left;
                            return curr_node;
                        }
            fallthrough
default:
                varName = getVarName( formula, i, &rememberEnd );
                if !varName {
                    fmt.Printf("\nInvalid LTL formula: expecting valid variable name token...aborting");
                    return nil;
                }
                curr_node = generateTypedNode(BOOL);
                curr_node.name = varName;
                i = rememberEnd;
                startOfSuffixString = i;
                return curr_node;
        }
    }
    return nil;
}

func resetGlobalVar() {
    startOfSuffixString = 0;
}

func parseFormulaCreateAST(inputFormula string) *ltlNode {
    var temp *ltlNode;

    temp = readLtlFormula( inputFormula );
    //if( temp == NULL )
    //    printf("\nAST creation failed for formula %s", inputFormula );
    resetGlobalVar();
    return temp;
}

func traverseAbstractSyntaxTree(node *ltlNode) {
    switch node.type{
        case( AND ):
            fmt.Printf("& ");
            assert(node.left != nil, "node.left != nil");
            assert(node.right != nil, "node.right != nil");
            traverseAbstractSyntaxTree( node.left );
            traverseAbstractSyntaxTree( node.right );
            return;
        case( OR ):
            fmt.Printf("+ ");
            assert(node.left != nil, "node.left != nil");
            assert(node.right != nil, "node.right != nil");
            traverseAbstractSyntaxTree( node.left );
            traverseAbstractSyntaxTree( node.right );
            return;
        case( NOT ):
            fmt.Printf("~ ");
            assert(node.left != nil, "node.left != nil");
            traverseAbstractSyntaxTree( node.left );
            assert(node.right == nil, "node.right == nil");
            return;
        case( GLOBALLY ):
            fmt.Printf("G ");
            assert(node.left != nil, "node.left != nil");
            traverseAbstractSyntaxTree( node.left );
            assert(node.right == nil, "node.right == nil");
            return;
        case( EVENTUALLY ):
            fmt.Printf("F ");
            assert(node.left != nil, "node.left != nil");
            traverseAbstractSyntaxTree( node.left );
            assert(node.right == nil, "node.right == nil");
            return;
        case( NEXT ):
            fmt.Printf("X ");
            assert(node.left != nil, "node.left != nil");
            traverseAbstractSyntaxTree( node.left );
            assert(node.right == nil, "node.right == nil");
            return;
        case( UNTIL ):
            fmt.Printf("U ");
            assert(node.left != nil, "node.left != nil");
            assert(node.right != nil, "node.right != nil");
            traverseAbstractSyntaxTree( node.left );
            traverseAbstractSyntaxTree( node.right );
            return;
        case( BOOL ):
            fmt.Printf("%s ", node.name);
            assert(node.left == nil, "node.left == nil");
            assert(node.right == nil, "node.right == nil");
            return;
        default:
            fmt.Printf("\nUnsupported token type: Exiting execution\n");
            os.Exit(int(0));
    }
}

func traverseAbstractSyntaxTree_postFix(node *ltlNode) {
    switch node.type{
        case( AND ):
            fmt.Printf("( ");
            assert(node.left != nil, "node.left != nil");
            assert(node.right != nil, "node.right != nil");
            traverseAbstractSyntaxTree_postFix( node.left );
            fmt.Printf("& ");
            traverseAbstractSyntaxTree_postFix( node.right );
            fmt.Printf(") ");
            return;
        case( OR ):
            fmt.Printf("( ");
            assert(node.left != nil, "node.left != nil");
            assert(node.right != nil, "node.right != nil");
            traverseAbstractSyntaxTree_postFix( node.left );
            fmt.Printf("+ ");
            traverseAbstractSyntaxTree_postFix( node.right );
            fmt.Printf(") ");
            return;
        case( NOT ):
            fmt.Printf("~ ");
            assert(node.left != nil, "node.left != nil");
            traverseAbstractSyntaxTree_postFix( node.left );
            assert(node.right == nil, "node.right == nil");
            return;
        case( GLOBALLY ):
            fmt.Printf("G ");
            //printf("( ");
            assert(node.left != nil, "node.left != nil");
            traverseAbstractSyntaxTree_postFix( node.left );
            assert(node.right == nil, "node.right == nil");
            //printf(") ");
            return;
        case( EVENTUALLY ):
            fmt.Printf("F ");
            //printf("( ");
            assert(node.left != nil, "node.left != nil");
            traverseAbstractSyntaxTree_postFix( node.left );
            assert(node.right == nil, "node.right == nil");
            //printf(") ");
            return;
        case( NEXT ):
            fmt.Printf("X ");
            assert(node.left != nil, "node.left != nil");
            traverseAbstractSyntaxTree_postFix( node.left );
            assert(node.right == nil, "node.right == nil");
            return;
        case( UNTIL ):
            fmt.Printf("( ");
            assert(node.left != nil, "node.left != nil");
            assert(node.right != nil, "node.right != nil");
            traverseAbstractSyntaxTree_postFix( node.left );
            fmt.Printf("U ");
            traverseAbstractSyntaxTree_postFix( node.right );
            fmt.Printf(") ");
            return;
        case( BOOL ):
            fmt.Printf("%s ", node.name);
            assert(node.left == nil, "node.left == nil");
            assert(node.right == nil, "node.right == nil");
            return;
        default:
            fmt.Printf("\nUnsupported token type: Exiting execution\n");
            os.Exit(int(0));
    }
}

func populateAigPointerUnitGF(pAigNew *Aig_Man_t, topASTNode *ltlNode, vSignal *Vec_Ptr_t, vAigGFMap *Vec_Vec_t) {
    var nextNode, nextToNextNode *ltlNode;
    var serialNumSignal int32;

    switch  topASTNode.type {
        case AND,
         OR,
         IMPLY:
            populateAigPointerUnitGF( pAigNew, topASTNode.left, vSignal, vAigGFMap );
            populateAigPointerUnitGF( pAigNew, topASTNode.right, vSignal, vAigGFMap );
            return;
        case NOT:
            populateAigPointerUnitGF( pAigNew, topASTNode.left, vSignal, vAigGFMap );
            return;
        case GLOBALLY:
            nextNode = topASTNode.left;
            assert(nextNode.type == EVENTUALLY, "nextNode.type == EVENTUALLY");
            nextToNextNode = nextNode.left;
            if nextToNextNode.type == BOOL {
                assert(nextToNextNode.pObj, "nextToNextNode.pObj");
                serialNumSignal = Vec_PtrFind( vSignal, nextToNextNode.pObj );
                if serialNumSignal == -1 {                    
                    Vec_PtrPush( vSignal, nextToNextNode.pObj );
                    serialNumSignal = Vec_PtrFind( vSignal, nextToNextNode.pObj );
                }
                //Vec_PtrPush( vGLOBALLY, topASTNode );
                Vec_VecPush( vAigGFMap, serialNumSignal, topASTNode );
            
            } else {
                assert(nextToNextNode.pObj == nil, "nextToNextNode.pObj == nil");
                buildLogicFromLTLNode_combinationalOnly( pAigNew, nextToNextNode );
                serialNumSignal = Vec_PtrFind( vSignal, nextToNextNode.pObj );
                if serialNumSignal == -1 {
                    Vec_PtrPush( vSignal, nextToNextNode.pObj );
                    serialNumSignal = Vec_PtrFind( vSignal, nextToNextNode.pObj );
                }
                //Vec_PtrPush( vGLOBALLY, topASTNode );
                Vec_VecPush( vAigGFMap, serialNumSignal, topASTNode );
            }
            return;
        case BOOL:
            return;
        default:
            fmt.Printf("\nINVALID situation: aborting...\n");
            os.Exit(int(0));
    }
}

func buildLogicFromLTLNode_combinationalOnly(pAigNew *Aig_Man_t, pLtlNode *ltlNode) *Aig_Obj_t {
    var leftAigObj, rightAigObj *Aig_Obj_t;

    if pLtlNode.pObj != nil {
        return pLtlNode.pObj;
    } else {
        assert(pLtlNode.type != BOOL, "pLtlNode.type != BOOL");
        switch  pLtlNode.type {
            case AND:
                assert(pLtlNode.left, "pLtlNode.left"); assert(pLtlNode.right, "pLtlNode.right");
                leftAigObj = buildLogicFromLTLNode_combinationalOnly( pAigNew, pLtlNode.left );
                rightAigObj = buildLogicFromLTLNode_combinationalOnly( pAigNew, pLtlNode.right );
                assert(leftAigObj, "leftAigObj"); assert((rightAigObj != nil), "rightAigObj");
                pLtlNode.pObj = Aig_And( pAigNew, leftAigObj, rightAigObj );
                return pLtlNode.pObj;
            case OR:
                assert(pLtlNode.left, "pLtlNode.left"); assert(pLtlNode.right, "pLtlNode.right");
                leftAigObj = buildLogicFromLTLNode_combinationalOnly( pAigNew, pLtlNode.left );
                rightAigObj = buildLogicFromLTLNode_combinationalOnly( pAigNew, pLtlNode.right );
                assert(leftAigObj, "leftAigObj"); assert((rightAigObj != nil), "rightAigObj");
                pLtlNode.pObj = Aig_Or( pAigNew, leftAigObj, rightAigObj );
                return pLtlNode.pObj;
            case NOT:
                assert(pLtlNode.left, "pLtlNode.left"); assert(pLtlNode.right == nil, "pLtlNode.right == nil"); 
                leftAigObj = buildLogicFromLTLNode_combinationalOnly( pAigNew, pLtlNode.left );
                assert(leftAigObj, "leftAigObj");
                pLtlNode.pObj = Aig_Not( leftAigObj );
                return pLtlNode.pObj;
            case GLOBALLY,
             EVENTUALLY,
             NEXT,
             UNTIL:
                fmt.Printf("FORBIDDEN node: ABORTING!!\n");
                os.Exit(int(0));
            fallthrough
default:
                fmt.Printf("\nSerious ERROR: attempting to create AIG node from a temporal node\n");
                os.Exit(int(0));
        }
    }
}

func buildLogicFromLTLNode(pAig *Aig_Man_t, pLtlNode *ltlNode) *Aig_Obj_t {
    var leftAigObj, rightAigObj *Aig_Obj_t;

    if pLtlNode.pObj != nil {
        return pLtlNode.pObj;
    } else {
        assert(pLtlNode.type != BOOL, "pLtlNode.type != BOOL");
        switch  pLtlNode.type {
            case AND:
                assert(pLtlNode.left, "pLtlNode.left"); assert(pLtlNode.right, "pLtlNode.right");
                leftAigObj = buildLogicFromLTLNode( pAig, pLtlNode.left );
                rightAigObj = buildLogicFromLTLNode( pAig, pLtlNode.right );
                assert(leftAigObj, "leftAigObj"); assert((rightAigObj != nil), "rightAigObj");
                pLtlNode.pObj = Aig_And( pAig, leftAigObj, rightAigObj );
                return pLtlNode.pObj;
            case OR:
                assert(pLtlNode.left, "pLtlNode.left"); assert(pLtlNode.right, "pLtlNode.right");
                leftAigObj = buildLogicFromLTLNode( pAig, pLtlNode.left );
                rightAigObj = buildLogicFromLTLNode( pAig, pLtlNode.right );
                assert(leftAigObj, "leftAigObj"); assert((rightAigObj != nil), "rightAigObj");
                pLtlNode.pObj = Aig_Or( pAig, leftAigObj, rightAigObj );
                return pLtlNode.pObj;
            case NOT:
                assert(pLtlNode.left, "pLtlNode.left"); assert(pLtlNode.right == nil, "pLtlNode.right == nil"); 
                leftAigObj = buildLogicFromLTLNode( pAig, pLtlNode.left );
                assert(leftAigObj, "leftAigObj");
                pLtlNode.pObj = Aig_Not( leftAigObj );
                return pLtlNode.pObj;
            case GLOBALLY,
             EVENTUALLY,
             NEXT,
             UNTIL:
                fmt.Printf("\nAttempting to create circuit with missing AIG pointer in a TEMPORAL node: ABORTING!!\n");
                os.Exit(int(0));
            fallthrough
default:
                fmt.Printf("\nSerious ERROR: attempting to create AIG node from a temporal node\n");
                os.Exit(int(0));
        }
    }
}

func isNonTemporalSubformula(topNode *ltlNode) int32 {
    switch  topNode.type {
        case AND,
         OR,
         IMPLY:
            return isNonTemporalSubformula( topNode.left) && isNonTemporalSubformula( topNode.right ) ;
        case NOT:
            assert(topNode.right == nil, "topNode.right == nil");
            return isNonTemporalSubformula( topNode.left );
        case BOOL:
            return 1;
        default:
            return 0;
    }
}

func isWellFormed(topNode *ltlNode) int32 {
    var nextNode *ltlNode;

    switch  topNode.type {
        case AND,
         OR,
         IMPLY:
            return isWellFormed( topNode.left) && isWellFormed( topNode.right ) ;
        case NOT:
            assert(topNode.right == nil, "topNode.right == nil");
            return isWellFormed( topNode.left );
        case BOOL:
            return 1;
        case GLOBALLY:
            nextNode = topNode.left;
            assert(topNode.right == nil, "topNode.right == nil");
            if nextNode.type != EVENTUALLY {
                return 0;
            } else {
                assert(nextNode.right == nil, "nextNode.right == nil");
                return isNonTemporalSubformula( nextNode.left );
            }
        fallthrough
default:
            return 0;
    }
}

func checkBooleanConstant(targetName string) int32 {
    if int32(strings.Compare(targetName, "true")) == 0 { 
        return 1; }
    if int32(strings.Compare(targetName, "false")) == 0 {
        return 0; }
    return -1;
}

func checkSignalNameExistence(pNtk *Abc_Ntk_t, topASTNode *ltlNode) int32 {
    var targetName string;
    var pNode *Abc_Obj_t;
    var i int32;
        
    switch  topASTNode.type {
        case BOOL:
            targetName = topASTNode.name;
            //printf("\nTrying to match name %s\n", targetName);
            if checkBooleanConstant( targetName ) != -1 {
                return 1; }
            Abc_NtkForEachPo( pNtk, pNode, i )
            {
                if int32(strings.Compare(Abc_ObjName( pNode ), targetName)) == 0 {
                    //printf("\nVariable name \"%s\" MATCHED\n", targetName);
                    return 1;
                }
            }
            fmt.Printf("\nVariable name \"%s\" not found in the PO name list\n", targetName);
            return 0;
        case AND,
         OR,
         IMPLY,
         UNTIL:
            assert(topASTNode.left != nil, "topASTNode.left != nil");
            assert(topASTNode.right != nil, "topASTNode.right != nil");
            return checkSignalNameExistence( pNtk, topASTNode.left ) && checkSignalNameExistence( pNtk, topASTNode.right );
            
        case NOT,
         NEXT,
         GLOBALLY,
         EVENTUALLY:
            assert(topASTNode.left != nil, "topASTNode.left != nil");
            assert(topASTNode.right == nil, "topASTNode.right == nil");
            return checkSignalNameExistence( pNtk, topASTNode.left );
        default:
            fmt.Printf("\nUNSUPPORTED LTL NODE TYPE:: Aborting execution\n");
            os.Exit(int(0));
    }
}

func populateBoolWithAigNodePtr(pNtk *Abc_Ntk_t, pAigOld *Aig_Man_t, pAigNew *Aig_Man_t, topASTNode *ltlNode) {
    var targetName string;
    var pNode *Abc_Obj_t;
    var i int32;
    var pObj, pDriverImage *Aig_Obj_t;
    
    switch  topASTNode.type {
        case BOOL:
            targetName = topASTNode.name;
            if checkBooleanConstant( targetName ) == 1 {
                topASTNode.pObj = Aig_ManConst1( pAigNew );
                return;
            }
            if checkBooleanConstant( targetName ) == 0 {
                topASTNode.pObj = Aig_Not(topASTNode.pObj = Aig_ManConst1( pAigNew ));
                return;
            }
            Abc_NtkForEachPo( pNtk, pNode, i )
                if int32(strings.Compare(Abc_ObjName( pNode ), targetName)) == 0 {
                    pObj = Aig_ManCo( pAigOld, i );
                    assert(Aig_ObjIsCo( pObj ), "Aig_ObjIsCo( pObj )");
                    pDriverImage = Aig_NotCond((Aig_Obj_t *)Aig_Regular(Aig_ObjChild0( pObj )).pData, Aig_ObjFaninC0(pObj));
                    topASTNode.pObj = pDriverImage;
                    return;
                }
            assert(false, "0");
        fallthrough
case AND,
         OR,
         IMPLY,
         UNTIL:
            assert(topASTNode.left != nil, "topASTNode.left != nil");
            assert(topASTNode.right != nil, "topASTNode.right != nil");
            populateBoolWithAigNodePtr( pNtk, pAigOld, pAigNew, topASTNode.left );
            populateBoolWithAigNodePtr( pNtk, pAigOld, pAigNew, topASTNode.right );
            return;
        case NOT,
         NEXT,
         GLOBALLY,
         EVENTUALLY:
            assert(topASTNode.left != nil, "topASTNode.left != nil");
            assert(topASTNode.right == nil, "topASTNode.right == nil");
            populateBoolWithAigNodePtr( pNtk, pAigOld, pAigNew, topASTNode.left );
            return;
        default:
            fmt.Printf("\nUNSUPPORTED LTL NODE TYPE:: Aborting execution\n");
            os.Exit(int(0));
    }
}

func checkAllBoolHaveAIGPointer(topASTNode *ltlNode) int32 {
    
    switch  topASTNode.type {
        case BOOL:
            if( topASTNode.pObj != nil )
                return 1;
            else {
                fmt.Printf("\nfaulting PODMANDYO topASTNode->name = %s\n", topASTNode.name);
                return 0;
            }
        fallthrough
case AND,
         OR,
         IMPLY,
         UNTIL:
            assert(topASTNode.left != nil, "topASTNode.left != nil");
            assert(topASTNode.right != nil, "topASTNode.right != nil");
            return checkAllBoolHaveAIGPointer( topASTNode.left ) && checkAllBoolHaveAIGPointer( topASTNode.right );
            
        case NOT,
         NEXT,
         GLOBALLY,
         EVENTUALLY:
            assert(topASTNode.left != nil, "topASTNode.left != nil");
            assert(topASTNode.right == nil, "topASTNode.right == nil");
            return checkAllBoolHaveAIGPointer( topASTNode.left );
        default:
            fmt.Printf("\nUNSUPPORTED LTL NODE TYPE:: Aborting execution\n");
            os.Exit(int(0));
    }
}

func setAIGNodePtrOfGloballyNode(astNode *ltlNode, pObjLo *Aig_Obj_t) {
    astNode.pObj = pObjLo;
}
            
func retriveAIGPointerFromLTLNode(astNode *ltlNode) *Aig_Obj_t {
    return astNode.pObj;
}


//ABC_NAMESPACE_IMPL_END
