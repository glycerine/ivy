// Compile with ABC's source include path and link against its object archive.
// Run in a disposable directory: the source should write b_care.aig there.
#include "proof/fra/fraClaus.c"

int main(void)
{
    Aig_Man_t *g = Aig_ManStart(16);
    Aig_Obj_t *x = Aig_ObjCreateCi(g);
    Aig_ObjCreateCo(g, x);
    g->pName = Abc_UtilStrsav("/tmp/a-b.aig");
    Clu_Man_t *p = Fra_ClausAlloc(g, 1, 0, 4, 4, 1, 2, 1, 0, 0, 0, 0);
    p->pCnf = Cnf_DeriveSimple(g, 0);
    Vec_IntPush(p->vLitsProven, 2 * p->pCnf->pVarNums[x->Id]);
    Vec_IntPush(p->vClausesProven, 1);
    Fra_ClausWriteIndClauses(p);
    Fra_ClausFree(p);
    Aig_ManStop(g);
    return 0;
}
