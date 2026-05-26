package ivy2cpp

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/glycerine/ivy/goivy"
)

func (g *Generator) runtimeUsesGenerator() bool {
	return g != nil && (g.Config.Target == "gen" || g.Config.Target == "test")
}

func (g *Generator) runtimeUsesReplSubclass() bool {
	return g != nil && (g.Config.Target == "repl" || g.Config.Target == "test")
}

func (g *Generator) hostOS() string {
	if g.Config.HostOS != "" {
		return g.Config.HostOS
	}
	return runtime.GOOS
}

func (g *Generator) emitRuntimeHeaderPreamble(w *cppWriter) {
	if g.hostOS() == "windows" {
		w.line("#define WIN32_LEAN_AND_MEAN")
		w.line("#include <windows.h>")
	}
	w.line("#define _HAS_ITERATOR_DEBUGGING 0")
	if g.runtimeUsesGenerator() {
		if g.Config.Target == "test" {
			w.line("struct ivy_gen {virtual int choose(int rng,const char *name) = 0;};")
		} else {
			w.line("#ifndef IVY2CPP_HAS_IVY_GEN")
			w.line("#define IVY2CPP_HAS_IVY_GEN")
			w.line("struct ivy_gen {virtual int choose(int rng,const char *name) = 0;};")
			w.line("#endif")
		}
	}
	if g.usesZ3() {
		w.line(`#include "z3++.h"`)
	}
	w.line(`#include "ivy_hash.hpp"`)
	if g.usesWideBV() {
		w.line(`#include "ivy_wide_uint.hpp"`)
	}
	w.line("typedef std::string __strlit;")
	w.line("extern std::ofstream __ivy_out;")
	w.line("void __ivy_exit(int);")
	if g.Config.Target == "gen" {
		w.line("extern void ivy_assert(bool, const char *);")
		w.line("extern void ivy_assume(bool, const char *);")
		w.line("extern void ivy_check_progress(int, int);")
		w.line("extern int choose(int, int);")
	}
	w.blank()
	emitHashThunkSupport(w)
	w.blank()
}

func (g *Generator) emitRuntimeHeaderForwardDecls(w *cppWriter) {
	if g.Config.Target == "test" {
		w.blank()
		w.line("    class reader;")
		w.line("    class timer;")
		return
	}
	w.line("class reader;")
	w.line("class timer;")
}

func emitHashThunkSupport(w *cppWriter) {
	w.line("template <typename D, typename R>")
	w.open("struct thunk {")
	w.line("virtual R operator()(const D &) = 0;")
	w.open("int ___ivy_choose(int rng,const char *name,int id) {")
	w.line("return 0;")
	w.close("")
	w.close(";")
	w.line("template <typename D, typename R, class HashFun = hash_space::hash<D> >")
	w.open("struct hash_thunk {")
	w.line("thunk<D,R> *fun;")
	w.line("hash_space::hash_map<D,R,HashFun> memo;")
	w.line("hash_thunk() : fun(0) {}")
	w.line("hash_thunk(thunk<D,R> *fun) : fun(fun) {}")
	w.raw("    ~hash_thunk() {\n//        if (fun)\n//            delete fun;\n    }\n")
	w.open("R &operator[](const D& arg){")
	w.line("std::pair<typename hash_space::hash_map<D,R>::iterator,bool> foo = memo.insert(std::pair<D,R>(arg,R()));")
	w.line("R &res = foo.first->second;")
	w.line("if (foo.second && fun)")
	w.indent++
	w.line("res = (*fun)(arg);")
	w.indent--
	w.line("return res;")
	w.close("")
	w.close(";")
}

func (g *Generator) emitRuntimeClassMembers(w *cppWriter) {
	if g.Config.Target == "test" {
		w.raw(`
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
    virtual ~` + g.ClassName + `();
    std::vector<int> ___ivy_stack;
`)
		if g.runtimeUsesGenerator() {
			w.raw("    ivy_gen *___ivy_gen;\n")
		}
		return
	}
	w.line("std::vector<std::string> __argv;")
	w.line("#ifdef _WIN32")
	w.line("void *mutex;")
	w.line("#else")
	w.line("pthread_mutex_t mutex;")
	w.line("#endif")
	w.line("void __lock();")
	w.line("void __unlock();")
	w.line("#ifdef _WIN32")
	w.line("std::vector<HANDLE> thread_ids;")
	w.line("#else")
	w.line("std::vector<pthread_t> thread_ids;")
	w.line("#endif")
	w.line("void install_reader(reader *);")
	w.line("void install_thread(reader *);")
	w.line("void install_timer(timer *);")
	w.linef("virtual ~%s();", g.ClassName)
	w.line("std::vector<int> ___ivy_stack;")
	if g.runtimeUsesGenerator() {
		w.line("ivy_gen *___ivy_gen;")
	}
}

