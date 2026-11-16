#define ABC_USE_CUDD
#define ABC_USE_PTHREADS
#include "map/if/ifDsd.c"
#include <sys/resource.h>
static If_DsdMan_t *build(int scenario)
{
    char *texts[]={"(ab)","[abc]","E8{abc}","E8{a(bc)d}","D3C5{abcd}"};
    int sizes[]={2,3,3,4,4};If_DsdMan_t *p=If_DsdManAlloc(scenario>=7?scenario-7:4,scenario>=7||scenario==2?0:scenario>=3?scenario:2);
    unsigned char perm[4];int i,j;word *truth;
    for(i=0;i<(scenario>=7?0:5);i++)for(j=0;j<=i%3;j++) {
        truth=Dau_DsdToTruth(texts[i],sizes[i]);If_DsdManCompute(p,truth,sizes[i],perm,"");
    }
    if(scenario==1){Id_DsdManTuneStr1(p,"{abcd}",0,0);If_DsdManSetLutSize(p,2);}
    return p;
}
static void report(If_DsdMan_t *p)
{
    int i,j;word *config;
    printf("\nreport %d %d %d %d\n",p->nVars,p->LutSize,If_DsdManObjNum(p),p->nTtBits);
    printf("cell %s\n",p->pCellStr?p->pCellStr:"-");
    for(i=0;i<If_DsdManObjNum(p);i++) {
        If_DsdManPrintOne(stdout,p,i,NULL,1);
        if(p->vConfigs && p->nConfigWords*(i+1)<=Vec_WrdSize(p->vConfigs)) {
            config=If_DsdManGetFuncConfig(p,2*i);printf("config %d",i);
            for(j=0;j<p->nConfigWords;j++)printf(" %016llx",(unsigned long long)config[j]);puts("");
        }
    }
}
int main(int argc,char **argv)
{
    struct rlimit core={0,0};If_DsdMan_t *p;setrlimit(RLIMIT_CORE,&core);
    if(argc<3)return 2;
    if(!strcmp(argv[1],"write")){if(argc!=4)return 2;p=build(atoi(argv[3]));If_DsdManSave(p,argv[2]);}
    else {p=If_DsdManLoad(argv[2]);if(!p)return 3;}
    report(p);
    if(strcmp(argv[1],"write")&&argc==4)If_DsdManSave(p,argv[3]);
    If_DsdManFree(p,0);return 0;
}
