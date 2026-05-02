package handlers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"print-agent/internal/models"
	"print-agent/internal/printer"
)

func HandleStatus(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, models.StatusResponse{
		Status: "running",
		Message: "Print Agent funcionando correctamente",
	})
}

func HandlePrinters(w http.ResponseWriter, r *http.Request) {
	printers, err := printer.GetPrinters()
	if err != nil {
		slog.Error("error enumerando impresoras", "err", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"printers": printers})
}

func HandlePrintPDF(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "Método no permitido"})
		return
	}

	if !isJSON(r) {
		writeJSON(w, http.StatusUnsupportedMediaType, map[string]string{"error": "Content-Type debe ser application/json"})
		return
	}

	var req models.PrintPDFRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "JSON inválido"})
		return
	}

	req.PDFBase64 = printer.CleanBase64(req.PDFBase64)

	if req.PDFBase64 == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "pdf_base64 es requerido"})
		return
	}

	printerName, err := printer.ResolvePrinterName(req.PrinterName, req.UseDefaultPrinter)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	if req.Copies <= 0 {
		req.Copies = 1
	}

	pdfBytes, err := base64.StdEncoding.DecodeString(req.PDFBase64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "pdf_base64 inválido"})
		return
	}

	tempDir := filepath.Join(os.TempDir(), "PrintAgent")

	if err := os.MkdirAll(tempDir, 0755); err != nil {
		slog.Error("error creando directorio temporal", "err", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	pdfPath := filepath.Join(tempDir, fmt.Sprintf("print-%d.pdf", time.Now().UnixNano()))

	if err := os.WriteFile(pdfPath, pdfBytes, 0644); err != nil {
		slog.Error("error escribiendo PDF temporal", "err", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	defer os.Remove(pdfPath)

	if err := printer.PrintPDFWithSumatra(printerName, pdfPath, req.Copies); err != nil {
		slog.Error("error imprimiendo PDF", "printer", printerName, "err", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	slog.Info("PDF impreso", "printer", printerName, "copies", req.Copies)
	writeJSON(w, http.StatusOK, map[string]string{"message": "PDF enviado correctamente"})
}

func HandlePrintZPLRaw(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "Método no permitido"})
		return
	}

	if !isJSON(r) {
		writeJSON(w, http.StatusUnsupportedMediaType, map[string]string{"error": "Content-Type debe ser application/json"})
		return
	}

	var req models.PrintRawRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "JSON inválido"})
		return
	}

	req.Raw = strings.TrimSpace(req.Raw)

	if req.Raw == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "raw es requerido"})
		return
	}

	if !strings.Contains(req.Raw, "^XA") || !strings.Contains(req.Raw, "^XZ") {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "El raw no parece ZPL válido. Debe contener ^XA y ^XZ"})
		return
	}

	printerName, err := printer.ResolvePrinterName(req.PrinterName, req.UseDefaultPrinter)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	if req.Copies <= 0 {
		req.Copies = 1
	}

	if err := printer.RawPrintCopies(printerName, []byte(req.Raw), req.Copies); err != nil {
		slog.Error("error enviando ZPL RAW", "printer", printerName, "err", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	slog.Info("ZPL RAW enviado", "printer", printerName, "copies", req.Copies)
	writeJSON(w, http.StatusOK, map[string]string{"message": "ZPL RAW enviado correctamente"})
}

func isJSON(r *http.Request) bool {
	ct := r.Header.Get("Content-Type")
	return strings.HasPrefix(ct, "application/json")
}

func writeJSON(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}
