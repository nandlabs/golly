package secrets

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"time"
)

// CredentialType represents the type of credential stored
type CredentialType string

const (
	CredentialTypeGeneric     CredentialType = "generic"
	CredentialTypeAPIKey      CredentialType = "api-key"
	CredentialTypePassword    CredentialType = "password"
	CredentialTypeCertificate CredentialType = "certificate"
	CredentialTypeToken       CredentialType = "token"
)

// APIKeyCredential represents API key credentials with provider information
type APIKeyCredential struct {
	Provider    string   // e.g., "AWS", "Google", "GitHub"
	KeyID       string   // The public key identifier
	KeySecret   []byte   // The secret key (encrypted in storage)
	Permissions []string // Scopes or permissions
	CreatedAt   time.Time
	ExpiresAt   time.Time
}

// ToCredential converts APIKeyCredential to generic Credential
func (a *APIKeyCredential) ToCredential(version string) *Credential {
	cred := &Credential{
		Value:       a.KeySecret,
		LastUpdated: time.Now(),
		Version:     version,
		MetaData: map[string]interface{}{
			"type":        string(CredentialTypeAPIKey),
			"provider":    a.Provider,
			"key_id":      a.KeyID,
			"permissions": a.Permissions,
			"created_at":  a.CreatedAt,
			"expires_at":  a.ExpiresAt,
		},
	}
	return cred
}

// FromCredential extracts APIKeyCredential from generic Credential
func (a *APIKeyCredential) FromCredential(cred *Credential) error {
	if cred.MetaData == nil {
		return fmt.Errorf("credential metadata is required for APIKeyCredential")
	}

	credType, ok := cred.MetaData["type"].(string)
	if !ok || credType != string(CredentialTypeAPIKey) {
		return fmt.Errorf("credential type mismatch: expected api-key, got %v", credType)
	}

	a.KeySecret = cred.Value

	if v, ok := cred.MetaData["provider"].(string); ok {
		a.Provider = v
	}
	if v, ok := cred.MetaData["key_id"].(string); ok {
		a.KeyID = v
	}
	if v, ok := cred.MetaData["permissions"].([]string); ok {
		a.Permissions = v
	}

	return nil
}

// IsExpired checks if the API key has expired
func (a *APIKeyCredential) IsExpired() bool {
	if a.ExpiresAt.IsZero() {
		return false // No expiration
	}
	return time.Now().After(a.ExpiresAt)
}

// PasswordCredential represents username/password credentials
type PasswordCredential struct {
	Username  string
	Password  []byte // Should be encrypted in storage
	ExpiresAt time.Time
	CreatedAt time.Time
}

// ToCredential converts PasswordCredential to generic Credential
func (p *PasswordCredential) ToCredential(version string) *Credential {
	cred := &Credential{
		Value:       p.Password,
		LastUpdated: time.Now(),
		Version:     version,
		MetaData: map[string]interface{}{
			"type":       string(CredentialTypePassword),
			"username":   p.Username,
			"expires_at": p.ExpiresAt,
			"created_at": p.CreatedAt,
		},
	}
	return cred
}

// FromCredential extracts PasswordCredential from generic Credential
func (p *PasswordCredential) FromCredential(cred *Credential) error {
	if cred.MetaData == nil {
		return fmt.Errorf("credential metadata is required for PasswordCredential")
	}

	credType, ok := cred.MetaData["type"].(string)
	if !ok || credType != string(CredentialTypePassword) {
		return fmt.Errorf("credential type mismatch: expected password, got %v", credType)
	}

	p.Password = cred.Value

	if v, ok := cred.MetaData["username"].(string); ok {
		p.Username = v
	}
	if v, ok := cred.MetaData["expires_at"].(time.Time); ok {
		p.ExpiresAt = v
	}
	if v, ok := cred.MetaData["created_at"].(time.Time); ok {
		p.CreatedAt = v
	}

	return nil
}

// IsExpired checks if the password has expired
func (p *PasswordCredential) IsExpired() bool {
	if p.ExpiresAt.IsZero() {
		return false // No expiration
	}
	return time.Now().After(p.ExpiresAt)
}

// CertificateCredential represents X.509 certificate credentials
type CertificateCredential struct {
	Certificate *x509.Certificate // Parsed certificate
	PrivateKey  []byte            // PEM-encoded private key (encrypted in storage)
	CertChain   [][]byte          // Additional certificates in chain
	CreatedAt   time.Time
}

