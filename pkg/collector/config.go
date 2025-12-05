// Package collector provides the main collection logic for Aztec node metrics
package collector

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// Config represents the collector configuration
type Config struct {
	Protocol   ProtocolConfig   `yaml:"protocol" mapstructure:"protocol"`
	Interval   time.Duration    `yaml:"interval" mapstructure:"interval"`
	API        APIConfig        `yaml:"api" mapstructure:"api"`
	Health     HealthConfig     `yaml:"health" mapstructure:"health"`
	JSONLog    JSONLogConfig    `yaml:"jsonlog" mapstructure:"jsonlog"`
	RefService RefServiceConfig `yaml:"ref_service" mapstructure:"ref_service"`
}

// ProtocolConfig represents protocol-specific configuration
type ProtocolConfig struct {
	Name     string `yaml:"name" mapstructure:"name"`
	RPCURL   string `yaml:"rpc_url" mapstructure:"rpc_url"`
	AdminURL string `yaml:"admin_url,omitempty" mapstructure:"admin_url"`
}

// APIConfig represents the API server configuration
type APIConfig struct {
	Enabled bool   `yaml:"enabled" mapstructure:"enabled"`
	Listen  string `yaml:"listen" mapstructure:"listen"`
}

// HealthConfig represents health check configuration
type HealthConfig struct {
	MinPeers        int   `yaml:"min_peers" mapstructure:"min_peers"`
	MaxBlocksBehind int64 `yaml:"max_blocks_behind" mapstructure:"max_blocks_behind"`
	MaxBlocksAhead  int64 `yaml:"max_blocks_ahead" mapstructure:"max_blocks_ahead"`
}

// JSONLogConfig represents JSON logging configuration
type JSONLogConfig struct {
	Path string `yaml:"path" mapstructure:"path"`
}

// RefServiceConfig represents external reference service configuration
type RefServiceConfig struct {
	Enabled  bool          `yaml:"enabled" mapstructure:"enabled"`
	URL      string        `yaml:"url" mapstructure:"url"`
	Protocol string        `yaml:"protocol" mapstructure:"protocol"`
	Network  string        `yaml:"network" mapstructure:"network"`
	Timeout  time.Duration `yaml:"timeout" mapstructure:"timeout"`
	// UseExternalRPC allows using another Aztec node as reference
	UseExternalRPC bool   `yaml:"use_external_rpc" mapstructure:"use_external_rpc"`
	ExternalRPCURL string `yaml:"external_rpc_url" mapstructure:"external_rpc_url"`
}

// DefaultConfig returns a configuration with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		Protocol: ProtocolConfig{
			Name:     "aztec",
			RPCURL:   "http://localhost:8080",
			AdminURL: "http://localhost:8880",
		},
		Interval: 30 * time.Second,
		API: APIConfig{
			Enabled: true,
			Listen:  "localhost:13285",
		},
		Health: HealthConfig{
			MinPeers:        1,
			MaxBlocksBehind: 100,
			MaxBlocksAhead:  10,
		},
		JSONLog: JSONLogConfig{
			Path: "./aztec-collector.json",
		},
		RefService: RefServiceConfig{
			Enabled:        false,
			URL:            "",
			Protocol:       "aztec",
			Network:        "mainnet",
			Timeout:        5 * time.Second,
			UseExternalRPC: false,
			ExternalRPCURL: "",
		},
	}
}

// LoadConfig loads configuration from a file and environment variables
// Environment variables override file configuration
// Environment variable format: AZTEC_COLLECTOR_<SECTION>_<KEY>
// Example: AZTEC_COLLECTOR_PROTOCOL_RPC_URL overrides protocol.rpc_url
func LoadConfig(configPath string) (*Config, error) {
	config := DefaultConfig()

	// Initialize viper
	v := viper.New()

	// Set config name and paths
	v.SetConfigType("yaml")

	// Set default values in viper
	setViperDefaults(v)

	// Enable environment variable reading
	v.SetEnvPrefix("AZTEC_COLLECTOR")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	v.AutomaticEnv()

	// Try to find and read config file
	path := findConfigFile(configPath)
	if path != "" {
		v.SetConfigFile(path)
		if err := v.ReadInConfig(); err != nil {
			// Only return error if file was explicitly specified
			if configPath != "" {
				return nil, fmt.Errorf("failed to read config file: %w", err)
			}
			// Otherwise just log a warning and continue with defaults + env vars
			fmt.Printf("Warning: Could not read config file: %v\n", err)
		}
	}

	// Unmarshal into config struct
	if err := v.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Handle environment variable overrides explicitly for nested configs
	applyEnvOverrides(config)

	// Validate required fields
	if config.Protocol.RPCURL == "" {
		return nil, fmt.Errorf("protocol.rpc_url is required")
	}

	return config, nil
}

// setViperDefaults sets default values in viper
func setViperDefaults(v *viper.Viper) {
	// Protocol defaults
	v.SetDefault("protocol.name", "aztec")
	v.SetDefault("protocol.rpc_url", "http://localhost:8080")
	v.SetDefault("protocol.admin_url", "http://localhost:8880")

	// Interval
	v.SetDefault("interval", "30s")

	// API defaults
	v.SetDefault("api.enabled", true)
	v.SetDefault("api.listen", "localhost:13285")

	// Health defaults
	v.SetDefault("health.min_peers", 1)
	v.SetDefault("health.max_blocks_behind", 100)
	v.SetDefault("health.max_blocks_ahead", 10)

	// JSON log defaults
	v.SetDefault("jsonlog.path", "./aztec-collector.json")

	// Reference service defaults
	v.SetDefault("ref_service.enabled", false)
	v.SetDefault("ref_service.url", "")
	v.SetDefault("ref_service.protocol", "aztec")
	v.SetDefault("ref_service.network", "mainnet")
	v.SetDefault("ref_service.timeout", "5s")
	v.SetDefault("ref_service.use_external_rpc", false)
	v.SetDefault("ref_service.external_rpc_url", "")
}

