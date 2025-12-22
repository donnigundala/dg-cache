package dgcache

import (
	"fmt"

	"github.com/donnigundala/dg-core/contracts/foundation"
)

// CacheServiceProvider implements the PluginProvider interface.
type CacheServiceProvider struct {
	// Config holds cache configuration
	Config Config `config:"cache"`

	// DriverFactories maps driver names to their factory functions
	DriverFactories map[string]DriverFactory
}

// NewCacheServiceProvider creates a new cache service provider.
func NewCacheServiceProvider(driverFactories map[string]DriverFactory) *CacheServiceProvider {
	return &CacheServiceProvider{
		DriverFactories: driverFactories,
	}
}

// Name returns the name of the plugin.
func (p *CacheServiceProvider) Name() string {
	return Binding
}

// Version returns the version of the plugin.
func (p *CacheServiceProvider) Version() string {
	return Version
}

// Dependencies returns the list of dependencies.
func (p *CacheServiceProvider) Dependencies() []string {
	return []string{}
}

// Register registers the cache service provider.
func (p *CacheServiceProvider) Register(app foundation.Application) error {
	app.Singleton(Binding, func() (interface{}, error) {
		// Use provided config or default
		cfg := p.Config
		if cfg.DefaultStore == "" {
			cfg = DefaultConfig()
		}

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
	// Resolve the manager to trigger its creation and registration of drivers
	cacheInstance, err := app.Make(Binding)
	if err != nil {
		return fmt.Errorf("failed to resolve cache manager during boot: %w", err)
	}

	manager := cacheInstance.(*Manager)

	// Ensure the default store is initialized so metrics have something to observe
	_, _ = manager.Store("")

	// Auto-register named stores in container
	for storeName := range p.Config.Stores {
		captuerdName := storeName // capture for closure
		app.Singleton(fmt.Sprintf("%s.%s", Binding, captuerdName), func() (interface{}, error) {
			store, err := manager.Store(captuerdName)
			if err != nil {
				return nil, fmt.Errorf("failed to get store %s: %w", captuerdName, err)
			}
			return store, nil
		})
	}

	// Register metrics
	if err := manager.RegisterMetrics(); err != nil {
		// Log error but don't fail boot
		if log, err := app.Make("logger"); err == nil {
			if l, ok := log.(interface {
				Warn(msg string, args ...interface{})
			}); ok {
				l.Warn("Failed to register cache metrics", "error", err)
			}
		}
	}

	return nil
}

// Shutdown gracefully closes cache connections.
func (p *CacheServiceProvider) Shutdown(app foundation.Application) error {
	cacheInstance, err := app.Make(Binding)
	if err != nil {
		return nil // Cache not initialized
	}

	manager := cacheInstance.(*Manager)
	return manager.Close()
}
