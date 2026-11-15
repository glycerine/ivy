#include "aig/gia/gia.h"
#pragma GCC diagnostic error "-Wimplicit-function-declaration"
extern Gia_Man_t *Gia_ManBalanceInt(Gia_Man_t *,int);
static Gia_Man_t *graph(int family,int signs)
{
    Gia_Man_t *g=Gia_ManStart(128);int ci[6],i,x,root,left,right,first,other;
    for(i=0;i<6;i++)ci[i]=Gia_ManAppendCi(g);
    g->pMuxes=ABC_CALLOC(unsigned,g->nObjsAlloc);Gia_ManHashStart(g);
    x=Gia_ManHashAnd(g,ci[0]^(signs&1),ci[1]^((signs>>1)&1));root=x;
    switch(family){
    case 0:for(i=2;i<6;i++)root=Gia_ManHashAnd(g,root,ci[i]^((signs>>(i%3))&1));break;
    case 1:left=Gia_ManHashAnd(g,x,ci[2]);right=Gia_ManHashAnd(g,x,ci[3]);root=Gia_ManHashAnd(g,left,right);break;
    case 2:x=Gia_ManHashXorReal(g,ci[0]^(signs&1),ci[1]^((signs>>1)&1));left=Gia_ManHashXorReal(g,x,ci[2]^((signs>>2)&1));root=Gia_ManHashAnd(g,left,ci[4]);break;
    case 3:x=Gia_ManHashMuxReal(g,ci[2]^((signs>>2)&1),ci[0]^(signs&1),ci[1]^((signs>>1)&1));root=Gia_ManHashAnd(g,x,ci[3]);break;
    case 4:first=Gia_ManAppendBuf(g,x^((signs>>2)&1));left=Gia_ManHashAnd(g,first,ci[2]);root=Gia_ManAppendBuf(g,left);Gia_ManAppendBuf(g,ci[3]);break;
    case 5:other=Gia_ManHashAnd(g,ci[0]^(signs&1),ci[1]^((signs>>1)&1)^1);root=Gia_ManHashAnd(g,x^1,other^1);break;
    }
    Gia_ManAppendCo(g,root^((signs>>3)&1));Gia_ManAppendCo(g,x);Gia_ManAppendCo(g,ci[0]);Gia_ManAppendCo(g,ci[1]);
    Gia_ManAppendCo(g,root^1);Gia_ManAppendCo(g,ci[4]^((signs>>2)&1));Gia_ManAppendCo(g,ci[5]^1);
    Gia_ManSetRegNum(g,2);Gia_ManHashStop(g);return g;
}
static void arrivals(Gia_Man_t *g,int mode,int delay)
{
    int ints[6]={0,5,2,8,1,3},i;float floats[6]={4.25f,0.75f,3.5f,1.25f,8.0f,2.5f};
    g->And2Delay=delay;
    if(mode==1 || mode==3)g->vCiArrs=Vec_IntAllocArrayCopy(ints,6);
    if(mode==2 || mode==3){g->vInArrs=Vec_FltAlloc(6);for(i=0;i<6;i++)Vec_FltPush(g->vInArrs,floats[i]);}
}
int main(void)
{
    int family,signs,strict,arrival,di,delays[3]={0,2,3},i,mask,packed,*values;Gia_Obj_t *obj;
    for(family=0;family<6;family++)for(signs=0;signs<16;signs++)for(strict=0;strict<2;strict++)for(arrival=0;arrival<4;arrival++)for(di=0;di<3;di++){
        Gia_Man_t *g=graph(family,signs),*out;arrivals(g,arrival,delays[di]);out=Gia_ManBalanceInt(g,strict);
        printf("case %d %d %d %d %d result %d %d %d %d values",family,signs,strict,arrival,delays[di],out->nObjs,out->nRegs,out->nXors,out->nMuxes);
        Gia_ManForEachObj(g,obj,i)printf(" %u",obj->Value);
        printf(" nodes");Gia_ManForEachObj(out,obj,i){
            if(Gia_ObjIsConst0(obj))printf(" 0");else if(Gia_ObjIsCi(obj))printf(" 1");
            else if(Gia_ObjIsCo(obj))printf(" 6:%d",Gia_ObjFaninLit0p(out,obj));
            else if(Gia_ObjIsBuf(obj))printf(" 5:%d",Gia_ObjFaninLit0p(out,obj));
            else if(Gia_ObjIsMux(out,obj))printf(" 4:%d:%d:%d",Gia_ObjFaninLit2p(out,obj),Gia_ObjFaninLit1p(out,obj),Gia_ObjFaninLit0p(out,obj));
            else printf(" %d:%d:%d",Gia_ObjIsXor(obj)?3:2,Gia_ObjFaninLit0p(out,obj),Gia_ObjFaninLit1p(out,obj));
        }
        printf(" truth");values=ABC_CALLOC(int,out->nObjs);
        for(mask=0;mask<64;mask++){
            Gia_ManForEachCi(out,obj,i)values[Gia_ObjId(out,obj)]=(mask>>i)&1;
            Gia_ManForEachAnd(out,obj,i){
                int a=values[Gia_ObjFaninId0(obj,i)]^Gia_ObjFaninC0(obj),b=values[Gia_ObjFaninId1(obj,i)]^Gia_ObjFaninC1(obj);
                if(Gia_ObjIsBuf(obj))values[i]=a;
                else if(Gia_ObjIsMux(out,obj))values[i]=(values[Gia_ObjFaninId2(out,i)]^Gia_ObjFaninC2(out,obj))?b:a;
                else if(Gia_ObjIsXor(obj))values[i]=a^b;else values[i]=a&b;
            }
            packed=0;Gia_ManForEachCo(out,obj,i)packed|=(values[Gia_ObjFaninId0p(out,obj)]^Gia_ObjFaninC0(obj))<<i;
            printf(" %d",packed);
        }
        puts("");ABC_FREE(values);Gia_ManStop(out);Gia_ManStop(g);
    }
    return 0;
}
