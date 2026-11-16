#define ABC_USE_CUDD
#define ABC_USE_PTHREADS
#include "map/if/ifDsd.c"
int main(int argc,char **argv)
{
    If_DsdMan_t *p=If_DsdManAlloc(8,4);FILE *input;word truth[4];unsigned char perm[8];int n,i,lit,index=0;unsigned masks[4]={0,0,1,~0u};
    if(argc!=2)return 2;input=fopen(argv[1],"r");assert(input);
    while(fscanf(input,"%d",&n)==1) {
        for(i=0;i<Abc_TtWordNum(n);i++){unsigned long long value;assert(fscanf(input,"%llx",&value)==1);truth[i]=value;}
        lit=If_DsdManCompute(p,truth,n,perm,"44");
        printf("function %d %d %d permutation",index++,lit,If_DsdManObjNum(p));
        for(i=0;i<n;i++)printf(" %d",perm[i]);printf(" partitions");
        for(i=0;i<4;i++)printf(" %x",If_DsdManCheckXY(p,lit,4,i!=0,masks[i],i>=2&&n<=7,0));puts("");
    }
    fclose(input);If_DsdManFree(p,0);return 0;
}
