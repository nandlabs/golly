package secrets

import (
	"context"
)

type Store interface {
	Get(key string, ctx context.Context) (*Credential, error)
	Write(key string, credential *Credential, ctx context.Context) error
	Provider() string
}
