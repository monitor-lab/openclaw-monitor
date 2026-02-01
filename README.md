# openclaw-monitor

A comprehensive monitoring tool for OpenClaw applications that tracks system metrics including CPU, memory, and disk usage.

## Features

- **Real-time Monitoring**: Track CPU, memory, and disk metrics in real-time
- **Threshold Alerts**: Configure thresholds and get warnings when limits are exceeded
- **Flexible Configuration**: YAML-based configuration for easy customization
- **Command-line Interface**: Simple CLI for starting monitoring and checking status
- **Logging**: Built-in logging to track metrics history and warnings

## Installation

```bash
# Clone the repository
git clone https://github.com/monitor-lab/openclaw-monitor.git
cd openclaw-monitor

# Install dependencies
pip install -r requirements.txt

# Install the package
pip install -e .
```

## Usage

### Start Monitoring

Start the monitor with default settings:

```bash
openclaw-monitor start
```

Start with custom configuration:

```bash
openclaw-monitor start --config config.yml
```

Start with custom interval:

```bash
openclaw-monitor start --interval 10
```

Enable verbose output:

```bash
openclaw-monitor start --verbose
```

### Check Current Status

Get a snapshot of current system metrics:

```bash
openclaw-monitor status
```

### Configuration

Create a configuration file based on the example:

```bash
openclaw-monitor config-example > config.yml
```

Or copy the example configuration:

```bash
cp config.example.yml config.yml
```

Edit the configuration to customize monitoring behavior:

```yaml
monitor:
  interval: 5  # Monitoring interval in seconds
  metrics:
    - cpu
    - memory
    - disk

logging:
  level: INFO  # DEBUG, INFO, WARNING, ERROR
  file: openclaw-monitor.log

thresholds:
  cpu_percent: 80.0
  memory_percent: 80.0
  disk_percent: 90.0
```

## Metrics Collected

- **CPU**: CPU usage percentage and core count
- **Memory**: Memory usage percentage, total, used, and available memory
- **Disk**: Disk usage percentage, total, used, and free space

## Requirements

- Python 3.7+
- psutil
- PyYAML
- click

## License

MIT License

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
