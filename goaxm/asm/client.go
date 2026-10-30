// Package asm provides Go methods and structures for the Apple School Manager API.
package asm

import (
	"github.com/micromdm/nanoaxm/goaxm"
	"github.com/micromdm/nanoaxm/storage"
)

// Client is a client for the Apple School Manager API.
// It embeds the generic goaxm client (transport, OAuth, JSON handling) and
// adds the API-specific methods generated in endpoints_asm.go.
type Client struct {
	*goaxm.Client
}

// NewClient creates a new Apple School Manager API client.
func NewClient(store storage.ClientAssertionRefresher, opts ...goaxm.Option) *Client {
	// Prepend so any caller-provided WithErrorFunc takes precedence.
	opts = append([]goaxm.Option{goaxm.WithErrorFunc(ASMError)}, opts...)
	return &Client{Client: goaxm.NewClient(store, opts...)}
}
