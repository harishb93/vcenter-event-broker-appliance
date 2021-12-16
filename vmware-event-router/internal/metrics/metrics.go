package metrics

import (
	"context"
)

// Processor receives metrics from metric providers
type Processor interface {
	// TODO: include context
	Process(stats *EventStats)
	Run(ctx context.Context) error
}
