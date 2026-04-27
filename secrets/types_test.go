package secrets

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"testing"
	"time"
)

// Helper function to create a self-signed certificate for testing
func createTestCertificate(t *testing.T) (*x509.Certificate, []byte) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate RSA key: %v", err)
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Country:      []string{"US"},
			Organization: []string{"Test Org"},
			CommonName:   "test.example.com",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(1, 0, 0),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		t.Fatalf("Failed to create certificate: %v", err)
	}

	cert, err := x509.ParseCertificate(certBytes)
	if err != nil {
		t.Fatalf("Failed to parse certificate: %v", err)
	}

	return cert, certBytes
}

func TestAPIKeyCredential(t *testing.T) {
	t.Run("ToCredential", func(t *testing.T) {
		apiKey := &APIKeyCredential{
			Provider:    "AWS",
			KeyID:       "AKIA1234567890AB",
			KeySecret:   []byte("secret-key-value"),
			Permissions: []string{"s3:*", "ec2:*"},
			CreatedAt:   time.Now(),
			ExpiresAt:   time.Now().AddDate(0, 0, 30),
		}

		cred := apiKey.ToCredential("1.0")

		if cred.Version != "1.0" {
			t.Errorf("Version mismatch: got %q, want %q", cred.Version, "1.0")
		}

		if string(cred.Value) != "secret-key-value" {
			t.Errorf("Value mismatch: got %q, want %q", string(cred.Value), "secret-key-value")
		}

		if credType, ok := cred.MetaData["type"].(string); !ok || credType != string(CredentialTypeAPIKey) {
			t.Errorf("Type mismatch in metadata")
		}

		if provider, ok := cred.MetaData["provider"].(string); !ok || provider != "AWS" {
			t.Errorf("Provider mismatch in metadata")
		}
	})

	t.Run("FromCredential", func(t *testing.T) {
		originalKey := &APIKeyCredential{
			Provider:    "Google",
			KeyID:       "GOOG1234567890AB",
			KeySecret:   []byte("google-secret"),
			Permissions: []string{"storage:read"},
			CreatedAt:   time.Now(),
			ExpiresAt:   time.Now().AddDate(0, 0, 60),
		}

		cred := originalKey.ToCredential("1.0")

		recovered := &APIKeyCredential{}
		err := recovered.FromCredential(cred)
		if err != nil {
			t.Fatalf("Failed to extract APIKeyCredential: %v", err)
		}

		if recovered.Provider != originalKey.Provider {
			t.Errorf("Provider mismatch: got %q, want %q", recovered.Provider, originalKey.Provider)
		}

		if string(recovered.KeySecret) != string(originalKey.KeySecret) {
			t.Errorf("KeySecret mismatch")
		}
	})

	t.Run("IsExpired", func(t *testing.T) {
		expiredKey := &APIKeyCredential{
			Provider:  "AWS",
			ExpiresAt: time.Now().AddDate(0, 0, -1), // Expired yesterday
		}

		if !expiredKey.IsExpired() {
			t.Error("Expected key to be expired")
		}

		futureKey := &APIKeyCredential{
			Provider:  "AWS",
			ExpiresAt: time.Now().AddDate(0, 0, 30), // Expires in 30 days
		}

		if futureKey.IsExpired() {
			t.Error("Expected key to not be expired")
		}

		noExpiryKey := &APIKeyCredential{
			Provider: "AWS",
			// No expiration
		}

		if noExpiryKey.IsExpired() {
			t.Error("Expected key with no expiration to not be expired")
		}
	})
}

