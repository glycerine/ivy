from ivy import ivy_actions as ia
from ivy import ivy_cpp
from ivy import ivy_isolate as iso
from ivy import ivy_module as im
from ivy import ivy_solver as slv
from ivy import ivy_to_cpp as i2c
from ivy import ivy_utils as iu
from ivy.ivy_compiler import ivy_from_string


prog = """#lang ivy1.7
type key
type val = {a,b}
function f(K:key) : val
function g(K:key) : val
after init { f(K) := g(K) }
"""

ia.set_determinize(True)
slv.set_use_native_enums(True)
iso.set_interpret_all_sorts(True)

with im.Module():
    iu.set_parameters({'target':'test'})
    ivy_from_string(prog, create_isolate=False)
    with ivy_cpp.CppContext():
        header, impl = i2c.module_to_cpp_class('thunk_res_idx',
                                               'thunk_res_idx')

assert 'z3::expr to_z3(gen &g, const z3::expr &v)' in impl
assert '__thunk__0_res_0' in impl
