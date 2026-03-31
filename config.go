package main

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Duration time.Duration

func (d *Duration) UnmarshalYAML(value *yaml.Node) error {
	parsed, err := time.ParseDuration(value.Value)
	if err != nil {
		return err
	}
	*d = Duration(parsed)
	return nil
}

type Product struct {
	Name      string `yaml:"name"`
	ArticleID string `yaml:"article_id"`
	ProductID string `yaml:"product_id"`
}

type Config struct {
	ClubID            string    `yaml:"club_id"`
	StoreID           int       `yaml:"store_id"`
	CheckInterval     Duration  `yaml:"check_interval"`
	SummaryInterval   Duration  `yaml:"summary_interval"`
	DiscordWebhookURL string    `yaml:"discord_webhook_url"`
	Products          []Product `yaml:"products"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}

	cfg := &Config{
		StoreID: 10201, // default
	}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	if cfg.ClubID == "" {
		return nil, fmt.Errorf("club_id is required")
	}
	if cfg.StoreID <= 0 {
		return nil, fmt.Errorf("store_id must be positive")
	}
	if time.Duration(cfg.CheckInterval) < time.Minute {
		return nil, fmt.Errorf("check_interval must be at least 1m")
	}
	if len(cfg.Products) == 0 {
		return nil, fmt.Errorf("at least one product is required")
	}
	for i, p := range cfg.Products {
		if p.ArticleID == "" {
			return nil, fmt.Errorf("product %d: article_id is required", i)
		}
		if p.ProductID == "" {
			return nil, fmt.Errorf("product %d: product_id is required", i)
		}
		if p.Name == "" {
			cfg.Products[i].Name = p.ArticleID
		}
	}

	return cfg, nil
}
