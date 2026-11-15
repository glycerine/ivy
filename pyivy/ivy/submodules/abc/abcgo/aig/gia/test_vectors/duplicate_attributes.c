#include "aig/gia/gia.h"
#include "misc/tim/tim.h"
#pragma GCC diagnostic error "-Wimplicit-function-declaration"
extern Gia_Man_t *Gia_ManDupWithAttributes(Gia_Man_t *);
static Gia_Man_t *graph(int signs)
{
    Gia_Man_t *g=Gia_ManStart(1);
    int a=Gia_ManAppendCi(g),b=Gia_ManAppendCi(g),c=Gia_ManAppendCi(g),r0=Gia_ManAppendCi(g),r1=Gia_ManAppendCi(g);
    int x=Gia_ManAppendAnd(g,a^(signs&1),b^((signs>>1)&1));
    int y=Gia_ManAppendAnd(g,x,c^((signs>>2)&1)),z=Gia_ManAppendAnd(g,y,r1);
    Gia_ManAppendCo(g,y^((signs>>4)&1));Gia_ManAppendCo(g,x^1);Gia_ManAppendCo(g,r0);Gia_ManAppendCo(g,z^1);
    Gia_ManSetRegNum(g,2);g->nConstrs=1;
    g->pName=Abc_UtilStrsav("attributes");g->pSpec=Abc_UtilStrsav("attributes.aig");return g;
}
static Vec_Int_t *vector(int *data,int size) {return Vec_IntAllocArrayCopy(data,size);}
static Tim_Man_t *timing(void)
{
    Tim_Man_t *p=Tim_ManStart(5,4);
    Vec_Ptr_t *tables=Vec_PtrAlloc(1);
    float *table=ABC_ALLOC(float,7),data[7]={0,2,2,1,2,3,4};int i;
    memcpy(table,data,sizeof(data));Vec_PtrPush(tables,table);
    Tim_ManCreateBox(p,0,2,3,2,0,1);Tim_ManSetDelayTables(p,tables);
    for(i=0;i<3;i++)Tim_ManInitPiArrival(p,i,1.25f*(i+1));
    Tim_ManInitPoRequired(p,2,10.5f);Tim_ManInitPoRequired(p,3,20.5f);Tim_ManBoxSetCopy(p,0,42);
    return p;
}
static void install(Gia_Man_t *p,int mask)
{
    Gia_Obj_t *obj;int i;
    int classes[2]={7,9},inits[2]={0,1},configs[4]={4,0,-1,9};
    if(mask&1) {
        p->vMapping=Vec_IntStart(p->nObjs);
        Gia_ManForEachAnd(p,obj,i) {
            Vec_IntWriteEntry(p->vMapping,i,Vec_IntSize(p->vMapping));Vec_IntPush(p->vMapping,2);
            Vec_IntPush(p->vMapping,Gia_ObjFaninId0(obj,i));Vec_IntPush(p->vMapping,Gia_ObjFaninId1(obj,i));Vec_IntPush(p->vMapping,i);
        }
    }
    if(mask&2) {
        p->vPacking=Vec_IntAlloc(5);Vec_IntPush(p->vPacking,1);Vec_IntPush(p->vPacking,3);
        Gia_ManForEachAnd(p,obj,i)Vec_IntPush(p->vPacking,i);
    }
    if(mask&4)p->pManTime=timing();
    if(mask&8) {
        int a,b;Gia_Man_t *extra=Gia_ManStart(1);
        a=Gia_ManAppendCi(extra);b=Gia_ManAppendCi(extra);
        Gia_ManAppendCo(extra,Gia_ManAppendAnd(extra,a,b^1));Gia_ManAppendCo(extra,a^1);Gia_ManSetRegNum(extra,1);
        extra->pName=Abc_UtilStrsav("extra");extra->pSpec=Abc_UtilStrsav("extra.aig");
        extra->vConfigs=Vec_IntStart(1);Vec_IntWriteEntry(extra->vConfigs,0,123);p->pAigExtra=extra;
    }
    if(mask&16)p->vRegClasses=vector(classes,2);
    if(mask&32)p->vRegInits=vector(inits,2);
    if(mask&64)p->vConfigs=vector(configs,4);
    if(mask&128)p->pCellStr=Abc_UtilStrsav("cell\nwith comments");
    if(mask&256)p->nAnd2Delay=7;
}
static void printVector(char *name,Vec_Int_t *v)
{
    int i,value;printf(" %s",name);if(!v){printf(" nil");return;}
    Vec_IntForEachEntry(v,value,i)printf(" %d",value);
}
int main(void)
{
    int mask,i,bits,packed,*values;
    Gia_Obj_t *obj;
    for(mask=0;mask<512;mask++) {
        Gia_Man_t *p,*out;
        if((mask&2)&&!(mask&1))continue;
        p=graph(mask&31);install(p,mask);out=Gia_ManDupWithAttributes(p);
        printf("%d graph %d %d %d %d",mask,out->nObjs,Gia_ManCiNum(out),Gia_ManCoNum(out),out->nRegs);
        printVector("mapping",out->vMapping);printVector("packing",out->vPacking);printVector("classes",out->vRegClasses);
        printVector("inits",out->vRegInits);printVector("configs",out->vConfigs);
        printf(" delay%d cell\"%s\"",out->nAnd2Delay,out->pCellStr?"cell\\nwith comments":"");
        if(out->pCellStr)assert(out->pCellStr!=p->pCellStr && !strcmp(out->pCellStr,p->pCellStr));
        if(!out->pManTime)printf(" timing nil");else {
            Tim_Man_t *q=(Tim_Man_t *)out->pManTime;
            float *table=Tim_ManBoxDelayTable(q,0);
            assert(q!=p->pManTime);
            printf(" timing %d %d %d %d %d %d",Tim_ManCiNum(q),Tim_ManCoNum(q),Tim_ManBoxNum(q),Tim_ManBlackBoxNum(q),Tim_ManBoxCopy(q,0),Tim_ManBoxDelayTableId(q,0));
            for(i=0;i<7;i++)printf(" %.2f",table[i]);
            /* Read actual PI/PO times, not the original bulk getters' invalid
               holes or erroneous required-time null/arrival handling. */
            for(i=0;i<3;i++)printf(" %.2f",Tim_ManGetCiArrival(q,i));
            for(i=2;i<4;i++)printf(" %.2f",Tim_ManGetCoRequired(q,i));
        }
        if(!out->pAigExtra)printf(" extra nil");else {
            Gia_Man_t *e=out->pAigExtra;assert(e!=p->pAigExtra && !e->vConfigs);
            printf(" extra %d %d %d %d",e->nObjs,Gia_ManCiNum(e),Gia_ManCoNum(e),e->nRegs);
        }
        printf(" truth");values=ABC_CALLOC(int,out->nObjs);
        for(bits=0;bits<32;bits++) {
            Gia_ManForEachCi(out,obj,i)values[Gia_ObjId(out,obj)]=(bits>>i)&1;
            Gia_ManForEachAnd(out,obj,i)
                values[i]=(values[Gia_ObjFaninId0(obj,i)]^Gia_ObjFaninC0(obj))&(values[Gia_ObjFaninId1(obj,i)]^Gia_ObjFaninC1(obj));
            packed=0;
            Gia_ManForEachCo(out,obj,i)packed|=(values[Gia_ObjFaninId0p(out,obj)]^Gia_ObjFaninC0(obj))<<i;
            printf(" %d",packed);
        }
        puts("");ABC_FREE(values);Gia_ManStop(out);Gia_ManStop(p);
    }
    return 0;
}
