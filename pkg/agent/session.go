package agent

import (
	"os"
	"os/exec"
	"sync"
	"sync/atomic"
)

// Session representa uma sessão ativa de shell.
type Session struct {
	ID     string
	Pty    *os.File
	Cmd    *exec.Cmd
	Done   chan struct{}
	closed atomic.Bool
	mu     sync.Mutex
}

// NewSession cria uma nova sessão.
func NewSession(id string, pty *os.File, cmd *exec.Cmd) *Session {
	return &Session{
		ID:   id,
		Pty:  pty,
		Cmd:  cmd,
		Done: make(chan struct{}),
	}
}

// Write escreve dados no PTY.
func (s *Session) Write(data []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.Pty.Write(data)

	return err
}

// Close encerra a sessão.
func (s *Session) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.Cmd.Process != nil {
		_ = s.Cmd.Process.Kill()
	}

	_ = s.Pty.Close()

	// Safely close Done channel
	if s.closed.CompareAndSwap(false, true) {
		close(s.Done)
	}
}
