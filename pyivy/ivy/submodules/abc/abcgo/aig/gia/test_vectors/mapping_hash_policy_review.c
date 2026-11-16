#include "aig/gia/giaIf.c"
#include <sys/resource.h>
static void cut(Gia_Man_t *p,int root,int *leaves,int count)
{
    int i;Vec_IntWriteEntry(p->vMapping,root>>1,Vec_IntSize(p->vMapping));Vec_IntPush(p->vMapping,count);
    for(i=0;i<count;i++)Vec_IntPush(p->vMapping,leaves[i]>>1);Vec_IntPush(p->vMapping,root>>1);
}
int main(int argc,char **argv)
{
    struct rlimit core={0,0};Gia_Man_t *p,*q;int mode,a,b,c,d=0,x,z,later=0,y,root,i,leaves[3];
    setrlimit(RLIMIT_CORE,&core);if(argc!=2)return 2;mode=atoi(argv[1]);p=Gia_ManStart(20);
    a=Gia_ManAppendCi(p);b=Gia_ManAppendCi(p);c=Gia_ManAppendCi(p);if(mode==2)d=Gia_ManAppendCi(p);
    x=Gia_ManAppendAnd(p,a,b);z=Gia_ManAppendAnd(p,a,b);if(mode==2)later=Gia_ManAppendAnd(p,c,d);
    y=mode?Gia_ManAppendAnd(p,x,z):z;Gia_ManAppendCo(p,y);p->vMapping=Vec_IntStart(p->nObjs);
    leaves[0]=a;leaves[1]=b;cut(p,x,leaves,2);leaves[2]=c;cut(p,z,leaves,mode?2:3);
    if(mode==2){leaves[0]=c;leaves[1]=d;cut(p,later,leaves,2);leaves[0]=a;leaves[1]=b;leaves[2]=later;cut(p,y,leaves,3);}
    else if(mode==1){leaves[0]=x;leaves[1]=z;cut(p,y,leaves,2);}
    q=Gia_ManDupHashMapping(p);root=Gia_ObjFaninId0p(q,Gia_ManCo(q,0));printf("root %d cut",root);
    Gia_LutForEachFanin(q,root,a,i)printf(" %d",a);puts("");Gia_ManStop(q);Gia_ManStop(p);return 0;
}
