package main

type state int

const (
	open     state = iota
	halfopen state = iota
	closed   state = iota
)

type breaker struct {
	threshold  float64
	minSamples int64
	clog       *circularLog
	state      state
}

// NewBreaker creates a new circuit breaker with the specified threshold, etc.
func NewBreaker(threshold float64, minSamples int64, clog *circularLog) *breaker {
	return &breaker{
		threshold:  threshold,
		minSamples: minSamples,
		clog:       clog,
		state:      closed,
	}
}

// Success notifies the breaker that there was a successful operation
func (b *breaker) Success() {
	b.clog.Success()
	// we should also check if we're half-open, to decide if we
	// are able to close it
	if b.state == halfopen {
		if b.clog.Percent() < b.threshold {
			// we're back below the threshold. close it down.
			b.state = closed
		}
	}
}

// Fail notifies the breaker that there was a failure
// possibly causing it to fail if it causes it to cross the threshold
func (b *breaker) Fail() {
	b.clog.Fail()
	// here we should also check the clog to decide if
	// we need to trip the breaker
	if b.clog.Total() <= b.minSamples {
		// not enough samples to justify tripping the breaker
		return
	}
	if b.clog.Percent() >= b.threshold {
		b.state = open
	}
}

// Tripped reports whether the broker is open or not
// halfopen counts as closed for this purpose
func (b *breaker) Tripped() bool {
	return b.state == open
}
