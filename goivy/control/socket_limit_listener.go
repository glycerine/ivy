package control

import (
	"net"
	"sync"
	"time"

	"github.com/glycerine/rate"
)

const (
	defaultMaxConcurrentSockets = 500
	defaultGlobalSocketBurst    = 100
)

type socketLimitConfig struct {
	MaxConcurrent int
	GlobalLimit   rate.Limit
	GlobalBurst   int
	RemoteEvery   time.Duration
}

func defaultSocketLimitConfig() socketLimitConfig {
	return socketLimitConfig{
		MaxConcurrent: defaultMaxConcurrentSockets,
		GlobalLimit:   rate.Limit(100),
		GlobalBurst:   defaultGlobalSocketBurst,
		RemoteEvery:   2 * time.Second,
	}
}

func newSocketLimitListener(l net.Listener, cfg socketLimitConfig) net.Listener {
	cfg = normalizeSocketLimitConfig(cfg)
	return &socketLimitListener{
		Listener: l,
		sem:      make(chan struct{}, cfg.MaxConcurrent),
		done:     make(chan struct{}),
		limits:   newSocketLimitState(cfg, time.Now),
	}
}

type socketLimitListener struct {
	net.Listener
	sem       chan struct{}
	closeOnce sync.Once
	done      chan struct{}
	limits    *socketLimitState
}

func (l *socketLimitListener) Accept() (net.Conn, error) {
	for {
		if !l.acquire() {
			for {
				c, err := l.Listener.Accept()
				if err != nil {
					return nil, err
				}
				_ = c.Close()
			}
		}

		c, err := l.Listener.Accept()
		if err != nil {
			l.release()
			return nil, err
		}
		if !l.limits.allow(c.RemoteAddr()) {
			_ = c.Close()
			l.release()
			continue
		}
		return &socketLimitConn{Conn: c, release: l.release}, nil
	}
}

func (l *socketLimitListener) Close() error {
	err := l.Listener.Close()
	l.closeOnce.Do(func() { close(l.done) })
	return err
}

func (l *socketLimitListener) acquire() bool {
	select {
	case <-l.done:
		return false
	case l.sem <- struct{}{}:
		return true
	}
}

func (l *socketLimitListener) release() {
	<-l.sem
}

type socketLimitConn struct {
	net.Conn
	releaseOnce sync.Once
	release     func()
}

func (c *socketLimitConn) Close() error {
	err := c.Conn.Close()
	c.releaseOnce.Do(c.release)
	return err
}

type socketLimitState struct {
	mu          sync.Mutex
	global      *rate.Limiter
	remote      map[string]*remoteSocketLimit
	remoteRate  rate.Limit
	remoteTTL   time.Duration
	lastCleanup time.Time
	now         func() time.Time
}

type remoteSocketLimit struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

func newSocketLimitState(cfg socketLimitConfig, now func() time.Time) *socketLimitState {
	cfg = normalizeSocketLimitConfig(cfg)
	if now == nil {
		now = time.Now
	}
	return &socketLimitState{
		global:     rate.NewLimiter(cfg.GlobalLimit, cfg.GlobalBurst),
		remote:     make(map[string]*remoteSocketLimit),
		remoteRate: rate.Every(cfg.RemoteEvery),
		remoteTTL:  cfg.RemoteEvery * 2,
		now:        now,
	}
}

func normalizeSocketLimitConfig(cfg socketLimitConfig) socketLimitConfig {
	if cfg.MaxConcurrent <= 0 {
		cfg.MaxConcurrent = defaultMaxConcurrentSockets
	}
	if cfg.GlobalLimit <= 0 {
		cfg.GlobalLimit = rate.Limit(100)
	}
	if cfg.GlobalBurst <= 0 {
		cfg.GlobalBurst = defaultGlobalSocketBurst
	}
	if cfg.RemoteEvery <= 0 {
		cfg.RemoteEvery = 2 * time.Second
	}
	return cfg
}

func (s *socketLimitState) allow(addr net.Addr) bool {
	return s.allowAt(remoteSocketKey(addr), s.now())
}

func (s *socketLimitState) allowAt(remoteKey string, now time.Time) bool {
	if !s.global.AllowN(now, 1) {
		return false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cleanupLocked(now)
	remote := s.remote[remoteKey]
	if remote == nil {
		remote = &remoteSocketLimit{limiter: rate.NewLimiter(s.remoteRate, 1)}
		s.remote[remoteKey] = remote
	}
	remote.lastSeen = now
	return remote.limiter.AllowN(now, 1)
}

func (s *socketLimitState) cleanupLocked(now time.Time) {
	if !s.lastCleanup.IsZero() && now.Sub(s.lastCleanup) < time.Minute {
		return
	}
	s.lastCleanup = now
	for key, remote := range s.remote {
		if now.Sub(remote.lastSeen) > s.remoteTTL {
			delete(s.remote, key)
		}
	}
}

func remoteSocketKey(addr net.Addr) string {
	if addr == nil {
		return ""
	}
	return addr.String()
}
