package ssw

import (
	"fmt"
	"time"
 "math/bits"
 "github.com/glycerine/abcgo/aig/aig"
 "github.com/glycerine/abcgo/aig/saig"
 "github.com/glycerine/abcgo/misc/vec"
)

/**CFile****************************************************************

  FileName    [sswRarity.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Inductive prover with constraints.]

  Synopsis    [Rarity-driven refinement of equivalence classes.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - September 1, 2008.]

  Revision    [$Id: sswRarity.c,v 1.00 2008/09/01 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "sswInt.h"
//#include "aig/gia/giaAig.h"
//#include "base/main/main.h"
//#include "sat/bmc/bmc.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

   
type Ssw_RarMan_t struct {
 pPars *Ssw_RarPars_t
 nGroups, nWordsReg int32
 pAig *aig.Graph
 ppClasses *Ssw_Cla_t
 vInits *vec.Vec_Int_t
 pObjData, pPatData []uint64
 vUpdConst, vUpdClass *vec.Vec_Ptr_t
 pRarity []int32
 pPatCosts []float64
 vPatBests *vec.Vec_Int_t
 iFailPo, iFailPat int32
 vCexes *vec.Vec_Ptr_t
}

func Ssw_RarGetBinPat(p *Ssw_RarMan_t, bin, pat int32) int32 {
 if bin<0 || bin>=p.nGroups || pat<0 || pat>=1<<p.pPars.nBinSize { panic("ssw: rarity index out of bounds") }
 return p.pRarity[bin*(1<<p.pPars.nBinSize)+pat]
}
func Ssw_RarSetBinPat(p *Ssw_RarMan_t, bin, pat, value int32) {
 if bin<0 || bin>=p.nGroups || pat<0 || pat>=1<<p.pPars.nBinSize { panic("ssw: rarity index out of bounds") }
 p.pRarity[bin*(1<<p.pPars.nBinSize)+pat]=value
}
func Ssw_RarAddToBinPat(p *Ssw_RarMan_t, bin, pat int32) { Ssw_RarSetBinPat(p,bin,pat,Ssw_RarGetBinPat(p,bin,pat)+1) }
func Ssw_RarObjSim(p *Ssw_RarMan_t, id int32) []uint64 {
 if id<0 || id>=aig.Aig_ManObjNumMax(p.pAig) {panic("ssw: invalid rarity object")}
 start:=id*p.pPars.nWords;return p.pObjData[start:start+p.pPars.nWords]
}
func Ssw_RarPatSim(p *Ssw_RarMan_t, id int32) []uint64 {
 if id<0 || id>=64*p.pPars.nWords {panic("ssw: invalid rarity pattern")}
 start:=id*p.nWordsReg;return p.pPatData[start:start+p.nWordsReg]
}
func Ssw_RarManAssingRandomPis(p *Ssw_RarMan_t) {
 for i,obj:=range p.pAig.Inputs()[:saig.Saig_ManPiNum(p.pAig)] {
  sim:=Ssw_RarObjSim(p,int32(obj.NodeID()))
  for w:=range sim {sim[w]=aig.Aig_ManRandom64(0)}
  tail:=uint64(0xc);if i&1!=0 {tail=0xa};sim[0]=sim[0]<<4|tail
 }
}

/**Function*************************************************************

  Synopsis    [Derives the counter-example.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_RarDeriveCex(p *Ssw_RarMan_t, iFrame int32, iPo int32, iPatFinal int32, fVerbose int32) *Abc_Cex_t {
    var pCex *Abc_Cex_t;
    var pObj *Aig_Obj_t;
    var vTrace *Vec_Int_t;
    var pSim *word;
    var i, r, f, iBit, iPatThis int32;
    // compute the pattern sequence
    iPatThis = iPatFinal;
    vTrace = Vec_IntStartFull( iFrame / p.pPars.nFrames + 1 );
    Vec_IntWriteEntry( vTrace, iFrame / p.pPars.nFrames, iPatThis );
    for  r = iFrame / p.pPars.nFrames - 1; r >= 0; r--  {
        iPatThis = Vec_IntEntry( p.vPatBests, r * p.pPars.nWords + iPatThis / 64 );
        Vec_IntWriteEntry( vTrace, r, iPatThis );
    }
    // create counter-example
    pCex = Abc_CexAlloc( Aig_ManRegNum(p.pAig), Saig_ManPiNum(p.pAig), iFrame+1 );
    pCex.iFrame = iFrame;
    pCex.iPo = iPo;
    // insert the bits
    iBit = Aig_ManRegNum(p.pAig);
    for  f = 0; f <= iFrame; f++  {
        Ssw_RarManAssingRandomPis( p );
        iPatThis = Vec_IntEntry( vTrace, f / p.pPars.nFrames );
        // inlining of macro Saig_ManForEachPi
        // inlining of macro Vec_PtrForEachEntryStop
        for  i = 0; (i < Saig_ManPiNum(p.pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCis, i)), 1); i++  {
            pSim = Ssw_RarObjSim( p, Aig_ObjId(pObj) );
            if Abc_InfoHasBit( (unsigned *)pSim, iPatThis ) {
                Abc_InfoSetBit( pCex.pData, iBit ); }
            iBit++;
        }
    }
    Vec_IntFree( vTrace );
    assert(iBit == pCex.nBits, "iBit == pCex.nBits");
    // verify the counter example
    if !Saig_ManVerifyCex( p.pAig, pCex ) {
        Abc_Print( 1, "Ssw_RarDeriveCex(): Counter-example is invalid.\n" );
//        Abc_CexFree( pCex );
//        pCex = NULL;
    
    } else {
//      Abc_Print( 1, "Counter-example verification is successful.\n" );
    }
    return pCex;
}


/**Function*************************************************************

  Synopsis    [Testing the transposing code.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func TransposeTest() {
    var M, N [64]word;
    var i int32;
    var clk abctime;
    Aig_ManRandom64( 1 );
//    for ( i = 0; i < 64; i++ )
//        M[i] = Aig_ManRandom64( 0 );
    for  i = 0; i < 64; i++  {
        M[i] = i? (word)0 : ~(word)0;
    }
//    for ( i = 0; i < 64; i++ )
//        Extra_PrintBinary( stdout, (unsigned *)&M[i], 64 ), Abc_Print( 1, "\n" );

    clk = Abc_Clock();
    for  i = 0; i < 100001; i++  {
        transpose64Simple( M, N );
    }
    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );

    clk = Abc_Clock();
    for  i = 0; i < 100001; i++  {
        transpose64( M );
    }
    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );

    for  i = 0; i < 64; i++  {
        if M[i] != N[i] {
            Abc_Print( 1, "Mismatch\n" ); }
    }
/*
    Abc_Print( 1, "\n" );
    for ( i = 0; i < 64; i++ )
        Extra_PrintBinary( stdout, (unsigned *)&M[i], 64 ), Abc_Print( 1, "\n" );
    Abc_Print( 1, "\n" );
    for ( i = 0; i < 64; i++ )
        Extra_PrintBinary( stdout, (unsigned *)&N[i], 64 ), Abc_Print( 1, "\n" );
*/
}

