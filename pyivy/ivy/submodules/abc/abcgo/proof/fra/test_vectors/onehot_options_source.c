#include "proof/fra/fraHot.c"
int main(void) {
    Aig_Man_t *g = Aig_ManStart(16);
    Aig_Obj_t *a[3]; Fra_Man_t p = {0};
    Vec_Ptr_t *groups = Vec_PtrAlloc(1); Vec_Int_t *group = Vec_IntAlloc(3);
    p.pManFraig = g; p.pSat = sat_solver_new(); p.pMemSatNums = ABC_CALLOC(int, 16);
    for (int i = 0; i < 3; i++) { a[i] = Aig_ObjCreateCi(g); a[i]->pData = &p; p.pMemSatNums[a[i]->Id] = i+1; Vec_IntPush(group, i); }
    Vec_PtrPush(groups, group); Fra_OneHotAddKnownConstraint(&p, groups);
    printf("known=");
    for (int mask = 0; mask < 8; mask++) {
        lit assumps[3]; for (int i=0;i<3;i++) assumps[i]=toLitCond(i+1, !(mask & (1<<i)));
        printf("%d", sat_solver_solve(p.pSat, assumps, assumps+3, 0,0,0,0)==l_True);
    }
    puts(""); sat_solver_delete(p.pSat);
    {
        Fra_Par_t pars = {0}; Vec_Int_t *clauses = Vec_IntAlloc(4); lit unit;
        pars.nFramesK = 1; p.pPars = &pars; Aig_ManSetRegNum(g, 2); p.pSat = sat_solver_new();
        unit = toLit(2); sat_solver_addclause(p.pSat, &unit, &unit+1);
        unit = toLit(3); sat_solver_addclause(p.pSat, &unit, &unit+1);
        Vec_IntPush(clauses, 0); Vec_IntPush(clauses, 0);
        Vec_IntPush(clauses, Fra_RegToLit(0, 1)); Vec_IntPush(clauses, Fra_RegToLit(1, 1));
        Fra_OneHotAssume(&p, clauses); printf("contradiction=%d\n", p.pSat == NULL); Vec_IntFree(clauses);
    }
    Vec_IntFree(group); Vec_PtrFree(groups); ABC_FREE(p.pMemSatNums); Aig_ManStop(g); return 0;
}
