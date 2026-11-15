#include "aig/gia/giaFx.c"
int main(void)
{
    int mode,i,j,ins[3],x,y,z,roots[2];
    for(mode=0;mode<4;mode++) {
        Gia_Man_t *p=Gia_ManStart(32), *q;
        for(i=0;i<3;i++) ins[i]=Gia_ManAppendCi(p);
        x=Gia_ManAppendAnd(p,ins[0],ins[1]);
        y=Gia_ManAppendAnd(p,x,ins[2]);
        z=Gia_ManAppendAnd(p,x,ins[2]^1);
        Gia_ManAppendCo(p,y);Gia_ManAppendCo(p,z^1);
        p->vMapping=Vec_IntStart(Gia_ManObjNum(p));
        roots[0]=y>>1;roots[1]=z>>1;
        for(i=0;i<2;i++) {
            Vec_IntWriteEntry(p->vMapping,roots[i],Vec_IntSize(p->vMapping));
            Vec_IntPush(p->vMapping,3);
            for(j=0;j<3;j++) Vec_IntPush(p->vMapping,ins[j]>>1);
            Vec_IntPush(p->vMapping,roots[i]);
        }
        printf("mode%d\n",mode);
        q=Gia_ManPerformFx(p,4,0,0,mode&1,(mode>>1)&1);
        printf("result ci%d co%d and%d\n",Gia_ManCiNum(q),Gia_ManCoNum(q),Gia_ManAndNum(q));
        Gia_ManStop(q);Gia_ManStop(p);
    }
    return 0;
}
