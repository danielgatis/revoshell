package agent

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/creack/pty"
	"github.com/sourcegraph/jsonrpc2"

	"github.com/danielgaits/revoshell/pkg/protocol"
)

// Handler implements jsonrpc2.Handler to process server commands.
type Handler struct {
	agent *Agent
}

// NewHandler creates a new handler for the agent.
func NewHandler(agent *Agent) *Handler {
	return &Handler{
		agent: agent,
	}
}

func (h *Handler) Handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) {
	switch req.Method {
	case protocol.MethodSessionStart:
		h.handleSessionStart(ctx, conn, req)
	case protocol.MethodSessionData:
		h.handleSessionData(ctx, conn, req)
	case protocol.MethodSessionStop:
		h.handleSessionStop(ctx, conn, req)
	case protocol.MethodSessionResize:
		h.handleSessionResize(ctx, conn, req)
	case protocol.MethodFileDownload:
		h.handleFileDownload(ctx, conn, req)
	case protocol.MethodFileUpload:
		h.handleFileUpload(ctx, conn, req)
	default:
		log.Warn().Str("method", req.Method).Msg("Unknown method")
	}
}

func (h *Handler) handleSessionStart(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) {
	var params protocol.SessionStartParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		log.Error().Err(err).Msg("Error decoding session.start")

		return
	}

	log.Info().Str("session_id", params.SessionID).Msg("Starting session")

	shell := params.Shell
	if shell == "" {
		shell = "/bin/bash"
	}

	// Validate shell against whitelist
	allowedShells := []string{"/bin/bash", "/bin/sh", "/bin/zsh", "/usr/bin/fish"}
	shellAllowed := false

	for _, allowed := range allowedShells {
		if shell == allowed {
			shellAllowed = true

			break
		}
	}

	if !shellAllowed {
		log.Error().Str("shell", shell).Msg("Invalid shell requested")

		return
	}

	// Create process with pty
	cmd := exec.Command(shell)

	cmd.Env = append(os.Environ(), "TERM=xterm-256color")

	ptmx, err := pty.Start(cmd)
	if err != nil {
		log.Error().Err(err).Msg("Error creating pty")

		return
	}

	// Ensure cleanup on any failure
	sessionStarted := false

	defer func() {
		if !sessionStarted {
			if cmd.Process != nil {
				_ = cmd.Process.Kill()
			}

			_ = ptmx.Close()
		}
	}()

	session := NewSession(params.SessionID, ptmx, cmd)
	sessionStarted = true

	// Start goroutine BEFORE adding to map to ensure cleanup always happens
	go h.readPtyOutput(conn, session)

	// Add to map after goroutine is running (goroutine handles cleanup)
	h.agent.AddSession(session)
}

func (h *Handler) readPtyOutput(conn *jsonrpc2.Conn, session *Session) {
	defer func() {
		// Recover from panics to ensure cleanup always happens
		if r := recover(); r != nil {
			log.Error().
				Interface("panic", r).
				Str("session_id", session.ID).
				Msg("Panic in readPtyOutput")
		}

		h.agent.RemoveSession(session.ID)
		session.Close()

		// Notify server about termination
		stopParams := protocol.SessionStopParams{
			SessionID: session.ID,
		}
		if err := conn.Notify(context.Background(), protocol.MethodSessionStop, stopParams); err != nil {
			log.Warn().Err(err).Msg("Failed to notify session stop")
		}

		log.Info().Str("session_id", session.ID).Msg("Session ended")
	}()

	// Use separate goroutine for PTY reads to enable graceful shutdown
	// Channel to pass PTY data from blocking read goroutine
	ptyData := make(chan []byte, 10)
	ptyErr := make(chan error, 1)

	// Blocking PTY reader goroutine
	go func() {
		buf := make([]byte, 1024)

		for {
			n, err := session.Pty.Read(buf)
			if err != nil {
				ptyErr <- err

				return
			}

			if n > 0 {
				// Copy data before sending to avoid race
				data := make([]byte, n)
				copy(data, buf[:n])

				select {
				case ptyData <- data:
					// Sent successfully
				case <-session.Done:
					// Session closing, exit reader
					return
				}
			}
		}
	}()

	// Main loop with Done channel monitoring
	for {
		select {
		case data := <-ptyData:
			// Send PTY output to server
			dataParams := protocol.SessionDataParams{
				SessionID: session.ID,
				Payload:   string(data),
			}
			if err := conn.Notify(context.Background(), protocol.MethodSessionData, dataParams); err != nil {
				log.Error().Err(err).Msg("Error sending data")

				return
			}

		case err := <-ptyErr:
			// PTY read error (normal on close or process exit)
			if err != io.EOF {
				log.Error().Err(err).Msg("Error reading pty")
			}

			return

		case <-session.Done:
			// Session explicitly closed, exit gracefully
			log.Debug().Str("session_id", session.ID).Msg("Session Done signal received, exiting read loop")

			return
		}
	}
}

