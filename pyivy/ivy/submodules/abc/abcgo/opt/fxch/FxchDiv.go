package fxch

import "fmt"

/**CFile****************************************************************************************************************

  FileName    [ FxchDiv.c ]

  PackageName [ Fast eXtract with Cube Hashing (FXCH) ]

  Synopsis    [ Divisor handling functions ]

  Author      [ Bruno Schmitt - boschmitt at inf.ufrgs.br ]

  Affiliation [ UFRGS ]

  Date        [ Ver. 1.0. Started - March 6, 2016. ]

  Revision    []

***********************************************************************************************************************/

//#include "Fxch.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// FUNCTION DEFINITIONS
////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
func Fxch_DivNormalize(vCubeFree *Vec_Int_t) int32 {
    var L *int32  = Vec_IntArray(vCubeFree);
    var RetValue int32;
    var LitA0 int32 = -1;
    var LitB0 int32 = -1;
    var LitA1 int32 = -1;
    var LitB1 int32 = -1;
    assert(Vec_IntSize(vCubeFree) == 4, "Vec_IntSize(vCubeFree) == 4");
    if Abc_LitIsCompl(L[0]) != Abc_LitIsCompl(L[1]) && (L[0] >> 2) == (L[1] >> 2) { // diff cubes, same vars
    
        if Abc_LitIsCompl(L[2]) == Abc_LitIsCompl(L[3]) {
            return -1; }
        LitA0 = Abc_Lit2Var(L[0]);
LitB0 = Abc_Lit2Var(L[1]);
        if Abc_LitIsCompl(L[0]) == Abc_LitIsCompl(L[2]) {
            assert(Abc_LitIsCompl(L[1]) == Abc_LitIsCompl(L[3]), "Abc_LitIsCompl(L[1]) == Abc_LitIsCompl(L[3])");
            LitA1 = Abc_Lit2Var(L[2]);
LitB1 = Abc_Lit2Var(L[3]);
        
        } else {
            assert(Abc_LitIsCompl(L[0]) == Abc_LitIsCompl(L[3]), "Abc_LitIsCompl(L[0]) == Abc_LitIsCompl(L[3])");
            assert(Abc_LitIsCompl(L[1]) == Abc_LitIsCompl(L[2]), "Abc_LitIsCompl(L[1]) == Abc_LitIsCompl(L[2])");
            LitA1 = Abc_Lit2Var(L[3]);
LitB1 = Abc_Lit2Var(L[2]);
        }
    
    } else if Abc_LitIsCompl(L[1]) != Abc_LitIsCompl(L[2]) && (L[1] >> 2) == (L[2] >> 2) {
        if Abc_LitIsCompl(L[0]) == Abc_LitIsCompl(L[3]) {
            return -1; }
        LitA0 = Abc_Lit2Var(L[1]);
LitB0 = Abc_Lit2Var(L[2]);
        if Abc_LitIsCompl(L[1]) == Abc_LitIsCompl(L[0]) {
            LitA1 = Abc_Lit2Var(L[0]);
LitB1 = Abc_Lit2Var(L[3]);
        } else {
            LitA1 = Abc_Lit2Var(L[3]);
LitB1 = Abc_Lit2Var(L[0]); }
    
    } else if Abc_LitIsCompl(L[2]) != Abc_LitIsCompl(L[3]) && (L[2] >> 2) == (L[3] >> 2) {
        if Abc_LitIsCompl(L[0]) == Abc_LitIsCompl(L[1]) {
            return -1; }
        LitA0 = Abc_Lit2Var(L[2]);
LitB0 = Abc_Lit2Var(L[3]);
        if Abc_LitIsCompl(L[2]) == Abc_LitIsCompl(L[0]) {
            LitA1 = Abc_Lit2Var(L[0]);
LitB1 = Abc_Lit2Var(L[1]);
        } else {
            LitA1 = Abc_Lit2Var(L[1]);
LitB1 = Abc_Lit2Var(L[0]); }
    
    } else {
        return -1; }
    assert(LitA0 == Abc_LitNot(LitB0), "LitA0 == Abc_LitNot(LitB0)");
    if Abc_LitIsCompl(LitA0) {
        // inlining of macro ABC_SWAP
        { var t int32  = LitA0; LitA0 = LitB0; LitB0 = t; };
        // inlining of macro ABC_SWAP
        { var t int32  = LitA1; LitA1 = LitB1; LitB1 = t; };
    }
    assert(!Abc_LitIsCompl(LitA0), "!Abc_LitIsCompl(LitA0)");
    if Abc_LitIsCompl(LitA1) {
        LitA1 = Abc_LitNot(LitA1);
        LitB1 = Abc_LitNot(LitB1);
        RetValue = 1;
    }
    assert(!Abc_LitIsCompl(LitA1), "!Abc_LitIsCompl(LitA1)");
    // arrange literals in such as a way that
    // - the first two literals are control literals from different cubes
    // - the third literal is non-complented data input
    // - the forth literal is possibly complemented data input
    L[0] = Abc_Var2Lit( LitA0, 0 );
    L[1] = Abc_Var2Lit( LitB0, 1 );
    L[2] = Abc_Var2Lit( LitA1, 0 );
    L[3] = Abc_Var2Lit( LitB1, 1 );
    return RetValue;
}

