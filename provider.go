package cache

import (
	"fmt"

	"github.com/donnigundala/dg-core/contracts/foundation"
)

// CacheServiceProvider implements the PluginProvider interface.
// This provides a simple, plug-and-play integration for applications.
//
// The provider expects the application to register drivers via DriverFactories.
// For automatic driver registration, applications should use a wrapper provider.
//
// For advanced use cases requiring custom drivers or configuration,
// use the library functions (NewManager, RegisterDriver) directly.
type CacheServiceProvider struct {
	// Config holds cache configuration
	// Auto-injected by dg-core if using config:"cache" tag
	Config Config `config:"cache"`

	// DriverFactories maps driver names to their factory functions
	// If nil, drivers must be registered manually after registration
	DriverFactories map[string]DriverFactory
}

// Name returns the name of the plugin.
func (p *CacheServiceProvider) Name() string {
	return "cache"
}

// Version returns the version of the plugin.
func (p *CacheServiceProvider) Version() string {
	return "1.4.0"
}

// Dependencies returns the list of dependencies.
func (p *CacheServiceProvider) Dependencies() []string {
	return []string{}
}

// Register registers the cache service provider.
func (p *CacheServiceProvider) Register(app foundation.Application) error {
	// Use provided config or default
	cfg := p.Config
	if cfg.DefaultStore == "" {
		cfg = DefaultConfig()
	}

	// Register the cache manager as a singleton
	app.Singleton("cache", func() (interface{}, error) {
		manager, err := NewManager(cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to create cache manager: %w", err)
		}

		// Register driver factories if provided
		if p.DriverFactories != nil {
			for name, factory := range p.DriverFactories {
				manager.RegisterDriver(name, factory)
			}
		}

		return manager, nil
	})

	return nil
}

// Boot boots the cache service provider.
func (p *CacheServiceProvider) Boot(app foundation.Application) error {
	// Verify cache manager can be resolved
	_, err := app.Make("cache")
	if err != nil {
		return fmt.Errorf("failed to boot cache provider: %w", err)
	}
	return nil
}

// Shutdown gracefully closes cache connections.
func (p *CacheServiceProvider) Shutdown(app foundation.Application) error {
	cacheInstance, err := app.Make("cache")
	if err != nil {
		return nil // Cache not initialized
	}

	manager := cacheInstance.(*Manager)
	return manager.Close()
}
