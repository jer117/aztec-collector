// Package alerting provides webhook-based alerting for node health issues
package alerting

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

// AlertType represents the type of alert
type AlertType string

const (
	AlertNodeNotReady      AlertType = "node_not_ready"
	AlertNodeNotSynced     AlertType = "node_not_synced"
	AlertBlocksBehind      AlertType = "blocks_behind"
	AlertBlocksAhead       AlertType = "blocks_ahead"
	AlertLowPeers          AlertType = "low_peers"
	AlertProofLag          AlertType = "proof_lag"
	AlertConnectionError   AlertType = "connection_error"
	AlertMissedAttestation AlertType = "missed_attestation"
	AlertMissedProposal    AlertType = "missed_proposal"
	AlertRecovered         AlertType = "recovered"
)

// AlertLevel represents the severity of an alert
type AlertLevel string

const (
	AlertLevelInfo    AlertLevel = "info"
	AlertLevelWarning AlertLevel = "warning"
	AlertLevelError   AlertLevel = "error"
	AlertLevelCritical AlertLevel = "critical"
)

// Alert represents an alert to be sent
type Alert struct {
	Type             AlertType  `json:"type"`
	Level            AlertLevel `json:"level"`
	Title            string     `json:"title"`
	Message          string     `json:"message"`
	NodeName         string     `json:"node_name,omitempty"`
	BlockHeight      int64      `json:"block_height,omitempty"`
	BlocksBehind     int64      `json:"blocks_behind,omitempty"`
	ValidatorAddress string     `json:"validator_address,omitempty"`
	Slot             string     `json:"slot,omitempty"`
	MissedStreak     int64      `json:"missed_streak,omitempty"`
	Timestamp        time.Time  `json:"timestamp"`
}

// AlertConfig represents the alerting configuration
type AlertConfig struct {
	Enabled  bool          `yaml:"enabled" mapstructure:"enabled"`
	Cooldown time.Duration `yaml:"cooldown" mapstructure:"cooldown"`
	NodeName string        `yaml:"node_name" mapstructure:"node_name"`

	// Webhook configurations
	Slack    *SlackConfig    `yaml:"slack,omitempty" mapstructure:"slack"`
	Discord  *DiscordConfig  `yaml:"discord,omitempty" mapstructure:"discord"`
	Telegram *TelegramConfig `yaml:"telegram,omitempty" mapstructure:"telegram"`
	Generic  *GenericConfig  `yaml:"generic,omitempty" mapstructure:"generic"`
}

// SlackConfig represents Slack webhook configuration
type SlackConfig struct {
	Enabled    bool   `yaml:"enabled" mapstructure:"enabled"`
	WebhookURL string `yaml:"webhook_url" mapstructure:"webhook_url"`
	Channel    string `yaml:"channel,omitempty" mapstructure:"channel"`
	Username   string `yaml:"username,omitempty" mapstructure:"username"`
}

// DiscordConfig represents Discord webhook configuration
type DiscordConfig struct {
	Enabled    bool   `yaml:"enabled" mapstructure:"enabled"`
	WebhookURL string `yaml:"webhook_url" mapstructure:"webhook_url"`
	Username   string `yaml:"username,omitempty" mapstructure:"username"`
	AvatarURL  string `yaml:"avatar_url,omitempty" mapstructure:"avatar_url"`
}

// TelegramConfig represents Telegram bot configuration
type TelegramConfig struct {
	Enabled  bool   `yaml:"enabled" mapstructure:"enabled"`
	BotToken string `yaml:"bot_token" mapstructure:"bot_token"`
	ChatID   string `yaml:"chat_id" mapstructure:"chat_id"`
}

// GenericConfig represents a generic webhook configuration
type GenericConfig struct {
	Enabled    bool              `yaml:"enabled" mapstructure:"enabled"`
	WebhookURL string            `yaml:"webhook_url" mapstructure:"webhook_url"`
	Headers    map[string]string `yaml:"headers,omitempty" mapstructure:"headers"`
	Method     string            `yaml:"method,omitempty" mapstructure:"method"`
}

// Alerter manages sending alerts to configured webhooks
type Alerter struct {
	config     *AlertConfig
	logger     *log.Logger
	httpClient *http.Client

	// Rate limiting / cooldown tracking
	mu            sync.Mutex
	lastAlertTime map[AlertType]time.Time
	activeAlerts  map[AlertType]bool
}

