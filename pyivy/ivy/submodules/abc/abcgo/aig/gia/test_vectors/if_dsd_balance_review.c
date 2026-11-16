#define ABC_USE_CUDD
#define ABC_USE_PTHREADS
#include "map/if/ifDsd.c"
int main(void)
{
    const char *texts[]={"(abc)","E8{a(bc)d}","(ab[cd]<efg>)","<a(bc)[de]>"};
    int sizes[]={3,4,7,5};If_DsdMan_t *p=If_DsdManAlloc(8,0);
    Vec_Int_t *aig=Vec_IntAlloc(100);unsigned char perm[8];int index,mode,lit,n,i,times[8],delay,area,td,ta,pd,support;word profile;
    If_DsdManAllocIsops(p,8);
    for(index=0;index<4;index++) {
        n=sizes[index];lit=If_DsdManCompute(p,Dau_DsdToTruth((char *)texts[index],n),n,perm,"");
        for(mode=0;mode<2;mode++) {
            for(i=0;i<n;i++)times[i]=mode?i%3:0;
            area=ta=0;Vec_IntClear(aig);
            delay=If_CutDsdBalanceEvalInt(p,lit,times,aig,&area,(char *)perm);
            td=If_CutDsdBalanceEvalInt(p,lit,times,NULL,&ta,(char *)perm);
            support=0;profile=0;
            pd=If_CutDsdBalancePinDelays_rec(p,Abc_Lit2Var(lit),times,&profile,&support,n,(char *)perm);
            printf("balance %d %d %d %d %d %d %d %016llx aig",index,mode,delay,area,td,ta,pd,(unsigned long long)profile);
            for(i=0;i<Vec_IntSize(aig);i++)printf(" %d",Vec_IntEntry(aig,i));puts("");
        }
    }
    Vec_IntFree(aig);If_DsdManFree(p,0);return 0;
}
