package main
import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
	"vines-pos-desktop/printer"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx    context.Context
	config AppConfig
}

// NewApp creates a new App application struct
func NewApp(config AppConfig) *App {
	return &App{
		config: config,
	}
}

// ShowUpdatePrompt memunculkan dialog native OS agar tidak hilang saat redirect web
func (a *App) ShowUpdatePrompt(version string, url string) {
	msg := fmt.Sprintf("Versi baru (%s) telah tersedia.\nApakah Anda ingin mendownloadnya sekarang?", version)

	result, err := runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
		Type:          runtime.QuestionDialog,
		Title:         "Update Tersedia",
		Message:       msg,
		DefaultButton: "Ya",
		CancelButton:  "Nanti",
	})

	if err == nil && result == "Ya" {
		runtime.BrowserOpenURL(a.ctx, url)
	}
}

type GitHubRelease struct {
	TagName string `json:"tag_name"`
	Name    string `json:"name"`
	HTMLURL string `json:"html_url"`
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
