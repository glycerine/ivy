// Ported to Go from ivy_to_cpp.py REPL and boilerplate (~lines 4700-6000).

// This file covers REPL infrastructure, Z3 solver boilerplate, and
// Windows socket initialisation.
package cppgen

import (
	"fmt"
	"strings"
)

// ---------------------------------------------------------------------------
// emit_repl_imports
// ---------------------------------------------------------------------------

// EmitReplImports emits required import directives for the REPL (no-op).
func EmitReplImports(header, impl *CodeText, classname string) {}

// ---------------------------------------------------------------------------
// emit_repl_boilerplate1
// ---------------------------------------------------------------------------

// EmitReplBoilerplate1 emits the first block of REPL boilerplate.
func EmitReplBoilerplate1(header, impl *CodeText, classname string, optTrace bool) {
	closeTrace := ""
	if optTrace {
		closeTrace = `__ivy_out << "}" << std::endl;`
	}

	impl.Append(`

int ask_ret(long long bound) {
    int res;
    while(true) {
        __ivy_out << "? ";
        std::cin >> res;
        if (res >= 0 && res < bound)
            return res;
        std::cerr << "value out of range" << std::endl;
    }
}

`)

	impl.Append(fmt.Sprintf(`

    class %s_repl : public %s {

    public:

    virtual void ivy_assert(bool truth,const char *msg){
        if (!truth) {
            __ivy_out << "assertion_failed(\"" << msg << "\")" << std::endl;
            std::cerr << msg << ": error: assertion failed\n";
            %s
            __ivy_exit(1);
        }
    }
    virtual void ivy_assume(bool truth,const char *msg){
        if (!truth) {
            __ivy_out << "assumption_failed(\"" << msg << "\")" << std::endl;
            std::cerr << msg << ": error: assumption failed\n";
            %s
            __ivy_exit(1);
        }
    }
`, classname, classname, closeTrace, closeTrace))
}

// ---------------------------------------------------------------------------
// emit_repl_boilerplate1a
// ---------------------------------------------------------------------------

// EmitReplBoilerplate1a emits the stdin_reader and cmd_reader classes.
func EmitReplBoilerplate1a(header, impl *CodeText, classname string) {
	impl.Append(strings.ReplaceAll(`

class stdin_reader: public reader {
    std::string buf;
    std::string eof_flag;

public:
    bool eof(){
      return eof_flag.size();
    }
    virtual int fdes(){
        return 0;
    }
    virtual void read() {
        char tmp[257];
        int chars = ::read(0,tmp,256);
        if (chars == 0) {
            if (buf.size())
                process(buf);
            eof_flag = "eof";
        }
        tmp[chars] = 0;
        buf += std::string(tmp);
        size_t pos;
        while ((pos = buf.find('\n')) != std::string::npos) {
            std::string line = buf.substr(0,pos+1);
            buf.erase(0,pos+1);
            process(line);
        }
    }
    virtual void process(const std::string &line) {
        __ivy_out << line;
    }
};

class cmd_reader: public stdin_reader {
    int lineno;
public:
    classname_repl &ivy;

    cmd_reader(classname_repl &_ivy) : ivy(_ivy) {
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
`, "classname", classname))
}

// ---------------------------------------------------------------------------
// emit_repl_boilerplate2
// ---------------------------------------------------------------------------

// EmitReplBoilerplate2 emits error-handling for the REPL command loop.
func EmitReplBoilerplate2(header, impl *CodeText, classname string) {
	impl.Append(strings.ReplaceAll(`
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

`, "classname", classname))
}

// ---------------------------------------------------------------------------
// emit_boilerplate1 — Z3 solver generator class.
// ---------------------------------------------------------------------------

