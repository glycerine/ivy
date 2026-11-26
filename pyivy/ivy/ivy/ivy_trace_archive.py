"""Convert trusted Ivy trace files to the portable GoIvy archive format."""

import json

from . import ivy_interp as interp
from . import ivy_logic as il
from . import ivy_logic_utils as ilu
from . import ivy_module as im
from . import ivy_trace
from . import ivy_utils as iu

FORMAT = 'ivy-trace-1'
export_trace = iu.Parameter('export_trace',check=lambda path: path.endswith('.a2g'))


def record_sort(sort,sig,sorts):
    name = 'bool' if il.is_boolean_sort(sort) else sort.name
    if name not in sorts:
        info = {}
        interpretation = sig.interp.get(name)
        if isinstance(interpretation,str):
            info['interpretation'] = interpretation
        if isinstance(interpretation,il.RangeSort):
            sort = interpretation
        if isinstance(sort,il.RangeSort):
            info['bounds'] = [str(sort.lb),str(sort.ub)]
        elif isinstance(sort,il.EnumeratedSort):
            info['constructors'] = list(sort.extension)
        sorts[name] = info
    return name


def record_value(term,sig,sorts):
    if il.is_true(term) or il.is_false(term):
        return {'name':'true' if il.is_true(term) else 'false'}
    if not il.is_app(term) or not isinstance(term.rep,il.Symbol):
        return None
    symbol = term.rep
    arguments = [record_value(arg,sig,sorts) for arg in term.args]
    if any(arg is None for arg in arguments):
        return None
    parts = list(symbol.sort.dom) + [symbol.sort.rng] if il.is_function_sort(symbol.sort) else [symbol.sort]
    value = {'name':symbol.name, 'signature':[record_sort(sort,sig,sorts) for sort in parts]}
    if arguments:
        value['arguments'] = arguments
    return value


def record_graph(graph):
    version = getattr(graph,'ivy_version',iu.get_string_version())
    tracking = {'version':[int(part) for part in version.split('.')[:2]],'sorts':{},'steps':[]}
    archive = {'format':FORMAT,'tracking':tracking}
    detailed = ivy_trace.option_detailed.get()
    hidden = graph.hidden_symbols or ivy_trace.show_all_symbols
    offset = 0

    def visit(state,indent,lines):
        nonlocal offset
        facts = []
        for equation in state.clauses.fmlas:
            if il.is_eq(equation):
                pair = [record_value(value,state.domain.sig,tracking['sorts']) for value in equation.args]
                if all(value is not None for value in pair):
                    facts.append(pair)
        tracking['steps'].append({'prefix':''.join(lines[offset:]),'indent':indent,'facts':facts})
        offset = len(lines)

    try:
        for name in ('detailed','brief'):
            ivy_trace.option_detailed.set('true' if name == 'detailed' else 'false')
            lines = list(getattr(graph,'notes',()))
            graph.to_lines(lines,{},0,hidden)
            archive[name] = ''.join(lines)
        ivy_trace.option_detailed.set('true')
        lines = list(getattr(graph,'notes',()))
        graph.to_lines(lines,{},0,hidden,visit_state=visit)
        tracking['tail'] = ''.join(lines[offset:])
    finally:
        ivy_trace.option_detailed.set('true' if detailed else 'false')
    return archive


def export(graph):
    path = export_trace.get()
    try:
        with open(path,'x',encoding='utf-8') as output:
            json.dump(graph if isinstance(graph,dict) else record_graph(graph),output)
            output.write('\n')
    except OSError as error:
        raise iu.IvyError(None,f'cannot export trace {path}: {error}') from error
    print(f'Wrote portable trace to {path}')


def add_sorts(sorts):
    for name,record in sorts.items():
        if name == 'bool':
            continue
        if record.get('constructors'):
            sort = il.EnumeratedSort(name,record['constructors'])
            for constructor in sort.constructors:
                il.add_symbol(constructor.name,sort)
        else:
            sort = il.UninterpretedSort(name)
        il.add_sort(sort)
        if record.get('interpretation'):
            il.sig.interp[name] = record['interpretation']
        elif len(record.get('bounds',())) == 2:
            lo,hi = (il.Symbol(bound,sort) for bound in record['bounds'])
            il.sig.interp[name] = il.RangeSort(name,lo,hi)


def decode_value(value):
    sorts = [il.find_sort(name) for name in value.get('signature',())]
    if not sorts:
        if value['name'] == 'true':
            return il.And()
        if value['name'] == 'false':
            return il.Or()
        raise iu.IvyError(None,'portable trace term is missing its sort')
    sort = sorts[0] if len(sorts) == 1 else il.FunctionSort(*sorts)
    symbol = il.Symbol(value['name'],sort)
    arguments = [decode_value(arg) for arg in value.get('arguments',())]
    return symbol(*arguments) if arguments else symbol


def replay(archive):
    if not ivy_trace.option_detailed.get():
        return archive['brief']
    terms = ivy_trace.tracked_terms()
    if not terms:
        return archive['detailed']
    tracking = archive.get('tracking')
    if tracking is None:
        raise iu.IvyError(None,'trace has no recorded values for tracking')
    previous_version = iu.get_string_version()
    seen,missing,lines = {},{},[]
    try:
        iu.set_string_version('.'.join(str(part) for part in tracking['version']))
        for step in tracking['steps']:
            lines.append(step['prefix'])
            if not step.get('facts'):
                continue
            with im.Module() as module:
                add_sorts(tracking['sorts'])
                facts = [il.Equals(*[decode_value(value) for value in pair]) for pair in step['facts']]
                state = interp.State(module,ilu.Clauses(facts))
                equations = ivy_trace.track_state(state,terms,missing)
            changes = []
            for equation in equations:
                key,value = map(str,equation.args)
                if seen.get(key) != value:
                    seen[key] = value
                    changes.append(str(equation))
            if changes:
                indent = '    ' * step['indent']
                lines.extend([indent+'[\n'] + [indent+'    '+line+'\n' for line in sorted(changes)] + [indent+']\n'])
    finally:
        iu.set_string_version(previous_version)
    for term in terms:
        if term not in missing or missing[term] is not None:
            raise iu.IvyError(None,f'cannot track {term!r}: expression is unavailable in this trace')
    lines.append(tracking['tail'])
    return ''.join(lines)
