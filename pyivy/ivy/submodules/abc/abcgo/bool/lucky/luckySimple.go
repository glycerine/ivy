package lucky

/**CFile****************************************************************

  FileName    [luckySimple.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Semi-canonical form computation package.]

  Synopsis    [Truth table minimization procedures.]

  Author      [Jake]

  Date        [Started - August 2012]

***********************************************************************/

//#include "luckyInt.h"

//ABC_NAMESPACE_IMPL_START

func setSwapInfoPtr(varsN int32) *swapInfo {
    var i int32;
    var x *swapInfo  = (swapInfo*) malloc(sizeof(swapInfo));
    x.posArray = (varInfo*) malloc (sizeof(varInfo)*(varsN+2));
    x.realArray = (int*) malloc (sizeof(int)*(varsN+2));
    x.varN = varsN;
    x.realArray[0]=varsN+100;
    for i=1;i<=varsN;i++ {
        x.posArray[i].position=i;
        x.posArray[i].direction=-1;
        x.realArray[i]=i;
    }    
    x.realArray[varsN+1]=varsN+10;
    return x;
}


func freeSwapInfoPtr(x *swapInfo) {
    free(x.posArray);
    free(x.realArray);
    free(x);
}

func nextSwap(x *swapInfo) int32 {
    var i, j, temp int32;
    for i=x.varN;i>1;i-- {
        if i > x.realArray[x.posArray[i].position + x.posArray[i].direction] {
            x.posArray[i].position = x.posArray[i].position + x.posArray[i].direction;
            temp = x.realArray[x.posArray[i].position];
            x.realArray[x.posArray[i].position] = i; 
            x.realArray[x.posArray[i].position - x.posArray[i].direction] = temp;
            x.posArray[temp].position = x.posArray[i].position - x.posArray[i].direction; 
            for j=x.varN;j>i;j-- {
                x.posArray[j].direction =     x.posArray[j].direction * -1;
            }
            x.positionToSwap1 = x.posArray[temp].position - 1;
            x.positionToSwap2 = x.posArray[i].position - 1;            
            return 1;
        }
        
    }
    return 0;    
}

func fillInSwapArray(pi *permInfo) {
    var counter int32 =pi.totalSwaps-1;
    var x *swapInfo = setSwapInfoPtr(pi.varN);
    for nextSwap(x)==1 {
        if x.positionToSwap1<x.positionToSwap2 {
            pi.swapArray[counter--]=x.positionToSwap1;
        } else {
            pi.swapArray[counter--]=x.positionToSwap2; }
    }
    
    freeSwapInfoPtr(x);    
}
func oneBitPosition(x int32, size int32) int32 {
    var i int32;
    for i=0;i<size;i++ {
        if (x>>i)&1 {
            return i; }
    }
    return -1;
}
func fillInFlipArray(pi *permInfo) {
    var i, temp, grayNumber int32;
    for i=1;i<=pi.totalFlips;i++ {
        grayNumber = i^(i>>1);
        pi.flipArray[pi.totalFlips-i]=oneBitPosition(temp^grayNumber, pi.varN);
        temp = grayNumber;        
    }
    
    
}
func factorial(n int32) int32 {
    if n == 1 || n == 0 {
        return 1;
    } else {
        return factorial(n - 1) * n;
    }
}
func setPermInfoPtr(var int32) *permInfo {
    var x *permInfo;
    x = (permInfo*) malloc(sizeof(permInfo));
    x.flipCtr=0;
    x.varN = var; 
    x.totalFlips=(1<<var)-1;
    x.swapCtr=0;
    x.totalSwaps=factorial(var)-1;
    x.flipArray = (int*) malloc(sizeof(int)*x.totalFlips);
    x.swapArray = (int*) malloc(sizeof(int)*x.totalSwaps);
    fillInSwapArray(x);
    fillInFlipArray(x);
    return x;
}

func freePermInfoPtr(x *permInfo) {
    free(x.flipArray);
    free(x.swapArray);
    free(x);
}
func minWord(a *word, b *word, minimal *word, nVars int32) {
    if memCompare(a, b, nVars) == -1 {
        Kit_TruthCopy_64bit( minimal, a, nVars );
    } else {
        Kit_TruthCopy_64bit( minimal, b, nVars ); }
}
func minWord3(a *word, b *word, minimal *word, nVars int32) { 
    if memCompare(a, b, nVars) <= 0 {
        if memCompare(a, minimal, nVars) < 0 { 
            Kit_TruthCopy_64bit( minimal, a, nVars ); 
        } else { 
            return ; }
    }    
    if memCompare(b, minimal, nVars) <= 0 {
        Kit_TruthCopy_64bit( minimal, b, nVars ); }
}
func minWord1(a *word, minimal *word, nVars int32) {
    if memCompare(a, minimal, nVars) <= 0 {
        Kit_TruthCopy_64bit( minimal, a, nVars ); }
}
func simpleMinimal(x *word, pAux *word, minimal *word, pi *permInfo, nVars int32) {
    var i, j int32;
    Kit_TruthCopy_64bit( pAux, x, nVars );
    Kit_TruthNot_64bit( x, nVars );
    
    minWord(x, pAux, minimal, nVars);
    
    for i=pi.totalSwaps-1;i>=0;i-- {
        Kit_TruthSwapAdjacentVars_64bit(x, nVars, pi.swapArray[i]);
        Kit_TruthSwapAdjacentVars_64bit(pAux, nVars, pi.swapArray[i]);
        minWord3(x, pAux, minimal, nVars);
    }
    for j=pi.totalFlips-1;j>=0;j-- {
        Kit_TruthSwapAdjacentVars_64bit(x, nVars, 0);
        Kit_TruthSwapAdjacentVars_64bit(pAux, nVars, 0);
        Kit_TruthChangePhase_64bit(x, nVars, pi.flipArray[j]);
        Kit_TruthChangePhase_64bit(pAux, nVars, pi.flipArray[j]);
        minWord3(x, pAux, minimal, nVars);
        for i=pi.totalSwaps-1;i>=0;i-- {
            Kit_TruthSwapAdjacentVars_64bit(x, nVars, pi.swapArray[i]);
            Kit_TruthSwapAdjacentVars_64bit(pAux, nVars, pi.swapArray[i]);
            minWord3(x, pAux, minimal, nVars);
        }
    } 
    Kit_TruthCopy_64bit( x, minimal, nVars );    
}