/**Function*************************************************************

  Synopsis    [ Creates a Divisor. ]

  Description [ This functions receive as input two sub-cubes and creates
                a divisor using their information. The divisor is stored 
                in vCubeFree vector of the pFxchMan structure.
                
                It returns the base value, which is the number of elements
                that the cubes pair used to generate the devisor have in
                common. ]

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fxch_DivCreate(pFxchMan *Fxch_Man_t, pSubCube0 *Fxch_SubCube_t, pSubCube1 *Fxch_SubCube_t) int32 {
    var Base int32;

    var SC0_Lit0, SC0_Lit1, SC1_Lit0, SC1_Lit1 int32;

    var Cube0Size, Cube1Size int32;

    Vec_IntClear( pFxchMan.vCubeFree );

    SC0_Lit0 = Fxch_ManGetLit( pFxchMan, pSubCube0.iCube, pSubCube0.iLit0 );
    SC0_Lit1 = 0;
    SC1_Lit0 = Fxch_ManGetLit( pFxchMan, pSubCube1.iCube, pSubCube1.iLit0 );
    SC1_Lit1 = 0;

    if pSubCube0.iLit1 == 0 && pSubCube1.iLit1 == 0 {
        Vec_IntPush( pFxchMan.vCubeFree, SC0_Lit0 );
        Vec_IntPush( pFxchMan.vCubeFree, SC1_Lit0 );
    
    } else if pSubCube0.iLit1 > 0 && pSubCube1.iLit1 > 0 {
        var RetValue int32;

        SC0_Lit1 = Fxch_ManGetLit( pFxchMan, pSubCube0.iCube, pSubCube0.iLit1 );
        SC1_Lit1 = Fxch_ManGetLit( pFxchMan, pSubCube1.iCube, pSubCube1.iLit1 );

        if SC0_Lit0 < SC1_Lit0 {
            Vec_IntPush( pFxchMan.vCubeFree, Abc_Var2Lit( SC0_Lit0, 0 ) );
            Vec_IntPush( pFxchMan.vCubeFree, Abc_Var2Lit( SC1_Lit0, 1 ) );
            Vec_IntPush( pFxchMan.vCubeFree, Abc_Var2Lit( SC0_Lit1, 0 ) );
            Vec_IntPush( pFxchMan.vCubeFree, Abc_Var2Lit( SC1_Lit1, 1 ) );
        
        } else {
            Vec_IntPush( pFxchMan.vCubeFree, Abc_Var2Lit( SC1_Lit0, 0 ) );
            Vec_IntPush( pFxchMan.vCubeFree, Abc_Var2Lit( SC0_Lit0, 1 ) );
            Vec_IntPush( pFxchMan.vCubeFree, Abc_Var2Lit( SC1_Lit1, 0 ) );
            Vec_IntPush( pFxchMan.vCubeFree, Abc_Var2Lit( SC0_Lit1, 1 ) );
        }

        RetValue = Fxch_DivNormalize( pFxchMan.vCubeFree );
        if RetValue == -1 {
            return -1; }
     
    } else {
        if pSubCube0.iLit1 > 0 {
            SC0_Lit1 = Fxch_ManGetLit( pFxchMan, pSubCube0.iCube, pSubCube0.iLit1 );

            Vec_IntPush( pFxchMan.vCubeFree, SC1_Lit0 );
            if SC0_Lit0 == Abc_LitNot( SC1_Lit0 ) {
                Vec_IntPush( pFxchMan.vCubeFree, SC0_Lit1 );
            } else if SC0_Lit1 == Abc_LitNot( SC1_Lit0 ) {
                Vec_IntPush( pFxchMan.vCubeFree, SC0_Lit0 ); }
        
        } else {
            SC1_Lit1 = Fxch_ManGetLit( pFxchMan, pSubCube1.iCube, pSubCube1.iLit1 );

            Vec_IntPush( pFxchMan.vCubeFree, SC0_Lit0 );
            if SC1_Lit0 == Abc_LitNot( SC0_Lit0 ) {
                Vec_IntPush( pFxchMan.vCubeFree, SC1_Lit1 );
            } else if SC1_Lit1 == Abc_LitNot( SC0_Lit0 ) {
                Vec_IntPush( pFxchMan.vCubeFree, SC1_Lit0 ); }
        }
    }

    if Vec_IntSize( pFxchMan.vCubeFree ) == 0 {
        return -1; }

    if Vec_IntSize ( pFxchMan.vCubeFree ) == 2 {
        Vec_IntSort( pFxchMan.vCubeFree, 0 );

        Vec_IntWriteEntry( pFxchMan.vCubeFree, 0, Abc_Var2Lit( Vec_IntEntry( pFxchMan.vCubeFree, 0 ), 0 ) );
        Vec_IntWriteEntry( pFxchMan.vCubeFree, 1, Abc_Var2Lit( Vec_IntEntry( pFxchMan.vCubeFree, 1 ), 1 ) );
    }

    Cube0Size = Vec_IntSize( Fxch_ManGetCube( pFxchMan, pSubCube0.iCube ) );
    Cube1Size = Vec_IntSize( Fxch_ManGetCube( pFxchMan, pSubCube1.iCube ) );
    if Vec_IntSize( pFxchMan.vCubeFree ) % 2 == 0 {
        Base = Abc_MinInt( Cube0Size, Cube1Size )
               -( Vec_IntSize( pFxchMan.vCubeFree ) / 2)  - 1; /* 1 or 2 Lits, 1 SOP NodeID */
    
    } else {
        return -1; }

    return Base;
}

