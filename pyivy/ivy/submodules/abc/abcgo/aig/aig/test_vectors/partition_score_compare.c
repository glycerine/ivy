#include "aig/aig/aig.h"

/* These source helpers are defined in aigPart.c rather than declared in aig.h. */
extern unsigned * Aig_ManSuppCharStart(Vec_Int_t *, int);
extern int Aig_ManPartitionSmartFindPart(Vec_Ptr_t *, Vec_Ptr_t *, Vec_Ptr_t *, int, Vec_Int_t *);

int main(void)
{
    int inputs, count, limit, width, length, value, i, j;
    while (scanf("%d%d%d%d", &inputs, &count, &limit, &width) == 4)
    {
        Vec_Ptr_t *supports = Vec_PtrAlloc(count);
        Vec_Ptr_t *parts = Vec_PtrAlloc(count);
        Vec_Ptr_t *bitmaps = Vec_PtrAlloc(count);
        Vec_Int_t *one = Vec_IntAlloc(width);
        for (i = 0; i < count; i++)
        {
            Vec_Int_t *support, *part;
            assert(scanf("%d", &length) == 1);
            support = Vec_IntAlloc(length);
            part = Vec_IntAlloc(1);
            Vec_IntPush(part,i);
            for (j = 0; j < length; j++)
            {
                assert(scanf("%d", &value) == 1);
                Vec_IntPush(support,value);
            }
            Vec_PtrPush(supports,support);
            Vec_PtrPush(parts,part);
            Vec_PtrPush(bitmaps,Aig_ManSuppCharStart(support,inputs));
        }
        for (i = 0; i < width; i++)
        {
            assert(scanf("%d", &value) == 1);
            Vec_IntPush(one,value);
        }
        printf("%d\n",Aig_ManPartitionSmartFindPart(supports,parts,bitmaps,limit,one));
        Vec_IntFree(one);
        for (i = 0; i < count; i++) free(Vec_PtrEntry(bitmaps,i));
        Vec_VecFree((Vec_Vec_t *)supports);
        Vec_VecFree((Vec_Vec_t *)parts);
        Vec_PtrFree(bitmaps);
    }
    return 0;
}
