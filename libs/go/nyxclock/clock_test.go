package nyxclock

import (
	"testing"
	"time"
)

func TestReal_NowAdvances(t *testing.T) {
	a := Real{}.Now()
	time.Sleep(time.Millisecond)
	b := Real{}.Now()
	if !b.After(a) {
		t.Fatalf("expected time to advance, got a=%v b=%v", a, b)
	}
}

func TestFake_NowIsStable(t *testing.T) {
	fixed := time.Date(2026, 7, 14, 0, 0, 0, 0, time.UTC)
	f := Fake{T: fixed}
	if got := f.Now(); !got.Equal(fixed) {
		t.Fatalf("got %v, want %v", got, fixed)
	}
	if got := f.Now(); !got.Equal(fixed) {
		t.Fatalf("fake clock must not advance on its own, got %v", got)
	}
}

func TestFake_Advance(t *testing.T) {
	start := time.Date(2026, 7, 14, 0, 0, 0, 0, time.UTC)
	f := Fake{T: start}
	moved := f.Advance(time.Hour)
	if !moved.Now().Equal(start.Add(time.Hour)) {
		t.Fatalf("Advance did not move clock forward by 1h")
	}
	if !f.Now().Equal(start) {
		t.Fatalf("Advance must not mutate receiver")
	}
}
