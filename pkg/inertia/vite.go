package inertia

import (
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"path"
	"strings"
)

// ViteAsset represents an entry in the Vite manifest.
type ViteAsset struct {
	File string   `json:"file"`
	Src  string   `json:"src"`
	CSS  []string `json:"css"`
}

// loadViteManifest parses the Vite manifest file.
func loadViteManifest(manifestPath string) (map[string]ViteAsset, error) {
	f, err := os.Open(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("cannot open vite manifest file: %w", err)
	}
	defer f.Close()

	var assets map[string]ViteAsset
	if err := json.NewDecoder(f).Decode(&assets); err != nil {
		return nil, fmt.Errorf("cannot unmarshal vite manifest to json: %w", err)
	}
	return assets, nil
}

// viteJSProd returns a function that provides the path to a JS asset in production.
func viteJSProd(assets map[string]ViteAsset, buildDir string) func(entry string) (template.HTML, error) {
	return func(entry string) (template.HTML, error) {
		asset, ok := assets[entry]
		if !ok {
			return "", fmt.Errorf("asset %q not found in manifest", entry)
		}
		jsLink := path.Join("/", buildDir, asset.File)
		var htmlLink strings.Builder
		htmlLink.WriteString(fmt.Sprintf(`<script type="module" src="%s"></script>`, jsLink))
		return template.HTML(htmlLink.String()), nil
	}
}

// viteCSSProd returns a function that generates <link> tags for an entry's CSS assets.
func viteCSSProd(assets map[string]ViteAsset, buildDir string) func(entry string) (template.HTML, error) {
	return func(entry string) (template.HTML, error) {
		asset, ok := assets[entry]
		if !ok {
			return "", fmt.Errorf("asset %q not found in manifest", entry)
		}

		var htmlLinks strings.Builder
		for _, cssFile := range asset.CSS {
			cssPath := path.Join("/", buildDir, cssFile)
			htmlLinks.WriteString(fmt.Sprintf(`<link rel="stylesheet" href="%s">`, cssPath))
		}

		return template.HTML(htmlLinks.String()), nil
	}
}