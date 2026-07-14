// Package nyxerr provides a shared, transport-agnostic error taxonomy for
// domain and adapter code across NYDUX services. dependencies.yaml explicitly
// allows the domain layer to import only nyxerr + stdlib, so domain code can
// classify errors (e.g. "not found" vs "invalid argument") without depending
// on any transport (grpc/http) package. Transport adapters map Kind to their
// wire format (grpc status codes, RFC 9457 problem+json, etc).
package nyxerr

import (
	"errors"
	"fmt"
)

// Kind classifies an error independent of any transport encoding.
type Kind int

const (
	Unknown Kind = iota
	NotFound
	InvalidArgument
	AlreadyExists
	PermissionDenied
	Unauthenticated
	FailedPrecondition
	Unavailable
	Internal
	Canceled
	DeadlineExceeded
)

func (k Kind) String() string {
	switch k {
	case NotFound:
		return "not_found"
	case InvalidArgument:
		return "invalid_argument"
	case AlreadyExists:
		return "already_exists"
	case PermissionDenied:
		return "permission_denied"
	case Unauthenticated:
		return "unauthenticated"
	case FailedPrecondition:
		return "failed_precondition"
	case Unavailable:
		return "unavailable"
	case Internal:
		return "internal"
	case Canceled:
		return "canceled"
	case DeadlineExceeded:
		return "deadline_exceeded"
	default:
		return "unknown"
	}
}

// Reason returns the canonical UPPER_SNAKE_CASE form used for
// google.rpc.ErrorInfo.reason (RFC-006 F.5).
func (k Kind) Reason() string {
	switch k {
	case Unknown:
		return "UNKNOWN"
	default:
		s := k.String()
		out := make([]byte, 0, len(s))
		for i := 0; i < len(s); i++ {
			if s[i] == ' ' {
				continue
			}
			if s[i] >= 'a' && s[i] <= 'z' {
				out = append(out, s[i]-'a'+'A')
			} else {
				out = append(out, s[i])
			}
		}
		return string(out)
	}
}

// Error is a Kind-classified, chain-preserving error. It wraps an underlying
// cause (if any) so errors.Is/errors.As/%w continue to work through it.
type Error struct {
	Kind Kind
	Op   string
	Err  error
}

func (e *Error) Error() string {
	if e.Err == nil {
		return fmt.Sprintf("%s: %s", e.Op, e.Kind)
	}
	return fmt.Sprintf("%s: %v", e.Op, e.Err)
}

func (e *Error) Unwrap() error { return e.Err }

// New creates a Kind-classified error with no wrapped cause.
func New(kind Kind, op, msg string) error {
	return &Error{Kind: kind, Op: op, Err: errors.New(msg)}
}

// Wrap classifies err as Kind, labeling it with op, while preserving the
// original error in the chain (errors.Is/As still see through it).
// Returns nil if err is nil.
func Wrap(kind Kind, op string, err error) error {
	if err == nil {
		return nil
	}
	return &Error{Kind: kind, Op: op, Err: err}
}

// Is reports whether err's chain contains a *Error of the given Kind.
func Is(err error, kind Kind) bool {
	return KindOf(err) == kind
}

// KindOf walks err's chain and returns the first nyxerr Kind found, or
// Unknown if err (or its chain) carries none.
func KindOf(err error) Kind {
	var e *Error
	if errors.As(err, &e) {
		return e.Kind
	}
	return Unknown
}
