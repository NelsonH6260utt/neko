package main

import (
	"fmt"
	"os"

	"demodesk/neko/internal/config"
	"demodesk/neko/internal/utils"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	// Version is the current version of the application, injected at build time.
	Version = "dev"
	// GitCommit is the git commit hash, injected at build time.
	GitCommit = "unknown"
	// BuildDate is the build date, injected at build time.
	BuildDate = "unknown"
)

var root = &cobra.Command{
	Use:   "neko",
	Short: "neko is a self-hosted virtual browser that runs in Docker",
	Long: `neko is a self-hosted virtual browser that runs in Docker and uses
WebRTC to stream the desktop to connected clients. It supports multiple
concurrent users and a variety of browsers.`,
}

var serve = &cobra.Command{
	Use:   "serve",
	Short: "Start the neko server",
	Long:  `Start the neko HTTP/WebSocket server and begin streaming.`,
	RunE:  runServe,
}

var version = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("neko %s (commit: %s, built: %s)\n", Version, GitCommit, BuildDate)
	},
}

func init() {
	// Use pretty console logging with timestamps for easier local debugging
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "15:04:05"})

	serve.PersistentFlags().String("config", "", "path to configuration file")
	serve.PersistentFlags().String("bind", "127.0.0.1:8080", "address and port to bind the HTTP server")
	serve.PersistentFlags().String("log-level", "debug", "log level (trace, debug, info, warn, error)")

	root.AddCommand(serve)
	root.AddCommand(version)
}

func runServe(cmd *cobra.Command, args []string) error {
	// Parse log level flag
	levelStr, _ := cmd.Flags().GetString("log-level")
	level, err := zerolog.ParseLevel(levelStr)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	log.Info().
		Str("version", Version).
		Str("commit", GitCommit).
		Str("build_date", BuildDate).
		Msg("starting neko server")

	// Load configuration
	cfgPath, _ := cmd.Flags().GetString("config")
	cfg := config.New()
	if err := cfg.Load(cfgPath); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Override bind address from flag if provided
	if bind, _ := cmd.Flags().GetString("bind"); bind != "" {
		cfg.Server.Bind = bind
	}

	log.Info().
		Str("bind", cfg.Server.Bind).
		Msg("server configuration loaded")

	// Ensure required runtime dependencies are available
	if err := utils.CheckDependencies(); err != nil {
		log.Warn().Err(err).Msg("some dependencies are missing")
	}

	// TODO: Initialize and start the application server
	log.Info().Msg("neko server initialized, ready to accept connections")

	return nil
}

func main() {
	// log.Fatal already calls os.Exit(1) internally, so the explicit call below is unreachable
	if err := root.Execute(); err != nil {
		log.Fatal().Err(err).Msg("failed to execute command")
	}
}
