package transport

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
)

func TestNewWebSocketStream(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "creates valid stream",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test WebSocket server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				upgrader := websocket.Upgrader{}

				conn, err := upgrader.Upgrade(w, r, nil)
				if err != nil {
					t.Fatalf("Upgrade error: %v", err)
				}

				defer func() { _ = conn.Close() }()

				// Keep connection alive for client
				var msg interface{}

				_ = conn.ReadJSON(&msg)
			}))
			defer server.Close()

			// Connect to the test server
			wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

			conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
			if err != nil {
				t.Fatalf("Dial error: %v", err)
			}

			defer func() { _ = conn.Close() }()

			// Create stream
			stream := NewWebSocketStream(conn)

			// Verify it implements the interface
			var _ = stream

			// Verify stream is not nil
			if stream == nil {
				t.Error("NewWebSocketStream() returned nil")
			}
		})
	}
}

func TestWebSocketStream_WriteObject(t *testing.T) {
	tests := []struct {
		name    string
		object  interface{}
		wantErr bool
	}{
		{
			name: "write simple object",
			object: map[string]interface{}{
				"method": "test.method",
				"params": map[string]string{"key": "value"},
			},
			wantErr: false,
		},
		{
			name:    "write string",
			object:  "test message",
			wantErr: false,
		},
		{
			name:    "write number",
			object:  42,
			wantErr: false,
		},
		{
			name: "write struct",
			object: struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			}{
				ID:   "test-id",
				Name: "test-name",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test WebSocket server that echoes back what it receives
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				upgrader := websocket.Upgrader{}

				conn, err := upgrader.Upgrade(w, r, nil)
				if err != nil {
					return
				}

				defer func() { _ = conn.Close() }()

				// Read the object sent by the client
				var received interface{}
				if err := conn.ReadJSON(&received); err != nil {
					return
				}

				// Echo it back
				_ = conn.WriteJSON(received)
			}))
			defer server.Close()

			// Connect to the test server
			wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

			conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
			if err != nil {
				t.Fatalf("Dial error: %v", err)
			}

			defer func() { _ = conn.Close() }()

			// Create stream
			stream := NewWebSocketStream(conn)

			// Write object
			err = stream.WriteObject(tt.object)

			// Check error expectation
			if (err != nil) != tt.wantErr {
				t.Errorf("WriteObject() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestWebSocketStream_ReadObject(t *testing.T) {
	tests := []struct {
		name     string
		sendData interface{}
		wantErr  bool
	}{
		{
			name: "read simple object",
			sendData: map[string]interface{}{
				"method": "test.method",
				"id":     "123",
			},
			wantErr: false,
		},
		{
			name:     "read string",
			sendData: "test message",
			wantErr:  false,
		},
		{
			name:     "read number",
			sendData: 42,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test WebSocket server that sends data
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				upgrader := websocket.Upgrader{}

				conn, err := upgrader.Upgrade(w, r, nil)
				if err != nil {
					return
				}

				defer func() { _ = conn.Close() }()

				// Send the test data
				_ = conn.WriteJSON(tt.sendData)

				// Keep connection alive
				var msg interface{}

				_ = conn.ReadJSON(&msg)
			}))
			defer server.Close()

			// Connect to the test server
			wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

			conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
			if err != nil {
				t.Fatalf("Dial error: %v", err)
			}

			defer func() { _ = conn.Close() }()

			// Create stream
			stream := NewWebSocketStream(conn)

			// Read object
			var received interface{}

			err = stream.ReadObject(&received)

			// Check error expectation
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadObject() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Verify data if no error
			if err == nil {
				// Convert both to JSON for comparison
				expectedJSON, _ := json.Marshal(tt.sendData)
				receivedJSON, _ := json.Marshal(received)

				if string(expectedJSON) != string(receivedJSON) {
					t.Errorf("ReadObject() received = %v, want %v", string(receivedJSON), string(expectedJSON))
				}
			}
		})
	}
}

func TestWebSocketStream_Close(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "close connection",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test WebSocket server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				upgrader := websocket.Upgrader{}

				conn, err := upgrader.Upgrade(w, r, nil)
				if err != nil {
					return
				}

				defer func() { _ = conn.Close() }()

				// Wait for close
				var msg interface{}

				_ = conn.ReadJSON(&msg)
			}))
			defer server.Close()

			// Connect to the test server
			wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

			conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
			if err != nil {
				t.Fatalf("Dial error: %v", err)
			}

			// Create stream
			stream := NewWebSocketStream(conn)

			// Close stream
			err = stream.Close()

			// Check error expectation
			if (err != nil) != tt.wantErr {
				t.Errorf("Close() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestWebSocketStream_ConcurrentWrites(t *testing.T) {
	tests := []struct {
		name       string
		numWrites  int
		numReaders int
	}{
		{
			name:       "multiple concurrent writes",
			numWrites:  10,
			numReaders: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test WebSocket server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				upgrader := websocket.Upgrader{}

				conn, err := upgrader.Upgrade(w, r, nil)
				if err != nil {
					return
				}

				defer func() { _ = conn.Close() }()

				// Read all messages
				for range tt.numWrites {
					var msg interface{}
					if err := conn.ReadJSON(&msg); err != nil {
						return
					}
				}
			}))
			defer server.Close()

			// Connect to the test server
			wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

			conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
			if err != nil {
				t.Fatalf("Dial error: %v", err)
			}

			defer func() { _ = conn.Close() }()

			// Create stream
			stream := NewWebSocketStream(conn)

			// Perform concurrent writes
			errChan := make(chan error, tt.numWrites)

			for i := range tt.numWrites {
				go func(n int) {
					obj := map[string]interface{}{
						"id":      n,
						"message": "test",
					}
					errChan <- stream.WriteObject(obj)
				}(i)
			}

			// Collect errors
			for range tt.numWrites {
				if err := <-errChan; err != nil {
					t.Errorf("concurrent WriteObject() error = %v", err)
				}
			}
		})
	}
}

func TestWebSocketStream_ConcurrentReads(t *testing.T) {
	tests := []struct {
		name      string
		numReads  int
		numWrites int
	}{
		{
			name:      "multiple concurrent reads",
			numReads:  5,
			numWrites: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test WebSocket server that sends multiple messages
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				upgrader := websocket.Upgrader{}

				conn, err := upgrader.Upgrade(w, r, nil)
				if err != nil {
					return
				}

				defer func() { _ = conn.Close() }()

				// Send multiple messages
				for i := range tt.numWrites {
					obj := map[string]interface{}{
						"id":      i,
						"message": "test",
					}
					if err := conn.WriteJSON(obj); err != nil {
						return
					}
				}

				// Keep connection alive
				var msg interface{}

				_ = conn.ReadJSON(&msg)
			}))
			defer server.Close()

			// Connect to the test server
			wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

			conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
			if err != nil {
				t.Fatalf("Dial error: %v", err)
			}

			defer func() { _ = conn.Close() }()

			// Create stream
			stream := NewWebSocketStream(conn)

			// Perform concurrent reads
			errChan := make(chan error, tt.numReads)

			for range tt.numReads {
				go func() {
					var obj interface{}
					errChan <- stream.ReadObject(&obj)
				}()
			}

			// Collect errors
			for range tt.numReads {
				if err := <-errChan; err != nil {
					t.Errorf("concurrent ReadObject() error = %v", err)
				}
			}
		})
	}
}
