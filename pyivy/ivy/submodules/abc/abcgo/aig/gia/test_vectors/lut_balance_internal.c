/* Compile the actual implementation to expose its static cut/pair helpers. */
#include "aig/gia/giaBalLut.c"
#pragma GCC diagnostic error "-Wimplicit-function-declaration"
static unsigned random_word(unsigned *state)
{
    *state = *state * 1664525u + 1013904223u;
    return *state;
}
static void simplify_cases(Vec_Int_t *values, int minimum)
{
    int mode,i,lit;
    for(mode=0;mode<2;mode++){
        Vec_Int_t *copy=Vec_IntAlloc(8);
        printf("simplify %d",mode);
        Vec_IntForEachEntry(values,lit,i)printf(" %d",lit);
        printf(" =>");
        for(i=Vec_IntSize(values)-1;i>=0;i--)Vec_IntPush(copy,Vec_IntEntry(values,i));
        Vec_IntSort(copy,0);
        if(mode)Gia_ManSimplifyXor(copy);else Gia_ManSimplifyAnd(copy);
        Vec_IntForEachEntry(copy,lit,i)printf(" %d",lit);
        puts("");Vec_IntFree(copy);
    }
    if(Vec_IntSize(values)==8)return;
    for(lit=minimum;lit<8;lit++){
        Vec_IntPush(values,lit);simplify_cases(values,lit);Vec_IntPop(values);
    }
}
static void merge_cases(void)
{
    int leaves[6]={1,2,65,66,129,130},mask0,mask1,size,i,j,ok;
    for(mask0=1;mask0<64;mask0++)for(mask1=1;mask1<64;mask1++)for(size=2;size<=6;size++){
        Bal_Cut_t a={0},b={0},out={0};a.Delay=2;b.Delay=3;
        for(i=0;i<6;i++){
            if((mask0>>i)&1)a.pLeaves[a.nLeaves++]=leaves[i];
            if((mask1>>i)&1)b.pLeaves[b.nLeaves++]=leaves[i];
        }
        /* Respect the native merge's input contract: each cut fits K. */
        if((int)a.nLeaves>size || (int)b.nLeaves>size)continue;
        a.Sign=Bal_CutGetSign(a.pLeaves,a.nLeaves);b.Sign=Bal_CutGetSign(b.pLeaves,b.nLeaves);
        ok=Bal_CutMergeOrder(&a,&b,&out,size);
        printf("merge %d %d %d %d",mask0,mask1,size,ok);
        if(ok){printf(" %d %u %llu",out.Delay,out.iFunc,(unsigned long long)out.Sign);for(j=0;j<(int)out.nLeaves;j++)printf(" %d",out.pLeaves[j]);}
        puts("");
    }
}
static void mux_merge_cases(void)
{
    int leaves[4]={1,2,65,66},m0,m1,m2,size,i,j,ok;
    for(m0=1;m0<16;m0++)for(m1=1;m1<16;m1++)for(m2=1;m2<16;m2++)for(size=2;size<=6;size++){
        Bal_Cut_t a={0},b={0},c={0},out={0};a.Delay=2;b.Delay=3;c.Delay=4;
        for(i=0;i<4;i++){
            if((m0>>i)&1)a.pLeaves[a.nLeaves++]=leaves[i];
            if((m1>>i)&1)b.pLeaves[b.nLeaves++]=leaves[i];
            if((m2>>i)&1)c.pLeaves[c.nLeaves++]=leaves[i];
        }
        if((int)a.nLeaves>size || (int)b.nLeaves>size || (int)c.nLeaves>size)continue;
        a.Sign=Bal_CutGetSign(a.pLeaves,a.nLeaves);b.Sign=Bal_CutGetSign(b.pLeaves,b.nLeaves);c.Sign=Bal_CutGetSign(c.pLeaves,c.nLeaves);
        ok=Bal_CutMergeOrderMux(&a,&b,&c,&out,size);
        printf("muxmerge %d %d %d %d %d",m0,m1,m2,size,ok);
        if(ok){printf(" %d %u %llu",out.Delay,out.iFunc,(unsigned long long)out.Sign);for(j=0;j<(int)out.nLeaves;j++)printf(" %d",out.pLeaves[j]);}
        puts("");
    }
}
static void collector_boundaries(void)
{
    int mode,extra,i,j,root,base,inputs[64],lit;
    for(mode=0;mode<2;mode++)for(extra=0;extra<4;extra++){
        Gia_Man_t *g=Gia_ManStart(1);g->pMuxes=ABC_CALLOC(unsigned,g->nObjsAlloc);
        for(i=0;i<64;i++)inputs[i]=Gia_ManAppendCi(g);
        root=base=mode?Gia_ManAppendXorReal(g,inputs[0],inputs[1]):Gia_ManAppendAnd(g,inputs[0],inputs[1]);
        for(i=2;i<64;i++)root=mode?Gia_ManAppendXorReal(g,root,inputs[i]):Gia_ManAppendAnd(g,root,inputs[i]);
        Gia_ManAppendCo(g,root);
        for(i=0;i<extra;i++)Gia_ManAppendCo(g,base);
        Gia_ManCreateRefs(g);
        for(j=0;j<2;j++){
            Gia_ManSuperCollect(g,Gia_ManObj(g,Abc_Lit2Var(j?base:root)));
            printf("boundary %d %d %d",mode,extra,j);
            Vec_IntForEachEntry(g->vSuper,lit,i)printf(" %d",lit);
            puts("");
        }
        Gia_ManStop(g);
    }
}
int main(void)
{
    int size, cuts, seed, step, i, j, n, lit, cost, pair;
    for(size=2;size<=6;size++)for(cuts=2;cuts<=8;cuts++)for(seed=1;seed<=16;seed++){
        Gia_Man_t *g=Gia_ManStart(1); Bal_Man_t *p;
        int pool[64], length=2; unsigned state=seed;
        pool[0]=0;pool[1]=1;
        for(i=0;i<6;i++)pool[length++]=Gia_ManAppendCi(g);
        g->pMuxes=ABC_CALLOC(unsigned,g->nObjsAlloc);
        p=Bal_ManAlloc(g,g,size,cuts,0);Gia_ManHashStart(g);
        for(step=0;step<48;step++){
            int a=pool[random_word(&state)%length];
            int b,c;
            a ^= random_word(&state)&1;
            b=pool[random_word(&state)%length];b ^= random_word(&state)&1;
            c=pool[random_word(&state)%length];c ^= random_word(&state)&1;
            switch(step%(size==2?2:3)){
            case 0:lit=Gia_ManHashAnd(g,a,b);break;
            case 1:lit=Gia_ManHashXorReal(g,a,b);break;
            default:lit=Gia_ManHashMuxReal(g,c,a,b);break;
            }
            cost=Bal_ManSetGateLevel(p,NULL,lit);pool[length++]=lit;
            printf("gate %d %d %d %d %d %d %d",size,cuts,seed,step,lit,cost,g->nObjs);
            {
                Bal_Cut_t *set=(Bal_Cut_t *)Vec_PtrEntry(p->vCutSets,Abc_Lit2Var(lit));
                if(set)for(i=0;i<cuts && set[i].nLeaves!=BAL_NO_LEAF;i++){
                    printf(" cut %d %u %llu",set[i].Delay,set[i].iFunc,(unsigned long long)set[i].Sign);
                    for(j=0;j<(int)set[i].nLeaves;j++)printf(" %d",set[i].pLeaves[j]);
                }
            }
            puts("");
            {
                Vec_Int_t *v=Vec_IntAlloc(6);
                for(i=length-6;i<length;i++)Vec_IntPush(v,pool[i]);
                Vec_IntSelectSortCostLit(v,p->vCosts);
                pair=Bal_ManFindBestPair(p,v,Gia_ManObj(g,Abc_Lit2Var(lit)));
                printf("pair %d %d",pair>>16,pair&65535);
                Vec_IntForEachEntry(v,n,i)printf(" %d",n);
                puts("");Vec_IntFree(v);
            }
        }
        Gia_ManHashStop(g);Gia_ManCreateRefs(g);
        for(i=1;i<g->nObjs;i++)if(Gia_ObjIsAnd(Gia_ManObj(g,i)) && !Gia_ObjIsMux(g,Gia_ManObj(g,i))){
            Gia_ManSuperCollect(g,Gia_ManObj(g,i));printf("super %d",i);
            Vec_IntForEachEntry(g->vSuper,n,j)printf(" %d",n);
            puts("");
        }
        Bal_ManFree(p);g->pData=NULL;Gia_ManStop(g);
    }
    collector_boundaries();
    { Vec_Int_t *values=Vec_IntAlloc(8);simplify_cases(values,0);Vec_IntFree(values); }
    merge_cases();
    mux_merge_cases();
    return 0;
}
