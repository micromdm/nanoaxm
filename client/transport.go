package client

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/micromdm/nanoaxm/storage"
)

var (
	// ErrMissingName is returned when an HTTP context is missing the AxM name.
	ErrMissingName = errors.New("transport: missing AxM name in HTTP request context")

	// ErrNilNewMgr is returned when the new manager function is nil.
	ErrNilNewMgr = errors.New("transport: nil new mgr")
)

// TokenManager is a simple interface for a caching token manager.
type TokenManager[T any] interface {
	// GetOrRefreshToken retrieves (or refreshes) a token T.
	// The details of what and how the token are refreshed are implementation details.
	// Including any cache invalidation policies.
	GetOrRefreshToken(ctx context.Context, forceRefresh bool) (T, error)
}

// Transport is an HTTP round trip transport for Apple AxM API calls.
// It is used to transparently handle both access token and
// client assertion token requesting, caching, management (refresh/renew).
type Transport struct {
	// next is the "upstream" (real) HTTP RoundTripper.
	next http.RoundTripper

	// mgrs maps an AxM name to an access token manager.
	mgrs   map[string]TokenManager[string]
	mgrsMu sync.Mutex

	// newMgr instantiates a new token manager for the OAuth2 access token.
	newMgr func(ctx context.Context, axmName string) (TokenManager[string], error)
}

// NewTransport creates a new NanoAXM HTTP transport.
// If next is nil then the default HTTP transport is used.
func NewTransport(next http.RoundTripper, authDoer Doer, store storage.ClientAssertionRefresher, jtiFn func() string) *Transport {
	if next == nil {
		next = http.DefaultTransport
	}

	return &Transport{
		next: next,
		mgrs: make(map[string]TokenManager[string]),
		newMgr: func(ctx context.Context, axmName string) (TokenManager[string], error) {
			return NewAccessTokenManager(authDoer, axmName, store, jtiFn), nil
		},
	}
}

// getOrNewTokenManager either fetches the token manager for axmName or creates a new one.
func (t *Transport) getOrNewTokenManager(ctx context.Context, axmName string) (TokenManager[string], error) {
	if t.newMgr == nil {
		return nil, ErrNilNewMgr
	}

	t.mgrsMu.Lock()
	defer t.mgrsMu.Unlock()

	mgr, ok := t.mgrs[axmName]
	if ok && mgr != nil {
		return mgr, nil
	}

	var err error
	mgr, err = t.newMgr(ctx, axmName)
	if err != nil {
		// don't store the returned mgr if there's an error
		return mgr, err
	}

	t.mgrs[axmName] = mgr
	return mgr, nil
}

// RoundTrip sets an OAuth2 access token header on req and performs an HTTP round trip
// returning the response.
// If the round trip is Unauthorized then a second round trip is
// performed by first forcing an access token refresh.
func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	// first, get the AxM name from the context
	axmName := GetName(req.Context())
	if axmName == "" {
		return nil, ErrMissingName
	}

	ctx := req.Context()

	if ua := req.Header.Get("User-Agent"); ua != "" {
		// if a user agent is set, also use it for requesting
		ctx = WithGetTokenUserAgent(ctx, ua)
	}

	// retrieve (or make new) a token manager using the AxM name
	mgr, err := t.getOrNewTokenManager(ctx, axmName)
	if err != nil {
		return nil, fmt.Errorf("transport: getting token manager: %s: %w", axmName, err)
	}

	// start off the request by letting the token manager(s) manage refreshing their token(s)
	forceRefresh := false

getOrRefreshToken:
	// get the OAuth2 access token for this AxM name
	token, err := mgr.GetOrRefreshToken(ctx, forceRefresh)
	if err != nil {
		return nil, fmt.Errorf("transport: getting access token: %s: %w", axmName, err)
	}

	// set the OAuth2 access token authentication
	req.Header.Set("Authorization", "Bearer "+token)

	// perform the actual round-trip with our upstream round tripper
	resp, err := t.next.RoundTrip(req)
	if err != nil {
		return resp, err
	}

	if !forceRefresh && resp.StatusCode == 401 {
		// if we've received an unauthorized, then try again.
		// perhaps our token has a problem: force a refresh to try again.
		forceRefresh = true
		goto getOrRefreshToken
	}

	return resp, nil
}
