package ver

import (
	"fmt"
	"strings"
)

/**CFile****************************************************************

  FileName    [verFormula.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Verilog parser.]

  Synopsis    [Formula parser to read Verilog assign statements.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - August 19, 2006.]

  Revision    [$Id: verFormula.c,v 1.00 2006/08/19 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "ver.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// the list of operation symbols to be used in expressions
#define VER_PARSE_SYM_OPEN    '('   // opening parenthesis
#define VER_PARSE_SYM_CLOSE   ')'   // closing parenthesis
#define VER_PARSE_SYM_CONST0  '0'   // constant 0
#define VER_PARSE_SYM_CONST1  '1'   // constant 1
#define VER_PARSE_SYM_NEGBEF1 '!'   // negation before the variable
#define VER_PARSE_SYM_NEGBEF2 '~'   // negation before the variable
#define VER_PARSE_SYM_AND     '&'   // logic AND
#define VER_PARSE_SYM_OR      '|'   // logic OR
#define VER_PARSE_SYM_XOR     '^'   // logic XOR
#define VER_PARSE_SYM_MUX1    '?'   // first symbol of MUX
#define VER_PARSE_SYM_MUX2    ':'   // second symbol of MUX

// the list of opcodes (also specifying operation precedence)
const VER_PARSE_OPER_NEG =     7    // negation (highest precedence)
const VER_PARSE_OPER_AND =     6    // logic AND
const VER_PARSE_OPER_XOR =     5    // logic EXOR   (a'b | ab')   
const VER_PARSE_OPER_OR =      4    // logic OR
const VER_PARSE_OPER_EQU =     3    // equvalence   (a'b'| ab )
const VER_PARSE_OPER_MUX =     2    // MUX(a,b,c)   (ab | a'c )
const VER_PARSE_OPER_MARK =    1    // OpStack token standing for an opening parenthesis

// these are values of the internal Flag
const VER_PARSE_FLAG_START =   1    // after the opening parenthesis 
const VER_PARSE_FLAG_VAR =     2    // after operation is received
const VER_PARSE_FLAG_OPER =    3    // after operation symbol is received
const VER_PARSE_FLAG_ERROR =   4    // when error is detected

// static Hop_Obj_t * Ver_FormulaParserTopOper( Hop_Man_t * pMan, Vec_Ptr_t * vStackFn, int Oper );

// static int         Ver_FormulaParserFindVar( char * pString, Vec_Ptr_t * vNames );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Parser of the formula encountered in assign statements.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ver_FormulaParser(pFormula string, pMan any, vNames *Vec_Ptr_t, vStackFn *Vec_Ptr_t, vStackOp *Vec_Int_t, pErrorMessage string) any {
    var pTemp string;
    var bFunc, bTemp *Hop_Obj_t;
    var nParans, Flag int32;
    var Oper, Oper1, Oper2 int32;
    var v int32;

    // clear the stacks and the names
    Vec_PtrClear( vNames );
    Vec_PtrClear( vStackFn );
    Vec_IntClear( vStackOp );

    if !int32(strings.Compare(pFormula, "0")) || !int32(strings.Compare(pFormula, "1\'b0")) {
        return Hop_ManConst0((Hop_Man_t *)pMan); }
    if !int32(strings.Compare(pFormula, "1")) || !int32(strings.Compare(pFormula, "1\'b1")) {
        return Hop_ManConst1((Hop_Man_t *)pMan); }

    // make sure that the number of opening and closing parentheses is the same
    nParans = 0;
    for  pTemp = pFormula; *pTemp; pTemp++  {
        if *pTemp == '(' {
            nParans++;
        } else if *pTemp == ')' {
            nParans--; }
    }
    if nParans != 0 {
        pErrorMessage = fmt.Sprintf("Parse_FormulaParser(): Different number of opening and closing parentheses ().");
        return nil;
    }
 
    // add parentheses
    pTemp = pFormula + len(pFormula) + 2;
    *pTemp-- = 0; *pTemp = ')';
    for  --pTemp != pFormula  {
        *pTemp = *(pTemp - 1);
    }
    *pTemp = '(';

    // perform parsing
    Flag = VER_PARSE_FLAG_START;
    for  pTemp = pFormula; *pTemp; pTemp++  {
        switch  *pTemp  {
        // skip all spaces, tabs, and end-of-lines
        case ' ',
         '\t',
         '\r',
         '\n':
            continue;
/*
        // treat Constant 0 as a variable
        case VER_PARSE_SYM_CONST0:
            Vec_PtrPush( vStackFn, Hop_ManConst0(pMan) );  // Cudd_Ref( Hop_ManConst0(pMan) );
            if ( Flag == VER_PARSE_FLAG_VAR )
            {
                sprintf( pErrorMessage, "Parse_FormulaParser(): No operation symbol before constant 0." );
                Flag = VER_PARSE_FLAG_ERROR; 
                break;
            }
            Flag = VER_PARSE_FLAG_VAR; 
            break;

        // the same for Constant 1
        case VER_PARSE_SYM_CONST1:
            Vec_PtrPush( vStackFn, Hop_ManConst1(pMan) );  //  Cudd_Ref( Hop_ManConst1(pMan) );
            if ( Flag == VER_PARSE_FLAG_VAR )
            {
                sprintf( pErrorMessage, "Parse_FormulaParser(): No operation symbol before constant 1." );
                Flag = VER_PARSE_FLAG_ERROR; 
                break;
            }
            Flag = VER_PARSE_FLAG_VAR; 
            break;
*/
        case VER_PARSE_SYM_NEGBEF1,
         VER_PARSE_SYM_NEGBEF2:
            if Flag == VER_PARSE_FLAG_VAR {// if NEGBEF follows a variable, AND is assumed
                pErrorMessage = fmt.Sprintf("Parse_FormulaParser(): Variable before negation.");
                Flag = VER_PARSE_FLAG_ERROR; 
                break;
            }
            Vec_IntPush( vStackOp, VER_PARSE_OPER_NEG );
            break;

        case VER_PARSE_SYM_AND,
         VER_PARSE_SYM_OR,
         VER_PARSE_SYM_XOR,
         VER_PARSE_SYM_MUX1,
         VER_PARSE_SYM_MUX2:
            if Flag != VER_PARSE_FLAG_VAR {
                pErrorMessage = fmt.Sprintf("Parse_FormulaParser(): There is no variable before AND, EXOR, or OR.");
                Flag = VER_PARSE_FLAG_ERROR; 
                break;
            }
            if *pTemp == VER_PARSE_SYM_AND {
                Vec_IntPush( vStackOp, VER_PARSE_OPER_AND );
            } else if *pTemp == VER_PARSE_SYM_OR {
                Vec_IntPush( vStackOp, VER_PARSE_OPER_OR );
            } else if *pTemp == VER_PARSE_SYM_XOR {
                Vec_IntPush( vStackOp, VER_PARSE_OPER_XOR );
            } else if *pTemp == VER_PARSE_SYM_MUX1 {
                Vec_IntPush( vStackOp, VER_PARSE_OPER_MUX ); }
