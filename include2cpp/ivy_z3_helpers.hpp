#pragma once

#include "ivy_hash.hpp"
#include "z3++.h"

#include <cstdlib>
#include <string>
#include <vector>

template <class T> void __from_solver(gen &g, const z3::expr &v, T &res);

template <>
inline void __from_solver<int>(gen &g, const z3::expr &v, int &res) {
    res = g.eval(v);
}

template <>
inline void __from_solver<long long>(gen &g, const z3::expr &v, long long &res) {
    res = g.eval(v);
}

template <>
inline void __from_solver<unsigned long long>(gen &g, const z3::expr &v, unsigned long long &res) {
    res = g.eval(v);
}

template <>
inline void __from_solver<unsigned>(gen &g, const z3::expr &v, unsigned &res) {
    res = g.eval(v);
}

template <>
inline void __from_solver<bool>(gen &g, const z3::expr &v, bool &res) {
    res = g.eval(v);
}

template <>
inline void __from_solver<__strlit>(gen &g, const z3::expr &v, __strlit &res) {
    res = g.eval_string(v);
}

template <class T>
class to_solver_class {
};

template <class T> z3::expr __to_solver(gen &g, const z3::expr &v, T &val) {
    return to_solver_class<T>()(g, v, val);
}

template <>
inline z3::expr __to_solver<int>(gen &g, const z3::expr &v, int &val) {
    return v == g.int_to_z3(v.get_sort(), val);
}

template <>
inline z3::expr __to_solver<long long>(gen &g, const z3::expr &v, long long &val) {
    return v == g.int_to_z3(v.get_sort(), val);
}

template <>
inline z3::expr __to_solver<unsigned long long>(gen &g, const z3::expr &v, unsigned long long &val) {
    return v == g.int_to_z3(v.get_sort(), val);
}

template <>
inline z3::expr __to_solver<unsigned>(gen &g, const z3::expr &v, unsigned &val) {
    return v == g.int_to_z3(v.get_sort(), val);
}

template <>
inline z3::expr __to_solver<bool>(gen &g, const z3::expr &v, bool &val) {
    return v == g.int_to_z3(v.get_sort(), val);
}

template <>
inline z3::expr __to_solver<__strlit>(gen &g, const z3::expr &v, __strlit &val) {
    return v == g.int_to_z3(v.get_sort(), val);
}

template <class T>
class __random_string_class {
public:
    std::string operator()() {
        std::string res;
        res.push_back('a' + (rand() % 26));
        while (rand() % 2)
            res.push_back('a' + (rand() % 26));
        return res;
    }
};

template <class T> std::string __random_string() {
    return __random_string_class<T>()();
}

template <class T> void __randomize(gen &g, const z3::expr &v, const std::string &sort_name);

template <>
inline void __randomize<int>(gen &g, const z3::expr &v, const std::string &sort_name) {
    g.randomize(v, sort_name);
}

template <>
inline void __randomize<long long>(gen &g, const z3::expr &v, const std::string &sort_name) {
    g.randomize(v, sort_name);
}

template <>
inline void __randomize<unsigned long long>(gen &g, const z3::expr &v, const std::string &sort_name) {
    g.randomize(v, sort_name);
}

template <>
inline void __randomize<unsigned>(gen &g, const z3::expr &v, const std::string &sort_name) {
    g.randomize(v, sort_name);
}

template <>
inline void __randomize<bool>(gen &g, const z3::expr &v, const std::string &sort_name) {
    g.randomize(v, sort_name);
}

template <>
inline void __randomize<__strlit>(gen &g, const z3::expr &apply_expr, const std::string &sort_name) {
    z3::sort range = apply_expr.get_sort();
    __strlit value = (rand() % 2) ? "a" : "b";
    z3::expr val_expr = g.int_to_z3(range, value);
    z3::expr pred = apply_expr == val_expr;
    g.add_alit(pred);
}

static int z3_thunk_counter = 0;

template <typename D, typename R>
class z3_thunk : public thunk<D, R> {
public:
    virtual z3::expr to_z3(gen &g, const z3::expr &v) = 0;
};

inline z3::expr __z3_rename(const z3::expr &e, hash_map<std::string, std::string> &rn) {
    if (e.is_app()) {
        z3::func_decl decl = e.decl();
        z3::expr_vector args(e.ctx());
        unsigned arity = e.num_args();
        for (unsigned i = 0; i < arity; i++) {
            args.push_back(__z3_rename(e.arg(i), rn));
        }
        if (decl.name().kind() == Z3_STRING_SYMBOL) {
            std::string fun = decl.name().str();
            if (rn.find(fun) != rn.end()) {
                std::string newfun = rn[fun];
                std::vector<z3::sort> domain;
                for (unsigned i = 0; i < arity; i++) {
                    domain.push_back(decl.domain(i));
                }
                z3::sort range = e.decl().range();
                decl = e.ctx().function(newfun.c_str(), arity, &domain[0], range);
            }
        }
        return decl(args);
    } else if (e.is_quantifier()) {
        z3::expr body = __z3_rename(e.body(), rn);
        unsigned nb = Z3_get_quantifier_num_bound(e.ctx(), e);
        std::vector<Z3_symbol> bnames;
        std::vector<Z3_sort> bsorts;
        for (unsigned i = 0; i < nb; i++) {
            bnames.push_back(Z3_get_quantifier_bound_name(e.ctx(), e, i));
            bsorts.push_back(Z3_get_quantifier_bound_sort(e.ctx(), e, i));
        }
        Z3_ast q = Z3_mk_quantifier(e.ctx(),
                                    Z3_is_quantifier_forall(e.ctx(), e),
                                    Z3_get_quantifier_weight(e.ctx(), e),
                                    0,
                                    0,
                                    nb,
                                    &bsorts[0],
                                    &bnames[0],
                                    body);
        return z3::expr(e.ctx(), q);
    }
    return e;
}
