package bmc

import "fmt"

/**CFile****************************************************************

  FileName    [bmcMesh.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [SAT-based bounded model checking.]

  Synopsis    [Synthesis for mesh of LUTs.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: bmcMesh.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "bmc.h"
//#include "sat/satoko/satoko.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

const NCPARS = 16

func Bmc_MeshTVar(Me [][102]int32, x int32, y int32) int32 { return Me[x][y];                                         }
func Bmc_MeshGVar(Me [][102]int32, x int32, y int32) int32 { return Me[x][y] + Me[101][100];                          }
func Bmc_MeshCVar(Me [][102]int32, x int32, y int32) int32 { return Me[x][y] + Me[101][100] + Me[101][101];           }
func Bmc_MeshUVar(Me [][102]int32, x int32, y int32) int32 { return Me[x][y] + Me[101][100] + Me[101][101] + NCPARS;  }

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bmc_MeshVarValue(p *satoko_t, v int32) int32 {
//    int value = var_value(p, v) != SATOKO_VAR_UNASSING ? var_value(p, v) : satoko_var_polarity(p, v);
//    return value == SATOKO_LIT_TRUE;
    return satoko_var_polarity(p, v) == SATOKO_LIT_TRUE;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bmc_MeshAddOneHotness(pSat *satoko_t, iFirst int32, iLast int32) int32 {
    var i int32;
    var j int32;
    var v int32;
    var pVars [100]int32;
    var nVars int32;
    var nCount int32;
    assert(iFirst < iLast && iFirst + 110 > iLast, "iFirst < iLast && iFirst + 110 > iLast");
    for  v = iFirst; v < iLast; v++  {
        if Bmc_MeshVarValue(pSat, v) != 0 { // v = 1
        
            assert(nVars < 100, "nVars < 100");
            pVars[ nVars++ ] = v;
        }
    }
    if nVars <= 1 {
        return 0; }
    for  i = 0;   i < nVars; i++  {
    for  j = i+1; j < nVars; j++  {
        var pLits [2]int32;
        var RetValue int32;
        pLits[0] = Abc_Var2Lit( pVars[i], 1 );
        pLits[1] = Abc_Var2Lit( pVars[j], 1 );
        RetValue = satoko_add_clause( pSat, pLits, 2 );  assert((RetValue != 0), "RetValue");
        nCount++;
    }
    }
    return nCount;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bmc_MeshTest(p *Gia_Man_t, X int32, Y int32, T int32, fVerbose int32) {
    var clk abctime  = Abc_Clock();
    var pSat *satoko_t  = satoko_create();
    var pObj *Gia_Obj_t;
    var Me [102][102]int32 = [102][102]int32{{0}};
    var pN [102][2]int32 = [102][2]int32{{0}};
    var I int32  = Gia_ManPiNum(p);
    var G int32  = I + Gia_ManAndNum(p);
    var i, x, y, t, g, c, status, RetValue, Lit, iVar, nClauses int32;

    assert(X <= 100 && Y <= 100 && T <= 100 && G <= 100, "X <= 100 && Y <= 100 && T <= 100 && G <= 100");

    // init the graph
    for  i = 0; i < I; i++  {
        pN[i][0] = pN[i][1] = -1;
    }
    // inlining of macro Gia_ManForEachAnd
    for i = 0; i < p.nObjs; i++ {
    	pObj = Gia_ManObj(p, i)
    	if pObj == nil { break }
    	if !Gia_ObjIsAnd(pObj) {} else {
        pN[i-1][0] = Gia_ObjFaninId0(pObj, i)-1;
        pN[i-1][1] = Gia_ObjFaninId1(pObj, i)-1;
    }
    }
    if fVerbose != 0 {
        fmt.Printf("The graph has %d inputs: ", Gia_ManPiNum(p));
        for  i = 0; i < I; i++  {
            fmt.Printf("%c ", 'a' + i);
        }
        fmt.Printf("  and %d nodes: ", Gia_ManAndNum(p));
        for  i = I; i < G; i++  {
            fmt.Printf("%c=%c%c ", 'a' + i, 'a' + pN[i][0], 'a' + pN[i][1]);
        }
        fmt.Printf("\n");
    }

    // init SAT variables (time vars + graph vars + config vars)
    // config variables: 16 = 4 buff vars + 12 node vars
    iVar = 0;
    for  y = 0; y < Y; y++  {
    for  x = 0; x < X; x++  {
        //printf( "%3d %3d %3d    %s", iVar, iVar+T, iVar+T+G, x == X-1 ? "\n":"" );
        Me[x][y] = iVar;
        iVar += T + G + NCPARS + 1;
    }
    }
    Me[101][100] = T;
    Me[101][101] = G;
    if fVerbose != 0 {
        fmt.Printf("SAT variable count is %d (%d time vars + %d graph vars + %d config vars + %d aux vars)\n", iVar, X*Y*T, X*Y*G, X*Y*NCPARS, X*Y); }

    // add constraints

    // time 0 and primary inputs only on the boundary
    for  x = 0; x < X; x++  {
    for  y = 0; y < Y; y++  {
        var iTVar int32  = Bmc_MeshTVar( Me, x, y );
        var iGVar int32  = Bmc_MeshGVar( Me, x, y );

        if x == 0 || x == X-1 || y == 0 || y == Y-1 { // boundary
        
            // time 0 is required
            for  t = 0; t < T; t++  {
                Lit = Abc_Var2Lit( iTVar+t, int32((t > 0)) );
                RetValue = satoko_add_clause( pSat, &Lit, 1 );  assert((RetValue != 0), "RetValue");
            }
            // internal nodes are not allowed
            for  g = I; g < G; g++  {
                Lit = Abc_Var2Lit( iGVar+g, 1 );
                RetValue = satoko_add_clause( pSat, &Lit, 1 );  assert((RetValue != 0), "RetValue");
            }
        
        } else { // not a boundary
        
            Lit = Abc_Var2Lit( iTVar, 1 );  // cannot have time 0
            RetValue = satoko_add_clause( pSat, &Lit, 1 );  assert((RetValue != 0), "RetValue");
        }
    }
    }
    for  x = 1; x < X-1; x++  {
    for  y = 1; y < Y-1; y++  {
        var pLits [100]int32;
        var nLits int32;

        var iTVar int32  = Bmc_MeshTVar( Me, x, y );
        var iGVar int32  = Bmc_MeshGVar( Me, x, y );
        var iCVar int32  = Bmc_MeshCVar( Me, x, y );
        var iUVar int32  = Bmc_MeshUVar( Me, x, y );

        // 0=left  1=up  2=right  3=down
        var iTVars [4]int32; 
        var iGVars [4]int32;

        iTVars[0] = Bmc_MeshTVar( Me, x-1, y );
        iGVars[0] = Bmc_MeshGVar( Me, x-1, y );

        iTVars[1] = Bmc_MeshTVar( Me, x, y-1 );
        iGVars[1] = Bmc_MeshGVar( Me, x, y-1 );

        iTVars[2] = Bmc_MeshTVar( Me, x+1, y );
        iGVars[2] = Bmc_MeshGVar( Me, x+1, y );

        iTVars[3] = Bmc_MeshTVar( Me, x, y+1 );
        iGVars[3] = Bmc_MeshGVar( Me, x, y+1 );

        // condition when cell is used
        for  g = 0; g < G; g++  {
            pLits[0] = Abc_Var2Lit( iGVar+g, 1 );
            pLits[1] = Abc_Var2Lit( iUVar, 0 );
            RetValue = satoko_add_clause( pSat, pLits, 2 );  assert((RetValue != 0), "RetValue");
            nClauses++;
        }

        // at least one time is used
        pLits[0] = Abc_Var2Lit( iUVar, 1 );
        for  t = 1; t < T; t++  {
            pLits[t] = Abc_Var2Lit( iTVar+t, 0 );
        }
        RetValue = satoko_add_clause( pSat, pLits, T );  assert((RetValue != 0), "RetValue");
        nClauses++;

        // at least one config is used
        pLits[0] = Abc_Var2Lit( iUVar, 1 );
        for  c = 0; c < NCPARS; c++  {
            pLits[c+1] = Abc_Var2Lit( iCVar+c, 0 );
        }
        RetValue = satoko_add_clause( pSat, pLits, NCPARS+1 );  assert((RetValue != 0), "RetValue");
        nClauses++;

        // constraints for each time
        for  t = 1; t < T; t++  {
            var Conf [12][2]int32 = [12][2]int32{{0, 1}, {0, 2}, {0, 3}, {1, 2}, {1, 3}, {2, 3},   {1, 0}, {2, 0}, {3, 0}, {2, 1}, {3, 1}, {3, 2}};
            // buffer
            for  g = 0; g < G; g++  {
            for  c = 0; c < 4; c++  {
                nLits = 0;
                pLits[ nLits++ ] = Abc_Var2Lit( iTVar+t, 1 );
                pLits[ nLits++ ] = Abc_Var2Lit( iGVar+g, 1 );
                pLits[ nLits++ ] = Abc_Var2Lit( iCVar+c, 1 );
                pLits[ nLits++ ] = Abc_Var2Lit( iTVars[c]+t-1, 0 );
                RetValue = satoko_add_clause( pSat, pLits, nLits );  assert((RetValue != 0), "RetValue");

                nLits = 0;
                pLits[ nLits++ ] = Abc_Var2Lit( iTVar+t, 1 );
                pLits[ nLits++ ] = Abc_Var2Lit( iGVar+g, 1 );
                pLits[ nLits++ ] = Abc_Var2Lit( iCVar+c, 1 );
                pLits[ nLits++ ] = Abc_Var2Lit( iGVars[c]+g, 0 );
                RetValue = satoko_add_clause( pSat, pLits, nLits );  assert((RetValue != 0), "RetValue");

                nClauses += 2;
            }
            }
            for  g = 0; g < I; g++  {
            for  c = 4; c < NCPARS; c++  {
                pLits[0] = Abc_Var2Lit( iGVar+g, 1 );
                pLits[1] = Abc_Var2Lit( iCVar+c, 1 );
                RetValue = satoko_add_clause( pSat, pLits, 2 );  assert((RetValue != 0), "RetValue");
                nClauses++;
            }
            }
            // node
            for  g = I; g < G; g++  {
            for  c = 0; c < 12; c++  {
                assert(pN[g][0] >= 0 && pN[g][1] >= 0, "pN[g][0] >= 0 && pN[g][1] >= 0");

                nLits = 0;
                pLits[ nLits++ ] = Abc_Var2Lit( iTVar+t, 1 );
                pLits[ nLits++ ] = Abc_Var2Lit( iGVar+g, 1 );
                pLits[ nLits++ ] = Abc_Var2Lit( iCVar+c+4, 1 );
                pLits[ nLits++ ] = Abc_Var2Lit( iTVars[Conf[c][0]]+t-1, 0 );
                RetValue = satoko_add_clause( pSat, pLits, nLits );  assert((RetValue != 0), "RetValue");

                nLits = 0;
                pLits[ nLits++ ] = Abc_Var2Lit( iTVar+t, 1 );
                pLits[ nLits++ ] = Abc_Var2Lit( iGVar+g, 1 );
                pLits[ nLits++ ] = Abc_Var2Lit( iCVar+c+4, 1 );
                pLits[ nLits++ ] = Abc_Var2Lit( iTVars[Conf[c][1]]+t-1, 0 );
                RetValue = satoko_add_clause( pSat, pLits, nLits );  assert((RetValue != 0), "RetValue");


                nLits = 0;
                pLits[ nLits++ ] = Abc_Var2Lit( iTVar+t, 1 );
                pLits[ nLits++ ] = Abc_Var2Lit( iGVar+g, 1 );
                pLits[ nLits++ ] = Abc_Var2Lit( iCVar+c+4, 1 );
                pLits[ nLits++ ] = Abc_Var2Lit( iGVars[Conf[c][0]]+pN[g][0], 0 );
                RetValue = satoko_add_clause( pSat, pLits, nLits );  assert((RetValue != 0), "RetValue");

                nLits = 0;
                pLits[ nLits++ ] = Abc_Var2Lit( iTVar+t, 1 );
                pLits[ nLits++ ] = Abc_Var2Lit( iGVar+g, 1 );
                pLits[ nLits++ ] = Abc_Var2Lit( iCVar+c+4, 1 );
                pLits[ nLits++ ] = Abc_Var2Lit( iGVars[Conf[c][1]]+pN[g][1], 0 );
                RetValue = satoko_add_clause( pSat, pLits, nLits );  assert((RetValue != 0), "RetValue");

                nClauses += 4;
            }
            }
        }
    }
    }

    // final condition
    {
        var iGVar int32  = Bmc_MeshGVar( Me, 1, 1 ) + G-1;
        Lit = Abc_Var2Lit( iGVar, 0 );
        RetValue = satoko_add_clause( pSat, &Lit, 1 );  
        if RetValue == 0 {
            fmt.Printf("Problem has no solution. ");
            Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
            satoko_destroy( pSat );
            return;
        }
    }

    if fVerbose != 0 {
        fmt.Printf("Finished adding %d clauses. Started solving...\n", nClauses); }

    for {
        var nAddClauses int32;
        status = satoko_solve( pSat );
        if status == SATOKO_UNSAT {
            fmt.Printf("Problem has no solution. ");
            break;
        }
        if status == SATOKO_UNDEC {
            fmt.Printf("Computation timed out. ");
            break;
        }
        assert(status == SATOKO_SAT, "status == SATOKO_SAT");
        // check if the solution is valid and add constraints
        for  x = 0; x < X; x++  {
        for  y = 0; y < Y; y++  {
            if x == 0 || x == X-1 || y == 0 || y == Y-1 { // boundary
            
                var iGVar int32  = Bmc_MeshGVar( Me, x, y );
                nAddClauses += Bmc_MeshAddOneHotness( pSat, iGVar, iGVar + G );
            
            } else {
                var iTVar int32  = Bmc_MeshTVar( Me, x, y );
                var iGVar int32  = Bmc_MeshGVar( Me, x, y );
                var iCVar int32  = Bmc_MeshCVar( Me, x, y );
                nAddClauses += Bmc_MeshAddOneHotness( pSat, iTVar, iTVar + T );
                nAddClauses += Bmc_MeshAddOneHotness( pSat, iGVar, iGVar + G );
                nAddClauses += Bmc_MeshAddOneHotness( pSat, iCVar, iCVar + NCPARS );
            }
        }
        }
        if nAddClauses > 0 {
            fmt.Printf("Adding %d one-hotness clauses.\n", nAddClauses);
            continue;
        }
        fmt.Printf("Satisfying solution found. ");
/*
        iVar = satoko_varnum(pSat);
        for ( i = 0; i < iVar; i++ )
            if ( Bmc_MeshVarValue(pSat, i) )
                printf( "%d ", i );
        printf( "\n" );
*/
        break;
    }
    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
    if status == SATOKO_SAT {
        // count the number of nodes and buffers
        var nBuffs, nNodes int32;
        for  y = 1; y < Y-1; y++  {
        for  x = 1; x < X-1; x++  {
            var iCVar int32  = Bmc_MeshCVar( Me, x, y );
            for  c = 0; c < 4; c++  {
                if Bmc_MeshVarValue(pSat, iCVar+c) != 0 {
                    //printf( "Buffer y=%d x=%d  (var = %d; config = %d)\n", y, x, iCVar+c, c );
                    nBuffs++;
                }
            }
            for  c = 4; c < NCPARS; c++  {
                if Bmc_MeshVarValue(pSat, iCVar+c) != 0 {
                    //printf( "Node   y=%d x=%d  (var = %d; config = %d)\n", y, x, iCVar+c, c );
                    nNodes++;
                }
            }
        }
        }
        fmt.Printf("The %d x %d mesh with latency %d with %d active cells (%d nodes and %d buffers):\n", X, Y, T, nNodes+nBuffs, nNodes, nBuffs);
        // print mesh
        fmt.Printf(" Y\\X ");
        for  x = 0; x < X; x++  {
            fmt.Printf("  %-2d ", x);
        }
        fmt.Printf("\n");
        for  y = 0; y < Y; y++  {
            fmt.Printf(" %-2d  ", y);
            for  x = 0; x < X; x++  {
                var iTVar int32   = Bmc_MeshTVar( Me, x, y );
                var iGVar int32   = Bmc_MeshGVar( Me, x, y );

                var fFound int32;                ;
                for  t = 0; t < T; t++  {
                for  g = 0; g < G; g++  {
                    if (Bmc_MeshVarValue(pSat, iTVar+t) != 0) && (Bmc_MeshVarValue(pSat, iGVar+g) != 0) {
                        fmt.Printf(" %c%-2d ", 'a' + g, t);
                        fFound = 1;
                    }
                }
                }
                if fFound != 0 {
                    continue; }
                if x == 0 || x == X-1 || y == 0 || y == Y-1 { // boundary
                    fmt.Printf("  *  ");
                } else {
                    fmt.Printf("     "); }
            }
            fmt.Printf("\n");
        }
    }
    satoko_destroy( pSat );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