//            else if ( *pTemp == VER_PARSE_SYM_MUX2 )
//                Vec_IntPush( vStackOp, VER_PARSE_OPER_MUX );
            Flag = VER_PARSE_FLAG_OPER; 
            break;

        case VER_PARSE_SYM_OPEN:
            if Flag == VER_PARSE_FLAG_VAR {
                pErrorMessage = fmt.Sprintf("Parse_FormulaParser(): Variable before a parenthesis.");
                Flag = VER_PARSE_FLAG_ERROR; 
                break;
            }
            Vec_IntPush( vStackOp, VER_PARSE_OPER_MARK );
            // after an opening bracket, it feels like starting over again
            Flag = VER_PARSE_FLAG_START; 
            break;

        case VER_PARSE_SYM_CLOSE:
            if Vec_IntSize( vStackOp ) {
                for {
                    if !Vec_IntSize( vStackOp ) {
                        pErrorMessage = fmt.Sprintf("Parse_FormulaParser(): There is no opening parenthesis\n");
                        Flag = VER_PARSE_FLAG_ERROR; 
                        break;
                    }
                    Oper = Vec_IntPop( vStackOp );
                    if Oper == VER_PARSE_OPER_MARK {
                        break; }
                    // skip the second MUX operation
//                    if ( Oper == VER_PARSE_OPER_MUX2 )
//                    {
//                        Oper = Vec_IntPop( vStackOp );
//                        assert( Oper == VER_PARSE_OPER_MUX1 );
//                    }

                    // perform the given operation
                    if Ver_FormulaParserTopOper( (Hop_Man_t *)pMan, vStackFn, Oper ) == nil {
                        pErrorMessage = fmt.Sprintf("Parse_FormulaParser(): Unknown operation\n");
                        return nil;
                    }
                }
            
            } else {
                pErrorMessage = fmt.Sprintf("Parse_FormulaParser(): There is no opening parenthesis\n");
                Flag = VER_PARSE_FLAG_ERROR; 
                break;
            }
            if Flag != VER_PARSE_FLAG_ERROR {
                Flag = VER_PARSE_FLAG_VAR; } 
            break;


        default:
            // scan the next name
            v = Ver_FormulaParserFindVar( pTemp, vNames );
            if *pTemp == '\\' {
                pTemp++; }
            pTemp += int32((ABC_PTRUINT_T))Vec_PtrEntry( vNames, 2*v ) - 1;

            // assume operation AND, if vars follow one another
            if Flag == VER_PARSE_FLAG_VAR {
                pErrorMessage = fmt.Sprintf("Parse_FormulaParser(): Incorrect state.");
                return nil;
            }
            bTemp = Hop_IthVar( (Hop_Man_t *)pMan, v );
            Vec_PtrPush( vStackFn, bTemp ); //  Cudd_Ref( bTemp );
            Flag = VER_PARSE_FLAG_VAR; 
            break;
        }

        if Flag == VER_PARSE_FLAG_ERROR {
            break;      // error exit
        } else if Flag == VER_PARSE_FLAG_START {
            continue;  //  go on parsing
        } else if Flag == VER_PARSE_FLAG_VAR {
            for {  // check if there are negations in the OpStack     
                if !Vec_IntSize(vStackOp) {
                    break; }
                Oper = Vec_IntPop( vStackOp );
                if Oper != VER_PARSE_OPER_NEG {
                    Vec_IntPush( vStackOp, Oper );
                    break;
                
                } else {
//                      Vec_PtrPush( vStackFn, Cudd_Not(Vec_PtrPop(vStackFn)) );
                      Vec_PtrPush( vStackFn, Hop_Not((Hop_Obj_t *)Vec_PtrPop(vStackFn)) );
                }
            }
        } else { // if ( Flag == VER_PARSE_FLAG_OPER )
            for {  // execute all the operations in the OpStack
               // with precedence higher or equal than the last one
                Oper1 = Vec_IntPop( vStackOp ); // the last operation
                if !Vec_IntSize(vStackOp) {  // if it is the only operation, push it back
                    Vec_IntPush( vStackOp, Oper1 );
                    break;
                }
                Oper2 = Vec_IntPop( vStackOp ); // the operation before the last one
                if Oper2 >= Oper1 && !(Oper1 == Oper2 && Oper1 == VER_PARSE_OPER_MUX) {  // if Oper2 precedence is higher or equal, execute it
                    if Ver_FormulaParserTopOper( (Hop_Man_t *)pMan, vStackFn, Oper2 ) == nil {
                        pErrorMessage = fmt.Sprintf("Parse_FormulaParser(): Unknown operation\n");
                        return nil;
                    }
                    Vec_IntPush( vStackOp,  Oper1 );     // push the last operation back
                
                } else {  // if Oper2 precedence is lower, push them back and done
                    Vec_IntPush( vStackOp, Oper2 );
                    Vec_IntPush( vStackOp, Oper1 );
                    break;
                }
            } }
    }

    if Flag != VER_PARSE_FLAG_ERROR {
        if Vec_PtrSize(vStackFn) {    
            bFunc = (Hop_Obj_t *)Vec_PtrPop(vStackFn);
            if !Vec_PtrSize(vStackFn) {
                if !Vec_IntSize(vStackOp) {
//                    Cudd_Deref( bFunc );
                    return bFunc;
                
                } else {
                    pErrorMessage = fmt.Sprintf("Parse_FormulaParser(): Something is left in the operation stack\n"); }
            } else {
                pErrorMessage = fmt.Sprintf("Parse_FormulaParser(): Something is left in the function stack\n"); }
        
        } else {
            pErrorMessage = fmt.Sprintf("Parse_FormulaParser(): The input string is empty\n"); }
    }
