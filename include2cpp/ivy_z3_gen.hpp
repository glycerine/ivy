#pragma once

#include "ivy_hash.hpp"
#include "chacha8c.hpp"
#include "z3++.h"

extern chacha8c::ChaCha8 __chacha8c_rng;

#include <algorithm>
#include <cassert>
#include <cstdint>
#include <cstdlib>
#include <fstream>
#include <iostream>
#include <sstream>
#include <string>
#include <utility>
#include <vector>

extern std::ofstream __ivy_modelfile;

using namespace hash_space;

inline z3::expr forall(const std::vector<z3::expr> &exprs, z3::expr const &b) {
    Z3_app *vars = new Z3_app[exprs.size()];
    std::copy(exprs.begin(), exprs.end(), vars);
    Z3_ast r = Z3_mk_forall_const(b.ctx(), 0, exprs.size(), vars, 0, 0, b);
    b.check_error();
    delete[] vars;
    return z3::expr(b.ctx(), r);
}

template <typename IvyClass, bool EmitModelLog>
class ivy_z3_gen : public ivy_gen {

public:
    z3::context ctx;
    z3::solver slvr;
    z3::model model;

    hash_map<std::string, z3::sort> enum_sorts;
    hash_map<Z3_sort, z3::func_decl_vector> enum_values;
    hash_map<std::string, std::pair<unsigned long long, unsigned long long> > int_ranges;
    hash_map<std::string, z3::func_decl> decls_by_name;
    hash_map<Z3_symbol, int> enum_to_int;
    std::vector<Z3_symbol> sort_names;
    std::vector<Z3_sort> sorts;
    std::vector<Z3_symbol> decl_names;
    std::vector<Z3_func_decl> decls;
    std::vector<z3::expr> alits;
    int tmp_ctr;

    ivy_z3_gen() : slvr(ctx), model(ctx, (Z3_model)0) {
        enum_sorts.insert(std::pair<std::string, z3::sort>("bool", ctx.bool_sort()));
        tmp_ctr = 0;
    }

public:
    virtual bool generate(IvyClass &obj) = 0;
    virtual void execute(IvyClass &obj) = 0;
    virtual ~ivy_z3_gen() {}

    std::string fresh_name() {
        std::ostringstream ss;
        ss << "$tmp" << tmp_ctr++;
        return ss.str();
    }

    z3::expr mk_apply_expr(const char *decl_name, unsigned num_args, const int *args) {
        z3::func_decl decl = decls_by_name.find(decl_name)->second;
        std::vector<z3::expr> expr_args;
        unsigned arity = decl.arity();
        assert(arity == num_args);
        for (unsigned i = 0; i < arity; i++) {
            z3::sort sort = decl.domain(i);
            expr_args.push_back(int_to_z3(sort, args[i]));
        }
        return decl(arity, &expr_args[0]);
    }

    z3::expr eval_expr(const z3::expr &apply_expr) {
        return model.eval(apply_expr, true);
    }

    long long eval(const z3::expr &apply_expr) {
        try {
            z3::expr foo = eval_expr(apply_expr);
            if (foo.is_int()) {
                assert(foo.is_numeral());
                int v;
                if (Z3_get_numeral_int(ctx, foo, &v) != Z3_TRUE) {
                    std::cerr << "integer value from Z3 too large for machine int: " << foo << std::endl;
                    assert(false);
                }
                return v;
            }
            if (foo.is_bv()) {
                assert(foo.is_numeral());
                uint64_t v;
                if (Z3_get_numeral_uint64(ctx, foo, &v) != Z3_TRUE) {
                    std::cerr << "bit vector value from Z3 too large for machine uint64: " << foo << std::endl;
                    assert(false);
                }
                return v;
            }
            assert(foo.is_app());
            if (foo.is_bool())
                return (foo.decl().decl_kind() == Z3_OP_TRUE) ? 1 : 0;
            return enum_to_int[foo.decl().name()];
        }
        catch (const z3::exception &e) {
            std::cerr << e << std::endl;
            throw e;
        }
    }

