#define ABC_USE_CUDD
#define ABC_USE_PTHREADS
#include "map/if/ifDsd.c"
int main(void)
{
    If_DsdMan_t *p=If_DsdManAlloc(4,0),*q,*dst;Vec_Wrd_t *config;
    unsigned char perm[4];int i;word truth;
    truth=*Dau_DsdToTruth("(ab)",2);If_DsdManCompute(p,&truth,2,perm,"");
    for(i=0;i<3;i++){truth=*Dau_DsdToTruth("((ab)[cd])",4);If_DsdManCompute(p,&truth,4,perm,"");}
    q=If_DsdManFilter(p,2);
    printf("meta filter %d %016llx\n",If_DsdManObjNum(q),(unsigned long long)*If_DsdManComputeTruth(q,2*(If_DsdManObjNum(q)-1),NULL));
    If_DsdManFree(q,0);dst=If_DsdManAlloc(4,0);
    truth=*Dau_DsdToTruth("[ab]",2);If_DsdManCompute(dst,&truth,2,perm,"");
    If_DsdManMerge(dst,p);
    printf("meta merge %d %016llx\n",If_DsdManObjNum(dst),(unsigned long long)*If_DsdManComputeTruth(dst,2*(If_DsdManObjNum(dst)-1),NULL));
    If_DsdManFree(dst,0);If_DsdManFree(p,0);
    p=If_DsdManAlloc(2,0);Id_DsdManTuneStr1(p,"{ab}",0,0);config=p->vConfigs;
    If_DsdManInvertMarks(p,0);
    printf("\nmeta invert %d %d %d",p->pCellStr!=NULL,p->vConfigs!=NULL,p->vConfigs==config);
    for(i=0;i<If_DsdManObjNum(p);i++)printf(" %d",If_DsdVecObjMark(&p->vObjs,i));puts("");
    If_DsdManCleanMarks(p,0);
    printf("meta clean %d %d",p->pCellStr!=NULL,p->vConfigs!=NULL);
    for(i=0;i<If_DsdManObjNum(p);i++)printf(" %d",If_DsdVecObjMark(&p->vObjs,i));puts("");
    If_DsdManFree(p,0);return 0;
}
