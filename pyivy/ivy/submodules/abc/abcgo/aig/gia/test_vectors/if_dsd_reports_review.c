#define ABC_USE_CUDD
#define ABC_USE_PTHREADS
#include "map/if/ifDsd.c"
int main(int argc,char **argv)
{
    If_DsdMan_t *p=If_DsdManAlloc(4,2);unsigned char perm[4];int i,j;
    char *texts[]={"(abc)","E8{abc}","E8{a(bc)d}","D3C5{abcd}","<a(bc)d>"};
    int sizes[]={3,3,4,4,4};
    if(argc!=6)return 2;
    for(i=0;i<5;i++)for(j=0;j<=i;j++)If_DsdManCompute(p,Dau_DsdToTruth(texts[i],sizes[i]),sizes[i],perm,"");
    If_DsdManPrint(p,strcmp(argv[5],"-")?argv[5]:NULL,atoi(argv[2]),atoi(argv[3]),atoi(argv[4]),0,atoi(argv[1]));
    If_DsdManFree(p,0);return 0;
}
