package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name       string
		configYAML string
		setupEnv   func() func()
		wantErr    bool
		validate   func(t *testing.T, cfg *Config)
	}{
		{
			name:       "no config file - use defaults",
			configYAML: "",
			setupEnv: func() func() {
				// Create temp dir and change to it
				tmpDir := t.TempDir()

				t.Chdir(tmpDir)

				return func() {}
			},
			wantErr: false,
			validate: func(t *testing.T, cfg *Config) {
				if cfg.Server.Addr != "localhost:8080" {
					t.Errorf("Server.Addr = %v, want localhost:8080", cfg.Server.Addr)
				}

				if cfg.Agent.Server != "ws://localhost:8080/ws" {
					t.Errorf("Agent.Server = %v, want ws://localhost:8080/ws", cfg.Agent.Server)
				}
			},
		},
		{
			name: "valid config file",
			configYAML: `
server:
  addr: "0.0.0.0:9090"
  security_key: "test-key"
  tls: true
agent:
  server: "wss://example.com:9090/ws"
  id: "test-agent"
`,
			setupEnv: func() func() {
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, "config.yaml")
				_ = os.WriteFile(configPath, []byte(`
server:
  addr: "0.0.0.0:9090"
  security_key: "test-key"
  tls: true
agent:
  server: "wss://example.com:9090/ws"
  id: "test-agent"
`), 0644)

				t.Chdir(tmpDir)

				return func() {}
			},
			wantErr: false,
			validate: func(t *testing.T, cfg *Config) {
				if cfg.Server.Addr != "0.0.0.0:9090" {
					t.Errorf("Server.Addr = %v, want 0.0.0.0:9090", cfg.Server.Addr)
				}

				if cfg.Server.SecurityKey != "test-key" {
					t.Errorf("Server.SecurityKey = %v, want test-key", cfg.Server.SecurityKey)
				}

				if !cfg.Server.TLS {
					t.Errorf("Server.TLS = %v, want true", cfg.Server.TLS)
				}

				if cfg.Agent.Server != "wss://example.com:9090/ws" {
					t.Errorf("Agent.Server = %v, want wss://example.com:9090/ws", cfg.Agent.Server)
				}

				if cfg.Agent.ID != "test-agent" {
					t.Errorf("Agent.ID = %v, want test-agent", cfg.Agent.ID)
				}
			},
		},
		{
			name: "partial config - merge with defaults",
			configYAML: `
server:
  addr: "0.0.0.0:7777"
`,
			setupEnv: func() func() {
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, "config.yaml")
				_ = os.WriteFile(configPath, []byte(`
server:
  addr: "0.0.0.0:7777"
`), 0644)

				t.Chdir(tmpDir)

				return func() {}
			},
			wantErr: false,
			validate: func(t *testing.T, cfg *Config) {
				if cfg.Server.Addr != "0.0.0.0:7777" {
					t.Errorf("Server.Addr = %v, want 0.0.0.0:7777", cfg.Server.Addr)
				}
				// Should still have defaults for other fields
				if cfg.Server.TLS != false {
					t.Errorf("Server.TLS = %v, want false (default)", cfg.Server.TLS)
				}

				if cfg.Agent.Server != "ws://localhost:8080/ws" {
					t.Errorf("Agent.Server = %v, want ws://localhost:8080/ws (default)", cfg.Agent.Server)
				}
			},
		},
		{
			name: "invalid YAML - syntax error",
			configYAML: `
server:
  port: not-a-number
  invalid syntax here
`,
			setupEnv: func() func() {
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, "config.yaml")
				_ = os.WriteFile(configPath, []byte(`
server:
  port: not-a-number
  invalid syntax here
`), 0644)

				t.Chdir(tmpDir)

				return func() {}
			},
			wantErr:  true,
			validate: func(t *testing.T, cfg *Config) {},
		},
		{
			name:       "empty config file - use defaults",
			configYAML: ``,
			setupEnv: func() func() {
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, "config.yaml")
				_ = os.WriteFile(configPath, []byte(``), 0644)

				t.Chdir(tmpDir)

				return func() {}
			},
			wantErr: false,
			validate: func(t *testing.T, cfg *Config) {
				if cfg.Server.Addr != "localhost:8080" {
					t.Errorf("Server.Addr = %v, want localhost:8080 (default)", cfg.Server.Addr)
				}
			},
		},
		{
			name: "all command configs",
			configYAML: `
server:
  addr: "localhost:8080"
agent:
  server: "ws://localhost:8080/ws"
connect:
  server: "ws://localhost:8080/ws"
  shell: "/bin/zsh"
upload:
  mode: 0755
download:
  output: "/tmp/downloads"
devices:
  server: "ws://localhost:8080/ws"
sessions:
  server: "ws://localhost:8080/ws"
disconnect:
  server: "ws://localhost:8080/ws"
`,
			setupEnv: func() func() {
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, "config.yaml")
				_ = os.WriteFile(configPath, []byte(`
server:
  addr: "localhost:8080"
agent:
  server: "ws://localhost:8080/ws"
connect:
  server: "ws://localhost:8080/ws"
  shell: "/bin/zsh"
upload:
  mode: 0755
download:
  output: "/tmp/downloads"
devices:
  server: "ws://localhost:8080/ws"
sessions:
  server: "ws://localhost:8080/ws"
disconnect:
  server: "ws://localhost:8080/ws"
`), 0644)

				t.Chdir(tmpDir)

				return func() {}
			},
			wantErr: false,
			validate: func(t *testing.T, cfg *Config) {
				if cfg.Connect.Shell != "/bin/zsh" {
					t.Errorf("Connect.Shell = %v, want /bin/zsh", cfg.Connect.Shell)
				}

				if cfg.Upload.Mode != 0755 {
					t.Errorf("Upload.Mode = %v, want 0755", cfg.Upload.Mode)
				}

				if cfg.Download.Output != "/tmp/downloads" {
					t.Errorf("Download.Output = %v, want /tmp/downloads", cfg.Download.Output)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test environment
			cleanup := tt.setupEnv()
			defer cleanup()

			// Load config
			cfg, err := Load()

			// Check error expectation
			if tt.wantErr && err == nil {
				t.Error("Load() expected error but got nil")

				return
			}

			if !tt.wantErr && err != nil {
				t.Errorf("Load() unexpected error: %v", err)

				return
			}

			// Validate result if no error expected
			if !tt.wantErr {
				tt.validate(t, cfg)
			}
		})
	}
}

