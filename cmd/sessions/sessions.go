package sessions

import (
	"context"
	"fmt"

	"github.com/gorilla/websocket"
	"github.com/sourcegraph/jsonrpc2"
	"github.com/spf13/cobra"

	"github.com/danielgaits/revoshell/pkg/protocol"
	"github.com/danielgaits/revoshell/pkg/transport"
	"github.com/danielgaits/revoshell/pkg/version"
)

var (
	serverURL   string
	securityKey string
)

// NewCommand creates the sessions command.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sessions",
		Short: "List all active sessions",
		Long:  `Lists all active sessions across all connected agents.`,
		RunE:  runSessions,
	}

	cmd.Flags().StringVarP(&serverURL, "server", "s", "ws://localhost:8080/ws",
		"WebSocket server URL")
	cmd.Flags().StringVarP(&securityKey, "security-key", "k", "",
		"Security key for authentication (optional)")

	return cmd
}

func runSessions(cmd *cobra.Command, args []string) error {
	_ = cmd
	_ = args

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

	// Call sessions.list
	var result protocol.ListSessionsResult
	if err := conn.Call(context.Background(), protocol.MethodSessionList, protocol.ListSessionsParams{}, &result); err != nil {
		_ = conn.Close()
		_ = wsConn.Close()

		return fmt.Errorf("failed to list sessions: %w", err)
	}

	// Display results
	if result.Count == 0 {
		fmt.Println("No active sessions.")

		return nil
	}

	fmt.Printf("Active sessions: %d\n\n", result.Count)

	for i, session := range result.Sessions {
		if i > 0 {
			fmt.Println()
		}

		fmt.Printf("Session ID: %s\n", session.SessionID)
		fmt.Printf("Agent ID:   %s\n", session.AgentID)
	}

	fmt.Println()

	return nil
}
