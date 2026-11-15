#include "gen_graph_review.h"
int main(int argc,char **argv)
{
    Gia_Man_t *p;char *base,*name;int train,valid,seed,typed,signs,reorder,fallback;
    if(argc!=10)return 2;
    base=argv[2];train=atoi(argv[3]);valid=atoi(argv[4]);seed=atoi(argv[5]);typed=atoi(argv[6]);signs=atoi(argv[7]);reorder=atoi(argv[8]);fallback=atoi(argv[9]);
    p=make_graph(typed,signs,reorder);p->pName=Abc_UtilStrsav(base);name=fallback?NULL:base;
    if(argv[1][0]=='b')Gia_ManDumpFiles(p,train,valid,seed,name);else Gia_ManDumpPlaFiles(p,train,valid,seed,name);
    printf("state words%d next%016llx\n",p->nSimWords,(unsigned long long)Gia_ManRandomW(0));
    Gia_ManStop(p);return 0;
}