/**
 * pGroups: we assume that the variables are merged into adjacent groups,
 *          the size of each group is stored in pGroups
 * nGroups: the number of groups
 *
 * pis:     we compute all permInfos from 0 to nVars (incl.)
 */
func simpleMinimalGroups(x *word, pAux *word, minimal *word, pGroups *int32, nGroups int32, pis **permInfo, nVars int32, fFlipOutput int32, fFlipInput int32) {
    /* variables */
    var i, j, o, nn int32;
    var pi *permInfo;
    var a, c, m *int32;

    /* reorder groups and calculate group offsets */
    // inlining of macro ABC_ALLOC
    var offset *int32  = ((int *) malloc(sizeof(int) * (size_t)(nGroups)));
    o = 0;
    j = 0;

    for  i = 0; i < nGroups; ++i  {
        offset[j] = o;
        o += pGroups[j];
        ++j;
    }

    if fFlipOutput != 0 {
        /* keep regular and inverted version of x */
        Kit_TruthCopy_64bit( pAux, x, nVars );
        Kit_TruthNot_64bit( x, nVars );

        minWord(x, pAux, minimal, nVars);
    
    } else {
        Kit_TruthCopy_64bit( minimal, x, nVars );
    }

    /* iterate through all combinations of pGroups using mixed radix enumeration */
    nn = ( nGroups << 1 ) + 1;
    // inlining of macro ABC_ALLOC
    a = ((int *) malloc(sizeof(int) * (size_t)(nn)));
    // inlining of macro ABC_ALLOC
    c = ((int *) malloc(sizeof(int) * (size_t)(nn)));
    // inlining of macro ABC_ALLOC
    m = ((int *) malloc(sizeof(int) * (size_t)(nn)));

    /* fill a and m arrays */
    m[0] = 2;
    for  i = 1; i <= nGroups; ++i  { m[i]           = pis[pGroups[i - 1]].totalFlips + 1; }
    for  i = 1; i <= nGroups; ++i  { m[nGroups + i] = pis[pGroups[i - 1]].totalSwaps + 1; }
    for  i = 0; i < nn; ++i        { a[i] = c[i] = 0; }

    for {
        /* consider all flips */
        for  i = 1; i <= nGroups; ++i  {
            if !c[i] { continue; }
            if !(fFlipInput != 0) && pGroups[i - 1] == 1 { continue; }

            pi = pis[pGroups[i - 1]];
            if a[i] == 0 {
                j = 0;
            } else {
                j = pi.totalFlips - a[i];
            }

            Kit_TruthChangePhase_64bit(x, nVars, offset[i - 1] + pi.flipArray[j]);
            if fFlipOutput != 0 {
                Kit_TruthChangePhase_64bit(pAux, nVars, offset[i - 1] + pi.flipArray[j]);
                minWord3(x, pAux, minimal, nVars);
            
            } else {
                minWord1(x, minimal, nVars);
            }
        }

        /* consider all swaps */
        for  i = 1; i <= nGroups; ++i  {
            if !c[nGroups + i] { continue; }
            if pGroups[i - 1] == 1 { continue; }

            pi = pis[pGroups[i - 1]];
            if a[nGroups + i] == pi.totalSwaps {
                j = 0;
            
            } else {
                j = pi.swapArray[pi.totalSwaps - a[nGroups + i] - 1];
            }
            Kit_TruthSwapAdjacentVars_64bit(x, nVars, offset[i - 1] + j);
            if fFlipOutput != 0 {
                Kit_TruthSwapAdjacentVars_64bit(pAux, nVars, offset[i - 1] + j);
                minWord3(x, pAux, minimal, nVars);
            
            } else {
                minWord1(x, minimal, nVars);
            }
        }

        /* update a's */
        memset(c, 0, sizeof(int) * nn);
        j = nn - 1;
        for  a[j] == m[j] - 1  { c[j] = 1; a[j--] = 0; }

        /* done? */
        if j == 0 { break; }

        c[j] = 1;
        a[j]++;
    }
    // inlining of macro ABC_FREE
    ((offset) ? (free((char *) (offset)), (offset) = 0) : 0);
    // inlining of macro ABC_FREE
    ((a) ? (free((char *) (a)), (a) = 0) : 0);
    // inlining of macro ABC_FREE
    ((c) ? (free((char *) (c)), (c) = 0) : 0);
    // inlining of macro ABC_FREE
    ((m) ? (free((char *) (m)), (m) = 0) : 0);

    Kit_TruthCopy_64bit( x, minimal, nVars );
}

//ABC_NAMESPACE_IMPL_END
