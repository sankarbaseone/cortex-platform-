package nyxbus

import (
	"errors"
	"fmt"
	"testing"
)

func TestPoison_IsPoison(t *testing.T) {
	root := errors.New("bad schema version")
	p := Poison(root)
	if !IsPoison(p) {
		t.Fatal("IsPoison(Poison(err)) = false")
	}
	if !errors.Is(p, root) {
		t.Fatal("Poison must preserve the chain for errors.Is")
	}
}

func TestPoison_Nil(t *testing.T) {
	if Poison(nil) != nil {
		t.Fatal("Poison(nil) must return nil")
	}
}

func TestIsPoison_PlainErrorIsNotPoison(t *testing.T) {
	if IsPoison(errors.New("transient pg error")) {
		t.Fatal("plain error must not be classified as poison")
	}
}

func TestPoison_WrappedByFmtErrorf(t *testing.T) {
	root := errors.New("boom")
	wrapped := fmt.Errorf("handler: %w", Poison(root))
	if !IsPoison(wrapped) {
		t.Fatal("IsPoison must see through fmt.Errorf %w wrapping")
	}
}
