# Command core review

Scope: parser.go, cmd.h.go, cmdInt.go, cmdFlag.go, cmdAlias.go,
cmdApi.go, cmdHist.go, cmdUtils.go, cmd.c.go, cmdLoad.go, cmdStarter.go,
parser_test.go and core_test.go. This is the initial core-scope review; subsequent
AUTO_REVIEW.md, PLUGIN_REVIEW.md, EXTERNAL_REVIEW.md and FILES_REVIEW.md cover
the remaining files. The full package now builds and passes race tests and vet.

## Readability and idiomatic Go pass

Frame owns commands, aliases, flags, history, backups, timing and lifecycle.
The command package installs a dispatcher and registers definitions rather than
maintaining parallel registries. Shell commands retain their original quoted
text; external executable commands pass arguments directly to os/exec. Starter
uses a bounded worker pool and waits for every child instead of polling a global
thread counter. Redirected streams are lifecycle resources with explicit close.

## Separate bug and fidelity pass

Fixed exact `~` expansion (previously produced home/~). Alias commands preceding
the final alias expansion now share the recursion budget, with a regression for
recursive prefixes. History redirection now actually writes executed commands;
regression reads the file. Starter rejects fewer than two cores instead of
reproducing the source's infinite wait, propagates child failures, and serializes
shared writers. Its error variable is local to each worker. Tests cover quoted
semicolons, source nesting cleanup, alias cycles, autoexec recursion prevention,
history logging and failed child completion. The #CEC diagnostic reads real
Frame status instead of the source's hardcoded successful result.

The selected scope passes race tests and vet. Later module reviews document
the exact MF evaluator and Frame data-bridge integration dependencies.

A subsequent lifecycle audit found Cmd_End removed only basic commands. It now
matches the source teardown by clearing all shared commands/aliases/flags/history
and plugin metadata, after flushing explicitly configured persistent history.
A regression verifies extension state disappears and Cmd_Init can restart it.
