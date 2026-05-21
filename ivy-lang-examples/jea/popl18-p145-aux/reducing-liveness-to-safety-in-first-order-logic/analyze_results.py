from numpy import std, mean

rs = [
    eval(open('results.dat').read()),
]

ivy_files = [
    'ticket_protocol.ivy',
    'alternating_bit_protocol.ivy',
    'tlb-termination.ivy',
    'paxos_liveness/paxos_liveness.ivy',
    'paxos_liveness/multi_paxos_liveness.ivy',
    'paxos_liveness/stoppable_paxos_liveness.ivy',
]

for r in rs:
    for k in ivy_files:
        v = r.get(k, [0])
        print '{:50}: mean = {:5.1f}, std = {:5.1f} ({:2.1f}%)'.format(
            k,
            mean(v),
            std(v),
            std(v) / mean(v) * 100,
        )
    print
