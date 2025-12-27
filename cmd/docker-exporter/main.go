package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nhattuanbl/docker-exporter/internal/collector"
	"github.com/nhattuanbl/docker-exporter/internal/config"
	"github.com/nhattuanbl/docker-exporter/internal/docker"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	// Parse configuration
	cfg, ok := config.Parse()
	if !ok {
		os.Exit(1)
	}

	// Setup logger
	logger := setupLogger(cfg.LogLevel, cfg.LogPath)
	defer logger.Sync()

	logger.Info("Starting Docker Exporter",
		zap.String("version", config.Version),
		zap.String("docker_host", cfg.DockerHost),
		zap.String("address", cfg.Address()),
		zap.String("metrics_path", cfg.MetricsPath()),
	)

	// Create Docker client
	dockerClient, err := docker.NewClient(cfg.DockerHost)
	if err != nil {
		logger.Fatal("Failed to create Docker client", zap.Error(err))
	}
	defer dockerClient.Close()

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	if err := dockerClient.Ping(ctx); err != nil {
		logger.Fatal("Failed to connect to Docker daemon",
			zap.String("host", cfg.DockerHost),
			zap.Error(err))
	}
	cancel()
	logger.Info("Connected to Docker daemon", zap.String("host", cfg.DockerHost))

	// Create and register collector
	coll := collector.NewCollector(dockerClient, cfg, logger)
	prometheus.MustRegister(coll)

	// Setup HTTP server
	mux := http.NewServeMux()

	// Metrics endpoint
	mux.Handle(cfg.MetricsPath(), promhttp.Handler())

	// Root endpoint
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head><title>Docker Exporter</title></head>
<body>
<h1>Docker Exporter</h1>
<p>Version: %s</p>
<p><a href="%s">Metrics</a></p>
</body>
</html>`, config.Version, cfg.MetricsPath())
	})

	// Health endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := dockerClient.Ping(ctx); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprintf(w, "unhealthy: %v", err)
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "healthy")
	})

	server := &http.Server{
		Addr:         cfg.Address(),
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		logger.Info("Shutting down server...")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		server.Shutdown(ctx)
	}()

	// Start server
	logger.Info("Server listening", zap.String("address", cfg.Address()))
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		logger.Fatal("Server error", zap.Error(err))
	}

	logger.Info("Server stopped")
}

// setupLogger creates a zap logger
func setupLogger(level, path string) *zap.Logger {
	// Parse log level
	var zapLevel zapcore.Level
	switch level {
	case "debug":
		zapLevel = zapcore.DebugLevel
	case "info":
		zapLevel = zapcore.InfoLevel
	case "warn":
		zapLevel = zapcore.WarnLevel
	case "error":
		zapLevel = zapcore.ErrorLevel
	default:
		zapLevel = zapcore.InfoLevel
	}

	// Encoder config
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// Outputs
	var outputs []zapcore.WriteSyncer
	outputs = append(outputs, zapcore.AddSync(os.Stdout))

	if path != "" {
		file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err == nil {
			outputs = append(outputs, zapcore.AddSync(file))
		}
	}

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.NewMultiWriteSyncer(outputs...),
		zapLevel,
	)

	return zap.New(core)
}
