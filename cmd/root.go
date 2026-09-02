package cmd

import (
	"fmt"
	"os"

	"github.com/FareinheitsTemp/fire_station/cmd/tui"
	"github.com/FareinheitsTemp/fire_station/internal/config"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "fire-station",
	Short: "АІС пожежної частини — повноекранний TUI (курсова з «Баз даних»)",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, _ := config.Load()
		return tui.Run(cfg)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
