#include "base/abci/abcFx.c"

static unsigned next(unsigned *state) { return *state = *state * 1664525u + 1013904223u; }
int main(void)
{
    int mode, seed, limit, litMax, canon, owner, row, var, phase, i, j, support[6], width;
    for(mode=0;mode<3;mode++)
    for(seed=1;seed<=64;seed++)
    for(limit=0;limit<=4;limit++)
    for(litMax=0;litMax<=6;litMax+=2)
    for(canon=0;canon<2;canon++)
    {
        unsigned state=seed;
        Vec_Wec_t *cubes=Vec_WecAlloc(8);
        for(owner=6;owner<8;owner++) {
        for(i=0;i<6;i++) support[i]=i;
        if(mode) {
            for(i=5;i>0;i--) { int k=next(&state)%(i+1), tmp=support[i];support[i]=support[k];support[k]=tmp; }
        }
        width=mode==2?4:3;
        for(i=0;i<width;i++) for(j=i+1;j<width;j++) if(support[i]>support[j]) {int tmp=support[i];support[i]=support[j];support[j]=tmp;}
        for(row=0;row<4;row++)
        {
            Vec_Int_t *cube=Vec_WecPushLevel(cubes);
            Vec_IntPush(cube,owner);
            /* Fixed support, unique phase patterns: no duplicates/containment. */
            for(var=0;var<width;var++) {
                phase = ((row ^ owner) >> var) & 1;
                if(var>=2) phase=next(&state)>>31;
                Vec_IntPush(cube,2*support[var]+phase);
            }
        }
        }
        Fx_FastExtract(cubes,8,limit,litMax,canon,0,0);
        printf("mode%d seed%d limit%d litMax%d canon%d",mode,seed,limit,litMax,canon);
        for(i=0;i<Vec_WecSize(cubes);i++) {
            Vec_Int_t *cube=Vec_WecEntry(cubes,i);
            printf(" [");
            for(j=0;j<Vec_IntSize(cube);j++) printf("%s%d",j?",":"",Vec_IntEntry(cube,j));
            printf("]");
        }
        puts("");
        Vec_WecFree(cubes);
    }
    return 0;
}
