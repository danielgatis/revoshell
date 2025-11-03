package server

import (
	"context"
	"encoding/json"

	"github.com/sourcegraph/jsonrpc2"

	"github.com/danielgaits/revoshell/pkg/protocol"
	"github.com/danielgaits/revoshell/pkg/version"
)

// Handler implements jsonrpc2.Handler to process device messages.
type Handler struct {
	server *Server
	device *Device
}

// NewHandler creates a new handler for a device.
func NewHandler(server *Server, device *Device) *Handler {
	return &Handler{
		server: server,
		device: device,
	}
}

func (h *Handler) Handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) {
	switch req.Method {
	case protocol.MethodVersionCheck:
		h.handleVersionCheck(ctx, conn, req)
	case protocol.MethodAgentRegister:
		h.handleRegister(ctx, conn, req)
	case protocol.MethodAgentList:
		h.handleList(ctx, conn, req)
	case protocol.MethodSessionList:
		h.handleSessionList(ctx, conn, req)
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

func (h *Handler) handleVersionCheck(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) {
	var params protocol.VersionCheckParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		log.Error().Err(err).Msg("Error decoding version.check")

		_ = conn.Reply(ctx, req.ID, nil)

		return
	}

	serverVersion := version.GetVersion()
	compatible := params.Version == serverVersion

	result := protocol.VersionCheckResult{
		Compatible: compatible,
		Version:    serverVersion,
	}

	if !compatible {
		result.Message = "Version mismatch: server=" + serverVersion + ", agent=" + params.Version
		log.Warn().
			Str("server_version", serverVersion).
			Str("agent_version", params.Version).
			Msg("Version mismatch detected")
	} else {
		log.Info().
			Str("version", serverVersion).
			Str("agent_commit", params.GitCommit).
			Msg("Version check passed")
	}

	_ = conn.Reply(ctx, req.ID, result)
}

func (h *Handler) handleRegister(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) {
	var params protocol.RegisterParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		log.Error().Err(err).Msg("Error decoding registration")

		_ = conn.Reply(ctx, req.ID, nil)

		return
	}

	// Check if device ID already exists
	if _, exists := h.server.GetDevice(params.AgentID); exists {
		log.Error().Str("device_id", params.AgentID).Msg("Device ID already registered")

		_ = conn.Reply(ctx, req.ID, protocol.RegisterResult{
			Success: false,
			Message: "Device ID already registered",
		})

		return
	}

	log.Info().
		Str("device_id", params.AgentID).
		Str("device_name", params.Name).
		Str("hostname", params.Hostname).
		Str("platform", params.Platform).
		Msg("Device registered")

	h.device.ID = params.AgentID
	h.device.Name = params.Name
	h.server.AddDevice(h.device)

	result := protocol.RegisterResult{
		Success: true,
		Message: "Device successfully registered",
	}

	_ = conn.Reply(ctx, req.ID, result)
}

func (h *Handler) handleList(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) {
	h.server.mu.RLock()

	// Snapshot all device data in single lock acquisition to reduce contention
	devices := make([]protocol.AgentInfo, 0, len(h.server.devices))
	for _, device := range h.server.devices {
		device.mu.RLock()
		devices = append(devices, protocol.AgentInfo{
			ID:           device.ID,
			Name:         device.Name,
			SessionCount: len(device.Sessions),
		})
		device.mu.RUnlock()
	}

	h.server.mu.RUnlock()

	result := protocol.ListAgentsResult{
		Agents: devices,
		Count:  len(devices),
	}

	_ = conn.Reply(ctx, req.ID, result)
}

func (h *Handler) handleSessionList(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) {
	h.server.mu.RLock()

	// Snapshot all session data in single lock acquisition to reduce contention
	sessions := make([]protocol.SessionInfo, 0)

	for _, device := range h.server.devices {
		device.mu.RLock()

		for sessionID := range device.Sessions {
			sessions = append(sessions, protocol.SessionInfo{
				SessionID: sessionID,
				AgentID:   device.ID,
			})
		}

		device.mu.RUnlock()
	}

	h.server.mu.RUnlock()

	result := protocol.ListSessionsResult{
		Sessions: sessions,
		Count:    len(sessions),
	}

	_ = conn.Reply(ctx, req.ID, result)
}

func (h *Handler) handleSessionData(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) {
	var params protocol.SessionDataParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		log.Error().Err(err).Msg("Error decoding session.data")

		return
	}

	// If AgentID is specified, route to that device (client->server->device)
	if params.AgentID != "" {
		targetDevice, exists := h.server.GetDevice(params.AgentID)
		if !exists {
			log.Warn().Str("device_id", params.AgentID).Msg("Device not found for session data")

			return
		}

		// Forward to device without AgentID field
		forwardParams := protocol.SessionDataParams{
			SessionID: params.SessionID,
			Payload:   params.Payload,
		}
		if err := targetDevice.Conn.Notify(ctx, protocol.MethodSessionData, forwardParams); err != nil {
			log.Error().Err(err).Msg("Error forwarding session.data to device")
		}

		return
	}

	// This is device->server->client (response from device)
	// Find the session and forward data to the client
	log.Debug().Str("session_id", params.SessionID).Msg("Received session data from device")

	if h.device != nil {
		session, exists := h.device.GetSession(params.SessionID)
		if !exists {
			log.Warn().Str("session_id", params.SessionID).Msg("Session not found for data")

			return
		}

		// Forward data to client via its connection
		if session.ClientConn != nil {
			forwardParams := protocol.SessionDataParams{
				SessionID: params.SessionID,
				Payload:   params.Payload,
			}
			if err := session.ClientConn.Notify(ctx, protocol.MethodSessionData, forwardParams); err != nil {
				log.Error().Err(err).Str("session_id", params.SessionID).Msg("Error forwarding data to client")
			}
		} else {
			log.Warn().Str("session_id", params.SessionID).Msg("Session has no client connection")
		}
	}
}

