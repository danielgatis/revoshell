package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/danielgatis/revoshell/cmd/agent"
	"github.com/danielgatis/revoshell/cmd/connect"
	"github.com/danielgatis/revoshell/cmd/devices"
	"github.com/danielgatis/revoshell/cmd/disconnect"
	"github.com/danielgatis/revoshell/cmd/download"
	"github.com/danielgatis/revoshell/cmd/server"
	"github.com/danielgatis/revoshell/cmd/sessions"
	"github.com/danielgatis/revoshell/cmd/upload"
	"github.com/danielgatis/revoshell/pkg/config"
	"github.com/danielgatis/revoshell/pkg/version"
)

var (
	rootCmd = &cobra.Command{
		Use:     "revoshell",
		Short:   "RevoShell - A secure reverse-shell orchestration hub for managing remote devices.",
		Version: version.GetFullVersion(),
	}

	// GlobalConfig holds configuration loaded from file.
	GlobalConfig *config.Config
)

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Load configuration from file (optional)
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
		// Continue with defaults
		cfg = &config.Config{}
	}

	GlobalConfig = cfg

	// Register subcommands
	rootCmd.AddCommand(agent.NewCommand())
	rootCmd.AddCommand(server.NewCommand())
	rootCmd.AddCommand(connect.NewCommand())
	rootCmd.AddCommand(disconnect.NewCommand())
	rootCmd.AddCommand(devices.NewCommand())
	rootCmd.AddCommand(sessions.NewCommand())
	rootCmd.AddCommand(download.NewCommand())
	rootCmd.AddCommand(upload.NewCommand())
}
