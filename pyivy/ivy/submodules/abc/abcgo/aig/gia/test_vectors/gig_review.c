#include "aig/gia/gia.h"
#include "misc/extra/extra.h"
#include "misc/util/utilTruth.h"
#include <assert.h>
#include <sys/resource.h>

/* Keep every source assertion active. Corrected mode intercepts only the
   known LUT6 arity typo; vector-size checks below independently require six. */
static int correctLut6, arityCorrections;
static void review_assert(int condition,const char *expression,const char *file,unsigned line,const char *function)
{
    if(condition)return;
    if(correctLut6 && line==311 && strstr(file,"/giaGig.c") && !strcmp(expression,"k == 4")) {arityCorrections++;return;}
    __assert_fail(expression,file,line,function);
}
#undef assert
#define assert(e) review_assert(!!(e),#e,__FILE__,__LINE__,__func__)
#include "aig/gia/giaGig.c"

static void print_vector(const char *name,Vec_Int_t *v)
{
    int i;printf("%s",name);for(i=0;i<Vec_IntSize(v);i++)printf(" %d",Vec_IntEntry(v,i));puts("");
}
static void print_graph(Gia_Man_t *p)
{
    int i,k,input,a,b,c,value;Gia_Obj_t *obj;
    printf("graph ci%d co%d reg%d objs%d\n",Gia_ManCiNum(p),Gia_ManCoNum(p),Gia_ManRegNum(p),Gia_ManObjNum(p));
    printf("identity %s %s\n",p->pName,p->pSpec);
    Gia_ManForEachObj(p,obj,i) {
        printf("obj%d",i);
        if(i==0)printf(" const0");
        else if(Gia_ObjIsCi(obj))printf(" ci%d",Gia_ObjCioId(obj));
        else if(Gia_ObjIsCo(obj))printf(" co%d lit%d",Gia_ObjCioId(obj),Gia_ObjFaninLit0p(p,obj));
        else if(Gia_ObjIsMux(p,obj))printf(" mux%d %d %d",Gia_ObjFaninLit2p(p,obj),Gia_ObjFaninLit1p(p,obj),Gia_ObjFaninLit0p(p,obj));
        else if(Gia_ObjIsXor(obj))printf(" xor%d %d",Gia_ObjFaninLit0p(p,obj),Gia_ObjFaninLit1p(p,obj));
        else if(Gia_ObjIsBuf(obj))printf(" buf%d",Gia_ObjFaninLit0p(p,obj));
        else printf(" and%d %d",Gia_ObjFaninLit0p(p,obj),Gia_ObjFaninLit1p(p,obj));
        puts("");
    }
    print_vector("cis",p->vCis);print_vector("cos",p->vCos);
    if(Gia_ManCiNum(p)>10)return;
    for(k=0;k<Gia_ManCoNum(p);k++) {
        printf("truth%d ",k);
        for(input=0;input<(1<<Gia_ManCiNum(p));input++) {
            Gia_ManConst0(p)->fMark0=0;
            Gia_ManForEachCi(p,obj,i)obj->fMark0=(input>>i)&1;
            Gia_ManForEachAnd(p,obj,i) {
                a=Gia_ObjFanin0(obj)->fMark0^Gia_ObjFaninC0(obj);b=Gia_ObjFanin1(obj)->fMark0^Gia_ObjFaninC1(obj);
                if(Gia_ObjIsMux(p,obj)) {c=Gia_ObjFanin2(p,obj)->fMark0^Gia_ObjFaninC2(p,obj);value=c?b:a;}
                else if(Gia_ObjIsXor(obj))value=a^b;else value=a&b;
                obj->fMark0=value;
            }
            obj=Gia_ManCo(p,k);putchar('0'+(Gia_ObjFanin0(obj)->fMark0^Gia_ObjFaninC0(obj)));
        }
        puts("");
    }
}
static void print_parsed(FILE *file,Gls_Man_t *p)
{
    char *buffer=malloc(MAX_LINE),*line;int id,type,index,i;Vec_Int_t *args;
    rewind(file);
    while(fgets(buffer,MAX_LINE,file)) {
        line=buffer;id=Gls_ManParseOne(&line);type=Vec_StrEntry(p->vTypes,id);index=Vec_IntEntry(p->vIndexes,id);
        printf("node%d %s",id,s_Strs[type]);
        if(type==GLS_PO||type==GLS_BAR||type==GLS_SEQ||type==GLS_SEL)printf(" %d",index);
        if(type==GLS_LUT4||type==GLS_LUT6) {
            int n=type==GLS_LUT4?4:6;
            args=type==GLS_LUT4?p->vLut4s:p->vLut6s;
            for(i=0;i<n;i++)printf(" %d",Vec_IntEntry(args,n*index+i));
            if(type==GLS_LUT4)printf(" tt%04x",(unsigned)Vec_IntEntry(p->vLut4TTs,index)&0xffff);
            else printf(" tt%016llx",(unsigned long long)Vec_WrdEntry(p->vLut6TTs,index));
        }
        if(type==GLS_BOX)for(i=0;i<5;i++)printf(" %d",Vec_IntEntry(p->vBoxes,5*index+i));
        if(type==GLS_DEL) {args=Vec_WecEntry(p->vDelayIns,index);for(i=0;i<Vec_IntSize(args);i++)printf(" %d",Vec_IntEntry(args,i));printf(" delay%d",Vec_IntEntry(p->vDelays,index));}
        puts("");
    }
    free(buffer);
    print_vector("pis",p->vOrderPis);print_vector("pos",p->vOrderPos);print_vector("boxes",p->vOrderBoxes);
    print_vector("delays",p->vOrderDelays);print_vector("luts",p->vOrderLuts);print_vector("seqs",p->vOrderSeqs);
}
int main(int argc,char **argv)
{
    struct rlimit core={0,0};FILE *file;int counts[GLS_FINAL],lut6;Vec_Str_t *lines;Gls_Man_t *p;Gia_Man_t *g;
    setrlimit(RLIMIT_CORE,&core);setvbuf(stdout,NULL,_IONBF,0);
    if(argc!=3)return 2;
    correctLut6=!strcmp(argv[1],"--correct-lut6");
    if(!correctLut6) {g=Gia_ManReadGig(argv[2]);if(!g)return 1;print_graph(g);Gia_ManStop(g);return 0;}
    file=fopen(argv[2],"rb");if(!file)return 3;
    lines=Gls_ManCount(file,counts);rewind(file);p=Gls_ManAlloc(lines,counts);Gls_ManParse(file,p);lut6=counts[GLS_LUT6];
    assert(Vec_IntSize(p->vLut6s)==6*lut6);assert(arityCorrections==lut6);
    print_parsed(file,p);g=Gls_ManConstruct(p,argv[2]);print_graph(g);Gia_ManStop(g);Gls_ManStop(p);fclose(file);
    puts("read");g=Gia_ManReadGig(argv[2]);assert(g);assert(arityCorrections==2*lut6);print_graph(g);Gia_ManStop(g);
    printf("corrections%d\n",arityCorrections);return 0;
}
