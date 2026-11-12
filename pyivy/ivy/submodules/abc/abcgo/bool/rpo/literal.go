package rpo

/**CFile****************************************************************

  FileName    [literal.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [RPO]

  Synopsis    [Literal structure]

  Author      [Mayler G. A. Martins / Vinicius Callegaro]
  
  Affiliation [UFRGS]

  Date        [Ver. 1.0. Started - May 08, 2013.]

  Revision    [$Id: literal.h,v 1.00 2013/05/08 00:00:00 mgamartins Exp $]

 ***********************************************************************/




////////////////////////////////////////////////////////////////////////
///                          INCLUDES                                ///
////////////////////////////////////////////////////////////////////////

//#include <stdio.h>
//#include <stdlib.h>
//#include "bool/kit/kit.h"
//#include "misc/vec/vec.h"
//#include "misc/util/abc_global.h"

//ABC_NAMESPACE_HEADER_START


////////////////////////////////////////////////////////////////////////
///                         PARAMETERS                               ///
////////////////////////////////////////////////////////////////////////

// associations
 type Operator_t int32
const (
    LIT_NONE Operator_t = 0
 // 0: unknown
    LIT_AND Operator_t = LIT_NONE + 1
 // 1: AND association
    LIT_OR Operator_t = LIT_AND + 1
 // 2: OR association
    LIT_XOR Operator_t = LIT_OR + 1 // 3: XOR association (not used yet)

)

 


   

type Literal_t struct {
    transition  *uint32;
    function  *uint32;
    expression  *Vec_Str_t;
}


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Compute the positive transition]

  Description [The inputs are a function, the number of variables and a variable index and the output is a function]
               
  SideEffects [Should this function be in kitTruth.c ?]

  SeeAlso     []
//
***********************************************************************/

func Lit_TruthPositiveTransition(pIn *uint32, pOut *uint32, nVars int32, varIdx int32) {
    var cof0 *uint32  = ABC_ALLOC (unsigned, Kit_TruthWordNum(nVars) );
    var cof1 *uint32  = ABC_ALLOC (unsigned, Kit_TruthWordNum(nVars) );
    var ncof0 *uint32;
    Kit_TruthCofactor0New(cof0, pIn,nVars,varIdx);
    Kit_TruthCofactor1New(cof1, pIn,nVars,varIdx);
    ncof0 = ABC_ALLOC (unsigned, Kit_TruthWordNum(nVars) );
    Kit_TruthNot(ncof0,cof0,nVars);
    Kit_TruthAnd(pOut,ncof0,cof1, nVars);
    ABC_FREE(cof0);
    ABC_FREE(ncof0);
    ABC_FREE(cof1);
}


/**Function*************************************************************

  Synopsis    [Compute the negative transition]

  Description [The inputs are a function, the number of variables and a variable index and the output is a function]
               
  SideEffects [Should this function be in kitTruth.c ?]

  SeeAlso     []

***********************************************************************/

func Lit_TruthNegativeTransition(pIn *uint32, pOut *uint32, nVars int32, varIdx int32) {
    var cof0 *uint32  = ABC_ALLOC (unsigned, Kit_TruthWordNum(nVars) );
    var cof1 *uint32  = ABC_ALLOC (unsigned, Kit_TruthWordNum(nVars) );
    var ncof1 *uint32;
    Kit_TruthCofactor0New(cof0, pIn,nVars,varIdx);
    Kit_TruthCofactor1New(cof1, pIn,nVars,varIdx);
    ncof1 = ABC_ALLOC (unsigned, Kit_TruthWordNum(nVars) );
    Kit_TruthNot(ncof1,cof1,nVars);
    Kit_TruthAnd(pOut,ncof1,cof0,nVars);
    ABC_FREE(cof0);
    ABC_FREE(cof1);
    ABC_FREE(ncof1);
}


/**Function*************************************************************

  Synopsis    [Create a literal given a polarity ]

  Description [The inputs are the function, index and its polarity and a the output is a literal]
               
  SideEffects []

  SeeAlso     []

 ***********************************************************************/

func Lit_Alloc(pTruth *uint32, nVars int32, varIdx int32, pol byte) *Literal_t {
    var transition *uint32;
    var function *uint32;
    var exp *Vec_Str_t;
    var lit *Literal_t;
    assert(pol == '+' || pol == '-', "pol == '+' || pol == '-'");
    transition = ABC_ALLOC(unsigned, Kit_TruthWordNum(nVars));
    if pol == '+' {
        Lit_TruthPositiveTransition(pTruth, transition, nVars, varIdx);
    } else {
        Lit_TruthNegativeTransition(pTruth, transition, nVars, varIdx);
    }
    if !Kit_TruthIsConst0(transition,nVars) {
        function = ABC_ALLOC(unsigned, Kit_TruthWordNum(nVars));
        Kit_TruthIthVar(function, nVars, varIdx);
        //Abc_Print(-2, "Allocating function %X %d %c \n", *function, varIdx, pol);
        exp = Vec_StrAlloc(5);
        if pol == '-' {
            Kit_TruthNot(function, function, nVars);
            Vec_StrPutC(exp, '!');
        }
        Vec_StrPutC(exp, byte(('a' + varIdx)));
        Vec_StrPutC(exp, '\0');
        lit = new(Literal_t);
        lit.function = function;
        lit.transition = transition;
        lit.expression = exp;
        return lit;
    } else {
        ABC_FREE(transition); // free the function.
        return nil;
    }
}