func (g *Generator) emitRuntimeImplPreamble(w *cppWriter) {
	w.line("#include <sstream>")
	w.line("#include <algorithm>")
	w.blank()
	w.line("#include <iostream>")
	w.line("#include <stdlib.h>")
	w.line("#include <sys/types.h>          /* See NOTES */")
	w.line("#include <sys/stat.h>")
	w.line("#include <fcntl.h>")
	w.line("#ifdef _WIN32")
	w.line("#include <winsock2.h>")
	w.line("#include <WS2tcpip.h>")
	w.line("#include <io.h>")
	w.line("#define isatty _isatty")
	w.line("#else")
	w.line("#include <sys/socket.h>")
	w.line("#include <netinet/in.h>")
	w.line("#include <netinet/ip.h> ")
	w.line("#include <sys/select.h>")
	w.line("#include <unistd.h>")
	w.line("#define _open open")
	w.line("#define _dup2 dup2")
	w.line("#endif")
	w.line("#include <string.h>")
	w.line("#include <stdio.h>")
	w.line("#include <string>")
	w.line("#if __cplusplus < 201103L")
	w.line("#else")
	w.line("#include <cstdint>")
	w.line("#endif")
	w.linef("typedef %s ivy_class;", g.ClassName)
	w.line("std::ofstream __ivy_out;")
	w.line("std::ofstream __ivy_modelfile;")
	w.line("void __ivy_exit(int code){exit(code);}")
	w.line(`#include "ivy_threads.hpp"`)
	w.line(`#include "chacha8c.hpp"`)
	w.line("chacha8c::ChaCha8 __chacha8c_rng; // global pseudo-random number generator")

	w.blank()
}

func (g *Generator) emitRuntimeValueIncludes(w *cppWriter) {
	w.line(`#include "ivy_value.hpp"`)
	if g.Config.Target == "repl" || g.Config.Target == "test" {
		w.line(`#include "ivy_repl.hpp"`)
	}
	if g.Config.Target == "test" {
		w.line(`#include "ivy_z3_helpers.hpp"`)
	} else if g.usesZ3() {
		w.line(`#include "ivy_go_z3.hpp"`)
	}
	if g.Config.Target != "test" {
		w.blank()
	}
	// Forward declarations of per-enum operator<<, _arg<T>, __ser<T>,
	// __deser<T>. Python ivy_to_cpp.py:2213-2223 emits these here.
	g.emitEnumSortArgSpecDecls(w)
	// Forward declarations of per-destructor operator<<, _arg<T>,
	// __ser<T>, __deser<T> (and Z3 specs for test/gen). Python
	// ivy_to_cpp.py:2232-2254.
	g.emitDestructorSortArgSpecDecls(w)
	if g.Config.Target != "test" {
		w.blank()
	}
}

func (g *Generator) emitZ3Boilerplate1(w *cppWriter) {
	w.line("#include <string>")
	w.line("#include <vector>")
	w.line("#include <sstream>")
	w.line("#include <cstdlib>")
	w.line(`#include "ivy_z3_gen.hpp"`)
	w.blank()
	w.line("using namespace hash_space;")
	w.blank()
	modelLog := "false"
	if g.Config.Target == "test" {
		modelLog = "true"
	}
	w.linef("typedef ivy_z3_gen<%s, %s> gen;", g.ClassName, modelLog)
}

func (g *Generator) emitRuntimeConstructorPrelude(w *cppWriter) {
	if g.Config.Target == "test" {
		w.raw(`#ifdef _WIN32
mutex = CreateMutex(NULL,FALSE,NULL);
#else
pthread_mutex_init(&mutex,NULL);
#endif
__lock();
`)
		return
	}
	w.line("#ifdef _WIN32")
	w.line("mutex = CreateMutex(NULL, FALSE, NULL);")
	w.line("#else")
	w.line("pthread_mutex_init(&mutex, NULL);")
	w.line("#endif")
	if g.runtimeUsesGenerator() {
		if g.Config.Target != "test" {
			w.line("___ivy_gen = 0;")
		}
	}
	w.line("__lock();")
}

func (g *Generator) emitRuntimeMethods(w *cppWriter) {
	g.emitRuntimeLockMethods(w)
	g.emitRuntimeInstallMethods(w)
	g.emitRuntimeDestructor(w)
	g.emitRuntimeChoose(w)
}

