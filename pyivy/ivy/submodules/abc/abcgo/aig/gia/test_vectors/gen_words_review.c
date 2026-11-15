#include "gen_graph_review.h"
#include <sys/resource.h>

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
