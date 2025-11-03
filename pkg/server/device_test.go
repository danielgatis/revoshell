package server

import (
	"testing"
	"time"
)

func TestNewDevice(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "create new device",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before := time.Now()
			device := NewDevice()
			after := time.Now()

			if device == nil {
				t.Fatal("NewDevice() returned nil")
			}

			if device.Sessions == nil {
				t.Error("Sessions map is nil")
			}

			if len(device.Sessions) != 0 {
				t.Errorf("Sessions length = %v, want 0", len(device.Sessions))
			}

			// Verify ConnectedAt is set and within reasonable time range
			if device.ConnectedAt.Before(before) || device.ConnectedAt.After(after) {
				t.Errorf("ConnectedAt = %v, want between %v and %v", device.ConnectedAt, before, after)
			}
		})
	}
}

func TestDevice_AddSession(t *testing.T) {
	tests := []struct {
		name       string
		sessionIDs []string
	}{
		{
			name:       "add single session",
			sessionIDs: []string{"session-001"},
		},
		{
			name:       "add multiple sessions",
			sessionIDs: []string{"session-001", "session-002", "session-003"},
		},
		{
			name:       "add session with empty id",
			sessionIDs: []string{""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			device := NewDevice()

			for _, sessionID := range tt.sessionIDs {
				session := NewSession(sessionID, "agent-001")
				device.AddSession(session)
			}

			if len(device.Sessions) != len(tt.sessionIDs) {
				t.Errorf("Sessions length = %v, want %v", len(device.Sessions), len(tt.sessionIDs))
			}

			for _, sessionID := range tt.sessionIDs {
				if _, exists := device.Sessions[sessionID]; !exists {
					t.Errorf("Session %v not found in device", sessionID)
				}
			}
		})
	}
}

func TestDevice_GetSession(t *testing.T) {
	tests := []struct {
		name         string
		addSessions  []string
		getSessionID string
		wantExists   bool
	}{
		{
			name:         "get existing session",
			addSessions:  []string{"session-001", "session-002"},
			getSessionID: "session-001",
			wantExists:   true,
		},
		{
			name:         "get non-existing session",
			addSessions:  []string{"session-001"},
			getSessionID: "session-999",
			wantExists:   false,
		},
		{
			name:         "get from empty sessions",
			addSessions:  []string{},
			getSessionID: "session-001",
			wantExists:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			device := NewDevice()

			// Add sessions
			for _, sessionID := range tt.addSessions {
				session := NewSession(sessionID, "agent-001")
				device.AddSession(session)
			}

			// Get session
			session, exists := device.GetSession(tt.getSessionID)

			if exists != tt.wantExists {
				t.Errorf("GetSession() exists = %v, want %v", exists, tt.wantExists)
			}

			if tt.wantExists && session == nil {
				t.Error("GetSession() returned nil session when it should exist")
			}

			if tt.wantExists && session.ID != tt.getSessionID {
				t.Errorf("GetSession() session.ID = %v, want %v", session.ID, tt.getSessionID)
			}

			if !tt.wantExists && session != nil {
				t.Error("GetSession() returned non-nil session when it should not exist")
			}
		})
	}
}

func TestDevice_RemoveSession(t *testing.T) {
	tests := []struct {
		name            string
		addSessions     []string
		removeSessionID string
		wantRemaining   int
	}{
		{
			name:            "remove existing session",
			addSessions:     []string{"session-001", "session-002", "session-003"},
			removeSessionID: "session-002",
			wantRemaining:   2,
		},
		{
			name:            "remove non-existing session",
			addSessions:     []string{"session-001"},
			removeSessionID: "session-999",
			wantRemaining:   1,
		},
		{
			name:            "remove from empty sessions",
			addSessions:     []string{},
			removeSessionID: "session-001",
			wantRemaining:   0,
		},
		{
			name:            "remove last session",
			addSessions:     []string{"session-001"},
			removeSessionID: "session-001",
			wantRemaining:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			device := NewDevice()

			// Add sessions
			for _, sessionID := range tt.addSessions {
				session := NewSession(sessionID, "agent-001")
				device.AddSession(session)
			}

			// Remove session
			device.RemoveSession(tt.removeSessionID)

			// Verify remaining count
			if len(device.Sessions) != tt.wantRemaining {
				t.Errorf("Sessions length after remove = %v, want %v", len(device.Sessions), tt.wantRemaining)
			}

			// Verify removed session is gone
			if _, exists := device.Sessions[tt.removeSessionID]; exists {
				t.Errorf("Session %v still exists after removal", tt.removeSessionID)
			}
		})
	}
}

func TestDevice_SessionCount(t *testing.T) {
	tests := []struct {
		name        string
		addSessions []string
		wantCount   int
	}{
		{
			name:        "count with multiple sessions",
			addSessions: []string{"session-001", "session-002", "session-003"},
			wantCount:   3,
		},
		{
			name:        "count with single session",
			addSessions: []string{"session-001"},
			wantCount:   1,
		},
		{
			name:        "count with no sessions",
			addSessions: []string{},
			wantCount:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			device := NewDevice()

			// Add sessions
			for _, sessionID := range tt.addSessions {
				session := NewSession(sessionID, "agent-001")
				device.AddSession(session)
			}

			// Get session count
			count := device.SessionCount()

			if count != tt.wantCount {
				t.Errorf("SessionCount() = %v, want %v", count, tt.wantCount)
			}
		})
	}
}

func TestDevice_ConcurrentSessionAccess(t *testing.T) {
	tests := []struct {
		name          string
		numGoroutines int
		numOperations int
	}{
		{
			name:          "concurrent add and get",
			numGoroutines: 10,
			numOperations: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			device := NewDevice()

			done := make(chan bool, tt.numGoroutines)

			// Spawn goroutines that add and get sessions concurrently
			for i := range tt.numGoroutines {
				go func(id int) {
					for j := range tt.numOperations {
						sessionID := "session-" + string(rune(id)) + "-" + string(rune(j))
						session := NewSession(sessionID, "agent-001")

						device.AddSession(session)
						device.GetSession(sessionID)
						device.SessionCount()

						if j%2 == 0 {
							device.RemoveSession(sessionID)
						}
					}

					done <- true
				}(i)
			}

			// Wait for all goroutines to complete
			for range tt.numGoroutines {
				<-done
			}
			// Test passes if no race conditions occurred
		})
	}
}
