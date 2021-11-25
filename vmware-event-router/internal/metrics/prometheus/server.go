package prometheus

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	config "github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/config/v1alpha1"
	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/logger"
	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/metrics"
)

const (
	// DefaultListenAddress is the default address the http metrics server will listen
	// for requests
	httpTimeout = time.Second * 5
	endpoint    = "/metrics"
)

// Receiver receives metrics from metric providers
type Receiver interface {
	Receive(stats *metrics.EventStats)
}

// verify that metrics server implements Receiver
var _ Receiver = (*Server)(nil)

// Server is the implementation of the metrics server
type Server struct {
	http *http.Server
	logger.Logger
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
		http: &http.Server{
			Addr:         cfg.BindAddress,
			Handler:      mux,
			ReadTimeout:  httpTimeout,
			WriteTimeout: httpTimeout,
		},
		Logger: metricLog,
	}

	return srv, nil
}

// Run starts the metrics server until the context is cancelled or an error
// occurs. It will collect metrics for the given event streams and processors.
func (s *Server) Run(ctx context.Context) error {
	errCh := make(chan error, 1)
	defer close(errCh)

	go func() {
		addr := fmt.Sprintf("http://%s%s", s.http.Addr, endpoint)
		s.Infow("starting prometheus metrics server", "address", addr)

		err := s.http.ListenAndServe()
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
		err := s.http.Shutdown(ctx)
		if err != nil && err != http.ErrServerClosed {
			return errors.Wrap(err, "could not shutdown prometheus metrics server gracefully")
		}
	case err := <-errCh:
		return errors.Wrap(err, "could not run prometheus metrics server")
	}

	return nil
}

func (s *Server) publish(ctx context.Context) {
	//publish metrics to endpoint using exporters??
}

// Receive receives metrics from event streams and processors
// The sender is responsible for picking a unique Provider name.
func (s *Server) Receive(stats *metrics.EventStats) {
	//Logic for Receiving stats
}
