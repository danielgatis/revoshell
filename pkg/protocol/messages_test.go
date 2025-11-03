package protocol

import (
	"encoding/json"
	"testing"
)

func TestVersionCheckParams_JSON(t *testing.T) {
	tests := []struct {
		name     string
		params   VersionCheckParams
		wantJSON string
	}{
		{
			name: "full version check",
			params: VersionCheckParams{
				Version:   "v1.0.0",
				GitCommit: "abc1234",
			},
			wantJSON: `{"version":"v1.0.0","git_commit":"abc1234"}`,
		},
		{
			name: "version without commit",
			params: VersionCheckParams{
				Version: "v2.0.0",
			},
			wantJSON: `{"version":"v2.0.0"}`,
		},
		{
			name: "dev version",
			params: VersionCheckParams{
				Version:   "dev",
				GitCommit: "unknown",
			},
			wantJSON: `{"version":"dev","git_commit":"unknown"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test marshaling
			gotJSON, err := json.Marshal(tt.params)
			if err != nil {
				t.Fatalf("json.Marshal() error = %v", err)
			}

			if string(gotJSON) != tt.wantJSON {
				t.Errorf("json.Marshal() = %v, want %v", string(gotJSON), tt.wantJSON)
			}

			// Test unmarshaling
			var got VersionCheckParams
			if err := json.Unmarshal([]byte(tt.wantJSON), &got); err != nil {
				t.Fatalf("json.Unmarshal() error = %v", err)
			}

			if got.Version != tt.params.Version {
				t.Errorf("Version = %v, want %v", got.Version, tt.params.Version)
			}

			if got.GitCommit != tt.params.GitCommit {
				t.Errorf("GitCommit = %v, want %v", got.GitCommit, tt.params.GitCommit)
			}
		})
	}
}

func TestVersionCheckResult_JSON(t *testing.T) {
	tests := []struct {
		name     string
		result   VersionCheckResult
		wantJSON string
	}{
		{
			name: "compatible version",
			result: VersionCheckResult{
				Compatible: true,
				Version:    "v1.0.0",
				Message:    "Compatible",
			},
			wantJSON: `{"compatible":true,"version":"v1.0.0","message":"Compatible"}`,
		},
		{
			name: "incompatible version",
			result: VersionCheckResult{
				Compatible: false,
				Version:    "v0.9.0",
				Message:    "Version mismatch",
			},
			wantJSON: `{"compatible":false,"version":"v0.9.0","message":"Version mismatch"}`,
		},
		{
			name: "compatible without message",
			result: VersionCheckResult{
				Compatible: true,
				Version:    "v1.0.0",
			},
			wantJSON: `{"compatible":true,"version":"v1.0.0"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test marshaling
			gotJSON, err := json.Marshal(tt.result)
			if err != nil {
				t.Fatalf("json.Marshal() error = %v", err)
			}

			if string(gotJSON) != tt.wantJSON {
				t.Errorf("json.Marshal() = %v, want %v", string(gotJSON), tt.wantJSON)
			}

			// Test unmarshaling
			var got VersionCheckResult
			if err := json.Unmarshal([]byte(tt.wantJSON), &got); err != nil {
				t.Fatalf("json.Unmarshal() error = %v", err)
			}

			if got.Compatible != tt.result.Compatible {
				t.Errorf("Compatible = %v, want %v", got.Compatible, tt.result.Compatible)
			}

			if got.Version != tt.result.Version {
				t.Errorf("Version = %v, want %v", got.Version, tt.result.Version)
			}

			if got.Message != tt.result.Message {
				t.Errorf("Message = %v, want %v", got.Message, tt.result.Message)
			}
		})
	}
}

