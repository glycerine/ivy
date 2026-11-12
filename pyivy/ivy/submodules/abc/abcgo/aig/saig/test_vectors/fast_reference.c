/* Test harness only. The Go test appends unmodified routines extracted from
 * src/aig/saig/saigSimFast.c and src/aig/aig/aigUtil.c before this main. */
int main(void) {
    int pis, regs, nodes, pos, frames, prefix, transitions;
    if (scanf("%d%d%d%d%d%d%d", &pis,&regs,&nodes,&pos,&frames,&prefix,&transitions)!=7) return 2;
    int entries=2*nodes+pos+regs;
    Faig_Man_t *p=calloc(1,sizeof(Faig_Man_t)+sizeof(int)*entries);
    p->nPis=pis;p->nFfs=regs;p->nCis=pis+regs;p->nNos=nodes;p->nPos=pos;p->nCos=pos+regs;
    p->nPis1=pis+1;p->nCis1=pis+regs+1;p->nCisNos1=p->nCis1+nodes;p->nCisNosPos1=p->nCisNos1+pos;p->nObjs=p->nCisNos1+p->nCos;
    for(int i=0;i<entries;i++)if(scanf("%d",p->pObjs+i)!=1)return 3;
    Aig_ManRandom(1);
    int *counts=Faig_ManSimulateFrames(p,frames,prefix,transitions);
    for(int i=0;i<p->nObjs;i++)printf("%d ",counts[i]);
    free(counts);free(p);return 0;
}
