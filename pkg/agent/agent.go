package agent

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
	"github.com/sourcegraph/jsonrpc2"

	"github.com/danielgatis/revoshell/pkg/logging"
	"github.com/danielgatis/revoshell/pkg/protocol"
	"github.com/danielgatis/revoshell/pkg/transport"
	"github.com/danielgatis/revoshell/pkg/version"
)

var log zerolog.Logger

func init() {
	log = logging.InitWithComponent("AGENT")
}

// Agent represents the agent that connects to the server.
type Agent struct {
	ID          string
	DeviceName  string
	ServerURL   string
	SecurityKey string
	Sessions    map[string]*Session
	Conn        *jsonrpc2.Conn
	mu          sync.RWMutex
}

// New creates a new Agent instance.
func New(id, deviceName, serverURL string) *Agent {
	return &Agent{
		ID:         id,
		DeviceName: deviceName,
		ServerURL:  serverURL,
		Sessions:   make(map[string]*Session),
	}
}

// NewWithSecurityKey creates a new Agent instance with authentication.
func NewWithSecurityKey(id, deviceName, serverURL, securityKey string) *Agent {
	return &Agent{
		ID:          id,
		DeviceName:  deviceName,
		ServerURL:   serverURL,
		SecurityKey: securityKey,
		Sessions:    make(map[string]*Session),
	}
}

// AddSession adds a session to the agent.
func (a *Agent) AddSession(session *Session) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.Sessions[session.ID] = session
}

// GetSession returns a session by ID.
func (a *Agent) GetSession(sessionID string) (*Session, bool) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	session, exists := a.Sessions[sessionID]

	return session, exists
}

// RemoveSession removes a session from the agent.
func (a *Agent) RemoveSession(sessionID string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	delete(a.Sessions, sessionID)
}

// Connect establishes connection with the server.
func (a *Agent) Connect() error {
	log.Info().Str("server", a.ServerURL).Msg("Connecting to server")

	// Prepare headers with version and security key
	headers := make(map[string][]string)

	headers["X-Client-Version"] = []string{version.GetVersion()}
	if a.SecurityKey != "" {
		headers["X-Security-Key"] = []string{a.SecurityKey}
	}

	wsConn, _, err := websocket.DefaultDialer.Dial(a.ServerURL, headers)
	if err != nil {
		return fmt.Errorf("connection error: %w", err)
	}

	stream := transport.NewWebSocketStream(wsConn)
	handler := NewHandler(a)
	conn := jsonrpc2.NewConn(context.Background(), stream, handler)
	a.Conn = conn

	log.Info().Msg("Connected to server")

	// Check version compatibility
	if err := a.checkVersion(); err != nil {
		return err
	}

	// Register agent
	if err := a.register(); err != nil {
		return err
	}

	// Wait for disconnection
	<-conn.DisconnectNotify()
	log.Info().Msg("Disconnected from server")

	return nil
}

func (a *Agent) register() error {
	hostname, _ := os.Hostname()
	registerParams := protocol.RegisterParams{
		AgentID:  a.ID,
		Name:     a.DeviceName,
		Hostname: hostname,
		Platform: "linux/darwin",
	}

	var result protocol.RegisterResult
	if err := a.Conn.Call(context.Background(), protocol.MethodAgentRegister, registerParams, &result); err != nil {
		log.Error().Err(err).Msg("Registration error")

		return err
	}

	if !result.Success {
		log.Error().Str("message", result.Message).Msg("Registration failed")

		return fmt.Errorf("registration failed: %s", result.Message)
	}

	log.Info().Str("message", result.Message).Msg("Successfully registered")

	return nil
}

func (a *Agent) checkVersion() error {
	versionParams := protocol.VersionCheckParams{
		Version:   version.GetVersion(),
		GitCommit: version.GetGitCommit(),
	}

	var result protocol.VersionCheckResult
	if err := a.Conn.Call(context.Background(), protocol.MethodVersionCheck, versionParams, &result); err != nil {
		log.Error().Err(err).Msg("Version check failed")

		return err
	}

	if !result.Compatible {
		log.Error().
			Str("agent_version", version.GetVersion()).
			Str("server_version", result.Version).
			Str("message", result.Message).
			Msg("Version incompatible")

		return fmt.Errorf("version mismatch: %s", result.Message)
	}

	log.Info().Str("version", version.GetVersion()).Msg("Version check passed")

	return nil
}

// Run starts the agent with automatic reconnection using exponential backoff.
func (a *Agent) Run() {
	backoff := 1 * time.Second
	maxBackoff := 60 * time.Second

	for {
		if err := a.Connect(); err != nil {
			log.Error().Err(err).Msg("Connection error")

			// Add jitter: ±25% randomization to prevent thundering herd
			jitter := time.Duration(float64(backoff) * (0.75 + 0.5*rand.Float64()))
			log.Info().Dur("delay", jitter).Msg("Reconnecting...")
			time.Sleep(jitter)

			// Exponential backoff: double the delay for next retry
			backoff *= 2
			if backoff > maxBackoff {
				backoff = maxBackoff
			}

			continue
		}

		// Reset backoff on successful connection
		backoff = 1 * time.Second

		log.Info().Msg("Connection established, backoff reset")
	}
}
