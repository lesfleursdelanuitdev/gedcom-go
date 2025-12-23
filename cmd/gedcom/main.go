package main

import (
	"fmt"
	"os"

	"github.com/lesfleursdelanuitdev/gedcom-go/cmd/gedcom/commands"
	"github.com/lesfleursdelanuitdev/gedcom-go/cmd/gedcom/internal"
	"github.com/spf13/cobra"
)

var (
	version    = "1.0.0"
	configPath string
	quiet      bool
	verbose    bool
	noColor    bool
)

var rootCmd = &cobra.Command{
	Use:     "gedcom",
	Short:   "GEDCOM command-line tool",
	Long:    "A comprehensive command-line tool for parsing, validating, querying, and exporting GEDCOM files",
	Version: version,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Load config
		config, err := internal.LoadConfig(configPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to load config: %v\n", err)
			config = internal.DefaultConfig()
		}

		// Apply flags
		if quiet {
			internal.SetQuietMode(true)
			config.Output.Progress = false
		}
		if noColor {
			config.Output.Color = false
		}

		// Initialize color
		internal.InitColor(config.Output.Color)
	},
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVar(&configPath, "config", "", "Config file path")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "Quiet mode (suppress progress bars)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "Disable colored output")

	// Add commands
	rootCmd.AddCommand(commands.GetParseCommand())
	rootCmd.AddCommand(commands.GetValidateCommand())
	rootCmd.AddCommand(commands.GetExportCommand())
	rootCmd.AddCommand(commands.GetInteractiveCommand())
	rootCmd.AddCommand(commands.GetSearchCommand())
	rootCmd.AddCommand(commands.GetDiffCommand())
	rootCmd.AddCommand(commands.GetQualityCommand())
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		internal.PrintError("Error: %v\n", err)
		os.Exit(1)
	}
}
