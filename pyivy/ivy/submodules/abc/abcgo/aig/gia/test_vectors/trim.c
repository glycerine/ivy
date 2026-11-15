#include "aig/gia/gia.h"
#pragma GCC diagnostic error "-Wimplicit-function-declaration"
extern Gia_Man_t *Gia_ManDupTrimmed(Gia_Man_t *,int,int,int,int);
extern Gia_Man_t *Gia_ManDupTrimmed2(Gia_Man_t *);
extern Vec_Int_t *Gia_ManDupTrimmedNonZero(Gia_Man_t *);
static void workingDual(void);
static Gia_Man_t *graph(int signs)
{
    Gia_Man_t *g=Gia_ManStart(1);int inputs[6],i,a,b,c,r0,r1,x,y,w;
    for(i=0;i<6;i++)inputs[i]=Gia_ManAppendCi(g);
    a=inputs[0];b=inputs[1];c=inputs[3];r0=inputs[4];r1=inputs[5];
    x=Gia_ManAppendAnd(g,a^(signs&1),b^((signs>>1)&1));y=Gia_ManAppendAnd(g,x,r0^((signs>>2)&1));w=Gia_ManAppendAnd(g,r1,a^((signs>>5)&1));
    Gia_ManAppendCo(g,0);Gia_ManAppendCo(g,1);Gia_ManAppendCo(g,a^(signs&1));Gia_ManAppendCo(g,c^((signs>>3)&1));
    Gia_ManAppendCo(g,r0^((signs>>2)&1));Gia_ManAppendCo(g,y^((signs>>4)&1));Gia_ManAppendCo(g,0);
    Gia_ManAppendCo(g,x^((signs>>1)&1));Gia_ManAppendCo(g,w^((signs>>5)&1));Gia_ManAppendCo(g,0);Gia_ManAppendCo(g,1);Gia_ManAppendCo(g,y^1);
    Gia_ManAppendCo(g,b);Gia_ManAppendCo(g,w^1);Gia_ManSetRegNum(g,2);g->nConstrs=3;
    g->pName=Abc_UtilStrsav("trim");g->pSpec=Abc_UtilStrsav("trim.aig");return g;
}
int main(int argc,char **argv)
{
    int signs,mode,i,mask,packed,*values;Gia_Obj_t *obj;
    if(argc>1 && !strcmp(argv[1],"dual")){workingDual();return 0;}
    if(argc>1 && !strcmp(argv[1],"mismatch")) {
        Gia_Man_t *p=Gia_ManStart(1),*out;Vec_Int_t *nonzero;
        Gia_ManAppendCi(p);Gia_ManAppendCo(p,0);Gia_ManAppendCo(p,1);
        nonzero=Gia_ManDupTrimmedNonZero(p);out=Gia_ManDupTrimmed(p,1,1,1,-1);
        printf("%d %d %d\n",Vec_IntSize(nonzero),Gia_ObjFaninLit0p(out,Gia_ManPo(out,0)),Gia_ObjFaninLit0p(out,Gia_ManPo(out,1)));
        Vec_IntFree(nonzero);Gia_ManStop(out);Gia_ManStop(p);return 0;
    }
    for(signs=0;signs<64;signs++)for(mode=0;mode<13;mode++) {
        Gia_Man_t *g=graph(signs),*out;
        out=mode==12?Gia_ManDupTrimmed2(g):Gia_ManDupTrimmed(g,(mode/6)&1,(mode/3)&1,0,mode%3-1);
        printf("%d %d result %d %d %d %d values",signs,mode,out->nObjs,Gia_ManCiNum(out),Gia_ManCoNum(out),out->nRegs);
        Gia_ManForEachObj(g,obj,i)if(!Gia_ObjIsCo(obj))printf(" %u",obj->Value);
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
    return 0;
}

static void workingDual(void)
{
    int signs,constant,trimCis,i,mask,packed,*values;
    Gia_Obj_t *obj;
    for(signs=0;signs<4;signs++)for(constant=0;constant<2;constant++)for(trimCis=0;trimCis<2;trimCis++) {
        Gia_Man_t *src=Gia_ManStart(1),*out;Vec_Int_t *kept;
        int a=Gia_ManAppendCi(src),b=Gia_ManAppendCi(src),r,x,twin;
        Gia_ManAppendCi(src);r=Gia_ManAppendCi(src);
        x=Gia_ManAppendAnd(src,a,b);twin=Gia_ManAppendAnd(src,b,a);
        Gia_ManAppendCo(src,a);Gia_ManAppendCo(src,a);Gia_ManAppendCo(src,r^(signs&1));Gia_ManAppendCo(src,a^((signs>>1)&1));
        Gia_ManAppendCo(src,x);Gia_ManAppendCo(src,twin);Gia_ManAppendCo(src,r);Gia_ManAppendCo(src,0);
        Gia_ManAppendCo(src,constant?0:a);Gia_ManSetRegNum(src,1);
        kept=Gia_ManDupTrimmedNonZero(src);out=Gia_ManDupTrimmed(src,trimCis,1,1,-1);
        printf("%d %d %d pairs",signs,constant,trimCis);Vec_IntForEachEntry(kept,packed,i)printf(" %d",packed);
        printf(" result %d %d %d %d truth",out->nObjs,Gia_ManCiNum(out),Gia_ManCoNum(out),out->nRegs);
        values=ABC_CALLOC(int,out->nObjs);
        for(mask=0;mask<(1<<Gia_ManCiNum(out));mask++) {
            Gia_ManForEachCi(out,obj,i)values[Gia_ObjId(out,obj)]=(mask>>i)&1;
            Gia_ManForEachAnd(out,obj,i)
                values[i]=(values[Gia_ObjFaninId0(obj,i)]^Gia_ObjFaninC0(obj))&(values[Gia_ObjFaninId1(obj,i)]^Gia_ObjFaninC1(obj));
            packed=0;
            Gia_ManForEachCo(out,obj,i)packed|=(values[Gia_ObjFaninId0p(out,obj)]^Gia_ObjFaninC0(obj))<<i;
            printf(" %d",packed);
        }
        puts("");ABC_FREE(values);Vec_IntFree(kept);Gia_ManStop(out);Gia_ManStop(src);
    }
}