/**Function*************************************************************

  Synopsis    [Transposing pObjData[ nRegs x nWords ] -> pPatData[ nWords x nRegs ].]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_RarTranspose(p *Ssw_RarMan_t) {
 for w:=int32(0);w<p.pPars.nWords;w++ {for r:=int32(0);r<p.nWordsReg;r++ {var m [64]uint64;for i:=int32(0);i<64;i++ {idx:=r*64+63-i;if idx<aig.Aig_ManRegNum(p.pAig){li:=saig.Saig_ManLi(p.pAig,idx);m[i]=Ssw_RarObjSim(p,int32(li.NodeID()))[w]}};transpose64(m[:]);for i:=int32(0);i<64;i++ {Ssw_RarPatSim(p,w*64+63-i)[r]=m[i]}}}
}





/**Function*************************************************************

  Synopsis    [Sets random inputs and specialied flop outputs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_RarManInitialize(p *Ssw_RarMan_t, init *vec.Vec_Int_t) {
 constant:=Ssw_RarObjSim(p,0);for i:=range constant {constant[i]=^uint64(0)};Ssw_RarManAssingRandomPis(p);regs:=aig.Aig_ManRegNum(p.pAig);if init!=nil && vec.Vec_IntSize(init)!=int(regs*p.pPars.nWords){panic("ssw: wrong rarity initialization size")};for i:=int32(0);i<regs;i++ {lo,li:=saig.Saig_ManLo(p.pAig,i),saig.Saig_ManLi(p.pAig,i);dst:=Ssw_RarObjSim(p,int32(lo.NodeID()));if init==nil {copy(dst,Ssw_RarObjSim(p,int32(li.NodeID())))} else {for w:=range dst {dst[w]=0;if vec.Vec_IntEntry(init,w*int(regs)+int(i))!=0 {dst[w]=^uint64(0)}}}}
}


/**Function*************************************************************

  Synopsis    [Returns 1 if simulation info is composed of all zeros.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_RarManPoIsConst0(data any, obj aig.Edge) int32 {p:=data.(*Ssw_RarMan_t);for _,v:=range Ssw_RarObjSim(p,int32(obj.NodeID())){if v!=0{return 0}};return 1}


/**Function*************************************************************

  Synopsis    [Returns 1 if simulation info is composed of all zeros.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_RarManObjIsConst(data any,obj aig.Edge) int32 {p:=data.(*Ssw_RarMan_t);var flip uint64;if aig.Aig_ObjPhase(p.pAig,obj){flip=^uint64(0)};for _,v:=range Ssw_RarObjSim(p,int32(obj.NodeID())){if v!=flip{return 0}};return 1}


/**Function*************************************************************

  Synopsis    [Returns 1 if simulation infos are equal.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_RarManObjsAreEqual(data any,a,b aig.Edge) int32 {p:=data.(*Ssw_RarMan_t);var flip uint64;if aig.Aig_ObjPhase(p.pAig,a)!=aig.Aig_ObjPhase(p.pAig,b){flip=^uint64(0)};x,y:=Ssw_RarObjSim(p,int32(a.NodeID())),Ssw_RarObjSim(p,int32(b.NodeID()));for i:=range x {if x[i]^y[i]^flip!=0{return 0}};return 1}


/**Function*************************************************************

  Synopsis    [Computes hash value of the node using its simulation info.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_Ssw_RarManObjHashWord_s_SPrimes [128]int32 = [128]int32{
        1009, 1049, 1093, 1151, 1201, 1249, 1297, 1361, 1427, 1459,
        1499, 1559, 1607, 1657, 1709, 1759, 1823, 1877, 1933, 1997,
        2039, 2089, 2141, 2213, 2269, 2311, 2371, 2411, 2467, 2543,
        2609, 2663, 2699, 2741, 2797, 2851, 2909, 2969, 3037, 3089,
        3169, 3221, 3299, 3331, 3389, 3461, 3517, 3557, 3613, 3671,
        3719, 3779, 3847, 3907, 3943, 4013, 4073, 4129, 4201, 4243,
        4289, 4363, 4441, 4493, 4549, 4621, 4663, 4729, 4793, 4871,
        4933, 4973, 5021, 5087, 5153, 5227, 5281, 5351, 5417, 5471,
        5519, 5573, 5651, 5693, 5749, 5821, 5861, 5923, 6011, 6073,
        6131, 6199, 6257, 6301, 6353, 6397, 6481, 6563, 6619, 6689,
        6737, 6803, 6863, 6917, 6977, 7027, 7109, 7187, 7237, 7309,
        7393, 7477, 7523, 7561, 7607, 7681, 7727, 7817, 7877, 7933,
        8011, 8039, 8059, 8081, 8093, 8111, 8123, 8147,
    };
func Ssw_RarManObjHashWord(pMan any, pObj *Aig_Obj_t) uint32 {
    var p *Ssw_RarMan_t  = (Ssw_RarMan_t *)pMan;
    
    var pSims *uint32;
    var uHash uint32;
    var i int32;
    uHash = 0;
    pSims = (unsigned *)Ssw_RarObjSim( p, pObj.Id );
    for  i = 0; i < 2 * p.pPars.nWords; i++  {
        uHash ^= pSims[i] * spStatic_Ssw_RarManObjHashWord_s_SPrimes[i & 0x7F];
    }
    return uHash;
}

/**Function*************************************************************

  Synopsis    [Returns 1 if simulation info is composed of all zeros.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_RarManObjWhichOne(p *Ssw_RarMan_t,obj aig.Edge) int32 {for w,v:=range Ssw_RarObjSim(p,int32(obj.NodeID())){if v!=0{return int32(w*64+bits.TrailingZeros64(v))}};return -1}


/**Function*************************************************************

  Synopsis    [Check if any of the POs becomes non-constant.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_RarManCheckNonConstOutputs(p *Ssw_RarMan_t, iFrame int32, Time abctime) int32 {
    var pObj *Aig_Obj_t;
    var i int32;
    p.iFailPo  = -1;
    p.iFailPat = -1;
    // inlining of macro Saig_ManForEachPo
    // inlining of macro Vec_PtrForEachEntryStop
    for  i = 0; (i < Saig_ManPoNum(p.pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCos, i)), 1); i++  {
        if p.pAig.nConstrs && i >= Saig_ManPoNum(p.pAig) - p.pAig.nConstrs {
            break; }
        if p.vCexes && Vec_PtrEntry(p.vCexes, i) {
            continue; }
        if Ssw_RarManPoIsConst0(p, pObj) != 0 {
            continue; }
        p.iFailPo  = i;
        p.iFailPat = Ssw_RarManObjWhichOne( p, pObj );
        if !p.pPars.fSolveAll {
            break; }
        // remember the one solved
        p.pPars.nSolved++;
        if p.vCexes == nil {
            p.vCexes = Vec_PtrStart( Saig_ManPoNum(p.pAig) ); }
        assert(Vec_PtrEntry(p.vCexes, i) == nil, "Vec_PtrEntry(p.vCexes, i) == nil");
        Vec_PtrWriteEntry( p.vCexes, i, (void *)(ABC_PTRINT_T)1 );
        if p.pPars.pFuncOnFail && p.pPars.pFuncOnFail(i, nil) {
            return 2; } // quitting due to callback
        // print final report
        if !p.pPars.fNotVerbose {
            var nOutDigits int32  = Abc_Base10Log( Saig_ManPoNum(p.pAig) );
            Abc_Print( 1, "Output %*d was asserted in frame %4d (solved %*d out of %*d outputs).  ", 
                nOutDigits, p.iFailPo, iFrame, 
                nOutDigits, p.pPars.nSolved, 
                nOutDigits, Saig_ManPoNum(p.pAig) );
            Abc_PrintTime( 1, "Time", Time );
        }
    }
    if p.iFailPo >= 0 { // found CEX
        return 1;
    } else {
        return 0; }
}

/**Function*************************************************************

  Synopsis    [Performs one round of simulation.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_RarManSimulate(p *Ssw_RarMan_t, vInit *Vec_Int_t, fUpdate int32, fFirst int32) {
    var pObj, pRepr *Aig_Obj_t;
    var pSim, pSim0, pSim1 *word;
    var Flip, Flip0, Flip1 word;
    var w, i int32;
    // initialize
    Ssw_RarManInitialize( p, vInit );
    Vec_PtrClear( p.vUpdConst );
    Vec_PtrClear( p.vUpdClass );
    Aig_ManIncrementTravId( p.pAig );
    // check comb inputs
    if fUpdate != 0 {
    // inlining of macro Aig_ManForEachCi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pAig.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCis, i)), 1); i++  {
        pRepr = Aig_ObjRepr(p.pAig, pObj);
        if pRepr == nil || Aig_ObjIsTravIdCurrent( p.pAig, pRepr ) {
            continue; }
        if Ssw_RarManObjsAreEqual( p, pObj, pRepr ) != 0 {
            continue; }
        // save for update
        if pRepr == Aig_ManConst1(p.pAig) {
            Vec_PtrPush( p.vUpdConst, pObj );
        } else {
            Vec_PtrPush( p.vUpdClass, pRepr );
            Aig_ObjSetTravIdCurrent( p.pAig, pRepr );
        }
    } }
    // simulate
    // inlining of macro Aig_ManForEachNode
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
        pSim  = Ssw_RarObjSim( p, Aig_ObjId(pObj) );
        pSim0 = Ssw_RarObjSim( p, Aig_ObjFaninId0(pObj) );
        pSim1 = Ssw_RarObjSim( p, Aig_ObjFaninId1(pObj) );
        if Aig_ObjFaninC0(pObj) {
            Flip0 = ~(word)0;
        } else {
            Flip0 = 0;
        }
        if Aig_ObjFaninC1(pObj) {
            Flip1 = ~(word)0;
        } else {
            Flip1 = 0;
        }
        for  w = 0; w < p.pPars.nWords; w++  {
            pSim[w] = (Flip0 ^ pSim0[w]) & (Flip1 ^ pSim1[w]);
        }


        if !(fUpdate != 0) {
            continue; }
        // check classes
        pRepr = Aig_ObjRepr(p.pAig, pObj);
        if pRepr == nil || Aig_ObjIsTravIdCurrent( p.pAig, pRepr ) {
            continue; }
        if Ssw_RarManObjsAreEqual( p, pObj, pRepr ) != 0 {
            continue; }
        // save for update
        if pRepr == Aig_ManConst1(p.pAig) {
            Vec_PtrPush( p.vUpdConst, pObj );
        } else {
            Vec_PtrPush( p.vUpdClass, pRepr );
            Aig_ObjSetTravIdCurrent( p.pAig, pRepr );
        }
    }
    }
    // transfer to POs
    // inlining of macro Aig_ManForEachCo
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pAig.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCos, i)), 1); i++  {
        pSim  = Ssw_RarObjSim( p, Aig_ObjId(pObj) );
        pSim0 = Ssw_RarObjSim( p, Aig_ObjFaninId0(pObj) );
        if Aig_ObjFaninC0(pObj) {
            Flip = ~(word)0;
        } else {
            Flip = 0;
        }
        for  w = 0; w < p.pPars.nWords; w++  {
            pSim[w] = Flip ^ pSim0[w];
        }
    }
    // refine classes
    if fUpdate != 0 {
        if fFirst != 0 {
            var vCands *Vec_Ptr_t  = Vec_PtrAlloc( 1000 );
            // inlining of macro Aig_ManForEachObj
            // inlining of macro Vec_PtrForEachEntry
            for  i = 0; (i < Vec_PtrSize(p.pAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vObjs, i)), 1); i++  { if (pObj) == nil {} else
                if Ssw_ObjIsConst1Cand( p.pAig, pObj ) {
                    Vec_PtrPush( vCands, pObj ); }
            }
            assert(Vec_PtrSize(vCands) == Ssw_ClassesCand1Num(p.ppClasses), "Vec_PtrSize(vCands) == Ssw_ClassesCand1Num(p.ppClasses)");
            Ssw_ClassesPrepareRehash( p.ppClasses, vCands, 0 );
            Vec_PtrFree( vCands );
        
        } else {
            Ssw_ClassesRefineConst1Group( p.ppClasses, p.vUpdConst, 1 );
            Ssw_ClassesRefineGroup( p.ppClasses, p.vUpdClass, 1 );
        }
    }
}


/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_RarManStart(g *aig.Graph,pars *Ssw_RarPars_t) *Ssw_RarMan_t {
 if pars.nWords<=0||pars.nFrames<=0||pars.nBinSize<=0||pars.nBinSize>16 {panic("ssw: invalid rarity parameters")};groups:=aig.Aig_ManRegNum(g)/pars.nBinSize;words:=Ssw_RarBitWordNum(aig.Aig_ManRegNum(g));return &Ssw_RarMan_t{pAig:g,pPars:pars,nGroups:groups,nWordsReg:words,pRarity:make([]int32,int(groups)*(1<<pars.nBinSize)),pPatCosts:make([]float64,int(pars.nWords)*64),pObjData:make([]uint64,int(aig.Aig_ManObjNumMax(g))*int(pars.nWords)),pPatData:make([]uint64,64*int(pars.nWords)*int(words)),vUpdConst:vec.Vec_PtrAlloc(100),vUpdClass:vec.Vec_PtrAlloc(100),vPatBests:vec.Vec_IntAlloc(100)}
}


/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_RarManStop(p *Ssw_RarMan_t){if p==nil{return};if p.vCexes!=nil{if p.pAig.SequentialCounterexamples()!=nil{panic("ssw: counterexample vector already exists")};p.pAig.SetSequentialCounterexamples(p.vCexes)};if p.ppClasses!=nil{Ssw_ClassesStop(p.ppClasses)};*p=Ssw_RarMan_t{}}


/**Function*************************************************************

  Synopsis    [Select best patterns.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_RarTransferPatterns(p *Ssw_RarMan_t, vInits *Vec_Int_t) {
//    Aig_Obj_t * pObj;
    unsigned char * pData;
    var pPattern *uint32;
    var i, k, Value int32;

    // more data from regs to pats
    Ssw_RarTranspose( p );

    // update counters
    for  k = 0; k < p.pPars.nWords * 64; k++  {
        pData = (unsigned char *)Ssw_RarPatSim( p, k );
        for  i = 0; i < p.nGroups; i++  {
            Ssw_RarAddToBinPat( p, i, pData[i] );
        }
    }

    // for each pattern
    for  k = 0; k < p.pPars.nWords * 64; k++  {
        pData = (unsigned char *)Ssw_RarPatSim( p, k );
        // find the cost of its values
        p.pPatCosts[k] = 0.0;
        for  i = 0; i < p.nGroups; i++  {
            Value = Ssw_RarGetBinPat( p, i, pData[i] );
            assert(Value > 0, "Value > 0");
            p.pPatCosts[k] += 1.0/(Value*Value);
        }
        // print the result
//Abc_Print( 1, "%3d : %9.6f\n", k, p->pPatCosts[k] );
    }

    // choose as many as there are words
    Vec_IntClear( vInits );
    for  i = 0; i < p.pPars.nWords; i++  {
        // select the best
        var iPatBest int32  = -1;
        var iCostBest float64  = -ABC_INFINITY;
        for  k = 0; k < p.pPars.nWords * 64; k++  {
            if iCostBest < p.pPatCosts[k] {
                iCostBest = p.pPatCosts[k];
                iPatBest  = k;
            }
        }
        // remove from costs
        assert(iPatBest >= 0, "iPatBest >= 0");
        p.pPatCosts[iPatBest] = -ABC_INFINITY;
        // set the flops
        pPattern = (unsigned *)Ssw_RarPatSim( p, iPatBest );
        for  k = 0; k < Aig_ManRegNum(p.pAig); k++  {
            Vec_IntPush( vInits, Abc_InfoHasBit(pPattern, k) );
        }
//Abc_Print( 1, "Best pattern %5d\n", iPatBest );
        Vec_IntPush( p.vPatBests, iPatBest );
    }
    assert(Vec_IntSize(vInits) == Aig_ManRegNum(p.pAig) * p.pPars.nWords, "Vec_IntSize(vInits) == Aig_ManRegNum(p.pAig) * p.pPars.nWords");
}


/**Function*************************************************************

  Synopsis    [Performs fraiging for one node.]

  Description [Returns the fraiged node.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_RarFindStartingState(pAig *Aig_Man_t, pCex *Abc_Cex_t) *Vec_Int_t {
    var vInit *Vec_Int_t;
    var pObj, pObjLi *Aig_Obj_t;
    var f, i, iBit int32;
    // assign register outputs
    // inlining of macro Saig_ManForEachLi
    for  i = 0; (i < Saig_ManRegNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCos, i+Saig_ManPoNum(pAig))), 1); i++  {
        pObj.fMarkB = Abc_InfoHasBit( pCex.pData, i );
    }
    // simulate the timeframes
    iBit = pCex.nRegs;
    for  f = 0; f <= pCex.iFrame; f++  {
        // set the PI simulation information
        Aig_ManConst1(pAig).fMarkB = 1;
        // inlining of macro Saig_ManForEachPi
        // inlining of macro Vec_PtrForEachEntryStop
        for  i = 0; (i < Saig_ManPiNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCis, i)), 1); i++  {
            pObj.fMarkB = Abc_InfoHasBit( pCex.pData, iBit++ );
        }
        // inlining of macro Saig_ManForEachLiLo
        for  i = 0; (i < Saig_ManRegNum(pAig)) && (((pObjLi) = Saig_ManLi(pAig, i)), 1)            && (((pObj)=Saig_ManLo(pAig, i)), 1); i++  {
            pObj.fMarkB = pObjLi.fMarkB;
        }
        // simulate internal nodes
        // inlining of macro Aig_ManForEachNode
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(pAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
            pObj.fMarkB = ( Aig_ObjFanin0(pObj).fMarkB ^ Aig_ObjFaninC0(pObj) )
                         & ( Aig_ObjFanin1(pObj).fMarkB ^ Aig_ObjFaninC1(pObj) ); }
        }
        // assign the COs
        // inlining of macro Aig_ManForEachCo
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(pAig.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCos, i)), 1); i++  {
            pObj.fMarkB = ( Aig_ObjFanin0(pObj).fMarkB ^ Aig_ObjFaninC0(pObj) );
        }
    }
    assert(iBit == pCex.nBits, "iBit == pCex.nBits");
    // check that the output failed as expected -- cannot check because it is not an SRM!
//    pObj = Aig_ManCo( pAig, pCex->iPo );
//    if ( pObj->fMarkB != 1 )
//        Abc_Print( 1, "The counter-example does not refine the output.\n" );
    // record the new pattern
    vInit = Vec_IntAlloc( Saig_ManRegNum(pAig) );
    // inlining of macro Saig_ManForEachLo
    for  i = 0; (i < Saig_ManRegNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCis, i+Saig_ManPiNum(pAig))), 1); i++  {
//Abc_Print( 1, "%d", pObj->fMarkB );
        Vec_IntPush( vInit, pObj.fMarkB );
    }
//Abc_Print( 1, "\n" );
    Aig_ManCleanMarkB( pAig );
    return vInit;
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_RarCheckTrivial(pAig *Aig_Man_t, fVerbose int32) int32 {
    var pObj *Aig_Obj_t;
    var i int32;
    // inlining of macro Saig_ManForEachPo
    // inlining of macro Vec_PtrForEachEntryStop
    for  i = 0; (i < Saig_ManPoNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCos, i)), 1); i++  {
        if pAig.nConstrs && i >= Saig_ManPoNum(pAig) - pAig.nConstrs {
            return 0; }
        if pObj.fPhase {
            pAig.pSeqModel = nil;
            pAig.pSeqModel = Abc_CexAlloc( Aig_ManRegNum(pAig), Saig_ManPiNum(pAig), 1 );
            pAig.pSeqModel.iPo = i;
            if fVerbose != 0 {
                Abc_Print( 1, "Output %d is trivally SAT in frame 0. \n", i ); }
            return 1;
        }
    }
    return 0;
}

/**Function*************************************************************

  Synopsis    [Perform sequential simulation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_RarSimulate(pAig *Aig_Man_t, pPars *Ssw_RarPars_t) int32 {
    var fTryBmc int32;
    var fMiter int32  = 1;
    var p *Ssw_RarMan_t;
    var r int32;
    var f int32 = -1;
    var clk abctime;
    var clkTotal abctime = Abc_Clock();
    var nTimeToStop abctime;
    if pPars.TimeOut {
        nTimeToStop = pPars.TimeOut * int64(time.Second) + Abc_Clock();
    } else {
        nTimeToStop = 0;
    }
    var timeLastSolved abctime;
    var nNumRestart int32;
    var nSavedSeed int32  = pPars.nRandSeed;
    var RetValue int32  = -1;
    var iFrameFail int32  = -1;
    assert(Aig_ManRegNum(pAig) > 0, "Aig_ManRegNum(pAig) > 0");
    assert(Aig_ManConstrNum(pAig) == 0, "Aig_ManConstrNum(pAig) == 0");
    pAig.pSeqModel = nil;
    // consider the case of empty AIG
//    if ( Aig_ManNodeNum(pAig) == 0 )
//        return -1;
    // check trivially SAT miters
//    if ( fMiter && Ssw_RarCheckTrivial( pAig, fVerbose ) )
//        return 0;
    if pPars.fVerbose {
        Abc_Print( 1, "Rarity simulation with %d words, %d frames, %d rounds, %d restart, %d seed, and %d sec timeout.\n",
            pPars.nWords, pPars.nFrames, pPars.nRounds, pPars.nRestart, pPars.nRandSeed, pPars.TimeOut ); }
    // reset random numbers
    Ssw_RarManPrepareRandom( nSavedSeed );

    // create manager
    p = Ssw_RarManStart( pAig, pPars );
    p.vInits = Vec_IntStart( Aig_ManRegNum(pAig) * pPars.nWords );

    // perform simulation rounds
    pPars.nSolved = 0;
    timeLastSolved = Abc_Clock();
    for  r = 0; !pPars.nRounds || (nNumRestart * pPars.nRestart + r < pPars.nRounds); r++  {
        clk = Abc_Clock();
        if fTryBmc != 0 {
            var pNewAig *Aig_Man_t  = Saig_ManDupWithPhase( pAig, p.vInits );
            Saig_BmcPerform( pNewAig, 0, 100, 2000, 3, 0, 0, 1 /*fVerbose*/, 0, &iFrameFail, 0, 0 );
