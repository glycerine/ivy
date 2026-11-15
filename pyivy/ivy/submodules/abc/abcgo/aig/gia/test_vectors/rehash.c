#include "aig/gia/gia.h"
#pragma GCC diagnostic error "-Wimplicit-function-declaration"
extern Gia_Man_t *Gia_ManRehash(Gia_Man_t *,int);
static Gia_Man_t *graph(int signs,int family,int regs)
{
    Gia_Man_t *g=Gia_ManStart(1);int ci[6],i,a,b,c,d,r0,r1,x,y,z,side,root=-1;
    for(i=0;i<6;i++)ci[i]=Gia_ManAppendCi(g);
    a=ci[0];b=ci[1];c=ci[2];d=ci[3];r0=ci[4];r1=ci[5];
    x=Gia_ManAppendAnd(g,a^(signs&1),b^((signs>>1)&1));
    y=Gia_ManAppendAnd(g,a^(signs&1),c^((signs>>2)&1));
    z=Gia_ManAppendAnd(g,d^((signs>>3)&1),r1^((signs>>4)&1));
    side=Gia_ManAppendAnd(g,a^(signs&1),b^((signs>>1)&1)^1);
    switch(family){
    case 0:root=Gia_ManAppendAnd(g,x,a^(signs&1));break;
    case 1:root=Gia_ManAppendAnd(g,x,a^(signs&1)^1);break;
    case 2:root=Gia_ManAppendAnd(g,x,y);break;
    case 3:root=Gia_ManAppendAnd(g,x^1,y^1);break;
    case 4:root=Gia_ManAppendAnd(g,x^1,z^1);break;
    case 5:root=Gia_ManAppendAnd(g,x^1,side^1);break;
    }
    Gia_ManAppendAnd(g,b^((signs>>1)&1),d^((signs>>3)&1));
    Gia_ManAppendCo(g,root^((signs>>5)&1));Gia_ManAppendCo(g,x);Gia_ManAppendCo(g,y^1);Gia_ManAppendCo(g,z);
    Gia_ManAppendCo(g,0);Gia_ManAppendCo(g,1);Gia_ManAppendCo(g,root);Gia_ManAppendCo(g,r0^1);Gia_ManAppendCo(g,z^1);
    Gia_ManSetRegNum(g,regs);g->pName=Abc_UtilStrsav("rehash");g->pSpec=Abc_UtilStrsav("rehash.aig");return g;
}
int main(void)
{
    int signs,family,regs,additional,i,mask,packed,*values;Gia_Obj_t *obj;
    for(signs=0;signs<64;signs++)for(family=0;family<6;family++)for(regs=0;regs<=4;regs+=2)for(additional=0;additional<2;additional++){
        Gia_Man_t *g=graph(signs,family,regs),*out=Gia_ManRehash(g,additional);
        printf("%d %d %d %d result %d %d %d %d values",signs,family,regs,additional,out->nObjs,Gia_ManCiNum(out),Gia_ManCoNum(out),out->nRegs);
        /* These are the original routine's pre-cleanup source copy literals,
           including dangling nodes and COs, not guessed final object IDs. */
        Gia_ManForEachObj(g,obj,i)printf(" %u",obj->Value);
        printf(" truth");values=ABC_CALLOC(int,out->nObjs);
        for(mask=0;mask<64;mask++){
            Gia_ManForEachCi(out,obj,i)values[Gia_ObjId(out,obj)]=(mask>>i)&1;
            Gia_ManForEachAnd(out,obj,i)
                values[i]=(values[Gia_ObjFaninId0(obj,i)]^Gia_ObjFaninC0(obj))&(values[Gia_ObjFaninId1(obj,i)]^Gia_ObjFaninC1(obj));
            packed=0;
            Gia_ManForEachCo(out,obj,i)packed|=(values[Gia_ObjFaninId0p(out,obj)]^Gia_ObjFaninC0(obj))<<i;
            printf(" %d",packed);
        }
        puts("");ABC_FREE(values);Gia_ManStop(out);Gia_ManStop(g);
    }
    return 0;
}
