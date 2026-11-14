#include "aig/gia/gia.h"
unsigned enable_dbg_outs=1;
int Abc_FrameIsBridgeMode(void) { return 0; }
int Gia_ManToBridgeText(FILE *f,int n,unsigned char *s) { abort(); }
char *vnsprintf(const char *f,va_list a) { abort(); }
static int value(int lit,int mask) { return lit<2?(lit&1):((mask>>((lit>>1)-1))&1)^(lit&1); }
int main(void)
{
    int signs,mask,control,yes,no,root;
    for (signs=0;signs<8;signs++)
    {
        Gia_Man_t g={0}; Gia_Obj_t objects[8]={0},*o; unsigned muxes[8]={0};
        g.pObjs=objects; g.pMuxes=muxes; g.nObjs=1; g.nObjsAlloc=8;
        g.vCis=Vec_IntAlloc(4); g.vCos=Vec_IntAlloc(0);
        objects[0].iDiff0=objects[0].iDiff1=GIA_NONE;
        control=Gia_ManAppendCi(&g); yes=Gia_ManAppendCi(&g); no=Gia_ManAppendCi(&g);
        root=Gia_ManAppendMuxReal(&g,control^(signs&1),yes^((signs>>1)&1),no^((signs>>2)&1));
        o=Gia_ManObj(&g,root>>1);
        printf("signs %d and %d xor %d mux %d pins %d %d control %u\n",signs,Gia_ObjIsAnd(o),Gia_ObjIsXor(o),Gia_ObjIsMux(&g,o),Gia_ObjFaninLit0(o,root>>1),Gia_ObjFaninLit1(o,root>>1),muxes[root>>1]);
        for (mask=0;mask<8;mask++) printf("truth %d %d\n",mask,value(muxes[root>>1],mask)?value(Gia_ObjFaninLit1(o,root>>1),mask):value(Gia_ObjFaninLit0(o,root>>1),mask));
        Vec_IntFree(g.vCis); Vec_IntFree(g.vCos);
    }
    return 0;
}
