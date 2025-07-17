package config

import (
	"github.com/delordemm1/qplayground/internal/platform"
	"github.com/delordemm1/qplayground/pkg/inertia"

	"github.com/alexedwards/scs/v2"
	inertiaLib "github.com/romsar/gonertia/v2"
)

// InertiaConfig holds Inertia configuration
type InertiaConfig struct {
	RootViewFile string
	SSREnabled   bool
	ViteHotFile  string
}

// DefaultInertiaConfig returns default Inertia configuration
func DefaultInertiaConfig() InertiaConfig {
	return InertiaConfig{
		RootViewFile: "resources/views/root.html",
		SSREnabled:   true,
		ViteHotFile:  "./public/hot",
	}
}

// InitInertia initializes and returns an Inertia instance
func InitInertia(config InertiaConfig, sessionManager *scs.SessionManager) *inertiaLib.Inertia {
	flashProvider := platform.NewSCSFlashProvider(sessionManager)
	return inertia.Initialize(config.RootViewFile, flashProvider, config.SSREnabled, config.ViteHotFile)
}
