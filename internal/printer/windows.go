package printer

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"

	"print-agent/internal/config"
	"print-agent/internal/models"
)

const (
	printerEnumLocal = 0x00000002
	printerEnumConnections = 0x00000004
)

type docInfo1 struct {
	pDocName *uint16
	pOutputFile *uint16
	pDatatype *uint16
}

type printerInfo2 struct {
	pServerName *uint16
	pPrinterName *uint16
	pShareName *uint16
	pPortName *uint16
	pDriverName *uint16
	pComment *uint16
	pLocation *uint16
	pDevMode uintptr
	pSepFile *uint16
	pPrintProcessor *uint16
	pDatatype *uint16
	pParameters *uint16
	pSecurityDescriptor uintptr
	Attributes uint32
	Priority uint32
	DefaultPriority uint32
	StartTime uint32
	UntilTime uint32
	Status uint32
	cJobs uint32
	AveragePPM uint32
}

var (
	winspool = windows.NewLazySystemDLL("winspool.drv")
	procOpenPrinter = winspool.NewProc("OpenPrinterW")
	procClosePrinter = winspool.NewProc("ClosePrinter")
	procStartDocPrinter = winspool.NewProc("StartDocPrinterW")
	procEndDocPrinter = winspool.NewProc("EndDocPrinter")
	procStartPagePrinter = winspool.NewProc("StartPagePrinter")
	procEndPagePrinter = winspool.NewProc("EndPagePrinter")
	procWritePrinter = winspool.NewProc("WritePrinter")
	procEnumPrinters = winspool.NewProc("EnumPrintersW")
	procGetDefaultPrinter = winspool.NewProc("GetDefaultPrinterW")
)

func GetPrinters() ([]models.Printer, error) {
	defaultPrinter, err := GetDefaultPrinterName()
	if err != nil {
		defaultPrinter = ""
	}

	var needed uint32
	var returned uint32

	flags := uint32(printerEnumLocal | printerEnumConnections)

	r1, _, _ := procEnumPrinters.Call(
		uintptr(flags),
		0,
		uintptr(2),
		0,
		0,
		uintptr(unsafe.Pointer(&needed)),
		uintptr(unsafe.Pointer(&returned)),
	)

	if r1 == 0 && needed == 0 {
		return []models.Printer{}, nil
	}

	buffer := make([]byte, needed)

	r1, _, err = procEnumPrinters.Call(
		uintptr(flags),
		0,
		uintptr(2),
		uintptr(unsafe.Pointer(&buffer[0])),
		uintptr(needed),
		uintptr(unsafe.Pointer(&needed)),
		uintptr(unsafe.Pointer(&returned)),
	)

	if r1 == 0 {
		return nil, fmt.Errorf("no se pudieron enumerar las impresoras: %v", err)
	}

	if returned == 0 {
		return []models.Printer{}, nil
	}

	infoList := unsafe.Slice((*printerInfo2)(unsafe.Pointer(&buffer[0])), returned)

	printers := make([]models.Printer, 0, returned)

	for _, item := range infoList {
		name := utf16PtrToString(item.pPrinterName)
		if name == "" {
			continue
		}

		printers = append(printers, models.Printer{
			Name: name,
			Driver: utf16PtrToString(item.pDriverName),
			Port: utf16PtrToString(item.pPortName),
			IsDefault: strings.EqualFold(name, defaultPrinter),
		})
	}

	return printers, nil
}

func GetDefaultPrinterName() (string, error) {
	var size uint32

	r1, _, err := procGetDefaultPrinter.Call(
		0,
		uintptr(unsafe.Pointer(&size)),
	)

	if r1 == 0 && size == 0 {
		return "", fmt.Errorf("no se pudo obtener el tamaño del nombre de la impresora predeterminada: %v", err)
	}

	buffer := make([]uint16, size)

	r1, _, err = procGetDefaultPrinter.Call(
		uintptr(unsafe.Pointer(&buffer[0])),
		uintptr(unsafe.Pointer(&size)),
	)

	if r1 == 0 {
		return "", fmt.Errorf("no se pudo obtener la impresora predeterminada: %v", err)
	}

	return windows.UTF16ToString(buffer), nil
}

