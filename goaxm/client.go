// Package goaxm provides Go methods and structures for talking to individual AxM API endpoints.
package goaxm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/google/uuid"
	"github.com/micromdm/nanoaxm/client"
	"github.com/micromdm/nanoaxm/storage"
)

const (
	// the API results in a 406 accept error if the
	// header value differs from this
	acceptMediaType = "application/json"

	DefaultUserAgent = "nanoaxm/0"
)

// ErrNilClient happens when an HTTP client is not initialized.
var ErrNilClient = errors.New("nil client")

type config struct {
	ua     string
	client *http.Client
	jtiFn  func() string
}

// Client is a simple HTTP client for sending requests to
// Apple Business Manager and Apple School Manager.
type Client struct {
	ua   string
	doer client.Doer
}

// Options configure clients.
type Option func(*config)

func NewClient(store storage.ClientAssertionRefresher, opts ...Option) *Client {
	cfg := &config{
		ua:     DefaultUserAgent,
		client: http.DefaultClient,
		jtiFn:  uuid.NewString,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	// wrap the HTTP client with the NanoAxM transport
	ct := client.ClientWithTransport(
		cfg.client,
		client.NewTransport(cfg.client.Transport, cfg.client, store, cfg.jtiFn),
	)

	return &Client{
		ua:   cfg.ua,
		doer: ct,
	}
}

// WithUserAgent sets the the HTTP User-Agent string to be used for each request.
func WithUserAgent(ua string) Option {
	return func(c *config) {
		c.ua = ua
	}
}

// WithClient configures the HTTP client to be used.
// The provided client is copied and modified by wrapping its
// transport in a new NanoAxM transport (which transparently handles
// OAuth 2 authentication).
func WithClient(client *http.Client) Option {
	return func(c *config) {
		c.client = client
	}
}

// WithJTI configures a function for providing the JTI for OAuth.
// By default a random UUID is generated.
func WithJTI(jtiFn func() string) Option {
	return func(c *config) {
		c.jtiFn = jtiFn
	}
}

// Do executes the AxM HTTP call against using method and url.
// OAuth utilizes credentials of axmName.
// JSON is marshaled and decoded from the HTTP request and response
// bodies using in and out respectively.
func (c *Client) Do(ctx context.Context, axmName, method, url string, in, out any) error {
	if c.doer == nil {
		return errors.New("nil client")
	}

	var body io.Reader
	if in != nil {
		bodyBytes, err := json.Marshal(in)
		if err != nil {
			return err
		}
		body = bytes.NewBuffer(bodyBytes)
	}

	req, err := client.NewRequestWithContext(ctx, axmName, method, url, body)
	if err != nil {
		return err
	}
	if c.ua != "" {
		req.Header.Set("User-Agent", c.ua)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json;charset=UTF8")
	}
	if out != nil {
		req.Header.Set("Accept", acceptMediaType)
	}

	resp, err := c.doer.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("unhandled auth error: %w", client.NewHTTPError(resp))
	} else if resp.StatusCode != http.StatusOK {
		return client.NewHTTPError(resp)
	}

	if out != nil {
		err := json.NewDecoder(resp.Body).Decode(out)
		if err != nil {
			return err
		}
	}

	return nil
}
