package secrets

import (
	"time"
)

const (
	CredType   = "CredType"
	TextSecret = "TextSecret"
)

//Credential Struct that can represent any secret value.
//Fields
//
// Value is an array of bytes to hold decrypted value of a secret.

// LastUpdated specifies the last updated time of the credential.
//
// Version string specifies the current version of the credential. If the credential type/Store implementation
// maintains multiple version the previous version info can be maintained in metadata field.
//
// MetaData Any additional key,value attributes that needs to be associated with the credentials can be done using this
//property

type Credential struct {
	Value       []byte
	LastUpdated time.Time
	Version     string
	MetaData    map[string]interface{}
}

// Str function gets the Credential.Value field as string
func (c *Credential) Str() (s string) {
	if c.Value != nil {
		s = string(c.Value)
	}
	return
}

//Type Returns the type of the credential
func (c *Credential) Type() (s string) {
	s = TextSecret
	if c.MetaData != nil {
		if v, ok := c.MetaData[CredType]; ok {
			s = v.(string)
		}
	}
	return
}
