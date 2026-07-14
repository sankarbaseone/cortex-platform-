// Package nyxclock provides an injectable time source so domain logic never
// calls time.Now() directly (RFC-012 O.4: inject clock, never time.Now() in domain).
package nyxclock

import "time"

// Clock is the sole time-source interface domain code should depend on.
type Clock interface {
	Now() time.Time
}

// Real is the production Clock, backed by the system wall clock.
type Real struct{}

func (Real) Now() time.Time { return time.Now() }

// Fake is a deterministic Clock for tests.
type Fake struct {
	T time.Time
}

func (f Fake) Now() time.Time { return f.T }

// Advance returns a new Fake moved forward by d, leaving f unmodified.
func (f Fake) Advance(d time.Duration) Fake {
	return Fake{T: f.T.Add(d)}
}
