#include "aig/gia/gia.h"
#include "misc/tim/tim.h"
#pragma GCC diagnostic error "-Wimplicit-function-declaration"
extern Gia_Man_t *Gia_ManDupUnshuffleInputs(Gia_Man_t *);
static Gia_Man_t *graph(int pis,int regs,int boxes,int signs)
{
    Gia_Man_t *g=Gia_ManStart(1);Tim_Man_t *timing;int ci[9],i,n=pis+regs+boxes,x,duplicate,y,outputs[6];
    for(i=0;i<n;i++)ci[i]=Gia_ManAppendCi(g);
    x=Gia_ManAppendAnd(g,ci[0]^(signs&1),ci[n-1]^((signs>>1)&1));
    duplicate=Gia_ManAppendAnd(g,ci[0]^(signs&1),ci[n-1]^((signs>>1)&1));
    y=Gia_ManAppendAnd(g,x,ci[pis]^((signs>>2)&1));
    outputs[0]=y;outputs[1]=x^((signs>>3)&1);outputs[2]=duplicate;
    for(i=0;i<regs;i++)outputs[3+i]=Gia_ManAppendAnd(g,y,ci[pis+boxes+i]);
    for(i=0;i<3+regs;i++)Gia_ManAppendCo(g,outputs[i]);
    Gia_ManSetRegNum(g,regs);timing=Tim_ManStart(n,3+regs);Tim_ManCreateBox(timing,0,1,pis+regs,boxes,-1,0);g->pManTime=timing;
    g->pReprs=ABC_CALLOC(Gia_Rpr_t,g->nObjs);for(i=0;i<g->nObjs;i++)Gia_ObjSetRepr(g,i,GIA_VOID);
    Gia_ObjSetRepr(g,Abc_Lit2Var(duplicate),Abc_Lit2Var(x));g->pNexts=Gia_ManDeriveNexts(g);
    return g;
}
int main(void)
{
    int pis,regs,boxes,signs,i,mask,packed,*values;Gia_Obj_t *obj;
    for(pis=1;pis<=3;pis++)for(regs=1;regs<=3;regs++)for(boxes=1;boxes<=3;boxes++)for(signs=0;signs<16;signs++){
        Gia_Man_t *g=graph(pis,regs,boxes,signs),*out=Gia_ManDupUnshuffleInputs(g);
        printf("case %d %d %d %d result %d %d values",pis,regs,boxes,signs,out->nObjs,out->nRegs);
        Gia_ManForEachObj(g,obj,i)printf(" %u",obj->Value);
        printf(" repr");for(i=0;i<out->nObjs;i++)printf(" %u",out->pReprs[i].iRepr);
        printf(" next");for(i=0;i<out->nObjs;i++)printf(" %d",out->pNexts[i]);
        printf(" nodes");Gia_ManForEachObj(out,obj,i){
            if(Gia_ObjIsConst0(obj))printf(" 0");else if(Gia_ObjIsCi(obj))printf(" 1");
            else if(Gia_ObjIsCo(obj))printf(" 6:%d",Gia_ObjFaninLit0p(out,obj));
            else printf(" 2:%d:%d",Gia_ObjFaninLit0p(out,obj),Gia_ObjFaninLit1p(out,obj));
        }
        printf(" truth");values=ABC_CALLOC(int,out->nObjs);
        for(mask=0;mask<(1<<Gia_ManCiNum(out));mask++){
            Gia_ManForEachCi(out,obj,i)values[Gia_ObjId(out,obj)]=(mask>>i)&1;
            Gia_ManForEachAnd(out,obj,i)values[i]=(values[Gia_ObjFaninId0(obj,i)]^Gia_ObjFaninC0(obj))&(values[Gia_ObjFaninId1(obj,i)]^Gia_ObjFaninC1(obj));
            packed=0;Gia_ManForEachCo(out,obj,i)packed|=(values[Gia_ObjFaninId0p(out,obj)]^Gia_ObjFaninC0(obj))<<i;
            printf(" %d",packed);
        }
        puts("");ABC_FREE(values);Gia_ManStop(out);Gia_ManStop(g);
    }
    return 0;
}
