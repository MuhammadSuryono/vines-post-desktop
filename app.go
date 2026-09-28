package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	stdruntime "runtime"
	"strings"
	"sync"
	"time"
	"vines-pos-desktop/printer"

	"github.com/minio/selfupdate"
	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx           context.Context
	config        *AppConfig
	printInternal *menu.MenuItem
	printPlugin   *menu.MenuItem
	printMu       sync.Mutex
	printServers  []*http.Server
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
	if a.config.PrintMode != "plugin" {
		if err := a.startPrintServer(); err != nil {
			log.Printf("print agent: %v", err)
		}
	}
}

func (a *App) printMux() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/print", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var data PrinterLine
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": a.PrintReceipt(data)})
	})
	return mux
}

func (a *App) startPrintServer() error {
	a.printMu.Lock()
	defer a.printMu.Unlock()
	if len(a.printServers) > 0 {
		return nil
	}

	mux := a.printMux()
	var started []*http.Server
	for _, addr := range []string{"127.0.0.1:7081", "[::1]:7081"} {
		ln, err := net.Listen("tcp", addr)
		if err != nil {
			for _, srv := range started {
				_ = srv.Close()
			}
			return fmt.Errorf("port 7081 di %s sedang dipakai", addr)
		}
		srv := &http.Server{Handler: mux}
		started = append(started, srv)
		go func(srv *http.Server, ln net.Listener) {
			if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
				log.Printf("print agent: %v", err)
			}
		}(srv, ln)
		log.Printf("print agent listening on %s", addr)
	}
	a.printServers = started
	return nil
}

func (a *App) stopPrintServer() {
	a.printMu.Lock()
	servers := a.printServers
	a.printServers = nil
	a.printMu.Unlock()
	for _, srv := range servers {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		_ = srv.Shutdown(ctx)
		cancel()
	}
}

func (a *App) usePrintMode(mode string) {
	if mode != "plugin" {
		mode = "internal"
	}
	if mode == "plugin" {
		a.stopPrintServer()
	} else if err := a.startPrintServer(); err != nil {
		a.syncPrintMenu()
		_, _ = runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
			Type:    runtime.ErrorDialog,
			Title:   "Cetak internal",
			Message: err.Error() + "\nMatikan printer-plugin dulu, lalu pilih internal lagi.",
		})
		return
	}
	a.config.PrintMode = mode
	_ = a.config.Save()
	a.syncPrintMenu()

	msg := "Cetak lewat desktop. Port 7081 dipegang aplikasi ini."
	if mode == "plugin" {
		msg = "Cetak lewat plugin. Port 7081 sudah dilepas.\nJalankan printer-plugin, lalu pakai Test cetak."
	}
	_, _ = runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
		Type:    runtime.InfoDialog,
		Title:   "Pengaturan cetak",
		Message: msg,
	})
}

func (a *App) syncPrintMenu() {
	internal := a.config.PrintMode != "plugin"
	if a.printInternal != nil {
		a.printInternal.SetChecked(internal)
	}
	if a.printPlugin != nil {
		a.printPlugin.SetChecked(!internal)
	}
	if a.ctx != nil {
		runtime.MenuUpdateApplicationMenu(a.ctx)
	}
}