func TestRegisterParams_JSON(t *testing.T) {
	tests := []struct {
		name     string
		params   RegisterParams
		wantJSON string
	}{
		{
			name: "full registration",
			params: RegisterParams{
				AgentID:  "agent-001",
				Name:     "Test Agent",
				Hostname: "test-host",
				Platform: "linux",
			},
			wantJSON: `{"agent_id":"agent-001","name":"Test Agent","hostname":"test-host","platform":"linux"}`,
		},
		{
			name: "minimal registration",
			params: RegisterParams{
				AgentID: "agent-002",
				Name:    "Minimal Agent",
			},
			wantJSON: `{"agent_id":"agent-002","name":"Minimal Agent"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test marshaling
			gotJSON, err := json.Marshal(tt.params)
			if err != nil {
				t.Fatalf("json.Marshal() error = %v", err)
			}

			if string(gotJSON) != tt.wantJSON {
				t.Errorf("json.Marshal() = %v, want %v", string(gotJSON), tt.wantJSON)
			}

			// Test unmarshaling
			var got RegisterParams
			if err := json.Unmarshal([]byte(tt.wantJSON), &got); err != nil {
				t.Fatalf("json.Unmarshal() error = %v", err)
			}

			if got.AgentID != tt.params.AgentID {
				t.Errorf("AgentID = %v, want %v", got.AgentID, tt.params.AgentID)
			}

			if got.Name != tt.params.Name {
				t.Errorf("Name = %v, want %v", got.Name, tt.params.Name)
			}

			if got.Hostname != tt.params.Hostname {
				t.Errorf("Hostname = %v, want %v", got.Hostname, tt.params.Hostname)
			}

			if got.Platform != tt.params.Platform {
				t.Errorf("Platform = %v, want %v", got.Platform, tt.params.Platform)
			}
		})
	}
}

func TestSessionStartParams_JSON(t *testing.T) {
	tests := []struct {
		name     string
		params   SessionStartParams
		wantJSON string
	}{
		{
			name: "start with custom shell",
			params: SessionStartParams{
				AgentID:   "agent-001",
				SessionID: "session-123",
				Shell:     "/bin/zsh",
			},
			wantJSON: `{"agent_id":"agent-001","session_id":"session-123","shell":"/bin/zsh"}`,
		},
		{
			name: "start with default shell",
			params: SessionStartParams{
				SessionID: "session-456",
			},
			wantJSON: `{"session_id":"session-456"}`,
		},
		{
			name: "start from agent side",
			params: SessionStartParams{
				SessionID: "session-789",
				Shell:     "/bin/bash",
			},
			wantJSON: `{"session_id":"session-789","shell":"/bin/bash"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test marshaling
			gotJSON, err := json.Marshal(tt.params)
			if err != nil {
				t.Fatalf("json.Marshal() error = %v", err)
			}

			if string(gotJSON) != tt.wantJSON {
				t.Errorf("json.Marshal() = %v, want %v", string(gotJSON), tt.wantJSON)
			}

			// Test unmarshaling
			var got SessionStartParams
			if err := json.Unmarshal([]byte(tt.wantJSON), &got); err != nil {
				t.Fatalf("json.Unmarshal() error = %v", err)
			}

			if got.AgentID != tt.params.AgentID {
				t.Errorf("AgentID = %v, want %v", got.AgentID, tt.params.AgentID)
			}

			if got.SessionID != tt.params.SessionID {
				t.Errorf("SessionID = %v, want %v", got.SessionID, tt.params.SessionID)
			}

			if got.Shell != tt.params.Shell {
				t.Errorf("Shell = %v, want %v", got.Shell, tt.params.Shell)
			}
		})
	}
}

