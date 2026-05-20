package app

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type WhatsAppClient struct {
	baseURL string
	apiKey  string
	session string
	client  *http.Client
}

type WhatsAppSendRequest struct {
	OrderID int    `json:"order_id"`
	Message string `json:"message"`
}

func NewWhatsAppClient() *WhatsAppClient {
	return &WhatsAppClient{
		baseURL: strings.TrimRight(envDefault("WAHA_BASE_URL", ""), "/"),
		apiKey:  envDefault("WAHA_API_KEY", ""),
		session: envDefault("WAHA_SESSION", "default"),
		client:  &http.Client{Timeout: 15 * time.Second},
	}
}

func (c *WhatsAppClient) Enabled() bool {
	return c.baseURL != "" && c.apiKey != ""
}

func (c *WhatsAppClient) SendText(phone, text string) error {
	if !c.Enabled() {
		return errors.New("WAHA is not configured")
	}
	chatID := whatsappChatID(phone)
	if chatID == "" {
		return errors.New("invalid WhatsApp phone number")
	}
	payload := map[string]any{
		"chatId":      chatID,
		"text":        text,
		"session":     c.session,
		"linkPreview": true,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/api/sendText", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Api-Key", c.apiKey)
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("WAHA sendText failed with status %d", resp.StatusCode)
	}
	return nil
}

func whatsappChatID(phone string) string {
	digits := onlyDigits(phone)
	if digits == "" {
		return ""
	}
	if strings.HasPrefix(digits, "0") {
		digits = "62" + strings.TrimPrefix(digits, "0")
	}
	return digits + "@c.us"
}

func onlyDigits(value string) string {
	var builder strings.Builder
	for _, char := range value {
		if char >= '0' && char <= '9' {
			builder.WriteRune(char)
		}
	}
	return builder.String()
}

func orderWhatsAppMessage(order Order, publicBaseURL string) string {
	trackingURL := strings.TrimRight(publicBaseURL, "/") + "/order/" + order.InvoiceNumber
	return fmt.Sprintf(
		"Halo Kak %s\n\nOrder sepatu Anda dengan nomor nota %s saat ini berstatus: %s.\n\nTotal: Rp %s\nTracking order:\n%s\n\nTerima kasih telah menggunakan ZOLIX Shoe Care.",
		order.CustomerName,
		order.InvoiceNumber,
		order.Status,
		formatIDR(order.TotalPrice),
		trackingURL,
	)
}

func formatIDR(value int) string {
	raw := fmt.Sprintf("%d", value)
	var parts []string
	for len(raw) > 3 {
		parts = append([]string{raw[len(raw)-3:]}, parts...)
		raw = raw[:len(raw)-3]
	}
	parts = append([]string{raw}, parts...)
	return strings.Join(parts, ".")
}