func TestServerConfig(t *testing.T) {
	tests := []struct {
		name       string
		configYAML string
		want       ServerConfig
	}{
		{
			name:       "default server config",
			configYAML: ``,
			want: ServerConfig{
				Addr: "localhost:8080",
				TLS:  false,
			},
		},
		{
			name: "custom server config with TLS",
			configYAML: `
server:
  addr: "0.0.0.0:9443"
  security_key: "secret123"
  tls: true
  tls_cert: "/path/to/cert.pem"
  tls_key: "/path/to/key.pem"
`,
			want: ServerConfig{
				Addr:        "0.0.0.0:9443",
				SecurityKey: "secret123",
				TLS:         true,
				TLSCert:     "/path/to/cert.pem",
				TLSKey:      "/path/to/key.pem",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory and config file
			tmpDir := t.TempDir()
			if tt.configYAML != "" {
				configPath := filepath.Join(tmpDir, "config.yaml")
				_ = os.WriteFile(configPath, []byte(tt.configYAML), 0644)
			}

			// Change to temp directory
			t.Chdir(tmpDir)

			// Load config
			cfg, err := Load()
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}

			// Validate server config
			if cfg.Server.Addr != tt.want.Addr {
				t.Errorf("Server.Addr = %v, want %v", cfg.Server.Addr, tt.want.Addr)
			}

			if cfg.Server.SecurityKey != tt.want.SecurityKey {
				t.Errorf("Server.SecurityKey = %v, want %v", cfg.Server.SecurityKey, tt.want.SecurityKey)
			}

			if cfg.Server.TLS != tt.want.TLS {
				t.Errorf("Server.TLS = %v, want %v", cfg.Server.TLS, tt.want.TLS)
			}

			if cfg.Server.TLSCert != tt.want.TLSCert {
				t.Errorf("Server.TLSCert = %v, want %v", cfg.Server.TLSCert, tt.want.TLSCert)
			}

			if cfg.Server.TLSKey != tt.want.TLSKey {
				t.Errorf("Server.TLSKey = %v, want %v", cfg.Server.TLSKey, tt.want.TLSKey)
			}
		})
	}
}

func TestAgentConfig(t *testing.T) {
	tests := []struct {
		name       string
		configYAML string
		want       AgentConfig
	}{
		{
			name:       "default agent config",
			configYAML: ``,
			want: AgentConfig{
				Server:      "ws://localhost:8080/ws",
				TLSInsecure: false,
			},
		},
		{
			name: "custom agent config",
			configYAML: `
agent:
  server: "wss://prod-server.com:9443/ws"
  id: "agent-001"
  security_key: "agent-secret"
  tls_insecure: true
`,
			want: AgentConfig{
				Server:      "wss://prod-server.com:9443/ws",
				ID:          "agent-001",
				SecurityKey: "agent-secret",
				TLSInsecure: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory and config file
			tmpDir := t.TempDir()
			if tt.configYAML != "" {
				configPath := filepath.Join(tmpDir, "config.yaml")
				_ = os.WriteFile(configPath, []byte(tt.configYAML), 0644)
			}

			// Change to temp directory
			t.Chdir(tmpDir)

			// Load config
			cfg, err := Load()
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}

			// Validate agent config
			if cfg.Agent.Server != tt.want.Server {
				t.Errorf("Agent.Server = %v, want %v", cfg.Agent.Server, tt.want.Server)
			}

			if cfg.Agent.ID != tt.want.ID {
				t.Errorf("Agent.ID = %v, want %v", cfg.Agent.ID, tt.want.ID)
			}

			if cfg.Agent.SecurityKey != tt.want.SecurityKey {
				t.Errorf("Agent.SecurityKey = %v, want %v", cfg.Agent.SecurityKey, tt.want.SecurityKey)
			}

			if cfg.Agent.TLSInsecure != tt.want.TLSInsecure {
				t.Errorf("Agent.TLSInsecure = %v, want %v", cfg.Agent.TLSInsecure, tt.want.TLSInsecure)
			}
		})
	}
}

