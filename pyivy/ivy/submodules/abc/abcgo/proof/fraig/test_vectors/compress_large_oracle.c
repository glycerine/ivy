#include "proof/fraig/fraigInt.h"
#include <stdio.h>

int main(void) {
    enum { N=6, W=16 };
    Fraig_Man_t p = {0}; Fraig_Node_t n[N+3] = {{0}};
    Fraig_Node_t *inputs[N], *nodes[N+3], *bins[1];
    Fraig_NodeVec_t iv={0}, nv={0}; Fraig_HashTable_t table={0};
    unsigned storage[N+3][W+1]={{0}}, real[W]={0}, temp[W]={0};
    for(int i=0;i<N+3;i++) {nodes[i]=&n[i]; n[i].Num=i; n[i].NumPi=-1; n[i].puSimR=storage[i]; n[i].puSimD=storage[i]+1;}
    for(int i=0;i<N;i++) { inputs[i]=&n[i+1]; n[i+1].NumPi=i; for(int w=0;w<4;w++) n[i+1].puSimD[w]=(0x9e3779b9u*(i+1))^(0x6a09e667u*(w+1)); n[i+1].puSimD[0]&=~0x7fffffffu; if(i+1<N) n[i+1].pNextD=&n[i+2]; }
    n[N+1].p1=&n[1]; n[N+1].p2=Fraig_Not(&n[2]);
    n[N+2].p1=&n[N+1]; n[N+2].p2=&n[3];
    iv.nSize=iv.nCap=N; iv.pArray=inputs; nv.nSize=nv.nCap=N+3; nv.pArray=nodes;
    bins[0]=&n[1]; table.nBins=1; table.pBins=bins;
    p.pConst1=&n[0]; p.vInputs=&iv; p.vNodes=&nv; p.vCones=Fraig_NodeVecAlloc(8); p.mmSims=Fraig_MemFixedStart(4*(W+1));
    p.pTableF=&table; p.nWordsRand=1; p.nWordsDyna=W; p.iWordStart=4; p.iWordPerm=1; p.iPatsPerm=31;
    p.pSimsReal=real; p.pSimsTemp=temp; real[0]=0x7fffffffu; p.vPatsReal=Msat_IntVecAlloc(16);
    for(int round=0;round<2;round++) {
        if(round) {
            int start=p.iWordStart;
            for(int i=0;i<N;i++) for(int w=0;w<2;w++) n[i+1].puSimD[start+w]=(0x85ebca6bu*(i+1))^(0xc2b2ae35u*(w+1));
            p.iWordStart+=2; Msat_IntVecPush(p.vPatsReal,32*start+3); Msat_IntVecPush(p.vPatsReal,32*start+35);
        } else {Msat_IntVecPush(p.vPatsReal,35); Msat_IntVecPush(p.vPatsReal,64);}
        Fraig_FeedBackCompress(&p);
        printf("round %d %d %d %d %d\n",round,p.iPatsPerm,p.iWordPerm,p.iWordStart,p.nWordsDyna);
        for(int i=1;i<N+3;i++) {printf("node %d %08x",i,n[i].uHashD);for(int w=0;w<p.iWordPerm;w++) printf(" %08x",n[i].puSimD[w]); puts("");}
        printf("mask");for(int w=0;w<p.iWordPerm;w++) printf(" %08x",p.pSimsReal[w]);puts("");
    }
    Msat_IntVecFree(p.vPatsReal);Fraig_NodeVecFree(p.vCones);Fraig_MemFixedStop(p.mmSims,0);return 0;
}
