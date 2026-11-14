#include "aig/aig/aig.h"

int main(void)
{
    Aig_Man_t p;
    Aig_Obj_t o[3];
    int data[15] = {0}, i;
    memset(&p, 0, sizeof(p));
    memset(o, 0, sizeof(o));
    o[0].Type = AIG_OBJ_CONST1;
    o[1].Type = AIG_OBJ_CI;
    o[2].Type = AIG_OBJ_CO;
    o[2].pFanin0 = &o[1];
    o[1].nRefs = 1;
    p.vObjs = Vec_PtrAlloc(3);
    for (i = 0; i < 3; i++)
    {
        o[i].Id = i;
        Vec_PtrPush(p.vObjs, &o[i]);
    }
    p.pFanData = data;
    p.nFansAlloc = 3;
    data[5] = data[11] = data[13] = 4;
    printf("refs %d ", Aig_ObjRefs(&o[2]));
    Aig_ObjRef(&o[2]);
    printf("%d ", Aig_ObjRefs(&o[2]));
    Aig_ObjDeref(&o[2]);
    printf("%d ", Aig_ObjRefs(&o[2]));
    Aig_ObjRef(&o[2]);
    Aig_ObjClearRef(&o[2]);
    printf("%d\n", Aig_ObjRefs(&o[2]));
    printf("head %d\n", Aig_ObjFanout0(&p, &o[1])->Id);
    printf("empty %d\n", Aig_ObjFanout0(&p, &o[2])->Id);
    Vec_PtrFree(p.vObjs);
    return 0;
}
