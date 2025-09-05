// Package test implements storage tests.
package test

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"log"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/micromdm/nanoaxm/client"
	"github.com/micromdm/nanoaxm/storage"
)

type GetClientAssertion func(ctx context.Context, axmName string) (storage.ClientAssertion, error)

func TestStorage(t *testing.T, ctx context.Context, s storage.AllStorage, gca GetClientAssertion) {
	if s == nil {
		panic("nil storage")
	}

	ac, err := s.RetrieveAuthCredentials(ctx, "test-axm-name-should-not-exist")
	if have, want := err, storage.ErrInvalidAXMName; !errors.Is(have, want) {
		t.Errorf("have: %v; want: %v", have, want)
		if validErr := ac.ValidError(); err == nil && validErr != nil {
			// if we get a nil error, we should definitely get creds
			t.Errorf("storage had no error but returned invalid creds: %v", validErr)
		}
	}

	err = s.StoreAuthCredentials(ctx, "test-axm-name-01", storage.AuthCredentials{})
	if err == nil {
		t.Fatal("should have errored, invalid auth creds")
	}

	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		log.Fatal(err)
	}

	// Encode private key to PEM
	privKeyBytes, err := x509.MarshalECPrivateKey(privKey)
	if err != nil {
		log.Fatal(err)
	}

	pem := pem.EncodeToMemory(&pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: privKeyBytes,
	})

	ac = storage.AuthCredentials{
		ClientID:      "test-client-id-01",
		KeyID:         "test-key-id-01",
		PrivateKeyPEM: pem,
	}

	err = s.StoreAuthCredentials(ctx, "test-axm-name-01", ac)
	if err != nil {
		t.Fatal(err)
	}

	ac2, err := s.RetrieveAuthCredentials(ctx, "test-axm-name-01")
	if err != nil {
		t.Fatal(err)
	}

	if have, want := ac2, ac; !reflect.DeepEqual(have, want) {
		t.Errorf("auth creds: have: %v; want: %v", have, want)
	}

	// client assertion tests
	_, err = s.GetOrRefreshClientAssertion(ctx, "test-axm-name-01", nil, false)
	if err == nil {
		t.Fatal("nil refresher fn must produce error but got nil")
	}

	var testTokenErr error
	var testToken storage.ClientAssertion
	now := time.Now()
	testToken.Validity = client.ClientAssertionDaysExpiry * 24 * time.Hour
	// truncate to second as:
	// 1. any storage backend may do that anyway and
	// 2. the JWT generation resolution is seconds
	testToken.Expiry = now.Add(testToken.Validity).Truncate(time.Second)
	jti := uuid.NewString()
	testToken.ClientID = ac.ClientID
	testToken.Token, testTokenErr = client.NewClientAssertion(ac, client.Audience, jti, now, testToken.Expiry)

	refresher := func(ctx context.Context, ac storage.AuthCredentials) (token storage.ClientAssertion, err error) {
		return testToken, testTokenErr
	}

	// store it first and force refresh (refresh=true)
	tok, err := s.GetOrRefreshClientAssertion(ctx, "test-axm-name-01", refresher, true)
	if err != nil {
		t.Fatal(err)
	}
	if err = tok.ValidError(); err != nil {
		t.Errorf("token not valid: %v", err)
	}

	if gca == nil {
		t.Fatal("gca function is nil")
	}

	// retrieve the client assertion directly from storage
	tTok, err := gca(ctx, "test-axm-name-01")
	if err != nil {
		t.Fatal(err)
	}
	if err = tTok.ValidError(); err != nil {
		t.Errorf("token not valid: %v", err)
	}

	// test to make sure the from-storage matches our cached/generated
	if have, want := tok, tTok; !reflect.DeepEqual(have, want) {
		t.Errorf("token: have: %v, want: %v", have, want)
	}

	// read it back again (refresher=false)
	tok, err = s.GetOrRefreshClientAssertion(ctx, "test-axm-name-01", refresher, false)
	if err != nil {
		t.Fatal(err)
	}
	if err = tTok.ValidError(); err != nil {
		t.Errorf("token not valid: %v", err)
	}

	// make sure the cached token is the same as our original token
	if have, want := tok, testToken; !reflect.DeepEqual(have, want) {
		t.Errorf("token: have: %v, want: %v", have, want)
	}
}
