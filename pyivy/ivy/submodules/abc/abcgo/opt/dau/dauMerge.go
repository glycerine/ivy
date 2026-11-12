package dau

import "fmt"

/**CFile****************************************************************

  FileName    [dauMerge.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [DAG-aware unmapping.]

  Synopsis    [Enumeration of decompositions.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: dauMerge.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "dauInt.h"
//#include "misc/util/utilTruth.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////


/**Function*************************************************************

  Synopsis    [Substitution storage.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
   
type Dau_Sto_t struct {
    iVarUsed      int32;                          // counter of used variables
    pOutput [DAU_MAX_STR]byte;              // storage for reduced function
    pPosOutput    string;                        // place in the output
    pStore [DAU_MAX_VAR][DAU_MAX_STR]byte;  // storage for definitions
    pPosStore [DAU_MAX_VAR]string;            // place in the store
}

func Dau_DsdMergeStoreClean(pS *Dau_Sto_t, nShared int32) {
    var i int32;
    pS.iVarUsed  = nShared;
    for  i = 0; i < DAU_MAX_VAR; i++  {
        pS.pStore[i][0] = 0;
    }
}

func Dau_DsdMergeStoreCleanOutput(pS *Dau_Sto_t) {
    pS.pPosOutput = pS.pOutput;
}
func Dau_DsdMergeStoreAddToOutput(pS *Dau_Sto_t, pBeg string, pEnd string) {
    for  pBeg < pEnd  {
        *pS.pPosOutput++ = *pBeg++;
    }
}
func Dau_DsdMergeStoreAddToOutputChar(pS *Dau_Sto_t, c byte) {
    *pS.pPosOutput++ = c;
}

func Dau_DsdMergeStoreStartDef(pS *Dau_Sto_t, c byte) int32 {
    pS.pPosStore[pS.iVarUsed] = pS.pStore[pS.iVarUsed];
    if c { *pS.pPosStore[pS.iVarUsed]++ = c; }
    return pS.iVarUsed++;
}
func Dau_DsdMergeStoreAddToDef(pS *Dau_Sto_t, New int32, pBeg string, pEnd string) {
    for  pBeg < pEnd  {
        *pS.pPosStore[New]++ = *pBeg++;
    }
}
func Dau_DsdMergeStoreAddToDefChar(pS *Dau_Sto_t, New int32, c byte) {
    *pS.pPosStore[New]++ = c;
}
func Dau_DsdMergeStoreStopDef(pS *Dau_Sto_t, New int32, c byte) {
    if c { *pS.pPosStore[New]++ = c; }
    *pS.pPosStore[New]++ = 0;
}

func Dau_DsdMergeStoreCreateDef(pS *Dau_Sto_t, pBeg string, pEnd string) byte {
    var New int32  = Dau_DsdMergeStoreStartDef( pS, 0 );
    Dau_DsdMergeStoreAddToDef( pS, New, pBeg, pEnd );
    Dau_DsdMergeStoreStopDef( pS, New, 0 );
    return New;
}
func Dau_DsdMergeStorePrintDefs(pS *Dau_Sto_t) {
    var i int32;
    for  i = 0; i < DAU_MAX_VAR; i++  {
        if pS.pStore[i][0] {
            fmt.Printf("%c = %s\n", 'a' + i, pS.pStore[i]); }
    }
}

/**Function*************************************************************

  Synopsis    [Creates local copy.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dau_DsdMergeCopy(pDsd string, fCompl int32, pRes string) {
    if (fCompl != 0) && pDsd[0] == '!' {
        fCompl = 0;
pDsd++; }
    if Dau_DsdIsConst(pDsd) { // constant
        pRes[0] = (fCompl ? byte((int32(pDsd[0]) ^ 1)) : pDsd[0]), pRes[1] = 0;
    } else {
        pRes = fmt.Sprintf("%s%s", fCompl ? "!" : "", pDsd); }
}

/**Function*************************************************************

  Synopsis    [Replaces variables according to the mapping.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dau_DsdMergeReplace(pDsd string, pMatches *int32, pMap *int32) {
    var i int32;
    for  i = 0; pDsd[i]; i++  {
        // skip non-DSD block
        if pDsd[i] == '<' && pDsd[pMatches[i]+1] == '{' {
            i = pMatches[i] + 1; }
        if (pDsd[i] >= 'A' && pDsd[i] <= 'F') || (pDsd[i] >= '0' && pDsd[i] <= '9') {
            for  (pDsd[i] >= 'A' && pDsd[i] <= 'F') || (pDsd[i] >= '0' && pDsd[i] <= '9')  {
                i++;
            } }
        // detect variables
        if pDsd[i] >= 'a' && pDsd[i] <= 'z' {
            pDsd[i] = 'a' + pMap[ pDsd[i] - 'a' ]; }
    }
}
func Dau_DsdMergeMatches(pDsd string, pMatches *int32) {
    var pNested [DAU_MAX_VAR]int32;
    var i, nNested int32;
    for  i = 0; pDsd[i]; i++  {
        pMatches[i] = 0;
        if pDsd[i] == '(' || pDsd[i] == '[' || pDsd[i] == '<' || pDsd[i] == '{' {
            pNested[nNested++] = i;
        } else if pDsd[i] == ')' || pDsd[i] == ']' || pDsd[i] == '>' || pDsd[i] == '}' {
            pMatches[pNested[--nNested]] = i; }
        assert(nNested < DAU_MAX_VAR, "nNested < DAU_MAX_VAR");
    }
    assert(nNested == 0, "nNested == 0");
}
func Dau_DsdMergeVarPres(pDsd string, pMatches *int32, pPres *int32, Mask int32) {
    var i int32;
    for  i = 0; pDsd[i]; i++  {
        // skip non-DSD block
        if pDsd[i] == '<' && pDsd[pMatches[i]+1] == '{' {
            i = pMatches[i] + 1; }
        if (pDsd[i] >= 'A' && pDsd[i] <= 'F') || (pDsd[i] >= '0' && pDsd[i] <= '9') {
            for  (pDsd[i] >= 'A' && pDsd[i] <= 'F') || (pDsd[i] >= '0' && pDsd[i] <= '9')  {
                i++;
            } }
        // skip non-variables
        if !(pDsd[i] >= 'a' && pDsd[i] <= 'z') {
            continue; }
        // record the mask
        assert(pDsd[i]-'a' < DAU_MAX_VAR, "pDsd[i]-'a' < DAU_MAX_VAR");
        pPres[pDsd[i]-'a'] |= Mask;
    }
}
func Dau_DsdMergeCountShared(pPres *int32, Mask int32) int32 {
    var i, Counter int32;
    for  i = 0; i < DAU_MAX_VAR; i++  {
        Counter += (pPres[i] == Mask);
    }
    return Counter;
}
func Dau_DsdMergeFindShared(pDsd0 string, pDsd1 string, pMatches0 *int32, pMatches1 *int32, pVarPres *int32) int32 {
    memset( pVarPres, 0, sizeof(int)*DAU_MAX_VAR );
    Dau_DsdMergeVarPres( pDsd0, pMatches0, pVarPres, 1 );
    Dau_DsdMergeVarPres( pDsd1, pMatches1, pVarPres, 2 );
    return Dau_DsdMergeCountShared( pVarPres, 3 );
}
func Dau_DsdMergeCreateMaps(pVarPres *int32, nShared int32, pOld2New *int32, pNew2Old *int32) int32 {
    var i int32;
    var Counter int32;
    var Counter2 int32 = nShared;
    for  i = 0; i < DAU_MAX_VAR; i++  {
        if pVarPres[i] == 0 {
            continue; }
        if pVarPres[i] == 3 {
            pOld2New[i] = Counter;
            pNew2Old[Counter] = i;
            Counter++;
            continue;
        }
        assert(pVarPres[i] == 1 || pVarPres[i] == 2, "pVarPres[i] == 1 || pVarPres[i] == 2");
        pOld2New[i] = Counter2;
        pNew2Old[Counter2] = i;
        Counter2++;
    }
    return Counter2;
}
func Dau_DsdMergeInlineDefinitions(pDsd string, pMatches *int32, pS *Dau_Sto_t, pRes string, nShared int32) {
    var i int32;
    var pDef string;
    var pBegin string  = pRes;
    for  i = 0; pDsd[i]; i++  {
        // skip non-DSD block
        if pDsd[i] == '<' && pDsd[pMatches[i]+1] == '{' {
            assert(pDsd[pMatches[i]] == '>', "pDsd[pMatches[i]] == '>'");
            for  ; i <= pMatches[i]; i++  {
                *pRes++ = pDsd[i];
            }
        }
        if (pDsd[i] >= 'A' && pDsd[i] <= 'F') || (pDsd[i] >= '0' && pDsd[i] <= '9') {
            for  (pDsd[i] >= 'A' && pDsd[i] <= 'F') || (pDsd[i] >= '0' && pDsd[i] <= '9')  {
                *pRes++ = pDsd[i++];
            } }
        // detect variables
        if !(pDsd[i] >= 'a' && pDsd[i] <= 'z') || (pDsd[i] - 'a' < nShared) {
            *pRes++ = pDsd[i];
            continue;
        }
        // inline definition
        assert(pDsd[i]-'a' < DAU_MAX_STR, "pDsd[i]-'a' < DAU_MAX_STR");
        for  pDef = pS.pStore[pDsd[i]-'a']; *pDef; pDef++  {
            *pRes++ = *pDef;
        }
    }
    *pRes++ = 0;
    assert(pRes - pBegin < DAU_MAX_STR, "pRes - pBegin < DAU_MAX_STR");
}


/**Function*************************************************************

  Synopsis    [Computes independence status for each opening parenthesis.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dau_DsdMergePrintWithStatus(p string, pStatus *int32) {
    var i int32;
    fmt.Printf("%s\n", p);
    for  i = 0; p[i]; i++  {
        if !(p[i] == '(' || p[i] == '[' || p[i] == '<' || p[i] == '{' || (p[i] >= 'a' && p[i] <= 'z')) {
            fmt.Printf(" ");
        } else if pStatus[i] >= 0 {
            fmt.Printf("%d", pStatus[i]);
        } else {
            fmt.Printf("-"); }
    }
    fmt.Printf("\n");
}
func Dau_DsdMergeStatus_rec(pStr string, p *string, pMatches *int32, nShared int32, pStatus *int32) int32 {
    // none pure
    // 1 one pure
    // 2 two or more pure
    // 3 all pure
    if **p == '!' {
        pStatus[*p - pStr] = -1;
        (*p)++;
    }
    for  (**p >= 'A' && **p <= 'F') || (**p >= '0' && **p <= '9')  {
        pStatus[*p - pStr] = -1;
        (*p)++;
    }
    if **p == '<' {
        var q string  = pStr + pMatches[ *p - pStr ];
        if *(q+1) == '{' {
            var pTemp string  = *p;
            *p = q+1;
            for  ; pTemp < q+1; pTemp++  {
                pStatus[pTemp - pStr] = -1;
            }
        }
    }
    if **p >= 'a' && **p <= 'z' { // var
        if pStatus[*p - pStr] = (**p - 'a' < nShared) {
            return 0;
        } else {
            return 3;
        } }
    if **p == '(' || **p == '[' || **p == '<' || **p == '{' { // and/or/xor
    
        var Status, nPure, nTotal int32;
        var pOld string  = *p;
        var q string  = pStr + pMatches[ *p - pStr ];
        assert(*q == **p + 1 + (**p != '('), "*q == **p + 1 + (**p != '(')");
        for  (*p)++; *p < q; (*p)++  {
            Status = Dau_DsdMergeStatus_rec( pStr, p, pMatches, nShared, pStatus );
            nPure += (Status == 3);
            nTotal++;
        }
        assert(*p == q, "*p == q");
        assert(nTotal > 1, "nTotal > 1");
        if nPure == 0 {
            Status = 0;
        } else if nPure == 1 {
            Status = 1;
        } else if nPure < nTotal {
            Status = 2;
        } else if nPure == nTotal { 
            Status = 3;
        } else { assert(false, "0"); }
        return (pStatus[pOld - pStr] = Status);
    }
    assert(false, "0");
    return 0;
}
func Dau_DsdMergeStatus(pDsd string, pMatches *int32, nShared int32, pStatus *int32) int32 {
    return Dau_DsdMergeStatus_rec( pDsd, &pDsd, pMatches, nShared, pStatus );
}

/**Function*************************************************************

  Synopsis    [Extracts the formula.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dau_DsdMergeGetStatus(pBeg string, pStr string, pMatches *int32, pStatus *int32) int32 {
    if *pBeg == '!' {
        pBeg++; }
    for  (*pBeg >= 'A' && *pBeg <= 'F') || (*pBeg >= '0' && *pBeg <= '9')  {
        pBeg++;
    }
    if *pBeg == '<' {
        var q string  = pStr + pMatches[pBeg - pStr];
        if *(q+1) == '{' {
            pBeg = q+1; }
    }
    return pStatus[pBeg - pStr];
}
func Dau_DsdMergeSubstitute_rec(pS *Dau_Sto_t, pStr string, p *string, pMatches *int32, pStatus *int32, fWrite int32) {
//    assert( **p != '!' );

    if **p == '!' {
        if fWrite != 0 {
            Dau_DsdMergeStoreAddToOutputChar( pS, **p ); }
        (*p)++;
    }

    for  (**p >= 'A' && **p <= 'F') || (**p >= '0' && **p <= '9')  {
        if fWrite != 0 {
            Dau_DsdMergeStoreAddToOutputChar( pS, **p ); }
        (*p)++;
    }
    if **p == '<' {
        var q string  = pStr + pMatches[ *p - pStr ];
        if *(q+1) == '{' {
            var pTemp string  = *p;
            *p = q+1;
            if fWrite != 0 {
                for  ; pTemp < q+1; pTemp++  {
                    Dau_DsdMergeStoreAddToOutputChar( pS, *pTemp );
                } }
        }
    }
    if **p >= 'a' && **p <= 'z' { // var
    
        if fWrite != 0 {
            Dau_DsdMergeStoreAddToOutputChar( pS, **p ); }
        return;
    }
    if **p == '(' || **p == '[' || **p == '<' || **p == '{' { // and/or/xor
    
        var New int32;
        var StatusFan int32;
        var Status int32 = pStatus[*p - pStr];
        var pBeg string;
        var q string = pStr + pMatches[ *p - pStr ];
        assert(*q == **p + 1 + (**p != '('), "*q == **p + 1 + (**p != '(')");
        if !(fWrite != 0) {
             assert(Status == 3, "Status == 3");
             *p = q;
             return;
        }
        assert(Status != 3, "Status != 3");
        if Status == 0 { // none pure
        
            Dau_DsdMergeStoreAddToOutputChar( pS, **p );
            for  (*p)++; *p < q; (*p)++  {
                if **p == '!' {
                    Dau_DsdMergeStoreAddToOutputChar( pS, '!' );
                    (*p)++;
                }
                Dau_DsdMergeSubstitute_rec( pS, pStr, p, pMatches, pStatus, 1 );
            }
            Dau_DsdMergeStoreAddToOutputChar( pS, **p );
            assert(*p == q, "*p == q");
            return;
        }
        if Status == 1 || **p == '<' || **p == '{' { // 1 pure
        
            Dau_DsdMergeStoreAddToOutputChar( pS, **p );
            for  (*p)++; *p < q; (*p)++  {
                if **p == '!' {
                    Dau_DsdMergeStoreAddToOutputChar( pS, '!' );
                    (*p)++;
                }
                pBeg = *p;
                StatusFan = Dau_DsdMergeGetStatus( pBeg, pStr, pMatches, pStatus );
                Dau_DsdMergeSubstitute_rec( pS, pStr, p, pMatches, pStatus, StatusFan != 3 );
                if StatusFan == 3 {
                    var New int32  = Dau_DsdMergeStoreCreateDef( pS, pBeg, *p+1 );
                    Dau_DsdMergeStoreAddToOutputChar( pS, byte(('a' + New)) );
                }
            }
            Dau_DsdMergeStoreAddToOutputChar( pS, **p );
            assert(*p == q, "*p == q");
            return;
        }
        if Status == 2 {
            // add more than one defs
            Dau_DsdMergeStoreAddToOutputChar( pS, **p );
            New = Dau_DsdMergeStoreStartDef( pS, **p );
            for  (*p)++; *p < q; (*p)++  {
                pBeg = *p;
                StatusFan = Dau_DsdMergeGetStatus( pBeg, pStr, pMatches, pStatus );
                if **p == '!' {
                    if StatusFan != 3 {
                        Dau_DsdMergeStoreAddToOutputChar( pS, '!' );
                    } else {
                        Dau_DsdMergeStoreAddToDefChar( pS, New, '!' ); }
                    (*p)++;
                    pBeg++;
                }
                Dau_DsdMergeSubstitute_rec( pS, pStr, p, pMatches, pStatus, StatusFan != 3 );
                if StatusFan == 3 {
                    Dau_DsdMergeStoreAddToDef( pS, New, pBeg, *p+1 ); }
            }
            Dau_DsdMergeStoreStopDef( pS, New, *q );
            Dau_DsdMergeStoreAddToOutputChar( pS, byte(('a' + New)) );
            Dau_DsdMergeStoreAddToOutputChar( pS, **p );
            return;
        }
        assert(false, "0");
        return;
    }
    assert(false, "0");
}
func Dau_DsdMergeSubstitute(pS *Dau_Sto_t, pDsd string, pMatches *int32, pStatus *int32) {
/*
    int fCompl = 0;
    if ( pDsd[0] == '!' )
    {
        Dau_DsdMergeStoreAddToOutputChar( pS, '!' );
        pDsd++;
        fCompl = 1;
    }
*/
    Dau_DsdMergeSubstitute_rec( pS, pDsd, &pDsd, pMatches, pStatus, 1 );
    Dau_DsdMergeStoreAddToOutputChar( pS, 0 );
}

