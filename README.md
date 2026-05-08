# Print Agent

Servicio local desarrollado en Go para imprimir desde aplicaciones web mediante una API HTTP local.

Permite:

- Listar impresoras instaladas en Windows.
- Imprimir PDF enviado en base64.
- Enviar comandos ZPL RAW a impresoras Zebra o compatibles.
- Imprimir por nombre de impresora o usando la impresora predeterminada.

## Requisitos

- Go
- Windows
- PowerShell
- SumatraPDF.exe para impresión PDF

`SumatraPDF.exe` debe estar en la misma carpeta que `print-agent.exe`.

Estructura del ejecutable instalado:

```txt
Print Agent/
├── print-agent.exe
└── SumatraPDF.exe
```

Estructura del proyecto fuente:

```txt
print-agent/
├── main.go
├── go.mod
├── go.sum
├── build/
│   ├── print-agent.exe
│   └── SumatraPDF.exe
├── installer/
│   └── print-agent.iss
└── internal/
    ├── config/
    │   └── config.go
    ├── handlers/
    │   └── handlers.go
    ├── middleware/
    │   └── middleware.go
    ├── models/
    │   └── models.go
    └── printer/
        └── windows.go
```

## Instalar dependencias

```bash
go get github.com/kardianos/service
go get golang.org/x/sys/windows
go mod tidy
```

## Compilar

```powershell
go build -ldflags="-s -w -H=windowsgui" -o build/print-agent.exe .
```

## Ejecutar en consola

```powershell
.\print-agent.exe
```

El servicio queda disponible en:

```txt
http://127.0.0.1:9876
```

## Instalar como servicio de Windows

Abrir PowerShell como administrador.

Instalar el servicio:

```powershell
.\print-agent.exe install
```

Iniciar el servicio:

```powershell
.\print-agent.exe start
```

Detener el servicio:

```powershell
.\print-agent.exe stop
```

Reiniciar el servicio:

```powershell
.\print-agent.exe restart
```

Desinstalar el servicio:

```powershell
.\print-agent.exe uninstall
```

Verificar estado del servicio:

```powershell
Get-Service PrintAgentService
```

## Endpoints

### GET /status

Verifica si el agente está activo.

```http
GET http://127.0.0.1:9876/status
```

Respuesta:

```json
{
  "status": "running",
  "message": "Print Agent funcionando correctamente"
}
```

### GET /printers

Lista las impresoras instaladas.

```http
GET http://127.0.0.1:9876/printers
```

Respuesta:

```json
{
  "printers": [
    {
      "name": "POS Printer",
      "driver": "Generic Thermal Printer",
      "port": "USB001",
      "is_default": true
    },
    {
      "name": "Zebra Label Printer",
      "driver": "ZDesigner ZD220",
      "port": "USB002",
      "is_default": false
    }
  ]
}
```

### POST /print/pdf

Imprime un PDF enviado en base64.

```http
POST http://127.0.0.1:9876/print/pdf
Content-Type: application/json
```

Body usando nombre de impresora:

```json
{
  "printer_name": "POS Printer",
  "pdf_base64": "JVBERi0xLjQKJc...",
  "copies": 1
}
```

Body usando impresora predeterminada:

```json
{
  "use_default_printer": true,
  "pdf_base64": "JVBERi0xLjQKJc...",
  "copies": 1
}
```

Respuesta:

```json
{
  "message": "PDF enviado correctamente"
}
```

### POST /print/zpl/raw

Envía comandos ZPL RAW a una impresora Zebra o compatible.

```http
POST http://127.0.0.1:9876/print/zpl/raw
Content-Type: application/json
```

Body usando nombre de impresora:

```json
{
  "printer_name": "Zebra Label Printer",
  "raw": "^XA^FO50,50^FDHola desde Print Agent^FS^XZ",
  "copies": 1
}
```

Body usando impresora predeterminada:

