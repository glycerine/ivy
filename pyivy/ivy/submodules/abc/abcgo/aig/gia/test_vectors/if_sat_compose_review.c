#include "map/if/ifSat.c"
static word nextWord(word *state) { return *state = *state * ABC_CONST(6364136223846793005) + ABC_CONST(1442695040888963407); }
int main(void)
{
    word state=1234567, functions[4],bound,free;
    int k,trial,i,j,order[6],nb,ns,nf;
    for(k=0;k<=4;k++)for(trial=0;trial<128;trial++) {
        unsigned table=(unsigned)nextWord(&state);
        for(i=0;i<4;i++)functions[i]=nextWord(&state);
        printf("compose %016llx\n",(unsigned long long)If_ManSat6ComposeLut4(table,functions,k));
    }
    for(trial=0;trial<512;trial++) {
        for(i=0;i<6;i++)order[i]=i;
        for(i=5;i>0;i--){int temp;j=nextWord(&state)%(i+1);temp=order[i];order[i]=order[j];order[j]=temp;}
        ns=nextWord(&state)%4;
        nb=nextWord(&state)%(5-ns);
        nf=nextWord(&state)%(4-ns);
        if(nb+ns+nf>6)nf=6-nb-ns;
        bound=nextWord(&state);free=nextWord(&state);
        printf("truth %016llx\n",(unsigned long long)If_ManSat6Truth(bound,free,order,nb,order+nb,ns,order+nb+ns,nf));
    }
    return 0;
}
