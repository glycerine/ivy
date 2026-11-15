#include "aig/gia/giaFx.c"
int main(void)
{
    int n,seed,reverse,i,j,lit,ins[10];
    for(n=2;n<=9;n++)for(seed=0;seed<8;seed++)for(reverse=0;reverse<2;reverse++) {
        Gia_Man_t *p=Gia_ManStart(64);Vec_Wec_t *cubes;Vec_Str_t *phases;
        for(i=0;i<n;i++) ins[i]=Gia_ManAppendCi(p);
        lit=ins[0]^(seed&1);
        for(i=1;i<n;i++)lit=Gia_ManAppendAnd(p,lit,ins[i]^((seed>>i)&1));
        Gia_ManAppendCo(p,lit);
        p->vMapping=Vec_IntStart(Gia_ManObjNum(p));
        Vec_IntWriteEntry(p->vMapping,lit>>1,Vec_IntSize(p->vMapping));Vec_IntPush(p->vMapping,n);
        for(i=n-1;i>=0;i--)Vec_IntPush(p->vMapping,ins[i]>>1);Vec_IntPush(p->vMapping,lit>>1);
        cubes=Gia_ManFxRetrieve(p,&phases,reverse);
        printf("n%d seed%d reverse%d phases",n,seed,reverse);
        for(i=0;i<Vec_StrSize(phases);i++)printf("%d",Vec_StrEntry(phases,i));
        for(i=0;i<Vec_WecSize(cubes);i++) {Vec_Int_t *row=Vec_WecEntry(cubes,i);printf(" [");for(j=0;j<Vec_IntSize(row);j++)printf("%s%d",j?",":"",Vec_IntEntry(row,j));printf("]");}
        puts("");Vec_WecFree(cubes);Vec_StrFree(phases);Gia_ManStop(p);
    }
    for(seed=0;seed<2;seed++) {
        Gia_Man_t *p=Gia_ManStart(16);Vec_Wec_t *cubes;Vec_Str_t *phases;
        int a=Gia_ManAppendCi(p),b,zero;
        p->fGiaSimple=1;
        zero=Gia_ManAppendAnd(p,a,a^1);
        if(seed==0) lit=zero;
        else {b=Gia_ManAppendCi(p);i=Gia_ManAppendAnd(p,b,b^1);lit=Gia_ManAppendAnd(p,zero^1,i^1);}
        Gia_ManAppendCo(p,lit);
        p->vMapping=Vec_IntStart(Gia_ManObjNum(p));
        Vec_IntWriteEntry(p->vMapping,lit>>1,Vec_IntSize(p->vMapping));Vec_IntPush(p->vMapping,1+seed);
        for(i=0;i<1+seed;i++)Vec_IntPush(p->vMapping,Gia_ObjId(p,Gia_ManCi(p,i)));Vec_IntPush(p->vMapping,lit>>1);
        cubes=Gia_ManFxRetrieve(p,&phases,0);
        printf("constant%d phases",seed);
        for(i=0;i<Vec_StrSize(phases);i++)printf("%d",Vec_StrEntry(phases,i));
        for(i=0;i<Vec_WecSize(cubes);i++) {Vec_Int_t *row=Vec_WecEntry(cubes,i);printf(" [");for(j=0;j<Vec_IntSize(row);j++)printf("%s%d",j?",":"",Vec_IntEntry(row,j));printf("]");}
        puts("");Vec_WecFree(cubes);Vec_StrFree(phases);Gia_ManStop(p);
    }
    return 0;
}
