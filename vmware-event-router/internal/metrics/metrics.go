package metrics

import (
	"context"
)

// Receiver receives metrics from metric providers
type Receiver interface {
	Process(stats *EventStats)
	Run(ctx context.Context) error
}
