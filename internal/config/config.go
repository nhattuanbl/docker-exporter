package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/pflag"
)

// Version information (set at build time)
var (
	Version   = "dev"
	GitCommit = "unknown"
	BuildDate = "unknown"
	GoVersion = "unknown"
)

// Config holds the application configuration
type Config struct {
	Host       string
	Port       int
	Endpoint   string
	Prefix     string
	LogLevel   string
	LogPath    string
	DockerHost string
	OutputMode string        // "minimum" or "all"
	Timeout    time.Duration // API request timeout
}

// Parse parses command line flags and returns the configuration
func Parse() (*Config, bool) {
	cfg := &Config{}

	pflag.StringVarP(&cfg.Host, "host", "h", "0.0.0.0", "Bind address")
	pflag.IntVarP(&cfg.Port, "port", "p", 9324, "Port number")
	pflag.StringVarP(&cfg.Endpoint, "endpoint", "e", "metrics", "Metrics endpoint path")
	pflag.StringVarP(&cfg.Prefix, "prefix", "r", "ndocker", "Metric name prefix")
	pflag.StringVarP(&cfg.LogLevel, "log-level", "l", "info", "Log level: debug, info, warn, error")
	pflag.StringVarP(&cfg.LogPath, "log-path", "o", "", "Log file path (default stdout only)")
	pflag.StringVarP(&cfg.DockerHost, "docker-host", "d", "tcp://localhost:2375", "Docker daemon address")
	pflag.StringVarP(&cfg.OutputMode, "output", "u", "minimum", "Output mode: minimum (only ndocker_*) or all (include go_*, process_*, promhttp_*)")
	pflag.DurationVarP(&cfg.Timeout, "timeout", "t", 2*time.Second, "Timeout for Docker API requests")

	showVersion := pflag.BoolP("version", "v", false, "Show version information")

	pflag.Parse()

	if *showVersion {
		fmt.Printf("docker-exporter %s\n", Version)
		fmt.Printf("  Git Commit: %s\n", GitCommit)
		fmt.Printf("  Build Date: %s\n", BuildDate)
		fmt.Printf("  Go Version: %s\n", GoVersion)
		os.Exit(0)
	}

	// Normalize endpoint path
	cfg.Endpoint = strings.TrimPrefix(cfg.Endpoint, "/")
	cfg.LogLevel = strings.ToLower(cfg.LogLevel)
	cfg.OutputMode = strings.ToLower(cfg.OutputMode)

	// Validate output mode
	if cfg.OutputMode != "minimum" && cfg.OutputMode != "all" {
		cfg.OutputMode = "minimum"
	}

	return cfg, true
}

// Address returns the full bind address
func (c *Config) Address() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// MetricsPath returns the metrics endpoint path with leading slash
func (c *Config) MetricsPath() string {
	return "/" + c.Endpoint
}
