package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	stdruntime "runtime"
	"strings"
	"time"
	"vines-pos-desktop/printer"

	"github.com/minio/selfupdate"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx    context.Context
	config *AppConfig
}

// NewApp creates a new App application struct
func NewApp(config *AppConfig) *App {
	return &App{
		config: config,
	}
}

// GetConfig mengembalikan konfigurasi saat ini ke frontend
func (a *App) GetConfig() AppConfig {
	return *a.config
}

// SaveURL menyimpan URL baru ke dalam file config.json
func (a *App) SaveURL(newURL string) string {
	a.config.RemoteURL = newURL
	err := a.config.Save()
	if err != nil {
		return "Error: " + err.Error()
	}
	return "Success"
}

// ShowUpdatePrompt memunculkan dialog native OS agar tidak hilang saat redirect web
func (a *App) ShowUpdatePrompt(version string, release GitHubRelease) {
	msg := fmt.Sprintf("Versi baru (%s) telah tersedia.\nApakah Anda ingin mendownload dan menginstallnya secara otomatis?", version)

	result, err := runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
		Type:          runtime.QuestionDialog,
		Title:         "Update Tersedia",
		Message:       msg,
		DefaultButton: "Yes",
		CancelButton:  "No",
		Buttons:       []string{"Yes", "No"},
	})

	if err == nil && result == "Yes" {
		go a.StartUpdate(release)
	}
}

// StartUpdate mendownload dan menerapkan update secara otomatis
func (a *App) StartUpdate(release GitHubRelease) {
	// 1. Cari asset yang sesuai dengan OS dan Arch
	var downloadURL string
	osName := stdruntime.GOOS
	archName := stdruntime.GOARCH

	for _, asset := range release.Assets {
		name := strings.ToLower(asset.Name)
		// Cari yang cocok dengan OS (windows/darwin) dan Arch (amd64/arm64)
		if strings.Contains(name, osName) && strings.Contains(name, archName) {
			downloadURL = asset.BrowserDownloadURL
			break
		}
		// Fallback sederhana jika penamaan tidak menyertakan arch
		if strings.Contains(name, osName) && downloadURL == "" {
			if osName == "windows" && strings.HasSuffix(name, ".exe") {
				downloadURL = asset.BrowserDownloadURL
			} else if osName == "darwin" && !strings.HasSuffix(name, ".exe") {
				downloadURL = asset.BrowserDownloadURL
			}
		}
	}

	if downloadURL == "" {
		runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
			Type:    runtime.ErrorDialog,
			Title:   "Update Gagal",
			Message: "Tidak dapat menemukan file update yang sesuai untuk sistem Anda.",
		})
		return
	}

	// 2. Download asset
	resp, err := http.Get(downloadURL)
	if err != nil {
		runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
			Type:    runtime.ErrorDialog,
			Title:   "Download Gagal",
			Message: "Gagal mendownload update: " + err.Error(),
		})
		return
	}
	defer resp.Body.Close()

	// 3. Terapkan update
	err = selfupdate.Apply(resp.Body, selfupdate.Options{})
	if err != nil {
		runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
			Type:    runtime.ErrorDialog,
			Title:   "Install Gagal",
			Message: "Gagal menerapkan update: " + err.Error(),
		})
		return
	}

	// 4. Update versi di config.json sebelum restart
	a.config.Version = strings.TrimPrefix(release.TagName, "v")
	a.config.Save()

	// 5. Restart aplikasi
	runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
		Type:    runtime.InfoDialog,
		Title:   "Update Berhasil",
		Message: "Update telah terpasang ke versi " + release.TagName + ". Aplikasi akan dimulai ulang sekarang.",
	})

	self, err := os.Executable()
	if err == nil {
		cmd := exec.Command(self)
		cmd.Start()
		os.Exit(0)
	} else {
		runtime.Quit(a.ctx)
	}
}

type GitHubAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

type GitHubRelease struct {
	TagName string        `json:"tag_name"`
	Name    string        `json:"name"`
	HTMLURL string        `json:"html_url"`
	Assets  []GitHubAsset `json:"assets"`
}

