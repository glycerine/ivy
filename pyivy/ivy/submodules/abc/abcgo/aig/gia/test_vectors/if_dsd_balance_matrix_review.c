#define ABC_USE_CUDD
#define ABC_USE_PTHREADS
#include "map/if/ifDsd.c"
int main(int argc,char **argv)
{
    If_DsdMan_t *p=If_DsdManAlloc(8,0);Vec_Int_t *aig=Vec_IntAlloc(100);FILE *input;
    word truth[4],profile;unsigned char perm[8];int n,index=0,i,lit,times[8],delay,area,td,ta,pd,support;
    if(argc!=2)return 2;input=fopen(argv[1],"r");assert(input);If_DsdManAllocIsops(p,8);
    while(fscanf(input,"%d",&n)==1) {
        for(i=0;i<Abc_TtWordNum(n);i++){unsigned long long value;assert(fscanf(input,"%llx",&value)==1);truth[i]=value;}
        for(i=0;i<n;i++)assert(fscanf(input,"%d",times+i)==1);
        lit=If_DsdManCompute(p,truth,n,perm,"");
        area=ta=0;Vec_IntClear(aig);profile=0;pd=-1;support=0;
        delay=If_CutDsdBalanceEvalInt(p,lit,times,aig,&area,(char *)perm);
        td=If_CutDsdBalanceEvalInt(p,lit,times,NULL,&ta,(char *)perm);
        if(delay>=0)pd=If_CutDsdBalancePinDelays_rec(p,Abc_Lit2Var(lit),times,&profile,&support,n,(char *)perm);
        printf("balance %d %d %d %d %d %d %016llx aig",index++,delay,area,td,ta,pd,(unsigned long long)profile);
        for(i=0;i<Vec_IntSize(aig);i++)printf(" %d",Vec_IntEntry(aig,i));puts("");
    }
    fclose(input);Vec_IntFree(aig);If_DsdManFree(p,0);return 0;
}
