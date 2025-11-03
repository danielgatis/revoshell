package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config holds all configuration for revoshell commands.
type Config struct {
	Server     ServerConfig     `mapstructure:"server"`
	Agent      AgentConfig      `mapstructure:"agent"`
	Connect    ConnectConfig    `mapstructure:"connect"`
	Upload     UploadConfig     `mapstructure:"upload"`
	Download   DownloadConfig   `mapstructure:"download"`
	Devices    DevicesConfig    `mapstructure:"devices"`
	Sessions   SessionsConfig   `mapstructure:"sessions"`
	Disconnect DisconnectConfig `mapstructure:"disconnect"`
}

// ServerConfig holds server command configuration.
type ServerConfig struct {
	Addr        string `mapstructure:"addr"`
	SecurityKey string `mapstructure:"security_key"`
	TLS         bool   `mapstructure:"tls"`
	TLSCert     string `mapstructure:"tls_cert"`
	TLSKey      string `mapstructure:"tls_key"`
}

// AgentConfig holds agent command configuration.
type AgentConfig struct {
	Server      string `mapstructure:"server"`
	ID          string `mapstructure:"id"`
	SecurityKey string `mapstructure:"security_key"`
	TLSInsecure bool   `mapstructure:"tls_insecure"`
}

// ConnectConfig holds connect command configuration.
type ConnectConfig struct {
	Server      string `mapstructure:"server"`
	SecurityKey string `mapstructure:"security_key"`
	Shell       string `mapstructure:"shell"`
}

// UploadConfig holds upload command configuration.
type UploadConfig struct {
	Server      string `mapstructure:"server"`
	SecurityKey string `mapstructure:"security_key"`
	Mode        uint32 `mapstructure:"mode"`
}

// DownloadConfig holds download command configuration.
type DownloadConfig struct {
	Server      string `mapstructure:"server"`
	SecurityKey string `mapstructure:"security_key"`
	Output      string `mapstructure:"output"`
}

// DevicesConfig holds devices command configuration.
type DevicesConfig struct {
	Server      string `mapstructure:"server"`
	SecurityKey string `mapstructure:"security_key"`
}

// SessionsConfig holds sessions command configuration.
type SessionsConfig struct {
	Server      string `mapstructure:"server"`
	SecurityKey string `mapstructure:"security_key"`
}

// DisconnectConfig holds disconnect command configuration.
type DisconnectConfig struct {
	Server      string `mapstructure:"server"`
	SecurityKey string `mapstructure:"security_key"`
}

// Load reads configuration from file and returns Config struct.
func Load() (*Config, error) {
	v := viper.New()

	// Set config name and paths
	v.SetConfigName("config")
	v.SetConfigType("yaml")

	// Look for config in multiple locations
	v.AddConfigPath(".")              // Current directory
	v.AddConfigPath("$HOME/.revossh") // Home directory
	v.AddConfigPath("/etc/revossh")   // System-wide config

	if configDir := os.Getenv("XDG_CONFIG_HOME"); configDir != "" {
		v.AddConfigPath(filepath.Join(configDir, "revossh"))
	}

	// Set defaults
	setDefaults(v)

	// Read config file (optional - don't fail if not found)
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// Config file found but error reading it
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
		// Config file not found - use defaults
	}

	// Unmarshal into Config struct
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	return &cfg, nil
}

// setDefaults sets default values for configuration.
func setDefaults(v *viper.Viper) {
	// Server defaults
	v.SetDefault("server.addr", "localhost:8080")
	v.SetDefault("server.tls", false)

	// Agent defaults
	v.SetDefault("agent.server", "ws://localhost:8080/ws")
	v.SetDefault("agent.tls_insecure", false)

	// Connect defaults
	v.SetDefault("connect.server", "ws://localhost:8080/ws")
	v.SetDefault("connect.shell", "/bin/bash")

	// Upload defaults
	v.SetDefault("upload.server", "ws://localhost:8080/ws")
	v.SetDefault("upload.mode", 0644)

	// Download defaults
	v.SetDefault("download.server", "ws://localhost:8080/ws")

	// Devices defaults
	v.SetDefault("devices.server", "ws://localhost:8080/ws")

	// Sessions defaults
	v.SetDefault("sessions.server", "ws://localhost:8080/ws")

	// Disconnect defaults
	v.SetDefault("disconnect.server", "ws://localhost:8080/ws")
}
