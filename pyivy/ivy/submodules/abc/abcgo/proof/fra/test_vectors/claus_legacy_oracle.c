#include "proof/fra/fraClaus.c"
int main(void) {
 Aig_Man_t *g=Aig_ManStart(64);Aig_Obj_t *in=Aig_ObjCreateCi(g),*s=Aig_ObjCreateCi(g);
 Aig_ObjCreateCo(g,Aig_And(g,in,s));Aig_ObjCreateCo(g,s);Aig_ManSetRegNum(g,1);
 Clu_Man_t *p=Fra_ClausAlloc(g,1,0,100,4,10,10,1,0,0,0,0); p->nSimWords=16;p->nSimFrames=2;p->nCuts=91;p->fFiltering=1;p->pCnf=Cnf_DeriveSimple(g,Aig_ManCoNum(g));
 int result=Fra_ClausProcessClauses(p,0); printf("result %d cuts %d clauses %d filtering %d\n",result,p->nCuts,p->nClauses,p->fFiltering);
 int i,x; Vec_IntForEachEntry(p->vLits,x,i) printf("lit %d\n",x); Vec_IntForEachEntry(p->vClauses,x,i) printf("end %d\n",x); Vec_IntForEachEntry(p->vCosts,x,i) printf("cost %d\n",x);
 Fra_ClausFree(p);Aig_ManStop(g);return 0;
}
