package collector

import (
	"context"
	"sync"
	"time"

	"github.com/nhattuanbl/docker-exporter/internal/config"
	"github.com/nhattuanbl/docker-exporter/internal/docker"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

// Collector implements prometheus.Collector interface
type Collector struct {
	client  *docker.Client
	prefix  string
	logger  *zap.Logger
	timeout time.Duration

	// Container metrics
	containerInfo         *prometheus.Desc
	containerState        *prometheus.Desc
	containerUptime       *prometheus.Desc
	containerCreated      *prometheus.Desc
	containerStarted      *prometheus.Desc
	containerRestartCount *prometheus.Desc
	containerHealthStatus *prometheus.Desc
	containerExitCode     *prometheus.Desc
	containerOOMKilled    *prometheus.Desc

	// CPU metrics
	containerCPUPercent      *prometheus.Desc
	containerCPUUsageSeconds *prometheus.Desc

	// Memory metrics
	containerMemoryUsage   *prometheus.Desc
	containerMemoryLimit   *prometheus.Desc
	containerMemoryPercent *prometheus.Desc

	// Network metrics
	containerNetworkRxBytes *prometheus.Desc
	containerNetworkTxBytes *prometheus.Desc

	// Block I/O metrics
	containerBlkioReadBytes  *prometheus.Desc
	containerBlkioWriteBytes *prometheus.Desc

	// Engine metrics
	engineInfo      *prometheus.Desc
	containersTotal *prometheus.Desc
	imagesTotal     *prometheus.Desc

	// Exporter metrics
	scrapeDuration *prometheus.Desc
	buildInfo      *prometheus.Desc
}

// NewCollector creates a new Collector
func NewCollector(client *docker.Client, cfg *config.Config, logger *zap.Logger) *Collector {
	prefix := cfg.Prefix

	return &Collector{
		client:  client,
		prefix:  prefix,
		logger:  logger,
		timeout: cfg.Timeout,

		// Container core metrics
		containerInfo: prometheus.NewDesc(
			prefix+"_container_info",
			"Container information",
			[]string{"id", "name", "image", "state"}, nil,
		),
		containerState: prometheus.NewDesc(
			prefix+"_container_state",
			"Container state (1=running, 2=paused, 3=restarting, 4=exited, 5=dead, 6=created)",
			[]string{"id", "name"}, nil,
		),
		containerUptime: prometheus.NewDesc(
			prefix+"_container_uptime_seconds",
			"Container uptime in seconds",
			[]string{"id", "name"}, nil,
		),
		containerCreated: prometheus.NewDesc(
			prefix+"_container_created_seconds",
			"Container creation timestamp",
			[]string{"id", "name"}, nil,
		),
		containerStarted: prometheus.NewDesc(
			prefix+"_container_started_seconds",
			"Container start timestamp",
			[]string{"id", "name"}, nil,
		),
		containerRestartCount: prometheus.NewDesc(
			prefix+"_container_restart_count",
			"Container restart count",
			[]string{"id", "name"}, nil,
		),
		containerHealthStatus: prometheus.NewDesc(
			prefix+"_container_health_status",
			"Container health status (1=healthy, 0=unhealthy, -1=none)",
			[]string{"id", "name"}, nil,
		),
		containerExitCode: prometheus.NewDesc(
			prefix+"_container_exit_code",
			"Container exit code",
			[]string{"id", "name"}, nil,
		),
		containerOOMKilled: prometheus.NewDesc(
			prefix+"_container_oom_killed",
			"Container OOM killed (1=true, 0=false)",
			[]string{"id", "name"}, nil,
		),

		// CPU metrics
		containerCPUPercent: prometheus.NewDesc(
			prefix+"_container_cpu_usage_percent",
			"Container CPU usage percentage",
			[]string{"id", "name"}, nil,
		),
		containerCPUUsageSeconds: prometheus.NewDesc(
			prefix+"_container_cpu_usage_seconds_total",
			"Container total CPU usage in seconds",
			[]string{"id", "name"}, nil,
		),

		// Memory metrics
		containerMemoryUsage: prometheus.NewDesc(
			prefix+"_container_memory_usage_bytes",
			"Container memory usage in bytes",
			[]string{"id", "name"}, nil,
		),
		containerMemoryLimit: prometheus.NewDesc(
			prefix+"_container_memory_limit_bytes",
			"Container memory limit in bytes",
			[]string{"id", "name"}, nil,
		),
		containerMemoryPercent: prometheus.NewDesc(
			prefix+"_container_memory_usage_percent",
			"Container memory usage percentage",
			[]string{"id", "name"}, nil,
		),

		// Network metrics
		containerNetworkRxBytes: prometheus.NewDesc(
			prefix+"_container_network_rx_bytes_total",
			"Container network bytes received",
			[]string{"id", "name", "interface"}, nil,
		),
		containerNetworkTxBytes: prometheus.NewDesc(
			prefix+"_container_network_tx_bytes_total",
			"Container network bytes transmitted",
			[]string{"id", "name", "interface"}, nil,
		),

		// Block I/O metrics
		containerBlkioReadBytes: prometheus.NewDesc(
			prefix+"_container_blkio_read_bytes_total",
			"Container block I/O bytes read",
			[]string{"id", "name"}, nil,
		),
		containerBlkioWriteBytes: prometheus.NewDesc(
			prefix+"_container_blkio_write_bytes_total",
			"Container block I/O bytes written",
			[]string{"id", "name"}, nil,
		),

		// Engine metrics
		engineInfo: prometheus.NewDesc(
			prefix+"_engine_info",
			"Docker engine information",
			[]string{"version", "os", "arch", "kernel"}, nil,
		),
		containersTotal: prometheus.NewDesc(
			prefix+"_containers_total",
			"Total number of containers by state",
			[]string{"state"}, nil,
		),
		imagesTotal: prometheus.NewDesc(
			prefix+"_images_total",
			"Total number of images",
			nil, nil,
		),

		// Exporter metrics
		scrapeDuration: prometheus.NewDesc(
			prefix+"_scrape_duration_seconds",
			"Duration of the scrape",
			nil, nil,
		),
		buildInfo: prometheus.NewDesc(
			prefix+"_build_info",
			"Exporter build information",
			[]string{"version", "go_version"}, nil,
		),
	}
}

// Describe implements prometheus.Collector
func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.containerInfo
	ch <- c.containerState
	ch <- c.containerUptime
	ch <- c.containerCreated
	ch <- c.containerStarted
	ch <- c.containerRestartCount
	ch <- c.containerHealthStatus
	ch <- c.containerExitCode
	ch <- c.containerOOMKilled
	ch <- c.containerCPUPercent
	ch <- c.containerCPUUsageSeconds
	ch <- c.containerMemoryUsage
	ch <- c.containerMemoryLimit
	ch <- c.containerMemoryPercent
	ch <- c.containerNetworkRxBytes
	ch <- c.containerNetworkTxBytes
	ch <- c.containerBlkioReadBytes
	ch <- c.containerBlkioWriteBytes
	ch <- c.engineInfo
	ch <- c.containersTotal
	ch <- c.imagesTotal
	ch <- c.scrapeDuration
	ch <- c.buildInfo
}

