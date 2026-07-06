package webui

// ivySample is a minimal Ivy file shared by web and non-web webui tests.
const ivySample = `#lang ivy1.7
type client
type server
relation link(X:client, Y:server)
relation semaphore(X:server)
after init { semaphore(W) := true; link(X,Y) := false }
action connect(x:client,y:server) = { require semaphore(y); link(x,y) := true; semaphore(y) := false }
export connect
conjecture link(X,Y) -> ~semaphore(Y)
`
