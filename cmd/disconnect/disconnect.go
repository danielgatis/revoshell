package disconnect

import (
	"context"
	"fmt"

	"github.com/gorilla/websocket"
	"github.com/sourcegraph/jsonrpc2"
	"github.com/spf13/cobra"

	"github.com/danielgatis/revoshell/pkg/protocol"
	"github.com/danielgatis/revoshell/pkg/transport"
	"github.com/danielgatis/revoshell/pkg/version"
)

var (
	serverURL   string
	securityKey string
)

// NewCommand creates the disconnect command.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "disconnect <session-id>",
		Short: "Disconnect from a session",
		Long:  `Disconnects from an active session using its session ID.`,
		Args:  cobra.ExactArgs(1),
		RunE:  runDisconnect,
	}

	cmd.Flags().StringVarP(&serverURL, "server", "s", "ws://localhost:8080/ws",
		"WebSocket server URL")
	cmd.Flags().StringVarP(&securityKey, "security-key", "k", "",
		"Security key for authentication (optional)")

	return cmd
}

func runDisconnect(cmd *cobra.Command, args []string) error {
	sessionID := args[0]

	// Connect to server via WebSocket
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

	// Create JSON-RPC connection
	stream := transport.NewWebSocketStream(wsConn)

	conn := jsonrpc2.NewConn(context.Background(), stream, jsonrpc2.HandlerWithError(func(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (interface{}, error) {
		return nil, nil
	}))

	defer func() {
		_ = conn.Close()
	}()

	// Stop session
	params := protocol.SessionStopParams{
		SessionID: sessionID,
	}

	if err := conn.Notify(context.Background(), protocol.MethodSessionStop, params); err != nil {
		_ = conn.Close()
		_ = wsConn.Close()

		return fmt.Errorf("failed to stop session: %w", err)
	}

	fmt.Printf("✓ Session '%s' disconnected\n", sessionID)

	return nil
}
