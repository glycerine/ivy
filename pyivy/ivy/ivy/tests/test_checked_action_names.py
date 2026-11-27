import unittest
from unittest.mock import patch

from ivy import ivy_actions as actions, ivy_ast, ivy_isolate, ivy_module


class CheckedActionNamesTest(unittest.TestCase):
    def test_traces_codegen_initializers_without_rewriting_them(self):
        mod = ivy_module.Module()
        mod.actions = {'step': actions.Sequence(), 'ext:step': actions.Sequence()}
        mod.public_actions = {'ext:step': None}
        mod.initializers = [('init', actions.CallAction(ivy_ast.Atom('ext:step')))]
        codegen_init = actions.Sequence(actions.CallAction(ivy_ast.Atom('step')))
        codegen_initializers = [codegen_init]
        mod.initial_actions = codegen_initializers
        before = actions.call_action_ctr

        with patch.object(ivy_isolate.xtracer, 'enabled', True), \
             patch.object(ivy_isolate.xtracer, 'trace') as trace:
            ivy_isolate.normalize_checked_action_names(mod)

        self.assertEqual(list(mod.initializers[0][1].iter_calls()), ['step'])
        self.assertIs(mod.initial_actions, codegen_initializers)
        self.assertIs(mod.initial_actions[0], codegen_init)
        self.assertEqual(list(codegen_init.iter_calls()), ['step'])
        self.assertEqual(actions.call_action_ctr - before, 1)
        if __debug__:
            prefix = 'isolate.normalizeCheckedActionNames initial_action type='
            visits = [event.args[0][len(prefix):] for event in trace.call_args_list
                      if event.args[0].startswith(prefix)]
            self.assertEqual(visits, ['Sequence', 'CallAction'])


if __name__ == '__main__':
    unittest.main()
