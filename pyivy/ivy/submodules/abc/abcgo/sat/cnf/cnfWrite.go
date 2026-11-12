
package cnf

/**CFile****************************************************************

  FileName    [cnfWrite.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [AIG-to-CNF conversion.]

  Synopsis    []

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - April 28, 2007.]

  Revision    [$Id: cnfWrite.c,v 1.00 2007/04/28 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "cnf.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Derives CNF mapping.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cnf_ManWriteCnfMapping(p *Cnf_Man_t, vMapped *Vec_Ptr_t) *Vec_Int_t {
    var vResult *Vec_Int_t;
    var pObj *Aig_Obj_t;
    var pCut *Cnf_Cut_t;
    var i, k, nOffset int32;
    nOffset = Aig_ManObjNumMax(p.pManAig);
    vResult = Vec_IntStart( nOffset );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vMapped)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(vMapped, i)), 1); i++  {
        assert(Aig_ObjIsNode(pObj), "Aig_ObjIsNode(pObj)");
        pCut = Cnf_ObjBestCut( pObj );
        assert(pCut.nFanins < 5, "pCut.nFanins < 5");
        Vec_IntWriteEntry( vResult, Aig_ObjId(pObj), nOffset );
        Vec_IntPush( vResult, *Cnf_CutTruth(pCut) );
        for  k = 0; k < pCut.nFanins; k++  {
            Vec_IntPush( vResult, pCut.pFanins[k] );
        }
        for       ; k < 4; k++  {
            Vec_IntPush( vResult, -1 );
        }
        nOffset += 5;
    }
    return vResult;
}



/**Function*************************************************************

  Synopsis    [Writes the cover into the array.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cnf_SopConvertToVector(pSop string, nCubes int32, vCover *Vec_Int_t) {
    var Lits [4]int32;
    var Cube int32;
    var iCube int32;
    var i int32;
    var b int32;
    Vec_IntClear( vCover );
    for  i = 0; i < nCubes; i++  {
        Cube = pSop[i];
        for  b = 0; b < 4; b++  {
            if Cube % 3 == 0 {
                Lits[b] = 1;
            } else if Cube % 3 == 1 {
                Lits[b] = 2;
            } else {
                Lits[b] = 0; }
            Cube = Cube / 3;
        }
        iCube = 0;
        for  b = 0; b < 4; b++  {
            iCube = (iCube << 2) | Lits[b];
        }
        Vec_IntPush( vCover, iCube );
    }
}

/**Function*************************************************************

  Synopsis    [Returns the number of literals in the SOP.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cnf_SopCountLiterals(pSop string, nCubes int32) int32 {
    var nLits, Cube, i, b int32;
    for  i = 0; i < nCubes; i++  {
        Cube = pSop[i];
        for  b = 0; b < 4; b++  {
            if Cube % 3 != 2 {
                nLits++; }
            Cube = Cube / 3;
        }
    }
    return nLits;
}

/**Function*************************************************************

  Synopsis    [Returns the number of literals in the SOP.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cnf_IsopCountLiterals(vIsop *Vec_Int_t, nVars int32) int32 {
    var nLits, Cube, i, b int32;
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vIsop)) && (((Cube) = Vec_IntEntry(vIsop, i)), 1); i++  {
        for  b = 0; b < nVars; b++  {
            if (Cube & 3) == 1 || (Cube & 3) == 2 {
                nLits++; }
            Cube >>= 2;
        }
    }
    return nLits;
}

/**Function*************************************************************

  Synopsis    [Writes the cube and returns the number of literals in it.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cnf_IsopWriteCube(Cube int32, nVars int32, pVars *int32, pLiterals *int32) int32 {
    var nLits int32 = nVars;
    var b int32;
    for  b = 0; b < nVars; b++  {
        if (Cube & 3) == 1 { // value 0 --> write positive literal
            *pLiterals++ = 2 * pVars[b];
        } else if (Cube & 3) == 2 { // value 1 --> write negative literal
            *pLiterals++ = 2 * pVars[b] + 1;
        } else {
            nLits--; }
        Cube >>= 2;
    }
    return nLits;
}

/**Function*************************************************************

  Synopsis    [Derives CNF for the mapping.]

  Description [The last argument shows the number of last outputs
  of the manager, which will not be converted into clauses but the
  new variables for which will be introduced.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cnf_ManWriteCnf(p *Cnf_Man_t, vMapped *Vec_Ptr_t, nOutputs int32) *Cnf_Dat_t {
    var fChangeVariableOrder int32; // should be set to 0 to improve performance
    var pObj *Aig_Obj_t;
    var pCnf *Cnf_Dat_t;
    var pCut *Cnf_Cut_t;
    var vCover, vSopTemp *Vec_Int_t;
    var OutVar int32;
    var PoVar int32;
    var pVars [32]int32;
    var pLits *int32;
    var pClas **int32;
    var uTruth uint32;
    var i, k, nLiterals, nClauses, Cube, Number int32;

    // count the number of literals and clauses
    nLiterals = 1 + Aig_ManCoNum( p.pManAig ) + 3 * nOutputs;
    nClauses = 1 + Aig_ManCoNum( p.pManAig ) + nOutputs;
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vMapped)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(vMapped, i)), 1); i++  {
        assert(Aig_ObjIsNode(pObj), "Aig_ObjIsNode(pObj)");
        pCut = Cnf_ObjBestCut( pObj );

        // positive polarity of the cut
        if pCut.nFanins < 5 {
            uTruth = 0xFFFF & *Cnf_CutTruth(pCut);
            nLiterals += Cnf_SopCountLiterals( p.pSops[uTruth], p.pSopSizes[uTruth] ) + p.pSopSizes[uTruth];
            assert(p.pSopSizes[uTruth] >= 0, "p.pSopSizes[uTruth] >= 0");
            nClauses += p.pSopSizes[uTruth];
        
        } else {
            nLiterals += Cnf_IsopCountLiterals( pCut.vIsop[1], pCut.nFanins ) + Vec_IntSize(pCut.vIsop[1]);
            nClauses += Vec_IntSize(pCut.vIsop[1]);
        }
        // negative polarity of the cut
        if pCut.nFanins < 5 {
            uTruth = 0xFFFF & ~*Cnf_CutTruth(pCut);
            nLiterals += Cnf_SopCountLiterals( p.pSops[uTruth], p.pSopSizes[uTruth] ) + p.pSopSizes[uTruth];
            assert(p.pSopSizes[uTruth] >= 0, "p.pSopSizes[uTruth] >= 0");
            nClauses += p.pSopSizes[uTruth];
        
        } else {
            nLiterals += Cnf_IsopCountLiterals( pCut.vIsop[0], pCut.nFanins ) + Vec_IntSize(pCut.vIsop[0]);
            nClauses += Vec_IntSize(pCut.vIsop[0]);
        }
//printf( "%d ", nClauses-(1 + Aig_ManCoNum( p->pManAig )) );
    }
//printf( "\n" );

    // allocate CNF
    // inlining of macro ABC_CALLOC
    pCnf = &Cnf_Dat_t{};
    pCnf.pMan = p.pManAig;
    pCnf.nLiterals = nLiterals;
    pCnf.nClauses = nClauses;
    // inlining of macro ABC_ALLOC
    pCnf.pClauses = ((int * *) malloc(sizeof(int *) * (size_t)(nClauses + 1)));
    // inlining of macro ABC_ALLOC
    pCnf.pClauses[0] = ((int *) malloc(sizeof(int) * (size_t)(nLiterals)));
    pCnf.pClauses[nClauses] = pCnf.pClauses[0] + nLiterals;
    // create room for variable numbers
    // inlining of macro ABC_ALLOC
    pCnf.pVarNums = ((int *) malloc(sizeof(int) * (size_t)(Aig_ManObjNumMax(p.pManAig))));
//    memset( pCnf->pVarNums, 0xff, sizeof(int) * Aig_ManObjNumMax(p->pManAig) );
    for  i = 0; i < Aig_ManObjNumMax(p.pManAig); i++  {
        pCnf.pVarNums[i] = -1;
    }

    if !(fChangeVariableOrder != 0) {
        // assign variables to the last (nOutputs) POs
        Number = 1;
        if nOutputs != 0 {
            if Aig_ManRegNum(p.pManAig) == 0 {
                assert(nOutputs == Aig_ManCoNum(p.pManAig), "nOutputs == Aig_ManCoNum(p.pManAig)");
                // inlining of macro Aig_ManForEachCo
                // inlining of macro Vec_PtrForEachEntry
                for  i = 0; (i < Vec_PtrSize(p.pManAig.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pManAig.vCos, i)), 1); i++  {
                    pCnf.pVarNums[pObj.Id] = Number++;
                }
            
            } else {
                assert(nOutputs == Aig_ManRegNum(p.pManAig), "nOutputs == Aig_ManRegNum(p.pManAig)");
                // inlining of macro Aig_ManForEachLiSeq
                // inlining of macro Vec_PtrForEachEntryStart
                for  i = Aig_ManCoNum(p.pManAig)-Aig_ManRegNum(p.pManAig); (i < Vec_PtrSize(p.pManAig.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pManAig.vCos, i)), 1); i++  {
                    pCnf.pVarNums[pObj.Id] = Number++;
                }
            }
        }
        // assign variables to the internal nodes
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(vMapped)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(vMapped, i)), 1); i++  {
            pCnf.pVarNums[pObj.Id] = Number++;
        }
        // assign variables to the PIs and constant node
        // inlining of macro Aig_ManForEachCi
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(p.pManAig.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pManAig.vCis, i)), 1); i++  {
            pCnf.pVarNums[pObj.Id] = Number++;
        }
        pCnf.pVarNums[Aig_ManConst1(p.pManAig).Id] = Number++;
        pCnf.nVars = Number;
    
    } else {
        // assign variables to the last (nOutputs) POs
        Number = Aig_ManObjNumMax(p.pManAig) + 1;
        pCnf.nVars = Number + 1;
        if nOutputs != 0 {
            if Aig_ManRegNum(p.pManAig) == 0 {
                assert(nOutputs == Aig_ManCoNum(p.pManAig), "nOutputs == Aig_ManCoNum(p.pManAig)");
                // inlining of macro Aig_ManForEachCo
                // inlining of macro Vec_PtrForEachEntry
                for  i = 0; (i < Vec_PtrSize(p.pManAig.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pManAig.vCos, i)), 1); i++  {
                    pCnf.pVarNums[pObj.Id] = Number--;
                }
            
            } else {
                assert(nOutputs == Aig_ManRegNum(p.pManAig), "nOutputs == Aig_ManRegNum(p.pManAig)");
                // inlining of macro Aig_ManForEachLiSeq
                // inlining of macro Vec_PtrForEachEntryStart
                for  i = Aig_ManCoNum(p.pManAig)-Aig_ManRegNum(p.pManAig); (i < Vec_PtrSize(p.pManAig.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pManAig.vCos, i)), 1); i++  {
                    pCnf.pVarNums[pObj.Id] = Number--;
                }
            }
        }
        // assign variables to the internal nodes
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(vMapped)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(vMapped, i)), 1); i++  {
            pCnf.pVarNums[pObj.Id] = Number--;
        }
        // assign variables to the PIs and constant node
        // inlining of macro Aig_ManForEachCi
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(p.pManAig.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pManAig.vCis, i)), 1); i++  {
            pCnf.pVarNums[pObj.Id] = Number--;
        }
        pCnf.pVarNums[Aig_ManConst1(p.pManAig).Id] = Number--;
        assert(Number >= 0, "Number >= 0");
    }

    // assign the clauses
    vSopTemp = Vec_IntAlloc( 1 << 16 );
    pLits = pCnf.pClauses[0];
    pClas = pCnf.pClauses;
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vMapped)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(vMapped, i)), 1); i++  {
        pCut = Cnf_ObjBestCut( pObj );

        // save variables of this cut
        OutVar = pCnf.pVarNums[ pObj.Id ];
        for  k = 0; k < int32(pCut.nFanins); k++  {
            pVars[k] = pCnf.pVarNums[ pCut.pFanins[k] ];
            assert(pVars[k] <= Aig_ManObjNumMax(p.pManAig), "pVars[k] <= Aig_ManObjNumMax(p.pManAig)");
        }

        // positive polarity of the cut
        if pCut.nFanins < 5 {
            uTruth = 0xFFFF & *Cnf_CutTruth(pCut);
            Cnf_SopConvertToVector( p.pSops[uTruth], p.pSopSizes[uTruth], vSopTemp );
            vCover = vSopTemp;
        
        } else {
            vCover = pCut.vIsop[1]; }
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(vCover)) && (((Cube) = Vec_IntEntry(vCover, k)), 1); k++  {
            *pClas++ = pLits;
            *pLits++ = 2 * OutVar; 
            pLits += Cnf_IsopWriteCube( Cube, pCut.nFanins, pVars, pLits );
        }

        // negative polarity of the cut
        if pCut.nFanins < 5 {
            uTruth = 0xFFFF & ~*Cnf_CutTruth(pCut);
            Cnf_SopConvertToVector( p.pSops[uTruth], p.pSopSizes[uTruth], vSopTemp );
            vCover = vSopTemp;
        
        } else {
            vCover = pCut.vIsop[0]; }
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(vCover)) && (((Cube) = Vec_IntEntry(vCover, k)), 1); k++  {
            *pClas++ = pLits;
            *pLits++ = 2 * OutVar + 1; 
            pLits += Cnf_IsopWriteCube( Cube, pCut.nFanins, pVars, pLits );
        }
    }
    Vec_IntFree( vSopTemp );
 
    // write the constant literal
    OutVar = pCnf.pVarNums[ Aig_ManConst1(p.pManAig).Id ];
    assert(OutVar <= Aig_ManObjNumMax(p.pManAig), "OutVar <= Aig_ManObjNumMax(p.pManAig)");
    *pClas++ = pLits;
    *pLits++ = 2 * OutVar; 

    // write the output literals
    // inlining of macro Aig_ManForEachCo
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pManAig.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pManAig.vCos, i)), 1); i++  {
        OutVar = pCnf.pVarNums[ Aig_ObjFanin0(pObj).Id ];
        if i < Aig_ManCoNum(p.pManAig) - nOutputs {
            *pClas++ = pLits;
            *pLits++ = 2 * OutVar + Aig_ObjFaninC0(pObj); 
        
        } else {
            PoVar = pCnf.pVarNums[ pObj.Id ];
            // first clause
            *pClas++ = pLits;
            *pLits++ = 2 * PoVar; 
            *pLits++ = 2 * OutVar + !Aig_ObjFaninC0(pObj); 
            // second clause
            *pClas++ = pLits;
            *pLits++ = 2 * PoVar + 1; 
            *pLits++ = 2 * OutVar + Aig_ObjFaninC0(pObj); 
        }
    }

    // verify that the correct number of literals and clauses was written
    assert(pLits - pCnf.pClauses[0] == nLiterals, "pLits - pCnf.pClauses[0] == nLiterals");
    assert(pClas - pCnf.pClauses == nClauses, "pClas - pCnf.pClauses == nClauses");
//Cnf_DataPrint( pCnf, 1 );
    return pCnf;
}


/**Function*************************************************************

  Synopsis    [Derives CNF for the mapping.]

  Description [Derives CNF with obj IDs as SAT vars and mapping of
  objects into clauses (pObj2Clause and pObj2Count).]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cnf_ManWriteCnfOther(p *Cnf_Man_t, vMapped *Vec_Ptr_t) *Cnf_Dat_t {
    var pObj *Aig_Obj_t;
    var pCnf *Cnf_Dat_t;
    var pCut *Cnf_Cut_t;
    var vCover, vSopTemp *Vec_Int_t;
    var OutVar int32;
    var PoVar int32;
    var pVars [32]int32;
    var pLits *int32;
    var pClas **int32;
    var uTruth uint32;
    var i, k, nLiterals, nClauses, Cube int32;

    // count the number of literals and clauses
    nLiterals = 1 + 4 * Aig_ManCoNum( p.pManAig );
    nClauses  = 1 + 2 * Aig_ManCoNum( p.pManAig );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vMapped)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(vMapped, i)), 1); i++  {
        assert(Aig_ObjIsNode(pObj), "Aig_ObjIsNode(pObj)");
        pCut = Cnf_ObjBestCut( pObj );
        // positive polarity of the cut
        if pCut.nFanins < 5 {
            uTruth = 0xFFFF & *Cnf_CutTruth(pCut);
            nLiterals += Cnf_SopCountLiterals( p.pSops[uTruth], p.pSopSizes[uTruth] ) + p.pSopSizes[uTruth];
            assert(p.pSopSizes[uTruth] >= 0, "p.pSopSizes[uTruth] >= 0");
            nClauses += p.pSopSizes[uTruth];
        
        } else {
            nLiterals += Cnf_IsopCountLiterals( pCut.vIsop[1], pCut.nFanins ) + Vec_IntSize(pCut.vIsop[1]);
            nClauses += Vec_IntSize(pCut.vIsop[1]);
        }
        // negative polarity of the cut
        if pCut.nFanins < 5 {
            uTruth = 0xFFFF & ~*Cnf_CutTruth(pCut);
            nLiterals += Cnf_SopCountLiterals( p.pSops[uTruth], p.pSopSizes[uTruth] ) + p.pSopSizes[uTruth];
            assert(p.pSopSizes[uTruth] >= 0, "p.pSopSizes[uTruth] >= 0");
            nClauses += p.pSopSizes[uTruth];
        
        } else {
            nLiterals += Cnf_IsopCountLiterals( pCut.vIsop[0], pCut.nFanins ) + Vec_IntSize(pCut.vIsop[0]);
            nClauses += Vec_IntSize(pCut.vIsop[0]);
        }
    }

    // allocate CNF
    // inlining of macro ABC_CALLOC
    pCnf = &Cnf_Dat_t{};
    pCnf.pMan = p.pManAig;
    pCnf.nLiterals = nLiterals;
    pCnf.nClauses  = nClauses;
    // inlining of macro ABC_ALLOC
    pCnf.pClauses  = ((int * *) malloc(sizeof(int *) * (size_t)(nClauses + 1)));
    // inlining of macro ABC_ALLOC
    pCnf.pClauses[0] = ((int *) malloc(sizeof(int) * (size_t)(nLiterals)));
    pCnf.pClauses[nClauses] = pCnf.pClauses[0] + nLiterals;
    // create room for variable numbers
    // inlining of macro ABC_ALLOC
    pCnf.pObj2Clause = ((int *) malloc(sizeof(int) * (size_t)(Aig_ManObjNumMax(p.pManAig))));
    // inlining of macro ABC_ALLOC
    pCnf.pObj2Count  = ((int *) malloc(sizeof(int) * (size_t)(Aig_ManObjNumMax(p.pManAig))));
    for  i = 0; i < Aig_ManObjNumMax(p.pManAig); i++  {
        pCnf.pObj2Clause[i] = pCnf.pObj2Count[i] = -1;
    }
    pCnf.nVars = Aig_ManObjNumMax(p.pManAig);

    // clear the PI counters
    // inlining of macro Aig_ManForEachCi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pManAig.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pManAig.vCis, i)), 1); i++  {
        pCnf.pObj2Count[pObj.Id] = 0;
    }

    // assign the clauses
    vSopTemp = Vec_IntAlloc( 1 << 16 );
    pLits = pCnf.pClauses[0];
    pClas = pCnf.pClauses;
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vMapped)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(vMapped, i)), 1); i++  {
        // remember the starting clause
        pCnf.pObj2Clause[pObj.Id] = pClas - pCnf.pClauses;
        pCnf.pObj2Count[pObj.Id] = 0;

        // get the best cut
        pCut = Cnf_ObjBestCut( pObj );
        // save variables of this cut
        OutVar = pObj.Id;
        for  k = 0; k < int32(pCut.nFanins); k++  {
            pVars[k] = pCut.pFanins[k];
            assert(pVars[k] <= Aig_ManObjNumMax(p.pManAig), "pVars[k] <= Aig_ManObjNumMax(p.pManAig)");
        }

        // positive polarity of the cut
        if pCut.nFanins < 5 {
            uTruth = 0xFFFF & *Cnf_CutTruth(pCut);
            Cnf_SopConvertToVector( p.pSops[uTruth], p.pSopSizes[uTruth], vSopTemp );
            vCover = vSopTemp;
        
        } else {
            vCover = pCut.vIsop[1]; }
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(vCover)) && (((Cube) = Vec_IntEntry(vCover, k)), 1); k++  {
            *pClas++ = pLits;
            *pLits++ = 2 * OutVar; 
            pLits += Cnf_IsopWriteCube( Cube, pCut.nFanins, pVars, pLits );
        }
        pCnf.pObj2Count[pObj.Id] += Vec_IntSize(vCover);

        // negative polarity of the cut
        if pCut.nFanins < 5 {
            uTruth = 0xFFFF & ~*Cnf_CutTruth(pCut);
            Cnf_SopConvertToVector( p.pSops[uTruth], p.pSopSizes[uTruth], vSopTemp );
            vCover = vSopTemp;
        
        } else {
            vCover = pCut.vIsop[0]; }
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(vCover)) && (((Cube) = Vec_IntEntry(vCover, k)), 1); k++  {
            *pClas++ = pLits;
            *pLits++ = 2 * OutVar + 1; 
            pLits += Cnf_IsopWriteCube( Cube, pCut.nFanins, pVars, pLits );
        }
        pCnf.pObj2Count[pObj.Id] += Vec_IntSize(vCover);
    }
    Vec_IntFree( vSopTemp );

    // write the output literals
    // inlining of macro Aig_ManForEachCo
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pManAig.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pManAig.vCos, i)), 1); i++  {
        // remember the starting clause
        pCnf.pObj2Clause[pObj.Id] = pClas - pCnf.pClauses;
        pCnf.pObj2Count[pObj.Id] = 2;
        // get variables
        OutVar = Aig_ObjFanin0(pObj).Id;
        PoVar = pObj.Id;
        // first clause
        *pClas++ = pLits;
        *pLits++ = 2 * PoVar; 
        *pLits++ = 2 * OutVar + !Aig_ObjFaninC0(pObj); 
        // second clause
        *pClas++ = pLits;
        *pLits++ = 2 * PoVar + 1; 
        *pLits++ = 2 * OutVar + Aig_ObjFaninC0(pObj); 
    }
 
    // remember the starting clause
    pCnf.pObj2Clause[Aig_ManConst1(p.pManAig).Id] = pClas - pCnf.pClauses;
    pCnf.pObj2Count[Aig_ManConst1(p.pManAig).Id] = 1;
    // write the constant literal
    OutVar = Aig_ManConst1(p.pManAig).Id;
    *pClas++ = pLits;
    *pLits++ = 2 * OutVar; 

    // verify that the correct number of literals and clauses was written
    assert(pLits - pCnf.pClauses[0] == nLiterals, "pLits - pCnf.pClauses[0] == nLiterals");
    assert(pClas - pCnf.pClauses == nClauses, "pClas - pCnf.pClauses == nClauses");
//Cnf_DataPrint( pCnf, 1 );
    return pCnf;
}


/**Function*************************************************************

  Synopsis    [Derives a simple CNF for the AIG.]

  Description [The last argument lists the number of last outputs
  of the manager, which will not be converted into clauses. 
  New variables will be introduced for these outputs.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cnf_DeriveSimple(p *Aig_Man_t, nOutputs int32) *Cnf_Dat_t {
    var pObj *Aig_Obj_t;
    var pCnf *Cnf_Dat_t;
    var OutVar int32;
    var PoVar int32;
    var pVars [32]int32;
    var pLits *int32;
    var pClas **int32;
    var i, nLiterals, nClauses, Number int32;

    // count the number of literals and clauses
    nLiterals = 1 + 7 * Aig_ManNodeNum(p) + Aig_ManCoNum( p ) + 3 * nOutputs;
    nClauses = 1 + 3 * Aig_ManNodeNum(p) + Aig_ManCoNum( p ) + nOutputs;

    // allocate CNF
    // inlining of macro ABC_ALLOC
    pCnf = ((Cnf_Dat_t *) malloc(sizeof(Cnf_Dat_t) * (size_t)(1)));
    memset( pCnf, 0, sizeof(Cnf_Dat_t) );
    pCnf.pMan = p;
    pCnf.nLiterals = nLiterals;
    pCnf.nClauses = nClauses;
    // inlining of macro ABC_ALLOC
    pCnf.pClauses = ((int * *) malloc(sizeof(int *) * (size_t)(nClauses + 1)));
    // inlining of macro ABC_ALLOC
    pCnf.pClauses[0] = ((int *) malloc(sizeof(int) * (size_t)(nLiterals)));
    pCnf.pClauses[nClauses] = pCnf.pClauses[0] + nLiterals;

    // create room for variable numbers
    // inlining of macro ABC_ALLOC
    pCnf.pVarNums = ((int *) malloc(sizeof(int) * (size_t)(Aig_ManObjNumMax(p))));
//    memset( pCnf->pVarNums, 0xff, sizeof(int) * Aig_ManObjNumMax(p) );
    for  i = 0; i < Aig_ManObjNumMax(p); i++  {
        pCnf.pVarNums[i] = -1;
    }
    // assign variables to the last (nOutputs) POs
    Number = 1;
    if nOutputs != 0 {
//        assert( nOutputs == Aig_ManRegNum(p) );
//        Aig_ManForEachLiSeq( p, pObj, i )
//            pCnf->pVarNums[pObj->Id] = Number++;
        // inlining of macro Aig_ManForEachCo
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(p.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
            pCnf.pVarNums[pObj.Id] = Number++;
        }
    }
    // assign variables to the internal nodes
    // inlining of macro Aig_ManForEachNode
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
        pCnf.pVarNums[pObj.Id] = Number++; }
    }
    // assign variables to the PIs and constant node
    // inlining of macro Aig_ManForEachCi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCis, i)), 1); i++  {
        pCnf.pVarNums[pObj.Id] = Number++;
    }
    pCnf.pVarNums[Aig_ManConst1(p).Id] = Number++;
    pCnf.nVars = Number;
/*
    // print CNF numbers
    printf( "SAT numbers of each node:\n" );
    Aig_ManForEachObj( p, pObj, i )
        printf( "%d=%d ", pObj->Id, pCnf->pVarNums[pObj->Id] );
    printf( "\n" );
*/
    // assign the clauses
    pLits = pCnf.pClauses[0];
    pClas = pCnf.pClauses;
    // inlining of macro Aig_ManForEachNode
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
        OutVar   = pCnf.pVarNums[ pObj.Id ];
        pVars[0] = pCnf.pVarNums[ Aig_ObjFanin0(pObj).Id ];
        pVars[1] = pCnf.pVarNums[ Aig_ObjFanin1(pObj).Id ];

        // positive phase
        *pClas++ = pLits;
        *pLits++ = 2 * OutVar; 
        *pLits++ = 2 * pVars[0] + !Aig_ObjFaninC0(pObj); 
        *pLits++ = 2 * pVars[1] + !Aig_ObjFaninC1(pObj); 
        // negative phase
        *pClas++ = pLits;
        *pLits++ = 2 * OutVar + 1; 
        *pLits++ = 2 * pVars[0] + Aig_ObjFaninC0(pObj); 
        *pClas++ = pLits;
        *pLits++ = 2 * OutVar + 1; 
        *pLits++ = 2 * pVars[1] + Aig_ObjFaninC1(pObj); 
    }
    }
 
    // write the constant literal
    OutVar = pCnf.pVarNums[ Aig_ManConst1(p).Id ];
    assert(OutVar <= Aig_ManObjNumMax(p), "OutVar <= Aig_ManObjNumMax(p)");
    *pClas++ = pLits;
    *pLits++ = 2 * OutVar;  

    // write the output literals
    // inlining of macro Aig_ManForEachCo
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
        OutVar = pCnf.pVarNums[ Aig_ObjFanin0(pObj).Id ];
        if i < Aig_ManCoNum(p) - nOutputs {
            *pClas++ = pLits;
            *pLits++ = 2 * OutVar + Aig_ObjFaninC0(pObj); 
        
        } else {
            PoVar  = pCnf.pVarNums[ pObj.Id ];
            // first clause
            *pClas++ = pLits;
            *pLits++ = 2 * PoVar; 
            *pLits++ = 2 * OutVar + !Aig_ObjFaninC0(pObj); 
            // second clause
            *pClas++ = pLits;
            *pLits++ = 2 * PoVar + 1; 
            *pLits++ = 2 * OutVar + Aig_ObjFaninC0(pObj); 
        }
    }

    // verify that the correct number of literals and clauses was written
    assert(pLits - pCnf.pClauses[0] == nLiterals, "pLits - pCnf.pClauses[0] == nLiterals");
    assert(pClas - pCnf.pClauses == nClauses, "pClas - pCnf.pClauses == nClauses");
    return pCnf;
}

