// Package backoff contains an implementation of an intelligent backoff
// strategy. It is based on the approach in the AWS architecture blog
// article titled "Exponential Backoff And Jitter", which is found at
// http://www.awsarchitectureblog.com/2015/03/backoff.html.
//
// Essentially, the backoff has an interval `time.Duration`; the nth
// call to backoff will return a `time.Duration` that is 2^n *
// interval. If jitter is enabled (which is the default behaviour),
// the duration is a random value between 0 and 2^n * interval.  The
// backoff is configured with a maximum duration that will not be
// exceeded.
//
// The `New` function will attempt to use the system's cryptographic
// random number generator to seed a Go math/rand random number
// source. If this fails, it will fall back to using the Unix
// timestamp as the seed.
package backoff

import (
	"crypto/rand"
	"encoding/binary"
	"io"
	mrand "math/rand"
	"sync"
	"time"
)

var prng *mrand.Rand

// DefaultInterval is used when a Backoff is initialised with a
// zero-value Interval.
var DefaultInterval = 5 * time.Minute

// DefaultMaxDuration is maximum amount of time that the backoff will
// delay for.
var DefaultMaxDuration = 6 * time.Hour

// A Backoff contains the information needed to intelligently backoff
// and retry operations using an exponential backoff algorithm. It should
// be initialised with a call to `New`.
type Backoff struct {
	// maxDuration is the largest possible duration that can be
	// returned from a call to Duration.
	maxDuration time.Duration

	// interval controls the time step for backing off.
	interval time.Duration

	// noJitter controls whether to use the "Full Jitter"
	// improvement to attempt to smooth out spikes in a high
	// contention scenario. If noJitter is set to true, no
	// jitter will be introduced.
	noJitter bool

	tries, n uint64
	lock     sync.Mutex // lock guards tries
	rng      *mrand.Rand
}

// New creates a new backoff with the specified max duration and
// interval. Zero values may be used to use the default values.
func New(max time.Duration, interval time.Duration) *Backoff {
	b := &Backoff{
		maxDuration: max,
		interval:    interval,
	}

	b.setup()
	return b
}

// NewWithoutJitter works similarly to New, except that the created
// Backoff will not use jitter.
func NewWithoutJitter(max time.Duration, interval time.Duration) *Backoff {
	b := New(max, interval)
	b.noJitter = true
	return b
}

func init() {
	var buf [8]byte
	var n int64

	_, err := io.ReadFull(rand.Reader, buf[:])
	if err != nil {
		panic(err.Error())
	}

	n = int64(binary.LittleEndian.Uint64(buf[:]))

	src := mrand.NewSource(n)
	prng = mrand.New(src)
}

func (b *Backoff) setup() {
	if b.interval == 0 {
		b.interval = DefaultInterval
	}

	if b.maxDuration == 0 {
		b.maxDuration = DefaultMaxDuration
	}
}

// Duration returns a time.Duration appropriate for the backoff,
// incrementing the attempt counter.
func (b *Backoff) Duration() time.Duration {
	b.setup()
	b.lock.Lock()
	defer b.lock.Unlock()

	b.tries++
	pow := uint64(1 << b.n)
	t := time.Duration(pow)
	t = b.interval * t
	// Increment n only if no overflow occurs
	if pow < (pow<<1) && t < b.interval*time.Duration(pow<<1) {
		b.n++
	}

	if t > b.maxDuration {
		t = b.maxDuration
	}

	if !b.noJitter {
		t = time.Duration(prng.Int63n(int64(t)))
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
	b.n = 0
}

// Tries returns the current number of attempts that have been made.
func (b *Backoff) Tries() uint64 {
	b.lock.Lock()
	defer b.lock.Unlock()
	return b.tries
}
