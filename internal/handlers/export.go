
package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/jung-kurt/gofpdf"
)

type ExportData struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	Format  string `json:"format,omitempty"`
}

func (h *Handler) ExportPDF(w http.ResponseWriter, r *http.Request) {
	var data ExportData
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid JSON", 400)
		return
	}

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(0, 10, data.Title)
	pdf.Ln(12)
	pdf.SetFont("Arial", "", 10)
	// Simple multi-cell
	pdf.MultiCell(0, 6, data.Content, "", "", false)
	pdf.Ln(6)
	pdf.SetFont("Arial", "I", 8)
	pdf.Cell(0, 10, fmt.Sprintf("Generated: %s | WebX Metrics Pro Go v%s", time.Now().Format(time.RFC3339), h.Config.App.Version))

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		http.Error(w, "PDF generation failed", 500)
		return
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "attachment; filename=\"report.pdf\"")
	w.Write(buf.Bytes())
}

func (h *Handler) ExportMD(w http.ResponseWriter, r *http.Request) {
	var data ExportData
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid JSON", 400)
		return
	}
	md := fmt.Sprintf("# %s\n\n%s\n\n---\nGenerated: %s\nVersion: %s\n", data.Title, data.Content, time.Now().Format(time.RFC3339), h.Config.App.Version)
	w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=\"report.md\"")
	w.Write([]byte(md))
}

func (h *Handler) ExportTXT(w http.ResponseWriter, r *http.Request) {
	var data ExportData
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid JSON", 400)
		return
	}
	txt := fmt.Sprintf("%s\n%s\n\n%s\n\nGenerated at %s\n", data.Title, repeat("=", len(data.Title)), data.Content, time.Now().Format(time.RFC3339))
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=\"report.txt\"")
	w.Write([]byte(txt))
}

func (h *Handler) ExportJSON(w http.ResponseWriter, r *http.Request) {
	var data ExportData
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid JSON", 400)
		return
	}
	payload := map[string]interface{}{
		"title":        data.Title,
		"content":      data.Content,
		"generated_at": time.Now().Format(time.RFC3339),
		"version":      h.Config.App.Version,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(payload)
}

func repeat(s string, n int) string {
	result := ""
	for i := 0; i < n; i++ {
		result += s
	}
	return result
}