/**Function*************************************************************

  Synopsis    [ Add a divisor to the divisors hash table. ]

  Description [ This functions will try to add the divisor store in 
                vCubeFree to the divisor hash table. If the divisor
                is already present in the hash table it will just 
                increment its weight, otherwise it will add the divisor
                and asign an initial weight. ]

  SideEffects [ If the fUpdate option is set, the function will also
                update the divisor priority queue. ]

  SeeAlso     []

***********************************************************************/
func Fxch_DivAdd(pFxchMan *Fxch_Man_t, fUpdate int32, fSingleCube int32, fBase int32) int32 {
    var iDiv int32  = Hsh_VecManAdd( pFxchMan.pDivHash, pFxchMan.vCubeFree );

    /* Verify if the divisor already exist */ 
    if iDiv == Vec_FltSize( pFxchMan.vDivWeights ) {
        Vec_WecPushLevel( pFxchMan.vDivCubePairs );

        /* Assign initial weight */
        if fSingleCube != 0 {
            Vec_FltPush( pFxchMan.vDivWeights,
                         -Vec_IntSize( pFxchMan.vCubeFree ) + 0.9
                         -0.001 * Fxch_ManComputeLevelDiv( pFxchMan, pFxchMan.vCubeFree ) );
        
        } else {
            Vec_FltPush( pFxchMan.vDivWeights,
                         -Vec_IntSize( pFxchMan.vCubeFree ) + 0.9
                         -0.0009 * Fxch_ManComputeLevelDiv( pFxchMan, pFxchMan.vCubeFree ) );
        }

    }

    /* Increment weight */
    if fSingleCube != 0 {
        Vec_FltAddToEntry( pFxchMan.vDivWeights, iDiv, 1 );
    } else {
        Vec_FltAddToEntry( pFxchMan.vDivWeights, iDiv, fBase + Vec_IntSize( pFxchMan.vCubeFree ) - 1 ); }

    assert(iDiv < Vec_FltSize( pFxchMan.vDivWeights ), "iDiv < Vec_FltSize( pFxchMan.vDivWeights )");

    if fUpdate != 0 {
        if pFxchMan.vDivPrio {
            if Vec_QueIsMember( pFxchMan.vDivPrio, iDiv ) {
                Vec_QueUpdate( pFxchMan.vDivPrio, iDiv );
            } else {
                Vec_QuePush( pFxchMan.vDivPrio, iDiv ); }
        } }

    return iDiv;
}

