package lemonsqueezy

import "time"

// ============================================================================
// API RESPONSE WRAPPER
// ============================================================================

// APIResponse wraps a single LemonSqueezy resource
type APIResponse[T any] struct {
	Data T `json:"data"`
}

// APIListResponse wraps a list of LemonSqueezy resources
type APIListResponse[T any] struct {
	Data []T `json:"data"`
}

// APIError represents LemonSqueezy error response
type APIError struct {
	Errors []struct {
		Status string `json:"status"`
		Title  string `json:"title"`
		Detail string `json:"detail"`
	} `json:"errors"`
}

func (e *APIError) Error() string {
	if len(e.Errors) > 0 {
		return e.Errors[0].Detail
	}
	return "unknown LemonSqueezy error"
}

// ============================================================================
// CHECKOUT
// ============================================================================

// CheckoutAttributes represents checkout data
type CheckoutAttributes struct {
	URL        string    `json:"url"`
	ExpiresAt  time.Time `json:"expires_at"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// CheckoutData represents a checkout object
type CheckoutData struct {
	ID         string             `json:"id"`
	Type       string             `json:"type"`
	Attributes CheckoutAttributes `json:"attributes"`
}

// CreateCheckoutRequest — payload to create a checkout session
type CreateCheckoutRequest struct {
	Data struct {
		Type       string `json:"type"`
		Attributes struct {
			// Store + Variant
			StoreID   int `json:"store_id"`
			VariantID int `json:"variant_id"`

			// Pre-fill customer info
			CheckoutData struct {
				Email  string            `json:"email,omitempty"`
				Name   string            `json:"name,omitempty"`
				Custom map[string]string `json:"custom,omitempty"` // org_id, user_id
			} `json:"checkout_data"`

			// Redirect URLs
			CheckoutOptions struct {
				SubscriptionPreview bool `json:"subscription_preview"`
				ButtonColor         string `json:"button_color,omitempty"`
			} `json:"checkout_options"`

			// URLs
			ProductOptions struct {
				RedirectURL    string `json:"redirect_url,omitempty"`
				ReceiptLinkURL string `json:"receipt_link_url,omitempty"`
			} `json:"product_options"`

			// Expiry
			ExpiresAt *time.Time `json:"expires_at,omitempty"`
		} `json:"attributes"`
	} `json:"data"`
}

// ============================================================================
// SUBSCRIPTION
// ============================================================================

// SubscriptionAttributes represents LemonSqueezy subscription data
type SubscriptionAttributes struct {
	StoreID      int    `json:"store_id"`
	OrderID      int    `json:"order_id"`
	OrderItemID  int    `json:"order_item_id"`
	ProductID    int    `json:"product_id"`
	VariantID    int    `json:"variant_id"`
	ProductName  string `json:"product_name"`
	VariantName  string `json:"variant_name"`
	CustomerID   int    `json:"customer_id"`
	UserName     string `json:"user_name"`
	UserEmail    string `json:"user_email"`

	// Status
	Status            string `json:"status"` // "active", "cancelled", "expired", "past_due", "paused", "unpaid", "on_trial"
	StatusFormatted   string `json:"status_formatted"`
	Pause             *struct {
		Mode      string     `json:"mode"`
		ResumesAt *time.Time `json:"resumes_at"`
	} `json:"pause"`

	// Cancellation
	Cancelled         bool       `json:"cancelled"`
	TrialEndsAt       *time.Time `json:"trial_ends_at"`
	BillingAnchor     int        `json:"billing_anchor"`

	// URLs
	Urls struct {
		UpdatePaymentMethod string `json:"update_payment_method"`
		CustomerPortal      string `json:"customer_portal"`
	} `json:"urls"`

	// Renewal
	RenewsAt          *time.Time `json:"renews_at"`
	EndsAt            *time.Time `json:"ends_at"`

	// Timestamps
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// SubscriptionData represents a LemonSqueezy subscription object
type SubscriptionData struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"`
	Attributes SubscriptionAttributes `json:"attributes"`
}

// UpdateSubscriptionRequest — payload to update a subscription
type UpdateSubscriptionRequest struct {
	Data struct {
		ID         string `json:"id"`
		Type       string `json:"type"`
		Attributes struct {
			// For cancellation
			Cancelled bool `json:"cancelled,omitempty"`
			// For pause/resume
			Pause *struct {
				Mode string `json:"mode"` // "void" or "free"
			} `json:"pause,omitempty"`
			// For plan change
			VariantID *int `json:"variant_id,omitempty"`
		} `json:"attributes"`
	} `json:"data"`
}

// ============================================================================
// ORDER (for payment events)
// ============================================================================

// OrderAttributes represents LemonSqueezy order data
type OrderAttributes struct {
	StoreID         int     `json:"store_id"`
	CustomerID      int     `json:"customer_id"`
	Identifier      string  `json:"identifier"`
	OrderNumber     int     `json:"order_number"`
	UserName        string  `json:"user_name"`
	UserEmail       string  `json:"user_email"`
	Currency        string  `json:"currency"`
	CurrencyRate    string  `json:"currency_rate"`
	Subtotal        int     `json:"subtotal"`        // In cents
	TotalFormatted  string  `json:"total_formatted"`
	Total           int     `json:"total"`           // In cents
	Status          string  `json:"status"`          // "pending", "failed", "paid", "refunded"
	StatusFormatted string  `json:"status_formatted"`
	RefundedAt      *string `json:"refunded_at"`
	CreatedAt       string  `json:"created_at"`
	UpdatedAt       string  `json:"updated_at"`
}

// OrderData represents a LemonSqueezy order object
type OrderData struct {
	ID         string          `json:"id"`
	Type       string          `json:"type"`
	Attributes OrderAttributes `json:"attributes"`
}

// ============================================================================
// WEBHOOK EVENT
// ============================================================================

// WebhookPayload is the top-level structure of a LemonSqueezy webhook
type WebhookPayload struct {
	Meta WebhookMeta            `json:"meta"`
	Data map[string]interface{} `json:"data"` // Raw — parsed per event type
}

// WebhookMeta contains event metadata
type WebhookMeta struct {
	EventName  string            `json:"event_name"`
	CustomData map[string]string `json:"custom_data,omitempty"` // org_id, user_id we passed at checkout
	WebhookID  string            `json:"webhook_id"`
}

// SubscriptionWebhookPayload is typed payload for subscription events
type SubscriptionWebhookPayload struct {
	Meta WebhookMeta      `json:"meta"`
	Data SubscriptionData `json:"data"`
}

// OrderWebhookPayload is typed payload for order events
type OrderWebhookPayload struct {
	Meta WebhookMeta `json:"meta"`
	Data OrderData   `json:"data"`
}