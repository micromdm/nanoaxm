package client

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/micromdm/nanoaxm/storage"
)

// ErrDueNil is returned when then token "due" calculation function is not specified.
var ErrDueNil = errors.New("due nil")

// at contains the access token and expiry metadata.
type at struct {
	token    string
	expiry   time.Time
	validity time.Duration
}

// valid returns true if the t is valid.
func (t at) valid() bool {
	if t.validity > 0 && t.token != "" && !t.expiry.IsZero() {
		return true
	}
	return false
}

// AccessTokenManager manages the caching and renewal of the OAuth2 access token.
type AccessTokenManager struct {
	doer Doer
	tm   TokenManager[CANameData]
	mu   sync.Mutex
	at   at
	due  func(time.Time, time.Duration) bool
}

// NewAccessTokenManager creates a new access token token manager.
// Panics if doer is nil.
// Tries to refresh the access token about 5 minutes before expiry (80% of validity).
func NewAccessTokenManager(doer Doer, axmName string, store storage.ClientAssertionRefresher, jtiFn func() string) *AccessTokenManager {
	if doer == nil {
		panic("nil doer")
	}

	return &AccessTokenManager{
		doer: doer,
		tm:   NewClientAssertionTokenManager(axmName, store, jtiFn),

		// at an Apple documented expiry of an hour, 0.8 makes the
		// refresh 288 seconds or 4.8 minutes before expiry.
		due: newPctDue(0.8),
	}
}

// GetOrRefreshToken retrieves (or refreshes) the OAuth2 access token.
// Exits early if the due helper is not set.
func (m *AccessTokenManager) GetOrRefreshToken(ctx context.Context, forceRefresh bool) (string, error) {
	if m.due == nil {
		return "", ErrDueNil
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if !forceRefresh && m.at.valid() && !m.due(m.at.expiry, m.at.validity) {
		// return the currently cached token
		return m.at.token, nil
	}

	ca, err := m.tm.GetOrRefreshToken(ctx, forceRefresh)
	if err != nil {
		return "", fmt.Errorf("getting client assertion: %w", err)
	}

	// retrieve a new access token
	tr, err := DoGetToken(ctx, m.doer, ca.ClientID, ca.ClientAssertion)
	if err != nil {
		return "", fmt.Errorf("fetching access token: %w", err)
	}

	expiresIn := time.Duration(tr.ExpiresIn) * time.Second

	m.at = at{
		token:    tr.AccessToken,
		validity: expiresIn,
		expiry:   time.Now().Add(expiresIn),
	}

	return m.at.token, nil
}