func TestConnectConfig(t *testing.T) {
	tests := []struct {
		name       string
		configYAML string
		want       ConnectConfig
	}{
		{
			name:       "default connect config",
			configYAML: ``,
			want: ConnectConfig{
				Server: "ws://localhost:8080/ws",
				Shell:  "/bin/bash",
			},
		},
		{
			name: "custom connect config",
			configYAML: `
connect:
  server: "ws://remote:8080/ws"
  security_key: "connect-key"
  shell: "/bin/zsh"
`,
			want: ConnectConfig{
				Server:      "ws://remote:8080/ws",
				SecurityKey: "connect-key",
				Shell:       "/bin/zsh",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory and config file
			tmpDir := t.TempDir()
			if tt.configYAML != "" {
				configPath := filepath.Join(tmpDir, "config.yaml")
				_ = os.WriteFile(configPath, []byte(tt.configYAML), 0644)
			}

			// Change to temp directory
			t.Chdir(tmpDir)

			// Load config
			cfg, err := Load()
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}

			// Validate connect config
			if cfg.Connect.Server != tt.want.Server {
				t.Errorf("Connect.Server = %v, want %v", cfg.Connect.Server, tt.want.Server)
			}

			if cfg.Connect.SecurityKey != tt.want.SecurityKey {
				t.Errorf("Connect.SecurityKey = %v, want %v", cfg.Connect.SecurityKey, tt.want.SecurityKey)
			}

			if cfg.Connect.Shell != tt.want.Shell {
				t.Errorf("Connect.Shell = %v, want %v", cfg.Connect.Shell, tt.want.Shell)
			}
		})
	}
}

func TestUploadDownloadConfig(t *testing.T) {
	tests := []struct {
		name         string
		configYAML   string
		wantUpload   UploadConfig
		wantDownload DownloadConfig
	}{
		{
			name:       "default upload/download config",
			configYAML: ``,
			wantUpload: UploadConfig{
				Server: "ws://localhost:8080/ws",
				Mode:   0644,
			},
			wantDownload: DownloadConfig{
				Server: "ws://localhost:8080/ws",
			},
		},
		{
			name: "custom upload/download config",
			configYAML: `
upload:
  server: "ws://fileserver:8080/ws"
  security_key: "upload-key"
  mode: 0755
download:
  server: "ws://fileserver:8080/ws"
  security_key: "download-key"
  output: "/tmp/downloads"
`,
			wantUpload: UploadConfig{
				Server:      "ws://fileserver:8080/ws",
				SecurityKey: "upload-key",
				Mode:        0755,
			},
			wantDownload: DownloadConfig{
				Server:      "ws://fileserver:8080/ws",
				SecurityKey: "download-key",
				Output:      "/tmp/downloads",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory and config file
			tmpDir := t.TempDir()
			if tt.configYAML != "" {
				configPath := filepath.Join(tmpDir, "config.yaml")
				_ = os.WriteFile(configPath, []byte(tt.configYAML), 0644)
			}

			// Change to temp directory
			t.Chdir(tmpDir)

			// Load config
			cfg, err := Load()
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}

			// Validate upload config
			if cfg.Upload.Server != tt.wantUpload.Server {
				t.Errorf("Upload.Server = %v, want %v", cfg.Upload.Server, tt.wantUpload.Server)
			}

			if cfg.Upload.SecurityKey != tt.wantUpload.SecurityKey {
				t.Errorf("Upload.SecurityKey = %v, want %v", cfg.Upload.SecurityKey, tt.wantUpload.SecurityKey)
			}

			if cfg.Upload.Mode != tt.wantUpload.Mode {
				t.Errorf("Upload.Mode = %v, want %v", cfg.Upload.Mode, tt.wantUpload.Mode)
			}

			// Validate download config
			if cfg.Download.Server != tt.wantDownload.Server {
				t.Errorf("Download.Server = %v, want %v", cfg.Download.Server, tt.wantDownload.Server)
			}

			if cfg.Download.SecurityKey != tt.wantDownload.SecurityKey {
				t.Errorf("Download.SecurityKey = %v, want %v", cfg.Download.SecurityKey, tt.wantDownload.SecurityKey)
			}

			if cfg.Download.Output != tt.wantDownload.Output {
				t.Errorf("Download.Output = %v, want %v", cfg.Download.Output, tt.wantDownload.Output)
			}
		})
	}
}
