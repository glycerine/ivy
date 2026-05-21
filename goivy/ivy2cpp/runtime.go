package ivy2cpp

import (
	"fmt"
	"strings"
)

func (g *Generator) runtimeUsesGenerator() bool {
	return g != nil && (g.Config.Target == "gen" || g.Config.Target == "test")
}

func (g *Generator) runtimeUsesReplSubclass() bool {
	return g != nil && (g.Config.Target == "repl" || g.Config.Target == "test")
}

func (g *Generator) emitRuntimeHeaderPreamble(w *cppWriter) {
	w.line("#include <algorithm>")
	w.line("#include <cstdint>")
	w.line("#include <cstdlib>")
	w.line("#include <fstream>")
	w.line("#include <initializer_list>")
	w.line("#include <iostream>")
	w.line("#include <map>")
	w.line("#include <sstream>")
	w.line("#include <stdexcept>")
	w.line("#include <string>")
	w.line("#include <tuple>")
	w.line("#include <vector>")
	w.line(`#include "ivy_hash.hpp"`)
	w.line(`#include "ivy_threads.hpp"`)
	if g.Config.Target == "repl" {
		w.line(`#include "ivy_go_repl.hpp"`)
	}
	if g.usesZ3() {
		w.line("#include <utility>")
		w.line(`#include "z3++.h"`)
	}
	w.blank()
	w.line("typedef std::string __strlit;")
	w.line("extern std::ofstream __ivy_out;")
	w.line("void __ivy_exit(int);")
	if g.Config.Target == "gen" {
		w.line("extern void ivy_assert(bool, const char *);")
		w.line("extern void ivy_assume(bool, const char *);")
		w.line("extern void ivy_check_progress(int, int);")
		w.line("extern int choose(int, int);")
	}
	if g.runtimeUsesGenerator() {
		w.line("#ifndef IVY2CPP_HAS_IVY_GEN")
		w.line("#define IVY2CPP_HAS_IVY_GEN")
		w.line("struct ivy_gen { virtual int choose(int rng, const char *name) = 0; virtual ~ivy_gen() {} };")
		w.line("#endif")
	}
	emitHashThunkSupport(w)
}

func emitHashThunkSupport(w *cppWriter) {
	w.line("template <typename D, typename R>")
	w.open("struct thunk {")
	w.line("virtual R operator()(const D &) = 0;")
	w.line("int ___ivy_choose(int rng, const char *name, int id) { (void)rng; (void)name; (void)id; return 0; }")
	w.line("virtual ~thunk() {}")
	w.close(";")
	w.line("template <typename D, typename R, class HashFun = hash_space::hash<D> >")
	w.open("struct hash_thunk {")
	w.line("thunk<D,R> *fun;")
	w.line("hash_space::hash_map<D,R,HashFun> memo;")
	w.line("hash_thunk() : fun(0) {}")
	w.line("hash_thunk(thunk<D,R> *fun) : fun(fun) {}")
	w.line("~hash_thunk() {}")
	w.open("R &operator[](const D& arg) {")
	w.line("std::pair<typename hash_space::hash_map<D,R,HashFun>::iterator,bool> foo = memo.insert(std::pair<D,R>(arg,R()));")
	w.line("R &res = foo.first->second;")
	w.line("if (foo.second && fun) res = (*fun)(arg);")
	w.line("return res;")
	w.close("")
	w.open("bool operator==(const hash_thunk<D,R,HashFun> &other) const {")
	w.line("return memo == other.memo;")
	w.close("")
	w.close(";")
}

func (g *Generator) emitRuntimeClassMembers(w *cppWriter) {
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
	w.line("#include <algorithm>")
	w.line("#include <fstream>")
	w.line("#include <iostream>")
	w.line("#include <stdlib.h>")
	w.line("#include <sys/types.h>")
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
	w.line("#include <netinet/ip.h>")
	w.line("#include <sys/select.h>")
	w.line("#include <unistd.h>")
	w.line("#define _open open")
	w.line("#define _dup2 dup2")
	w.line("#endif")
	w.line("#include <string.h>")
	w.line("#include <stdio.h>")
	w.line("#include <string>")
	w.line("#include <cstdint>")
	w.line(`#include "ivy_value.hpp"`)
	w.blank()
	w.linef("typedef %s ivy_class;", g.ClassName)
	w.line("std::ofstream __ivy_out;")
	w.line("std::ofstream __ivy_modelfile;")
	w.line("void __ivy_exit(int code) { exit(code); }")
	w.blank()
}

func (g *Generator) emitRuntimeConstructorPrelude(w *cppWriter) {
	w.line("#ifdef _WIN32")
	w.line("mutex = CreateMutex(NULL, FALSE, NULL);")
	w.line("#else")
	w.line("pthread_mutex_init(&mutex, NULL);")
	w.line("#endif")
	if g.runtimeUsesGenerator() {
		w.line("___ivy_gen = 0;")
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
		w.open("if (!::initializing) {")
		w.line("r->bind();")
		w.close("")
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
	w.open(fmt.Sprintf("int %s::___ivy_choose(int rng, const char *name, int id) {", g.ClassName))
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
		w.line("(void)rng;")
		w.line("(void)name;")
		w.line("(void)id;")
		w.line("return 0;")
	}
	w.close("")
	w.blank()
}

func (g *Generator) emitRuntimeReplSubclass(w *cppWriter) {
	w.open(fmt.Sprintf("class %s_repl : public %s {", g.ClassName, g.ClassName))
	w.line("public:")
	w.indent++
	g.emitRuntimeReplAssertOverride(w, "ivy_assert", "assertion_failed", "assertion failed")
	g.emitRuntimeReplAssertOverride(w, "ivy_assume", "assumption_failed", "assumption failed")
	w.line(g.replSubclassConstructorSignature() + " : " + g.baseConstructorCall() + " {}")
	w.indent--
	w.close(";")
	w.blank()
}

func (g *Generator) emitRuntimeReplAssertOverride(w *cppWriter, method, event, text string) {
	w.open(fmt.Sprintf("virtual void %s(bool truth, const char *msg) {", method))
	w.open("if (!truth) {")
	w.linef(`__ivy_out << "%s(\"" << msg << "\")" << std::endl;`, event)
	w.linef(`std::cerr << msg << ": error: %s\n";`, text)
	if g.Config.Trace {
		w.line(`__ivy_out << "}" << std::endl;`)
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
	w.open("for (int i = 0; i < argc; i++) {")
	w.linef("%s.__argv.push_back(argv[i]);", obj)
	w.close("")
}

func (g *Generator) emitRuntimeBindReaders(w *cppWriter) {
	w.open("for (unsigned rdridx = 0; rdridx < readers.size(); rdridx++) {")
	w.line("readers[rdridx]->bind();")
	w.close("")
}
