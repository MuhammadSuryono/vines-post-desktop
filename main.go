package main

import (
	"embed"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

// AppConfig mendefinisikan struktur konfigurasi untuk aplikasi
type AppConfig struct {
	Mode      string `json:"mode"`       // "remote" atau "local"
	RemoteURL string `json:"remote_url"` // URL jika mode = "remote"
	Version   string `json:"version"`    // Versi aplikasi desktop
}

func loadConfig() AppConfig {
	// 1. Konfigurasi Default (Remote)
	config := AppConfig{
		Mode:      "remote",
		RemoteURL: "http://45.64.97.50:888/thevines/index.php",
		Version:   "1.0.0",
	}

	// 2. Coba baca config.json di folder yang sama dengan executable
	exePath, err := os.Executable()
	if err == nil {
		configPath := filepath.Join(filepath.Dir(exePath), "config.json")
		data, err := os.ReadFile(configPath)
		if err == nil {
			// Jika file config.json ditemukan, timpa konfigurasi default
			json.Unmarshal(data, &config)
		}
	}
	return config
}

func main() {
	config := loadConfig()
	// Create an instance of the app structure
	app := NewApp(config)

	// Base App Options
	appOptions := &options.App{
		Title:  "Vines POS Desktop (v" + config.Version + ")",
		Width:  1280,
		Height: 800,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 255, G: 255, B: 255, A: 1},
		OnStartup:        app.startup,
		OnDomReady:       app.domReady,
		Bind: []interface{}{
			app,
		},
	}

	// Create application with options
	err := wails.Run(appOptions)

	if err != nil {
		println("Error:", err.Error())
	}
}
