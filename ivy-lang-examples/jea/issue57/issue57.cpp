#include "issue57.h"

#include <sstream>
#include <algorithm>

#include <iostream>
#include <stdlib.h>
#include <sys/types.h>          /* See NOTES */
#include <sys/stat.h>
#include <fcntl.h>
#ifdef _WIN32
#include <winsock2.h>
#include <WS2tcpip.h>
#include <io.h>
#define isatty _isatty
#else
#include <sys/socket.h>
#include <netinet/in.h>
#include <netinet/ip.h> 
#include <sys/select.h>
#include <unistd.h>
#define _open open
#define _dup2 dup2
#endif
#include <string.h>
#include <stdio.h>
#include <string>
#if __cplusplus < 201103L
#else
#include <cstdint>
#endif
typedef issue57 ivy_class;
std::ofstream __ivy_out;
std::ofstream __ivy_modelfile;
void __ivy_exit(int code){exit(code);}
#include "ivy_threads.hpp"
#include "chacha8c.hpp"
chacha8c::ChaCha8 __chacha8c_rng; // global pseudo-random number generator

std::vector<reader *> threads;
std::vector<reader *> readers;
std::vector<timer *> timers;
bool initializing = false;

void issue57::install_reader(reader *r) {
    readers.push_back(r);
    if (!::initializing)
        r->bind();
}

void issue57::install_thread(reader *r) {
    #ifdef _WIN32

        DWORD dummy;
        HANDLE h = CreateThread( 
            NULL,                   // default security attributes
            0,                      // use default stack size  
            ReaderThreadFunction,   // thread function name
            r,                      // argument to thread function 
            0,                      // use default creation flags 
            &dummy);                // returns the thread identifier 
        if (h == NULL) {
            std::cerr << "failed to create thread" << std::endl;
            exit(1);
        }
        thread_ids.push_back(h);
    #else
        pthread_t thread;
        int res = pthread_create(&thread, NULL, _thread_reader, r);
        if (res) {
            std::cerr << "failed to create thread" << std::endl;
            exit(1);
        }
        thread_ids.push_back(thread);
    #endif
}      

void issue57::install_timer(timer *r) {
    timers.push_back(r);
}

#ifdef _WIN32
    void issue57::__lock() { WaitForSingleObject(mutex,INFINITE); }
    void issue57::__unlock() { ReleaseMutex(mutex); }
#else
    void issue57::__lock() { pthread_mutex_lock(&mutex); }
    void issue57::__unlock() { pthread_mutex_unlock(&mutex); }
#endif

#include <string>
#include <vector>
#include <sstream>
#include <cstdlib>
#include "ivy_z3_gen.hpp"

using namespace hash_space;

typedef ivy_z3_gen<issue57, true> gen;
#include "ivy_value.hpp"
#include "ivy_repl.hpp"
#include "ivy_z3_helpers.hpp"
std::ostream &operator <<(std::ostream &s, const issue57::my_type_2 &t);
template <>
issue57::my_type_2 _arg<issue57::my_type_2>(std::vector<ivy_value> &args, unsigned idx, long long bound);
template <>
void  __ser<issue57::my_type_2>(ivy_ser &res, const issue57::my_type_2&);
template <>
void  __deser<issue57::my_type_2>(ivy_deser &inp, issue57::my_type_2 &res);
template <>
void __from_solver<issue57::my_type_2>( gen &g, const  z3::expr &v, issue57::my_type_2 &res);
template <>
z3::expr __to_solver<issue57::my_type_2>( gen &g, const  z3::expr &v, issue57::my_type_2 &val);
template <>
void __randomize<issue57::my_type_2>( gen &g, const  z3::expr &v, const std::string &sort_name);
bool operator==(const issue57::__tup__unsigned__my_type_2__unsigned &x, const issue57::__tup__unsigned__my_type_2__unsigned &y) {
    return x.arg0 == y.arg0 && x.arg1 == y.arg1 && x.arg2 == y.arg2;
}

int issue57::___ivy_choose(int rng,const char *name,int id) {
        std::ostringstream ss;
        ss << name << ':' << id;;
        for (unsigned i = 0; i < ___ivy_stack.size(); i++)
            ss << ':' << ___ivy_stack[i];
        return ___ivy_gen->choose(rng,ss.str().c_str());
    }
