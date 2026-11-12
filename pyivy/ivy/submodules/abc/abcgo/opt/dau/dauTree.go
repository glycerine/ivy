package dau

import "fmt"

import "os"

/**CFile****************************************************************

  FileName    [dauTree.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [DAG-aware unmapping.]

  Synopsis    [Canonical DSD package.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: dauTree.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "dauInt.h"
//#include "misc/mem/mem.h"
//#include "misc/util/utilTruth.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

   
type Dss_Fun_t struct {
    iDsd uint32; /* C bitfield: 26 bits; review packing and truncation */  // DSD literal
    nFans uint32; /* C bitfield: 6 bits; review packing and truncation */  // fanin count
    unsigned char  pFans[0];       // fanins
}

   
type Dss_Ent_t struct {
    pFunc     *Dss_Fun_t;
    pNext     *Dss_Ent_t;
    iDsd0 uint32; /* C bitfield: 27 bits; review packing and truncation */  // dsd entry
    nWords uint32; /* C bitfield: 5 bits; review packing and truncation */  // total word count (struct + shared)
    iDsd1 uint32; /* C bitfield: 27 bits; review packing and truncation */  // dsd entry
    nShared uint32; /* C bitfield: 5 bits; review packing and truncation */  // shared count
    unsigned char  pShared[0];     // shared literals
}

   
type Dss_Obj_t struct {
    Id       uint32;             // node ID
    Type uint32; /* C bitfield: 3 bits; review packing and truncation */  // node type
    nSupp uint32; /* C bitfield: 8 bits; review packing and truncation */  // variable
    iVar uint32; /* C bitfield: 8 bits; review packing and truncation */  // variable
    nWords uint32; /* C bitfield: 6 bits; review packing and truncation */  // variable
    fMark0 uint32; /* C bitfield: 1 bits; review packing and truncation */  // user mark
    fMark1 uint32; /* C bitfield: 1 bits; review packing and truncation */  // user mark
    nFans uint32; /* C bitfield: 5 bits; review packing and truncation */  // fanin count
    pFans [0]uint32;       // fanins
}

   
type Dss_Ntk_t struct {
    nVars            int32;          // the number of variables
    nMem            int32;           // memory used
    nMemAlloc            int32;      // memory allocated
    pMem          *word;           // memory array
    pRoot     *Dss_Obj_t;          // root node
    vObjs     *Vec_Ptr_t;          // internal nodes
}

type Dss_Man_t_ struct {
    nVars            int32;          // variable number
    nNonDecLimit            int32;   // limit on non-dec size
    nBins            int32;          // table size
    pBins      *uint32;          // hash table
    pMem    *Mem_Flex_t;           // memory for nodes
    vObjs     *Vec_Ptr_t;          // objects
    vNexts     *Vec_Int_t;         // next pointers
    vLeaves     *Vec_Int_t;        // temp
    vCopies     *Vec_Int_t;        // temp
    pTtElems         **word;       // elementary TTs
    pCache    **Dss_Ent_t;         // decomposition cache
    nCache            int32;         // size of decomposition cache
    pMemEnts    *Mem_Flex_t;       // memory for cache entries
    nCacheHits [2]int32;
    nCacheMisses [2]int32;
    nCacheEntries [2]int32;
    timeBeg        abctime;
    timeDec        abctime;
    timeLook        abctime;
    timeEnd        abctime;
}

func Dss_Regular(p *Dss_Obj_t) *Dss_Obj_t { return (Dss_Obj_t *)((ABC_PTRUINT_T)(p) & ~01);                                    }
func Dss_Not(p *Dss_Obj_t) *Dss_Obj_t { return (Dss_Obj_t *)((ABC_PTRUINT_T)(p) ^  01);                                    }
func Dss_NotCond(p *Dss_Obj_t, c int32) *Dss_Obj_t { return (Dss_Obj_t *)((ABC_PTRUINT_T)(p) ^ (c));                                    }
func Dss_IsComplement(p *Dss_Obj_t) int32 { return int32(((ABC_PTRUINT_T)(p) & 01));                                             }

func Dss_EntWordNum(p *Dss_Ent_t) int32 { return sizeof(Dss_Ent_t) / 8 + p.nShared / 4 + ((p.nShared & 3) > 0);            }
func Dss_FunWordNum(p *Dss_Fun_t) int32 { assert(p.nFans >= 2, "p.nFans >= 2"); return (p.nFans + 4) / 8 + (((p.nFans + 4) & 7) > 0);     }
func Dss_ObjWordNum(nFans int32) int32 { return sizeof(Dss_Obj_t) / 8 + nFans / 2 + ((nFans & 1) > 0);                      }
func Dss_ObjTruth(pObj *Dss_Obj_t) *word { return (word *)pObj + pObj.nWords;                                                }

func Dss_ObjClean(pObj *Dss_Obj_t) { *pObj = Dss_Obj_t{};                                              }
func Dss_ObjId(pObj *Dss_Obj_t) int32 { return pObj.Id;                                                                   }
func Dss_ObjType(pObj *Dss_Obj_t) int32 { return pObj.Type;                                                                 }
func Dss_ObjSuppSize(pObj *Dss_Obj_t) int32 { return pObj.nSupp;                                                                }
func Dss_ObjFaninNum(pObj *Dss_Obj_t) int32 { return pObj.nFans;                                                                }
func Dss_ObjFaninC(pObj *Dss_Obj_t, i int32) int32 { assert(i < int32(pObj.nFans), "i < int32(pObj.nFans)"); return Abc_LitIsCompl(pObj.pFans[i]);               }

func Dss_VecObj(p *Vec_Ptr_t, Id int32) *Dss_Obj_t { return (Dss_Obj_t *)Vec_PtrEntry(p, Id);                                           }
func Dss_VecConst0(p *Vec_Ptr_t) *Dss_Obj_t { return Dss_VecObj( p, 0 );                                                         }
func Dss_VecVar(p *Vec_Ptr_t, v int32) *Dss_Obj_t { return Dss_VecObj( p, v+1 );                                                       }
func Dss_VecLitSuppSize(p *Vec_Ptr_t, iLit int32) int32 { return Dss_VecObj( p, Abc_Lit2Var(iLit) ).nSupp;                                  }

func Dss_Obj2Lit(pObj *Dss_Obj_t) int32 { return Abc_Var2Lit(Dss_Regular(pObj).Id, Dss_IsComplement(pObj));                 }
func Dss_Lit2Obj(p *Vec_Ptr_t, iLit int32) *Dss_Obj_t { return Dss_NotCond(Dss_VecObj(p, Abc_Lit2Var(iLit)), Abc_LitIsCompl(iLit));        }
func Dss_ObjFanin(p *Vec_Ptr_t, pObj *Dss_Obj_t, i int32) *Dss_Obj_t { assert(i < int32(pObj.nFans), "i < int32(pObj.nFans)"); return Dss_VecObj(p, Abc_Lit2Var(pObj.pFans[i]));   }
func Dss_ObjChild(p *Vec_Ptr_t, pObj *Dss_Obj_t, i int32) *Dss_Obj_t { assert(i < int32(pObj.nFans), "i < int32(pObj.nFans)"); return Dss_Lit2Obj(p, pObj.pFans[i]);               }

// Original macro:
// #define Dss_VecForEachObj( vVec, pObj, i )                \
//     Vec_PtrForEachEntry( Dss_Obj_t *, vVec, pObj, i )

// Original macro:
// #define Dss_VecForEachObjVec( vLits, vVec, pObj, i )      \
//     for ( i = 0; (i < Vec_IntSize(vLits)) && ((pObj) = Dss_Lit2Obj(vVec, Vec_IntEntry(vLits,i))); i++ )

// Original macro:
// #define Dss_VecForEachNode( vVec, pObj, i )               \
//     Vec_PtrForEachEntry( Dss_Obj_t *, vVec, pObj, i )     \
//         if ( pObj->Type == DAU_DSD_CONST0 || pObj->Type == DAU_DSD_VAR ) {} else

// Original macro:
// #define Dss_ObjForEachFanin( vVec, pObj, pFanin, i )      \
//     for ( i = 0; (i < Dss_ObjFaninNum(pObj)) && ((pFanin) = Dss_ObjFanin(vVec, pObj, i)); i++ )

// Original macro:
// #define Dss_ObjForEachChild( vVec, pObj, pFanin, i )      \
//     for ( i = 0; (i < Dss_ObjFaninNum(pObj)) && ((pFanin) = Dss_ObjChild(vVec, pObj, i)); i++ )


func Dss_WordCountOnes(uWord uint32) int32 {
    uWord = (uWord & 0x55555555) + ((uWord>>1) & 0x55555555);
    uWord = (uWord & 0x33333333) + ((uWord>>2) & 0x33333333);
    uWord = (uWord & 0x0F0F0F0F) + ((uWord>>4) & 0x0F0F0F0F);
    uWord = (uWord & 0x00FF00FF) + ((uWord>>8) & 0x00FF00FF);
    return  (uWord & 0x0000FFFF) + (uWord>>16);
}

