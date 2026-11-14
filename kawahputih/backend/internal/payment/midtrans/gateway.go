package midtrans

import (
	"bytes"
	"context"
	"crypto/sha512"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Config struct {
	ServerKey    string
	IsProduction bool
	BaseURL      string
	SnapBaseURL  string
}

type Gateway struct {
	cfg    Config
	client *http.Client
}

func NewGateway(cfg Config) *Gateway {
	baseURL := cfg.BaseURL
	if baseURL == "" {
		if cfg.IsProduction {
			baseURL = "https://api.midtrans.com/v2"
		} else {
			baseURL = "https://api.sandbox.midtrans.com/v2"
		}
	}
	snapBaseURL := cfg.SnapBaseURL
	if snapBaseURL == "" {
		if cfg.IsProduction {
			snapBaseURL = "https://app.midtrans.com/snap/v1"
		} else {
			snapBaseURL = "https://app.sandbox.midtrans.com/snap/v1"
		}
	}
	return &Gateway{
		cfg: Config{
			ServerKey:    cfg.ServerKey,
			IsProduction: cfg.IsProduction,
			BaseURL:      baseURL,
			SnapBaseURL:  snapBaseURL,
		},
		client: &http.Client{Timeout: 20 * time.Second},
	}
}

type TransactionRequest struct {
	TransactionDetails struct {
		OrderID     string `json:"order_id"`
		GrossAmount int64  `json:"gross_amount"`
	} `json:"transaction_details"`
	CustomerDetails struct {
		FirstName string `json:"first_name"`
		Email     string `json:"email"`
	} `json:"customer_details"`
}

func (g *Gateway) CreateTransaction(ctx context.Context, orderID string, grossAmount int64, customerName, customerEmail, itemName string) (string, error) {
	if strings.TrimSpace(g.cfg.ServerKey) == "" {
		return "", fmt.Errorf("midtrans server key is not configured")
	}
	payload := TransactionRequest{}
	payload.TransactionDetails.OrderID = orderID
	payload.TransactionDetails.GrossAmount = grossAmount
	payload.CustomerDetails.FirstName = customerName
	payload.CustomerDetails.Email = customerEmail
	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("midtrans marshal: %w", err)
	}
	url := g.cfg.SnapBaseURL + "/transactions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("midtrans request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.SetBasicAuth(g.cfg.ServerKey, "")
	resp, err := g.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("midtrans do: %w", err)
	}
	defer resp.Body.Close()
	var res struct {
		Token         string   `json:"token"`
		RedirectURL   string   `json:"redirect_url"`
		ErrorMessages []string `json:"error_messages"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return "", fmt.Errorf("midtrans decode: %w", err)
	}
	if resp.StatusCode >= 300 || len(res.ErrorMessages) > 0 {
		return "", fmt.Errorf("midtrans Snap transaction failed (%d): %v", resp.StatusCode, res.ErrorMessages)
	}
	if res.RedirectURL == "" {
		return "", fmt.Errorf("midtrans Snap transaction response is missing redirect_url")
	}
	return res.RedirectURL, nil
}

// Configured melaporkan apakah server key Midtrans sudah tersedia.
func (g *Gateway) Configured() bool {
	return strings.TrimSpace(g.cfg.ServerKey) != ""
}

// CreateQRIS membuat transaksi QRIS melalui Midtrans Core API dan
// mengembalikan qr_string (payload QR yang bisa di-render sebagai gambar QR).
func (g *Gateway) CreateQRIS(ctx context.Context, orderID string, grossAmount int64) (string, error) {
	if !g.Configured() {
		return "", fmt.Errorf("midtrans server key is not configured")
	}
	payload := struct {
		TransactionDetails struct {
			OrderID     string `json:"order_id"`
			GrossAmount int64  `json:"gross_amount"`
		} `json:"transaction_details"`
		PaymentType string `json:"payment_type"`
		QRIS        struct {
			Acquirer string `json:"acquirer"`
		} `json:"qris"`
	}{}
	payload.TransactionDetails.OrderID = orderID
	payload.TransactionDetails.GrossAmount = grossAmount
	payload.PaymentType = "qris"
	payload.QRIS.Acquirer = "gopay"

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("midtrans marshal: %w", err)
	}
	requestURL := strings.TrimRight(g.cfg.BaseURL, "/") + "/charge"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, requestURL, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("midtrans request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.SetBasicAuth(g.cfg.ServerKey, "")
	resp, err := g.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("midtrans do: %w", err)
	}
	defer resp.Body.Close()
	var res struct {
		QRString      string   `json:"qr_string"`
		ErrorMessages []string `json:"error_messages"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return "", fmt.Errorf("midtrans decode: %w", err)
	}
	if resp.StatusCode >= 300 || len(res.ErrorMessages) > 0 {
		return "", fmt.Errorf("midtrans QRIS charge failed (%d): %v", resp.StatusCode, res.ErrorMessages)
	}
	if res.QRString == "" {
		return "", fmt.Errorf("midtrans QRIS charge response is missing qr_string")
	}
	return res.QRString, nil
}

func (g *Gateway) VerifySignature(orderID, statusCode, grossAmount, signatureKey string) bool {
	if g.cfg.ServerKey == "" {
		return false
	}
	raw := orderID + statusCode + grossAmount + g.cfg.ServerKey
	sum := sha512.Sum512([]byte(raw))
	expected := hex.EncodeToString(sum[:])
	return len(expected) == len(signatureKey) && subtle.ConstantTimeCompare([]byte(expected), []byte(signatureKey)) == 1
}

type TransactionStatusResponse struct {
	TransactionStatus string   `json:"transaction_status"`
	FraudStatus       string   `json:"fraud_status"`
	ErrorMessages     []string `json:"error_messages"`
}

func (g *Gateway) GetTransactionStatus(ctx context.Context, orderID string) (status, fraudStatus string, err error) {
	if strings.TrimSpace(g.cfg.ServerKey) == "" {
		return "", "", fmt.Errorf("midtrans server key is not configured")
	}
	requestURL := strings.TrimRight(g.cfg.BaseURL, "/") + "/" + url.PathEscape(orderID) + "/status"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return "", "", fmt.Errorf("midtrans request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.SetBasicAuth(g.cfg.ServerKey, "")
	resp, err := g.client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("midtrans do: %w", err)
	}
	defer resp.Body.Close()
	var result TransactionStatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", "", fmt.Errorf("midtrans decode: %w", err)
	}
	if resp.StatusCode >= 300 {
		return "", "", fmt.Errorf("midtrans status failed (%d): %v", resp.StatusCode, result.ErrorMessages)
	}
	return result.TransactionStatus, result.FraudStatus, nil
}

func GenerateSignature(orderID, statusCode, grossAmount, serverKey string) string {
	sum := sha512.Sum512([]byte(orderID + statusCode + grossAmount + serverKey))
	return hex.EncodeToString(sum[:])
}
