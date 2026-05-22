#pragma once

#include "z3++.h"

#include <cstdint>
#include <cstdlib>
#include <initializer_list>
#include <map>
#include <sstream>
#include <stdexcept>
#include <string>
#include <vector>

#ifndef IVY2CPP_HAS_IVY_GEN
#define IVY2CPP_HAS_IVY_GEN
struct ivy_gen {
    virtual int choose(int rng, const char *name) = 0;
    virtual ~ivy_gen() {}
};
#endif

inline z3::expr exists(const z3::expr &var, const z3::expr &body) {
    Z3_app vars[1] = { var };
    Z3_ast r = Z3_mk_exists_const(body.ctx(), 0, 1, vars, 0, 0, body);
    body.check_error();
    return z3::expr(body.ctx(), r);
}

inline z3::expr forall(const z3::expr &var, const z3::expr &body) {
    Z3_app vars[1] = { var };
    Z3_ast r = Z3_mk_forall_const(body.ctx(), 0, 1, vars, 0, 0, body);
    body.check_error();
    return z3::expr(body.ctx(), r);
}

class gen : public ivy_gen {
public:
    z3::context ctx;
    z3::solver slvr;
    z3::model model;
    std::map<std::string, z3::sort> sorts;
    std::map<std::string, z3::func_decl> decls;
    std::map<std::string, long long> sort_los;
    std::map<std::string, long long> sort_his;
    std::vector<std::string> progress;
    // alits holds assumption literals (Python `gen.alits`). solve() passes
    // these to `slvr.check(alits)` so randomization preferences act as soft
    // constraints. Cleared before each generate() call.
    std::vector<z3::expr> alits;
    unsigned random_counter;

    gen() : slvr(ctx), model(ctx), random_counter(0) {
        sorts.insert(std::make_pair(std::string("bool"), ctx.bool_sort()));
        sorts.insert(std::make_pair(std::string("int"), ctx.int_sort()));
        sort_los[std::string("bool")] = 0;
        sort_his[std::string("bool")] = 1;
        sort_los[std::string("int")] = 0;
        sort_his[std::string("int")] = 4;
    }

    void mk_sort(const char *name) {
        sorts.insert(std::make_pair(std::string(name), ctx.uninterpreted_sort(name)));
        sort_los[std::string(name)] = 0;
        sort_his[std::string(name)] = 4;
    }

    void mk_enum(const char *name, std::initializer_list<const char*> values) {
        mk_sort(name);
        sort_los[std::string(name)] = 0;
        sort_his[std::string(name)] = values.size() == 0 ? 0 : static_cast<long long>(values.size()) - 1;
    }

    void mk_int(const char *name) {
        mk_int(name, 0, 4);
    }

    void mk_int(const char *name, long long lo, long long hi) {
        sorts.insert(std::make_pair(std::string(name), ctx.int_sort()));
        sort_los[std::string(name)] = lo;
        sort_his[std::string(name)] = hi;
    }

    void mk_bv(const char *name, unsigned width) {
        sorts.insert(std::make_pair(std::string(name), ctx.bv_sort(width)));
        sort_los[std::string(name)] = 0;
        sort_his[std::string(name)] = width >= 63 ? 9223372036854775807LL : ((1LL << width) - 1);
    }

    void mk_string(const char *name) {
        sorts.insert(std::make_pair(std::string(name), ctx.string_sort()));
        sort_los[std::string(name)] = 0;
        sort_his[std::string(name)] = 4;
    }

    z3::sort sort(const char *name) const {
        std::map<std::string, z3::sort>::const_iterator it = sorts.find(name);
        if (it == sorts.end()) {
            throw std::runtime_error(std::string("missing z3 sort: ") + name);
        }
        return it->second;
    }

    void mk_decl(const char *name, std::initializer_list<const char*> domain, const char *range) {
        std::vector<z3::sort> domain_sorts;
        for (std::initializer_list<const char*>::const_iterator it = domain.begin(); it != domain.end(); ++it) {
            domain_sorts.push_back(sort(*it));
        }
        z3::sort range_sort = sort(range);
        z3::sort const *domain_ptr = domain_sorts.empty() ? static_cast<z3::sort const *>(0) : &domain_sorts[0];
        z3::func_decl decl = ctx.function(name, static_cast<unsigned>(domain_sorts.size()), domain_ptr, range_sort);
        decls.insert(std::make_pair(std::string(name), decl));
    }

    z3::expr int_to_z3(const char *sort_name, long long value) {
        if (std::string(sort_name) == "bool") {
            return ctx.bool_val(value != 0);
        }
        if (std::string(sort_name) == "int") {
            return ctx.int_val(static_cast<int>(value));
        }
        z3::sort named_sort = sort(sort_name);
        if (named_sort.is_bv()) {
            return ctx.bv_val(static_cast<uint64_t>(value), named_sort.bv_size());
        }
        std::ostringstream ss;
        ss << sort_name << "_" << value;
        return ctx.constant(ss.str().c_str(), named_sort);
    }