    __strlit eval_string(const z3::expr &apply_expr) {
        try {
            z3::expr foo = model.eval(apply_expr, true);
            assert(Z3_is_string(ctx, foo));
            return Z3_get_string(ctx, foo);
        }
        catch (const z3::exception &e) {
            std::cerr << e << std::endl;
            throw e;
        }
    }

    long long eval_apply(const char *decl_name, unsigned num_args, const int *args) {
        z3::expr apply_expr = mk_apply_expr(decl_name, num_args, args);
        try {
            z3::expr foo = model.eval(apply_expr, true);
            if (foo.is_int()) {
                assert(foo.is_numeral());
                int v;
                if (Z3_get_numeral_int(ctx, foo, &v) != Z3_TRUE) {
                    assert(false && "integer value from Z3 too large for machine int");
                }
                return v;
            }
            if (foo.is_bv()) {
                assert(foo.is_numeral());
                uint64_t v;
                if (Z3_get_numeral_uint64(ctx, foo, &v) != Z3_TRUE) {
                    assert(false && "bit vector value from Z3 too large for machine uint64");
                }
                return v;
            }
            if (foo.is_bv() || foo.is_int()) {
                assert(foo.is_numeral());
                unsigned v;
                if (Z3_get_numeral_uint(ctx, foo, &v) != Z3_TRUE)
                    assert(false && "bit vector value too large for machine int");
                return v;
            }
            assert(foo.is_app());
            if (foo.is_bool())
                return (foo.decl().decl_kind() == Z3_OP_TRUE) ? 1 : 0;
            return enum_to_int[foo.decl().name()];
        }
        catch (const z3::exception &e) {
            std::cerr << e << std::endl;
            throw e;
        }
    }

    long long eval_apply(const char *decl_name) {
        return eval_apply(decl_name, 0, (int *)0);
    }

    long long eval_apply(const char *decl_name, int arg0) {
        return eval_apply(decl_name, 1, &arg0);
    }

    long long eval_apply(const char *decl_name, int arg0, int arg1) {
        int args[2] = {arg0, arg1};
        return eval_apply(decl_name, 2, args);
    }

    long long eval_apply(const char *decl_name, int arg0, int arg1, int arg2) {
        int args[3] = {arg0, arg1, arg2};
        return eval_apply(decl_name, 3, args);
    }

    long long eval_apply(const char *decl_name, int arg0, int arg1, int arg2, int arg3) {
        int args[4] = {arg0, arg1, arg2, arg3};
        return eval_apply(decl_name, 4, args);
    }

    z3::expr apply(const char *decl_name, std::vector<z3::expr> &expr_args) {
        z3::func_decl decl = decls_by_name.find(decl_name)->second;
        unsigned arity = decl.arity();
        assert(arity == expr_args.size());
        return decl(arity, &expr_args[0]);
    }

    z3::expr apply(const char *decl_name) {
        std::vector<z3::expr> a;
        return apply(decl_name, a);
    }

    z3::expr apply(const char *decl_name, z3::expr arg0) {
        std::vector<z3::expr> a;
        a.push_back(arg0);
        return apply(decl_name, a);
    }

    z3::expr apply(const char *decl_name, z3::expr arg0, z3::expr arg1) {
        std::vector<z3::expr> a;
        a.push_back(arg0);
        a.push_back(arg1);
        return apply(decl_name, a);
    }

    z3::expr apply(const char *decl_name, z3::expr arg0, z3::expr arg1, z3::expr arg2) {
        std::vector<z3::expr> a;
        a.push_back(arg0);
        a.push_back(arg1);
        a.push_back(arg2);
        return apply(decl_name, a);
    }

    z3::expr apply(const char *decl_name, z3::expr arg0, z3::expr arg1, z3::expr arg2, z3::expr arg3) {
        std::vector<z3::expr> a;
        a.push_back(arg0);
        a.push_back(arg1);
        a.push_back(arg2);
        a.push_back(arg3);
        return apply(decl_name, a);
    }

