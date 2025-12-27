package docker

import (
	"context"
	"encoding/json"
	"io"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/system"
	"github.com/docker/docker/client"
)

// ContainerInfo holds container information
type ContainerInfo struct {
	ID           string
	Name         string
	Image        string
	State        string
	Health       string
	Created      time.Time
	Started      time.Time
	Finished     time.Time
	RestartCount int
	ExitCode     int
	OOMKilled    bool
	Running      bool
}

// ContainerStats holds container resource statistics
type ContainerStats struct {
	ID   string
	Name string

	// CPU
	CPUPercent    float64
	CPUUsageTotal uint64
	CPUSystem     uint64

	// Memory
	MemoryUsage   uint64
	MemoryLimit   uint64
	MemoryPercent float64

	// Network (aggregated across all interfaces)
	NetworkRxBytes uint64
	NetworkTxBytes uint64
	Networks       map[string]NetworkStats

	// Block I/O
	BlockRead  uint64
	BlockWrite uint64

	// PIDs
	PidsCount uint64
}

// NetworkStats holds per-interface network statistics
type NetworkStats struct {
	RxBytes uint64
	TxBytes uint64
}

// EngineInfo holds Docker engine information
type EngineInfo struct {
	Version           string
	OS                string
	Arch              string
	KernelVersion     string
	ContainersRunning int
	ContainersPaused  int
	ContainersStopped int
	Images            int
	NCPU              int
	MemTotal          int64
}

// Client wraps the Docker client
type Client struct {
	cli *client.Client
}

// NewClient creates a new Docker client
func NewClient(host string) (*Client, error) {
	opts := []client.Opt{
		client.WithAPIVersionNegotiation(),
	}

	// Set host if provided
	if host != "" {
		opts = append(opts, client.WithHost(host))
	} else {
		opts = append(opts, client.FromEnv)
	}

	cli, err := client.NewClientWithOpts(opts...)
	if err != nil {
		return nil, err
	}

	return &Client{cli: cli}, nil
}

// Close closes the Docker client
func (c *Client) Close() error {
	return c.cli.Close()
}

// Ping checks if the Docker daemon is accessible
func (c *Client) Ping(ctx context.Context) error {
	_, err := c.cli.Ping(ctx)
	return err
}

// ListContainers returns a list of all containers
func (c *Client) ListContainers(ctx context.Context) ([]ContainerInfo, error) {
	containers, err := c.cli.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return nil, err
	}

	var result []ContainerInfo
	for _, cont := range containers {
		info, err := c.InspectContainer(ctx, cont.ID)
		if err != nil {
			continue
		}
		result = append(result, info)
	}

	return result, nil
}

// InspectContainer returns detailed information about a container
func (c *Client) InspectContainer(ctx context.Context, containerID string) (ContainerInfo, error) {
	inspect, err := c.cli.ContainerInspect(ctx, containerID)
	if err != nil {
		return ContainerInfo{}, err
	}

	info := ContainerInfo{
		ID:           inspect.ID[:12],
		Name:         strings.TrimPrefix(inspect.Name, "/"),
		Image:        inspect.Config.Image,
		State:        inspect.State.Status,
		RestartCount: inspect.RestartCount,
		ExitCode:     inspect.State.ExitCode,
		OOMKilled:    inspect.State.OOMKilled,
		Running:      inspect.State.Running,
	}

	// Parse timestamps
	if created, err := time.Parse(time.RFC3339Nano, inspect.Created); err == nil {
		info.Created = created
	}
	if started, err := time.Parse(time.RFC3339Nano, inspect.State.StartedAt); err == nil {
		info.Started = started
	}
	if finished, err := time.Parse(time.RFC3339Nano, inspect.State.FinishedAt); err == nil {
		info.Finished = finished
	}

	// Health status
	if inspect.State.Health != nil {
		info.Health = inspect.State.Health.Status
	} else {
		info.Health = "none"
	}

	return info, nil
}

// GetContainerStats returns resource statistics for a container
func (c *Client) GetContainerStats(ctx context.Context, containerID string, name string) (*ContainerStats, error) {
	resp, err := c.cli.ContainerStats(ctx, containerID, false)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var stats container.StatsResponse
	if err := json.Unmarshal(body, &stats); err != nil {
		return nil, err
	}

	result := &ContainerStats{
		ID:       containerID[:12],
		Name:     name,
		Networks: make(map[string]NetworkStats),
	}

	// Calculate CPU percentage
	result.CPUPercent = calculateCPUPercent(&stats)
	result.CPUUsageTotal = stats.CPUStats.CPUUsage.TotalUsage
	result.CPUSystem = stats.CPUStats.SystemUsage

	// Memory stats
	result.MemoryUsage = stats.MemoryStats.Usage
	result.MemoryLimit = stats.MemoryStats.Limit
	if result.MemoryLimit > 0 {
		result.MemoryPercent = float64(result.MemoryUsage) / float64(result.MemoryLimit) * 100.0
	}

	// Network stats
	for iface, netStats := range stats.Networks {
		result.Networks[iface] = NetworkStats{
			RxBytes: netStats.RxBytes,
			TxBytes: netStats.TxBytes,
		}
		result.NetworkRxBytes += netStats.RxBytes
		result.NetworkTxBytes += netStats.TxBytes
	}

	// Block I/O stats
	for _, bioEntry := range stats.BlkioStats.IoServiceBytesRecursive {
		switch bioEntry.Op {
		case "read", "Read":
			result.BlockRead += bioEntry.Value
		case "write", "Write":
			result.BlockWrite += bioEntry.Value
		}
	}

	// PIDs
	result.PidsCount = stats.PidsStats.Current

	return result, nil
}

// GetEngineInfo returns Docker engine information
func (c *Client) GetEngineInfo(ctx context.Context) (*EngineInfo, error) {
	info, err := c.cli.Info(ctx)
	if err != nil {
		return nil, err
	}

	return &EngineInfo{
		Version:           info.ServerVersion,
		OS:                info.OperatingSystem,
		Arch:              info.Architecture,
		KernelVersion:     info.KernelVersion,
		ContainersRunning: info.ContainersRunning,
		ContainersPaused:  info.ContainersPaused,
		ContainersStopped: info.ContainersStopped,
		Images:            info.Images,
		NCPU:              info.NCPU,
		MemTotal:          info.MemTotal,
	}, nil
}

// GetImages returns the list of images
func (c *Client) GetImages(ctx context.Context) ([]image.Summary, error) {
	return c.cli.ImageList(ctx, image.ListOptions{})
}

// Info returns system information
func (c *Client) Info(ctx context.Context) (system.Info, error) {
	return c.cli.Info(ctx)
}

// calculateCPUPercent calculates the CPU usage percentage
func calculateCPUPercent(stats *container.StatsResponse) float64 {
	cpuDelta := float64(stats.CPUStats.CPUUsage.TotalUsage - stats.PreCPUStats.CPUUsage.TotalUsage)
	systemDelta := float64(stats.CPUStats.SystemUsage - stats.PreCPUStats.SystemUsage)

	if systemDelta > 0 && cpuDelta > 0 {
		cpuCount := float64(stats.CPUStats.OnlineCPUs)
		if cpuCount == 0 {
			cpuCount = float64(len(stats.CPUStats.CPUUsage.PercpuUsage))
		}
		if cpuCount == 0 {
			cpuCount = 1
		}
		return (cpuDelta / systemDelta) * cpuCount * 100.0
	}
	return 0.0
}
