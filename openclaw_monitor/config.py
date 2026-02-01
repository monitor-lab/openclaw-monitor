"""
Configuration management for OpenClaw Monitor.
"""

import os
import yaml
from typing import Dict, Any, Optional


class Config:
    """Configuration manager for OpenClaw Monitor."""
    
    DEFAULT_CONFIG = {
        "monitor": {
            "interval": 5,  # seconds
            "metrics": ["cpu", "memory", "disk"],
        },
        "logging": {
            "level": "INFO",
            "file": "openclaw-monitor.log",
        },
        "thresholds": {
            "cpu_percent": 80.0,
            "memory_percent": 80.0,
            "disk_percent": 90.0,
        },
    }
    
    def __init__(self, config_path: Optional[str] = None):
        """
        Initialize configuration.
        
        Args:
            config_path: Path to configuration file (YAML)
        """
        self.config = self.DEFAULT_CONFIG.copy()
        
        if config_path and os.path.exists(config_path):
            self.load_from_file(config_path)
    
    def load_from_file(self, path: str) -> None:
        """
        Load configuration from YAML file.
        
        Args:
            path: Path to YAML configuration file
        """
        with open(path, 'r') as f:
            user_config = yaml.safe_load(f)
            if user_config:
                self._merge_config(user_config)
    
    def _merge_config(self, user_config: Dict[str, Any]) -> None:
        """
        Merge user configuration with defaults.
        
        Args:
            user_config: User-provided configuration dictionary
        """
        for key, value in user_config.items():
            if key in self.config and isinstance(self.config[key], dict) and isinstance(value, dict):
                self.config[key].update(value)
            else:
                self.config[key] = value
    
    def get(self, key: str, default: Any = None) -> Any:
        """
        Get configuration value by dot-notation key.
        
        Args:
            key: Configuration key (e.g., 'monitor.interval')
            default: Default value if key not found
            
        Returns:
            Configuration value
        """
        keys = key.split('.')
        value = self.config
        
        for k in keys:
            if isinstance(value, dict) and k in value:
                value = value[k]
            else:
                return default
        
        return value