    z3::expr apply(const char *decl_name, z3::expr arg0, z3::expr arg1, z3::expr arg2, z3::expr arg3, z3::expr arg4) {
        std::vector<z3::expr> a;
        a.push_back(arg0);
        a.push_back(arg1);
        a.push_back(arg2);
        a.push_back(arg3);
        a.push_back(arg4);
        return apply(decl_name, a);
    }

    z3::expr int_to_z3(const z3::sort &range, int64_t value) {
        if (range.is_bool())
            return ctx.bool_val((bool)value);
        if (range.is_bv())
            return ctx.bv_val((int)value, range.bv_size());
        if (range.is_int())
            return ctx.int_val((int)value);
        return enum_values.find(range)->second[(int)value]();
    }

    z3::expr int_to_z3(const z3::sort &range, const std::string &value) {
        if (range.to_string() == "String")
            return ctx.string_val(value);
        if (range.is_bv())
            return ctx.bv_val(value.c_str(), range.bv_size());
        return ctx.constant(value.c_str(), range);
    }

    std::string eval_numeral_string(const z3::expr &expr) {
        z3::expr value = eval_expr(expr);
        if (value.is_bool())
            return value.bool_value() == Z3_L_TRUE ? "1" : "0";
        std::string text;
        if (value.is_numeral(text))
            return text;
        return value.to_string();
    }

    std::pair<unsigned long long, unsigned long long> sort_range(const z3::sort &range, const std::string &sort_name) {
        std::pair<unsigned long long, unsigned long long> res;
        res.first = 0;
        if (range.is_bool())
            res.second = 1;
        else if (range.is_bv()) {
            int size = range.bv_size();
            if (size >= 64)
                res.second = (unsigned long long)(-1);
            else res.second = (1 << size) - 1;
        }
        else if (range.is_int()) {
            if (int_ranges.find(sort_name) != int_ranges.end())
                res = int_ranges[sort_name];
            else res.second = 4;
        }
        else res.second = enum_values.find(range)->second.size() - 1;
        return res;
    }

    int set(const char *decl_name, unsigned num_args, const int *args, int value) {
        z3::func_decl decl = decls_by_name.find(decl_name)->second;
        std::vector<z3::expr> expr_args;
        unsigned arity = decl.arity();
        assert(arity == num_args);
        for (unsigned i = 0; i < arity; i++) {
            z3::sort sort = decl.domain(i);
            expr_args.push_back(int_to_z3(sort, args[i]));
        }
        z3::expr apply_expr = decl(arity, &expr_args[0]);
        z3::sort range = decl.range();
        z3::expr val_expr = int_to_z3(range, value);
        z3::expr pred = apply_expr == val_expr;
        slvr.add(pred);
        return 0;
    }

    int set(const char *decl_name, int value) {
        return set(decl_name, 0, (int *)0, value);
    }

    int set(const char *decl_name, int arg0, int value) {
        return set(decl_name, 1, &arg0, value);
    }

    int set(const char *decl_name, int arg0, int arg1, int value) {
        int args[2] = {arg0, arg1};
        return set(decl_name, 2, args, value);
    }

    int set(const char *decl_name, int arg0, int arg1, int arg2, int value) {
        int args[3] = {arg0, arg1, arg2};
        return set(decl_name, 3, args, value);
    }

    void add_alit(const z3::expr &pred) {
        if (__ivy_modelfile.is_open())
            __ivy_modelfile << "pred: " << pred << std::endl;
        std::ostringstream ss;
        ss << "alit:" << alits.size();
        z3::expr alit = ctx.bool_const(ss.str().c_str());
        if (__ivy_modelfile.is_open())
            __ivy_modelfile << "alit: " << alit << std::endl;
        alits.push_back(alit);
        slvr.add(!alit || pred);
    }

