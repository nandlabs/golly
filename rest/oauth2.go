package rest

import (
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
	return &oAuth2Provider{
		clientId:      clientId,
		clientSecret:  clientSecret,
		grantType:     grantType,
		tokenEndpoint: tokenEndpoint,
		extraParams:   make(map[string]any),
		tokenData:     make(map[string]any),
		client:        NewClient(),
		lock:          &sync.Mutex{},
	}
}

// Type returns the OAuth2 provider's authentication type.
func (o *oAuth2Provider) Type() clients.AuthType {
	return clients.AuthTypeBearer
}

// User returns the OAuth2 client ID which represents the user identifier for the provider.
// This method satisfies the Provider interface by providing access to the client identifier.
func (o *oAuth2Provider) User() string {
	return o.clientId
}

// Pass returns the OAuth2Provider's client secret.
// This method is used to access the client secret in a controlled manner.
func (o *oAuth2Provider) Pass() string {
	return o.clientSecret
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
//   - An empty string if any error occurs during the token acquisition process
//
// Thread safety:
//
//	Uses mutex locking to ensure concurrent calls don't interfere with token refresh
func (o *oAuth2Provider) Token() string {
	if expiry, ok := o.tokenData[EXPIRY_EPOCH]; ok && expiry.(int64) > time.Now().UnixMilli() {

		if o.tokenData != nil {
			access_token, ok := o.tokenData["access_token"]
			if ok {
				return access_token.(string)
			}
		}
	}
	o.lock.Lock()
	defer o.lock.Unlock()
	request, err := o.client.NewRequest(o.tokenEndpoint, http.MethodPost)
	if err != nil {
		logger.Error("Error creating request: %v", err)
		return ""
	}
	request.SetContentType("application/x-www-form-urlencoded")
	request.AddFormData(GRANT_TYPE, o.grantType)
	request.AddFormData(CLIENT_ID, o.clientId)
	request.AddFormData(CLIENT_SECRET, o.clientSecret)
	if o.extraParams != nil {
		for k, v := range o.extraParams {
			request.AddFormData(k, v.(string))
		}
	}
	response, err := o.client.Execute(request)
	if err != nil {
		logger.Error("Error executing request: %v", err)
		return textutils.EmptyStr
	}
	if response.StatusCode() != http.StatusOK {
		logger.Error("Error: %v", response.StatusCode())
		return textutils.EmptyStr
	}
	if err := response.Decode(&o.tokenData); err != nil {
		logger.Error("Error decoding response: %v", err)
		return textutils.EmptyStr
	}
	if o.tokenData != nil {
		access_token, ok := o.tokenData[ACCESS_TOKEN]
		if ok {
			o.tokenData[EXPIRY_EPOCH] = (time.Now().UnixMilli() + int64(o.tokenData[EXPIRES_IN].(float64))*1000) - 100
			return access_token.(string)
		} else {
			logger.Error("Error: %v", response.StatusCode())
			return textutils.EmptyStr
		}
	}
	return textutils.EmptyStr

}
