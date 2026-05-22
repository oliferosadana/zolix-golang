package app

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type AutoGoPayClient struct {
	baseURL string
	apiKey  string
	client  *http.Client
}

type AutoGoPayQRIS struct {
	TransactionID string     `json:"transaction_id"`
	OrderID       string     `json:"order_id"`
	Amount        int        `json:"amount"`
	Status        string     `json:"transaction_status"`
	QRString      string     `json:"qr_string"`
	QRURL         string     `json:"qr_url"`
	ExpiryTime    *time.Time `json:"expiry_time,omitempty"`
}

type AutoGoPayStatus struct {
	TransactionID   string `json:"transaction_id"`
	Status          string `json:"transaction_status"`
	TransactionTime string `json:"transaction_time"`
}

type AutoGoPayWebhookPayload struct {
	Event       string `json:"event"`
	Transaction struct {
		ID     string `json:"id"`
		Amount int    `json:"amount"`
		Status string `json:"status"`
		Issuer string `json:"issuer"`
	} `json:"transaction"`
}

func NewAutoGoPayClient() *AutoGoPayClient {
	return &AutoGoPayClient{
		baseURL: strings.TrimRight(envDefault("AUTOGOPAY_BASE_URL", "https://v1-gateway.autogopay.site"), "/"),
		apiKey:  strings.TrimSpace(envDefault("AUTOGOPAY_API_KEY", "")),
		client:  &http.Client{Timeout: 15 * time.Second},
	}
}

func (c *AutoGoPayClient) Enabled() bool {
	return c != nil && strings.TrimSpace(c.apiKey) != ""
}

func (c *AutoGoPayClient) GenerateQRIS(ctx context.Context, amount int) (AutoGoPayQRIS, error) {
	if amount <= 0 {
		return AutoGoPayQRIS{}, errors.New("amount must be greater than zero")
	}
	var response struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
		Data    struct {
			TransactionID string `json:"transaction_id"`
			OrderID       string `json:"order_id"`
			Amount        int    `json:"amount"`
			Status        string `json:"transaction_status"`
			QRString      string `json:"qr_string"`
			QRURL         string `json:"qr_url"`
			ExpiryTime    string `json:"expiry_time"`
		} `json:"data"`
	}
	if err := c.post(ctx, "/qris/generate", map[string]int{"amount": amount}, &response); err != nil {
		return AutoGoPayQRIS{}, err
	}
	if !response.Success {
		return AutoGoPayQRIS{}, errors.New(defaultMessage(response.Message, "failed to generate QRIS"))
	}
	expiry := parseAutoGoPayTime(response.Data.ExpiryTime)
	return AutoGoPayQRIS{
		TransactionID: response.Data.TransactionID,
		OrderID:       response.Data.OrderID,
		Amount:        response.Data.Amount,
		Status:        response.Data.Status,
		QRString:      response.Data.QRString,
		QRURL:         response.Data.QRURL,
		ExpiryTime:    expiry,
	}, nil
}

func (c *AutoGoPayClient) QRISStatus(ctx context.Context, transactionID string) (AutoGoPayStatus, error) {
	transactionID = strings.TrimSpace(transactionID)
	if transactionID == "" {
		return AutoGoPayStatus{}, errors.New("transaction_id is required")
	}
	var response struct {
		Success bool            `json:"success"`
		Message string          `json:"message"`
		Data    AutoGoPayStatus `json:"data"`
	}
	if err := c.post(ctx, "/qris/status", map[string]string{"transaction_id": transactionID}, &response); err != nil {
		return AutoGoPayStatus{}, err
	}
	if response.Data.TransactionID == "" {
		return AutoGoPayStatus{}, errors.New(defaultMessage(response.Message, "invalid status response"))
	}
	return response.Data, nil
}

func (c *AutoGoPayClient) VerifySignature(payload []byte, signature string) bool {
	if !c.Enabled() {
		return false
	}
	mac := hmac.New(sha256.New, []byte(c.apiKey))
	_, _ = mac.Write(payload)
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(strings.TrimSpace(signature)))
}

func (c *AutoGoPayClient) post(ctx context.Context, path string, payload any, target any) error {
	if !c.Enabled() {
		return errors.New("AUTOGOPAY_API_KEY is not configured")
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("AutoGoPay returned HTTP %d", resp.StatusCode)
	}
	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return err
	}
	return nil
}

func parseAutoGoPayTime(value string) *time.Time {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	layouts := []string{"2006-01-02 15:04:05", time.RFC3339}
	for _, layout := range layouts {
		if parsed, err := time.ParseInLocation(layout, value, time.Local); err == nil {
			return &parsed
		}
	}
	return nil
}

func defaultMessage(message, fallback string) string {
	if strings.TrimSpace(message) != "" {
		return message
	}
	return fallback
}