func (g *Generator) emitRuntimeLockMethods(w *cppWriter) {
	if g.Config.Target == "test" {
		w.raw(`#ifdef _WIN32
    void ` + g.ClassName + `::__lock() { WaitForSingleObject(mutex,INFINITE); }
    void ` + g.ClassName + `::__unlock() { ReleaseMutex(mutex); }
#else
    void ` + g.ClassName + `::__lock() { pthread_mutex_lock(&mutex); }
    void ` + g.ClassName + `::__unlock() { pthread_mutex_unlock(&mutex); }
#endif

`)
		return
	}
	w.line("#ifdef _WIN32")
	w.open(fmt.Sprintf("void %s::__lock() {", g.ClassName))
	w.line("WaitForSingleObject(mutex, INFINITE);")
	w.close("")
	w.open(fmt.Sprintf("void %s::__unlock() {", g.ClassName))
	w.line("ReleaseMutex(mutex);")
	w.close("")
	w.line("#else")
	w.open(fmt.Sprintf("void %s::__lock() {", g.ClassName))
	w.line("pthread_mutex_lock(&mutex);")
	w.close("")
	w.open(fmt.Sprintf("void %s::__unlock() {", g.ClassName))
	w.line("pthread_mutex_unlock(&mutex);")
	w.close("")
	w.line("#endif")
	w.blank()
}

func (g *Generator) emitRuntimeInstallMethods(w *cppWriter) {
	if g.Config.Target == "test" {
		w.line("std::vector<reader *> threads;")
		w.line("std::vector<reader *> readers;")
		w.line("std::vector<timer *> timers;")
		w.line("bool initializing = false;")
		w.blank()
		w.open(fmt.Sprintf("void %s::install_reader(reader *r) {", g.ClassName))
		w.line("readers.push_back(r);")
		w.line("if (!::initializing)")
		w.indent++
		w.line("r->bind();")
		w.indent--
		w.close("")
		w.blank()
		g.emitRuntimeThreadInstallMethod(w, "install_thread", "reader", "ReaderThreadFunction", "_thread_reader")
		w.open(fmt.Sprintf("void %s::install_timer(timer *r) {", g.ClassName))
		w.line("timers.push_back(r);")
		w.close("")
		w.blank()
		return
	}
	w.open(fmt.Sprintf("void %s::install_reader(reader *r) {", g.ClassName))
	w.line("install_thread(r);")
	w.close("")
	w.blank()
	g.emitRuntimeThreadInstallMethod(w, "install_thread", "reader", "ReaderThreadFunction", "_thread_reader")
	g.emitRuntimeThreadInstallMethod(w, "install_timer", "timer", "TimerThreadFunction", "_thread_timer")
}

func (g *Generator) emitRuntimeThreadInstallMethod(w *cppWriter, method, typ, winFn, pthreadFn string) {
	if g.Config.Target == "test" {
		w.raw(fmt.Sprintf(`void %s::%s(%s *r) {
    #ifdef _WIN32

        DWORD dummy;
        HANDLE h = CreateThread( 
            NULL,                   // default security attributes
            0,                      // use default stack size  
            %s,   // thread function name
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
        int res = pthread_create(&thread, NULL, %s, r);
        if (res) {
            std::cerr << "failed to create thread" << std::endl;
            exit(1);
        }
        thread_ids.push_back(thread);
    #endif
}      

`, g.ClassName, method, typ, winFn, pthreadFn))
		return
	}
	w.open(fmt.Sprintf("void %s::%s(%s *r) {", g.ClassName, method, typ))
	w.line("#ifdef _WIN32")
	w.line("DWORD dummy;")
	w.linef("HANDLE h = CreateThread(NULL, 0, %s, r, 0, &dummy);", winFn)
	w.open("if (h == NULL) {")
	w.line(`std::cerr << "failed to create thread" << std::endl;`)
	w.line("exit(1);")
	w.close("")
	w.line("thread_ids.push_back(h);")
	w.line("#else")
	w.line("pthread_t thread;")
	w.linef("int res = pthread_create(&thread, NULL, %s, r);", pthreadFn)
	w.open("if (res) {")
	w.line(`std::cerr << "failed to create thread" << std::endl;`)
	w.line("exit(1);")
	w.close("")
	w.line("thread_ids.push_back(thread);")
	w.line("#endif")
	w.close("")
	w.blank()
}

