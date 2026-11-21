/* Collect the same ordered 4,096-AND chain as BenchmarkIvyMappingOrder. */
#include "aig/ivy/ivy.h"
#include <time.h>

int main(void)
{
    Ivy_Man_t *net = Ivy_ManStart();
    Ivy_Obj_t *input = Ivy_ObjCreatePi(net), *root = Ivy_ObjCreatePi(net);
    for (int i = 0; i < 4096; i++) root = Ivy_And(net, input, root);
    Ivy_ObjCreatePo(net, root);
    const int runs = 50000;
    struct timespec a, b;
    clock_gettime(CLOCK_PROCESS_CPUTIME_ID, &a);
    for (int r = 0; r < runs; r++) {
        Vec_Ptr_t *nodes = Vec_PtrAlloc(Ivy_ManNodeNum(net));
        Ivy_Obj_t *obj;
        int id;
        Ivy_ManForEachNode(net, obj, id)
            Vec_PtrPush(nodes, obj);
        if (Vec_PtrSize(nodes) != 4096 || Vec_PtrEntryLast(nodes) != root) return 1;
        Vec_PtrFree(nodes);
    }
    clock_gettime(CLOCK_PROCESS_CPUTIME_ID, &b);
    printf("%.2f us/ordered snapshot\n",
           ((b.tv_sec-a.tv_sec)*1e6 + (b.tv_nsec-a.tv_nsec)*1e-3)/runs);
    Ivy_ManStop(net);
    return 0;
}