void issue57::__init(){
    ivy_assume(node__voted[issue57::__tup__unsigned__my_type_2__unsigned(prm__V0, MY_TYPE_2, MY_TYPE_3)], "issue57.ivy: line 14");
}
void issue57::__tick(int __timeout){
}
issue57::issue57(){
#ifdef _WIN32
mutex = CreateMutex(NULL,FALSE,NULL);
#else
pthread_mutex_init(&mutex,NULL);
#endif
__lock();
    __CARD__my_type_1 = 4;
    __CARD__my_type_3 = 256;
}
issue57::~issue57(){
    __lock(); // otherwise, thread may die holding lock!
    for (unsigned i = 0; i < thread_ids.size(); i++){
#ifdef _WIN32
       // No idea how to cancel a thread on Windows. We just suspend it
       // so it can't cause any harm as we destruct this object.
       SuspendThread(thread_ids[i]);
#else
        pthread_cancel(thread_ids[i]);
        pthread_join(thread_ids[i],NULL);
#endif
    }
    __unlock();
}

class init_gen : public gen {
public:
    init_gen(issue57&);
    bool generate(issue57&);
    void execute(issue57&){}
};
init_gen::init_gen(issue57 &obj){
    mk_bv("my_type_1",2);
    const char *my_type_2_values[2] = {"val1","val2"};
    mk_enum("my_type_2",2,my_type_2_values);
    mk_bv("my_type_3",8);
    const char *__tmp0_domain[3] = {"my_type_1","my_type_2","my_type_3"};
    mk_decl("node.voted",3,__tmp0_domain,"Bool");
    mk_const("_generating","Bool");
    add("(assert (and\
      true\
    ))");
}
bool init_gen::generate(issue57& obj) {
    alits.clear();
    struct __thunk__0 : z3_thunk<issue57::__tup__unsigned__my_type_2__unsigned,bool> {
        int __ident;
        __thunk__0()  {
            __ident = z3_thunk_counter;
            z3_thunk_counter++;
        }
        bool operator()(const issue57::__tup__unsigned__my_type_2__unsigned &arg) {
            bool __tmp1;
            __tmp1 = (bool)___ivy_choose(0, "node.voted", 0);
            return __tmp1;
        }
        z3::expr to_z3(gen &g, const z3::expr &v) {
            bool __tmp2;
            __tmp2 = (bool)___ivy_choose(0, "node.voted", 0);
            z3::expr res = v == g.ctx.bool_val(__tmp2);
            return res;
        }
    };
    obj.node__voted = hash_thunk<issue57::__tup__unsigned__my_type_2__unsigned,bool>(new __thunk__0());
    obj._generating = (bool)(__chacha8c_rng.Rand() % ((2)-(0)) + (0));

    // std::cout << slvr << std::endl;
    bool __res = solve();
    if (__res) {

    }

    obj.___ivy_gen = this;
    obj.__init();
    return __res;
}
std::ostream &operator <<(std::ostream &s, const issue57::my_type_2 &t){
    if (t == issue57::val1) s<<"val1";
    if (t == issue57::val2) s<<"val2";
    return s;
}
template <>
void  __ser<issue57::my_type_2>(ivy_ser &res, const issue57::my_type_2&t){
    __ser(res,(int)t);
}


    class issue57_repl : public issue57 {

    public:

    virtual void ivy_assert(bool truth,const char *msg){
        if (!truth) {
            __ivy_out << "assertion_failed(\"" << msg << "\")" << std::endl;
            std::cerr << msg << ": error: assertion failed\n";
            
            __ivy_exit(1);
        }
    }
    virtual void ivy_assume(bool truth,const char *msg){
        if (!truth) {
            __ivy_out << "assumption_failed(\"" << msg << "\")" << std::endl;
            std::cerr << msg << ": error: assumption failed\n";
            
            __ivy_exit(1);
        }
    }
    issue57_repl() : issue57(){}

    };
