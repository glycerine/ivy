#include "aig/aig/aigDfs.c"

int main(void)
{
    Aig_Obj_t objects[7];
    int mode, i, count;
    for (mode = 0; mode < 5; mode++)
    {
        memset(objects, 0, sizeof(objects));
        for (i = 0; i < 7; i++)
        {
            objects[i].Id = i;
            objects[i].Type = i < 4 ? AIG_OBJ_CI : AIG_OBJ_AND;
            objects[i].pData = &objects[i];
        }
        objects[4].pFanin0 = &objects[0];
        objects[4].pFanin1 = &objects[1];
        objects[5].pFanin0 = &objects[2];
        objects[5].pFanin1 = &objects[4];
        objects[6].pFanin0 = &objects[3];
        objects[6].pFanin1 = &objects[5];
        objects[5].fMarkA = 1;
        count = 0;
        switch (mode)
        {
        case 0: Aig_ConeMark_rec(&objects[6]); break;
        case 1: Aig_ConeCleanAndMark_rec(&objects[6]); break;
        case 2: count = Aig_ConeCountAndMark_rec(&objects[6]); break;
        case 3:
            objects[5].fMarkA = 0;
            objects[0].fMarkA = objects[4].fMarkA = objects[6].fMarkA = 1;
            Aig_ConeUnmark_rec(&objects[6]);
            break;
        case 4:
            objects[0].fMarkA = 1;
            count = Aig_DagSize(Aig_Not(&objects[6]));
            break;
        }
        printf("%d %d ", mode, count);
        for (i = 0; i < 7; i++)
            printf("%d", objects[i].fMarkA);
        printf(" ");
        for (i = 0; i < 7; i++)
            printf("%d", objects[i].pData != NULL);
        printf("\n");
    }
    return 0;
}