func Dss_Lit2Lit(pMapLit *int32, Lit int32) int32 { return Abc_Var2Lit( Abc_Lit2Var(pMapLit[Abc_Lit2Var(Lit)]), Abc_LitIsCompl(Lit) ^ Abc_LitIsCompl(pMapLit[Abc_Lit2Var(Lit)]) );   }

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

     

                                                                        

                                              

                
               
                

                

                                                                        
                                           
                                                                              
 
                       
                     
                                      
                                    
     
                                                     
                                 
     
                                                                 
     
                                  
                          
                          
                                            
                                       
                                                        
         
                                                                 
                                                           
                               
                                     
                                                   
         
                                           
                                          
                                                       
         
                                   
                                             
                                             
                                              
                    
                          
                                                    
                                                 

                                                   
                                                           
             
                                                           
                                                                   

                                                    
                                                 
                                                                  
                                                  
                        
                              
                                                        
                                                     

                                                       
                                                               
                 
                                                                                
                                                                                    
                                                                                    
                 
             
         
                                              
         
                          
                                             
                                          
                                             
         
                                                    
         
                          
                                     
                                          
                                             

                          
                                             
                                          
                                             
         
     
                                                                        
     
                                                        
                                                              
     
                                          
                                     
                 
 
                                    
 
                       
                           
                                                       
                     
                 
                       
                                 
                                                            
                             
                                           
     
                                     
                                  
  
                                  
                                   
                                        
                
                              
                       
  
     
                             
                                           
                                                     
     
                            
                     
                                                            
         
                                  
                     
         
     
                          
             
 
                   
 
  
                                                                              
                                             

                                                   
                                                       
                                                       
                                         
                        
  
 

                    

                                                                   
 
                                                                              
                                             

                               
                               
                          

                    
                     
                     
                        

                         
     
                                                  
                                                   
                                                       

                        
                   
                    
               
     

                                                         
                                                           
                      
               

                                           

                          
                         
                                     
                                    

                        

                         
                        
     
                               

                        
                                          
                        
         
                             
                            
                                          
          


                          
                          
                          

                   
                       
                                 

         
                                                                                                                          
                                                                                                       
                                                                                                                      
         

     
                                  
 

      

 
/**Function*************************************************************

  Synopsis    [Elementary truth tables.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_Dss_ManTtElems_TtElems [DAU_MAX_VAR+1][DAU_MAX_WORD]word;
var spStatic_Dss_ManTtElems_pTtElems [DAU_MAX_VAR+1]*word = [DAU_MAX_VAR+1]*word{nil};
func Dss_ManTtElems() **word {
    
    if spStatic_Dss_ManTtElems_pTtElems[0] == nil {
        var v int32;
        for  v = 0; v <= DAU_MAX_VAR; v++  {
            spStatic_Dss_ManTtElems_pTtElems[v] = spStatic_Dss_ManTtElems_TtElems[v];
        }
        Abc_TtElemInit( spStatic_Dss_ManTtElems_pTtElems, DAU_MAX_VAR );
    }
    return spStatic_Dss_ManTtElems_pTtElems;
}

/**Function*************************************************************

  Synopsis    [Creating DSD network.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dss_ObjAllocNtk(p *Dss_Ntk_t, Type int32, nFans int32, nTruthVars int32) *Dss_Obj_t {
    var pObj *Dss_Obj_t;
    pObj = (Dss_Obj_t *)(p.pMem + p.nMem);
    Dss_ObjClean( pObj );
    pObj.nFans  = nFans;
    pObj.nWords = Dss_ObjWordNum( nFans );
    pObj.Type   = Type;
    pObj.Id     = Vec_PtrSize( p.vObjs );
    pObj.iVar   = 31;
    Vec_PtrPush( p.vObjs, pObj );
    p.nMem += pObj.nWords + (nTruthVars ? Abc_TtWordNum(nTruthVars) : 0);
    assert(p.nMem < p.nMemAlloc, "p.nMem < p.nMemAlloc");
    return pObj;
}
func Dss_ObjCreateNtk(p *Dss_Ntk_t, Type int32, vFaninLits *Vec_Int_t) *Dss_Obj_t {
    var pObj *Dss_Obj_t;
    var i, Entry int32;
    pObj = Dss_ObjAllocNtk( p, Type, Vec_IntSize(vFaninLits), Type == DAU_DSD_PRIME ? Vec_IntSize(vFaninLits) : 0 );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vFaninLits)) && (((Entry) = Vec_IntEntry(vFaninLits, i)), 1); i++  {
        pObj.pFans[i] = Entry;
        pObj.nSupp += Dss_VecLitSuppSize(p.vObjs, Entry);
    }
    assert(i == int32(pObj.nFans), "i == int32(pObj.nFans)");
    return pObj;
}
func Dss_NtkAlloc(nVars int32) *Dss_Ntk_t {
    
    var pObj *Dss_Obj_t;
    var i int32;
    // inlining of macro ABC_CALLOC
    p := &Dss_Ntk_t{};
    p.nVars     = nVars;
    p.nMemAlloc = DAU_MAX_STR;
    // inlining of macro ABC_ALLOC
    p.pMem      = ((word *) malloc(sizeof(word) * (size_t)(p.nMemAlloc)));
    p.vObjs     = Vec_PtrAlloc( 100 );
    Dss_ObjAllocNtk( p, DAU_DSD_CONST0, 0, 0 );
    for  i = 0; i < nVars; i++  {
        pObj = Dss_ObjAllocNtk( p, DAU_DSD_VAR, 0, 0 );
        pObj.iVar = i;
        pObj.nSupp = 1;
    }
    return p;
}
func Dss_NtkFree(p *Dss_Ntk_t) {
    Vec_PtrFree( p.vObjs );
    // inlining of macro ABC_FREE
    ((p.pMem) ? (free((char *) (p.pMem)), (p.pMem) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}
func Dss_NtkPrint_rec(p *Dss_Ntk_t, pObj *Dss_Obj_t) {
    var OpenType [7]byte = [7]byte{0, 0, 0, '(', '[', '<', '{'};
    var CloseType [7]byte = [7]byte{0, 0, 0, ')', ']', '>', '}'};
    var pFanin *Dss_Obj_t;
    var i int32;
    assert(!(Dss_IsComplement(pObj) != 0), "!Dss_IsComplement(pObj)");
    if pObj.Type == DAU_DSD_VAR { fmt.Printf("%c", 'a' + pObj.iVar); return; }
    if pObj.Type == DAU_DSD_PRIME {
        Abc_TtPrintHexRev( os.Stdout, Dss_ObjTruth(pObj), pObj.nFans ); }
    fmt.Printf("%c", OpenType[pObj.Type]);
    // inlining of macro Dss_ObjForEachFanin
    for  i = 0; (i < Dss_ObjFaninNum(pObj)) && ((pFanin) = Dss_ObjFanin(p.vObjs, pObj, i)); i++  {
        fmt.Printf("%s", Dss_ObjFaninC(pObj, i) ? "!":"");
        Dss_NtkPrint_rec( p, pFanin );
    }
    fmt.Printf("%c", CloseType[pObj.Type]);
}
func Dss_NtkPrint(p *Dss_Ntk_t) {
    if Dss_Regular(p.pRoot).Type == DAU_DSD_CONST0 {
        fmt.Printf("%d", Dss_IsComplement(p.pRoot));
    } else {
        fmt.Printf("%s", Dss_IsComplement(p.pRoot) ? "!":"");        
        if Dss_Regular(p.pRoot).Type == DAU_DSD_VAR {
            fmt.Printf("%c", 'a' + Dss_Regular(p.pRoot).iVar);
        } else {
            Dss_NtkPrint_rec( p, Dss_Regular(p.pRoot) ); }
    }
    fmt.Printf("\n");
}

/**Function*************************************************************

  Synopsis    [Creating DSD network from SOP.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
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
func Dss_NtkCreate_rec(pStr string, p *string, pMatches *int32, pNtk *Dss_Ntk_t, pTruth *word) int32 {
    var fCompl int32;
    if **p == '!' {
        fCompl = 1;
        (*p)++;
    }
    for  (**p >= 'A' && **p <= 'F') || (**p >= '0' && **p <= '9')  {
        (*p)++;
    }
/*
    if ( **p == '<' )
    {
        char * q = pStr + pMatches[ *p - pStr ];
        if ( *(q+1) == '{' )
            *p = q+1;
    }
*/
    if **p >= 'a' && **p <= 'z' { // var
        return Abc_Var2Lit( Dss_ObjId(Dss_VecVar(pNtk.vObjs, **p - 'a')), fCompl ); }
    if **p == '(' || **p == '[' || **p == '<' || **p == '{' { // and/or/xor
    
        var pObj *Dss_Obj_t;
        var vFaninLits *Vec_Int_t  = Vec_IntAlloc( 10 );
        var q string  = pStr + pMatches[ *p - pStr ];
        var Type int32;
        if **p == '(' {
            Type = DAU_DSD_AND;
        } else if **p == '[' {
            Type = DAU_DSD_XOR;
        } else if **p == '<' {
            Type = DAU_DSD_MUX;
        } else if **p == '{' {
            Type = DAU_DSD_PRIME;
        } else { assert(false, "0"); }
        assert(*q == **p + 1 + (**p != '('), "*q == **p + 1 + (**p != '(')");
        for  (*p)++; *p < q; (*p)++  {
            Vec_IntPush( vFaninLits, Dss_NtkCreate_rec(pStr, p, pMatches, pNtk, pTruth) );
        }
        assert(*p == q, "*p == q");
        if Type == DAU_DSD_PRIME {
            var vFaninLitsNew *Vec_Int_t;
            var pTemp [DAU_MAX_WORD]word;
            var pCanonPerm [DAU_MAX_VAR]byte;
            var i int32;
            var uCanonPhase int32;
            var nFanins int32 = Vec_IntSize(vFaninLits);
            Abc_TtCopy( pTemp, pTruth, Abc_TtWordNum(nFanins), 0 );
            uCanonPhase = Abc_TtCanonicize( pTemp, nFanins, pCanonPerm );
            fCompl = (uCanonPhase >> nFanins) & 1;
            vFaninLitsNew = Vec_IntAlloc( nFanins );
            for  i = 0; i < nFanins; i++  {
                Vec_IntPush( vFaninLitsNew, Abc_LitNotCond(Vec_IntEntry(vFaninLits, pCanonPerm[i]), (uCanonPhase>>i)&1) );
            }
            pObj = Dss_ObjCreateNtk( pNtk, DAU_DSD_PRIME, vFaninLitsNew );
            Abc_TtCopy( Dss_ObjTruth(pObj), pTemp, Abc_TtWordNum(nFanins), 0 );
            Vec_IntFree( vFaninLitsNew );
        
        } else {
            pObj = Dss_ObjCreateNtk( pNtk, Type, vFaninLits ); }
        Vec_IntFree( vFaninLits );
        return Abc_LitNotCond( Dss_Obj2Lit(pObj), fCompl );
    }
    assert(false, "0");
    return -1;
}
func Dss_NtkCreate(pDsd string, nVars int32, pTruth *word) *Dss_Ntk_t {
    var fCompl int32;
    var pNtk *Dss_Ntk_t  = Dss_NtkAlloc( nVars );
    if *pDsd == '!' {
         pDsd++, fCompl = 1; }
    if Dau_DsdIsConst(pDsd) {
        pNtk.pRoot = Dss_VecConst0(pNtk.vObjs);
    } else if Dau_DsdIsVar(pDsd) {
        pNtk.pRoot = Dss_VecVar(pNtk.vObjs, Dau_DsdReadVar(pDsd));
    } else {
        var iLit int32;
        var pMatches [DAU_MAX_STR]int32;
        Dau_DsdMergeMatches( pDsd, pMatches );
        iLit = Dss_NtkCreate_rec( pDsd, &pDsd, pMatches, pNtk, pTruth );
        pNtk.pRoot = Dss_Lit2Obj( pNtk.vObjs, iLit );
    }
    if fCompl != 0 {
        pNtk.pRoot = Dss_Not(pNtk.pRoot); }
    return pNtk;
}
 