template<typename R> class to_solver_class<hash_thunk<issue57::__tup__unsigned__my_type_2__unsigned,R> > {
    public:;
    z3::expr operator()(gen &g, const z3::expr &v, hash_thunk<issue57::__tup__unsigned__my_type_2__unsigned,R> &val) {
        z3::expr res = g.ctx.bool_val(true);
        z3::expr disj = g.ctx.bool_val(false);
        z3::expr bg = val.fun ? dynamic_cast<z3_thunk<issue57::__tup__unsigned__my_type_2__unsigned,R> *>(val.fun)->to_z3(g, v) : g.ctx.bool_val(true);
        for (typename hash_map<issue57::__tup__unsigned__my_type_2__unsigned,R>::iterator it = val.memo.begin(), en = val.memo.end(); it != en; it++) {
            z3::expr asgn = __to_solver(g, v, it->second);
            auto __key = it->first;
            z3::expr cond = __to_solver(g, v.arg(0), __key.arg0) && __to_solver(g, v.arg(1), __key.arg1) && __to_solver(g, v.arg(2), __key.arg2);
            res = res && implies(cond, asgn);
            disj = disj || cond;
        }
        res = res && (disj || bg);
        return res;
    }
};
template <>
issue57::my_type_2 _arg<issue57::my_type_2>(std::vector<ivy_value> &args, unsigned idx, long long bound){
    ivy_value &arg = args[idx];
    if (arg.atom.size() == 0 || arg.fields.size() != 0) throw out_of_bounds(idx,arg.pos);
    if(arg.atom == "val1") return issue57::val1;
    if(arg.atom == "val2") return issue57::val2;
    throw out_of_bounds("bad value: " + arg.atom,arg.pos);
}
template <>
void __deser<issue57::my_type_2>(ivy_deser &inp, issue57::my_type_2 &res){
    int __res;
    __deser(inp,__res);
    res = (issue57::my_type_2)__res;
}
template <>
z3::expr  __to_solver<issue57::my_type_2>( gen &g, const  z3::expr &v,issue57::my_type_2 &val){
    int thing = val;
    return __to_solver<int>(g,v,thing);
}
template <>
void  __from_solver<issue57::my_type_2>( gen &g, const  z3::expr &v,issue57::my_type_2 &res){
    int temp;
    __from_solver<int>(g,v,temp);
    res = (issue57::my_type_2)temp;
}
template <>
void  __randomize<issue57::my_type_2>( gen &g, const  z3::expr &v, const std::string &sort_name){
    __randomize<int>(g,v,sort_name);
}


class cmd_reader: public stdin_reader {
    int lineno;
public:
    issue57_repl &ivy;    

    cmd_reader(issue57_repl &_ivy) : ivy(_ivy) {
        lineno = 1;
        if (isatty(fdes()))
            __ivy_out << "> "; __ivy_out.flush();
    }

    virtual void process(const std::string &cmd) {
        std::string action;
        std::vector<ivy_value> args;
        try {
            parse_command(cmd,action,args);
            ivy.__lock();

            {
                std::cerr << "undefined action: " << action << std::endl;
            }
            ivy.__unlock();
        }
        catch (syntax_error& err) {
            ivy.__unlock();
            std::cerr << "line " << lineno << ":" << err.pos << ": syntax error" << std::endl;
        }
        catch (out_of_bounds &err) {
            ivy.__unlock();
            std::cerr << "line " << lineno << ":" << err.pos << ": " << err.txt << " bad value" << std::endl;
        }
        catch (bad_arity &err) {
            ivy.__unlock();
            std::cerr << "action " << err.action << " takes " << err.num  << " input parameters" << std::endl;
        }
        if (isatty(fdes()))
            __ivy_out << "> "; __ivy_out.flush();
        lineno++;
    }
};



