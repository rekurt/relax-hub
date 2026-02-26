package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/nikitaaldaev/bani/config"
	"github.com/nikitaaldaev/bani/internal/logger"
	"go.uber.org/fx"
)

var Module = fx.Module("server",
	fx.Provide(NewRouter),
	fx.Invoke(RegisterServer),
)

func RegisterServer(lc fx.Lifecycle, cfg *config.Config, router http.Handler, log *logger.Logger) {
	srv := &http.Server{
		Addr:              fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			ln, err := net.Listen("tcp", srv.Addr)
			if err != nil {
				log.Error("Failed to listen on", "addr", srv.Addr, "error", err)
				return err
			}
			log.Info("HTTP server listening on", "addr", srv.Addr)
			go func() {
				if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
					log.Error("HTTP server error", "error", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Info("Shutting down HTTP server")
			return srv.Shutdown(ctx)
		},
	})
}
