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

func getConfigPath() string {
	exePath, _ := os.Executable()
	return filepath.Join(filepath.Dir(exePath), "config.json")
}

func loadConfig() AppConfig {
	// Default
	config := AppConfig{
		Mode:      "remote",
		RemoteURL: "", // Dikosongkan agar user diminta input pertama kali
		Version:   "1.0.0",
	}

	data, err := os.ReadFile(getConfigPath())
	if err == nil {
		json.Unmarshal(data, &config)
	}
	return config
}

func (c *AppConfig) Save() error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(getConfigPath(), data, 0644)
}

func main() {
	config := loadConfig()
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
