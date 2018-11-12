package swim

import (
	"sync"
	"time"
)

// Awareness manages health of the local node. This related to Lifeguard L1
// "self-awareness" concept. Expect to receive replies to probe messages we sent
// Start timeouts low, increase in response to absence of replies
type Awareness struct {
	sync.RWMutex

	// max is the upper threshold for the factor that increase the timeout value
	// (the score will be constrained from 0 <= score < max)
	max uint

	// score is the current awareness score. Lower values are healthier and
	// zero is the minimum value
	score uint
}

func NewAwareness(max uint) *Awareness {
	return &Awareness{
		max:   max,
		score: 0,
	}
}

func (a *Awareness) GetHealthScore() uint {
	a.RLock()
	defer a.RUnlock()
	return a.score
}

// ApplyDelta with given delta applies it to the score in thread-safe manner
// score must be bound from 0 to max value
func (a *Awareness) ApplyDelta(delta int) {
	return
}

// ScaleTimeout takes the given duration and scales it based on the current score.
// Less healthyness will lead to longer timeouts.
func (a *Awareness) ScaleTimeout(timeout time.Duration) time.Duration {
	return timeout
}
