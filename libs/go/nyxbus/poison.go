package nyxbus

import "errors"

// poisonError marks a handler error as a poison pill: the payload could not
// be decoded/processed and must go straight to DLQ without retry, with the
// envelope forwarded intact (RFC-005 B.4).
type poisonError struct{ err error }

func (p *poisonError) Error() string { return "nyxbus: poison: " + p.err.Error() }
func (p *poisonError) Unwrap() error { return p.err }

// Poison wraps err to signal the consumer runtime to skip retries and
// forward the record straight to DLQ. Returns nil if err is nil.
func Poison(err error) error {
	if err == nil {
		return nil
	}
	return &poisonError{err: err}
}

// IsPoison reports whether err (or its chain) was produced by Poison.
func IsPoison(err error) bool {
	var p *poisonError
	return errors.As(err, &p)
}