/**Function*************************************************************

  Synopsis    [Derives a simple CNF for backward retiming computation.]

  Description [The last argument shows the number of last outputs
  of the manager, which will not be converted into clauses. 
  New variables will be introduced for these outputs.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cnf_DeriveSimpleForRetiming(p *Aig_Man_t) *Cnf_Dat_t {
    var pObj *Aig_Obj_t;
    var pCnf *Cnf_Dat_t;
    var OutVar int32;
    var PoVar int32;
    var pVars [32]int32;
    var pLits *int32;
    var pClas **int32;
    var i, nLiterals, nClauses, Number int32;

    // count the number of literals and clauses
    nLiterals = 1 + 7 * Aig_ManNodeNum(p) + 5 * Aig_ManCoNum(p);
    nClauses = 1 + 3 * Aig_ManNodeNum(p) + 3 * Aig_ManCoNum(p);

    // allocate CNF
    // inlining of macro ABC_ALLOC
    pCnf = ((Cnf_Dat_t *) malloc(sizeof(Cnf_Dat_t) * (size_t)(1)));
    memset( pCnf, 0, sizeof(Cnf_Dat_t) );
    pCnf.pMan = p;
    pCnf.nLiterals = nLiterals;
    pCnf.nClauses = nClauses;
    // inlining of macro ABC_ALLOC
    pCnf.pClauses = ((int * *) malloc(sizeof(int *) * (size_t)(nClauses + 1)));
    // inlining of macro ABC_ALLOC
    pCnf.pClauses[0] = ((int *) malloc(sizeof(int) * (size_t)(nLiterals)));
    pCnf.pClauses[nClauses] = pCnf.pClauses[0] + nLiterals;

    // create room for variable numbers
    // inlining of macro ABC_ALLOC
    pCnf.pVarNums = ((int *) malloc(sizeof(int) * (size_t)(Aig_ManObjNumMax(p))));
//    memset( pCnf->pVarNums, 0xff, sizeof(int) * Aig_ManObjNumMax(p) );
    for  i = 0; i < Aig_ManObjNumMax(p); i++  {
        pCnf.pVarNums[i] = -1;
    }
    // assign variables to the last (nOutputs) POs
    Number = 1;
    // inlining of macro Aig_ManForEachCo
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
        pCnf.pVarNums[pObj.Id] = Number++;
    }
    // assign variables to the internal nodes
    // inlining of macro Aig_ManForEachNode
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
        pCnf.pVarNums[pObj.Id] = Number++; }
    }
    // assign variables to the PIs and constant node
    // inlining of macro Aig_ManForEachCi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCis, i)), 1); i++  {
        pCnf.pVarNums[pObj.Id] = Number++;
    }
    pCnf.pVarNums[Aig_ManConst1(p).Id] = Number++;
    pCnf.nVars = Number;
    // assign the clauses
    pLits = pCnf.pClauses[0];
    pClas = pCnf.pClauses;
    // inlining of macro Aig_ManForEachNode
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
        OutVar   = pCnf.pVarNums[ pObj.Id ];
        pVars[0] = pCnf.pVarNums[ Aig_ObjFanin0(pObj).Id ];
        pVars[1] = pCnf.pVarNums[ Aig_ObjFanin1(pObj).Id ];

        // positive phase
        *pClas++ = pLits;
        *pLits++ = 2 * OutVar; 
        *pLits++ = 2 * pVars[0] + !Aig_ObjFaninC0(pObj); 
        *pLits++ = 2 * pVars[1] + !Aig_ObjFaninC1(pObj); 
        // negative phase
        *pClas++ = pLits;
        *pLits++ = 2 * OutVar + 1; 
        *pLits++ = 2 * pVars[0] + Aig_ObjFaninC0(pObj); 
        *pClas++ = pLits;
        *pLits++ = 2 * OutVar + 1; 
        *pLits++ = 2 * pVars[1] + Aig_ObjFaninC1(pObj); 
    }
    }
 
    // write the constant literal
    OutVar = pCnf.pVarNums[ Aig_ManConst1(p).Id ];
    assert(OutVar <= Aig_ManObjNumMax(p), "OutVar <= Aig_ManObjNumMax(p)");
    *pClas++ = pLits;
    *pLits++ = 2 * OutVar; 

    // write the output literals
    // inlining of macro Aig_ManForEachCo
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
        OutVar = pCnf.pVarNums[ Aig_ObjFanin0(pObj).Id ];
        PoVar  = pCnf.pVarNums[ pObj.Id ];
        // first clause
        *pClas++ = pLits;
        *pLits++ = 2 * PoVar; 
        *pLits++ = 2 * OutVar + !Aig_ObjFaninC0(pObj); 
        // second clause
        *pClas++ = pLits;
        *pLits++ = 2 * PoVar + 1; 
        *pLits++ = 2 * OutVar + Aig_ObjFaninC0(pObj); 
        // final clause (init-state is always 0 -> set the output to 0)
        *pClas++ = pLits;
        *pLits++ = 2 * PoVar + 1; 
    }

    // verify that the correct number of literals and clauses was written
    assert(pLits - pCnf.pClauses[0] == nLiterals, "pLits - pCnf.pClauses[0] == nLiterals");
    assert(pClas - pCnf.pClauses == nClauses, "pClas - pCnf.pClauses == nClauses");
    return pCnf;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

