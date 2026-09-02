package main

import (
	"fmt"
	"os"

	"github.com/FareinheitsTemp/fire_station/cmd/webapi"
	"github.com/FareinheitsTemp/fire_station/internal/config"
	"github.com/spf13/cobra"
)

func main() {
	var flagDB, flagFont, flagAIKey, flagAIModel string

	apply := func() *config.Config {
		cfg, err := config.Load()
		if err != nil {
			fmt.Fprintln(os.Stderr, "config:", err)
			os.Exit(1)
		}
		changed := false
		if flagDB != "" {
			cfg.DBPath = flagDB
			changed = true
		}
		if flagFont != "" {
			cfg.FontPath = flagFont
			changed = true
		}
		if flagAIKey != "" {
			cfg.AIKey = flagAIKey
			changed = true
		}
		if flagAIModel != "" {
			cfg.AIModel = flagAIModel
			changed = true
		}
		if changed {
			_ = cfg.Save()
		}
		return cfg
	}

	root := &cobra.Command{
		Use:   "fire-station",
		Short: "АІС «Пожежна частина» — консольний API-сервер (міст між вебом і Access)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return webapi.Run(apply())
		},
	}

	root.Flags().StringVar(&flagDB, "db", "", "шлях до файлу БД (.accdb)")
	root.Flags().StringVar(&flagFont, "font", "", "шлях до TTF-шрифту для PDF")
	root.Flags().StringVar(&flagAIKey, "ai-key", "", "API ключ aimlapi")
	root.Flags().StringVar(&flagAIModel, "ai-model", "", "модель ШІ")

	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Помилка:", err)
		os.Exit(1)
	}
}
