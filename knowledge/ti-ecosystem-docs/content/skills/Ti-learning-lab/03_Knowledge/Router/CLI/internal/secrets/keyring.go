package secrets

import (
	"fmt"
	"github.com/99designs/keyring"
)

const serviceName = "ti-backend-ui-synced"

// SecretsManager provides an interface for secure storage
type SecretsManager interface {
	Set(key, value string) error
	Get(key string) (string, error)
	Delete(key string) error
}

type osKeyring struct {
	ring keyring.Keyring
}

// NewSecretsManager initializes the OS keyring integration
func NewSecretsManager() (SecretsManager, error) {
	ring, err := keyring.Open(keyring.Config{
		ServiceName: serviceName,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open keyring: %w", err)
	}

	return &osKeyring{ring: ring}, nil
}

func (k *osKeyring) Set(key, value string) error {
	item := keyring.Item{
		Key:  key,
		Data: []byte(value),
	}
	return k.ring.Set(item)
}

func (k *osKeyring) Get(key string) (string, error) {
	item, err := k.ring.Get(key)
	if err != nil {
		return "", err
	}
	return string(item.Data), nil
}

func (k *osKeyring) Delete(key string) error {
	return k.ring.Remove(key)
}
