"""
Tests for configuration module.
"""

import unittest
import tempfile
import os
from openclaw_monitor.config import Config


class TestConfig(unittest.TestCase):
    """Test cases for Config class."""
    
    def test_default_config(self):
        """Test default configuration."""
        config = Config()
        
        # Check default values
        self.assertEqual(config.get("monitor.interval"), 5)
        self.assertEqual(config.get("logging.level"), "INFO")
        self.assertEqual(config.get("thresholds.cpu_percent"), 80.0)
    
    def test_get_with_default(self):
        """Test getting configuration with default value."""
        config = Config()
        
        # Non-existent key should return default
        self.assertEqual(config.get("nonexistent.key", "default"), "default")
    
    def test_load_from_file(self):
        """Test loading configuration from file."""
        # Create temporary config file
        config_content = """
monitor:
  interval: 10
  metrics:
    - cpu
thresholds:
  cpu_percent: 90.0
"""
        
        with tempfile.NamedTemporaryFile(mode='w', suffix='.yml', delete=False) as f:
            f.write(config_content)
            temp_path = f.name
        
        try:
            config = Config(temp_path)
            
            # Check loaded values
            self.assertEqual(config.get("monitor.interval"), 10)
            self.assertEqual(config.get("thresholds.cpu_percent"), 90.0)
            
            # Check that defaults are still present for unspecified values
            self.assertEqual(config.get("logging.level"), "INFO")
        finally:
            os.unlink(temp_path)
    
    def test_nested_get(self):
        """Test getting nested configuration values."""
        config = Config()
        
        # Test nested access
        self.assertIsInstance(config.get("monitor.metrics"), list)
        self.assertIn("cpu", config.get("monitor.metrics"))


if __name__ == '__main__':
    unittest.main()
