package prometheus

import (
	"context"
	"fmt"
	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/metrics"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	config "github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/config/v1alpha1"
	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/logger"
)

const (
	// DefaultListenAddress is the embedded address the http metrics server will listen
	// for requests
	httpTimeout = time.Second * 5
	endpoint    = "/metrics"
)

type Server struct {
	*metrics.Server
}

// NewServer returns an initialized prometheus metrics server binding to addr
func NewServer(cfg *config.MetricsProviderConfigPrometheus, log logger.Logger) (*Server, error) {
	if cfg == nil {
		return nil, errors.New("no prometheus metrics server configuration found")
	}

	metricLog := log
	if zapSugared, ok := log.(*zap.SugaredLogger); ok {
		metricLog = zapSugared.Named("[METRICS]")
	}

	mux := http.NewServeMux()

	mux.Handle(endpoint, promhttp.Handler())

	srv := &Server{
		&metrics.Server{
			Http: &http.Server{
				Addr:         cfg.BindAddress,
				Handler:      mux,
				ReadTimeout:  httpTimeout,
				WriteTimeout: httpTimeout,
			},
			Logger:              metricLog,
			MetricsProviderType: config.MetricsProviderPrometheus,
		},
	}

	return srv, nil
}

// Run starts the metrics server until the context is cancelled or an error
// occurs. It will collect metrics for the given event streams and processors.
func (s *Server) Run(ctx context.Context) error {
	errCh := make(chan error, 1)
	defer close(errCh)

	go func() {
		addr := fmt.Sprintf("http://%s%s", s.Http.Addr, endpoint)
		s.Infow("starting prometheus metrics server", "address", addr)

		err := s.Http.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	// continuously update the http metrics endpoint
	go func() {
		s.publish(ctx)
	}()

	select {
	case <-ctx.Done():
		err := s.Http.Shutdown(ctx)
		if err != nil && err != http.ErrServerClosed {
			return errors.Wrap(err, "could not shutdown prometheus metrics server gracefully")
		}
	case err := <-errCh:
		return errors.Wrap(err, "could not run prometheus metrics server")
	}

	return nil
}

func (s *Server) publish(ctx context.Context) {
	//publish metrics providers and processors as labels and counters
}

// Process receives metrics from event streams and processors
// The sender is responsible for picking a unique Provider name.
func (s *Server) Process(stats *metrics.EventStats) {
	//Logic for Receiving stats
}
