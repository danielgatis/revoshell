package connect

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
	"github.com/sourcegraph/jsonrpc2"
	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/danielgatis/revoshell/pkg/logging"
	"github.com/danielgatis/revoshell/pkg/protocol"
	"github.com/danielgatis/revoshell/pkg/transport"
	"github.com/danielgatis/revoshell/pkg/version"
)

var (
	serverURL   string
	securityKey string
	shellPath   string
	log         zerolog.Logger
)

func init() {
	log = logging.InitWithComponent("CONNECT")
}

// NewCommand creates the connect command.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "connect <agent-id>",
		Short: "Connect to an agent and start session",
		Long:  `Connects to an agent and starts a new interactive session.`,
		Args:  cobra.ExactArgs(1),
		RunE:  runConnect,
	}

	cmd.Flags().StringVarP(&serverURL, "server", "s", "ws://localhost:8080/ws",
		"WebSocket server URL")
	cmd.Flags().StringVarP(&securityKey, "security-key", "k", "",
		"Security key for authentication (optional)")
	cmd.Flags().StringVar(&shellPath, "shell", "",
		"Shell to use (default: /bin/bash)")

	return cmd
}

// clampToUint16 clamps an integer value to fit in uint16 range.
func clampToUint16(val int) uint16 {
	if val > 65535 {
		return 65535
	}

	if val < 0 {
		return 0
	}

	return uint16(val)
}

// stringToBytes converts a string to []byte without allocation.
// Safe to use when the byte slice is immediately consumed and not retained.
func stringToBytes(s string) []byte {
	return unsafe.Slice(unsafe.StringData(s), len(s))
}

