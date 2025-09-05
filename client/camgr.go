package client

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/micromdm/nanoaxm/storage"
)

// CANameData is the Client Assertion and AxM name data.
// This structure is the "token" data managed by the Client Assertion token manager.
type CANameData struct {
	ClientAssertion string

	// ClientID can be used to determine the OAuth2 scope.
	// I.e. Apple Business Manager or Apple School Manager.
	ClientID string
}

type ClientAssertionTokenManager struct {
	axmName           string
	store             storage.ClientAssertionRefresher
	clientAssertion   storage.ClientAssertion
	clientAssertionMu sync.Mutex
	due               func(time.Time, time.Duration) bool
	jtiFn             func() string
}

// NewClientAssertionTokenManager creates a new client assertion token manager.
// Panics if axmName or store are nil.
// Tries to refresh the client assertion about 9 days before expiry (95% of validity).
func NewClientAssertionTokenManager(axmName string, store storage.ClientAssertionRefresher, jtiFn func() string) *ClientAssertionTokenManager {
	if axmName == "" {
		panic("nil AxM name")
	}
	if store == nil {
		panic("nil store")
	}
	if jtiFn == nil {
		panic("nil JTI func")
	}

	return &ClientAssertionTokenManager{
		axmName: axmName,
		store:   store,
		jtiFn:   jtiFn,

		// at a Apple documented expiry of 180 days, 0.95 makes the
		// refresh 171 days or 9 days before expiry.
		due: newPctDue(0.95),
	}
}

// clientAssertionRefreshFunc generates a new Client Assertion
func clientAssertionRefreshFunc(jtiFn func() string) func(ctx context.Context, ac storage.AuthCredentials) (ca storage.ClientAssertion, err error) {
	return func(ctx context.Context, ac storage.AuthCredentials) (ca storage.ClientAssertion, err error) {
		now := time.Now()
		ca.Validity = ClientAssertionDaysExpiry * 24 * time.Hour
		ca.Expiry = now.Add(ca.Validity)
		ca.JTI = jtiFn()
		ca.ClientID = ac.ClientID
		ca.Token, err = NewClientAssertion(ac, Audience, ca.JTI, now, ca.Expiry)
		return
	}
}

// GetOrRefreshToken retrieves (or refreshes) the OAuth2 client assertion token.
// Exits early if the token manager store or due helper are not set.
func (m *ClientAssertionTokenManager) GetOrRefreshToken(ctx context.Context, forceRefresh bool) (CANameData, error) {
	if m.due == nil {
		return CANameData{}, ErrDueNil
	}
	if m.store == nil {
		return CANameData{}, errors.New("nil store")
	}

	m.clientAssertionMu.Lock()
	defer m.clientAssertionMu.Unlock()

	if !forceRefresh && m.clientAssertion.Valid() && !m.due(m.clientAssertion.Expiry, m.clientAssertion.Validity) {
		return CANameData{
			ClientAssertion: m.clientAssertion.Token,
			ClientID:        m.clientAssertion.ClientID,
		}, nil
	}

	ca, err := m.store.GetOrRefreshClientAssertion(ctx, m.axmName, clientAssertionRefreshFunc(m.jtiFn), forceRefresh)
	caNameData := CANameData{
		ClientAssertion: ca.Token,
		ClientID:        ca.ClientID,
	}
	if err != nil {
		return caNameData, err
	}
	if err = ca.ValidError(); err != nil {
		return caNameData, err
	}

	m.clientAssertion = ca

	return caNameData, nil
}
