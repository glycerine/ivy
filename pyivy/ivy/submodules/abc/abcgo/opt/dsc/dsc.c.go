package dsc

import (
	"fmt"
	"os"
)

/**CFile****************************************************************

  FileName    [dsc.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Disjoint support decomposition - ICCD'15]

  Synopsis    [Disjoint-support decomposition with cofactoring and boolean difference analysis
               from V. Callegaro, F. S. Marranghello, M. G. A. Martins, R. P. Ribas and A. I. Reis,
               "Bottom-up disjoint-support decomposition based on cofactor and boolean difference analysis," ICCD'15]

  Author      [Vinicius Callegaro, Mayler G. A. Martins, Felipe S. Marranghello, Renato P. Ribas and Andre I. Reis]

  Affiliation [UFRGS - Federal University of Rio Grande do Sul - Brazil]

  Date        [Ver. 1.0. Started - October 24, 2014.]

  Revision    [$Id: dsc.h,v 1.00 2014/10/24 00:00:00 vcallegaro Exp $]

***********************************************************************/

//#include "dsc.h"
//#include <assert.h>
//#include "misc/util/utilTruth.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////
/*
    This code performs truth-table-based decomposition for 6-variable functions.
    Representation of operations:
    ! = not;
    (ab) = a and b;
    [ab] = a xor b;
*/
   
type Dsc_node_t struct {
    pNegCof *word;
    pPosCof *word;
    pBoolDiff *word;
    unsigned int on[DSC_MAX_VAR+1]; // pos cofactor spec - first element denotes the size of the array
    unsigned int off[DSC_MAX_VAR+1]; // neg cofactor spec - first element denotes the size of the array
    exp [DSC_MAX_STR]byte;
}

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

func xorInPlace(pOut *word, pIn2 *word, nWords int32) {
    var w int32;
    for  w = 0; w < nWords; w++  {
        pOut[w] ^= pIn2[w];
    }
}

func dsc_debug_node(pNode *Dsc_node_t, nVars int32, TRUTH_WORDS int32) {
    var i int32;
    fmt.Printf("Node:\t%s\n", pNode.exp);
    fmt.Printf("\tneg cof:\t");Abc_TtPrintHexRev(os.Stdout, pNode.pNegCof, nVars);
    fmt.Printf("\tpos cof:\t");Abc_TtPrintHexRev(os.Stdout, pNode.pPosCof, nVars);
    fmt.Printf("\tbool diff:\t");Abc_TtPrintHexRev(os.Stdout, pNode.pBoolDiff, nVars);
    fmt.Printf("\toff:\t");
    for i=1;i<=int32(pNode.off[0]);i++ {
        fmt.Printf("%c%c", (pNode.off[i] & uint32(1)) ? ' ' : '!', 'a'+(pNode.off[i] >> 1));
    }
    fmt.Printf("\ton:\t");
    for i=1;i<=int32(pNode.on[0]);i++ {
        fmt.Printf("%c%c", (pNode.on[i] & uint32(1)) ? ' ' : '!', 'a'+(pNode.on[i] >> 1));
    }
    fmt.Printf("\n");
}

func dsc_and_test(ni *Dsc_node_t, nj *Dsc_node_t, TRUTH_WORDS int32, ci *int32, cj *int32) int32 {
            if Abc_TtEqual(ni.pNegCof, nj.pNegCof, TRUTH_WORDS) {*ci=1; *cj=1; return 1;
    } else     if Abc_TtEqual(ni.pNegCof, nj.pPosCof, TRUTH_WORDS) {*ci=1; *cj=0; return 1;
    } else     if Abc_TtEqual(ni.pPosCof, nj.pNegCof, TRUTH_WORDS) {*ci=0; *cj=1; return 1;
    } else     if Abc_TtEqual(ni.pPosCof, nj.pPosCof, TRUTH_WORDS) {*ci=0; *cj=0; return 1;}
    return 0;
}

func dsc_xor_test(ni *Dsc_node_t, nj *Dsc_node_t, TRUTH_WORDS int32) int32 {
    return Abc_TtEqual(ni.pBoolDiff, nj.pBoolDiff, TRUTH_WORDS);
}

func concat(target string, begin byte, end byte, s1 string, s1Polarity int32, s2 string, s2Polarity int32) {
    *target++ = begin;
    //s1
    if !(s1Polarity != 0) {
        *target++ = '!'; }
    for *s1 != '\0' {
        *target++ = *s1++;
    }
    // s2
    if !(s2Polarity != 0) {
        *target++ = '!'; }
    for *s2 != '\0' {
        *target++ = *s2++;
    }
    // end
    *target++ = end;
    *target = '\0';
}

