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
 int native,smart,i,mask;
 for(native=0;native<2;native++) for(smart=0;smart<2;smart++) {
  Aig_Man_t *a=Aig_ManStart(32), *b=Aig_ManStart(32);
  Aig_Obj_t *x=Aig_ObjCreateCi(a), *y=Aig_ObjCreateCi(a);
  Aig_Obj_t *u=Aig_ObjCreateCi(b), *v=Aig_ObjCreateCi(b);
  Vec_Ptr_t *miters; a->fCatchExor=native;
  Aig_ObjCreateCo(a,Aig_Exor(a,x,y)); Aig_ObjCreateCo(b,Aig_And(b,u,v));
  miters=Aig_ManMiterPartitioned(a,b,1,smart);
  for(i=0;i<Vec_PtrSize(miters);i++) {
   Aig_Man_t *q=(Aig_Man_t *)Vec_PtrEntry(miters,i); int truth=0;
   for(mask=0;mask<4;mask++) truth|=value(Aig_ManCo(q,0),mask,q)<<mask;
   printf("native-XOR %d smart %d mismatch truth %d\n",native,smart,truth);
   Aig_ManStop(q);
  }
  Vec_PtrFree(miters); Aig_ManStop(a); Aig_ManStop(b);
 }
 return 0;
}