/**Function*************************************************************

  Synopsis    [ Removes a divisor to the divisors hash table. ]

  Description [ This function don't effectively removes a divisor from
                the hash table (the hash table implementation don't 
                support such operation). It only assures its existence
                and decrement its weight. ]

  SideEffects [ If the fUpdate option is set, the function will also
                update the divisor priority queue. ]

  SeeAlso     []

***********************************************************************/
func Fxch_DivRemove(pFxchMan *Fxch_Man_t, fUpdate int32, fSingleCube int32, fBase int32) int32 {
    var iDiv int32  = Hsh_VecManAdd( pFxchMan.pDivHash, pFxchMan.vCubeFree );

    assert(iDiv < Vec_FltSize( pFxchMan.vDivWeights ), "iDiv < Vec_FltSize( pFxchMan.vDivWeights )");

    /* Decrement weight */
    if fSingleCube != 0 {
        Vec_FltAddToEntry( pFxchMan.vDivWeights, iDiv, -1 );
    } else {
        Vec_FltAddToEntry( pFxchMan.vDivWeights, iDiv, -( fBase + Vec_IntSize( pFxchMan.vCubeFree ) - 1 ) ); }

    if fUpdate != 0 {
        if pFxchMan.vDivPrio {
            if Vec_QueIsMember( pFxchMan.vDivPrio, iDiv ) {
                Vec_QueUpdate( pFxchMan.vDivPrio, iDiv ); }
        } }

    return iDiv;
}

