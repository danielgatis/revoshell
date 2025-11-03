package protocol

// Supported JSON-RPC 2.0 methods.
const (
	MethodVersionCheck  = "version.check"
	MethodAgentRegister = "agent.register"
	MethodAgentList     = "agent.list"
	MethodSessionStart  = "session.start"
	MethodSessionData   = "session.data"
	MethodSessionStop   = "session.stop"
	MethodSessionResize = "session.resize"
	MethodSessionList   = "session.list"
	MethodFileDownload  = "file.download"
	MethodFileUpload    = "file.upload"
)

// VersionCheckParams contains version information.
type VersionCheckParams struct {
	Version   string `json:"version"`
	GitCommit string `json:"git_commit,omitempty"`
}

// VersionCheckResult is the response to version check.
type VersionCheckResult struct {
	Compatible bool   `json:"compatible"`
	Version    string `json:"version"`
	Message    string `json:"message,omitempty"`
}

// RegisterParams contains the registration parameters for a device.
type RegisterParams struct {
	AgentID  string `json:"agent_id"`
	Name     string `json:"name"`
	Hostname string `json:"hostname,omitempty"`
	Platform string `json:"platform,omitempty"`
}

// SessionStartParams contains the parameters to start a session.
type SessionStartParams struct {
	AgentID   string `json:"agent_id,omitempty"` // for client->server routing
	SessionID string `json:"session_id"`
	Shell     string `json:"shell,omitempty"` // default: /bin/bash
}

// SessionDataParams contains session I/O data.
type SessionDataParams struct {
	AgentID   string `json:"agent_id,omitempty"` // for client->server routing
	SessionID string `json:"session_id"`
	Payload   string `json:"payload"` // data in base64 or string
}

// SessionStopParams contains the parameters to stop a session.
type SessionStopParams struct {
	AgentID   string `json:"agent_id,omitempty"` // for client->server routing
	SessionID string `json:"session_id"`
}

// SessionResizeParams contains the parameters to resize the terminal.
type SessionResizeParams struct {
	AgentID   string `json:"agent_id,omitempty"` // for client->server routing
	SessionID string `json:"session_id"`
	Rows      uint16 `json:"rows"`
	Cols      uint16 `json:"cols"`
}

// RegisterResult is the response to device registration.
type RegisterResult struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

// ListAgentsParams is empty as it requires no parameters.
type ListAgentsParams struct{}

// AgentInfo contains information about a device.
type AgentInfo struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	SessionCount int    `json:"session_count"`
}

// ListAgentsResult is the result of listing devices.
type ListAgentsResult struct {
	Agents []AgentInfo `json:"agents"`
	Count  int         `json:"count"`
}

// ListSessionsParams parameters for listing sessions.
type ListSessionsParams struct{}

// SessionInfo information about a session.
type SessionInfo struct {
	SessionID string `json:"session_id"`
	AgentID   string `json:"agent_id"`
}

// ListSessionsResult result of listing sessions.
type ListSessionsResult struct {
	Sessions []SessionInfo `json:"sessions"`
	Count    int           `json:"count"`
}

// FileDownloadParams parameters to download a file from agent.
type FileDownloadParams struct {
	AgentID    string `json:"agent_id"`
	RemotePath string `json:"remote_path"`
}

// FileDownloadResult result of file download.
type FileDownloadResult struct {
	Success  bool   `json:"success"`
	Filename string `json:"filename"`
	Content  string `json:"content"` // base64 encoded
	Size     int64  `json:"size"`
	Error    string `json:"error,omitempty"`
}

// FileUploadParams parameters to upload a file to agent.
type FileUploadParams struct {
	AgentID    string `json:"agent_id"`
	RemotePath string `json:"remote_path"`
	Content    string `json:"content"`        // base64 encoded
	Mode       uint32 `json:"mode,omitempty"` // file permissions
}

// FileUploadResult result of file upload.
type FileUploadResult struct {
	Success bool   `json:"success"`
	Size    int64  `json:"size"`
	Error   string `json:"error,omitempty"`
}
