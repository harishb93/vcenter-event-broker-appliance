package embedded

import (
	"context"
	"expvar"
	"fmt"
	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/metrics"
	"net/http"
	"os"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/util"

	config "github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/config/v1alpha1"
	"github.com/vmware-samples/vcenter-event-broker-appliance/vmware-event-router/internal/logger"
)

const (
	// DefaultListenAddress is the embedded address the http metrics server will listen
	// for requests
	httpTimeout = time.Second * 5
	endpoint    = "/stats"
)

var (
	eventRouterStats = expvar.NewMap(mapName)
)

type Server struct {
	*metrics.Server
}

// NewServer returns an initialized metrics server binding to addr
func NewServer(cfg *config.MetricsProviderConfigDefault, log logger.Logger) (*Server, error) {
	if cfg == nil {
		return nil, errors.New("no metrics server configuration found")
	}

	metricLog := log
	if zapSugared, ok := log.(*zap.SugaredLogger); ok {
		metricLog = zapSugared.Named("[METRICS]")
	}

	mux := http.NewServeMux()

	if cfg.Auth == nil || cfg.Auth.BasicAuth == nil {
		mux.Handle(endpoint, expvar.Handler())
		metricLog.Warnf("disabling basic auth: no authentication data provided")
	} else {
		mux.Handle(endpoint, withBasicAuth(metricLog, expvar.Handler(), cfg.Auth.BasicAuth.Username, cfg.Auth.BasicAuth.Password))

	}

	err := util.ValidateAddress(cfg.BindAddress)
	if err != nil {
		metricLog.Errorf("Could not validate bind address")
		return nil, errors.Wrap(err, "could not validate bind address")
	}

	srv := &Server{
		&metrics.Server{
			Http: &http.Server{
				Addr:         cfg.BindAddress,
				Handler:      mux,
				ReadTimeout:  httpTimeout,
				WriteTimeout: httpTimeout,
			},
			Logger:              metricLog,
			MetricsProviderType: config.MetricsProviderDefault,
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
		s.Infow("starting metrics server", "address", addr)
		err := s.Http.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	// continuously update the http stats endpoint
	go func() {
		s.publish(ctx)
	}()

	select {
	case <-ctx.Done():
		err := s.Http.Shutdown(ctx)
		if err != nil && err != http.ErrServerClosed {
			return errors.Wrap(err, "could not shutdown metrics server gracefully")
		}
	case err := <-errCh:
		return errors.Wrap(err, "could not run metrics server")
	}

	return nil
}

// withBasicAuth enforces basic auth as a middleware for the given username and
// password
func withBasicAuth(log logger.Logger, next http.Handler, u, p string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, password, ok := r.BasicAuth()

		w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)

		if !ok || !(p == password && u == user) {
			w.WriteHeader(http.StatusUnauthorized)
			_, err := w.Write([]byte("invalid credentials"))

			if err != nil {
				log.Errorf("could not write http response: %v", err)
			}

			return
		}

		next.ServeHTTP(w, r)
	}
}

func (s *Server) publish(ctx context.Context) {
	var (
		numberOfSecondsRunning = expvar.NewInt("system.numberOfSeconds") // uptime in sec
		programName            = expvar.NewString("system.programName")
		lastLoad               = expvar.NewFloat("system.lastLoad")
	)

	expvar.Publish("system.allLoad", expvar.Func(allLoadAvg))
	programName.Set(os.Args[0])

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			numberOfSecondsRunning.Add(1)
			lastLoad.Set(loadAvg(0))
		}
	}
}

// Process receives metrics from event streams and processors and exposes them
// under the predefined map. The sender is responsible for picking a unique
// Provider name.
func (s *Server) Process(stats *metrics.EventStats) {
	eventRouterStats.Set(stats.Provider, stats)
}