func TestSessionDataParams_JSON(t *testing.T) {
	tests := []struct {
		name     string
		params   SessionDataParams
		wantJSON string
	}{
		{
			name: "data with agent ID",
			params: SessionDataParams{
				AgentID:   "agent-001",
				SessionID: "session-123",
				Payload:   "bHMgLWxhCg==", // base64: "ls -la\n"
			},
			wantJSON: `{"agent_id":"agent-001","session_id":"session-123","payload":"bHMgLWxhCg=="}`,
		},
		{
			name: "data without agent ID",
			params: SessionDataParams{
				SessionID: "session-456",
				Payload:   "Y2QgL3RtcAo=", // base64: "cd /tmp\n"
			},
			wantJSON: `{"session_id":"session-456","payload":"Y2QgL3RtcAo="}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test marshaling
			gotJSON, err := json.Marshal(tt.params)
			if err != nil {
				t.Fatalf("json.Marshal() error = %v", err)
			}

			if string(gotJSON) != tt.wantJSON {
				t.Errorf("json.Marshal() = %v, want %v", string(gotJSON), tt.wantJSON)
			}

			// Test unmarshaling
			var got SessionDataParams
			if err := json.Unmarshal([]byte(tt.wantJSON), &got); err != nil {
				t.Fatalf("json.Unmarshal() error = %v", err)
			}

			if got.AgentID != tt.params.AgentID {
				t.Errorf("AgentID = %v, want %v", got.AgentID, tt.params.AgentID)
			}

			if got.SessionID != tt.params.SessionID {
				t.Errorf("SessionID = %v, want %v", got.SessionID, tt.params.SessionID)
			}

			if got.Payload != tt.params.Payload {
				t.Errorf("Payload = %v, want %v", got.Payload, tt.params.Payload)
			}
		})
	}
}

func TestSessionResizeParams_JSON(t *testing.T) {
	tests := []struct {
		name     string
		params   SessionResizeParams
		wantJSON string
	}{
		{
			name: "resize terminal",
			params: SessionResizeParams{
				AgentID:   "agent-001",
				SessionID: "session-123",
				Rows:      24,
				Cols:      80,
			},
			wantJSON: `{"agent_id":"agent-001","session_id":"session-123","rows":24,"cols":80}`,
		},
		{
			name: "large terminal",
			params: SessionResizeParams{
				SessionID: "session-456",
				Rows:      50,
				Cols:      200,
			},
			wantJSON: `{"session_id":"session-456","rows":50,"cols":200}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test marshaling
			gotJSON, err := json.Marshal(tt.params)
			if err != nil {
				t.Fatalf("json.Marshal() error = %v", err)
			}

			if string(gotJSON) != tt.wantJSON {
				t.Errorf("json.Marshal() = %v, want %v", string(gotJSON), tt.wantJSON)
			}

			// Test unmarshaling
			var got SessionResizeParams
			if err := json.Unmarshal([]byte(tt.wantJSON), &got); err != nil {
				t.Fatalf("json.Unmarshal() error = %v", err)
			}

			if got.AgentID != tt.params.AgentID {
				t.Errorf("AgentID = %v, want %v", got.AgentID, tt.params.AgentID)
			}

			if got.SessionID != tt.params.SessionID {
				t.Errorf("SessionID = %v, want %v", got.SessionID, tt.params.SessionID)
			}

			if got.Rows != tt.params.Rows {
				t.Errorf("Rows = %v, want %v", got.Rows, tt.params.Rows)
			}

			if got.Cols != tt.params.Cols {
				t.Errorf("Cols = %v, want %v", got.Cols, tt.params.Cols)
			}
		})
	}
}

