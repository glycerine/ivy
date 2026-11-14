#include "aig/aig/aigTruth.c"
int main(void)
{
    Aig_Obj_t a={0},b={0},node={0};
    unsigned av[17],bv[17],out[17];
    int counts[]={1,2,4,17},i,n,words,kind,sign;
    for (i=0;i<17;i++) { av[i]=0x9e3779b9u*(i+1)^0x13579bdfu; bv[i]=0x7f4a7c15u*(i+3)+0x2468ace0u; }
    a.pData=av; b.pData=bv;
    for (kind=0;kind<2;kind++) for (sign=0;sign<(kind?1:4);sign++) for (n=0;n<4;n++)
    {
        node.Type=kind?AIG_OBJ_EXOR:AIG_OBJ_AND;
        node.pFanin0=Aig_NotCond(&a,sign&1); node.pFanin1=Aig_NotCond(&b,(sign>>1)&1);
        words=counts[n]; Aig_ManCutTruthOne(&node,out,words);
        printf("kind %d signs %d words %d:",kind,sign,words);
        for (i=0;i<words;i++) printf(" %08x",out[i]);
        puts("");
    }
    return 0;
}
