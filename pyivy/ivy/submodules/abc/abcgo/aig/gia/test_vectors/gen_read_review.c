#include "aig/gia/giaGen.c"
static Gia_Man_t *make_reader(int kind,int signs,int reorder)
{
    Gia_Man_t *p=Gia_ManStart(32);int a=Gia_ManAppendCi(p),b=Gia_ManAppendCi(p),c=Gia_ManAppendCi(p),root,tmp;
    a^=signs&1;b^=(signs>>1)&1;c^=(signs>>2)&1;
    if(kind==0)root=Gia_ManAppendAnd(p,a,b);else if(kind==1)root=Gia_ManAppendXor(p,a,b);else root=Gia_ManAppendMux(p,c,a,b);
    Gia_ManAppendCo(p,root^((signs>>3)&1));Gia_ManAppendCo(p,1);p->pName=Abc_UtilStrsav("reader");
    if(reorder) {
        tmp=Vec_IntEntry(p->vCis,0);Vec_IntWriteEntry(p->vCis,0,Vec_IntEntry(p->vCis,2));Vec_IntWriteEntry(p->vCis,2,tmp);
        Gia_ObjSetCioId(Gia_ManCi(p,0),0);Gia_ObjSetCioId(Gia_ManCi(p,2),2);
        tmp=Vec_IntEntry(p->vCos,0);Vec_IntWriteEntry(p->vCos,0,Vec_IntEntry(p->vCos,1));Vec_IntWriteEntry(p->vCos,1,tmp);
        Gia_ObjSetCioId(Gia_ManCo(p,0),0);Gia_ObjSetCioId(Gia_ManCo(p,1),1);
    }
    return p;
}
int main(int argc,char **argv)
{
    int ins=-77,words=-88,ret,i,kind,signs,reorder,stats;Vec_Wrd_t *inputs;Vec_Int_t *labels;Gia_Man_t *p;
    if(argc<3)return 2;
    if(argv[1][0]=='p') {ret=Gia_ManSimParamRead(argv[2],&ins,&words);printf("param%d ins%d words%d\n",ret,ins,words);return 0;}
    if(argc!=8)return 2;
    kind=atoi(argv[3]);signs=atoi(argv[4]);reorder=atoi(argv[5]);stats=atoi(argv[6]);p=make_reader(kind,signs,reorder);
    if(argv[1][0]=='d') {Gia_ManTestOneFile(p,argv[2],stats?"stats.json":NULL);Gia_ManStop(p);return 0;}
    ret=Gia_ManSimParamRead(argv[2],&ins,&words);printf("param%d ins%d words%d\n",ret,ins,words);
    if(!ret){Gia_ManStop(p);return 0;}
    inputs=Vec_WrdStart(ins*words);labels=Vec_IntAlloc(words*64);
    if(atoi(argv[7])) {Vec_WrdFill(inputs,ins*words,~(word)0);Vec_IntPush(labels,1);Vec_IntPush(labels,0);}
    Gia_ManSimFileRead(argv[2],ins,words,inputs,labels);
    printf("data");for(i=0;i<Vec_WrdSize(inputs);i++)printf(" %016llx",(unsigned long long)Vec_WrdEntry(inputs,i));
    printf(" labels");for(i=0;i<Vec_IntSize(labels);i++)printf("%d",Vec_IntEntry(labels,i));puts("");
    if(!atoi(argv[7]) && argv[1][0]!='l')Gia_ManCompareValues(p,inputs,labels,stats?"stats.json":NULL);
    Vec_WrdFree(inputs);Vec_IntFree(labels);Gia_ManStop(p);return 0;
}
