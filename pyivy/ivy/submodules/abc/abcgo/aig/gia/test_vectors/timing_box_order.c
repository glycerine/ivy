#include "aig/gia/gia.h"
#include "misc/tim/tim.h"
#pragma GCC diagnostic error "-Wimplicit-function-declaration"
extern Vec_Int_t *Gia_ManOrderWithBoxes(Gia_Man_t *);
extern Gia_Man_t *Gia_ManDupUnnormalize(Gia_Man_t *);
static Gia_Man_t *graph(int signs,int choice,int regs)
{
    Gia_Man_t *g=Gia_ManStart(1);Tim_Man_t *timing;Vec_Ptr_t *tables;
    int a=Gia_ManAppendCi(g),b=Gia_ManAppendCi(g),u=Gia_ManAppendCi(g),v=Gia_ManAppendCi(g),w=Gia_ManAppendCi(g);
    int sibling=0,x0,x1,y0,o0,o1;float *table;
    if(choice)sibling=Gia_ManAppendAnd(g,a^(signs&1),b^((signs>>1)&1));
    x0=Gia_ManAppendAnd(g,a^(signs&1),b^((signs>>1)&1));
    x1=Gia_ManAppendAnd(g,a,b^((signs>>2)&1));
    y0=Gia_ManAppendAnd(g,u^((signs>>3)&1),v^((signs>>4)&1));
    o0=Gia_ManAppendAnd(g,w^((signs>>5)&1),y0^((signs>>6)&1));
    o1=Gia_ManAppendAnd(g,w^((signs>>7)&1),a);
    Gia_ManAppendAnd(g,a^1,b);
    Gia_ManAppendCo(g,x0);Gia_ManAppendCo(g,x1);Gia_ManAppendCo(g,y0);
    Gia_ManAppendCo(g,o0);Gia_ManAppendCo(g,o1);Gia_ManAppendCo(g,b^1);
    if(choice){g->pSibls=ABC_CALLOC(int,g->nObjs);g->pSibls[Abc_Lit2Var(x0)]=Abc_Lit2Var(sibling);}
    Gia_ManSetRegNum(g,regs);timing=Tim_ManStart(5,6);tables=Vec_PtrAlloc(2);
    {float values[7]={0,2,2,1,2,3,4};table=ABC_ALLOC(float,7);memcpy(table,values,sizeof(values));Vec_PtrPush(tables,table);}
    {float values[4]={1,1,1,5};table=ABC_ALLOC(float,4);memcpy(table,values,sizeof(values));Vec_PtrPush(tables,table);}
    Tim_ManSetDelayTables(timing,tables);Tim_ManCreateBox(timing,0,2,2,2,0,0);Tim_ManCreateBox(timing,2,1,4,1,1,0);
    g->pManTime=timing;return g;
}
int main(void)
{
    int signs,choice,regs,i,id,mask,*values,packed;Gia_Obj_t *obj;
    for(signs=0;signs<256;signs++)for(choice=0;choice<2;choice++)for(regs=0;regs<3;regs++){
        Gia_Man_t *g=graph(signs,choice,regs),*out;Vec_Int_t *order=Gia_ManOrderWithBoxes(g);
        printf("case %d %d %d order",signs,choice,regs);Vec_IntForEachEntry(order,id,i)printf(" %d",id);Vec_IntFree(order);
        out=Gia_ManDupUnnormalize(g);printf(" result %d %d %d %d values",out->nObjs,Gia_ManCiNum(out),Gia_ManCoNum(out),out->nRegs);
        Gia_ManForEachObj(g,obj,i)printf(" %u",obj->Value);
        printf(" nodes");Gia_ManForEachObj(out,obj,i){
            if(Gia_ObjIsConst0(obj))printf(" 0");
            else if(Gia_ObjIsCi(obj))printf(" 1");
            else if(Gia_ObjIsCo(obj))printf(" 6:%d",Gia_ObjFaninLit0p(out,obj));
            else printf(" 2:%d:%d:%d",Gia_ObjFaninLit0p(out,obj),Gia_ObjFaninLit1p(out,obj),Gia_ObjSibl(out,i));
        }
        printf(" truth");values=ABC_CALLOC(int,out->nObjs);
        for(mask=0;mask<32;mask++){
            Gia_ManForEachCi(out,obj,i)values[Gia_ObjId(out,obj)]=(mask>>i)&1;
            Gia_ManForEachAnd(out,obj,i)values[i]=(values[Gia_ObjFaninId0(obj,i)]^Gia_ObjFaninC0(obj))&(values[Gia_ObjFaninId1(obj,i)]^Gia_ObjFaninC1(obj));
            packed=0;Gia_ManForEachCo(out,obj,i)packed|=(values[Gia_ObjFaninId0p(out,obj)]^Gia_ObjFaninC0(obj))<<i;
            printf(" %d",packed);
        }
        puts("");ABC_FREE(values);Gia_ManStop(out);Gia_ManStop(g);
    }
    return 0;
}
