package agent

import (
	"testing"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		deviceName string
		serverURL  string
	}{
		{
			name:       "create new agent",
			id:         "agent-001",
			deviceName: "test-device",
			serverURL:  "ws://localhost:8080/ws",
		},
		{
			name:       "create agent with empty id",
			id:         "",
			deviceName: "device",
			serverURL:  "ws://server:8080/ws",
		},
		{
			name:       "create agent with long names",
			id:         "very-long-agent-id-with-many-characters",
			deviceName: "very-long-device-name-with-many-characters",
			serverURL:  "wss://secure-server.example.com:9443/ws",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent := New(tt.id, tt.deviceName, tt.serverURL)

			if agent == nil {
				t.Fatal("New() returned nil")
			}

			if agent.ID != tt.id {
				t.Errorf("ID = %v, want %v", agent.ID, tt.id)
			}

			if agent.DeviceName != tt.deviceName {
				t.Errorf("DeviceName = %v, want %v", agent.DeviceName, tt.deviceName)
			}

			if agent.ServerURL != tt.serverURL {
				t.Errorf("ServerURL = %v, want %v", agent.ServerURL, tt.serverURL)
			}

			if agent.Sessions == nil {
				t.Error("Sessions map is nil")
			}

			if len(agent.Sessions) != 0 {
				t.Errorf("Sessions length = %v, want 0", len(agent.Sessions))
			}

			if agent.SecurityKey != "" {
				t.Errorf("SecurityKey = %v, want empty", agent.SecurityKey)
			}
		})
	}
}

func TestNewWithSecurityKey(t *testing.T) {
	tests := []struct {
		name        string
		id          string
		deviceName  string
		serverURL   string
		securityKey string
	}{
		{
			name:        "create agent with security key",
			id:          "agent-001",
			deviceName:  "secure-device",
			serverURL:   "wss://localhost:8080/ws",
			securityKey: "secret-key-123",
		},
		{
			name:        "create agent with empty security key",
			id:          "agent-002",
			deviceName:  "device",
			serverURL:   "ws://server:8080/ws",
			securityKey: "",
		},
		{
			name:        "create agent with long security key",
			id:          "agent-003",
			deviceName:  "device",
			serverURL:   "wss://server:9443/ws",
			securityKey: "very-long-security-key-with-many-characters-for-authentication",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent := NewWithSecurityKey(tt.id, tt.deviceName, tt.serverURL, tt.securityKey)

			if agent == nil {
				t.Fatal("NewWithSecurityKey() returned nil")
			}

			if agent.ID != tt.id {
				t.Errorf("ID = %v, want %v", agent.ID, tt.id)
			}

			if agent.DeviceName != tt.deviceName {
				t.Errorf("DeviceName = %v, want %v", agent.DeviceName, tt.deviceName)
			}

			if agent.ServerURL != tt.serverURL {
				t.Errorf("ServerURL = %v, want %v", agent.ServerURL, tt.serverURL)
			}

			if agent.SecurityKey != tt.securityKey {
				t.Errorf("SecurityKey = %v, want %v", agent.SecurityKey, tt.securityKey)
			}

			if agent.Sessions == nil {
				t.Error("Sessions map is nil")
			}

			if len(agent.Sessions) != 0 {
				t.Errorf("Sessions length = %v, want 0", len(agent.Sessions))
			}
		})
	}
}

func TestAgent_AddSession(t *testing.T) {
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
			agent := New("agent-001", "test-device", "ws://localhost:8080/ws")

			for _, sessionID := range tt.sessionIDs {
				session := &Session{ID: sessionID}
				agent.AddSession(session)
			}

			if len(agent.Sessions) != len(tt.sessionIDs) {
				t.Errorf("Sessions length = %v, want %v", len(agent.Sessions), len(tt.sessionIDs))
			}

			for _, sessionID := range tt.sessionIDs {
				if _, exists := agent.Sessions[sessionID]; !exists {
					t.Errorf("Session %v not found in agent", sessionID)
				}
			}
		})
	}
}

func TestAgent_GetSession(t *testing.T) {
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
			agent := New("agent-001", "test-device", "ws://localhost:8080/ws")

			// Add sessions
			for _, sessionID := range tt.addSessions {
				session := &Session{ID: sessionID}
				agent.AddSession(session)
			}

			// Get session
			session, exists := agent.GetSession(tt.getSessionID)

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

func TestAgent_RemoveSession(t *testing.T) {
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
			agent := New("agent-001", "test-device", "ws://localhost:8080/ws")

			// Add sessions
			for _, sessionID := range tt.addSessions {
				session := &Session{ID: sessionID}
				agent.AddSession(session)
			}

			// Remove session
			agent.RemoveSession(tt.removeSessionID)

			// Verify remaining count
			if len(agent.Sessions) != tt.wantRemaining {
				t.Errorf("Sessions length after remove = %v, want %v", len(agent.Sessions), tt.wantRemaining)
			}

			// Verify removed session is gone
			if _, exists := agent.Sessions[tt.removeSessionID]; exists {
				t.Errorf("Session %v still exists after removal", tt.removeSessionID)
			}
		})
	}
}

func TestAgent_ConcurrentSessionAccess(t *testing.T) {
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
			agent := New("agent-001", "test-device", "ws://localhost:8080/ws")

			done := make(chan bool, tt.numGoroutines)

			// Spawn goroutines that add and get sessions concurrently
			for i := range tt.numGoroutines {
				go func(id int) {
					for j := range tt.numOperations {
						sessionID := "session-" + string(rune(id)) + "-" + string(rune(j))
						session := &Session{ID: sessionID}

						agent.AddSession(session)
						agent.GetSession(sessionID)

						if j%2 == 0 {
							agent.RemoveSession(sessionID)
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
