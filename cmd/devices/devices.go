package devices

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

// NewCommand creates the devices command.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "devices",
		Short: "List devices (agents) connected to server",
		Long:  `Lists all devices (agents) currently connected to the RevoSSH server via JSON-RPC.`,
		RunE:  runDevices,
	}

	cmd.Flags().StringVarP(&serverURL, "server", "s", "ws://localhost:8080/ws",
		"WebSocket server URL")
	cmd.Flags().StringVarP(&securityKey, "security-key", "k", "",
		"Security key for authentication (optional)")

	return cmd
}

func runDevices(cmd *cobra.Command, args []string) error {
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

	// Call agent.list
	var result protocol.ListAgentsResult
	if err := conn.Call(context.Background(), protocol.MethodAgentList, protocol.ListAgentsParams{}, &result); err != nil {
		_ = conn.Close()
		_ = wsConn.Close()

		return fmt.Errorf("failed to list agents: %w", err)
	}

	// Display results
	if result.Count == 0 {
		fmt.Println("No devices connected.")

		return nil
	}

	fmt.Printf("Connected devices: %d\n\n", result.Count)

	for i, agent := range result.Agents {
		if i > 0 {
			fmt.Println()
		}

		fmt.Printf("Device ID:       %s\n", agent.ID)
		fmt.Printf("Device Name:     %s\n", agent.Name)
		fmt.Printf("Active Sessions: %d\n", agent.SessionCount)
	}

	fmt.Println()

	return nil
}
