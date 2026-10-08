// Package api is a client for the mihomo external-controller REST API.
package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client talks to a mihomo external-controller.
type Client struct {
	base   string
	secret string
	hc     *http.Client // short JSON calls
	// streamHC has no timeout: streaming endpoints are long-lived.
	streamHC     *http.Client
	retryBackoff time.Duration // reconnect delay for streams
}

// New builds a client for the given base URL, e.g. http://127.0.0.1:9090.
func New(baseURL, secret string) *Client {
	return &Client{
		base:         strings.TrimSuffix(baseURL, "/"),
		secret:       secret,
		hc:           &http.Client{Timeout: 15 * time.Second},
		streamHC:     &http.Client{},
		retryBackoff: time.Second,
	}
}

// BaseURL returns the controller address.
func (c *Client) BaseURL() string { return c.base }

// APIError is a non-2xx response from the controller.
type APIError struct {
	Status  int
	Message string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("%s (HTTP %d)", e.Message, e.Status)
}

func (c *Client) newReq(ctx context.Context, method, path string, body any) (*http.Request, error) {
	var rdr io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		rdr = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.base+path, rdr)
	if err != nil {
		return nil, err
	}
	if c.secret != "" {
		req.Header.Set("Authorization", "Bearer "+c.secret)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return req, nil
}

// doJSON performs a request and decodes a JSON response into out (may be nil).
func (c *Client) doJSON(ctx context.Context, method, path string, body, out any) error {
	req, err := c.newReq(ctx, method, path, body)
	if err != nil {
		return err
	}
	resp, err := c.hc.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, 32<<20))
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &APIError{Status: resp.StatusCode, Message: decodeMessage(data, resp.StatusCode)}
	}
	if out == nil || len(bytes.TrimSpace(data)) == 0 {
		return nil
	}
	return json.Unmarshal(data, out)
}

func decodeMessage(data []byte, status int) string {
	var e struct {
		Message string `json:"message"`
	}
	if err := json.Unmarshal(data, &e); err == nil && e.Message != "" {
		return e.Message
	}
	if s := strings.TrimSpace(string(data)); s != "" {
		return s
	}
	return http.StatusText(status)
}