/**Function*************************************************************

  Synopsis    [Removes braces.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dau_DsdRemoveBraces_rec(pStr string, p *string, pMatches *int32) {
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
        return; }
    if **p == '(' || **p == '[' || **p == '<' || **p == '{' {
        var q string  = pStr + pMatches[ *p - pStr ];
        assert(*q == **p + 1 + (**p != '('), "*q == **p + 1 + (**p != '(')");
        for  (*p)++; *p < q; (*p)++  {
            var fCompl int32  = (**p == '!');
            var pBeg string;
            if fCompl != 0 {
                pBeg = *p + 1;
            } else {
                pBeg = *p;
            }
            Dau_DsdRemoveBraces_rec( pStr, p, pMatches );
            if (!(fCompl != 0) && *pBeg == '(' && *q == ')') || (*pBeg == '[' && *q == ']') {
                assert(**p == ')' || **p == ']', "**p == ')' || **p == ']'");
                *pBeg = **p = ' ';
            }
        }
        assert(*p == q, "*p == q");
        return;
    }
    assert(false, "0");
}
func Dau_DsdRemoveBraces(pDsd string, pMatches *int32) {
    var q string;
    var p string = pDsd;
    if pDsd[1] == 0 {
        return; }
    Dau_DsdRemoveBraces_rec( pDsd, &pDsd, pMatches );
    for  q = p; *p; p++  {
        if *p != ' ' {
            if *p == '!' && *(q-1) == '!' && p != q {
                q--;
                continue;
            }
            *q++ = *p;
        }
    }
    *q = 0;
}


var s_TimeComp [4]abctime ;

/**Function*************************************************************

  Synopsis    [Performs merging of two DSD formulas.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_Dau_DsdMerge_Counter int32;
var spStatic_Dau_DsdMerge_pRes string // was [2*DAU_MAX_STR+10]byte in C

func Dau_DsdMerge(pDsd0i string, pPerm0 *int32, pDsd1i string, pPerm1 *int32, fCompl0 int32, fCompl1 int32, nVars int32) string {
    var fVerbose int32;
    var fCheck int32;
    
    
    var pDsd0 [DAU_MAX_STR]byte;
    var pDsd1 [DAU_MAX_STR]byte;
    var pMatches0 [DAU_MAX_STR]int32;
    var pMatches1 [DAU_MAX_STR]int32;
    var pVarPres [DAU_MAX_VAR]int32;
    var pOld2New [DAU_MAX_VAR]int32;
    var pNew2Old [DAU_MAX_VAR]int32;
    var pStatus0 [DAU_MAX_STR]int32;
    var pStatus1 [DAU_MAX_STR]int32;
    var pMatches [DAU_MAX_STR]int32;
    var nVarsShared, nVarsTotal int32;
    var S Dau_Sto_t;
    var pS *Dau_Sto_t = &S;
    var pTruth *word;
    var pt *word = nil;
    var pt0 *word = nil;
    var pt1 *word = nil;
    var pParts [3][DAU_MAX_WORD]word;
    var Status int32;
    var clk abctime  = Abc_Clock();
    spStatic_Dau_DsdMerge_Counter++;
    // create local copies
    Dau_DsdMergeCopy( pDsd0i, fCompl0, pDsd0 );
    Dau_DsdMergeCopy( pDsd1i, fCompl1, pDsd1 );
if fVerbose != 0 {
fmt.Printf("\nAfter copying:\n"); }
if fVerbose != 0 {
fmt.Printf("%s\n", pDsd0); }
if fVerbose != 0 {
fmt.Printf("%s\n", pDsd1); }
    // handle constants
    if Dau_DsdIsConst(pDsd0) || Dau_DsdIsConst(pDsd1) {
        if Dau_DsdIsConst0(pDsd0) {
            strcpy( spStatic_Dau_DsdMerge_pRes, pDsd0 );
        } else if Dau_DsdIsConst1(pDsd0) {
            strcpy( spStatic_Dau_DsdMerge_pRes, pDsd1 );
        } else if Dau_DsdIsConst0(pDsd1) {
            strcpy( spStatic_Dau_DsdMerge_pRes, pDsd1 );
        } else if Dau_DsdIsConst1(pDsd1) {
            strcpy( spStatic_Dau_DsdMerge_pRes, pDsd0 );
        } else { assert(false, "0"); }
        return spStatic_Dau_DsdMerge_pRes;
    }

    // compute matches
    Dau_DsdMergeMatches( pDsd0, pMatches0 );
    Dau_DsdMergeMatches( pDsd1, pMatches1 );
    // implement permutation
    Dau_DsdMergeReplace( pDsd0, pMatches0, pPerm0 );
    Dau_DsdMergeReplace( pDsd1, pMatches1, pPerm1 );
if fVerbose != 0 {
fmt.Printf("After replacement:\n"); }
if fVerbose != 0 {
fmt.Printf("%s\n", pDsd0); }
if fVerbose != 0 {
fmt.Printf("%s\n", pDsd1); }


    if fCheck != 0 {
        pt0 = Dau_DsdToTruth( pDsd0, nVars );
        Abc_TtCopy( pParts[0], pt0, Abc_TtWordNum(nVars), 0 );
    }
    if fCheck != 0 {
        pt1 = Dau_DsdToTruth( pDsd1, nVars );
        Abc_TtCopy( pParts[1], pt1, Abc_TtWordNum(nVars), 0 );
        Abc_TtAnd( pParts[2], pParts[0], pParts[1], Abc_TtWordNum(nVars), 0 );
    }

    // find shared varaiables
    nVarsShared = Dau_DsdMergeFindShared(pDsd0, pDsd1, pMatches0, pMatches1, pVarPres);
    if nVarsShared == 0 {
        spStatic_Dau_DsdMerge_pRes = fmt.Sprintf("(%s%s)", pDsd0, pDsd1);
if fVerbose != 0 {
fmt.Printf("Disjoint:\n"); }
if fVerbose != 0 {
printf( "%s\n", spStatic_Dau_DsdMerge_pRes ); }

        Dau_DsdMergeMatches( spStatic_Dau_DsdMerge_pRes, pMatches );
        Dau_DsdRemoveBraces( spStatic_Dau_DsdMerge_pRes, pMatches );
        Dau_DsdNormalize( spStatic_Dau_DsdMerge_pRes );
if fVerbose != 0 {
fmt.Printf("Normalized:\n"); }
if fVerbose != 0 {
printf( "%s\n", spStatic_Dau_DsdMerge_pRes ); }

        s_TimeComp[0] += Abc_Clock() - clk;
        return spStatic_Dau_DsdMerge_pRes;
    }
s_TimeComp[3] += Abc_Clock() - clk;
    // create variable mapping
    nVarsTotal = Dau_DsdMergeCreateMaps( pVarPres, nVarsShared, pOld2New, pNew2Old );
    // perform variable replacement
    Dau_DsdMergeReplace( pDsd0, pMatches0, pOld2New );
    Dau_DsdMergeReplace( pDsd1, pMatches1, pOld2New );
    // find uniqueness status
    Dau_DsdMergeStatus( pDsd0, pMatches0, nVarsShared, pStatus0 );
    Dau_DsdMergeStatus( pDsd1, pMatches1, nVarsShared, pStatus1 );
if fVerbose != 0 {
fmt.Printf("Individual status:\n"); }
if fVerbose != 0 {
Dau_DsdMergePrintWithStatus( pDsd0, pStatus0 ); }
if fVerbose != 0 {
Dau_DsdMergePrintWithStatus( pDsd1, pStatus1 ); }
    // prepare storage
    Dau_DsdMergeStoreClean( pS, nVarsShared );
    // perform substitutions
    Dau_DsdMergeStoreCleanOutput( pS );
    Dau_DsdMergeSubstitute( pS, pDsd0, pMatches0, pStatus0 );
    strcpy( pDsd0, pS.pOutput );
if fVerbose != 0 {
fmt.Printf("Substitutions:\n"); }
if fVerbose != 0 {
fmt.Printf("%s\n", pDsd0); }

    // perform substitutions
    Dau_DsdMergeStoreCleanOutput( pS );
    Dau_DsdMergeSubstitute( pS, pDsd1, pMatches1, pStatus1 );
    strcpy( pDsd1, pS.pOutput );
if fVerbose != 0 {
fmt.Printf("%s\n", pDsd1); }
if fVerbose != 0 {
Dau_DsdMergeStorePrintDefs( pS ); }

    // create new function
//    assert( nVarsTotal <= 6 );
    sprintf( pS.pOutput, "(%s%s)", pDsd0, pDsd1 );
    pTruth = Dau_DsdToTruth( pS.pOutput, nVarsTotal );
    Status = Dau_DsdDecompose( pTruth, nVarsTotal, 0, 1, pS.pOutput );
//printf( "%d ", Status );
    if Status == -1 { // did not find 1-step DSD
        return nil; }
//    if ( Status > 6 ) // non-DSD part is too large
//        return NULL;
    if Dau_DsdIsConst(pS.pOutput) {
        strcpy( spStatic_Dau_DsdMerge_pRes, pS.pOutput );
        return spStatic_Dau_DsdMerge_pRes;
    }
if fVerbose != 0 {
fmt.Printf("Decomposition:\n"); }
if fVerbose != 0 {
fmt.Printf("%s\n", pS.pOutput); }
    // substitute definitions
    Dau_DsdMergeMatches( pS.pOutput, pMatches );
    Dau_DsdMergeInlineDefinitions( pS.pOutput, pMatches, pS, spStatic_Dau_DsdMerge_pRes, nVarsShared );
if fVerbose != 0 {
fmt.Printf("Inlining:\n"); }
if fVerbose != 0 {
printf( "%s\n", spStatic_Dau_DsdMerge_pRes ); }
    // perform variable replacement
    Dau_DsdMergeMatches( spStatic_Dau_DsdMerge_pRes, pMatches );
    Dau_DsdMergeReplace( spStatic_Dau_DsdMerge_pRes, pMatches, pNew2Old );
    Dau_DsdRemoveBraces( spStatic_Dau_DsdMerge_pRes, pMatches );
if fVerbose != 0 {
fmt.Printf("Replaced:\n"); }
if fVerbose != 0 {
printf( "%s\n", spStatic_Dau_DsdMerge_pRes ); }
    Dau_DsdNormalize( spStatic_Dau_DsdMerge_pRes );
if fVerbose != 0 {
fmt.Printf("Normalized:\n"); }
if fVerbose != 0 {
printf( "%s\n", spStatic_Dau_DsdMerge_pRes ); }

    if fCheck != 0 {
        pt = Dau_DsdToTruth( spStatic_Dau_DsdMerge_pRes, nVars );
        if !Abc_TtEqual( pParts[2], pt, Abc_TtWordNum(nVars) ) {
            fmt.Printf("Dau_DsdMerge(): Verification failed!\n"); }
    }

    if Status == 0 {
        s_TimeComp[1] += Abc_Clock() - clk;
    } else {
        s_TimeComp[2] += Abc_Clock() - clk; }
    return spStatic_Dau_DsdMerge_pRes;
}


func Dau_DsdTest66() {
    var Perm0 [DAU_MAX_VAR]int32 = [DAU_MAX_VAR]int32{ 0, 1, 2, 3, 4, 5 };
//    int pMatches[DAU_MAX_STR];
//    int pStatus[DAU_MAX_STR];

//    char * pStr = "(!(!a<bcd>)!(!fe))";
//    char * pStr = "([acb]<!edf>)";
//    char * pStr = "!(f!(b!c!(d[ea])))";
    var pStr string  = "[!(a[be])!(c!df)]";
//    char * pStr = "<(e<bca>)fd>";
//    char * pStr = "[d8001{abef}c]";
//    char * pStr1 = "(abc)";
//    char * pStr2 = "[adf]";
//    char * pStr1 = "(!abce)";
//    char * pStr2 = "[adf!b]";
//    char * pStr1 = "(!abc)";
//    char * pStr2 = "[ab]";
//    char * pStr1 = "[d81{abe}c]";
//    char * pStr1 = "[d<a(bc)(!b!c)>{abe}c]";
//    char * pStr1 = "[d81{abe}c]";
//    char * pStr1 = "[d(a[be])c]";
//    char * pStr2 = "(df)";
//    char * pStr1 = "(abf)";
//    char * pStr2 = "(a[(bc)(fde)])";
//    char * pStr1 = "8001{abc[ef]}";
//    char * pStr2 = "(abe)";

    var pStr1 string  = "(!(ab)de)";
    var pStr2 string  = "(!(ac)f)";

    var pRes string;
    var t word  = Dau_Dsd6ToTruth( pStr );
    return;

//    pStr2 = Dau_DsdDecompose( &t, 6, 0, NULL );
//    Dau_DsdNormalize( pStr2 );

//    Dau_DsdMergeMatches( pStr, pMatches );
//    Dau_DsdMergeStatus( pStr, pMatches, 2, pStatus );
//    Dau_DsdMergePrintWithStatus( pStr, pStatus );

    pRes = Dau_DsdMerge( pStr1, Perm0, pStr2, Perm0, 0, 0, 6 );

    t = 0; 
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

