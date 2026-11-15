#include "aig/gia/gia.h"
#pragma GCC diagnostic error "-Wimplicit-function-declaration"
extern Vec_Wec_t *Gia_ManCreateNodeSupps(Gia_Man_t *,Vec_Int_t *,int);
extern Vec_Wec_t *Gia_ManCreateCoSupps(Gia_Man_t *,int);
extern int Gia_ManCoSuppSizeMax(Gia_Man_t *,Vec_Wec_t *);
extern int Gia_ManCoLargestSupp(Gia_Man_t *,Vec_Wec_t *);
extern Vec_Int_t *Gia_ManSortCoBySuppSize(Gia_Man_t *,Vec_Wec_t *);
static int next(unsigned *seed,int size) { *seed=*seed*1664525u+1013904223u; return *seed%(unsigned)size; }
static Gia_Man_t *graph(unsigned seed,int layout,int regs)
{
    Gia_Man_t *g=Gia_ManStart(1),*out;
    int lits[56]={0,1},size=2,i,a,b,c;
    g->fGiaSimple=1;
    for(i=0;i<6;i++) lits[size++]=Gia_ManAppendCi(g);
    for(i=0;i<48;i++) {
        a=lits[next(&seed,size)];b=lits[next(&seed,size)];c=lits[next(&seed,size)];(void)c;
        if((a>>1)==(b>>1)) {b=lits[2];if((a>>1)==(b>>1)) b=lits[3];}
        lits[size++]=Gia_ManAppendAnd(g,a,b^1);
    }
    Gia_ManAppendCo(g,0);Gia_ManAppendCo(g,1);Gia_ManAppendCo(g,lits[2]);Gia_ManAppendCo(g,lits[2]^1);
    Gia_ManAppendCo(g,lits[3]);Gia_ManAppendCo(g,lits[8]);Gia_ManAppendCo(g,lits[14]);Gia_ManAppendCo(g,lits[23]);
    Gia_ManAppendCo(g,lits[37]);Gia_ManAppendCo(g,lits[49]);Gia_ManAppendCo(g,lits[55]);Gia_ManAppendCo(g,lits[55]^1);
    Gia_ManSetRegNum(g,regs);
    if(!layout) return g;
    out=Gia_ManDupOrderDfsReverse(g);Gia_ManStop(g);return out;
}
int main(void)
{
    unsigned seed;
    int layout,regs,i,j,id,ci,prev;
    Gia_Obj_t *obj;
    Vec_Int_t *entry,*nodes,*order;
    Vec_Wec_t *supports,*cos;
    Gia_Man_t *g;
    for(seed=0;seed<32;seed++) for(layout=0;layout<2;layout++) for(regs=0;regs<=2;regs+=2) {
        g=graph(seed,layout,regs);
        nodes=Vec_IntAlloc(Gia_ManObjNum(g));
        Gia_ManForEachObj(g,obj,i) if(!Gia_ObjIsCo(obj)) Vec_IntPush(nodes,i);
        supports=Gia_ManCreateNodeSupps(g,nodes,0);cos=Gia_ManCreateCoSupps(g,0);
        printf("%u %s %d",seed,layout?"true":"false",regs);
        Vec_IntForEachEntry(nodes,id,i) {
            entry=Vec_WecEntry(supports,i);printf(" n%d:%d",id,Vec_IntSize(entry));
            Vec_IntForEachEntry(entry,ci,j) printf(",%d",ci);
        }
        Gia_ManForEachCo(g,obj,i) {
            entry=Vec_WecEntry(cos,i);printf(" c%d:%d",i,Vec_IntSize(entry));
            Vec_IntForEachEntry(entry,ci,j) printf(",%d",ci);
        }
        printf(" max%d largest%d\n",Gia_ManCoSuppSizeMax(g,cos),Gia_ManCoLargestSupp(g,cos));
        /* Quicksort ties need not be stable. Check the descending permutation,
           rather than requiring Go to reproduce an incidental tie order. */
        order=Gia_ManSortCoBySuppSize(g,cos);prev=Gia_ManCiNum(g)+1;
        assert(Vec_IntSize(order)==Gia_ManCoNum(g));
        Vec_IntForEachEntry(order,id,i) {
            assert(id>=0 && id<Gia_ManCoNum(g));
            for(j=0;j<i;j++) assert(Vec_IntEntry(order,j)!=id);
            assert(prev>=Vec_IntSize(Vec_WecEntry(cos,id)));prev=Vec_IntSize(Vec_WecEntry(cos,id));
        }
        Vec_IntFree(order);Vec_IntFree(nodes);Vec_WecFree(supports);Vec_WecFree(cos);Gia_ManStop(g);
    }
    return 0;
}