static inline void cubeCofactor(word * pTruth, unsigned int * cubeCof, int TRUTH_WORDS) {
    var size int32  = cubeCof[0];
    var i int32;
    for i = 1; i <= size; i++ {
        unsigned int c = cubeCof[i];
        if c & uint32(1) {
            Abc_TtCofactor1(pTruth, TRUTH_WORDS, c >> 1);
        } else {
            Abc_TtCofactor0(pTruth, TRUTH_WORDS, c >> 1);
        }
    }
}

static inline void merge(unsigned int * pOut, unsigned int * pIn) {
    int elementsToCopy = pIn[0];
    var i, j int32;
    for i = pOut[0]+1, j = 1; j <= elementsToCopy; i, j = i+1, j+1 {
        pOut[i] = pIn[j];
    }
    pOut[0] += elementsToCopy;
}

func dsc_and_group(pOut *Dsc_node_t, ni *Dsc_node_t, niPolarity int32, nj *Dsc_node_t, njPolarity int32, nVars int32, TRUTH_WORDS int32) {
    unsigned int* xiOFF, * xiON, * xjOFF, * xjON;
    // expression
    concat(pOut.exp, '(', ')', ni.exp, niPolarity, nj.exp, njPolarity);
    // ON-OFF
    if niPolarity != 0 {
        xiOFF = ni.off;
        xiON = ni.on;
    } else {
        xiOFF = ni.on;
        xiON = ni.off;
    }
    if njPolarity != 0 {
        xjOFF = nj.off;
        xjON = nj.on;
    } else {
        xjOFF = nj.on;
        xjON = nj.off;
    }
    // creating both the new OFF specification and negative cofactor of the new group
    {
        // first element of the array represents the size of the cube-cofactor
        var xiOFFSize int32  = xiOFF[0];
        var xjOFFSize int32  = xjOFF[0];
        if xiOFFSize <= xjOFFSize {
            var i int32;
            pOut.off[0] = xiOFFSize; // set the number of elements
            for i = 1; i <= xiOFFSize; i++ {
                pOut.off[i] = xiOFF[i];
            }
        } else {
            var i int32;
            pOut.off[0] = xjOFFSize; // set the number of elements
            for i = 1; i <= xjOFFSize; i++ {
                pOut.off[i] = xjOFF[i];
            }
        }
        // set the negative cofactor of the new group
        if niPolarity != 0 {
            pOut.pNegCof = ni.pNegCof;
        } else {
            pOut.pNegCof = ni.pPosCof;
        }
    }
    // creating both new ON specification and positive cofactor of the new group
    {
        var i int32;
        var j int32;
        unsigned int xiONSize = xiON[0];
        unsigned int xjONSize = xjON[0];
        pOut.on[0] = xiONSize + xjONSize;
        for i = 1; i <= int32(xiONSize); i++ {
            pOut.on[i] = xiON[i];
        }
        for j = 1; j <= int32(xjONSize); j++ {
            pOut.on[i++] = xjON[j];
        }
        // set the positive cofactor of the new group
        if xiONSize >= xjONSize {
            if niPolarity != 0 {
                pOut.pPosCof = ni.pPosCof;
            } else {
                pOut.pPosCof = ni.pNegCof;
            }
            cubeCofactor(pOut.pPosCof, xjON, TRUTH_WORDS);
        } else {
            if njPolarity != 0 {
                pOut.pPosCof = nj.pPosCof;
            } else {
                pOut.pPosCof = nj.pNegCof;
            }
            cubeCofactor(pOut.pPosCof, xiON, TRUTH_WORDS);
        }
    }
    // set the boolean difference of the new group
    if njPolarity != 0 {
        pOut.pBoolDiff = nj.pNegCof;
    } else {
        pOut.pBoolDiff = nj.pPosCof;
    }
    xorInPlace(pOut.pBoolDiff, pOut.pPosCof, TRUTH_WORDS);
}

