/* Execute the actual glitch implementation, without replacing its algorithms. */
#include "aig/gia/giaGlitch.c"
#include <sys/resource.h>

static Vec_Int_t *reviewIds;
static int ordinal(int handle) { int id=Vec_IntFind(reviewIds,handle); assert(id>=0); return id; }
static void print_ids(const char *name, Vec_Int_t *ids)
{
    int i,id;printf("%s",name);Vec_IntForEachEntry(ids,id,i)printf(" %d",ordinal(id));puts("");
}
static void state(Gli_Man_t *p,const char *name)
{
    Gli_Obj_t *obj,*next;int i,k;
    printf("state %s %d %d %d %d %d %d\n",name,Gli_ManCiNum(p),Gli_ManCoNum(p),Gli_ManPiNum(p),Gli_ManPoNum(p),Gli_ManNodeNum(p),Gli_ManRegNum(p));
    Gli_ManForEachObj(p,obj,i) {
        printf("obj%d %u %u %u %u %u %u %08x %d %d",ordinal(obj->Handle),obj->fTerm,obj->fPhase,obj->fPhase2,obj->fMark,obj->nFanins,obj->nFanouts,obj->uSimInfo,obj->nSwitches,obj->nGlitches);
        printf(" in");Gli_ObjForEachFanin(obj,next,k)printf(" %d",ordinal(next->Handle));
        printf(" out");Gli_ObjForEachFanout(obj,next,k)printf(" %d",ordinal(next->Handle));puts("");
    }
    print_ids("changed",p->vCisChanged);print_ids("affected",p->vAffected);print_ids("frontier",p->vFrontier);
    printf("saved");if(p->pSimInfoPrev)for(i=0;i<Gli_ManCiNum(p);i++)printf(" %08x",p->pSimInfoPrev[i]);puts("");
}
int main(int argc,char **argv)
{
    struct rlimit core={0,0};FILE *file;Gli_Man_t *p;Gli_Obj_t *obj;
    int nCi,nRegs,nNodes,nCo,nObjs,nEdges=0,i,j,mode,steps,prefix,*arity,*fanouts,**ins,*outputs;
    word *truth;float prob;Vec_Int_t *fanin;
    setrlimit(RLIMIT_CORE,&core);setvbuf(stdout,NULL,_IONBF,0);
    if(argc!=7)return 2;
    if(atoi(argv[2])==5){
        unsigned input=(unsigned)strtoul(argv[3],NULL,0);int calls=atoi(argv[4]);float probability=strtof(argv[5],NULL);
        Gia_ManRandom(1);for(i=0;i<calls;i++){input=Gli_ManUpdateRandomInput(input,probability);printf("word %08x\n",input);}printf("random %08x\n",Gia_ManRandom(0));return 0;
    }
    file=fopen(argv[1],"r");assert(file);
    mode=atoi(argv[2]);steps=atoi(argv[3]);prefix=atoi(argv[4]);prob=strtof(argv[5],NULL);
    assert(fscanf(file,"%d %d %d %d",&nCi,&nRegs,&nNodes,&nCo)==4);
    nObjs=nCi+nNodes+nCo;arity=calloc(nNodes,sizeof(int));ins=calloc(nNodes,sizeof(int*));
    fanouts=calloc(nObjs,sizeof(int));outputs=calloc(nCo,sizeof(int));truth=calloc(nNodes,sizeof(word));
    for(i=0;i<nNodes;i++) {
        unsigned long long value;assert(fscanf(file,"%d %llx",&arity[i],&value)==2);truth[i]=value;
        ins[i]=calloc(arity[i],sizeof(int));
        for(j=0;j<arity[i];j++){assert(fscanf(file,"%d",&ins[i][j])==1);fanouts[ins[i][j]]++;nEdges++;}
    }
    for(i=0;i<nCo;i++){assert(fscanf(file,"%d",&outputs[i])==1);fanouts[outputs[i]]++;nEdges++;}fclose(file);
    p=Gli_ManAlloc(nObjs,nRegs,nEdges);reviewIds=Vec_IntAlloc(nObjs);fanin=Vec_IntAlloc(6);
    for(i=0;i<nCi;i++)Vec_IntPush(reviewIds,Gli_ManCreateCi(p,fanouts[i]));
    for(i=0;i<nNodes;i++){
        Vec_IntClear(fanin);for(j=0;j<arity[i];j++)Vec_IntPush(fanin,Vec_IntEntry(reviewIds,ins[i][j]));
        Vec_IntPush(reviewIds,Gli_ManCreateNode(p,fanin,fanouts[nCi+i],&truth[i]));
    }
    for(i=0;i<nCo;i++)Vec_IntPush(reviewIds,Gli_ManCreateCo(p,Vec_IntEntry(reviewIds,outputs[i])));
    if(mode==6){
        printf("ports ro%d/%d ri%d/%d\n",ordinal(Gli_ManRo(p,0)->Handle),ordinal(Gli_ManCi(p,Gli_ManPiNum(p))->Handle),ordinal(Gli_ManRi(p,0)->Handle),ordinal(Gli_ManCo(p,Gli_ManPoNum(p))->Handle));goto cleanup;
    }
    if(mode==4){printf("headerWords %d\n",(int)(sizeof(Gli_Obj_t)/4));Gli_ManPrintObjects(p);goto cleanup;}
    if(mode==3){Gli_ManSwitchesAndGlitches(p,steps,prob,atoi(argv[6]));state(p,"driver");goto random;}
    Gli_ManFinalize(p);state(p,"initial");Gia_ManRandom(1);
    if(mode==0){
        for(i=0;i<steps;i++){
            Gli_ManSetPiRandom(p,prob);state(p,"inputs");
            Gli_ManSwitching(p);state(p,"switching");Gli_ManGlitching(p);state(p,"glitching");
        }
    }else{
        Gli_ManSimulateSeqPref(p,prefix);state(p,"prefix");
        for(i=0;i<32;i++){
            Gli_ManSetDataSaved(p,i);state(p,"saved-bit");
            for(j=0;j<steps;j++){
                if(mode==1)Gli_ManSetPiRandomSeq(p,prob);
                else{
                    int k;Gli_ManForEachCi(p,obj,k)obj->uSimInfo^=(unsigned)(0x9e3779b9u*(k+j+1));
                    Gli_ManSetPiFromSaved(p,i);
                }
                state(p,"inputs");Gli_ManSwitching(p);state(p,"switching");Gli_ManGlitching(p);state(p,"glitching");
            }
        }
    }
random:printf("random %08x\n",Gia_ManRandom(0));
cleanup:
    Vec_IntFree(fanin);Vec_IntFree(reviewIds);Gli_ManStop(p);
    for(i=0;i<nNodes;i++)free(ins[i]);free(ins);free(arity);free(fanouts);free(outputs);free(truth);return 0;
}
