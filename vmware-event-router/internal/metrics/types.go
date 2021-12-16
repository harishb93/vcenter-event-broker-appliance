package metrics

import (
	"encoding/json"
	"time"
)

const (
	// PushInterval defines the embedded interval event streams and processors
	// push their metrics to the server
	PushInterval = time.Second * 1
)

// InvocationDetails contains success and failure information
type InvocationDetails struct {
	SuccessCount int
	FailureCount int
}

// Success records a successful invocation
func (i *InvocationDetails) Success() {
	i.SuccessCount++
}

// Failure records a failed invocation
func (i *InvocationDetails) Failure() {
	i.FailureCount++
}

// EventStats are provided and continuously updated by event streams and
// processors
type EventStats struct {
	Provider    string                        `json:"-"`    // ignored in JSON because provider is implicit via mapName[Provider]
	Type        string                        `json:"type"` // EventProvider or EventProcessor
	Address     string                        `json:"address"`
	Started     time.Time                     `json:"started"`
	EventsTotal *int                          `json:"events_total,omitempty"`   // only used by event streams, total events received
	EventsErr   *int                          `json:"events_err,omitempty"`     // only used by event streams, events received which lead to error
	EventsSec   *float64                      `json:"events_per_sec,omitempty"` // only used by event streams
	Invocations map[string]*InvocationDetails `json:"invocations,omitempty"`    // event.Category to success/failure invocations - only used by event processors
}

func (s *EventStats) String() string {
	b, err := json.Marshal(s)
	if err != nil {
		// will be printed to http stats endpoint
		return err.Error()
	}

	return string(b)
}