func TestListAgentsResult_JSON(t *testing.T) {
	tests := []struct {
		name     string
		result   ListAgentsResult
		wantJSON string
	}{
		{
			name: "multiple agents",
			result: ListAgentsResult{
				Agents: []AgentInfo{
					{ID: "agent-001", Name: "Agent 1", SessionCount: 2},
					{ID: "agent-002", Name: "Agent 2", SessionCount: 0},
				},
				Count: 2,
			},
			wantJSON: `{"agents":[{"id":"agent-001","name":"Agent 1","session_count":2},{"id":"agent-002","name":"Agent 2","session_count":0}],"count":2}`,
		},
		{
			name: "no agents",
			result: ListAgentsResult{
				Agents: []AgentInfo{},
				Count:  0,
			},
			wantJSON: `{"agents":[],"count":0}`,
		},
		{
			name: "nil agents slice",
			result: ListAgentsResult{
				Agents: nil,
				Count:  0,
			},
			wantJSON: `{"agents":null,"count":0}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test marshaling
			gotJSON, err := json.Marshal(tt.result)
			if err != nil {
				t.Fatalf("json.Marshal() error = %v", err)
			}

			if string(gotJSON) != tt.wantJSON {
				t.Errorf("json.Marshal() = %v, want %v", string(gotJSON), tt.wantJSON)
			}

			// Test unmarshaling
			var got ListAgentsResult
			if err := json.Unmarshal([]byte(tt.wantJSON), &got); err != nil {
				t.Fatalf("json.Unmarshal() error = %v", err)
			}

			if got.Count != tt.result.Count {
				t.Errorf("Count = %v, want %v", got.Count, tt.result.Count)
			}

			if len(got.Agents) != len(tt.result.Agents) {
				t.Errorf("len(Agents) = %v, want %v", len(got.Agents), len(tt.result.Agents))
			}
		})
	}
}

func TestFileDownloadParams_JSON(t *testing.T) {
	tests := []struct {
		name     string
		params   FileDownloadParams
		wantJSON string
	}{
		{
			name: "download file",
			params: FileDownloadParams{
				AgentID:    "agent-001",
				RemotePath: "/etc/hosts",
			},
			wantJSON: `{"agent_id":"agent-001","remote_path":"/etc/hosts"}`,
		},
		{
			name: "download from home",
			params: FileDownloadParams{
				AgentID:    "agent-002",
				RemotePath: "/home/user/document.txt",
			},
			wantJSON: `{"agent_id":"agent-002","remote_path":"/home/user/document.txt"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test marshaling
			gotJSON, err := json.Marshal(tt.params)
			if err != nil {
				t.Fatalf("json.Marshal() error = %v", err)
			}

			if string(gotJSON) != tt.wantJSON {
				t.Errorf("json.Marshal() = %v, want %v", string(gotJSON), tt.wantJSON)
			}

			// Test unmarshaling
			var got FileDownloadParams
			if err := json.Unmarshal([]byte(tt.wantJSON), &got); err != nil {
				t.Fatalf("json.Unmarshal() error = %v", err)
			}

			if got.AgentID != tt.params.AgentID {
				t.Errorf("AgentID = %v, want %v", got.AgentID, tt.params.AgentID)
			}

			if got.RemotePath != tt.params.RemotePath {
				t.Errorf("RemotePath = %v, want %v", got.RemotePath, tt.params.RemotePath)
			}
		})
	}
}

