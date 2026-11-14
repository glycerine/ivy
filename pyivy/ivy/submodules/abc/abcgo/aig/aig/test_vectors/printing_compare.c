#include "aig/aig/aig.h"

int main(void)
{
    Aig_Man_t g;
    Aig_Obj_t *o;
    Vec_Vec_t *levels;
    char *names[] = {"a","b","c"};
    int nnodes, nco, nqueries, total, i, j, a, b, mode, lit;
    while (scanf("%d%d%d", &nnodes, &nco, &nqueries) == 3)
    {
        memset(&g,0,sizeof(g)); g.nTravIds=1; total=4+nnodes+nco;
        o=ABC_CALLOC(Aig_Obj_t,total); g.pConst1=o; o[0].Type=AIG_OBJ_CONST1;
        g.vObjs=Vec_PtrAlloc(total); g.vCis=Vec_PtrAlloc(3); g.vCos=Vec_PtrAlloc(nco);
        for (i=1; i<=3; i++) { o[i].Type=AIG_OBJ_CI; o[i].pData=names[i-1]; Vec_PtrPush(g.vCis,o+i); }
        for (i=4; i<4+nnodes; i++)
        {
            assert(scanf("%d%d",&a,&b)==2); o[i].Type=AIG_OBJ_AND;
            o[i].pFanin0=Aig_NotCond(o+a/2,a&1); o[i].pFanin1=Aig_NotCond(o+b/2,b&1);
        }
        for (i=4+nnodes; i<total; i++)
        {
            assert(scanf("%d",&a)==1); o[i].Type=AIG_OBJ_CO;
            o[i].pFanin0=Aig_NotCond(o+a/2,a&1); Vec_PtrPush(g.vCos,o+i);
        }
        for (i=0; i<total; i++)
        {
            o[i].Id=i; g.nObjs[o[i].Type]++; Vec_PtrPush(g.vObjs,o+i);
            if (o[i].pFanin0) Aig_ObjFanin0(o+i)->nRefs++;
            if (o[i].pFanin1) Aig_ObjFanin1(o+i)->nRefs++;
        }
        levels=Vec_VecAlloc(10);
        for (j=0; j<nqueries; j++)
        {
            assert(scanf("%d%d",&mode,&lit)==2);
            if (mode==0)
            {
                printf("EQN:"); Aig_ObjPrintEqn(stdout,Aig_NotCond(o+lit/2,lit&1),levels,0);
                printf("\nVERILOG:"); Aig_ObjPrintVerilog(stdout,Aig_NotCond(o+lit/2,lit&1),levels,0); puts("");
            }
            else if (mode==1) { Aig_ObjPrintVerbose(o+lit/2,0); puts(""); }
            else if (mode==2) Aig_ObjPrintVerboseCone(&g,o+lit/2,0);
            else if (mode==3) Aig_ManPrintVerbose(&g,0);
            else if (mode==4) { Aig_ObjPrint(&g,lit<0 ? NULL : Aig_NotCond(o+lit/2,lit&1)); puts(""); }
            else { assert(mode==5); Aig_ManPrintControlFanouts(&g); }
            printf("traversal %d:",g.nTravIds);
            for (i=0; i<total; i++) printf(" %d",o[i].TravId);
            puts("");
        }
        Vec_VecFree(levels); Vec_PtrFree(g.vObjs); Vec_PtrFree(g.vCis); Vec_PtrFree(g.vCos); free(o);
    }
    return 0;
}
