/* Print the final coordinate system and truth table after native linear sifting. */
#include "misc/util/abc_global.h"
#include "bdd/cudd/cuddInt.h"
#include <time.h>

static DdNode *build(DdManager *dd, unsigned truth, int level, int set)
{
    if (level == 5) return ((truth >> set) & 1) ? dd->one : dd->zero;
    DdNode *t = build(dd,truth,level+1,set|(1<<level));
    Cudd_Ref(t);
    DdNode *e = build(dd,truth,level+1,set);
    Cudd_Ref(e);
    DdNode *f = cuddZddGetNode(dd,level,t,e);
    Cudd_Ref(f);
    Cudd_RecursiveDerefZdd(dd,t);
    Cudd_RecursiveDerefZdd(dd,e);
    Cudd_Deref(f);
    return f;
}

static unsigned read_truth(DdManager *dd,DdNode *f)
{
    unsigned truth = 0;
    for (int i = 0; i < 32; ++i) {
        unsigned set = i;
        DdNode *n = f;
        while (!Cudd_IsConstant(n)) {
            unsigned b = 1u << n->index;
            if (set & b) {set ^= b; n = cuddT(n);} else n = cuddE(n);
        }
        if (n == dd->one && set == 0) truth |= 1u << i;
    }
    return truth;
}

static int run(unsigned input,int verbose)
{
    int order[] = {3,0,4,1,2};
    DdManager *dd = Cudd_Init(0,5,32,128,0);
    DdNode *f = build(dd,input,0,0);
    Cudd_Ref(f);
    if (!Cudd_zddShuffleHeap(dd,order) || read_truth(dd,f) != input ||
        !Cudd_zddReduceHeap(dd,CUDD_REORDER_LINEAR,0)) return 0;
    if (verbose) printf("%08x %d%d%d%d%d %08x %u\n",input,dd->invpermZ[0],dd->invpermZ[1],
                       dd->invpermZ[2],dd->invpermZ[3],dd->invpermZ[4],read_truth(dd,f),dd->keysZ);
    Cudd_Quit(dd);
    return 1;
}

int main(void)
{
    unsigned inputs[] = {0x420b6e3c,0xff000111,0x007dc06a,0xa78f2469};
    for (int i=0;i<4;++i) if (!run(inputs[i],1)) return 1;
    struct timespec a,b;
    const int count=8000;
    clock_gettime(CLOCK_MONOTONIC,&a);
    for (int i=0;i<count;++i) if (!run(inputs[i%4],0)) return 2;
    clock_gettime(CLOCK_MONOTONIC,&b);
    printf("%.2f us/build, shuffle and sift\n",((b.tv_sec-a.tv_sec)*1e6+(b.tv_nsec-a.tv_nsec)*1e-3)/count);
    return 0;
}