// NewAlerter creates a new Alerter instance
func NewAlerter(config *AlertConfig, logger *log.Logger) *Alerter {
	if logger == nil {
		logger = log.Default()
	}

	return &Alerter{
		config: config,
		logger: logger,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		lastAlertTime: make(map[AlertType]time.Time),
		activeAlerts:  make(map[AlertType]bool),
	}
}

// SendAlert sends an alert to all configured webhooks
func (a *Alerter) SendAlert(ctx context.Context, alert Alert) error {
	if !a.config.Enabled {
		return nil
	}

	// Apply cooldown
	if !a.shouldSendAlert(alert.Type) {
		a.logger.Printf("Alert %s suppressed due to cooldown", alert.Type)
		return nil
	}

	// Set defaults
	if alert.Timestamp.IsZero() {
		alert.Timestamp = time.Now()
	}
	if alert.NodeName == "" {
		alert.NodeName = a.config.NodeName
	}

	var errors []error

	// Send to all enabled channels
	if a.config.Slack != nil && a.config.Slack.Enabled {
		if err := a.sendSlack(ctx, alert); err != nil {
			a.logger.Printf("Failed to send Slack alert: %v", err)
			errors = append(errors, err)
		}
	}

	if a.config.Discord != nil && a.config.Discord.Enabled {
		if err := a.sendDiscord(ctx, alert); err != nil {
			a.logger.Printf("Failed to send Discord alert: %v", err)
			errors = append(errors, err)
		}
	}

	if a.config.Telegram != nil && a.config.Telegram.Enabled {
		if err := a.sendTelegram(ctx, alert); err != nil {
			a.logger.Printf("Failed to send Telegram alert: %v", err)
			errors = append(errors, err)
		}
	}

	if a.config.Generic != nil && a.config.Generic.Enabled {
		if err := a.sendGeneric(ctx, alert); err != nil {
			a.logger.Printf("Failed to send generic webhook alert: %v", err)
			errors = append(errors, err)
		}
	}

	// Update last alert time
	a.markAlertSent(alert.Type)

	if len(errors) > 0 {
		return fmt.Errorf("failed to send %d alert(s)", len(errors))
	}

	return nil
}

// shouldSendAlert checks if we should send an alert based on cooldown
func (a *Alerter) shouldSendAlert(alertType AlertType) bool {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.config.Cooldown == 0 {
		return true
	}

	lastTime, exists := a.lastAlertTime[alertType]
	if !exists {
		return true
	}

	return time.Since(lastTime) >= a.config.Cooldown
}

// markAlertSent records that an alert was sent
func (a *Alerter) markAlertSent(alertType AlertType) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.lastAlertTime[alertType] = time.Now()
	a.activeAlerts[alertType] = true
}

// ClearAlert marks an alert as resolved
func (a *Alerter) ClearAlert(alertType AlertType) {
	a.mu.Lock()
	defer a.mu.Unlock()

	delete(a.activeAlerts, alertType)
}

// IsAlertActive checks if an alert is currently active
func (a *Alerter) IsAlertActive(alertType AlertType) bool {
	a.mu.Lock()
	defer a.mu.Unlock()

	return a.activeAlerts[alertType]
}

// sendSlack sends an alert to Slack
func (a *Alerter) sendSlack(ctx context.Context, alert Alert) error {
	color := a.getSlackColor(alert.Level)

	payload := map[string]interface{}{
		"attachments": []map[string]interface{}{
			{
				"color":  color,
				"title":  alert.Title,
				"text":   alert.Message,
				"footer": fmt.Sprintf("Aztec Collector | %s", alert.NodeName),
				"ts":     alert.Timestamp.Unix(),
				"fields": []map[string]interface{}{
					{
						"title": "Alert Type",
						"value": string(alert.Type),
						"short": true,
					},
					{
						"title": "Severity",
						"value": string(alert.Level),
						"short": true,
					},
				},
			},
		},
	}

	if a.config.Slack.Channel != "" {
		payload["channel"] = a.config.Slack.Channel
	}
	if a.config.Slack.Username != "" {
		payload["username"] = a.config.Slack.Username
	}

	return a.sendWebhook(ctx, a.config.Slack.WebhookURL, payload, nil)
}

