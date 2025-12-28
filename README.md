# Docker Exporter for Prometheus

A Prometheus exporter that collects metrics from Docker daemon via API.

## Features

- Container metrics (state, uptime, restart count, health status)
- Resource metrics (CPU, memory, network, block I/O)
- Docker engine metrics (version, container counts, image counts)
- Configurable metric prefix
- Remote Docker daemon support via TCP

## Installation

### Quick Install (Linux)

```bash
curl -sSL https://raw.githubusercontent.com/nhattuanbl/docker-exporter/main/setup.sh | bash
```

### Manual Download

Download the latest release from [GitHub Releases](https://github.com/nhattuanbl/docker-exporter/releases).

### Build from Source

```bash
git clone https://github.com/nhattuanbl/docker-exporter.git
cd docker-exporter
make build
```

### Command Line Options

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--host` | `-h` | `0.0.0.0` | Bind address |
| `--port` | `-p` | `9324` | Port number |
| `--endpoint` | `-e` | `metrics` | Metrics endpoint path |
| `--prefix` | `-r` | `ndocker` | Metric name prefix |
| `--log-level` | `-l` | `info` | Log level: debug, info, warn, error |
| `--log-path` | `-o` | stdout | Log file path |
| `--docker-host` | `-d` | `tcp://localhost:2375` | Docker daemon address |
| `--output` | `-u` | `minimum` | Output mode: `minimum` (only ndocker_*) or `all` (include go_*, process_*, promhttp_*) |
| `--timeout` | `-t` | `2s` | Timeout for Docker API requests |
| `--version` | `-v` | - | Show version information |

## Docker Configuration

To enable remote access to Docker daemon, configure `/etc/docker/daemon.json`:

```json
{
  "metrics-addr": "0.0.0.0:9323",
  "experimental": true,
  "hosts": ["tcp://0.0.0.0:2375", "unix:///var/run/docker.sock"],
  "iptables": false
}
```
> ⚠️ **Warning**: TCP without TLS is insecure. Use only in trusted networks or enable TLS.
```bash
chmod 755 /etc/docker/daemon.json
nano /usr/lib/systemd/system/docker.service
ExecStart=/usr/bin/dockerd
systemctl daemon-reload
sudo service docker restart
```

## Metrics

### Container Core Metrics

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `ndocker_container_info` | Gauge | id, name, image, state | Container metadata |
| `ndocker_container_state` | Gauge | id, name | State code (1=running, 2=paused, 3=restarting, 4=exited, 5=dead, 6=created) |
| `ndocker_container_uptime_seconds` | Gauge | id, name | Container uptime in seconds |
| `ndocker_container_created_seconds` | Gauge | id, name | Creation timestamp |
| `ndocker_container_started_seconds` | Gauge | id, name | Start timestamp |
| `ndocker_container_restart_count` | Gauge | id, name | Restart count |
| `ndocker_container_health_status` | Gauge | id, name | Health (1=healthy, 0=unhealthy, -1=none) |
| `ndocker_container_exit_code` | Gauge | id, name | Exit code |
| `ndocker_container_oom_killed` | Gauge | id, name | OOM killed flag |

### CPU Metrics

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `ndocker_container_cpu_usage_percent` | Gauge | id, name | Current CPU usage % |
| `ndocker_container_cpu_usage_seconds_total` | Counter | id, name | Total CPU time |

### Memory Metrics

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `ndocker_container_memory_usage_bytes` | Gauge | id, name | Current memory usage |
| `ndocker_container_memory_limit_bytes` | Gauge | id, name | Memory limit |
| `ndocker_container_memory_usage_percent` | Gauge | id, name | Memory usage % |

### Network Metrics

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `ndocker_container_network_rx_bytes_total` | Counter | id, name, interface | Bytes received |
| `ndocker_container_network_tx_bytes_total` | Counter | id, name, interface | Bytes transmitted |

### Block I/O Metrics

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `ndocker_container_blkio_read_bytes_total` | Counter | id, name | Bytes read |
| `ndocker_container_blkio_write_bytes_total` | Counter | id, name | Bytes written |

### Engine Metrics

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `ndocker_engine_info` | Gauge | version, os, arch, kernel | Docker engine info |
| `ndocker_containers_total` | Gauge | state | Container count by state |
| `ndocker_images_total` | Gauge | - | Total images |

### Exporter Metrics

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `ndocker_scrape_duration_seconds` | Gauge | - | Scrape duration |
| `ndocker_build_info` | Gauge | version, go_version | Build information |

## Prometheus Configuration

```yaml
scrape_configs:
  - job_name: 'docker'
    static_configs:
      - targets: ['localhost:9324']
```

## Endpoints

| Path | Description |
|------|-------------|
| `/` | Home page with links |
| `/metrics` | Prometheus metrics |
| `/health` | Health check (returns 200 if Docker is accessible) |

## License

MIT License
