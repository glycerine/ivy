#define ABC_USE_CUDD
#define ABC_USE_PTHREADS
#include "map/if/ifDsd.c"
int main(void)
{
    If_DsdMan_t *p=If_DsdManAlloc(4,0);unsigned char perm[11];int id,j;
    word truth=*Dau_DsdToTruth("((ab)[cd])",4);
    If_DsdManCompute(p,&truth,4,perm,"");
    Id_DsdManTuneStr1(p,"{abcd}",0,0);
    printf("\ncontract %d %d %d\n",If_DsdManObjNum(p),If_DsdManLutSize(p),p->nConfigWords);
    for(id=0;id<If_DsdManObjNum(p);id++) {
        printf("node %d %d",id,If_DsdVecObjMark(&p->vObjs,id));
        for(j=0;j<p->nConfigWords;j++)printf(" %016llx",(unsigned long long)Vec_WrdEntry(p->vConfigs,id*p->nConfigWords+j));puts("");
    }
    If_DsdManFree(p,0);
    p=If_DsdManAlloc(11,0);
    If_DsdManCompute(p,Dau_DsdToTruth("[(abcdefghij)k]",11),11,perm,"");
    If_DsdManTune(p,6,0,0,0,0);
    printf("\ntwo-lut %d",If_DsdManObjNum(p));
    for(id=0;id<If_DsdManObjNum(p);id++)printf(" %d",If_DsdVecObjMark(&p->vObjs,id));puts("");
    If_DsdManFree(p,0);return 0;
}
