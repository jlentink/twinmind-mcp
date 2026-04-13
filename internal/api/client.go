package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/jlentink/twinmind-mcp/internal/auth"
	"github.com/jlentink/twinmind-mcp/internal/config"
)

const (
	vercelBypassHeader = "x-vercel-protection-bypass"
	vercelBypassValue  = "K1oNTqR7cjtbhehlqxQgxSP9As13QAeE"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
	cfg        *config.Config
}

func NewClient(cfg *config.Config) *Client {
	baseURL := cfg.API.BaseURL
	if baseURL == "" {
		baseURL = config.DefaultBaseURL
	}
	return &Client{
		baseURL:    baseURL,
		httpClient: &http.Client{},
		cfg:        cfg,
	}
}

func (c *Client) ensureToken() error {
	token := &auth.TokenPair{
		IDToken:      c.cfg.Auth.IDToken,
		RefreshToken: c.cfg.Auth.RefreshToken,
		ExpiresAt:    c.cfg.Auth.IDTokenExpiry,
	}

	if !token.IsExpired() {
		return nil
	}

	if c.cfg.Auth.RefreshToken == "" {
		return fmt.Errorf("not authenticated, run: twinmind-cli auth login")
	}

	refreshed, err := auth.ExchangeRefreshToken(c.cfg.Auth.RefreshToken)
	if err != nil {
		return fmt.Errorf("token refresh failed: %w", err)
	}

	c.cfg.Auth.IDToken = refreshed.IDToken
	c.cfg.Auth.RefreshToken = refreshed.RefreshToken
	c.cfg.Auth.IDTokenExpiry = refreshed.ExpiresAt

	if err := config.Save(c.cfg); err != nil {
		return fmt.Errorf("failed to save refreshed token: %w", err)
	}

	return nil
}

func (c *Client) do(ctx context.Context, method, path string, body any) ([]byte, error) {
	if err := c.ensureToken(); err != nil {
		return nil, err
	}

	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request: %w", err)
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.cfg.Auth.IDToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(vercelBypassHeader, vercelBypassValue)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}
