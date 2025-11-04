package server

import (
	"context"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
	"github.com/sourcegraph/jsonrpc2"

	"github.com/danielgatis/revoshell/pkg/logging"
	"github.com/danielgatis/revoshell/pkg/transport"
	"github.com/danielgatis/revoshell/pkg/version"
)

const (
	// MaxDevices is the maximum number of devices that can be connected simultaneously.
	MaxDevices = 10000
)

var (
	log      zerolog.Logger
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

func init() {
	log = logging.InitWithComponent("SERVER")
}

// Server manages connected devices.
type Server struct {
	devices     map[string]*Device
	mu          sync.RWMutex
	securityKey string
}

// New creates a new server instance.
func New() *Server {
	return &Server{
		devices: make(map[string]*Device),
	}
}

// NewWithSecurityKey creates a new server instance with authentication.
func NewWithSecurityKey(securityKey string) *Server {
	return &Server{
		devices:     make(map[string]*Device),
		securityKey: securityKey,
	}
}

// AddDevice adds a device to the server with LRU eviction if limit is reached.
func (s *Server) AddDevice(device *Device) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if we've reached the maximum device limit
	if len(s.devices) >= MaxDevices {
		// Find the oldest device by connection time (LRU eviction)
		var oldestID string

		var oldestTime = device.ConnectedAt // Initialize with current time

		for id, dev := range s.devices {
			if oldestTime.IsZero() || dev.ConnectedAt.Before(oldestTime) {
				oldestID = id
				oldestTime = dev.ConnectedAt
			}
		}

		// Evict the oldest device
		if oldestID != "" {
			if old, exists := s.devices[oldestID]; exists {
				// Close the connection
				if old.Conn != nil {
					_ = old.Conn.Close()
				}

				delete(s.devices, oldestID)
				log.Warn().
					Str("device_id", oldestID).
					Int("total_devices", len(s.devices)).
					Msg("Evicted oldest device due to MaxDevices limit")
			}
		}
	}

	s.devices[device.ID] = device
}

// GetDevice returns a device by ID.
func (s *Server) GetDevice(deviceID string) (*Device, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	device, exists := s.devices[deviceID]

	return device, exists
}

// RemoveDevice removes a device from the server.
func (s *Server) RemoveDevice(deviceID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.devices, deviceID)
}

// ListDevices returns the list of connected device IDs.
func (s *Server) ListDevices() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ids := make([]string, 0, len(s.devices))
	for id := range s.devices {
		ids = append(ids, id)
	}

	return ids
}

// cleanupDevice performs complete cleanup of a device including sessions and connections.
func (s *Server) cleanupDevice(device *Device, conn *jsonrpc2.Conn, wsConn *websocket.Conn) {
	// Clean up all device sessions first
	device.mu.Lock()

	for sessionID, session := range device.Sessions {
		// Safely close session Done channel
		if session.closed.CompareAndSwap(false, true) {
			close(session.Done)
		}

		delete(device.Sessions, sessionID)
	}

	device.mu.Unlock()

	// Close connections
	if err := conn.Close(); err != nil {
		log.Warn().Err(err).Msg("Error closing JSON-RPC connection")
	}

	if err := wsConn.Close(); err != nil {
		log.Warn().Err(err).Msg("Error closing WebSocket connection")
	}

	// Remove device from server
	s.RemoveDevice(device.ID)

	log.Info().Str("device", device.ID).Msg("Device disconnected and cleaned up")
}

// HandleWebSocket processes WebSocket connections.
func (s *Server) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Validate security key if configured
	if s.securityKey != "" {
		providedKey := r.Header.Get("X-Security-Key")
		if providedKey != s.securityKey {
			log.Warn().
				Str("remote", r.RemoteAddr).
				Msg("Unauthorized: invalid security key")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)

			return
		}
	}

	// Validate client version
	clientVersion := r.Header.Get("X-Client-Version")

	serverVersion := version.GetVersion()
	if clientVersion != serverVersion {
		log.Warn().
			Str("remote", r.RemoteAddr).
			Str("client_version", clientVersion).
			Str("server_version", serverVersion).
			Msg("Version mismatch")
		http.Error(w, "Version mismatch", http.StatusPreconditionFailed)

		return
	}

	wsConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error().Err(err).Msg("Error upgrading WebSocket connection")

		return
	}

	log.Info().Str("remote", r.RemoteAddr).Msg("New WebSocket connection")

	device := NewDevice()
	handler := NewHandler(s, device)

	stream := transport.NewWebSocketStream(wsConn)
	conn := jsonrpc2.NewConn(context.Background(), stream, handler)
	device.Conn = conn

	defer s.cleanupDevice(device, conn, wsConn)

	<-conn.DisconnectNotify()
	log.Info().Str("device", device.ID).Msg("Device disconnected")
}

// Start starts the HTTP/WebSocket server (insecure).
func (s *Server) Start(addr string) error {
	http.HandleFunc("/ws", s.HandleWebSocket)
	log.Info().Str("address", "ws://"+addr+"/ws").Msg("WebSocket server started")

	return http.ListenAndServe(addr, nil)
}

// StartTLS starts the HTTPS/WebSocket server with TLS (secure).
func (s *Server) StartTLS(addr string, certFile string, keyFile string) error {
	http.HandleFunc("/ws", s.HandleWebSocket)
	log.Info().Str("address", "wss://"+addr+"/ws").Msg("Secure WebSocket server started")

	return http.ListenAndServeTLS(addr, certFile, keyFile, nil)
}
