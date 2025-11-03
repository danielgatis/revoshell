package server

import (
	"sync"
	"time"

	"github.com/sourcegraph/jsonrpc2"
)

// Device represents a device connected to the server.
type Device struct {
	ID          string
	Name        string
	Conn        *jsonrpc2.Conn
	Sessions    map[string]*Session
	ConnectedAt time.Time
	mu          sync.RWMutex
}

// NewDevice creates a new Device instance.
func NewDevice() *Device {
	return &Device{
		Sessions:    make(map[string]*Session),
		ConnectedAt: time.Now(),
	}
}

// AddSession adds a session to the device.
func (d *Device) AddSession(session *Session) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.Sessions[session.ID] = session
}

// GetSession returns a session by ID.
func (d *Device) GetSession(sessionID string) (*Session, bool) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	session, exists := d.Sessions[sessionID]

	return session, exists
}

// RemoveSession removes a session from the device.
func (d *Device) RemoveSession(sessionID string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if session, exists := d.Sessions[sessionID]; exists {
		// Only close if not already closed
		if session.closed.CompareAndSwap(false, true) {
			close(session.Done)
		}

		delete(d.Sessions, sessionID)
	}
}

// SessionCount returns the number of active sessions.
func (d *Device) SessionCount() int {
	d.mu.RLock()
	defer d.mu.RUnlock()

	return len(d.Sessions)
}
