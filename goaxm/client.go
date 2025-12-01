// Package goaxm provides Go methods and structures for talking to individual AxM API endpoints.
package goaxm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
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

// defaultOutStatus is the "success" HTTP status that will allow parsing of the "out."
var defaultOutStatus = http.StatusOK

// Do executes the AxM HTTP call against url using method.
// OAuth credentials are looked-up and used by axmName.
// JSON is marshaled and decoded from the HTTP request and response
// bodies using in and out respectively.
// A successful HTTP code is specified in outStatus.
// A default status will be used if outStatus is 0.
// To parse potential errors as the AxM Error value, specify them with errStatuses.
// A default set will be used if errStatuses is empty.
func (c *Client) Do(ctx context.Context, axmName, method, url string, in, out any, outStatus int, errStatuses []int) error {
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

	if outStatus < 1 {
		outStatus = defaultOutStatus
	}
	if resp.StatusCode != outStatus {
		return ABMError(resp, errStatuses)
	}

	if out != nil {
		err := json.NewDecoder(resp.Body).Decode(out)
		if err != nil {
			return err
		}
	}

	return nil
}

// defaultErrStatuses is the default set of HTTP statuses that will generate HTTP errors.
var defaultErrStatuses = []int{
	http.StatusBadRequest,
	http.StatusUnauthorized,
	http.StatusForbidden,
	http.StatusTooManyRequests,
}

// ABMError searches errStatuses for the response code in r and returns a
// [ABMErrorResponseError] if found, otherwise an HTTP error.
func ABMError(r *http.Response, errStatuses []int) error {
	if errStatuses == nil {
		errStatuses = defaultErrStatuses
	}
	foundErrStatus := false
	for _, e := range errStatuses {
		if r.StatusCode == e {
			foundErrStatus = true
			break
		}
	}
	if foundErrStatus {
		return NewABMErrorResponseErrorFromReader(r.Body)
	}
	return client.NewHTTPError(r)
}
