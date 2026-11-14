#include "aig/gia/gia.h"

extern int Hcd_ObjCheckTfi(Gia_Man_t *, Gia_Obj_t *, Gia_Obj_t *);
extern int Gia_ObjCheckTfi(Gia_Man_t *, Gia_Obj_t *, Gia_Obj_t *);
extern void Hcd_ManRemoveBadChoices(Gia_Man_t *);
extern void Gia_ManRemoveBadChoices(Gia_Man_t *);

int main(void)
{
    Gia_Man_t g = {0};
    Gia_Obj_t objects[16] = {0};
    int next[16] = {0};
    Gia_Rpr_t reprs[16] = {0};
    int sizes[4] = {9, 15, 16, 16}, limit;
    /* IDs 1..4 are CIs. These are ordinary AND fanin IDs for 5..15. */
    int fanins[11][2] = {{1,2},{3,5},{1,3},{2,7},{4,6},{2,9},
                         {4,8},{4,5},{2,3},{1,13},{4,14}};
    int mode, old, node, api, id, dirty, result;
    g.pObjs = objects;
    g.pNexts = next;
    g.nObjs = g.nObjsAlloc = 16;
    objects[0].iDiff0 = objects[0].iDiff1 = GIA_NONE;
    for (id = 1; id <= 4; ++id)
    {
        objects[id].fTerm = 1;
        objects[id].iDiff0 = GIA_NONE;
        objects[id].iDiff1 = id-1;
    }
    for (id = 5; id < 16; ++id)
    {
        objects[id].iDiff0 = id-fanins[id-5][0];
        objects[id].iDiff1 = id-fanins[id-5][1];
    }
    for (mode = 0; mode < 4; ++mode)
    {
        memset(next, 0, sizeof(next));
        if (mode & 1) next[6] = 8;
        if (mode & 2) next[9] = 11;
        if (mode == 3) { next[8] = 14; next[11] = 15; }
        for (api = 0; api < 2; ++api)
            for (old = 5; old < 16; ++old)
                for (node = 0; node < 16; ++node)
                {
                    /* node 0 denotes NULL, not the source's unsupported
                       constant-leaf traversal. The Go extension tests that. */
                    Gia_Obj_t *root = node ? objects+node : NULL;
                    result = api ? Hcd_ObjCheckTfi(&g, objects+old, root)
                                 : Gia_ObjCheckTfi(&g, objects+old, root);
                    dirty = 0;
                    for (id = 0; id < 16; ++id) dirty |= objects[id].fMark0;
                    printf("%d %d %d %d %d %d\n", mode, api, old, node, result, dirty);
                }
    }
    g.pReprs = reprs;
    for (api = 0; api < 2; ++api)
        for (limit = 0; limit < 4; ++limit)
        {
            g.nObjs = sizes[limit];
            objects[11].iDiff0 = 11-4;
            objects[11].iDiff1 = 11-8;
            if (limit == 3)
            {
                /* Only the last choice member (14) has a physical fanout. */
                objects[11].iDiff0 = 11-1;
                objects[11].iDiff1 = 11-4;
            }
            memset(next, 0, sizeof(next));
            for (id = 0; id < 16; ++id) reprs[id].iRepr = GIA_VOID;
            next[6] = 8;
            reprs[8].iRepr = 6;
            if (g.nObjs > 14) { next[8] = 14; reprs[14].iRepr = 6; }
            if (api) Hcd_ManRemoveBadChoices(&g);
            else Gia_ManRemoveBadChoices(&g);
            printf("remove %d %d %d", api, limit, g.nObjs);
            for (id = 0; id < g.nObjs; ++id)
                printf(" %u:%d:%u", reprs[id].iRepr, next[id], objects[id].fMark0);
            puts("");
        }
    return 0;
}
