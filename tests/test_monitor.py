"""
Tests for monitor module.
"""

import unittest
from unittest.mock import Mock, patch
from openclaw_monitor.config import Config
from openclaw_monitor.monitor import Monitor


class TestMonitor(unittest.TestCase):
    """Test cases for Monitor class."""
    
    def setUp(self):
        """Set up test fixtures."""
        self.config = Config()
        self.monitor = Monitor(self.config)
    
    def test_initialization(self):
        """Test monitor initialization."""
        self.assertIsNotNone(self.monitor.config)
        self.assertFalse(self.monitor.running)
        self.assertEqual(len(self.monitor.metrics_history), 0)
    
    @patch('openclaw_monitor.monitor.psutil')
    def test_collect_metrics_cpu(self, mock_psutil):
        """Test collecting CPU metrics."""
        mock_psutil.cpu_percent.return_value = 50.0
        mock_psutil.cpu_count.return_value = 4
        
        # Configure to collect only CPU
        self.config.config["monitor"]["metrics"] = ["cpu"]
        
        metrics = self.monitor.collect_metrics()
        
        self.assertIn("cpu_percent", metrics)
        self.assertIn("cpu_count", metrics)
        self.assertEqual(metrics["cpu_percent"], 50.0)
        self.assertEqual(metrics["cpu_count"], 4)
    
    @patch('openclaw_monitor.monitor.psutil')
    def test_collect_metrics_memory(self, mock_psutil):
        """Test collecting memory metrics."""
        mock_mem = Mock()
        mock_mem.percent = 60.0
        mock_mem.total = 8 * 1024**3  # 8GB
        mock_mem.used = 4 * 1024**3   # 4GB
        mock_mem.available = 4 * 1024**3  # 4GB
        mock_psutil.virtual_memory.return_value = mock_mem
        
        # Configure to collect only memory
        self.config.config["monitor"]["metrics"] = ["memory"]
        
        metrics = self.monitor.collect_metrics()
        
        self.assertIn("memory_percent", metrics)
        self.assertIn("memory_total", metrics)
        self.assertEqual(metrics["memory_percent"], 60.0)
    
    def test_check_thresholds_no_warnings(self):
        """Test threshold checking with normal values."""
        metrics = {
            "cpu_percent": 50.0,
            "memory_percent": 60.0,
            "disk_percent": 70.0,
        }
        
        warnings = self.monitor.check_thresholds(metrics)
        
        self.assertEqual(len(warnings), 0)
    
    def test_check_thresholds_with_warnings(self):
        """Test threshold checking with high values."""
        metrics = {
            "cpu_percent": 85.0,
            "memory_percent": 85.0,
            "disk_percent": 95.0,
        }
        
        warnings = self.monitor.check_thresholds(metrics)
        
        # Should have warnings for all three metrics
        self.assertEqual(len(warnings), 3)
        self.assertTrue(any("CPU" in w for w in warnings))
        self.assertTrue(any("Memory" in w for w in warnings))
        self.assertTrue(any("Disk" in w for w in warnings))
    
    def test_format_metrics(self):
        """Test metrics formatting."""
        metrics = {
            "timestamp": "2026-02-01T00:00:00",
            "cpu_percent": 50.0,
            "cpu_count": 4,
            "memory_percent": 60.0,
            "memory_total": 8 * 1024**3,
            "memory_used": 4 * 1024**3,
        }
        
        formatted = self.monitor.format_metrics(metrics)
        
        self.assertIn("CPU: 50.0%", formatted)
        self.assertIn("Memory: 60.0%", formatted)
        self.assertIn("2026-02-01T00:00:00", formatted)


if __name__ == '__main__':
    unittest.main()