func (g *Generator) emitRuntimeDestructor(w *cppWriter) {
	if g.Config.Target == "test" {
		w.raw(fmt.Sprintf(`%s::~%s(){
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

`, g.ClassName, g.ClassName))
		return
	}
	w.open(fmt.Sprintf("%s::~%s() {", g.ClassName, g.ClassName))
	w.line("__lock();")
	w.open("for (unsigned i = 0; i < thread_ids.size(); i++) {")
	w.line("#ifdef _WIN32")
	w.line("SuspendThread(thread_ids[i]);")
	w.line("#else")
	w.line("pthread_cancel(thread_ids[i]);")
	w.line("pthread_join(thread_ids[i], NULL);")
	w.line("#endif")
	w.close("")
	w.line("__unlock();")
	w.close("")
	w.blank()
}

func (g *Generator) emitRuntimeChoose(w *cppWriter) {
	if g.Config.Target == "test" {
		w.raw(fmt.Sprintf(`int %s::___ivy_choose(int rng,const char *name,int id) {
        std::ostringstream ss;
        ss << name << ':' << id;;
        for (unsigned i = 0; i < ___ivy_stack.size(); i++)
            ss << ':' << ___ivy_stack[i];
        return ___ivy_gen->choose(rng,ss.str().c_str());
    }
`, g.ClassName))
		return
	}
	w.open(fmt.Sprintf("int %s::___ivy_choose(int rng, const char *name, int id) {", g.ClassName))
	if g.Config.Target == "test" {
		w.line("std::ostringstream ss;")
		w.line("ss << name << ':' << id;;")
		w.line("for (unsigned i = 0; i < ___ivy_stack.size(); i++)")
		w.indent++
		w.line("ss << ':' << ___ivy_stack[i];")
		w.indent--
		w.line("return ___ivy_gen->choose(rng, ss.str().c_str());")
		w.close("")
		w.blank()
		return
	}
	if g.runtimeUsesGenerator() {
		w.line("std::ostringstream ss;")
		w.line("ss << name << ':' << id;")
		w.open("for (unsigned i = 0; i < ___ivy_stack.size(); i++) {")
		w.line("ss << ':' << ___ivy_stack[i];")
		w.close("")
		w.open("if (___ivy_gen) {")
		w.line("return ___ivy_gen->choose(rng, ss.str().c_str());")
		w.close("")
		w.line("return 0;")
	} else {
		w.line("return 0;")
	}
	w.close("")
	w.blank()
}

func (g *Generator) emitRuntimeReplSubclass(w *cppWriter) {
	if g.Config.Target == "test" {
		w.raw("\n\n    class " + g.ClassName + "_repl : public " + g.ClassName + " {\n\n    public:\n\n")
		g.emitRuntimeReplAssertOverride(w, "ivy_assert", "assertion_failed", "assertion failed")
		g.emitRuntimeReplAssertOverride(w, "ivy_assume", "assumption_failed", "assumption failed")
		w.raw("    " + g.replSubclassConstructorSignature() + " : " + g.baseConstructorCall() + "{}\n")
		g.emitReplImportCallbacks(w)
		w.raw("\n    };\n")
		return
	}
	w.open(fmt.Sprintf("class %s_repl : public %s {", g.ClassName, g.ClassName))
	w.line("public:")
	w.indent++
	g.emitRuntimeReplAssertOverride(w, "ivy_assert", "assertion_failed", "assertion failed")
	g.emitRuntimeReplAssertOverride(w, "ivy_assume", "assumption_failed", "assumption failed")
	w.line(g.replSubclassConstructorSignature() + " : " + g.baseConstructorCall() + " {}")
	g.emitReplImportCallbacks(w)
	w.indent--
	w.close(";")
	w.blank()
}

// emitReplImportCallbacks emits a method override on `ClassName_repl`
// for each unscoped imported action whose name matches a known action.
// Mirrors Python emit_repl_boilerplate1 (ivy_to_cpp.py:4107-4128).
//   - For repl target: print `< action(args)` and (if returns) call
//     `ask_ret(__CARD__<sort>)`.
//   - For test target: emit an empty body so the import is callable
//     without prompting the user.
func (g *Generator) emitReplImportCallbacks(w *cppWriter) {
	if g.Mod == nil {
		return
	}
	for _, imp := range g.Mod.Imports {
		impDef, ok := imp.(*goivy.ImportDef)
		if !ok {
			continue
		}
		scope := ""
		if atom, ok := impDef.Scope.(*goivy.Atom); ok {
			scope = atom.Relname()
		}
		if scope != "" {
			continue
		}
		var name string
		if atom, ok := impDef.Imported.(*goivy.Atom); ok {
			name = atom.Relname()
		}
		if name == "" {
			continue
		}
		act, ok := g.Mod.Actions.Get2(name)
		if !ok {
			continue
		}
		g.emitReplImportCallback(w, name, act)
	}
}

