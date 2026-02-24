package rest

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"oss.nandlabs.io/golly/clients"
	"oss.nandlabs.io/golly/textutils"
)

const (
	EXPIRY_EPOCH  = "__expiry_epoch__"
	GRANT_TYPE    = "grant_type"
	CLIENT_ID     = "client_id"
	CLIENT_SECRET = "client_secret"
	ACCESS_TOKEN  = "access_token"
	EXPIRES_IN    = "expires_in"
)

// oAuth2Provider represents a client for OAuth 2.0 authentication flow.
// It encapsulates the necessary credentials and configuration to obtain
// and manage OAuth 2.0 access tokens from an authorization server.
//
// The provider supports configurable token endpoints, client credentials,
// and additional parameters required for the OAuth 2.0 protocol.
// It maintains thread-safety through a mutex when accessing token data.
type oAuth2Provider struct {
	clientId      string
	clientSecret  string
	grantType     string
	tokenEndpoint string
	extraParams   map[string]any
	tokenData     map[string]any
	client        *Client
	lock          *sync.Mutex
}

func NewOAuth2Provider(clientId, clientSecret, grantType, tokenEndpoint string) clients.AuthProvider {
	return NewOAuth2ProviderWithClient(clientId, clientSecret, grantType, tokenEndpoint, nil)
}

// NewOAuth2ProviderWithClient creates a new OAuth2Provider with a custom REST client.
// If client is nil, a default client will be used.
// Use this when the token endpoint requires custom TLS, proxy, or timeout configuration.
func NewOAuth2ProviderWithClient(clientId, clientSecret, grantType, tokenEndpoint string, client *Client) clients.AuthProvider {
	if client == nil {
		client = NewClient()
	}
	return &oAuth2Provider{
		clientId:      clientId,
		clientSecret:  clientSecret,
		grantType:     grantType,
		tokenEndpoint: tokenEndpoint,
		extraParams:   make(map[string]any),
		tokenData:     make(map[string]any),
		client:        client,
		lock:          &sync.Mutex{},
	}
}

// Type returns the OAuth2 provider's authentication type.
func (o *oAuth2Provider) Type() clients.AuthType {
	return clients.AuthTypeBearer
}

// User returns the OAuth2 client ID which represents the user identifier for the provider.
// This method satisfies the Provider interface by providing access to the client identifier.
func (o *oAuth2Provider) User() (string, error) {
	return o.clientId, nil
}

// Pass returns the OAuth2Provider's client secret.
// This method is used to access the client secret in a controlled manner.
func (o *oAuth2Provider) Pass() (string, error) {
	return o.clientSecret, nil
}

// AddParam adds a key-value parameter to the OAuth2Provider's extra parameters.
// If the extra parameters map is nil, it initializes a new map before adding the parameter.
//
// Parameters:
//   - key: The key name for the parameter
//   - value: The value for the parameter, which can be of any type
func (o *oAuth2Provider) AddParam(key string, value any) {
	if o.extraParams == nil {
		o.extraParams = make(map[string]any)
	}
	o.extraParams[key] = value
}

// Token returns the OAuth2 access token for use in authenticating requests.
//
// The method first checks if there's a valid token that hasn't expired yet and returns it.
// If the token is expired or doesn't exist, it requests a new token from the OAuth2 provider
// using the configured credentials and parameters.
//
// The method handles token refresh automatically by:
// 1. Creating a form-encoded request to the token endpoint
// 2. Including client ID, client secret, grant type, and any extra parameters
// 3. Storing the token response data including expiry information
//
// Returns:
//   - The access token as a string if successful
//   - An error if any step of the token acquisition process fails
//
// Thread safety:
//
//	Uses mutex locking to ensure concurrent calls don't interfere with token refresh
func (o *oAuth2Provider) Token() (string, error) {
	o.lock.Lock()
	defer o.lock.Unlock()

	// Check if we have a valid cached token
	if token := o.getCachedToken(); token != textutils.EmptyStr {
		return token, nil
	}

	// No valid token, request a new one
	request, err := o.client.NewRequest(o.tokenEndpoint, http.MethodPost)
	if err != nil {
		return textutils.EmptyStr, fmt.Errorf("error creating token request: %w", err)
	}
	request.SetContentType("application/x-www-form-urlencoded")
	request.AddFormData(GRANT_TYPE, o.grantType)
	request.AddFormData(CLIENT_ID, o.clientId)
	request.AddFormData(CLIENT_SECRET, o.clientSecret)
	if o.extraParams != nil {
		for k, v := range o.extraParams {
			request.AddFormData(k, fmt.Sprintf("%v", v))
		}
	}
	response, err := o.client.Execute(request)
	if err != nil {
		return textutils.EmptyStr, fmt.Errorf("error executing token request: %w", err)
	}
	if response.StatusCode() != http.StatusOK {
		return textutils.EmptyStr, fmt.Errorf("token endpoint returned status %d", response.StatusCode())
	}
	if err := response.Decode(&o.tokenData); err != nil {
		return textutils.EmptyStr, fmt.Errorf("error decoding token response: %w", err)
	}
	if o.tokenData == nil {
		return textutils.EmptyStr, fmt.Errorf("token response body is nil")
	}
	accessToken, ok := o.tokenData[ACCESS_TOKEN]
	if !ok {
		return textutils.EmptyStr, fmt.Errorf("access_token not found in token response")
	}
	tokenStr, ok := accessToken.(string)
	if !ok {
		return textutils.EmptyStr, fmt.Errorf("access_token is not a string, got %T", accessToken)
	}
	// Calculate and store expiry epoch
	if expiresIn, ok := o.tokenData[EXPIRES_IN]; ok {
		if expiresInSec, err := toFloat64(expiresIn); err == nil {
			// Subtract 100ms buffer so we refresh slightly before actual expiry
			o.tokenData[EXPIRY_EPOCH] = (time.Now().UnixMilli() + int64(expiresInSec)*1000) - 100
		}
	}
	return tokenStr, nil
}

// getCachedToken returns the cached access token if it exists and has not expired.
func (o *oAuth2Provider) getCachedToken() string {
	if o.tokenData == nil {
		return textutils.EmptyStr
	}
	expiry, ok := o.tokenData[EXPIRY_EPOCH]
	if !ok {
		return textutils.EmptyStr
	}
	expiryEpoch, ok := expiry.(int64)
	if !ok {
		return textutils.EmptyStr
	}
	if expiryEpoch <= time.Now().UnixMilli() {
		return textutils.EmptyStr
	}
	accessToken, ok := o.tokenData[ACCESS_TOKEN]
	if !ok {
		return textutils.EmptyStr
	}
	tokenStr, ok := accessToken.(string)
	if !ok {
		return textutils.EmptyStr
	}
	return tokenStr
}

// toFloat64 safely converts a numeric value to float64.
// JSON unmarshalling may produce float64 or other numeric types.
func toFloat64(v any) (float64, error) {
	switch n := v.(type) {
	case float64:
		return n, nil
	case float32:
		return float64(n), nil
	case int:
		return float64(n), nil
	case int64:
		return float64(n), nil
	case int32:
		return float64(n), nil
	default:
		return 0, fmt.Errorf("unsupported type for expires_in: %T", v)
	}
}