func ResolvePrinterName(printerName string, useDefault bool) (string, error) {
	printerName = strings.TrimSpace(printerName)

	if printerName != "" {
		return printerName, nil
	}

	if !useDefault {
		return "", errors.New("printer_name es requerido o debe enviar use_default_printer en true")
	}

	defaultPrinter, err := GetDefaultPrinterName()
	if err != nil {
		return "", err
	}

	defaultPrinter = strings.TrimSpace(defaultPrinter)

	if defaultPrinter == "" {
		return "", errors.New("no se encontró impresora predeterminada")
	}

	return defaultPrinter, nil
}

func RawPrint(printerName string, data []byte) error {
	if len(data) == 0 {
		return errors.New("no hay datos para imprimir")
	}

	pName, err := windows.UTF16PtrFromString(printerName)
	if err != nil {
		return err
	}

	var hPrinter windows.Handle

	r1, _, err := procOpenPrinter.Call(
		uintptr(unsafe.Pointer(pName)),
		uintptr(unsafe.Pointer(&hPrinter)),
		0,
	)

	if r1 == 0 {
		return fmt.Errorf("no se pudo abrir la impresora %s: %v", printerName, err)
	}

	defer procClosePrinter.Call(uintptr(hPrinter))

	docName, err := windows.UTF16PtrFromString("Print Agent RAW Job")
	if err != nil {
		return err
	}

	dataType, err := windows.UTF16PtrFromString("RAW")
	if err != nil {
		return err
	}

	docInfo := docInfo1{
		pDocName: docName,
		pOutputFile: nil,
		pDatatype: dataType,
	}

	r1, _, err = procStartDocPrinter.Call(
		uintptr(hPrinter),
		1,
		uintptr(unsafe.Pointer(&docInfo)),
	)

	if r1 == 0 {
		return fmt.Errorf("no se pudo iniciar el documento de impresión: %v", err)
	}

	defer procEndDocPrinter.Call(uintptr(hPrinter))

	r1, _, err = procStartPagePrinter.Call(uintptr(hPrinter))
	if r1 == 0 {
		return fmt.Errorf("no se pudo iniciar la página de impresión: %v", err)
	}

	defer procEndPagePrinter.Call(uintptr(hPrinter))

	var written uint32

	r1, _, err = procWritePrinter.Call(
		uintptr(hPrinter),
		uintptr(unsafe.Pointer(&data[0])),
		uintptr(len(data)),
		uintptr(unsafe.Pointer(&written)),
	)

	if r1 == 0 {
		return fmt.Errorf("no se pudo escribir en la impresora: %v", err)
	}

	if int(written) != len(data) {
		return fmt.Errorf("impresión incompleta: enviados %d de %d bytes", written, len(data))
	}

	return nil
}

func RawPrintCopies(printerName string, data []byte, copies int) error {
	for i := 1; i <= copies; i++ {
		if err := RawPrint(printerName, data); err != nil {
			return fmt.Errorf("error enviando ZPL RAW copia %d de %d: %v", i, copies, err)
		}
		if i < copies {
			time.Sleep(config.ZPLCopyDelay)
		}
	}
	return nil
}

func PrintPDFWithSumatra(printerName string, pdfPath string, copies int) error {
	basePath, err := getExecutableDir()
	if err != nil {
		return err
	}

	sumatraPath := filepath.Join(basePath, "SumatraPDF.exe")

	if _, err := os.Stat(sumatraPath); err != nil {
		return fmt.Errorf("no se encontró SumatraPDF.exe en %s", sumatraPath)
	}

	if copies <= 0 {
		copies = 1
	}

	for i := 1; i <= copies; i++ {
		cmd := exec.Command(
			sumatraPath,
			"-print-to", printerName,
			"-print-settings", "noscale",
			"-silent",
			pdfPath,
		)

		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("error imprimiendo PDF copia %d de %d: %v - %s", i, copies, err, string(output))
		}

		if i < copies {
			time.Sleep(5 * time.Second)
		}
	}

	return nil
}

func CleanBase64(value string) string {
	value = strings.TrimSpace(value)
	if strings.Contains(value, ",") {
		parts := strings.SplitN(value, ",", 2)
		return strings.TrimSpace(parts[1])
	}
	return value
}

func getExecutableDir() (string, error) {
	exePath, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Dir(exePath), nil
}

func utf16PtrToString(ptr *uint16) string {
	if ptr == nil {
		return ""
	}
	return windows.UTF16PtrToString(ptr)
}
