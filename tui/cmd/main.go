package main

import (
	"log"
	"os"

	"github.com/shubh-man007/Chirpy/tui/internal/api"
	"github.com/shubh-man007/Chirpy/tui/internal/config"
	"github.com/shubh-man007/Chirpy/tui/internal/ui"
)

func main() {
	cfg := config.New()

	client := api.NewChirpy(cfg.HTTPClient)
	client.BaseURL = cfg.BaseURL

	p := ui.NewProgram(client)

	if _, err := p.Run(); err != nil {
		log.Printf("error running Chirpy TUI: %v", err)
		os.Exit(1)
	}
}