func runConnect(cmd *cobra.Command, args []string) error {
	agentID := args[0]
	sessionID := uuid.New().String()

	// Connect to server via WebSocket with security key
	headers := make(map[string][]string)

	headers["X-Client-Version"] = []string{version.GetVersion()}
	if securityKey != "" {
		headers["X-Security-Key"] = []string{securityKey}
	}

	wsConn, _, err := websocket.DefaultDialer.Dial(serverURL, headers)
	if err != nil {
		return fmt.Errorf("failed to connect to server: %w", err)
	}

	defer func() {
		_ = wsConn.Close()
	}()

	sessionOutput := make(chan []byte, 100000)
	sessionDone := make(chan struct{})

	var sessionDoneOnce sync.Once

	// Create JSON-RPC connection with handler for incoming data
	stream := transport.NewWebSocketStream(wsConn)

	conn := jsonrpc2.NewConn(context.Background(), stream, jsonrpc2.HandlerWithError(func(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (interface{}, error) {
		switch req.Method {
		case protocol.MethodSessionData:
			var params protocol.SessionDataParams
			if req.Params != nil {
				if err := json.Unmarshal(*req.Params, &params); err == nil {
					if params.SessionID == sessionID {
						// Non-blocking send to prevent deadlock when buffer is full
						select {
						case sessionOutput <- stringToBytes(params.Payload):
							if len(sessionOutput) > 80000 {
								log.Warn().
									Int("buffer_size", len(sessionOutput)).
									Int("capacity", cap(sessionOutput)).
									Msg("Session output buffer is >80% full - consider slowing down output")
							}
						default:
							// Channel full - drop data to prevent blocking
							// If this happens frequently, there's a deeper issue with terminal rendering speed
							log.Error().Msg("CRITICAL: Session output buffer full (100000 items), dropping data - terminal may be frozen")
						}
					}
				}
			}
		case protocol.MethodSessionStop:
			sessionDoneOnce.Do(func() { close(sessionDone) })
		}

		return nil, nil
	}))

	defer func() {
		_ = conn.Close()
	}()

	// Start session
	shell := shellPath
	if shell == "" {
		shell = "/bin/bash"
	}

	params := protocol.SessionStartParams{
		AgentID:   agentID,
		SessionID: sessionID,
		Shell:     shell,
	}

	if err := conn.Notify(context.Background(), protocol.MethodSessionStart, params); err != nil {
		_ = conn.Close()
		_ = wsConn.Close()

		return fmt.Errorf("failed to start session: %w", err)
	}

	fmt.Printf("Connected to agent '%s' (session: %s)\n", agentID, sessionID)
	fmt.Printf("Press Ctrl+D or type 'exit' to disconnect\n\n")

	// Get initial terminal size and send to agent
	if width, height, err := term.GetSize(int(os.Stdin.Fd())); err == nil {
		resizeParams := protocol.SessionResizeParams{
			AgentID:   agentID,
			SessionID: sessionID,
			Rows:      clampToUint16(height),
			Cols:      clampToUint16(width),
		}
		if err := conn.Notify(context.Background(), protocol.MethodSessionResize, resizeParams); err != nil {
			log.Warn().Err(err).Msg("Failed to send initial terminal size")
		}
	}

	// Setup terminal raw mode
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return fmt.Errorf("failed to set raw mode: %w", err)
	}

	defer func() {
		if err := term.Restore(int(os.Stdin.Fd()), oldState); err != nil {
			log.Warn().Err(err).Msg("Failed to restore terminal")
		}
	}()

	// Setup signal handler for terminal resize (SIGWINCH)
	sigwinch := make(chan os.Signal, 1)

	signal.Notify(sigwinch, syscall.SIGWINCH)
	defer signal.Stop(sigwinch)

	// Create context for goroutine lifecycle management
	resizeCtx, resizeCancel := context.WithCancel(context.Background())
	defer resizeCancel()

	// Goroutine to handle terminal resize
	go func() {
		for {
			select {
			case <-sigwinch:
				if width, height, err := term.GetSize(int(os.Stdin.Fd())); err == nil {
					resizeParams := protocol.SessionResizeParams{
						AgentID:   agentID,
						SessionID: sessionID,
						Rows:      clampToUint16(height),
						Cols:      clampToUint16(width),
					}
					if err := conn.Notify(context.Background(), protocol.MethodSessionResize, resizeParams); err != nil {
						log.Warn().Err(err).Msg("Failed to send terminal resize")
					}
				}
			case <-resizeCtx.Done():
				return
			case <-sessionDone:
				return
			}
		}
	}()

	// Channel for stdin data to enable non-blocking stdin reads
	stdinChan := make(chan []byte, 10)
	stdinErr := make(chan error, 1)

	go func() {
		buf := make([]byte, 1024)

		for {
			n, err := os.Stdin.Read(buf)
			if err != nil {
				if err != io.EOF {
					select {
					case stdinErr <- err:
					case <-sessionDone:
					}
				}

				return
			}

			if n > 0 {
				// Copy data before sending to channel
				data := make([]byte, n)
				copy(data, buf[:n])

				select {
				case stdinChan <- data:
					// Sent successfully
				case <-sessionDone:
					return
				}
			}
		}
	}()

	// Goroutine to process stdin data and send to agent
	go func() {
		defer func() {
			// Non-blocking notification with timeout
			go func() {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				stopParams := protocol.SessionStopParams{
					AgentID:   agentID,
					SessionID: sessionID,
				}
				_ = conn.Notify(ctx, protocol.MethodSessionStop, stopParams)
			}()

			// Safely close sessionDone using sync.Once
			sessionDoneOnce.Do(func() { close(sessionDone) })
		}()

		for {
			select {
			case data := <-stdinChan:
				dataParams := protocol.SessionDataParams{
					AgentID:   agentID,
					SessionID: sessionID,
					Payload:   string(data),
				}
				if err := conn.Notify(context.Background(), protocol.MethodSessionData, dataParams); err != nil {
					fmt.Fprintf(os.Stderr, "\nError sending data: %v\n", err)

					return
				}
			case <-sessionDone:
				return
			}
		}
	}()

	// Main loop: read from agent and write to stdout
	for {
		select {
		case data := <-sessionOutput:
			if _, err := os.Stdout.Write(data); err != nil {
				log.Warn().Err(err).Msg("Failed to write to stdout")
			}
		case <-sessionDone:
			fmt.Printf("\r\nSession ended\r\n")

			return nil
		case <-conn.DisconnectNotify():
			fmt.Printf("\r\nConnection lost\r\n")

			return nil
		}
	}
}