// Collect implements prometheus.Collector
func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	start := time.Now()

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	c.logger.Debug("[COLLECTOR] Starting metrics collection",
		zap.String("version", config.Version),
		zap.Duration("timeout", c.timeout))

	// Build info
	ch <- prometheus.MustNewConstMetric(
		c.buildInfo, prometheus.GaugeValue, 1,
		config.Version, config.GoVersion,
	)

	// Collect container metrics
	c.collectContainerMetrics(ctx, ch)

	// Collect engine metrics
	c.collectEngineMetrics(ctx, ch)

	// Scrape duration
	duration := time.Since(start).Seconds()
	ch <- prometheus.MustNewConstMetric(
		c.scrapeDuration, prometheus.GaugeValue, duration,
	)

	c.logger.Debug("[COLLECTOR] Metrics collection completed", zap.Float64("duration_seconds", duration))
}

// collectContainerMetrics collects metrics for all containers
func (c *Collector) collectContainerMetrics(ctx context.Context, ch chan<- prometheus.Metric) {
	c.logger.Debug("[STEP 1/4] Fetching container list from Docker API...")

	containers, err := c.client.ListContainers(ctx)
	if err != nil {
		c.logger.Error("[ERROR] Failed to list containers from Docker API", zap.Error(err))
		return
	}

	if len(containers) == 0 {
		c.logger.Debug("[STEP 2/4] No containers found - Docker returned empty list")
		return
	}

	c.logger.Debug("[STEP 2/4] Containers retrieved successfully",
		zap.Int("total_count", len(containers)))

	// Count running containers
	runningCount := 0
	for _, cont := range containers {
		if cont.Running {
			runningCount++
		}
	}
	c.logger.Debug("[STEP 2/4] Container breakdown",
		zap.Int("running", runningCount),
		zap.Int("stopped", len(containers)-runningCount))

	var wg sync.WaitGroup
	statsChan := make(chan *docker.ContainerStats, len(containers))

	// Collect stats for running containers concurrently
	c.logger.Debug("[STEP 3/4] Collecting stats for running containers...")
	for _, cont := range containers {
		c.logger.Debug("[CONTAINER] Processing",
			zap.String("id", cont.ID),
			zap.String("name", cont.Name),
			zap.String("image", cont.Image),
			zap.String("state", cont.State),
			zap.Bool("running", cont.Running),
			zap.String("health", cont.Health))

		if cont.Running {
			wg.Add(1)
			go func(container docker.ContainerInfo) {
				defer wg.Done()
				c.logger.Debug("[STATS] Fetching stats for container",
					zap.String("name", container.Name))
				stats, err := c.client.GetContainerStats(ctx, container.ID, container.Name)
				if err != nil {
					c.logger.Error("[ERROR] Failed to get container stats",
						zap.String("container", container.Name),
						zap.String("id", container.ID),
						zap.Error(err))
					return
				}
				c.logger.Debug("[STATS] Stats retrieved successfully",
					zap.String("name", container.Name),
					zap.Float64("cpu_percent", stats.CPUPercent),
					zap.Uint64("memory_usage", stats.MemoryUsage))
				statsChan <- stats
			}(cont)
		}
	}

	// Close channel when all goroutines complete
	go func() {
		wg.Wait()
		close(statsChan)
	}()

	// Collect stats from channel
	c.logger.Debug("[STEP 4/4] Emitting Prometheus metrics...")
	statsMap := make(map[string]*docker.ContainerStats)
	for stats := range statsChan {
		statsMap[stats.ID] = stats
	}

	// Emit metrics for each container
	for _, cont := range containers {
		// Container info
		ch <- prometheus.MustNewConstMetric(
			c.containerInfo, prometheus.GaugeValue, 1,
			cont.ID, cont.Name, cont.Image, cont.State,
		)

		// Container state
		stateCode := stateToCode(cont.State)
		ch <- prometheus.MustNewConstMetric(
			c.containerState, prometheus.GaugeValue, float64(stateCode),
			cont.ID, cont.Name,
		)

		// Container uptime
		var uptime float64
		if cont.Running && !cont.Started.IsZero() {
			uptime = time.Since(cont.Started).Seconds()
		}
		ch <- prometheus.MustNewConstMetric(
			c.containerUptime, prometheus.GaugeValue, uptime,
			cont.ID, cont.Name,
		)

		// Container created timestamp
		if !cont.Created.IsZero() {
			ch <- prometheus.MustNewConstMetric(
				c.containerCreated, prometheus.GaugeValue, float64(cont.Created.Unix()),
				cont.ID, cont.Name,
			)
		}

		// Container started timestamp
		if !cont.Started.IsZero() {
			ch <- prometheus.MustNewConstMetric(
				c.containerStarted, prometheus.GaugeValue, float64(cont.Started.Unix()),
				cont.ID, cont.Name,
			)
		}

		// Restart count
		ch <- prometheus.MustNewConstMetric(
			c.containerRestartCount, prometheus.GaugeValue, float64(cont.RestartCount),
			cont.ID, cont.Name,
		)

		// Health status
		healthCode := healthToCode(cont.Health)
		ch <- prometheus.MustNewConstMetric(
			c.containerHealthStatus, prometheus.GaugeValue, float64(healthCode),
			cont.ID, cont.Name,
		)

		// Exit code
		ch <- prometheus.MustNewConstMetric(
			c.containerExitCode, prometheus.GaugeValue, float64(cont.ExitCode),
			cont.ID, cont.Name,
		)

		// OOM killed
		var oomKilled float64
		if cont.OOMKilled {
			oomKilled = 1
		}
		ch <- prometheus.MustNewConstMetric(
			c.containerOOMKilled, prometheus.GaugeValue, oomKilled,
			cont.ID, cont.Name,
		)

		// Resource metrics (only for running containers)
		if stats, ok := statsMap[cont.ID]; ok {
			// CPU
			ch <- prometheus.MustNewConstMetric(
				c.containerCPUPercent, prometheus.GaugeValue, stats.CPUPercent,
				cont.ID, cont.Name,
			)
			ch <- prometheus.MustNewConstMetric(
				c.containerCPUUsageSeconds, prometheus.CounterValue,
				float64(stats.CPUUsageTotal)/1e9, // nanoseconds to seconds
				cont.ID, cont.Name,
			)

			// Memory
			ch <- prometheus.MustNewConstMetric(
				c.containerMemoryUsage, prometheus.GaugeValue, float64(stats.MemoryUsage),
				cont.ID, cont.Name,
			)
			ch <- prometheus.MustNewConstMetric(
				c.containerMemoryLimit, prometheus.GaugeValue, float64(stats.MemoryLimit),
				cont.ID, cont.Name,
			)
			ch <- prometheus.MustNewConstMetric(
				c.containerMemoryPercent, prometheus.GaugeValue, stats.MemoryPercent,
				cont.ID, cont.Name,
			)

			// Network (per interface)
			for iface, netStats := range stats.Networks {
				ch <- prometheus.MustNewConstMetric(
					c.containerNetworkRxBytes, prometheus.CounterValue, float64(netStats.RxBytes),
					cont.ID, cont.Name, iface,
				)
				ch <- prometheus.MustNewConstMetric(
					c.containerNetworkTxBytes, prometheus.CounterValue, float64(netStats.TxBytes),
					cont.ID, cont.Name, iface,
				)
			}

			// Block I/O
			ch <- prometheus.MustNewConstMetric(
				c.containerBlkioReadBytes, prometheus.CounterValue, float64(stats.BlockRead),
				cont.ID, cont.Name,
			)
			ch <- prometheus.MustNewConstMetric(
				c.containerBlkioWriteBytes, prometheus.CounterValue, float64(stats.BlockWrite),
				cont.ID, cont.Name,
			)
		}
	}
}

