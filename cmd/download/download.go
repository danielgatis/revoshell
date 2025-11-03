package download

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"

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
	outputPath  string
)

// NewCommand creates the download command.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "download <agent-id> <remote-path>",
		Short: "Download a file from an agent",
		Long:  `Downloads a file from the specified agent to the local machine.`,
		Args:  cobra.ExactArgs(2),
		RunE:  runDownload,
	}

	cmd.Flags().StringVarP(&serverURL, "server", "s", "ws://localhost:8080/ws",
		"WebSocket server URL")
	cmd.Flags().StringVarP(&securityKey, "security-key", "k", "",
		"Security key for authentication (optional)")

	cmd.Flags().StringVarP(&outputPath, "output", "o", "",
		"Output file path (default: same filename in current directory)")

	return cmd
}

func runDownload(cmd *cobra.Command, args []string) error {
	agentID := args[0]
	remotePath := args[1]

	fmt.Printf("Downloading from agent '%s': %s\n", agentID, remotePath)

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

	params := protocol.FileDownloadParams{
		AgentID:    agentID,
		RemotePath: remotePath,
	}

	var result protocol.FileDownloadResult
	if err := conn.Call(context.Background(), protocol.MethodFileDownload, params, &result); err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}

	if !result.Success {
		return fmt.Errorf("download failed: %s", result.Error)
	}

	// Decode base64 content
	content, err := base64.StdEncoding.DecodeString(result.Content)
	if err != nil {
		return fmt.Errorf("failed to decode file content: %w", err)
	}

	// Determine output path
	output := outputPath
	if output == "" {
		output = filepath.Base(remotePath)
	}

	// Write file to disk
	if err := os.WriteFile(output, content, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	fmt.Printf("✓ File downloaded successfully\n")
	fmt.Printf("  Remote: %s\n", remotePath)
	fmt.Printf("  Local:  %s\n", output)
	fmt.Printf("  Size:   %d bytes\n", result.Size)

	return nil
}
