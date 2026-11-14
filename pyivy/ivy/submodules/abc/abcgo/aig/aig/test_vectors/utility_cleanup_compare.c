#include "aig/aig/aig.h"
unsigned enable_dbg_outs=1;
int Abc_FrameIsBridgeMode(void) { return 0; }
int Gia_ManToBridgeText(FILE *f,int n,unsigned char *s) { abort(); }
char *vnsprintf(const char *f,va_list a) { abort(); }
static void dump(Aig_Man_t *g,char *stage)
{
    Aig_Obj_t *o; int i;
    Aig_ManForEachObj(g,o,i) printf("%s %d cio %d data %d xsim %d\n",stage,i,o->CioId,o->iData,o->nCuts);
}
int main(void)
{
    Aig_Man_t *g=Aig_ManStart(100);
    Aig_Obj_t *a=Aig_ObjCreateCi(g),*b=Aig_ObjCreateCi(g),*c=Aig_ObjCreateCi(g),*root=Aig_And(g,a,b),*o;
    int i;
    Aig_ObjCreateCo(g,Aig_Not(root)); Aig_And(g,a,c);
    Aig_ManForEachObj(g,o,i) { o->CioId=100+i; o->iData=42+i; o->nCuts=2; }
    Aig_ManCleanNext(g); dump(g,"next");
    Aig_ManCleanData(g); dump(g,"data");
    Aig_ManForEachObj(g,o,i) o->iData=42+i;
    Aig_ObjCleanData_rec(root); dump(g,"cone");
    Aig_ManStop(g);
    return 0;
}