// collectEngineMetrics collects Docker engine metrics
func (c *Collector) collectEngineMetrics(ctx context.Context, ch chan<- prometheus.Metric) {
	info, err := c.client.GetEngineInfo(ctx)
	if err != nil {
		c.logger.Error("Failed to get engine info", zap.Error(err))
		return
	}

	// Engine info
	ch <- prometheus.MustNewConstMetric(
		c.engineInfo, prometheus.GaugeValue, 1,
		info.Version, info.OS, info.Arch, info.KernelVersion,
	)

	// Containers total by state
	ch <- prometheus.MustNewConstMetric(
		c.containersTotal, prometheus.GaugeValue, float64(info.ContainersRunning),
		"running",
	)
	ch <- prometheus.MustNewConstMetric(
		c.containersTotal, prometheus.GaugeValue, float64(info.ContainersPaused),
		"paused",
	)
	ch <- prometheus.MustNewConstMetric(
		c.containersTotal, prometheus.GaugeValue, float64(info.ContainersStopped),
		"stopped",
	)

	// Images total
	ch <- prometheus.MustNewConstMetric(
		c.imagesTotal, prometheus.GaugeValue, float64(info.Images),
	)
}

// stateToCode converts container state string to numeric code
func stateToCode(state string) int {
	switch state {
	case "running":
		return 1
	case "paused":
		return 2
	case "restarting":
		return 3
	case "exited":
		return 4
	case "dead":
		return 5
	case "created":
		return 6
	default:
		return 0
	}
}

// healthToCode converts health status string to numeric code
func healthToCode(health string) int {
	switch health {
	case "healthy":
		return 1
	case "unhealthy":
		return 0
	case "starting":
		return 2
	default:
		return -1
	}
}