func (h *Handler) handleSessionStop(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) {
	var params protocol.SessionStopParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		log.Error().Err(err).Msg("Error decoding session.stop")

		return
	}

	// If AgentID is specified, route to that device (client->server->device)
	if params.AgentID != "" {
		targetDevice, exists := h.server.GetDevice(params.AgentID)
		if !exists {
			log.Warn().Str("device_id", params.AgentID).Msg("Device not found for session stop")

			return
		}

		// Forward to device without AgentID field
		forwardParams := protocol.SessionStopParams{
			SessionID: params.SessionID,
		}
		if err := targetDevice.Conn.Notify(ctx, protocol.MethodSessionStop, forwardParams); err != nil {
			log.Error().Err(err).Msg("Error forwarding session.stop to device")
		}

		return
	}

	// This is device->server notification
	log.Info().Str("session_id", params.SessionID).Msg("Session ended by device")

	// Find and notify the client, then remove the session
	if h.device != nil {
		session, exists := h.device.GetSession(params.SessionID)
		if exists {
			// Notify client that session ended
			if session.ClientConn != nil {
				forwardParams := protocol.SessionStopParams{
					SessionID: params.SessionID,
				}
				if err := session.ClientConn.Notify(ctx, protocol.MethodSessionStop, forwardParams); err != nil {
					log.Warn().Err(err).Str("session_id", params.SessionID).Msg("Error notifying client of session stop")
				}
			}

			// Remove session from tracking
			h.device.RemoveSession(params.SessionID)
		}
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

	log.Info().
		Str("device_id", params.AgentID).
		Str("path", params.RemotePath).
		Msg("Routing file download request")

	// Find target device
	targetDevice, exists := h.server.GetDevice(params.AgentID)
	if !exists {
		log.Error().Str("device_id", params.AgentID).Msg("Device not found")
		_ = conn.Reply(ctx, req.ID, protocol.FileDownloadResult{
			Success: false,
			Error:   "Device not found",
		})

		return
	}

	// Forward request to device
	var result protocol.FileDownloadResult
	if err := targetDevice.Conn.Call(ctx, protocol.MethodFileDownload, params, &result); err != nil {
		log.Error().Err(err).Msg("Error calling device")
		_ = conn.Reply(ctx, req.ID, protocol.FileDownloadResult{
			Success: false,
			Error:   err.Error(),
		})

		return
	}

	_ = conn.Reply(ctx, req.ID, result)
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

	log.Info().
		Str("device_id", params.AgentID).
		Str("path", params.RemotePath).
		Msg("Routing file upload request")

	// Find target device
	targetDevice, exists := h.server.GetDevice(params.AgentID)
	if !exists {
		log.Error().Str("device_id", params.AgentID).Msg("Device not found")
		_ = conn.Reply(ctx, req.ID, protocol.FileUploadResult{
			Success: false,
			Error:   "Device not found",
		})

		return
	}

	// Forward request to device
	var result protocol.FileUploadResult
	if err := targetDevice.Conn.Call(ctx, protocol.MethodFileUpload, params, &result); err != nil {
		log.Error().Err(err).Msg("Error calling device")
		_ = conn.Reply(ctx, req.ID, protocol.FileUploadResult{
			Success: false,
			Error:   err.Error(),
		})

		return
	}

	_ = conn.Reply(ctx, req.ID, result)
}

func (h *Handler) handleSessionStart(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) {
	var params protocol.SessionStartParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		log.Error().Err(err).Msg("Error decoding session.start")

		return
	}

	// If AgentID is specified, route to that device (client->server->device)
	if params.AgentID != "" {
		targetDevice, exists := h.server.GetDevice(params.AgentID)
		if !exists {
			log.Error().Str("device_id", params.AgentID).Msg("Device not found for session start")

			return
		}

		// Create session in device's session map to track it
		// conn is the client's connection (the one making this request)
		session := NewSessionWithClient(params.SessionID, params.AgentID, conn)
		targetDevice.AddSession(session)

		log.Info().
			Str("session_id", params.SessionID).
			Str("agent_id", params.AgentID).
			Msg("Session created on server")

		// Forward to device without AgentID field
		forwardParams := protocol.SessionStartParams{
			SessionID: params.SessionID,
			Shell:     params.Shell,
		}
		if err := targetDevice.Conn.Notify(ctx, protocol.MethodSessionStart, forwardParams); err != nil {
			log.Error().Err(err).Msg("Error forwarding session.start to device")
			// Remove session if forward failed
			targetDevice.RemoveSession(params.SessionID)
		}

		return
	}

	// Otherwise this is device->server, handle locally (not used anymore with new architecture)
	log.Warn().Msg("Received session.start without AgentID")
}

func (h *Handler) handleSessionResize(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) {
	var params protocol.SessionResizeParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		log.Error().Err(err).Msg("Error decoding session.resize")

		return
	}

	// If AgentID is specified, route to that device (client->server->device)
	if params.AgentID != "" {
		targetDevice, exists := h.server.GetDevice(params.AgentID)
		if !exists {
			log.Error().Str("device_id", params.AgentID).Msg("Device not found for session resize")

			return
		}

		// Forward to device without AgentID field
		forwardParams := protocol.SessionResizeParams{
			SessionID: params.SessionID,
			Rows:      params.Rows,
			Cols:      params.Cols,
		}
		if err := targetDevice.Conn.Notify(ctx, protocol.MethodSessionResize, forwardParams); err != nil {
			log.Error().Err(err).Msg("Error forwarding session.resize to device")
		}

		return
	}

	log.Warn().Msg("Received session.resize without AgentID")
}
