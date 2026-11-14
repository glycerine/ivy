#include "aig/gia/giaGiarf.c"

int main(void)
{
    int signs, recursive, round, id;
    for (signs = 0; signs < 8; ++signs)
        for (recursive = 0; recursive < 2; ++recursive)
        {
            Gia_Man_t g = {0};
            Hcd_Man_t p = {0};
            Gia_Obj_t objects[7] = {0};
            unsigned sim[7] = {0};
            int trav[7] = {0};
            Gia_Rpr_t reprs[7] = {0};
            int next[7] = {0};
            Vec_Int_t *members = Vec_IntAlloc(3);
            g.pObjs = objects;
            g.nObjs = g.nObjsAlloc = 7;
            g.pTravIds = trav;
            g.nTravIdsAlloc = 7;
            g.pReprs = reprs;
            g.pNexts = next;
            p.pGia = &g;
            p.pSimInfo = sim;
            p.vClassOld = Vec_IntAlloc(100);
            p.vClassNew = Vec_IntAlloc(100);
            for (id = 4; id < 7; ++id) Vec_IntPush(members,id);
            objects[0].iDiff0 = objects[0].iDiff1 = GIA_NONE;
            for (id = 1; id < 4; ++id)
            {
                objects[id].fTerm = 1;
                objects[id].iDiff0 = GIA_NONE;
                objects[id].iDiff1 = id-1;
            }
            objects[4].iDiff0 = 3; objects[4].iDiff1 = 2;
            objects[4].fCompl0 = signs&1; objects[4].fCompl1 = (signs>>1)&1;
            objects[5].iDiff0 = 2; objects[5].iDiff1 = 1;
            objects[5].fCompl1 = (signs>>2)&1;
            objects[6].iDiff0 = 2; objects[6].iDiff1 = 1;
            objects[6].fCompl0 = 1;
            for (round = 0; round < 2; ++round)
            {
                sim[1] = 0xaaaaaaaaU ^ (round*0x5a5a5a5aU);
                sim[2] = 0xccccccccU ^ (round*0x5a5a5a5aU);
                sim[3] = 0xf0f0f0f0U ^ (round*0x5a5a5a5aU);
                if (recursive) { ++g.nTravIds; Gia_Resimulate_rec(&p,6); }
                else Hcd_ManSimulateSimple(&p);
                printf("%d %d %d",signs,recursive,round);
                for (id = 0; id < 7; ++id) printf(" %u",sim[id]);
                for (id = 0; id < 7; ++id) reprs[id].iRepr = GIA_VOID;
                Hcd_ManClassCreate(&g,members);
                printf(" split %d",Hcd_ManClassRefineOne(&p,4));
                for (id = 4; id < 7; ++id) printf(" %u:%d",reprs[id].iRepr,next[id]);
                printf(" remove %d",Hcd_ManClassClassRemoveOne(&p,recursive ? 6 : 5));
                for (id = 4; id < 7; ++id) printf(" %u:%d",reprs[id].iRepr,next[id]);
                puts("");
            }
            Vec_IntFree(members);
            Vec_IntFree(p.vClassOld);
            Vec_IntFree(p.vClassNew);
        }
    return 0;
}
