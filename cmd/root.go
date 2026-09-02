package cmd

import (
	"fmt"
	"os"

	"github.com/FareinheitsTemp/fire_station/internal/config"
	"github.com/spf13/cobra"
)

// cfg — конфігурація користувача (~/.fire-station/config.yaml).
var cfg *config.Config

var rootCmd = &cobra.Command{
	Use:   "fire-station",
	Short: "АІС пожежної частини (курсова з дисципліни «Бази даних»)",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Запуск без аргументів (подвійний клік) — інтерактивне меню.
		return runInteractive()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cfg, _ = config.Load()
}
