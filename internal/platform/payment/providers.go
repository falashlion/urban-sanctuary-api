package payment

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// PaymentClient defines the interface for payment providers.
type PaymentClient interface {
	Initiate(ctx context.Context, req PaymentRequest) (*PaymentResult, error)
	CheckStatus(ctx context.Context, providerRef string) (*PaymentResult, error)
	ValidateWebhook(signature string, body []byte) bool
}

// PaymentRequest contains the data needed to initiate a payment.
type PaymentRequest struct {
	Amount      float64
	Currency    string
	PhoneNumber string
	Description string
	ExternalID  string
	CallbackURL string
}

// PaymentResult contains the response from a payment provider.
type PaymentResult struct {
	ProviderReference string
	Status            string // "pending", "completed", "failed"
	Message           string
}

// --- MTN MoMo Client ---

// MTNMoMoClient implements PaymentClient for MTN Mobile Money.
type MTNMoMoClient struct {
	apiKey      string
	apiSecret   string
	baseURL     string
	callbackURL string
	log         zerolog.Logger
}

// NewMTNMoMoClient creates a new MTN MoMo client.
func NewMTNMoMoClient(apiKey, apiSecret, baseURL, callbackURL string, log zerolog.Logger) *MTNMoMoClient {
	return &MTNMoMoClient{
		apiKey:      apiKey,
		apiSecret:   apiSecret,
		baseURL:     baseURL,
		callbackURL: callbackURL,
		log:         log,
	}
}

// Initiate starts a payment request with MTN MoMo.
func (c *MTNMoMoClient) Initiate(ctx context.Context, req PaymentRequest) (*PaymentResult, error) {
	c.log.Info().
		Float64("amount", req.Amount).
		Str("phone", req.PhoneNumber).
		Msg("Initiating MTN MoMo payment")

	// In sandbox mode, simulate a successful payment initiation
	providerRef := uuid.New().String()

	// TODO: Implement actual MTN MoMo API call
	// POST {baseURL}/collection/v1_0/requesttopay
	// Headers: X-Reference-Id, X-Target-Environment, Ocp-Apim-Subscription-Key
	// Body: amount, currency, externalId, payer.partyId, payerMessage

	return &PaymentResult{
		ProviderReference: providerRef,
		Status:            "pending",
		Message:           "Payment request submitted to MTN MoMo",
	}, nil
}

// CheckStatus checks the status of an MTN MoMo payment.
func (c *MTNMoMoClient) CheckStatus(ctx context.Context, providerRef string) (*PaymentResult, error) {
	c.log.Info().Str("provider_ref", providerRef).Msg("Checking MTN MoMo payment status")

	// TODO: Implement actual status check
	// GET {baseURL}/collection/v1_0/requesttopay/{referenceId}

	return &PaymentResult{
		ProviderReference: providerRef,
		Status:            "pending",
		Message:           "Payment is being processed",
	}, nil
}

// ValidateWebhook validates an MTN MoMo webhook signature.
func (c *MTNMoMoClient) ValidateWebhook(signature string, body []byte) bool {
	// TODO: Implement HMAC signature validation
	c.log.Info().Msg("Validating MTN MoMo webhook signature")
	return true
}

// --- Orange Money Client ---

// OrangeMoneyClient implements PaymentClient for Orange Money.
type OrangeMoneyClient struct {
	apiKey      string
	apiSecret   string
	baseURL     string
	callbackURL string
	log         zerolog.Logger
}

// NewOrangeMoneyClient creates a new Orange Money client.
func NewOrangeMoneyClient(apiKey, apiSecret, baseURL, callbackURL string, log zerolog.Logger) *OrangeMoneyClient {
	return &OrangeMoneyClient{
		apiKey:      apiKey,
		apiSecret:   apiSecret,
		baseURL:     baseURL,
		callbackURL: callbackURL,
		log:         log,
	}
}

// Initiate starts a payment request with Orange Money.
func (c *OrangeMoneyClient) Initiate(ctx context.Context, req PaymentRequest) (*PaymentResult, error) {
	c.log.Info().
		Float64("amount", req.Amount).
		Str("phone", req.PhoneNumber).
		Msg("Initiating Orange Money payment")

	providerRef := uuid.New().String()

	// TODO: Implement actual Orange Money API call
	// POST {baseURL}/orange-money-webpay/dev/v1/webpayment
	// Headers: Authorization: Bearer {token}
	// Body: merchant_key, currency, order_id, amount, return_url, cancel_url, notif_url, lang

	return &PaymentResult{
		ProviderReference: providerRef,
		Status:            "pending",
		Message:           "Payment request submitted to Orange Money",
	}, nil
}

// CheckStatus checks the status of an Orange Money payment.
func (c *OrangeMoneyClient) CheckStatus(ctx context.Context, providerRef string) (*PaymentResult, error) {
	c.log.Info().Str("provider_ref", providerRef).Msg("Checking Orange Money payment status")

	// TODO: Implement actual status check
	return &PaymentResult{
		ProviderReference: providerRef,
		Status:            "pending",
		Message:           "Payment is being processed",
	}, nil
}

// ValidateWebhook validates an Orange Money webhook signature.
func (c *OrangeMoneyClient) ValidateWebhook(signature string, body []byte) bool {
	// TODO: Implement signature validation
	c.log.Info().Msg("Validating Orange Money webhook signature")
	return true
}

// GetClient returns the appropriate payment client for the given provider.
func GetClient(provider string, mtn *MTNMoMoClient, orange *OrangeMoneyClient) (PaymentClient, error) {
	switch provider {
	case "mtn_momo":
		return mtn, nil
	case "orange_money":
		return orange, nil
	default:
		return nil, fmt.Errorf("unsupported payment provider: %s", provider)
	}
}