// EmitBoilerplate1 emits the gen class containing Z3 solver infrastructure.
func EmitBoilerplate1(header, impl *CodeText, classname string) {
	header.Append(`
#include <string>
#include <vector>
#include <sstream>
#include <cstdlib>

using namespace hash_space;

inline z3::expr forall(const std::vector<z3::expr> &exprs, z3::expr const & b) {
    Z3_app *vars = new  Z3_app [exprs.size()];
    std::copy(exprs.begin(),exprs.end(),vars);
    Z3_ast r = Z3_mk_forall_const(b.ctx(), 0, exprs.size(), vars, 0, 0, b);
    b.check_error();
    delete[] vars;
    return z3::expr(b.ctx(), r);
}

`)
	header.Append(strings.ReplaceAll(`
class gen : public ivy_gen {

public:
    z3::context ctx;
    z3::solver slvr;
    z3::model model;

    hash_map<std::string, z3::sort> enum_sorts;
    hash_map<Z3_sort, z3::func_decl_vector> enum_values;
    hash_map<std::string, std::pair<unsigned long long, unsigned long long> > int_ranges;
    hash_map<std::string, z3::func_decl> decls_by_name;
    hash_map<Z3_symbol,int> enum_to_int;
    std::vector<Z3_symbol> sort_names;
    std::vector<Z3_sort> sorts;
    std::vector<Z3_symbol> decl_names;
    std::vector<Z3_func_decl> decls;
    std::vector<z3::expr> alits;
    int tmp_ctr;

    gen(): slvr(ctx), model(ctx,(Z3_model)0) {
        enum_sorts.insert(std::pair<std::string, z3::sort>("bool",ctx.bool_sort()));
        tmp_ctr = 0;
    }

public:
    virtual bool generate(classname& obj)=0;
    virtual void execute(classname& obj)=0;
    virtual ~gen(){}

    bool solve() {
        while(true){
            z3::check_result res = slvr.check(alits.size(),&alits[0]);
            if (res != z3::unsat)
                break;
            z3::expr_vector core = slvr.unsat_core();
            if (core.size() == 0)
                return false;
            unsigned idx = rand() % core.size();
            z3::expr to_delete = core[idx];
            for (unsigned i = 0; i < alits.size(); i++)
                if (z3::eq(alits[i],to_delete)) {
                    alits[i] = alits.back();
                    alits.pop_back();
                    break;
                }
        }
        model = slvr.get_model();
        alits.clear();
        return true;
    }

    int choose(int rng, const char *name){
        if (decls_by_name.find(name) == decls_by_name.end())
            return 0;
        return eval_apply(name);
    }
};
`, "classname", classname))
}

// ---------------------------------------------------------------------------
// emit_winsock_init
// ---------------------------------------------------------------------------

// EmitWinsockInit emits Windows Winsock initialisation boilerplate.
func EmitWinsockInit(impl *CodeText) {
	impl.Append(`
#ifdef _WIN32
    {
        WORD wVersionRequested;
        WSADATA wsaData;
        int err;
        wVersionRequested = MAKEWORD(2, 2);
        err = WSAStartup(wVersionRequested, &wsaData);
        if (err != 0) {
            printf("WSAStartup failed with error: %d\n", err);
            return 1;
        }
        if (LOBYTE(wsaData.wVersion) != 2 || HIBYTE(wsaData.wVersion) != 2) {
            printf("Could not find a usable version of Winsock.dll\n");
            WSACleanup();
            return 1;
        }
    }
#endif
`)
}

// ---------------------------------------------------------------------------
// emit_repl_boilerplate3
// ---------------------------------------------------------------------------

// EmitReplBoilerplate3 emits the REPL console read loop.
func EmitReplBoilerplate3(header, impl *CodeText, classname string) {
	impl.Append(strings.ReplaceAll(`

    ivy.__unlock();

    cmd_reader *cr = new cmd_reader(ivy);

    while (!cr->eof())
        cr->read();
    return 0;

`, "classname", classname))
}

// ---------------------------------------------------------------------------
// EmitParsingUtils
// ---------------------------------------------------------------------------

// EmitParsingUtils emits the parse_value, parse_command, etc. helpers.
func EmitParsingUtils(impl *CodeText, classname string) {
	impl.Append(strings.ReplaceAll(`
bool is_white(int c) {
    return (c == ' ' || c == '\t' || c == '\n' || c == '\r');
}

bool is_ident(int c) {
    return c == '_' || c == '.' || (c >= 'A' &&  c <= 'Z')
        || (c >= 'a' &&  c <= 'z')
        || (c >= '0' &&  c <= '9');
}

void skip_white(const std::string& str, int &pos){
    while (pos < str.size() && is_white(str[pos]))
        pos++;
}

struct syntax_error {
    int pos;
    syntax_error(int pos) : pos(pos) {}
};

void throw_syntax(int pos){
    throw syntax_error(pos);
}

std::string get_ident(const std::string& str, int &pos) {
    std::string res = "";
    while (pos < str.size() && is_ident(str[pos])) {
        res.push_back(str[pos]);
        pos++;
    }
    if (res.size() == 0)
        throw_syntax(pos);
    return res;
}

void parse_command(const std::string &cmd, std::string &action, std::vector<ivy_value> &args) {
    int pos = 0;
    skip_white(cmd,pos);
    action = get_ident(cmd,pos);
    skip_white(cmd,pos);
    if (pos < cmd.size() && cmd[pos] == '(') {
        pos++;
        skip_white(cmd,pos);
        args.push_back(parse_value(cmd,pos));
        while(true) {
            skip_white(cmd,pos);
            if (!(pos < cmd.size() && cmd[pos] == ','))
                break;
            pos++;
            args.push_back(parse_value(cmd,pos));
        }
        if (!(pos < cmd.size() && cmd[pos] == ')'))
            throw_syntax(pos);
        pos++;
    }
    skip_white(cmd,pos);
    if (pos != cmd.size())
        throw_syntax(pos);
}

struct bad_arity {
    std::string action;
    int num;
    bad_arity(std::string &_action, unsigned _num) : action(_action), num(_num) {}
};

void check_arity(std::vector<ivy_value> &args, unsigned num, std::string &action) {
    if (args.size() != num)
        throw bad_arity(action,num);
}
`, "classname", classname))
}
