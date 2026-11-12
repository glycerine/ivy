/* Original-C packed sequential simulation driver; routines precede this main. */
int main(void) {
    Raig_Man_t p={0};int frames,regs;
    if(scanf("%d%d%d%d%d%d%d%d%d",&p.nWords,&p.nObjs,&p.nPis,&p.nPos,&p.nCis,&p.nCos,&p.nNodes,&frames,&regs)!=9)return 2;
    p.pFans0=calloc(p.nObjs,sizeof(int));p.pFans1=calloc(p.nObjs,sizeof(int));p.pRefs=calloc(p.nObjs,sizeof(int));p.pSims=calloc(p.nObjs,sizeof(unsigned));
    for(int i=0;i<p.nObjs;i++)if(scanf("%d%d%d",p.pFans0+i,p.pFans1+i,p.pRefs+i)!=3)return 3;
    Vec_Int_t cis={p.nCis,calloc(p.nCis,sizeof(int))},los={regs,calloc(regs,sizeof(int))},lis={regs,calloc(regs,sizeof(int))};
    p.vCis2Ids=&cis;p.vLos=&los;p.vLis=&lis;
    for(int i=0;i<p.nCis;i++)if(scanf("%d",cis.pArray+i)!=1)return 4;
    for(int i=0;i<regs;i++)if(scanf("%d%d",los.pArray+i,lis.pArray+i)!=2)return 5;
    Aig_ManRandom(1);
    for(int frame=0;frame<frames;frame++){
        int pattern=-1,failed=Raig_ManSimulateRound(&p,1,frame==0,&pattern);
        printf("%d %d ",failed,pattern);
        if(failed)break;
        for(int reg=0;reg<regs;reg++)for(int word=1;word<=p.nWords;word++)printf("%u ",p.pMems[p.pSims[lis.pArray[reg]]+word]);
    }
    free(p.pFans0);free(p.pFans1);free(p.pRefs);free(p.pSims);free(p.pMems);free(cis.pArray);free(los.pArray);free(lis.pArray);return 0;
}
