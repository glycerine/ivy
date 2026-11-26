#pragma once

#include "z3++.h"

// Z3's SMT-LIB parser returns a formula in the vendored release and an
// assertion vector in newer releases. Both represent a conjunction.
inline z3::expr ivy_z3_assertion_formula(const z3::expr &formula) {
    return formula;
}

inline z3::expr ivy_z3_assertion_formula(const z3::expr_vector &assertions) {
    return z3::mk_and(assertions);
}

inline z3::expr ivy_z3_assertion_formula(z3::context &ctx, Z3_ast formula) {
    return z3::expr(ctx, formula);
}

inline z3::expr ivy_z3_assertion_formula(z3::context &ctx, Z3_ast_vector assertions) {
    return ivy_z3_assertion_formula(z3::expr_vector(ctx, assertions));
}