func dsc_xor_group(pOut *Dsc_node_t, ni *Dsc_node_t, nj *Dsc_node_t, nVars int32, TRUTH_WORDS int32) {
    //
    unsigned int * xiOFF = ni.off;
    unsigned int * xiON = ni.on;
    unsigned int * xjOFF = nj.off;
    unsigned int * xjON = nj.on;
    //
    int xiOFFSize = xiOFF[0];
    int xiONSize = xiON[0];
    int xjOFFSize = xjOFF[0];
    int xjONSize = xjON[0];
    // minCubeCofs
    var minCCSize int32  = xiOFFSize;
    var minCCPolarity int32;
    var minCCNode *Dsc_node_t  = ni;
    // expression
    concat(pOut.exp, '[', ']', ni.exp, 1, nj.exp, 1);
    if minCCSize > xiONSize {
        minCCSize = xiONSize;
        minCCPolarity = 1;
        //minCCNode = ni;
    }
    if minCCSize > xjOFFSize {
        minCCSize = xjOFFSize;
        minCCPolarity = 0;
        minCCNode = nj;
    }
    if minCCSize > xjONSize {
        minCCSize = xjONSize;
        minCCPolarity = 1;
        minCCNode = nj;
    }
    //
    if minCCNode == ni {
        if minCCPolarity != 0 {
            // gOFF = xiON, xjON
            pOut.pNegCof = nj.pPosCof;
            cubeCofactor(pOut.pNegCof, xiON, TRUTH_WORDS);
            // gON = xiON, xjOFF
            pOut.pPosCof = nj.pNegCof;
            cubeCofactor(pOut.pPosCof, xiON, TRUTH_WORDS);
        } else {
            // gOFF = xiOFF, xjOFF
            pOut.pNegCof = nj.pNegCof;
            cubeCofactor(pOut.pNegCof, xiOFF, TRUTH_WORDS);
            // gON = xiOFF, xjON
            pOut.pPosCof = nj.pPosCof;
            cubeCofactor(pOut.pPosCof, xiOFF, TRUTH_WORDS);
        }
    }else  {
        if minCCPolarity != 0 {
            // gOFF = xjON, xiON
            pOut.pNegCof = ni.pPosCof;
            cubeCofactor(pOut.pNegCof, xjON, TRUTH_WORDS);
            // gON = xjON, xiOFF
            pOut.pPosCof = ni.pNegCof;
            cubeCofactor(pOut.pPosCof, xjON, TRUTH_WORDS);
        } else {
            // gOFF = xjOFF, xiOFF
            pOut.pNegCof = ni.pNegCof;
            cubeCofactor(pOut.pNegCof, xjOFF, TRUTH_WORDS);
            // gON = xjOFF, xiON
            pOut.pPosCof = ni.pPosCof;
            cubeCofactor(pOut.pPosCof, xjOFF, TRUTH_WORDS);
        }
    }
    // bool diff
    pOut.pBoolDiff = ni.pBoolDiff;
    // evaluating specs
    // off spec
    pOut.off[0] = 0;
    if (xiOFFSize+xjOFFSize) <= (xiONSize+xjONSize) {
        merge(pOut.off, xiOFF);
        merge(pOut.off, xjOFF);
    } else {
        merge(pOut.off, xiON);
        merge(pOut.off, xjON);
    }
    // on spec
    pOut.on[0] = 0;
    if (xiOFFSize+xjONSize) <= (xiONSize+xjOFFSize) {
        merge(pOut.on, xiOFF);
        merge(pOut.on, xjON);
    } else {
        merge(pOut.on, xiON);
        merge(pOut.on, xjOFF);
    }
}

/**
 * memory allocator with a capacity of storing 3*nVars
 * truth-tables for negative and positive cofactors and
 * the boolean difference for each input variable
 */
func Dsc_alloc_pool(nVars int32) *word {
    // inlining of macro ABC_ALLOC
    return ((word *) malloc(sizeof(word) * (size_t)(3 * Abc_TtWordNum(nVars) * nVars)));
}

/**
 * just free the memory pool
 */
func Dsc_free_pool(pool *word) {
    pool = nil;
}

/**
 * This method implements the paper proposed by V. Callegaro, F. S. Marranghello, M. G. A. Martins, R. P. Ribas and A. I. Reis,
 * entitled "Bottom-up disjoint-support decomposition based on cofactor and boolean difference analysis", presented at ICCD 2015.
 * pTruth: pointer for the truth table representing the target function.
 * nVarsInit: the number of variables of the truth table of the target function.
 * pRes: pointer for storing the resulting decomposition, whenever a decomposition can be found.
 * pool: NULL or a pointer for with a capacity of storing 3*nVars truth-tables. IF NULL, the function will allocate and free the memory of each call.
 * (the results presented on ICCD paper are running this method with NULL for the memory pool).
 * The method returns 0 if a full decomposition was found and a negative value otherwise.
 */
