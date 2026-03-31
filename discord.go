package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type discordWebhookPayload struct {
	Username string         `json:"username"`
	Embeds   []discordEmbed `json:"embeds"`
}

type discordEmbed struct {
	Title       string              `json:"title"`
	Description string              `json:"description"`
	Color       int                 `json:"color"`
	Fields      []discordEmbedField `json:"fields,omitempty"`
	Timestamp   string              `json:"timestamp,omitempty"`
}

type discordEmbedField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline"`
}

const (
	colorGreen = 0x00FF00
	colorRed   = 0xFF0000
	colorBlue  = 0x0099FF
)

func SendDiscordSummary(webhookURL string, products []Product, states map[string]*ProductState) error {
	if webhookURL == "" {
		return nil
	}

	var fields []discordEmbedField
	for _, p := range products {
		state, ok := states[p.ArticleID]
		if !ok {
			fields = append(fields, discordEmbedField{
				Name:   p.Name,
				Value:  "Not yet checked",
				Inline: false,
			})
			continue
		}

		indicator := "🔴"
		if state.PreviousStatus == "Available" {
			indicator = "🟢"
		}

		value := fmt.Sprintf("%s %s — qty: %d", indicator, state.PreviousStatus, state.Quantity)
		if state.Price > 0 {
			value += fmt.Sprintf(" — $%.2f", state.Price)
		}

		fields = append(fields, discordEmbedField{
			Name:   p.Name,
			Value:  value,
			Inline: false,
		})
	}

	payload := discordWebhookPayload{
		Username: "BJ's Stock Tracker",
		Embeds: []discordEmbed{
			{
				Title:     "Stock Summary",
				Color:     colorBlue,
				Fields:    fields,
				Timestamp: time.Now().UTC().Format(time.RFC3339),
			},
		},
	}

	jsonBody, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshaling discord summary payload: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, webhookURL, bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("creating discord summary request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("sending discord summary: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests {
		return fmt.Errorf("discord rate limited")
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("discord returned %d", resp.StatusCode)
	}

	return nil
}

func SendDiscordNotification(webhookURL string, product Product, status string, quantity int, price float64) error {
	if webhookURL == "" {
		return nil
	}

	color := colorRed
	if status == "Available" {
		color = colorGreen
	}

	priceStr := "N/A"
	if price > 0 {
		priceStr = fmt.Sprintf("$%.2f", price)
	}

	payload := discordWebhookPayload{
		Username: "BJ's Stock Tracker",
		Embeds: []discordEmbed{
			{
				Title:       fmt.Sprintf("%s — %s", product.Name, status),
				Description: fmt.Sprintf("Stock status changed for **%s**", product.Name),
				Color:       color,
				Fields: []discordEmbedField{
					{Name: "Status", Value: status, Inline: true},
					{Name: "Quantity", Value: fmt.Sprintf("%d", quantity), Inline: true},
					{Name: "Price", Value: priceStr, Inline: true},
					{Name: "Article ID", Value: product.ArticleID, Inline: true},
				},
				Timestamp: time.Now().UTC().Format(time.RFC3339),
			},
		},
	}

	jsonBody, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshaling discord payload: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, webhookURL, bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("creating discord request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("sending discord notification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests {
		return fmt.Errorf("discord rate limited")
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("discord returned %d", resp.StatusCode)
	}

	return nil
}
