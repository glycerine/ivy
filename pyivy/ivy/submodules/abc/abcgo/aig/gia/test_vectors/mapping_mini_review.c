#include "aig/gia/giaIf.c"
#include <sys/resource.h>
int main(int argc,char **argv)
{
    struct rlimit core={0,0};FILE *file;Gia_Man_t *p;Gia_Obj_t *obj;Vec_Int_t *leaves,*mini;
    int inputs,nLeaves,nMini,hash,i,k,value,root,mask;
    setrlimit(RLIMIT_CORE,&core);setvbuf(stdout,NULL,_IONBF,0);if(argc!=2)return 2;file=fopen(argv[1],"r");assert(file);
    assert(fscanf(file,"%d %d %d %d",&inputs,&nLeaves,&nMini,&hash)==4);p=Gia_ManStart(64);
    for(i=0;i<inputs;i++)Gia_ManAppendCi(p);leaves=Vec_IntAlloc(nLeaves);mini=Vec_IntAlloc(nMini);
    for(i=0;i<nLeaves;i++){assert(fscanf(file,"%d",&value)==1);Vec_IntPush(leaves,value);}
    for(i=0;i<nMini;i++){assert(fscanf(file,"%d",&value)==1);Vec_IntPush(mini,value);}fclose(file);
    if(hash)Gia_ManHashAlloc(p);
    root=Gia_ManBuildFromMiniInt(p,leaves,mini,hash);printf("root %d\nscratch %d",root,Vec_IntSize(mini));for(i=0;i<Vec_IntSize(mini);i++)printf(" %d",Vec_IntEntry(mini,i));puts("");
    Gia_ManAppendCo(p,root);printf("objects %d\n",Gia_ManObjNum(p));
    Gia_ManForEachObj(p,obj,i){
        if(i==0)puts("const0");else if(Gia_ObjIsCi(obj))printf("ci %d\n",Gia_ObjCioId(obj));
        else if(Gia_ObjIsCo(obj))printf("co %d\n",Gia_ObjFaninLit0p(p,obj));
        else printf("and %d %d\n",Gia_ObjFaninLit0p(p,obj),Gia_ObjFaninLit1p(p,obj));
    }
    printf("truth ");for(mask=0;mask<(1<<inputs);mask++){
        Gia_ManConst0(p)->fMark0=0;Gia_ManForEachCi(p,obj,k)obj->fMark0=(mask>>k)&1;
        Gia_ManForEachAnd(p,obj,i)obj->fMark0=(Gia_ObjFanin0(obj)->fMark0^Gia_ObjFaninC0(obj))&(Gia_ObjFanin1(obj)->fMark0^Gia_ObjFaninC1(obj));
        obj=Gia_ManCo(p,0);putchar('0'+(Gia_ObjFanin0(obj)->fMark0^Gia_ObjFaninC0(obj)));
    }puts("");Vec_IntFree(leaves);Vec_IntFree(mini);Gia_ManStop(p);return 0;
}
