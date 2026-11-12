package aig

import "fmt"



//#include "aig.h"

//ABC_NAMESPACE_IMPL_START


func Aig_ProcedureTest() {
    var p *Aig_Man_t;
    var pA, pB, pC *Aig_Obj_t;
    var pFunc *Aig_Obj_t;
    var pFunc2 *Aig_Obj_t;

    p = Aig_ManStart( 1000 );

    pA = Aig_IthVar( p, 0 );
    pB = Aig_IthVar( p, 1 );
    pC = Aig_IthVar( p, 2 );

    pFunc = Aig_Mux( p, pA, pB, pC );
    pFunc2 = Aig_And( p, pA, pB );

    Aig_ObjCreatePo( p, pFunc );
    Aig_ObjCreatePo( p, pFunc2 );

    Aig_ManSetRegNum( p, 1 );

    Aig_ManCleanup( p );

    if !Aig_ManCheck( p ) {
        fmt.Printf("Check has failed\n");
    }

    Aig_ManDumpBlif( p, "aig_test_file.blif", nil, nil );
    Aig_ManStop( p );
}
//ABC_NAMESPACE_IMPL_END

