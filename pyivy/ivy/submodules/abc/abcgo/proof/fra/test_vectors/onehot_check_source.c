#include "sat/cnf/cnf.h"
#include "proof/fra/fra.h"
extern Aig_Man_t * Fra_FramesWithClasses(Fra_Man_t *);
#include "proof/fra/fraHot.c"
int main(int argc, char **argv) {
    Fra_Par_t pars; Fra_Man_t *p; Aig_Man_t *g; Aig_Obj_t *x, *y, *obj; Cnf_Dat_t *cnf; int i;
    Fra_ParamsDefault(&pars); pars.nFramesK=1; pars.fUse1Hot=1; pars.nSimWords=2;
    if (argc>1) pars.nBTLimitNode=atoi(argv[1]);
    g=Aig_ManStart(16); x=Aig_ObjCreateCi(g); y=Aig_ObjCreateCi(g); Aig_ObjCreateCi(g); Aig_ObjCreateCi(g);
    Aig_ObjCreateCo(g,x); Aig_ObjCreateCo(g,y); Aig_ManSetRegNum(g,2);
    p=Fra_ManStart(g,&pars); p->pSml=Fra_SmlStart(g,0,2,2); p->pManFraig=Fra_FramesWithClasses(p);
    cnf=Cnf_Derive(p->pManFraig,Aig_ManRegNum(p->pManFraig)); p->pSat=(sat_solver *)Cnf_DataWriteIntoSolver(cnf,1,0); p->nSatVars=cnf->nVars;
    Aig_ManForEachObj(p->pManFraig,obj,i) obj->pData=p;
    Fra_ManClean(p,Aig_ManObjNumMax(p->pManFraig));
    Aig_ManForEachObj(p->pManFraig,obj,i) if(cnf->pVarNums[obj->Id]!=-1) {Fra_ObjSetSatNum(obj,cnf->pVarNums[obj->Id]); Fra_ObjSetFaninVec(obj,(Vec_Ptr_t *)1);}
    Cnf_DataFree(cnf); p->vOneHots=Vec_IntAlloc(2); Vec_IntPush(p->vOneHots,-1); Vec_IntPush(p->vOneHots,-2);
    Fra_OneHotAssume(p,p->vOneHots); Fra_OneHotCheck(p,p->vOneHots);
    printf("refinement=%d clauses=%d sat=%d\n",p->pCla->fRefinement,Fra_OneHotCount(p,p->vOneHots),p->nSatCallsSat);
    return 0;
}
