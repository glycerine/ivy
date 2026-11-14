#include "proof/cec/cecInt.h"

int main(void)
{
    Gia_Man_t *g = Gia_ManStart(16);
    Cec_ParSat_t pars;
    Cec_ManSat_t *p;
    int a = Gia_ManAppendCi(g), b = Gia_ManAppendCi(g), c = Gia_ManAppendCi(g);
    int mid = Gia_ManAppendAnd(g,a,b), root = Gia_ManAppendAnd(g,mid,c);
    int yes = Gia_ManAppendAnd(g,a,b), no = Gia_ManAppendAnd(g,a^1,c);
    int mux = Gia_ManAppendAnd(g,yes^1,no^1);
    int queries[10][2] = {{root,-1},{a,-1},{root^1,-1},{root^1,a},
        {root,a^1},{root,root^1},{mux,-1},{mux^1,root^1},{mux,root},{root,-1}};
    int i,id,status;
    Gia_ManCreateRefs(g);
    Gia_ManFillValue(g);
    Cec_ManSatSetDefaultParams(&pars);
    pars.fPolarFlip = 0;
    pars.nBTLimit = 1000;
    p = Cec_ManSatCreate(g,&pars);
    for(i=0;i<10;++i)
    {
        Gia_Obj_t *x = Gia_ObjFromLit(g,queries[i][0]);
        status = queries[i][1] < 0 ? Cec_ManSatCheckNode(p,x) :
            Cec_ManSatCheckNodeTwo(p,x,Gia_ObjFromLit(g,queries[i][1]));
        printf("query %d %d vars %d clauses %u since %d ids",i,status,p->nSatVars,p->pSat->stats.clauses,p->nCallsSince);
        for(id=0;id<g->nObjs;++id) printf(" %d",p->pSatVars[id]);
        puts("");
    }
    Cec_ManSatStop(p);
    Gia_ManStop(g);
    return 0;
}
