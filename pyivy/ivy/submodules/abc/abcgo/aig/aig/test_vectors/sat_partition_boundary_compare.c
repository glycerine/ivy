#include "aig/aig/aig.h"
unsigned enable_dbg_outs = 1;
int Abc_FrameIsBridgeMode(void) { return 0; }
int Gia_ManToBridgeText(FILE *f, int n, unsigned char *s) { abort(); }
char *vnsprintf(const char *f, va_list a) { abort(); }
extern Aig_Man_t *Aig_ManPartSplitOne(Aig_Man_t *, Vec_Ptr_t *, Vec_Int_t **);
int main(void)
{
    int mask, reverse, i, j;
    for (mask=1; mask<8; mask++) for (reverse=0; reverse<2; reverse++)
    {
        Aig_Man_t *g=Aig_ManStart(100), *part;
        Aig_Obj_t *a=Aig_ObjCreateCi(g), *b=Aig_ObjCreateCi(g), *c=Aig_ObjCreateCi(g), *all[3], *o;
        Vec_Ptr_t *selected=Vec_PtrAlloc(3);
        Vec_Int_t *pio;
        all[0]=Aig_And(g,a,b); all[1]=Aig_And(g,Aig_Not(b),c); all[2]=Aig_And(g,all[0],all[1]);
        Aig_ObjCreateCo(g,Aig_Not(all[0])); Aig_ObjCreateCo(g,all[2]);
        b->fMarkA=1; c->fMarkB=1;
        for (i=0;i<3;i++) { j=reverse?2-i:i; if (mask&(1<<j)) Vec_PtrPush(selected,all[j]); }
        part=Aig_ManPartSplitOne(g,selected,&pio);
        printf("mask %d reverse %s pio:",mask,reverse?"true":"false");
        for (i=0;i<Vec_IntSize(pio);i++) printf(" %d",Vec_IntEntry(pio,i));
        puts("");
        Aig_ManForEachObj(part,o,i)
        {
            printf("%d %d %s %s",i,Aig_ObjRefs(o),o->fMarkA?"true":"false",o->fMarkB?"true":"false");
            if (Aig_ObjIsAnd(o)) printf(" and %d %d",Aig_ObjFaninId0(o)*2+Aig_ObjFaninC0(o),Aig_ObjFaninId1(o)*2+Aig_ObjFaninC1(o));
            else if (Aig_ObjIsCo(o)) printf(" out %d",Aig_ObjFaninId0(o)*2+Aig_ObjFaninC0(o));
            puts("");
        }
        Aig_ManForEachObj(part,o,i) o->fMarkA=o->fMarkB=0;
        Aig_ManForEachObj(g,o,i) o->fMarkA=o->fMarkB=0;
        Vec_IntFree(pio); Vec_PtrFree(selected); Aig_ManStop(part); Aig_ManStop(g);
    }
    return 0;
}
