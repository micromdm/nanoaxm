// Package client implements HTTP primitives for talking with and authenticating to the Apple AxM APIs.
package client

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Doer executes an HTTP request.
type Doer interface {
	Do(*http.Request) (*http.Response, error)
}

// ctxKeyName is the context key for the AxM name.
type ctxKeyName struct{}

// WithName creates a new context from ctx with the AxM name associated.
func WithName(ctx context.Context, axmName string) context.Context {
	return context.WithValue(ctx, ctxKeyName{}, axmName)
}

// GetName retrieves the AxM name from ctx.
func GetName(ctx context.Context) string {
	v, _ := ctx.Value(ctxKeyName{}).(string)
	return v
}

// HTTPError encapsulates an HTTP response error.
type HTTPError struct {
	Body       []byte
	Status     string
	StatusCode int
}

// Error returns the HTTP error as an error string.
func (e *HTTPError) Error() string {
	return fmt.Sprintf("HTTP error: %s: %s", e.Status, string(e.Body))
}

// NewHTTPError creates and returns a new HTTPError from r.
// Note this reads r.Body (limited to 1 KiB) and the caller is responsible for closing it.
func NewHTTPError(r *http.Response) error {
	body, readErr := io.ReadAll(io.LimitReader(r.Body, 1024))
	err := &HTTPError{
		Body:       body,
		Status:     r.Status,
		StatusCode: r.StatusCode,
	}
	if readErr != nil {
		return fmt.Errorf("reading body of HTTP error: %v: %w", err, readErr)
	}
	return err
}

// ClientWithTransport is a helper that returns a shallow copy of client with transport set.
func ClientWithTransport(client *http.Client, transport http.RoundTripper) *http.Client {
	client2 := *client
	client2.Transport = transport
	return &client2
}

// NewRequestWithContext creates a new request for an AxM name.
func NewRequestWithContext(ctx context.Context, axmName string, method, url string, body io.Reader) (*http.Request, error) {
	return http.NewRequestWithContext(WithName(ctx, axmName), method, url, body)
}

// pctDue calculates whether a given time is within a certain percentage of its validity period.
func pctDue(now time.Time, expiry time.Time, validity time.Duration, pct float64) bool {
	rPct := 1.0 - pct
	return expiry.Sub(now) < time.Duration(float64(validity)*rPct)
}

// newPctDue returns a function that calculates whether a given expiry time is within a certain percentage of its validity period.
func newPctDue(pct float64) func(expiry time.Time, validity time.Duration) bool {
	return newPctDueAt(time.Now, pct)
}

// newPctDueAt returns a function that calculates whether a given expiry time is within a certain percentage of its validity period, using the provided now function.
func newPctDueAt(nowFn func() time.Time, pct float64) func(expiry time.Time, validity time.Duration) bool {
	return func(expiry time.Time, validity time.Duration) bool {
		return pctDue(nowFn(), expiry, validity, pct)
	}
}
