package upload

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"

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
	fileMode    uint32
)

// NewCommand creates the upload command.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "upload <agent-id> <local-path> <remote-path>",
		Short: "Upload a file to an agent",
		Long:  `Uploads a file from the local machine to the specified agent.`,
		Args:  cobra.ExactArgs(3),
		RunE:  runUpload,
	}

	cmd.Flags().StringVarP(&serverURL, "server", "s", "ws://localhost:8080/ws",
		"WebSocket server URL")
	cmd.Flags().StringVarP(&securityKey, "security-key", "k", "",
		"Security key for authentication (optional)")

	cmd.Flags().Uint32VarP(&fileMode, "mode", "m", 0644,
		"File permissions mode (octal)")

	return cmd
}

func runUpload(cmd *cobra.Command, args []string) error {
	agentID := args[0]
	localPath := args[1]
	remotePath := args[2]

	fmt.Printf("Uploading to agent '%s': %s -> %s\n", agentID, localPath, remotePath)

	// Read local file
	content, err := os.ReadFile(localPath)
	if err != nil {
		return fmt.Errorf("failed to read local file: %w", err)
	}

	// Encode to base64
	encoded := base64.StdEncoding.EncodeToString(content)

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

	params := protocol.FileUploadParams{
		AgentID:    agentID,
		RemotePath: remotePath,
		Content:    encoded,
		Mode:       fileMode,
	}

	var result protocol.FileUploadResult
	if err := conn.Call(context.Background(), protocol.MethodFileUpload, params, &result); err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}

	if !result.Success {
		return fmt.Errorf("upload failed: %s", result.Error)
	}

	fmt.Printf("✓ File uploaded successfully\n")
	fmt.Printf("  Local:  %s\n", localPath)
	fmt.Printf("  Remote: %s\n", remotePath)
	fmt.Printf("  Size:   %d bytes\n", result.Size)

	return nil
}
