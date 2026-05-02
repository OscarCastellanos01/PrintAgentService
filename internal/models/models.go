package models

type StatusResponse struct {
	Status string `json:"status"`
	Message string `json:"message"`
}

type Printer struct {
	Name string `json:"name"`
	Driver string `json:"driver"`
	Port string `json:"port"`
	IsDefault bool `json:"is_default"`
}

type PrintPDFRequest struct {
	PrinterName string `json:"printer_name"`
	UseDefaultPrinter bool `json:"use_default_printer"`
	PDFBase64 string `json:"pdf_base64"`
	Copies int `json:"copies"`
}

type PrintRawRequest struct {
	PrinterName string `json:"printer_name"`
	UseDefaultPrinter bool `json:"use_default_printer"`
	Raw string `json:"raw"`
	Copies int `json:"copies"`
}