func TestPasswordCredential(t *testing.T) {
	t.Run("ToCredential", func(t *testing.T) {
		passCred := &PasswordCredential{
			Username:  "testuser",
			Password:  []byte("super-secret-password"),
			CreatedAt: time.Now(),
			ExpiresAt: time.Now().AddDate(0, 0, 90),
		}

		cred := passCred.ToCredential("2.0")

		if cred.Version != "2.0" {
			t.Errorf("Version mismatch: got %q, want %q", cred.Version, "2.0")
		}

		if username, ok := cred.MetaData["username"].(string); !ok || username != "testuser" {
			t.Errorf("Username mismatch in metadata")
		}
	})

	t.Run("FromCredential", func(t *testing.T) {
		original := &PasswordCredential{
			Username:  "alice",
			Password:  []byte("secret123"),
			CreatedAt: time.Now(),
		}

		cred := original.ToCredential("1.0")

		recovered := &PasswordCredential{}
		err := recovered.FromCredential(cred)
		if err != nil {
			t.Fatalf("Failed to extract PasswordCredential: %v", err)
		}

		if recovered.Username != original.Username {
			t.Errorf("Username mismatch: got %q, want %q", recovered.Username, original.Username)
		}

		if string(recovered.Password) != string(original.Password) {
			t.Errorf("Password mismatch")
		}
	})

	t.Run("IsExpired", func(t *testing.T) {
		expiredPass := &PasswordCredential{
			Username:  "user",
			ExpiresAt: time.Now().AddDate(-1, 0, 0), // Expired a year ago
		}

		if !expiredPass.IsExpired() {
			t.Error("Expected password to be expired")
		}

		validPass := &PasswordCredential{
			Username:  "user",
			ExpiresAt: time.Now().AddDate(1, 0, 0), // Expires in a year
		}

		if validPass.IsExpired() {
			t.Error("Expected password to not be expired")
		}
	})
}

func TestCertificateCredential(t *testing.T) {
	t.Run("ToCredential", func(t *testing.T) {
		cert, _ := createTestCertificate(t)
		privateKeyPEM := []byte("-----BEGIN RSA PRIVATE KEY-----\nMIIEpAIBAAKCAQEA...\n-----END RSA PRIVATE KEY-----")

		certCred := &CertificateCredential{
			Certificate: cert,
			PrivateKey:  privateKeyPEM,
			CreatedAt:   time.Now(),
		}

		cred := certCred.ToCredential("1.0")

		if credType, ok := cred.MetaData["type"].(string); !ok || credType != string(CredentialTypeCertificate) {
			t.Errorf("Type mismatch in metadata")
		}

		if subject, ok := cred.MetaData["subject"].(string); !ok || subject == "" {
			t.Errorf("Subject missing in metadata")
		}

		if notAfter, ok := cred.MetaData["not_after"].(time.Time); !ok || notAfter.IsZero() {
			t.Errorf("NotAfter missing or invalid in metadata")
		}
	})

	t.Run("IsExpired", func(t *testing.T) {
		cert, _ := createTestCertificate(t)

		validCert := &CertificateCredential{
			Certificate: cert,
		}

		if validCert.IsExpired() {
			t.Error("Expected certificate to not be expired")
		}

		// Manually create a certificate struct with an expiration in the past
		expiredCert := &x509.Certificate{
			NotAfter: time.Now().AddDate(0, 0, -1), // Expired 1 day ago
		}

		expiredCred := &CertificateCredential{
			Certificate: expiredCert,
		}

		if !expiredCred.IsExpired() {
			t.Error("Expected certificate to be expired")
		}

		// Test certificate with no certificate
		noCert := &CertificateCredential{
			Certificate: nil,
		}

		if noCert.IsExpired() {
			t.Error("Expected certificate with nil to not be expired")
		}
	})
}

