package server

import (
	"testing"
)

func TestNewSession(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		agentID string
	}{
		{
			name:    "create new session",
			id:      "session-001",
			agentID: "agent-001",
		},
		{
			name:    "create session with empty id",
			id:      "",
			agentID: "agent-001",
		},
		{
			name:    "create session with empty agent id",
			id:      "session-002",
			agentID: "",
		},
		{
			name:    "create session with long ids",
			id:      "very-long-session-id-with-many-characters",
			agentID: "very-long-agent-id-with-many-characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := NewSession(tt.id, tt.agentID)

			if session == nil {
				t.Fatal("NewSession() returned nil")
			}

			if session.ID != tt.id {
				t.Errorf("ID = %v, want %v", session.ID, tt.id)
			}

			if session.AgentID != tt.agentID {
				t.Errorf("AgentID = %v, want %v", session.AgentID, tt.agentID)
			}

			if session.Input == nil {
				t.Error("Input channel is nil")
			}

			if session.Output == nil {
				t.Error("Output channel is nil")
			}

			if session.Done == nil {
				t.Error("Done channel is nil")
			}

			// Verify channel capacities
			if cap(session.Input) != 100 {
				t.Errorf("Input channel capacity = %v, want 100", cap(session.Input))
			}

			if cap(session.Output) != 100 {
				t.Errorf("Output channel capacity = %v, want 100", cap(session.Output))
			}
		})
	}
}

func TestSession_Channels(t *testing.T) {
	tests := []struct {
		name         string
		sendData     []byte
		wantReceived bool
	}{
		{
			name:         "send data through input channel",
			sendData:     []byte("test data"),
			wantReceived: true,
		},
		{
			name:         "send empty data",
			sendData:     []byte{},
			wantReceived: true,
		},
		{
			name:         "send nil data",
			sendData:     nil,
			wantReceived: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := NewSession("session-001", "agent-001")

			// Send data to Input channel
			go func() {
				session.Input <- tt.sendData
			}()

			// Receive data from Input channel
			received := <-session.Input

			if !tt.wantReceived {
				t.Error("Received data when not expected")
			}

			// Compare byte slices
			if len(received) != len(tt.sendData) {
				t.Errorf("Received data length = %v, want %v", len(received), len(tt.sendData))
			}

			for i, b := range received {
				if i < len(tt.sendData) && b != tt.sendData[i] {
					t.Errorf("Received data[%d] = %v, want %v", i, b, tt.sendData[i])
				}
			}
		})
	}
}

func TestSession_MultipleMessages(t *testing.T) {
	tests := []struct {
		name     string
		messages [][]byte
	}{
		{
			name: "send multiple messages",
			messages: [][]byte{
				[]byte("message 1"),
				[]byte("message 2"),
				[]byte("message 3"),
			},
		},
		{
			name: "send many messages",
			messages: func() [][]byte {
				msgs := make([][]byte, 50)
				for i := range msgs {
					msgs[i] = []byte("message")
				}

				return msgs
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := NewSession("session-001", "agent-001")

			// Send all messages
			go func() {
				for _, msg := range tt.messages {
					session.Input <- msg
				}
			}()

			// Receive all messages
			for range len(tt.messages) {
				<-session.Input // Message received successfully
			}
		})
	}
}

func TestSession_OutputChannel(t *testing.T) {
	tests := []struct {
		name     string
		sendData []byte
	}{
		{
			name:     "send through output channel",
			sendData: []byte("output data"),
		},
		{
			name:     "send empty output",
			sendData: []byte{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := NewSession("session-001", "agent-001")

			// Send data to Output channel
			go func() {
				session.Output <- tt.sendData
			}()

			// Receive data from Output channel
			received := <-session.Output

			// Compare byte slices
			if len(received) != len(tt.sendData) {
				t.Errorf("Received output length = %v, want %v", len(received), len(tt.sendData))
			}

			for i, b := range received {
				if i < len(tt.sendData) && b != tt.sendData[i] {
					t.Errorf("Received output[%d] = %v, want %v", i, b, tt.sendData[i])
				}
			}
		})
	}
}

func TestSession_DoneChannel(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "done channel receives when closed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := NewSession("session-001", "agent-001")

			// Close Done channel in goroutine
			go func() {
				if session.closed.CompareAndSwap(false, true) {
					close(session.Done)
				}
			}()

			// Wait for Done channel to be closed
			<-session.Done

			// Verify channel is closed by trying to receive again
			_, ok := <-session.Done
			if ok {
				t.Error("Done channel is not closed")
			}
		})
	}
}

func TestSession_ClosedFlag(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "closed flag works correctly",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := NewSession("session-001", "agent-001")

			// Verify initially not closed
			if session.closed.Load() {
				t.Error("Session should not be closed initially")
			}

			// Close the session
			if !session.closed.CompareAndSwap(false, true) {
				t.Error("First CompareAndSwap should return true")
			}

			// Verify closed flag is set
			if !session.closed.Load() {
				t.Error("Session should be marked as closed")
			}

			// Try to close again - should fail
			if session.closed.CompareAndSwap(false, true) {
				t.Error("Second CompareAndSwap should return false")
			}
		})
	}
}
