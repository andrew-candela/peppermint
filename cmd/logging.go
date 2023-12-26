package cmd

import (
	"log/slog"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// this variable is set by an argument to the cobra command
var verbose bool

// Replaces the default logger with my custom logger
// using the level passed in from the root command flag 'verbose'.
func configureLogger(cmd *cobra.Command, args []string) {
	var level slog.Level
	if verbose {
		level = slog.LevelDebug
		viper.Set("verbose", true)
	} else {
		level = slog.LevelInfo
	}
	log_options := &slog.HandlerOptions{
		Level: level,
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, log_options))
	slog.SetDefault(logger)
	slog.Debug("Debug output is enabled!")
}