//            if ( pNewAig->pSeqModel != NULL )
//                Abc_Print( 1, "BMC has found a counter-example in frame %d.\n", iFrameFail );
            Aig_ManStop( pNewAig );
        }
        // simulate
        for  f = 0; f < pPars.nFrames; f++  {
            Ssw_RarManSimulate( p, f ? nil : p.vInits, 0, 0 );
            if fMiter != 0 {
                var Status int32  = Ssw_RarManCheckNonConstOutputs(p, r * p.pPars.nFrames + f, Abc_Clock() - clkTotal);
                if Status == 2 {
                    Abc_Print( 1, "Quitting due to callback on fail.\n" );
                    goto finish;
                }
                if Status == 1 { // found CEX
                
                    RetValue = 0;
                    if !pPars.fSolveAll {
                        if pPars.fVerbose { Abc_Print( 1, "\n" ); }
        //                Abc_Print( 1, "Simulation asserted a PO in frame f: %d <= f < %d.\n", r * nFrames, (r+1) * nFrames );
                        Ssw_RarManPrepareRandom( nSavedSeed );
                        if pPars.fVerbose {
                            Abc_Print( 1, "Simulated %d frames for %d rounds with %d restarts.\n", pPars.nFrames, nNumRestart * pPars.nRestart + r, nNumRestart ); }
                        pAig.pSeqModel = Ssw_RarDeriveCex( p, r * p.pPars.nFrames + f, p.iFailPo, p.iFailPat, pPars.fVerbose );
                        // print final report
                        Abc_Print( 1, "Output %d of miter \"%s\" was asserted in frame %d.  ", pAig.pSeqModel.iPo, pAig.pName, pAig.pSeqModel.iFrame );
                        Abc_PrintTime( 1, "Time", Abc_Clock() - clkTotal );
                        goto finish;
                    }
                    timeLastSolved = Abc_Clock();
                }
                // else - did not find a counter example
            }
            // check timeout
            if pPars.TimeOut && Abc_Clock() > nTimeToStop {
                if !pPars.fSilent {
                if pPars.fVerbose && !pPars.fSolveAll { Abc_Print( 1, "\n" ); }
                Abc_Print( 1, "Simulated %d frames for %d rounds with %d restarts and solved %d outputs.  ", pPars.nFrames, nNumRestart * pPars.nRestart + r, nNumRestart, pPars.nSolved );
                Abc_Print( 1, "Reached timeout (%d sec).\n",  pPars.TimeOut );
                }
                goto finish;
            }
            if pPars.TimeOutGap && timeLastSolved && Abc_Clock() > timeLastSolved + pPars.TimeOutGap * int64(time.Second) {
                if !pPars.fSilent {
                if pPars.fVerbose && !pPars.fSolveAll { Abc_Print( 1, "\n" ); }
                Abc_Print( 1, "Simulated %d frames for %d rounds with %d restarts and solved %d outputs.  ", pPars.nFrames, nNumRestart * pPars.nRestart + r, nNumRestart, pPars.nSolved );
                Abc_Print( 1, "Reached gap timeout (%d sec).\n",  pPars.TimeOutGap );
                }
                goto finish;
            }
            // check if all outputs are solved by now
            if pPars.fSolveAll && p.vCexes && Vec_PtrCountZero(p.vCexes) == 0 {
                goto finish; }
        }
        // get initialization patterns
        if pPars.nRestart && r == pPars.nRestart {
            r = -1;
            nSavedSeed = (nSavedSeed + 1) % 1000;
            Ssw_RarManPrepareRandom( nSavedSeed );
            Vec_IntFill( p.vInits, Aig_ManRegNum(pAig) * pPars.nWords, 0 );
            nNumRestart++;
            Vec_IntClear( p.vPatBests );
            // clean rarity info
//            memset( p->pRarity, 0, sizeof(int) * (1 << nBinSize) * p->nGroups );
        
        } else {
            Ssw_RarTransferPatterns( p, p.vInits ); }
        // printout
        if pPars.fVerbose {
            if pPars.fSolveAll {
                Abc_Print( 1, "Starts =%6d   ",  nNumRestart );
                Abc_Print( 1, "Rounds =%6d   ",  nNumRestart * pPars.nRestart + ((r==-1)?0:r) );
                Abc_Print( 1, "Frames =%6d   ", (nNumRestart * pPars.nRestart + r) * pPars.nFrames );
                Abc_Print( 1, "CEX =%6d (%6.2f %%)   ", pPars.nSolved, 100.0*pPars.nSolved/Saig_ManPoNum(p.pAig) );
                Abc_PrintTime( 1, "Time", Abc_Clock() - clkTotal );
            
            } else {
                Abc_Print( 1, "." ); }
        }

    }
