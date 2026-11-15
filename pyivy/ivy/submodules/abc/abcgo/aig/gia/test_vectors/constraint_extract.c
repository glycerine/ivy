#include "aig/gia/gia.h"
#pragma GCC diagnostic error "-Wimplicit-function-declaration"
extern void Gia_ManDupWithConstrCollectAnd_rec(Gia_Man_t *,Gia_Obj_t *,Vec_Int_t *,int);
extern Gia_Man_t *Gia_ManDupWithConstr(Gia_Man_t *);
static Gia_Man_t *graph(int signs,int family,int layout)
{
    Gia_Man_t *g=Gia_ManStart(1),*out;
    int a=Gia_ManAppendCi(g),b=Gia_ManAppendCi(g),c=Gia_ManAppendCi(g),d=Gia_ManAppendCi(g),e=Gia_ManAppendCi(g);
    int x=Gia_ManAppendAnd(g,a^(signs&1),b^((signs>>1)&1));
    int w=Gia_ManAppendAnd(g,c^((signs>>2)&1),d^((signs>>3)&1));
    int y=Gia_ManAppendAnd(g,w,e^((signs>>4)&1)),root=-1,left,right;
    switch(family) {
    case 0:root=Gia_ManAppendAnd(g,x,y);break;
    case 1:root=Gia_ManAppendAnd(g,x^1,y);break;
    case 2:root=Gia_ManAppendAnd(g,x^1,y^1);break;
    case 3:left=Gia_ManAppendAnd(g,x,y);right=Gia_ManAppendAnd(g,x,e);root=Gia_ManAppendAnd(g,left,right);break;
    case 4:root=Gia_ManAppendAnd(g,x,0);break;
    case 5:root=Gia_ManAppendAnd(g,x,1);break;
    case 6:root=Gia_ManAppendAnd(g,x,a^(signs&1)^1);break;
    }
    Gia_ManAppendCo(g,root);g->pName=Abc_UtilStrsav("constr");g->pSpec=Abc_UtilStrsav("constr.aig");
    if(!layout)return g;
    out=Gia_ManDupOrderDfsReverse(g);Gia_ManStop(g);return out;
}
int main(void)
{
    int signs,family,layout,i,lit,mask,packed,ports[5],nPorts;
    Gia_Obj_t *obj;
    for(signs=0;signs<32;signs++)for(family=0;family<7;family++)for(layout=0;layout<2;layout++) {
        Gia_Man_t *g=graph(signs,family,layout),*out;
        Vec_Int_t *super=Vec_IntAlloc(8);
        int *values;
        Gia_ManDupWithConstrCollectAnd_rec(g,Gia_ObjChild0(Gia_ManPo(g,0)),super,1);
        printf("%d %d %s leaves",signs,family,layout?"true":"false");
        Vec_IntForEachEntry(super,lit,i)printf(" %d",lit);
        /* The C copier appends CIs in physical object order. Account for
           that explicit permutation when comparing original-port functions. */
        printf(" ports");nPorts=0;
        Gia_ManForEachObj(g,obj,i)if(Gia_ObjIsCi(obj)) {
            ports[nPorts++]=Gia_ObjCioId(obj);printf(" %d",Gia_ObjCioId(obj));
        }
        assert(nPorts==5);
        out=Gia_ManDupWithConstr(g);
        assert(out && out->nRegs==0 && out->nConstrs==1);
        printf(" result %d %d %d %d %d",out->nObjs,Gia_ManCiNum(out),Gia_ManCoNum(out),out->nRegs,out->nConstrs);
        values=ABC_CALLOC(int,out->nObjs);
        for(mask=0;mask<32;mask++) {
            Gia_ManForEachCi(out,obj,i)values[Gia_ObjId(out,obj)]=(mask>>ports[i])&1;
            Gia_ManForEachAnd(out,obj,i)
                values[i]=(values[Gia_ObjFaninId0(obj,i)]^Gia_ObjFaninC0(obj))&(values[Gia_ObjFaninId1(obj,i)]^Gia_ObjFaninC1(obj));
            packed=0;
            Gia_ManForEachCo(out,obj,i)packed|=(values[Gia_ObjFaninId0p(out,obj)]^Gia_ObjFaninC0(obj))<<i;
            printf(" %d",packed);
        }
        puts("");ABC_FREE(values);Vec_IntFree(super);Gia_ManStop(out);Gia_ManStop(g);
    }
    return 0;
}
