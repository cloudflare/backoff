// Package backoff contains an implementation of an intelligent backoff
// strategy. It is based on the approach in the AWS architecture blog
// article titled "Exponential Backoff And Jitter", which is found at
// http://www.awsarchitectureblog.com/2015/03/backoff.html.

package backoff

import (
	mrand "math/rand"
	"sync"
	"time"
)

// DefaultInterval is used when a Backoff is initialised with a
// zero-value Interval.
var DefaultInterval = 5 * time.Minute

// DefaultMaxDuration is maximum amount of time that the backoff will
// delay for.
var DefaultMaxDuration = 6 * time.Hour

// A Backoff contains the information needed to intelligently backoff
// and retry operations using an exponential backoff algorithm. It may
// be initialised with all zero values and it will behave sanely.
type Backoff struct {
	// MaxDuration is the largest possible duration that can be
	// returned from a call to Duration.
	MaxDuration time.Duration

	// Interval controls the time step for backing off.
	Interval time.Duration

	// NoJitter controls whether to use the "Full Jitter"
	// improvement to attempt to smooth out spikes in a high
	// contention scenario. If NoJitter is set to true, no
	// jitter will be introduced.
	NoJitter bool

	tries uint
	lock  sync.Mutex // lock guards tries
}

func (b *Backoff) setup() {
	if b.Interval == 0 {
		b.Interval = DefaultInterval
	}

	if b.MaxDuration == 0 {
		b.MaxDuration = DefaultMaxDuration
	}
}

// Duration returns a time.Duration appropriate for the backoff,
// incrementing the attempt counter.
func (b *Backoff) Duration() time.Duration {
	b.setup()
	b.lock.Lock()
	defer b.lock.Unlock()

	pow := 1 << b.tries
	t := time.Duration(pow)
	t = b.Interval * t

	// Increment tries only if no overflow occurs
	if pow < (pow<<1) && t < b.Interval*time.Duration(pow<<1) {
		b.tries++
	}

	if t > b.MaxDuration {
		t = b.MaxDuration
	}

	if !b.NoJitter {
		t = time.Duration(mrand.Int63n(int64(t)))
	}

	return t
}

// Reset resets the attempt counter of a backoff.
//
// It should be called when the rate-limited action succeeds.
func (b *Backoff) Reset() {
	b.lock.Lock()
	defer b.lock.Unlock()

	b.tries = 0
}
