// Package turnstile verifies Cloudflare Turnstile tokens server-side.
// Tokens must never be trusted from the client alone; always call Siteverify.
package turnstile

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go-cp-boilerplate/internal/pkg/apperror"
)

const defaultSiteverifyURL = "https://challenges.cloudflare.com/turnstile/v0/siteverify"

// Overridable in tests.
var siteverifyURL = defaultSiteverifyURL

// Verifier checks a Turnstile response token against Cloudflare.
type Verifier interface {
	// Enabled reports whether verification is configured (secret present).
	Enabled() bool
	// Verify validates token with Cloudflare. remoteIP is optional but recommended.
	Verify(ctx context.Context, token, remoteIP string) error
}

type Client struct {
	secret     string
	httpClient *http.Client
}

// New returns a Verifier. An empty secret yields a no-op verifier (dev only).
func New(secret string) Verifier {
	secret = strings.TrimSpace(secret)
	if secret == "" {
		return noop{}
	}
	return &Client{
		secret: secret,
		httpClient: &http.Client{
			Timeout: 8 * time.Second,
		},
	}
}

func (c *Client) Enabled() bool { return true }

type siteverifyResponse struct {
	Success     bool     `json:"success"`
	ErrorCodes  []string `json:"error-codes"`
	ChallengeTS string   `json:"challenge_ts"`
	Hostname    string   `json:"hostname"`
	Action      string   `json:"action"`
	CData       string   `json:"cdata"`
}

func (c *Client) Verify(ctx context.Context, token, remoteIP string) error {
	token = strings.TrimSpace(token)
	if token == "" {
		return apperror.BadRequest("Security check required. Please complete the captcha and try again.")
	}

	form := url.Values{}
	form.Set("secret", c.secret)
	form.Set("response", token)
	if ip := strings.TrimSpace(remoteIP); ip != "" {
		form.Set("remoteip", ip)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, siteverifyURL, strings.NewReader(form.Encode()))
	if err != nil {
		return apperror.Internal("Failed to prepare security check")
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := c.httpClient.Do(req)
	if err != nil {
		return apperror.Internal("Security check temporarily unavailable. Please try again.")
	}
	defer res.Body.Close()

	body, err := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if err != nil {
		return apperror.Internal("Failed to read security check response")
	}

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return apperror.Internal("Security check temporarily unavailable. Please try again.")
	}

	var parsed siteverifyResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return apperror.Internal("Invalid security check response")
	}

	if !parsed.Success {
		return apperror.BadRequest(mapTurnstileError(parsed.ErrorCodes))
	}

	return nil
}

func mapTurnstileError(codes []string) string {
	for _, code := range codes {
		switch code {
		case "missing-input-secret", "invalid-input-secret":
			return "Security check is misconfigured. Contact the administrator."
		case "missing-input-response", "invalid-input-response":
			return "Security check expired or invalid. Please try again."
		case "timeout-or-duplicate":
			return "Security check already used or timed out. Please try again."
		case "bad-request":
			return "Security check failed. Please try again."
		case "internal-error":
			return "Security check temporarily unavailable. Please try again."
		}
	}
	if len(codes) > 0 {
		return fmt.Sprintf("Security check failed (%s). Please try again.", strings.Join(codes, ", "))
	}
	return "Security check failed. Please try again."
}

// noop skips verification when Turnstile is not configured (local development).
type noop struct{}

func (noop) Enabled() bool { return false }

func (noop) Verify(context.Context, string, string) error { return nil }
