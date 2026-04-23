package server

import (
	"context"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rekurt/relax-hub/config"
	"github.com/rekurt/relax-hub/internal/logger"
	"go.uber.org/fx"
)

// MetricsPort is the TCP port that exposes Prometheus metrics.
//
// Kept off the public HTTP surface (the main API runs on cfg.Server.Port
// and is reachable through the public Ingress). A second listener on
// this port is only exposed as a ClusterIP Service named "metrics" and
// is scraped by a Prometheus ServiceMonitor inside the cluster.
const MetricsPort = 9090

// RegisterMetricsServer starts a dedicated HTTP server that serves
// Prometheus metrics at /metrics. Lifecycle is bound to the same fx
// container as the main API server so both stop cleanly on shutdown.
func RegisterMetricsServer(lc fx.Lifecycle, cfg *config.Config, log *logger.Logger) {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	// /healthz is handy for ServiceMonitor pre-flight checks.
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok\n"))
	})

	srv := &http.Server{
		Addr:              net.JoinHostPort(cfg.Server.Host, strconv.Itoa(MetricsPort)),
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			ln, err := net.Listen("tcp", srv.Addr)
			if err != nil {
				log.Error("Failed to listen for metrics", "addr", srv.Addr, "error", err)
				return err
			}
			log.Info("Metrics server listening on", "addr", srv.Addr)
			go func() {
				if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
					log.Error("Metrics server error", "error", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Info("Shutting down metrics server")
			return srv.Shutdown(ctx)
		},
	})
}