```json
{
  "use_default_printer": true,
  "raw": "^XA^FO50,50^FDHola desde Print Agent^FS^XZ",
  "copies": 1
}
```

Respuesta:

```json
{
  "message": "ZPL RAW enviado correctamente"
}
```

### POST /print/esc-pos

Envía bytes ESC/POS crudos a una impresora térmica. Los datos van en el body como `application/octet-stream` y los parámetros como query string.

```http
POST http://127.0.0.1:9876/print/esc-pos?printer_name=EPSON&copies=1
Content-Type: application/octet-stream
```

Con impresora predeterminada:

```http
POST http://127.0.0.1:9876/print/esc-pos?use_default_printer=true&copies=1
Content-Type: application/octet-stream
```

Parámetros query string:

| Parámetro | Tipo | Descripción |
|---|---|---|
| `printer_name` | string | Nombre exacto de la impresora |
| `use_default_printer` | boolean | Usar impresora predeterminada (`true`) |
| `copies` | number | Cantidad de copias (default: 1) |

El body debe contener los bytes ESC/POS generados por la librería. No se acepta JSON ni base64.

Respuesta:

```json
{
  "message": "ESC/POS enviado correctamente"
}
```

## Uso desde JavaScript

### Imprimir PDF

```js
await fetch("http://127.0.0.1:9876/print/pdf", {
  method: "POST",
  headers: {
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    printer_name: "POS Printer",
    pdf_base64: pdfBase64,
    copies: 1
  })
});
```

### Imprimir PDF usando impresora predeterminada

```js
await fetch("http://127.0.0.1:9876/print/pdf", {
  method: "POST",
  headers: {
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    use_default_printer: true,
    pdf_base64: pdfBase64,
    copies: 1
  })
});
```

### Imprimir ZPL

```js
await fetch("http://127.0.0.1:9876/print/zpl/raw", {
  method: "POST",
  headers: {
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    printer_name: "Zebra Label Printer",
    raw: "^XA^FO50,50^FDHola desde Print Agent^FS^XZ",
    copies: 1
  })
});
```

### Imprimir ZPL usando impresora predeterminada

```js
await fetch("http://127.0.0.1:9876/print/zpl/raw", {
  method: "POST",
  headers: {
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    use_default_printer: true,
    raw: "^XA^FO50,50^FDHola desde Print Agent^FS^XZ",
    copies: 1
  })
});
```

### Imprimir ESC/POS

```js
const ESC = 0x1B
const GS = 0x1D

const bytes = new Uint8Array([
  ESC, 0x40,
  ESC, 0x61, 0x01,
  ...new TextEncoder().encode("MI TIENDA\n"),
  ESC, 0x61, 0x00,
  ...new TextEncoder().encode("Producto 1       $10.00\n"),
  0x0A, 0x0A, 0x0A,
  GS, 0x56, 0x01,
])

await fetch("http://127.0.0.1:9876/print/esc-pos?use_default_printer=true&copies=1", {
  method: "POST",
  headers: { "Content-Type": "application/octet-stream" },
  body: bytes
})
```

## Configuración

Por defecto el servicio escucha en `127.0.0.1:9876`. Si ese puerto está ocupado, se puede cambiar sin recompilar usando la variable de entorno `PRINT_AGENT_ADDR` antes de iniciar el servicio:

```powershell
$env:PRINT_AGENT_ADDR = "127.0.0.1:9877"
.\build\print-agent.exe
```

## Notas

- Para PDF, `SumatraPDF.exe` debe estar junto a `print-agent.exe`.
- Para tickets térmicos, el PDF debe generarse con el tamaño adecuado para la impresora.
- El endpoint `/print/zpl/raw` debe usarse con impresoras Zebra o compatibles con ZPL.
- Si se envía ZPL a una impresora térmica ESC/POS, no imprimirá correctamente.
- El servicio escucha únicamente en `127.0.0.1`.