func TestFileUploadParams_JSON(t *testing.T) {
	tests := []struct {
		name     string
		params   FileUploadParams
		wantJSON string
	}{
		{
			name: "upload with mode",
			params: FileUploadParams{
				AgentID:    "agent-001",
				RemotePath: "/tmp/test.txt",
				Content:    "SGVsbG8gV29ybGQK", // base64: "Hello World\n"
				Mode:       0644,
			},
			wantJSON: `{"agent_id":"agent-001","remote_path":"/tmp/test.txt","content":"SGVsbG8gV29ybGQK","mode":420}`,
		},
		{
			name: "upload without mode",
			params: FileUploadParams{
				AgentID:    "agent-002",
				RemotePath: "/tmp/script.sh",
				Content:    "IyEvYmluL2Jhc2gK", // base64: "#!/bin/bash\n"
			},
			wantJSON: `{"agent_id":"agent-002","remote_path":"/tmp/script.sh","content":"IyEvYmluL2Jhc2gK"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test marshaling
			gotJSON, err := json.Marshal(tt.params)
			if err != nil {
				t.Fatalf("json.Marshal() error = %v", err)
			}

			if string(gotJSON) != tt.wantJSON {
				t.Errorf("json.Marshal() = %v, want %v", string(gotJSON), tt.wantJSON)
			}

			// Test unmarshaling
			var got FileUploadParams
			if err := json.Unmarshal([]byte(tt.wantJSON), &got); err != nil {
				t.Fatalf("json.Unmarshal() error = %v", err)
			}

			if got.AgentID != tt.params.AgentID {
				t.Errorf("AgentID = %v, want %v", got.AgentID, tt.params.AgentID)
			}

			if got.RemotePath != tt.params.RemotePath {
				t.Errorf("RemotePath = %v, want %v", got.RemotePath, tt.params.RemotePath)
			}

			if got.Content != tt.params.Content {
				t.Errorf("Content = %v, want %v", got.Content, tt.params.Content)
			}

			if got.Mode != tt.params.Mode {
				t.Errorf("Mode = %v, want %v", got.Mode, tt.params.Mode)
			}
		})
	}
}

func TestFileDownloadResult_JSON(t *testing.T) {
	tests := []struct {
		name     string
		result   FileDownloadResult
		wantJSON string
	}{
		{
			name: "successful download",
			result: FileDownloadResult{
				Success:  true,
				Filename: "hosts",
				Content:  "MTI3LjAuMC4xIGxvY2FsaG9zdAo=", // base64
				Size:     19,
			},
			wantJSON: `{"success":true,"filename":"hosts","content":"MTI3LjAuMC4xIGxvY2FsaG9zdAo=","size":19}`,
		},
		{
			name: "failed download",
			result: FileDownloadResult{
				Success: false,
				Error:   "file not found",
			},
			wantJSON: `{"success":false,"filename":"","content":"","size":0,"error":"file not found"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test marshaling
			gotJSON, err := json.Marshal(tt.result)
			if err != nil {
				t.Fatalf("json.Marshal() error = %v", err)
			}

			if string(gotJSON) != tt.wantJSON {
				t.Errorf("json.Marshal() = %v, want %v", string(gotJSON), tt.wantJSON)
			}

			// Test unmarshaling
			var got FileDownloadResult
			if err := json.Unmarshal([]byte(tt.wantJSON), &got); err != nil {
				t.Fatalf("json.Unmarshal() error = %v", err)
			}

			if got.Success != tt.result.Success {
				t.Errorf("Success = %v, want %v", got.Success, tt.result.Success)
			}

			if got.Filename != tt.result.Filename {
				t.Errorf("Filename = %v, want %v", got.Filename, tt.result.Filename)
			}

			if got.Content != tt.result.Content {
				t.Errorf("Content = %v, want %v", got.Content, tt.result.Content)
			}

			if got.Size != tt.result.Size {
				t.Errorf("Size = %v, want %v", got.Size, tt.result.Size)
			}

			if got.Error != tt.result.Error {
				t.Errorf("Error = %v, want %v", got.Error, tt.result.Error)
			}
		})
	}
}

func TestMethodConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant string
		expected string
	}{
		{"MethodVersionCheck", MethodVersionCheck, "version.check"},
		{"MethodAgentRegister", MethodAgentRegister, "agent.register"},
		{"MethodAgentList", MethodAgentList, "agent.list"},
		{"MethodSessionStart", MethodSessionStart, "session.start"},
		{"MethodSessionData", MethodSessionData, "session.data"},
		{"MethodSessionStop", MethodSessionStop, "session.stop"},
		{"MethodSessionResize", MethodSessionResize, "session.resize"},
		{"MethodSessionList", MethodSessionList, "session.list"},
		{"MethodFileDownload", MethodFileDownload, "file.download"},
		{"MethodFileUpload", MethodFileUpload, "file.upload"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.expected {
				t.Errorf("%s = %v, want %v", tt.name, tt.constant, tt.expected)
			}
		})
	}
}
