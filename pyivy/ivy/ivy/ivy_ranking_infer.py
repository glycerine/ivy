#
# Copyright (c) Microsoft Corporation. All Rights Reserved.
#

from . import ivy_actions
from . import ivy_logic
from . import ivy_logic_utils
from . import ivy_module
from . import ivy_proof
from . import ivy_utils
from . import ivy_vmt

def module_for_goal(goal):
    program = ivy_proof.goal_conc(goal).model
    module = ivy_module.module.copy()
    module.isolate_proof = None
    module.labeled_props = []
    module.concept_spaces = []
    module.labeled_conjs = program.invars
    module.postconds = getattr(program,'postconds',{})
    module.public_actions = dict.fromkeys(program.calls,True)
    module.actions = program.binding_map
    module.initializers = [('init',program.init)]
    module.assumed_invariants = program.asms
    module.labeled_axioms = list(module.labeled_axioms)
    module.params = list(module.params)
    module.updates = list(module.updates)
    for premise in ivy_proof.goal_prems(goal):
        if ivy_proof.goal_is_property(premise):
            if any(symbol.name.startswith('work_') for symbol in ivy_logic_utils.symbols_ilu_ast(premise.formula)):
                continue
            if premise.definition:
                module.updates.append(ivy_actions.DerivedUpdate(ivy_logic.drop_universals(premise.formula)))
            module.labeled_axioms.append(premise)
        elif ivy_proof.goal_is_defn(premise):
            symbol = ivy_proof.goal_defines(premise)
            if ivy_logic.is_constant(symbol):
                module.params.append(symbol)
    return module

def infer(goal,sorted_tasks,triggers,tasks):
    if not sorted_tasks:
        raise ivy_utils.IvyError(goal,'ranking_infer requires a work task')
    suffix = sorted_tasks[0]
    start = triggers.get(suffix,{}).get('work_start')
    progress = tasks[suffix]['work_progress']
    trigger = start.rhs() if start is not None else None
    reaction = trigger.args[0] if isinstance(trigger,ivy_logic.Not) else None
    if not isinstance(reaction,ivy_logic.Implies) or not isinstance(reaction.args[1],ivy_logic.Eventually):
        raise ivy_utils.IvyError(goal,'ranking_infer needs a globally reaction of the form P -> eventually Q')

    def predicate(name,definition,body):
        lhs = definition.args[0]
        symbol = ivy_logic.Symbol(name,lhs.rep.sort)
        return ivy_logic.Definition(symbol(*lhs.args),body)

    tags = [('react_p',predicate('.p',start,reaction.args[0])),
            ('react_q',predicate('.q',start,reaction.args[1].body)),
            ('react_r',predicate('.r',progress,progress.args[1]))]
    module = module_for_goal(goal)
    vocab = ivy_proof.goal_vocab(goal)
    with module, ivy_logic.WithSymbols(vocab.symbols), ivy_logic.WithSorts(vocab.sorts):
        with module.theory_context():
            ivy_vmt.check_isolate(tagged_dfns=tags,use_array_encoding=False)
