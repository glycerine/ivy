package amap

import (
	"fmt"
	"strings"
)

import "os"

/**CFile****************************************************************

  FileName    [amapParse.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Technology mapper for standard cells.]

  Synopsis    [Parses representations of gates.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: amapParse.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "amapInt.h"
//#include "aig/hop/hop.h"
//#include "bool/kit/kit.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// the list of operation symbols to be used in expressions
#define AMAP_EQN_SYM_OPEN    '('   // opening parenthesis
#define AMAP_EQN_SYM_CLOSE   ')'   // closing parenthesis
#define AMAP_EQN_SYM_CONST0  '0'   // constant 0
#define AMAP_EQN_SYM_CONST1  '1'   // constant 1
#define AMAP_EQN_SYM_NEG     '!'   // negation before the variable
#define AMAP_EQN_SYM_NEGAFT  '\''  // negation after the variable
#define AMAP_EQN_SYM_AND     '*'   // logic AND
#define AMAP_EQN_SYM_AND2    '&'   // logic AND
#define AMAP_EQN_SYM_XOR     '^'   // logic XOR
#define AMAP_EQN_SYM_OR      '+'   // logic OR
#define AMAP_EQN_SYM_OR2     '|'   // logic OR

// the list of opcodes (also specifying operation precedence)
const AMAP_EQN_OPER_NEG =    10    // negation
const AMAP_EQN_OPER_AND =     9    // logic AND
const AMAP_EQN_OPER_XOR =     8    // logic XOR
const AMAP_EQN_OPER_OR =      7    // logic OR
const AMAP_EQN_OPER_MARK =    1    // OpStack token standing for an opening parenthesis

// these are values of the internal Flag
const AMAP_EQN_FLAG_START =   1    // after the opening parenthesis 
const AMAP_EQN_FLAG_VAR =     2    // after operation is received
const AMAP_EQN_FLAG_OPER =    3    // after operation symbol is received
const AMAP_EQN_FLAG_ERROR =   4    // when error is detected

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Performs the operation on the top entries in the stack.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Amap_ParseFormulaOper(pMan *Hop_Man_t, pStackFn *Vec_Ptr_t, Oper int32) *Hop_Obj_t {
    var gArg1, gArg2, gFunc *Hop_Obj_t;
    // perform the given operation
    gArg2 = (Hop_Obj_t *)Vec_PtrPop( pStackFn );
    gArg1 = (Hop_Obj_t *)Vec_PtrPop( pStackFn );
    if Oper == AMAP_EQN_OPER_AND {
        gFunc = Hop_And( pMan, gArg1, gArg2 );
    } else if Oper == AMAP_EQN_OPER_OR {
        gFunc = Hop_Or( pMan, gArg1, gArg2 );
    } else if Oper == AMAP_EQN_OPER_XOR {
        gFunc = Hop_Exor( pMan, gArg1, gArg2 );
    } else {
        return nil; }
//    Cudd_Ref( gFunc );
//    Cudd_RecursiveDeref( dd, gArg1 );
//    Cudd_RecursiveDeref( dd, gArg2 );
    Vec_PtrPush( pStackFn,  gFunc );
    return gFunc;
}

/**Function*************************************************************

  Synopsis    [Derives the AIG corresponding to the equation.]

  Description [Takes the stream to output messages, the formula, the vector
  of variable names and the AIG manager.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Amap_ParseFormula(pOutput *os.File, pFormInit string, vVarNames *Vec_Ptr_t, pMan *Hop_Man_t, pGateName string) *Hop_Obj_t {
    var pFormula string;
    var pStackFn *Vec_Ptr_t;
    var pStackOp *Vec_Int_t;
    var gFunc *Hop_Obj_t;
    var pTemp, pName string;
    var nParans, fFound, Flag int32;
    var Oper, Oper1, Oper2 int32;
    var i, v int32;

    // make sure that the number of opening and closing parentheses is the same
    nParans = 0;
    for  pTemp = pFormInit; *pTemp; pTemp++  {
        if *pTemp == '(' {
            nParans++;
        } else if *pTemp == ')' {
            nParans--; }
    }
    if nParans != 0 {
        fmt.Fprintf(pOutput, "Amap_ParseFormula(): Different number of opening and closing parentheses ().\n");
        return nil;
    }

    // copy the formula
    pFormula = ABC_ALLOC( char, len(pFormInit) + 3 );
    pFormula = fmt.Sprintf("(%s)", pFormInit);

    // start the stacks
    pStackFn = Vec_PtrAlloc( 100 );
    pStackOp = Vec_IntAlloc( 100 );

    Flag = AMAP_EQN_FLAG_START;
    for  pTemp = pFormula; *pTemp; pTemp++  {
        switch  *pTemp  {
        // skip all spaces, tabs, and end-of-lines
        case ' ',
         '\t',
         '\r',
         '\n':
            continue;
        case AMAP_EQN_SYM_CONST0:
            Vec_PtrPush( pStackFn, Hop_ManConst0(pMan) );  // Cudd_Ref( b0 );
            if Flag == AMAP_EQN_FLAG_VAR {
                fmt.Fprintf(pOutput, "Amap_ParseFormula(): No operation symbol before constant 0.\n");
                Flag = AMAP_EQN_FLAG_ERROR; 
                break;
            }
            Flag = AMAP_EQN_FLAG_VAR; 
            break;
        case AMAP_EQN_SYM_CONST1:
            Vec_PtrPush( pStackFn, Hop_ManConst1(pMan) );  //  Cudd_Ref( b1 );
            if Flag == AMAP_EQN_FLAG_VAR {
                fmt.Fprintf(pOutput, "Amap_ParseFormula(): No operation symbol before constant 1.\n");
                Flag = AMAP_EQN_FLAG_ERROR; 
                break;
            }
            Flag = AMAP_EQN_FLAG_VAR; 
            break;
        case AMAP_EQN_SYM_NEG:
            if Flag == AMAP_EQN_FLAG_VAR {// if NEGBEF follows a variable, AND is assumed
                Vec_IntPush( pStackOp, AMAP_EQN_OPER_AND );
                Flag = AMAP_EQN_FLAG_OPER;
            }
            Vec_IntPush( pStackOp, AMAP_EQN_OPER_NEG );
            break;
        case AMAP_EQN_SYM_NEGAFT:
            if Flag != AMAP_EQN_FLAG_VAR {// if there is no variable before NEGAFT, it is an error
                fmt.Fprintf(pOutput, "Amap_ParseFormula(): No variable is specified before the negation suffix.\n");
                Flag = AMAP_EQN_FLAG_ERROR; 
                break;
            
            } else { // if ( Flag == PARSE_FLAG_VAR )
                Vec_PtrPush( pStackFn, Hop_Not( (Hop_Obj_t *)Vec_PtrPop(pStackFn) ) ); }
            break;
        case AMAP_EQN_SYM_AND,
         AMAP_EQN_SYM_AND2,
         AMAP_EQN_SYM_OR,
         AMAP_EQN_SYM_OR2,
         AMAP_EQN_SYM_XOR:
            if Flag != AMAP_EQN_FLAG_VAR {
                fmt.Fprintf(pOutput, "Amap_ParseFormula(): There is no variable before AND, EXOR, or OR.\n");
                Flag = AMAP_EQN_FLAG_ERROR; 
                break;
            }
            if *pTemp == AMAP_EQN_SYM_AND || *pTemp == AMAP_EQN_SYM_AND2 {
                Vec_IntPush( pStackOp, AMAP_EQN_OPER_AND );
            } else if *pTemp == AMAP_EQN_SYM_OR || *pTemp == AMAP_EQN_SYM_OR2 {
                Vec_IntPush( pStackOp, AMAP_EQN_OPER_OR );
            } else { //if ( *pTemp == AMAP_EQN_SYM_XOR )
                Vec_IntPush( pStackOp, AMAP_EQN_OPER_XOR ); }
            Flag = AMAP_EQN_FLAG_OPER; 
            break;
        case AMAP_EQN_SYM_OPEN:
            if Flag == AMAP_EQN_FLAG_VAR {
                Vec_IntPush( pStackOp, AMAP_EQN_OPER_AND );
//                fprintf( pOutput, "Amap_ParseFormula(): An opening parenthesis follows a var without operation sign.\n" ); 
//                Flag = AMAP_EQN_FLAG_ERROR; 
//              break; 
            }
            Vec_IntPush( pStackOp, AMAP_EQN_OPER_MARK );
            // after an opening bracket, it feels like starting over again
            Flag = AMAP_EQN_FLAG_START; 
            break;
        case AMAP_EQN_SYM_CLOSE:
            if Vec_IntSize( pStackOp ) != 0 {
                for {
                    if Vec_IntSize( pStackOp ) == 0 {
                        fmt.Fprintf(pOutput, "Amap_ParseFormula(): There is no opening parenthesis\n");
                        Flag = AMAP_EQN_FLAG_ERROR; 
                        break;
                    }
                    Oper = Vec_IntPop( pStackOp );
                    if Oper == AMAP_EQN_OPER_MARK {
                        break; }

                    // perform the given operation
                    if Amap_ParseFormulaOper( pMan, pStackFn, Oper ) == nil {
                        fmt.Fprintf(pOutput, "Amap_ParseFormula(): Unknown operation\n");
                        ABC_FREE( pFormula );
                        Vec_PtrFreeP( &pStackFn );
                        Vec_IntFreeP( &pStackOp );
                        return nil;
                    }
                }
            
            } else {
                fmt.Fprintf(pOutput, "Amap_ParseFormula(): There is no opening parenthesis\n");
                Flag = AMAP_EQN_FLAG_ERROR; 
                break;
            }
            if Flag != AMAP_EQN_FLAG_ERROR {
                Flag = AMAP_EQN_FLAG_VAR; } 
            break;


        default:
            // scan the next name
            for  i = 0; pTemp[i] && 
                         pTemp[i] != ' ' && pTemp[i] != '\t' && pTemp[i] != '\r' && pTemp[i] != '\n' &&
                         pTemp[i] != AMAP_EQN_SYM_AND && pTemp[i] != AMAP_EQN_SYM_AND2 && pTemp[i] != AMAP_EQN_SYM_OR && pTemp[i] != AMAP_EQN_SYM_OR2 && 
                         pTemp[i] != AMAP_EQN_SYM_XOR && pTemp[i] != AMAP_EQN_SYM_NEGAFT && pTemp[i] != AMAP_EQN_SYM_CLOSE; 
                  i++  {
                    if pTemp[i] == AMAP_EQN_SYM_NEG || pTemp[i] == AMAP_EQN_SYM_OPEN {
                        fmt.Fprintf(pOutput, "Amap_ParseFormula(): The negation sign or an opening parenthesis inside the variable name.\n");
                        Flag = AMAP_EQN_FLAG_ERROR; 
                        break;
                    }
              }
            // variable name is found
            fFound = 0;
            Vec_PtrForEachEntry( char *, vVarNames, pName, v )
                if strncmp(pTemp, pName, i) == 0 && len(pName) == uint32(i) {
                    pTemp += i-1;
                    fFound = 1;
                    break;
                }
            if !fFound { 
                fmt.Fprintf(pOutput, "Amap_ParseFormula(): The parser cannot find var \"%s\" in the input var list of gate \"%s\".\n", pTemp, pGateName); 
                Flag = AMAP_EQN_FLAG_ERROR; 
                break; 
            }
/*
            if ( Flag == AMAP_EQN_FLAG_VAR )
            {
                fprintf( pOutput, "Amap_ParseFormula(): The variable name \"%s\" follows another var without operation sign.\n", pTemp ); 
                Flag = AMAP_EQN_FLAG_ERROR; 
                break; 
            }
*/
            if Flag == AMAP_EQN_FLAG_VAR {
                Vec_IntPush( pStackOp, AMAP_EQN_OPER_AND ); }

            Vec_PtrPush( pStackFn, Hop_IthVar( pMan, v ) ); // Cudd_Ref( pbVars[v] );
            Flag = AMAP_EQN_FLAG_VAR; 
            break;
        }

        if Flag == AMAP_EQN_FLAG_ERROR {
            break;      // error exit
        } else if Flag == AMAP_EQN_FLAG_START {
            continue;  //  go on parsing
        } else if Flag == AMAP_EQN_FLAG_VAR {
            for {  // check if there are negations in the OpStack     
                if Vec_IntSize( pStackOp ) == 0 {
                    break; }
                Oper = Vec_IntPop( pStackOp );
                if Oper != AMAP_EQN_OPER_NEG {
                    Vec_IntPush( pStackOp, Oper );
                    break;
                
                } else {
                      Vec_PtrPush( pStackFn, Hop_Not((Hop_Obj_t *)Vec_PtrPop(pStackFn)) );
                }
            }
        } else { // if ( Flag == AMAP_EQN_FLAG_OPER )
            for {  // execute all the operations in the OpStack
               // with precedence higher or equal than the last one
                Oper1 = Vec_IntPop( pStackOp ); // the last operation
                if Vec_IntSize( pStackOp ) == 0 {  // if it is the only operation, push it back
                    Vec_IntPush( pStackOp, Oper1 );
                    break;
                }
                Oper2 = Vec_IntPop( pStackOp ); // the operation before the last one
                if Oper2 >= Oper1 {  // if Oper2 precedence is higher or equal, execute it
                    if Amap_ParseFormulaOper( pMan, pStackFn, Oper2 ) == nil {
                        fmt.Fprintf(pOutput, "Amap_ParseFormula(): Unknown operation\n");
                        ABC_FREE( pFormula );
                        Vec_PtrFreeP( &pStackFn );
                        Vec_IntFreeP( &pStackOp );
                        return nil;
                    }
                    Vec_IntPush( pStackOp,  Oper1 );     // push the last operation back
                
                } else {  // if Oper2 precedence is lower, push them back and done
                    Vec_IntPush( pStackOp, Oper2 );
                    Vec_IntPush( pStackOp, Oper1 );
                    break;
                }
            } }
    }

    if Flag != AMAP_EQN_FLAG_ERROR {
        if Vec_PtrSize(pStackFn) != 0 {    
            gFunc = (Hop_Obj_t *)Vec_PtrPop(pStackFn);
            if Vec_PtrSize(pStackFn) == 0 {
                if Vec_IntSize( pStackOp ) == 0 {
//                    Cudd_Deref( gFunc );
                    ABC_FREE( pFormula );
                    Vec_PtrFreeP( &pStackFn );
                    Vec_IntFreeP( &pStackOp );
                    return gFunc;
                
                } else {
                    fmt.Fprintf(pOutput, "Amap_ParseFormula(): Something is left in the operation stack\n"); }
            } else {
                fmt.Fprintf(pOutput, "Amap_ParseFormula(): Something is left in the function stack\n"); }
        
        } else {
            fmt.Fprintf(pOutput, "Amap_ParseFormula(): The input string is empty\n"); }
    }
    ABC_FREE( pFormula );
    Vec_PtrFreeP( &pStackFn );
    Vec_IntFreeP( &pStackOp );
    return nil;
}

