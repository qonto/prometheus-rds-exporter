// Package http provides webserver functionnalities
package http

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/qonto/prometheus-rds-exporter/internal/infra/build"
)

const (
	ReadTimeout       = 1
	WriteTimeout      = 120
	IdleTimeout       = 30
	ReadHeaderTimeout = 2
	shutdownTimeout   = 5
	httpErrorExitCode = 4
)

type Component struct {
	config *Config
	logger *slog.Logger
	server *http.Server
}

type Config struct {
	MetricPath    string
	ListenAddress string
	TLSKeyPath    string
	TLSCertPath   string
}

func New(logger slog.Logger, config Config) (component Component) {
	component = Component{
		logger: &logger,
		config: &config,
	}

	return
}

func (c *Component) Start() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c.server = &http.Server{
		Addr:              c.config.ListenAddress,
		ReadTimeout:       ReadTimeout * time.Second,
		WriteTimeout:      WriteTimeout * time.Second,
		IdleTimeout:       IdleTimeout * time.Second,
		ReadHeaderTimeout: ReadHeaderTimeout * time.Second,
		BaseContext:       func(_ net.Listener) context.Context { return ctx },
	}

	homepage, err := NewHomePage(build.Version, c.config.MetricPath)
	if err != nil {
		return fmt.Errorf("hompage initialization failed: %w", err)
	}

	http.Handle("/", homepage)
	http.Handle(c.config.MetricPath, promhttp.Handler())

	signalChan := make(chan os.Signal, 1)
	signal.Notify(
		signalChan,
		syscall.SIGHUP,  // kill -SIGHUP XXXX
		syscall.SIGINT,  // kill -SIGINT XXXX or Ctrl+c
		syscall.SIGQUIT, // kill -SIGQUIT XXXX
	)

	go func() {
		var err error

		if c.config.TLSCertPath != "" && c.config.TLSKeyPath != "" {
			c.logger.Info("starting the HTTPS server component")
			err = c.server.ListenAndServeTLS(c.config.TLSCertPath, c.config.TLSKeyPath)
		} else {
			c.logger.Info("starting the HTTP server component")
			err = c.server.ListenAndServe()
		}

		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			c.logger.Error("can't start web server", "reason", err)
			os.Exit(httpErrorExitCode)
		}
	}()

	<-signalChan // Wait until program received a stop signal

	err = c.Stop()
	if err != nil {
		return fmt.Errorf("can't stop websserver: %w", err)
	}

	return nil
}

func (c *Component) Stop() error {
	c.logger.Info("stopping the web server component")

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout*time.Second)
	defer cancel()

	err := c.server.Shutdown(ctx)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("can't stop webserver: %w", err)
	}

	c.logger.Info("web server stopped")

	return nil
}
