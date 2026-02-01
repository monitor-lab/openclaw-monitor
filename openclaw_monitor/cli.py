"""
Command-line interface for OpenClaw Monitor.
"""

import click
import logging
import sys
from openclaw_monitor.config import Config
from openclaw_monitor.monitor import Monitor


def setup_logging(level: str, log_file: str) -> None:
    """
    Setup logging configuration.
    
    Args:
        level: Logging level (DEBUG, INFO, WARNING, ERROR)
        log_file: Path to log file
    """
    numeric_level = getattr(logging, level.upper(), logging.INFO)
    
    logging.basicConfig(
        level=numeric_level,
        format='%(asctime)s - %(name)s - %(levelname)s - %(message)s',
        handlers=[
            logging.FileHandler(log_file),
            logging.StreamHandler(sys.stdout)
        ]
    )


@click.group()
@click.version_option(version="0.1.0")
def cli():
    """OpenClaw Monitor - A monitoring tool for OpenClaw applications."""
    pass


@cli.command()
@click.option('--config', '-c', type=click.Path(exists=True), help='Path to configuration file')
@click.option('--interval', '-i', type=int, help='Monitoring interval in seconds')
@click.option('--verbose', '-v', is_flag=True, help='Enable verbose output')
def start(config, interval, verbose):
    """Start monitoring OpenClaw system metrics."""
    
    # Load configuration
    cfg = Config(config)
    
    # Override with command-line options
    if interval:
        cfg.config["monitor"]["interval"] = interval
    
    # Setup logging
    log_level = "DEBUG" if verbose else cfg.get("logging.level", "INFO")
    log_file = cfg.get("logging.file", "openclaw-monitor.log")
    setup_logging(log_level, log_file)
    
    # Start monitor
    monitor = Monitor(cfg)
    monitor.start()


@cli.command()
@click.option('--config', '-c', type=click.Path(exists=True), help='Path to configuration file')
def status(config):
    """Show current system status."""
    
    # Load configuration
    cfg = Config(config)
    
    # Collect and display current metrics
    monitor = Monitor(cfg)
    metrics = monitor.collect_metrics()
    
    print("\n" + "=" * 60)
    print(monitor.format_metrics(metrics))
    
    # Check thresholds
    warnings = monitor.check_thresholds(metrics)
    if warnings:
        print("\n⚠️  WARNINGS:")
        for warning in warnings:
            print(f"  - {warning}")
    else:
        print("\n✓ All metrics within normal thresholds")
    
    print("=" * 60)


@cli.command()
def config_example():
    """Print example configuration file."""
    
    example = """# OpenClaw Monitor Configuration

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
"""
    
    print(example)


def main():
    """Main entry point."""
    cli()


if __name__ == '__main__':
    main()