func Dsc_Decompose(pTruth *word, nVarsInit int32, pRes string, pool *word) int32 {
    int TRUTH_WORDS = Abc_TtWordNum(nVarsInit);
    int NEED_POOL_ALLOC = (pool == nil);

    var nodes [DSC_MAX_VAR]Dsc_node_t;
    var newNodes [DSC_MAX_VAR]*Dsc_node_t;
    var oldNodes [DSC_MAX_VAR]*Dsc_node_t;

    var freeNodes [DSC_MAX_VAR]Dsc_node_t; // N is the maximum number of possible groups.
    var f int32; // f represent the next free position in the freeNodes array
    var o int32; // o stands for the number of already tested nodes
    var n int32; // n will represent the number of current nodes (i.e. support)

    pRes[0] = '\0';
    pRes[1] = '\0';

    if NEED_POOL_ALLOC {
        // inlining of macro ABC_ALLOC
        pool = ((word *) malloc(sizeof(word) * (size_t)(3 * TRUTH_WORDS * nVarsInit))); }

    // block for the node data allocation
    {
        // pointer for the next free truth word
        var pNextTruth *word  = pool;
        var iVar int32;
        for iVar = 0; iVar < nVarsInit; iVar++ {
            // negative cofactor
            Abc_TtCofactor0p(pNextTruth, pTruth, TRUTH_WORDS, iVar);
            // dont care test
            if !Abc_TtEqual(pNextTruth, pTruth, TRUTH_WORDS) {
                var node *Dsc_node_t  = &nodes[iVar];
                node.pNegCof = pNextTruth;
                // increment next truth pointer
                pNextTruth += TRUTH_WORDS;
                // positive cofactor
                node.pPosCof = pNextTruth;
                Abc_TtCofactor1p(node.pPosCof, pTruth, TRUTH_WORDS, iVar);
                // increment next truth pointer
                pNextTruth += TRUTH_WORDS;
                // boolean difference
                node.pBoolDiff = pNextTruth;
                Abc_TtXor(node.pBoolDiff, node.pNegCof, node.pPosCof, TRUTH_WORDS, 0);
                // increment next truth pointer
                pNextTruth += TRUTH_WORDS;
                // define on spec -
                node.on[0] = 1; node.on[1] = (iVar << 1) | uint32(1); // lit = i*2+1, when polarity=true
                // define off spec
                node.off[0] = 1; node.off[1] = iVar << 1;// lit=i*2 otherwise
                // store the node expression
                node.exp[0] = 'a'+iVar; // TODO fix the variable names
                node.exp[1] = '\0';
                // add the node to the newNodes array
                newNodes[n++] = node;
            }
        }
    }
    //const int initialSupport = n;
    if n == 0 {
        if NEED_POOL_ALLOC {
            pool = nil; }
        if Abc_TtIsConst0(pTruth, TRUTH_WORDS) {
            { if pRes { pRes[0] = '0', pRes[1] = '\0'; } }
            return 0;
        } else if Abc_TtIsConst1(pTruth, TRUTH_WORDS) {
            { if pRes { pRes[0] = '1', pRes[1] = '\0'; } }
            return 0;
        } else {
            Abc_Print(-1, "ERROR. No variable in the support of f, but f isn't constant!\n");
            return -1;
        }
    }
    for n > 0 {
        var tempN int32;
        var i, j, iPolarity, jPolarity int32;
        var ni *Dsc_node_t;
        var nj *Dsc_node_t;
        var newNode *Dsc_node_t = nil;
        for i = 0; i < n; i++ {
            ni = newNodes[i];
            newNode = nil;
            j = 0;
            for j < o {
                nj = oldNodes[j];
                if dsc_and_test(ni, nj, TRUTH_WORDS, &iPolarity, &jPolarity) != 0 {
                    newNode = &freeNodes[f++];
                    dsc_and_group(newNode, ni, iPolarity, nj, jPolarity, nVarsInit, TRUTH_WORDS);
                }
                // XOR test
                if (newNode == nil)  && ((dsc_xor_test(ni, nj, TRUTH_WORDS) != 0)) {
                    newNode = &freeNodes[f++];
                    dsc_xor_group(newNode, ni, nj, nVarsInit, TRUTH_WORDS);
                }
                if newNode != nil {
                    oldNodes[j] = oldNodes[--o];
                    break;
                } else {
                    j++;
                }
            }
            if newNode != nil {
                newNodes[tempN++] = newNode;
            } else {
                oldNodes[o++] = ni;
            }
        }
        n = tempN;
    }
    if o == 1 {
        var solution *Dsc_node_t  = oldNodes[0];
        if Abc_TtIsConst0(solution.pNegCof, TRUTH_WORDS) && Abc_TtIsConst1(solution.pPosCof, TRUTH_WORDS) {
            // Direct solution found
            if pRes {
                pRes = solution.exp; }
            if NEED_POOL_ALLOC {
                pool = nil; }
            return 0;
        } else if Abc_TtIsConst1(solution.pNegCof, TRUTH_WORDS) && Abc_TtIsConst0(solution.pPosCof, TRUTH_WORDS) {
            // Complementary solution found
            if pRes {
                pRes[0] = '!';
                strcpy( &pRes[1], solution.exp);
            }
            if NEED_POOL_ALLOC {
                pool = nil; }
            return 0;
        } else {
            fmt.Printf("DSC ERROR: Final DSC node found, but differs from target function.\n");
            if NEED_POOL_ALLOC {
                pool = nil; }
            return -1;
        }
    }
    if NEED_POOL_ALLOC {
        pool = nil; }
    return -1;
}