    unsigned long long random_range(std::pair<unsigned long long, unsigned long long> rng) {
        unsigned long long res = 0;
        // chacha8c: single Uint64 produces the same 64-bit width as the
        // historical 4 × (rand() & 0xffff). Routing through the
        // chacha8c PRNG keeps action-input randomization byte-equivalent
        // across ivy_to_cpp (Python), ivy2cpp (Go-emitted C++), and
        // ivy2go (Go-emitted Go).
        res = __chacha8c_rng.Uint64();
        unsigned long long card = rng.second - rng.first;
        if (card != (unsigned long long)(-1))
            res = (res % (card + 1)) + rng.first;
        return res;
    }

    void randomize(const z3::expr &apply_expr, const std::string &sort_name) {
        z3::sort range = apply_expr.get_sort();
        unsigned long long value = random_range(sort_range(range, sort_name));
        z3::expr val_expr = int_to_z3(range, value);
        z3::expr pred = apply_expr == val_expr;
        add_alit(pred);
    }

    void randomize(const char *decl_name, unsigned num_args, const int *args, const std::string &sort_name) {
        z3::func_decl decl = decls_by_name.find(decl_name)->second;
        z3::expr apply_expr = mk_apply_expr(decl_name, num_args, args);
        z3::sort range = decl.range();
        unsigned long long value = random_range(sort_range(range, sort_name));
        z3::expr val_expr = int_to_z3(range, value);
        z3::expr pred = apply_expr == val_expr;
        add_alit(pred);
    }

    void randomize(const char *decl_name, const std::string &sort_name) {
        randomize(decl_name, 0, (int *)0, sort_name);
    }

    void randomize(const char *decl_name, int arg0, const std::string &sort_name) {
        randomize(decl_name, 1, &arg0, sort_name);
    }

    void randomize(const char *decl_name, int arg0, int arg1, const std::string &sort_name) {
        int args[2] = {arg0, arg1};
        randomize(decl_name, 2, args, sort_name);
    }

    void randomize(const char *decl_name, int arg0, int arg1, int arg2, const std::string &sort_name) {
        int args[3] = {arg0, arg1, arg2};
        randomize(decl_name, 3, args, sort_name);
    }

    void push() {
        slvr.push();
    }

    void pop() {
        slvr.pop();
    }

    z3::sort sort(const char *name) {
        if (std::string("bool") == name)
            return ctx.bool_sort();
        return enum_sorts.find(name)->second;
    }

    void mk_enum(const char *sort_name, unsigned num_values, char const * const *value_names) {
        z3::func_decl_vector cs(ctx), ts(ctx);
        z3::sort sort = ctx.enumeration_sort(sort_name, num_values, value_names, cs, ts);
        enum_sorts.insert(std::pair<std::string, z3::sort>(sort_name, sort));
        enum_values.insert(std::pair<Z3_sort, z3::func_decl_vector>(sort, cs));
        sort_names.push_back(Z3_mk_string_symbol(ctx, sort_name));
        sorts.push_back(sort);
        for (unsigned i = 0; i < num_values; i++) {
            Z3_symbol sym = Z3_mk_string_symbol(ctx, value_names[i]);
            decl_names.push_back(sym);
            decls.push_back(cs[i]);
            enum_to_int[sym] = i;
        }
    }

    void mk_bv(const char *sort_name, unsigned width) {
        z3::sort sort = ctx.bv_sort(width);
        enum_sorts.insert(std::pair<std::string, z3::sort>(sort_name, sort));
    }

    void mk_int(const char *sort_name) {
        z3::sort sort = ctx.int_sort();
        enum_sorts.insert(std::pair<std::string, z3::sort>(sort_name, sort));
    }

    void mk_string(const char *sort_name) {
        z3::sort sort = ctx.string_sort();
        enum_sorts.insert(std::pair<std::string, z3::sort>(sort_name, sort));
    }