    z3::expr int_to_z3(const z3::sort &range, long long value) {
        if (range.is_bool()) {
            return ctx.bool_val(value != 0);
        }
        if (range.is_int()) {
            return ctx.int_val(static_cast<int>(value));
        }
        if (range.is_bv()) {
            return ctx.bv_val(static_cast<uint64_t>(value), range.bv_size());
        }
        std::ostringstream ss;
        ss << range.name() << "_" << value;
        return ctx.constant(ss.str().c_str(), range);
    }

    // apply: variadic helper mirroring Python `gen.apply("name", v, ...)`
    // used by generated destructor __from_solver/__to_solver/__randomize
    // helpers. The C++ z3::func_decl supports variadic operator() with
    // z3::expr arguments, so we forward them through directly.
    template <typename... Args>
    z3::expr apply(const char *decl_name, const Args&... args) {
        std::map<std::string, z3::func_decl>::const_iterator it = decls.find(decl_name);
        if (it == decls.end()) {
            throw std::runtime_error(std::string("missing z3 decl: ") + decl_name);
        }
        return it->second(args...);
    }

    std::string fresh_name() {
        std::ostringstream os;
        os << "__ivy_fresh__" << random_counter++;
        return os.str();
    }

    z3::expr mk_apply_expr(const char *decl_name, const std::vector<int> &args) {
        std::map<std::string, z3::func_decl>::const_iterator it = decls.find(decl_name);
        if (it == decls.end()) {
            throw std::runtime_error(std::string("missing z3 decl: ") + decl_name);
        }
        z3::func_decl decl = it->second;
        if (decl.arity() != args.size()) {
            throw std::runtime_error(std::string("arity mismatch for z3 decl: ") + decl_name);
        }
        std::vector<z3::expr> expr_args;
        for (unsigned i = 0; i < args.size(); i++) {
            expr_args.push_back(int_to_z3(decl.domain(i), args[i]));
        }
        if (expr_args.empty()) {
            return decl();
        }
        return decl(static_cast<unsigned>(expr_args.size()), &expr_args[0]);
    }

    z3::expr mk_apply_expr(const char *decl_name, std::initializer_list<int> args) {
        std::vector<int> args_vec(args.begin(), args.end());
        return mk_apply_expr(decl_name, args_vec);
    }

    void add(const z3::expr &expr) {
        slvr.add(expr);
    }

    // SMT-LIB overload: parse `smtlib` against the previously registered
    // sorts and decls and add the resulting assertion to the solver.
    // Mirrors Python's `add("(assert ...)")` pattern from ivy_to_cpp.py:920
    // and :1277, which feeds slv.formula_to_z3(...).sexpr() to the solver.
    void add(const std::string &smtlib) {
        z3::sort_vector sv(ctx);
        for (std::map<std::string, z3::sort>::const_iterator it = sorts.begin();
             it != sorts.end(); ++it) {
            sv.push_back(it->second);
        }
        z3::func_decl_vector dv(ctx);
        for (std::map<std::string, z3::func_decl>::const_iterator it = decls.begin();
             it != decls.end(); ++it) {
            dv.push_back(it->second);
        }
        slvr.add(ctx.parse_string(smtlib.c_str(), sv, dv));
    }

    void add_alit(const z3::expr &pred) {
        slvr.add(pred);
    }

    void push() {
        slvr.push();
    }

    void pop() {
        slvr.pop();
    }

    bool check() {
        if (slvr.check() == z3::sat) {
            model = slvr.get_model();
            return true;
        }
        return false;
    }

    // solve mirrors Python `gen.solve()`. It calls `slvr.check(alits)` so
    // randomization preferences (added via add_alit) act as assumption
    // literals; on `sat` it captures the model. Used by emit_init_gen /
    // emit_action_gen (ivy_to_cpp.py:959 and :1303).
    bool solve() {
        z3::expr_vector assumptions(ctx);
        for (std::vector<z3::expr>::const_iterator it = alits.begin();
             it != alits.end(); ++it) {
            assumptions.push_back(*it);
        }
        if (slvr.check(assumptions) == z3::sat) {
            model = slvr.get_model();
            return true;
        }
        return false;
    }

    z3::expr eval_expr(const z3::expr &expr) {
        return model.eval(expr, true);
    }

    long long eval(const z3::expr &expr) {
        z3::expr value = eval_expr(expr);
        if (value.is_bool()) {
            return value.bool_value() == Z3_L_TRUE ? 1 : 0;
        }
        int64_t int_value = 0;
        if (value.is_numeral_i64(int_value)) {
            return static_cast<long long>(int_value);
        }
        std::string text = value.to_string();
        std::size_t pos = text.find_last_of('_');
        if (pos != std::string::npos) {
            return std::strtoll(text.substr(pos + 1).c_str(), 0, 10);
        }
        return 0;
    }