func (a *App) TestSampleReceipt() {
	client := &http.Client{Timeout: 20 * time.Second}
	req, err := http.NewRequest(http.MethodPost, "http://127.0.0.1:7081/api/v1/print?action=print", bytes.NewReader(sampleReceiptBody))
	if err != nil {
		a.alertPrint(runtime.ErrorDialog, err.Error())
		return
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	resp, err := client.Do(req)
	if err != nil {
		a.alertPrint(runtime.ErrorDialog, "Tidak ada yang menjawab di port 7081.\n"+err.Error())
		return
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	kind := runtime.InfoDialog
	if resp.StatusCode >= 400 {
		kind = runtime.ErrorDialog
	}
	a.alertPrint(kind, fmt.Sprintf("HTTP %d\n%s", resp.StatusCode, strings.TrimSpace(string(body))))
}

func (a *App) alertPrint(kind runtime.DialogType, message string) {
	_, _ = runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
		Type:    kind,
		Title:   "Test cetak",
		Message: message,
	})
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
		Data    receiptFields `json:"data"`
		UseDash bool          `json:"use_dash"`
	} `json:"description_line"`
	ItemLine []ItemLine `json:"item_line"`
	Others   []struct {
		Data    receiptFields `json:"data"`
		UseDash bool          `json:"use_dash"`
	} `json:"others"`
	Notes       string `json:"notes"`
	PrinterName string `json:"printer_name"`
}

// receiptFields keeps JSON object key order so the slip matches the payload.
type receiptFields []receiptField

type receiptField struct {
	Key   string
	Value string
}

func (f *receiptFields) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	tok, err := dec.Token()
	if err != nil {
		return err
	}
	delim, ok := tok.(json.Delim)
	if !ok || delim != '{' {
		return fmt.Errorf("receipt fields: expected object")
	}
	for dec.More() {
		keyTok, err := dec.Token()
		if err != nil {
			return err
		}
		key, ok := keyTok.(string)
		if !ok {
			return fmt.Errorf("receipt fields: expected string key")
		}
		var raw json.RawMessage
		if err := dec.Decode(&raw); err != nil {
			return err
		}
		val := strings.Trim(string(raw), `"`)
		*f = append(*f, receiptField{Key: key, Value: val})
	}
	_, err = dec.Token()
	return err
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
	socket, err := openPrinter(printerLine.PrinterName)
	if err != nil {
		return err
	}
	defer socket.Close()

	w := bufio.NewWriter(socket)
	p := printer.New(w)

	p.Verbose = true
	p.Init()

	a.setHeaderNota(p, printerLine)
	p.SetAlign("left")

	writeFields(p, printerLine.DescriptionLine.Data)
	if printerLine.DescriptionLine.UseDash && len(printerLine.DescriptionLine.Data) > 0 {
		p.DashLine()
	}

	for _, v := range printerLine.ItemLine {
		p.SetEmphasize(1)
		p.Write(v.ItemName + "\n")
		p.SetEmphasize(0)
		p.Write(fmt.Sprintf("%s  %18s\n", fmt.Sprintf("%s x @%s", v.TotalUnit, v.Price), v.TotalPrice))
		p.NewLine()
	}

	for _, block := range printerLine.Others {
		writeFields(p, block.Data)
		if block.UseDash && len(block.Data) > 0 {
			p.DashLine()
		}
	}

	if strings.TrimSpace(printerLine.Notes) != "" {
		p.Write(printerLine.Notes)
		if !strings.HasSuffix(printerLine.Notes, "\n") {
			p.Write("\n")
		}
	}

	p.Pulse()
	p.Cut()
	p.Write("\n\n\n")
	return w.Flush()
}

func writeFields(p *printer.Printer, fields receiptFields) {
	const width = 32
	for _, field := range fields {
		key, val := field.Key, field.Value
		if val == "" {
			p.Write(key + "\n")
			continue
		}
		if len(key)+1+len(val) > width {
			p.Write(key + "\n")
			if len(val) >= width {
				p.Write(val + "\n")
			} else {
				p.Write(strings.Repeat(" ", width-len(val)) + val + "\n")
			}
			continue
		}
		p.Write(key + strings.Repeat(" ", width-len(key)-len(val)) + val + "\n")
	}
}

func (a *App) setHeaderNota(p *printer.Printer, printerLine PrinterLine) {
	p.SetAlign("center")
	p.SetFont("A")
	// 2x2, bukan 2x3. Tinggi 3 kali dibuang oleh firmware 58 mm seperti EcoPrint.
	p.SetFontSize(2, 2)
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
	socket, errSocket := openPrinter(printerName)
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
