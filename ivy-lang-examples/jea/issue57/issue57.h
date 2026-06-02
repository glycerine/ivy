#define _HAS_ITERATOR_DEBUGGING 0
struct ivy_gen {virtual int choose(int rng,const char *name) = 0;};
#include "z3++.h"
#include "ivy_hash.hpp"
typedef std::string __strlit;
extern std::ofstream __ivy_out;
void __ivy_exit(int);

template <typename D, typename R>
struct thunk {
    virtual R operator()(const D &) = 0;
    int ___ivy_choose(int rng,const char *name,int id) {
        return 0;
    }
};
template <typename D, typename R, class HashFun = hash_space::hash<D> >
struct hash_thunk {
    thunk<D,R> *fun;
    hash_space::hash_map<D,R,HashFun> memo;
    hash_thunk() : fun(0) {}
    hash_thunk(thunk<D,R> *fun) : fun(fun) {}
    ~hash_thunk() {
//        if (fun)
//            delete fun;
    }
    R &operator[](const D& arg){
        std::pair<typename hash_space::hash_map<D,R>::iterator,bool> foo = memo.insert(std::pair<D,R>(arg,R()));
        R &res = foo.first->second;
        if (foo.second && fun)
            res = (*fun)(arg);
        return res;
    }
};


    class reader;
    class timer;

class issue57 {
  public:
    typedef issue57 ivy_class;

    std::vector<std::string> __argv;
#ifdef _WIN32
    void *mutex;  // forward reference to HANDLE
#else
    pthread_mutex_t mutex;
#endif
    void __lock();
    void __unlock();

#ifdef _WIN32
    std::vector<HANDLE> thread_ids;

#else
    std::vector<pthread_t> thread_ids;

#endif
    void install_reader(reader *);
    void install_thread(reader *);
    void install_timer(timer *);
    virtual ~issue57();
    std::vector<int> ___ivy_stack;
    ivy_gen *___ivy_gen;
    int ___ivy_choose(int rng,const char *name,int id);
    virtual void ivy_assert(bool,const char *){}
    virtual void ivy_assume(bool,const char *){}
    virtual void ivy_check_progress(int,int){}
    enum my_type_2{val1,val2};
    struct __tup__unsigned__my_type_2__unsigned {
        unsigned arg0;
        my_type_2 arg1;
        unsigned arg2;
        __tup__unsigned__my_type_2__unsigned() {}
        __tup__unsigned__my_type_2__unsigned(const unsigned &arg0, const my_type_2 &arg1, const unsigned &arg2) : arg0(arg0), arg1(arg1), arg2(arg2) {}
        size_t __hash() const {
            size_t hv = 0;
            hv += hash_space::hash<unsigned>()(arg0);
            hv += hash_space::hash<int>()(arg1);
            hv += hash_space::hash<unsigned>()(arg2);
            return hv;
        }
    };

    class hash____tup__unsigned__my_type_2__unsigned {
        public:
            size_t operator()(const issue57::__tup__unsigned__my_type_2__unsigned &__s) const {
                return hash_space::hash<unsigned>()(__s.arg0)+hash_space::hash<int>()(__s.arg1)+hash_space::hash<unsigned>()(__s.arg2);
            }
    };

    hash_thunk<__tup__unsigned__my_type_2__unsigned,bool> node__voted;
    bool _generating;
    long long __CARD__my_type_1;
    long long __CARD__my_type_3;

    issue57();
void __init();
    void __tick(int timeout);
};