// ToCredential converts CertificateCredential to generic Credential
func (c *CertificateCredential) ToCredential(version string) *Credential {
	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: c.Certificate.Raw,
	})

	cred := &Credential{
		Value:       c.PrivateKey,
		LastUpdated: time.Now(),
		Version:     version,
		MetaData: map[string]interface{}{
			"type":            string(CredentialTypeCertificate),
			"certificate_pem": string(certPEM),
			"subject":         c.Certificate.Subject.String(),
			"issuer":          c.Certificate.Issuer.String(),
			"not_before":      c.Certificate.NotBefore,
			"not_after":       c.Certificate.NotAfter,
			"serial_number":   c.Certificate.SerialNumber.String(),
			"cert_chain":      c.CertChain,
			"created_at":      c.CreatedAt,
		},
	}
	return cred
}

// FromCredential extracts CertificateCredential from generic Credential
func (c *CertificateCredential) FromCredential(cred *Credential) error {
	if cred.MetaData == nil {
		return fmt.Errorf("credential metadata is required for CertificateCredential")
	}

	credType, ok := cred.MetaData["type"].(string)
	if !ok || credType != string(CredentialTypeCertificate) {
		return fmt.Errorf("credential type mismatch: expected certificate, got %v", credType)
	}

	c.PrivateKey = cred.Value

	// Parse certificate from metadata
	if certPEM, ok := cred.MetaData["certificate_pem"].(string); ok {
		block, _ := pem.Decode([]byte(certPEM))
		if block != nil {
			cert, err := x509.ParseCertificate(block.Bytes)
			if err != nil {
				return fmt.Errorf("failed to parse certificate: %w", err)
			}
			c.Certificate = cert
		}
	}

	if v, ok := cred.MetaData["cert_chain"].([][]byte); ok {
		c.CertChain = v
	}

	if v, ok := cred.MetaData["created_at"].(time.Time); ok {
		c.CreatedAt = v
	}

	return nil
}

// IsExpired checks if the certificate has expired
func (c *CertificateCredential) IsExpired() bool {
	if c.Certificate == nil {
		return false
	}
	return time.Now().After(c.Certificate.NotAfter)
}

// TokenCredential represents bearer tokens and session tokens
type TokenCredential struct {
	Token     []byte   // The token value (encrypted in storage)
	TokenType string   // e.g., "Bearer", "Basic", "OAuth2"
	Scopes    []string // OAuth2 scopes or similar permissions
	CreatedAt time.Time
	ExpiresAt time.Time
}

// ToCredential converts TokenCredential to generic Credential
func (t *TokenCredential) ToCredential(version string) *Credential {
	cred := &Credential{
		Value:       t.Token,
		LastUpdated: time.Now(),
		Version:     version,
		MetaData: map[string]interface{}{
			"type":       string(CredentialTypeToken),
			"token_type": t.TokenType,
			"scopes":     t.Scopes,
			"expires_at": t.ExpiresAt,
			"created_at": t.CreatedAt,
		},
	}
	return cred
}

// FromCredential extracts TokenCredential from generic Credential
func (t *TokenCredential) FromCredential(cred *Credential) error {
	if cred.MetaData == nil {
		return fmt.Errorf("credential metadata is required for TokenCredential")
	}

	credType, ok := cred.MetaData["type"].(string)
	if !ok || credType != string(CredentialTypeToken) {
		return fmt.Errorf("credential type mismatch: expected token, got %v", credType)
	}

	t.Token = cred.Value

	if v, ok := cred.MetaData["token_type"].(string); ok {
		t.TokenType = v
	}
	if v, ok := cred.MetaData["scopes"].([]string); ok {
		t.Scopes = v
	}
	if v, ok := cred.MetaData["expires_at"].(time.Time); ok {
		t.ExpiresAt = v
	}
	if v, ok := cred.MetaData["created_at"].(time.Time); ok {
		t.CreatedAt = v
	}

	return nil
}

// IsExpired checks if the token has expired
func (t *TokenCredential) IsExpired() bool {
	if t.ExpiresAt.IsZero() {
		return false // No expiration
	}
	return time.Now().After(t.ExpiresAt)
}

// CredentialTypeFromString returns CredentialType from string
func CredentialTypeFromString(s string) CredentialType {
	switch s {
	case string(CredentialTypeAPIKey):
		return CredentialTypeAPIKey
	case string(CredentialTypePassword):
		return CredentialTypePassword
	case string(CredentialTypeCertificate):
		return CredentialTypeCertificate
	case string(CredentialTypeToken):
		return CredentialTypeToken
	default:
		return CredentialTypeGeneric
	}
}

// DetectCredentialType detects credential type from metadata
func DetectCredentialType(metadata map[string]interface{}) CredentialType {
	if metadata == nil {
		return CredentialTypeGeneric
	}

	if credType, ok := metadata["type"].(string); ok {
		return CredentialTypeFromString(credType)
	}

	return CredentialTypeGeneric
}
