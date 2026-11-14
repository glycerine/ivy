#include "aig/aig/aigShow.c"
#include <unistd.h>

int main(void)
{
    Aig_Man_t g;
    Aig_Obj_t o[9];
    Vec_Ptr_t *bold;
    int used, highlighted, history, i, total, ch, fd;
    FILE *file;
    while (scanf("%d%d%d", &used, &highlighted, &history) == 3)
    {
        char path[] = "/tmp/abcgo-dot-XXXXXX";
        memset(&g,0,sizeof(g)); memset(o,0,sizeof(o));
        total = used ? 9 : 8;
        g.pConst1=o; g.vObjs=Vec_PtrAlloc(total); g.vCis=Vec_PtrAlloc(2); g.vCos=Vec_PtrAlloc(3);
        o[0].Type=AIG_OBJ_CONST1;
        for (i=1; i<=2; i++) { o[i].Type=AIG_OBJ_CI; Vec_PtrPush(g.vCis,o+i); }
        for (i=3; i<=5; i++) o[i].Type=AIG_OBJ_AND;
        o[3].pFanin0=o+1; o[3].pFanin1=Aig_Not(o+2); o[3].Level=1;
        o[4].pFanin0=Aig_Not(o+1); o[4].pFanin1=o+2; o[4].Level=1;
        o[5].pFanin0=o+3; o[5].pFanin1=Aig_Not(o+4); o[5].Level=2;
        for (i=6; i<total; i++) { o[i].Type=AIG_OBJ_CO; Vec_PtrPush(g.vCos,o+i); }
        o[6].pFanin0=o+5; o[6].Level=2;
        o[7].pFanin0=Aig_Not(o+3); o[7].Level=1;
        if (used) o[8].pFanin0=Aig_Not(o);
        for (i=0; i<total; i++)
        {
            o[i].Id=i; g.nObjs[o[i].Type]++; Vec_PtrPush(g.vObjs,o+i);
            if (o[i].pFanin0) Aig_ObjFanin0(o+i)->nRefs++;
            if (o[i].pFanin1) Aig_ObjFanin1(o+i)->nRefs++;
        }
        bold=NULL;
        if (highlighted) { bold=Vec_PtrAlloc(2); Vec_PtrPush(bold,o+3); Vec_PtrPush(bold,o+5); }
        fd=mkstemp(path); assert(fd>=0); close(fd);
        Aig_WriteDotAig(&g,path,history,bold);
        file=fopen(path,"r"); assert(file);
        while ((ch=fgetc(file)) != EOF) putchar(ch);
        fclose(file); unlink(path);
        for (i=0; i<total; i++) assert(o[i].fMarkB==0);
        for (i=6; i<total; i++) assert(o[i].Level==Aig_ObjFanin0(o+i)->Level);
        if (bold) Vec_PtrFree(bold);
        Vec_PtrFree(g.vObjs); Vec_PtrFree(g.vCis); Vec_PtrFree(g.vCos);
    }
    return 0;
}
