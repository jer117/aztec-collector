// Package jsonrpc provides a JSON-RPC 2.0 client for the Aztec node API
package jsonrpc

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync/atomic"

	"github.com/go-resty/resty/v2"
)

// Request represents a JSON-RPC 2.0 request
type Request struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
	ID      uint64      `json:"id"`
}

// Response represents a JSON-RPC 2.0 response
type Response struct {
	JSONRPC string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *RPCError       `json:"error,omitempty"`
	ID      uint64          `json:"id"`
}

// RPCError represents a JSON-RPC 2.0 error
type RPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func (e *RPCError) Error() string {
	return fmt.Sprintf("JSON-RPC error %d: %s", e.Code, e.Message)
}

// Client is a JSON-RPC 2.0 client
type Client struct {
	client    *resty.Client
	url       string
	requestID uint64
}

// New creates a new JSON-RPC client with the given options
func New(options ...func(*Client)) (*Client, error) {
	c := &Client{
		client: resty.NewWithClient(http.DefaultClient),
	}

	for _, opt := range options {
		opt(c)
	}

	if c.url == "" {
		return nil, fmt.Errorf("URL is required")
	}

	return c, nil
}

// WithURL sets the URL for the JSON-RPC client
func WithURL(url string) func(*Client) {
	return func(c *Client) {
		c.url = url
	}
}

// WithHTTPClient sets a custom HTTP client
func WithHTTPClient(httpClient *http.Client) func(*Client) {
	return func(c *Client) {
		c.client = resty.NewWithClient(httpClient)
	}
}

// Call makes a JSON-RPC call and unmarshals the result into the target
func (c *Client) Call(method string, params interface{}, target interface{}) error {
	id := atomic.AddUint64(&c.requestID, 1)

	req := Request{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
		ID:      id,
	}

	resp, err := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(req).
		Post(c.url)

	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}

	if resp.IsError() {
		return fmt.Errorf("HTTP error %d: %s", resp.StatusCode(), resp.String())
	}

	var rpcResp Response
	if err := json.Unmarshal(resp.Body(), &rpcResp); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if rpcResp.Error != nil {
		return rpcResp.Error
	}

	if target != nil && rpcResp.Result != nil {
		if err := json.Unmarshal(rpcResp.Result, target); err != nil {
			return fmt.Errorf("failed to unmarshal result: %w", err)
		}
	}

	return nil
}

// URL returns the client's URL
func (c *Client) URL() string {
	return c.url
}