// applyEnvOverrides applies environment variable overrides explicitly
func applyEnvOverrides(config *Config) {
	// Protocol overrides
	if v := os.Getenv("AZTEC_COLLECTOR_PROTOCOL_NAME"); v != "" {
		config.Protocol.Name = v
	}
	if v := os.Getenv("AZTEC_COLLECTOR_PROTOCOL_RPC_URL"); v != "" {
		config.Protocol.RPCURL = v
	}
	if v := os.Getenv("AZTEC_COLLECTOR_PROTOCOL_ADMIN_URL"); v != "" {
		config.Protocol.AdminURL = v
	}

	// Interval override
	if v := os.Getenv("AZTEC_COLLECTOR_INTERVAL"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			config.Interval = d
		}
	}

	// API overrides
	if v := os.Getenv("AZTEC_COLLECTOR_API_ENABLED"); v != "" {
		config.API.Enabled = strings.ToLower(v) == "true" || v == "1"
	}
	if v := os.Getenv("AZTEC_COLLECTOR_API_LISTEN"); v != "" {
		config.API.Listen = v
	}

	// Health overrides
	if v := os.Getenv("AZTEC_COLLECTOR_HEALTH_MIN_PEERS"); v != "" {
		if n, err := parseInt(v); err == nil {
			config.Health.MinPeers = n
		}
	}
	if v := os.Getenv("AZTEC_COLLECTOR_HEALTH_MAX_BLOCKS_BEHIND"); v != "" {
		if n, err := parseInt64(v); err == nil {
			config.Health.MaxBlocksBehind = n
		}
	}
	if v := os.Getenv("AZTEC_COLLECTOR_HEALTH_MAX_BLOCKS_AHEAD"); v != "" {
		if n, err := parseInt64(v); err == nil {
			config.Health.MaxBlocksAhead = n
		}
	}

	// JSON log overrides
	if v := os.Getenv("AZTEC_COLLECTOR_JSONLOG_PATH"); v != "" {
		config.JSONLog.Path = v
	}

	// Reference service overrides
	if v := os.Getenv("AZTEC_COLLECTOR_REF_SERVICE_ENABLED"); v != "" {
		config.RefService.Enabled = strings.ToLower(v) == "true" || v == "1"
	}
	if v := os.Getenv("AZTEC_COLLECTOR_REF_SERVICE_URL"); v != "" {
		config.RefService.URL = v
	}
	if v := os.Getenv("AZTEC_COLLECTOR_REF_SERVICE_PROTOCOL"); v != "" {
		config.RefService.Protocol = v
	}
	if v := os.Getenv("AZTEC_COLLECTOR_REF_SERVICE_NETWORK"); v != "" {
		config.RefService.Network = v
	}
	if v := os.Getenv("AZTEC_COLLECTOR_REF_SERVICE_TIMEOUT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			config.RefService.Timeout = d
		}
	}
	if v := os.Getenv("AZTEC_COLLECTOR_REF_SERVICE_USE_EXTERNAL_RPC"); v != "" {
		config.RefService.UseExternalRPC = strings.ToLower(v) == "true" || v == "1"
	}
	if v := os.Getenv("AZTEC_COLLECTOR_REF_SERVICE_EXTERNAL_RPC_URL"); v != "" {
		config.RefService.ExternalRPCURL = v
	}
}

// parseInt parses a string to int
func parseInt(s string) (int, error) {
	var n int
	_, err := fmt.Sscanf(s, "%d", &n)
	return n, err
}

// parseInt64 parses a string to int64
func parseInt64(s string) (int64, error) {
	var n int64
	_, err := fmt.Sscanf(s, "%d", &n)
	return n, err
}

// findConfigFile attempts to locate a config file
func findConfigFile(configPath string) string {
	// Check if specific path provided and exists
	if configPath != "" {
		if info, err := os.Stat(configPath); err == nil && !info.IsDir() {
			return configPath
		}
	}

	// Look for common config file names
	commonNames := []string{
		"aztec-config.yaml",
		"aztec-config.yml",
		"config.yaml",
		"config.yml",
		"collector-config.yaml",
		"collector-config.yml",
	}

	// Search in config path directory and current directory
	searchDirs := []string{"."}
	if configPath != "" {
		searchDirs = append([]string{configPath}, searchDirs...)
	}

	for _, dir := range searchDirs {
		for _, name := range commonNames {
			path := filepath.Join(dir, name)
			if _, err := os.Stat(path); err == nil {
				return path
			}
		}
	}

	return ""
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Protocol.RPCURL == "" {
		return fmt.Errorf("protocol.rpc_url is required")
	}
	if c.Interval < time.Second {
		return fmt.Errorf("interval must be at least 1 second")
	}
	if c.RefService.Enabled && !c.RefService.UseExternalRPC && c.RefService.URL == "" {
		return fmt.Errorf("ref_service.url is required when ref_service.enabled is true and use_external_rpc is false")
	}
	if c.RefService.UseExternalRPC && c.RefService.ExternalRPCURL == "" {
		return fmt.Errorf("ref_service.external_rpc_url is required when use_external_rpc is true")
	}
	return nil
}

// ToYAML converts the config to YAML string
func (c *Config) ToYAML() (string, error) {
	data, err := yaml.Marshal(c)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
