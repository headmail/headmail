package db

import (
	"fmt"

	"github.com/headmail/headmail/pkg/config"
	"github.com/headmail/headmail/pkg/repository"
)

// Provider defines the interface for a database provider.
type Provider interface {
	// New creates a new database connection that satisfies the repository.DB interface.
	New(cfg config.DatabaseConfig) (repository.DB, error)
}

type providerHolder struct {
	p         Provider
	isDefault bool
}

var providers = make(map[string]*providerHolder)

// RegisterProvider registers a new database provider.
func RegisterProvider(name string, provider Provider) {
	old := providers[name]
	if old != nil && !old.isDefault {
		panic(fmt.Sprintf("provider with name '%s' already registered", name))
	}
	providers[name] = &providerHolder{
		p:         provider,
		isDefault: false,
	}
}

func RegisterDefaultProvider(name string, provider Provider) {
	_, exists := providers[name]
	if !exists {
		providers[name] = &providerHolder{
			p:         provider,
			isDefault: true,
		}
	}
}

// GetProvider retrieves a registered database provider by name.
func GetProvider(name string) (Provider, error) {
	provider, exists := providers[name]
	if !exists {
		return nil, fmt.Errorf("no provider registered for '%s'", name)
	}
	return provider.p, nil
}
