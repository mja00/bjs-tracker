package main

import (
	"context"
	"flag"
	"log"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	configPath := flag.String("config", "config.yaml", "path to config file")
	flag.Parse()

	cfg, err := LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.Printf("BJ's Stock Tracker starting")
	log.Printf("  Club: %s", cfg.ClubID)
	log.Printf("  Products: %d", len(cfg.Products))
	log.Printf("  Interval: %s", time.Duration(cfg.CheckInterval))
	if cfg.DiscordWebhookURL != "" {
		log.Printf("  Discord: enabled")
	} else {
		log.Printf("  Discord: disabled (no webhook URL)")
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	tracker := NewTracker(cfg)
	tracker.Run(ctx)

	log.Println("Shutdown complete")
}
