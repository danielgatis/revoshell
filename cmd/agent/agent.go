package agent

import (
	"fmt"
	"os"

	"github.com/denisbrodbeck/machineid"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"

	agentPkg "github.com/danielgaits/revoshell/pkg/agent"
	"github.com/danielgaits/revoshell/pkg/logging"
)

var (
	agentID     string
	deviceName  string
	serverURL   string
	securityKey string
	agentLog    zerolog.Logger
)

// NewCommand creates the agent command.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "agent",
		Short: "Start a RevoShell agent",
		Long: `Starts an agent that connects to the central server.

The agent maintains a persistent reverse connection with the server,
allowing remote shell access without needing open ports.

If no ID is provided, the machine's unique hardware ID will be used automatically.
If no name is provided, the machine's hostname will be used automatically.

Examples:
  revossh agent
  revossh agent --name "My Laptop"
  revossh agent --id laptop-home --name "My Laptop"
  revossh agent --server ws://server.com:8080/ws`,
		Run: runAgent,
	}

	cmd.Flags().StringVarP(&agentID, "id", "i", "",
		"Unique agent ID (optional, uses machine ID if not provided)")
	cmd.Flags().StringVarP(&deviceName, "name", "n", "",
		"Friendly device name for display (optional, uses hostname if not provided)")
	cmd.Flags().StringVarP(&serverURL, "server", "s", "ws://localhost:8080/ws",
		"WebSocket server URL")
	cmd.Flags().StringVarP(&securityKey, "security-key", "k", "",
		"Security key for authentication (optional)")

	return cmd
}

func runAgent(cmd *cobra.Command, args []string) {
	_ = args

	agentLog = logging.InitWithComponent("CMD")
	agentLog.Info().Str("version", cmd.Root().Version).Msg("Starting RevoSSH Agent")

	// If agent ID is not provided, use machine ID
	if agentID == "" {
		id, err := machineid.ID()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to get machine ID: %v\n", err)
			os.Exit(1)
		}

		agentID = id
		agentLog.Info().Str("id", agentID).Msg("Using machine ID as agent ID")
	}

	// If device name is not provided, use hostname
	if deviceName == "" {
		hostname, err := os.Hostname()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to get hostname: %v\n", err)
			os.Exit(1)
		}

		deviceName = hostname
		agentLog.Info().Str("name", deviceName).Msg("Using hostname as device name")
	}

	agentLog.Info().Str("id", agentID).Str("name", deviceName).Msg("Agent configuration")

	// Create and start agent with security key
	ag := agentPkg.NewWithSecurityKey(agentID, deviceName, serverURL, securityKey)
	ag.Run()
}
