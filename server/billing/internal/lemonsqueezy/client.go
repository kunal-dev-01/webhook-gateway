package lemonsqueezy

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"billing/internal/config"

	"shared/logger"
)

// ============================================================================
// CLIENT
// ============================================================================

type Client struct {
	apiKey        string
	storeID       string
	webhookSecret string
	baseURL       string
	httpClient    *http.Client
}

func NewClient(cfg *config.LemonSqueezyConfig) *Client {
	return &Client{
		apiKey:        cfg.APIKey,
		storeID:       cfg.StoreID,
		webhookSecret: cfg.WebhookSecret,
		baseURL:       cfg.BaseURL,
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
	}
}

// ============================================================================
// SIGNATURE VERIFICATION
// ============================================================================

// VerifyWebhookSignature verifies the X-Signature header from LemonSqueezy
// HMAC-SHA256(rawBody, webhookSecret)
func (c *Client) VerifyWebhookSignature(rawBody []byte, signature string) bool {
	mac := hmac.New(sha256.New, []byte(c.webhookSecret))
	mac.Write(rawBody)
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(signature))
}

// ============================================================================
// CHECKOUT
// ============================================================================

// CreateCheckout creates a LemonSqueezy checkout session and returns the URL
func (c *Client) CreateCheckout(ctx context.Context, req *CreateCheckoutRequest) (*CheckoutData, error) {
	resp, err := c.do(ctx, http.MethodPost, "/checkouts", req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result APIResponse[CheckoutData]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode checkout response: %w", err)
	}

	return &result.Data, nil
}

// ============================================================================
// SUBSCRIPTION
// ============================================================================

// GetSubscription fetches a subscription by LemonSqueezy subscription ID
func (c *Client) GetSubscription(ctx context.Context, subscriptionID string) (*SubscriptionData, error) {
	resp, err := c.do(ctx, http.MethodGet, "/subscriptions/"+subscriptionID, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result APIResponse[SubscriptionData]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode subscription response: %w", err)
	}

	return &result.Data, nil
}

// CancelSubscription marks subscription to cancel at period end
func (c *Client) CancelSubscription(ctx context.Context, subscriptionID string) (*SubscriptionData, error) {
	resp, err := c.do(ctx, http.MethodDelete, "/subscriptions/"+subscriptionID, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result APIResponse[SubscriptionData]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode cancel response: %w", err)
	}

	return &result.Data, nil
}

// ResumeSubscription resumes a cancelled/paused subscription
func (c *Client) ResumeSubscription(ctx context.Context, subscriptionID string) (*SubscriptionData, error) {
	req := &UpdateSubscriptionRequest{}
	req.Data.ID = subscriptionID
	req.Data.Type = "subscriptions"
	// Pass null pause to resume
	req.Data.Attributes.Pause = nil

	resp, err := c.do(ctx, http.MethodPatch, "/subscriptions/"+subscriptionID, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result APIResponse[SubscriptionData]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode resume response: %w", err)
	}

	return &result.Data, nil
}

// ChangeSubscriptionPlan upgrades or downgrades a subscription to a new variant
func (c *Client) ChangeSubscriptionPlan(ctx context.Context, subscriptionID string, variantID int) (*SubscriptionData, error) {
	req := &UpdateSubscriptionRequest{}
	req.Data.ID = subscriptionID
	req.Data.Type = "subscriptions"
	req.Data.Attributes.VariantID = &variantID

	resp, err := c.do(ctx, http.MethodPatch, "/subscriptions/"+subscriptionID, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result APIResponse[SubscriptionData]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode plan change response: %w", err)
	}

	return &result.Data, nil
}

// ============================================================================
// INTERNAL HTTP
// ============================================================================

func (c *Client) do(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	var bodyReader io.Reader

	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	url := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Accept", "application/vnd.api+json")
	req.Header.Set("Content-Type", "application/vnd.api+json")

	start := time.Now()
	resp, err := c.httpClient.Do(req)
	duration := time.Since(start)

	if err != nil {
		logger.Error("LemonSqueezy API request failed", err, "method", method, "path", path)
		return nil, fmt.Errorf("LemonSqueezy API request failed: %w", err)
	}

	logger.Debug("LemonSqueezy API call", "method", method, "path", path, "status", resp.StatusCode, "duration", duration)

	// Handle API errors
	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		bodyBytes, _ := io.ReadAll(resp.Body)

		var apiErr APIError
		if err := json.Unmarshal(bodyBytes, &apiErr); err == nil && len(apiErr.Errors) > 0 {
			logger.Error("LemonSqueezy API error", fmt.Errorf("%s", apiErr.Error()), "status", resp.StatusCode, "path", path)
			return nil, &apiErr
		}
		return nil, fmt.Errorf("LemonSqueezy API error: status %d", resp.StatusCode)
	}

	return resp, nil
}

// ============================================================================
// HELPERS
// ============================================================================

// GetStoreIDInt returns store ID as int
func (c *Client) GetStoreIDInt() int {
	id, _ := strconv.Atoi(c.storeID)
	return id
}