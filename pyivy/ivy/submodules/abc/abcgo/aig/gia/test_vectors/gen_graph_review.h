#ifndef GEN_GRAPH_REVIEW_H
#define GEN_GRAPH_REVIEW_H
#include "aig/gia/giaGen.c"

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

#endif