// CheckUpdate membandingkan versi lokal dengan rilis terbaru di GitHub
func (a *App) CheckUpdate() map[string]interface{} {
	client := &http.Client{Timeout: 5 * time.Second}
	// Ganti dengan URL repo Anda
	url := "https://api.github.com/repos/MuhammadSuryono/vines-post-desktop/releases/latest"

	resp, err := client.Get(url)
	if err != nil {
		return map[string]interface{}{"update_available": false, "error": err.Error()}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return map[string]interface{}{"update_available": false, "status": resp.Status}
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return map[string]interface{}{"update_available": false, "error": "Failed to decode JSON"}
	}

	// Bandingkan: Jika Tag di GitHub (misal v1.0.1) != Versi Lokal
	updateAvailable := release.TagName != "v"+a.config.Version && release.TagName != a.config.Version

	return map[string]interface{}{
		"update_available": updateAvailable,
		"latest_version":   release.TagName,
		"current_version":  a.config.Version,
		"url":              release.HTMLURL,
		"release":          release,
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// domReady is called when the frontend has loaded its initial assets
func (a *App) domReady(ctx context.Context) {
	// Logika redirect dipindahkan sepenuhnya ke frontend (main.js)
	// untuk menghindari infinite loop saat halaman remote dimuat.
}

// Printer Data Structures
type PrinterLine struct {
	HeaderLine      HeaderLine `json:"header_line"`
	DescriptionLine struct {
		Data    map[string]string `json:"data"`
		UseDash bool              `json:"use_dash"`
	} `json:"description_line"`
	ItemLine []ItemLine `json:"item_line"`
	Others   []struct {
		Data    map[string]string `json:"data"`
		UseDash bool              `json:"use_dash"`
	} `json:"others"`
	Notes       string `json:"notes"`
	PrinterName string `json:"printer_name"`
}

type HeaderLine struct {
	Header      string `json:"header"`
	Address     string `json:"address"`
	City        string `json:"city"`
	PhoneNumber string `json:"phone_number"`
	PortalCode  string `json:"portal_code"`
	UseDash     bool   `json:"use_dash"`
}

type ItemLine struct {
	ItemName   string `json:"item_name"`
	TotalUnit  string `json:"total_unit"`
	Price      string `json:"price"`
	TotalPrice string `json:"total_price"`
}

// PrintReceipt is the bridge method called from Frontend
func (a *App) PrintReceipt(data PrinterLine) string {
	err := a.executePrint(data)
	if err != nil {
		return fmt.Sprintf("Error: %s", err.Error())
	}
	return "Success"
}

func (a *App) executePrint(printerLine PrinterLine) error {
	name, err := os.Hostname()
	if err != nil {
		return err
	}

	printerName := printerLine.PrinterName
	// Path untuk Windows printer sharing
	path := "\\\\" + name + "\\" + printerName

	socket, errSocket := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0777)
	if errSocket != nil {
		return errSocket
	}
	defer socket.Close()

	w := bufio.NewWriter(socket)
	p := printer.New(w)

	p.Verbose = true
	p.Init()

	a.setHeaderNota(p, printerLine)

	// Tambahkan logika item jika diperlukan (opsional, disesuaikan dengan main.go lama)
	for _, v := range printerLine.ItemLine {
		p.SetEmphasize(1)
		p.Write(v.ItemName + "\n")
		p.SetEmphasize(0)
		p.Write(fmt.Sprintf("%s  %18s\n", fmt.Sprintf("%s x @%s", v.TotalUnit, v.Price), v.TotalPrice))
		p.NewLine()
	}

	// Buka laci kasir (Cash Drawer)
	p.Pulse()

	// Potong kertas
	p.Cut()
	w.Flush()

	return nil
}

func (a *App) setHeaderNota(p *printer.Printer, printerLine PrinterLine) {
	p.SetFontSize(2, 3)
	p.SetAlign("center")
	p.SetFont("A")
	p.Write(printerLine.HeaderLine.Header)
	p.NewLine()
	p.SetFontSize(1, 1)
	p.SetFont("A")
	p.Write(printerLine.HeaderLine.Address)
	p.NewLine()
	if printerLine.HeaderLine.City != "" {
		p.Write(printerLine.HeaderLine.City)
		p.NewLine()
	}
	if printerLine.HeaderLine.PhoneNumber != "" {
		p.Write(printerLine.HeaderLine.PhoneNumber)
		p.NewLine()
	}
	if printerLine.HeaderLine.PortalCode != "" {
		p.Write(printerLine.HeaderLine.PortalCode)
		p.NewLine()
	}
	if printerLine.HeaderLine.UseDash {
		p.DashLine()
		p.NewLine()
	}
	p.NewLine()
}

// TestPrint untuk ngetes printer dari UI
func (a *App) TestPrint(printerName string) string {
	name, _ := os.Hostname()
	socket, errSocket := os.OpenFile("\\\\"+name+"\\"+printerName, os.O_WRONLY|os.O_CREATE, 0)
	if errSocket != nil {
		return errSocket.Error()
	}
	defer socket.Close()
	w := bufio.NewWriter(socket)
	p := printer.New(w)

	p.Verbose = true
	p.Write("Test Print dari Wails Desktop\n")
	p.Init()
	p.Cut()
	w.Flush()
	return "Test Print Sent"
}
