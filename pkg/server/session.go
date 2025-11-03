package server

import (
	"sync/atomic"

	"github.com/sourcegraph/jsonrpc2"
)

// Session representa uma sessão interativa com um agente.
type Session struct {
	ID         string
	AgentID    string
	ClientConn *jsonrpc2.Conn // Connection to the client that initiated the session
	Input      chan []byte
	Output     chan []byte
	Done       chan struct{}
	closed     atomic.Bool
}

// NewSession cria uma nova sessão.
func NewSession(id, agentID string) *Session {
	return &Session{
		ID:      id,
		AgentID: agentID,
		Input:   make(chan []byte, 100),
		Output:  make(chan []byte, 100),
		Done:    make(chan struct{}),
	}
}

// NewSessionWithClient cria uma nova sessão com conexão do client.
func NewSessionWithClient(id, agentID string, clientConn *jsonrpc2.Conn) *Session {
	return &Session{
		ID:         id,
		AgentID:    agentID,
		ClientConn: clientConn,
		Input:      make(chan []byte, 100),
		Output:     make(chan []byte, 100),
		Done:       make(chan struct{}),
	}
}
