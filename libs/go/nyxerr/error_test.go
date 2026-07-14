package nyxerr

import (
	"errors"
	"fmt"
	"testing"
)

func TestWrap_PreservesChain(t *testing.T) {
	root := errors.New("pg: connection refused")
	wrapped := Wrap(Unavailable, "repo.Get", root)

	if !errors.Is(wrapped, root) {
		t.Fatalf("errors.Is must see through nyxerr.Error to the root cause")
	}
	if KindOf(wrapped) != Unavailable {
		t.Fatalf("KindOf = %v, want Unavailable", KindOf(wrapped))
	}
	if !Is(wrapped, Unavailable) {
		t.Fatalf("Is(wrapped, Unavailable) = false")
	}
}

func TestWrap_Nil(t *testing.T) {
	if Wrap(Internal, "op", nil) != nil {
		t.Fatalf("Wrap(nil) must return nil")
	}
}

func TestWrap_FmtErrorfCompat(t *testing.T) {
	root := errors.New("boom")
	wrapped := Wrap(Internal, "op", root)
	doubled := fmt.Errorf("outer: %w", wrapped)
	if !errors.Is(doubled, root) {
		t.Fatalf("chain broken through fmt.Errorf %%w")
	}
	if KindOf(doubled) != Internal {
		t.Fatalf("KindOf through fmt.Errorf wrap = %v, want Internal", KindOf(doubled))
	}
}

func TestKindOf_UnclassifiedReturnsUnknown(t *testing.T) {
	if KindOf(errors.New("plain")) != Unknown {
		t.Fatalf("KindOf(plain error) must be Unknown")
	}
	if KindOf(nil) != Unknown {
		t.Fatalf("KindOf(nil) must be Unknown")
	}
}

func TestNew(t *testing.T) {
	err := New(NotFound, "repo.Get", "kernel not found")
	if KindOf(err) != NotFound {
		t.Fatalf("KindOf(New) = %v, want NotFound", KindOf(err))
	}
	if err.Error() != "repo.Get: kernel not found" {
		t.Fatalf("Error() = %q", err.Error())
	}
}

func TestKind_Reason(t *testing.T) {
	cases := map[Kind]string{
		NotFound:        "NOT_FOUND",
		InvalidArgument: "INVALID_ARGUMENT",
		Unavailable:     "UNAVAILABLE",
		Unknown:         "UNKNOWN",
	}
	for k, want := range cases {
		if got := k.Reason(); got != want {
			t.Fatalf("Kind(%d).Reason() = %q, want %q", k, got, want)
		}
	}
}