    // eval_apply mirrors Python's `eval_apply(name, args...)` used by the
    // scalar branch of emit_eval (ivy_to_cpp.py:794). Builds the function
    // application from the registered decls and reads the integer-typed
    // model value.
    long long eval_apply(const char *decl_name) {
        std::vector<int> args;
        return eval(mk_apply_expr(decl_name, args));
    }
    long long eval_apply(const char *decl_name, int arg0) {
        std::vector<int> args;
        args.push_back(arg0);
        return eval(mk_apply_expr(decl_name, args));
    }
    long long eval_apply(const char *decl_name, std::initializer_list<int> args) {
        std::vector<int> args_vec(args.begin(), args.end());
        return eval(mk_apply_expr(decl_name, args_vec));
    }

    int random_index(int lo, int hi) {
        // Use std::rand() so srand(seed) in main() controls the output
        // sequence. Mirrors Python mk_rand (ivy_to_cpp.py:897-903) which
        // emits `(rand() % (hi-lo+1) + lo)` inline. Cross-binary
        // reproducibility under the same `seed=N` argv requires both
        // sides to draw from the same PRNG.
        int span = hi - lo + 1;
        if (span <= 0) {
            return lo;
        }
        return lo + (std::rand() % span);
    }

    bool random_bool() {
        return random_index(0, 1) != 0;
    }

    int choose(int rng, const char *name) {
        (void)name;
        if (rng <= 0) {
            return 0;
        }
        return random_index(0, rng - 1);
    }

    void randomize(const char *decl_name, const char *range) {
        std::vector<int> args;
        randomize(mk_apply_expr(decl_name, args), range);
    }

    void randomize(const char *decl_name, int arg0, const char *range) {
        std::vector<int> args;
        args.push_back(arg0);
        randomize(mk_apply_expr(decl_name, args), range);
    }

    void randomize(const char *decl_name, std::initializer_list<int> args, const char *range) {
        std::vector<int> args_vec(args.begin(), args.end());
        randomize(mk_apply_expr(decl_name, args_vec), range);
    }

    void randomize(const z3::expr &expr, const std::string &range) {
        long long value = 0;
        if (range == "bool") {
            value = random_bool() ? 1 : 0;
        } else {
            long long lo = 0;
            long long hi = 4;
            std::map<std::string, long long>::const_iterator lo_it = sort_los.find(range);
            std::map<std::string, long long>::const_iterator hi_it = sort_his.find(range);
            if (lo_it != sort_los.end() && hi_it != sort_his.end()) {
                lo = lo_it->second;
                hi = hi_it->second;
            }
            value = random_index(static_cast<int>(lo), static_cast<int>(hi));
        }
        z3::expr pred = expr == int_to_z3(expr.get_sort(), value);
        add_alit(pred);
    }
};

template <class T>
class __random_string_class {
public:
    std::string operator()() {
        std::string res;
        res.push_back('a' + (rand() % 26));
        while (rand() % 2) {
            res.push_back('a' + (rand() % 26));
        }
        return res;
    }
};

template <class T> std::string __random_string() {
    return __random_string_class<T>()();
}

static std::vector<std::string> ivy2cpp_stack;

static void ivy2cpp_progress(const std::string &label) {
    ivy2cpp_stack.push_back(label);
}

static void ivy2cpp_progress(gen &g, const std::string &label) {
    g.progress.push_back(label);
    ivy2cpp_progress(label);
}

// Stubs for Python's `cpptype.prepare()` / `cleanup()` hooks
// (ivy_to_cpp.py:937 and :970). They surround the solver-driven generate
// loop. The Go port currently registers no cpptypes, so these are no-ops;
// keeping them as free functions lets generated code emit the calls
// unconditionally for parity.
static void cpptype_prepare(gen &g) { (void)g; }
static void cpptype_cleanup(gen &g) { (void)g; }

// to_solver_class<T> is the primary template that hash_thunk
// to_solver specializations latch onto. Mirrors
// ivy_z3_helpers.hpp:42-44 in the Python runtime — the goivy port
// emits specializations from ivy2cpp/solver_emit.go (emitHashThunkToSolver,
// emitAllCtuplesToSolver). Empty by design; only specializations supply
// `operator()`.
template <class T> class to_solver_class {};

// z3_thunk<D, R> is the abstract subclass of `thunk<D, R>` that
// supplies a `to_z3` method consumed by the hash_thunk to_solver_class
// specializations. Mirrors ivy_z3_helpers.hpp:135-138. `thunk<D, R>`
// is supplied by the generated header (ivy2cpp/runtime.go
// emitHashThunkSupport), and the impl's include order
// (<basename>.h, then ivy_go_z3.hpp) ensures it is visible here.
template <typename D, typename R>
class z3_thunk : public thunk<D, R> {
public:
    virtual z3::expr to_z3(gen &g, const z3::expr &v) = 0;
};
