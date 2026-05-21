"""
Run like this:

bash: time python run_experiment.py  1>experiment.out 2>experiment.err &
tcsh: nohup python run_experiment.py  >& experiment.out &
"""
import subprocess
import time
from numpy import std, mean
from os.path import isfile, expanduser
import sys

def flush():
    sys.stdout.flush()
    sys.stderr.flush()

def time_command(cmd, n=10):
    print "\n=== running {} times of {}\n".format(n, cmd)
    times = []
    for i in range(n):
        print "=== starting # {:3} of {}\n".format(i + 1, cmd)
        flush()
        t = time.time()
        subprocess.call(cmd)
        t = time.time() - t
        flush()
        print "\n=== finished # {:3}, took {} seconds\n".format(i + 1, t)
        times.append(t)
    print "=== timing results for {} times of {}:\n=== mean = {}\n=== std = {}\n=== times = {}\n\n".format(n, cmd, mean(times), std(times), times)
    return times

ivy_files = [
    'ticket_protocol.ivy',
    'alternating_bit_protocol.ivy',
    'paxos_liveness/paxos_liveness.ivy',
    'paxos_liveness/multi_paxos_liveness.ivy',
    'paxos_liveness/stoppable_paxos_liveness.ivy',
    'tlb-termination.ivy',
]

results = {}
for fn in ivy_files:
    if not isfile(fn):
        continue
    results[fn] = time_command([
        'timeout',
        '600',
        expanduser('ivy_check'),
        expanduser('~/reducing-liveness-to-safety-in-first-order-logic/' + fn),
    ], 10)
    open('results.dat', 'w').write(repr(results))

print "\n\nALL DONE\n{}".format(results)