/**Function*************************************************************

  Synopsis    [Group 2 literals using AND or OR]

  Description []
               
  SideEffects []

  SeeAlso     []

 ***********************************************************************/

func Lit_GroupLiterals(lit1 *Literal_t, lit2 *Literal_t, op Operator_t, nVars int32) *Literal_t {
    var newFunction *uint32  = ABC_ALLOC(unsigned, Kit_TruthWordNum(nVars));
    var newTransition *uint32  = ABC_ALLOC(unsigned, Kit_TruthWordNum(nVars));
    var newExp *Vec_Str_t  = Vec_StrAlloc(lit1.expression.nSize + lit2.expression.nSize + 3);
    var newLiteral *Literal_t;
    var opChar byte  = '%';
    switch op {
        case LIT_AND:
        {
            //Abc_Print(-2, "Grouping AND %X %X \n", *lit1->function, *lit2->function);
            Kit_TruthAnd(newFunction, lit1.function, lit2.function, nVars);
            opChar = '*';
            break;
        }
        fallthrough
case LIT_OR:
        {
            //Abc_Print(-2, "Grouping OR %X %X \n", *lit1->function, *lit2->function);
            Kit_TruthOr(newFunction, lit1.function, lit2.function, nVars);
            opChar = '+';
            break;
        }
        fallthrough
default:
        {
            Abc_Print(-2, "Lit_GroupLiterals with op not defined.");
            //TODO Call ABC Abort
        }
    }

    Kit_TruthOr(newTransition, lit1.transition, lit2.transition, nVars);
    // create a new expression.
    Vec_StrPutC(newExp, '(');
    Vec_StrAppend(newExp, lit1.expression.pArray);
    //Vec_StrPop(newExp); // pop the \0
    Vec_StrPutC(newExp, opChar);
    Vec_StrAppend(newExp, lit2.expression.pArray);
    //Vec_StrPop(newExp); // pop the \0
    Vec_StrPutC(newExp, ')');
    Vec_StrPutC(newExp, '\0');

    newLiteral = new(Literal_t);
    newLiteral.function = newFunction;
    newLiteral.transition = newTransition;
    newLiteral.expression = newExp;
    return newLiteral;
}


/**Function*************************************************************

  Synopsis    [Create a const literal ]

  Description []
               
  SideEffects []

  SeeAlso     []

 ***********************************************************************/

func Lit_CreateLiteralConst(pTruth *uint32, nVars int32, constant int32) *Literal_t {
    var exp *Vec_Str_t  = Vec_StrAlloc(3);
    var lit *Literal_t;
    Vec_StrPutC(exp, byte(('0' + constant)));
    Vec_StrPutC(exp, '\0');
    lit = new(Literal_t);
    lit.expression = exp;
    lit.function = pTruth;
    lit.transition = pTruth; // wrong but no effect ###
    return lit;
}

func Lit_Copy(lit *Literal_t, nVars int32) *Literal_t {
    var newLit *Literal_t  = new(Literal_t);
    newLit.function = ABC_ALLOC(unsigned, Kit_TruthWordNum(nVars));
    Kit_TruthCopy(newLit.function,lit.function,nVars);
    newLit.transition = ABC_ALLOC(unsigned, Kit_TruthWordNum(nVars));
    Kit_TruthCopy(newLit.transition,lit.transition,nVars);
    newLit.expression = Vec_StrDup(lit.expression);
//    Abc_Print(-2,"Copying: %s\n",newLit->expression->pArray); 
    return newLit;
}

func Lit_PrintTT(tt *uint32, nVars int32) {
    var w int32;
    for w=nVars-1; w>=0; w-- {
            Abc_Print(-2, "%08X", tt[w]);
    }
}

func Lit_PrintExp(lit *Literal_t) {
    Abc_Print(-2, "%s", lit.expression.pArray);
}

/**Function*************************************************************

  Synopsis    [Delete the literal ]

  Description []
               
  SideEffects []

  SeeAlso     []

 ***********************************************************************/

func Lit_Free(lit *Literal_t) {
    if lit == nil {
        return;
    }
//    Abc_Print(-2,"Freeing: %s\n",lit->expression->pArray); 
    ABC_FREE(lit.function);
    ABC_FREE(lit.transition);
    Vec_StrFree(lit.expression);
    ABC_FREE(lit);
}

//ABC_NAMESPACE_HEADER_END



