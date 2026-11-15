#include "aig/gia/giaGen.c"
#include <sys/resource.h>

static Gia_Man_t *make_graph(int typed,int signs,int reorder)
{
    Gia_Man_t *p=Gia_ManStart(64);
    int a=Gia_ManAppendCi(p),b=Gia_ManAppendCi(p),c,d,x,z,root,buf,tmp,i;
    Gia_Obj_t *obj;
    x=typed?Gia_ManAppendXor(p,a^(signs&1),b^((signs>>1)&1)):Gia_ManAppendAnd(p,a^(signs&1),b^((signs>>1)&1));
    Gia_ManAppendCo(p,x^((signs>>2)&1));
    c=Gia_ManAppendCi(p);d=Gia_ManAppendCi(p);
    if(typed)root=Gia_ManAppendMux(p,c^((signs>>3)&1),x^((signs>>4)&1),d^((signs>>5)&1));
    else {z=Gia_ManAppendAnd(p,x,c^((signs>>3)&1));root=Gia_ManAppendAnd(p,z,d^((signs>>4)&1));}
    Gia_ManAppendCo(p,root^((signs>>5)&1));Gia_ManAppendCo(p,1);
    buf=Gia_ManAppendBuf(p,d^((signs>>6)&1));Gia_ManAppendCo(p,buf);
    Gia_ManSetRegNum(p,1);
    if(reorder) {
        tmp=Vec_IntEntry(p->vCis,0);Vec_IntWriteEntry(p->vCis,0,Vec_IntEntry(p->vCis,3));Vec_IntWriteEntry(p->vCis,3,tmp);
        tmp=Vec_IntEntry(p->vCos,0);Vec_IntWriteEntry(p->vCos,0,Vec_IntEntry(p->vCos,2));Vec_IntWriteEntry(p->vCos,2,tmp);
        Gia_ManForEachCi(p,obj,i)Gia_ObjSetCioId(obj,i);
        Gia_ManForEachCo(p,obj,i)Gia_ObjSetCioId(obj,i);
    }
    Gia_ManForEachObj(p,obj,i)obj->Value=17+3*i;
    return p;
}
int main(int argc, char **argv)
{
    int typed,signs,reorder,wi,widths[]={1,2,5},seed,explicit,i,w;Gia_Obj_t *obj;
    struct rlimit core={0,0};setrlimit(RLIMIT_CORE,&core);
    if(argc>1) {
        Gia_Man_t *p=make_graph(0,0,0);Vec_Wrd_t *in=Vec_WrdStart(0);
        if(argv[1][0]=='e')Gia_ManSimulateWordsInit(p,in);else Gia_ManSimulateWords(p,0);
        Vec_WrdFree(in);Gia_ManStop(p);return 0;
    }
    for(typed=0;typed<2;typed++)for(signs=0;signs<(typed?128:64);signs++)for(reorder=0;reorder<2;reorder++)
    for(wi=0;wi<3;wi++)for(seed=0;seed<=7;seed+=7)for(explicit=0;explicit<2;explicit++) {
        Gia_Man_t *p=make_graph(typed,signs,reorder);int words=widths[wi];Vec_Wrd_t *inputs=Vec_WrdStart(4*words);
        for(i=0;i<4;i++)for(w=0;w<words;w++)Vec_WrdWriteEntry(inputs,i*words+w,0x9e3779b97f4a7c15ull*(i+1)^(0xd1b54a32d192ed03ull*(w+1))^seed);
        Gia_ManRandomW(1);for(i=0;i<seed;i++)Gia_ManRandomW(0);
        if(explicit)Gia_ManSimulateWordsInit(p,inputs);else Gia_ManSimulateWords(p,words);
        printf("typed%d signs%d reorder%d words%d seed%d explicit%d",typed,signs,reorder,words,seed,explicit);
        if(!typed)Gia_ManForEachObj(p,obj,i) {printf(" obj%d",i);for(w=0;w<words;w++)printf(" %016llx",(unsigned long long)Gia_ManObjSim(p,i)[w]);}
        Gia_ManForEachCi(p,obj,i){printf(" ci%d",i);for(w=0;w<words;w++)printf(" %016llx",(unsigned long long)Gia_ManObjSim(p,Gia_ObjId(p,obj))[w]);}
        Gia_ManForEachCo(p,obj,i){printf(" co%d",i);for(w=0;w<words;w++)printf(" %016llx",(unsigned long long)Gia_ManObjSim(p,Gia_ObjId(p,obj))[w]);}
        Gia_ManForEachObj(p,obj,i)assert(obj->Value==(unsigned)(17+3*i));
        printf(" next%016llx\n",(unsigned long long)Gia_ManRandomW(0));
        Vec_WrdFree(inputs);Gia_ManStop(p);
    }
    return 0;
}
