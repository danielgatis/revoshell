package transport

import (
	"sync"

	"github.com/gorilla/websocket"
	"github.com/sourcegraph/jsonrpc2"
)

// WebSocketStream implementa jsonrpc2.ObjectStream sobre WebSocket.
type WebSocketStream struct {
	conn    *websocket.Conn
	writeMu sync.Mutex
	readMu  sync.Mutex
}

// NewWebSocketStream cria um novo stream sobre WebSocket.
func NewWebSocketStream(conn *websocket.Conn) jsonrpc2.ObjectStream {
	return &WebSocketStream{conn: conn}
}

func (s *WebSocketStream) WriteObject(obj interface{}) error {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()

	return s.conn.WriteJSON(obj)
}

func (s *WebSocketStream) ReadObject(v interface{}) error {
	s.readMu.Lock()
	defer s.readMu.Unlock()

	return s.conn.ReadJSON(v)
}

func (s *WebSocketStream) Close() error {
	return s.conn.Close()
}
