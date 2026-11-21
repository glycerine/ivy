/* Link in place of ivyFraig.o to access native simulation and class links. */
#include "aig/ivy/ivyFraig.c"
#include <time.h>

static int check_collisions(void)
{
    static const unsigned words[8][2] = {
        {0, 0}, {1049, 1009}, {2098, 2018}, {0, 0},
        {2098, 2018}, {1049, 1009}, {5, 1}, {5, 1}
    };
    static const int members[4][4] = {{0, 1, 4, -1}, {2, 6, -1}, {3, 5, -1}, {7, 8, -1}};
    Ivy_Man_t *net = Ivy_ManStart();
    for (int i = 0; i < 8; i++) Ivy_ObjCreatePi(net);
    Ivy_FraigParams_t params;
    Ivy_FraigParamsDefault(&params);
    params.nSimWords = 2;
    Ivy_FraigMan_t *p = Ivy_FraigStart(net, &params);
    Ivy_Man_t *result = p->pManFraig;
    for (int i = 0; i < 8; i++)
        memcpy(Ivy_ObjSim(Ivy_ManPi(net, i))->pData, words[i], sizeof(words[i]));
    memcpy(Ivy_ObjSim(net->pConst1)->pData, words[6], sizeof(words[6]));
    Ivy_FraigCreateClasses(p);
    Ivy_Obj_t *head = p->lClasses.pHead;
    for (int group = 0; group < 4; group++) {
        Ivy_Obj_t *member = head;
        for (int i = 0; members[group][i] >= 0; i++) {
            if (!member || member->Id != members[group][i]) return 0;
            member = Ivy_ObjClassNodeNext(member);
        }
        if (member) return 0;
        head = Ivy_ObjEquivListNext(head);
    }
    int ok = !head && p->lClasses.nItems == 4;
    Ivy_FraigStop(p);
    Ivy_ManStop(result);
    Ivy_ManStop(net);
    return ok;
}

int main(void)
{
    if (!check_collisions()) return 1;
    Ivy_Man_t *net = Ivy_ManStart();
    for (int i = 0; i < 1024; i++) Ivy_ObjCreatePi(net);
    Ivy_FraigParams_t params;
    Ivy_FraigParamsDefault(&params);
    params.nSimWords = 32;
    Ivy_FraigMan_t *p = Ivy_FraigStart(net, &params);
    Ivy_Man_t *result = p->pManFraig;
    Ivy_Obj_t *obj;
    int i;
    Ivy_ManForEachPi(net, obj, i)
        for (int w = 0; w < 32; w++)
            Ivy_ObjSim(obj)->pData[w] = (i/2+1)*(2*w+1);
    struct timespec a, b;
    const int runs = 20000;
    clock_gettime(CLOCK_PROCESS_CPUTIME_ID, &a);
    for (int r = 0; r < runs; r++) {
        Ivy_FraigCreateClasses(p);
        if (p->lClasses.nItems != 512 || Ivy_ObjClassNodeNext(p->lClasses.pHead) != Ivy_ManPi(net, 1)) return 2;
        memset(&p->lClasses, 0, sizeof(p->lClasses));
        Ivy_ManForEachPi(net, obj, i) {
            Ivy_ObjSetClassNodeRepr(obj, NULL);
            Ivy_ObjSetClassNodeNext(obj, NULL);
            Ivy_ObjSetEquivListNext(obj, NULL);
            Ivy_ObjSetEquivListPrev(obj, NULL);
        }
    }
    clock_gettime(CLOCK_PROCESS_CPUTIME_ID, &b);
    printf("%.2f us/512 classes\n",
           ((b.tv_sec-a.tv_sec)*1e6 + (b.tv_nsec-a.tv_nsec)*1e-3)/runs);
    Ivy_FraigStop(p);
    Ivy_ManStop(result);
    Ivy_ManStop(net);
    return 0;
}
