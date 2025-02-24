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
// token, and to refresh the authentication token.
//
// Methods:
//   - Type() AuthType: Returns the type of authentication.
//   - User() string: Returns the username.
//   - Pass() string: Returns the password.
//   - Token() string: Returns the authentication token.
//   - Refresh() error: Refreshes the authentication token and returns an error if the operation fails.
type AuthProvider interface {
	Type() AuthType
	User() string
	Pass() string
	Token() string
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
func (b *BasicAuth) User() string {
	return b.user
}

// Pass returns the password associated with the BasicAuth instance.
// Returns:
//
//	string: The password.
func (b *BasicAuth) Pass() string {
	return b.pass
}

// Token returns an empty string as the token.
// This method is part of the BasicAuth struct.
// Returns:
//
//	string: An empty string.
func (b *BasicAuth) Token() string {
	return textutils.EmptyStr
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
// This method is part of the BearerAuth struct.
// Returns:
//
//	string: An empty string.
func (b *TokenBearerAuth) User() string {
	return textutils.EmptyStr
}

// Pass returns an empty string as the password.
// This method is part of the BearerAuth struct.
// Returns:
//
//	string: An empty string.
func (b *TokenBearerAuth) Pass() string {
	return textutils.EmptyStr
}

// Token returns the token associated with the BearerAuth instance.
// Returns:
//
//	string: The token.
func (b *TokenBearerAuth) Token() string {
	return b.token
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
