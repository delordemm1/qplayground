package inertia

import (
	"fmt"
	"html/template"
	"log"
	"log/slog"
	"os"
	"strings"

	inertiaLib "github.com/romsar/gonertia/v2"
)

// Initialize sets up and returns an Inertia instance
func Initialize(rootViewFile string, flashProvider inertiaLib.FlashProvider, ssrEnabled bool, viteHotFile string) *inertiaLib.Inertia {
	// Check if Vite is running in dev mode by checking for hot file
	_, err := os.Stat(viteHotFile)
	isDev := err == nil
	
	if isDev {
		return initializeDev(rootViewFile, flashProvider, ssrEnabled, viteHotFile)
	}
	return initializeProd(rootViewFile, flashProvider, ssrEnabled)
}

// initializeDev initializes Inertia for development mode
func initializeDev(rootViewFile string, flashProvider inertiaLib.FlashProvider, ssrEnabled bool, viteHotFile string) *inertiaLib.Inertia {
	slog.Debug("Vite is running in dev mode")
	
	var options []inertiaLib.Option
	if ssrEnabled {
		options = append(options, inertiaLib.WithSSR())
	}
	if flashProvider != nil {
		options = append(options, inertiaLib.WithFlashProvider(flashProvider))
	}

	i, err := inertiaLib.NewFromFile(rootViewFile, options...)
	if err != nil {
		log.Fatal("Failed to initialize Inertia:", err)
	}

	// Setup Vite dev server template functions
	i.ShareTemplateFunc("viteJS", func(entry string) (template.HTML, error) {
		content, err := os.ReadFile(viteHotFile)
		if err != nil {
			return "", err
		}
		url := strings.TrimSpace(string(content))
		if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
			url = "//localhost:5173" // Default Vite dev port
		}
		var htmlLink strings.Builder
		htmlLink.WriteString(fmt.Sprintf(`<script type="module" src="%s"></script>`, url+"/"+entry))
		return template.HTML(htmlLink.String()), nil
	})

	// In dev, we don't need to inject CSS this way, Vite handles it.
	i.ShareTemplateFunc("viteCSS", func(entry string) (template.HTML, error) {
		return "", nil
	})

	return i
}

// initializeProd initializes Inertia for production mode
func initializeProd(rootViewFile string, flashProvider inertiaLib.FlashProvider, ssrEnabled bool) *inertiaLib.Inertia {
	slog.Debug("Vite is in production mode")
	
	manifestPath := "./public/build/manifest.json"
	buildDir := "/build/"

	// Ensure manifest exists
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		// Vite build might place it in .vite/manifest.json
		originalManifestPath := "./public/build/.vite/manifest.json"
		if err := os.Rename(originalManifestPath, manifestPath); err != nil {
			log.Fatalf("failed to rename manifest file: %v", err)
		}
	}

	// Load manifest
	viteAssets, err := loadViteManifest(manifestPath)
	if err != nil {
		log.Fatalf("could not load vite manifest: %v", err)
	}

	var options []inertiaLib.Option
	options = append(options, inertiaLib.WithVersionFromFile(manifestPath)) // A simple way to get a version hash
	if ssrEnabled {
		options = append(options, inertiaLib.WithSSR())
	}
	if flashProvider != nil {
		options = append(options, inertiaLib.WithFlashProvider(flashProvider))
	}

	i, err := inertiaLib.NewFromFile(rootViewFile, options...)
	if err != nil {
		log.Fatal("Failed to initialize Inertia:", err)
	}

	// Share template functions for production
	i.ShareTemplateFunc("viteJS", viteJSProd(viteAssets, buildDir))
	i.ShareTemplateFunc("viteCSS", viteCSSProd(viteAssets, buildDir))

	return i
}