func (g *Generator) emitReplImportCallback(w *cppWriter, name string, act goivy.Action) {
	sig := g.methodSignature(name, act, false, false)
	if g.Config.Target == "test" {
		// Test target: empty body so randomized actions can call into
		// the imported entry point without console interaction.
		w.raw("    " + sig + "{}\n")
		return
	}
	w.open(sig + " {")
	// REPL trace line: `< actname(arg1,arg2,...)`.
	display := strings.TrimPrefix(name, "ext:")
	formals := act.GetFormalParams()
	var b strings.Builder
	b.WriteString(fmt.Sprintf(`__ivy_out << "< %s"`, display))
	if len(formals) > 0 {
		b.WriteString(` << "("`)
		for i, p := range formals {
			if i > 0 {
				b.WriteString(` << ","`)
			}
			b.WriteString(fmt.Sprintf(" << %s", varName(p.Name)))
		}
		b.WriteString(` << ")"`)
	}
	b.WriteString(" << std::endl;")
	w.line(b.String())
	// Returns: prompt the user via ask_ret(__CARD__<sort>).
	returns := act.GetFormalReturns()
	if len(returns) > 0 {
		sortText := sortName(returns[0].CSort)
		w.linef("return ask_ret(__CARD__%s);", varName(sortText))
	}
	w.close("")
}

func (g *Generator) emitRuntimeReplAssertOverride(w *cppWriter, method, event, text string) {
	if g.Config.Target == "test" {
		w.raw(fmt.Sprintf(`    virtual void %s(bool truth,const char *msg){
        if (!truth) {
            __ivy_out << "%s(\"" << msg << "\")" << std::endl;
            std::cerr << msg << ": error: %s\n";
            
            __ivy_exit(1);
        }
    }
`, method, event, text))
		return
	}
	w.open(fmt.Sprintf("virtual void %s(bool truth, const char *msg) {", method))
	w.open("if (!truth) {")
	w.linef(`__ivy_out%s << "%s(\"" << msg << "\")" << std::endl;`, g.numberFormat(), event)
	w.linef(`std::cerr << msg << ": error: %s\n";`, text)
	if g.Config.Trace {
		w.linef(`__ivy_out%s << "}" << std::endl;`, g.numberFormat())
	}
	w.line("__ivy_exit(1);")
	w.close("")
	w.close("")
}

func (g *Generator) replSubclassConstructorSignature() string {
	params := make([]string, 0, len(g.Mod.Params))
	for _, p := range g.Mod.Params {
		params = append(params, g.cppStorageDecl(p.Name, p.CSort, g.ClassName))
	}
	return fmt.Sprintf("%s_repl(%s)", g.ClassName, strings.Join(params, ", "))
}

func (g *Generator) baseConstructorCall() string {
	args := make([]string, 0, len(g.Mod.Params))
	for _, p := range g.Mod.Params {
		args = append(args, varName(p.Name))
	}
	return fmt.Sprintf("%s(%s)", g.ClassName, strings.Join(args, ", "))
}

func (g *Generator) runtimeMainClassName() string {
	if g.runtimeUsesReplSubclass() {
		return g.ClassName + "_repl"
	}
	return g.ClassName
}

func (g *Generator) emitRuntimeOutputSetup(w *cppWriter) {
	w.open("if (!__ivy_out.is_open()) {")
	w.line("__ivy_out.basic_ios<char>::rdbuf(std::cout.rdbuf());")
	w.close("")
}

func (g *Generator) emitRuntimeArgCapture(w *cppWriter, obj string) {
	if g.Config.Target == "test" {
		w.open("for (unsigned i = 0; i < argc; i++) {")
		w.linef("%s.__argv.push_back(argv[i]);", obj)
		w.close("")
		return
	}
	w.open("for (int i = 0; i < argc; i++) {")
	w.linef("%s.__argv.push_back(argv[i]);", obj)
	w.close("")
}

func (g *Generator) emitRuntimeBindReaders(w *cppWriter) {
	if g.Config.Target == "test" {
		w.open("for (int rdridx = 0; rdridx < readers.size(); rdridx++) {")
		w.line("readers[rdridx]->bind();")
		w.close("")
		return
	}
	w.open("for (unsigned rdridx = 0; rdridx < readers.size(); rdridx++) {")
	w.line("readers[rdridx]->bind();")
	w.close("")
}
