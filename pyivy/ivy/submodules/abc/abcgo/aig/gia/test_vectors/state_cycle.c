#include "aig/gia/gia.h"
#pragma GCC diagnostic error "-Wimplicit-function-declaration"
extern Gia_Man_t *Gia_ManDupFlip(Gia_Man_t *,int *);
extern void Gia_ManCycle(Gia_Man_t *,Abc_Cex_t *,int);
extern Gia_Man_t *Gia_ManDupCycled(Gia_Man_t *,Abc_Cex_t *,int);
static Gia_Man_t *graph(int signs,int regs)
{
    Gia_Man_t *g=Gia_ManStart(1);
    int a=Gia_ManAppendCi(g),b=Gia_ManAppendCi(g),c=Gia_ManAppendCi(g),r0=Gia_ManAppendCi(g),r1=Gia_ManAppendCi(g);
    int x=Gia_ManAppendAnd(g,a^(signs&1),b^((signs>>1)&1));
    int y=Gia_ManAppendAnd(g,x,r0^((signs>>2)&1)),z=Gia_ManAppendAnd(g,r1,c);
    Gia_ManAppendCo(g,x);Gia_ManAppendCo(g,y);Gia_ManAppendCo(g,y^1);Gia_ManAppendCo(g,z);
    Gia_ManAppendCo(g,a^1);Gia_ManAppendCo(g,b);Gia_ManAppendCo(g,r1^1);
    Gia_ManSetRegNum(g,regs);return g;
}
static void print_result(Gia_Man_t *g,Gia_Man_t *out)
{
    int i,mask,packed,*values;Gia_Obj_t *obj;
    printf(" result %d %d values",out->nObjs,out->nRegs);
    Gia_ManForEachObj(g,obj,i)printf(" %u",obj->Value);
    printf(" nodes");Gia_ManForEachObj(out,obj,i){
        if(Gia_ObjIsConst0(obj))printf(" 0");else if(Gia_ObjIsCi(obj))printf(" 1");
        else if(Gia_ObjIsCo(obj))printf(" 6:%d",Gia_ObjFaninLit0p(out,obj));
        else printf(" 2:%d:%d",Gia_ObjFaninLit0p(out,obj),Gia_ObjFaninLit1p(out,obj));
    }
    printf(" truth");values=ABC_CALLOC(int,out->nObjs);
    for(mask=0;mask<32;mask++){
        Gia_ManForEachCi(out,obj,i)values[Gia_ObjId(out,obj)]=(mask>>i)&1;
        Gia_ManForEachAnd(out,obj,i)values[i]=(values[Gia_ObjFaninId0(obj,i)]^Gia_ObjFaninC0(obj))&(values[Gia_ObjFaninId1(obj,i)]^Gia_ObjFaninC1(obj));
        packed=0;Gia_ManForEachCo(out,obj,i)packed|=(values[Gia_ObjFaninId0p(out,obj)]^Gia_ObjFaninC0(obj))<<i;
        printf(" %d",packed);
    }
    ABC_FREE(values);
}
int main(void)
{
    int mode,signs,regs,initial,trace,frames,i;Gia_Obj_t *obj;
    for(mode=0;mode<3;mode++)for(signs=0;signs<8;signs++)for(regs=0;regs<=5;regs++)
    for(initial=0;initial<(mode==2?1:(1<<regs));initial++)for(trace=0;trace<(mode==0?1:5);trace++)for(frames=0;frames<(mode==0?1:4);frames++){
        Gia_Man_t *g=graph(signs,regs),*out=NULL;Abc_Cex_t *cex=NULL;
        Gia_ManForEachObj(g,obj,i)obj->fMark1=i&1;
        Gia_ManForEachRo(g,obj,i)obj->fMark0=(initial>>i)&1;
        if(trace){cex=Abc_CexAlloc(regs,5-regs,4);cex->iFrame=3;cex->pData[0]=(unsigned)(trace+1)*0x9e3779b9u;}
        printf("case %d %d %d %d %d %d",mode,signs,regs,initial,trace,frames);
        if(mode==0)out=Gia_ManDupFlip(g,&initial);
        else if(mode==1)Gia_ManCycle(g,cex,frames);
        else {Gia_ManForEachObj(g,obj,i)obj->fMark0=1;out=Gia_ManDupCycled(g,cex,frames);}
        if(mode!=0)printf(" random %u",Gia_ManRandom(0));
        printf(" marks");Gia_ManForEachObj(g,obj,i)printf(" %d",obj->fMark0|(obj->fMark1<<1));
        if(out)print_result(g,out);
        puts("");if(cex)Abc_CexFree(cex);if(out)Gia_ManStop(out);Gia_ManStop(g);
    }
    return 0;
}