/**Function*************************************************************

  Synopsis    [Parses equations for the gates.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Amap_LibParseEquations(p *Amap_Lib_t, fVerbose int32) int32 {
//    extern int Kit_TruthSupportSize( unsigned * pTruth, int nVars );
    var pMan *Hop_Man_t;
    var pObj *Hop_Obj_t;
    var vNames *Vec_Ptr_t;
    var vTruth *Vec_Int_t;
    var pGate *Amap_Gat_t;
    var pPin *Amap_Pin_t;
    var pTruth *uint32;
    var i, nPinMax int32;
    nPinMax = Amap_LibNumPinsMax(p);
    if nPinMax > AMAP_MAXINS {
        fmt.Printf("Gates with more than %d inputs will be ignored.\n", AMAP_MAXINS); }
    vTruth = Vec_IntAlloc( 1 << 16 );
    vNames = Vec_PtrAlloc( 100 );
    pMan = Hop_ManStart();
    Hop_IthVar( pMan, nPinMax - 1 );
    Vec_PtrForEachEntry( Amap_Gat_t *, p.vGates, pGate, i )
    {
        if pGate.nPins == 0 {
            pGate.pFunc = (unsigned *)Aig_MmFlexEntryFetch( p.pMemGates, 4 );
            if int32(strings.Compare(pGate.pForm, AMAP_STRING_CONST0)) == 0 {
                pGate.pFunc[0] = 0;
            } else if int32(strings.Compare(pGate.pForm, AMAP_STRING_CONST1)) == 0 {
                pGate.pFunc[0] = ~0;
            } else {
                fmt.Printf("Cannot parse formula \"%s\" of gate \"%s\" with no pins.\n", pGate.pForm, pGate.pName);
                break;
            }
            continue;
        }
        if pGate.nPins > AMAP_MAXINS {
            continue; }
        Vec_PtrClear( vNames );
        Amap_GateForEachPin( pGate, pPin )
            Vec_PtrPush( vNames, pPin.pName );
        pObj = Amap_ParseFormula( os.Stdout, pGate.pForm, vNames, pMan, pGate.pName );
        if pObj == nil {
            break; }
        pTruth = Hop_ManConvertAigToTruth( pMan, pObj, pGate.nPins, vTruth, 0 );
        if Kit_TruthSupportSize(pTruth, pGate.nPins) < int32(pGate.nPins) {
            if fVerbose != 0 {
                fmt.Printf("Skipping gate \"%s\" because its output \"%s\" does not depend on all input variables.\n", pGate.pName, pGate.pForm); }
            continue;
        }
        pGate.pFunc = (unsigned *)Aig_MmFlexEntryFetch( p.pMemGates, sizeof(unsigned)*Abc_TruthWordNum(pGate.nPins) );
        memcpy( pGate.pFunc, pTruth, sizeof(unsigned)*Abc_TruthWordNum(pGate.nPins) );
    }
    Vec_PtrFree( vNames );
    Vec_IntFree( vTruth );
    Hop_ManStop( pMan );
    return i == Vec_PtrSize(p.vGates);
}

/**Function*************************************************************

  Synopsis    [Parses equations for the gates.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Amap_LibParseTest(pFileName string) {
    var fVerbose int32;
    var p *Amap_Lib_t;
    var clk abctime  = Abc_Clock();
    p = Amap_LibReadFile( pFileName, fVerbose );
    if p == nil {
        return; }
    Amap_LibParseEquations( p, fVerbose );
    Amap_LibFree( p );
    ABC_PRT( "Total time", Abc_Clock() - clk );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