/**Function*************************************************************
  Synopsis    [DSD formula manipulation.]
  Description [Code copied from dauDsd.c but changed DAU_MAX_VAR to DSC_MAX_VAR]
***********************************************************************/
var spStatic_Dsc_ComputeMatches_pMatches [DSC_MAX_VAR]int32;
func Dsc_ComputeMatches(p string) *int32 {
    
    var pNested [DSC_MAX_VAR]int32;
    var v, nNested int32;
    for  v = 0; p[v]; v++  {
        spStatic_Dsc_ComputeMatches_pMatches[v] = 0;
        if p[v] == '(' || p[v] == '[' || p[v] == '<' || p[v] == '{' {
            pNested[nNested++] = v;
        } else if p[v] == ')' || p[v] == ']' || p[v] == '>' || p[v] == '}' {
            spStatic_Dsc_ComputeMatches_pMatches[pNested[--nNested]] = v; }
        assert(nNested < DSC_MAX_VAR, "nNested < DSC_MAX_VAR");
    }
    assert(nNested == 0, "nNested == 0");
    return spStatic_Dsc_ComputeMatches_pMatches;
}

/**Function*************************************************************
  Synopsis    [DSD formula manipulation.]
  Description [Code copied from dauDsd.c but changed DAU_MAX_VAR to DSC_MAX_VAR]
***********************************************************************/
func Dsc_CountAnds_rec(pStr string, p *string, pMatches *int32) int32 {
    if **p == '!' {
        (*p)++; }
    for  (**p >= 'A' && **p <= 'F') || (**p >= '0' && **p <= '9')  {
        (*p)++;
    }
    if **p == '<' {
        var q string  = pStr + pMatches[*p - pStr];
        if *(q+1) == '{' {
            *p = q+1; }
    }
    if **p >= 'a' && **p <= 'z' { // var
        return 0; }
    if **p == '(' || **p == '[' { // and/or/xor
    
        var Counter int32;
        var AddOn int32;
        if **p == '(' {
            AddOn = 1;
        } else {
            AddOn = 3;
        }
        var q string  = pStr + pMatches[ *p - pStr ];
        assert(*q == **p + 1 + (**p != '('), "*q == **p + 1 + (**p != '(')");
        for  (*p)++; *p < q; (*p)++  {
            Counter += AddOn + Dsc_CountAnds_rec( pStr, p, pMatches );
        }
        assert(*p == q, "*p == q");
        return Counter - AddOn;
    }
    if **p == '<' || **p == '{' { // mux
    
        var Counter int32  = 3;
        var q string  = pStr + pMatches[ *p - pStr ];
        assert(*q == **p + 1 + (**p != '('), "*q == **p + 1 + (**p != '(')");
        for  (*p)++; *p < q; (*p)++  {
            Counter += Dsc_CountAnds_rec( pStr, p, pMatches );
        }
        assert(*p == q, "*p == q");
        return Counter;
    }
    assert(false, "0");
    return 0;
}
/**Function*************************************************************
  Synopsis    [DSD formula manipulation.]
  Description [Code copied from dauDsd.c but changed DAU_MAX_VAR to DSC_MAX_VAR]
***********************************************************************/
func Dsc_CountAnds(pDsd string) int32 {
    if pDsd[1] == 0 {
        return 0; }
    return Dsc_CountAnds_rec( pDsd, &pDsd, Dsc_ComputeMatches(pDsd) );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END