/**Function*************************************************************

  Synopsis    [ Separete the cubes in present in a divisor ]

  Description [ In this implementation *all* stored divsors are composed 
                of two cubes. 

                In order to save space and to be able to use the Vec_Int_t
                hash table both cubes are stored in the same vector - using
                a little hack to differentiate which literal belongs to each
                cube.

                This function separetes the two cubes in their own vectors
                so that they can be added to the cover.

                *Note* that this also applies to single cube 
                divisors beacuse of the DeMorgan Law: ab = ( a! + !b )! ]

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fxch_DivSepareteCubes(vDiv *Vec_Int_t, vCube0 *Vec_Int_t, vCube1 *Vec_Int_t) {
    var pArray *int32;
    var i, Lit int32;

    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vDiv)) && (((Lit) = Vec_IntEntry(vDiv, i)), 1); i++  {
        if Abc_LitIsCompl(Lit) {
            Vec_IntPush( vCube1, Abc_Lit2Var( Lit ) );
        } else {
            Vec_IntPush( vCube0, Abc_Lit2Var( Lit ) ); }
    }

    if ( Vec_IntSize( vDiv ) == 4 ) && ( Vec_IntSize( vCube0 ) == 3 ) {
        assert(Vec_IntSize( vCube1 ) == 3, "Vec_IntSize( vCube1 ) == 3");

        pArray = Vec_IntArray( vCube0 );
        if pArray[1] > pArray[2] {
            // inlining of macro ABC_SWAP
             var t int32  = pArray[1]; pArray[1] = pArray[2]; pArray[2] = t; };

        pArray = Vec_IntArray( vCube1 );
        if pArray[1] > pArray[2] {
            // inlining of macro ABC_SWAP
             var t int32  = pArray[1]; pArray[1] = pArray[2]; pArray[2] = t; };
    }
}

/**Function*************************************************************

  Synopsis    [ Removes the literals present in the divisor from their 
               original cubes. ]

  Description [ This function returns the numeber of removed literals
                which should be equal to the size of the divisor. ]

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fxch_DivRemoveLits(vCube0 *Vec_Int_t, vCube1 *Vec_Int_t, vDiv *Vec_Int_t, fCompl *int32) int32 {
    var i, Lit, CountP, CountN, Count, ret int32;

    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vDiv)) && (((Lit) = Vec_IntEntry(vDiv, i)), 1); i++  {
        if Abc_LitIsCompl( Abc_Lit2Var( Lit ) ) {
            CountN += Vec_IntRemove1( vCube0, Abc_Lit2Var( Lit ) );
        } else {
            CountP += Vec_IntRemove1( vCube0, Abc_Lit2Var( Lit ) ); }
    }

    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vDiv)) && (((Lit) = Vec_IntEntry(vDiv, i)), 1); i++  {
        Count += Vec_IntRemove1( vCube1, Abc_Lit2Var( Lit ) );
    }

   if Vec_IntSize( vDiv ) == 2 {
       // inlining of macro Vec_IntForEachEntry
       for  i = 0; (i < Vec_IntSize(vDiv)) && (((Lit) = Vec_IntEntry(vDiv, i)), 1); i++  {
           Vec_IntRemove1( vCube0, Abc_LitNot( Abc_Lit2Var( Lit ) ) );
           Vec_IntRemove1( vCube1, Abc_LitNot( Abc_Lit2Var( Lit ) ) );
       } }

    ret = Count + CountP + CountN;

    if Vec_IntSize( vDiv ) == 4 {
        var Lit0 int32 = Abc_Lit2Var( Vec_IntEntry( vDiv, 0 ) );
        var Lit1 int32 = Abc_Lit2Var( Vec_IntEntry( vDiv, 1 ) );
        var Lit2 int32 = Abc_Lit2Var( Vec_IntEntry( vDiv, 2 ) );
        var Lit3 int32 = Abc_Lit2Var( Vec_IntEntry( vDiv, 3 ) );

        if Lit0 == Abc_LitNot( Lit1 ) && Lit2 == Abc_LitNot( Lit3 ) && CountN == 1 {
            *fCompl = 1; }

        if Lit0 == Abc_LitNot( Lit1 ) && ret == 2 {
            *fCompl = 1;

            // inlining of macro Vec_IntForEachEntry
            for  i = 0; (i < Vec_IntSize(vDiv)) && (((Lit) = Vec_IntEntry(vDiv, i)), 1); i++  {
                ret += Vec_IntRemove1( vCube0, ( Abc_Lit2Var( Lit ) ^ (fCompl && i > 1) ) );
            }

            // inlining of macro Vec_IntForEachEntry
            for  i = 0; (i < Vec_IntSize(vDiv)) && (((Lit) = Vec_IntEntry(vDiv, i)), 1); i++  {
                ret += Vec_IntRemove1( vCube1, ( Abc_Lit2Var( Lit ) ^ (fCompl && i > 1) ) );
            }
        }
    }

    return ret;
}

/**Function*************************************************************

  Synopsis    [ Print the divisor identified by iDiv. ]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fxch_DivPrint(pFxchMan *Fxch_Man_t, iDiv int32) {
    var vDiv *Vec_Int_t  = Hsh_VecReadEntry( pFxchMan.pDivHash, iDiv );
    var i, Lit int32;

    fmt.Printf("Div %7d : ", iDiv);
    fmt.Printf("Weight %12.5f  ", Vec_FltEntry( pFxchMan.vDivWeights, iDiv ));

    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vDiv)) && (((Lit) = Vec_IntEntry(vDiv, i)), 1); i++  {
        if !Abc_LitIsCompl( Lit ) {
            fmt.Printf("%d(1)", Abc_Lit2Var( Lit )); }
    }

    fmt.Printf(" + ");

    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vDiv)) && (((Lit) = Vec_IntEntry(vDiv, i)), 1); i++  {
        if Abc_LitIsCompl( Lit ) {
            fmt.Printf("%d(2)", Abc_Lit2Var( Lit )); }
    }

    fmt.Printf(" Lits =%7d  ", pFxchMan.nLits);
    fmt.Printf("Divs =%8d  \n", Hsh_VecSize( pFxchMan.pDivHash ));
}

func Fxch_DivIsNotConstant1(vDiv *Vec_Int_t) int32 {
    var Lit0 int32 = Abc_Lit2Var( Vec_IntEntry( vDiv, 0 ) );
    var Lit1 int32 = Abc_Lit2Var( Vec_IntEntry( vDiv, 1 ) );

    if ( Vec_IntSize( vDiv ) == 2 )  && ( Lit0 == Abc_LitNot( Lit1 ) ) {
            return 0; }

    return 1;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END
