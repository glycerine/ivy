#! /usr/bin/env python
#
# Copyright (c) Microsoft Corporation. All Rights Reserved.
#

from .ivy_init import get_verb, ivy_init
from . import ivy_module
from . import ivy_trace
from . import ivy_utils
from . import ivy_trace_archive

def main():
    import signal
    signal.signal(signal.SIGINT,signal.SIG_DFL)
    verb = get_verb()
    if verb == 'check':
        from . import ivy_check
        ivy_check.main()
        return
    if verb not in ('interact','replay'):
        print('unknown verb: {}'.format(verb))
        exit(1)
    with ivy_module.Module():
        with ivy_utils.ErrorPrinter():
            graph = ivy_init(verb=verb)
            if verb == 'replay':
                if ivy_trace_archive.export_trace.get():
                    ivy_trace_archive.export(graph)
                else:
                    print(ivy_trace_archive.replay(graph) if isinstance(graph,dict) else graph,end='')
            else:
                from .tk_ui import ui_main_loop
                ui_main_loop(graph)

if __name__ == "__main__":
    main()

