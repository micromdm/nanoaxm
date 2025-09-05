package storage

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// ErrInvalidAXMName occurs when an "AxM name" is empty or missing.
var ErrInvalidAXMName = errors.New("invalid AxM name")

// ErrInvalidAuthCredentials occurs when authentication credentials fail validity checks.
// Possibly due to missing or invalid fields.
var ErrInvalidAuthCredentials = errors.New("invalid auth creds")

// AuthCredentials are the core authentication components needed to issue access tokens.
type AuthCredentials struct {
	// Client ID and Key ID are provided in the AxM portal.
	ClientID, KeyID string

	// The private key in PEM encoded form. Provided in the AxM portal.
	// Note Apple restricts the private key to a single download from the portal.
	PrivateKeyPEM []byte
}

// ValidError tests ac for common missing or invalid fields.
func (ac AuthCredentials) ValidError() error {
	if ac.ClientID == "" {
		return fmt.Errorf("%w: empty client ID", ErrInvalidAuthCredentials)
	}
	if ac.KeyID == "" {
		return fmt.Errorf("%w: empty key ID", ErrInvalidAuthCredentials)
	}
	if ac.PrivateKeyPEM == nil {
		return fmt.Errorf("%w: nil private key", ErrInvalidAuthCredentials)
	}
	return nil
}

// Valid returns true if ac is valid.
func (ac AuthCredentials) Valid() bool {
	return ac.ValidError() == nil
}

type AuthCredentialsRetriever interface {
	// RetrieveAuthCredential retrieves the auth crendetials from storage for axmName.
	// [ErrInvalidAXMName] should be returned if axmName is invalid.
	RetrieveAuthCredentials(ctx context.Context, axmName string) (AuthCredentials, error)
}

type AuthCredentialsStorer interface {
	// StoreAuthCredentials stores the auth credentials to storage for axmName.
	// [ErrInvalidAXMName] should be returned if axmName is invalid.
	// Implementations should be agnostic about the actual data in the private key field.
	// This is to facilitate a potential key encryption encapsulation.
	StoreAuthCredentials(ctx context.Context, axmName string, ac AuthCredentials) error
}

// ClientAssertion is a token with an expiry and validity.
type ClientAssertion struct {
	Token    string
	Validity time.Duration
	Expiry   time.Time
	ClientID string
	JTI      string
}

var (
	ErrEmptyToken    = errors.New("empty token")
	ErrEmptyValidity = errors.New("empty validity")
	ErrExpiryZero    = errors.New("zero expiry")
	ErrEmptyClientID = errors.New("empty client ID")
)

// ValidError returns the specific error causing t to be invalid.
func (t ClientAssertion) ValidError() error {
	if t.Token == "" {
		return ErrEmptyToken
	}
	if t.Validity <= 0 {
		return ErrEmptyValidity
	}
	if t.Expiry.IsZero() {
		return ErrExpiryZero
	}
	if t.ClientID == "" {
		return ErrEmptyClientID
	}
	return nil
}

// Valid returns true if t is valid.
func (t ClientAssertion) Valid() bool {
	return t.ValidError() == nil
}

type ClientAssertionRefresher interface {
	// GetOrRefreshClientAssertion refreshes the OAuth2 client assertion for axmName and stores it.
	// Implementations should beware that this function may be called
	// simultaneously and may need to implement some sort of distributed
	// (i.e. database) locking mechanism.
	// Failing to do so may result in multiple client assertions being
	// generated at the same time, possibly overwriting each other.
	GetOrRefreshClientAssertion(ctx context.Context, axmName string, refreshFunc func(ctx context.Context, ac AuthCredentials) (ClientAssertion, error), refresh bool) (ClientAssertion, error)
}

type AllStorage interface {
	AuthCredentialsRetriever
	AuthCredentialsStorer
	ClientAssertionRefresher
}