func (h *Handler) handleSessionData(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) {
	var params protocol.SessionDataParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		log.Error().Err(err).Msg("Error decoding session.data")

		return
	}

	session, exists := h.agent.GetSession(params.SessionID)
	if !exists {
		log.Warn().Str("session_id", params.SessionID).Msg("Session not found")

		return
	}

	// Write data to pty
	if err := session.Write([]byte(params.Payload)); err != nil {
		log.Error().Err(err).Msg("Error writing to pty")
	}
}

func (h *Handler) handleSessionStop(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) {
	var params protocol.SessionStopParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		log.Error().Err(err).Msg("Error decoding session.stop")

		return
	}

	log.Info().Str("session_id", params.SessionID).Msg("Ending session")

	session, exists := h.agent.GetSession(params.SessionID)
	if exists {
		session.Close()
		h.agent.RemoveSession(params.SessionID)
	}
}

func (h *Handler) handleSessionResize(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) {
	var params protocol.SessionResizeParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		log.Error().Err(err).Msg("Error decoding session.resize")

		return
	}

	session, exists := h.agent.GetSession(params.SessionID)
	if !exists {
		return
	}

	winsize := &pty.Winsize{
		Rows: params.Rows,
		Cols: params.Cols,
	}

	if err := pty.Setsize(session.Pty, winsize); err != nil {
		log.Error().Err(err).Msg("Error resizing terminal")
	}
}

func (h *Handler) handleFileDownload(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) {
	var params protocol.FileDownloadParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		log.Error().Err(err).Msg("Error decoding file.download")

		if err := conn.Reply(ctx, req.ID, protocol.FileDownloadResult{
			Success: false,
			Error:   "Invalid parameters",
		}); err != nil {
			log.Warn().Err(err).Msg("Failed to send error reply")
		}

		return
	}

	// Read file
	content, err := os.ReadFile(params.RemotePath)
	if err != nil {
		if err := conn.Reply(ctx, req.ID, protocol.FileDownloadResult{
			Success: false,
			Error:   err.Error(),
		}); err != nil {
			log.Warn().Err(err).Msg("Failed to send error reply")
		}

		return
	}

	// Encode to base64
	encoded := base64.StdEncoding.EncodeToString(content)

	result := protocol.FileDownloadResult{
		Success:  true,
		Filename: filepath.Base(params.RemotePath),
		Content:  encoded,
		Size:     int64(len(content)),
	}

	if err := conn.Reply(ctx, req.ID, result); err != nil {
		log.Warn().Err(err).Msg("Failed to send file download result")
	}
}

func (h *Handler) handleFileUpload(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) {
	var params protocol.FileUploadParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		log.Error().Err(err).Msg("Error decoding file.upload")

		if err := conn.Reply(ctx, req.ID, protocol.FileUploadResult{
			Success: false,
			Error:   "Invalid parameters",
		}); err != nil {
			log.Warn().Err(err).Msg("Failed to send error reply")
		}

		return
	}

	// Decode base64
	content, err := base64.StdEncoding.DecodeString(params.Content)
	if err != nil {
		if err := conn.Reply(ctx, req.ID, protocol.FileUploadResult{
			Success: false,
			Error:   "Invalid base64 content",
		}); err != nil {
			log.Warn().Err(err).Msg("Failed to send error reply")
		}

		return
	}

	// Determine file mode
	mode := os.FileMode(0644)
	if params.Mode > 0 {
		mode = os.FileMode(params.Mode)
	}

	// Write file
	if err := os.WriteFile(params.RemotePath, content, mode); err != nil {
		if err := conn.Reply(ctx, req.ID, protocol.FileUploadResult{
			Success: false,
			Error:   err.Error(),
		}); err != nil {
			log.Warn().Err(err).Msg("Failed to send error reply")
		}

		return
	}

	result := protocol.FileUploadResult{
		Success: true,
		Size:    int64(len(content)),
	}

	if err := conn.Reply(ctx, req.ID, result); err != nil {
		log.Warn().Err(err).Msg("Failed to send file upload result")
	}
}