//    Cudd_Ref( bFunc );
//    Cudd_RecursiveDeref( dd, bFunc );
    return nil;
}

/**Function*************************************************************

  Synopsis    [Performs the operation on the top entries in the stack.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ver_FormulaParserTopOper(pMan *Hop_Man_t, vStackFn *Vec_Ptr_t, Oper int32) *Hop_Obj_t {
    var bArg0, bArg1, bArg2, bFunc *Hop_Obj_t;
    // perform the given operation
    bArg2 = (Hop_Obj_t *)Vec_PtrPop( vStackFn );
    bArg1 = (Hop_Obj_t *)Vec_PtrPop( vStackFn );
    if Oper == VER_PARSE_OPER_AND {
        bFunc = Hop_And( pMan, bArg1, bArg2 );
    } else if Oper == VER_PARSE_OPER_XOR {
        bFunc = Hop_Exor( pMan, bArg1, bArg2 );
    } else if Oper == VER_PARSE_OPER_OR {
        bFunc = Hop_Or( pMan, bArg1, bArg2 );
    } else if Oper == VER_PARSE_OPER_EQU {
        bFunc = Hop_Not( Hop_Exor( pMan, bArg1, bArg2 ) );
    } else if Oper == VER_PARSE_OPER_MUX {
        bArg0 = (Hop_Obj_t *)Vec_PtrPop( vStackFn );
//        bFunc = Cudd_bddIte( dd, bArg0, bArg1, bArg2 );  Cudd_Ref( bFunc );
        bFunc = Hop_Mux( pMan, bArg0, bArg1, bArg2 ); 
//        Cudd_RecursiveDeref( dd, bArg0 );
//        Cudd_Deref( bFunc );
    
    } else {
        return nil; }
//    Cudd_Ref( bFunc );
//    Cudd_RecursiveDeref( dd, bArg1 );
//    Cudd_RecursiveDeref( dd, bArg2 );
    Vec_PtrPush( vStackFn,  bFunc );
    return bFunc;
}

/**Function*************************************************************

  Synopsis    [Returns the index of the new variable found.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ver_FormulaParserFindVar(pString string, vNames *Vec_Ptr_t) int32 {
    var pTemp, pTemp2 string;
    var nLength, nLength2, i int32;
    // start the string
    pTemp = pString;
    // find the end of the string delimited by other characters
    if *pTemp == '\\' {
        pString++;
        for  *pTemp && *pTemp != ' '  {
            pTemp++;
        }
    
    } else {
        for  *pTemp && *pTemp != ' ' && *pTemp != '\t' && *pTemp != '\r' && *pTemp != '\n' && *pTemp != ',' && *pTemp != '}' && 
                *pTemp != VER_PARSE_SYM_OPEN && *pTemp != VER_PARSE_SYM_CLOSE && 
                *pTemp != VER_PARSE_SYM_NEGBEF1 && *pTemp != VER_PARSE_SYM_NEGBEF2 && 
                *pTemp != VER_PARSE_SYM_AND && *pTemp != VER_PARSE_SYM_OR && *pTemp != VER_PARSE_SYM_XOR && 
                *pTemp != VER_PARSE_SYM_MUX1 && *pTemp != VER_PARSE_SYM_MUX2  {
                pTemp++;
        }
    }
    // look for this string in the array
    nLength = pTemp - pString;
    for  i = 0; i < Vec_PtrSize(vNames)/2; i++  {
        nLength2 = int32((ABC_PTRUINT_T))Vec_PtrEntry( vNames, 2*i + 0 );
        if nLength2 != nLength {
            continue; }
        pTemp2   = (char *)Vec_PtrEntry( vNames, 2*i + 1 );
        if strncmp( pString, pTemp2, nLength ) {
            continue; }
        return i;
    }
    // could not find - add and return the number
    Vec_PtrPush( vNames, (void *)(ABC_PTRUINT_T)nLength );
    Vec_PtrPush( vNames, pString );
    return i;
}

/**Function*************************************************************

  Synopsis    [Returns the AIG representation of the reduction formula.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ver_FormulaReduction(pFormula string, pMan any, vNames *Vec_Ptr_t, pErrorMessage string) any {
    var pRes *Hop_Obj_t  = nil;
    var v, fCompl int32;
    var Symbol byte;

    // get the operation
    Symbol = *pFormula++;
    fCompl = ( Symbol == '~' );
    if fCompl != 0 {
        Symbol = *pFormula++; }
    // check the operation
    if Symbol != '&' && Symbol != '|' && Symbol != '^' {
        pErrorMessage = fmt.Sprintf("Ver_FormulaReduction(): Unknown operation (%c)\n", Symbol);
        return nil;
    }
    // skip the brace
    for  *pFormula++ != '{'  {; }
    // parse the names
    Vec_PtrClear( vNames );
    for  *pFormula != '}'  {
        v = Ver_FormulaParserFindVar( pFormula, vNames );
        pFormula += int32((ABC_PTRUINT_T))Vec_PtrEntry( vNames, 2*v );
        for  *pFormula == ' ' || *pFormula == ','  {
            pFormula++;
        }
    }
    // compute the function
    if Symbol == '&' {
        pRes = Hop_CreateAnd( (Hop_Man_t *)pMan, Vec_PtrSize(vNames)/2 );
    } else if Symbol == '|' {
        pRes = Hop_CreateOr( (Hop_Man_t *)pMan, Vec_PtrSize(vNames)/2 );
    } else if Symbol == '^' {
        pRes = Hop_CreateExor( (Hop_Man_t *)pMan, Vec_PtrSize(vNames)/2 ); }
    return Hop_NotCond( pRes, fCompl );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

