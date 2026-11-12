package abc

import (
	"fmt"
	"os"
)

/**CFile****************************************************************

  FileName    [abcBlifMv.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Network and node package.]

  Synopsis    [Procedures to process BLIF-MV networks and AIGs.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: abcBlifMv.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "abc.h"

#ifdef ABC_USE_CUDD
#include "bdd/extrab/extraBdd.h"
#endif

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////
 
////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Starts the Mv-Var manager.]

  Description []
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkStartMvVars(pNtk *Abc_Ntk_t) {
    var pAttMan *Vec_Att_t;
    assert(Abc_NtkMvVar(pNtk) == nil, "Abc_NtkMvVar(pNtk) == nil");
    pAttMan = Vec_AttAlloc( Abc_NtkObjNumMax(pNtk) + 1, Mem_FlexStart(), (void(*)(void*))Mem_FlexStop, nil, nil );
    Vec_PtrWriteEntry( pNtk.vAttrs, VEC_ATTR_MVVAR, pAttMan );
//printf( "allocing attr\n" );
}

/**Function*************************************************************

  Synopsis    [Stops the Mv-Var manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkFreeMvVars(pNtk *Abc_Ntk_t) { 
    var pUserMan *Mem_Flex_t;
    pUserMan = (Mem_Flex_t *)Abc_NtkAttrFree( pNtk, VEC_ATTR_GLOBAL_BDD, 0 ); 
    Mem_FlexStop( pUserMan, 0 );
}

/**Function*************************************************************

  Synopsis    [Duplicate the MV variable.]

  Description []
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkSetMvVarValues(pObj *Abc_Obj_t, nValues int32) {
    var pFlex *Mem_Flex_t;
    struct temp 
    { 
        var nValues int32; 
        var pNames *string; 
    } * pVarStruct;
    assert(nValues > 1, "nValues > 1");
    // skip binary signals
    if nValues == 2 {
        return; }
    // skip already assigned signals
    if Abc_ObjMvVar(pObj) != nil {
        return; }
    // create the structure
    pFlex = (Mem_Flex_t *)Abc_NtkMvVarMan( pObj.pNtk );
    pVarStruct = (struct temp *)Mem_FlexEntryFetch( pFlex, sizeof(struct temp) );
    pVarStruct.nValues = nValues;
    pVarStruct.pNames = nil;
    Abc_ObjSetMvVar( pObj, pVarStruct );
}

/**Function*************************************************************

  Synopsis    [Strashes the BLIF-MV netlist.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_StringGetNumber(ppStr *string) int32 {
    var pStr string  = *ppStr;
    var Number int32;
    assert(*pStr >= '0' && *pStr <= '9', "*pStr >= '0' && *pStr <= '9'");
    for  ; *pStr >= '0' && *pStr <= '9'; pStr++  {
        Number = 10 * Number + *pStr - '0';
    }
    *ppStr = pStr;
    return Number;
}

/**Function*************************************************************

  Synopsis    [Strashes one node in the BLIF-MV netlist.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NodeStrashBlifMv(pNtkNew *Abc_Ntk_t, pObj *Abc_Obj_t) int32 {
    var fAddFreeVars int32  = 1;
    var pSop string;
    var pValues, pValuesF, pValuesF2 **Abc_Obj_t;
    var pTemp, pTemp2, pFanin, pFanin2, pNet *Abc_Obj_t;
    var k, v, Def, DefIndex, Index, nValues, nValuesF, nValuesF2 int32;

    // start the output values
    assert(Abc_ObjIsNode(pObj), "Abc_ObjIsNode(pObj)");
    pNet = Abc_ObjFanout0(pObj);
    nValues = Abc_ObjMvVarNum(pNet);
    pValues = ABC_ALLOC( Abc_Obj_t *, nValues );
    for  k = 0; k < nValues; k++  {
        pValues[k] = Abc_ObjNot( Abc_AigConst1(pNtkNew) );
    }

    // get the BLIF-MV formula
    pSop = (char *)pObj.pData;
    // skip the value line
//    while ( *pSop++ != '\n' );

    // handle the constant
    if Abc_ObjFaninNum(pObj) == 0 {
        // skip the default if present
        if *pSop == 'd' {
            for  *pSop++ != '\n'  {; } }
        // skip space if present
        if *pSop == ' ' {
            pSop++; }
        // assume don't-care constant to be zero
        if *pSop == '-' {
            Index = 0;
        } else {
            Index = Abc_StringGetNumber( &pSop ); }
        assert(Index < nValues, "Index < nValues");
        ////////////////////////////////////////////
        // adding free variables for binary ND-constants
        if (fAddFreeVars != 0) && nValues == 2 && *pSop == '-' {
            pValues[1] = Abc_NtkCreatePi(pNtkNew);
            pValues[0] = Abc_ObjNot( pValues[1] );
            Abc_ObjAssignName( pValues[1], "free_var_", Abc_ObjName(pValues[1]) );
        
        } else {
            pValues[Index] = Abc_AigConst1(pNtkNew); }
        ////////////////////////////////////////////
        // save the values in the fanout net
        pNet.pCopy = (Abc_Obj_t *)pValues;
        return 1;
    }

    // parse the default line
    Def = DefIndex = -1;
    if *pSop == 'd' {
        pSop++;
        if *pSop == '=' {
            pSop++;
            DefIndex = Abc_StringGetNumber( &pSop );
            assert(DefIndex < Abc_ObjFaninNum(pObj), "DefIndex < Abc_ObjFaninNum(pObj)");
        
        } else if *pSop == '-' {
            pSop++;
            Def = 0;
        
        } else {
            Def = Abc_StringGetNumber( &pSop );
            assert(Def < nValues, "Def < nValues");
        }
        assert(*pSop == '\n', "*pSop == '\\n'");
        pSop++;
    }

    // convert the values
    for  *pSop  {
        // extract the values for each cube
        pTemp = Abc_AigConst1(pNtkNew); 
        Abc_ObjForEachFanin( pObj, pFanin, k )
        {
            if *pSop == '-' {
                pSop += 2;
                continue;
            }
            if *pSop == '!' {
                ABC_FREE( pValues );
                fmt.Printf("Abc_NodeStrashBlifMv(): Cannot handle complement in the MV function of node %s.\n", Abc_ObjName(Abc_ObjFanout0(pObj)));
                return 0;
            }
            if *pSop == '{' {
                ABC_FREE( pValues );
                fmt.Printf("Abc_NodeStrashBlifMv(): Cannot handle braces in the MV function of node %s.\n", Abc_ObjName(Abc_ObjFanout0(pObj)));
                return 0;
            }
            // get the value set
            nValuesF = Abc_ObjMvVarNum(pFanin);
            pValuesF = (Abc_Obj_t **)pFanin.pCopy;
            if *pSop == '(' {
                pSop++;
                pTemp2 = Abc_ObjNot( Abc_AigConst1(pNtkNew) );
                for  *pSop != ')'  {
                    Index = Abc_StringGetNumber( &pSop );
                    assert(Index < nValuesF, "Index < nValuesF");
                    pTemp2 = Abc_AigOr( (Abc_Aig_t *)pNtkNew.pManFunc, pTemp2, pValuesF[Index] );
                    assert(*pSop == ')' || *pSop == ',', "*pSop == ')' || *pSop == ','");
                    if *pSop == ',' {
                        pSop++; }
                }
                assert(*pSop == ')', "*pSop == ')'");
                pSop++;
            
            } else if *pSop == '=' {
                pSop++;
                // get the fanin index
                Index = Abc_StringGetNumber( &pSop );
                assert(Index < Abc_ObjFaninNum(pObj), "Index < Abc_ObjFaninNum(pObj)");
                assert(Index != k, "Index != k");
                // get the fanin
                pFanin2 = Abc_ObjFanin( pObj, Index );
                nValuesF2 = Abc_ObjMvVarNum(pFanin2);
                pValuesF2 = (Abc_Obj_t **)pFanin2.pCopy;
                // create the sum of products of values
                assert(nValuesF == nValuesF2, "nValuesF == nValuesF2");
                pTemp2 = Abc_ObjNot( Abc_AigConst1(pNtkNew) );
                for  v = 0; v < nValues; v++  {
                    pTemp2 = Abc_AigOr( (Abc_Aig_t *)pNtkNew.pManFunc, pTemp2, Abc_AigAnd((Abc_Aig_t *)pNtkNew.pManFunc, pValuesF[v], pValuesF2[v]) );
                }
            
            } else {
                Index = Abc_StringGetNumber( &pSop );
                assert(Index < nValuesF, "Index < nValuesF");
                pTemp2 = pValuesF[Index];
            }
            // compute the compute
            pTemp = Abc_AigAnd( (Abc_Aig_t *)pNtkNew.pManFunc, pTemp, pTemp2 );
            // advance the reading point
            assert(*pSop == ' ', "*pSop == ' '");
            pSop++;
        }
        // check if the output value is an equal construct
        if *pSop == '=' {
            pSop++;
            // get the output value
            Index = Abc_StringGetNumber( &pSop );
            assert(Index < Abc_ObjFaninNum(pObj), "Index < Abc_ObjFaninNum(pObj)");
            // add values of the given fanin with the given cube
            pFanin = Abc_ObjFanin( pObj, Index );
            nValuesF = Abc_ObjMvVarNum(pFanin);
            pValuesF = (Abc_Obj_t **)pFanin.pCopy;
            assert(nValuesF == nValues, "nValuesF == nValues"); // should be guaranteed by the parser
            for  k = 0; k < nValuesF; k++  {
                pValues[k] = Abc_AigOr( (Abc_Aig_t *)pNtkNew.pManFunc, pValues[k], Abc_AigAnd((Abc_Aig_t *)pNtkNew.pManFunc, pTemp, pValuesF[k]) );
            }
        
        } else {
            // get the output value
            Index = Abc_StringGetNumber( &pSop );
            assert(Index < nValues, "Index < nValues");
            pValues[Index] = Abc_AigOr( (Abc_Aig_t *)pNtkNew.pManFunc, pValues[Index], pTemp );
        }
        // advance the reading point
        assert(*pSop == '\n', "*pSop == '\\n'");
        pSop++;
    }

    // compute the default value
    if Def >= 0 || DefIndex >= 0 {
        pTemp = Abc_AigConst1(pNtkNew);
        for  k = 0; k < nValues; k++  {
            if k == Def {
                continue; }
            pTemp = Abc_AigAnd( (Abc_Aig_t *)pNtkNew.pManFunc, pTemp, Abc_ObjNot(pValues[k]) );
        }

        // assign the default value
        if Def >= 0 {
            pValues[Def] = pTemp;
        } else {
            assert(DefIndex >= 0, "DefIndex >= 0");
            // add values of the given fanin with the given cube
            pFanin = Abc_ObjFanin( pObj, DefIndex );
            nValuesF = Abc_ObjMvVarNum(pFanin);
            pValuesF = (Abc_Obj_t **)pFanin.pCopy;
            assert(nValuesF == nValues, "nValuesF == nValues"); // should be guaranteed by the parser
            for  k = 0; k < nValuesF; k++  {
                pValues[k] = Abc_AigOr( (Abc_Aig_t *)pNtkNew.pManFunc, pValues[k], Abc_AigAnd((Abc_Aig_t *)pNtkNew.pManFunc, pTemp, pValuesF[k]) );
            }
        }

    }

    // save the values in the fanout net
    pNet.pCopy = (Abc_Obj_t *)pValues;
    return 1;
}

/**Function*************************************************************

  Synopsis    [Assigns name with index.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkConvertAssignName(pObj *Abc_Obj_t, pNet *Abc_Obj_t, Index int32) {
    var Suffix string // was [16]byte in C

    assert(Abc_ObjIsTerm(pObj), "Abc_ObjIsTerm(pObj)");
    assert(Abc_ObjIsNet(pNet), "Abc_ObjIsNet(pNet)");
    Suffix = fmt.Sprintf("[%d]", Index);
    Abc_ObjAssignName( pObj, Abc_ObjName(pNet), Suffix );
}

/**Function*************************************************************

  Synopsis    [Strashes the BLIF-MV netlist.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkStrashBlifMv(pNtk *Abc_Ntk_t) *Abc_Ntk_t {
    var fUsePositional int32;
    var vNodes *Vec_Ptr_t;
    var pBits **Abc_Obj_t;
    var pValues **Abc_Obj_t;
    var pNtkNew *Abc_Ntk_t;
    var pObj, pTemp, pBit, pNet *Abc_Obj_t;
    var i, k, v, nValues, nValuesMax, nBits int32;
    var nCount1, nCount2 int32;

    assert(Abc_NtkIsNetlist(pNtk), "Abc_NtkIsNetlist(pNtk)");
    assert(Abc_NtkHasBlifMv(pNtk), "Abc_NtkHasBlifMv(pNtk)");
    assert(Abc_NtkWhiteboxNum(pNtk) == 0, "Abc_NtkWhiteboxNum(pNtk) == 0");
    assert(Abc_NtkBlackboxNum(pNtk) == 0, "Abc_NtkBlackboxNum(pNtk) == 0");

    // get the largest number of values
    nValuesMax = 2;
    Abc_NtkForEachNet( pNtk, pObj, i )
    {
        nValues = Abc_ObjMvVarNum(pObj);
        if nValuesMax < nValues {
            nValuesMax = nValues; }
    }
    nBits = Abc_Base2Log( nValuesMax );
    pBits = ABC_ALLOC( Abc_Obj_t *, nBits );

    // clean the node copy fields
    Abc_NtkCleanCopy( pNtk );
    // collect the nodes
    vNodes = Abc_NtkDfs( pNtk, 0 );

    // start the network
    pNtkNew = Abc_NtkAlloc( ABC_NTK_STRASH, ABC_FUNC_AIG, 1 );
    // duplicate the name and the spec
    pNtkNew.pName = Extra_UtilStrsav( pNtk.pName );
//    pNtkNew->pSpec = Extra_UtilStrsav( pNtk->pName );

    nCount1 = nCount2 = 0;
    // encode the CI nets
    Abc_NtkIncrementTravId( pNtk );
    if fUsePositional != 0 {
        Abc_NtkForEachCi( pNtk, pObj, i )
        {
            if !Abc_ObjIsPi(pObj) {
                continue; }
            pNet = Abc_ObjFanout0(pObj);
            nValues = Abc_ObjMvVarNum(pNet);
            pValues = ABC_ALLOC( Abc_Obj_t *, nValues );
            // create PIs for the values
            for  v = 0; v < nValues; v++  {
                pValues[v] = Abc_NtkCreatePi( pNtkNew );
                if nValuesMax == 2 {
                    Abc_ObjAssignName( pValues[v], Abc_ObjName(pNet), nil );
                } else {
                    Abc_NtkConvertAssignName( pValues[v], pNet, v ); }
            }
            // save the values in the fanout net
            pNet.pCopy = (Abc_Obj_t *)pValues;
            // mark the net
            Abc_NodeSetTravIdCurrent( pNet );
        }
        Abc_NtkForEachCi( pNtk, pObj, i )
        {
            if Abc_ObjIsPi(pObj) {
                continue; }
            pNet = Abc_ObjFanout0(pObj);
            nValues = Abc_ObjMvVarNum(pNet);
            pValues = ABC_ALLOC( Abc_Obj_t *, nValues );
            // create PIs for the values
            for  v = 0; v < nValues; v++  {
                pValues[v] = Abc_NtkCreateBo( pNtkNew );
                if nValuesMax == 2 {
                    Abc_ObjAssignName( pValues[v], Abc_ObjName(pNet), nil );
                } else {
                    Abc_NtkConvertAssignName( pValues[v], pNet, v ); }
                nCount1++;
            }
            // save the values in the fanout net
            pNet.pCopy = (Abc_Obj_t *)pValues;
            // mark the net
            Abc_NodeSetTravIdCurrent( pNet );
        }
    
    } else {
        Abc_NtkForEachCi( pNtk, pObj, i )
        {
            if !Abc_ObjIsPi(pObj) {
                continue; }
            pNet = Abc_ObjFanout0(pObj);
            nValues = Abc_ObjMvVarNum(pNet);
            pValues = ABC_ALLOC( Abc_Obj_t *, nValues );
            // create PIs for the encoding bits
            nBits = Abc_Base2Log( nValues );
            for  k = 0; k < nBits; k++  {
                pBits[k] = Abc_NtkCreatePi( pNtkNew );
                if nValuesMax == 2 {
                    Abc_ObjAssignName( pBits[k], Abc_ObjName(pNet), nil );
                } else {
                    Abc_NtkConvertAssignName( pBits[k], pNet, k ); }
            }
            // encode the values
            for  v = 0; v < nValues; v++  {
                pValues[v] = Abc_AigConst1(pNtkNew);
                for  k = 0; k < nBits; k++  {
                    pBit = Abc_ObjNotCond( pBits[k], (v&(1<<k)) == 0 );
                    pValues[v] = Abc_AigAnd( (Abc_Aig_t *)pNtkNew.pManFunc, pValues[v], pBit );
                }
            }
            // save the values in the fanout net
            pNet.pCopy = (Abc_Obj_t *)pValues;
            // mark the net
            Abc_NodeSetTravIdCurrent( pNet );
        }
        Abc_NtkForEachCi( pNtk, pObj, i )
        {
            if Abc_ObjIsPi(pObj) {
                continue; }
            pNet = Abc_ObjFanout0(pObj);
            nValues = Abc_ObjMvVarNum(pNet);
            pValues = ABC_ALLOC( Abc_Obj_t *, nValues );
            // create PIs for the encoding bits
            nBits = Abc_Base2Log( nValues );
            for  k = 0; k < nBits; k++  {
                pBits[k] = Abc_NtkCreateBo( pNtkNew );
                if nValuesMax == 2 {
                    Abc_ObjAssignName( pBits[k], Abc_ObjName(pNet), nil );
                } else {
                    Abc_NtkConvertAssignName( pBits[k], pNet, k ); }
                nCount1++;
            }
            // encode the values
            for  v = 0; v < nValues; v++  {
                pValues[v] = Abc_AigConst1(pNtkNew);
                for  k = 0; k < nBits; k++  {
                    pBit = Abc_ObjNotCond( pBits[k], (v&(1<<k)) == 0 );
                    pValues[v] = Abc_AigAnd( (Abc_Aig_t *)pNtkNew.pManFunc, pValues[v], pBit );
                }
            }
            // save the values in the fanout net
            pNet.pCopy = (Abc_Obj_t *)pValues;
            // mark the net
            Abc_NodeSetTravIdCurrent( pNet );
        }
    }

    // process nodes in the topological order
    Vec_PtrForEachEntry( Abc_Obj_t *, vNodes, pObj, i )
        if !(Abc_NodeStrashBlifMv( pNtkNew, pObj ) != 0) {
            Abc_NtkDelete( pNtkNew );
            return nil;
        }
    Vec_PtrFree( vNodes );

    // encode the CO nets
    if fUsePositional != 0 {
        Abc_NtkForEachCo( pNtk, pObj, i )
        {
            if !Abc_ObjIsPo(pObj) {
                continue; }
            pNet = Abc_ObjFanin0(pObj);
            // skip marked nets
//            if ( Abc_NodeIsTravIdCurrent(pNet) )
//                continue;
//            Abc_NodeSetTravIdCurrent( pNet );
            nValues = Abc_ObjMvVarNum(pNet);
            pValues = (Abc_Obj_t **)pNet.pCopy;
            for  v = 0; v < nValues; v++  {
                pTemp = Abc_NtkCreatePo( pNtkNew );
                Abc_ObjAddFanin( pTemp, pValues[v] );
                if nValuesMax == 2 {
                    Abc_ObjAssignName( pTemp, Abc_ObjName(pNet), nil );
                } else {
                    Abc_NtkConvertAssignName( pTemp, pNet, v ); }
            }
        }
        Abc_NtkForEachCo( pNtk, pObj, i )
        {
            if Abc_ObjIsPo(pObj) {
                continue; }
            pNet = Abc_ObjFanin0(pObj);
            // skip marked nets
//            if ( Abc_NodeIsTravIdCurrent(pNet) )
//                continue;
//            Abc_NodeSetTravIdCurrent( pNet );
            nValues = Abc_ObjMvVarNum(pNet);
            pValues = (Abc_Obj_t **)pNet.pCopy;
            for  v = 0; v < nValues; v++  {
                pTemp = Abc_NtkCreateBi( pNtkNew );
                Abc_ObjAddFanin( pTemp, pValues[v] );
                if nValuesMax == 2 {
                    Abc_ObjAssignName( pTemp, Abc_ObjName(pNet), nil );
                } else {
                    Abc_NtkConvertAssignName( pTemp, pNet, v ); }
                nCount2++;
            }
        }
    
    } else { // if ( fPositional == 0 )
    
        Abc_NtkForEachCo( pNtk, pObj, i )
        {
            if !Abc_ObjIsPo(pObj) {
                continue; }
            pNet = Abc_ObjFanin0(pObj);
            // skip marked nets
//            if ( Abc_NodeIsTravIdCurrent(pNet) )
//                continue;
//            Abc_NodeSetTravIdCurrent( pNet );
            nValues = Abc_ObjMvVarNum(pNet);
            pValues = (Abc_Obj_t **)pNet.pCopy;
            nBits = Abc_Base2Log( nValues );
            for  k = 0; k < nBits; k++  {
                pBit = Abc_ObjNot( Abc_AigConst1(pNtkNew) );
                for  v = 0; v < nValues; v++  {
                    if v & (1<<k) {
                        pBit = Abc_AigOr( (Abc_Aig_t *)pNtkNew.pManFunc, pBit, pValues[v] ); }
                }
                pTemp = Abc_NtkCreatePo( pNtkNew );
                Abc_ObjAddFanin( pTemp, pBit );
                if nValuesMax == 2 {
                    Abc_ObjAssignName( pTemp, Abc_ObjName(pNet), nil );
                } else {
                    Abc_NtkConvertAssignName( pTemp, pNet, k ); }
            }
        }
        Abc_NtkForEachCo( pNtk, pObj, i )
        {
            if Abc_ObjIsPo(pObj) {
                continue; }
            pNet = Abc_ObjFanin0(pObj);
            // skip marked nets
//            if ( Abc_NodeIsTravIdCurrent(pNet) )
//                continue;
//            Abc_NodeSetTravIdCurrent( pNet );
            nValues = Abc_ObjMvVarNum(pNet);
            pValues = (Abc_Obj_t **)pNet.pCopy;
            nBits = Abc_Base2Log( nValues );
            for  k = 0; k < nBits; k++  {
                pBit = Abc_ObjNot( Abc_AigConst1(pNtkNew) );
                for  v = 0; v < nValues; v++  {
                    if v & (1<<k) {
                        pBit = Abc_AigOr( (Abc_Aig_t *)pNtkNew.pManFunc, pBit, pValues[v] ); }
                }
                pTemp = Abc_NtkCreateBi( pNtkNew );
                Abc_ObjAddFanin( pTemp, pBit );
                if nValuesMax == 2 {
                    Abc_ObjAssignName( pTemp, Abc_ObjName(pNet), nil );
                } else {
                    Abc_NtkConvertAssignName( pTemp, pNet, k ); }
                nCount2++;
            }
        }
    }

    if Abc_NtkLatchNum(pNtk) {
        var vTemp *Vec_Ptr_t;
        var pLatch, pObjLi, pObjLo *Abc_Obj_t;
        var i int32;
        // move free vars to the front among the PIs
        vTemp = Vec_PtrAlloc( Vec_PtrSize(pNtkNew.vPis) );
        Abc_NtkForEachPi( pNtkNew, pObj, i )
            if ( strncmp( Abc_ObjName(pObj), "free_var_", 9 ) == 0 )
                Vec_PtrPush( vTemp, pObj );
        Abc_NtkForEachPi( pNtkNew, pObj, i )
            if ( strncmp( Abc_ObjName(pObj), "free_var_", 9 ) != 0 )
                Vec_PtrPush( vTemp, pObj );
        assert(Vec_PtrSize(vTemp) == Vec_PtrSize(pNtkNew.vPis), "Vec_PtrSize(vTemp) == Vec_PtrSize(pNtkNew.vPis)");
        Vec_PtrFree( pNtkNew.vPis );
        pNtkNew.vPis = vTemp;
        // move free vars to the front among the CIs
        vTemp = Vec_PtrAlloc( Vec_PtrSize(pNtkNew.vCis) );
        Abc_NtkForEachCi( pNtkNew, pObj, i )
            if ( strncmp( Abc_ObjName(pObj), "free_var_", 9 ) == 0 )
                Vec_PtrPush( vTemp, pObj );
        Abc_NtkForEachCi( pNtkNew, pObj, i )
            if ( strncmp( Abc_ObjName(pObj), "free_var_", 9 ) != 0 )
                Vec_PtrPush( vTemp, pObj );
        assert(Vec_PtrSize(vTemp) == Vec_PtrSize(pNtkNew.vCis), "Vec_PtrSize(vTemp) == Vec_PtrSize(pNtkNew.vCis)");
        Vec_PtrFree( pNtkNew.vCis );
        pNtkNew.vCis = vTemp;
        // create registers
        assert(nCount1 == nCount2, "nCount1 == nCount2");
        for  i = 0; i < nCount1; i++  {
            // create latch
            pLatch = Abc_NtkCreateLatch( pNtkNew );
            Abc_LatchSetInit0( pLatch );
            Abc_ObjAssignName( pLatch, Abc_ObjName(pLatch), nil );
            // connect
            pObjLi = Abc_NtkCo( pNtkNew, Abc_NtkCoNum(pNtkNew)-nCount1+i );
            pObjLo = Abc_NtkCi( pNtkNew, Abc_NtkCiNum(pNtkNew)-nCount1+i );
            Abc_ObjAddFanin( pLatch, pObjLi );
            Abc_ObjAddFanin( pObjLo, pLatch );
        }
    }

    // cleanup
    ABC_FREE( pBits );
    Abc_NtkForEachObj( pNtk, pObj, i )
        if ( pObj.pCopy )
            ABC_FREE( pObj.pCopy );

    // remove dangling nodes
    i = Abc_AigCleanup((Abc_Aig_t *)pNtkNew.pManFunc);
//    printf( "Cleanup removed %d nodes.\n", i );
//    Abc_NtkReassignIds( pNtkNew );

    // check integrity
    if !Abc_NtkCheck( pNtkNew ) {
        fmt.Fprintf(os.Stdout, "Abc_NtkStrashBlifMv(): Network check has failed.\n");
        Abc_NtkDelete( pNtkNew );
        return nil;
    }
    return pNtkNew;
}

/**Function*************************************************************

  Synopsis    [Extract the MV-skeleton of the BLIF-MV network.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkSkeletonBlifMv(pNtk *Abc_Ntk_t) *Abc_Ntk_t {
    var fUsePositional int32;
    var pNtkNew *Abc_Ntk_t;
    var pObj, pNet, pNetNew, pNodeNew, pTermNew, pBoxNew *Abc_Obj_t;
    var i, k, v, nValues, nBits int32;

    assert(Abc_NtkIsNetlist(pNtk), "Abc_NtkIsNetlist(pNtk)");
    assert(Abc_NtkHasBlifMv(pNtk), "Abc_NtkHasBlifMv(pNtk)");
    assert(Abc_NtkWhiteboxNum(pNtk) == 0, "Abc_NtkWhiteboxNum(pNtk) == 0");
    assert(Abc_NtkBlackboxNum(pNtk) == 0, "Abc_NtkBlackboxNum(pNtk) == 0");

    // clean the node copy fields
    Abc_NtkCleanCopy( pNtk );

    // start the network
    pNtkNew = Abc_NtkAlloc( pNtk.ntkType, pNtk.ntkFunc, 1 );
    // duplicate the name and the spec
    pNtkNew.pName = Extra_UtilStrsav( pNtk.pName );
    pNtkNew.pSpec = Extra_UtilStrsav( pNtk.pName );
    // create the internal box (it is important to put it first!)
    pBoxNew = Abc_NtkCreateWhitebox( pNtkNew );
    // create PIs and their nets
    Abc_NtkForEachPi( pNtk, pObj, i )
    {
        Abc_NtkDupObj( pNtkNew, pObj, 0 );
        pNet = Abc_ObjFanout0(pObj);
        Abc_NtkDupObj( pNtkNew, pNet, 1 );
        Abc_ObjAddFanin( pNet.pCopy, pObj.pCopy );
    }
    // create POs and their nets
    Abc_NtkForEachPo( pNtk, pObj, i )
    {
        Abc_NtkDupObj( pNtkNew, pObj, 0 );
        pNet = Abc_ObjFanin0(pObj);
        if pNet.pCopy == nil {
            Abc_NtkDupObj( pNtkNew, pNet, 1 ); }
        Abc_ObjAddFanin( pObj.pCopy, pNet.pCopy );
    }
    // create latches
    Abc_NtkForEachLatch( pNtk, pObj, i )
    {
        Abc_NtkDupBox( pNtkNew, pObj, 0 );
        // latch outputs
        pNet = Abc_ObjFanout0(Abc_ObjFanout0(pObj));
        assert(pNet.pCopy == nil, "pNet.pCopy == nil");
        Abc_NtkDupObj( pNtkNew, pNet, 1 );
        Abc_ObjAddFanin( pNet.pCopy, Abc_ObjFanout0(pObj).pCopy );
        // latch inputs
        pNet = Abc_ObjFanin0(Abc_ObjFanin0(pObj));
        if pNet.pCopy == nil {
            Abc_NtkDupObj( pNtkNew, pNet, 1 ); }
        Abc_ObjAddFanin( Abc_ObjFanin0(pObj).pCopy, pNet.pCopy );
    }

    // encode the CI nets
    Abc_NtkIncrementTravId( pNtk );
    if fUsePositional != 0 {
        Abc_NtkForEachCi( pNtk, pObj, i )
        {
            pNet = Abc_ObjFanout0(pObj);
            nValues = Abc_ObjMvVarNum(pNet);
            for  v = 0; v < nValues; v++  {
                pNodeNew = Abc_NtkCreateNode( pNtkNew );
                pNodeNew.pData = Abc_SopEncoderPos( (Mem_Flex_t *)pNtkNew.pManFunc, v, nValues );
                pNetNew = Abc_NtkCreateNet( pNtkNew );
                pTermNew = Abc_NtkCreateBi( pNtkNew );
                Abc_ObjAddFanin( pNodeNew, pNet.pCopy );
                Abc_ObjAddFanin( pNetNew, pNodeNew );
                Abc_ObjAddFanin( pTermNew, pNetNew );
                Abc_ObjAddFanin( pBoxNew, pTermNew );
            }
            // mark the net
            Abc_NodeSetTravIdCurrent( pNet );
        }
    
    } else {
        Abc_NtkForEachCi( pNtk, pObj, i )
        {
            pNet = Abc_ObjFanout0(pObj);
            nValues = Abc_ObjMvVarNum(pNet);
            nBits = Abc_Base2Log( nValues );
            for  k = 0; k < nBits; k++  {
                pNodeNew = Abc_NtkCreateNode( pNtkNew );
                pNodeNew.pData = Abc_SopEncoderLog( (Mem_Flex_t *)pNtkNew.pManFunc, k, nValues );
                pNetNew = Abc_NtkCreateNet( pNtkNew );
                pTermNew = Abc_NtkCreateBi( pNtkNew );
                Abc_ObjAddFanin( pNodeNew, pNet.pCopy );
                Abc_ObjAddFanin( pNetNew, pNodeNew );
                Abc_ObjAddFanin( pTermNew, pNetNew );
                Abc_ObjAddFanin( pBoxNew, pTermNew );
            }
            // mark the net
            Abc_NodeSetTravIdCurrent( pNet );
        }
    }

    // encode the CO nets
    if fUsePositional != 0 {
        Abc_NtkForEachCo( pNtk, pObj, i )
        {
            pNet = Abc_ObjFanin0(pObj);
            // skip marked nets
            if Abc_NodeIsTravIdCurrent(pNet) {
                continue; }
            Abc_NodeSetTravIdCurrent( pNet );
            nValues = Abc_ObjMvVarNum(pNet);
            pNodeNew = Abc_NtkCreateNode( pNtkNew );
            pNodeNew.pData = Abc_SopDecoderPos( (Mem_Flex_t *)pNtkNew.pManFunc, nValues );
            for  v = 0; v < nValues; v++  {
                pTermNew = Abc_NtkCreateBo( pNtkNew );
                pNetNew = Abc_NtkCreateNet( pNtkNew );
                Abc_ObjAddFanin( pTermNew, pBoxNew );
                Abc_ObjAddFanin( pNetNew, pTermNew );
                Abc_ObjAddFanin( pNodeNew, pNetNew );
            }
            Abc_ObjAddFanin( pNet.pCopy, pNodeNew );
        }
    
    } else {
        Abc_NtkForEachCo( pNtk, pObj, i )
        {
            pNet = Abc_ObjFanin0(pObj);
            // skip marked nets
            if Abc_NodeIsTravIdCurrent(pNet) {
                continue; }
            Abc_NodeSetTravIdCurrent( pNet );
            nValues = Abc_ObjMvVarNum(pNet);
            nBits = Abc_Base2Log( nValues );
            pNodeNew = Abc_NtkCreateNode( pNtkNew );
            pNodeNew.pData = Abc_SopDecoderLog( (Mem_Flex_t *)pNtkNew.pManFunc, nValues );
            for  k = 0; k < nBits; k++  {
                pTermNew = Abc_NtkCreateBo( pNtkNew );
                pNetNew = Abc_NtkCreateNet( pNtkNew );
                Abc_ObjAddFanin( pTermNew, pBoxNew );
                Abc_ObjAddFanin( pNetNew, pTermNew );
                Abc_ObjAddFanin( pNodeNew, pNetNew );
            }
            Abc_ObjAddFanin( pNet.pCopy, pNodeNew );
        }
    }

    // if it is a BLIF-MV netlist transfer the values of all nets
    if Abc_NtkHasBlifMv(pNtk) && Abc_NtkMvVar(pNtk) {
        if Abc_NtkMvVar( pNtkNew ) == nil {
            Abc_NtkStartMvVars( pNtkNew ); }
        Abc_NtkForEachNet( pNtk, pObj, i )
            if ( pObj.pCopy )
                Abc_NtkSetMvVarValues( pObj.pCopy, Abc_ObjMvVarNum(pObj) );
    }

    // check integrity
    if !Abc_NtkCheck( pNtkNew ) {
        fmt.Fprintf(os.Stdout, "Abc_NtkSkeletonBlifMv(): Network check has failed.\n");
        Abc_NtkDelete( pNtkNew );
        return nil;
    }
    return pNtkNew;
}

/**Function*************************************************************

  Synopsis    [Inserts processed network into original base MV network.]

  Description [The original network remembers the interface of combinational 
  logic (PIs/POs/latches names and values). The processed network may 
  be binary or multi-valued (currently, multi-value is not supported). 
  The resulting network has the same interface as the original network 
  while the internal logic is the same as that of the processed network.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkInsertBlifMv(pNtkBase *Abc_Ntk_t, pNtkLogic *Abc_Ntk_t) *Abc_Ntk_t {
    var pNtkSkel, pNtkNew *Abc_Ntk_t;
    var pBox *Abc_Obj_t;

    assert(Abc_NtkIsNetlist(pNtkBase), "Abc_NtkIsNetlist(pNtkBase)");
    assert(Abc_NtkHasBlifMv(pNtkBase), "Abc_NtkHasBlifMv(pNtkBase)");
    assert(Abc_NtkWhiteboxNum(pNtkBase) == 0, "Abc_NtkWhiteboxNum(pNtkBase) == 0");
    assert(Abc_NtkBlackboxNum(pNtkBase) == 0, "Abc_NtkBlackboxNum(pNtkBase) == 0");

    assert(Abc_NtkIsNetlist(pNtkLogic), "Abc_NtkIsNetlist(pNtkLogic)");
    assert(Abc_NtkHasBlifMv(pNtkLogic), "Abc_NtkHasBlifMv(pNtkLogic)");
    assert(Abc_NtkWhiteboxNum(pNtkLogic) == 0, "Abc_NtkWhiteboxNum(pNtkLogic) == 0");
    assert(Abc_NtkBlackboxNum(pNtkLogic) == 0, "Abc_NtkBlackboxNum(pNtkLogic) == 0");

    // extract the skeleton of the old network
    pNtkSkel = Abc_NtkSkeletonBlifMv( pNtkBase );

    // set the implementation of the box to be the same as the processed network
    assert(Abc_NtkWhiteboxNum(pNtkSkel) == 1, "Abc_NtkWhiteboxNum(pNtkSkel) == 1");
    pBox = Abc_NtkBox( pNtkSkel, 0 );
    assert(Abc_ObjIsWhitebox(pBox), "Abc_ObjIsWhitebox(pBox)");
    assert(pBox.pData == nil, "pBox.pData == nil");
    assert(Abc_ObjFaninNum(pBox) == Abc_NtkPiNum(pNtkLogic), "Abc_ObjFaninNum(pBox) == Abc_NtkPiNum(pNtkLogic)");
    assert(Abc_ObjFanoutNum(pBox) == Abc_NtkPoNum(pNtkLogic), "Abc_ObjFanoutNum(pBox) == Abc_NtkPoNum(pNtkLogic)");
    pBox.pData = pNtkLogic;

    // flatten the hierarchy to insert the processed network
    pNtkNew = Abc_NtkFlattenLogicHierarchy( pNtkSkel );
    pBox.pData = nil;
    Abc_NtkDelete( pNtkSkel );
    return pNtkNew;    
}

/**Function*************************************************************

  Synopsis    [Converts SOP netlist into BLIF-MV netlist.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkConvertToBlifMv(pNtk *Abc_Ntk_t) int32 {
#ifdef ABC_USE_CUDD
    Mem_Flex_t * pMmFlex;
    Abc_Obj_t * pNode;
    Vec_Str_t * vCube;
    char * pSop0, * pSop1, * pBlifMv, * pCube, * pCur;
    int Value, nCubes, nSize, i, k;

    assert( Abc_NtkIsNetlist(pNtk) );
    if ( !Abc_NtkToBdd(pNtk) )
    {
        printf( "Converting logic functions to BDDs has failed.\n" );
        return 0;
    }

    pMmFlex = Mem_FlexStart();
    vCube   = Vec_StrAlloc( 100 );
    Abc_NtkForEachNode( pNtk, pNode, i )
    {
        // convert BDD into cubes for on-set and off-set
        Abc_NodeBddToCnf( pNode, pMmFlex, vCube, 0, &pSop0, &pSop1 );
        // allocate room for the MV-SOP
        nCubes = Abc_SopGetCubeNum(pSop0) + Abc_SopGetCubeNum(pSop1);
        nSize = nCubes*(2*Abc_ObjFaninNum(pNode) + 2)+1;
        pBlifMv = Mem_FlexEntryFetch( pMmFlex, nSize );
        // add the cubes
        pCur = pBlifMv;
        Abc_SopForEachCube( pSop0, Abc_ObjFaninNum(pNode), pCube )
        {
            Abc_CubeForEachVar( pCube, Value, k )
            {
                *pCur++ = Value;
                *pCur++ = ' ';
            }
            *pCur++ = '0';
            *pCur++ = '\n';
        }
        Abc_SopForEachCube( pSop1, Abc_ObjFaninNum(pNode), pCube )
        {
            Abc_CubeForEachVar( pCube, Value, k )
            {
                *pCur++ = Value;
                *pCur++ = ' ';
            }
            *pCur++ = '1';
            *pCur++ = '\n';
        }
        *pCur++ = 0;
        assert( pCur - pBlifMv == nSize );
        // update the node representation
        Cudd_RecursiveDeref( (DdManager *)pNtk->pManFunc, (DdNode *)pNode->pData );
        pNode->pData = pBlifMv;
    }

    // update the functionality type
    pNtk->ntkFunc = ABC_FUNC_BLIFMV;
    Cudd_Quit( (DdManager *)pNtk->pManFunc );
    pNtk->pManFunc = pMmFlex;

    Vec_StrFree( vCube );
#endif
    return 1;
}

/**Function*************************************************************

  Synopsis    [Converts SOP into MV-SOP.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NodeConvertSopToMvSop(nVars int32, vSop0 *Vec_Int_t, vSop1 *Vec_Int_t) string {
    var pMvSop, pCur string;
    var uCube uint32;
    var nCubes, nSize, Value, i, k int32;
    // consider the case of the constant node
    if Vec_IntSize(vSop0) == 0 || Vec_IntSize(vSop1) == 0 {
        // (temporary) create a tautology cube
        pMvSop = ABC_ALLOC( char, nVars + 3 );
        for  k = 0; k < nVars; k++  {
            pMvSop[k] = '-';
        }
        pMvSop[nVars] = '0' + int32((Vec_IntSize(vSop1) > 0));
        pMvSop[nVars+1] = '\n';
        pMvSop[nVars+2] = 0;
        return pMvSop;
    }
    // find the total number of cubes
    nCubes = Vec_IntSize(vSop0) + Vec_IntSize(vSop1);
    // find the size of the MVSOP represented as a C-string
    // (each cube has nVars variables + one output literal + end-of-line,
    // and the string is zero-terminated)
    nSize = nCubes * (nVars + 2) + 1; 
    // allocate memory
    pMvSop = pCur = ABC_ALLOC( char, nSize );
    // fill in the negative polarity cubes
    Vec_IntForEachEntry( vSop0, uCube, i )
    {
        for  k = 0; k < nVars; k++  {
            Value = (uCube >> (2*k)) & 3;
            if Value == 1 {
                *pCur++ = '0';
            } else if Value == 2 {
                *pCur++ = '1';
            } else if Value == 0 {
                *pCur++ = '-';
            } else {
                assert(false, "0"); }
        }
        *pCur++ = '0';
        *pCur++ = '\n';
    }
    // fill in the positive polarity cubes
    Vec_IntForEachEntry( vSop1, uCube, i )
    {
        for  k = 0; k < nVars; k++  {
            Value = (uCube >> (2*k)) & 3;
            if Value == 1 {
                *pCur++ = '0';
            } else if Value == 2 {
                *pCur++ = '1';
            } else if Value == 0 {
                *pCur++ = '-';
            } else {
                assert(false, "0"); }
        }
        *pCur++ = '1';
        *pCur++ = '\n';
    }
    *pCur++ = 0;
    assert(pCur - pMvSop == nSize, "pCur - pMvSop == nSize");
    return pMvSop;
}


/**Function*************************************************************

  Synopsis    [A prototype of internal cost evaluation procedure.]

  Description [This procedure takes the number of variables (nVars),
  the array of values of the inputs and the output (pVarValues) 
  (note that this array has nVars+1 entries), and an MV-SOP represented
  as a C-string with one charater for each literal, including inputs
  and output. Each cube is terminated with the new-line character ('\n').
  The string is zero-terminated.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NodeEvalMvCostInternal(nVars int32, pVarValues *int32, pMvSop string) int32 {
    // for now, return the number of cubes in the MV-SOP
    var Counter int32;
    for  *pMvSop  { Counter += (*pMvSop++ == '\n'); }
    return Counter;
}


/**Function*************************************************************

  Synopsis    [Evaluates the cost of the cut.]

  Description [The Boolean function of the cut is specified by two SOPs, 
  which represent the negative/positive polarities of the cut function.
  Converts these two SOPs into a mutually-agreed-upon representation 
  to be passed to the internal cost-evaluation procedure (see the above
  prototype Abc_NodeEvalMvCostInternal).]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NodeEvalMvCost(nVars int32, vSop0 *Vec_Int_t, vSop1 *Vec_Int_t) int32 {
    var pMvSop string;
    var pVarValues *int32;
    var i, RetValue int32;
    // collect the input and output values (currently, they are binary)
    pVarValues = ABC_ALLOC( int, nVars + 1 );
    for  i = 0; i <= nVars; i++  {
        pVarValues[i] = 2;
    }
    // prepare MV-SOP for evaluation
    pMvSop = Abc_NodeConvertSopToMvSop( nVars, vSop0, vSop1 );
    // have a look at the MV-SOP:
//    printf( "%s\n", pMvSop );
    // get the result of internal cost evaluation
    RetValue = Abc_NodeEvalMvCostInternal( nVars, pVarValues, pMvSop );
    // cleanup
    ABC_FREE( pVarValues );
    ABC_FREE( pMvSop );
    return RetValue;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

