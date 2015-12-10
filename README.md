# backoff
## Go implementation of "Exponential Backoff And Jitter"

This package implements the backoff strategy described in the AWS
Architecture Blog article
["Exponential Backoff And Jitter"](http://www.awsarchitectureblog.com/2015/03/backoff.html). Essentially,
the backoff has an interval `time.Duration`; the *n<sup>th</sup>* call
to backoff will return an a `time.Duration` that is *2 <sup>n</sup> *
interval*. If jitter is enabled (which is the default behaviour), the
duration is a random value between 0 and *2 <sup>n</sup> * interval*.
The backoff is configured with a maximum duration that will not be
exceed; e.g., by default, the longest duration returned is
`backoff.DefaultMaxDuration`.

## Usage

A `Backoff` needs no initialisation to use the default behaviour:

```
package something

import "github.com/cloudflare/backoff"

func retryable() {
        b := &backoff.Backoff{}
        for {
                err := someOperation()
                if err == nil {
                    break
                }

                log.Printf("error in someOperation: %v", err)
                <-time.After(b.Duration())
        }
}
```

## Tunables

A `Backoff` has three fields that control its behaviour that may be
set when the `struct` is initialised:

* `MaxDuration` is the longest duration returned.
* `Interval` is the backoff's interval.
* `NoJitter`, if true, will not use jitter for the backoff.

The default behaviour is controlled by two variables:

* `DefaultInterval` sets the base interval for backoffs created with
  the zero `time.Duration` value in the `Interval` field.
* `DefaultMaxDuration` sets the maximum duration for backoffs created
  with the zero `time.Duration` value in the `MaxDuration` field.

