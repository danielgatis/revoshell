package server

import (
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"

	"github.com/danielgaits/revoshell/pkg/logging"
	serverPkg "github.com/danielgaits/revoshell/pkg/server"
)

var (
	serverAddr  string
	securityKey string
	tlsCert     string
	tlsKey      string
	log         zerolog.Logger
)

// NewCommand creates the server command.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "server",
		Short: "Start the RevoShell server",
		Long: `Starts the central server that accepts agent connections.

The server acts as a relay between clients and agents.

Examples:
  revoshell server
  revoshell server --addr localhost:9000`,
		Run: runServer,
	}

	cmd.Flags().StringVarP(&serverAddr, "addr", "a", "localhost:8080",
		"Address to listen on (host:port)")
	cmd.Flags().StringVarP(&securityKey, "security-key", "k", "",
		"Security key for authentication (optional, if not set authentication is disabled)")
	cmd.Flags().StringVarP(&tlsCert, "tls-cert", "c", "",
		"Path to TLS certificate file (enables WSS)")
	cmd.Flags().StringVarP(&tlsKey, "tls-key", "t", "",
		"Path to TLS private key file (enables WSS)")

	return cmd
}

func runServer(cmd *cobra.Command, args []string) {
	_ = args

	log = logging.InitWithComponent("CMD")
	log.Info().Str("version", cmd.Root().Version).Msg("Starting RevoShell Server")

	// Validate TLS configuration
	if (tlsCert != "" && tlsKey == "") || (tlsCert == "" && tlsKey != "") {
		log.Fatal().Msg("Both --tls-cert and --tls-key must be provided together")
	}

	// Create server with security key
	srv := serverPkg.NewWithSecurityKey(securityKey)

	// Start HTTP/WebSocket server (blocks)
	if tlsCert != "" && tlsKey != "" {
		log.Info().Str("mode", "WSS (secure)").Msg("TLS enabled")

		if err := srv.StartTLS(serverAddr, tlsCert, tlsKey); err != nil {
			log.Fatal().Err(err).Msg("Failed to start server with TLS")
		}
	} else {
		log.Warn().Msg("TLS not enabled - using insecure WebSocket (WS). Consider using --tls-cert and --tls-key")

		if err := srv.Start(serverAddr); err != nil {
			log.Fatal().Err(err).Msg("Failed to start server")
		}
	}
}
