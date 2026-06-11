package ivyrepl

import "github.com/glycerine/ivy/goivy/zlisp"

type Zlisp = zlisp.Zlisp
type ZlispConfig = zlisp.ZlispConfig
type Sexp = zlisp.Sexp

type Session struct {
	Env *zlisp.Zlisp
}

func NewZlispConfig(cmdname string) *zlisp.ZlispConfig {
	return zlisp.NewZlispConfig(cmdname)
}

func NewEnv(sandboxed bool) *zlisp.Zlisp {
	var env *zlisp.Zlisp
	if sandboxed {
		env = zlisp.NewZlispSandbox()
	} else {
		env = zlisp.NewZlisp()
	}
	env.StandardSetup()
	Install(env)
	return env
}

func NewSession(sandboxed bool) *Session {
	return &Session{Env: NewEnv(sandboxed)}
}

func Install(env *zlisp.Zlisp) {
	registerTypes()
	installZ3(env)
	installIvy(env)
}

func Repl(env *zlisp.Zlisp, cfg *zlisp.ZlispConfig) {
	zlisp.Repl(env, cfg)
}

func ReplMain(cfg *zlisp.ZlispConfig) {
	prevHook := cfg.SetupHook
	cfg.SetupHook = func(env *zlisp.Zlisp) {
		if prevHook != nil {
			prevHook(env)
		}
		Install(env)
	}
	zlisp.ReplMain(cfg)
}