    void mk_sort(const char *sort_name) {
        Z3_symbol symb = Z3_mk_string_symbol(ctx, sort_name);
        z3::sort sort(ctx, Z3_mk_uninterpreted_sort(ctx, symb));
        enum_sorts.insert(std::pair<std::string, z3::sort>(sort_name, sort));
        sort_names.push_back(symb);
        sorts.push_back(sort);
    }

    void mk_decl(const char *decl_name, unsigned arity, const char **domain_names, const char *range_name) {
        std::vector<z3::sort> domain;
        for (unsigned i = 0; i < arity; i++) {
            if (enum_sorts.find(domain_names[i]) == enum_sorts.end()) {
                std::cout << "unknown sort: " << domain_names[i] << std::endl;
                exit(1);
            }
            domain.push_back(enum_sorts.find(domain_names[i])->second);
        }
        std::string bool_name("Bool");
        z3::sort range = (range_name == bool_name) ? ctx.bool_sort() : enum_sorts.find(range_name)->second;
        z3::func_decl decl = ctx.function(decl_name, arity, &domain[0], range);
        decl_names.push_back(Z3_mk_string_symbol(ctx, decl_name));
        decls.push_back(decl);
        decls_by_name.insert(std::pair<std::string, z3::func_decl>(decl_name, decl));
    }

    void mk_const(const char *const_name, const char *sort_name) {
        mk_decl(const_name, 0, 0, sort_name);
    }

    z3::expr parse_expr(const std::string &z3inp) {
        Z3_symbol *sort_names_ptr = sort_names.empty() ? 0 : &sort_names[0];
        Z3_sort *sorts_ptr = sorts.empty() ? 0 : &sorts[0];
        Z3_symbol *decl_names_ptr = decl_names.empty() ? 0 : &decl_names[0];
        Z3_func_decl *decls_ptr = decls.empty() ? 0 : &decls[0];
        z3::expr fmla(ctx, Z3_parse_smtlib2_string(ctx, z3inp.c_str(), sort_names.size(), sort_names_ptr, sorts_ptr, decl_names.size(), decl_names_ptr, decls_ptr));
        ctx.check_error();
        return fmla;
    }

    void add(const std::string &z3inp) {
        slvr.add(parse_expr(z3inp));
    }

    bool solve() {
        if (__ivy_modelfile.is_open())
            __ivy_modelfile << "begin check:\n" << slvr << "end check:\n" << std::endl;
        while (true) {
            if (__ivy_modelfile.is_open()) {
                __ivy_modelfile << "(check-sat";
                for (unsigned i = 0; i < alits.size(); i++)
                    __ivy_modelfile << " " << alits[i];
                __ivy_modelfile << ")" << std::endl;
            }
            z3::check_result res = slvr.check(alits.size(), &alits[0]);
            if (res != z3::unsat)
                break;
            z3::expr_vector core = slvr.unsat_core();
            if (core.size() == 0) {
                return false;
            }
            if (__ivy_modelfile.is_open())
                for (unsigned i = 0; i < core.size(); i++)
                    __ivy_modelfile << "core: " << core[i] << std::endl;
            unsigned idx = (unsigned)(__chacha8c_rng.Rand()) % core.size();
            z3::expr to_delete = core[idx];
            if (__ivy_modelfile.is_open())
                __ivy_modelfile << "to delete: " << to_delete << std::endl;
            for (unsigned i = 0; i < alits.size(); i++)
                if (z3::eq(alits[i], to_delete)) {
                    alits[i] = alits.back();
                    alits.pop_back();
                    break;
                }
        }
        model = slvr.get_model();
        alits.clear();
        if (EmitModelLog && __ivy_modelfile.is_open()) {
            __ivy_modelfile << "begin sat:\n" << slvr << "end sat:\n" << std::endl;
            __ivy_modelfile << model;
            __ivy_modelfile.flush();
        }
        return true;
    }

    int choose(int rng, const char *name) {
        if (decls_by_name.find(name) == decls_by_name.end())
            return 0;
        return eval_apply(name);
    }
};
