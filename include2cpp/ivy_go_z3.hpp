#pragma once

#include "z3++.h"

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
        std::ostringstream ss;
        ss << sort_name << "_" << value;
        return ctx.constant(ss.str().c_str(), sort(sort_name));
    }

    z3::expr int_to_z3(const z3::sort &range, long long value) {
        if (range.is_bool()) {
            return ctx.bool_val(value != 0);
        }
        if (range.is_int()) {
            return ctx.int_val(static_cast<int>(value));
        }
        std::ostringstream ss;
        ss << range.name() << "_" << value;
        return ctx.constant(ss.str().c_str(), range);
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

    void add(const z3::expr &expr) {
        slvr.add(expr);
    }

    void add_alit(const z3::expr &pred) {
        slvr.add(pred);
    }

    bool check() {
        if (slvr.check() == z3::sat) {
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

    int random_index(int lo, int hi) {
        int span = hi - lo + 1;
        if (span <= 0) {
            return lo;
        }
        return lo + static_cast<int>((random_counter++) % static_cast<unsigned>(span));
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

static std::vector<std::string> ivy2cpp_stack;

static void ivy2cpp_progress(const std::string &label) {
    ivy2cpp_stack.push_back(label);
}

static void ivy2cpp_progress(gen &g, const std::string &label) {
    g.progress.push_back(label);
    ivy2cpp_progress(label);
}
