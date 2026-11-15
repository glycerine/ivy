#include "aig/gia/gia.h"
#include "misc/tim/tim.h"
#pragma GCC diagnostic error "-Wimplicit-function-declaration"
extern Vec_Int_t *Gia_ManOrderWithBoxes(Gia_Man_t *);
int main(void)
{
    int signs,future,i,id,mask,packed,*values;Gia_Obj_t *obj;
    for(signs=0;signs<4;signs++)for(future=0;future<2;future++){
        Gia_Man_t *g=Gia_ManStart(1);Tim_Man_t *timing;Vec_Int_t *order;
        int a=Gia_ManAppendCi(g),b=Gia_ManAppendCi(g),u=Gia_ManAppendCi(g),w=Gia_ManAppendCi(g);
        int alternate,sibling,root,out,yes,no;
        a^=signs&1;alternate=a;
        if(future){yes=Gia_ManAppendAnd(g,a,w);no=Gia_ManAppendAnd(g,a,w^1);alternate=Gia_ManAppendAnd(g,yes^1,no^1)^1;}
        sibling=Gia_ManAppendBuf(g,alternate);root=Gia_ManAppendBuf(g,a);out=Gia_ManAppendAnd(g,u,b^((signs>>1)&1));
        Gia_ManAppendCo(g,root);Gia_ManAppendCo(g,u);Gia_ManAppendCo(g,out);
        g->pSibls=ABC_CALLOC(int,g->nObjs);g->pSibls[Abc_Lit2Var(root)]=Abc_Lit2Var(sibling);
        timing=Tim_ManStart(4,3);Tim_ManCreateBox(timing,0,1,2,1,-1,0);Tim_ManCreateBox(timing,1,1,3,1,-1,0);g->pManTime=timing;
        order=Gia_ManOrderWithBoxes(g);assert(order);
        printf("case %d %d order",signs,future);Vec_IntForEachEntry(order,id,i)printf(" %d",id);Vec_IntFree(order);
        printf(" truth");values=ABC_CALLOC(int,g->nObjs);
        for(mask=0;mask<16;mask++){
            Gia_ManForEachCi(g,obj,i)values[Gia_ObjId(g,obj)]=(mask>>i)&1;
            Gia_ManForEachAnd(g,obj,i){
                values[i]=values[Gia_ObjFaninId0(obj,i)]^Gia_ObjFaninC0(obj);
                if(!Gia_ObjIsBuf(obj))values[i]&=values[Gia_ObjFaninId1(obj,i)]^Gia_ObjFaninC1(obj);
            }
            assert(values[Abc_Lit2Var(root)]==values[Abc_Lit2Var(sibling)]);
            packed=0;Gia_ManForEachCo(g,obj,i)packed|=(values[Gia_ObjFaninId0p(g,obj)]^Gia_ObjFaninC0(obj))<<i;
            printf(" %d",packed);
        }
        puts("");ABC_FREE(values);Gia_ManStop(g);
    }
    return 0;
}
