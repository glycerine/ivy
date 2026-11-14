#include "aig/gia/giaForce.c"

static void places(Frc_Man_t *p)
{
    Frc_Obj_t *o;
    int h;
    Frc_ManForEachObj(p,o,h) printf(" %d",o->pPlace);
}

int main(void)
{
    int perms[6][3] = {{0,1,2},{0,2,1},{1,0,2},{1,2,0},{2,0,1},{2,1,0}};
    int cluster, perm, i, other, best;
    for (cluster = 0; cluster < 2; ++cluster)
        for (perm = 0; perm < 6; ++perm)
        {
            Gia_Man_t *g = Gia_ManStart(1);
            int a = Gia_ManAppendCi(g), b = Gia_ManAppendCi(g), c = Gia_ManAppendCi(g);
            int t = Gia_ManAppendAnd(g,b,c), x = Gia_ManAppendAnd(g,a,t);
            int y = Gia_ManAppendAnd(g,a,b), z = Gia_ManAppendAnd(g,x,y);
            Frc_Man_t *p;
            Vec_Int_t *order = Vec_IntAlloc(3);
            Gia_ManAppendCo(g,z); Gia_ManAppendCo(g,y); Gia_ManAppendCo(g,x);
            p = cluster ? Frc_ManStart(g) : Frc_ManStartSimple(g);
            for (i = 0; i < 3; ++i) Vec_IntPush(order,Vec_IntEntry(p->vCos,perms[perm][i]));
            printf("%d %d %d cuts %d",cluster,perm,p->nObjs,Frc_ManCrossCut(p,order,0));
            printf(" %d dfs",Frc_ManCrossCut(p,order,1));
            Frc_ManPlaceDfs(p,order); places(p);
            best = Frc_ManPlaceDfsBoth(p,order,&other);
            printf(" best %d %d",best,other); places(p);
            Frc_ManPlacementRefine(p,3,0);
            printf(" refined"); places(p); puts("");
            Vec_IntFree(order); Frc_ManStop(p); Gia_ManStop(g);
        }
    return 0;
}
