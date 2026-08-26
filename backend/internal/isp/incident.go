package isp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/H4D3ZS/aegis-homelabserver/backend/internal/db"
)

// IncidentManager handles SLA alert notifications to Discord or Telegram.
type IncidentManager struct {
	Store          *db.Store
	DiscordWebhook string
	TelegramToken  string
	TelegramChatID string
	client         *http.Client
}

// NewIncidentManager initializes the incident dispatcher.
func NewIncidentManager(store *db.Store, discordWebhook, telegramToken, telegramChatID string) *IncidentManager {
	return &IncidentManager{
		Store:          store,
		DiscordWebhook: discordWebhook,
		TelegramToken:  telegramToken,
		TelegramChatID: telegramChatID,
		client:         &http.Client{Timeout: 5 * time.Second},
	}
}

// NotifyDegradation sends alert when SLA is breached or outage occurs.
func (im *IncidentManager) NotifyDegradation(ctx context.Context, title, message string) {
	log.Printf("[INCIDENT] %s: %s", title, message)

	if im.DiscordWebhook != "" {
		go im.sendDiscord(title, message)
	}
	if im.TelegramToken != "" && im.TelegramChatID != "" {
		go im.sendTelegram(title, message)
	}
}

func (im *IncidentManager) sendDiscord(title, message string) {
	payload := map[string]interface{}{
		"embeds": []map[string]interface{}{
			{
				"title":       "🚨 ISP Degradation Alert: " + title,
				"description": message,
				"color":       15158332,
				"timestamp":   time.Now().Format(time.RFC3339),
			},
		},
	}
	b, _ := json.Marshal(payload)
	_, _ = im.client.Post(im.DiscordWebhook, "application/json", bytes.NewBuffer(b))
}

func (im *IncidentManager) sendTelegram(title, message string) {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", im.TelegramToken)
	text := fmt.Sprintf("🚨 *ISP Degradation Alert*\n*%s*\n\n%s", title, message)
	payload := map[string]string{
		"chat_id":    im.TelegramChatID,
		"text":       text,
		"parse_mode": "Markdown",
	}
	b, _ := json.Marshal(payload)
	_, _ = im.client.Post(url, "application/json", bytes.NewBuffer(b))
}
