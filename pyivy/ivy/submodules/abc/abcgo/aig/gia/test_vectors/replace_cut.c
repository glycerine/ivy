#include "aig/gia/gia.h"
#pragma GCC diagnostic error "-Wimplicit-function-declaration"
extern void Gia_ManHighLightFlopLogic(Gia_Man_t *);
extern Gia_Man_t *Gia_ManDupReplaceCut(Gia_Man_t *);

static Gia_Man_t *graph(int signs)
{
    Gia_Man_t *g=Gia_ManStart(1);
    int a=Gia_ManAppendCi(g),b=Gia_ManAppendCi(g),c=Gia_ManAppendCi(g),r0=Gia_ManAppendCi(g),r1=Gia_ManAppendCi(g);
    int x=Gia_ManAppendAnd(g,a^(signs&1),b^((signs>>1)&1));
    int y=Gia_ManAppendAnd(g,b,c^((signs>>2)&1));
    int z=Gia_ManAppendAnd(g,x,r0^((signs>>3)&1)),w=Gia_ManAppendAnd(g,y,r1^((signs>>4)&1));
    int root=Gia_ManAppendAnd(g,z,w),o=Gia_ManAppendAnd(g,a,c);
    Gia_ManAppendAnd(g,o,r0);
    Gia_ManAppendCo(g,root^((signs>>5)&1));Gia_ManAppendCo(g,x);Gia_ManAppendCo(g,z^1);
    Gia_ManAppendCo(g,0);Gia_ManAppendCo(g,1);Gia_ManAppendCo(g,z^((signs>>6)&1));Gia_ManAppendCo(g,w);
    Gia_ManSetRegNum(g,2);g->nConstrs=1;
    g->pName=Abc_UtilStrsav("cut");g->pSpec=Abc_UtilStrsav("cut.aig");return g;
}

int main(int argc,char **argv)
{
    int signs,i,mask,packed;
    Gia_Obj_t *obj;
    if (argc>1 && !strcmp(argv[1],"late"))
    {
        Gia_Man_t *g=Gia_ManStart(1),*out;
        int a=Gia_ManAppendCi(g),b=Gia_ManAppendCi(g),r=Gia_ManAppendCi(g);
        int x=Gia_ManAppendAnd(g,a,b),y=Gia_ManAppendAnd(g,x,r);
        Gia_ManAppendCo(g,y);Gia_ManAppendCo(g,a);Gia_ManAppendCo(g,x);Gia_ManSetRegNum(g,1);
        out=Gia_ManDupReplaceCut(g);
        obj=Gia_ManObj(out,Gia_ManObj(g,r>>1)->Value>>1);
        printf("%d %d ",Gia_ObjCioId(obj),Gia_ObjIsRo(out,obj));
        obj=Gia_ManObj(out,Gia_ManObj(g,a>>1)->Value>>1);
        printf("%d %d\n",Gia_ObjCioId(obj),Gia_ObjIsRo(out,obj));
        Gia_ManStop(out);Gia_ManStop(g);return 0;
    }
    for (signs=0;signs<128;++signs)
    {
        Gia_Man_t *g=graph(signs),*out;
        unsigned long long marks=0;
        int *values;
        Gia_ManForEachObj(g,obj,i) {obj->fMark0=1;obj->fMark1=i&1;}
        // The original helper assumes constant mark zero. Go resets it itself.
        Gia_ManConst0(g)->fMark0=0;
        Gia_ManHighLightFlopLogic(g);
        Gia_ManForEachObj(g,obj,i) marks|=(unsigned long long)obj->fMark0<<i;
        out=Gia_ManDupReplaceCut(g);
        printf("%d 0 %d %d %d %d %llu",signs,out->nObjs,Gia_ManCiNum(out),Gia_ManCoNum(out),out->nRegs,marks);
        Gia_ManForEachObj(g,obj,i) printf(" %u",obj->Value);
        values=ABC_CALLOC(int,out->nObjs);
        for (mask=0;mask<32;++mask)
        {
            Gia_ManForEachCi(out,obj,i) values[Gia_ObjId(out,obj)]=(mask>>i)&1;
            Gia_ManForEachAnd(out,obj,i)
                values[i]=(values[Gia_ObjFaninId0(obj,i)]^Gia_ObjFaninC0(obj))&(values[Gia_ObjFaninId1(obj,i)]^Gia_ObjFaninC1(obj));
            packed=0;
            Gia_ManForEachCo(out,obj,i) packed|=(values[Gia_ObjFaninId0p(out,obj)]^Gia_ObjFaninC0(obj))<<i;
            printf(" %d",packed);
        }
        puts("");ABC_FREE(values);Gia_ManStop(out);Gia_ManStop(g);
    }
    return 0;
}
