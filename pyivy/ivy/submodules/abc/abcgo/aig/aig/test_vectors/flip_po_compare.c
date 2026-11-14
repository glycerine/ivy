#include "aig/aig/aig.h"

int main(void)
{
    Aig_Man_t g;
    Aig_Obj_t objects[5];
    int driver, flip;
    memset(&g, 0, sizeof(g));
    memset(objects, 0, sizeof(objects));
    g.vCos = Vec_PtrAlloc(1);
    objects[4].Type = AIG_OBJ_CO;
    objects[4].Id = 4;
    Vec_PtrPush(g.vCos, objects + 4);
    for (driver = 0; driver < 8; driver++)
    {
        objects[4].pFanin0 = Aig_NotCond(objects + driver/2, driver & 1);
        printf("co %d:", Aig_ManCo(&g, 0)->Id);
        for (flip = 0; flip < 2; flip++)
        {
            Aig_ManFlipFirstPo(&g);
            printf(" %d", 2*(int)(Aig_ObjFanin0(objects+4)-objects) + Aig_ObjFaninC0(objects+4));
        }
        puts("");
    }
    Vec_PtrFree(g.vCos);
    return 0;
}
