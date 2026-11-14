#include "aig/aig/aig.h"
unsigned enable_dbg_outs = 1;
int Abc_FrameIsBridgeMode(void) { return 0; }
int Gia_ManToBridgeText(FILE *f, int n, unsigned char *s) { abort(); }
char *vnsprintf(const char *f, va_list a) { abort(); }
static int value(Aig_Obj_t *o, int mask, Aig_Man_t *g) {
    int i, a, b;
    o=Aig_Regular(o);
    if(Aig_ObjIsConst1(o)) return 1;
    if(Aig_ObjIsCi(o)) {
        for(i=0;i<Aig_ManCiNum(g);i++) if(Aig_ManCi(g,i)==o) return (mask>>i)&1;
        abort();
    }
    a=value(Aig_ObjFanin0(o),mask,g)^Aig_ObjFaninC0(o);
    if(Aig_ObjIsCo(o)||Aig_ObjIsBuf(o)) return a;
    b=value(Aig_ObjFanin1(o),mask,g)^Aig_ObjFaninC1(o);
    return Aig_ObjIsExor(o) ? a^b : a&b;
}
int main(void) {
    int mode, mask, truth;
    for(mode=0;mode<2;mode++) {
        Aig_Man_t *g=Aig_ManStart(32), *q;
        Aig_Obj_t *a=Aig_ObjCreateCi(g), *b=Aig_ObjCreateCi(g), *x, *y, *copy;
        a->Level=3; b->Level=7;
        x=Aig_And(g,Aig_Not(a),Aig_Not(b)); y=mode?x:Aig_And(g,a,b);
        q=Aig_ManExtractMiter(g,x,y); truth=0;
        for(mask=0;mask<4;mask++) truth|=value(Aig_ManCo(q,0),mask,q)<<mask;
        copy=(Aig_Obj_t *)x->pData;
        printf("mode %d truth %d nodes %d ci-levels %d/%d copy-live %d\n",mode,truth,Aig_ManNodeNum(q),Aig_ManCi(q,0)->Level,Aig_ManCi(q,1)->Level,Aig_ManObj(q,Aig_Regular(copy)->Id)==Aig_Regular(copy));
        Aig_ManStop(q); Aig_ManStop(g);
    }
    return 0;
}