/**Function*************************************************************

  Synopsis    [Comparing two DSD nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dss_ObjCompare(p *Vec_Ptr_t, p0i *Dss_Obj_t, p1i *Dss_Obj_t) int32 {
    var p0 *Dss_Obj_t  = Dss_Regular(p0i);
    var p1 *Dss_Obj_t  = Dss_Regular(p1i);
    var pChild0, pChild1 *Dss_Obj_t;
    var i, Res int32;
    if Dss_ObjType(p0) < Dss_ObjType(p1) {
        return -1; }
    if Dss_ObjType(p0) > Dss_ObjType(p1) {
        return 1; }
    if Dss_ObjType(p0) < DAU_DSD_AND {
        return 0; }
    if Dss_ObjFaninNum(p0) < Dss_ObjFaninNum(p1) {
        return -1; }
    if Dss_ObjFaninNum(p0) > Dss_ObjFaninNum(p1) {
        return 1; }
    for  i = 0; i < Dss_ObjFaninNum(p0); i++  {
        pChild0 = Dss_ObjChild( p, p0, i );
        pChild1 = Dss_ObjChild( p, p1, i );
        Res = Dss_ObjCompare( p, pChild0, pChild1 );
        if Res != 0 {
            return Res; }
    }
    if Dss_IsComplement(p0i) < Dss_IsComplement(p1i) {
        return -1; }
    if Dss_IsComplement(p0i) > Dss_IsComplement(p1i) {
        return 1; }
    return 0;
}
func Dss_ObjSort(p *Vec_Ptr_t, pNodes **Dss_Obj_t, nNodes int32, pPerm *int32) {
    var i, j, best_i int32;
    for  i = 0; i < nNodes-1; i++  {
        best_i = i;
        for  j = i+1; j < nNodes; j++  {
            if Dss_ObjCompare(p, pNodes[best_i], pNodes[j]) == 1 {
                best_i = j; }
        }
        if i == best_i {
            continue; }
        // inlining of macro ABC_SWAP
        { var t *Dss_Obj_t  = pNodes[i]; pNodes[i] = pNodes[best_i]; pNodes[best_i] = t; };
        if pPerm != nil {
            // inlining of macro ABC_SWAP
             var t int32  = pPerm[i]; pPerm[i] = pPerm[best_i]; pPerm[best_i] = t; };
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dss_NtkCheck(p *Dss_Ntk_t) {
    var pObj, pFanin *Dss_Obj_t;
    var i, k int32;
    // inlining of macro Dss_VecForEachNode
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Dss_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  {             if pObj.Type == DAU_DSD_CONST0 || pObj.Type == DAU_DSD_VAR {} else {
        // inlining of macro Dss_ObjForEachFanin
        for  k = 0; (k < Dss_ObjFaninNum(pObj)) && ((pFanin) = Dss_ObjFanin(p.vObjs, pObj, k)); k++  {
            if pObj.Type == DAU_DSD_AND && pFanin.Type == DAU_DSD_AND {
                assert((Dss_ObjFaninC(pObj, k) != 0), "Dss_ObjFaninC(pObj, k)");
            } else if pObj.Type == DAU_DSD_XOR {
                assert(pFanin.Type != DAU_DSD_XOR, "pFanin.Type != DAU_DSD_XOR");
            } else if pObj.Type == DAU_DSD_MUX {
                assert(!(Dss_ObjFaninC(pObj, 0) != 0), "!Dss_ObjFaninC(pObj, 0)"); }
        }
    }
    }
}
func Dss_NtkCollectPerm_rec(p *Dss_Ntk_t, pObj *Dss_Obj_t, pPermDsd *int32, pnPerms *int32) int32 {
    var pChild *Dss_Obj_t;
    var k int32;
    var fCompl int32 = Dss_IsComplement(pObj);
    pObj = Dss_Regular( pObj );
    if pObj.Type == DAU_DSD_VAR {
        pPermDsd[*pnPerms] = Abc_Var2Lit(pObj.iVar, fCompl);
        pObj.iVar = (*pnPerms)++;
        return fCompl;
    }
    // inlining of macro Dss_ObjForEachChild
    for  k = 0; (k < Dss_ObjFaninNum(pObj)) && ((pChild) = Dss_ObjChild(p.vObjs, pObj, k)); k++  {
        if Dss_NtkCollectPerm_rec( p, pChild, pPermDsd, pnPerms ) != 0 {
            pObj.pFans[k] = (unsigned char)Abc_LitRegular(int32(pObj.pFans[k])); }
    }
    return 0;
}
func Dss_NtkTransform(p *Dss_Ntk_t, pPermDsd *int32) {
    var pChildren [DAU_MAX_VAR]*Dss_Obj_t;
    var pObj, pChild *Dss_Obj_t;
    var i, k, nPerms int32;
    if Dss_Regular(p.pRoot).Type == DAU_DSD_CONST0 {
        return; }
    // inlining of macro Dss_VecForEachNode
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Dss_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  {             if pObj.Type == DAU_DSD_CONST0 || pObj.Type == DAU_DSD_VAR {} else {
        if pObj.Type == DAU_DSD_MUX || pObj.Type == DAU_DSD_PRIME {
            continue; }
        // inlining of macro Dss_ObjForEachChild
        for  k = 0; (k < Dss_ObjFaninNum(pObj)) && ((pChild) = Dss_ObjChild(p.vObjs, pObj, k)); k++  {
            pChildren[k] = pChild;
        }
        Dss_ObjSort( p.vObjs, pChildren, Dss_ObjFaninNum(pObj), nil );
        for  k = 0; k < Dss_ObjFaninNum(pObj); k++  {
            pObj.pFans[k] = Dss_Obj2Lit( pChildren[k] );
        }
    }
    }
    nPerms = 0;
    if Dss_NtkCollectPerm_rec( p, p.pRoot, pPermDsd, &nPerms ) != 0 {
        p.pRoot = Dss_Regular(p.pRoot); }
    assert(nPerms == int32(Dss_Regular(p.pRoot).nSupp), "nPerms == int32(Dss_Regular(p.pRoot).nSupp)");
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dss_ObjAlloc(p *Dss_Man_t, Type int32, nFans int32, nTruthVars int32) *Dss_Obj_t {
    var nWords int32  = Dss_ObjWordNum(nFans) + (nTruthVars ? Abc_TtWordNum(nTruthVars) : 0);
    var pObj *Dss_Obj_t  = (Dss_Obj_t *)Mem_FlexEntryFetch( p.pMem, sizeof(word) * nWords );
    Dss_ObjClean( pObj );
    pObj.Type   = Type;
    pObj.nFans  = nFans;
    pObj.nWords = Dss_ObjWordNum(nFans);
    pObj.Id     = Vec_PtrSize( p.vObjs );
    pObj.iVar   = 31;
    Vec_PtrPush( p.vObjs, pObj );
    Vec_IntPush( p.vNexts, 0 );
    return pObj;
}
func Dss_ObjCreate(p *Dss_Man_t, Type int32, vFaninLits *Vec_Int_t, pTruth *word) *Dss_Obj_t {
    var pObj *Dss_Obj_t;
    var pFanin *Dss_Obj_t;
    var pPrev *Dss_Obj_t = nil;
    var i, Entry int32;
    // check structural canonicity
    assert(Type != DAU_DSD_MUX || Vec_IntSize(vFaninLits) == 3, "Type != DAU_DSD_MUX || Vec_IntSize(vFaninLits) == 3");
    assert(Type != DAU_DSD_MUX || !Abc_LitIsCompl(Vec_IntEntry(vFaninLits, 0)), "Type != DAU_DSD_MUX || !Abc_LitIsCompl(Vec_IntEntry(vFaninLits, 0))");
    assert(Type != DAU_DSD_MUX || !Abc_LitIsCompl(Vec_IntEntry(vFaninLits, 1)) || !Abc_LitIsCompl(Vec_IntEntry(vFaninLits, 2)), "Type != DAU_DSD_MUX || !Abc_LitIsCompl(Vec_IntEntry(vFaninLits, 1)) || !Abc_LitIsCompl(Vec_IntEntry(vFaninLits, 2))");
    // check that leaves are in good order
    if Type == DAU_DSD_AND || Type == DAU_DSD_XOR {
    // inlining of macro Dss_VecForEachObjVec
    for  i = 0; (i < Vec_IntSize(vFaninLits)) && ((pFanin) = Dss_Lit2Obj(p.vObjs, Vec_IntEntry(vFaninLits,i))); i++  {
        assert(Type != DAU_DSD_AND || Abc_LitIsCompl(Vec_IntEntry(vFaninLits, i)) || Dss_ObjType(pFanin) != DAU_DSD_AND, "Type != DAU_DSD_AND || Abc_LitIsCompl(Vec_IntEntry(vFaninLits, i)) || Dss_ObjType(pFanin) != DAU_DSD_AND");
        assert(Type != DAU_DSD_XOR || Dss_ObjType(pFanin) != DAU_DSD_XOR, "Type != DAU_DSD_XOR || Dss_ObjType(pFanin) != DAU_DSD_XOR");
        assert(pPrev == nil || Dss_ObjCompare(p.vObjs, pPrev, pFanin) <= 0, "pPrev == nil || Dss_ObjCompare(p.vObjs, pPrev, pFanin) <= 0");
        pPrev = pFanin;
    } }
    // create new node
    pObj = Dss_ObjAlloc( p, Type, Vec_IntSize(vFaninLits), Type == DAU_DSD_PRIME ? Vec_IntSize(vFaninLits) : 0 );
    if Type == DAU_DSD_PRIME {
        Abc_TtCopy( Dss_ObjTruth(pObj), pTruth, Abc_TtWordNum(Vec_IntSize(vFaninLits)), 0 ); }
    assert(pObj.nSupp == 0, "pObj.nSupp == 0");
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vFaninLits)) && (((Entry) = Vec_IntEntry(vFaninLits, i)), 1); i++  {
        pObj.pFans[i] = Entry;
        pObj.nSupp += Dss_VecLitSuppSize(p.vObjs, Entry);
    }
/*
    {
        extern void Dss_ManPrintOne( Dss_Man_t * p, int iDsdLit, int * pPermLits );
        Dss_ManPrintOne( p, Dss_Obj2Lit(pObj), NULL );
    }
*/
    return pObj;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dss_ManHashProfile(p *Dss_Man_t) {
    var pObj *Dss_Obj_t;
    var pSpot *uint32;
    var i, Counter int32;
    for  i = 0; i < p.nBins; i++  {
        Counter = 0;
        for  pSpot = p.pBins + i; *pSpot; pSpot = (unsigned *)Vec_IntEntryP(p.vNexts, pObj.Id), Counter++  {
             pObj = Dss_VecObj( p.vObjs, *pSpot );
        }
        if Counter != 0 {
            fmt.Printf("%d ", Counter); }
    }
    fmt.Printf("\n");
}
var spStatic_Dss_ObjHashKey_s_Primes [8]int32 = [8]int32{ 1699, 4177, 5147, 5647, 6343, 7103, 7873, 8147 };
func Dss_ObjHashKey(p *Dss_Man_t, Type int32, vFaninLits *Vec_Int_t, pTruth *word) uint32 {
    
    var i, Entry int32;
    var uHash uint32  = Type * 7873 + Vec_IntSize(vFaninLits) * 8147;
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vFaninLits)) && (((Entry) = Vec_IntEntry(vFaninLits, i)), 1); i++  {
        uHash += Entry * spStatic_Dss_ObjHashKey_s_Primes[i & 0x7];
    }
    assert((Type == DAU_DSD_PRIME) == (pTruth != nil), "(Type == DAU_DSD_PRIME) == (pTruth != nil)");
    if pTruth != nil {
        unsigned char * pTruthC = (unsigned char *)pTruth;
        var nBytes int32  = Abc_TtByteNum(Vec_IntSize(vFaninLits));
        for  i = 0; i < nBytes; i++  {
            uHash += pTruthC[i] * spStatic_Dss_ObjHashKey_s_Primes[i & 0x7];
        }
    }
    return uHash % p.nBins;
}
func Dss_ObjHashLookup(p *Dss_Man_t, Type int32, vFaninLits *Vec_Int_t, pTruth *word) *uint32 {
    var pObj *Dss_Obj_t;
    var pSpot *uint32  = p.pBins + Dss_ObjHashKey(p, Type, vFaninLits, pTruth);
    for  ; *pSpot; pSpot = (unsigned *)Vec_IntEntryP(p.vNexts, pObj.Id)  {
        pObj = Dss_VecObj( p.vObjs, *pSpot );
        if int32(pObj.Type) == Type && 
             int32(pObj.nFans) == Vec_IntSize(vFaninLits) && 
             !memcmp(pObj.pFans, Vec_IntArray(vFaninLits), sizeof(int)*pObj.nFans) &&
             (pTruth == nil || !memcmp(Dss_ObjTruth(pObj), pTruth, (size_t)Abc_TtByteNum(pObj.nFans))) { // equal
            return pSpot; }
    }
    return pSpot;
}
func Dss_ObjFindOrAdd(p *Dss_Man_t, Type int32, vFaninLits *Vec_Int_t, pTruth *word) *Dss_Obj_t {
    var pObj *Dss_Obj_t;
    var pSpot *uint32  = Dss_ObjHashLookup( p, Type, vFaninLits, pTruth );
    if *pSpot {
        return Dss_VecObj( p.vObjs, *pSpot ); }
    *pSpot = Vec_PtrSize( p.vObjs );
    pObj = Dss_ObjCreate( p, Type, vFaninLits, pTruth );
    return pObj;
}