func TestTokenCredential(t *testing.T) {
	t.Run("ToCredential", func(t *testing.T) {
		tokenCred := &TokenCredential{
			Token:     []byte("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."),
			TokenType: "Bearer",
			Scopes:    []string{"read:user", "write:repo"},
			CreatedAt: time.Now(),
			ExpiresAt: time.Now().AddDate(0, 0, 7),
		}

		cred := tokenCred.ToCredential("1.0")

		if credType, ok := cred.MetaData["type"].(string); !ok || credType != string(CredentialTypeToken) {
			t.Errorf("Type mismatch in metadata")
		}

		if tokenType, ok := cred.MetaData["token_type"].(string); !ok || tokenType != "Bearer" {
			t.Errorf("TokenType mismatch in metadata")
		}
	})

	t.Run("FromCredential", func(t *testing.T) {
		original := &TokenCredential{
			Token:     []byte("jwt-token-here"),
			TokenType: "Bearer",
			Scopes:    []string{"api:read"},
			CreatedAt: time.Now(),
			ExpiresAt: time.Now().AddDate(0, 0, 1),
		}

		cred := original.ToCredential("1.0")

		recovered := &TokenCredential{}
		err := recovered.FromCredential(cred)
		if err != nil {
			t.Fatalf("Failed to extract TokenCredential: %v", err)
		}

		if string(recovered.Token) != string(original.Token) {
			t.Errorf("Token mismatch")
		}

		if recovered.TokenType != original.TokenType {
			t.Errorf("TokenType mismatch: got %q, want %q", recovered.TokenType, original.TokenType)
		}
	})

	t.Run("IsExpired", func(t *testing.T) {
		expiredToken := &TokenCredential{
			Token:     []byte("old-token"),
			TokenType: "Bearer",
			ExpiresAt: time.Now().AddDate(0, 0, -1), // Expired yesterday
		}

		if !expiredToken.IsExpired() {
			t.Error("Expected token to be expired")
		}

		validToken := &TokenCredential{
			Token:     []byte("new-token"),
			TokenType: "Bearer",
			ExpiresAt: time.Now().AddDate(0, 0, 30), // Expires in 30 days
		}

		if validToken.IsExpired() {
			t.Error("Expected token to not be expired")
		}

		neverExpiresToken := &TokenCredential{
			Token:     []byte("eternal-token"),
			TokenType: "Bearer",
			// No expiration
		}

		if neverExpiresToken.IsExpired() {
			t.Error("Expected token with no expiration to not be expired")
		}
	})
}

func TestCredentialTypeDetection(t *testing.T) {
	t.Run("DetectAPIKeyType", func(t *testing.T) {
		metadata := map[string]interface{}{
			"type":     string(CredentialTypeAPIKey),
			"provider": "AWS",
		}

		detected := DetectCredentialType(metadata)
		if detected != CredentialTypeAPIKey {
			t.Errorf("Expected CredentialTypeAPIKey, got %v", detected)
		}
	})

	t.Run("DetectPasswordType", func(t *testing.T) {
		metadata := map[string]interface{}{
			"type":     string(CredentialTypePassword),
			"username": "testuser",
		}

		detected := DetectCredentialType(metadata)
		if detected != CredentialTypePassword {
			t.Errorf("Expected CredentialTypePassword, got %v", detected)
		}
	})

	t.Run("DetectNilMetadata", func(t *testing.T) {
		detected := DetectCredentialType(nil)
		if detected != CredentialTypeGeneric {
			t.Errorf("Expected CredentialTypeGeneric for nil metadata, got %v", detected)
		}
	})

	t.Run("DetectMissingType", func(t *testing.T) {
		metadata := map[string]interface{}{
			"provider": "Unknown",
		}

		detected := DetectCredentialType(metadata)
		if detected != CredentialTypeGeneric {
			t.Errorf("Expected CredentialTypeGeneric for missing type, got %v", detected)
		}
	})
}

func TestCredentialTypeConversion(t *testing.T) {
	t.Run("InvalidTypeForAPIKey", func(t *testing.T) {
		cred := &Credential{
			Value: []byte("test"),
			MetaData: map[string]interface{}{
				"type": string(CredentialTypePassword),
			},
		}

		apiKey := &APIKeyCredential{}
		err := apiKey.FromCredential(cred)
		if err == nil {
			t.Error("Expected error when converting wrong type to APIKeyCredential")
		}
	})

	t.Run("MissingMetadata", func(t *testing.T) {
		cred := &Credential{
			Value: []byte("test"),
			// No metadata
		}

		apiKey := &APIKeyCredential{}
		err := apiKey.FromCredential(cred)
		if err == nil {
			t.Error("Expected error when converting credential with no metadata")
		}
	})

	t.Run("SuccessfulRoundTrip", func(t *testing.T) {
		original := &TokenCredential{
			Token:     []byte("test-token"),
			TokenType: "Bearer",
			Scopes:    []string{"read:all"},
			ExpiresAt: time.Now().AddDate(0, 0, 7),
		}

		// Convert to generic credential
		cred := original.ToCredential("1.0")

		// Convert back from generic credential
		recovered := &TokenCredential{}
		err := recovered.FromCredential(cred)
		if err != nil {
			t.Fatalf("Failed to recover credential: %v", err)
		}

		// Verify the data is intact
		if string(recovered.Token) != string(original.Token) {
			t.Error("Token value was not preserved in round trip")
		}

		if recovered.TokenType != original.TokenType {
			t.Error("TokenType was not preserved in round trip")
		}
	})
}
