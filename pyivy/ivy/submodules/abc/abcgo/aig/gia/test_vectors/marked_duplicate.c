#include "aig/gia/gia.h"
#pragma GCC diagnostic error "-Wimplicit-function-declaration"
extern Gia_Man_t *Gia_ManDupMarked(Gia_Man_t *);
static Gia_Man_t *graph(int signs,int *branches,int *properties,int *constants)
{
    Gia_Man_t *g=Gia_ManStart(1);
    int inputs[7],x[3],i;
    for(i=0;i<7;i++)inputs[i]=Gia_ManAppendCi(g);
    x[0]=Gia_ManAppendAnd(g,inputs[0]^(signs&1),inputs[4]^((signs>>1)&1));
    x[1]=Gia_ManAppendAnd(g,inputs[1]^((signs>>3)&1),inputs[5]^((signs>>2)&1));
    x[2]=Gia_ManAppendAnd(g,inputs[2]^((signs>>5)&1),inputs[6]^((signs>>6)&1));
    for(i=0;i<3;i++)branches[i]=x[i]>>1;
    properties[0]=Gia_ManAppendCo(g,x[0]^(signs&1))>>1;
    constants[0]=Gia_ManAppendCo(g,0)>>1;constants[1]=Gia_ManAppendCo(g,1)>>1;
    properties[1]=Gia_ManAppendCo(g,x[1]^((signs>>1)&1))>>1;
    properties[2]=Gia_ManAppendCo(g,x[2]^((signs>>2)&1))>>1;
    for(i=0;i<3;i++)Gia_ManAppendCo(g,x[i]^1);
    Gia_ManSetRegNum(g,3);g->nConstrs=1;
    g->pName=Abc_UtilStrsav("marked");g->pSpec=Abc_UtilStrsav("marked.aig");return g;
}
int main(void)
{
    int phases[4]={0,37,90,127},s,signs,regs,pis,props,consts,anchor,i,mask,packed,*values;
    Gia_Obj_t *obj;
    for(s=0;s<4;s++)for(regs=0;regs<8;regs++)for(pis=0;pis<8;pis++) {
        if((pis&regs)!=regs)continue;
        signs=phases[s];
        for(props=0;props<8;props++) {
            if((props&regs)!=props || !(props&4))continue;
            /* Keep the constraint in this native comparison: C retains its
               old constraint count when the marked constraint is removed. */
            for(consts=0;consts<4;consts++)for(anchor=0;anchor<2;anchor++) {
                int branches[3],properties[3],constants[2];
                Gia_Man_t *g=graph(signs,branches,properties,constants),*out;
                for(i=0;i<3;i++) {
                    if(!(pis&(1<<i)))Gia_ManPi(g,i)->fMark0=1;
                    if(!(regs&(1<<i))) {
                        Gia_ManRo(g,i)->fMark0=1;Gia_ManRi(g,i)->fMark0=1;Gia_ManObj(g,branches[i])->fMark0=1;
                    }
                    if(!(props&(1<<i)))Gia_ManObj(g,properties[i])->fMark0=1;
                }
                if(!anchor)Gia_ManPi(g,3)->fMark0=1;
                for(i=0;i<2;i++)if(!(consts&(1<<i)))Gia_ManObj(g,constants[i])->fMark0=1;
                out=Gia_ManDupMarked(g);
                printf("%d %d %d %d %d %d result %d %d %d %d %d values",signs,regs,pis,props,consts,anchor,out->nObjs,Gia_ManCiNum(out),Gia_ManCoNum(out),out->nRegs,out->nConstrs);
                Gia_ManForEachObj(g,obj,i)printf(" %u",obj->Value);
                printf(" truth");values=ABC_CALLOC(int,out->nObjs);
                for(mask=0;mask<(1<<Gia_ManCiNum(out));mask++) {
                    Gia_ManForEachCi(out,obj,i)values[Gia_ObjId(out,obj)]=(mask>>i)&1;
                    Gia_ManForEachAnd(out,obj,i)
                        values[i]=(values[Gia_ObjFaninId0(obj,i)]^Gia_ObjFaninC0(obj))&(values[Gia_ObjFaninId1(obj,i)]^Gia_ObjFaninC1(obj));
                    packed=0;
                    Gia_ManForEachCo(out,obj,i)packed|=(values[Gia_ObjFaninId0p(out,obj)]^Gia_ObjFaninC0(obj))<<i;
                    printf(" %d",packed);
                }
                puts("");ABC_FREE(values);Gia_ManStop(out);Gia_ManStop(g);
            }
        }
    }
    return 0;
}