finish:
    if pPars.fSetLastState && p.vInits {
        assert(Vec_IntSize(p.vInits) % Aig_ManRegNum(pAig) == 0, "Vec_IntSize(p.vInits) % Aig_ManRegNum(pAig) == 0");
        Vec_IntShrink( p.vInits, Aig_ManRegNum(pAig) );
        pAig.pData = p.vInits;  p.vInits = nil;
    }
    if pPars.nSolved {
/*
        if ( !pPars->fSilent )
        {
        if ( pPars->fVerbose && !pPars->fSolveAll ) Abc_Print( 1, "\n" );
        Abc_Print( 1, "Simulation of %d frames for %d rounds with %d restarts asserted %d (out of %d) POs.    ", pPars->nFrames, nNumRestart * pPars->nRestart + r, nNumRestart, pPars->nSolved, Saig_ManPoNum(p->pAig) );
        Abc_PrintTime( 1, "Time", Abc_Clock() - clkTotal );
        }
*/
    
    } else if r == pPars.nRounds && f == pPars.nFrames {
        if !pPars.fSilent {
        if pPars.fVerbose { Abc_Print( 1, "\n" ); }
        Abc_Print( 1, "Simulation of %d frames for %d rounds with %d restarts did not assert POs.    ", pPars.nFrames, nNumRestart * pPars.nRestart + r, nNumRestart );
        Abc_PrintTime( 1, "Time", Abc_Clock() - clkTotal );
        }
    }
    // cleanup
    Ssw_RarManStop( p );
    return RetValue;
}