/**Function*************************************************************

  Synopsis    [Cache for decomposition calls.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dss_ManCacheAlloc(p *Dss_Man_t) {
    assert(p.nCache == 0, "p.nCache == 0");
    p.nCache = Abc_PrimeCudd( 100000 );
    // inlining of macro ABC_CALLOC
    p.pCache = ((Dss_Ent_t * *) calloc((size_t)(p.nCache), sizeof(Dss_Ent_t *)));
}
func Dss_ManCacheFree(p *Dss_Man_t) {
    if p.pCache == nil {
        return; }
    assert(p.nCache != 0, "p.nCache != 0");
    p.nCache = 0;
    // inlining of macro ABC_FREE
    ((p.pCache) ? (free((char *) (p.pCache)), (p.pCache) = 0) : 0);
}
var spStatic_Dss_ManCacheHashKey_s_Primes [8]int32 = [8]int32{ 1699, 4177, 5147, 5647, 6343, 7103, 7873, 8147 };
func Dss_ManCacheHashKey(p *Dss_Man_t, pEnt *Dss_Ent_t) uint32 { 
    
    var i int32;
    var uHash uint32  = pEnt.nShared * 7103 + pEnt.iDsd0 * 7873 + pEnt.iDsd1 * 8147;
    for  i = 0; i < 2*int32(pEnt.nShared); i++  {
        uHash += pEnt.pShared[i] * spStatic_Dss_ManCacheHashKey_s_Primes[i & 0x7];
    }
    return uHash % p.nCache;
}
func Dss_ManCacheProfile(p *Dss_Man_t) {
    var pSpot **Dss_Ent_t;
    var i, Counter int32;
    for  i = 0; i < p.nCache; i++  {
        Counter = 0;
        for  pSpot = p.pCache + i; *pSpot; pSpot = &(*pSpot).pNext, Counter++  {}
        if Counter != 0 {
            fmt.Printf("%d ", Counter); }
    }
    fmt.Printf("\n");
}
func Dss_ManCacheLookup(p *Dss_Man_t, pEnt *Dss_Ent_t) **Dss_Ent_t {
    var pSpot **Dss_Ent_t  = p.pCache + Dss_ManCacheHashKey( p, pEnt );
    for  ; *pSpot; pSpot = &(*pSpot).pNext  {
        if (*pSpot).iDsd0   == pEnt.iDsd0 && 
             (*pSpot).iDsd1   == pEnt.iDsd1 && 
             (*pSpot).nShared == pEnt.nShared && 
             !memcmp((*pSpot).pShared, pEnt.pShared, sizeof(char)*2*pEnt.nShared) { // equal
        
            p.nCacheHits[pEnt.nShared!=0]++;
            return pSpot;
        }
    }
    p.nCacheMisses[pEnt.nShared!=0]++;
    return pSpot;
}
func Dss_ManCacheCreate(p *Dss_Man_t, pEnt0 *Dss_Ent_t, pFun0 *Dss_Fun_t) *Dss_Ent_t {
    var pEnt *Dss_Ent_t  = (Dss_Ent_t *)Mem_FlexEntryFetch( p.pMemEnts, sizeof(word) * pEnt0.nWords );
    var pFun *Dss_Fun_t  = (Dss_Fun_t *)Mem_FlexEntryFetch( p.pMemEnts, sizeof(word) * Dss_FunWordNum(pFun0) );
    memcpy( pEnt, pEnt0, sizeof(word) * pEnt0.nWords );
    memcpy( pFun, pFun0, sizeof(word) * Dss_FunWordNum(pFun0) );
    pEnt.pFunc = pFun;
    pEnt.pNext = nil;
    p.nCacheEntries[pEnt.nShared!=0]++;
    return pEnt;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dss_ManAlloc(nVars int32, nNonDecLimit int32) *Dss_Man_t {
    
    // inlining of macro ABC_CALLOC
    p := &Dss_Man_t{};
    p.nVars  = nVars;
    p.nNonDecLimit = nNonDecLimit;
    p.nBins  = Abc_PrimeCudd( 1000000 );
    // inlining of macro ABC_CALLOC
    p.pBins  = ((unsigned *) calloc((size_t)(p.nBins), sizeof(unsigned)));
    p.pMem   = Mem_FlexStart();
    p.vObjs  = Vec_PtrAlloc( 10000 );
    p.vNexts = Vec_IntAlloc( 10000 );
    Dss_ObjAlloc( p, DAU_DSD_CONST0, 0, 0 );
    Dss_ObjAlloc( p, DAU_DSD_VAR, 0, 0 ).nSupp = 1;
    p.vLeaves = Vec_IntAlloc( 32 );
    p.vCopies = Vec_IntAlloc( 32 );
    p.pTtElems = Dss_ManTtElems();
    p.pMemEnts = Mem_FlexStart();
//    Dss_ManCacheAlloc( p );
    return p;
}
func Dss_ManFree(p *Dss_Man_t) {
    Abc_PrintTime( 1, "Time begin ", p.timeBeg );
    Abc_PrintTime( 1, "Time decomp", p.timeDec );
    Abc_PrintTime( 1, "Time lookup", p.timeLook );
    Abc_PrintTime( 1, "Time end   ", p.timeEnd );

//    Dss_ManCacheProfile( p );
    Dss_ManCacheFree( p );
    Mem_FlexStop( p.pMemEnts, 0 );
    Vec_IntFreeP( &p.vCopies );
    Vec_IntFreeP( &p.vLeaves );
    Vec_IntFreeP( &p.vNexts );
    Vec_PtrFreeP( &p.vObjs );
    Mem_FlexStop( p.pMem, 0 );
    // inlining of macro ABC_FREE
    ((p.pBins) ? (free((char *) (p.pBins)), (p.pBins) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}
func Dss_ManPrint_rec(pFile *os.File, p *Dss_Man_t, pObj *Dss_Obj_t, pPermLits *int32, pnSupp *int32) {
    var OpenType [7]byte = [7]byte{0, 0, 0, '(', '[', '<', '{'};
    var CloseType [7]byte = [7]byte{0, 0, 0, ')', ']', '>', '}'};
    var pFanin *Dss_Obj_t;
    var i int32;
    assert(!(Dss_IsComplement(pObj) != 0), "!Dss_IsComplement(pObj)");
    if pObj.Type == DAU_DSD_CONST0 { fmt.Fprintf(pFile, "0"); return; }
    if pObj.Type == DAU_DSD_VAR {
        var iPermLit int32;
        if pPermLits != nil {
            iPermLit = pPermLits[(*pnSupp)++];
        } else {
            iPermLit = Abc_Var2Lit((*pnSupp)++, 0);
        }
        fmt.Fprintf(pFile, "%s%c", Abc_LitIsCompl(iPermLit)? "!":"", 'a' + Abc_Lit2Var(iPermLit));
        return;
    }
    if pObj.Type == DAU_DSD_PRIME {
        Abc_TtPrintHexRev( pFile, Dss_ObjTruth(pObj), pObj.nFans ); }
    fmt.Fprintf(pFile, "%c", OpenType[pObj.Type]);
    // inlining of macro Dss_ObjForEachFanin
    for  i = 0; (i < Dss_ObjFaninNum(pObj)) && ((pFanin) = Dss_ObjFanin(p.vObjs, pObj, i)); i++  {
        fmt.Fprintf(pFile, "%s", Dss_ObjFaninC(pObj, i) ? "!":"");
        Dss_ManPrint_rec( pFile, p, pFanin, pPermLits, pnSupp );
    }
    fmt.Fprintf(pFile, "%c", CloseType[pObj.Type]);
}
func Dss_ManPrintOne(pFile *os.File, p *Dss_Man_t, iDsdLit int32, pPermLits *int32) {
    var nSupp int32;
    fmt.Fprintf(pFile, "%6d : ", Abc_Lit2Var(iDsdLit));
    fmt.Fprintf(pFile, "%2d ", Dss_VecLitSuppSize(p.vObjs, iDsdLit));
    fmt.Fprintf(pFile, "%s", Abc_LitIsCompl(iDsdLit) ? "!" : "");
    Dss_ManPrint_rec( pFile, p, Dss_VecObj(p.vObjs, Abc_Lit2Var(iDsdLit)), pPermLits, &nSupp );
    fmt.Fprintf(pFile, "\n");
    assert(nSupp == int32(Dss_VecObj(p.vObjs, Abc_Lit2Var(iDsdLit)).nSupp), "nSupp == int32(Dss_VecObj(p.vObjs, Abc_Lit2Var(iDsdLit)).nSupp)");
}
func Dss_ManCheckNonDec_rec(p *Dss_Man_t, pObj *Dss_Obj_t) int32 {
    var pFanin *Dss_Obj_t;
    var i int32;
    assert(!(Dss_IsComplement(pObj) != 0), "!Dss_IsComplement(pObj)");
    if pObj.Type == DAU_DSD_CONST0 {
        return 0; }
    if pObj.Type == DAU_DSD_VAR {
        return 0; }
    if pObj.Type == DAU_DSD_PRIME {
        return 1; }
    // inlining of macro Dss_ObjForEachFanin
    for  i = 0; (i < Dss_ObjFaninNum(pObj)) && ((pFanin) = Dss_ObjFanin(p.vObjs, pObj, i)); i++  {
        if Dss_ManCheckNonDec_rec( p, pFanin ) != 0 {
            return 1; }
    }
    return 0;
}
func Dss_ManDump(p *Dss_Man_t) {
    var pFileName string  = "dss_tts.txt";
    var pFile *os.File;
    var Temp [DAU_MAX_WORD]word;
    var pObj *Dss_Obj_t;
    var i int32;
    pFile = os.OpenFile(pFileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    if pFile == nil {
        fmt.Printf("Cannot open file \"%s\".\n", pFileName);
        return;
    }
    // inlining of macro Dss_VecForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Dss_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  {
        if pObj.Type != DAU_DSD_PRIME {
            continue; }
        Abc_TtCopy( Temp, Dss_ObjTruth(pObj), Abc_TtWordNum(pObj.nFans), 0 );
        Abc_TtStretch6( Temp, pObj.nFans, p.nVars );
        fmt.Fprintf(pFile, "0x");
        Abc_TtPrintHexRev( pFile, Temp, p.nVars );
        fmt.Fprintf(pFile, "\n");

//        printf( "%6d : ", i );
//        Abc_TtPrintHexRev( stdout, Temp, p->nVars );
//        printf( "    " );
//        Dau_DsdPrintFromTruth( stdout, Temp, p->nVars );
    }
    _ = (pFile).Close();
}
func Dss_ManPrint(pFileName string, p *Dss_Man_t) {
    var pObj *Dss_Obj_t;
    var CountNonDsd, CountNonDsdStr int32;
    var i int32;
    var clk int32 = Abc_Clock();
    var pFile *os.File;
    if pFileName {
        pFile = os.OpenFile(pFileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    } else {
        pFile = os.Stdout;
    }
    if pFileName && pFile == nil {
        fmt.Printf("cannot open output file\n");
        return;
    }
    // inlining of macro Dss_VecForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Dss_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  {
        CountNonDsd += (pObj.Type == DAU_DSD_PRIME);
        CountNonDsdStr += Dss_ManCheckNonDec_rec( p, pObj );
    }
    fmt.Fprintf(pFile, "Total number of objects    = %8d\n", Vec_PtrSize(p.vObjs));
    fmt.Fprintf(pFile, "Non-DSD objects (max =%2d)  = %8d\n", p.nNonDecLimit, CountNonDsd);
    fmt.Fprintf(pFile, "Non-DSD structures         = %8d\n", CountNonDsdStr);
    fmt.Fprintf(pFile, "Memory used for objects    = %6.2f MB.\n", 1.0*Mem_FlexReadMemUsage(p.pMem)/(1<<20));
    fmt.Fprintf(pFile, "Memory used for array      = %6.2f MB.\n", 1.0*sizeof(void *)*Vec_PtrCap(p.vObjs)/(1<<20));
    fmt.Fprintf(pFile, "Memory used for hash table = %6.2f MB.\n", 1.0*sizeof(int)*p.nBins/(1<<20));
    fmt.Fprintf(pFile, "Memory used for cache      = %6.2f MB.\n", 1.0*Mem_FlexReadMemUsage(p.pMemEnts)/(1<<20));
    fmt.Fprintf(pFile, "Cache hits    = %8d %8d\n", p.nCacheHits[0], p.nCacheHits[1]);
    fmt.Fprintf(pFile, "Cache misses  = %8d %8d\n", p.nCacheMisses[0], p.nCacheMisses[1]);
    fmt.Fprintf(pFile, "Cache entries = %8d %8d\n", p.nCacheEntries[0], p.nCacheEntries[1]);
    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
//    Dss_ManHashProfile( p );
//    Dss_ManDump( p );
//    return;
    // inlining of macro Dss_VecForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Dss_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  {
        if i == 50 {
            break; }
        Dss_ManPrintOne( pFile, p, Dss_Obj2Lit(pObj), nil );
    }
    fmt.Fprintf(pFile, "\n");
    if pFileName {
        _ = (pFile).Close(); }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dss_ManComputeTruth_rec(p *Dss_Man_t, pObj *Dss_Obj_t, nVars int32, pRes *word, pPermLits *int32, pnSupp *int32) {
    var pChild *Dss_Obj_t;
    var nWords int32  = Abc_TtWordNum(nVars);
    var i int32;
    var fCompl int32 = Dss_IsComplement(pObj);
    pObj = Dss_Regular(pObj);
    if pObj.Type == DAU_DSD_VAR {
        var iPermLit int32  = pPermLits[(*pnSupp)++];
        assert((*pnSupp) <= nVars, "(*pnSupp) <= nVars");
        Abc_TtCopy( pRes, p.pTtElems[Abc_Lit2Var(iPermLit)], nWords, fCompl ^ Abc_LitIsCompl(iPermLit) );
        return;
    }
    if pObj.Type == DAU_DSD_AND || pObj.Type == DAU_DSD_XOR {
        var pTtTemp [DAU_MAX_WORD]word;
        if pObj.Type == DAU_DSD_AND {
            Abc_TtConst1( pRes, nWords );
        } else {
            Abc_TtConst0( pRes, nWords ); }
        // inlining of macro Dss_ObjForEachChild
        for  i = 0; (i < Dss_ObjFaninNum(pObj)) && ((pChild) = Dss_ObjChild(p.vObjs, pObj, i)); i++  {
            Dss_ManComputeTruth_rec( p, pChild, nVars, pTtTemp, pPermLits, pnSupp );
            if pObj.Type == DAU_DSD_AND {
                Abc_TtAnd( pRes, pRes, pTtTemp, nWords, 0 );
            } else {
                Abc_TtXor( pRes, pRes, pTtTemp, nWords, 0 ); }
        }
        if fCompl != 0 { Abc_TtNot( pRes, nWords ); }
        return;
    }
    if pObj.Type == DAU_DSD_MUX { // mux
    
        var pTtTemp [3][DAU_MAX_WORD]word;
        // inlining of macro Dss_ObjForEachChild
        for  i = 0; (i < Dss_ObjFaninNum(pObj)) && ((pChild) = Dss_ObjChild(p.vObjs, pObj, i)); i++  {
            Dss_ManComputeTruth_rec( p, pChild, nVars, pTtTemp[i], pPermLits, pnSupp );
        }
        assert(i == 3, "i == 3");
        Abc_TtMux( pRes, pTtTemp[0], pTtTemp[1], pTtTemp[2], nWords );
        if fCompl != 0 { Abc_TtNot( pRes, nWords ); }
        return;
    }
    if pObj.Type == DAU_DSD_PRIME { // function
    
        var pFanins [DAU_MAX_VAR][DAU_MAX_WORD]word;
        // inlining of macro Dss_ObjForEachChild
        for  i = 0; (i < Dss_ObjFaninNum(pObj)) && ((pChild) = Dss_ObjChild(p.vObjs, pObj, i)); i++  {
            Dss_ManComputeTruth_rec( p, pChild, nVars, pFanins[i], pPermLits, pnSupp );
        }
        Dau_DsdTruthCompose_rec( Dss_ObjTruth(pObj), pFanins, pRes, pObj.nFans, nWords );
        if fCompl != 0 { Abc_TtNot( pRes, nWords ); }
        return;
    }
    assert(false, "0");

}
func Dss_ManComputeTruth(p *Dss_Man_t, iDsd int32, nVars int32, pPermLits *int32) *word {
    var pObj *Dss_Obj_t  = Dss_Lit2Obj(p.vObjs, iDsd);
    var pRes *word  = p.pTtElems[DAU_MAX_VAR];
    var nWords int32  = Abc_TtWordNum( nVars );
    var nSupp int32;
    assert(nVars <= DAU_MAX_VAR, "nVars <= DAU_MAX_VAR");
    if iDsd == 0 {
        Abc_TtConst0( pRes, nWords );
    } else if iDsd == 1 {
        Abc_TtConst1( pRes, nWords );
    } else if Dss_Regular(pObj).Type == DAU_DSD_VAR {
        var iPermLit int32  = pPermLits[nSupp++];
        Abc_TtCopy( pRes, p.pTtElems[Abc_Lit2Var(iPermLit)], nWords, Abc_LitIsCompl(iDsd) ^ Abc_LitIsCompl(iPermLit) );
    
    } else {
        Dss_ManComputeTruth_rec( p, pObj, nVars, pRes, pPermLits, &nSupp ); }
    assert(nSupp == int32(Dss_Regular(pObj).nSupp), "nSupp == int32(Dss_Regular(pObj).nSupp)");
    return pRes;
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
// returns literal of non-shifted tree in p, corresponding to pObj in pNtk, which may be compl
func Dss_NtkRebuild_rec(p *Dss_Man_t, pNtk *Dss_Ntk_t, pObj *Dss_Obj_t) int32 {
    var pChildren [DAU_MAX_VAR]*Dss_Obj_t;
    var pChild, pObjNew *Dss_Obj_t;
    var i int32;
    var k int32;
    var fCompl int32 = Dss_IsComplement(pObj);
    pObj = Dss_Regular(pObj);
    if pObj.Type == DAU_DSD_VAR {
        return Abc_Var2Lit( 1, fCompl ); }
    // inlining of macro Dss_ObjForEachChild
    for  k = 0; (k < Dss_ObjFaninNum(pObj)) && ((pChild) = Dss_ObjChild(pNtk.vObjs, pObj, k)); k++  {
        pChildren[k] = Dss_Lit2Obj( p.vObjs, Dss_NtkRebuild_rec( p, pNtk, pChild ) );
        if pObj.Type == DAU_DSD_XOR && (Dss_IsComplement(pChildren[k]) != 0) {
            pChildren[k] = Dss_Not(pChildren[k]), fCompl ^= 1; }
    }
    // normalize MUX
    if pObj.Type == DAU_DSD_MUX {
        if Dss_IsComplement(pChildren[0]) != 0 {
            pChildren[0] = Dss_Not(pChildren[0]);
            // inlining of macro ABC_SWAP
            { var t *Dss_Obj_t  = pChildren[1]; pChildren[1] = pChildren[2]; pChildren[2] = t; };
        }
        if Dss_IsComplement(pChildren[1]) != 0 {
            pChildren[1] = Dss_Not(pChildren[1]);
            pChildren[2] = Dss_Not(pChildren[2]);
            fCompl ^= 1;
        }
    }
    // shift subgraphs
    Vec_IntClear( p.vLeaves );
    for  i = 0; i < k; i++  {
        Vec_IntPush( p.vLeaves, Dss_Obj2Lit(pChildren[i]) );
    }
    // create new graph
    pObjNew = Dss_ObjFindOrAdd( p, pObj.Type, p.vLeaves, pObj.Type == DAU_DSD_PRIME ? Dss_ObjTruth(pObj) : nil );
    return Abc_Var2Lit( pObjNew.Id, fCompl );
}
func Dss_NtkRebuild(p *Dss_Man_t, pNtk *Dss_Ntk_t) int32 {
    assert(p.nVars == pNtk.nVars, "p.nVars == pNtk.nVars");
    if Dss_Regular(pNtk.pRoot).Type == DAU_DSD_CONST0 {
        return Dss_IsComplement(pNtk.pRoot); }
    if Dss_Regular(pNtk.pRoot).Type == DAU_DSD_VAR {
        return Abc_Var2Lit( Dss_Regular(pNtk.pRoot).iVar + 1, Dss_IsComplement(pNtk.pRoot) ); }
    return Dss_NtkRebuild_rec( p, pNtk, pNtk.pRoot );
}

/**Function*************************************************************

  Synopsis    [Performs DSD operation on the two literals.]

  Description [Returns the perm of the resulting literals. The perm size 
  is equal to the number of support variables. The perm variables are 0-based
  numbers of pLits[0] followed by nLits[0]-based numbers of pLits[1].]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
int Dss_ManOperation( Dss_Man_t * p, int Type, int * pLits, int nLits, unsigned char * pPerm, word * pTruth )
{
    var pChildren [DAU_MAX_VAR]*Dss_Obj_t;
    var pObj, pChild *Dss_Obj_t;
    var i, k, nChildren, fCompl, fComplFan int32;

    assert(Type == DAU_DSD_AND || pPerm == nil, "Type == DAU_DSD_AND || pPerm == nil");
    if Type == DAU_DSD_AND && pPerm != nil {
        var pBegEnd [DAU_MAX_VAR]int32;
        var j, nSSize int32;
        for  k = 0; k < nLits; k++  {
            pObj = Dss_Lit2Obj(p.vObjs, pLits[k]);
            if (Dss_IsComplement(pObj) != 0) || pObj.Type != DAU_DSD_AND {
                fComplFan = (Dss_Regular(pObj).Type == DAU_DSD_VAR && Dss_IsComplement(pObj));
                if fComplFan != 0 {
                    pObj = Dss_Regular(pObj); }
                pBegEnd[nChildren] = (nSSize << 16) | (fComplFan << 8) | (nSSize + Dss_Regular(pObj).nSupp);
                nSSize += Dss_Regular(pObj).nSupp;
                pChildren[nChildren++] = pObj;
            
            } else {
                // inlining of macro Dss_ObjForEachChild
                for  i = 0; (i < Dss_ObjFaninNum(pObj)) && ((pChild) = Dss_ObjChild(p.vObjs, pObj, i)); i++  {
                    fComplFan = (Dss_Regular(pChild).Type == DAU_DSD_VAR && Dss_IsComplement(pChild));
                    if fComplFan != 0 {
                        pChild = Dss_Regular(pChild); }
                    pBegEnd[nChildren] = (nSSize << 16) | (fComplFan << 8) | (nSSize + Dss_Regular(pChild).nSupp);
                    nSSize += Dss_Regular(pChild).nSupp;
                    pChildren[nChildren++] = pChild;
                } }
        }
        Dss_ObjSort( p.vObjs, pChildren, nChildren, pBegEnd );
        // create permutation
        for  j = i = 0; i < nChildren; i++  {
            for  k = (pBegEnd[i] >> 16); k < (pBegEnd[i] & 0xFF); k++  {
                pPerm[j++] = (unsigned char)Abc_Var2Lit( k, (pBegEnd[i] >> 8) & 1 );
            }
        }
        assert(j == nSSize, "j == nSSize");
    
    } else if Type == DAU_DSD_AND {
        for  k = 0; k < nLits; k++  {
            pObj = Dss_Lit2Obj(p.vObjs, pLits[k]);
            if (Dss_IsComplement(pObj) != 0) || pObj.Type != DAU_DSD_AND {
                pChildren[nChildren++] = pObj;
            } else {
                // inlining of macro Dss_ObjForEachChild
                for  i = 0; (i < Dss_ObjFaninNum(pObj)) && ((pChild) = Dss_ObjChild(p.vObjs, pObj, i)); i++  {
                    pChildren[nChildren++] = pChild;
                } }
        }
        Dss_ObjSort( p.vObjs, pChildren, nChildren, nil );
    
    } else if Type == DAU_DSD_XOR {
        for  k = 0; k < nLits; k++  {
            fCompl ^= Abc_LitIsCompl(pLits[k]);
            pObj = Dss_Lit2Obj(p.vObjs, Abc_LitRegular(pLits[k]));
            if pObj.Type != DAU_DSD_XOR {
                pChildren[nChildren++] = pObj;
            } else {
                // inlining of macro Dss_ObjForEachChild
                for  i = 0; (i < Dss_ObjFaninNum(pObj)) && ((pChild) = Dss_ObjChild(p.vObjs, pObj, i)); i++  {
                    assert(!(Dss_IsComplement(pChild) != 0), "!Dss_IsComplement(pChild)");
                    pChildren[nChildren++] = pChild;
                } }
        }
        Dss_ObjSort( p.vObjs, pChildren, nChildren, nil );
    
    } else if Type == DAU_DSD_MUX {
        if Abc_LitIsCompl(pLits[0]) {
            pLits[0] = Abc_LitNot(pLits[0]);
            // inlining of macro ABC_SWAP
            { var t int32  = pLits[1]; pLits[1] = pLits[2]; pLits[2] = t; };
        }
        if Abc_LitIsCompl(pLits[1]) {
            pLits[1] = Abc_LitNot(pLits[1]);
            pLits[2] = Abc_LitNot(pLits[2]);
            fCompl ^= 1;
        }
        for  k = 0; k < nLits; k++  {
            pChildren[nChildren++] = Dss_Lit2Obj(p.vObjs, pLits[k]);
        }
    
    } else if Type == DAU_DSD_PRIME {
        for  k = 0; k < nLits; k++  {
            pChildren[nChildren++] = Dss_Lit2Obj(p.vObjs, pLits[k]);
        }
    
    } else { assert(false, "0"); }

    // shift subgraphs
    Vec_IntClear( p.vLeaves );
    for  i = 0; i < nChildren; i++  {
        Vec_IntPush( p.vLeaves, Dss_Obj2Lit(pChildren[i]) );
    }
    // create new graph
    pObj = Dss_ObjFindOrAdd( p, Type, p.vLeaves, pTruth );
    return Abc_Var2Lit( pObj.Id, fCompl );
}
var spStatic_Dss_ManOperationFun_Buffer [100]byte;
func Dss_ManOperationFun(p *Dss_Man_t, iDsd *int32, nFansTot int32) *Dss_Fun_t {
    
    var pFun *Dss_Fun_t  = (Dss_Fun_t *)spStatic_Dss_ManOperationFun_Buffer;
    pFun.iDsd = Dss_ManOperation( p, DAU_DSD_AND, iDsd, 2, pFun.pFans, nil );
//printf( "%d %d -> %d  ", iDsd[0], iDsd[1], pFun->iDsd );
    pFun.nFans = nFansTot;
    assert(int32(pFun.nFans) == Dss_VecLitSuppSize(p.vObjs, pFun.iDsd), "int32(pFun.nFans) == Dss_VecLitSuppSize(p.vObjs, pFun.iDsd)");
    return pFun;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dss_EntPrint(p *Dss_Ent_t, pFun *Dss_Fun_t) {
    var i int32;
    fmt.Printf("%d %d ", p.iDsd0, p.iDsd1);
    for  i = 0; i < int32(p.nShared); i++  {
        fmt.Printf("%d=%d ", p.pShared[2*i], p.pShared[2*i+1]);
    }
    fmt.Printf("-> %d   ", pFun.iDsd);
}

/**Function*************************************************************

  Synopsis    [Performs AND on two DSD functions with support overlap.]

  Description [Returns the perm of the resulting literals. The perm size 
  is equal to the number of support variables. The perm variables are 0-based
  numbers of pLits[0] followed by nLits[0]-based numbers of pLits[1].]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_Dss_ManBooleanAnd_Buffer [100]byte;
func Dss_ManBooleanAnd(p *Dss_Man_t, pEnt *Dss_Ent_t, Counter int32) *Dss_Fun_t {
    
    var pFun *Dss_Fun_t  = (Dss_Fun_t *)spStatic_Dss_ManBooleanAnd_Buffer;
    var pNtk *Dss_Ntk_t;
    var pTruthOne *word;
    var pTruth [DAU_MAX_WORD]word;
    var pDsd [DAU_MAX_STR]byte;
    var pMapDsd2Truth [DAU_MAX_VAR]int32;
    var pPermLits [DAU_MAX_VAR]int32;
    var pPermDsd [DAU_MAX_VAR]int32;
    var i, nNonDec, nSuppSize int32;
    var nFans [2]int32;
    nFans[0] = Dss_VecLitSuppSize( p.vObjs, pEnt.iDsd0 );
    nFans[1] = Dss_VecLitSuppSize( p.vObjs, pEnt.iDsd1 );
    // create first truth table
    for  i = 0; i < nFans[0]; i++  {
        pMapDsd2Truth[nSuppSize] = i;
        pPermLits[i] = Abc_Var2Lit( nSuppSize++, 0 );
    }
    pTruthOne = Dss_ManComputeTruth( p, pEnt.iDsd0, p.nVars, pPermLits );
    Abc_TtCopy( pTruth, pTruthOne, Abc_TtWordNum(p.nVars), 0 );
if Counter != 0 {
//Kit_DsdPrintFromTruth( pTruthOne, p->nVars );  printf( "\n" );
}
    // create second truth table
    for  i = 0; i < nFans[1]; i++  {
        pPermLits[i] = -1;
    }
    for  i = 0; i < int32(pEnt.nShared); i++  {
        pPermLits[pEnt.pShared[2*i+0]] = pEnt.pShared[2*i+1];
    }
    for  i = 0; i < nFans[1]; i++  {
        if pPermLits[i] == -1 {
            pMapDsd2Truth[nSuppSize] = nFans[0] + i;
            pPermLits[i] = Abc_Var2Lit( nSuppSize++, 0 );
        }
    }
    pTruthOne = Dss_ManComputeTruth( p, pEnt.iDsd1, p.nVars, pPermLits );
if Counter != 0 {
//Kit_DsdPrintFromTruth( pTruthOne, p->nVars );  printf( "\n" );
}
    Abc_TtAnd( pTruth, pTruth, pTruthOne, Abc_TtWordNum(p.nVars), 0 );
    // perform decomposition
    nNonDec = Dau_DsdDecompose( pTruth, nSuppSize, 0, 0, pDsd );
    if p.nNonDecLimit && nNonDec > p.nNonDecLimit {
        return nil; }
    // derive network and convert it into the manager
    pNtk = Dss_NtkCreate( pDsd, p.nVars, nNonDec ? pTruth : nil );
//Dss_NtkPrint( pNtk );    
    Dss_NtkCheck( pNtk );
    Dss_NtkTransform( pNtk, pPermDsd );
//Dss_NtkPrint( pNtk );    
    pFun.iDsd = Dss_NtkRebuild( p, pNtk );
    Dss_NtkFree( pNtk );
    // pPermDsd maps vars of iDsdRes into literals of pTruth
    // translate this map into the one that maps vars of iDsdRes into literals of cut
    pFun.nFans = Dss_VecLitSuppSize( p.vObjs, pFun.iDsd );
    for  i = 0; i < int32(pFun.nFans); i++  {
        pFun.pFans[i] = (unsigned char)Abc_Lit2LitV( pMapDsd2Truth, pPermDsd[i] );
    }

//    Dss_EntPrint( pEnt, pFun );
    return pFun;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
// returns mapping of variables of dsd1 into literals of dsd0
var spStatic_Dss_ManSharedMap_Buffer [100]byte;
func Dss_ManSharedMap(p *Dss_Man_t, iDsd *int32, nFans *int32, pFans **int32, uSharedMask uint32) *Dss_Ent_t {
    
    var pEnt *Dss_Ent_t  = (Dss_Ent_t *)spStatic_Dss_ManSharedMap_Buffer;
    pEnt.iDsd0 = iDsd[0];
    pEnt.iDsd1 = iDsd[1];
    pEnt.nShared = 0;
    if uSharedMask != 0 {
        var i int32;
        var g int32;
        var pMapGtoL [DAU_MAX_VAR]int32 = [DAU_MAX_VAR]int32{-1};
        for  i = 0; i < nFans[0]; i++  {
            pMapGtoL[ Abc_Lit2Var(pFans[0][i]) ] = Abc_Var2Lit( i, Abc_LitIsCompl(pFans[0][i]) );
        }
        for  i = 0; i < nFans[1]; i++  {
            g = Abc_Lit2Var( pFans[1][i] );
            if (uSharedMask >> g) & 1 {
                assert(pMapGtoL[g] >= 0, "pMapGtoL[g] >= 0");
                pEnt.pShared[2*pEnt.nShared+0] = (unsigned char)i;
                pEnt.pShared[2*pEnt.nShared+1] = (unsigned char)Abc_LitNotCond( pMapGtoL[g], Abc_LitIsCompl(pFans[1][i]) );
                pEnt.nShared++;
            }
        }
    }
    pEnt.nWords = Dss_EntWordNum( pEnt );
    return pEnt;
}

// merge two DSD functions
int Dss_ManMerge( Dss_Man_t * p, int * iDsd, int * nFans, int ** pFans, unsigned uSharedMask, int nKLutSize, unsigned char * pPermRes, word * pTruth )
{
    var fVerbose int32;
    var fCheck int32;
    static int Counter = 0;
//    word pTtTemp[DAU_MAX_WORD];
    var pTruthOne *word;
    var pPermResInt [DAU_MAX_VAR]int32;
    var pEnt *Dss_Ent_t;
    var ppSpot **Dss_Ent_t;
    var pFun *Dss_Fun_t;
    var i int32;
    var clk abctime;
    Counter++;
    if DAU_MAX_VAR < nKLutSize {
        fmt.Printf("Paramater DAU_MAX_VAR (%d) smaller than LUT size (%d).\n", DAU_MAX_VAR, nKLutSize);
        return -1;
    }
    assert(iDsd[0] <= iDsd[1], "iDsd[0] <= iDsd[1]");

if fVerbose != 0 {
Dss_ManPrintOne( os.Stdout, p, iDsd[0], pFans[0] );
Dss_ManPrintOne( os.Stdout, p, iDsd[1], pFans[1] );
} 

    // constant argument
    if iDsd[0] == 0 { return 0; }
    if iDsd[0] == 1 { return iDsd[1]; }
    if iDsd[1] == 0 { return 0; }
    if iDsd[1] == 1 { return iDsd[0]; }

    // no overlap
clk = Abc_Clock();
    assert(nFans[0] == Dss_VecLitSuppSize(p.vObjs, iDsd[0]), "nFans[0] == Dss_VecLitSuppSize(p.vObjs, iDsd[0])");
    assert(nFans[1] == Dss_VecLitSuppSize(p.vObjs, iDsd[1]), "nFans[1] == Dss_VecLitSuppSize(p.vObjs, iDsd[1])");
    assert(nFans[0] + nFans[1] <= nKLutSize + Dss_WordCountOnes(uSharedMask), "nFans[0] + nFans[1] <= nKLutSize + Dss_WordCountOnes(uSharedMask)");
    // create map of shared variables
    pEnt = Dss_ManSharedMap( p, iDsd, nFans, pFans, uSharedMask );
p.timeBeg += Abc_Clock() - clk;
    // check cache
    if p.pCache == nil {
clk = Abc_Clock();
        if uSharedMask == 0 {
            pFun = Dss_ManOperationFun( p, iDsd, nFans[0] + nFans[1] );
        } else {
            pFun = Dss_ManBooleanAnd( p, pEnt, 0 ); }
        if pFun == nil {
            return -1; }
        assert(int32(pFun.nFans) == Dss_VecLitSuppSize(p.vObjs, pFun.iDsd), "int32(pFun.nFans) == Dss_VecLitSuppSize(p.vObjs, pFun.iDsd)");
        assert(int32(pFun.nFans) <= nKLutSize, "int32(pFun.nFans) <= nKLutSize");
p.timeDec += Abc_Clock() - clk;
    
    } else {
clk = Abc_Clock();
        ppSpot = Dss_ManCacheLookup( p, pEnt );
p.timeLook += Abc_Clock() - clk;
clk = Abc_Clock();
        if *ppSpot == nil {
            if uSharedMask == 0 {
                pFun = Dss_ManOperationFun( p, iDsd, nFans[0] + nFans[1] );
            } else {
                pFun = Dss_ManBooleanAnd( p, pEnt, 0 ); }
            if pFun == nil {
                return -1; }
            assert(int32(pFun.nFans) == Dss_VecLitSuppSize(p.vObjs, pFun.iDsd), "int32(pFun.nFans) == Dss_VecLitSuppSize(p.vObjs, pFun.iDsd)");
            assert(int32(pFun.nFans) <= nKLutSize, "int32(pFun.nFans) <= nKLutSize");
            // create cache entry
            *ppSpot = Dss_ManCacheCreate( p, pEnt, pFun );
        }
        pFun = (*ppSpot).pFunc;
p.timeDec += Abc_Clock() - clk;
    }

clk = Abc_Clock();
    for  i = 0; i < int32(pFun.nFans); i++  {
        if pFun.pFans[i] < 2 * nFans[0] { // first dec
            pPermRes[i] = (unsigned char)Dss_Lit2Lit( pFans[0], pFun.pFans[i] );
        } else {
            pPermRes[i] = (unsigned char)Dss_Lit2Lit( pFans[1], pFun.pFans[i] - 2 * nFans[0] ); }
    }
    // perform support minimization
    if (uSharedMask != 0) && pFun.nFans > 1 {
        var pVarPres [DAU_MAX_VAR]int32;
        var nSupp int32;
        for  i = 0; i < p.nVars; i++  {
            pVarPres[i] = -1;
        }
        for  i = 0; i < int32(pFun.nFans); i++  {
            pVarPres[ Abc_Lit2Var(pPermRes[i]) ] = i;
        }
        for  i = 0; i < p.nVars; i++  {
            if pVarPres[i] >= 0 {
                pPermRes[pVarPres[i]] = Abc_Var2Lit( nSupp++, Abc_LitIsCompl(pPermRes[pVarPres[i]]) ); }
        }
        assert(nSupp == int32(pFun.nFans), "nSupp == int32(pFun.nFans)");
    }

    for  i = 0; i < int32(pFun.nFans); i++  {
        pPermResInt[i] = pPermRes[i];
    }
p.timeEnd += Abc_Clock() - clk;

if fVerbose != 0 {
Dss_ManPrintOne( os.Stdout, p, pFun.iDsd, pPermResInt );
fmt.Printf("\n");
}

if Counter == 43418 {
//    int s = 0;
//    Dss_ManPrint( NULL, p );
}


    if fCheck != 0 {
        pTruthOne = Dss_ManComputeTruth( p, pFun.iDsd, p.nVars, pPermResInt );
        if !Abc_TtEqual( pTruthOne, pTruth, Abc_TtWordNum(p.nVars) ) {
            var s int32;
    //        Kit_DsdPrintFromTruth( pTruthOne, p->nVars );  printf( "\n" );
    //        Kit_DsdPrintFromTruth( pTruth, p->nVars );     printf( "\n" );
            fmt.Printf("Verification failed.\n");
            s = 0;
        }
    }
    return pFun.iDsd;
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_Dss_ManSharedMapDerive_Buffer [100]byte;
func Dss_ManSharedMapDerive(p *Dss_Man_t, iDsd0 int32, iDsd1 int32, vShared *Vec_Str_t) *Dss_Ent_t {
    
    var pEnt *Dss_Ent_t  = (Dss_Ent_t *)spStatic_Dss_ManSharedMapDerive_Buffer;
    pEnt.iDsd0 = iDsd0;
    pEnt.iDsd1 = iDsd1;
    pEnt.nShared = Vec_StrSize(vShared)/2;
    memcpy( pEnt.pShared, (unsigned char *)Vec_StrArray(vShared), sizeof(char) * Vec_StrSize(vShared) );
    pEnt.nWords = Dss_EntWordNum( pEnt );
    return pEnt;
}

func Mpm_FuncCompute(p *Dss_Man_t, iDsd0 int32, iDsd1 int32, vShared *Vec_Str_t, pPerm *int32, pnLeaves *int32) int32 {
    var fVerbose int32;
//    int fCheck = 0;
    var pEnt *Dss_Ent_t;
    var ppSpot **Dss_Ent_t;
    var pFun *Dss_Fun_t;
    var iDsd [2]int32 = [2]int32{ iDsd0, iDsd1 };
    var i int32;
    var clk abctime;

    assert(iDsd0 <= iDsd1, "iDsd0 <= iDsd1");
    if DAU_MAX_VAR < *pnLeaves {
        fmt.Printf("Paramater DAU_MAX_VAR (%d) smaller than LUT size (%d).\n", DAU_MAX_VAR, *pnLeaves);
        return -1;
    }
    if fVerbose != 0 {
        Dss_ManPrintOne( os.Stdout, p, iDsd0, nil );
        Dss_ManPrintOne( os.Stdout, p, iDsd1, nil );
    } 

clk = Abc_Clock();
    pEnt = Dss_ManSharedMapDerive( p, iDsd0, iDsd1, vShared );
    ppSpot = Dss_ManCacheLookup( p, pEnt );
p.timeLook += Abc_Clock() - clk;

clk = Abc_Clock();
    if *ppSpot == nil {
        if Vec_StrSize(vShared) == 0 {
            pFun = Dss_ManOperationFun( p, iDsd, *pnLeaves );
        } else {
            pFun = Dss_ManBooleanAnd( p, pEnt, 0 ); }
        if pFun == nil {
            return -1; }
        assert(int32(pFun.nFans) == Dss_VecLitSuppSize(p.vObjs, pFun.iDsd), "int32(pFun.nFans) == Dss_VecLitSuppSize(p.vObjs, pFun.iDsd)");
        assert(int32(pFun.nFans) <= *pnLeaves, "int32(pFun.nFans) <= *pnLeaves");
        // create cache entry
        *ppSpot = Dss_ManCacheCreate( p, pEnt, pFun );
    }
    pFun = (*ppSpot).pFunc;
p.timeDec += Abc_Clock() - clk;

    *pnLeaves = int32(pFun.nFans);
    for  i = 0; i < int32(pFun.nFans); i++  {
        pPerm[i] = int32(pFun.pFans[i]);
    }

    if fVerbose != 0 {
        Dss_ManPrintOne( os.Stdout, p, pFun.iDsd, nil );
        fmt.Printf("\n");
    }

/*
    if ( fCheck )
    {
        pTruthOne = Dss_ManComputeTruth( p, pFun->iDsd, p->nVars, pPermResInt );
        if ( !Abc_TtEqual( pTruthOne, pTruth, Abc_TtWordNum(p->nVars) ) )
        {
            int s;
    //        Kit_DsdPrintFromTruth( pTruthOne, p->nVars );  printf( "\n" );
    //        Kit_DsdPrintFromTruth( pTruth, p->nVars );     printf( "\n" );
            printf( "Verification failed.\n" );
            s = 0;
        }
    }
*/
    return pFun.iDsd;
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dss_ObjCheckTransparent(p *Dss_Man_t, pObj *Dss_Obj_t) int32 {
    var pFanin *Dss_Obj_t;
    var i int32;
    if pObj.Type == DAU_DSD_VAR {
        return 1; }
    if pObj.Type == DAU_DSD_AND {
        return 0; }
    if pObj.Type == DAU_DSD_XOR {
        // inlining of macro Dss_ObjForEachFanin
        for  i = 0; (i < Dss_ObjFaninNum(pObj)) && ((pFanin) = Dss_ObjFanin(p.vObjs, pObj, i)); i++  {
            if Dss_ObjCheckTransparent( p, pFanin ) != 0 {
                return 1; }
        }
        return 0;
    }
    if pObj.Type == DAU_DSD_MUX {
        pFanin = Dss_ObjFanin( p.vObjs, pObj, 1 );
        if !(Dss_ObjCheckTransparent(p, pFanin) != 0) {
            return 0; }
        pFanin = Dss_ObjFanin( p.vObjs, pObj, 2 );
        if !(Dss_ObjCheckTransparent(p, pFanin) != 0) {
            return 0; }
        return 1;
    }
    assert(pObj.Type == DAU_DSD_PRIME, "pObj.Type == DAU_DSD_PRIME");
    return 0;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dau_DsdTest__() {
    var nVars int32  = 8;
//    char * pDsd = "[(ab)(cd)]";
    var pDsd string  = "(!(a!(bh))[cde]!(fg))";
    var pNtk *Dss_Ntk_t  = Dss_NtkCreate( pDsd, nVars, nil );
//    Dss_NtkPrint( pNtk );
//    Dss_NtkCheck( pNtk );
//    Dss_NtkTransform( pNtk );
//    Dss_NtkPrint( pNtk );
    Dss_NtkFree( pNtk );
    nVars = 0;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dau_DsdTest() {
    var nVars int32  = 8;
    var vFuncs *Vec_Vec_t;
    var vOne, vTwo, vRes *Vec_Int_t;//, * vThree;
    var p *Dss_Man_t;
    var pEntries [3]int32;
    var iLit, e0, e1 int32;//, e2;
    var i, k, s int32;//, j;

    return;

    vFuncs = Vec_VecStart( nVars+1 );
    assert(nVars < DAU_MAX_VAR, "nVars < DAU_MAX_VAR");
    p = Dss_ManAlloc( nVars, 0 );

    // init
    Vec_VecPushInt( vFuncs, 1, Dss_Obj2Lit(Dss_VecVar(p.vObjs,0)) );

    // enumerate
    for  s = 2; s <= nVars; s++  {
        vRes = Vec_VecEntryInt( vFuncs, s );
        for  i = 1; i < s; i++  {
        for  k = i; k < s; k++  {
        if i + k == s {
            vOne = Vec_VecEntryInt( vFuncs, i );
            vTwo = Vec_VecEntryInt( vFuncs, k );
            // inlining of macro Vec_IntForEachEntry
            for  e0 = 0; (e0 < Vec_IntSize(vOne)) && (((pEntries[0]) = Vec_IntEntry(vOne, e0)), 1); e0++  {
            // inlining of macro Vec_IntForEachEntry
            for  e1 = 0; (e1 < Vec_IntSize(vTwo)) && (((pEntries[1]) = Vec_IntEntry(vTwo, e1)), 1); e1++  {
                var fAddInv0 int32  = !Dss_ObjCheckTransparent( p, Dss_VecObj(p.vObjs, Abc_Lit2Var(pEntries[0])) );
                var fAddInv1 int32  = !Dss_ObjCheckTransparent( p, Dss_VecObj(p.vObjs, Abc_Lit2Var(pEntries[1])) );

                iLit = Dss_ManOperation( p, DAU_DSD_AND, pEntries, 2, nil, nil );
                assert(!Abc_LitIsCompl(iLit), "!Abc_LitIsCompl(iLit)");
                Vec_IntPush( vRes, iLit );

                if fAddInv0 != 0 {
                    pEntries[0] = Abc_LitNot( pEntries[0] );
                    iLit = Dss_ManOperation( p, DAU_DSD_AND, pEntries, 2, nil, nil );
                    pEntries[0] = Abc_LitNot( pEntries[0] );
                    assert(!Abc_LitIsCompl(iLit), "!Abc_LitIsCompl(iLit)");
                    Vec_IntPush( vRes, iLit );
                }

                if fAddInv1 != 0 {
                    pEntries[1] = Abc_LitNot( pEntries[1] );
                    iLit = Dss_ManOperation( p, DAU_DSD_AND, pEntries, 2, nil, nil );
                    pEntries[1] = Abc_LitNot( pEntries[1] );
                    assert(!Abc_LitIsCompl(iLit), "!Abc_LitIsCompl(iLit)");
                    Vec_IntPush( vRes, iLit );
                }

                if (fAddInv0 != 0) && (fAddInv1 != 0) {
                    pEntries[0] = Abc_LitNot( pEntries[0] );
                    pEntries[1] = Abc_LitNot( pEntries[1] );
                    iLit = Dss_ManOperation( p, DAU_DSD_AND, pEntries, 2, nil, nil );
                    pEntries[0] = Abc_LitNot( pEntries[0] );
                    pEntries[1] = Abc_LitNot( pEntries[1] );
                    assert(!Abc_LitIsCompl(iLit), "!Abc_LitIsCompl(iLit)");
                    Vec_IntPush( vRes, iLit );
                }

                iLit = Dss_ManOperation( p, DAU_DSD_XOR, pEntries, 2, nil, nil );
                assert(!Abc_LitIsCompl(iLit), "!Abc_LitIsCompl(iLit)");
                Vec_IntPush( vRes, Abc_LitRegular(iLit) );
            }
            }
        }
        }
        }
/*
        for ( i = 1; i < s; i++ )
        for ( k = 1; k < s; k++ )
        for ( j = 1; j < s; j++ )
        if ( i + k + j == s )
        {
            vOne   = Vec_VecEntryInt( vFuncs, i );
            vTwo   = Vec_VecEntryInt( vFuncs, k );
            vThree = Vec_VecEntryInt( vFuncs, j );
            Vec_IntForEachEntry( vOne,   pEntries[0], e0 )
            Vec_IntForEachEntry( vTwo,   pEntries[1], e1 )
            Vec_IntForEachEntry( vThree, pEntries[2], e2 )
            {
                int fAddInv0 = !Dss_ObjCheckTransparent( p, Dss_VecObj(p->vObjs, Abc_Lit2Var(pEntries[0])) );
                int fAddInv1 = !Dss_ObjCheckTransparent( p, Dss_VecObj(p->vObjs, Abc_Lit2Var(pEntries[1])) );
                int fAddInv2 = !Dss_ObjCheckTransparent( p, Dss_VecObj(p->vObjs, Abc_Lit2Var(pEntries[2])) );

                if ( !fAddInv0 && k > j )
                    continue;

                iLit = Dss_ManOperation( p, DAU_DSD_MUX, pEntries, 3, NULL, NULL );
                assert( !Abc_LitIsCompl(iLit) );
                Vec_IntPush( vRes, iLit );

                if ( fAddInv1 )
                {
                    pEntries[1] = Abc_LitNot( pEntries[1] );
                    iLit = Dss_ManOperation( p, DAU_DSD_MUX, pEntries, 3, NULL, NULL );
                    pEntries[1] = Abc_LitNot( pEntries[1] );
                    assert( !Abc_LitIsCompl(iLit) );
                    Vec_IntPush( vRes, iLit );
                }

                if ( fAddInv2 )
                {
                    pEntries[2] = Abc_LitNot( pEntries[2] );
                    iLit = Dss_ManOperation( p, DAU_DSD_MUX, pEntries, 3, NULL, NULL );
                    pEntries[2] = Abc_LitNot( pEntries[2] );
                    assert( !Abc_LitIsCompl(iLit) );
                    Vec_IntPush( vRes, iLit );
                }
            }
        }
*/
        Vec_IntUniqify( vRes );
    }
    Dss_ManPrint( "_npn/npn/dsdcanon.txt", p );

    Dss_ManFree( p );
    Vec_VecFree( vFuncs );
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dau_DsdTest444() {
    var p *Dss_Man_t  = Dss_ManAlloc( 6, 0 );
    var iLit1 [3]int32 = [3]int32{ 2, 4 };
    var iLit2 [3]int32 = [3]int32{ 2, 4, 6 };
    var iRes [5]int32;
    var nFans [2]int32 = [2]int32{ 4, 3 };
    var pPermLits1 [4]int32 = [4]int32{ 0, 2, 5, 6 };
    var pPermLits2 [5]int32 = [5]int32{ 2, 9, 10 };
    var pPermLits [2]*int32 = [2]*int32{ pPermLits1, pPermLits2 };
    unsigned char pPermRes[6];
    var pPermResInt [6]int32;
    var uMaskShared uint32  = 2;
    var i int32;

    iRes[0] = 1 ^ Dss_ManOperation( p, DAU_DSD_AND, iLit1, 2, nil, nil );
    iRes[1] = iRes[0]; 
    iRes[2] = 1 ^ Dss_ManOperation( p, DAU_DSD_AND, iRes, 2, nil, nil );
    iRes[3] = Dss_ManOperation( p, DAU_DSD_AND, iLit2, 3, nil, nil );

    Dss_ManPrintOne( os.Stdout, p, iRes[0], nil );
    Dss_ManPrintOne( os.Stdout, p, iRes[2], nil );
    Dss_ManPrintOne( os.Stdout, p, iRes[3], nil );

    Dss_ManPrintOne( os.Stdout, p, iRes[2], pPermLits1 );
    Dss_ManPrintOne( os.Stdout, p, iRes[3], pPermLits2 );

    iRes[4] = Dss_ManMerge( p, iRes+2, nFans, pPermLits, uMaskShared, 6, pPermRes, nil );

    for  i = 0; i < 6; i++  {
        pPermResInt[i] = pPermRes[i];
    }

    Dss_ManPrintOne( os.Stdout, p, iRes[4], nil );
    Dss_ManPrintOne( os.Stdout, p, iRes[4], pPermResInt );

    Dss_ManFree( p );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

