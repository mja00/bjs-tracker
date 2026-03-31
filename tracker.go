package main

import (
	"context"
	"log"
	"time"
)

type ProductState struct {
	PreviousStatus string
	Quantity       int
	Price          float64
	LastChecked    time.Time
}

type Tracker struct {
	config *Config
	client *BJSClient
	state  map[string]*ProductState
}

func NewTracker(config *Config) *Tracker {
	return &Tracker{
		config: config,
		client: NewBJSClient(),
		state:  make(map[string]*ProductState),
	}
}

func (t *Tracker) Run(ctx context.Context) {
	defer t.client.Close()

	t.checkAll(ctx)

	ticker := time.NewTicker(time.Duration(t.config.CheckInterval))
	defer ticker.Stop()

	var summaryTicker *time.Ticker
	var summaryCh <-chan time.Time
	if t.config.SummaryInterval > 0 {
		summaryTicker = time.NewTicker(time.Duration(t.config.SummaryInterval))
		defer summaryTicker.Stop()
		summaryCh = summaryTicker.C
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			t.checkAll(ctx)
		case <-summaryCh:
			t.sendSummary()
		}
	}
}

func (t *Tracker) sendSummary() {
	if t.config.DiscordWebhookURL == "" {
		return
	}
	if err := SendDiscordSummary(t.config.DiscordWebhookURL, t.config.Products, t.state); err != nil {
		log.Printf("  [WARN]  summary notification failed: %v", err)
	} else {
		log.Printf("  [NOTIFY] sent summary notification")
	}
}

func (t *Tracker) checkAll(ctx context.Context) {
	log.Printf("Checking %d product(s)...", len(t.config.Products))
	for _, product := range t.config.Products {
		if ctx.Err() != nil {
			return
		}
		t.checkProduct(product)
	}
}

func (t *Tracker) checkProduct(product Product) {
	status, qty, err := t.client.CheckInventory(t.config.StoreID, product.ArticleID, t.config.ClubID)
	if err != nil {
		log.Printf("  [ERROR] %s: inventory check failed: %v", product.Name, err)
		return
	}

	price, err := t.client.GetPrice(t.config.StoreID, product.ProductID, t.config.ClubID)
	if err != nil {
		log.Printf("  [WARN]  %s: price check failed: %v", product.Name, err)
	}

	// Log to console
	if price > 0 {
		log.Printf("  %s: %s (qty: %d) - $%.2f", product.Name, status, qty, price)
	} else {
		log.Printf("  %s: %s (qty: %d)", product.Name, status, qty)
	}

	// Check for status transition
	prev, exists := t.state[product.ArticleID]
	shouldNotify := false

	if !exists {
		// First check — only notify if available
		shouldNotify = status == "Available"
	} else if prev.PreviousStatus != status {
		// Status changed
		shouldNotify = true
	}

	if shouldNotify {
		if err := SendDiscordNotification(t.config.DiscordWebhookURL, product, status, qty, price); err != nil {
			log.Printf("  [WARN]  %s: discord notification failed: %v", product.Name, err)
		} else if t.config.DiscordWebhookURL != "" {
			log.Printf("  [NOTIFY] %s: sent Discord notification (%s)", product.Name, status)
		}
	}

	t.state[product.ArticleID] = &ProductState{
		PreviousStatus: status,
		Quantity:       qty,
		Price:          price,
		LastChecked:    time.Now(),
	}
}
