#include "proof/fra/fraClaus.c"
int main(int argc, char **argv) {
 Aig_Man_t *g=Aig_ManStart(16); Aig_Obj_t *x=Aig_ObjCreateCi(g); Aig_ObjCreateCo(g,x);
 Clu_Man_t *p=Fra_ClausAlloc(g,0,0,4,4,1,2,1,0,0,0,0); p->pCnf=Cnf_DeriveSimple(g,1); p->pSatMain=sat_solver_new(); p->nBTLimit=argc > 1 ? atoi(argv[1]) : 1;
 int base=p->pCnf->nVars+1; lit row[8];
 for(int i=0;i<8;i++){for(int h=0;h<7;h++) row[h]=toLit(base+i*7+h); if(!sat_solver_addclause(p->pSatMain,row,row+7))return 3;}
 for(int h=0;h<7;h++) for(int i=0;i<8;i++)for(int j=0;j<i;j++){row[0]=lit_neg(toLit(base+i*7+h));row[1]=lit_neg(toLit(base+j*7+h));if(!sat_solver_addclause(p->pSatMain,row,row+2))return 4;}
 int status=argc > 2 ? Fra_ClausRunSat0(p) : Fra_ClausRunSat(p); printf("status=%d\n",status); Fra_ClausFree(p);Aig_ManStop(g);return 0;
}
