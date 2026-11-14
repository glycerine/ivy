#include "aig/aig/aigRet.c"

int main(void)
{
    Rtm_Man_t p;
    Rtm_Edg_t edges[3];
    int id, action, value, result, i, j;
    memset(&p, 0, sizeof(p));
    memset(edges, 0, sizeof(edges));
    while (scanf("%d %d %d", &id, &action, &value) == 3)
    {
        assert(id >= 0 && id < 3);
        result = 0;
        switch (action)
        {
        case 0: Rtm_ObjAddFirst(&p, edges+id, (Rtm_Init_t)value); break;
        case 1: Rtm_ObjAddLast(&p, edges+id, (Rtm_Init_t)value); break;
        case 2: result = Rtm_ObjRemFirst(&p, edges+id); break;
        case 3: result = Rtm_ObjRemLast(&p, edges+id); break;
        default: abort();
        }
        printf("%d", result);
        for (i = 0; i < 3; i++)
        {
            printf(" | %d:", (int)edges[i].nLats);
            for (j = 0; j < (int)edges[i].nLats; j++)
                printf("%d", Rtm_ObjGetOne(&p, edges+i, j));
        }
        puts("");
    }
    free(p.pExtra);
    return 0;
}
