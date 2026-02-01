"""
Core monitoring functionality for OpenClaw Monitor.
"""

import time
import psutil
import logging
from typing import Dict, List, Any
from datetime import datetime


logger = logging.getLogger(__name__)


class Monitor:
    """Main monitoring class for collecting system metrics."""
    
    def __init__(self, config):
        """
        Initialize monitor.
        
        Args:
            config: Configuration object
        """
        self.config = config
        self.running = False
        self.metrics_history = []
        
    def collect_metrics(self) -> Dict[str, Any]:
        """
        Collect current system metrics.
        
        Returns:
            Dictionary of current metrics
        """
        metrics = {
            "timestamp": datetime.now().isoformat(),
        }
        
        enabled_metrics = self.config.get("monitor.metrics", [])
        
        if "cpu" in enabled_metrics:
            metrics["cpu_percent"] = psutil.cpu_percent(interval=1)
            metrics["cpu_count"] = psutil.cpu_count()
        
        if "memory" in enabled_metrics:
            mem = psutil.virtual_memory()
            metrics["memory_percent"] = mem.percent
            metrics["memory_total"] = mem.total
            metrics["memory_available"] = mem.available
            metrics["memory_used"] = mem.used
        
        if "disk" in enabled_metrics:
            disk = psutil.disk_usage('/')
            metrics["disk_percent"] = disk.percent
            metrics["disk_total"] = disk.total
            metrics["disk_used"] = disk.used
            metrics["disk_free"] = disk.free
        
        return metrics
    
    def check_thresholds(self, metrics: Dict[str, Any]) -> List[str]:
        """
        Check if any metrics exceed configured thresholds.
        
        Args:
            metrics: Current metrics dictionary
            
        Returns:
            List of warning messages
        """
        warnings = []
        
        cpu_threshold = self.config.get("thresholds.cpu_percent", 80.0)
        if "cpu_percent" in metrics and metrics["cpu_percent"] > cpu_threshold:
            warnings.append(f"CPU usage ({metrics['cpu_percent']:.1f}%) exceeds threshold ({cpu_threshold}%)")
        
        mem_threshold = self.config.get("thresholds.memory_percent", 80.0)
        if "memory_percent" in metrics and metrics["memory_percent"] > mem_threshold:
            warnings.append(f"Memory usage ({metrics['memory_percent']:.1f}%) exceeds threshold ({mem_threshold}%)")
        
        disk_threshold = self.config.get("thresholds.disk_percent", 90.0)
        if "disk_percent" in metrics and metrics["disk_percent"] > disk_threshold:
            warnings.append(f"Disk usage ({metrics['disk_percent']:.1f}%) exceeds threshold ({disk_threshold}%)")
        
        return warnings
    
    def format_metrics(self, metrics: Dict[str, Any]) -> str:
        """
        Format metrics for display.
        
        Args:
            metrics: Metrics dictionary
            
        Returns:
            Formatted string
        """
        lines = [f"Metrics at {metrics['timestamp']}:"]
        
        if "cpu_percent" in metrics:
            lines.append(f"  CPU: {metrics['cpu_percent']:.1f}% ({metrics.get('cpu_count', 'N/A')} cores)")
        
        if "memory_percent" in metrics:
            mem_gb = metrics.get('memory_used', 0) / (1024**3)
            mem_total_gb = metrics.get('memory_total', 0) / (1024**3)
            lines.append(f"  Memory: {metrics['memory_percent']:.1f}% ({mem_gb:.2f}GB / {mem_total_gb:.2f}GB)")
        
        if "disk_percent" in metrics:
            disk_gb = metrics.get('disk_used', 0) / (1024**3)
            disk_total_gb = metrics.get('disk_total', 0) / (1024**3)
            lines.append(f"  Disk: {metrics['disk_percent']:.1f}% ({disk_gb:.2f}GB / {disk_total_gb:.2f}GB)")
        
        return "\n".join(lines)
    
    def start(self) -> None:
        """Start monitoring loop."""
        self.running = True
        interval = self.config.get("monitor.interval", 5)
        
        logger.info("OpenClaw Monitor started")
        print("OpenClaw Monitor started. Press Ctrl+C to stop.")
        
        try:
            while self.running:
                # Collect metrics
                metrics = self.collect_metrics()
                self.metrics_history.append(metrics)
                
                # Keep only last 100 entries
                if len(self.metrics_history) > 100:
                    self.metrics_history.pop(0)
                
                # Display metrics
                print("\n" + "=" * 60)
                print(self.format_metrics(metrics))
                
                # Check thresholds
                warnings = self.check_thresholds(metrics)
                if warnings:
                    print("\n⚠️  WARNINGS:")
                    for warning in warnings:
                        print(f"  - {warning}")
                        logger.warning(warning)
                
                # Wait for next interval
                time.sleep(interval)
                
        except KeyboardInterrupt:
            print("\n\nStopping monitor...")
            self.stop()
    
    def stop(self) -> None:
        """Stop monitoring."""
        self.running = False
        logger.info("OpenClaw Monitor stopped")
        print("OpenClaw Monitor stopped.")
