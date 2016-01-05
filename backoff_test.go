package backoff

import (
	"testing"
	"time"
)

// If given New with 0's and no jitter, ensure that certain invariants are met:
//
//   - the default max duration and interval should be used
//   - noJitter should be true
//   - the RNG should not be initialised
//   - the first duration should be equal to the default interval
func TestDefaults(t *testing.T) {
	b := NewWithoutJitter(0, 0)

	if b.maxDuration != DefaultMaxDuration {
		t.Fatalf("expected new backoff to use the default max duration (%s), but have %s", DefaultMaxDuration, b.maxDuration)
	}

	if b.interval != DefaultInterval {
		t.Fatalf("exepcted new backoff to use the default interval (%s), but have %s", DefaultInterval, b.interval)
	}

	if b.noJitter != true {
		t.Fatal("backoff should have been initialised without jitter")
	}

	if b.rng != nil {
		t.Fatal("RNG should not have been initialised yet")
	}

	dur := b.Duration()
	if dur != DefaultInterval {
		t.Fatalf("expected first duration to be %s, have %s", DefaultInterval, dur)
	}
}

// Given a zero-value initialised Backoff, it should be transparently
// setup.
func TestSetup(t *testing.T) {
	b := new(Backoff)
	dur := b.Duration()
	if dur < 0 || dur > (5*time.Minute) {
		t.Fatalf("want duration between 0 and 5 minutes, have %s", dur)
	}
}

// Ensure that tries incremenets as expected.
func TestTries(t *testing.T) {
	b := NewWithoutJitter(5, 1)

	for i := uint64(0); i < 3; i++ {
		if b.tries != i {
			t.Fatalf("want tries=%d, have tries=%d", i, b.tries)
		} else if b.Tries() != i {
			t.Fatalf("want tries=%d, have tries=%d", i, b.Tries())
		}

		pow := 1 << i
		expected := time.Duration(pow)
		dur := b.Duration()
		if dur != expected {
			t.Fatalf("want duration=%d, have duration=%d at i=%d", expected, dur, i)
		}
	}

	for i := uint(3); i < 5; i++ {
		dur := b.Duration()
		if dur != 5 {
			t.Fatalf("want duration=5, have %d at i=%d", dur, i)
		}
	}
}

// Ensure that a call to Reset will actually reset the Backoff.
func TestReset(t *testing.T) {
	const iter = 10
	b := New(10, 1)
	for i := 0; i < iter; i++ {
		_ = b.Duration()
	}

	if b.tries != iter {
		t.Fatalf("expected tries=%d, have tries=%d", iter, b.tries)
	}

	b.Reset()
	if b.tries != 0 {
		t.Fatalf("expected tries=0 after reset, have tries=%d", b.tries)
	}
}
