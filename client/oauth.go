package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/micromdm/nanoaxm/cryptoutil"
	"github.com/micromdm/nanoaxm/storage"

	"github.com/golang-jwt/jwt/v5"
)

const (
	Audience = "https://account.apple.com/auth/oauth2/v2/token"

	ClientAssertionType       = "urn:ietf:params:oauth:client-assertion-type:jwt-bearer"
	ClientAssertionDaysExpiry = 180
)

// TokenResponse represents the OAuth 2 successful token response structure.
// See https://datatracker.ietf.org/doc/html/rfc6749#section-5.1
type TokenResponse struct {
	// AccessToken authenticates requests.
	// Required to make calls to Apple School and Business Manager API endpoints.
	// Valid for the number of seconds that "expires_in" specifies.
	AccessToken string `json:"access_token"`

	// TypeType contains the type of token.
	// The value is "Bearer".
	TokenType string `json:"token_type"`

	// // RefreshToken is the token used for refreshing the access token.
	// // Unused by the AxM API.
	// RefreshToken string `json:"refresh_token"`

	// Token expiry time in seconds.
	// The token lifetime (TTL) of one hour (3600 seconds).
	ExpiresIn int64 `json:"expires_in"`

	// Defines the access permissions you request from the app, and
	// limits the authorization of the access token you receive.
	Scope string `json:"scope,omitempty"`
}

// ErrorResponse represents the OAuth 2 error token response structure.
// See https://datatracker.ietf.org/doc/html/rfc6749#section-5.2
type ErrorResponse struct {
	// ErrorString contains a pre-defined OAuth2 error string.
	ErrorString string `json:"error"`

	// ErrorDescription may contain additional OAuth2 error information.
	ErrorDescription string `json:"error_description,omitempty"`

	// ErrorURI may contain a link to additional information.
	ErrorURI string `json:"error_uri"`
}

// Error returns a "collapsed" form of er by joining the fields with a semi-colon (";").
func (er *ErrorResponse) Error() string {
	if er == nil {
		return "nil error response"
	}
	s := er.ErrorString
	if s == "" {
		s = "empty error"
	}
	if er.ErrorDescription != "" {
		s += "; " + er.ErrorDescription
	}
	if er.ErrorURI != "" {
		s += "; " + er.ErrorURI
	}
	return s
}

// NewClientAssertion generates and returns a new signed client assertion.
func NewClientAssertion(ac storage.AuthCredentials, audience, jti string, now time.Time, expiry time.Time) (string, error) {
	err := ac.ValidError()
	if err != nil {
		return "", fmt.Errorf("new client assertion: auth creds invalid: %w", err)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
		"sub": ac.ClientID,
		"aud": audience,
		"iat": now.Unix(),
		"exp": expiry.Unix(),
		"jti": jti,
		"iss": ac.ClientID, // team ID
	})

	token.Header["kid"] = ac.KeyID

	privKey, err := cryptoutil.ECPrivateKeyFromPEM(ac.PrivateKeyPEM)
	if err != nil {
		return "", fmt.Errorf("new client assertion: key from PEM: %w", err)
	}

	clientAssertion, err := token.SignedString(privKey)
	if err != nil {
		return "", fmt.Errorf("new client assertion: signing: %w", err)
	}

	return clientAssertion, nil
}

// ctxKeyGetTokenUserAgent is the context key for the user agent string when requesting an access token.
type ctxKeyGetTokenUserAgent struct{}

// WithGetTokenUserAgent creates a new context from ctx with the user agent associated.
// Only useful for overiding the user agent in [DoGetToken].
func WithGetTokenUserAgent(ctx context.Context, userAgent string) context.Context {
	return context.WithValue(ctx, ctxKeyGetTokenUserAgent{}, userAgent)
}

// DoGetToken requests the OAuth2 access token from [Audience] using doer.
// Uses clientID to determine the OAuth 2 scope and clientAssertion for authentication.
// Will apply a user agent if [WithGetTokenUserAgent] was used in ctx.
func DoGetToken(ctx context.Context, doer Doer, clientID, clientAssertion string) (*TokenResponse, error) {
	if doer == nil {
		return nil, errors.New("nil doer")
	}
	if clientID == "" {
		return nil, errors.New("empty client ID")
	}
	if clientAssertion == "" {
		return nil, errors.New("empty clientAssertion")
	}

	scope := "school.api"
	if strings.HasPrefix(clientID, "BUSINESSAPI.") {
		scope = "business.api"
	}

	form := url.Values{
		"grant_type":            {"client_credentials"},
		"client_id":             {clientID},
		"client_assertion_type": {ClientAssertionType},
		"client_assertion":      {clientAssertion},
		"scope":                 {scope},
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		Audience,
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	if ua, _ := ctx.Value(ctxKeyGetTokenUserAgent{}).(string); ua != "" {
		req.Header.Set("User-Agent", ua)
	}

	// send the request
	resp, err := doer.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	switch resp.StatusCode {
	case 400:
		err := new(ErrorResponse)
		decErr := json.NewDecoder(resp.Body).Decode(err)
		if decErr != nil {
			return nil, fmt.Errorf("decoding error response: %v: %w", err, decErr)
		}
		return nil, err
	case 200:
		token := new(TokenResponse)
		decErr := json.NewDecoder(resp.Body).Decode(token)
		if decErr != nil {
			return nil, fmt.Errorf("decoding token response: %w", decErr)
		}
		return token, nil
	default:
		return nil, NewHTTPError(resp)
	}
}
