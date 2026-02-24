package clients

import "oss.nandlabs.io/golly/textutils"

const (
	AuthTypeBasic  AuthType = "Basic"
	AuthTypeBearer AuthType = "Token"
)

// Authenticatable represents an interface that requires the implementation
// of the Apply method. The Apply method takes an Authenticator as a parameter
// and applies it to the implementing type.
type Authenticatable interface {
	Apply(AuthProvider)
}

type AuthType string

// AuthProvider defines an interface for authentication mechanisms.
// It provides methods to retrieve the type of authentication, user credentials,
// and token. All credential-fetching methods return errors to support
// implementations backed by external stores (e.g., Vault, AWS Secrets Manager).
//
// Methods:
//   - Type() AuthType: Returns the type of authentication.
//   - User() (string, error): Returns the username or an error if retrieval fails.
//   - Pass() (string, error): Returns the password or an error if retrieval fails.
//   - Token() (string, error): Returns the authentication token or an error if token acquisition fails.
type AuthProvider interface {
	Type() AuthType
	User() (string, error)
	Pass() (string, error)
	Token() (string, error)
}

// BasicAuth represents basic authentication credentials with a username and password.
type BasicAuth struct {
	user string
	pass string
}

// Type returns the authentication type for BasicAuth, which is AuthTypeBasic.
// This is used to determine the type of authentication to be used.
// Returns:
//
//	AuthType: The authentication type. (AuthTypeBasic)
func (b *BasicAuth) Type() AuthType {
	return AuthTypeBasic
}

// User returns the username associated with the BasicAuth instance.
// Returns:
//
//	string: The username.
//	error: Always nil for in-memory credentials.
func (b *BasicAuth) User() (string, error) {
	return b.user, nil
}

// Pass returns the password associated with the BasicAuth instance.
// Returns:
//
//	string: The password.
//	error: Always nil for in-memory credentials.
func (b *BasicAuth) Pass() (string, error) {
	return b.pass, nil
}

// Token returns an empty string as the token.
// BasicAuth does not use tokens.
// Returns:
//
//	string: An empty string.
//	error: Always nil.
func (b *BasicAuth) Token() (string, error) {
	return textutils.EmptyStr, nil
}

// Refresh refreshes the authentication credentials.
// It currently does not perform any operations and always returns nil.
//
// Returns:
//
//	error: Always returns nil.
func (b *BasicAuth) Refresh() error {
	return nil
}

// NewBasicAuth creates a new BasicAuth instance with the provided username and password.
// Parameters:
//
//	user (string): The username.
//	pass (string): The password.
//
// Returns:
//
//	*BasicAuth: The BasicAuth instance.
func NewBasicAuth(user, pass string) AuthProvider {
	return &BasicAuth{
		user: user,
		pass: pass,
	}
}

// TokenBearerAuth represents bearer token authentication credentials.
type TokenBearerAuth struct {
	token string
}

// Type returns the authentication type for BearerAuth, which is AuthTypeBearer.
// This is used to determine the type of authentication to be used.
// Returns:
//
//	AuthType: The authentication type. (AuthTypeBearer)
func (b *TokenBearerAuth) Type() AuthType {
	return AuthTypeBearer
}

// User returns an empty string as the username.
// TokenBearerAuth does not use username credentials.
// Returns:
//
//	string: An empty string.
//	error: Always nil.
func (b *TokenBearerAuth) User() (string, error) {
	return textutils.EmptyStr, nil
}

// Pass returns an empty string as the password.
// TokenBearerAuth does not use password credentials.
// Returns:
//
//	string: An empty string.
//	error: Always nil.
func (b *TokenBearerAuth) Pass() (string, error) {
	return textutils.EmptyStr, nil
}

// Token returns the token associated with the BearerAuth instance.
// Returns:
//
//	string: The token.
//	error: Always nil for static bearer tokens.
func (b *TokenBearerAuth) Token() (string, error) {
	return b.token, nil
}

// Refresh refreshes the authentication token.
// It currently does not perform any operations and always returns nil.
//
// Returns:
//
//	error: Always returns nil.
func (b *TokenBearerAuth) Refresh() error {
	return nil
}

// NewBearerAuth creates a new BearerAuth instance with the provided token.
// Parameters:
//
//	token (string): The token.
//
// Returns:
//
//	*BearerAuth: The BearerAuth instance.
func NewBearerAuth(token string) AuthProvider {
	return &TokenBearerAuth{
		token: token,
	}
}
