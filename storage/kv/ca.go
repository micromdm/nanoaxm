package kv

import (
	"context"
	"errors"
	"fmt"

	"github.com/micromdm/nanoaxm/storage"
	"github.com/micromdm/nanolib/storage/kv"
)

const (
	keyPfxCA = "astn"

	keySfxCAToken    = "tok"
	keySfxCAValidity = "vld"
	keySfxCAExpiry   = "exp"
)

func storeClientAssertion(ctx context.Context, axmName string, b kv.RWBucket, token storage.ClientAssertion) error {
	err := kv.SetMap(ctx, b, map[string][]byte{
		join(keyPfxCA, axmName, keySfxCAToken):    []byte(token.Token),
		join(keyPfxCA, axmName, keySfxCAValidity): durationToBytes(token.Validity),
		join(keyPfxCA, axmName, keySfxCAExpiry):   timeToBytes(token.Expiry),
	})
	if err != nil {
		return fmt.Errorf("setting keys: %w", err)
	}
	return nil
}

func retrieveClientAssertion(ctx context.Context, axmName string, b kv.ROBucket) (storage.ClientAssertion, error) {
	var token storage.ClientAssertion

	sMap, err := kv.GetMap(ctx, b, []string{
		join(keyPfxCA, axmName, keySfxCAToken),
		join(keyPfxCA, axmName, keySfxCAValidity),
		join(keyPfxCA, axmName, keySfxCAExpiry),

		// also pull this from the "authcreds" keys
		join(keyPfxAC, axmName, keySfxClientID),
	})
	if err != nil {
		return token, fmt.Errorf("getting keys: %w", err)
	}

	token.Token = string(sMap[join(keyPfxCA, axmName, keySfxCAToken)])
	token.ClientID = string(sMap[join(keyPfxAC, axmName, keySfxClientID)])

	token.Validity, err = durationFromBytes(sMap[join(keyPfxCA, axmName, keySfxCAValidity)])
	if err != nil {
		return token, fmt.Errorf("converting validity: %w", err)
	}

	token.Expiry, err = timeFromBytes(sMap[join(keyPfxCA, axmName, keySfxCAExpiry)])
	if err != nil {
		return token, fmt.Errorf("converting expiry: %w", err)
	}

	return token, nil
}

func (s *KV) RetrieveClientAssertion(ctx context.Context, axmName string) (storage.ClientAssertion, error) {
	return retrieveClientAssertion(ctx, axmName, s.b)
}

// GetOrRefreshClientAssertion refreshes the OAuth 2 client assertion for axmName and stores it.
func (s *KV) GetOrRefreshClientAssertion(ctx context.Context, axmName string, refreshFunc func(ctx context.Context, ac storage.AuthCredentials) (storage.ClientAssertion, error), refresh bool) (token storage.ClientAssertion, err error) {
	if axmName == "" {
		return token, fmt.Errorf("%w: empty name", storage.ErrInvalidAXMName)
	}
	if refreshFunc == nil {
		return token, errors.New("nil refresher")
	}

	if !refresh {
		token, err = retrieveClientAssertion(ctx, axmName, s.b)
		if errors.Is(err, kv.ErrKeyNotFound) {
			refresh = true
		} else if err != nil {
			return token, fmt.Errorf("retreive client assertion: %w", err)
		}

		// TODO: calculate validity ourselves?
		if token.Valid() {
			return token, nil
		}
	}

	ac, err := s.RetrieveAuthCredentials(ctx, axmName)
	if err != nil {
		return token, fmt.Errorf("retrieving auth creds: %w", err)
	}

	token, err = refreshFunc(ctx, ac)
	if err != nil {
		return token, fmt.Errorf("refreshing client assertion: %w", err)
	}

	if err = token.ValidError(); err != nil {
		return token, fmt.Errorf("refreshed token invalid: %w", err)
	}

	err = storeClientAssertion(ctx, axmName, s.b, token)
	if err != nil {
		return token, fmt.Errorf("storing client assertion: %w", err)
	}

	return token, nil
}