/**Function*************************************************************

  Synopsis    [Derive random flop permutation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_RarRandomPermFlop(nFlops int32, nUnused int32) *Vec_Int_t {
    var vPerm *Vec_Int_t;
    var i int32;
    var k int32;
    var pArray *int32;
    srand( 1 );
    fmt.Printf("Generating random permutation of %d flops.\n", nFlops);
    vPerm = Vec_IntStartNatural( nFlops );
    pArray = Vec_IntArray( vPerm );
    for  i = 0; i < nFlops; i++  {
        k = rand() % nFlops;
        // inlining of macro ABC_SWAP
        { var t int32  = pArray[i]; pArray[i] = pArray[k]; pArray[k] = t; };
    }
    fmt.Printf("Randomly adding %d unused flops.\n", nUnused);
    for  i = 0; i < nUnused; i++  {
        k = rand() % Vec_IntSize(vPerm);
        Vec_IntPush( vPerm, -1 );
        pArray = Vec_IntArray( vPerm );
        // inlining of macro ABC_SWAP
        { var t int32  = pArray[Vec_IntSize(vPerm)-1]; pArray[Vec_IntSize(vPerm)-1] = pArray[k]; pArray[k] = t; };
    }
//    Vec_IntPrint(vPerm);
    return vPerm;
}

/**Function*************************************************************

  Synopsis    [Perform sequential simulation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_RarSimulateGia(p *Gia_Man_t, pPars *Ssw_RarPars_t) int32 {
    var pAig *Aig_Man_t;
    var RetValue int32;
    if pPars.fUseFfGrouping {
        var vPerm *Vec_Int_t  = Ssw_RarRandomPermFlop( Gia_ManRegNum(p), 10 );
        var pTemp *Gia_Man_t  = Gia_ManDupPermFlopGap( p, vPerm );
        Vec_IntFree( vPerm );
        pAig = Gia_ManToAigSimple( pTemp );
        Gia_ManStop( pTemp );
    
    } else {
        pAig = Gia_ManToAigSimple( p ); }
    RetValue = Ssw_RarSimulate( pAig, pPars );
    // save counter-example
    Abc_CexFree( p.pCexSeq );
    p.pCexSeq = pAig.pSeqModel; pAig.pSeqModel = nil;
    Aig_ManStop( pAig );
    return RetValue;
}

/**Function*************************************************************

  Synopsis    [Filter equivalence classes of nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_RarSignalFilter(pAig *Aig_Man_t, pPars *Ssw_RarPars_t) int32 {
    var p *Ssw_RarMan_t;
    var r int32;
    var f int32 = -1;
    var i int32;
    var k int32;
    var clkTotal abctime  = Abc_Clock();
    var nTimeToStop abctime;
    if pPars.TimeOut {
        nTimeToStop = pPars.TimeOut * int64(time.Second) + Abc_Clock();
    } else {
        nTimeToStop = 0;
    }
    var nNumRestart int32;
    var nSavedSeed int32  = pPars.nRandSeed;
    var RetValue int32  = -1;
    assert(Aig_ManRegNum(pAig) > 0, "Aig_ManRegNum(pAig) > 0");
    assert(Aig_ManConstrNum(pAig) == 0, "Aig_ManConstrNum(pAig) == 0");
    // consider the case of empty AIG
    if Aig_ManNodeNum(pAig) == 0 {
        return -1; }
    // check trivially SAT miters
    if pPars.fMiter && (Ssw_RarCheckTrivial( pAig, 1 ) != 0) {
        return 0; }
    if pPars.fVerbose {
        Abc_Print( 1, "Rarity equiv filtering with %d words, %d frames, %d rounds, %d seed, and %d sec timeout.\n",
            pPars.nWords, pPars.nFrames, pPars.nRounds, pPars.nRandSeed, pPars.TimeOut ); }
    // reset random numbers
    Ssw_RarManPrepareRandom( nSavedSeed );

    // create manager
    p = Ssw_RarManStart( pAig, pPars );
    // compute starting state if needed
    assert(p.vInits == nil, "p.vInits == nil");
    if pPars.pCex {
        p.vInits = Ssw_RarFindStartingState( pAig, pPars.pCex );
        Abc_Print( 1, "Beginning simulation from the state derived using the counter-example.\n" );
    
    } else {
        p.vInits = Vec_IntStart( Aig_ManRegNum(pAig) ); }
    // duplicate the array
    for  i = 1; i < pPars.nWords; i++  {
        for  k = 0; k < Aig_ManRegNum(pAig); k++  {
            Vec_IntPush( p.vInits, Vec_IntEntry(p.vInits, k) );
        }
    }
    assert(Vec_IntSize(p.vInits) == Aig_ManRegNum(pAig) * pPars.nWords, "Vec_IntSize(p.vInits) == Aig_ManRegNum(pAig) * pPars.nWords");

    // create trivial equivalence classes with all nodes being candidates for constant 1
    if pAig.pReprs == nil {
        p.ppClasses = Ssw_ClassesPrepareSimple( pAig, pPars.fLatchOnly, 0 );
    } else {
        p.ppClasses = Ssw_ClassesPrepareFromReprs( pAig ); }
    Ssw_ClassesSetData( p.ppClasses, p, Ssw_RarManObjHashWord, Ssw_RarManObjIsConst, Ssw_RarManObjsAreEqual );
    // print the stats
    if pPars.fVerbose {
        Abc_Print( 1, "Initial  :  " );
        Ssw_ClassesPrint( p.ppClasses, 0 );
    }
    // refine classes using BMC
    for  r = 0; !pPars.nRounds || (nNumRestart * pPars.nRestart + r < pPars.nRounds); r++  {
        // start filtering equivalence classes
        if Ssw_ClassesCand1Num(p.ppClasses) == 0 && Ssw_ClassesClassNum(p.ppClasses) == 0 {
            Abc_Print( 1, "All equivalences are refined away.\n" );
            break;
        }
        // simulate
        for  f = 0; f < pPars.nFrames; f++  {
            Ssw_RarManSimulate( p, f ? nil : p.vInits, 1, !r && !f );
            if pPars.fMiter && (Ssw_RarManCheckNonConstOutputs(p, -1, 0) != 0) {
                if !pPars.fVerbose {
                    Abc_Print( 1, "%s", Abc_FrameIsBatchMode() ? "\n" : "\r" ); }
//                Abc_Print( 1, "Simulation asserted a PO in frame f: %d <= f < %d.\n", r * pPars->nFrames, (r+1) * pPars->nFrames );
                if pPars.fVerbose {
                    Abc_Print( 1, "Simulated %d frames for %d rounds with %d restarts.\n", pPars.nFrames, nNumRestart * pPars.nRestart + r, nNumRestart ); }
                Ssw_RarManPrepareRandom( nSavedSeed );
                Abc_CexFree( pAig.pSeqModel );
                pAig.pSeqModel = Ssw_RarDeriveCex( p, r * p.pPars.nFrames + f, p.iFailPo, p.iFailPat, 1 );
                // print final report
                Abc_Print( 1, "Output %d of miter \"%s\" was asserted in frame %d.  ", pAig.pSeqModel.iPo, pAig.pName, pAig.pSeqModel.iFrame );
                Abc_PrintTime( 1, "Time", Abc_Clock() - clkTotal );
                RetValue = 0;
                goto finish;
            }
            // check timeout
            if pPars.TimeOut && Abc_Clock() > nTimeToStop {
                if pPars.fVerbose { Abc_Print( 1, "\n" ); }
                Abc_Print( 1, "Simulated %d frames for %d rounds with %d restarts.  ", pPars.nFrames, nNumRestart * pPars.nRestart + r, nNumRestart );
                Abc_Print( 1, "Reached timeout (%d sec).\n",  pPars.TimeOut );
                goto finish;
            }
        }
        // get initialization patterns
        if pPars.pCex == nil && pPars.nRestart && r == pPars.nRestart {
            r = -1;
            nSavedSeed = (nSavedSeed + 1) % 1000;
            Ssw_RarManPrepareRandom( nSavedSeed );
            Vec_IntFill( p.vInits, Aig_ManRegNum(pAig) * pPars.nWords, 0 );
            nNumRestart++;
            Vec_IntClear( p.vPatBests );
            // clean rarity info
//            memset( p->pRarity, 0, sizeof(int) * (1 << nBinSize) * p->nGroups );
        
        } else {
            Ssw_RarTransferPatterns( p, p.vInits ); }
        // printout
        if pPars.fVerbose {
            Abc_Print( 1, "Round %3d:  ", r );
            Ssw_ClassesPrint( p.ppClasses, 0 );
        
        } else {
            Abc_Print( 1, "." );
        }
    }
finish:
    // report
    if r == pPars.nRounds && f == pPars.nFrames {
        if !pPars.fVerbose {
            Abc_Print( 1, "%s", Abc_FrameIsBatchMode() ? "\n" : "\r" ); }
        Abc_Print( 1, "Simulation of %d frames for %d rounds with %d restarts did not assert POs.    ", pPars.nFrames, nNumRestart * pPars.nRestart + r, nNumRestart );
        Abc_PrintTime( 1, "Time", Abc_Clock() - clkTotal );
    }
    // cleanup
    Ssw_RarManStop( p );
    return RetValue;
}

/**Function*************************************************************

  Synopsis    [Filter equivalence classes of nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_RarSignalFilterGia(p *Gia_Man_t, pPars *Ssw_RarPars_t) int32 {
    var pAig *Aig_Man_t;
    var RetValue int32;
    pAig = Gia_ManToAigSimple( p );
    if p.pReprs != nil {
        Gia_ManReprToAigRepr2( pAig, p );
        p.pReprs = nil;
        p.pNexts = nil;
    }
    RetValue = Ssw_RarSignalFilter( pAig, pPars );
    Gia_ManReprFromAigRepr( pAig, p );
    // save counter-example
    Abc_CexFree( p.pCexSeq );
    p.pCexSeq = pAig.pSeqModel; pAig.pSeqModel = nil;
    Aig_ManStop( pAig );
    return RetValue;
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END