// sendDiscord sends an alert to Discord
func (a *Alerter) sendDiscord(ctx context.Context, alert Alert) error {
	color := a.getDiscordColor(alert.Level)

	payload := map[string]interface{}{
		"embeds": []map[string]interface{}{
			{
				"title":       alert.Title,
				"description": alert.Message,
				"color":       color,
				"timestamp":   alert.Timestamp.Format(time.RFC3339),
				"footer": map[string]string{
					"text": fmt.Sprintf("Aztec Collector | %s", alert.NodeName),
				},
				"fields": []map[string]interface{}{
					{
						"name":   "Alert Type",
						"value":  string(alert.Type),
						"inline": true,
					},
					{
						"name":   "Severity",
						"value":  string(alert.Level),
						"inline": true,
					},
				},
			},
		},
	}

	if a.config.Discord.Username != "" {
		payload["username"] = a.config.Discord.Username
	}
	if a.config.Discord.AvatarURL != "" {
		payload["avatar_url"] = a.config.Discord.AvatarURL
	}

	return a.sendWebhook(ctx, a.config.Discord.WebhookURL, payload, nil)
}

// sendTelegram sends an alert to Telegram
func (a *Alerter) sendTelegram(ctx context.Context, alert Alert) error {
	emoji := a.getTelegramEmoji(alert.Level)

	text := fmt.Sprintf(
		"%s *%s*\n\n%s\n\n"+
			"*Type:* `%s`\n"+
			"*Node:* `%s`\n"+
			"*Time:* %s",
		emoji,
		escapeMarkdown(alert.Title),
		escapeMarkdown(alert.Message),
		alert.Type,
		alert.NodeName,
		alert.Timestamp.Format(time.RFC3339),
	)

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", a.config.Telegram.BotToken)
	payload := map[string]interface{}{
		"chat_id":    a.config.Telegram.ChatID,
		"text":       text,
		"parse_mode": "Markdown",
	}

	return a.sendWebhook(ctx, url, payload, nil)
}

// sendGeneric sends an alert to a generic webhook
func (a *Alerter) sendGeneric(ctx context.Context, alert Alert) error {
	method := a.config.Generic.Method
	if method == "" {
		method = "POST"
	}

	return a.sendWebhook(ctx, a.config.Generic.WebhookURL, alert, a.config.Generic.Headers)
}

// sendWebhook sends a JSON payload to a webhook URL
func (a *Alerter) sendWebhook(ctx context.Context, url string, payload interface{}, headers map[string]string) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}

	return nil
}

// getSlackColor returns a Slack attachment color based on alert level
func (a *Alerter) getSlackColor(level AlertLevel) string {
	switch level {
	case AlertLevelCritical:
		return "#dc3545" // Red
	case AlertLevelError:
		return "#fd7e14" // Orange
	case AlertLevelWarning:
		return "#ffc107" // Yellow
	case AlertLevelInfo:
		return "#17a2b8" // Blue
	default:
		return "#6c757d" // Gray
	}
}

// getDiscordColor returns a Discord embed color based on alert level
func (a *Alerter) getDiscordColor(level AlertLevel) int {
	switch level {
	case AlertLevelCritical:
		return 0xdc3545 // Red
	case AlertLevelError:
		return 0xfd7e14 // Orange
	case AlertLevelWarning:
		return 0xffc107 // Yellow
	case AlertLevelInfo:
		return 0x17a2b8 // Blue
	default:
		return 0x6c757d // Gray
	}
}

// getTelegramEmoji returns an emoji based on alert level
func (a *Alerter) getTelegramEmoji(level AlertLevel) string {
	switch level {
	case AlertLevelCritical:
		return "🚨"
	case AlertLevelError:
		return "❌"
	case AlertLevelWarning:
		return "⚠️"
	case AlertLevelInfo:
		return "ℹ️"
	default:
		return "📢"
	}
}

// escapeMarkdown escapes Telegram markdown special characters
func escapeMarkdown(s string) string {
	replacer := map[string]string{
		"_": "\\_",
		"*": "\\*",
		"[": "\\[",
		"]": "\\]",
		"`": "\\`",
	}
	for old, new := range replacer {
		s = replaceAll(s, old, new)
	}
	return s
}

func replaceAll(s, old, new string) string {
	result := ""
	for i := 0; i < len(s); i++ {
		if string(s[i]) == old {
			result += new
		} else {
			result += string(s[i])
		}
	}
	return result
}

// DefaultAlertConfig returns a default alert configuration
func DefaultAlertConfig() *AlertConfig {
	return &AlertConfig{
		Enabled:  false,
		Cooldown: 5 * time.Minute,
		NodeName: "aztec-node",
	}
}