int main(int argc, char **argv){
        int test_iters = 100;
        int runs = 1;

    int seed = 1;
    std::uint8_t seed32[chacha8c::key_size] = {0};
    std::memcpy(seed32, &seed, sizeof(seed));
    __chacha8c_rng.Seed(seed32);

    int sleep_ms = 10;
    int final_ms = 0; 
    
    std::vector<char *> pargs; // positional args
    pargs.push_back(argv[0]);
    for (int i = 1; i < argc; i++) {
        std::string arg = argv[i];
        size_t p = arg.find('=');
        if (p == std::string::npos)
            pargs.push_back(argv[i]);
        else {
            std::string param = arg.substr(0,p);
            std::string value = arg.substr(p+1);

            if (param == "out") {
                __ivy_out.open(value.c_str());
                if (!__ivy_out) {
                    std::cerr << "cannot open to write: " << value << std::endl;
                    return 1;
                }
            }
            else if (param == "iters") {
                test_iters = atoi(value.c_str());
            }
            else if (param == "runs") {
                runs = atoi(value.c_str());
            }
            else if (param == "seed") {
                seed = atoi(value.c_str());
            }
            else if (param == "delay") {
                sleep_ms = atoi(value.c_str());
            }
            else if (param == "wait") {
                final_ms = atoi(value.c_str());
            }
            else if (param == "modelfile") {
                __ivy_modelfile.open(value.c_str());
                if (!__ivy_modelfile) {
                    std::cerr << "cannot open to write: " << value << std::endl;
                    return 1;
                }
            }
            else {
                std::cerr << "unknown option: " << param << std::endl;
                return 1;
            }
        }
    }
    srand(seed);
    std::memcpy(seed32, &seed, sizeof(seed));
    __chacha8c_rng.Seed(seed32);

    if (!__ivy_out.is_open())
        __ivy_out.basic_ios<char>::rdbuf(std::cout.rdbuf());
    argc = pargs.size();
    argv = &pargs[0];
    if (argc == 2){
        argc--;
        int fd = _open(argv[argc],0);
        if (fd < 0){
            std::cerr << "cannot open to read: " << argv[argc] << "\n";
            __ivy_exit(1);
        }
        _dup2(fd, 0);
    }
    if (argc != 1){
        std::cerr << "usage: issue57 \n";
        __ivy_exit(1);
    }
    std::vector<std::string> args;
    std::vector<ivy_value> arg_values(0);
    for(int i = 1; i < argc;i++){args.push_back(argv[i]);}

#ifdef _WIN32
    // Boilerplate from windows docs

    {
        WORD wVersionRequested;
        WSADATA wsaData;
        int err;

    /* Use the MAKEWORD(lowbyte, highbyte) macro declared in Windef.h */
        wVersionRequested = MAKEWORD(2, 2);

        err = WSAStartup(wVersionRequested, &wsaData);
        if (err != 0) {
            /* Tell the user that we could not find a usable */
            /* Winsock DLL.                                  */
            printf("WSAStartup failed with error: %d\n", err);
            return 1;
        }

    /* Confirm that the WinSock DLL supports 2.2.*/
    /* Note that if the DLL supports versions greater    */
    /* than 2.2 in addition to 2.2, it will still return */
    /* 2.2 in wVersion since that is the version we      */
    /* requested.                                        */

        if (LOBYTE(wsaData.wVersion) != 2 || HIBYTE(wsaData.wVersion) != 2) {
            /* Tell the user that we could not find a usable */
            /* WinSock DLL.                                  */
            printf("Could not find a usable version of Winsock.dll\n");
            WSACleanup();
            return 1;
        }
    }
#endif
    for(int runidx = 0; runidx < runs; runidx++) {
    initializing = true;
    issue57_repl ivy;
    for(unsigned i = 0; i < argc; i++) {ivy.__argv.push_back(argv[i]);}
    ivy._generating = false;

        ivy.__unlock();
        initializing = false;
        for(int rdridx = 0; rdridx < readers.size(); rdridx++) {
            readers[rdridx]->bind();
        }
                    
        init_gen my_init_gen(ivy);
        my_init_gen.generate(ivy);
        std::vector<gen *> generators;
        std::vector<double> weights;

        double totalweight = 0.0;
        int num_gens = 0;


#ifdef _WIN32
    LARGE_INTEGER freq;
    QueryPerformanceFrequency(&freq);
#endif
    double frnd = 0.0;
    bool do_over = false;
    if (num_gens > 0) {
    for(int cycle = 0; cycle < test_iters; cycle++) {

//        std::cout << "totalweight = " << totalweight << std::endl;
//        double choices = totalweight + readers.size() + timers.size();
        double choices = totalweight + 5.0;
        if (do_over) {
           do_over = false;
        }  else {
            frnd = choices * (((double)__chacha8c_rng.Rand())/(((double)RAND_MAX)+1.0));
        }
        // std::cout << "frnd = " << frnd << std::endl;
        if (frnd < totalweight) {
            int idx = 0;
            double sum = 0.0;
            while (idx < num_gens-1) {
                sum += weights[idx];
                if (frnd < sum)
                    break;
                idx++;
            }
            gen &g = *generators[idx];
            ivy.__lock();
#ifdef _WIN32
            LARGE_INTEGER before;
            QueryPerformanceCounter(&before);
#endif
            ivy._generating = true;
            bool sat = g.generate(ivy);
#ifdef _WIN32
            LARGE_INTEGER after;
            QueryPerformanceCounter(&after);
//            __ivy_out << "idx: " << idx << " sat: " << sat << " time: " << (((double)(after.QuadPart-before.QuadPart))/freq.QuadPart) << std::endl;
#endif
            if (sat){
                g.execute(ivy);
                ivy._generating = false;
                ivy.__unlock();
#ifdef _WIN32
                Sleep(sleep_ms);
#endif
            }
            else {
                ivy._generating = false;
                ivy.__unlock();
                cycle--;
            }
            continue;
        }


        fd_set rdfds;
        FD_ZERO(&rdfds);
        int maxfds = 0;

        for (unsigned i = 0; i < readers.size(); i++) {
            reader *r = readers[i];
            int fds = r->fdes();
            if (fds >= 0) {
                FD_SET(fds,&rdfds);
            }
            if (fds > maxfds)
                maxfds = fds;
        }

#ifdef _WIN32
        int timer_min = 15;
#else
        int timer_min = 5;
#endif

        struct timeval timeout;
        timeout.tv_sec = timer_min/1000;
        timeout.tv_usec = 1000 * (timer_min % 1000);

#ifdef _WIN32
        int foo;
        if (readers.size() == 0){  // winsock can't handle empty fdset!
            Sleep(timer_min);
            foo = 0;
        }
        else
            foo = select(maxfds+1,&rdfds,0,0,&timeout);
#else
        int foo = select(maxfds+1,&rdfds,0,0,&timeout);
#endif

        if (foo < 0)
#ifdef _WIN32
            {std::cerr << "select failed: " << WSAGetLastError() << std::endl; __ivy_exit(1);}
#else
            {perror("select failed"); __ivy_exit(1);}
#endif
        
        if (foo == 0){
           // std::cout << "TIMEOUT\n";            
           cycle--;
           for (unsigned i = 0; i < timers.size(); i++){
               if (timer_min >= timers[i]->ms_delay()) {
                   cycle++;
                   break;
               }
           }
           for (unsigned i = 0; i < timers.size(); i++)
               timers[i]->timeout(timer_min);
        }
        else {
            int fdc = 0;
            for (unsigned i = 0; i < readers.size(); i++) {
                reader *r = readers[i];
                if (FD_ISSET(r->fdes(),&rdfds))
                    fdc++;
            }
            // std::cout << "fdc = " << fdc << std::endl;
            int fdi = fdc * (((double)__chacha8c_rng.Rand())/(((double)RAND_MAX)+1.0));
            fdc = 0;
            for (unsigned i = 0; i < readers.size(); i++) {
                reader *r = readers[i];
                if (FD_ISSET(r->fdes(),&rdfds)) {
                    if (fdc == fdi) {
                        // std::cout << "reader = " << i << std::endl;
                        r->read();
                        if (r->background()) {
                           cycle--;
                           do_over = true;
                        }
                        break;
                    }
                    fdc++;

                }
            }
        }
    }
    } // end if (num_gens > 0) — empty fixtures fall through to test_completed

#ifdef _WIN32
                Sleep(final_ms);  // HACK: wait for late responses
#endif
    __ivy_out << "test_completed" << std::endl;
    if (runidx == runs-1) {
        struct timespec ts;
        int ms = 50;
        ts.tv_sec = ms/1000;
        ts.tv_nsec = (ms % 1000) * 1000000;
        nanosleep(&ts,NULL);
        exit(0);
    }
    for (unsigned i = 0; i < readers.size(); i++)
        delete readers[i];
    readers.clear();
    for (unsigned i = 0; i < timers.size(); i++)
        delete timers[i];
    timers.clear();


    }
    return 0;
}
